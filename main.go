package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/lakerszhy/thn/config"
	"github.com/lakerszhy/thn/hn"
	"github.com/lakerszhy/thn/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	logFile, err := config.InitLogger()
	if err != nil {
		return err
	}
	defer logFile.Close()

	slog.Info("THN starting up....")

	hotkeys, err := config.LoadHotkeys()
	if err != nil {
		return err
	}

	client, err := hn.New(context.Background())
	if err != nil {
		return err
	}

	app := ui.NewApp(client, config.HackerNewsTheme, hotkeys)
	if _, err = tea.NewProgram(app).Run(); err != nil {
		return err
	}
	return nil
}
