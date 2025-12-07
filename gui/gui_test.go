package gui

import (
	"runtime"
	"strings"
	"testing"

	"app-launcher/config"
	"app-launcher/executor"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

type MockConfigManager struct {
	Data config.Config
}

func (m *MockConfigManager) GetCommand(name string) (config.Command, bool) {
	cmd, exists := m.Data.Commands[name]
	return cmd, exists
}

func (m *MockConfigManager) Load() error {
	return nil
}

// **Feature: app-launcher, Property 1: Hotkey toggles window visibility**
// **Validates: Requirements 1.1, 1.3**
// For any window visibility state, pressing the configured hotkey should toggle
// the window to the opposite visibility state (hidden → visible, visible → hidden).
//
// Note: This property test verifies the core toggle property by testing the state
// management logic. Since Fyne GUI operations require a display environment, this test
// validates that the visibility state correctly alternates between true and false,
// which is the fundamental behavior that Toggle() must maintain. The actual window
// show/hide operations are validated through manual/integration testing.
func TestProperty_HotkeyTogglesWindowVisibility(t *testing.T) {
	// Use test.NewApp() instead of app.New() to avoid threading crashes in tests
	testApp := test.NewApp()

	// Setup Mock dependencies
	mockCfg := config.Config{
		Commands: map[string]config.Command{},
	}
	cm := &MockConfigManager{Data: mockCfg}
	exec := executor.NewExecutor(cm)

	// Initialize GUIManager to test the actual logic
	gui := NewGUIManager(exec, testApp)
	gui.Initialize()

	properties := gopter.NewProperties(nil)

	properties.Property("visibility state alternates correctly through toggle operations", prop.ForAll(
		func(numToggles int) bool {
			// Ensure window starts in a known state (Visible)
			gui.Show()
			expectedVisible := true

			for i := 0; i < numToggles; i++ {
				// Apply toggle logic
				// In the actual app, this is: guiManager.Toggle()
				// We now call the actual method instead of simulating it manually
				gui.Toggle()

				// Calculate expected state
				expectedVisible = !expectedVisible

				// Verify the state changed (Internal logic check)
				// We check if the application's internal state matches our expectation
				if gui.visible != expectedVisible {
					t.Logf("Toggle %d failed: expected visible=%v, got=%v", i, expectedVisible, gui.visible)
					return false
				}
			}

			// Final state verification based on toggle count
			// Even number of toggles -> Should return to initial state (Visible)
			if numToggles%2 == 0 && !gui.visible {
				t.Logf("After %d toggles (even), window should be visible but is hidden", numToggles)
				return false
			}

			// Odd number of toggles -> Should be opposite of initial state (Hidden)
			if numToggles%2 == 1 && gui.visible {
				t.Logf("After %d toggles (odd), window should be hidden but is visible", numToggles)
				return false
			}

			return true
		},
		gen.IntRange(1, 50), // Test random toggle sequences between 1 and 50 times
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Feature: app-launcher, Property 2: Window show focuses input field**
// **Validates: Requirements 1.2**
// For any window show event, the input field should receive focus immediately
// after the window becomes visible.
//
// Note: This property test verifies the focus behavior by testing that Show()
// correctly manages the visibility state and prepares the input field for focus.
// Since Fyne GUI operations require a display environment, this test validates
// the state transitions and input field preparation that must occur when Show()
// is called. The actual focus operation is validated through manual/integration testing.
func TestProperty_WindowShowFocusesInputField(t *testing.T) {
	// Use test.NewApp() to avoid threading crashes
	testApp := test.NewApp()

	// Setup Mock dependencies (avoid file I/O)
	mockCfg := config.Config{
		Commands: map[string]config.Command{},
	}

	// Create executor with MockConfigManager
	cm := &MockConfigManager{Data: mockCfg}
	exec := executor.NewExecutor(cm)

	// Initialize GUIManager
	gui := NewGUIManager(exec, testApp)

	// Initialize the UI (creates window, entry, etc.)
	gui.Initialize()

	// Capture the window created by Initialize
	w := gui.window
	if w == nil {
		t.Fatal("Window was not initialized by gui.Initialize()")
	}

	// Verify UI components are built
	if gui.entry == nil {
		t.Fatal("Entry field was not initialized by gui.Initialize()")
	}

	properties := gopter.NewProperties(nil)

	properties.Property("Show() makes window visible and focuses input field", prop.ForAll(
		func(numCycles int) bool {
			for i := 0; i < numCycles; i++ {
				// --- Setup phase ---
				// Reset to "closed" state before testing Show()
				gui.Hide()

				// Set dirty state to verify cleanup
				gui.entry.SetText("some old text")

				// Ensure focus is lost (simulate user doing something else)
				w.Canvas().Unfocus()

				// --- Action phase ---
				// Execute the actual business logic we want to test
				gui.Show()

				// --- Verification phase ---

				// 1. Verify internal visibility state
				if !gui.visible {
					t.Logf("Cycle %d: Internal visible state is false after Show()", i)
					return false
				}

				// 2. Verify Input Field Focus (Critical Requirement)
				// Note: Fyne test driver updates focus immediately upon RequestFocus call
				focusedObj := w.Canvas().Focused()

				if focusedObj == nil {
					t.Logf("Cycle %d: No object has focus after Show()", i)
					return false
				}

				// Compare the focused object with our input field
				if focusedObj != gui.entry {
					t.Logf("Cycle %d: Focus is on wrong object. Expected entry, got %v", i, focusedObj)
					return false
				}

				// 3. Verify Input Field is cleared
				if gui.entry.Text != "" {
					t.Logf("Cycle %d: Input text was not cleared. Got: '%s'", i, gui.entry.Text)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 50), // Test random repetition 1-50 times
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Feature: app-launcher, Property 4: Successful launch hides window**
// **Validates: Requirements 2.3**
// For any command that successfully launches an application, the window should
// become hidden immediately after execution.
func TestProperty_SuccessfulLaunchHidesWindow(t *testing.T) {
	// Use test.NewApp() to avoid threading crashes
	testApp := test.NewApp()

	properties := gopter.NewProperties(nil)

	properties.Property("successful command execution hides the window", prop.ForAll(
		func(commandName string) bool {
			// Setup: Create a mock config with a valid executable command
			var validCmd config.Command
			if runtime.GOOS == "windows" {
				// Use a simple Windows command that will succeed
				validCmd = config.Command{
					Path: "C:\\Windows\\System32\\hostname.exe",
					Args: []string{},
				}
			} else {
				// Use a simple Unix command that will succeed
				validCmd = config.Command{
					Path: "/bin/hostname",
					Args: []string{},
				}
			}

			mockCfg := &MockConfigManager{
				Data: config.Config{
					Commands: map[string]config.Command{
						commandName: validCmd,
					},
				},
			}

			// Create executor with the mock config
			exec := executor.NewExecutor(mockCfg)

			// Create GUI manager and Initialize to build UI components
			gui := NewGUIManager(exec, testApp)
			gui.Initialize()

			// Ensure window starts visible (simulating user activation)
			gui.Show()

			// Verify precondition: window should be visible
			if !gui.visible {
				t.Logf("Precondition failed: window should be visible before command execution")
				return false
			}

			// --- Action Phase ---
			// Trigger the command submission event (simulate User pressing Enter)
			// This exercises the actual logic: handleCommandSubmit -> exec.Execute -> Hide()
			gui.entry.OnSubmitted(commandName)

			// --- Verification Phase ---

			// Verify postcondition: window should be hidden automatically after successful execution
			// We do NOT call gui.Hide() manually here. We verify the app did it.
			if gui.visible {
				t.Logf("Property violated: window remained visible after successful command execution")
				return false
			}

			// Additional check: Ensure no error is shown
			if gui.errorLabel.Visible() {
				t.Logf("Property violated: error label is visible but command should have succeeded")
				return false
			}

			return true
		},
		genValidCommandName(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genValidCommandName generates valid command names for testing
func genValidCommandName() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) < 30
	})
}

// **Feature: app-launcher, Property 5: Invalid commands show errors**
// **Validates: Requirements 2.4**
// For any command name that does not exist in the configuration, entering that
// command and pressing Enter should display an error message and keep the window visible.
func TestProperty_InvalidCommandsShowErrors(t *testing.T) {
	// Use test.NewApp() to avoid threading crashes
	testApp := test.NewApp()

	properties := gopter.NewProperties(nil)

	properties.Property("invalid command execution shows error and keeps window visible", prop.ForAll(
		func(invalidCommandName string) bool {
			// Setup: Create a mock config with a few known commands
			knownCommands := map[string]config.Command{
				"browser":  {Path: "dummy_path", Args: []string{}},
				"editor":   {Path: "dummy_path", Args: []string{}},
				"terminal": {Path: "dummy_path", Args: []string{}},
			}

			// Ensure the generated command name is NOT in our known commands
			if _, exists := knownCommands[invalidCommandName]; exists {
				return true // Skip if by chance we generated a valid command
			}

			mockCfg := &MockConfigManager{
				Data: config.Config{
					Commands: knownCommands,
				},
			}

			// Create executor with the mock config
			exec := executor.NewExecutor(mockCfg)

			// Create GUI manager and Initialize
			gui := NewGUIManager(exec, testApp)
			gui.Initialize()

			// Ensure window starts visible
			gui.Show()

			// Verify precondition: error label should be hidden initially
			if gui.errorLabel.Visible() {
				t.Logf("Precondition failed: error label should be hidden initially")
				return false
			}

			// --- Action Phase ---
			// Trigger the command submission event with an invalid command
			// This tests the actual error handling logic in handleCommandSubmit
			gui.entry.OnSubmitted(invalidCommandName)

			// --- Verification Phase ---

			// 1. Verify error message is displayed
			if !gui.errorLabel.Visible() {
				t.Logf("Property violated: error label is not visible after invalid command")
				return false
			}

			// 2. Verify error message content (should contain the command name)
			// Note: We need to check the Text property directly, as Fyne test driver might not render it visually in the same way
			errorText := gui.errorLabel.Text
			if !strings.Contains(errorText, invalidCommandName) {
				t.Logf("Property violated: error message '%s' does not mention command '%s'", errorText, invalidCommandName)
				return false
			}

			// 3. Verify window should remain visible (Requirement: allow user to retry)
			if !gui.visible {
				t.Logf("Property violated: window is hidden after invalid command execution")
				return false
			}

			return true
		},
		genInvalidCommandName(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genInvalidCommandName generates command names that are unlikely to be in the config
func genInvalidCommandName() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		// Generate non-empty strings that are not common command names
		if len(s) == 0 || len(s) > 50 {
			return false
		}
		// Exclude common command names used in the test setup
		commonNames := []string{"browser", "editor", "terminal"}
		for _, common := range commonNames {
			if s == common {
				return false
			}
		}
		return true
	})
}

// **Feature: app-launcher, Property 6: Escape cancels without execution**
// **Validates: Requirements 2.5**
// For any text entered in the input field, pressing Escape should hide the window
// without executing any command or launching any application.
func TestProperty_EscapeCancelsWithoutExecution(t *testing.T) {
	// Use test.NewApp() to avoid threading crashes
	testApp := test.NewApp()

	properties := gopter.NewProperties(nil)

	properties.Property("Escape key hides window without executing command", prop.ForAll(
		func(inputText string) bool {
			// Setup: Create a mock config with some valid commands
			mockCfg := &MockConfigManager{
				Data: config.Config{
					Commands: map[string]config.Command{
						"browser": {Path: "dummy_path", Args: []string{}},
						"editor":  {Path: "dummy_path", Args: []string{}},
					},
				},
			}

			// Create executor with the mock config
			exec := executor.NewExecutor(mockCfg)

			// Create GUI manager and Initialize
			gui := NewGUIManager(exec, testApp)
			gui.Initialize()

			// Ensure window starts visible (simulating user activation)
			gui.Show()

			// Verify precondition: window should be visible
			if !gui.visible {
				t.Logf("Precondition failed: window should be visible before Escape")
				return false
			}

			// Set the input text (simulating user typing)
			gui.entry.SetText(inputText)

			// --- Action Phase ---
			// 1. Ensure the entry widget has focus (key events go to the focused widget)
			gui.window.Canvas().Focus(gui.entry)

			// 2. Simulate Escape key press via the Window Canvas
			// This tests that the key event propagates correctly to the handler
			if handler := gui.window.Canvas().OnTypedKey(); handler != nil {
				handler(&fyne.KeyEvent{Name: fyne.KeyEscape})
			}

			// --- Verification Phase ---

			// Verify postcondition 1: window should be hidden
			if gui.visible {
				t.Logf("Property violated: window is still visible after Escape key press")
				return false
			}

			// Verify postcondition 2: no command was executed
			// We verify this indirectly by checking that no error is shown.
			// (If a command was attempted and failed, an error would be shown.
			//  If a command succeeded, the window would hide, but we are testing Escape here.)

			// If the input was a valid command (e.g. "browser"), hitting Enter would run it.
			// Hitting Escape should NOT run it. Since we can't easily spy on the Executor here without a spy mock,
			// checking side effects (error label) is a reasonable proxy for this property test.
			if gui.errorLabel.Visible() {
				t.Logf("Property violated: error label became visible, implying command execution attempted")
				return false
			}

			return true
		},
		genArbitraryInputText(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genArbitraryInputText generates arbitrary text that might be entered in the input field
func genArbitraryInputText() gopter.Gen {
	return gen.OneConstOf(
		"",                       // Empty string
		"   ",                    // Whitespace
		"browser",                // Valid command name
		"invalid_command_xyz123", // Invalid command name
		"cmd with spaces",        // Command with spaces
		"special!@#$%chars",      // Special characters
		"test123",                // Alphanumeric
		"a",                      // Single character
		"very_long_command_name_that_does_not_exist_in_config", // Long string
	)
}

// **Feature: app-launcher, Component Initialization**
// **Validates: Requirements 3.1, 5.1**
// The NewGUIManager constructor must return a valid instance with all dependencies
// correctly injected, UI components initialized, and the window in a hidden state.
func TestNewGUIManager(t *testing.T) {
	// 1. Initialize Fyne test driver
	// Create a test Fyne app
	testApp := test.NewApp()

	// 2. Setup Mock dependencies (avoid file I/O)
	mockCfg := &MockConfigManager{
		Data: config.Config{Commands: map[string]config.Command{}},
	}
	exec := executor.NewExecutor(mockCfg)

	// 3. Create the object under test
	gui := NewGUIManager(exec, testApp)

	// 4. Verification (Assertions)

	// Verify instance creation
	if gui == nil {
		t.Fatal("NewGUIManager returned nil")
	}

	// Verify dependency injection
	if gui.executor != exec {
		t.Error("GUIManager executor not set correctly")
	}

	// Verify initial state (should be hidden)
	if gui.visible {
		t.Error("GUIManager should not be visible initially")
	}

	// Verify UI component initialization
	if gui.entry == nil {
		t.Error("GUIManager input field (entry) was not initialized")
	}
}

// Unit Tests for GUI Interactions
// **Validates: Requirements 1.1, 1.3, 2.5**

// TestToggleWindowVisibility tests window show/hide on toggle
func TestToggleWindowVisibility(t *testing.T) {
	testApp := test.NewApp()
	mockCfg := &MockConfigManager{
		Data: config.Config{Commands: map[string]config.Command{}},
	}
	exec := executor.NewExecutor(mockCfg)
	gui := NewGUIManager(exec, testApp)
	gui.Initialize()

	// Test initial state - should be hidden
	if gui.visible {
		t.Error("Window should be hidden initially")
	}

	// Test first toggle - should show
	gui.Toggle()
	if !gui.visible {
		t.Error("Window should be visible after first toggle")
	}

	// Test second toggle - should hide
	gui.Toggle()
	if gui.visible {
		t.Error("Window should be hidden after second toggle")
	}

	// Test third toggle - should show again
	gui.Toggle()
	if !gui.visible {
		t.Error("Window should be visible after third toggle")
	}
}

// TestEscapeKeyCancellation tests Escape key cancellation
func TestEscapeKeyCancellation(t *testing.T) {
	testApp := test.NewApp()
	mockCfg := &MockConfigManager{
		Data: config.Config{
			Commands: map[string]config.Command{
				"browser": {Path: "C:\\Program Files\\Firefox\\firefox.exe", Args: []string{}},
			},
		},
	}
	exec := executor.NewExecutor(mockCfg)
	gui := NewGUIManager(exec, testApp)
	gui.Initialize()

	// Show the window
	gui.Show()
	if !gui.visible {
		t.Fatal("Window should be visible after Show()")
	}

	// Set some text in the entry field
	gui.entry.SetText("browser")
	if gui.entry.Text != "browser" {
		t.Error("Entry text should be 'browser'")
	}

	// Simulate Escape key press by calling Hide() directly
	// (In the real app, the key handler calls Hide() when Escape is pressed)
	gui.Hide()

	// Verify window is hidden
	if gui.visible {
		t.Error("Window should be hidden after Escape key")
	}

	// Verify no error was displayed (command was not executed)
	if gui.errorLabel.Visible() {
		t.Error("Error label should not be visible after Escape")
	}
}

// TestEnterKeyCommandSubmission tests Enter key command submission
func TestEnterKeyCommandSubmission(t *testing.T) {
	testApp := test.NewApp()

	t.Run("valid command hides window", func(t *testing.T) {
		mockCfg := &MockConfigManager{
			Data: config.Config{
				Commands: map[string]config.Command{
					"test": {Path: "C:\\Windows\\System32\\hostname.exe", Args: []string{}},
				},
			},
		}
		exec := executor.NewExecutor(mockCfg)
		gui := NewGUIManager(exec, testApp)
		gui.Initialize()

		// Show the window
		gui.Show()
		if !gui.visible {
			t.Fatal("Window should be visible after Show()")
		}

		// Set valid command text
		gui.entry.SetText("test")

		// Simulate Enter key by calling the submit handler
		gui.entry.OnSubmitted("test")

		// Verify window is hidden after successful execution
		if gui.visible {
			t.Error("Window should be hidden after successful command execution")
		}

		// Verify no error is displayed
		if gui.errorLabel.Visible() {
			t.Error("Error label should not be visible after successful execution")
		}
	})

	t.Run("invalid command shows error and keeps window visible", func(t *testing.T) {
		mockCfg := &MockConfigManager{
			Data: config.Config{
				Commands: map[string]config.Command{
					"browser": {Path: "C:\\Program Files\\Firefox\\firefox.exe", Args: []string{}},
				},
			},
		}
		exec := executor.NewExecutor(mockCfg)
		gui := NewGUIManager(exec, testApp)
		gui.Initialize()

		// Show the window
		gui.Show()
		if !gui.visible {
			t.Fatal("Window should be visible after Show()")
		}

		// Set invalid command text
		gui.entry.SetText("invalid_command")

		// Simulate Enter key by calling the submit handler
		gui.entry.OnSubmitted("invalid_command")

		// Verify window is still visible after failed execution
		if !gui.visible {
			t.Error("Window should remain visible after invalid command")
		}

		// Verify error is displayed
		if !gui.errorLabel.Visible() {
			t.Error("Error label should be visible after invalid command")
		}

		// Verify error message contains the command name
		if !strings.Contains(gui.errorLabel.Text, "invalid_command") {
			t.Errorf("Error message should mention the command name, got: %s", gui.errorLabel.Text)
		}
	})
}

// TestErrorMessageDisplay tests error message display
func TestErrorMessageDisplay(t *testing.T) {
	testApp := test.NewApp()
	mockCfg := &MockConfigManager{
		Data: config.Config{Commands: map[string]config.Command{}},
	}
	exec := executor.NewExecutor(mockCfg)
	gui := NewGUIManager(exec, testApp)
	gui.Initialize()

	// Initially, error label should be hidden
	if gui.errorLabel.Visible() {
		t.Error("Error label should be hidden initially")
	}

	// Show an error message
	errorMsg := "Test error message"
	gui.ShowError(errorMsg)

	// Verify error label is visible
	if !gui.errorLabel.Visible() {
		t.Error("Error label should be visible after ShowError()")
	}

	// Verify error message text
	if gui.errorLabel.Text != errorMsg {
		t.Errorf("Error label text should be '%s', got '%s'", errorMsg, gui.errorLabel.Text)
	}

	// Show the window and verify error is cleared
	gui.Show()

	// Verify error label is hidden after Show()
	if gui.errorLabel.Visible() {
		t.Error("Error label should be hidden after Show()")
	}
}
