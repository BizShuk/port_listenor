package svc

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/bizshuk/port_listenor/config"
)

// monitorInterval 套件級變數，記錄監控間隔時間（預設為 0）
var monitorInterval time.Duration = 0

// RunMonitor 啟動持續監控循環與指標服務
func RunMonitor(ctx context.Context, entries []config.PortEntry, timeout time.Duration) error {
	globalConfig := config.Get()
	monitorInterval = parseDuration(globalConfig.CheckInterval)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err := RunOneTimeCheck(ctx, entries, timeout); err != nil {
			return err
		}

		select {
		case <-time.After(monitorInterval):
		case <-ctx.Done():
			return nil
		}
	}
}

// RunOneTimeCheck 執行單次檢查邏輯，印出表格結果。
func RunOneTimeCheck(ctx context.Context, entries []config.PortEntry, timeout time.Duration) error {
	results := CheckPorts(entries, timeout)
	config.UpdateStatuses(ctx, results)
	RenderDashboard(results)
	return nil
}

// RenderDashboard 渲染儀表板
func RenderDashboard(statuses []config.PortStatus) {
	fmt.Print("\033[H\033[2J")
	fmt.Println("================================================================================")
	fmt.Printf(" PORT HEALTH CHECKER - Last Update: %s (Interval: %v)\n",
		time.Now().Format("2006-01-02 15:04:05"), monitorInterval)
	fmt.Println("================================================================================")

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 3, ' ', 0)
	fmt.Fprintln(w, "PORT\tSERVICE\tSTATUS\tLATENCY\tPID\tPROCESS NAME")

	for _, s := range statuses {
		statusStr := "\033[31mCLOSED\033[0m"
		if s.IsOpen {
			statusStr = "\033[32mOPEN\033[0m"
		}
		latencyStr := "-"
		if s.IsOpen {
			latencyStr = fmt.Sprintf("%.2fms", s.LatencyMs)
		}
		pidStr := "-"
		if s.IsOpen && s.PID != "" {
			pidStr = s.PID
		}
		procStr := "-"
		if s.IsOpen && s.ProcessName != "" {
			procStr = s.ProcessName
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			s.Port, s.Service, statusStr, latencyStr, pidStr, procStr)
	}

	w.Flush()
	fmt.Println("================================================================================")
}
