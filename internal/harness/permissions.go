package harness

import "fmt"

// PermissionLevel controls how a tool execution is gated.
type PermissionLevel int

const (
	PermAllow PermissionLevel = iota // auto-allow (read-only tools)
	PermAsk                          // ask user before executing (destructive tools)
	PermDeny                         // never allow
)

// PermissionCallback is invoked when a tool needs user approval.
// toolName is the tool being called, args is the raw JSON arguments.
// Returns true if the user approves execution.
type PermissionCallback func(toolName string, args string) (approved bool)

// PermissionConfig controls tool execution gating.
type PermissionConfig struct {
	DefaultMode PermissionLevel
	AlwaysAllow map[string]bool      // tool names that auto-allow
	AlwaysDeny  map[string]bool      // tool names that auto-deny
	AskCallback PermissionCallback   // called when tool needs approval
}

// DefaultPermissions returns a config where read-only tools auto-allow
// and destructive tools require user approval.
func DefaultPermissions() *PermissionConfig {
	return &PermissionConfig{
		DefaultMode: PermAsk,
		AlwaysAllow: map[string]bool{
			"read_file":  true,
			"list_files": true,
			"grep":       true,
			"think":      true,
		},
		AlwaysDeny:  map[string]bool{},
		AskCallback: nil,
	}
}

// checkPermission determines whether a tool call should proceed.
// Returns ("", true) if allowed, or (reason, false) if denied.
func (pc *PermissionConfig) checkPermission(toolName string, argsStr string) (string, bool) {
	if pc == nil {
		return "", true
	}

	if pc.AlwaysDeny[toolName] {
		return fmt.Sprintf("tool %q is denied by policy", toolName), false
	}

	if pc.AlwaysAllow[toolName] {
		return "", true
	}

	level := pc.DefaultMode

	switch level {
	case PermAllow:
		return "", true
	case PermDeny:
		return fmt.Sprintf("tool %q is denied by default policy", toolName), false
	case PermAsk:
		if pc.AskCallback == nil {
			// no callback set, default to allow to avoid blocking
			return "", true
		}
		if pc.AskCallback(toolName, argsStr) {
			return "", true
		}
		return "tool execution denied by user", false
	}

	return "", true
}
