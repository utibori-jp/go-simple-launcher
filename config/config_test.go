package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: app-launcher, Property 7: Valid JSON configurations load completely**
// **Validates: Requirements 3.2**
// For any valid JSON configuration file with proper structure, loading the configuration
// should successfully parse all command mappings and make them available for lookup.
func TestProperty_ValidJSONConfigurationsLoadCompletely(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("valid JSON configurations load all commands", prop.ForAll(
		func(commands map[string]Command) bool {
			// Skip empty command names as they're invalid
			for name := range commands {
				if name == "" {
					return true // Skip this test case
				}
			}

			// Create a valid Config structure
			cfg := Config{
				Commands: commands,
			}

			// Marshal to JSON
			jsonData, err := json.Marshal(cfg)
			if err != nil {
				t.Logf("Failed to marshal config: %v", err)
				return false
			}

			// Write to a temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test_config.json")
			if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil {
				t.Logf("Failed to write temp file: %v", err)
				return false
			}

			// Create ConfigManager and load
			cm, err := NewConfigManager(tmpFile)
			if err != nil {
				t.Logf("Failed to create ConfigManager: %v", err)
				return false
			}

			if err := cm.Load(); err != nil {
				t.Logf("Failed to load config: %v", err)
				return false
			}

			// Verify all commands are loaded and accessible
			for name, expectedCmd := range commands {
				loadedCmd, exists := cm.GetCommand(name)
				if !exists {
					t.Logf("Command '%s' not found after loading", name)
					return false
				}

				// Verify path matches
				if loadedCmd.Path != expectedCmd.Path {
					t.Logf("Command '%s' path mismatch: expected '%s', got '%s'",
						name, expectedCmd.Path, loadedCmd.Path)
					return false
				}

				// Verify args match (handle nil vs empty slice)
				expectedArgs := expectedCmd.Args
				if expectedArgs == nil {
					expectedArgs = []string{}
				}
				loadedArgs := loadedCmd.Args
				if loadedArgs == nil {
					loadedArgs = []string{}
				}

				if len(expectedArgs) != len(loadedArgs) {
					t.Logf("Command '%s' args length mismatch: expected %d, got %d",
						name, len(expectedArgs), len(loadedArgs))
					return false
				}

				for i := range expectedArgs {
					if expectedArgs[i] != loadedArgs[i] {
						t.Logf("Command '%s' arg[%d] mismatch: expected '%s', got '%s'",
							name, i, expectedArgs[i], loadedArgs[i])
						return false
					}
				}
			}

			return true
		},
		genValidCommandMap(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genValidCommandMap generates random valid command maps for property testing
func genValidCommandMap() gopter.Gen {
	return gen.MapOf(
		genNonEmptyString(),
		genValidCommand(),
	).SuchThat(func(m map[string]Command) bool {
		// Ensure all command names are non-empty and all paths are non-empty
		for name, cmd := range m {
			if name == "" || cmd.Path == "" {
				return false
			}
		}
		return true
	})
}

// genNonEmptyString generates non-empty strings for command names
func genNonEmptyString() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) < 50
	})
}

// genValidCommand generates valid Command structures
func genValidCommand() gopter.Gen {
	return gopter.CombineGens(
		genNonEmptyPath(),
		gen.SliceOf(gen.AlphaString()),
	).Map(func(values []interface{}) Command {
		return Command{
			Path: values[0].(string),
			Args: values[1].([]string),
		}
	})
}

// genNonEmptyPath generates non-empty path strings
func genNonEmptyPath() gopter.Gen {
	return gen.OneConstOf(
		"C:\\Windows\\System32\\cmd.exe",
		"C:\\Program Files\\App\\app.exe",
		"/usr/bin/app",
		"/opt/application/bin/app",
		"./relative/path/app.exe",
	)
}

// Unit Tests for Configuration Management

// TestLoadValidConfiguration tests loading a valid configuration file with multiple commands
func TestLoadValidConfiguration(t *testing.T) {
	// Use the valid test configuration
	cm, err := NewConfigManager("../testdata/valid_config.json")
	if err != nil {
		t.Fatalf("Failed to create ConfigManager: %v", err)
	}

	err = cm.Load()
	if err != nil {
		t.Fatalf("Failed to load valid configuration: %v", err)
	}

	// Verify all expected commands are loaded
	expectedCommands := []string{"browser", "editor", "terminal", "notepad"}
	for _, cmdName := range expectedCommands {
		cmd, exists := cm.GetCommand(cmdName)
		if !exists {
			t.Errorf("Expected command '%s' not found", cmdName)
		}
		if cmd.Path == "" {
			t.Errorf("Command '%s' has empty path", cmdName)
		}
	}

	// Verify specific command details
	editor, exists := cm.GetCommand("editor")
	if !exists {
		t.Fatal("Editor command not found")
	}
	if editor.Path != "C:\\Program Files\\Microsoft VS Code\\Code.exe" {
		t.Errorf("Editor path mismatch: got '%s'", editor.Path)
	}
	if len(editor.Args) != 1 || editor.Args[0] != "-n" {
		t.Errorf("Editor args mismatch: got %v", editor.Args)
	}
}

// TestLoadMissingConfigFile tests handling of a missing configuration file
func TestLoadMissingConfigFile(t *testing.T) {
	// Try to load a non-existent file
	cm, err := NewConfigManager("../testdata/nonexistent_config.json")
	if err != nil {
		t.Fatalf("Failed to create ConfigManager: %v", err)
	}

	err = cm.Load()
	if err == nil {
		t.Fatal("Expected error when loading missing config file, got nil")
	}

	// Verify error message indicates file not found
	expectedSubstring := "failed to read config file"
	if !contains(err.Error(), expectedSubstring) {
		t.Errorf("Error message should contain '%s', got: %v", expectedSubstring, err)
	}
}

// TestLoadMalformedJSON tests handling of malformed JSON
func TestLoadMalformedJSON(t *testing.T) {
	// Use the invalid test configuration (malformed JSON)
	cm, err := NewConfigManager("../testdata/invalid_config.json")
	if err != nil {
		t.Fatalf("Failed to create ConfigManager: %v", err)
	}

	err = cm.Load()
	if err == nil {
		t.Fatal("Expected error when loading malformed JSON, got nil")
	}

	// Verify error message indicates JSON parsing failure
	expectedSubstring := "failed to parse config file"
	if !contains(err.Error(), expectedSubstring) {
		t.Errorf("Error message should contain '%s', got: %v", expectedSubstring, err)
	}
}

// TestLoadEmptyConfiguration tests handling of an empty configuration
func TestLoadEmptyConfiguration(t *testing.T) {
	// Use the empty test configuration
	cm, err := NewConfigManager("../testdata/empty_config.json")
	if err != nil {
		t.Fatalf("Failed to create ConfigManager: %v", err)
	}

	err = cm.Load()
	if err != nil {
		t.Fatalf("Failed to load empty configuration: %v", err)
	}

	// Verify that no commands are available
	cmd, exists := cm.GetCommand("anycommand")
	if exists {
		t.Errorf("Expected no commands in empty config, but found: %v", cmd)
	}
}

// **Feature: app-launcher, Property 8: Invalid configurations fail gracefully**
// **Validates: Requirements 3.3**
// For any missing or malformed configuration file, the launcher should display
// a clear error message and exit without crashing.
func TestProperty_InvalidConfigurationsFailGracefully(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("missing config files fail with clear error", prop.ForAll(
		func(filename string) bool {
			// Create a path to a non-existent file
			tmpDir := t.TempDir()
			nonExistentPath := filepath.Join(tmpDir, filename)

			// Create ConfigManager with non-existent file
			cm, err := NewConfigManager(nonExistentPath)
			if err != nil {
				// Should not fail at creation, only at Load()
				t.Logf("Unexpected error at creation: %v", err)
				return false
			}

			// Attempt to load - should fail gracefully
			err = cm.Load()
			if err == nil {
				t.Logf("Expected error for missing file, got nil")
				return false
			}

			// Verify error message is clear and contains relevant information
			errMsg := err.Error()
			if errMsg == "" {
				t.Logf("Error message is empty")
				return false
			}

			// Error should mention it's a config file issue
			if !contains(errMsg, "config") && !contains(errMsg, "file") {
				t.Logf("Error message should mention config or file: %s", errMsg)
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) > 0 && len(s) < 100
		}),
	))

	properties.Property("malformed JSON fails with clear error", prop.ForAll(
		func(invalidJSON string) bool {
			// Create a temporary file with invalid JSON
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "invalid.json")

			if err := os.WriteFile(tmpFile, []byte(invalidJSON), 0644); err != nil {
				t.Logf("Failed to write temp file: %v", err)
				return false
			}

			// Create ConfigManager and attempt to load
			cm, err := NewConfigManager(tmpFile)
			if err != nil {
				t.Logf("Unexpected error at creation: %v", err)
				return false
			}

			err = cm.Load()
			if err == nil {
				t.Logf("Expected error for malformed JSON, got nil")
				return false
			}

			// Verify error message is clear
			errMsg := err.Error()
			if errMsg == "" {
				t.Logf("Error message is empty")
				return false
			}

			// Error should mention parsing or config issue
			if !contains(errMsg, "parse") && !contains(errMsg, "config") {
				t.Logf("Error message should mention parse or config: %s", errMsg)
				return false
			}

			return true
		},
		genInvalidJSON(),
	))

	properties.Property("invalid structure fails with clear error", prop.ForAll(
		func(invalidConfig string) bool {
			// Create a temporary file with valid JSON but invalid structure
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "invalid_structure.json")

			if err := os.WriteFile(tmpFile, []byte(invalidConfig), 0644); err != nil {
				t.Logf("Failed to write temp file: %v", err)
				return false
			}

			// Create ConfigManager and attempt to load
			cm, err := NewConfigManager(tmpFile)
			if err != nil {
				t.Logf("Unexpected error at creation: %v", err)
				return false
			}

			err = cm.Load()
			if err == nil {
				t.Logf("Expected error for invalid structure, got nil")
				return false
			}

			// Verify error message is clear
			errMsg := err.Error()
			if errMsg == "" {
				t.Logf("Error message is empty")
				return false
			}

			// Error should be informative
			if !contains(errMsg, "command") && !contains(errMsg, "path") && !contains(errMsg, "commands") {
				t.Logf("Error message should mention command or path: %s", errMsg)
				return false
			}

			return true
		},
		genInvalidStructure(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genInvalidJSON generates various forms of invalid JSON
func genInvalidJSON() gopter.Gen {
	return gen.OneConstOf(
		"not json at all",
		"{incomplete json",
		"{'single': 'quotes'}",
		"{\"missing\": }",
		"{\"trailing\": \"comma\",}",
		"[\"array\", \"not\", \"object\"]",
		"",
		"null",
		"123",
		"{\"nested\": {\"incomplete\": }",
	)
}

// genInvalidStructure generates valid JSON with invalid configuration structure
func genInvalidStructure() gopter.Gen {
	return gen.OneConstOf(
		// Missing commands field
		"{}",
		// Commands is not an object
		"{\"commands\": []}",
		"{\"commands\": \"string\"}",
		"{\"commands\": 123}",
		// Empty command name (handled by validation)
		"{\"commands\": {\"\": {\"path\": \"test.exe\", \"args\": []}}}",
		// Missing path field
		"{\"commands\": {\"test\": {\"args\": []}}}",
		// Empty path
		"{\"commands\": {\"test\": {\"path\": \"\", \"args\": []}}}",
		// Path is not a string
		"{\"commands\": {\"test\": {\"path\": 123, \"args\": []}}}",
		// Args is not an array
		"{\"commands\": {\"test\": {\"path\": \"test.exe\", \"args\": \"string\"}}}",
	)
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
