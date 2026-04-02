package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coah80/coahgpt/internal/harness"
)

type harnessEventMsg struct {
	event harness.Event
}

type harnessCompleteMsg struct{}

func waitForEvent(ch <-chan harness.Event) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-ch
		if !ok {
			return harnessCompleteMsg{}
		}
		return harnessEventMsg{event: event}
	}
}

func runHarness(agent *harness.Agent, history []harness.Message, userMsg string) (context.CancelFunc, <-chan harness.Event) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := agent.RunChat(ctx, history, userMsg)
	return cancel, ch
}

func toolIcon(name string) string {
	switch name {
	case "bash":
		return "⏺"
	case "read_file":
		return "⏺"
	case "write_file":
		return "⏺"
	case "edit_file":
		return "⏺"
	case "list_files":
		return "⏺"
	case "grep":
		return "⏺"
	case "think":
		return "⏺"
	default:
		return "⏺"
	}
}

func toolVerb(name string) string {
	switch name {
	case "bash":
		return "Running bash"
	case "read_file":
		return "Reading"
	case "write_file":
		return "Writing"
	case "edit_file":
		return "Editing"
	case "list_files":
		return "Listing"
	case "grep":
		return "Searching"
	case "think":
		return "Thinking"
	default:
		return name
	}
}
