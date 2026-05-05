package ui

import "github.com/charmbracelet/lipgloss"

const (
	colorAccent = lipgloss.Color("#BD93F9")
	colorMuted  = lipgloss.Color("#666666")
	colorText   = lipgloss.Color("#DDDDDD")
	colorDim    = lipgloss.Color("#444444")
	colorBg     = lipgloss.Color("#141414")
)

type Styles struct {
	Renderer      *lipgloss.Renderer
	Caps          Capabilities
	NavBar        lipgloss.Style
	NavItem       lipgloss.Style
	NavItemActive lipgloss.Style
	Footer        lipgloss.Style
	Accent        lipgloss.Style
	Muted         lipgloss.Style
	Dim           lipgloss.Style
	InputLabel    lipgloss.Style
	Error         lipgloss.Style
	Success       lipgloss.Style
}

func NewStyles(r *lipgloss.Renderer, caps Capabilities) Styles {
	return Styles{
		Renderer: r,
		Caps:     caps,
		NavBar:   r.NewStyle().Background(colorBg),
		NavItem: r.NewStyle().
			Foreground(colorMuted).
			Background(colorBg).
			Padding(0, 2),
		NavItemActive: r.NewStyle().
			Foreground(colorAccent).
			Background(colorBg).
			Bold(true).
			Padding(0, 2),
		Footer: r.NewStyle().
			Foreground(colorDim).
			Background(colorBg).
			Padding(0, 1),
		Accent:     r.NewStyle().Foreground(colorAccent).Bold(true),
		Muted:      r.NewStyle().Foreground(colorMuted),
		Dim:        r.NewStyle().Foreground(colorDim),
		InputLabel: r.NewStyle().Foreground(colorMuted),
		Error:      r.NewStyle().Foreground(lipgloss.Color("#FF5555")),
		Success:    r.NewStyle().Foreground(lipgloss.Color("#50FA7B")),
	}
}

func (s Styles) Bullet() string {
	if s.Caps.HasUnicode {
		return "▸"
	}
	return ">"
}

func (s Styles) Check() string {
	if s.Caps.HasUnicode {
		return "✓"
	}
	return "[OK]"
}

func (s Styles) Cross() string {
	if s.Caps.HasUnicode {
		return "✗"
	}
	return "[ERR]"
}

func (s Styles) Dot() string {
	if s.Caps.HasUnicode {
		return "·"
	}
	return "-"
}

func (s Styles) ArrowUp() string {
	if s.Caps.HasUnicode {
		return "↑"
	}
	return "^"
}

func (s Styles) ArrowDown() string {
	if s.Caps.HasUnicode {
		return "↓"
	}
	return "v"
}

func (s Styles) NowPlaying() string {
	if s.Caps.HasUnicode {
		return "▶"
	}
	return ">"
}

func (s Styles) Sep() string {
	if s.Caps.HasUnicode {
		return "─"
	}
	return "-"
}
