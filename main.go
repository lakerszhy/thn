package main

import (
	"context"
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/lakerszhy/thn/config"
	"github.com/lakerszhy/thn/hn"
	"github.com/lakerszhy/thn/ui"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	client, err := hn.New(context.Background())
	if err != nil {
		return err
	}

	app := ui.NewApp(client, config.HackerNewsTheme, config.Hotkeys)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		return err
	}
	return nil
}
