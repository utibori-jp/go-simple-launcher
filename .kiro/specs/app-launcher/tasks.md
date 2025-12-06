# Implementation Plan

- [x] 1. Set up project structure and dependencies
  - Initialize Go module with appropriate name
  - Add Fyne dependency for GUI
  - Add keyboard shortcut library for Windows (e.g., robotgo or golang-design/hotkey)
  - Create directory structure: config/, executor/, gui/, hotkey/
  - Create testdata/ directory with sample configuration files
  - _Requirements: 6.1, 6.2, 6.3_

- [x] 2. Implement configuration management
  - Create Config and Command data structures matching the JSON schema
  - Implement ConfigManager with Load() method to read and parse JSON
  - Implement GetCommand() method for O(1) command lookup
  - Add validation for required fields (command name, path)
  - _Requirements: 3.1, 3.2, 4.1, 4.2, 4.3_

- [x] 2.1 Write property test for configuration loading
  - **Property 7: Valid JSON configurations load completely**
  - **Validates: Requirements 3.2**

- [x] 2.2 Write unit tests for configuration management
  - Test loading valid configuration with multiple commands
  - Test handling missing configuration file
  - Test handling malformed JSON
  - Test handling empty configuration
  - _Requirements: 3.1, 3.2, 3.3_

- [x] 3. Implement application executor
  - Create Executor struct with reference to ConfigManager
  - Implement Execute() method to lookup commands and launch applications
  - Use exec.Command() to start processes without blocking
  - Handle Windows-specific path formats and .exe extensions
  - Return detailed errors for launch failures
  - _Requirements: 2.1, 2.2, 5.1, 5.2, 5.4_

- [x] 3.1 Write property test for valid command execution
  - **Property 3: Valid commands execute correctly**
  - **Validates: Requirements 2.1, 2.2, 5.1, 5.2**

- [x] 3.2 Write property test for non-blocking execution
  - **Property 10: Application launch is non-blocking**
  - **Validates: Requirements 5.4**

- [x] 3.3 Write property test for launch failure handling
  - **Property 9: Launch failures show error details**
  - **Validates: Requirements 5.3**

- [x] 3.4 Write unit tests for executor
  - Test executing command without arguments
  - Test executing command with multiple arguments
  - Test handling non-existent executable path
  - Test handling permission denied errors
  - _Requirements: 5.1, 5.2, 5.3_

- [x] 4. Implement GUI manager with Fyne
  - Create GUIManager struct with Fyne app and window
  - Implement Initialize() to create window with text entry widget
  - Implement Show() and Hide() methods for window visibility
  - Implement Toggle() method for hotkey activation
  - Set up Enter key handler to execute commands via Executor
  - Set up Escape key handler to hide window
  - Implement ShowError() to display error messages in GUI
  - Configure window to be always on top and centered
  - _Requirements: 1.2, 1.4, 2.1, 2.3, 2.4, 2.5_

- [x] 4.1 Write property test for window visibility toggle
  - **Property 1: Hotkey toggles window visibility**
  - **Validates: Requirements 1.1, 1.3**

- [x] 4.2 Write property test for input focus
  - **Property 2: Window show focuses input field**
  - **Validates: Requirements 1.2**

- [ ]* 4.3 Write property test for successful launch hiding window
  - **Property 4: Successful launch hides window**
  - **Validates: Requirements 2.3**

- [ ]* 4.4 Write property test for invalid command error display
  - **Property 5: Invalid commands show errors**
  - **Validates: Requirements 2.4**

- [ ]* 4.5 Write property test for Escape cancellation
  - **Property 6: Escape cancels without execution**
  - **Validates: Requirements 2.5**

- [ ]* 4.6 Write unit tests for GUI interactions
  - Test window show/hide on toggle
  - Test Escape key cancellation
  - Test Enter key command submission
  - Test error message display
  - _Requirements: 1.1, 1.3, 2.5_

- [x] 5. Implement hotkey manager
  - Create HotkeyManager struct with callback function
  - Implement Register() to register global Windows hotkey
  - Implement Start() to begin listening for hotkey events
  - Implement Stop() to unregister hotkey on shutdown
  - Handle hotkey registration errors (already in use, invalid format)
  - Parse hotkey string format (e.g., "Ctrl+Space", "Alt+Space")
  - _Requirements: 1.1, 1.3_


- [x] 5.1 Write unit tests for hotkey manager
  - Test hotkey registration with valid format
  - Test handling invalid hotkey format
  - Test callback invocation on hotkey press
  - _Requirements: 1.1_

- [x] 6. Implement main application and integration
  - Create App struct to coordinate all components
  - Implement NewApp() to initialize ConfigManager, Executor, GUIManager, HotkeyManager
  - Load configuration file from default location (%APPDATA%\launcher\config.json)
  - Support --config flag to override configuration path
  - Support --hotkey flag to override default hotkey
  - Implement Run() to start hotkey listener and Fyne app
  - Implement Shutdown() for graceful cleanup
  - Handle configuration errors with error dialogs and exit
  - _Requirements: 3.1, 3.3, 6.4_

- [x] 6.1 Write property test for invalid configuration handling
  - **Property 8: Invalid configurations fail gracefully**
  - **Validates: Requirements 3.3**

- [ ] 7. Add error handling and logging
  - Implement error logging to stderr with timestamps
  - Add detailed error messages for all failure scenarios
  - Ensure all errors are user-friendly in GUI
  - Log all execution attempts for audit purposes
  - _Requirements: 2.4, 3.3, 5.3_

- [ ] 8. Create sample configuration and documentation
  - Create example config.json with common Windows applications
  - Add comments in code explaining configuration format
  - Document command-line flags (--config, --hotkey)
  - _Requirements: 4.1_

- [ ] 9. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.
