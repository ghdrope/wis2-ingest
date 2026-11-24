package main

import (
	"context"
	"fmt"
	"time"
	"wis2-ingest/internal/config"
	"wis2-ingest/internal/filesystem"
	"wis2-ingest/internal/metrics"
	"wis2-ingest/internal/mqtt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newDoctorCommand creates the "doctor" subcommand.
//
// Currently it performs basic health checks:
// 1. Configuration validation
// 2. Output directory existence and writability
// 3. MQTT connection validation
func newDoctorCommand() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check WIS2 ingest health",
		Long:  "Performs health checks for WIS2 ingest.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			// -------------------------------
			// 1. Load Configuration
			// -------------------------------
			cfg := config.NewIngestOptions()
			if err := cfg.Load(); err != nil {
				return fmt.Errorf("❌ configuration check failed: %w", err)
			}
			if verbose {
				fmt.Print(cfg.PrintVerbose())
			}
			fmt.Println("✅ Configuration check passed")

			// -------------------------------
			// 2. Check output directory
			// -------------------------------
			if err := filesystem.CheckDirWritable(cfg.OutputDir); err != nil {
				return fmt.Errorf("❌ output directory check failed: %w", err)
			}
			fmt.Println("✅ Output directory check passed")

			// -------------------------------
			// 3. Check MQTT connection
			// -------------------------------

			// Metrics
			metricsInstance, _, err := metrics.New("wis2-ingest-doctor")
			if err != nil {
				return err
			}

			mqttClient := mqtt.NewClient(cfg, zap.L(), metricsInstance)

			// To avoid hanging indefinitely a context with timeout is placed here
			ctx, cancel := context.WithTimeout(context.Background(), doctorMQTTTimeout)
			defer cancel()

			// Start MQTT connection asynchronously and wait for it
			mqttClient.Start(ctx)

			if err := waitConnected(ctx, mqttClient); err != nil {
				return fmt.Errorf("❌ MQTT connection check failed: %w", err)
			}

			fmt.Println("✅ MQTT connection check passed")

			fmt.Println("\n🚀 All checks passed")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Print configuration values")
	return cmd
}

// doctorMQTTTimeout bounds how long doctor waits for the MQTT connection.
var doctorMQTTTimeout = 5 * time.Second

// waitConnected polls c until it reports a connection or ctx is done.
func waitConnected(ctx context.Context, c interface{ Connected() bool }) error {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		if c.Connected() {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("not connected to MQTT broker: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
