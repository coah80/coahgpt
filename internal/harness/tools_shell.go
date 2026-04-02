package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`rm\s+(-rf|-fr)\s+/\s*$`),
	regexp.MustCompile(`rm\s+(-rf|-fr)\s+~/?\s*$`),
	regexp.MustCompile(`rm\s+(-rf|-fr)\s+\.\s*$`),
	regexp.MustCompile(`mkfs\.`),
	regexp.MustCompile(`dd\s+.*of=/dev/`),
	regexp.MustCompile(`>\s*/dev/sd`),
	regexp.MustCompile(`:(){.*};:`),
	regexp.MustCompile(`chmod\s+777\s+/`),
	regexp.MustCompile(`curl.*\|\s*(sudo\s+)?bash`),
	regexp.MustCompile(`wget.*\|\s*(sudo\s+)?bash`),
}

func RegisterShellTool(r *Registry) {
	r.Register(ToolDef{
		Name:        "bash",
		Description: "run a shell command. returns stdout+stderr. use for builds, tests, git, installs. do NOT use for reading files.",
		Destructive: true,
		Parameters: schemaObject(
			schemaProp("command", "string", "the shell command to run"),
			schemaProp("timeout", "integer", "timeout in milliseconds (default 30000)"),
			schemaRequired("command"),
		),
	}, bashHandler(r))
}

func bashHandler(registry *Registry) ToolHandler {
	return func(ctx context.Context, raw json.RawMessage) (ToolResult, error) {
		var args struct {
			Command string `json:"command"`
			Timeout int    `json:"timeout"`
		}
		if err := json.Unmarshal(raw, &args); err != nil {
			return ToolResult{}, fmt.Errorf("invalid arguments: %w", err)
		}

		if args.Command == "" {
			return ToolResult{Content: "command is required", IsError: true}, nil
		}

		for _, pat := range dangerousPatterns {
			if pat.MatchString(args.Command) {
				return ToolResult{
					Content: fmt.Sprintf("blocked: matches dangerous pattern (%s)", pat.String()),
					IsError: true,
				}, nil
			}
		}

		timeout := 30 * time.Second
		if args.Timeout > 0 {
			timeout = time.Duration(args.Timeout) * time.Millisecond
		}

		cmdCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		cmd := exec.CommandContext(cmdCtx, "sh", "-c", args.Command)
		cmd.Env = sanitizedEnv()

		out, err := cmd.CombinedOutput()
		result := string(out)

		result = truncateOutput(result, 10*1024)

		if err != nil {
			if cmdCtx.Err() == context.DeadlineExceeded {
				return ToolResult{
					Content: fmt.Sprintf("command timed out after %s\n%s", timeout, result),
					IsError: true,
				}, nil
			}
			return ToolResult{
				Content: result + "\nexit: " + err.Error(),
				IsError: true,
			}, nil
		}

		return ToolResult{Content: result}, nil
	}
}

func truncateOutput(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	half := maxBytes / 2
	head := s[:half]
	tail := s[len(s)-half:]
	trimmed := len(s) - maxBytes
	return head + fmt.Sprintf("\n\n[truncated %d bytes]\n\n", trimmed) + tail
}

func sanitizedEnv() []string {
	allowed := map[string]bool{
		"PATH": true, "HOME": true, "USER": true, "SHELL": true,
		"TMPDIR": true, "TEMP": true, "TMP": true,
		"LANG": true, "LC_ALL": true, "LC_CTYPE": true,
		"GOPATH": true, "GOROOT": true, "GOBIN": true,
		"NODE_PATH": true, "NVM_DIR": true,
		"RUST_BACKTRACE": true, "CARGO_HOME": true,
		"TERM": true, "COLORTERM": true,
		"EDITOR": true, "VISUAL": true,
		"XDG_DATA_HOME": true, "XDG_CONFIG_HOME": true, "XDG_CACHE_HOME": true,
	}

	var env []string
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 && allowed[parts[0]] {
			env = append(env, e)
		}
	}
	return env
}
