package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

func New(level string, logFile string) (*slog.Logger, *os.File, error) {
	logLevel := slog.LevelInfo

	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	var writer io.Writer = os.Stdout
	var file *os.File

	if logFile != "" {
		var err error

		file, err = os.OpenFile(
			logFile,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0644,
		)
		if err != nil {
			return nil, nil, err
		}

		writer = io.MultiWriter(os.Stdout, file)
	}

	log := slog.New(
		slog.NewTextHandler(writer, &slog.HandlerOptions{
			Level: logLevel,
		}),
	)

	return log, file, nil
}
