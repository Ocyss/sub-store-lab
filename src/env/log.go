package env

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
)

var (
	logger *slog.Logger
	level  slog.Level
)

func initLog() {
	defaultLevel := slog.LevelInfo
	if Conf.Debug {
		defaultLevel = slog.LevelDebug
	}
	level = parseLever(Conf.LogLevel, defaultLevel)
	logger = slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			AddSource:  true,
			Level:      level,
			TimeFormat: time.DateTime,
		}),
	)
	slog.SetDefault(logger)
}

func parseLever(str string, def slog.Level) slog.Level {
	switch strings.ToLower(str) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return def
	}
}

func GetLogger() *slog.Logger {
	return logger
}
