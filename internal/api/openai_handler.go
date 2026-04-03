package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/coah80/coahgpt/internal/ollama"
	"github.com/coah80/coahgpt/internal/persona"
)

// OpenAI-compatible chat completions endpoint
// This allows coah code CLI (OpenCode fork) to use coahgpt.com as a provider

type openAIChatRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []openAIChoice `json:"choices"`
}

type openAIChoice struct {
	Index        int           `json:"index"`
	Message      openAIMessage `json:"message,omitempty"`
	Delta        openAIMessage `json:"delta,omitempty"`
	FinishReason *string       `json:"finish_reason"`
}

func HandleOpenAIChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":{"message":"method not allowed"}}`, http.StatusMethodNotAllowed)
		return
	}

	var req openAIChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"message":"invalid request"}}`, http.StatusBadRequest)
		return
	}

	// inject system prompt if not present
	hasSystem := false
	for _, m := range req.Messages {
		if m.Role == "system" {
			hasSystem = true
			break
		}
	}
	if !hasSystem {
		req.Messages = append([]openAIMessage{{Role: "system", Content: persona.ChatPrompt}}, req.Messages...)
	}

	// convert to Ollama format
	ollamaMsgs := make([]ollamaMsg, len(req.Messages))
	for i, m := range req.Messages {
		ollamaMsgs[i] = ollamaMsg{Role: m.Role, Content: m.Content}
	}

	ollamaReq := ollamaReq{
		Model:    ollama.Model,
		Messages: ollamaMsgs,
		Stream:   req.Stream,
		Options: map[string]interface{}{
			"repeat_penalty": 1.3,
			"temperature":    0.7,
			"top_p":          0.9,
			"num_predict":    1024,
		},
	}

	body, _ := json.Marshal(ollamaReq)
	ollamaURL := ollama.DefaultBaseURL + "/api/chat"

	httpReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, ollamaURL, bytes.NewReader(body))
	if err != nil {
		http.Error(w, `{"error":{"message":"internal error"}}`, http.StatusInternalServerError)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		http.Error(w, `{"error":{"message":"model unavailable"}}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if req.Stream {
		streamOpenAIResponse(w, resp.Body)
	} else {
		nonStreamOpenAIResponse(w, resp.Body)
	}
}

func streamOpenAIResponse(w http.ResponseWriter, body io.Reader) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, `{"error":{"message":"streaming not supported"}}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var chunk ollamaChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			continue
		}

		finishReason := (*string)(nil)
		if chunk.Done {
			s := "stop"
			finishReason = &s
		}

		openAIChunk := openAIChatResponse{
			ID:      "chatcmpl-coahgpt",
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   "CoahGPT One",
			Choices: []openAIChoice{{
				Index:        0,
				Delta:        openAIMessage{Role: "assistant", Content: chunk.Message.Content},
				FinishReason: finishReason,
			}},
		}

		data, _ := json.Marshal(openAIChunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()

		if chunk.Done {
			fmt.Fprintf(w, "data: [DONE]\n\n")
			flusher.Flush()
			break
		}
	}
}

func nonStreamOpenAIResponse(w http.ResponseWriter, body io.Reader) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)

	var fullContent string
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var chunk ollamaChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			continue
		}
		fullContent += chunk.Message.Content
		if chunk.Done {
			break
		}
	}

	stop := "stop"
	resp := openAIChatResponse{
		ID:      "chatcmpl-coahgpt",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "CoahGPT One",
		Choices: []openAIChoice{{
			Index:        0,
			Message:      openAIMessage{Role: "assistant", Content: fullContent},
			FinishReason: &stop,
		}},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type ollamaMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaReq struct {
	Model    string                 `json:"model"`
	Messages []ollamaMsg            `json:"messages"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type ollamaChunk struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

// Also serve /v1/models for model listing
func HandleOpenAIModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data": []map[string]interface{}{
			{
				"id":       "coahgpt-one",
				"object":   "model",
				"created":  time.Now().Unix(),
				"owned_by": "coah",
			},
		},
	})
}
