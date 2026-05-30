package cmd

import (
	"fmt"
	"os"

	"github.com/bizshuk/port_listenor/config"
	"github.com/spf13/cobra"
)

var cfgFile string

// RootCmd 是CLI的根命令
var RootCmd = &cobra.Command{
	Use:   "port",
	Short: "Port Health Checker CLI",
	Long:  `A CLI tool to check the status of specific ports and export metrics.`,
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig 初始化全域設定
func initConfig() {
	// 載入設定（會自動處理設定檔、環境變數、預設值）
	if err := config.Default(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
}

// Execute 執行根命令
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
