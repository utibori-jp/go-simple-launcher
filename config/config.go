package config

import (
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
		return nil, fmt.Errorf("config path cannot be empty")
	}
	
	return &ConfigManager{
		configPath: configPath,
		commands:   make(map[string]Command),
	}, nil
}

// Load reads and parses the JSON configuration file
func (c *ConfigManager) Load() error {
	// Read the configuration file
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate and store commands
	if cfg.Commands == nil {
		return fmt.Errorf("configuration must contain 'commands' field")
	}

	for name, cmd := range cfg.Commands {
		// Validate command name
		if name == "" {
			return fmt.Errorf("command name cannot be empty")
		}

		// Validate required fields
		if cmd.Path == "" {
			return fmt.Errorf("command '%s' must have a non-empty path", name)
		}

		// Args can be nil or empty, but if present must be a valid slice
		if cmd.Args == nil {
			cmd.Args = []string{}
		}

		c.commands[name] = cmd
	}

	return nil
}

// GetCommand retrieves a command by name with O(1) lookup
func (c *ConfigManager) GetCommand(name string) (Command, bool) {
	cmd, exists := c.commands[name]
	return cmd, exists
}
