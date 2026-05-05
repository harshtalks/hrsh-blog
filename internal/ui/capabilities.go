package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/muesli/termenv"
)

type Capabilities struct {
	Profile    termenv.Profile
	HasUnicode bool
	Term       string
}

func DetectCapabilities(sess ssh.Session, renderer *lipgloss.Renderer) Capabilities {
	pty, _, active := sess.Pty()
	term := ""
	if active {
		term = pty.Term
	}

	profile := renderer.ColorProfile()

	// Heuristic: most terminals that support colors also support Unicode,
	// except explicitly limited ones.
	hasUnicode := profile != termenv.Ascii
	low := strings.ToLower(term)
	if strings.Contains(low, "dumb") || strings.Contains(low, "ascii") || strings.Contains(low, "cygwin") {
		hasUnicode = false
	}

	return Capabilities{
		Profile:    profile,
		HasUnicode: hasUnicode,
		Term:       term,
	}
}
