package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coah80/coahgpt/internal/harness"
	"github.com/coah80/coahgpt/internal/session"
)

func (m Model) handleCommand(cmd string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(cmd)
	command := strings.ToLower(parts[0])

	switch command {
	case "/quit", "/exit", "/q":
		return m, tea.Quit

	case "/clear", "/c":
		m.messages = nil
		m.renderedCache = nil
		m.sessionID = ""
		m.harnessHistory = nil
		m.err = nil
		m.sess = session.NewSession(ModelDisplayName)
		m.inputTokens = 0
		m.outputTokens = 0
		return m.updateViewport(), nil

	case "/help", "/h":
		return m.cmdHelp()

	case "/model":
		return m.cmdModel()

	case "/save":
		return m.cmdSave(parts)

	case "/sessions", "/history":
		return m.cmdSessions()

	case "/resume":
		return m.cmdResume(parts)

	case "/tokens":
		m.messages = append(m.messages, ChatMessage{
			Role: "system",
			Content: fmt.Sprintf("tokens (estimated): input ~%d, output ~%d, total ~%d",
				m.inputTokens, m.outputTokens, m.inputTokens+m.outputTokens),
		})
		return m.updateViewport(), nil

	case "/export":
		return m.cmdExport()

	case "/compact":
		m.messages = append(m.messages, ChatMessage{
			Role: "system", Content: "compact: not implemented yet",
		})
		return m.updateViewport(), nil

	case "/verbose":
		m.verbose = !m.verbose
		state := "off"
		if m.verbose {
			state = "on"
		}
		m.messages = append(m.messages, ChatMessage{
			Role: "system", Content: fmt.Sprintf("verbose: %s", state),
		})
		return m.updateViewport(), nil

	case "/config":
		return m.cmdConfig()

	default:
		m.messages = append(m.messages, ChatMessage{
			Role:    "system",
			Content: "unknown command: " + command + "  (try /help)",
		})
		return m.updateViewport(), nil
	}
}

func (m Model) cmdHelp() (tea.Model, tea.Cmd) {
	help := strings.Join([]string{
		"/help       show commands",
		"/clear      reset conversation",
		"/model      show current model",
		"/save       save session with optional title",
		"/sessions   list recent sessions",
		"/resume     resume a saved session",
		"/tokens     show token usage estimate",
		"/export     export session as markdown",
		"/compact    summarize old messages",
		"/verbose    toggle verbose tool output",
		"/config     show loaded project config",
		"/quit       exit",
		"",
		"!command    run bash directly (not sent to AI)",
		"alt+enter   insert newline (multi-line input)",
		"ctrl+c      cancel or exit",
		"pgup/pgdn   scroll",
	}, "\n")
	m.messages = append(m.messages, ChatMessage{Role: "system", Content: help})
	return m.updateViewport(), nil
}

func (m Model) cmdModel() (tea.Model, tea.Cmd) {
	mode := "remote"
	if m.localMode {
		mode = "local"
	}
	m.messages = append(m.messages, ChatMessage{
		Role:    "system",
		Content: fmt.Sprintf("%s (%s)", ModelDisplayName, mode),
	})
	return m.updateViewport(), nil
}

func (m Model) cmdSave(parts []string) (tea.Model, tea.Cmd) {
	if m.sess == nil {
		m.messages = append(m.messages, ChatMessage{
			Role: "system", Content: "no active session to save",
		})
		return m.updateViewport(), nil
	}
	if len(parts) > 1 {
		title := strings.Join(parts[1:], " ")
		m.sess = m.sess.WithTitle(title)
	} else if m.sess.Title == "" {
		for _, msg := range m.sess.Messages {
			if msg.Role == "user" {
				title := msg.Content
				if len(title) > 50 {
					title = title[:50] + "..."
				}
				m.sess = m.sess.WithTitle(title)
				break
			}
		}
	}
	if err := session.Save(m.sess); err != nil {
		m.messages = append(m.messages, ChatMessage{
			Role: "system", Content: fmt.Sprintf("save failed: %s", err),
		})
	} else {
		m.messages = append(m.messages, ChatMessage{
			Role: "system", Content: fmt.Sprintf("saved: %s", m.sess.ID[:8]),
		})
	}
	return m.updateViewport(), nil
}

func (m Model) cmdSessions() (tea.Model, tea.Cmd) {
	summaries, err := session.List()
	if err != nil {
		m.messages = append(m.messages, ChatMessage{
			Role: "system", Content: fmt.Sprintf("error: %s", err),
		})
		return m.updateViewport(), nil
	}
	if len(summaries) == 0 {
		m.messages = append(m.messages, ChatMessage{
			Role: "system", Content: "no saved sessions",
		})
		return m.updateViewport(), nil
	}
	limit := 10
	if len(summaries) < limit {
		limit = len(summaries)
	}
	var sb strings.Builder
	sb.WriteString("recent sessions:\n")
	for i := 0; i < limit; i++ {
		s := summaries[i]
		title := s.Title
		if title == "" {
			title = s.Preview
		}
		if title == "" {
			title = "(untitled)"
		}
		sb.WriteString(fmt.Sprintf("  %s  %s  (%d msgs)\n",
			s.ID[:8], title, s.MsgCount))
	}
	sb.WriteString("\nuse /resume <id> to load")
	m.messages = append(m.messages, ChatMessage{
		Role: "system", Content: sb.String(),
	})
	return m.updateViewport(), nil
}

func (m Model) cmdResume(parts []string) (tea.Model, tea.Cmd) {
	if len(parts) < 2 {
		m.messages = append(m.messages, ChatMessage{
			Role: "system", Content: "usage: /resume <session-id-prefix>",
		})
		return m.updateViewport(), nil
	}
	prefix := parts[1]
	summaries, err := session.List()
	if err != nil {
		m.messages = append(m.messages, ChatMessage{
			Role: "system", Content: fmt.Sprintf("error: %s", err),
		})
		return m.updateViewport(), nil
	}
	var match *session.SessionSummary
	for _, s := range summaries {
		if strings.HasPrefix(s.ID, prefix) {
			match = s
			break
		}
	}
	if match == nil {
		m.messages = append(m.messages, ChatMessage{
			Role: "system", Content: "no session matching: " + prefix,
		})
		return m.updateViewport(), nil
	}
	loaded, err := session.Load(match.ID)
	if err != nil {
		m.messages = append(m.messages, ChatMessage{
			Role: "system", Content: fmt.Sprintf("load failed: %s", err),
		})
		return m.updateViewport(), nil
	}
	m.sess = loaded
	m.messages = nil
	m.renderedCache = nil
	m.harnessHistory = nil
	for _, msg := range loaded.Messages {
		m.messages = append(m.messages, ChatMessage{Role: msg.Role, Content: msg.Content})
		if msg.Role == "user" {
			m.harnessHistory = append(m.harnessHistory, harness.Message{
				Role: harness.RoleUser, Content: msg.Content,
			})
		} else if msg.Role == "assistant" {
			m.harnessHistory = append(m.harnessHistory, harness.Message{
				Role: harness.RoleAssistant, Content: msg.Content,
			})
		}
	}
	m.messages = append(m.messages, ChatMessage{
		Role: "system", Content: fmt.Sprintf("resumed session %s (%d messages)", match.ID[:8], match.MsgCount),
	})
	return m.updateViewport(), nil
}

func (m Model) cmdExport() (tea.Model, tea.Cmd) {
	if len(m.messages) == 0 {
		m.messages = append(m.messages, ChatMessage{
			Role: "system", Content: "nothing to export",
		})
		return m.updateViewport(), nil
	}

	var sb strings.Builder
	sb.WriteString("# Session Export\n\n")
	sb.WriteString(fmt.Sprintf("*Exported: %s*\n\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString("---\n\n")

	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			sb.WriteString("**You:**\n\n")
			sb.WriteString(msg.Content)
			sb.WriteString("\n\n")
		case "assistant":
			sb.WriteString("**Assistant:**\n\n")
			sb.WriteString(msg.Content)
			sb.WriteString("\n\n")
		case "tool":
			sb.WriteString(fmt.Sprintf("*[tool: %s]*\n\n", msg.Tool))
		case "system":
			sb.WriteString(fmt.Sprintf("*%s*\n\n", msg.Content))
		}
	}

	outPath := filepath.Join(".", "session.md")
	if err := os.WriteFile(outPath, []byte(sb.String()), 0o644); err != nil {
		m.messages = append(m.messages, ChatMessage{
			Role: "system", Content: fmt.Sprintf("export failed: %s", err),
		})
		return m.updateViewport(), nil
	}

	m.messages = append(m.messages, ChatMessage{
		Role: "system", Content: "exported to ./session.md",
	})
	return m.updateViewport(), nil
}

func (m Model) cmdConfig() (tea.Model, tea.Cmd) {
	if m.projectCfg == nil {
		m.messages = append(m.messages, ChatMessage{
			Role: "system", Content: "no project config loaded",
		})
		return m.updateViewport(), nil
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("source: %s\n", m.projectCfg.Source))
	if m.projectCfg.Model != "" {
		sb.WriteString(fmt.Sprintf("model: %s\n", m.projectCfg.Model))
	}
	if m.projectCfg.System != "" {
		preview := m.projectCfg.System
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		sb.WriteString(fmt.Sprintf("system: %s\n", preview))
	}
	if len(m.projectCfg.Permissions) > 0 {
		sb.WriteString("permissions:\n")
		for k, v := range m.projectCfg.Permissions {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}
	m.messages = append(m.messages, ChatMessage{
		Role: "system", Content: strings.TrimRight(sb.String(), "\n"),
	})
	return m.updateViewport(), nil
}
