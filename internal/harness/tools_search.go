package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

func RegisterSearchTool(r *Registry) {
	r.Register(ToolDef{
		Name:        "grep",
		Description: "search for a pattern in files using regex. returns matching lines with file:line format.",
		ReadOnly:    true,
		Parameters: schemaObject(
			schemaProp("pattern", "string", "regex pattern to search for"),
			schemaProp("path", "string", "directory or file to search in (default: current directory)"),
			schemaProp("include", "string", "glob filter for filenames (e.g. *.go, *.ts)"),
			schemaRequired("pattern"),
		),
	}, grepHandler())
}

func grepHandler() ToolHandler {
	return func(ctx context.Context, raw json.RawMessage) (ToolResult, error) {
		var args struct {
			Pattern string `json:"pattern"`
			Path    string `json:"path"`
			Include string `json:"include"`
		}
		if err := json.Unmarshal(raw, &args); err != nil {
			return ToolResult{}, fmt.Errorf("invalid arguments: %w", err)
		}

		if args.Pattern == "" {
			return ToolResult{Content: "pattern is required", IsError: true}, nil
		}

		if len(args.Pattern) > 500 {
			return ToolResult{Content: "pattern too long (max 500 chars)", IsError: true}, nil
		}

		searchPath := "."
		if args.Path != "" {
			searchPath = args.Path
		}

		cmdArgs := []string{"-rn", "-m", "50", "-E", args.Pattern}
		if args.Include != "" {
			cmdArgs = append(cmdArgs, "--include="+args.Include)
		}
		cmdArgs = append(cmdArgs, searchPath)

		cmd := exec.CommandContext(ctx, "grep", cmdArgs...)
		out, err := cmd.Output()

		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				return ToolResult{Content: "no matches found"}, nil
			}
			// grep not available, try manual approach
			if isCommandNotFound(err) {
				return ToolResult{
					Content: "grep not available on this system",
					IsError: true,
				}, nil
			}
			return ToolResult{
				Content: fmt.Sprintf("grep error: %s", err),
				IsError: true,
			}, nil
		}

		result := string(out)
		lines := strings.Split(strings.TrimSpace(result), "\n")
		if len(lines) > 50 {
			lines = lines[:50]
			result = strings.Join(lines, "\n") + "\n... (truncated to 50 results)"
		}

		return ToolResult{Content: strings.TrimSpace(result)}, nil
	}
}

func isCommandNotFound(err error) bool {
	if exitErr, ok := err.(*exec.Error); ok {
		return exitErr.Err == exec.ErrNotFound
	}
	return false
}
