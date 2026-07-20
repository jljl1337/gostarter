package log

import (
	"fmt"
	"log/slog"
)

// Debugf logs a debug message with formatting.
func Debugf(format string, args ...any) {
	Debug(fmt.Sprintf(format, args...))
}

// Debug logs a debug message.
func Debug(msg string) {
	slog.Debug(msg)
}

// Infof logs an info message with formatting.
func Infof(format string, args ...any) {
	Info(fmt.Sprintf(format, args...))
}

// Info logs an info message.
func Info(msg string) {
	slog.Info(msg)
}

// Warnf logs a warning message with formatting.
func Warnf(format string, args ...any) {
	Warn(fmt.Sprintf(format, args...))
}

// Warn logs a warning message.
func Warn(msg string) {
	slog.Warn(msg)
}

// Errorf logs an error message with formatting.
func Errorf(format string, args ...any) {
	Error(fmt.Sprintf(format, args...))
}

// Error logs an error message.
func Error(msg string) {
	slog.Error(msg)
}
