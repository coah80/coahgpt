package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coah80/coahgpt/internal/config"
	"github.com/coah80/coahgpt/internal/harness"
	"github.com/coah80/coahgpt/internal/persona"
	"github.com/coah80/coahgpt/internal/session"
	"github.com/coah80/coahgpt/internal/tui"
)

func main() {
	serverURL := flag.String("server", "https://coahgpt.com", "coah code API server URL")
	flag.StringVar(serverURL, "s", "https://coahgpt.com", "coah code API server URL (shorthand)")

	ollamaURL := flag.String("ollama", "", "Ollama API URL for local agent mode (e.g. http://localhost:11434)")
	flag.StringVar(ollamaURL, "o", "", "Ollama API URL (shorthand)")

	local := flag.Bool("local", false, "run in local agent mode (tools execute on your machine)")
	flag.BoolVar(local, "l", false, "local agent mode (shorthand)")

	model := flag.String("model", "qwen2.5:7b", "model name for local mode")
	flag.StringVar(model, "m", "qwen2.5:7b", "model name (shorthand)")

	resume := flag.String("resume", "", "resume a saved session by ID prefix")
	flag.StringVar(resume, "r", "", "resume session (shorthand)")

	version := flag.Bool("version", false, "print version and exit")
	flag.BoolVar(version, "v", false, "print version and exit (shorthand)")

	flag.Usage = printUsage
	flag.Parse()

	if *version {
		fmt.Printf("coah code v%s (%s)\n", tui.Version, tui.ModelDisplayName)
		os.Exit(0)
	}

	var m tui.Model

	if *local || *ollamaURL != "" || *resume != "" {
		url := *ollamaURL
		if url == "" {
			url = "http://localhost:11434"
		}

		systemPrompt := persona.SystemPrompt

		// load project config
		cwd, _ := os.Getwd()
		projCfg, err := config.LoadProjectConfig(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: project config error: %v\n", err)
		}

		if projCfg != nil {
			if projCfg.Model != "" {
				*model = projCfg.Model
			}
			if projCfg.System != "" {
				systemPrompt = systemPrompt + "\n\n# Project Context\n" + projCfg.System
			}
			fmt.Fprintf(os.Stderr, "loaded project config from %s\n", projCfg.Source)
		}

		// create permission channels before agent so the callback can capture them
		permReqCh := make(chan tui.PermissionRequest, 1)
		permResCh := make(chan bool, 1)

		askFn := func(toolName string, args string) bool {
			permReqCh <- tui.PermissionRequest{ToolName: toolName, Args: args}
			return <-permResCh
		}

		agent := harness.DefaultAgentWithPermissionCallback(harness.AgentConfig{
			OllamaURL: url,
			Model:     *model,
			System:    systemPrompt,
		}, askFn)

		// handle --resume
		if *resume != "" {
			sess := findSession(*resume)
			if sess == nil {
				fmt.Fprintf(os.Stderr, "no session matching: %s\n", *resume)
				os.Exit(1)
			}
			m = tui.NewResumedModelWithChannels(agent, sess, projCfg, permReqCh, permResCh)
		} else {
			m = tui.NewLocalModelWithChannels(agent, projCfg, permReqCh, permResCh)
		}
	} else {
		m = tui.NewModel(*serverURL)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func findSession(prefix string) *session.Session {
	summaries, err := session.List()
	if err != nil || len(summaries) == 0 {
		return nil
	}
	for _, s := range summaries {
		if strings.HasPrefix(s.ID, prefix) {
			sess, err := session.Load(s.ID)
			if err != nil {
				return nil
			}
			return sess
		}
	}
	return nil
}

func printUsage() {
	fmt.Printf(`
  coah code v%s (%s)

Usage: coah [flags]

Flags:
  -s, --server <url>   API server URL (default: https://coahgpt.com)
  -l, --local          run in local agent mode (tools execute locally)
  -o, --ollama <url>   Ollama API URL (default: http://localhost:11434)
  -m, --model <name>   model name (default: qwen2.5:7b)
  -r, --resume <id>    resume a saved session by ID prefix
  -v, --version        print version and exit
  -h, --help           show this help

Modes:
  remote (default)     connects to coah code server for chat
  local (-l)           runs agent loop locally with tool execution
                       reads, edits, writes files, runs bash, greps code

Project Config:
  Place COAH.md or .coah/config.json in your project root.
  COAH.md content is appended to the system prompt.
  config.json can override model, permissions, and exclude paths.

Examples:
  coahgpt                                  chat via server
  coahgpt -l                               local agent mode
  coahgpt -l -o http://myserver:11434      custom Ollama URL
  coahgpt -l -m llama3.1:8b               different model
  coahgpt -r abc123                        resume saved session
`, tui.Version, tui.ModelDisplayName)
}
