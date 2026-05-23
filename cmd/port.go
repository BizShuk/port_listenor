package cmd

import (
	"github.com/spf13/cobra"
)

var portCmd = &cobra.Command{
	Use:   "port",
	Short: "Port operations",
	Long:  `Subcommands for checking specific ports.`,
}

func init() {
	RootCmd.AddCommand(portCmd)
}
