package logging

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

// minLogLevel is the global threshold below which messages are dropped. It is
// stored atomically because the log level can be changed at runtime while other
// goroutines are logging.
var minLogLevel atomic.Int32

// Write failures get a best-effort diagnostic independent of the configured
// destination. The existing logging API cannot return errors to callers.
var writeErrorLogger = log.New(os.Stderr, "logger: ", 0)

func init() {
	minLogLevel.Store(LogLevelInfo)
}

func SetMinLogLevel(level int) {
	minLogLevel.Store(int32(level))
}

type Logger struct {
	// Keep one logger so writes and destination changes share its output lock.
	logger    *log.Logger
	component string
	hostname  string
}

func NewLogger(component string) *Logger {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	l := &Logger{
		logger:    log.New(os.Stdout, "", 0), // omit default prefixes
		component: strings.ToUpper(component),
		hostname:  hostname,
	}
	return l
}

func (l *Logger) SetWriter(writer io.Writer) {
	l.logger.SetOutput(writer)
}

// Shutdown logs a farewell message. It does not flush or close the writer;
// callers retain ownership of the destination and its cleanup.
func (l *Logger) Shutdown() {
	if len(logExitMessages) == 0 {
		return
	}
	l.Info(logExitMessages[rand.Intn(len(logExitMessages))])
}

func (l *Logger) constructLogLine(logLevel int, msg string, v ...any) string {
	if len(v) > 0 {
		msg = fmt.Sprintf(msg, v...)
	}
	severity, err := LevelToName(logLevel)
	if err != nil {
		severity = logPrefixUnknown
	}
	return fmt.Sprintf("%s [%s][%-5s][%-4s]: %s", time.Now().UTC().Format(logTimeFormat), l.hostname, severity, l.component, msg)
}

func (l *Logger) Log(logLevel int, exit bool, msg string, v ...any) {
	if minLogLevel.Load() <= int32(logLevel) {
		msg = l.constructLogLine(logLevel, msg, v...)
		if err := l.logger.Output(2, msg); err != nil {
			writeErrorLogger.Printf("write failed: %s", err)
		}
	}

	if exit {
		l.Shutdown()
		code := 0
		if logLevel >= LogLevelError {
			code = 1
		}
		os.Exit(code)
	}
}

func (l *Logger) Debug(msg string)               { l.Log(LogLevelDebug, false, msg) }
func (l *Logger) Debugln(msg string)             { l.Debug(msg) }
func (l *Logger) Debugf(format string, v ...any) { l.Log(LogLevelDebug, false, format, v...) }

func (l *Logger) Info(msg string)               { l.Log(LogLevelInfo, false, msg) }
func (l *Logger) Infoln(msg string)             { l.Info(msg) }
func (l *Logger) Infof(format string, v ...any) { l.Log(LogLevelInfo, false, format, v...) }

func (l *Logger) Warn(msg string)               { l.Log(LogLevelWarn, false, msg) }
func (l *Logger) Warnln(msg string)             { l.Warn(msg) }
func (l *Logger) Warnf(format string, v ...any) { l.Log(LogLevelWarn, false, format, v...) }

func (l *Logger) Error(msg string)               { l.Log(LogLevelError, false, msg) }
func (l *Logger) Errorln(msg string)             { l.Error(msg) }
func (l *Logger) Errorf(format string, v ...any) { l.Log(LogLevelError, false, format, v...) }

func (l *Logger) Fatal(msg string)               { l.Log(LogLevelFatal, true, msg) }
func (l *Logger) Fatalln(msg string)             { l.Fatal(msg) }
func (l *Logger) Fatalf(format string, v ...any) { l.Log(LogLevelFatal, true, format, v...) }

// Print* functions are just a redirect to their Info* counterparts.
func (l *Logger) Print(msg string)               { l.Info(msg) }
func (l *Logger) Println(msg string)             { l.Infoln(msg) }
func (l *Logger) Printf(format string, v ...any) { l.Infof(format, v...) }
