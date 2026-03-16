package main

import (
	"context"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command for the WIS2 CLI.
//
// It acts as the entry point for all subcommands. If executed
// without a subcommand, it displays the CLI help.
var rootCmd = &cobra.Command{
	Use:   "wis2",
	Short: "WIS2 CLI",
	Long:  `CLI for WIS2.`,
	Run: func(cmd *cobra.Command, args []string) {

		// Display help when no subcommand is provided.
		cmd.HelpFunc()(cmd, args)
	},
}

// Execute initializes the CLI and executes the root command.
//
// The provided context is propagated to Cobra commands and allows
// graceful shutdown handling when receiving termination signals.
func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

// init registers CLI subcommands.
func init() {
	rootCmd.AddCommand(newIngestCommand())
}
