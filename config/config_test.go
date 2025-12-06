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
