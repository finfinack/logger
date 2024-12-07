# Go logging library with a twist

This repo provides a logging library for Go which has a few benefits:

- it adds additional default formatting better suited to identify where a log line came from
- unlike e.g. `glog`, it has minimal dependencies (most importantly, none with C bindings)
- it's based on the default logger but provides `debug`, `warn` and `error` logging too
- the log level can be set dynamically

## Log Format

The log format is as follows:

`<LOG LEVEL>: <hostname> - [<COMPONENT>] YYYY/MM/DD hh:mm:ss <source file>:<line>: message`

And example of the above would be:

`WARN: srv01 - [MAIN] 2024/12/07 14:46:38 logging.go:70: test message`

**Note**: The component is always formatted as 4 characters so you want to choose an appropriate abbreviation.

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
	ctx, cancel := context.WithCancel(context.Background())
	logger := logging.NewLogger(ctx, nil, logging.LogLevelInfo, "MAIN", "hostname")

	// Regular logging.
	for i := 0; i < 3; i++ {
		logger.Warnf("test(%d)", i)
		logger.Debug("test") // this should not be logged due to the log level
	}

	// Cancel the context and wait for tasks to complete.
	cancel()
	time.Sleep(3 * time.Second)

	// Finally demo the os.Exit(1) of fatal.
	logger.Fatal("final error message")
}
```
