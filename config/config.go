package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bizshuk/gosdk/config"
	sdkutils "github.com/bizshuk/gosdk/utils"
	"github.com/bizshuk/port_listenor/svc"
	"github.com/spf13/viper"
)

var defaultJson = `{
  "check_interval": "30s",
  "timeout": "5s",
  "metrics_port": 10235,
  "mimir_endpoint": "",
  "log_level": "info",
  "ports": [
    { "port": 8080, "name": "web" },
    { "port": 8081, "name": "api" },
    { "port": 5432, "name": "postgres" },
    { "port": 6379, "name": "redis" },
    { "port": 6378, "name": "redis-cluster" },
    { "port": 9090, "name": "prometheus" },
    { "port": 9093, "name": "alertmanager" },
    { "port": 3000, "name": "grafana" },
    { "port": 3100, "name": "loki" },
    { "port": 9009, "name": "mimir" },
    { "port": 3200, "name": "tempo" },
    { "port": 22, "name": "ssh" }
  ]
}`

// 全域設定實例
var globalSettings *Settings

// Settings 定義所有設定項目
type Settings struct {
	CheckInterval string      `mapstructure:"check_interval"`
	Timeout       string      `mapstructure:"timeout"`
	MetricsPort   int         `mapstructure:"metrics_port"`
	MimirEndpoint string      `mapstructure:"mimir_endpoint"`
	LogLevel      string      `mapstructure:"log_level"`
	Ports         []PortEntry `mapstructure:"ports"`
}

// PortEntry 定義單一連接埠設定
type PortEntry struct {
	Port int    `mapstructure:"port" json:"port"`
	Name string `mapstructure:"name" json:"name"`
}

// Get 返回全域設定單例
// 首次調用時自動初始化
func Get() *Settings {
	if globalSettings == nil {
		if err := Default(); err != nil {
			panic(fmt.Sprintf("failed to load config: %v", err))
		}
	}
	return globalSettings
}

func Default() error {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, "config", "port_listenor")
	configFilePath := filepath.Join(configDir, "settings.json")

	err := sdkutils.CreateIfNotExist(configFilePath, defaultJson)
	if err != nil {
		return err
	}

	setDefaultSettings()
	config.DefaultWithDir("~/config/port_listenor")

	// 將 viper 內容解碼到 Settings 結構
	globalSettings = &Settings{}
	if err := viper.Unmarshal(globalSettings); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// setDefaultSettings 設定預設值
func setDefaultSettings() {
	viper.SetDefault("check_interval", "30s")
	viper.SetDefault("timeout", "5s")
	viper.SetDefault("metrics_port", 10235)
	viper.SetDefault("log_level", "info")
	viper.SetDefault("ports", []PortEntry{
		{Port: 8080, Name: "web"},
		{Port: 8081, Name: "api"},
		{Port: 5432, Name: "postgres"},
		{Port: 6379, Name: "redis"},
		{Port: 6378, Name: "redis-cluster"},
		{Port: 9090, Name: "prometheus"},
		{Port: 9093, Name: "alertmanager"},
		{Port: 3000, Name: "grafana"},
		{Port: 3100, Name: "loki"},
		{Port: 9009, Name: "mimir"},
		{Port: 3200, Name: "tempo"},
		{Port: 22, Name: "ssh"},
	})
}

// GetViper 返回 viper 實例，允許直接存取
func GetViper() *viper.Viper {
	return viper.GetViper()
}

// Reset 清除全域設定（主要用於測試）
func Reset() {
	globalSettings = nil
}

// ToSvcConfig 轉換為 svc.Config 格式
// 用於與 svc 套件保持相容
func (s *Settings) ToSvcConfig() *svc.Config {
	ports := make([]svc.PortEntry, len(s.Ports))
	for i, p := range s.Ports {
		ports[i] = svc.PortEntry{Port: p.Port, Name: p.Name}
	}
	return &svc.Config{
		CheckInterval: s.CheckInterval,
		Timeout:       s.Timeout,
		MetricsPort:   s.MetricsPort,
		MimirEndpoint: s.MimirEndpoint,
		LogLevel:      s.LogLevel,
		Ports:         ports,
	}
}
