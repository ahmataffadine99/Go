package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecommerce-cli/internal/app"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	lm "github.com/charmbracelet/wish/logging"
)

const host = "0.0.0.0"
const port = 2222

func main() {
	baseURL := "http://localhost:8080"
	
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			func(h ssh.Handler) ssh.Handler {
				return func(s ssh.Session) {
					// We pass the SSH session (which acts as io.Reader and io.Writer) to our app!
					app.RunClientApp(baseURL, s, s)
					h(s)
				}
			},
			lm.Middleware(),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting SSH server on %s:%d", host, port)
	go func() {
		if err = s.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	<-done
	log.Println("Stopping SSH server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Close(); err != nil {
		log.Fatalln(err)
	}
	_ = ctx
}
