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
	if minLogLevel > logLevel {
		return // log level is set higher, so ignoring this message
	}

	msg = l.constructLogLine(logLevel, msg, v...)
	l.logger.Print(msg)
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
