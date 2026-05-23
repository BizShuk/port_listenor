package config

import (
	"fmt"

	"github.com/bizshuk/gosdk/config"
	"github.com/bizshuk/port_listenor/svc"
	"github.com/spf13/viper"
)

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
