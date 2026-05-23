package svc

import (
	"fmt"
	"io"
	"sort"
	"sync"
	"text/tabwriter"

	"github.com/bizshuk/port_listenor/config"
)

// CheckConfig 定義單次檢查的參數
type CheckConfig struct {
	PortsToCheck []config.PortEntry
	TimeoutVal   string
	Writer       io.Writer
}

// RunOneTimeCheck 執行單次檢查邏輯，印出表格結果
func RunOneTimeCheck(cfg *CheckConfig, globalConfig *config.Settings) error {
	ports := cfg.PortsToCheck
	if len(ports) == 0 {
		ports = globalConfig.Ports
	}
	if len(ports) == 0 {
		return fmt.Errorf("no ports specified to check. Use -c to specify a config file or --ports to specify ports directly")
	}

	timeoutVal := globalConfig.Timeout
	if cfg.TimeoutVal != "" {
		timeoutVal = cfg.TimeoutVal
	}
	timeout := ParseDuration(timeoutVal)

	c := NewChecker(globalConfig)
	var results []PortStatus
	var resultsLock sync.Mutex
	var wg sync.WaitGroup

	for _, entry := range ports {
		wg.Add(1)
		go func(e config.PortEntry) {
			defer wg.Done()
			status := c.CheckPortWithProcess(e, timeout)
			resultsLock.Lock()
			results = append(results, status)
			resultsLock.Unlock()
		}(entry)
	}
	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		return results[i].Port < results[j].Port
	})

	w := tabwriter.NewWriter(cfg.Writer, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "PORT\tSERVICE\tSTATUS\tLATENCY\tPID\tPROCESS NAME")
	for _, s := range results {
		statusStr := "CLOSED"
		if s.IsOpen {
			statusStr = "OPEN"
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
	return nil
}
