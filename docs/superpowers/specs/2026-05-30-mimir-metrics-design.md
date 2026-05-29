# Mimir Metrics & Grafana Dashboard Design

## 1. Overview

Send port listening metrics (including closed ports) to Mimir via OpenTelemetry, and provide a Grafana dashboard for visualization.

## 2. Metrics Design

### New Metrics

| Metric Name | Type | Labels | Description |
|-------------|------|--------|-------------|
| `net_host_port` | Gauge | `host`, `port` | 1 = port open, 0 = port closed |
| `net_host_port_latency_ms` | Gauge | `host`, `port` | Port check latency in milliseconds |

### Label Standardization

- **`host`**: Target hostname (e.g., `localhost`)
- **`port`**: Port number (e.g., `8080`)
- **Removed**: `service`, `domain`, `pid`, `process_name` (not needed for Mimir/Grafana use case)

### Replacement

These two new metrics **replace** the existing Prometheus-only metrics:
- `port_check_status`
- `port_check_latency_ms`
- `port_check_process_info`

## 3. Code Changes

### `svc/checker.go`

1. **`NewChecker()`** — Register new metric names:
   - `net_host_port` GaugeVec with labels `{host, port}`
   - `net_host_port_latency_ms` GaugeVec with labels `{host, port}`

2. **`UpdateMetrics(status PortStatus)`** — Use simplified labels:
   - `net_host_port{host, port}` → set 1.0 or 0.0
   - `net_host_port_latency_ms{host, port}` → set latency value

3. **`registerOTelMetrics(meter metric.Meter)`** — Update OTel callbacks:
   - Observable gauge `net_host_port` with `{host, port}` attributes
   - Observable gauge `net_host_port_latency_ms` with `{host, port}` attributes

## 4. Dashboard Design

### File

`docs/dashboard.json` — Grafana JSON Dashboard model

### Panels

#### Panel 1: Host Overview (Stat)
- **Type**: Stat
- **Repeat**: Horizontal by host
- **Display**: Host name / IP

#### Panel 2: Port Status Table (Table)
- **Type**: Table
- **Repeat**: Horizontal by host
- **Columns**: Port | Status | Latency (ms)
- **Status column**: Color-coded — 1 = Green (Open), 0 = Red (Closed)
- **Repeat direction**: Horizontal

## 5. Configuration

No configuration changes required. The existing `config.PortEntry` structure is sufficient.

## 6. Files Summary

| File | Action |
|------|--------|
| `svc/checker.go` | Modify — update metric names and labels |
| `docs/dashboard.json` | Create — Grafana dashboard JSON |
