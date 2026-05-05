package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/harshtalks/hrsh-blog/internal/content"
	"github.com/muesli/termenv"
)

type blogsState int

const (
	blogsStateList blogsState = iota
	blogsStatePost
)

const (
	linesPerPost = 4
	listHeaderH  = 3
	listFooterH  = 3
	listTopPad   = 2
)

type BlogsModel struct {
	styles       Styles
	posts        []content.Post
	cursor       int
	scrollOffset int
	state        blogsState
	vp           viewport.Model
	width        int
	height       int
}

const postOverhead = 6

func newBlogs(styles Styles, posts []content.Post, width, height int) BlogsModel {
	return BlogsModel{
		styles: styles,
		posts:  posts,
		state:  blogsStateList,
		vp:     viewport.New(width-4, height-postOverhead),
		width:  width,
		height: height,
	}
}

func (b BlogsModel) resize(width, height int) BlogsModel {
	b.width = width
	b.height = height
	b.vp.Width = width - 4
	b.vp.Height = height - postOverhead
	b.scrollOffset = b.clampedOffset(b.visibleCount())
	return b
}

func (b BlogsModel) inPost() bool { return b.state == blogsStatePost }

func (b BlogsModel) visibleCount() int {
	available := b.height - listTopPad - listHeaderH - listFooterH
	n := available / linesPerPost
	if n < 1 {
		n = 1
	}
	return n
}

func (b BlogsModel) clampedOffset(visibleCount int) int {
	offset := b.scrollOffset
	if b.cursor < offset {
		offset = b.cursor
	}
	if b.cursor >= offset+visibleCount {
		offset = b.cursor - visibleCount + 1
	}
	maxOffset := max(0, len(b.posts)-visibleCount)
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		offset = 0
	}
	return offset
}

func (b BlogsModel) handleMsg(msg tea.Msg) (BlogsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if b.state == blogsStateList && b.cursor > 0 {
				b.cursor--
				b.scrollOffset = b.clampedOffset(b.visibleCount())
			}
		case "down", "j":
			if b.state == blogsStateList && b.cursor < len(b.posts)-1 {
				b.cursor++
				b.scrollOffset = b.clampedOffset(b.visibleCount())
			}
		case "enter":
			if b.state == blogsStateList && len(b.posts) > 0 {
				rendered := b.renderPost(b.posts[b.cursor], b.vp.Width)
				b.vp.SetContent(rendered)
				b.vp.GotoTop()
				b.state = blogsStatePost
			}
		case "q", "esc", "backspace":
			if b.state == blogsStatePost {
				b.state = blogsStateList
			}
		}
		if b.state == blogsStatePost {
			var cmd tea.Cmd
			b.vp, cmd = b.vp.Update(msg)
			return b, cmd
		}
	}
	return b, nil
}

func (b BlogsModel) View() string {
	if b.state == blogsStatePost {
		return b.postView()
	}
	return b.listView()
}

func (b BlogsModel) listView() string {
	s := b.styles
	r := s.Renderer

	if len(b.posts) == 0 {
		return lipgloss.Place(b.width, b.height, lipgloss.Center, lipgloss.Center,
			s.Muted.Render("No posts yet."))
	}

	listW := min(70, b.width-8)
	visible := b.visibleCount()
	offset := b.clampedOffset(visible)
	end := min(offset+visible, len(b.posts))
	window := b.posts[offset:end]

	posLabel := s.Dim.Render(fmt.Sprintf("%d / %d", b.cursor+1, len(b.posts)))

	var scrollHints string
	if offset > 0 {
		scrollHints += s.Dim.Render(s.ArrowUp() + " more  ")
	}
	if end < len(b.posts) {
		scrollHints += s.Dim.Render(s.ArrowDown() + " more")
	}

	var rows []string
	rows = append(rows, s.Accent.Render("Writing")+"  "+posLabel)
	rows = append(rows, s.Dim.Render(strings.Repeat(s.Sep(), listW)))
	rows = append(rows, "")

	for i, post := range window {
		absIdx := offset + i
		date := s.Muted.Render(post.Date)
		readTime := s.Dim.Render(s.Dot() + " " + post.ReadingTime())

		maxTitleLen := listW - 6
		titleText := post.Title
		runes := []rune(titleText)
		if len(runes) > maxTitleLen {
			if s.Caps.HasUnicode {
				titleText = string(runes[:maxTitleLen-1]) + "…"
			} else {
				titleText = string(runes[:maxTitleLen-3]) + "..."
			}
		}

		var title string
		if absIdx == b.cursor {
			title = r.NewStyle().Foreground(colorAccent).Bold(true).Render(s.Bullet() + "  " + titleText)
		} else {
			title = r.NewStyle().Foreground(colorText).Render("   " + titleText)
		}

		meta := r.NewStyle().Foreground(colorMuted).PaddingLeft(3).Render(date + "  " + readTime)

		rows = append(rows, title)
		rows = append(rows, meta)
		rows = append(rows, "")
		rows = append(rows, "")
	}

	rows = append(rows, s.Dim.Render(strings.Repeat(s.Sep(), listW)))
	rows = append(rows, "")

	hintLine := "[" + s.ArrowUp() + s.ArrowDown() + " / j k] navigate   [enter] read"
	if scrollHints != "" {
		hintLine = hintLine + "   " + scrollHints
	}
	rows = append(rows, s.Dim.Render(hintLine))

	block := r.NewStyle().Width(listW).Render(strings.Join(rows, "\n"))
	return lipgloss.Place(b.width, b.height, lipgloss.Center, lipgloss.Top,
		r.NewStyle().PaddingTop(listTopPad).Render(block))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (b BlogsModel) postView() string {
	s := b.styles
	r := s.Renderer
	post := b.posts[b.cursor]
	postW := min(80, b.width-8)
	b.vp.Width = postW
	b.vp.Height = b.height - postOverhead

	header := lipgloss.JoinVertical(lipgloss.Left,
		s.Accent.Render(post.Title),
		s.Muted.Render(post.Date+"  "+s.Dim.Render(s.Dot()+" "+post.ReadingTime())),
		s.Dim.Render(strings.Repeat(s.Sep(), postW)),
	)

	scrollPct := fmt.Sprintf("%d%%", int(b.vp.ScrollPercent()*100))
	footer := r.NewStyle().
		Width(postW).
		Render(s.Dim.Render("[q / esc] back") +
			strings.Repeat(" ", max(0, postW-lipgloss.Width("[q / esc] back")-len(scrollPct))) +
			s.Dim.Render(scrollPct))

	block := lipgloss.JoinVertical(lipgloss.Left,
		header, "", b.vp.View(), "", footer,
	)

	return lipgloss.Place(b.width, b.height, lipgloss.Center, lipgloss.Top,
		r.NewStyle().Width(postW).Render(block))
}

func (b BlogsModel) renderPost(post content.Post, width int) string {
	var style string
	switch b.styles.Caps.Profile {
	case termenv.TrueColor, termenv.ANSI256:
		style = "dark"
	case termenv.ANSI:
		style = "ascii"
	default:
		style = "notty"
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return post.Body
	}
	rendered, err := renderer.Render(post.Body)
	if err != nil {
		return post.Body
	}
	return rendered
}
