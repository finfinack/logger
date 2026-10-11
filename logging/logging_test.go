package logging

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"
)

func TestLogFormat(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(t, "main", Config{}) // lower-case on purpose: should be upper-cased
	if err := l.SetWriter(&buf); err != nil {
		t.Fatal(err)
	}
	setTestLevel(t, LogLevelDebug)

	l.Warn(context.Background(), "hello world")

	// timestamp [hostname][WARN ][MAIN]: hello world
	re := regexp.MustCompile(`^\S+ \[[^\]]*\]\[WARN \]\[MAIN\]: hello world\n$`)
	if got := buf.String(); !re.MatchString(got) {
		t.Errorf("Warn output = %q, does not match %v", got, re)
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(t, "t", Config{})
	if err := l.SetWriter(&buf); err != nil {
		t.Fatal(err)
	}
	setTestLevel(t, LogLevelWarn)

	l.Info(context.Background(), "dropped") // below threshold
	if buf.Len() != 0 {
		t.Errorf("Info below threshold logged %q, want nothing", buf.String())
	}

	l.Error(context.Background(), "kept") // at/above threshold
	if !strings.Contains(buf.String(), "kept") {
		t.Errorf("Error output = %q, want it to contain %q", buf.String(), "kept")
	}
}

// TestFormatArgSafety verifies that messages containing verbs are printed
// literally when no format arguments are supplied, and formatted otherwise.
func TestFormatArgSafety(t *testing.T) {
	tests := []struct {
		name string
		log  func(l *Logger)
		want string
	}{
		{"literal percent", func(l *Logger) { l.Info(context.Background(), "100% done") }, "100% done"},
		{"formatted", func(l *Logger) { l.Infof(context.Background(), "val=%d", 5) }, "val=5"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := newTestLogger(t, "t", Config{})
			if err := l.SetWriter(&buf); err != nil {
				t.Fatal(err)
			}
			setTestLevel(t, LogLevelDebug)

			tc.log(l)

			if !strings.HasSuffix(strings.TrimRight(buf.String(), "\n"), tc.want) {
				t.Errorf("output = %q, want suffix %q", buf.String(), tc.want)
			}
		})
	}
}

func TestLevelToName(t *testing.T) {
	if _, err := LevelToName(999); err == nil {
		t.Error("LevelToName(999) err = nil, want error for unknown level")
	}
	for lvl, want := range logLevelToName {
		got, err := LevelToName(lvl)
		if err != nil {
			t.Errorf("LevelToName(%d) unexpected err: %v", lvl, err)
			continue
		}
		if got != want {
			t.Errorf("LevelToName(%d) = %q, want %q", lvl, got, want)
		}
	}
}

func TestLevelToValue(t *testing.T) {
	if _, err := LevelToValue("nope"); err == nil {
		t.Error("LevelToValue(\"nope\") err = nil, want error for unknown name")
	}
	// Round-trip every known level through name and back, case-insensitively.
	for lvl := range logLevelToName {
		name, err := LevelToName(lvl)
		if err != nil {
			t.Fatalf("LevelToName(%d): %v", lvl, err)
		}
		got, err := LevelToValue(strings.ToLower(name))
		if err != nil {
			t.Errorf("LevelToValue(%q): %v", name, err)
			continue
		}
		if got != lvl {
			t.Errorf("LevelToValue(%q) = %d, want %d", name, got, lvl)
		}
	}
}

func TestShutdown(t *testing.T) {
	setTestLevel(t, LogLevelInfo)
	var buf bytes.Buffer
	logger := newTestLogger(t, "test", Config{Writer: &buf})
	logger.Shutdown(context.Background())
	if len(logExitMessages) == 0 {
		if buf.Len() != 0 {
			t.Errorf("unexpected farewell: %q", buf.String())
		}
		return
	}
	if strings.Count(buf.String(), "\n") != 1 {
		t.Fatalf("expected one farewell: %q", buf.String())
	}
	for _, message := range logExitMessages {
		if strings.HasSuffix(buf.String(), message+"\n") {
			return
		}
	}
	t.Errorf("unknown farewell: %q", buf.String())
}
