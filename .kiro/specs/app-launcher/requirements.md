# Requirements Document

## Introduction

This document specifies the requirements for a simple application launcher built with Golang. The launcher provides a keyboard-activated text input interface that allows users to execute predefined commands to launch applications. The system reads command configurations from a JSON file that users edit manually, without providing built-in configuration management.

## Glossary

- **Launcher**: The application that provides command execution functionality
- **Command**: A text string that maps to an application execution instruction
- **Configuration File**: A JSON file containing command-to-application mappings
- **Text Box**: A GUI input field where users enter commands
- **Hotkey**: A keyboard shortcut that activates the Launcher

## Configuration Schema
The configuration file (`config.json`) MUST adhere to the following structure to support application paths and arguments.

**Example JSON:**
```json
{
  "commands": {
    "browser": {
      "path": "C:\\Program Files\\Mozilla Firefox\\firefox.exe",
      "args": []
    },
    "editor": {
      "path": "C:\\Program Files\\Microsoft VS Code\\Code.exe",
      "args": ["-n"]
    },
    "terminal": {
      "path": "C:\\Windows\\System32\\cmd.exe",
      "args": []
    }
  }
}
```

## Requirements

### Requirement 1

**User Story:** As a user, I want to activate the launcher with a keyboard shortcut, so that I can quickly access the command interface without using a mouse.

#### Acceptance Criteria

1. WHEN the user presses the configured hotkey, THE Launcher SHALL display the Text Box interface
2. THE Launcher SHALL use `Alt+Space` as the default hotkey if no configuration is provided
3. THE Launcher SHALL accept a command-line argument (e.g., `--hotkey`) to override the default hotkey assignment
2. WHEN the Text Box is displayed, THE Launcher SHALL focus the input field for immediate text entry
3. WHEN the user presses the hotkey while the Text Box is already visible, THE Launcher SHALL hide the Text Box
4. WHEN the Text Box is displayed, THE Launcher SHALL position it prominently on the screen

### Requirement 2

**User Story:** As a user, I want to enter a command in the text box and launch the corresponding application, so that I can quickly start programs without navigating menus.

#### Acceptance Criteria

1. WHEN the user types a command and presses Enter, THE Launcher SHALL look up the command in the Configuration File
2. WHEN a valid command is found, THE Launcher SHALL execute the corresponding application
3. WHEN the application is successfully launched, THE Launcher SHALL hide the Text Box
4. WHEN an invalid command is entered, THE Launcher SHALL display an error message to the user
5. WHEN the user presses Escape, THE Launcher SHALL hide the Text Box without executing any command

### Requirement 3

**User Story:** As a user, I want the launcher to read commands from a JSON configuration file, so that I can define which applications to launch for each command.

#### Acceptance Criteria

1. WHEN the Launcher starts, THE Launcher SHALL read the Configuration File from a specified location
2. WHEN the Configuration File is valid JSON, THE Launcher SHALL parse and load all command mappings
3. WHEN the Configuration File is missing or invalid, THE Launcher SHALL display an error message and exit gracefully

### Requirement 4

**User Story:** As a user, I want to manually edit the configuration file with a text editor, so that I can add, remove, or modify command mappings.

#### Acceptance Criteria

1. THE Configuration File SHALL use JSON format matching the Configuration Schema defined above.
2. WHEN a user edits the Configuration File, THE Configuration File SHALL support specifying the application path for each command
3. WHEN a user edits the Configuration File, THE Configuration File SHALL support specifying command-line arguments for each application
4. THE Launcher SHALL NOT provide any built-in interface for editing the Configuration File

### Requirement 5

**User Story:** As a user, I want the launcher to execute applications with their specified arguments, so that I can launch programs with specific configurations.

#### Acceptance Criteria

1. WHEN executing a command, THE Launcher SHALL start the application with the path specified in the Configuration File
2. THE Launcher SHALL normalize the application path (e.g., converting forward slashes `/` to backslashes `\`) to ensure compatibility with the Windows operating system
2. WHEN the Configuration File specifies arguments for a command, THE Launcher SHALL pass those arguments to the application
3. WHEN the application fails to launch, THE Launcher SHALL display an error message with details
4. WHEN launching an application, THE Launcher SHALL not block or wait for the application to complete

### Requirement 6

**User Story:** As a developer, I want the launcher built with Golang and Fyne, initially targeting Windows, so that I can validate the core functionality with a standalone executable.

#### Acceptance Criteria

1. THE Launcher SHALL be implemented using the Go programming language
2. THE Launcher SHALL use the Fyne library for GUI components
3. THE Launcher SHALL use a keyboard shortcut library compatible with Windows API
4. THE Launcher SHALL compile to a standalone executable(.exe)
5. Cross-platform support for other OSs (macOS/Linux) is out of scope for this iteration.
