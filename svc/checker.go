package svc

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/bizshuk/port_listenor/config"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
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

type Checker struct {
	Config           *config.Settings
	StatusLock       sync.RWMutex
	LatestStatuses   []PortStatus
	PortStatusGauge  *prometheus.GaugeVec
	PortLatencyGauge *prometheus.GaugeVec
	PortInfoGauge    *prometheus.GaugeVec
	Registry         *prometheus.Registry
	OtelProvider     *sdkmetric.MeterProvider
}

func NewChecker() *Checker {
	reg := prometheus.NewRegistry()

	statusGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "port_check_status",
			Help: "1 = port open, 0 = port closed",
		},
		[]string{"host", "port", "service", "domain", "pid", "process_name"},
	)

	latencyGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "port_check_latency_ms",
			Help: "Port check latency in milliseconds",
		},
		[]string{"host", "port", "service"},
	)

	infoGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "port_check_process_info",
			Help: "Process info for open ports (value = 1 if present)",
		},
		[]string{"host", "port", "service", "pid", "process_name"},
	)

	reg.MustRegister(statusGauge)
	reg.MustRegister(latencyGauge)
	reg.MustRegister(infoGauge)

	return &Checker{
		Config:           config.Get(),
		PortStatusGauge:  statusGauge,
		PortLatencyGauge: latencyGauge,
		PortInfoGauge:    infoGauge,
		Registry:         reg,
	}
}

func (c *Checker) InitOTel(ctx context.Context) error {
	u, err := url.Parse(c.Config.MimirURL)
	if err != nil {
		return fmt.Errorf("failed to parse mimir url: %w", err)
	}

	endpoint := u.Host
	path := u.Path
	if path == "" || path == "/" {
		path = "/otlp/v1/metrics"
	}

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithURLPath(path),
	}
	if u.Scheme != "https" {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("port-checker"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(10*time.Second))),
		sdkmetric.WithResource(res),
	)

	otel.SetMeterProvider(provider)
	c.OtelProvider = provider

	meter := otel.Meter("port-checker")
	return c.registerOTelMetrics(meter)
}

func (c *Checker) registerOTelMetrics(meter metric.Meter) error {
	_, err := meter.Float64ObservableGauge(
		"port_check_status",
		metric.WithDescription("1 = port open, 0 = port closed"),
		metric.WithFloat64Callback(func(ctx context.Context, observer metric.Float64Observer) error {
			c.StatusLock.RLock()
			defer c.StatusLock.RUnlock()
			for _, status := range c.LatestStatuses {
				val := 0.0
				if status.IsOpen {
					val = 1.0
				}
				pid := status.PID
				if pid == "" {
					pid = "unknown"
				}
				procName := status.ProcessName
				if procName == "" {
					procName = "unknown"
				}
				observer.Observe(val, metric.WithAttributes(
					attribute.String("host", status.Host),
					attribute.String("port", fmt.Sprintf("%d", status.Port)),
					attribute.String("service", status.Service),
					attribute.String("domain", status.Domain),
					attribute.String("pid", pid),
					attribute.String("process_name", procName),
				))
			}
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Float64ObservableGauge(
		"port_check_latency_ms",
		metric.WithDescription("Port check latency in milliseconds"),
		metric.WithFloat64Callback(func(ctx context.Context, observer metric.Float64Observer) error {
			c.StatusLock.RLock()
			defer c.StatusLock.RUnlock()
			for _, status := range c.LatestStatuses {
				observer.Observe(status.LatencyMs, metric.WithAttributes(
					attribute.String("host", status.Host),
					attribute.String("port", fmt.Sprintf("%d", status.Port)),
					attribute.String("service", status.Service),
				))
			}
			return nil
		}),
	)
	return err
}

func (c *Checker) ShutdownOTel(ctx context.Context) error {
	if c.OtelProvider != nil {
		return c.OtelProvider.Shutdown(ctx)
	}
	return nil
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

func (c *Checker) CheckPortWithProcess(entry config.PortEntry, timeout time.Duration) PortStatus {
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

func (c *Checker) UpdateMetrics(status PortStatus) {
	portStr := fmt.Sprintf("%d", status.Port)
	var pid, procName string
	if status.IsOpen {
		pid = status.PID
		procName = status.ProcessName
	}
	if pid == "" {
		pid = "unknown"
	}
	if procName == "" {
		procName = "unknown"
	}

	val := 0.0
	if status.IsOpen {
		val = 1.0
	}

	c.PortStatusGauge.WithLabelValues(status.Host, portStr, status.Service, status.Domain, pid, procName).Set(val)
	c.PortLatencyGauge.WithLabelValues(status.Host, portStr, status.Service).Set(status.LatencyMs)

	if status.IsOpen {
		c.PortInfoGauge.WithLabelValues(status.Host, portStr, status.Service, pid, procName).Set(1.0)
	}
}

func ParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 5 * time.Second
	}
	return d
}
