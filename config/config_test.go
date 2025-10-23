package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.AWSRegion != "us-east-2" {
		t.Errorf("Expected default AWS region to be us-east-2, got %s", cfg.AWSRegion)
	}

	if cfg.MaxParallelExecutions != 10 {
		t.Errorf("Expected default max parallel executions to be 10, got %d", cfg.MaxParallelExecutions)
	}

	if !cfg.SaveOutputs {
		t.Error("Expected SaveOutputs to be true by default")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Create a temporary file for testing
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")

	// Create a custom config
	originalConfig := &Config{
		AWSRegion:             "us-west-1",
		AWSProfile:            "production",
		MaxParallelExecutions: 5,
		AutoApprove:           true,
		SaveOutputs:           false,
		OutputDirectory:       "/tmp/outputs",
		ShowStatusIcons:       false,
		ColorScheme:           "dark",
	}

	// Save the config
	if err := originalConfig.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load the config back
	loadedConfig, err := loadConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify all fields match
	if loadedConfig.AWSRegion != originalConfig.AWSRegion {
		t.Errorf("AWSRegion mismatch: got %s, want %s", loadedConfig.AWSRegion, originalConfig.AWSRegion)
	}

	if loadedConfig.AWSProfile != originalConfig.AWSProfile {
		t.Errorf("AWSProfile mismatch: got %s, want %s", loadedConfig.AWSProfile, originalConfig.AWSProfile)
	}

	if loadedConfig.MaxParallelExecutions != originalConfig.MaxParallelExecutions {
		t.Errorf("MaxParallelExecutions mismatch: got %d, want %d", loadedConfig.MaxParallelExecutions, originalConfig.MaxParallelExecutions)
	}

	if loadedConfig.AutoApprove != originalConfig.AutoApprove {
		t.Errorf("AutoApprove mismatch: got %t, want %t", loadedConfig.AutoApprove, originalConfig.AutoApprove)
	}
}

func TestLoadConfigNoFile(t *testing.T) {
	// Change to a temp directory where no config exists
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Load config (should return default)
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Should return default config
	defaultCfg := DefaultConfig()
	if cfg.AWSRegion != defaultCfg.AWSRegion {
		t.Errorf("Expected default config, got custom config")
	}
}
