package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func RegisterFileTools(r *Registry) {
	r.Register(ToolDef{
		Name:        "read_file",
		Description: "read a file and return its contents with line numbers. use offset/limit for large files.",
		ReadOnly:    true,
		Parameters: schemaObject(
			schemaProp("file_path", "string", "absolute path to the file"),
			schemaProp("offset", "integer", "line number to start from (1-based, optional)"),
			schemaProp("limit", "integer", "max lines to read (default 2000, optional)"),
			schemaRequired("file_path"),
		),
	}, readFileHandler())

	r.Register(ToolDef{
		Name:        "write_file",
		Description: "create or overwrite a file. creates parent directories if needed.",
		Destructive: true,
		Parameters: schemaObject(
			schemaProp("file_path", "string", "absolute path to the file"),
			schemaProp("content", "string", "the full file content to write"),
			schemaRequired("file_path", "content"),
		),
	}, writeFileHandler())

	r.Register(ToolDef{
		Name:        "edit_file",
		Description: "replace an exact string in a file. fails if old_string is not found or matches multiple locations.",
		Destructive: true,
		Parameters: schemaObject(
			schemaProp("file_path", "string", "absolute path to the file"),
			schemaProp("old_string", "string", "the exact text to find (must be unique in the file)"),
			schemaProp("new_string", "string", "the replacement text"),
			schemaRequired("file_path", "old_string", "new_string"),
		),
	}, editFileHandler())

	r.Register(ToolDef{
		Name:        "list_files",
		Description: "list files matching a glob pattern. returns up to 100 results.",
		ReadOnly:    true,
		Parameters: schemaObject(
			schemaProp("pattern", "string", "glob pattern like /path/to/*.go or /path/**/*.ts"),
			schemaRequired("pattern"),
		),
	}, listFilesHandler())
}

func readFileHandler() ToolHandler {
	return func(ctx context.Context, raw json.RawMessage) (ToolResult, error) {
		var args struct {
			FilePath string `json:"file_path"`
			Offset   int    `json:"offset"`
			Limit    int    `json:"limit"`
		}
		if err := json.Unmarshal(raw, &args); err != nil {
			return ToolResult{}, fmt.Errorf("invalid arguments: %w", err)
		}

		if args.FilePath == "" {
			return ToolResult{Content: "file_path is required", IsError: true}, nil
		}

		data, err := os.ReadFile(args.FilePath)
		if err != nil {
			if os.IsNotExist(err) {
				return ToolResult{Content: fmt.Sprintf("file not found: %s", args.FilePath), IsError: true}, nil
			}
			return ToolResult{}, fmt.Errorf("reading file: %w", err)
		}

		lines := strings.Split(string(data), "\n")

		start := 0
		if args.Offset > 0 {
			start = args.Offset - 1
			if start >= len(lines) {
				return ToolResult{Content: "offset beyond end of file", IsError: true}, nil
			}
		}

		limit := 2000
		if args.Limit > 0 {
			limit = args.Limit
		}

		end := start + limit
		if end > len(lines) {
			end = len(lines)
		}

		var sb strings.Builder
		for i := start; i < end; i++ {
			sb.WriteString(fmt.Sprintf("%6d\t%s\n", i+1, lines[i]))
		}

		out := sb.String()
		if end < len(lines) {
			out += fmt.Sprintf("\n... (%d more lines)", len(lines)-end)
		}

		return ToolResult{Content: out}, nil
	}
}

func writeFileHandler() ToolHandler {
	return func(ctx context.Context, raw json.RawMessage) (ToolResult, error) {
		var args struct {
			FilePath string `json:"file_path"`
			Content  string `json:"content"`
		}
		if err := json.Unmarshal(raw, &args); err != nil {
			return ToolResult{}, fmt.Errorf("invalid arguments: %w", err)
		}

		if args.FilePath == "" {
			return ToolResult{Content: "file_path is required", IsError: true}, nil
		}

		parentDir := filepath.Dir(args.FilePath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return ToolResult{}, fmt.Errorf("creating directories: %w", err)
		}

		if err := os.WriteFile(args.FilePath, []byte(args.Content), 0644); err != nil {
			return ToolResult{}, fmt.Errorf("writing file: %w", err)
		}

		return ToolResult{
			Content: fmt.Sprintf("wrote %d bytes to %s", len(args.Content), args.FilePath),
		}, nil
	}
}

func editFileHandler() ToolHandler {
	return func(ctx context.Context, raw json.RawMessage) (ToolResult, error) {
		var args struct {
			FilePath  string `json:"file_path"`
			OldString string `json:"old_string"`
			NewString string `json:"new_string"`
		}
		if err := json.Unmarshal(raw, &args); err != nil {
			return ToolResult{}, fmt.Errorf("invalid arguments: %w", err)
		}

		if args.FilePath == "" || args.OldString == "" {
			return ToolResult{Content: "file_path and old_string are required", IsError: true}, nil
		}

		data, err := os.ReadFile(args.FilePath)
		if err != nil {
			if os.IsNotExist(err) {
				return ToolResult{Content: fmt.Sprintf("file not found: %s", args.FilePath), IsError: true}, nil
			}
			return ToolResult{}, fmt.Errorf("reading file: %w", err)
		}

		content := string(data)
		count := strings.Count(content, args.OldString)

		if count == 0 {
			return ToolResult{
				Content: "old_string not found in file. make sure it matches exactly, including whitespace.",
				IsError: true,
			}, nil
		}

		if count > 1 {
			return ToolResult{
				Content: fmt.Sprintf("old_string found %d times. provide more context to make it unique.", count),
				IsError: true,
			}, nil
		}

		newContent := strings.Replace(content, args.OldString, args.NewString, 1)
		if err := os.WriteFile(args.FilePath, []byte(newContent), 0644); err != nil {
			return ToolResult{}, fmt.Errorf("writing file: %w", err)
		}

		result := "edit applied successfully"
		diff := UnifiedDiff(content, newContent, filepath.Base(args.FilePath))
		if diff != "" {
			result += "\n" + diff
		}

		return ToolResult{Content: result}, nil
	}
}

func listFilesHandler() ToolHandler {
	return func(ctx context.Context, raw json.RawMessage) (ToolResult, error) {
		var args struct {
			Pattern string `json:"pattern"`
		}
		if err := json.Unmarshal(raw, &args); err != nil {
			return ToolResult{}, fmt.Errorf("invalid arguments: %w", err)
		}

		if args.Pattern == "" {
			return ToolResult{Content: "pattern is required", IsError: true}, nil
		}

		matches, err := filepath.Glob(args.Pattern)
		if err != nil {
			return ToolResult{Content: fmt.Sprintf("invalid glob pattern: %s", err), IsError: true}, nil
		}

		const maxResults = 100
		if len(matches) > maxResults {
			matches = matches[:maxResults]
		}

		if len(matches) == 0 {
			return ToolResult{Content: "no files matched"}, nil
		}

		return ToolResult{Content: strings.Join(matches, "\n")}, nil
	}
}
