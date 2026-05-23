package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/bizshuk/port_listenor/svc"
	"github.com/spf13/cobra"
)

var (
	monitorInterval string
	monitorTimeout  string
	metricsPort     int
	mimirEndpoint   string
)

var monitorPortCmd = &cobra.Command{
	Use:   "port",
	Short: "Start continuous port health monitoring",
	Long:  `Run in daemon mode to check port status periodically and expose metrics.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		cfg := &svc.MonitorConfig{
			Interval:    monitorInterval,
			Timeout:     monitorTimeout,
			MetricsPort: metricsPort,
			MimirURL:    mimirEndpoint,
		}

		if !cmd.Flags().Changed("metrics-port") {
			cfg.MetricsPort = 0 // 代表使用設定檔的值
		}

		return svc.RunMonitor(ctx, cfg)
	},
}

func init() {
	monitorPortCmd.Flags().StringVar(&monitorInterval, "interval", "", "check interval (e.g. 10s, 1m)")
	monitorPortCmd.Flags().StringVar(&monitorTimeout, "timeout", "", "connection timeout (e.g. 2s, 5s)")
	monitorPortCmd.Flags().IntVar(&metricsPort, "metrics-port", 0, "prometheus metrics port")
	monitorPortCmd.Flags().StringVar(&mimirEndpoint, "mimir-endpoint", "", "OTLP mimir endpoint URL")
	monitorCmd.AddCommand(monitorPortCmd)
}
