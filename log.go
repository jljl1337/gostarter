package gostarter

import (
	"fmt"

	"github.com/jljl1337/gostarter/pkg/shared/env"
	"github.com/jljl1337/gostarter/pkg/shared/log"
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

	log.SetCustomLogger(env.LogLevel)

	return nil
}

func Debugf(format string, args ...any) {
	log.Debugf(format, args...)
}

func Debug(msg string) {
	log.Debug(msg)
}

func Infof(format string, args ...any) {
	log.Infof(format, args...)
}

func Info(msg string) {
	log.Info(msg)
}

func Warnf(format string, args ...any) {
	log.Warnf(format, args...)
}

func Warn(msg string) {
	log.Warn(msg)
}

func Errorf(format string, args ...any) {
	log.Errorf(format, args...)
}

func Error(msg string) {
	log.Error(msg)
}
