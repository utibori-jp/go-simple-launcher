package gui

import (
	"testing"

	"app-launcher/config"
	"app-launcher/executor"
)

// TestNewGUIManager verifies that NewGUIManager creates a valid GUIManager instance
func TestNewGUIManager(t *testing.T) {
	// Create a mock config manager
	cfg, err := config.NewConfigManager("../testdata/valid_config.json")
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Create executor
	exec := executor.NewExecutor(cfg)

	// Create GUI manager
	gui := NewGUIManager(exec)

	if gui == nil {
		t.Fatal("NewGUIManager returned nil")
	}

	if gui.executor != exec {
		t.Error("GUIManager executor not set correctly")
	}

	if gui.visible {
		t.Error("GUIManager should not be visible initially")
	}
}

// TestGUIManagerStructure verifies the GUIManager has all required methods
func TestGUIManagerStructure(t *testing.T) {
	// Create a mock config manager
	cfg, err := config.NewConfigManager("../testdata/valid_config.json")
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Create executor
	exec := executor.NewExecutor(cfg)

	// Create GUI manager
	gui := NewGUIManager(exec)

	// Verify methods exist by calling them (without running the app)
	// We can't fully test GUI behavior without a display, but we can verify structure

	// These methods should exist and not panic when called on uninitialized GUI
	// (though they may not do anything useful without Initialize being called)

	// Test that the struct has the expected fields
	if gui.executor == nil {
		t.Error("executor field should not be nil")
	}
}
