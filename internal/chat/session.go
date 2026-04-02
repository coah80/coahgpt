package chat

import (
	"crypto/rand"
	"encoding/hex"
	"sort"
	"sync"
	"time"

	"github.com/coah80/coahgpt/internal/ollama"
	"github.com/coah80/coahgpt/internal/persona"
)

type Session struct {
	ID       string           `json:"id"`
	Messages []ollama.Message `json:"messages"`
	Created  time.Time        `json:"created"`
}

type Store struct {
	sessions sync.Map
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) NewSession() *Session {
	session := &Session{
		ID:      generateID(),
		Created: time.Now(),
		Messages: []ollama.Message{
			{Role: "system", Content: persona.ChatPrompt},
		},
	}
	s.sessions.Store(session.ID, session)
	return session
}

func (s *Store) GetSession(id string) *Session {
	val, ok := s.sessions.Load(id)
	if !ok {
		return nil
	}
	return val.(*Session)
}

func (s *Store) AddMessage(sessionID string, msg ollama.Message) *Session {
	val, ok := s.sessions.Load(sessionID)
	if !ok {
		return nil
	}
	session := val.(*Session)
	updated := &Session{
		ID:       session.ID,
		Created:  session.Created,
		Messages: append(copyMessages(session.Messages), msg),
	}
	s.sessions.Store(sessionID, updated)
	return updated
}

func (s *Store) ListSessions() []*Session {
	var sessions []*Session
	s.sessions.Range(func(_, value any) bool {
		sessions = append(sessions, value.(*Session))
		return true
	})
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Created.After(sessions[j].Created)
	})
	return sessions
}

func copyMessages(msgs []ollama.Message) []ollama.Message {
	cp := make([]ollama.Message, len(msgs))
	copy(cp, msgs)
	return cp
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
