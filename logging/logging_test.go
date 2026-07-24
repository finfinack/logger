package logging

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestLogFormat(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger("main") // lower-case on purpose: should be upper-cased
	l.SetWriter(&buf)
	SetMinLogLevel(LogLevelDebug)

	l.Warn("hello world")

	// timestamp [hostname][WARN ][MAIN]: hello world
	re := regexp.MustCompile(`^\S+ \[[^\]]*\]\[WARN \]\[MAIN\]: hello world\n$`)
	if got := buf.String(); !re.MatchString(got) {
		t.Errorf("Warn output = %q, does not match %v", got, re)
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger("t")
	l.SetWriter(&buf)
	SetMinLogLevel(LogLevelWarn)

	l.Info("dropped") // below threshold
	if buf.Len() != 0 {
		t.Errorf("Info below threshold logged %q, want nothing", buf.String())
	}

	l.Error("kept") // at/above threshold
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
		{"literal percent", func(l *Logger) { l.Info("100% done") }, "100% done"},
		{"formatted", func(l *Logger) { l.Infof("val=%d", 5) }, "val=5"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := NewLogger("t")
			l.SetWriter(&buf)
			SetMinLogLevel(LogLevelDebug)

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

// TestShutdownReachesAllMessages guards against the off-by-one that previously
// made the last exit message unreachable.
func TestShutdownReachesAllMessages(t *testing.T) {
	if len(logExitMessages) == 0 {
		t.Skip("no exit messages configured")
	}
	last := logExitMessages[len(logExitMessages)-1]
	SetMinLogLevel(LogLevelDebug)

	seen := false
	for i := 0; i < 100000 && !seen; i++ {
		var buf bytes.Buffer
		l := NewLogger("t")
		l.SetWriter(&buf)
		l.Shutdown()
		if strings.Contains(buf.String(), last) {
			seen = true
		}
	}
	if !seen {
		t.Errorf("last exit message %q was never selected by Shutdown", last)
	}
}
