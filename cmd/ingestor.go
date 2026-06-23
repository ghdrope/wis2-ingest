package main

import (
	"net/http"
	"os"
	"wis2-ingest/internal/config"
	"wis2-ingest/internal/filesystem"
	"wis2-ingest/internal/health"
	"wis2-ingest/internal/metrics"
	"wis2-ingest/internal/mqtt"
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
		Use:  "ingestor",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {

			bootstrapLogger.Info("Starting WIS2 Ingest")

			cfg := config.NewIngestOptions()
			if err := cfg.Load(); err != nil {
				return err
			}

			bootstrapLogger.Info("Configuration loaded")
			bootstrapLogger.Info(cfg.PrintVerbose()) // Print all config fields dynamically

			logLevel, logFormat := utils.GetLogVars()

			runtimeLogger := logging.NewLoggerOrDie(logLevel, logFormat)

			if err := filesystem.EnsureDirs(cfg.OutputDir); err != nil {
				return err
			}
			runtimeLogger.Info(
				"Output directory created",
				"output_dir", cfg.OutputDir,
			)

			mux := http.NewServeMux()

			// Metrics
			metricsInstance, metricsHandler, err := metrics.New("wis2-ingest")
			if err != nil {
				return err
			}
			mux.Handle("/metrics", metricsHandler)

			// MQTT
			mqttClient := mqtt.NewClient(cfg, runtimeLogger, metricsInstance)
			mqttClient.Start(cmd.Context()) // start asynchronously

			// Health probes
			health.RegisterProbeHandlers(mux, mqttClient)

			// Server
			port := os.Getenv("PORT")
			if port == "" {
				port = "8080"
			}
			addr := ":" + port
			server := &http.Server{
				Addr:    addr,
				Handler: mux,
			}

			// Server
			go func() {
				runtimeLogger.Info("http server running", "addr", addr)
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					runtimeLogger.Error(err, "http server failed")
				}
			}()
			defer func() {
				if err := server.Shutdown(cmd.Context()); err != nil {
					runtimeLogger.Error(err, "http shutdown failed")
				}
			}()

			runtimeLogger.Info("MQTT client running. Waiting for messages...")

			<-cmd.Context().Done()

			runtimeLogger.Info("Shutting down WIS2 Ingest...")

			return nil
		},
	}

	return cmd
}
