package executor

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"app-launcher/config"
)

// Executor handles command execution and application launching
type Executor struct {
	config *config.ConfigManager
}

// NewExecutor creates a new Executor with the specified ConfigManager
func NewExecutor(cfg *config.ConfigManager) *Executor {
	return &Executor{
		config: cfg,
	}
}

// Execute looks up a command by name and launches the corresponding application
// Returns an error if the command is not found or if the application fails to launch
func (e *Executor) Execute(commandName string) error {
	// Lookup command in configuration
	cmd, exists := e.config.GetCommand(commandName)
	if !exists {
		return fmt.Errorf("command '%s' not found", commandName)
	}

	// Normalize path for Windows (convert forward slashes to backslashes)
	normalizedPath := normalizePath(cmd.Path)

	// Create the command with arguments
	execCmd := exec.Command(normalizedPath, cmd.Args...)

	// Start the process without blocking (don't wait for it to complete)
	if err := execCmd.Start(); err != nil {
		return fmt.Errorf("failed to launch '%s': %w", commandName, err)
	}

	// Return immediately without waiting for the process to complete
	return nil
}

// normalizePath converts forward slashes to backslashes for Windows compatibility
func normalizePath(path string) string {
	// Replace forward slashes with backslashes
	normalized := strings.ReplaceAll(path, "/", string(filepath.Separator))
	
	// Use filepath.Clean to normalize the path further
	return filepath.Clean(normalized)
}
