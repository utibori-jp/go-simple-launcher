package config

import (
	"app-launcher/logger"
	"encoding/json"
	"fmt"
	"os"
)

// Command represents a single command configuration with path and arguments.
//
// Configuration Format:
// The command structure maps a command name to an executable path and optional arguments.
//
// Example JSON:
//
//	{
//	  "commands": {
//	    "chrome": {
//	      "path": "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
//	      "args": []
//	    },
//	    "vscode": {
//	      "path": "C:\\Program Files\\Microsoft VS Code\\Code.exe",
//	      "args": ["-n"]
//	    }
//	  }
//	}
//
// Fields:
//   - Path: Absolute path to the executable. Forward slashes (/) are automatically
//     converted to backslashes (\) on Windows for compatibility.
//   - Args: Array of command-line arguments to pass to the application.
//     Can be empty ([]) if no arguments are needed.
type Command struct {
	Path string   `json:"path"` // Absolute path to executable
	Args []string `json:"args"` // Command-line arguments (can be empty)
}

// Config represents the root configuration structure.
//
// The configuration file must be valid JSON with a "commands" object at the root.
// Each key in "commands" is the command name that users will type in the launcher,
// and the value is a Command object specifying the executable path and arguments.
//
// Configuration File Location:
//   - Default: %APPDATA%\launcher\config.json
//   - Override with --config flag: launcher.exe --config="C:\path\to\config.json"
//
// Validation Rules:
//   - Command names must be non-empty strings
//   - Each command must have a non-empty "path" field
//   - The "args" field can be empty but must be present
//   - Duplicate command names are not allowed (enforced by JSON object structure)
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
