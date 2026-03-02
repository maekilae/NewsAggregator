package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
)

// MultiHandler sends records to multiple handlers
type MultiHandler struct {
	handlers []slog.Handler
}

func (m *MultiHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, l) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			h.Handle(ctx, r)
		}
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}

func Init(logFile string) {
	f, _ := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	// 1. Console Handler (Standard Text)
	h1 := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo, AddSource: false})

	// 2. File Handler (JSON for large scale ingestion)
	h2 := slog.NewJSONHandler(f, &slog.HandlerOptions{Level: slog.LevelInfo, AddSource: true})

	combined := &MultiHandler{handlers: []slog.Handler{h1, h2}}
	slog.SetDefault(slog.New(combined))
}

// Err adds a stack trace and error to the log
func Err(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}
	trace := make([]byte, 1024)
	n := runtime.Stack(trace, false)
	return slog.Group("exception",
		slog.String("message", err.Error()),
		slog.String("stacktrace", string(trace[:n])),
	)
}
