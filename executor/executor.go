package executor

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"app-launcher/config"
	"app-launcher/logger"
)

type ConfigProvider interface {
	GetCommand(name string) (config.Command, bool)
	Load() error
}

// Executor handles command execution and application launching
type Executor struct {
	config ConfigProvider
}

// NewExecutor creates a new Executor with the specified ConfigManager
func NewExecutor(cfg ConfigProvider) *Executor {
	return &Executor{
		config: cfg,
	}
}

// Execute looks up a command by name and launches the corresponding application
// Returns an error if the command is not found or if the application fails to launch
func (e *Executor) Execute(commandName string) error {
	logger.Info("Attempting to execute command: '%s'", commandName)

	// Lookup command in configuration
	cmd, exists := e.config.GetCommand(commandName)
	if !exists {
		err := fmt.Errorf("command '%s' not found", commandName)
		logger.Error("Command execution failed: %v", err)
		return err
	}

	// Normalize path for Windows (convert forward slashes to backslashes)
	normalizedPath := normalizePath(cmd.Path)
	logger.Info("Normalized path for '%s': %s (args: %v)", commandName, normalizedPath, cmd.Args)

	// Create the command with arguments
	execCmd := exec.Command(normalizedPath, cmd.Args...)

	// Start the process without blocking (don't wait for it to complete)
	if err := execCmd.Start(); err != nil {
		// Provide detailed error information
		detailedErr := fmt.Errorf("failed to launch '%s': %w", commandName, err)
		logger.Error("Application launch failed for '%s' (path: %s): %v", commandName, normalizedPath, err)
		return detailedErr
	}

	logger.Info("Successfully launched application for command '%s' (PID: %d)", commandName, execCmd.Process.Pid)
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
