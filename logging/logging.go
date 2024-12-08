package logging

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"strings"
	"time"
)

type Logger struct {
	ctx    context.Context
	logger *log.Logger

	logLevel int

	component string
	hostname  string
}

func NewLogger(ctx context.Context, writer io.Writer, logLevel int, component string, hostname string) *Logger {
	out := writer
	if out == nil {
		out = os.Stdout
	}
	logger := log.New(out, "", 0) // omit default prefixes

	l := &Logger{
		ctx:       ctx,
		logger:    logger,
		logLevel:  logLevel,
		component: strings.ToUpper(component),
		hostname:  hostname,
	}
	go l.waitForContext()
	return l
}

func (l *Logger) waitForContext() {
	for {
		select {
		case <-time.After(contextCheckDelay):
			continue // the context is still valid
		case <-l.ctx.Done():
			idx := big.NewInt(0)
			if len(logExitMessages) > 1 {
				idx, _ = rand.Int(rand.Reader, big.NewInt(int64(len(logExitMessages)-1)))
			}
			l.Info(logExitMessages[idx.Int64()])
			return // the context has been cancelled
		}
	}
}

func (l *Logger) constructLogLine(severity, msg string, v ...any) string {
	msg = fmt.Sprintf(msg, v...)
	return fmt.Sprintf("[%s][%-5s][%-4s][%s]: %s", time.Now().UTC().Format(logTimeFormat), severity, l.component, l.hostname, msg)
}

func (l *Logger) log(severity string, exit bool, msg string, v ...any) {
	msg = l.constructLogLine(severity, msg, v...)
	if exit {
		l.logger.Fatal(msg) // Fatal prints and then calls os.Exit(1)
	}
	l.logger.Print(msg)
}

func (l *Logger) Debug(msg string) {
	if l.logLevel > LogLevelDebug {
		return // log level is set higher, so ignoring this message
	}
	l.log(LogSeverityDebug, false, msg)
}

func (l *Logger) Debugf(format string, v ...any) {
	if l.logLevel > LogLevelDebug {
		return // log level is set higher, so ignoring this message
	}
	l.log(LogSeverityDebug, false, format, v...)
}

func (l *Logger) Info(msg string) {
	if l.logLevel > LogLevelInfo {
		return // log level is set higher, so ignoring this message
	}
	l.log(LogSeverityInfo, false, msg)
}

func (l *Logger) Infof(format string, v ...any) {
	if l.logLevel > LogLevelInfo {
		return // log level is set higher, so ignoring this message
	}
	l.log(LogSeverityInfo, false, format, v...)
}

func (l *Logger) Warn(msg string) {
	if l.logLevel > LogLevelWarn {
		return // log level is set higher, so ignoring this message
	}
	l.log(LogSeverityWarn, false, msg)
}

func (l *Logger) Warnf(format string, v ...any) {
	if l.logLevel > LogLevelWarn {
		return // log level is set higher, so ignoring this message
	}
	l.log(LogSeverityWarn, false, format, v...)
}

func (l *Logger) Error(msg string) {
	if l.logLevel > LogLevelError {
		return // log level is set higher, so ignoring this message
	}
	l.log(LogSeverityError, false, msg)
}

func (l *Logger) Errorf(format string, v ...any) {
	if l.logLevel > LogLevelError {
		return // log level is set higher, so ignoring this message
	}
	l.log(LogSeverityError, false, format, v...)
}

func (l *Logger) Fatal(msg string) {
	if l.logLevel > LogLevelFatal {
		return // log level is set higher, so ignoring this message
	}
	l.log(LogSeverityFatal, true, msg)
}

func (l *Logger) Fatalf(format string, v ...any) {
	if l.logLevel > LogLevelFatal {
		return // log level is set higher, so ignoring this message
	}
	l.log(LogSeverityFatal, true, format, v...)
}
