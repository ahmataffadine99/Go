package main

import (
	"errors"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"ecommerce-cli/internal/app"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	lm "github.com/charmbracelet/wish/logging"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func sessionInput(s ssh.Session) (*os.File, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	go func() {
		io.Copy(w, s)
		w.Close()
	}()
	return r, nil
}

func main() {
	baseURL := getEnv("API_URL", "http://localhost:8080")
	addr := net.JoinHostPort("0.0.0.0", getEnv("SSH_PORT", "2222"))

	s, err := wish.NewServer(
		wish.WithAddress(addr),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			func(h ssh.Handler) ssh.Handler {
				return func(s ssh.Session) {
					in, err := sessionInput(s)
					if err != nil {
						wish.Fatalln(s, err)
						return
					}
					defer in.Close()

					if s.User() == "admin" {
						app.RunAdminApp(baseURL, in, s)
					} else {
						app.RunClientApp(baseURL, in, s)
					}
					h(s)
				}
			},
			activeterm.Middleware(),
			lm.Middleware(),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting SSH server on %s (ssh admin@... pour l'administration)", addr)
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Fatalln(err)
		}
	}()

	<-done
	log.Println("Stopping SSH server...")
	if err := s.Close(); err != nil {
		log.Fatalln(err)
	}
}
