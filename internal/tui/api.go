package tui

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type TokenMsg struct {
	Token string
}

type DoneMsg struct {
	SessionID string
}

type ErrMsg struct {
	Err error
}

type chatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
}

func StreamChat(serverURL, sessionID, message string) tea.Cmd {
	return func() tea.Msg {
		body := chatRequest{
			Message:   message,
			SessionID: sessionID,
		}

		payload, err := json.Marshal(body)
		if err != nil {
			return ErrMsg{Err: fmt.Errorf("failed to marshal request: %w", err)}
		}

		client := &http.Client{Timeout: 120 * time.Second}
		req, err := http.NewRequest("POST", serverURL+"/api/chat", bytes.NewReader(payload))
		if err != nil {
			return ErrMsg{Err: fmt.Errorf("failed to create request: %w", err)}
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")

		resp, err := client.Do(req)
		if err != nil {
			return ErrMsg{Err: fmt.Errorf("connection failed: %w", err)}
		}

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return ErrMsg{Err: fmt.Errorf("server error (%d): %s", resp.StatusCode, string(respBody))}
		}

		return streamSSE(resp, sessionID)
	}
}

type batchMsg struct {
	tokens    []string
	done      bool
	sessionID string
	err       error
}

func streamSSE(resp *http.Response, sessionID string) tea.Msg {
	defer resp.Body.Close()

	var tokens []string
	newSessionID := sessionID

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		if data == "[DONE]" {
			break
		}

		var event sseEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			// might just be a raw token string
			tokens = append(tokens, data)
			continue
		}

		if event.SessionID != "" {
			newSessionID = event.SessionID
		}

		if event.Token != "" {
			tokens = append(tokens, event.Token)
		}

		if event.Done {
			break
		}

		if event.Error != "" {
			return ErrMsg{Err: fmt.Errorf("server: %s", event.Error)}
		}
	}

	if err := scanner.Err(); err != nil {
		return ErrMsg{Err: fmt.Errorf("stream read error: %w", err)}
	}

	return batchMsg{
		tokens:    tokens,
		done:      true,
		sessionID: newSessionID,
	}
}

type sseEvent struct {
	Token     string `json:"token,omitempty"`
	Done      bool   `json:"done,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Error     string `json:"error,omitempty"`
}
