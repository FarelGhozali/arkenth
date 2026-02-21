package config

import (
	"log"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Target          string
	Depth           int
	MobileEmulation string
	AuthJSON        string
	RecordVideo     bool
	FastMode        bool
}

// LoadConfig reads the configuration set by Cobra/Viper flags
func LoadConfig() *AppConfig {
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalf("Error reading config file: %v", err)
		}
	}

	return &AppConfig{
		Target:          viper.GetString("target"),
		Depth:           viper.GetInt("depth"),
		MobileEmulation: viper.GetString("mobile-emulation"),
		AuthJSON:        viper.GetString("auth-json"),
		RecordVideo:     viper.GetBool("record-video"),
		FastMode:        viper.GetBool("fast-mode"),
	}
}
