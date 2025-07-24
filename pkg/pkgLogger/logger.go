package pkgLogger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
)

// Interface -.
type Interface interface {
	Debug(message interface{}, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
	Fatal(message interface{}, args ...interface{})
}

// Logger -.
type Logger struct {
	logger *slog.Logger
}

var _ Interface = (*Logger)(nil)

// New создает новый экземпляр логгера.
func New(env string) *Logger {
	var log *slog.Logger

	switch strings.ToLower(env) {
	case "local":
		log = setupPrettySlog()
	case "dev":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "prod":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return &Logger{
		logger: log,
	}
}

// Debug -.
func (l *Logger) Debug(message interface{}, args ...interface{}) {
	l.msg(slog.LevelDebug, message, args...)
}

// Info -.
func (l *Logger) Info(message string, args ...interface{}) {
	l.log(slog.LevelInfo, message, args...)
}

// Warn -.
func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(slog.LevelWarn, message, args...)
}

// Error -.
func (l *Logger) Error(message interface{}, args ...interface{}) {
	l.msg(slog.LevelError, message, args...)
}

// Fatal -.
func (l *Logger) Fatal(message interface{}, args ...interface{}) {
	l.msg(slog.LevelError, message, args...)
	os.Exit(1)
}

func (l *Logger) log(level slog.Level, message string, args ...interface{}) {
	if !l.logger.Enabled(context.TODO(), level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip [Callers, log, Interface.Method]

	r := slog.NewRecord(time.Now(), level, message, pcs[0])
	if len(args) > 0 {
		r.Add(args...)
	}
	_ = l.logger.Handler().Handle(context.TODO(), r)
}

func (l *Logger) msg(level slog.Level, message interface{}, args ...interface{}) {
	if !l.logger.Enabled(context.TODO(), level) {
		return
	}

	var msg string
	switch m := message.(type) {
	case error:
		msg = m.Error()
	case string:
		msg = m
	default:
		msg = fmt.Sprintf("%s message %v has unknown type %T", level, message, message)
	}

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip [Callers, msg, Interface.Method]

	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	if len(args) > 0 {
		r.Add(args...)
	}
	_ = l.logger.Handler().Handle(context.TODO(), r)
}
