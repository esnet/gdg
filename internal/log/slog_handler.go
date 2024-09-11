package log

import (
	"context"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

type ContextHandler struct {
	handler      slog.Handler
	errorHandler slog.Handler
	outStream    *os.File
	errStream    *os.File
	options      *tint.Options
}

func NewContextHandler(h slog.Handler, out *os.File, err *os.File, opts *tint.Options) *ContextHandler {
	ch := &ContextHandler{outStream: out, errStream: err, options: opts}
	if lh, ok := h.(*ContextHandler); ok {
		if lh.outStream == out && lh.errStream == err && lh.options.Level == opts.Level {
			return lh
		}
	}

	errOpts := *opts
	errOpts.Level = slog.LevelWarn
	// create a error logger
	// set global logger with custom options
	errorHandler := tint.NewHandler(err, &errOpts)
	ch.errorHandler = errorHandler

	outHandler := tint.NewHandler(out, opts)
	ch.handler = outHandler
	return ch
}

func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if level >= slog.LevelWarn {
		return h.errorHandler.Enabled(ctx, level)
	}

	return h.handler.Enabled(ctx, level)
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= slog.LevelWarn {
		return h.errorHandler.Handle(ctx, r)
	}
	return h.handler.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewContextHandler(h.handler.WithAttrs(attrs), h.outStream, h.errStream, nil)
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return NewContextHandler(h.handler.WithGroup(name), h.outStream, h.errStream, nil)
}
