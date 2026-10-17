package main

import (
	"errors"
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

func main() {
	baseURL := getEnv("API_URL", "http://localhost:8080")
	addr := net.JoinHostPort("0.0.0.0", getEnv("SSH_PORT", "2222"))

	s, err := wish.NewServer(
		wish.WithAddress(addr),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			func(h ssh.Handler) ssh.Handler {
				return func(s ssh.Session) {
					if s.User() == "admin" {
						app.RunAdminApp(baseURL, s, s)
					} else {
						app.RunClientApp(baseURL, s, s)
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
