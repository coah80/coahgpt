package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coah80/coahgpt/internal/auth"
	"github.com/coah80/coahgpt/internal/chat"
	"github.com/coah80/coahgpt/internal/ollama"
)

type Handler struct {
	store       *chat.Store
	client      *ollama.Client
	authDB      *auth.DB
	authService *auth.Service
}

func NewHandler(store *chat.Store, client *ollama.Client, authDB *auth.DB, authService *auth.Service) *Handler {
	return &Handler{
		store:       store,
		client:      client,
		authDB:      authDB,
		authService: authService,
	}
}

type chatRequestBody struct {
	SessionID      string `json:"session_id"`
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id"`
}

type sseEvent struct {
	Token     string `json:"token"`
	Done      bool   `json:"done"`
	SessionID string `json:"session_id,omitempty"`
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

	// check if the user is authenticated and has a conversation_id for DB persistence
	var dbUserID int64
	var dbConvID string
	if token := extractBearerToken(r); token != "" && body.ConversationID != "" && h.authService != nil {
		user, err := h.authService.GetCurrentUser(r.Context(), token)
		if err == nil {
			dbUserID = user.ID
			dbConvID = body.ConversationID
		}
	}

	var session *chat.Session
	if body.SessionID != "" {
		session = h.store.GetSession(body.SessionID)
		if session == nil {
			// session doesn't exist in memory — create a new one
			session = h.store.NewSession()
		}
	} else {
		session = h.store.NewSession()
	}

	session = h.store.AddMessage(session.ID, ollama.Message{
		Role:    "user",
		Content: body.Message,
	})

	// persist user message to DB if authenticated
	if dbConvID != "" {
		_ = h.authDB.AddChatMessage(dbConvID, "user", body.Message)
		_ = h.authDB.UpdateConversationTimestamp(dbConvID)

		// auto-generate title from first user message
		conv, err := h.authDB.GetConversation(dbConvID, dbUserID)
		if err == nil && conv.Title == "" {
			title := body.Message
			if len(title) > 50 {
				title = title[:50]
			}
			_ = h.authDB.UpdateConversationTitle(dbConvID, title)
		}
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, `{"error":"streaming not supported"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	var fullResponse strings.Builder

	err := h.client.StreamChat(r.Context(), session.Messages, func(token string, done bool) {
		if !done {
			fullResponse.WriteString(token)
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

	// persist assistant response to DB if authenticated
	if dbConvID != "" {
		_ = h.authDB.AddChatMessage(dbConvID, "assistant", fullResponse.String())
		_ = h.authDB.UpdateConversationTimestamp(dbConvID)
	}
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
		"model":  ollama.Model,
	})
}
