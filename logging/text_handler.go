package logging

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// textHandler retains the original prefix and delegates attribute encoding to slog.
// Each record has its own buffer; shared state is immutable except for the writer.
type textHandler struct {
	writer              io.Writer
	hostname, component string
	attrs               []slog.Attr
	groups              []string
}

func (*textHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *textHandler) Handle(ctx context.Context, record slog.Record) error {
	severity := record.Level.String()
	if record.Level == slogLevelFatal {
		severity = logPrefixFatal
	}
	line := fmt.Sprintf("%s [%s][%-5s][%-4s]: %s",
		record.Time.UTC().Format(logTimeFormat), h.hostname, severity,
		h.component, record.Message)
	attrs := slices.Clone(h.attrs)
	var eventAttrs []slog.Attr
	record.Attrs(func(attr slog.Attr) bool { eventAttrs = append(eventAttrs, attr); return true })
	attrs = append(attrs, groupAttrs(h.groups, eventAttrs)...)
	if len(attrs) > 0 {
		var buf bytes.Buffer
		encoder := slog.NewTextHandler(&buf, &slog.HandlerOptions{
			ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
				if len(groups) == 0 && (attr.Key == slog.LevelKey || attr.Key == slog.MessageKey) {
					return slog.Attr{}
				}
				return attr
			},
		})
		attributes := slog.NewRecord(time.Time{}, slog.LevelInfo, "", 0)
		attributes.AddAttrs(attrs...)
		if err := encoder.Handle(ctx, attributes); err != nil {
			return err
		}
		if encoded := strings.TrimSuffix(buf.String(), "\n"); encoded != "" {
			line += " " + encoded
		}
	}
	_, err := io.WriteString(h.writer, escapeControls(line)+"\n")
	return err
}

func (h *textHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.attrs = append(slices.Clone(h.attrs), groupAttrs(h.groups, slices.Clone(attrs))...)
	return &clone
}

func (h *textHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	clone := *h
	clone.groups = append(slices.Clone(h.groups), name)
	return &clone
}

func groupAttrs(groups []string, attrs []slog.Attr) []slog.Attr {
	for i := len(groups) - 1; i >= 0; i-- {
		attrs = []slog.Attr{slog.GroupAttrs(groups[i], attrs...)}
	}
	return attrs
}

func escapeControls(value string) string {
	control := func(r rune) bool { return unicode.IsControl(r) || r == '\u2028' || r == '\u2029' }
	if strings.IndexFunc(value, control) < 0 {
		return value
	}
	var escaped strings.Builder
	for _, r := range value {
		if control(r) {
			quoted := strconv.QuoteToASCII(string(r))
			escaped.WriteString(quoted[1 : len(quoted)-1])
		} else {
			escaped.WriteRune(r)
		}
	}
	return escaped.String()
}
