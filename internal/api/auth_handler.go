package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/coah80/coahgpt/internal/auth"
)

type AuthHandler struct {
	service *auth.Service
}

func NewAuthHandler(service *auth.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

type signupRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type verifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type resendRequest struct {
	Email string `json:"email"`
}

type userResponse struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Verified bool   `json:"verified"`
}

func (h *AuthHandler) HandleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body signupRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	_, err := h.service.Signup(r.Context(), body.Email, body.Name, body.Password)
	if err != nil {
		if strings.Contains(err.Error(), "email already registered") {
			writeError(w, "email already registered", http.StatusConflict)
			return
		}
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "verification email sent",
	})
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, user, err := h.service.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "not verified") {
			writeError(w, "email not verified", http.StatusForbidden)
			return
		}
		writeError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"token": token,
		"user":  userResponse{Email: user.Email, Name: user.Name, Verified: user.Verified},
	})
}

func (h *AuthHandler) HandleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body verifyRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.VerifyEmail(r.Context(), body.Email, body.Code); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "email verified",
	})
}

func (h *AuthHandler) HandleResend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body resendRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.ResendVerification(r.Context(), body.Email); err != nil {
		writeError(w, "failed to resend verification", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "verification email sent",
	})
}

func (h *AuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := extractBearerToken(r)
	if token == "" {
		writeError(w, "authorization required", http.StatusUnauthorized)
		return
	}

	user, err := h.service.GetCurrentUser(r.Context(), token)
	if err != nil {
		writeError(w, "invalid or expired session", http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": userResponse{Email: user.Email, Name: user.Name, Verified: user.Verified},
	})
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := extractBearerToken(r)
	if token == "" {
		writeError(w, "authorization required", http.StatusUnauthorized)
		return
	}

	if err := h.service.Logout(r.Context(), token); err != nil {
		writeError(w, "logout failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

// RegisterRoutes wires all auth endpoints onto the given mux
func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/auth/signup", h.HandleSignup)
	mux.HandleFunc("/api/auth/login", h.HandleLogin)
	mux.HandleFunc("/api/auth/verify", h.HandleVerify)
	mux.HandleFunc("/api/auth/resend", h.HandleResend)
	mux.HandleFunc("/api/auth/me", h.HandleMe)
	mux.HandleFunc("/api/auth/logout", h.HandleLogout)
}

func extractBearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(header, "Bearer ")
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
