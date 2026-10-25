// SPDX-License-Identifier: Apache-2.0
// Copyright Tony Burns

package embedded

import (
	"context"
	"log/slog"

	"go.temporal.io/server/common/log/tag"
)

// logger passes the embedded server's log records to an slog logger. The server logs its
// routine progress at info level. The adapter lowers those records to debug level, and the
// process log then shows only the server's warnings and errors.
type logger struct {
	slog *slog.Logger
}

// newLogger returns a server logger that writes to l.
func newLogger(l *slog.Logger) *logger {
	return &logger{slog: l}
}

// Debug writes the message at debug level.
func (l *logger) Debug(msg string, tags ...tag.Tag) {
	l.log(slog.LevelDebug, msg, tags)
}

// Info lowers a routine progress message to debug level.
func (l *logger) Info(msg string, tags ...tag.Tag) {
	l.log(slog.LevelDebug, msg, tags)
}

// Warn keeps the server's warnings at their own level.
func (l *logger) Warn(msg string, tags ...tag.Tag) {
	l.log(slog.LevelWarn, msg, tags)
}

// Error passes a server error through unchanged.
func (l *logger) Error(msg string, tags ...tag.Tag) {
	l.log(slog.LevelError, msg, tags)
}

// DPanic logs at error level. The server panics on these only in development builds.
func (l *logger) DPanic(msg string, tags ...tag.Tag) {
	l.log(slog.LevelError, msg, tags)
}

// Panic logs the message as an error before it panics with the same text.
func (l *logger) Panic(msg string, tags ...tag.Tag) {
	l.log(slog.LevelError, msg, tags)
	panic(msg)
}

// Fatal logs and panics like Panic. The server calls it only for a state it can't continue
// from, and a panic lets the process's deferred cleanup run where an exit wouldn't.
func (l *logger) Fatal(msg string, tags ...tag.Tag) {
	l.log(slog.LevelError, msg, tags)
	panic(msg)
}

// log writes one record at level, with one attribute for each tag. It checks the level first,
// because encoding a tag's value costs more than a dropped record should.
func (l *logger) log(level slog.Level, msg string, tags []tag.Tag) {
	ctx := context.Background()
	if !l.slog.Enabled(ctx, level) {
		return
	}
	attrs := make([]slog.Attr, 0, len(tags))
	for _, t := range tags {
		attrs = append(attrs, slog.Any(t.Key(), t.Value()))
	}
	l.slog.LogAttrs(ctx, level, msg, attrs...)
}
