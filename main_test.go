package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultConfigPath(t *testing.T) {
	// Save original APPDATA value
	originalAppData := os.Getenv("APPDATA")
	defer os.Setenv("APPDATA", originalAppData)

	// Test with APPDATA set
	testAppData := "C:\\Users\\TestUser\\AppData\\Roaming"
	os.Setenv("APPDATA", testAppData)

	expected := filepath.Join(testAppData, "launcher", "config.json")
	result := getDefaultConfigPath()

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Test with APPDATA not set
	os.Setenv("APPDATA", "")
	result = getDefaultConfigPath()
	expected = "config.json"

	if result != expected {
		t.Errorf("Expected %s when APPDATA is empty, got %s", expected, result)
	}
}

func TestNewApp_InvalidConfigPath(t *testing.T) {
	// Test with non-existent config file
	_, err := NewApp("nonexistent_config.json", "Alt+Space")
	if err == nil {
		t.Error("Expected error when loading non-existent config file, got nil")
	}
}

func TestNewApp_EmptyConfigPath(t *testing.T) {
	// Test with empty config path
	_, err := NewApp("", "Alt+Space")
	if err == nil {
		t.Error("Expected error when config path is empty, got nil")
	}
}

func TestNewApp_InvalidJSON(t *testing.T) {
	// Create a temporary invalid config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configContent := `{invalid json}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	_, err := NewApp(configPath, "Alt+Space")
	if err == nil {
		t.Error("Expected error when loading invalid JSON config, got nil")
	}
}

func TestNewApp_MissingCommandsField(t *testing.T) {
	// Create a config file without commands field
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configContent := `{"other": "field"}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	_, err := NewApp(configPath, "Alt+Space")
	if err == nil {
		t.Error("Expected error when config missing commands field, got nil")
	}
}
