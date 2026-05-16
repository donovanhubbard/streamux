package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		fmt.Errorf("invalid log level: %s", level)
		os.Exit(1)
	}
	return slog.LevelDebug
}

func setupOutput(cfg *Config) *lumberjack.Logger {
	logRotator := &lumberjack.Logger{
		Filename:   cfg.LogFile,
		MaxSize:    50,   // Max size in MB
		MaxBackups: 3,    // Number of backups
		MaxAge:     30,   // Days
		Compress:   true, // Enable compression
	}

	logLevel := parseLogLevel(cfg.LogLevel)

	opts := &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Check if the current attribute is the 'time' key
			if a.Key == slog.TimeKey {
				// Get the time value and format it using a 12-hour layout
				// "03:04:05PM" is the Go reference for 12-hour format with AM/PM
				t := a.Value.Time()
				return slog.String(slog.TimeKey, t.Format("2006-01-02 03:04:05PM -7:00"))
			}
			return a
		},
	}

	var handler slog.Handler
	if cfg.Stdout && cfg.LogFile != "" {
		multiWriter := io.MultiWriter(os.Stdout, logRotator)
		handler = slog.NewTextHandler(multiWriter, opts)
	} else if cfg.LogFile != "" {
		handler = slog.NewTextHandler(logRotator, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logRotator
}
