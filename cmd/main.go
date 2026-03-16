package main

import (
	"os"

	"github.com/akuity/kargo/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	// Setup a context that is automatically cancelled on SIGINT/SIGTERM.
	ctx := signals.SetupSignalHandler()

	// Execute the CLI root command.
	if err := Execute(ctx); err != nil {
		logging.LoggerFromContext(ctx).Error(err, "")
		os.Exit(1)
	}
}
