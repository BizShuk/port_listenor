package svc

import (
	"fmt"
	"net"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bizshuk/port_listenor/config"
)

func CheckPortWithProcess(entry config.PortEntry, timeout time.Duration) config.PortStatus {
	status := config.PortStatus{
		Host:    "localhost",
		Port:    entry.Port,
		Service: entry.Name,
		Domain:  "",
	}

	isOpen, latency, _ := checkPort(status.Host, entry.Port, timeout)
	status.IsOpen = isOpen
	status.LatencyMs = latency
	status.LastCheckTime = time.Now()

	if isOpen {
		pid, procName := getProcessInfo(entry.Port)
		status.PID = pid
		status.ProcessName = procName
	}

	return status
}

func checkPort(host string, port int, timeout time.Duration) (bool, float64, error) {
	start := time.Now()
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		elapsed := time.Since(start).Seconds() * 1000
		return false, elapsed, nil
	}
	defer conn.Close()
	elapsed := time.Since(start).Seconds() * 1000
	return true, elapsed, nil
}

func getProcessInfo(port int) (pid string, processName string) {
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	output, err := cmd.Output()
	if err != nil {
		return "", ""
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return "", ""
	}

	for i := 1; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		if len(fields) < 2 {
			continue
		}
		pidStr := fields[1]
		if pidStr == "" || pidStr == "PID" {
			continue
		}

		cmd = exec.Command("ps", "-p", pidStr, "-o", "comm=")
		out, err := cmd.Output()
		if err != nil {
			return pidStr, ""
		}
		procName := strings.TrimSpace(string(out))
		return pidStr, procName
	}
	return "", ""
}

func ParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 5 * time.Second
	}
	return d
}

// CheckPorts 併發檢查多個連接埠，並依連接埠號碼排序回傳結果
func CheckPorts(entries []config.PortEntry, timeout time.Duration) []config.PortStatus {
	var results []config.PortStatus
	var resultsLock sync.Mutex
	var wg sync.WaitGroup

	for _, entry := range entries {
		wg.Add(1)
		go func(e config.PortEntry) {
			defer wg.Done()
			status := CheckPortWithProcess(e, timeout)
			resultsLock.Lock()
			results = append(results, status)
			resultsLock.Unlock()
		}(entry)
	}
	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		return results[i].Port < results[j].Port
	})

	return results
}

// ResolvePorts 解析傳入的 port 列表，若為空則從設定檔讀取，並回傳 PortEntry 列表、Timeout 期間與錯誤。
func ResolvePorts(ports []int) ([]config.PortEntry, time.Duration, error) {
	globalConfig := config.Get()
	timeout := ParseDuration(globalConfig.Timeout)

	var entries []config.PortEntry
	if len(ports) > 0 {
		for _, port := range ports {
			entries = append(entries, config.PortEntry{
				Port: port,
				Name: fmt.Sprintf("port-%d", port),
			})
		}
	} else {
		entries = globalConfig.Ports
	}

	if len(entries) == 0 {
		return nil, 0, fmt.Errorf("no ports specified in config or arguments")
	}

	return entries, timeout, nil
}
