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
	styles   Styles
	section  section
	width    int
	height   int
	overview OverviewModel
	blogs    BlogsModel
	contact  ContactModel
	music    MusicModel
}

func New(posts []content.Post, width, height int, renderer *lipgloss.Renderer, caps Capabilities) App {
	styles := NewStyles(renderer, caps)
	contentH := height - 4
	return App{
		styles:   styles,
		section:  sectionOverview,
		width:    width,
		height:   height,
		overview: newOverview(styles, width, contentH),
		blogs:    newBlogs(styles, posts, width, contentH),
		contact:  newContact(styles, width, contentH),
		music:    newMusic(styles, width, contentH),
	}
}

func (a App) Init() tea.Cmd { return nil }

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tracksFetchedMsg, tracksFetchErrMsg:
		var cmd tea.Cmd
		a.music, cmd = a.music.handleMsg(msg)
		return a, cmd
	}

	switch msg := msg.(type) {
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
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

		if msg.String() == "q" {
			if a.section == sectionBlogs && a.blogs.inPost() {
				var cmd tea.Cmd
				a.blogs, cmd = a.blogs.handleMsg(msg)
				return a, cmd
			}
			if a.section == sectionContact && a.contact.hasFocusedInput() {
				var cmd tea.Cmd
				a.contact, cmd = a.contact.handleMsg(msg)
				return a, cmd
			}
			return a, tea.Quit
		}

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

	s := a.styles
	r := s.Renderer

	topPad := r.NewStyle().Background(colorBg).Width(a.width).Render("")
	nav := a.navBar()
	sep := s.Dim.Render(strings.Repeat(s.Sep(), a.width))
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
	s := a.styles
	r := s.Renderer

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
			parts = append(parts, s.NavItemActive.Render(label))
		} else {
			parts = append(parts, s.NavItem.Render(label))
		}
	}

	bar := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	return r.NewStyle().
		Background(colorBg).
		Width(a.width).
		Align(lipgloss.Center).
		Render(bar)
}

func (a App) footerBar() string {
	s := a.styles
	var hints string
	switch a.section {
	case sectionBlogs:
		if a.blogs.inPost() {
			hints = "[" + s.ArrowUp() + s.ArrowDown() + "] scroll  [q] back  [ctrl+c] quit"
		} else {
			hints = "[" + s.ArrowUp() + s.ArrowDown() + "] navigate  [enter] read  [1/2/3/4] sections  [q] quit"
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
	return s.Footer.Width(a.width).Render(hints + strings.Repeat(" ", padding))
}
