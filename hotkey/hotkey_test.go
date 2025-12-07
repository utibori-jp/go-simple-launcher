package hotkey

import (
	"strings"
	"testing"
	"time"
)

// Unit Tests for Hotkey Manager

// TestNewHotkeyManager tests creating a new hotkey manager with valid callback
func TestNewHotkeyManager(t *testing.T) {
	callbackInvoked := false
	callback := func() {
		callbackInvoked = true
	}

	hm, err := NewHotkeyManager(callback)
	if err != nil {
		t.Fatalf("Failed to create HotkeyManager: %v", err)
	}

	if hm == nil {
		t.Fatal("Expected non-nil HotkeyManager")
	}

	if hm.callback == nil {
		t.Fatal("Expected callback to be set")
	}

	// Verify callback works
	hm.callback()
	if !callbackInvoked {
		t.Error("Callback was not invoked")
	}
}

// TestNewHotkeyManagerNilCallback tests that creating a hotkey manager with nil callback fails
func TestNewHotkeyManagerNilCallback(t *testing.T) {
	hm, err := NewHotkeyManager(nil)
	if err == nil {
		t.Fatal("Expected error when creating HotkeyManager with nil callback, got nil")
	}

	if hm != nil {
		t.Error("Expected nil HotkeyManager when callback is nil")
	}

	expectedSubstring := "callback function cannot be nil"
	if !strings.Contains(err.Error(), expectedSubstring) {
		t.Errorf("Error message should contain '%s', got: %v", expectedSubstring, err)
	}
}

// TestRegisterValidHotkey tests hotkey registration with valid formats
func TestRegisterValidHotkey(t *testing.T) {
	testCases := []struct {
		name        string
		hotkeyStr   string
		shouldError bool
	}{
		{"Alt+Space", "Alt+Space", false},
		{"Ctrl+Space", "Ctrl+Space", false},
		{"Ctrl+Alt+L", "Ctrl+Alt+L", false},
		{"Shift+F1", "Shift+F1", false},
		{"Ctrl+Shift+A", "Ctrl+Shift+A", false},
		{"Alt+Enter", "Alt+Enter", false},
		{"Ctrl+Escape", "Ctrl+Escape", false},
		{"Win+D", "Win+D", false},
		{"Ctrl+1", "Ctrl+1", false},
		{"Alt+F4", "Alt+F4", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hm, err := NewHotkeyManager(func() {})
			if err != nil {
				t.Fatalf("Failed to create HotkeyManager: %v", err)
			}

			err = hm.Register(tc.hotkeyStr)

			// Note: Registration might fail if the hotkey is already in use by the system
			// or another application. We'll check for specific error types.
			if err != nil {
				// If it's an "already in use" error, that's acceptable for this test
				// since we're testing the parsing logic, not actual system registration
				if strings.Contains(err.Error(), "already be in use") ||
					strings.Contains(err.Error(), "already in use") {
					t.Logf("Hotkey %s is already in use (acceptable for test): %v", tc.hotkeyStr, err)
					return
				}

				if tc.shouldError {
					// Expected error
					return
				}
				t.Errorf("Unexpected error registering hotkey '%s': %v", tc.hotkeyStr, err)
			}

			// Clean up
			hm.Stop()
		})
	}
}

// TestRegisterInvalidHotkeyFormat tests handling of invalid hotkey formats
func TestRegisterInvalidHotkeyFormat(t *testing.T) {
	testCases := []struct {
		name          string
		hotkeyStr     string
		errorContains string
	}{
		{
			name:          "Empty string",
			hotkeyStr:     "",
			errorContains: "hotkey string cannot be empty",
		},
		{
			name:          "No modifier",
			hotkeyStr:     "Space",
			errorContains: "at least one modifier",
		},
		{
			name:          "No key",
			hotkeyStr:     "Ctrl+Alt",
			errorContains: "unknown key",
		},
		{
			name:          "Invalid modifier",
			hotkeyStr:     "Invalid+Space",
			errorContains: "unknown modifier",
		},
		{
			name:          "Invalid key",
			hotkeyStr:     "Ctrl+InvalidKey",
			errorContains: "unknown key",
		},
		{
			name:          "Only key no plus",
			hotkeyStr:     "A",
			errorContains: "at least one modifier",
		},
		{
			name:          "Trailing plus",
			hotkeyStr:     "Ctrl+",
			errorContains: "unknown key",
		},
		{
			name:          "Leading plus",
			hotkeyStr:     "+Space",
			errorContains: "unknown modifier",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hm, err := NewHotkeyManager(func() {})
			if err != nil {
				t.Fatalf("Failed to create HotkeyManager: %v", err)
			}

			err = hm.Register(tc.hotkeyStr)
			if err == nil {
				t.Fatalf("Expected error when registering invalid hotkey '%s', got nil", tc.hotkeyStr)
			}

			if !strings.Contains(err.Error(), tc.errorContains) {
				t.Errorf("Error message should contain '%s', got: %v", tc.errorContains, err)
			}

			// Clean up
			hm.Stop()
		})
	}
}

// TestStartWithoutRegister tests that Start() fails when called before Register()
func TestStartWithoutRegister(t *testing.T) {
	hm, err := NewHotkeyManager(func() {})
	if err != nil {
		t.Fatalf("Failed to create HotkeyManager: %v", err)
	}

	err = hm.Start()
	if err == nil {
		t.Fatal("Expected error when calling Start() before Register(), got nil")
	}

	expectedSubstring := "hotkey not registered"
	if !strings.Contains(err.Error(), expectedSubstring) {
		t.Errorf("Error message should contain '%s', got: %v", expectedSubstring, err)
	}
}

// TestCallbackInvocation tests that the callback is invoked when hotkey is pressed
// Note: This test simulates callback invocation since we cannot actually press keys in unit tests
func TestCallbackInvocation(t *testing.T) {
	callbackInvoked := false
	callbackCount := 0

	callback := func() {
		callbackInvoked = true
		callbackCount++
	}

	hm, err := NewHotkeyManager(callback)
	if err != nil {
		t.Fatalf("Failed to create HotkeyManager: %v", err)
	}

	// Verify callback can be invoked
	hm.callback()
	if !callbackInvoked {
		t.Error("Callback was not invoked")
	}

	if callbackCount != 1 {
		t.Errorf("Expected callback count to be 1, got %d", callbackCount)
	}

	// Invoke again to test multiple invocations
	hm.callback()
	if callbackCount != 2 {
		t.Errorf("Expected callback count to be 2, got %d", callbackCount)
	}
}

// TestStopWithoutRegister tests that Stop() can be called safely without Register()
func TestStopWithoutRegister(t *testing.T) {
	hm, err := NewHotkeyManager(func() {})
	if err != nil {
		t.Fatalf("Failed to create HotkeyManager: %v", err)
	}

	// Should not panic
	hm.Stop()
}

// TestParseModifier tests the parseModifier function
func TestParseModifier(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"Ctrl", true},
		{"ctrl", true},
		{"Control", true},
		{"Alt", true},
		{"alt", true},
		{"Shift", true},
		{"shift", true},
		{"Win", true},
		{"Windows", true},
		{"Super", true},
		{"Cmd", true},
		{"Command", true},
		{"Invalid", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			_, ok := parseModifier(tc.input)
			if ok != tc.expected {
				t.Errorf("parseModifier(%s) = %v, expected %v", tc.input, ok, tc.expected)
			}
		})
	}
}

// TestParseKey tests the parseKey function
func TestParseKey(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		// Special keys
		{"Space", true},
		{"space", true},
		{"Enter", true},
		{"Return", true},
		{"Tab", true},
		{"Escape", true},
		{"Esc", true},
		{"Up", true},
		{"Down", true},
		{"Left", true},
		{"Right", true},

		// Function keys
		{"F1", true},
		{"f1", true},
		{"F12", true},

		// Number keys
		{"0", true},
		{"5", true},
		{"9", true},

		// Letter keys
		{"A", true},
		{"a", true},
		{"Z", true},
		{"z", true},

		// Invalid keys
		{"Invalid", false},
		{"F13", false},
		{"", false},
		{"AB", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			_, ok := parseKey(tc.input)
			if ok != tc.expected {
				t.Errorf("parseKey(%s) = %v, expected %v", tc.input, ok, tc.expected)
			}
		})
	}
}

// TestParseHotkey tests the parseHotkey function
func TestParseHotkey(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Valid Alt+Space",
			input:       "Alt+Space",
			shouldError: false,
		},
		{
			name:        "Valid Ctrl+Alt+L",
			input:       "Ctrl+Alt+L",
			shouldError: false,
		},
		{
			name:        "No modifier",
			input:       "Space",
			shouldError: true,
			errorMsg:    "at least one modifier",
		},
		{
			name:        "Invalid modifier",
			input:       "Invalid+Space",
			shouldError: true,
			errorMsg:    "unknown modifier",
		},
		{
			name:        "Invalid key",
			input:       "Ctrl+InvalidKey",
			shouldError: true,
			errorMsg:    "unknown key",
		},
		{
			name:        "Empty string",
			input:       "",
			shouldError: true,
			errorMsg:    "at least one modifier",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := parseHotkey(tc.input)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for input '%s', got nil", tc.input)
				} else if tc.errorMsg != "" && !strings.Contains(err.Error(), tc.errorMsg) {
					t.Errorf("Error message should contain '%s', got: %v", tc.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input '%s': %v", tc.input, err)
				}
			}
		})
	}
}

// TestMultipleStartStop tests starting and stopping the hotkey manager multiple times
func TestMultipleStartStop(t *testing.T) {
	callbackCount := 0
	callback := func() {
		callbackCount++
	}

	hm, err := NewHotkeyManager(callback)
	if err != nil {
		t.Fatalf("Failed to create HotkeyManager: %v", err)
	}

	// Try to register a hotkey (may fail if already in use, which is okay)
	err = hm.Register("Ctrl+Shift+F12")
	if err != nil {
		if strings.Contains(err.Error(), "already in use") {
			t.Skip("Hotkey already in use, skipping test")
		}
		t.Fatalf("Failed to register hotkey: %v", err)
	}

	// Start listening
	err = hm.Start()
	if err != nil {
		t.Fatalf("Failed to start hotkey manager: %v", err)
	}

	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)

	// Stop
	hm.Stop()

	// Give it a moment to stop
	time.Sleep(10 * time.Millisecond)

	// Should be able to stop again without error
	hm.Stop()
}
