package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const asciiArt = `╦ ╦   ╦═╗   ╔═╗   ╦ ╦
╠═╣   ╠╦╝   ╚═╗   ╠═╣
╩ ╩   ╩╚═   ╚═╝   ╩ ╩`

type OverviewModel struct {
	styles Styles
	width  int
	height int
}

func newOverview(styles Styles, width, height int) OverviewModel {
	return OverviewModel{styles: styles, width: width, height: height}
}

func (o OverviewModel) resize(width, height int) OverviewModel {
	o.width = width
	o.height = height
	return o
}

func (o OverviewModel) View() string {
	s := o.styles
	r := s.Renderer

	var photo string
	if s.Caps.Profile == termenv.TrueColor && s.Caps.HasUnicode {
		photo = r.NewStyle().Align(lipgloss.Center).Render(photoArt)
	} else {
		photo = s.Muted.Render("[harsh@hrsh]")
	}

	var art string
	if s.Caps.HasUnicode {
		art = s.Accent.Render(asciiArt)
	} else {
		art = s.Accent.Render("H  R  S  H")
	}

	role := s.Muted.Render("Frontend Engineer 2 @ Intelxlabs")

	bio := r.NewStyle().
		Foreground(colorText).
		Align(lipgloss.Center).
		Render("I love Frontend Engineering and Capybaras (In that order). Based in India.")

	exploring := r.NewStyle().
		Foreground(colorText).
		Align(lipgloss.Center).
		Render("These days exploring HTTP Caching, REST API Design Patterns,\nand how to roll out a robust and secure Auth.")

	tagline := s.Accent.Render("Let's build awesome things together.")
	divider := s.Dim.Render(s.Dot() + "  " + s.Dot() + "  " + s.Dot())

	links := lipgloss.JoinVertical(lipgloss.Center,
		s.Muted.Render("web    ")+s.Accent.Render("hrshwrites.vercel.app"),
		s.Muted.Render("email  ")+s.Accent.Render("harshpareek.works@gmail.com"),
		s.Muted.Render("ssh    ")+s.Accent.Render("ssh switchback.proxy.rlwy.net -p 18516"),
	)

	block := lipgloss.JoinVertical(lipgloss.Center,
		photo,
		"",
		art,
		"",
		role,
		"",
		bio,
		"",
		exploring,
		"",
		divider,
		"",
		tagline,
		"",
		divider,
		"",
		links,
	)

	return lipgloss.Place(o.width, o.height, lipgloss.Center, lipgloss.Center, block)
}
