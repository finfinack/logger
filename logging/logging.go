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
	"sync"
	"time"
)

type Logger struct {
	ctx    context.Context
	logger *log.Logger

	mu sync.Mutex

	component string
	hostname  string
}

func NewLogger(ctx context.Context, writer io.Writer, component string, hostname string) *Logger {
	out := writer
	if out == nil {
		out = os.Stdout
	}
	logger := log.New(out, "", log.Ldate|log.Ltime|log.Lshortfile)

	l := &Logger{
		ctx:       ctx,
		logger:    logger,
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

func (l *Logger) logf(pfx string, exit bool, format string, v ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logger.SetPrefix(fmt.Sprintf("%s%s - [%-4s] ", pfx, l.hostname, l.component))
	if exit {
		l.logger.Fatalf(format, v...) // Fatal prints and then calls os.Exit(1)
	}
	l.logger.Printf(format, v...)
}

func (l *Logger) log(pfx string, exit bool, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logger.SetPrefix(fmt.Sprintf("%s%s - [%-4s] ", pfx, l.hostname, l.component))
	if exit {
		l.logger.Fatal(msg) // Fatal prints and then calls os.Exit(1)
	}
	l.logger.Print(msg)
}

func (l *Logger) Debug(msg string) {
	l.log(LogPrefixDebug, false, msg)
}

func (l *Logger) Debugf(format string, v ...any) {
	l.logf(LogPrefixDebug, false, format, v...)
}

func (l *Logger) Info(msg string) {
	l.log(LogPrefixInfo, false, msg)
}

func (l *Logger) Infof(format string, v ...any) {
	l.logf(LogPrefixInfo, false, format, v...)
}

func (l *Logger) Warn(msg string) {
	l.log(LogPrefixWarn, false, msg)
}

func (l *Logger) Warnf(format string, v ...any) {
	l.logf(LogPrefixWarn, false, format, v...)
}

func (l *Logger) Error(msg string) {
	l.log(LogPrefixError, false, msg)
}

func (l *Logger) Errorf(format string, v ...any) {
	l.logf(LogPrefixError, false, format, v...)
}

func (l *Logger) Fatal(msg string) {
	l.log(LogPrefixFatal, true, msg)
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.logf(LogPrefixFatal, true, format, v...)
}
