package config

import (
	_ "embed"
	"fmt"

	"github.com/bizshuk/gosdk/config"
	"github.com/bizshuk/gosdk/log"
	sdkutils "github.com/bizshuk/gosdk/utils"
	"github.com/spf13/viper"
)

//go:embed default_settings.json
var defaultJson string

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
	config.DefaultWithDir(SETTINGS_PATH)

	// 將 viper 內容解碼到 Settings 結構
	globalSettings = &Settings{}
	if err := viper.Unmarshal(globalSettings); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return nil
}
