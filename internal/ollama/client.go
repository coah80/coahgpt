package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	DefaultBaseURL = "http://localhost:11434"
	Model          = "qwen3:8b"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string                 `json:"model"`
	Messages []Message              `json:"messages"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type chatStreamChunk struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func StreamChat(ctx context.Context, messages []Message, onToken func(token string, done bool)) error {
	client := NewClient(DefaultBaseURL)
	return client.StreamChat(ctx, messages, onToken)
}

func (c *Client) StreamChat(ctx context.Context, messages []Message, onToken func(token string, done bool)) error {
	reqBody := chatRequest{
		Model:    Model,
		Messages: messages,
		Stream:   true,
		Options: map[string]interface{}{
			"repeat_penalty":    1.3,
			"temperature":      0.7,
			"top_p":            0.9,
			"num_predict":      1024,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshaling chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request to ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var chunk chatStreamChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			return fmt.Errorf("decoding stream chunk: %w", err)
		}

		onToken(chunk.Message.Content, chunk.Done)

		if chunk.Done {
			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading stream: %w", err)
	}

	return nil
}
