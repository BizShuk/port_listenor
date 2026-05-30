package config

import (
	_ "embed"
	"fmt"

	"github.com/bizshuk/gosdk/config"
	"github.com/spf13/viper"
)

//go:embed default_settings.json
var defaultSettingsJSON string

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
	if globalSettings == nil {
		if err := Default(); err != nil {
			panic(fmt.Sprintf("failed to load config: %v", err))
		}
	}
	return globalSettings
}

func Default() error {
	config.Default(config.WithAppName("port_listenor"), config.WithDefaultValue(defaultSettingsJSON))
	// 將 viper 內容解碼到 Settings 結構
	globalSettings = &Settings{}
	if err := viper.Unmarshal(globalSettings); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return nil
}
