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
	severity, ok := logLevelToMessage[logLevel]
	if !ok {
		severity = logPrefixUnknown
	}
	return fmt.Sprintf("%s [%s][%-5s][%-4s]: %s", time.Now().UTC().Format(logTimeFormat), l.hostname, severity, l.component, msg)
}

func (l *Logger) log(logLevel int, exit bool, msg string, v ...any) {
	if minLogLevel > logLevel {
		return // log level is set higher, so ignoring this message
	}

	msg = l.constructLogLine(logLevel, msg, v...)
	l.logger.Print(msg)
	if exit {
		l.Shutdown()
		os.Exit(1)
	}
}

func (l *Logger) Debug(msg string) {
	l.log(LogLevelDebug, false, msg)
}

func (l *Logger) Debugln(msg string) {
	l.Debug(fmt.Sprintln(msg))
}

func (l *Logger) Debugf(format string, v ...any) {
	l.log(LogLevelDebug, false, format, v...)
}

func (l *Logger) Info(msg string) {
	l.log(LogLevelInfo, false, msg)
}

func (l *Logger) Infoln(msg string) {
	l.Info(fmt.Sprintln(msg))
}

func (l *Logger) Infof(format string, v ...any) {
	l.log(LogLevelInfo, false, format, v...)
}

func (l *Logger) Warn(msg string) {
	l.log(LogLevelWarn, false, msg)
}

func (l *Logger) Warnln(msg string) {
	l.Warn(fmt.Sprintln(msg))
}

func (l *Logger) Warnf(format string, v ...any) {
	l.log(LogLevelWarn, false, format, v...)
}

func (l *Logger) Error(msg string) {
	l.log(LogLevelError, false, msg)
}

func (l *Logger) Errorln(msg string) {
	l.Error(fmt.Sprintln(msg))
}

func (l *Logger) Errorf(format string, v ...any) {
	l.log(LogLevelError, false, format, v...)
}

func (l *Logger) Fatal(msg string) {
	l.log(LogLevelFatal, true, msg)
}

func (l *Logger) Fatalln(msg string) {
	l.Fatal(fmt.Sprintln(msg))
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.log(LogLevelFatal, true, format, v...)
}
