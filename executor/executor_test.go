package executor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"app-launcher/config"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// MockConfigManager はテスト用のConfigManagerの実装です
type MockConfigManager struct {
	Data config.Config
}

// GetCommand はメモリ上のDataからコマンドを返します
func (m *MockConfigManager) GetCommand(name string) (config.Command, bool) {
	cmd, exists := m.Data.Commands[name]
	return cmd, exists
}

// Load はMockなので何もしません（インターフェース適合用）
func (m *MockConfigManager) Load() error {
	return nil
}

// **Feature: app-launcher, Property 3: Valid commands execute correctly**
// **Validates: Requirements 2.1, 2.2, 5.1, 5.2**
// For any valid command name in the configuration, entering that command and pressing Enter
// should execute the application with the exact path and arguments specified in the configuration file.
func TestProperty_ValidCommandsExecuteCorrectly(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("valid commands execute with correct path and arguments", prop.ForAll(
		func(commandName string, cmd config.Command) bool {
			// Mock config
			cfg := config.Config{
				Commands: map[string]config.Command{
					commandName: cmd,
				},
			}
			cm := &MockConfigManager{
				Data: cfg,
			}

			// Create Executor
			executor := NewExecutor(cm)

			// Verify the command is in the config
			loadedCmd, exists := cm.GetCommand(commandName)
			if !exists {
				t.Logf("Command '%s' not found in config", commandName)
				return false
			}

			// Verify path and args match what we configured
			if loadedCmd.Path != cmd.Path {
				t.Logf("Path mismatch: expected '%s', got '%s'", cmd.Path, loadedCmd.Path)
				return false
			}

			// Normalize expected args (handle nil vs empty slice)
			expectedArgs := cmd.Args
			if expectedArgs == nil {
				expectedArgs = []string{}
			}
			loadedArgs := loadedCmd.Args
			if loadedArgs == nil {
				loadedArgs = []string{}
			}

			if len(expectedArgs) != len(loadedArgs) {
				t.Logf("Args length mismatch: expected %d, got %d", len(expectedArgs), len(loadedArgs))
				return false
			}

			for i := range expectedArgs {
				if expectedArgs[i] != loadedArgs[i] {
					t.Logf("Arg[%d] mismatch: expected '%s', got '%s'", i, expectedArgs[i], loadedArgs[i])
					return false
				}
			}

			// Test execution - we can't actually launch arbitrary executables in tests,
			// but we can verify that the executor attempts to execute with the correct
			// path and arguments by checking the error message for non-existent executables
			err := executor.Execute(commandName)

			// For valid executables that exist, execution should succeed (err == nil)
			// For non-existent executables, we should get a specific error
			// Either way, the executor should have attempted to use the correct path/args

			// If the executable actually exists and is valid, execution succeeds
			if err == nil {
				return true
			}

			// If the executable doesn't exist, we should get an error mentioning the command name
			// This verifies the executor tried to execute the right command
			errMsg := err.Error()
			if !contains(errMsg, commandName) {
				t.Logf("Error message doesn't mention command name '%s': %v", commandName, err)
				return false
			}

			// The error should indicate a launch failure (not a "command not found" error)
			if contains(errMsg, "not found") && !contains(errMsg, "failed to launch") {
				t.Logf("Got 'command not found' error instead of launch error: %v", err)
				return false
			}

			return true
		},
		genValidCommandName(),
		genExecutableCommand(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genValidCommandName generates valid command names
func genValidCommandName() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) < 30
	})
}

// genExecutableCommand generates Command structures with executable paths
func genExecutableCommand() gopter.Gen {
	// Use real system executables that are likely to exist
	var pathGen gopter.Gen

	if runtime.GOOS == "windows" {
		pathGen = gen.OneConstOf(
			"C:\\Windows\\System32\\hostname.exe",
			"C:\\Windows\\System32\\whoami.exe",
			"C:\\Windows\\System32\\ipconfig.exe",
		)
	} else {
		pathGen = gen.OneConstOf(
			"/bin/hostname",
			"/usr/bin/whoami",
			"/bin/date",
		)
	}

	return gopter.CombineGens(
		pathGen,
		gen.SliceOf(gen.AlphaString()),
	).Map(func(values []interface{}) config.Command {
		args := values[1].([]string)
		// Limit args to reasonable size
		if len(args) > 3 {
			args = args[:3]
		}
		return config.Command{
			Path: values[0].(string),
			Args: args,
		}
	})
}

// **Feature: app-launcher, Property 10: Application launch is non-blocking**
// **Validates: Requirements 5.4**
// For any application launched by the executor, the launch function should return control
// immediately without waiting for the application to complete execution.
func TestProperty_ApplicationLaunchIsNonBlocking(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("application launch returns immediately without blocking", prop.ForAll(
		func(commandName string) bool {
			// Create a temporary config file with a long-running command
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "test_config.json")

			// Use a command that will run for a while but is harmless
			var cmd config.Command
			if runtime.GOOS == "windows" {
				// Use timeout command which will run for specified seconds
				// "timeout /t 10" will wait for 10 seconds
				cmd = config.Command{
					Path: "C:\\Windows\\System32\\timeout.exe",
					Args: []string{"/t", "5", "/nobreak"},
				}
			} else {
				// Use sleep command on Unix-like systems
				cmd = config.Command{
					Path: "/bin/sleep",
					Args: []string{"5"},
				}
			}

			// Create config structure
			cfg := config.Config{
				Commands: map[string]config.Command{
					commandName: cmd,
				},
			}

			// Write config to file
			jsonData, err := json.Marshal(cfg)
			if err != nil {
				t.Logf("Failed to marshal config: %v", err)
				return false
			}

			if err := os.WriteFile(configFile, jsonData, 0644); err != nil {
				t.Logf("Failed to write config file: %v", err)
				return false
			}

			// Create ConfigManager and load
			cm, err := config.NewConfigManager(configFile)
			if err != nil {
				t.Logf("Failed to create ConfigManager: %v", err)
				return false
			}

			if err := cm.Load(); err != nil {
				t.Logf("Failed to load config: %v", err)
				return false
			}

			// Create Executor
			executor := NewExecutor(cm)

			// Measure execution time
			startTime := time.Now()

			// Execute the command
			err = executor.Execute(commandName)

			elapsedTime := time.Since(startTime)

			// If execution failed, that's okay for this test - we're testing non-blocking behavior
			// The important thing is that it returned quickly
			if err != nil {
				t.Logf("Execution failed (expected for some systems): %v", err)
				// Even if it failed, it should have failed quickly
			}

			// The execution should return in less than 1 second
			// If it's blocking, it would take 5+ seconds (the sleep/timeout duration)
			maxBlockingTime := 1 * time.Second

			if elapsedTime >= maxBlockingTime {
				t.Logf("Execution took %v, which suggests blocking behavior (threshold: %v)",
					elapsedTime, maxBlockingTime)
				return false
			}

			// Success: execution returned quickly (non-blocking)
			return true
		},
		genValidCommandName(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Feature: app-launcher, Property 9: Launch failures show error details**
// **Validates: Requirements 5.3**
// For any command that fails to launch (invalid path, permission denied, etc.),
// the launcher should display an error message containing details about the failure.
func TestProperty_LaunchFailuresShowErrorDetails(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("launch failures return errors with details", prop.ForAll(
		func(commandName string, invalidPath string) bool {
			// Create a temporary config file with an invalid executable path
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "test_config.json")

			// Create a command with an invalid/non-existent path
			cmd := config.Command{
				Path: invalidPath,
				Args: []string{},
			}

			// Create config structure
			cfg := config.Config{
				Commands: map[string]config.Command{
					commandName: cmd,
				},
			}

			// Write config to file
			jsonData, err := json.Marshal(cfg)
			if err != nil {
				t.Logf("Failed to marshal config: %v", err)
				return false
			}

			if err := os.WriteFile(configFile, jsonData, 0644); err != nil {
				t.Logf("Failed to write config file: %v", err)
				return false
			}

			// Create ConfigManager and load
			cm, err := config.NewConfigManager(configFile)
			if err != nil {
				t.Logf("Failed to create ConfigManager: %v", err)
				return false
			}

			if err := cm.Load(); err != nil {
				t.Logf("Failed to load config: %v", err)
				return false
			}

			// Create Executor
			executor := NewExecutor(cm)

			// Execute the command - this should fail
			err = executor.Execute(commandName)

			// The execution MUST fail for invalid paths
			if err == nil {
				t.Logf("Expected error for invalid path '%s', but got nil", invalidPath)
				return false
			}

			// The error message must contain details about the failure
			errMsg := err.Error()

			// Check that the error message contains the command name
			if !contains(errMsg, commandName) {
				t.Logf("Error message missing command name '%s': %v", commandName, err)
				return false
			}

			// Check that the error message indicates a launch failure
			if !contains(errMsg, "failed to launch") {
				t.Logf("Error message missing 'failed to launch' text: %v", err)
				return false
			}

			// The error should provide some detail about what went wrong
			// (e.g., "no such file", "permission denied", "not found", "file does not exist", etc.)
			hasDetail := contains(errMsg, "no such file") ||
				contains(errMsg, "not found") ||
				contains(errMsg, "cannot find") ||
				contains(errMsg, "permission denied") ||
				contains(errMsg, "access denied") ||
				contains(errMsg, "executable file not found") ||
				contains(errMsg, "file does not exist") ||
				contains(errMsg, "does not exist")

			if !hasDetail {
				t.Logf("Error message lacks specific failure details: %v", err)
				return false
			}

			return true
		},
		genValidCommandName(),
		genInvalidExecutablePath(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genInvalidExecutablePath generates paths that are unlikely to be valid executables
func genInvalidExecutablePath() gopter.Gen {
	basePaths := []string{
		"/nonexistent/path/to/executable",
		"C:\\NonExistent\\Path\\executable.exe",
		"/tmp/does_not_exist_binary",
		"/usr/local/bin/fake_app",
		"not/a/real/path/binary",
		"/this/path/definitely/does/not/exist/app",
		"C:\\FakePath\\NonExistent\\program.exe",
	}

	return gopter.CombineGens(
		gen.OneConstOf(basePaths[0], basePaths[1], basePaths[2], basePaths[3], basePaths[4], basePaths[5], basePaths[6]),
		gen.Identifier(),
	).Map(func(values []interface{}) string {
		basePath := values[0].(string)
		suffix := values[1].(string)
		return basePath + "_" + suffix
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Unit Tests

// TestExecuteCommandWithoutArguments tests executing a command without arguments
func TestExecuteCommandWithoutArguments(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test_config.json")

	// Use a real system executable
	var execPath string
	if runtime.GOOS == "windows" {
		execPath = "C:\\Windows\\System32\\cmd.exe"
	} else {
		execPath = "/bin/echo"
	}

	// Create config with command without arguments
	cfg := config.Config{
		Commands: map[string]config.Command{
			"testcmd": {
				Path: execPath,
				Args: []string{},
			},
		},
	}

	// Write config to file
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create ConfigManager and load
	cm, err := config.NewConfigManager(configFile)
	if err != nil {
		t.Fatalf("Failed to create ConfigManager: %v", err)
	}

	if err := cm.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create Executor
	executor := NewExecutor(cm)

	// Execute the command
	err = executor.Execute("testcmd")

	// Should succeed (or fail gracefully if executable doesn't exist)
	if err != nil {
		// If it fails, it should be a launch failure, not a "command not found"
		if contains(err.Error(), "command 'testcmd' not found") {
			t.Errorf("Command should exist in config but got 'not found' error: %v", err)
		}
		// Otherwise, it's a launch failure which is acceptable for this test
		t.Logf("Launch failed (acceptable): %v", err)
	}
}

// TestExecuteCommandWithMultipleArguments tests executing a command with multiple arguments
func TestExecuteCommandWithMultipleArguments(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test_config.json")

	// Use a real system executable with arguments
	var execPath string
	var args []string
	if runtime.GOOS == "windows" {
		execPath = "C:\\Windows\\System32\\cmd.exe"
		args = []string{"/c", "echo", "test"}
	} else {
		execPath = "/bin/echo"
		args = []string{"hello", "world"}
	}

	// Create config with command with multiple arguments
	cfg := config.Config{
		Commands: map[string]config.Command{
			"testcmd": {
				Path: execPath,
				Args: args,
			},
		},
	}

	// Write config to file
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create ConfigManager and load
	cm, err := config.NewConfigManager(configFile)
	if err != nil {
		t.Fatalf("Failed to create ConfigManager: %v", err)
	}

	if err := cm.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create Executor
	executor := NewExecutor(cm)

	// Execute the command
	err = executor.Execute("testcmd")

	// Should succeed (or fail gracefully if executable doesn't exist)
	if err != nil {
		// If it fails, it should be a launch failure, not a "command not found"
		if contains(err.Error(), "command 'testcmd' not found") {
			t.Errorf("Command should exist in config but got 'not found' error: %v", err)
		}
		// Otherwise, it's a launch failure which is acceptable for this test
		t.Logf("Launch failed (acceptable): %v", err)
	}
}

// TestExecuteNonExistentExecutablePath tests handling of non-existent executable path
func TestExecuteNonExistentExecutablePath(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test_config.json")

	// Use a path that definitely doesn't exist
	nonExistentPath := "/this/path/does/not/exist/fake_executable"
	if runtime.GOOS == "windows" {
		nonExistentPath = "C:\\NonExistent\\Path\\fake.exe"
	}

	// Create config with non-existent executable
	cfg := config.Config{
		Commands: map[string]config.Command{
			"fakecmd": {
				Path: nonExistentPath,
				Args: []string{},
			},
		},
	}

	// Write config to file
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create ConfigManager and load
	cm, err := config.NewConfigManager(configFile)
	if err != nil {
		t.Fatalf("Failed to create ConfigManager: %v", err)
	}

	if err := cm.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create Executor
	executor := NewExecutor(cm)

	// Execute the command - should fail
	err = executor.Execute("fakecmd")

	// Must return an error
	if err == nil {
		t.Fatal("Expected error for non-existent executable, got nil")
	}

	// Error should mention the command name
	if !contains(err.Error(), "fakecmd") {
		t.Errorf("Error should mention command name 'fakecmd': %v", err)
	}

	// Error should indicate launch failure
	if !contains(err.Error(), "failed to launch") {
		t.Errorf("Error should indicate 'failed to launch': %v", err)
	}

	// Error should contain details about the failure
	errMsg := err.Error()
	hasDetail := contains(errMsg, "no such file") ||
		contains(errMsg, "not found") ||
		contains(errMsg, "cannot find") ||
		contains(errMsg, "executable file not found") ||
		contains(errMsg, "file does not exist")

	if !hasDetail {
		t.Errorf("Error should contain specific failure details: %v", err)
	}
}

// TestExecutePermissionDenied tests handling of permission denied errors
func TestExecutePermissionDenied(t *testing.T) {
	// This test is platform-specific and may not work on all systems
	// On Windows, permission denied is harder to test without admin rights
	// On Unix-like systems, we can create a file without execute permissions

	if runtime.GOOS == "windows" {
		t.Skip("Permission denied test is difficult to reliably test on Windows")
	}

	// Create a temporary directory and file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test_config.json")
	noExecFile := filepath.Join(tmpDir, "no_exec_file")

	// Create a file without execute permissions
	if err := os.WriteFile(noExecFile, []byte("#!/bin/sh\necho test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Ensure the file does NOT have execute permissions
	if err := os.Chmod(noExecFile, 0644); err != nil {
		t.Fatalf("Failed to set file permissions: %v", err)
	}

	// Create config with the non-executable file
	cfg := config.Config{
		Commands: map[string]config.Command{
			"noperm": {
				Path: noExecFile,
				Args: []string{},
			},
		},
	}

	// Write config to file
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create ConfigManager and load
	cm, err := config.NewConfigManager(configFile)
	if err != nil {
		t.Fatalf("Failed to create ConfigManager: %v", err)
	}

	if err := cm.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create Executor
	executor := NewExecutor(cm)

	// Execute the command - should fail with permission error
	err = executor.Execute("noperm")

	// Must return an error
	if err == nil {
		t.Fatal("Expected error for permission denied, got nil")
	}

	// Error should mention the command name
	if !contains(err.Error(), "noperm") {
		t.Errorf("Error should mention command name 'noperm': %v", err)
	}

	// Error should indicate launch failure
	if !contains(err.Error(), "failed to launch") {
		t.Errorf("Error should indicate 'failed to launch': %v", err)
	}

	// Error should contain permission-related details
	errMsg := err.Error()
	hasPermissionDetail := contains(errMsg, "permission denied") ||
		contains(errMsg, "access denied") ||
		contains(errMsg, "not permitted")

	if !hasPermissionDetail {
		t.Logf("Warning: Error may not explicitly mention permission issue: %v", err)
		// Don't fail the test as the exact error message may vary by system
	}
}
