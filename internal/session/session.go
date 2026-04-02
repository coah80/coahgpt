package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
}

type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type SessionSummary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Preview   string    `json:"preview"`
	MsgCount  int       `json:"msg_count"`
}

func SessionDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, ".coah", "sessions")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

func NewSession(model string) *Session {
	return &Session{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Model:     model,
		Messages:  []Message{},
	}
}

func (s *Session) AddMessage(role, content string) *Session {
	return &Session{
		ID:        s.ID,
		Title:     s.Title,
		CreatedAt: s.CreatedAt,
		UpdatedAt: time.Now(),
		Model:     s.Model,
		Messages: append(append([]Message{}, s.Messages...), Message{
			Role:      role,
			Content:   content,
			Timestamp: time.Now(),
		}),
	}
}

func (s *Session) WithTitle(title string) *Session {
	return &Session{
		ID:        s.ID,
		Title:     title,
		CreatedAt: s.CreatedAt,
		UpdatedAt: time.Now(),
		Model:     s.Model,
		Messages:  s.Messages,
	}
}

func Save(s *Session) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling session: %w", err)
	}
	path := filepath.Join(SessionDir(), s.ID+".json")
	return os.WriteFile(path, data, 0o644)
}

func Load(id string) (*Session, error) {
	path := filepath.Join(SessionDir(), id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading session %s: %w", id, err)
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing session %s: %w", id, err)
	}
	return &s, nil
}

func List() ([]*SessionSummary, error) {
	dir := SessionDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading sessions dir: %w", err)
	}

	var summaries []*SessionSummary
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		s, err := Load(id)
		if err != nil {
			continue
		}
		preview := ""
		if len(s.Messages) > 0 {
			preview = s.Messages[0].Content
			if len(preview) > 80 {
				preview = preview[:80] + "..."
			}
		}
		summaries = append(summaries, &SessionSummary{
			ID:        s.ID,
			Title:     s.Title,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
			Preview:   preview,
			MsgCount:  len(s.Messages),
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].UpdatedAt.After(summaries[j].UpdatedAt)
	})

	return summaries, nil
}

func Delete(id string) error {
	path := filepath.Join(SessionDir(), id+".json")
	return os.Remove(path)
}
