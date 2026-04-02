package harness

import (
	"context"
	"encoding/json"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role      Role       `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ToolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	ReadOnly    bool
	Destructive bool
}

type ToolResult struct {
	Content string
	IsError bool
}

type ToolHandler func(ctx context.Context, args json.RawMessage) (ToolResult, error)

type EventType int

const (
	EventToken         EventType = iota // streaming token from model
	EventToolStart                      // tool execution starting
	EventToolProgress                   // tool progress (e.g. bash stdout)
	EventToolComplete                   // tool finished
	EventToolError                      // tool failed
	EventPermissionAsk                  // permission needed before tool execution
	EventLoopStart                      // new agent loop iteration
	EventDone                           // agent finished, no more tool calls
	EventError                          // fatal error
)

type Event struct {
	Type         EventType
	Content      string // token text, progress text, or error message
	ToolName     string // which tool (for tool events)
	ToolArgs     string // tool arguments (for tool events)
	Done         bool   // final event
	InputTokens  int    // approximate input tokens (set on EventDone)
	OutputTokens int    // approximate output tokens (set on EventDone)
}

type toolNameKey struct{}

func ContextWithToolName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, toolNameKey{}, name)
}

func ToolNameFromContext(ctx context.Context) string {
	if name, ok := ctx.Value(toolNameKey{}).(string); ok {
		return name
	}
	return ""
}
