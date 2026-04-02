package tui

import "github.com/charmbracelet/lipgloss"

// catppuccin mocha
var (
	ColorBase     = lipgloss.Color("#1e1e2e")
	ColorMantle   = lipgloss.Color("#181825")
	ColorSurface0 = lipgloss.Color("#313244")
	ColorSurface1 = lipgloss.Color("#45475a")
	ColorText     = lipgloss.Color("#cdd6f4")
	ColorSubtext  = lipgloss.Color("#a6adc8")
	ColorMauve    = lipgloss.Color("#cba6f7")
	ColorLavender = lipgloss.Color("#b4befe")
	ColorPink     = lipgloss.Color("#f5c2e7")
	ColorGreen    = lipgloss.Color("#a6e3a1")
	ColorRed      = lipgloss.Color("#f38ba8")
	ColorYellow   = lipgloss.Color("#f9e2af")
)

var (
	UserPromptStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext)

	AssistantResponseGutter = lipgloss.NewStyle().
				Foreground(ColorSurface1)

	SystemMsgStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			Italic(true)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorMauve)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)

	ToolLabelStyle = lipgloss.NewStyle().
			Foreground(ColorMauve).
			Bold(true)

	ToolContentStyle = lipgloss.NewStyle().
				Foreground(ColorSubtext)

	ToolSuccessStyle = lipgloss.NewStyle().
				Foreground(ColorGreen)

	ToolErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed)

	WelcomeNameStyle = lipgloss.NewStyle().
				Foreground(ColorMauve).
				Bold(true)

	WelcomeDimStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext)

	PermissionStyle = lipgloss.NewStyle().
			Foreground(ColorYellow).
			Bold(true)
)
