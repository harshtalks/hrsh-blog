package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/harshtalks/hrsh-blog/internal/content"
)

type blogsState int

const (
	blogsStateList blogsState = iota
	blogsStatePost
)

const (
	linesPerPost  = 4 // title + meta + 2 blank lines
	listHeaderH   = 3 // heading + separator + blank
	listFooterH   = 3 // separator + blank + hints
	listTopPad    = 2
)

type BlogsModel struct {
	posts        []content.Post
	cursor       int
	scrollOffset int
	state        blogsState
	vp           viewport.Model
	width        int
	height       int
}

const postOverhead = 6 // header(3) + blank(1) + blank(1) + footer(1)

func newBlogs(posts []content.Post, width, height int) BlogsModel {
	return BlogsModel{
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

func (b BlogsModel) inPost() bool {
	return b.state == blogsStatePost
}

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
				rendered := renderPost(b.posts[b.cursor], b.vp.Width)
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
	if len(b.posts) == 0 {
		return lipgloss.Place(b.width, b.height, lipgloss.Center, lipgloss.Center,
			mutedStyle.Render("No posts yet."))
	}

	listW := min(70, b.width-8)
	visible := b.visibleCount()
	offset := b.clampedOffset(visible)
	end := min(offset+visible, len(b.posts))
	window := b.posts[offset:end]

	// position indicator: "4 / 13"
	posLabel := dimStyle.Render(fmt.Sprintf("%d / %d", b.cursor+1, len(b.posts)))

	// scroll hints
	var scrollHints string
	if offset > 0 {
		scrollHints += dimStyle.Render("↑ more  ")
	}
	if end < len(b.posts) {
		scrollHints += dimStyle.Render("↓ more")
	}

	var rows []string
	rows = append(rows, accentStyle.Render("Writing")+"  "+posLabel)
	rows = append(rows, dimStyle.Render(strings.Repeat("─", listW)))
	rows = append(rows, "")

	for i, post := range window {
		absIdx := offset + i
		date := mutedStyle.Render(post.Date)
		readTime := dimStyle.Render("· " + post.ReadingTime())

		// truncate title so it never wraps — each post must be exactly linesPerPost lines
		maxTitleLen := listW - 6
		titleText := post.Title
		runes := []rune(titleText)
		if len(runes) > maxTitleLen {
			titleText = string(runes[:maxTitleLen-1]) + "…"
		}

		var title string
		if absIdx == b.cursor {
			title = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true).
				Render("▸  " + titleText)
		} else {
			title = lipgloss.NewStyle().
				Foreground(colorText).
				Render("   " + titleText)
		}

		meta := lipgloss.NewStyle().
			Foreground(colorMuted).
			PaddingLeft(3).
			Render(date + "  " + readTime)

		rows = append(rows, title)
		rows = append(rows, meta)
		rows = append(rows, "")
		rows = append(rows, "")
	}

	rows = append(rows, dimStyle.Render(strings.Repeat("─", listW)))
	rows = append(rows, "")

	hintLine := "[↑↓ / j k] navigate   [enter] read"
	if scrollHints != "" {
		hintLine = hintLine + "   " + scrollHints
	}
	rows = append(rows, dimStyle.Render(hintLine))

	block := lipgloss.NewStyle().
		Width(listW).
		Render(strings.Join(rows, "\n"))

	return lipgloss.Place(b.width, b.height, lipgloss.Center, lipgloss.Top,
		lipgloss.NewStyle().PaddingTop(listTopPad).Render(block))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (b BlogsModel) postView() string {
	post := b.posts[b.cursor]
	postW := min(80, b.width-8)
	b.vp.Width = postW
	b.vp.Height = b.height - postOverhead

	header := lipgloss.JoinVertical(lipgloss.Left,
		accentStyle.Render(post.Title),
		mutedStyle.Render(post.Date+"  "+dimStyle.Render("· "+post.ReadingTime())),
		dimStyle.Render(strings.Repeat("─", postW)),
	)

	scrollPct := fmt.Sprintf("%d%%", int(b.vp.ScrollPercent()*100))
	footer := lipgloss.NewStyle().
		Width(postW).
		Render(dimStyle.Render("[q / esc] back") +
			strings.Repeat(" ", max(0, postW-lipgloss.Width("[q / esc] back")-len(scrollPct))) +
			dimStyle.Render(scrollPct))

	block := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		b.vp.View(),
		"",
		footer,
	)

	return lipgloss.Place(b.width, b.height, lipgloss.Center, lipgloss.Top,
		lipgloss.NewStyle().Width(postW).Render(block))
}

func renderPost(post content.Post, width int) string {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
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
