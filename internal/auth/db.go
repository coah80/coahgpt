package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	db *sql.DB
}

type User struct {
	ID           int64
	Email        string
	Name         string
	PasswordHash string
	Verified     bool
	VerifyCode   string
	VerifyExpiry time.Time
	CreatedAt    time.Time
}

type Session struct {
	Token     string
	UserID    int64
	CreatedAt time.Time
	ExpiresAt time.Time
}

type Conversation struct {
	ID        string
	UserID    int64
	Title     string
	Messages  []ChatMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ChatMessage struct {
	Role    string
	Content string
}

type ConversationSummary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Preview   string    `json:"preview"`
	UpdatedAt time.Time `json:"updated_at"`
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    verified INTEGER DEFAULT 0,
    verify_code TEXT,
    verify_expiry DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS conversations (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    title TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS chat_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

func NewDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// sqlite tuning for concurrent access
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set journal mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &DB{db: db}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) CreateUser(ctx context.Context, email, name, passwordHash string) (*User, error) {
	result, err := d.db.ExecContext(ctx,
		"INSERT INTO users (email, name, password_hash) VALUES (?, ?, ?)",
		email, name, passwordHash,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	return &User{
		ID:           id,
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		Verified:     false,
		CreatedAt:    time.Now(),
	}, nil
}

func (d *DB) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return d.scanUser(d.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, verified,
		        COALESCE(verify_code, ''), COALESCE(verify_expiry, ''),
		        created_at
		 FROM users WHERE email = ?`, email,
	))
}

func (d *DB) GetUserByID(ctx context.Context, id int64) (*User, error) {
	return d.scanUser(d.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, verified,
		        COALESCE(verify_code, ''), COALESCE(verify_expiry, ''),
		        created_at
		 FROM users WHERE id = ?`, id,
	))
}

func (d *DB) scanUser(row *sql.Row) (*User, error) {
	var u User
	var verified int
	var verifyExpiry string

	err := row.Scan(
		&u.ID, &u.Email, &u.Name, &u.PasswordHash, &verified,
		&u.VerifyCode, &verifyExpiry, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	u.Verified = verified == 1
	if verifyExpiry != "" {
		u.VerifyExpiry, _ = time.Parse("2006-01-02 15:04:05", verifyExpiry)
	}

	return &u, nil
}

func (d *DB) SetVerifyCode(ctx context.Context, userID int64, code string, expiry time.Time) error {
	_, err := d.db.ExecContext(ctx,
		"UPDATE users SET verify_code = ?, verify_expiry = ? WHERE id = ?",
		code, expiry.UTC().Format("2006-01-02 15:04:05"), userID,
	)
	if err != nil {
		return fmt.Errorf("set verify code: %w", err)
	}
	return nil
}

func (d *DB) VerifyUser(ctx context.Context, email, code string) error {
	var storedCode string
	var expiryStr string
	var verified int

	err := d.db.QueryRowContext(ctx,
		`SELECT COALESCE(verify_code, ''), COALESCE(verify_expiry, ''), verified
		 FROM users WHERE email = ?`, email,
	).Scan(&storedCode, &expiryStr, &verified)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}

	if verified == 1 {
		return fmt.Errorf("already verified")
	}
	if storedCode == "" || storedCode != code {
		return fmt.Errorf("invalid verification code")
	}
	if expiryStr != "" {
		expiry, _ := time.Parse("2006-01-02 15:04:05", expiryStr)
		if time.Now().UTC().After(expiry) {
			return fmt.Errorf("verification code expired")
		}
	}

	_, err = d.db.ExecContext(ctx,
		"UPDATE users SET verified = 1, verify_code = NULL, verify_expiry = NULL WHERE email = ?",
		email,
	)
	if err != nil {
		return fmt.Errorf("verify user: %w", err)
	}
	return nil
}

func (d *DB) CreateSession(ctx context.Context, userID int64) (*Session, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	now := time.Now().UTC()
	expires := now.Add(30 * 24 * time.Hour)

	_, err := d.db.ExecContext(ctx,
		"INSERT INTO sessions (token, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)",
		token, userID,
		now.Format("2006-01-02 15:04:05"),
		expires.Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return nil, fmt.Errorf("insert session: %w", err)
	}

	return &Session{
		Token:     token,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: expires,
	}, nil
}

func (d *DB) GetSession(ctx context.Context, token string) (*Session, error) {
	var s Session
	err := d.db.QueryRowContext(ctx,
		`SELECT token, user_id, created_at, expires_at
		 FROM sessions WHERE token = ? AND expires_at > ?`,
		token, time.Now().UTC().Format("2006-01-02 15:04:05"),
	).Scan(&s.Token, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (d *DB) DeleteSession(ctx context.Context, token string) error {
	_, err := d.db.ExecContext(ctx, "DELETE FROM sessions WHERE token = ?", token)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (d *DB) DeleteExpiredSessions(ctx context.Context) error {
	_, err := d.db.ExecContext(ctx,
		"DELETE FROM sessions WHERE expires_at <= ?",
		time.Now().UTC().Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return fmt.Errorf("delete expired sessions: %w", err)
	}
	return nil
}

func (d *DB) CreateConversation(userID int64, id string) (*Conversation, error) {
	now := time.Now().UTC()
	nowStr := now.Format("2006-01-02 15:04:05")

	_, err := d.db.Exec(
		"INSERT INTO conversations (id, user_id, title, created_at, updated_at) VALUES (?, ?, '', ?, ?)",
		id, userID, nowStr, nowStr,
	)
	if err != nil {
		return nil, fmt.Errorf("insert conversation: %w", err)
	}

	return &Conversation{
		ID:        id,
		UserID:    userID,
		Title:     "",
		Messages:  nil,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (d *DB) GetConversation(id string, userID int64) (*Conversation, error) {
	var conv Conversation
	err := d.db.QueryRow(
		"SELECT id, user_id, title, created_at, updated_at FROM conversations WHERE id = ? AND user_id = ?",
		id, userID,
	).Scan(&conv.ID, &conv.UserID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get conversation: %w", err)
	}

	rows, err := d.db.Query(
		"SELECT role, content FROM chat_messages WHERE conversation_id = ? ORDER BY id ASC",
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		if err := rows.Scan(&msg.Role, &msg.Content); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate messages: %w", err)
	}

	conv.Messages = messages
	return &conv, nil
}

func (d *DB) ListConversations(userID int64) ([]*ConversationSummary, error) {
	rows, err := d.db.Query(`
		SELECT c.id, c.title, COALESCE(
			(SELECT content FROM chat_messages
			 WHERE conversation_id = c.id AND role = 'user'
			 ORDER BY id ASC LIMIT 1), ''
		) AS preview, c.updated_at
		FROM conversations c
		WHERE c.user_id = ?
		ORDER BY c.updated_at DESC
		LIMIT 50`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	var summaries []*ConversationSummary
	for rows.Next() {
		var s ConversationSummary
		if err := rows.Scan(&s.ID, &s.Title, &s.Preview, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan conversation summary: %w", err)
		}
		if len(s.Preview) > 100 {
			s.Preview = s.Preview[:100] + "..."
		}
		summaries = append(summaries, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate conversations: %w", err)
	}

	return summaries, nil
}

func (d *DB) DeleteConversation(id string, userID int64) error {
	result, err := d.db.Exec(
		"DELETE FROM conversations WHERE id = ? AND user_id = ?",
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("delete conversation: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("conversation not found")
	}
	return nil
}

func (d *DB) AddChatMessage(conversationID, role, content string) error {
	_, err := d.db.Exec(
		"INSERT INTO chat_messages (conversation_id, role, content) VALUES (?, ?, ?)",
		conversationID, role, content,
	)
	if err != nil {
		return fmt.Errorf("insert chat message: %w", err)
	}
	return nil
}

func (d *DB) UpdateConversationTitle(id, title string) error {
	_, err := d.db.Exec(
		"UPDATE conversations SET title = ? WHERE id = ?",
		title, id,
	)
	if err != nil {
		return fmt.Errorf("update conversation title: %w", err)
	}
	return nil
}

func (d *DB) UpdateConversationTimestamp(id string) error {
	_, err := d.db.Exec(
		"UPDATE conversations SET updated_at = ? WHERE id = ?",
		time.Now().UTC().Format("2006-01-02 15:04:05"), id,
	)
	if err != nil {
		return fmt.Errorf("update conversation timestamp: %w", err)
	}
	return nil
}
