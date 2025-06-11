package log

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

func NewLogger(enabled bool, level string) *slog.Logger {
	if !enabled {
		return slog.New(slog.DiscardHandler)
	}

	slogLevel := levelToSlogLevel(level)
	handler := newMuxHandler(slogLevel)

	return slog.New(handler)
}

type muxHandler struct {
	infoHandler slog.Handler
	errHandler  slog.Handler
	baseLevel   slog.Level
}

func newMuxHandler(level slog.Level) *muxHandler {
	outLevel := new(slog.LevelVar)
	outLevel.Set(level)

	errLevel := new(slog.LevelVar)
	errLevel.Set(max(level, slog.LevelWarn))

	outHandler := tint.NewHandler(colorable.NewColorable(os.Stdout), &tint.Options{
		NoColor:    !isatty.IsTerminal(os.Stdout.Fd()),
		Level:      outLevel,
		TimeFormat: time.DateTime,
	})

	errHandler := tint.NewHandler(colorable.NewColorable(os.Stderr), &tint.Options{
		NoColor:    !isatty.IsTerminal(os.Stderr.Fd()),
		Level:      errLevel,
		TimeFormat: time.DateTime,
	})

	return &muxHandler{outHandler, errHandler, level}
}

func (h *muxHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if level < h.baseLevel {
		return false
	}
	if level >= slog.LevelWarn {
		return h.errHandler.Enabled(ctx, level)
	}
	return h.infoHandler.Enabled(ctx, level)
}

func (h *muxHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= slog.LevelWarn {
		return h.errHandler.Handle(ctx, r)
	}
	return h.infoHandler.Handle(ctx, r)
}

func (h *muxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &muxHandler{
		infoHandler: h.infoHandler.WithAttrs(attrs),
		errHandler:  h.errHandler.WithAttrs(attrs),
		baseLevel:   h.baseLevel,
	}
}

func (h *muxHandler) WithGroup(name string) slog.Handler {
	return &muxHandler{
		infoHandler: h.infoHandler.WithGroup(name),
		errHandler:  h.errHandler.WithGroup(name),
		baseLevel:   h.baseLevel,
	}
}

func levelToSlogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
