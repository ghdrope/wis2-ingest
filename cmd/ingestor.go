package main

import (
	"wis2-ingest/internal/config"
	"wis2-ingest/internal/mqtt"
	"wis2-ingest/pkg/filesystem"
	"wis2-ingest/pkg/health"
	"wis2-ingest/pkg/utils"

	"github.com/akuity/kargo/pkg/logging"
	"github.com/spf13/cobra"
)

// newIngestCommand creates the "ingest" subcommand.
//
// The ingest command is responsible for initializing configuration,
// setting up logging, and starting the WIS2 data ingestion workflow.
// Configuration is loaded from environment variables via IngestOptions.
func newIngestorCommand() *cobra.Command {

	// During startup we enforce an info-level logger to ensure
	// that important initialization messages are always visible.
	_, format := utils.GetLogVars()

	bootstrapLogger := logging.NewLoggerOrDie(logging.InfoLevel, format)

	cmd := &cobra.Command{
		Use: "ingestor",
		RunE: func(cmd *cobra.Command, _ []string) error {

			bootstrapLogger.Info("Starting WIS2 Ingest")

			cfg := &config.IngestOptions{}

			if err := cfg.Load(); err != nil {
				return err
			}
			bootstrapLogger.Info("Configuration loaded",
				"host", cfg.Host,
				"port", cfg.Port,
				"topics", cfg.Topics,
				"output_dir", cfg.OutputDir)

			logLevel, logFormat := utils.GetLogVars()
			runtimeLogger := logging.NewLoggerOrDie(logLevel, logFormat)

			if err := filesystem.EnsureDirs(cfg.OutputDir); err != nil {
				return err
			}
			runtimeLogger.Info("Output directory created",
				"output_dir", cfg.OutputDir)

			mqttClient := mqtt.NewClient(cfg, runtimeLogger)
			if err := mqttClient.ConnectAndSubscribe(cmd.Context()); err != nil {
				return err
			}

			health.StartProbes(8080,
				func() bool { return mqttClient.Connected() },
				func() bool { return mqttClient.Connected() },
			)

			runtimeLogger.Info("MQTT client running. Waiting for messages...")

			ctx := cmd.Context()
			<-ctx.Done()

			runtimeLogger.Info("Shutting down WIS2 Ingest...")

			return nil
		},
	}

	return cmd
}
