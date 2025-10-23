package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	// AWS Configuration
	AWSRegion  string `json:"aws_region"`
	AWSProfile string `json:"aws_profile"`

	// Execution Configuration
	MaxParallelExecutions int  `json:"max_parallel_executions"`
	AutoApprove           bool `json:"auto_approve"`
	SaveOutputs           bool `json:"save_outputs"`
	OutputDirectory       string `json:"output_directory"`

	// UI Configuration
	ShowStatusIcons bool `json:"show_status_icons"`
	ColorScheme     string `json:"color_scheme"` // "default", "light", "dark"
}

func DefaultConfig() *Config {
	return &Config{
		AWSRegion:             "us-east-2",
		AWSProfile:            "default",
		MaxParallelExecutions: 10,
		AutoApprove:           false,
		SaveOutputs:           true,
		OutputDirectory:       "terragrunt-outputs",
		ShowStatusIcons:       true,
		ColorScheme:           "default",
	}
}

func LoadConfig() (*Config, error) {
	configPaths := []string{
		".terragrunt-vision.json",
		filepath.Join(os.Getenv("HOME"), ".terragrunt-vision.json"),
		"/etc/terragrunt-vision/config.json",
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return loadConfigFromFile(path)
		}
	}

	// Return default config if no config file found
	return DefaultConfig(), nil
}

func loadConfigFromFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	config := DefaultConfig()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}

func (c *Config) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}
