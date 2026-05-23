package cmd

import (
	"github.com/spf13/cobra"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Continuous monitoring operations",
	Long:  `Subcommands for continuous port monitoring and metrics exporting.`,
}

func init() {
	RootCmd.AddCommand(monitorCmd)
}
