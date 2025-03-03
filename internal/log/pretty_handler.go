package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
)

const (
	reset = "\033[0m"

	black        = 30
	red          = 31
	green        = 32
	yellow       = 33
	blue         = 34
	magenta      = 35
	cyan         = 36
	lightGray    = 37
	darkGray     = 90
	lightRed     = 91
	lightGreen   = 92
	lightYellow  = 93
	lightBlue    = 94
	lightMagenta = 95
	lightCyan    = 96
	white        = 97
)

func colorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
}

func newPrettyHandler(opts *slog.HandlerOptions) *prettyHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	b := &bytes.Buffer{}
	return &prettyHandler{
		b: b,
		h: slog.NewJSONHandler(b, &slog.HandlerOptions{
			Level:       opts.Level,
			AddSource:   opts.AddSource,
			ReplaceAttr: suppressDefaults(opts.ReplaceAttr),
		}),
		m: &sync.Mutex{},
	}
}

type prettyHandler struct {
	h slog.Handler
	b *bytes.Buffer
	m *sync.Mutex
}

func (h *prettyHandler) Handle(ctx context.Context, rec slog.Record) error {
	level := rec.Level.String()
	switch rec.Level {
	case slog.LevelDebug:
		level = colorize(darkGray, level)
	case slog.LevelInfo:
		level = colorize(cyan, level)
	case slog.LevelWarn:
		level = colorize(lightYellow, level)
	case slog.LevelError:
		level = colorize(lightRed, level)
	}

	attrs, err := h.computeAttrs(ctx, rec)
	if err != nil {
		return err
	}

	if len(attrs) == 0 {
		fmt.Println( //nolint:forbidigo
			colorize(lightGray, rec.Time.Format("15:04:05.000")),
			level,
			colorize(white, rec.Message),
		)
		return nil
	}

	bytes, err := json.Marshal(attrs)
	if err != nil {
		return fmt.Errorf("error when marshaling attrs: %w", err)
	}

	fmt.Println( //nolint:forbidigo
		colorize(lightGray, rec.Time.Format("15:04:05.000")),
		level,
		colorize(white, rec.Message),
		colorize(darkGray, string(bytes)),
	)

	return nil
}

func (h *prettyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &prettyHandler{h: h.h.WithAttrs(attrs), b: h.b, m: h.m}
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	return &prettyHandler{h: h.h.WithGroup(name), b: h.b, m: h.m}
}

func (h *prettyHandler) computeAttrs(ctx context.Context, rec slog.Record) (map[string]any, error) {
	h.m.Lock()
	defer func() {
		h.b.Reset()
		h.m.Unlock()
	}()
	if err := h.h.Handle(ctx, rec); err != nil {
		return nil, fmt.Errorf("error when calling inner handler's Handle: %w", err)
	}

	var attrs map[string]any
	err := json.Unmarshal(h.b.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshaling inner handler's Handle result: %w", err)
	}
	return attrs, nil
}

func suppressDefaults(next func([]string, slog.Attr) slog.Attr) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, attr slog.Attr) slog.Attr {
		if attr.Key == slog.TimeKey ||
			attr.Key == slog.LevelKey ||
			attr.Key == slog.MessageKey {
			return slog.Attr{}
		}
		if next == nil {
			return attr
		}
		return next(groups, attr)
	}
}
