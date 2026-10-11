// Package logging provides context-aware text and JSON logging through slog handlers.
package logging

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var minLogLevel atomic.Int64

// Diagnostics use a separate destination and omit the failed log message.
// Reporting is best-effort because the logging methods cannot return errors.
var writeErrorLogger = log.New(os.Stderr, "logger: ", 0)

func init() { minLogLevel.Store(LogLevelInfo) }

// SetMinLogLevel changes the global threshold, including for existing loggers.
func SetMinLogLevel(level int) { minLogLevel.Store(int64(level)) }

// Format selects the output representation for a built-in handler.
type Format int

const (
	// FormatText preserves the timestamp and hostname/severity/component prefix.
	FormatText Format = iota
	// FormatJSON emits one JSON object per line.
	FormatJSON
)

// Config configures a logger. Its zero value selects text output on stdout.
type Config struct {
	// Format defaults to FormatText.
	Format Format
	// Writer defaults to os.Stdout and remains owned by the caller.
	Writer io.Writer
	// Handler replaces the built-in output. When set, Writer must be nil and
	// Format must be FormatText (the zero value). Its lifecycle belongs to the caller.
	Handler slog.Handler
}

// Logger sends records to a concurrency-safe slog handler.
// Construct loggers with NewLogger; the zero value is not usable.
type Logger struct {
	handler slog.Handler
	writer  *syncWriter
}

// NewLogger creates a logger and adds hostname and uppercased component metadata.
// Invalid or conflicting output configuration is rejected.
func NewLogger(component string, config Config) (*Logger, error) {
	if config.Handler != nil && (config.Writer != nil || config.Format != FormatText) {
		return nil, fmt.Errorf("a custom handler cannot be combined with Writer or Format")
	}
	if config.Format != FormatText && config.Format != FormatJSON {
		return nil, fmt.Errorf("unsupported log format %d", config.Format)
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	component = strings.ToUpper(component)
	if config.Handler != nil {
		return &Logger{handler: config.Handler.WithAttrs([]slog.Attr{
			slog.String("hostname", hostname), slog.String("component", component),
		})}, nil
	}
	writer := config.Writer
	if writer == nil {
		writer = os.Stdout
	}
	output := &syncWriter{writer: writer}
	var handler slog.Handler
	if config.Format == FormatJSON {
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level: slog.LevelDebug,
			ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
				if len(groups) == 0 && attr.Key == slog.LevelKey && attr.Value.Any() == slogLevelFatal {
					return slog.String(slog.LevelKey, logPrefixFatal)
				}
				return attr
			},
		}).WithAttrs([]slog.Attr{
			slog.String("hostname", hostname), slog.String("component", component),
		})
	} else {
		handler = &textHandler{writer: output, hostname: hostname, component: component}
	}
	return &Logger{handler: handler, writer: output}, nil
}

// SetWriter changes a built-in handler's destination, waiting for in-flight writes.
// Nil writers and custom handlers are rejected; custom handlers own their output.
func (l *Logger) SetWriter(writer io.Writer) error {
	if writer == nil {
		return fmt.Errorf("log writer must not be nil")
	}
	if l.writer == nil {
		return fmt.Errorf("cannot change the writer of a custom handler")
	}
	l.writer.mu.Lock()
	defer l.writer.mu.Unlock()
	l.writer.writer = writer
	return nil
}

// Shutdown logs a farewell. It does not flush or close writers or handlers.
func (l *Logger) Shutdown(ctx context.Context) {
	if len(logExitMessages) > 0 {
		l.Info(ctx, logExitMessages[rand.Intn(len(logExitMessages))])
	}
}

// Log emits a record at a supported LogLevel with slog-style attributes.
// The keys time, level, msg, hostname, and component are reserved at the top level.
// Invalid levels and reserved attributes produce a stderr diagnostic.
// A nil context is treated as context.Background(); cancellation does not drop logs.
func (l *Logger) Log(ctx context.Context, level int, msg string, args ...any) {
	l.log(ctx, level, msg, args, false)
}

func (l *Logger) log(ctx context.Context, level int, msg string, args []any, formatted bool) {
	severity, ok := toSlogLevel(level)
	if !ok {
		writeErrorLogger.Printf("unsupported log level %d", level)
		return
	}
	if minLogLevel.Load() > int64(level) {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if !l.handler.Enabled(ctx, severity) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	if formatted {
		msg = fmt.Sprintf(msg, args...)
		args = nil
	}
	record := slog.NewRecord(time.Now().UTC(), severity, msg, pcs[0])
	var attributes slog.Record
	attributes.Add(args...)
	attributes.Attrs(func(attr slog.Attr) bool {
		addAttribute(&record, attr)
		return true
	})
	if err := l.handler.Handle(ctx, record); err != nil {
		writeErrorLogger.Printf("write failed: %q", err.Error())
	}
}

func addAttribute(record *slog.Record, attr slog.Attr) {
	attr.Value = attr.Value.Resolve()
	// slog inlines unnamed groups, so validate their keys at the same scope.
	if attr.Key == "" && attr.Value.Kind() == slog.KindGroup {
		for _, nested := range attr.Value.Group() {
			addAttribute(record, nested)
		}
		return
	}
	switch attr.Key {
	case slog.TimeKey, slog.LevelKey, slog.MessageKey, "hostname", "component":
		writeErrorLogger.Printf("attribute key %q is reserved", attr.Key)
	default:
		record.AddAttrs(attr)
	}
}

const slogLevelFatal slog.Level = 12

func toSlogLevel(level int) (slog.Level, bool) {
	switch level {
	case LogLevelDebug:
		return slog.LevelDebug, true
	case LogLevelInfo:
		return slog.LevelInfo, true
	case LogLevelWarn:
		return slog.LevelWarn, true
	case LogLevelError:
		return slog.LevelError, true
	case LogLevelFatal:
		return slogLevelFatal, true
	default:
		return 0, false
	}
}

type syncWriter struct {
	mu     sync.Mutex
	writer io.Writer
}

func (w *syncWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n, err := w.writer.Write(p)
	if n < len(p) && err == nil {
		err = io.ErrShortWrite
	}
	return n, err
}
