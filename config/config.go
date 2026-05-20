package config

import (
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

const (
	AppName = "thn"
)

func InitLogger() (io.Closer, error) {
	dir, err := xdg.StateFile(AppName)
	if err != nil {
		return nil, err
	}

	if _, err = os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		if err = os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}

	filePath := filepath.Join(dir, AppName+".log")
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
