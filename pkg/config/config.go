package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	API        APIConfig        `yaml:"api"`
	Protection ProtectionConfig `yaml:"protection"`
	File       FileConfig       `yaml:"file"`
	Batch      BatchConfig      `yaml:"batch"`
	Parallel   ParallelConfig   `yaml:"parallel"`
	Output     OutputConfig     `yaml:"output"`
}

// APIConfig represents CRDP API connection settings
type APIConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Timeout int    `yaml:"timeout"`
	TLS     bool   `yaml:"tls"`
}

// ProtectionConfig represents data protection settings
type ProtectionConfig struct {
	Policy string `yaml:"policy"`
}

// FileConfig represents file processing settings
type FileConfig struct {
	Delimiter  string `yaml:"delimiter"`
	Column     int    `yaml:"column"`
	SkipHeader bool   `yaml:"skip_header"`
}

// BatchConfig represents batch processing settings
type BatchConfig struct {
	Enabled bool `yaml:"enabled"`
	Size    int  `yaml:"size"`
}

// ParallelConfig represents parallel processing settings
type ParallelConfig struct {
	Workers int `yaml:"workers"`
}

// OutputConfig represents output settings
type OutputConfig struct {
	File        string `yaml:"file"`
	ShowProgress bool   `yaml:"show_progress"`
	ShowBody    bool   `yaml:"show_body"`
	Verbose     bool   `yaml:"verbose"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		API: APIConfig{
			Host:    "192.168.0.231",
			Port:    32082,
			Timeout: 30,
			TLS:     false,
		},
		Protection: ProtectionConfig{
			Policy: "P03",
		},
		File: FileConfig{
			Delimiter:  ",",
			Column:     0,
			SkipHeader: false,
		},
		Batch: BatchConfig{
			Enabled: true,
			Size:    100,
		},
		Parallel: ParallelConfig{
			Workers: 1,
		},
		Output: OutputConfig{
			File:        "",
			ShowProgress: true,
			ShowBody:    false,
			Verbose:     false,
		},
	}
}

// LoadConfig loads configuration from file
// If configPath is empty, tries default locations
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// If no path provided, try default locations
	if configPath == "" {
		configPath = findConfigFile()
	}

	// If still no path found, return default config
	if configPath == "" {
		return config, nil
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Read and parse config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// findConfigFile searches for config file in default locations
func findConfigFile() string {
	locations := []string{
		"config.yaml",
		"config.yml",
		filepath.Join(os.Getenv("HOME"), ".crdp", "config.yaml"),
		"/etc/crdp/config.yaml",
	}

	for _, loc := range locations {
		if loc == "" {
			continue
		}
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	return ""
}

// SaveConfig saves configuration to file
func (c *Config) SaveConfig(filePath string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.API.Port <= 0 || c.API.Port > 65535 {
		return fmt.Errorf("invalid api.port: %d (must be 1-65535)", c.API.Port)
	}

	if c.API.Timeout <= 0 {
		return fmt.Errorf("invalid api.timeout: %d (must be positive)", c.API.Timeout)
	}

	if c.File.Column < 0 {
		return fmt.Errorf("invalid file.column: %d (must be non-negative)", c.File.Column)
	}

	if c.Batch.Size <= 0 {
		return fmt.Errorf("invalid batch.size: %d (must be positive)", c.Batch.Size)
	}

	if c.Parallel.Workers < 0 {
		return fmt.Errorf("invalid parallel.workers: %d (must be non-negative)", c.Parallel.Workers)
	}

	return nil
}
