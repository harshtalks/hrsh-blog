package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harshtalks/hrsh-blog/internal/config"
)

type musicState int

const (
	musicStateIdle musicState = iota
	musicStateLoading
	musicStateLoaded
	musicStateError
)

type Track struct {
	Name       string
	Artist     string
	PlayedAt   time.Time
	NowPlaying bool
}

type tracksFetchedMsg struct{ tracks []Track }
type tracksFetchErrMsg struct{ err error }

type MusicModel struct {
	state   musicState
	tracks  []Track
	errMsg  string
	spinner spinner.Model
	width   int
	height  int
}

func newMusic(width, height int) MusicModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorAccent)
	return MusicModel{
		state:   musicStateIdle,
		spinner: sp,
		width:   width,
		height:  height,
	}
}

func (m MusicModel) resize(width, height int) MusicModel {
	m.width = width
	m.height = height
	return m
}

func (m MusicModel) startLoading() (MusicModel, tea.Cmd) {
	m.state = musicStateLoading
	return m, tea.Batch(m.spinner.Tick, fetchTracks())
}

func fetchTracks() tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf(
			"https://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks&user=%s&api_key=%s&format=json&limit=%d",
			config.LastFMUser, config.LastFMKey, config.LastFMLimit,
		)
		resp, err := http.Get(url) //nolint:noctx
		if err != nil {
			return tracksFetchErrMsg{err}
		}
		defer resp.Body.Close()

		var raw struct {
			RecentTracks struct {
				Track []struct {
					Name   string `json:"name"`
					Artist struct {
						Text string `json:"#text"`
					} `json:"artist"`
					Date struct {
						UTS string `json:"uts"`
					} `json:"date"`
					Attr struct {
						NowPlaying string `json:"nowplaying"`
					} `json:"@attr"`
				} `json:"track"`
			} `json:"recenttracks"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
			return tracksFetchErrMsg{err}
		}

		tracks := make([]Track, 0, len(raw.RecentTracks.Track))
		for _, t := range raw.RecentTracks.Track {
			track := Track{
				Name:       t.Name,
				Artist:     t.Artist.Text,
				NowPlaying: t.Attr.NowPlaying == "true",
			}
			if t.Date.UTS != "" {
				var uts int64
				fmt.Sscan(t.Date.UTS, &uts)
				track.PlayedAt = time.Unix(uts, 0)
			}
			tracks = append(tracks, track)
		}
		return tracksFetchedMsg{tracks}
	}
}

func (m MusicModel) handleMsg(msg tea.Msg) (MusicModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tracksFetchedMsg:
		m.state = musicStateLoaded
		m.tracks = msg.tracks
		return m, nil
	case tracksFetchErrMsg:
		m.state = musicStateError
		m.errMsg = msg.err.Error()
		return m, nil
	case spinner.TickMsg:
		if m.state == musicStateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	case d < 24*time.Hour:
		hrs := int(d.Hours())
		if hrs == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hrs)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2")
	}
}

func (m MusicModel) View() string {
	switch m.state {
	case musicStateIdle, musicStateLoading:
		block := m.spinner.View() + "  " + mutedStyle.Render("Loading recent tracks...")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, block)
	case musicStateError:
		block := lipgloss.JoinVertical(lipgloss.Center,
			errorStyle.Render("✗  Could not load tracks"),
			"",
			mutedStyle.Render(m.errMsg),
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, block)
	}
	return m.tracksView()
}

func (m MusicModel) tracksView() string {
	listW := min(70, m.width-8)

	var rows []string
	rows = append(rows, accentStyle.Render("Listening")+"  "+dimStyle.Render("recent plays · @"+config.LastFMUser))
	rows = append(rows, dimStyle.Render(strings.Repeat("─", listW)))
	rows = append(rows, "")

	for _, t := range m.tracks {
		var when string
		if t.NowPlaying {
			when = accentStyle.Render("▶  now playing")
		} else {
			when = dimStyle.Render(relativeTime(t.PlayedAt))
		}

		trackLine := lipgloss.NewStyle().
			Foreground(colorText).
			Bold(t.NowPlaying).
			Render(t.Name)

		artistLine := mutedStyle.Render(t.Artist)

		metaLine := lipgloss.NewStyle().Width(listW).Render(
			trackLine + "  " + when,
		)

		rows = append(rows, metaLine)
		rows = append(rows, artistLine)
		rows = append(rows, "")
	}

	rows = append(rows, dimStyle.Render(strings.Repeat("─", listW)))
	rows = append(rows, dimStyle.Render("scrobbled via Last.fm"))

	block := lipgloss.NewStyle().Width(listW).Render(strings.Join(rows, "\n"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top,
		lipgloss.NewStyle().PaddingTop(2).Render(block))
}
