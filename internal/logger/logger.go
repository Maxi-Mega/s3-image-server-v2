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

	return nil
}

func Trace(a ...any) {
	logger.Trace().Msg(fmt.Sprint(a...))
}

func Debug(a ...any) {
	logger.Debug().Msg(fmt.Sprint(a...))
}

func Info(a ...any) {
	logger.Info().Msg(fmt.Sprint(a...))
}

func Warn(a ...any) {
	logger.Warn().Msg(fmt.Sprint(a...))
}

func Error(a ...any) {
	logger.Error().Msg(fmt.Sprint(a...))
}

func Fatal(a ...any) {
	logger.Fatal().Msg(fmt.Sprint(a...))
}
