package logging

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"strings"
	"time"
)

var (
	minLogLevel = LogLevelInfo
)

func SetMinLogLevel(level int) {
	minLogLevel = level
}

type Logger struct {
	logger    *log.Logger
	component string
	hostname  string
}

func NewLogger(component string) *Logger {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return &Logger{
		logger:    log.New(os.Stdout, "", 0), // omit default prefixes
		component: strings.ToUpper(component),
		hostname:  hostname,
	}
}

func (l *Logger) SetWriter(writer io.Writer) {
	l.logger = log.New(writer, "", 0) // omit default prefixes
}

func (l *Logger) Shutdown() {
	idx := big.NewInt(0)
	if len(logExitMessages) > 1 {
		idx, _ = rand.Int(rand.Reader, big.NewInt(int64(len(logExitMessages)-1)))
	}
	l.Info(logExitMessages[idx.Int64()])
}

func (l *Logger) constructLogLine(logLevel int, msg string, v ...any) string {
	msg = fmt.Sprintf(msg, v...)
	severity, err := LevelToName(logLevel)
	if err != nil {
		severity = logPrefixUnknown
	}
	return fmt.Sprintf("%s [%s][%-5s][%-4s]: %s", time.Now().UTC().Format(logTimeFormat), l.hostname, severity, l.component, msg)
}

func (l *Logger) Log(logLevel int, exit bool, msg string, v ...any) {
	if minLogLevel > logLevel {
		return // log level is set higher, so ignoring this message
	}

	msg = l.constructLogLine(logLevel, msg, v...)
	l.logger.Print(msg)
	if exit {
		l.Terminate()
		code := 0
		if logLevel >= LogLevelError {
			code = 1
		}
		os.Exit(code)
	}
}

func (l *Logger) Debug(a ...any)                 { l.Log(LogLevelDebug, false, "", a...) }
func (l *Logger) Debugln(a ...any)               { l.Debug(a...) }
func (l *Logger) Debugf(format string, v ...any) { l.Log(LogLevelDebug, false, format, v...) }

func (l *Logger) Info(a ...any)                 { l.Log(LogLevelInfo, false, "", a...) }
func (l *Logger) Infoln(a ...any)               { l.Info(a...) }
func (l *Logger) Infof(format string, v ...any) { l.Log(LogLevelInfo, false, format, v...) }

func (l *Logger) Warn(a ...any)                 { l.Log(LogLevelWarn, false, "", a...) }
func (l *Logger) Warnln(a ...any)               { l.Warn(a...) }
func (l *Logger) Warnf(format string, v ...any) { l.Log(LogLevelWarn, false, format, v...) }

func (l *Logger) Error(a ...any)                 { l.Log(LogLevelError, false, "", a...) }
func (l *Logger) Errorln(a ...any)               { l.Error(a...) }
func (l *Logger) Errorf(format string, v ...any) { l.Log(LogLevelError, false, format, v...) }

func (l *Logger) Fatal(a ...any)                 { l.Log(LogLevelFatal, true, "", a...) }
func (l *Logger) Fatalln(a ...any)               { l.Fatal(a...) }
func (l *Logger) Fatalf(format string, v ...any) { l.Log(LogLevelFatal, true, format, v...) }

// Print* functions are just a redirect to their Info* counterparts.
func (l *Logger) Print(a ...any)                 { l.Info(a...) }
func (l *Logger) Println(a ...any)               { l.Infoln(a...) }
func (l *Logger) Printf(format string, v ...any) { l.Infof(format, v...) }
