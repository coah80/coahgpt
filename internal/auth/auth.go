package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db     *DB
	mailer Mailer
}

type Mailer interface {
	SendVerification(to, name, code string) error
}

func NewService(db *DB, mailer Mailer) *Service {
	return &Service{db: db, mailer: mailer}
}

func (s *Service) Signup(ctx context.Context, email, name, password string) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	name = strings.TrimSpace(name)

	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.db.CreateUser(ctx, email, name, string(hash))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, fmt.Errorf("email already registered")
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	if err := s.sendVerificationCode(ctx, user); err != nil {
		return nil, fmt.Errorf("send verification: %w", err)
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, *User, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil, fmt.Errorf("invalid credentials")
		}
		return "", nil, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("invalid credentials")
	}

	if !user.Verified {
		return "", nil, fmt.Errorf("email not verified")
	}

	session, err := s.db.CreateSession(ctx, user.ID)
	if err != nil {
		return "", nil, fmt.Errorf("create session: %w", err)
	}

	return session.Token, user, nil
}

func (s *Service) VerifyEmail(ctx context.Context, email, code string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	code = strings.TrimSpace(code)

	if err := s.db.VerifyUser(ctx, email, code); err != nil {
		return err
	}
	return nil
}

func (s *Service) ResendVerification(ctx context.Context, email string) error {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			// don't reveal whether email exists
			return nil
		}
		return fmt.Errorf("find user: %w", err)
	}

	if user.Verified {
		return nil
	}

	return s.sendVerificationCode(ctx, user)
}

func (s *Service) GetCurrentUser(ctx context.Context, token string) (*User, error) {
	session, err := s.db.GetSession(ctx, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid session")
		}
		return nil, fmt.Errorf("get session: %w", err)
	}

	user, err := s.db.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	return s.db.DeleteSession(ctx, token)
}

func (s *Service) sendVerificationCode(ctx context.Context, user *User) error {
	code, err := generateVerifyCode()
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	expiry := time.Now().UTC().Add(10 * time.Minute)
	if err := s.db.SetVerifyCode(ctx, user.ID, code, expiry); err != nil {
		return err
	}

	if err := s.mailer.SendVerification(user.Email, user.Name, code); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}

func generateVerifyCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func validateEmail(email string) error {
	if !strings.Contains(email, "@") {
		return fmt.Errorf("invalid email address")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	return nil
}

