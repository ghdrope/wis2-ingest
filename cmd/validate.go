package main

import (
	"fmt"
	"wis2-ingest/internal/config"

	"github.com/spf13/cobra"
)

// newValidateCommand creates the "validate" subcommand.
//
// It verifies that the ingest configuration is correctly defined and
// optionally points all configuration values if --verbose is passed.
func newValidateCommand() *cobra.Command {

	var verbose bool

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate WIS2 configuration",
		Long:  "Validates the WIS2 ingest configuration loaded from environment variables.",
		RunE: func(cmd *cobra.Command, args []string) error {

			// Load configuration
			cfg := &config.IngestOptions{}
			if err := cfg.Load(); err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// If verbose, print all config values
			if verbose {
				fmt.Println("Configuration values:")
				fmt.Printf("  Host      : %s\n", cfg.Host)
				fmt.Printf("  Port      : %s\n", cfg.Port)
				fmt.Printf("  Topics    : %v\n", cfg.Topics)
				fmt.Printf("  OutputDir : %s\n", cfg.OutputDir)
				fmt.Printf("  Username  : %s\n", cfg.Username)
			}

			fmt.Println("✅ Configuration validation successful")
			return nil
		},
	}

	// Add the --verbose flag
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Print configuration values")
	return cmd
}
