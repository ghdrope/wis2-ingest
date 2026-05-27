package main

import (
	"os"
	"wis2-ingest/pkg/utils"

	"github.com/akuity/kargo/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	// Setup a context that is automatically cancelled on SIGINT/SIGTERM.
	ctx := signals.SetupSignalHandler()

	if utils.IsDebug() {
		logging.LoggerFromContext(ctx).Info("🐛 DEBUG MODE ENABLED")
	}

	// Execute the CLI root command.
	if err := Execute(ctx); err != nil {
		logging.LoggerFromContext(ctx).Error(err, "")
		os.Exit(1)
	}
}
