package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/bizshuk/port_listenor/svc"
)

var (
	cfgFile string
	config  svc.Config
)

var RootCmd = &cobra.Command{
	Use:   "port-checker",
	Short: "Port Health Checker CLI",
	Long:  `A CLI tool to check the status of specific ports and export metrics.`,
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "settings.json", "config file path")
}

func initConfig() {
	if cfgFile == "" {
		cfgFile = "settings.json"
	}
	cfg, err := svc.LoadConfig(cfgFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && cfgFile == "settings.json" {
			config = svc.Config{
				CheckInterval: "30s",
				Timeout:       "5s",
				MetricsPort:   10235,
			}
			return
		}
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	config = *cfg
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// GetGlobalConfig 提供給同套件的其他 command 檔案獲取全域設定
func GetGlobalConfig() *svc.Config {
	return &config
}
