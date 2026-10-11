# Go logging library with a twist — v2

Context-aware logging with structured attributes, the original hostname/component
text prefix, and opt-in JSON output. The implementation uses standard-library
`slog.Handler` interfaces and has no external dependencies.

The v1 implementation is maintained on the `v1` branch. v2 uses the module path
`github.com/finfinack/logger/v2`.

## Usage

```go
package main

import (
    "context"
    "log"

    "github.com/finfinack/logger/v2/logging"
)

func main() {
    logger, err := logging.NewLogger("MAIN", logging.Config{
        Format: logging.FormatJSON, // omit for the original text format
    })
    if err != nil {
        log.Fatal(err)
    }
    ctx := context.Background()
    defer logger.Shutdown(ctx)

    logger.Info(ctx, "request completed", "status", 200, "duration_ms", 42)
    logger.Infof(ctx, "processed %d items", 5)
    logger.Debug(ctx, "diagnostics") // filtered at the default info threshold
}
```

Every logging method takes `context.Context` first, including `*f`, `*ln`,
`Print*`, `Fatal*`, `Log`, and `Shutdown`. Pass the request context when available,
or `context.Background()` for startup and other background work. Nil contexts
are treated as background contexts. Cancellation does not suppress logging.

`Debug`, `Info`, `Warn`, `Error`, `Fatal`, and their `*ln` aliases accept alternating
keys and values or `slog.Attr` values, including `slog.Group`. The `*f` methods
use printf formatting; their arguments do not become structured fields.

## Output formats

`Config{}` defaults to `FormatText` and `os.Stdout`. Set `Writer` in the config
to choose a different built-in destination.

Text retains the original prefix and appends quoted key/value attributes:

```text
2026-10-11T12:00:00Z [hostname][INFO ][MAIN]: request completed status=200 duration_ms=42
```

Severity is padded to five characters, and the uppercased component is padded to
four. Control characters throughout the rendered text record are escaped to keep
each record on one physical line. Attribute encoding uses slog's text format.

`FormatJSON` preserves typed values and emits one JSON object per line:

```json
{"time":"2026-10-11T12:00:00Z","level":"INFO","hostname":"hostname","component":"MAIN","msg":"request completed","status":200,"duration_ms":42}
```

Timestamps are UTC: RFC3339 in text and RFC3339 with fractional precision in JSON.
The top-level attribute names `time`, `level`, `msg`, `hostname`, and `component`
are reserved. Conflicting attributes are dropped with a stderr diagnostic;
those names remain usable inside named groups. Unmatched attribute arguments
follow slog's `!BADKEY` convention.

## Levels and output errors

The existing level constants keep their values: debug 0, info 10, warn 20,
error 30, and fatal 40. Change the global threshold at runtime:

```go
logging.SetMinLogLevel(logging.LogLevelDebug)
logger.Log(ctx, logging.LogLevelWarn, "retrying", "attempt", 2)
```

The threshold applies to existing loggers and custom handlers. A custom handler
can impose additional filtering. `Log` accepts the supported level constants;
unsupported levels produce a stderr diagnostic without emitting a record.

Writer and handler errors, including short writes, are reported to stderr on a
best-effort basis. Diagnostics quote errors and omit the failed log message.
Error reporting cannot guarantee delivery if stderr also fails.

`SetWriter` returns an error for nil writers or custom handlers:

```go
if err := logger.SetWriter(os.Stdout); err != nil {
    // Handle the configuration error.
}
```

Built-in writes and destination changes on one logger share a lock. If separate
loggers share a writer, that writer must support concurrent writes.

## Custom handlers and OpenTelemetry

An application can supply any concurrency-safe `slog.Handler`:

```go
logger, err := logging.NewLogger("API", logging.Config{
    Handler: applicationHandler,
})
```

A custom handler receives the context, standard slog levels, hostname/component
attributes, and the source location of the logging call. Fatal records use a
custom slog level of 12. When `Handler` is set, leave `Writer` unset and `Format`
at its zero value; custom handlers own their output and lifecycle.

For OpenTelemetry, configure a LoggerProvider in the application and supply an
[`otelslog.NewHandler`](https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/otelslog#NewHandler)
handler. The application owns the SDK/exporter dependencies, flushing, and
shutdown. This library does not add an OpenTelemetry dependency or automatically
copy context values into attributes.

`Shutdown(ctx)` only logs a random farewell; it does not flush or close anything.
`Fatal(ctx, ...)` logs at fatal level, logs a farewell subject to filtering, and
exits with code 1 even if fatal output is filtered. Deferred functions do not run.
For buffered or telemetry destinations, applications should log the error,
perform their own bounded cleanup, then exit.

## Migrating from v1

1. Change imports from `github.com/finfinack/logger/logging` to
   `github.com/finfinack/logger/v2/logging`.
2. Pass a `Config` to `NewLogger` and handle its returned error.
3. Add a context as the first argument of every logging call and `Shutdown`.
4. Handle errors returned by `SetWriter`, or set `Config.Writer` at construction.
5. Replace `Log(level, false, format, args...)` with `Log(ctx, level, msg, attrs...)`
   for attributes, or a level-specific `*f` method for formatting. Replace
   exit-enabled `Log` calls with application-managed termination or `Fatal`.
6. Account for escaped control characters in text output and fractional JSON
   timestamps when parsing logs.

After v2 is published, install a tagged release with:

```sh
go get github.com/finfinack/logger/v2@v2.0.0
```

## Example

```sh
go run ./example
go run ./example -json
```
