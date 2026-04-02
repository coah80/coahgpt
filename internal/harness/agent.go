package harness

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultMaxTurns = 25

type AgentConfig struct {
	OllamaURL string // e.g. "http://localhost:11434"
	Model     string // e.g. "qwen2.5:7b"
	System    string // system prompt
	MaxTurns  int    // max tool-call loops before bailing
}

type Agent struct {
	registry  *Registry
	ollamaURL string
	model     string
	system    string
	maxTurns  int
	tokens    *TokenTracker
}

func NewAgent(config AgentConfig, registry *Registry) *Agent {
	maxTurns := config.MaxTurns
	if maxTurns <= 0 {
		maxTurns = defaultMaxTurns
	}
	return &Agent{
		registry:  registry,
		ollamaURL: config.OllamaURL,
		model:     config.Model,
		system:    config.System,
		maxTurns:  maxTurns,
		tokens:    NewTokenTracker(),
	}
}

// RunChat appends a user message to history and runs the agent loop.
func (a *Agent) RunChat(ctx context.Context, history []Message, userMessage string) <-chan Event {
	msgs := make([]Message, len(history), len(history)+1)
	copy(msgs, history)
	msgs = append(msgs, Message{Role: RoleUser, Content: userMessage})
	return a.Run(ctx, msgs)
}

// Run executes the agentic loop. Returns a channel of events for the UI to consume.
// The loop: call ollama -> stream tokens -> if tool calls, execute them -> feed results back -> repeat.
// Closes the channel when done.
func (a *Agent) Run(ctx context.Context, messages []Message) <-chan Event {
	events := make(chan Event, 64)

	go func() {
		defer close(events)
		a.loop(ctx, messages, events)
	}()

	return events
}

// GetTokenStats returns approximate input and output token counts for this session.
func (a *Agent) GetTokenStats() (input, output int) {
	return a.tokens.Stats()
}

// SetPermissionCallback sets the callback invoked when a tool needs user approval.
func (a *Agent) SetPermissionCallback(cb PermissionCallback) {
	a.registry.mu.Lock()
	if a.registry.permissions == nil {
		a.registry.permissions = DefaultPermissions()
	}
	a.registry.permissions.AskCallback = cb
	a.registry.mu.Unlock()
}

func (a *Agent) loop(ctx context.Context, messages []Message, events chan<- Event) {
	for turn := 0; turn < a.maxTurns; turn++ {
		if ctx.Err() != nil {
			events <- Event{Type: EventError, Content: "cancelled", Done: true}
			return
		}

		events <- Event{Type: EventLoopStart, Content: fmt.Sprintf("turn %d", turn+1)}

		assistantMsg, err := a.streamCompletion(ctx, messages, events)
		if err != nil {
			events <- Event{Type: EventError, Content: err.Error(), Done: true}
			return
		}

		if len(assistantMsg.ToolCalls) == 0 {
			in, out := a.tokens.Stats()
			events <- Event{Type: EventDone, Done: true, InputTokens: in, OutputTokens: out}
			return
		}

		// append the assistant message with tool calls
		messages = appendMessage(messages, assistantMsg)

		// execute tools and collect results
		toolResults := a.registry.ExecuteParallel(ctx, assistantMsg.ToolCalls)

		// append each tool result as a tool message
		for i, result := range toolResults {
			toolName := assistantMsg.ToolCalls[i].Function.Name
			content := result.Content
			if result.IsError {
				content = fmt.Sprintf("ERROR: %s", content)
			}
			messages = appendMessage(messages, Message{
				Role:    RoleTool,
				Content: content,
				ToolCalls: []ToolCall{{
					Function: ToolCallFunction{Name: toolName},
				}},
			})

			events <- Event{
				Type:     EventToolComplete,
				ToolName: toolName,
				Content:  truncateForEvent(content, 500),
			}
		}
	}

	events <- Event{
		Type:    EventError,
		Content: fmt.Sprintf("hit max turns (%d) without completing", a.maxTurns),
		Done:    true,
	}
}

// ollamaChatRequest is the request body for /api/chat
type ollamaChatRequest struct {
	Model    string                   `json:"model"`
	Messages []ollamaMessage          `json:"messages"`
	Tools    []map[string]interface{} `json:"tools,omitempty"`
	Stream   bool                     `json:"stream"`
}

// ollamaMessage maps to Ollama's message format
type ollamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
}

type ollamaToolCall struct {
	Function ollamaToolCallFunction `json:"function"`
}

type ollamaToolCallFunction struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ollamaStreamChunk is a single line from Ollama's streaming response
type ollamaStreamChunk struct {
	Message ollamaChunkMessage `json:"message"`
	Done    bool               `json:"done"`
}

type ollamaChunkMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
}

func (a *Agent) streamCompletion(ctx context.Context, messages []Message, events chan<- Event) (Message, error) {
	ollamaMsgs := convertMessages(messages)
	tools := a.registry.OllamaToolDefs()

	// count input tokens from all messages
	for _, m := range messages {
		a.tokens.AddInput(m.Content)
	}

	reqBody := ollamaChatRequest{
		Model:    a.model,
		Messages: ollamaMsgs,
		Tools:    tools,
		Stream:   true,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return Message{}, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.ollamaURL+"/api/chat", bytes.NewReader(bodyBytes))
	if err != nil {
		return Message{}, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return Message{}, fmt.Errorf("ollama returned %d: %s", resp.StatusCode, string(body))
	}

	var fullContent string
	var toolCalls []ToolCall

	scanner := bufio.NewScanner(resp.Body)
	// bump scanner buffer for big responses
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var chunk ollamaStreamChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			continue
		}

		if chunk.Message.Content != "" {
			fullContent += chunk.Message.Content
			events <- Event{
				Type:    EventToken,
				Content: chunk.Message.Content,
			}
		}

		if len(chunk.Message.ToolCalls) > 0 {
			for _, tc := range chunk.Message.ToolCalls {
				toolCalls = append(toolCalls, ToolCall{
					Function: ToolCallFunction{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				})
			}
		}

		if chunk.Done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return Message{}, fmt.Errorf("reading stream: %w", err)
	}

	// count output tokens from response
	a.tokens.AddOutput(fullContent)

	return Message{
		Role:      RoleAssistant,
		Content:   fullContent,
		ToolCalls: toolCalls,
	}, nil
}

func convertMessages(messages []Message) []ollamaMessage {
	out := make([]ollamaMessage, len(messages))
	for i, m := range messages {
		msg := ollamaMessage{
			Role:    string(m.Role),
			Content: m.Content,
		}
		for _, tc := range m.ToolCalls {
			msg.ToolCalls = append(msg.ToolCalls, ollamaToolCall{
				Function: ollamaToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}
		out[i] = msg
	}
	return out
}

func appendMessage(messages []Message, msg Message) []Message {
	result := make([]Message, len(messages), len(messages)+1)
	copy(result, messages)
	return append(result, msg)
}

func truncateForEvent(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
