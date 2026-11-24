package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
	"wis2-ingest/internal/config"
	"wis2-ingest/internal/filesystem"
	"wis2-ingest/internal/health"
	"wis2-ingest/internal/metrics"
	"wis2-ingest/internal/mqtt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newIngestCommand creates the "ingest" subcommand.
//
// The ingest command is responsible for initializing configuration,
// setting up logging, and starting the WIS2 data ingestion workflow.
// Configuration is loaded from environment variables via IngestOptions.
func newIngestorCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:  "ingestor",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {

			logger := zap.L()

			logger.Info("Starting WIS2 Ingest")

			cfg := config.NewIngestOptions()
			if err := cfg.Load(); err != nil {
				return err
			}

			logger.Info("Configuration loaded")
			logger.Info(cfg.PrintVerbose())

			if err := filesystem.EnsureDirs(cfg.OutputDir); err != nil {
				return err
			}
			logger.Info("Output directory created", zap.String("output_dir", cfg.OutputDir))

			mux := http.NewServeMux()

			// Metrics
			metricsInstance, metricsHandler, err := metrics.New("wis2-ingest")
			if err != nil {
				return err
			}
			mux.Handle("/metrics", metricsHandler)

			// MQTT
			mqttClient := mqtt.NewClient(cfg, logger, metricsInstance)
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
			serverErr := make(chan error, 1)
			go func() {
				logger.Info("http server running", zap.String("addr", addr))
				if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					serverErr <- err
				}
			}()

			logger.Info("MQTT client running. Waiting for messages...")

			// Without the HTTP server there are no probes or metrics, so a
			// failure (e.g. port already in use) stops the ingestor.
			select {
			case <-cmd.Context().Done():
			case err := <-serverErr:
				return fmt.Errorf("http server failed: %w", err)
			}

			logger.Info("Shutting down WIS2 Ingest...")

			// cmd.Context() is already cancelled here, so the shutdown gets
			// its own timeout to let in-flight requests finish.
			ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				logger.Error("http shutdown failed", zap.Error(err))
			}

			return nil
		},
	}

	return cmd
}

// shutdownTimeout bounds how long the HTTP server waits for in-flight
// requests on shutdown.
const shutdownTimeout = 10 * time.Second
