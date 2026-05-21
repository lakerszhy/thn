package config

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

const (
	AppName = "thn"
)

func InitLogger() (io.Closer, error) {
	filePath, err := xdg.StateFile(filepath.Join(AppName, AppName+".log"))
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(file, opts)

	slog.SetDefault(slog.New(handler))

	return file, nil
}
