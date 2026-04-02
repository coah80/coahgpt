package tui

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/coah80/coahgpt/internal/config"
	"github.com/coah80/coahgpt/internal/harness"
	"github.com/coah80/coahgpt/internal/session"

	// cached markdown renderer
	styles "github.com/charmbracelet/glamour/styles"
)

const Version = "1.0.0"

type ChatMessage struct {
	Role    string
	Content string
	Tool    string
	IsError bool
}

type PermissionRequest struct {
	ToolName string
	Args     string
}

type permissionRequestMsg struct {
	req PermissionRequest
}

type bashResultMsg struct {
	cmd    string
	output string
}

type Model struct {
	messages    []ChatMessage
	input       textarea.Model
	spinner     spinner.Model
	viewport    viewport.Model
	streaming   bool
	currentResp strings.Builder
	width       int
	height      int
	err         error
	ready       bool

	serverURL string
	sessionID string

	localMode      bool
	agent          *harness.Agent
	harnessHistory []harness.Message
	eventCh        <-chan harness.Event
	cancelFn       context.CancelFunc
	thinkingTool   string

	// session persistence
	sess *session.Session

	// project config
	projectCfg *config.ProjectConfig

	// permission dialog
	permissionReqCh  chan PermissionRequest
	permissionResCh  chan bool
	pendingPermission *PermissionRequest

	// verbose mode
	verbose bool

	// token tracking
	inputTokens  int
	outputTokens int

	// cached glamour renderer (avoids ANSI leak from repeated OSC 11 queries)
	mdRenderer      *glamour.TermRenderer
	mdRendererWidth int

	// rendered markdown cache: avoids O(n^2) re-rendering during streaming
	renderedCache []string // parallel to messages, cached rendered markdown per assistant msg
}

func NewModel(serverURL string) Model {
	return newModel(false, serverURL, nil, nil)
}

func NewLocalModel(agent *harness.Agent) Model {
	return newModel(true, "", agent, nil)
}

func NewLocalModelWithConfig(agent *harness.Agent, cfg *config.ProjectConfig) Model {
	return newModel(true, "", agent, cfg)
}

func NewLocalModelWithChannels(agent *harness.Agent, cfg *config.ProjectConfig, reqCh chan PermissionRequest, resCh chan bool) Model {
	m := newModel(true, "", agent, cfg)
	m.permissionReqCh = reqCh
	m.permissionResCh = resCh
	return m
}

func NewResumedModel(agent *harness.Agent, sess *session.Session, cfg *config.ProjectConfig) Model {
	m := newModel(true, "", agent, cfg)
	return restoreSession(m, sess)
}

func NewResumedModelWithChannels(agent *harness.Agent, sess *session.Session, cfg *config.ProjectConfig, reqCh chan PermissionRequest, resCh chan bool) Model {
	m := newModel(true, "", agent, cfg)
	m.permissionReqCh = reqCh
	m.permissionResCh = resCh
	return restoreSession(m, sess)
}

func restoreSession(m Model, sess *session.Session) Model {
	m.sess = sess
	for _, msg := range sess.Messages {
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
	return m
}

func newTextArea() textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "send a message..."
	ta.Focus()
	ta.CharLimit = 8192
	ta.MaxHeight = 6
	ta.ShowLineNumbers = false
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(ColorSubtext)
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(ColorText)
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(ColorMauve).Bold(true)
	ta.Prompt = "> "
	ta.SetHeight(1)
	return ta
}

func newModel(local bool, serverURL string, agent *harness.Agent, cfg *config.ProjectConfig) Model {
	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = SpinnerStyle

	reqCh := make(chan PermissionRequest, 1)
	resCh := make(chan bool, 1)

	return Model{
		messages:        []ChatMessage{},
		input:           newTextArea(),
		spinner:         sp,
		serverURL:       serverURL,
		localMode:       local,
		agent:           agent,
		projectCfg:      cfg,
		permissionReqCh: reqCh,
		permissionResCh: resCh,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.spinner.Tick, m.pollPermissions())
}

func (m Model) pollPermissions() tea.Cmd {
	return func() tea.Msg {
		req, ok := <-m.permissionReqCh
		if !ok {
			return nil
		}
		return permissionRequestMsg{req: req}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m.updateLayout(), nil
	case batchMsg:
		return m.handleBatch(msg)
	case harnessEventMsg:
		return m.handleHarnessEvent(msg)
	case harnessCompleteMsg:
		return m.finishStreaming(), nil
	case bashResultMsg:
		m.messages = append(m.messages, ChatMessage{
			Role:    "system",
			Content: msg.output,
		})
		return m.updateViewport(), nil
	case permissionRequestMsg:
		m.pendingPermission = &msg.req
		return m.updateViewport(), nil
	case ErrMsg:
		m.streaming = false
		m.thinkingTool = ""
		m.err = msg.Err
		m.messages = append(m.messages, ChatMessage{
			Role:    "system",
			Content: fmt.Sprintf("error: %s", msg.Err),
		})
		return m.updateViewport(), nil
	case spinner.TickMsg:
		if m.streaming {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	if m.pendingPermission == nil {
		var tiCmd tea.Cmd
		m.input, tiCmd = m.input.Update(msg)
		cmds = append(cmds, tiCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// handle permission dialog input
	if m.pendingPermission != nil {
		return m.handlePermissionKey(msg)
	}

	switch msg.Type {
	case tea.KeyCtrlC:
		if m.cancelFn != nil {
			m.cancelFn()
		}
		if m.streaming {
			m.streaming = false
			m.thinkingTool = ""
			return m.updateViewport(), nil
		}
		return m, tea.Quit
	case tea.KeyEnter:
		// alt+enter -> insert newline (multi-line input)
		if msg.Alt {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
		// plain enter -> submit
		text := strings.TrimSpace(m.input.Value())
		if text == "" {
			return m, nil
		}
		m.input.Reset()
		m.input.SetHeight(1)

		if strings.HasPrefix(text, "/") {
			return m.handleCommand(text)
		}

		// bash mode: ! prefix runs command directly
		if strings.HasPrefix(text, "!") {
			cmd := strings.TrimSpace(strings.TrimPrefix(text, "!"))
			if cmd == "" {
				return m, nil
			}
			m.messages = append(m.messages, ChatMessage{Role: "user", Content: "! " + cmd})
			return m.updateViewport(), m.runDirectBash(cmd)
		}

		if m.streaming {
			return m, nil
		}

		m.messages = append(m.messages, ChatMessage{Role: "user", Content: text})
		m.streaming = true
		m.currentResp.Reset()
		m.err = nil

		// session auto-save: create on first message
		if m.sess == nil && m.localMode {
			m.sess = session.NewSession(ModelDisplayName)
		}
		if m.sess != nil {
			m.sess = m.sess.AddMessage("user", text)
			m.inputTokens += len(text) / 4
			_ = session.Save(m.sess)
		}

		if m.localMode {
			return m.startHarness(text)
		}

		updated := m.updateViewport()
		return updated, tea.Batch(
			StreamChat(m.serverURL, m.sessionID, text),
			m.spinner.Tick,
		)
	case tea.KeyPgUp:
		m.viewport.HalfViewUp()
		return m, nil
	case tea.KeyPgDown:
		m.viewport.HalfViewDown()
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	// auto-grow textarea height
	lines := strings.Count(m.input.Value(), "\n") + 1
	if lines > 6 {
		lines = 6
	}
	if lines < 1 {
		lines = 1
	}
	m.input.SetHeight(lines)

	return m, cmd
}

func (m Model) handlePermissionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.permissionResCh <- true
		perm := m.pendingPermission
		m.pendingPermission = nil
		m.messages = append(m.messages, ChatMessage{
			Role:    "system",
			Content: fmt.Sprintf("allowed: %s", perm.ToolName),
		})
		return m.updateViewport(), m.pollPermissions()
	case "n", "N":
		m.permissionResCh <- false
		perm := m.pendingPermission
		m.pendingPermission = nil
		m.messages = append(m.messages, ChatMessage{
			Role:    "system",
			Content: fmt.Sprintf("denied: %s", perm.ToolName),
		})
		return m.updateViewport(), m.pollPermissions()
	case "a", "A":
		m.permissionResCh <- true
		perm := m.pendingPermission
		m.pendingPermission = nil
		m.messages = append(m.messages, ChatMessage{
			Role:    "system",
			Content: fmt.Sprintf("always allow: %s", perm.ToolName),
		})
		return m.updateViewport(), m.pollPermissions()
	}
	return m, nil
}

func (m Model) runDirectBash(cmd string) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
		output := strings.TrimRight(string(out), "\n")
		if err != nil {
			if output != "" {
				output += "\n"
			}
			output += fmt.Sprintf("exit: %s", err)
		}
		if output == "" {
			output = "(no output)"
		}
		return bashResultMsg{cmd: cmd, output: output}
	}
}

func (m Model) startHarness(text string) (Model, tea.Cmd) {
	cancel, ch := runHarness(m.agent, m.harnessHistory, text)
	m.cancelFn = cancel
	m.eventCh = ch
	m.harnessHistory = append(m.harnessHistory, harness.Message{
		Role:    harness.RoleUser,
		Content: text,
	})

	updated := m.updateViewport()
	return updated, tea.Batch(waitForEvent(ch), m.spinner.Tick)
}

func (m Model) handleHarnessEvent(msg harnessEventMsg) (Model, tea.Cmd) {
	evt := msg.event

	switch evt.Type {
	case harness.EventToken:
		m.currentResp.WriteString(evt.Content)
	case harness.EventToolStart:
		m.flushCurrentResp()
		m.thinkingTool = evt.ToolName
		m.messages = append(m.messages, ChatMessage{
			Role:    "tool",
			Content: fmt.Sprintf("%s %s", toolIcon(evt.ToolName), toolVerb(evt.ToolName)),
			Tool:    evt.ToolName,
		})
	case harness.EventToolProgress:
		if len(m.messages) > 0 && m.messages[len(m.messages)-1].Role == "tool" {
			last := &m.messages[len(m.messages)-1]
			preview := truncateLines(evt.Content, 3)
			last.Content = fmt.Sprintf("%s %s\n%s", toolIcon(last.Tool), toolVerb(last.Tool), preview)
		}
	case harness.EventToolComplete:
		m.thinkingTool = ""
		if len(m.messages) > 0 && m.messages[len(m.messages)-1].Role == "tool" {
			last := &m.messages[len(m.messages)-1]
			preview := truncateLines(evt.Content, 5)
			last.Content = fmt.Sprintf("%s %s\n%s", toolIcon(last.Tool), toolVerb(last.Tool), preview)
		}
	case harness.EventToolError:
		m.thinkingTool = ""
		m.messages = append(m.messages, ChatMessage{
			Role:    "tool",
			Content: fmt.Sprintf("error: %s", evt.Content),
			Tool:    evt.ToolName,
			IsError: true,
		})
	case harness.EventDone:
		m.flushCurrentResp()
		m.saveAssistantHistory()
		if evt.InputTokens > 0 {
			m.inputTokens = evt.InputTokens
		}
		if evt.OutputTokens > 0 {
			m.outputTokens += evt.OutputTokens
		}
		return m.finishStreaming(), nil
	case harness.EventError:
		m.flushCurrentResp()
		m.messages = append(m.messages, ChatMessage{
			Role:    "system",
			Content: fmt.Sprintf("error: %s", evt.Content),
		})
		return m.finishStreaming(), nil
	case harness.EventPermissionAsk:
		m.permissionReqCh <- PermissionRequest{
			ToolName: evt.ToolName,
			Args:     evt.ToolArgs,
		}
	case harness.EventLoopStart:
		// silent
	}

	updated := m.updateViewport()
	return updated, waitForEvent(m.eventCh)
}

func (m *Model) flushCurrentResp() {
	content := strings.TrimSpace(m.currentResp.String())
	if content != "" {
		m.messages = append(m.messages, ChatMessage{Role: "assistant", Content: content})
		m.currentResp.Reset()
	}
}

func (m *Model) saveAssistantHistory() {
	var fullResp strings.Builder
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].Role == "user" {
			break
		}
		if m.messages[i].Role == "assistant" {
			fullResp.WriteString(m.messages[i].Content)
			fullResp.WriteString("\n")
		}
	}
	if fullResp.Len() > 0 {
		m.harnessHistory = append(m.harnessHistory, harness.Message{
			Role:    harness.RoleAssistant,
			Content: strings.TrimSpace(fullResp.String()),
		})
	}
}

func (m Model) finishStreaming() Model {
	m.flushCurrentResp()
	m.streaming = false
	m.thinkingTool = ""
	m.cancelFn = nil
	m.eventCh = nil

	// session auto-save: persist assistant response
	if m.sess != nil {
		for i := len(m.messages) - 1; i >= 0; i-- {
			if m.messages[i].Role == "user" {
				break
			}
			if m.messages[i].Role == "assistant" {
				m.sess = m.sess.AddMessage("assistant", m.messages[i].Content)
				m.outputTokens += len(m.messages[i].Content) / 4
			}
		}
		_ = session.Save(m.sess)
	}

	return m.updateViewport()
}

func (m Model) handleBatch(msg batchMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		m.streaming = false
		m.err = msg.err
		m.messages = append(m.messages, ChatMessage{
			Role:    "system",
			Content: fmt.Sprintf("error: %s", msg.err),
		})
		return m.updateViewport(), nil
	}

	for _, token := range msg.tokens {
		m.currentResp.WriteString(token)
	}
	if msg.sessionID != "" {
		m.sessionID = msg.sessionID
	}
	if msg.done {
		m.streaming = false
		content := strings.TrimSpace(m.currentResp.String())
		if content != "" {
			m.messages = append(m.messages, ChatMessage{Role: "assistant", Content: content})
		}
		m.currentResp.Reset()
	}

	return m.updateViewport(), nil
}

func (m Model) footerHeight() int {
	if m.pendingPermission != nil {
		return 2 // permission prompt takes 2 lines
	}
	if m.streaming {
		return 1 // spinner line
	}
	// textarea height (auto-grows with content) + 0 padding
	lines := strings.Count(m.input.Value(), "\n") + 1
	if lines > 6 {
		lines = 6
	}
	if lines < 1 {
		lines = 1
	}
	return lines
}

func (m Model) updateLayout() Model {
	footer := m.footerHeight()

	vpHeight := m.height - footer
	if vpHeight < 1 {
		vpHeight = 1
	}
	vpWidth := m.width
	if vpWidth < 10 {
		vpWidth = 10
	}

	if !m.ready {
		m.viewport = viewport.New(vpWidth, vpHeight)
		m.viewport.Style = lipgloss.NewStyle()
		m.ready = true
	} else {
		m.viewport.Width = vpWidth
		m.viewport.Height = vpHeight
	}

	m.input.SetWidth(vpWidth - 4)

	// invalidate cached renderer if width changed
	if m.mdRendererWidth != vpWidth-6 {
		m.mdRenderer = nil
		// clear rendered cache since width changed
		m.renderedCache = nil
	}

	return m.updateViewport()
}

func (m Model) updateViewport() Model {
	updated := m
	content := updated.renderMessages()
	updated.viewport.SetContent(content)
	updated.viewport.GotoBottom()
	return updated
}

func (m *Model) renderMessages() string {
	if len(m.messages) == 0 && !m.streaming {
		return m.renderWelcome()
	}

	var sb strings.Builder
	w := m.width - 6
	if w < 20 {
		w = 20
	}

	// grow cache slice to match messages
	for len(m.renderedCache) < len(m.messages) {
		m.renderedCache = append(m.renderedCache, "")
	}

	for i, msg := range m.messages {
		switch msg.Role {
		case "user":
			sb.WriteString(renderUserMessage(msg.Content))

		case "assistant":
			// use cached render for finalized assistant messages
			if m.renderedCache[i] == "" {
				m.renderedCache[i] = m.renderMarkdownCached(msg.Content, w)
			}
			sb.WriteString(renderAssistantBlock(m.renderedCache[i]))

		case "system":
			sb.WriteString(renderSystemMessage(msg.Content))

		case "tool":
			sb.WriteString(renderToolMessage(msg, w))
		}
	}

	// streaming partial: always re-render (only the in-progress tail)
	if m.streaming {
		partial := m.currentResp.String()
		if partial != "" {
			rendered := m.renderMarkdownCached(partial, w)
			sb.WriteString(renderAssistantBlock(rendered))
		}
	}

	return sb.String()
}

func renderUserMessage(content string) string {
	return UserPromptStyle.Render("> "+content) + "\n"
}

func renderAssistantBlock(rendered string) string {
	var sb strings.Builder
	lines := strings.Split(rendered, "\n")
	for i, line := range lines {
		gutter := AssistantResponseGutter.Render("  \u23bf ")
		sb.WriteString(gutter)
		sb.WriteString(line)
		if i < len(lines)-1 {
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")
	return sb.String()
}

func renderSystemMessage(content string) string {
	var sb strings.Builder
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		sb.WriteString("  ")
		sb.WriteString(SystemMsgStyle.Render(line))
		sb.WriteString("\n")
	}
	return sb.String()
}

func renderToolMessage(msg ChatMessage, w int) string {
	var sb strings.Builder
	if msg.IsError {
		sb.WriteString("  ")
		sb.WriteString(ToolErrorStyle.Render(msg.Content))
		sb.WriteString("\n")
		return sb.String()
	}
	lines := strings.SplitN(msg.Content, "\n", 2)
	sb.WriteString("  ")
	sb.WriteString(ToolLabelStyle.Render(lines[0]))
	sb.WriteString("\n")
	if len(lines) > 1 && strings.TrimSpace(lines[1]) != "" {
		preview := ToolContentStyle.Width(w).Render(lines[1])
		for _, pl := range strings.Split(preview, "\n") {
			sb.WriteString("    ")
			sb.WriteString(pl)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (m Model) renderWelcome() string {
	var sb strings.Builder
	sb.WriteString("\n")

	// render cat logo in mauve
	for _, line := range strings.Split(CatASCII, "\n") {
		if line == "" {
			continue
		}
		sb.WriteString(WelcomeNameStyle.Render(line))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(WelcomeNameStyle.Render("  coah code"))
	sb.WriteString(WelcomeDimStyle.Render(" v"+Version))
	sb.WriteString(WelcomeDimStyle.Render(" ("+ModelDisplayName+")"))
	sb.WriteString("\n")
	if m.localMode {
		sb.WriteString(WelcomeDimStyle.Render("  local agent mode"))
		sb.WriteString("\n")
	}
	if m.projectCfg != nil {
		sb.WriteString(WelcomeDimStyle.Render(fmt.Sprintf("  project config: %s", filepath.Base(m.projectCfg.Source))))
		sb.WriteString("\n")
	}
	sb.WriteString(WelcomeDimStyle.Render("  type /help for commands"))
	sb.WriteString("\n")
	return sb.String()
}

// newMarkdownRenderer creates a glamour renderer with a fixed dark style.
// uses glamour.WithStandardStyle instead of WithAutoStyle to avoid sending
// OSC 11 terminal queries that leak ANSI escape bytes into bubbletea's
// input stream.
func newMarkdownRenderer(width int) *glamour.TermRenderer {
	if width < 20 {
		width = 20
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(styles.DarkStyle),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}
	return r
}

// getRenderer returns the cached renderer, rebuilding only when width changes.
func (m *Model) getRenderer(width int) *glamour.TermRenderer {
	if m.mdRenderer == nil || m.mdRendererWidth != width {
		m.mdRenderer = newMarkdownRenderer(width)
		m.mdRendererWidth = width
	}
	return m.mdRenderer
}

func (m *Model) renderMarkdownCached(content string, width int) string {
	r := m.getRenderer(width)
	if r == nil {
		return content
	}
	rendered, err := r.Render(content)
	if err != nil {
		return content
	}
	return strings.TrimRight(rendered, "\n")
}

func (m Model) View() string {
	if !m.ready {
		return "..."
	}

	var sb strings.Builder

	sb.WriteString(m.viewport.View())
	sb.WriteString("\n")

	if m.pendingPermission != nil {
		prompt := fmt.Sprintf("  %s: %s  [y/n/a(lways)]",
			ToolLabelStyle.Render(m.pendingPermission.ToolName),
			truncateForDisplay(m.pendingPermission.Args, 60))
		sb.WriteString(PermissionStyle.Render(prompt))
	} else if m.streaming {
		indicator := m.spinner.View() + " "
		if m.thinkingTool != "" {
			indicator += toolIcon(m.thinkingTool) + " " + toolVerb(m.thinkingTool) + "..."
		} else {
			indicator += "thinking..."
		}
		sb.WriteString(SpinnerStyle.Render(indicator))
	} else {
		sb.WriteString(m.input.View())
	}

	return sb.String()
}

func truncateForDisplay(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func truncateLines(s string, maxLines int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[:maxLines], "\n") + "\n..."
}
