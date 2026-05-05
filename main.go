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

	"github.com/charmbracelet/keygen"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	gossh "golang.org/x/crypto/ssh"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
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

func hostKeyOption() ssh.Option {
	// Production: load raw PEM from env var (e.g., Railway/Fly.io secret)
	if pem := os.Getenv("SSH_HOST_KEY"); pem != "" {
		log.Info("Using SSH host key from SSH_HOST_KEY env var")
		return wish.WithHostKeyPEM([]byte(pem))
	}

	// Local dev: load from file, auto-generate if missing
	path := ".ssh/term_info_ed25519"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Info("Host key not found, generating new one", "path", path)
		if _, err := keygen.New(path, keygen.WithKeyType(keygen.Ed25519), keygen.WithWrite()); err != nil {
			log.Fatal("Failed to generate host key", "error", err)
		}
	}
	log.Info("Using SSH host key from file", "path", path)
	return wish.WithHostKeyPath(path)
}

func main() {
	posts, err := content.LoadPosts(postsFS, "src/content/blog")
	if err != nil {
		log.Error("Failed to load posts", "error", err)
		os.Exit(1)
	}
	log.Info("Loaded posts", "count", len(posts))

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		func(srv *ssh.Server) error {
			// Accept ALL connections without authentication. We do this by
			// setting NoClientAuth (handles "none" method) AND permissive
			// callbacks for password/keyboard-interactive. OpenSSH may send
			// "publickey" first even when the user has no keys; without a
			// PublicKeyCallback the server advertises password/keyboard-
			// interactive instead, which we accept silently.
			log.Info("Enabling permissive SSH auth: NoClientAuth + password + keyboard-interactive")
			srv.ServerConfigCallback = func(ctx ssh.Context) *gossh.ServerConfig {
				return &gossh.ServerConfig{
					NoClientAuth: true,
					PasswordCallback: func(conn gossh.ConnMetadata, password []byte) (*gossh.Permissions, error) {
						return &gossh.Permissions{}, nil
					},
					KeyboardInteractiveCallback: func(conn gossh.ConnMetadata, client gossh.KeyboardInteractiveChallenge) (*gossh.Permissions, error) {
						return &gossh.Permissions{}, nil
					},
				}
			}
			return nil
		},
		hostKeyOption(),
		wish.WithMiddleware(
			rateLimitMiddleware,
			maxConnMiddleware,
			bm.Middleware(func(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
				pty, _, active := sess.Pty()
				if !active {
					fmt.Fprintln(sess, "This app requires an interactive terminal.")
					return nil, nil
				}
				renderer := bm.MakeRenderer(sess)
				caps := ui.DetectCapabilities(sess, renderer)
				return ui.New(posts, pty.Window.Width, pty.Window.Height, renderer, caps), []tea.ProgramOption{
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
