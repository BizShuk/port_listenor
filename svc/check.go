package svc

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"text/tabwriter"

	"github.com/bizshuk/port_listenor/config"
)

// RunOneTimeCheck 執行單次檢查邏輯，印出表格結果
func RunOneTimeCheck() error {
	globalConfig := config.Get()
	ports := globalConfig.Ports
	if len(ports) == 0 {
		return fmt.Errorf("no ports specified in config")
	}

	timeout := ParseDuration(globalConfig.Timeout)

	c := NewChecker()
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

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "PORT\tSERVICE\tSTATUS\tLATENCY\tPID\tPROCESS NAME")
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
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\n",
			s.Port, s.Service, statusStr, latencyStr, pidStr, procStr)
	}
	tw.Flush()
	return nil
}

// RunOneTimeCheckWithPorts 執行單次檢查，只檢查指定的連接埠
func RunOneTimeCheckWithPorts(ports []int) error {
	globalConfig := config.Get()
	timeout := ParseDuration(globalConfig.Timeout)

	c := NewChecker()
	var results []PortStatus
	var resultsLock sync.Mutex
	var wg sync.WaitGroup

	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			entry := config.PortEntry{Port: p, Name: fmt.Sprintf("port-%d", p)}
			status := c.CheckPortWithProcess(entry, timeout)
			resultsLock.Lock()
			results = append(results, status)
			resultsLock.Unlock()
		}(port)
	}
	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		return results[i].Port < results[j].Port
	})

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "PORT\tSERVICE\tSTATUS\tLATENCY\tPID\tPROCESS NAME")
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
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\n",
			s.Port, s.Service, statusStr, latencyStr, pidStr, procStr)
	}
	tw.Flush()
	return nil
}

// RunOneTimeCheckSimple 執行單次檢查，使用標準輸出
func RunOneTimeCheckSimple() error {
	return RunOneTimeCheck()
}
