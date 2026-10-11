package logging

import (
	"context"
	"os"
)

// Debug logs at debug level with slog-style attributes.
func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelDebug, msg, args, false)
}

// Debugln logs at debug level with slog-style attributes.
func (l *Logger) Debugln(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelDebug, msg, args, false)
}

// Debugf logs at debug level using printf formatting.
func (l *Logger) Debugf(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelDebug, msg, args, true)
}

// Info logs at info level with slog-style attributes.
func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelInfo, msg, args, false)
}

// Infoln logs at info level with slog-style attributes.
func (l *Logger) Infoln(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelInfo, msg, args, false)
}

// Infof logs at info level using printf formatting.
func (l *Logger) Infof(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelInfo, msg, args, true)
}

// Warn logs at warn level with slog-style attributes.
func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelWarn, msg, args, false)
}

// Warnln logs at warn level with slog-style attributes.
func (l *Logger) Warnln(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelWarn, msg, args, false)
}

// Warnf logs at warn level using printf formatting.
func (l *Logger) Warnf(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelWarn, msg, args, true)
}

// Error logs at error level with slog-style attributes.
func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelError, msg, args, false)
}

// Errorln logs at error level with slog-style attributes.
func (l *Logger) Errorln(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelError, msg, args, false)
}

// Errorf logs at error level using printf formatting.
func (l *Logger) Errorf(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelError, msg, args, true)
}

// Fatal logs at fatal level with slog-style attributes. It then logs a farewell and exits with code 1, even if filtered. Deferred functions do not run.
func (l *Logger) Fatal(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelFatal, msg, args, false)
	l.Shutdown(ctx)
	os.Exit(1)
}

// Fatalln logs at fatal level with slog-style attributes. It then logs a farewell and exits with code 1, even if filtered. Deferred functions do not run.
func (l *Logger) Fatalln(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelFatal, msg, args, false)
	l.Shutdown(ctx)
	os.Exit(1)
}

// Fatalf logs at fatal level using printf formatting. It then logs a farewell and exits with code 1, even if filtered. Deferred functions do not run.
func (l *Logger) Fatalf(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelFatal, msg, args, true)
	l.Shutdown(ctx)
	os.Exit(1)
}

// Print logs at info level with slog-style attributes.
func (l *Logger) Print(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelInfo, msg, args, false)
}

// Println logs at info level with slog-style attributes.
func (l *Logger) Println(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelInfo, msg, args, false)
}

// Printf logs at info level using printf formatting.
func (l *Logger) Printf(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LogLevelInfo, msg, args, true)
}
