package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coah80/coahgpt/internal/chat"
	"github.com/coah80/coahgpt/internal/ollama"
	"github.com/coah80/coahgpt/internal/search"
)

type Handler struct {
	store  *chat.Store
	client *ollama.Client
}

func NewHandler(store *chat.Store, client *ollama.Client) *Handler {
	return &Handler{
		store:  store,
		client: client,
	}
}

type chatRequestBody struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

type sseEvent struct {
	Token     string `json:"token"`
	Thinking  string `json:"thinking,omitempty"`
	Done      bool   `json:"done"`
	SessionID string `json:"session_id,omitempty"`
	Sources   string `json:"sources,omitempty"`
}

type sessionSummary struct {
	ID      string    `json:"id"`
	Preview string    `json:"preview"`
	Created time.Time `json:"created"`
}

func (h *Handler) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var body chatRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(body.Message) == "" {
		http.Error(w, `{"error":"message is required"}`, http.StatusBadRequest)
		return
	}

	var session *chat.Session
	if body.SessionID != "" {
		session = h.store.GetSession(body.SessionID)
		if session == nil {
			session = h.store.NewSession()
		}
	} else {
		session = h.store.NewSession()
	}

	filteredMessage := filterInput(body.Message)

	isWebSearch := strings.Contains(filteredMessage, "[Web Search]")
	isDeepResearch := strings.Contains(filteredMessage, "[Deep Research]")
	cleanMessage := strings.ReplaceAll(strings.ReplaceAll(filteredMessage, " [Web Search]", ""), " [Deep Research]", "")

	searchContext := ""
	if isWebSearch || isDeepResearch {
		numResults := 5
		if isDeepResearch {
			numResults = 10
		}
		results, err := search.WebSearch(cleanMessage, numResults)
		if err == nil && len(results) > 0 {
			searchContext = search.FormatResults(results)
		}
	}

	userContent := cleanMessage
	if searchContext != "" {
		userContent = cleanMessage + "\n\n---\n" + searchContext + "\n---\nUse the search results above to answer. Cite sources when relevant."
	}

	session = h.store.AddMessage(session.ID, ollama.Message{
		Role:    "user",
		Content: userContent,
	})

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, `{"error":"streaming not supported"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	if searchContext != "" {
		srcEvt := sseEvent{Thinking: "searching the web..."}
		data, _ := json.Marshal(srcEvt)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	var fullResponse strings.Builder
	leakDetected := false

	err := h.client.StreamChat(r.Context(), session.Messages, func(token string, done bool) {
		if !done {
			fullResponse.WriteString(token)

			filtered, leaked := filterLeakedContent(fullResponse.String(), token)
			if leaked && !leakDetected {
				leakDetected = true
				evt := sseEvent{Token: "\n\nnah bro nice try lol", Done: false}
				data, _ := json.Marshal(evt)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
				return
			}
			if leakDetected {
				return
			}
			token = filtered
		}

		evt := sseEvent{
			Token: token,
			Done:  done,
		}
		if done {
			evt.SessionID = session.ID
		}

		data, _ := json.Marshal(evt)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	})

	if err != nil {
		errEvt := sseEvent{Token: fmt.Sprintf("\n\n[error: %s]", err.Error()), Done: true, SessionID: session.ID}
		data, _ := json.Marshal(errEvt)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
		return
	}

	h.store.AddMessage(session.ID, ollama.Message{
		Role:    "assistant",
		Content: fullResponse.String(),
	})
}

func (h *Handler) HandleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	sessions := h.store.ListSessions()
	summaries := make([]sessionSummary, 0, len(sessions))

	for _, s := range sessions {
		preview := ""
		for _, msg := range s.Messages {
			if msg.Role == "user" {
				preview = msg.Content
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}
				break
			}
		}
		summaries = append(summaries, sessionSummary{
			ID:      s.ID,
			Preview: preview,
			Created: s.Created,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"model":  "CoahGPT One",
	})
}
