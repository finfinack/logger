package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"
)

func newTestLogger(t *testing.T, component string, config Config) *Logger {
	t.Helper()
	logger, err := NewLogger(component, config)
	if err != nil {
		t.Fatal(err)
	}
	return logger
}

func setTestLevel(t *testing.T, level int) {
	t.Helper()
	previous := minLogLevel.Load()
	t.Cleanup(func() { minLogLevel.Store(previous) })
	SetMinLogLevel(level)
}

func captureDiagnostics(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	previous := writeErrorLogger.Writer()
	writeErrorLogger.SetOutput(&buf)
	t.Cleanup(func() { writeErrorLogger.SetOutput(previous) })
	return &buf
}

func decodeRecord(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("decode record %q: %v", data, err)
	}
	return record
}

func TestJSONRecords(t *testing.T) {
	setTestLevel(t, LogLevelDebug)
	tests := []struct {
		name, level string
		log         func(*Logger, context.Context)
		formatted   bool
	}{
		{"Debug", "DEBUG", func(l *Logger, ctx context.Context) { l.Debug(ctx, "count", "count", 5) }, false},
		{"Debugln", "DEBUG", func(l *Logger, ctx context.Context) { l.Debugln(ctx, "count", "count", 5) }, false},
		{"Debugf", "DEBUG", func(l *Logger, ctx context.Context) { l.Debugf(ctx, "count=%d", 5) }, true},
		{"Info", "INFO", func(l *Logger, ctx context.Context) { l.Info(ctx, "count", "count", 5) }, false},
		{"Infoln", "INFO", func(l *Logger, ctx context.Context) { l.Infoln(ctx, "count", "count", 5) }, false},
		{"Infof", "INFO", func(l *Logger, ctx context.Context) { l.Infof(ctx, "count=%d", 5) }, true},
		{"Warn", "WARN", func(l *Logger, ctx context.Context) { l.Warn(ctx, "count", "count", 5) }, false},
		{"Warnln", "WARN", func(l *Logger, ctx context.Context) { l.Warnln(ctx, "count", "count", 5) }, false},
		{"Warnf", "WARN", func(l *Logger, ctx context.Context) { l.Warnf(ctx, "count=%d", 5) }, true},
		{"Error", "ERROR", func(l *Logger, ctx context.Context) { l.Error(ctx, "count", "count", 5) }, false},
		{"Errorln", "ERROR", func(l *Logger, ctx context.Context) { l.Errorln(ctx, "count", "count", 5) }, false},
		{"Errorf", "ERROR", func(l *Logger, ctx context.Context) { l.Errorf(ctx, "count=%d", 5) }, true},
		{"Print", "INFO", func(l *Logger, ctx context.Context) { l.Print(ctx, "count", "count", 5) }, false},
		{"Println", "INFO", func(l *Logger, ctx context.Context) { l.Println(ctx, "count", "count", 5) }, false},
		{"Printf", "INFO", func(l *Logger, ctx context.Context) { l.Printf(ctx, "count=%d", 5) }, true},
		{"Log", "WARN", func(l *Logger, ctx context.Context) { l.Log(ctx, LogLevelWarn, "count", "count", 5) }, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := newTestLogger(t, "main", Config{Format: FormatJSON, Writer: &buf})
			tc.log(logger, context.Background())
			record := decodeRecord(t, buf.Bytes())
			if record["level"] != tc.level || record["component"] != "MAIN" {
				t.Errorf("unexpected metadata: %v", record)
			}
			hostname, ok := record["hostname"].(string)
			if !ok || hostname == "" {
				t.Errorf("missing hostname: %v", record)
			}
			timestamp, ok := record["time"].(string)
			if !ok {
				t.Fatalf("missing timestamp: %v", record)
			}
			if parsed, err := time.Parse(time.RFC3339Nano, timestamp); err != nil || parsed.Location() != time.UTC {
				t.Errorf("timestamp %q is not RFC3339 UTC", timestamp)
			}
			if tc.formatted {
				if record["msg"] != "count=5" {
					t.Errorf("formatted record: %v", record)
				}
				if _, exists := record["count"]; exists {
					t.Errorf("format args became attributes: %v", record)
				}
			} else if record["msg"] != "count" || record["count"] != float64(5) {
				t.Errorf("attributes lost their type: %v", record)
			}
		})
	}
}

func TestStructuredAttributes(t *testing.T) {
	setTestLevel(t, LogLevelInfo)
	for _, format := range []Format{FormatText, FormatJSON} {
		t.Run(map[Format]string{FormatText: "text", FormatJSON: "json"}[format], func(t *testing.T) {
			var buf bytes.Buffer
			logger := newTestLogger(t, "main", Config{Format: format, Writer: &buf})
			logger.Info(context.Background(), "request completed", "status", 200,
				slog.Bool("ok", true), slog.Group("request", slog.String("id", "hello world")),
				"error", errors.New("request failed"))
			if format == FormatJSON {
				record := decodeRecord(t, buf.Bytes())
				if record["status"] != float64(200) || record["ok"] != true || record["error"] != "request failed" {
					t.Errorf("record = %v", record)
				}
				request, ok := record["request"].(map[string]any)
				if !ok || request["id"] != "hello world" {
					t.Errorf("group = %v", record["request"])
				}
			} else {
				for _, want := range []string{"status=200", "ok=true", `request.id="hello world"`, `error="request failed"`} {
					if !strings.Contains(buf.String(), want) {
						t.Errorf("output %q missing %q", buf.String(), want)
					}
				}
			}
		})
	}
}

func TestSingleLineRecords(t *testing.T) {
	setTestLevel(t, LogLevelInfo)
	message := "line one\nline two\r\x1b[31m\u2028\u2029"
	for _, format := range []Format{FormatText, FormatJSON} {
		t.Run(map[Format]string{FormatText: "text", FormatJSON: "json"}[format], func(t *testing.T) {
			var buf bytes.Buffer
			logger := newTestLogger(t, "m\nain", Config{Format: format, Writer: &buf})
			logger.Info(context.Background(), message, "value", "untrusted\n\x1b[31m")
			output := buf.String()
			if strings.Count(output, "\n") != 1 || strings.ContainsAny(output, "\r\x1b\u2028\u2029") {
				t.Errorf("unsafe physical record: %q", output)
			}
			if format == FormatJSON {
				record := decodeRecord(t, buf.Bytes())
				if record["msg"] != message || record["component"] != "M\nAIN" {
					t.Errorf("JSON did not preserve values: %v", record)
				}
			} else if !strings.Contains(output, `line one\nline two\r\x1b[31m\u2028\u2029`) {
				t.Errorf("message not escaped: %q", output)
			}
		})
	}
}

func TestConfigAndWriterValidation(t *testing.T) {
	handler := slog.NewJSONHandler(io.Discard, nil)
	tests := []struct {
		name   string
		config Config
	}{
		{"invalid format", Config{Format: Format(99)}},
		{"custom writer", Config{Handler: handler, Writer: io.Discard}},
		{"custom format", Config{Handler: handler, Format: FormatJSON}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if logger, err := NewLogger("test", tc.config); err == nil || logger != nil {
				t.Fatalf("NewLogger = %v, %v; want nil and error", logger, err)
			}
		})
	}
	logger := newTestLogger(t, "test", Config{Writer: io.Discard})
	if err := logger.SetWriter(nil); err == nil {
		t.Error("nil writer accepted")
	}
	custom := newTestLogger(t, "test", Config{Handler: handler})
	if err := custom.SetWriter(io.Discard); err == nil {
		t.Error("custom handler writer accepted")
	}
}

type probeState struct {
	enabledCtx, handledCtx context.Context
	record                 slog.Record
	disabled               bool
}

type probeHandler struct {
	inner slog.Handler
	state *probeState
}

func (h *probeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	h.state.enabledCtx = ctx
	return !h.state.disabled && h.inner.Enabled(ctx, level)
}
func (h *probeHandler) Handle(ctx context.Context, record slog.Record) error {
	h.state.handledCtx, h.state.record = ctx, record.Clone()
	return h.inner.Handle(ctx, record)
}
func (h *probeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &probeHandler{inner: h.inner.WithAttrs(attrs), state: h.state}
}
func (h *probeHandler) WithGroup(name string) slog.Handler {
	return &probeHandler{inner: h.inner.WithGroup(name), state: h.state}
}

func TestCustomHandlerContext(t *testing.T) {
	setTestLevel(t, LogLevelDebug)
	var buf bytes.Buffer
	state := &probeState{}
	handler := &probeHandler{inner: slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}), state: state}
	logger := newTestLogger(t, "main", Config{Handler: handler})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	logger.Info(ctx, "canceled request", "status", 499)
	if state.enabledCtx != ctx || state.handledCtx != ctx {
		t.Error("request context was not forwarded to both handler methods")
	}
	record := decodeRecord(t, buf.Bytes())
	if record["component"] != "MAIN" || record["status"] != float64(499) {
		t.Errorf("record = %v", record)
	}
	source := state.record.Source()
	if source == nil || !strings.HasSuffix(source.File, "structured_test.go") {
		t.Errorf("source points to wrapper: %v", source)
	}
	buf.Reset()
	logger.Info(context.TODO(), "nil context")
	if state.enabledCtx == nil || state.handledCtx == nil {
		t.Error("nil context was not normalized")
	}
	state.disabled = true
	buf.Reset()
	logger.Info(ctx, "handler filtered")
	if buf.Len() != 0 {
		t.Errorf("disabled handler received output: %q", buf.String())
	}
}

func TestDynamicFiltering(t *testing.T) {
	setTestLevel(t, LogLevelWarn)
	for _, format := range []Format{FormatText, FormatJSON} {
		var buf bytes.Buffer
		logger := newTestLogger(t, "test", Config{Format: format, Writer: &buf})
		SetMinLogLevel(LogLevelWarn)
		logger.Debug(context.Background(), "dropped")
		if buf.Len() != 0 {
			t.Errorf("format %d did not filter debug", format)
		}
		SetMinLogLevel(LogLevelDebug)
		logger.Debug(context.Background(), "enabled dynamically")
		if !strings.Contains(buf.String(), "enabled dynamically") {
			t.Errorf("format %d did not update threshold", format)
		}
	}
	// Thresholds retain the full range of int rather than narrowing to int32.
	maximum := int(^uint(0) >> 1)
	var buf bytes.Buffer
	logger := newTestLogger(t, "test", Config{Writer: &buf})
	SetMinLogLevel(maximum)
	logger.Info(context.Background(), "dropped")
	if buf.Len() != 0 {
		t.Error("large threshold overflowed")
	}
	SetMinLogLevel(-maximum - 1)
	logger.Info(context.Background(), "enabled")
	if buf.Len() == 0 {
		t.Error("small threshold overflowed")
	}
}

func TestReservedAttributes(t *testing.T) {
	setTestLevel(t, LogLevelInfo)
	diagnostics := captureDiagnostics(t)
	var buf bytes.Buffer
	logger := newTestLogger(t, "main", Config{Format: FormatJSON, Writer: &buf})
	logger.Info(context.Background(), "original", "msg", "forged",
		slog.Group("", slog.String("component", "forged"), slog.Int("count", 5)),
		slog.Group("request", slog.String("msg", "allowed")))
	record := decodeRecord(t, buf.Bytes())
	if record["msg"] != "original" || record["component"] != "MAIN" || record["count"] != float64(5) {
		t.Errorf("reserved field overwritten: %v", record)
	}
	if strings.Count(buf.String(), `"msg":`) != 2 || strings.Count(buf.String(), `"component":`) != 1 {
		t.Errorf("unexpected duplicate fields: %q", buf.String())
	}
	if !strings.Contains(diagnostics.String(), `attribute key "msg" is reserved`) || !strings.Contains(diagnostics.String(), `attribute key "component" is reserved`) {
		t.Errorf("diagnostics = %q", diagnostics.String())
	}
}

type failingWriter struct{ short bool }

func (w failingWriter) Write(p []byte) (int, error) {
	if w.short {
		return len(p) - 1, nil
	}
	return 0, errors.New("disk full\nforged diagnostic")
}

func TestOutputErrors(t *testing.T) {
	setTestLevel(t, LogLevelInfo)
	diagnostics := captureDiagnostics(t)
	for _, format := range []Format{FormatText, FormatJSON} {
		for _, short := range []bool{false, true} {
			diagnostics.Reset()
			logger := newTestLogger(t, "test", Config{Format: format, Writer: failingWriter{short: short}})
			logger.Error(context.Background(), "sensitive message")
			want := `logger: write failed: "disk full\nforged diagnostic"` + "\n"
			if short {
				want = `logger: write failed: "short write"` + "\n"
			}
			if diagnostics.String() != want {
				t.Errorf("format %d short=%v: diagnostics = %q, want %q", format, short, diagnostics.String(), want)
			}
		}
	}
	diagnostics.Reset()
	var buf bytes.Buffer
	logger := newTestLogger(t, "test", Config{Writer: &buf})
	logger.Log(context.Background(), 999, "invalid level")
	if buf.Len() != 0 || !strings.Contains(diagnostics.String(), "unsupported log level 999") {
		t.Errorf("invalid level output=%q diagnostics=%q", buf.String(), diagnostics.String())
	}
	diagnostics.Reset()
	logger.Info(context.Background(), "unmatched", "orphan")
	if !strings.Contains(buf.String(), "!BADKEY=orphan") {
		t.Errorf("malformed attributes disappeared: %q", buf.String())
	}
}

func TestConcurrentWriterChanges(t *testing.T) {
	setTestLevel(t, LogLevelInfo)
	for _, format := range []Format{FormatText, FormatJSON} {
		t.Run(map[Format]string{FormatText: "text", FormatJSON: "json"}[format], func(t *testing.T) {
			var buf bytes.Buffer
			logger := newTestLogger(t, "test", Config{Format: format, Writer: &buf})
			var wg sync.WaitGroup
			for worker := 0; worker < 5; worker++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < 100; i++ {
						logger.Info(context.Background(), "record", "count", i)
					}
				}()
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 100; i++ {
					if err := logger.SetWriter(&buf); err != nil {
						t.Error(err)
						return
					}
					SetMinLogLevel(LogLevelInfo)
				}
			}()
			wg.Wait()
			lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
			if len(lines) != 500 {
				t.Fatalf("got %d records, want 500", len(lines))
			}
			for _, line := range lines {
				if format == FormatJSON {
					decodeRecord(t, []byte(line))
				} else if !strings.Contains(line, "[INFO ][TEST]: record count=") {
					t.Fatalf("corrupt text record: %q", line)
				}
			}
			var next bytes.Buffer
			if err := logger.SetWriter(&next); err != nil {
				t.Fatal(err)
			}
			logger.Info(context.Background(), "new destination")
			if !strings.Contains(next.String(), "new destination") {
				t.Error("SetWriter did not change destination")
			}
		})
	}
}

func TestFatalExit(t *testing.T) {
	for _, method := range []string{"Fatal", "Fatalln", "Fatalf"} {
		for _, filtered := range []bool{false, true} {
			t.Run(method+map[bool]string{false: "/enabled", true: "/filtered"}[filtered], func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestFatalHelper$")
				cmd.Env = append(os.Environ(), "LOGGER_TEST_FATAL="+method)
				if filtered {
					cmd.Env = append(cmd.Env, "LOGGER_TEST_FILTERED=1")
				}
				output, err := cmd.CombinedOutput()
				var exitErr *exec.ExitError
				if ctx.Err() != nil || !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
					t.Fatalf("Fatal did not exit with 1: %v; output %s", err, output)
				}
				if filtered {
					if len(output) != 0 {
						t.Errorf("filtered fatal output: %q", output)
					}
				} else {
					record := decodeRecord(t, output)
					if record["level"] != "FATAL" || record["msg"] != "fatal record" {
						t.Errorf("fatal record = %v", record)
					}
				}
			})
		}
	}
}

func TestFatalHelper(t *testing.T) {
	method := os.Getenv("LOGGER_TEST_FATAL")
	if method == "" {
		t.Skip("subprocess helper")
	}
	SetMinLogLevel(LogLevelFatal)
	if os.Getenv("LOGGER_TEST_FILTERED") == "1" {
		SetMinLogLevel(LogLevelFatal + 1)
	}
	logger := newTestLogger(t, "test", Config{Format: FormatJSON})
	ctx := context.Background()
	switch method {
	case "Fatal":
		logger.Fatal(ctx, "fatal record")
	case "Fatalln":
		logger.Fatalln(ctx, "fatal record")
	case "Fatalf":
		logger.Fatalf(ctx, "fatal %s", "record")
	default:
		t.Fatalf("unknown method %q", method)
	}
	os.Exit(42)
}

func TestTextHandlerGroups(t *testing.T) {
	var buf bytes.Buffer
	base := &textHandler{writer: &buf, hostname: "host", component: "TEST"}
	handler := base.WithAttrs([]slog.Attr{slog.String("bound", "root")}).WithGroup("request").WithAttrs([]slog.Attr{slog.Int("bound", 1)}).WithGroup("nested")
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "hello", 0)
	record.AddAttrs(slog.Int("value", 2))
	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"bound=root", "request.bound=1", "request.nested.value=2"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("output %q missing %q", buf.String(), want)
		}
	}
	buf.Reset()
	if err := base.WithGroup("").Handle(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "hello value=2") || strings.Contains(buf.String(), "bound=") {
		t.Errorf("base handler mutated: %q", buf.String())
	}
}

func TestFilteredFormatting(t *testing.T) {
	setTestLevel(t, LogLevelWarn)
	logger := newTestLogger(t, "test", Config{Writer: io.Discard})
	value := &observedStringer{}
	logger.Infof(context.Background(), "%s", value)
	if value.called {
		t.Error("filtered log evaluated formatting")
	}
}

type observedStringer struct{ called bool }

func (v *observedStringer) String() string { v.called = true; return "value" }

func FuzzSingleLineRecords(f *testing.F) {
	f.Add("hello", "main")
	f.Add("untrusted\n\r\x1b[31m\u2028", "m\nain")
	f.Add("untrusted\x7f", "main")
	f.Add(string([]byte{0xff, 0, '\n'}), "test")
	f.Fuzz(func(t *testing.T, message, component string) {
		for _, format := range []Format{FormatText, FormatJSON} {
			var buf bytes.Buffer
			logger := newTestLogger(t, component, Config{Format: format, Writer: &buf})
			logger.Info(context.Background(), message, "value", component)
			output := buf.String()
			if strings.Count(output, "\n") != 1 {
				t.Fatalf("format %d emitted multiple physical records: %q", format, output)
			}
			if format == FormatText {
				for _, r := range output {
					if (unicode.IsControl(r) && r != '\n') || r == '\u2028' || r == '\u2029' {
						t.Fatalf("text output emitted raw control character %U", r)
					}
				}
			}
			if format == FormatJSON {
				decodeRecord(t, buf.Bytes())
			}
		}
	})
}
