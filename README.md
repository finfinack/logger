# Go logging library with a twist

This repo provides a logging library for Go which has a few benefits:

- it adds additional default formatting better suited to identify where a log line came from
- unlike e.g. `glog`, it has minimal dependencies (most importantly, none with C bindings)
- it's based on the default logger but provides `debug`, `warn` and `error` logging too
- the log level can be set dynamically
- the timestamp is RFC3339 compliant

## Log Format

The log format is as follows:

`YYYY-MM-DDThh:mm:ss [hostname][severity][component]: message`

And example of the above would be:

`2024-12-08T12:37:46Z [hostname][WARN ][MAIN]: test`

**Notes**:

- The severity is always formatted as 5 characters.
- The component is always formatted as 4 characters so you want to choose an appropriate abbreviation.

## Usage

The library can be used by including `github.com/finfinack/logger/logging`:

```go
package main

import (
	"context"
	"time"

	"github.com/finfinack/logger/logging"
)

func main() {
	// Global settings
	logging.SetMinLogLevel(logging.LogLevelInfo) // Info is the default level, just a demo here

	logger := logging.NewLogger("MAIN")
	logger.SetWriter(os.Stdout) // stdout is the default, just a demo here
	// Shutdown / cleanup before termination.
	defer logger.Shutdown()

	// Regular logging.
	for i := 0; i < 3; i++ {
		logger.Warnf("test(%d)", i)
		logger.Debug("test") // this should not be logged due to the log level
	}

	// Finally demo the os.Exit(1) of fatal.
	logger.Fatal("final error message")
}
```

You can also run the provided example:

```bash
go run example/example.go
```
