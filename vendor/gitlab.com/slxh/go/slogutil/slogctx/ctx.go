package slogctx

import (
	"context"
	"log/slog"
	"sync"
)

// Add adds the given log arguments in the given context.
// It returns the original context if the context already includes logging information,
// or a new one if it does not.
func Add(ctx context.Context, args ...any) context.Context {
	if e := From(ctx); e != nil {
		e.Add(args...)
		return ctx
	}

	return NewContext(args...).Store(ctx)
}

// With returns a new context with the given log parameters.
// The log information in the context is extended with the given information.
func With(ctx context.Context, args ...any) context.Context {
	return From(ctx).CopyWith(args...).Store(ctx)
}

type contextKey struct{}

// Context is used for storing logging arguments in context.
// It is thread-safe, and may be used to share metadata between multiple threads.
type Context struct {
	args []any
	mu   sync.Mutex
}

// From returns Context from [context.Context] if set, nil otherwise.
func From(ctx context.Context) *Context {
	if v, ok := ctx.Value(contextKey{}).(*Context); ok {
		return v
	}

	return nil
}

// NewContext returns an initialised Context that holds the given arguments.
func NewContext(args ...any) *Context {
	return &Context{args: args}
}

// AddToRecord adds the arguments contained in the context to the given [slog.Record].
// It does nothing on a nil Context.
func (c *Context) AddToRecord(r *slog.Record) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	r.Add(c.args...)
}

// Add adds the given arguments to the Context.
// It panics when called on a nil Context.
func (c *Context) Add(args ...any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.args = append(c.args, args...)
}

// Args returns a copy of the arguments stored in the Context.
// It returns nil for a nil context.
func (c *Context) Args() []any {
	if c == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return c.args
}

// CopyWith returns a copy of the current Context with the given values added.
// It returns a new Context when called on a nil Context.
func (c *Context) CopyWith(args ...any) *Context {
	if c == nil {
		return NewContext(args...)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	next := make([]any, len(args)+len(c.args))
	copy(next, c.args)
	copy(next[len(c.args):], args)

	return &Context{args: next}
}

// Store stores the Context in the given [context.Context].
// It overwrites any existing value.
func (c *Context) Store(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey{}, c)
}
