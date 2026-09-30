package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
)

const host = "0.0.0.0"
const port = 2222

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			bm.Middleware(teaHandler),
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

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	m := sshModel{
		term: s.User(),
	}
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

type sshModel struct {
	term string
}

func (m sshModel) Init() tea.Cmd {
	return nil
}

func (m sshModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m sshModel) View() string {
	return fmt.Sprintf("Bienvenue sur le serveur SSH E-Commerce CLI !\n\nAppuyez sur 'q' pour vous déconnecter.\n")
}
