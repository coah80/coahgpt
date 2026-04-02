package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/coah80/coahgpt/internal/auth"
)

type ConversationHandler struct {
	db      *auth.DB
	service *auth.Service
}

func NewConversationHandler(db *auth.DB, service *auth.Service) *ConversationHandler {
	return &ConversationHandler{db: db, service: service}
}

func (h *ConversationHandler) getUserFromRequest(r *http.Request) (*auth.User, error) {
	token := extractBearerToken(r)
	if token == "" {
		return nil, errUnauthorized
	}
	return h.service.GetCurrentUser(r.Context(), token)
}

var errUnauthorized = &apiError{status: http.StatusUnauthorized, message: "authorization required"}

type apiError struct {
	status  int
	message string
}

func (e *apiError) Error() string { return e.message }

func (h *ConversationHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/conversations", h.handleConversations)
	mux.HandleFunc("/api/conversations/", h.handleConversationByID)
}

func (h *ConversationHandler) handleConversations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listConversations(w, r)
	case http.MethodPost:
		h.createConversation(w, r)
	default:
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ConversationHandler) handleConversationByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/conversations/")
	if path == "" {
		writeError(w, "conversation id required", http.StatusBadRequest)
		return
	}

	parts := strings.SplitN(path, "/", 2)
	convID := parts[0]

	if len(parts) == 2 && parts[1] == "messages" {
		if r.Method != http.MethodPost {
			writeError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.addMessage(w, r, convID)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getConversation(w, r, convID)
	case http.MethodDelete:
		h.deleteConversation(w, r, convID)
	default:
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ConversationHandler) listConversations(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromRequest(r)
	if err != nil {
		writeError(w, "authorization required", http.StatusUnauthorized)
		return
	}

	summaries, err := h.db.ListConversations(user.ID)
	if err != nil {
		writeError(w, "failed to list conversations", http.StatusInternalServerError)
		return
	}

	if summaries == nil {
		summaries = []*auth.ConversationSummary{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"conversations": summaries,
	})
}

type createConversationRequest struct {
	ID string `json:"id"`
}

func (h *ConversationHandler) createConversation(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromRequest(r)
	if err != nil {
		writeError(w, "authorization required", http.StatusUnauthorized)
		return
	}

	var body createConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(body.ID) == "" {
		writeError(w, "id is required", http.StatusBadRequest)
		return
	}

	conv, err := h.db.CreateConversation(user.ID, body.ID)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			writeError(w, "conversation already exists", http.StatusConflict)
			return
		}
		writeError(w, "failed to create conversation", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         conv.ID,
		"created_at": conv.CreatedAt,
	})
}

func (h *ConversationHandler) getConversation(w http.ResponseWriter, r *http.Request, convID string) {
	user, err := h.getUserFromRequest(r)
	if err != nil {
		writeError(w, "authorization required", http.StatusUnauthorized)
		return
	}

	conv, err := h.db.GetConversation(convID, user.ID)
	if err != nil {
		writeError(w, "conversation not found", http.StatusNotFound)
		return
	}

	msgs := conv.Messages
	if msgs == nil {
		msgs = []auth.ChatMessage{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":       conv.ID,
		"title":    conv.Title,
		"messages": msgs,
	})
}

func (h *ConversationHandler) deleteConversation(w http.ResponseWriter, r *http.Request, convID string) {
	user, err := h.getUserFromRequest(r)
	if err != nil {
		writeError(w, "authorization required", http.StatusUnauthorized)
		return
	}

	if err := h.db.DeleteConversation(convID, user.ID); err != nil {
		writeError(w, "conversation not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

type addMessageRequest struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (h *ConversationHandler) addMessage(w http.ResponseWriter, r *http.Request, convID string) {
	user, err := h.getUserFromRequest(r)
	if err != nil {
		writeError(w, "authorization required", http.StatusUnauthorized)
		return
	}

	var body addMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.Role == "" || body.Content == "" {
		writeError(w, "role and content are required", http.StatusBadRequest)
		return
	}

	// verify the conversation belongs to this user
	_, err = h.db.GetConversation(convID, user.ID)
	if err != nil {
		writeError(w, "conversation not found", http.StatusNotFound)
		return
	}

	if err := h.db.AddChatMessage(convID, body.Role, body.Content); err != nil {
		writeError(w, "failed to add message", http.StatusInternalServerError)
		return
	}

	if err := h.db.UpdateConversationTimestamp(convID); err != nil {
		writeError(w, "failed to update conversation", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}
