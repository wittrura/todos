package logging

import (
	"io"
	"log/slog"
	"slices"
)

type logger struct {
	logger *slog.Logger
}

func NewLogger(buf io.Writer) *logger {
	return &logger{
		logger: slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{})),
	}
}

func (l *logger) Info(msg string, fields map[string]any) {
	var keys []string
	for k := range fields {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var args []any
	for _, key := range keys {
		args = append(args, []any{key, fields[key]}...)
	}
	l.logger.Info(msg, args...)
}
