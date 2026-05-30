package svc

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/bizshuk/port_listenor/config"
)

// RunMonitor 啟動持續監控循環與指標服務
func RunMonitor(ctx context.Context) error {
	globalConfig := config.Get()

	if err := InitMeterProvider(ctx, globalConfig.MimirURL); err != nil {
		log.Printf("Warning: Failed to initialize OpenTelemetry: %v", err)
	} else {
		defer func() {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			if err := ShutdownOTel(shutdownCtx); err != nil {
				log.Printf("Error shutting down MeterProvider: %v", err)
			}
		}()
	}

	timeout := ParseDuration(globalConfig.Timeout)
	interval := ParseDuration(globalConfig.CheckInterval)
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			log.Println("Received shutdown signal, shutting down gracefully...")
			return nil
		default:
		}

		var currentStatuses []PortStatus
		var currentLock sync.Mutex

		for _, entry := range globalConfig.Ports {
			wg.Add(1)
			go func(e config.PortEntry) {
				defer wg.Done()
				status := CheckPortWithProcess(e, timeout)

				currentLock.Lock()
				currentStatuses = append(currentStatuses, status)
				currentLock.Unlock()
			}(entry)
		}

		wg.Wait()

		sort.Slice(currentStatuses, func(i, j int) bool {
			return currentStatuses[i].Port < currentStatuses[j].Port
		})

		UpdateStatuses(ctx, currentStatuses)

		RenderDashboard(currentStatuses, interval)

		select {
		case <-time.After(interval):
		case <-ctx.Done():
			log.Println("Received shutdown signal, shutting down gracefully...")
			return nil
		}
	}
}

// RenderDashboard 渲染儀表板
func RenderDashboard(statuses []PortStatus, interval time.Duration) {
	fmt.Print("\033[H\033[2J")
	fmt.Println("================================================================================")
	fmt.Printf(" PORT HEALTH CHECKER - Last Update: %s (Interval: %v)\n",
		time.Now().Format("2006-01-02 15:04:05"), interval)
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
