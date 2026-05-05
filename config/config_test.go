package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestLoadConfig(t *testing.T) {
	// Setup viper test values
	viper.Set("target", "https://example.com")
	viper.Set("depth", 3)
	viper.Set("mobile-emulation", "iPhone 13")
	viper.Set("auth-json", "auth.json")
	viper.Set("record-video", true)
	viper.Set("fast-mode", true)

	// Clean up after test
	defer viper.Reset()

	cfg := LoadConfig()

	if cfg.Target != "https://example.com" {
		t.Errorf("Expected Target to be 'https://example.com', got '%s'", cfg.Target)
	}

	if cfg.Depth != 3 {
		t.Errorf("Expected Depth to be 3, got %d", cfg.Depth)
	}

	if cfg.MobileEmulation != "iPhone 13" {
		t.Errorf("Expected MobileEmulation to be 'iPhone 13', got '%s'", cfg.MobileEmulation)
	}

	if cfg.AuthJSON != "auth.json" {
		t.Errorf("Expected AuthJSON to be 'auth.json', got '%s'", cfg.AuthJSON)
	}

	if !cfg.RecordVideo {
		t.Errorf("Expected RecordVideo to be true, got %v", cfg.RecordVideo)
	}

	if !cfg.FastMode {
		t.Errorf("Expected FastMode to be true, got %v", cfg.FastMode)
	}
}
