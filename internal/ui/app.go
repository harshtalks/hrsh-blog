package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harshtalks/hrsh-blog/internal/content"
)

type section int

const (
	sectionOverview section = iota
	sectionBlogs
	sectionContact
	sectionMusic
)

type App struct {
	section  section
	width    int
	height   int
	overview OverviewModel
	blogs    BlogsModel
	contact  ContactModel
	music    MusicModel
}

func New(posts []content.Post, width, height int) App {
	contentH := height - 4 // nav bar + separator line + footer bar
	return App{
		section:  sectionOverview,
		width:    width,
		height:   height,
		overview: newOverview(width, contentH),
		blogs:    newBlogs(posts, width, contentH),
		contact:  newContact(width, contentH),
		music:    newMusic(width, contentH),
	}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tracksFetchedMsg, tracksFetchErrMsg:
		var cmd tea.Cmd
		a.music, cmd = a.music.handleMsg(msg)
		return a, cmd

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		contentH := msg.Height - 4
		a.overview = a.overview.resize(msg.Width, contentH)
		a.blogs = a.blogs.resize(msg.Width, contentH)
		a.contact = a.contact.resize(msg.Width, contentH)
		a.music = a.music.resize(msg.Width, contentH)
		return a, nil

	case tea.KeyMsg:
		// ctrl+c always quits
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

		// q quits unless contact form is active or blogs is in post view
		if msg.String() == "q" {
			if a.section == sectionBlogs && a.blogs.inPost() {
				var cmd tea.Cmd
				a.blogs, cmd = a.blogs.handleMsg(msg)
				return a, cmd
			}
			if a.section == sectionContact && a.contact.hasFocusedInput() {
				// let contact handle q as a character
				var cmd tea.Cmd
				a.contact, cmd = a.contact.handleMsg(msg)
				return a, cmd
			}
			return a, tea.Quit
		}

		// nav shortcuts
		switch msg.String() {
		case "1":
			a.section = sectionOverview
			return a, nil
		case "2":
			a.section = sectionBlogs
			return a, nil
		case "3":
			a.section = sectionContact
			return a, nil
		case "4":
			a.section = sectionMusic
			if a.music.state == musicStateIdle {
				var cmd tea.Cmd
				a.music, cmd = a.music.startLoading()
				return a, cmd
			}
			return a, nil
		case "tab":
			// Only cycle nav if not inside a focused contact input
			if a.section != sectionContact || !a.contact.hasFocusedInput() {
				next := (a.section + 1) % 4
				a.section = next
				if next == sectionMusic && a.music.state == musicStateIdle {
					var cmd tea.Cmd
					a.music, cmd = a.music.startLoading()
					return a, cmd
				}
				return a, nil
			}
		case "shift+tab":
			if a.section != sectionContact || !a.contact.hasFocusedInput() {
				next := (a.section + 3) % 4
				a.section = next
				if next == sectionMusic && a.music.state == musicStateIdle {
					var cmd tea.Cmd
					a.music, cmd = a.music.startLoading()
					return a, cmd
				}
				return a, nil
			}
		}
	}

	// delegate to active section
	var cmd tea.Cmd
	switch a.section {
	case sectionBlogs:
		a.blogs, cmd = a.blogs.handleMsg(msg)
	case sectionContact:
		a.contact, cmd = a.contact.handleMsg(msg)
	case sectionMusic:
		a.music, cmd = a.music.handleMsg(msg)
	}
	return a, cmd
}

func (a App) View() string {
	if a.width < 40 || a.height < 10 {
		return "Terminal too small. Please resize to at least 40×10."
	}

	topPad := lipgloss.NewStyle().Background(colorBg).Width(a.width).Render("")
	nav := a.navBar()
	sep := dimStyle.Render(strings.Repeat("─", a.width))
	footer := a.footerBar()

	var body string
	switch a.section {
	case sectionOverview:
		body = a.overview.View()
	case sectionBlogs:
		body = a.blogs.View()
	case sectionContact:
		body = a.contact.View()
	case sectionMusic:
		body = a.music.View()
	}

	return strings.Join([]string{topPad, nav, sep, body, footer}, "\n")
}

func (a App) navBar() string {
	type item struct {
		label   string
		key     string
		section section
	}
	items := []item{
		{"Overview", "1", sectionOverview},
		{"Blogs", "2", sectionBlogs},
		{"Contact", "3", sectionContact},
		{"Listening", "4", sectionMusic},
	}

	var parts []string
	for _, it := range items {
		label := fmt.Sprintf("[%s] %s", it.key, it.label)
		if it.section == a.section {
			parts = append(parts, navItemActiveStyle.Render(label))
		} else {
			parts = append(parts, navItemStyle.Render(label))
		}
	}

	bar := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	return lipgloss.NewStyle().
		Background(colorBg).
		Width(a.width).
		Align(lipgloss.Center).
		Render(bar)
}

func (a App) footerBar() string {
	var hints string
	switch a.section {
	case sectionBlogs:
		if a.blogs.inPost() {
			hints = "[↑↓] scroll  [q] back  [ctrl+c] quit"
		} else {
			hints = "[↑↓] navigate  [enter] read  [1/2/3/4] sections  [q] quit"
		}
	case sectionContact:
		hints = "[tab] next field  [enter] submit  [1/2/3/4] sections  [ctrl+c] quit"
	case sectionMusic:
		hints = "[tab] navigate  [1/2/3/4] sections  [q] quit"
	default:
		hints = "[tab] navigate  [1/2/3/4] sections  [q] quit"
	}

	padding := a.width - lipgloss.Width(hints) - 2
	if padding < 0 {
		padding = 0
	}
	return footerStyle.Width(a.width).Render(hints + strings.Repeat(" ", padding))
}
