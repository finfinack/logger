package main

import (
	"os"

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
