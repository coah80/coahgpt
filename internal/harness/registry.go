package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type registeredTool struct {
	def     ToolDef
	handler ToolHandler
}

type EventCallback func(Event)

type Registry struct {
	mu           sync.RWMutex
	tools        map[string]registeredTool
	callbacks    []EventCallback
	loopDetector *LoopDetector
	permissions  *PermissionConfig
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]registeredTool),
	}
}

func (r *Registry) Register(def ToolDef, handler ToolHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[def.Name] = registeredTool{def: def, handler: handler}
}

func (r *Registry) OnEvent(cb EventCallback) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.callbacks = append(r.callbacks, cb)
}

func (r *Registry) SetLoopDetector(ld *LoopDetector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.loopDetector = ld
}

func (r *Registry) SetPermissions(config *PermissionConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.permissions = config
}

func (r *Registry) emit(evt Event) {
	r.mu.RLock()
	cbs := make([]EventCallback, len(r.callbacks))
	copy(cbs, r.callbacks)
	r.mu.RUnlock()

	for _, cb := range cbs {
		cb(evt)
	}
}

func (r *Registry) Execute(ctx context.Context, name string, args json.RawMessage) (ToolResult, error) {
	r.mu.RLock()
	tool, ok := r.tools[name]
	ld := r.loopDetector
	perms := r.permissions
	r.mu.RUnlock()

	if !ok {
		return ToolResult{
			Content: fmt.Sprintf("unknown tool: %s", name),
			IsError: true,
		}, nil
	}

	if ld != nil {
		action := ld.Record(name, string(args))
		switch action {
		case LoopAbort:
			return ToolResult{
				Content: "tool loop detected. you've called this with the same args too many times. try a completely different approach.",
				IsError: true,
			}, nil
		case LoopForceText:
			return ToolResult{
				Content: "you're repeating yourself. stop calling tools and explain what you're trying to do instead.",
				IsError: true,
			}, nil
		case LoopWarn:
			r.emit(Event{
				Type:     EventToolProgress,
				ToolName: name,
				Content:  "possible loop detected, same tool+args called multiple times",
			})
		}
	}

	argsStr := string(args)

	// permission check before execution
	if perms != nil {
		r.emit(Event{
			Type:     EventPermissionAsk,
			ToolName: name,
			ToolArgs: argsStr,
		})
		if reason, allowed := perms.checkPermission(name, argsStr); !allowed {
			return ToolResult{
				Content: reason,
				IsError: true,
			}, nil
		}
	}

	r.emit(Event{
		Type:     EventToolStart,
		ToolName: name,
		ToolArgs: argsStr,
	})

	toolCtx := ContextWithToolName(ctx, name)
	result, err := tool.handler(toolCtx, args)

	if err != nil {
		r.emit(Event{
			Type:     EventToolError,
			ToolName: name,
			Content:  err.Error(),
		})
		return ToolResult{
			Content: err.Error(),
			IsError: true,
		}, nil
	}

	r.emit(Event{
		Type:     EventToolComplete,
		ToolName: name,
		Content:  result.Content,
	})

	return result, nil
}

func (r *Registry) ExecuteParallel(ctx context.Context, calls []ToolCall) []ToolResult {
	// figure out if all calls are read-only
	r.mu.RLock()
	allReadOnly := true
	for _, call := range calls {
		if tool, ok := r.tools[call.Function.Name]; ok {
			if !tool.def.ReadOnly {
				allReadOnly = false
				break
			}
		}
	}
	r.mu.RUnlock()

	results := make([]ToolResult, len(calls))

	if allReadOnly && len(calls) > 1 {
		var wg sync.WaitGroup
		for i, call := range calls {
			wg.Add(1)
			go func(idx int, c ToolCall) {
				defer wg.Done()
				res, _ := r.Execute(ctx, c.Function.Name, c.Function.Arguments)
				results[idx] = res
			}(i, call)
		}
		wg.Wait()
	} else {
		for i, call := range calls {
			res, _ := r.Execute(ctx, call.Function.Name, call.Function.Arguments)
			results[i] = res
		}
	}

	return results
}

// OllamaToolDefs converts registered tools to the format Ollama expects.
func (r *Registry) OllamaToolDefs() []map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]map[string]interface{}, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        t.def.Name,
				"description": t.def.Description,
				"parameters":  t.def.Parameters,
			},
		})
	}
	return defs
}

func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.tools[name]
	return ok
}

func (r *Registry) List() []ToolDef {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]ToolDef, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, t.def)
	}
	return defs
}
