package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jljl1337/gostarter/env"
)

/*
SetCustomLoggerFromEnv sets a custom logger based on the log level defined in
the environment variables. It uses the slog package to create a new logger with
a custom handler that formats log messages with a timestamp, log level, and
message. The log level is set based on the value of env.LogLevel.
*/
func SetCustomLoggerFromEnv() error {
	if !env.ConstantsSet {
		return fmt.Errorf("environment variables not set, cannot set custom logger")
	}

	SetCustomLogger(env.LogLevel)

	return nil
}

/*
SetCustomLogger sets a custom logger with the specified log level. It uses the
slog package to create a new logger with a custom handler that formats log
messages with a timestamp, log level, and message. The log level is set based
on the provided logLevel parameter.
*/
func SetCustomLogger(logLevel int) {
	slog.SetDefault(newCustomLogger(logLevel))
}

// newCustomLogger creates a new custom logger with the specified log level.
func newCustomLogger(logLevel int) *slog.Logger {
	return slog.New(&customHandler{
		level: slog.Level(logLevel),
	})
}

/*
A customHandler is a custom implementation of slog.Handler that formats log
messages with a timestamp, log level, and message.
*/
type customHandler struct {
	level slog.Leveler
}

func (h *customHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *customHandler) Handle(ctx context.Context, r slog.Record) error {
	// Format timestamp like default log
	ts := time.Now().Format("2006-01-02 15:04:05.000")

	lvl := r.Level.String()
	switch lvl {
	case "DEBUG":
		lvl = "DBG"
	case "INFO":
		lvl = "INF"
	case "WARN":
		lvl = "WRN"
	case "ERROR":
		lvl = "ERR"
	}

	msg := r.Message
	formatted := fmt.Sprintf("%s %s %s\n", ts, lvl, msg)

	// Write using standard log or directly to stderr
	_, err := fmt.Fprint(os.Stdout, formatted)
	return err
}

func (h *customHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *customHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *customHandler) Close() error { return nil }

func (h *customHandler) Flush() error { return nil }
