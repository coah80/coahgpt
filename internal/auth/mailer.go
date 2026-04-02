package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"time"
)

type ResendMailer struct {
	apiKey   string
	fromAddr string
	client   *http.Client
}

func NewResendMailer(apiKey string) *ResendMailer {
	return &ResendMailer{
		apiKey:   apiKey,
		fromAddr: "coahGPT <noreply@coahgpt.com>",
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (m *ResendMailer) SendVerification(to, name, code string) error {
	body := resendRequest{
		From:    m.fromAddr,
		To:      []string{to},
		Subject: "Verify your coahGPT account",
		HTML:    verificationEmailHTML(name, code),
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal email body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}

	return nil
}

func verificationEmailHTML(name, code string) string {
	return `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0;padding:0;background-color:#1e1e2e;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="background-color:#1e1e2e;padding:40px 0;">
<tr><td align="center">
<table width="480" cellpadding="0" cellspacing="0" style="background-color:#313244;border-radius:16px;overflow:hidden;">

<tr><td style="padding:32px 40px 24px;text-align:center;">
  <img src="https://coahgpt.com/cat-normal.png" alt="coahGPT" width="48" height="48" style="display:inline-block;border-radius:8px;margin-bottom:12px;" />
  <h1 style="margin:0;font-size:28px;font-weight:700;color:#cba6f7;">coahGPT</h1>
</td></tr>

<tr><td style="padding:0 40px;">
  <div style="height:1px;background-color:#45475a;"></div>
</td></tr>

<tr><td style="padding:32px 40px;">
  <p style="margin:0 0 20px;font-size:16px;color:#cdd6f4;">Hey ` + name + `,</p>
  <p style="margin:0 0 24px;font-size:16px;color:#cdd6f4;">Your verification code is:</p>

  <div style="background-color:#1e1e2e;border:2px solid #cba6f7;border-radius:12px;padding:24px;text-align:center;margin:0 0 24px;">
    <span style="font-size:36px;font-weight:700;letter-spacing:8px;color:#cba6f7;font-family:'Courier New',monospace;">` + code + `</span>
  </div>

  <p style="margin:0 0 8px;font-size:14px;color:#a6adc8;">This code expires in 10 minutes.</p>
  <p style="margin:0;font-size:14px;color:#a6adc8;">If you didn't create an account, ignore this email.</p>
</td></tr>

<tr><td style="padding:24px 40px 32px;">
  <div style="height:1px;background-color:#45475a;"></div>
  <p style="margin:16px 0 0;font-size:12px;color:#6c7086;text-align:center;">
    <img src="https://coahgpt.com/cat-normal.png" alt="" width="16" height="16" style="vertical-align:middle;border-radius:3px;margin-right:4px;" />coahGPT &mdash; locally hosted AI with taste
  </p>
</td></tr>

</table>
</td></tr>
</table>
</body>
</html>`
}

// SMTPMailer sends emails via a local or remote SMTP server (e.g. Proton Bridge)
type SMTPMailer struct {
	host     string // e.g. "127.0.0.1"
	port     string // e.g. "1025"
	username string
	password string
	from     string // e.g. "noreply@coahgpt.com"
}

func NewSMTPMailer(host, port, username, password, from string) *SMTPMailer {
	return &SMTPMailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (m *SMTPMailer) SendVerification(to, name, code string) error {
	auth := smtp.PlainAuth("", m.username, m.password, m.host)

	subject := "Verify your coahGPT account"
	htmlBody := verificationEmailHTML(name, code)

	msg := "From: coahGPT <" + m.from + ">\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		htmlBody

	addr := m.host + ":" + m.port
	err := smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}

// NoopMailer logs verification codes to stdout for development
type NoopMailer struct{}

func NewNoopMailer() *NoopMailer {
	return &NoopMailer{}
}

func (m *NoopMailer) SendVerification(to, name, code string) error {
	fmt.Printf("[dev mailer] verification code for %s (%s): %s\n", name, to, code)
	return nil
}
