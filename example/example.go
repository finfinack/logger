package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/finfinack/logger/v2/logging"
)

var (
	jsonOutput = flag.Bool("json", false, "emit structured JSON records")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	format := logging.FormatText
	if *jsonOutput {
		format = logging.FormatJSON
	}

	logger, err := logging.NewLogger("MAIN", logging.Config{Format: format})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer logger.Shutdown(ctx)

	logging.SetMinLogLevel(logging.LogLevelInfo)
	logger.Info(ctx, "service started", "port", 8080)
	logger.Warn(ctx, "retrying request", "attempt", 2)
	logger.Infof(ctx, "processed %d items", 5)
	logger.Debug(ctx, "filtered debug message")
}
