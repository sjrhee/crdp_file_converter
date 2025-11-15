package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Processing ProcessingConfig `yaml:"processing"`
	Output     OutputConfig     `yaml:"output"`
}

// ServerConfig represents CRDP server settings
type ServerConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Policy  string `yaml:"policy"`
	Timeout int    `yaml:"timeout"`
}

// ProcessingConfig represents file processing settings
type ProcessingConfig struct {
	Delimiter        string `yaml:"delimiter"`
	Column           int    `yaml:"column"`
	BatchSize        int    `yaml:"batch_size"`
	SkipHeader       bool   `yaml:"skip_header"`
	ParallelWorkers  int    `yaml:"parallel_workers"`
}

// OutputConfig represents output settings
type OutputConfig struct {
	File string `yaml:"file"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:    "localhost",
			Port:    8080,
			Policy:  "P03",
			Timeout: 30,
		},
		Processing: ProcessingConfig{
			Delimiter:       ",",
			Column:          0,
			BatchSize:       100,
			SkipHeader:      false,
			ParallelWorkers: 1,
		},
		Output: OutputConfig{
			File: "",
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
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", c.Server.Port)
	}

	if c.Server.Timeout <= 0 {
		return fmt.Errorf("invalid timeout: %d (must be positive)", c.Server.Timeout)
	}

	if c.Processing.Column < 0 {
		return fmt.Errorf("invalid column: %d (must be non-negative)", c.Processing.Column)
	}

	if c.Processing.BatchSize <= 0 {
		return fmt.Errorf("invalid batch_size: %d (must be positive)", c.Processing.BatchSize)
	}

	if c.Processing.ParallelWorkers < 0 {
		return fmt.Errorf("invalid parallel_workers: %d (must be non-negative)", c.Processing.ParallelWorkers)
	}

	return nil
}
