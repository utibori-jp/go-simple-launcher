package config

import (
	"app-launcher/logger"
	"encoding/json"
	"fmt"
	"os"
)

// Command represents a single command configuration with path and arguments
type Command struct {
	Path string   `json:"path"`
	Args []string `json:"args"`
}

// Config represents the root configuration structure
type Config struct {
	Commands map[string]Command `json:"commands"`
}

// ConfigManager handles loading and accessing configuration
type ConfigManager struct {
	configPath string
	commands   map[string]Command
}

// NewConfigManager creates a new ConfigManager with the specified config file path
func NewConfigManager(configPath string) (*ConfigManager, error) {
	if configPath == "" {
		err := fmt.Errorf("config path cannot be empty")
		logger.Error("Failed to create ConfigManager: %v", err)
		return nil, err
	}

	logger.Info("Creating ConfigManager with path: %s", configPath)
	return &ConfigManager{
		configPath: configPath,
		commands:   make(map[string]Command),
	}, nil
}

// Load reads and parses the JSON configuration file
func (c *ConfigManager) Load() error {
	logger.Info("Loading configuration from: %s", c.configPath)

	// Read the configuration file
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		logger.Error("Failed to read configuration file '%s': %v", c.configPath, err)
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		logger.Error("Failed to parse configuration file '%s': %v", c.configPath, err)
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate and store commands
	if cfg.Commands == nil {
		err := fmt.Errorf("configuration must contain 'commands' field")
		logger.Error("Invalid configuration structure in '%s': %v", c.configPath, err)
		return err
	}

	for name, cmd := range cfg.Commands {
		// Validate command name
		if name == "" {
			err := fmt.Errorf("command name cannot be empty")
			logger.Error("Configuration validation failed: %v", err)
			return err
		}

		// Validate required fields
		if cmd.Path == "" {
			err := fmt.Errorf("command '%s' must have a non-empty path", name)
			logger.Error("Configuration validation failed: %v", err)
			return err
		}

		// Args can be nil or empty, but if present must be a valid slice
		if cmd.Args == nil {
			cmd.Args = []string{}
		}

		c.commands[name] = cmd
	}

	logger.Info("Successfully loaded %d commands from configuration", len(c.commands))
	return nil
}

// GetCommand retrieves a command by name with O(1) lookup
func (c *ConfigManager) GetCommand(name string) (Command, bool) {
	cmd, exists := c.commands[name]
	if !exists {
		logger.Warn("Command lookup failed: '%s' not found in configuration", name)
	}
	return cmd, exists
}
