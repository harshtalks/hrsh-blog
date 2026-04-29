package main

import (
	"context"
	"embed"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"
	"golang.org/x/time/rate"

	"github.com/harshtalks/hrsh-blog/internal/content"
	"github.com/harshtalks/hrsh-blog/internal/ui"
)

//go:embed src/content/blog
var postsFS embed.FS

const (
	host    = "0.0.0.0"
	port    = "2222"
	maxConn = 50
)

var (
	connCount int
	connMu    sync.Mutex
	limiters  = make(map[string]*rate.Limiter)
	limiterMu sync.Mutex
)

func getLimiter(ip string) *rate.Limiter {
	limiterMu.Lock()
	defer limiterMu.Unlock()
	if l, ok := limiters[ip]; ok {
		return l
	}
	l := rate.NewLimiter(rate.Every(time.Minute/5), 5)
	limiters[ip] = l
	return l
}

func rateLimitMiddleware(next ssh.Handler) ssh.Handler {
	return func(s ssh.Session) {
		host, _, _ := net.SplitHostPort(s.RemoteAddr().String())
		if !getLimiter(host).Allow() {
			fmt.Fprintln(s, "Rate limit exceeded. Please try again in a minute.")
			return
		}
		next(s)
	}
}

func maxConnMiddleware(next ssh.Handler) ssh.Handler {
	return func(s ssh.Session) {
		connMu.Lock()
		if connCount >= maxConn {
			connMu.Unlock()
			fmt.Fprintln(s, "Server is at capacity. Please try again shortly.")
			return
		}
		connCount++
		connMu.Unlock()
		defer func() {
			connMu.Lock()
			connCount--
			connMu.Unlock()
		}()
		next(s)
	}
}

func main() {
	// The server process has no TTY (especially in Docker), so termenv would
	// default to Ascii and strip all ANSI colors. Force TrueColor globally.
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(os.Stdout, termenv.WithProfile(termenv.TrueColor)))

	posts, err := content.LoadPosts(postsFS, "src/content/blog")
	if err != nil {
		log.Error("Failed to load posts", "error", err)
		os.Exit(1)
	}
	log.Info("Loaded posts", "count", len(posts))

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			rateLimitMiddleware,
			maxConnMiddleware,
			bm.Middleware(func(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
				pty, _, active := sess.Pty()
				if !active {
					fmt.Fprintln(sess, "This app requires an interactive terminal.")
					return nil, nil
				}
				return ui.New(posts, pty.Window.Width, pty.Window.Height), []tea.ProgramOption{
					tea.WithAltScreen(),
				}
			}),
			lm.Middleware(),
		),
	)
	if err != nil {
		log.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("SSH server started", "host", host, "port", port)

	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Error("Server stopped", "error", err)
		}
	}()

	<-done
	log.Info("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Error("Shutdown error", "error", err)
	}
}
