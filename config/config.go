package config

import (
	"fmt"

	"github.com/bizshuk/gosdk/config"
	"github.com/bizshuk/gosdk/log"
	sdkutils "github.com/bizshuk/gosdk/utils"
	"github.com/spf13/viper"
)

var defaultJson = `{
  "check_interval": "30s",
  "timeout": "5s",
  "metrics_port": 10235,
  "mimir_url": "http://localhost:9009/api/v1/push",
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

const SETTINGS_PATH = "~/.config/port_listenor"

// 全域設定實例
var globalSettings *Settings

// Settings 定義所有設定項目
type Settings struct {
	CheckInterval string      `mapstructure:"check_interval"`
	Timeout       string      `mapstructure:"timeout"`
	MetricsPort   int         `mapstructure:"metrics_port"`
	MimirURL      string      `mapstructure:"mimir_url"`
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
	log.Info("1")
	if globalSettings == nil {
		log.Info("2")
		if err := Default(); err != nil {
			panic(fmt.Sprintf("failed to load config: %v", err))
		}
	}
	log.Info("3", globalSettings)
	return globalSettings
}

func Default() error {
	err := sdkutils.CreateIfNotExist(SETTINGS_PATH, defaultJson)
	if err != nil {
		return err
	}
	log.Info("4")
	config.DefaultWithDir(SETTINGS_PATH)

	// 將 viper 內容解碼到 Settings 結構
	globalSettings = &Settings{}
	if err := viper.Unmarshal(globalSettings); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return nil
}
