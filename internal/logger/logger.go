package logger

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/rs/zerolog"
)

const logTimeFormat = "2006-01-02T15:04:05.000Z"

var errInvalidConfig = errors.New("invalid configuration")

var logger = defaultLogger() //nolint:gochecknoglobals

func Init(logLevel string, colorLogs, jsonLogFormat bool, jsonLogFields map[string]any) error {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("%w: %w", errInvalidConfig, err)
	}

	zerolog.TimeFieldFormat = logTimeFormat

	if jsonLogFormat {
		logCtx := zerolog.New(os.Stdout).With().Timestamp()

		for field, value := range jsonLogFields {
			logCtx = logCtx.Any(field, value)
		}

		logger = logCtx.Logger()
	} else {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			NoColor:    !colorLogs,
			TimeFormat: logTimeFormat,
		}

		logger = zerolog.New(consoleWriter).Level(level).With().Timestamp().Logger()
	}

	Trace("Logger is initialized")

	return nil
}

func defaultLogger() zerolog.Logger {
	if testing.Testing() {
		return zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.TimeFormat = logTimeFormat
		})).With().Timestamp().Logger()
	}

	return zerolog.Logger{}
}

func Trace(a ...any) {
	logger.Trace().Msg(fmt.Sprint(a...))
}

func Tracef(format string, a ...any) {
	logger.Trace().Msgf(format, a...)
}

func Debug(a ...any) {
	logger.Debug().Msg(fmt.Sprint(a...))
}

func Debugf(format string, a ...any) {
	logger.Debug().Msgf(format, a...)
}

func Info(a ...any) {
	logger.Info().Msg(fmt.Sprint(a...))
}

func Infof(format string, a ...any) {
	logger.Info().Msgf(format, a...)
}

func Warn(a ...any) {
	logger.Warn().Msg(fmt.Sprint(a...))
}

func Warnf(format string, a ...any) {
	logger.Warn().Msgf(format, a...)
}

func Error(a ...any) {
	logger.Error().Msg(fmt.Sprint(a...))
}

func Errorf(format string, a ...any) {
	logger.Error().Msgf(format, a...)
}

func Fatal(a ...any) {
	logger.Fatal().Msg(fmt.Sprint(a...))
}

func Fatalf(format string, a ...any) {
	logger.Fatal().Msgf(format, a...)
}
