package harness

import (
	"context"
	"encoding/json"
	"fmt"
)

func RegisterThinkTool(r *Registry) {
	r.Register(ToolDef{
		Name:        "think",
		Description: "use this to reason through a problem before acting. call before: critical decisions, changing approach, complex multi-step tasks. output is private.",
		ReadOnly:    true,
		Parameters: schemaObject(
			schemaProp("thought", "string", "your reasoning"),
			schemaRequired("thought"),
		),
	}, thinkHandler())
}

func thinkHandler() ToolHandler {
	return func(ctx context.Context, raw json.RawMessage) (ToolResult, error) {
		var args struct {
			Thought string `json:"thought"`
		}
		if err := json.Unmarshal(raw, &args); err != nil {
			return ToolResult{}, fmt.Errorf("invalid arguments: %w", err)
		}

		if args.Thought == "" {
			return ToolResult{Content: "thought is required", IsError: true}, nil
		}

		return ToolResult{
			Content: fmt.Sprintf("[thought recorded, %d chars]", len(args.Thought)),
		}, nil
	}
}
