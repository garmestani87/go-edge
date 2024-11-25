package logging

import (
	"edge-app/configs"
	"edge-app/pkg/constant"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var (
	once    sync.Once
	zLogger *zerolog.Logger
)

type zeroLogger struct {
	cfg    *configs.Config
	logger *zerolog.Logger
}

type URLFilterHook struct {
	ignoredURLs []string
}

func (u URLFilterHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	for _, url := range u.ignoredURLs {
		if msg == url {
			// Skip logging this URL
			e.Discard()
			return
		}
	}
}

var zeroLogLevelMapping = map[string]zerolog.Level{
	"debug": zerolog.DebugLevel,
	"info":  zerolog.InfoLevel,
	"warn":  zerolog.WarnLevel,
	"error": zerolog.ErrorLevel,
	"fatal": zerolog.FatalLevel,
}

func newZeroLogger(cfg *configs.Config) *zeroLogger {
	logger := &zeroLogger{cfg: cfg}
	if cfg.Logging.Console {
		logger.InitConsoleLogger()
	} else {
		logger.InitFileLogger(cfg)
	}
	return logger
}

func logParamsToZeroParams(keys map[ExtraKey]interface{}) map[string]interface{} {
	params := map[string]interface{}{}

	for k, v := range keys {
		params[string(k)] = v
	}

	return params
}

func (l *zeroLogger) getLogLevel() zerolog.Level {
	level, exists := zeroLogLevelMapping[l.cfg.Logging.Level]
	if !exists {
		return zerolog.DebugLevel
	}
	return level
}

func (l *zeroLogger) InitConsoleLogger() {
	once.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC1123}).
			With().
			Timestamp().
			Str("AppName", l.cfg.Application.Name).
			Str("LoggerName", "Zerolog").
			Logger().
			Hook(URLFilterHook{ignoredURLs: []string{constant.Metrics}})

		zerolog.SetGlobalLevel(l.getLogLevel())
		zLogger = &logger
	})
	l.logger = zLogger
}

func (l *zeroLogger) InitFileLogger(cfg *configs.Config) {
	once.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

		fileName := fmt.Sprintf("%s%s", cfg.Logging.FilePath, cfg.Logging.FileName)

		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o666)
		if err != nil {
			panic("could not open log file")
		}

		logger := zerolog.New(file).
			With().
			Timestamp().
			Str("AppName", l.cfg.Application.Name).
			Str("LoggerName", "Zerolog").
			Logger().
			Hook(URLFilterHook{ignoredURLs: []string{constant.Metrics}})

		zerolog.SetGlobalLevel(l.getLogLevel())
		zLogger = &logger
	})
	l.logger = zLogger
}

func (l *zeroLogger) Debug(cat Category, sub SubCategory, msg string, extra map[ExtraKey]interface{}) {
	l.logger.
		Debug().
		Str("Category", string(cat)).
		Str("SubCategory", string(sub)).
		Fields(logParamsToZeroParams(extra)).
		Msg(msg)
}

func (l *zeroLogger) Debugf(template string, args ...interface{}) {
	l.logger.
		Debug().
		Msgf(template, args...)
}

func (l *zeroLogger) Info(cat Category, sub SubCategory, msg string, extra map[ExtraKey]interface{}) {
	l.logger.
		Info().
		Str("Category", string(cat)).
		Str("SubCategory", string(sub)).
		Fields(logParamsToZeroParams(extra)).
		Msg(msg)
}

func (l *zeroLogger) Infof(template string, args ...interface{}) {
	l.logger.
		Info().
		Msgf(template, args...)
}

func (l *zeroLogger) Warn(cat Category, sub SubCategory, msg string, extra map[ExtraKey]interface{}) {
	l.logger.
		Warn().
		Str("Category", string(cat)).
		Str("SubCategory", string(sub)).
		Fields(logParamsToZeroParams(extra)).
		Msg(msg)
}

func (l *zeroLogger) Warnf(template string, args ...interface{}) {
	l.logger.
		Warn().
		Msgf(template, args...)
}

func (l *zeroLogger) Error(cat Category, sub SubCategory, msg string, extra map[ExtraKey]interface{}) {
	l.logger.
		Error().
		Str("Category", string(cat)).
		Str("SubCategory", string(sub)).
		Fields(logParamsToZeroParams(extra)).
		Msg(msg)
}

func (l *zeroLogger) Errorf(template string, args ...interface{}) {
	l.logger.
		Error().
		Msgf(template, args...)
}

func (l *zeroLogger) Fatal(cat Category, sub SubCategory, msg string, extra map[ExtraKey]interface{}) {
	l.logger.
		Fatal().
		Str("Category", string(cat)).
		Str("SubCategory", string(sub)).
		Fields(logParamsToZeroParams(extra)).
		Msg(msg)
}

func (l *zeroLogger) Fatalf(template string, args ...interface{}) {
	l.logger.
		Fatal().
		Msgf(template, args...)
}
