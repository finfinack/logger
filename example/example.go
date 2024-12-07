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
