package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// RootCmd 是CLI的根命令
var RootCmd = &cobra.Command{
	Use:   "port",
	Short: "Port Health Checker CLI",
	Long:  `A CLI tool to check the status of specific ports and export metrics.`,
}

// Execute 執行根命令
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
