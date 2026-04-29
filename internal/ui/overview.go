package ui

import (
	"github.com/charmbracelet/lipgloss"
)

const asciiArt = `╦ ╦   ╦═╗   ╔═╗   ╦ ╦
╠═╣   ╠╦╝   ╚═╗   ╠═╣
╩ ╩   ╩╚═   ╚═╝   ╩ ╩`

type OverviewModel struct {
	width  int
	height int
}

func newOverview(width, height int) OverviewModel {
	return OverviewModel{width: width, height: height}
}

func (o OverviewModel) resize(width, height int) OverviewModel {
	o.width = width
	o.height = height
	return o
}

func (o OverviewModel) View() string {
	photo := lipgloss.NewStyle().Align(lipgloss.Center).Render(photoArt)

	art := accentStyle.Render(asciiArt)

	role := mutedStyle.Render("Frontend Engineer 2 @ Intelxlabs")

	bio := lipgloss.NewStyle().
		Foreground(colorText).
		Align(lipgloss.Center).
		Render("I love Frontend Engineering and Capybaras (In that order). Based in India.")

	exploring := lipgloss.NewStyle().
		Foreground(colorText).
		Align(lipgloss.Center).
		Render("These days exploring HTTP Caching, REST API Design Patterns,\nand how to roll out a robust and secure Auth.")

	tagline := accentStyle.Render("Let's build awesome things together.")

	divider := dimStyle.Render("·  ·  ·")

	links := lipgloss.JoinVertical(lipgloss.Center,
		mutedStyle.Render("web    ")+accentStyle.Render("hrshwrites.vercel.app"),
		mutedStyle.Render("email  ")+accentStyle.Render("harshpareek.works@gmail.com"),
		mutedStyle.Render("ssh    ")+accentStyle.Render("ssh hrsh-ssh.fly.dev"),
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
