package logger

import (
	"errors"
	"fmt"
	"os"

	"github.com/Maxi-Mega/s3-image-server-v2/config"

	"github.com/rs/zerolog"
)

const logTimeFormat = "2006-01-02T15:04:05.000Z"

var errInvalidConfig = errors.New("invalid configuration")

var logger zerolog.Logger //nolint:gochecknoglobals

func Init(cfg config.Log) error {
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("%w: %w", errInvalidConfig, err)
	}

	zerolog.TimeFieldFormat = logTimeFormat

	if cfg.JSONLogFormat {
		logCtx := zerolog.New(os.Stdout).With().Timestamp()

		for field, value := range cfg.JSONLogFields {
			logCtx = logCtx.Interface(field, value)
		}

		logger = logCtx.Logger()
	} else {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			NoColor:    !cfg.ColorLogs,
			TimeFormat: logTimeFormat,
		}

		logger = zerolog.New(consoleWriter).Level(level).With().Timestamp().Logger()
	}

	Trace("Logger is initialized")

	return nil
}

func Trace(a ...any) {
	logger.Trace().Msg(fmt.Sprint(a...))
}

func Tracef(format string, a ...any) {
	logger.Trace().Msg(fmt.Sprintf(format, a...))
}

func Debug(a ...any) {
	logger.Debug().Msg(fmt.Sprint(a...))
}

func Debugf(format string, a ...any) {
	logger.Debug().Msg(fmt.Sprintf(format, a...))
}

func Info(a ...any) {
	logger.Info().Msg(fmt.Sprint(a...))
}

func Infof(format string, a ...any) {
	logger.Info().Msg(fmt.Sprintf(format, a...))
}

func Warn(a ...any) {
	logger.Warn().Msg(fmt.Sprint(a...))
}

func Warnf(format string, a ...any) {
	logger.Warn().Msg(fmt.Sprintf(format, a...))
}

func Error(a ...any) {
	logger.Error().Msg(fmt.Sprint(a...))
}

func Errorf(format string, a ...any) {
	logger.Error().Msg(fmt.Sprintf(format, a...))
}

func Fatal(a ...any) {
	logger.Fatal().Msg(fmt.Sprint(a...))
}

func Fatalf(format string, a ...any) {
	logger.Fatal().Msg(fmt.Sprintf(format, a...))
}
