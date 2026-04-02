package harness

// RegisterBuiltins registers all built-in tools on the registry.
// call this once during setup, before starting the agent loop.
func RegisterBuiltins(r *Registry) {
	RegisterFileTools(r)
	RegisterShellTool(r)
	RegisterSearchTool(r)
	RegisterThinkTool(r)
}

// DefaultAgent creates a fully wired agent with all built-in tools, loop detection, and default permissions.
func DefaultAgent(config AgentConfig) *Agent {
	registry := NewRegistry()
	RegisterBuiltins(registry)

	ld := NewLoopDetector()
	registry.SetLoopDetector(ld)

	perms := DefaultPermissions()
	registry.SetPermissions(perms)

	return NewAgent(config, registry)
}

// DefaultAgentWithPermissionCallback creates an agent that calls askFn for destructive tools.
// askFn blocks until the user responds (true = allow, false = deny).
func DefaultAgentWithPermissionCallback(config AgentConfig, askFn PermissionCallback) *Agent {
	registry := NewRegistry()
	RegisterBuiltins(registry)

	ld := NewLoopDetector()
	registry.SetLoopDetector(ld)

	perms := DefaultPermissions()
	perms.AskCallback = askFn
	registry.SetPermissions(perms)

	return NewAgent(config, registry)
}
