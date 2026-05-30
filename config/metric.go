package config

import (
	"context"
	"fmt"
	"time"

	gosdkmetric "github.com/bizshuk/gosdk/metric"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// PortStatus 代表連接埠檢查的結果狀態
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

var (
	// Cached meter instruments (created in init)
	statusGauge  metric.Float64Gauge
	latencyGauge metric.Float64Gauge
	initErr      error
)

func init() {
	// Pre-create meter instruments using global meter (binds to real provider via otel.SetMeterProvider later)
	meter := otel.Meter("port_listenor")

	statusGauge, initErr = meter.Float64Gauge(
		"port_check_status",
		metric.WithDescription("1 = port open, 0 = port closed"),
	)
	if initErr != nil {
		panic(fmt.Errorf("failed to create status gauge in init: %w", initErr))
	}

	latencyGauge, initErr = meter.Float64Gauge(
		"port_check_latency_ms",
		metric.WithDescription("Port check latency in milliseconds"),
	)
	if initErr != nil {
		panic(fmt.Errorf("failed to create latency gauge in init: %w", initErr))
	}

	gosdkmetric.InitMeterProvider(context.Background())
}

// UpdateStatuses records the latest port statuses to OpenTelemetry metrics.
func UpdateStatuses(ctx context.Context, statuses []PortStatus) {
	for _, s := range statuses {
		val := 0.0
		if s.IsOpen {
			val = 1.0
		}
		pid := s.PID
		if pid == "" {
			pid = "unknown"
		}
		procName := s.ProcessName
		if procName == "" {
			procName = "unknown"
		}
		statusGauge.Record(ctx, val, metric.WithAttributes(
			attribute.String("host", s.Host),
			attribute.String("port", fmt.Sprintf("%d", s.Port)),
			attribute.String("service", s.Service),
			attribute.String("domain", s.Domain),
			attribute.String("pid", pid),
			attribute.String("process_name", procName),
		))

		latencyGauge.Record(ctx, s.LatencyMs, metric.WithAttributes(
			attribute.String("host", s.Host),
			attribute.String("port", fmt.Sprintf("%d", s.Port)),
			attribute.String("service", s.Service),
		))
	}
}
