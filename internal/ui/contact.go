package ui

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harshtalks/hrsh-blog/internal/config"
)

type contactState int

const (
	contactStateForm contactState = iota
	contactStateSubmitting
	contactStateSuccess
	contactStateError
)

type submitOkMsg struct{}
type submitErrMsg struct{ err error }

type ContactModel struct {
	state   contactState
	name    textinput.Model
	email   textinput.Model
	message textarea.Model
	focused int
	errMsg  string
	width   int
	height  int
}

func newContact(width, height int) ContactModel {
	name := textinput.New()
	name.Placeholder = "Your name"
	name.CharLimit = 100
	name.Width = 40
	name.Focus()

	email := textinput.New()
	email.Placeholder = "your@email.com"
	email.CharLimit = 100
	email.Width = 40

	msg := textarea.New()
	msg.Placeholder = "What's on your mind?"
	msg.SetWidth(60)
	msg.SetHeight(5)
	msg.CharLimit = 1000

	return ContactModel{
		state:   contactStateForm,
		name:    name,
		email:   email,
		message: msg,
		focused: 0,
		width:   width,
		height:  height,
	}
}

func (c ContactModel) resize(width, height int) ContactModel {
	c.width = width
	c.height = height
	return c
}

func (c ContactModel) hasFocusedInput() bool {
	return c.state == contactStateForm
}

func (c ContactModel) handleMsg(msg tea.Msg) (ContactModel, tea.Cmd) {
	if c.state == contactStateSuccess || c.state == contactStateError {
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "r" {
			c = newContact(c.width, c.height)
		}
		return c, nil
	}

	if c.state == contactStateSubmitting {
		switch msg := msg.(type) {
		case submitOkMsg:
			c.state = contactStateSuccess
		case submitErrMsg:
			c.state = contactStateError
			c.errMsg = msg.err.Error()
		}
		return c, nil
	}

	switch msg := msg.(type) {
	case submitOkMsg:
		c.state = contactStateSuccess
		return c, nil
	case submitErrMsg:
		c.state = contactStateError
		c.errMsg = msg.err.Error()
		return c, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			c.focused = (c.focused + 1) % 4
			c.syncFocus()
			return c, nil
		case "shift+tab", "up":
			c.focused = (c.focused + 3) % 4
			c.syncFocus()
			return c, nil
		case "enter":
			if c.focused == 3 {
				return c.submit()
			}
			if c.focused == 2 {
				var cmd tea.Cmd
				c.message, cmd = c.message.Update(msg)
				return c, cmd
			}
		}
	}

	var cmd tea.Cmd
	switch c.focused {
	case 0:
		c.name, cmd = c.name.Update(msg)
	case 1:
		c.email, cmd = c.email.Update(msg)
	case 2:
		c.message, cmd = c.message.Update(msg)
	}
	return c, cmd
}

func (c *ContactModel) syncFocus() {
	c.name.Blur()
	c.email.Blur()
	c.message.Blur()
	switch c.focused {
	case 0:
		c.name.Focus()
	case 1:
		c.email.Focus()
	case 2:
		c.message.Focus()
	}
}

func (c ContactModel) submit() (ContactModel, tea.Cmd) {
	name := strings.TrimSpace(c.name.Value())
	email := strings.TrimSpace(c.email.Value())
	message := strings.TrimSpace(c.message.Value())

	if name == "" || email == "" || message == "" {
		c.errMsg = "All fields are required."
		c.state = contactStateError
		return c, nil
	}

	c.state = contactStateSubmitting
	return c, func() tea.Msg {
		formData := url.Values{
			config.GoogleFormEntryName:    {name},
			config.GoogleFormEntryEmail:   {email},
			config.GoogleFormEntryMessage: {message},
		}
		resp, err := http.PostForm(config.GoogleFormURL, formData)
		if err != nil {
			return submitErrMsg{err}
		}
		resp.Body.Close()
		return submitOkMsg{}
	}
}

func (c ContactModel) View() string {
	formW := 60

	switch c.state {
	case contactStateSuccess:
		block := lipgloss.JoinVertical(lipgloss.Center,
			successStyle.Render("✓  Message sent!"),
			"",
			mutedStyle.Render("Thanks for reaching out. I'll get back to you soon."),
			"",
			dimStyle.Render("[r] send another"),
		)
		return lipgloss.Place(c.width, c.height, lipgloss.Center, lipgloss.Center, block)

	case contactStateError:
		block := lipgloss.JoinVertical(lipgloss.Center,
			errorStyle.Render("✗  Something went wrong"),
			"",
			mutedStyle.Render(c.errMsg),
			"",
			dimStyle.Render("[r] try again"),
		)
		return lipgloss.Place(c.width, c.height, lipgloss.Center, lipgloss.Center, block)

	case contactStateSubmitting:
		return lipgloss.Place(c.width, c.height, lipgloss.Center, lipgloss.Center,
			mutedStyle.Render("Sending..."))
	}

	label := func(s string, active bool) string {
		if active {
			return accentStyle.Render(s)
		}
		return inputLabelStyle.Render(s)
	}

	submitLabel := "  Send message  "
	submitBtn := func() string {
		if c.focused == 3 {
			return lipgloss.NewStyle().
				Foreground(colorBg).
				Background(colorAccent).
				Bold(true).
				Padding(0, 1).
				Render(submitLabel)
		}
		return lipgloss.NewStyle().
			Foreground(colorAccent).
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorDim).
			Padding(0, 1).
			Render(submitLabel)
	}()

	divider := dimStyle.Render(strings.Repeat("─", formW))

	form := lipgloss.JoinVertical(lipgloss.Left,
		accentStyle.Render("Get in touch"),
		mutedStyle.Render("I usually reply within a day or two."),
		divider,
		"",
		"",
		label("Name", c.focused == 0),
		c.name.View(),
		"",
		"",
		label("Email", c.focused == 1),
		c.email.View(),
		"",
		"",
		label("Message", c.focused == 2),
		c.message.View(),
		"",
		"",
		submitBtn,
		"",
		divider,
		dimStyle.Render("[tab] next field   [enter] submit   [ctrl+c] quit"),
	)

	block := lipgloss.NewStyle().Width(formW).Render(form)
	return lipgloss.Place(c.width, c.height, lipgloss.Center, lipgloss.Top,
		lipgloss.NewStyle().PaddingTop(2).Render(block))
}
