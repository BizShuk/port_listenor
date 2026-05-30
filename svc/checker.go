package svc

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/bizshuk/port_listenor/config"
)

type PortStatus struct {
	Host          string
	Port          int
	Service       string
	Domain        string
	IsOpen        bool
	LatencyMs     float64
	PID           string
	ProcessName   string
	LastCheckTime time.Time
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

func CheckPortWithProcess(entry config.PortEntry, timeout time.Duration) PortStatus {
	status := PortStatus{
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

func ParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 5 * time.Second
	}
	return d
}
