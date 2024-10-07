// Package log encapsulates structured logging using a singleton logger
// instance. Messages are emitted via the Debug, Info, Warn, Error functions.
// The logging output is set only once: either by the Init function or invoking
// a logging function, whichever happens first. By default, it will write to
// stderr in a text format. To control the format and/or the output sink, call
// the Init function with a handler. as early as possible.
package log

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sort"
	"sync"
)

var (
	theLogger  *slog.Logger
	initLogger sync.Once
)

// Init sets up a singleton logger just once. Subsequent invocations after the
// first invocation are pretty much no ops, If h is empty, then it effectively
// shuts off logging.
func Init(h slog.Handler) {
	initLogger.Do(func() {
		if h == nil {
			// this value should be greater than the highest value provided by the standard library
			level := slog.LevelError + 1
			h = slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: level})
		}

		theLogger = slog.New(h)
		theLogger.Debug("initialized logger")
	})
}

func Debug(ctx context.Context, fields map[string]any, msg string) {
	log(ctx, slog.LevelDebug, fields, nil, msg)
}

func Info(ctx context.Context, fields map[string]any, msg string) {
	log(ctx, slog.LevelInfo, fields, nil, msg)
}

func Warn(ctx context.Context, fields map[string]any, msg string) {
	log(ctx, slog.LevelWarn, fields, nil, msg)
}

func Error(ctx context.Context, fields map[string]any, err error, msg string) {
	log(ctx, slog.LevelError, fields, err, msg)
}

func log(ctx context.Context, v slog.Level, fields map[string]any, err error, msg string) {
	if theLogger == nil {
		h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})
		Init(h)
	}
	if !theLogger.Enabled(ctx, v) {
		return
	}

	attrs := make([]slog.Attr, 0, len(fields))
	for key, val := range fields {
		attrs = append(attrs, slog.Attr{Key: key, Value: slog.AnyValue(val)})
	}
	if err != nil {
		attrs = append(attrs, slog.Attr{Key: "error", Value: slog.AnyValue(err)})
	}

	sort.Slice(attrs, func(i, j int) bool { return attrs[i].Key < attrs[j].Key })
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}

	theLogger.LogAttrs(ctx, v, msg, slog.Group("data", args...))
}
