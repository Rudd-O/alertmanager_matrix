package slogctx

import (
	"context"
	"log/slog"
)

var _ slog.Handler = (*Handler)(nil)

// Handler is a [slog.Handler] that provides context when available.
// It only adds context for the slog calls with context, such as [slog.InfoContext].
type Handler struct {
	h slog.Handler
}

// Default returns a Handler that wraps the handler used by [slog.Default].
func Default() Handler {
	return NewHandler(slog.Default().Handler())
}

// NewHandler returns an initialised Handler for the given [slog.Handler].
func NewHandler(handler slog.Handler) Handler {
	return Handler{h: handler}
}

// Enabled returns true if the handler logs at the given level.
// See [slog.Handler] for details.
func (h Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

// Handle handles the given Record.
// See [slog.Handler] for details.
func (h Handler) Handle(ctx context.Context, record slog.Record) error {
	From(ctx).AddToRecord(&record)

	return h.h.Handle(ctx, record) //nolint:wrapcheck // no additional context to provide here
}

// WithAttrs returns a new Handler with the given attributes set.
// See [slog.Handler] for details.
func (h Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewHandler(h.h.WithAttrs(attrs))
}

// WithGroup returns a new Handler with the given name appended to existing fields.
// See [slog.Handler] for details.
func (h Handler) WithGroup(name string) slog.Handler {
	return NewHandler(h.h.WithGroup(name))
}
