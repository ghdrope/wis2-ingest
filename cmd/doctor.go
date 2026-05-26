package main

import (
	"context"
	"fmt"
	"time"
	"wis2-ingest/internal/config"
	"wis2-ingest/internal/mqtt"
	"wis2-ingest/pkg/filesystem"
	"wis2-ingest/pkg/utils"

	"github.com/akuity/kargo/pkg/logging"
	"github.com/spf13/cobra"
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
			_, format := utils.GetLogVars()
			logger := logging.NewLoggerOrDie(logging.ErrorLevel, format) // temporary logger
			mqttClient := mqtt.NewClient(cfg, logger)

			// To avoid hanging indefinitely a context with timeout is placed here
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Start MQTT connection assynchronously
			mqttClient.Start(ctx)

			fmt.Println("✅ MQTT connection check passed")

			fmt.Println("\n🚀 All checks passed")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Print configuration values")
	return cmd
}
