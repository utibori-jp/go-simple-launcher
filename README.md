# Application Launcher

A lightweight, keyboard-driven application launcher for Windows built with Go and Fyne. Launch your favorite applications quickly using simple text commands activated by a global hotkey.

## Features

- **Global Hotkey Activation**: Press `Alt+Space` (default) to bring up the launcher from anywhere
- **Simple Command Interface**: Type a command name and press Enter to launch applications
- **JSON Configuration**: Easy-to-edit configuration file for defining custom commands
- **Non-blocking Execution**: Applications launch immediately without blocking the launcher
- **Error Handling**: Clear error messages for invalid commands or launch failures

## Installation

1. Download the latest `launcher.exe` from the releases page
2. Place it in a convenient location (e.g., `C:\Program Files\Launcher\`)
3. Create your configuration file (see Configuration section below)
4. Run `launcher.exe`

## Quick Start

1. **Create a configuration file** at `%APPDATA%\launcher\config.json`:

```json
{
  "commands": {
    "chrome": {
      "path": "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
      "args": []
    },
    "vscode": {
      "path": "C:\\Program Files\\Microsoft VS Code\\Code.exe",
      "args": ["-n"]
    },
    "notepad": {
      "path": "C:\\Windows\\System32\\notepad.exe",
      "args": []
    }
  }
}
```

2. **Run the launcher**:
```cmd
launcher.exe
```

3. **Use the launcher**:
   - Press `Alt+Space` to open the launcher window
   - Type a command name (e.g., `chrome`)
   - Press `Enter` to launch the application
   - Press `Escape` to cancel and close the window

## Command-Line Flags

The launcher supports the following command-line flags:

### `--config`

Specify a custom path to the configuration file.

**Default**: `%APPDATA%\launcher\config.json`

**Example**:
```cmd
launcher.exe --config="C:\MyConfigs\launcher-config.json"
```

### `--hotkey`

Specify a custom hotkey to activate the launcher.

**Default**: `Alt+Space`

**Supported Formats**:
- `Alt+Space`
- `Ctrl+Space`
- `Ctrl+Alt+L`
- `Shift+Alt+Space`

**Example**:
```cmd
launcher.exe --hotkey="Ctrl+Alt+L"
```

### Combined Usage

You can combine multiple flags:

```cmd
launcher.exe --config="C:\custom\config.json" --hotkey="Ctrl+Space"
```

## Configuration

### Configuration File Format

The configuration file is a JSON file with the following structure:

```json
{
  "commands": {
    "command-name": {
      "path": "C:\\Path\\To\\Application.exe",
      "args": ["arg1", "arg2"]
    }
  }
}
```

### Configuration Fields

- **commands**: Object containing all command definitions
  - **command-name**: The name you'll type in the launcher (e.g., "chrome", "vscode")
    - **path**: Absolute path to the executable file
      - Must be a non-empty string
      - Use double backslashes (`\\`) in Windows paths
      - Forward slashes (`/`) are automatically converted to backslashes
    - **args**: Array of command-line arguments to pass to the application
      - Can be an empty array `[]` if no arguments are needed
      - Each argument is a separate string in the array

### Example Configuration

A comprehensive example configuration with common Windows applications:

```json
{
  "commands": {
    "chrome": {
      "path": "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
      "args": []
    },
    "firefox": {
      "path": "C:\\Program Files\\Mozilla Firefox\\firefox.exe",
      "args": []
    },
    "vscode": {
      "path": "C:\\Program Files\\Microsoft VS Code\\Code.exe",
      "args": ["-n"]
    },
    "cmd": {
      "path": "C:\\Windows\\System32\\cmd.exe",
      "args": []
    },
    "powershell": {
      "path": "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
      "args": []
    },
    "notepad": {
      "path": "C:\\Windows\\System32\\notepad.exe",
      "args": []
    },
    "calc": {
      "path": "C:\\Windows\\System32\\calc.exe",
      "args": []
    },
    "explorer": {
      "path": "C:\\Windows\\explorer.exe",
      "args": []
    },
    "paint": {
      "path": "C:\\Windows\\System32\\mspaint.exe",
      "args": []
    },
    "wordpad": {
      "path": "C:\\Program Files\\Windows NT\\Accessories\\wordpad.exe",
      "args": []
    }
  }
}
```

### Configuration File Location

**Default Location**: `%APPDATA%\launcher\config.json`

On most Windows systems, this expands to:
```
C:\Users\<YourUsername>\AppData\Roaming\launcher\config.json
```

**Custom Location**: Use the `--config` flag to specify a different location.

### Editing Configuration

1. Close the launcher application
2. Open the configuration file in any text editor
3. Add, remove, or modify command entries
4. Save the file
5. Restart the launcher

**Note**: The launcher does not support hot-reload. You must restart the application for configuration changes to take effect.

## Usage

### Basic Workflow

1. **Activate**: Press the configured hotkey (default: `Alt+Space`)
2. **Enter Command**: Type the command name (e.g., `chrome`)
3. **Execute**: Press `Enter` to launch the application
4. **Cancel**: Press `Escape` to close without launching

### Keyboard Shortcuts

- **Hotkey** (default `Alt+Space`): Toggle launcher window visibility
- **Enter**: Execute the entered command
- **Escape**: Close the launcher window without executing

### Error Messages

The launcher provides clear error messages for common issues:

- **"Command 'xyz' not found"**: The command name doesn't exist in your configuration
- **"Failed to launch 'xyz': ..."**: The application couldn't be started (check the path)
- **"Configuration file not found at ..."**: The config file doesn't exist at the specified location
- **"Invalid configuration file: ..."**: The JSON syntax is incorrect
- **"Cannot register hotkey: already in use"**: Another application is using the same hotkey

## Building from Source

### Prerequisites

- Go 1.23 or later
- Windows operating system

### Project Structure

```
app-launcher/
├── config/          # Configuration management
├── executor/        # Application execution logic
├── gui/             # Fyne-based GUI components
├── hotkey/          # Global hotkey registration
├── logger/          # Logging utilities
├── testdata/        # Test fixtures
├── main.go          # Application entry point
├── config.json      # Example configuration
└── go.mod           # Go module dependencies
```

### Dependencies

The project uses the following main dependencies:

```cmd
go get fyne.io/fyne/v2
go get github.com/moutend/go-hook
go get github.com/leanovate/gopter  # For property-based testing
```

Or simply run:

```cmd
go mod download
```

### Build

```cmd
go build -o launcher.exe
```

### Run Tests

Run all tests including unit tests and property-based tests:

```cmd
go test ./...
```

Run tests with verbose output:

```cmd
go test -v ./...
```

Run tests for a specific package:

```cmd
go test ./gui
go test ./config
go test ./executor
```

## Troubleshooting

### Launcher doesn't start

- Check that the configuration file exists at the expected location
- Verify the JSON syntax is valid (use a JSON validator)
- Check the console output for error messages

### Hotkey doesn't work

- Ensure another application isn't using the same hotkey
- Try a different hotkey combination using the `--hotkey` flag
- Run the launcher as administrator if needed

### Application doesn't launch

- Verify the executable path in the configuration is correct
- Check that the executable file exists at the specified path
- Ensure you have permission to execute the application
- Check the launcher logs for detailed error messages

### Configuration changes not taking effect

- Restart the launcher application (configuration is loaded at startup only)
- Verify the configuration file syntax is valid JSON

## Logging

The launcher logs all operations to stderr with timestamps in the format `[YYYY-MM-DD HH:MM:SS] LEVEL: message`:

- Application startup and initialization
- Configuration loading
- Command execution attempts
- Errors and warnings

**Log Levels**:
- `INFO`: Normal operations and status messages
- `WARN`: Warning messages (non-critical issues)
- `ERROR`: Error messages (operation failures)
- `FATAL`: Critical errors that cause the application to exit

**Example log output**:
```
[2025-12-07 09:31:30] INFO: Application launcher starting
[2025-12-07 09:31:30] INFO: Configuration path: C:\Users\...\config.json
[2025-12-07 09:31:30] INFO: Hotkey: Alt+Space
[2025-12-07 09:31:30] INFO: Successfully launched application for command 'chrome' (PID: 12345)
```

To capture logs to a file:

```cmd
launcher.exe 2> launcher.log
```

## Security Considerations

- **Validate Paths**: Only add trusted applications to your configuration
- **No Shell Interpretation**: Commands are executed directly without shell interpretation
- **Audit Logging**: All execution attempts are logged for audit purposes
- **Manual Configuration**: Configuration must be edited manually (no remote updates)

## Limitations

- **Windows Only**: Currently supports Windows only (macOS/Linux support planned)
- **No Hot-Reload**: Configuration changes require application restart
- **No Built-in Editor**: Configuration must be edited with external text editor
- **No Command History**: Previous commands are not saved or suggested

## Future Enhancements

Planned features for future versions:

- Configuration hot-reload
- Built-in configuration editor GUI
- Command history and autocomplete
- Command aliases
- Environment variable substitution
- Working directory specification per command
- Visual feedback for running applications

## License

[Add your license information here]

## Contributing

[Add contribution guidelines here]

## Support

For issues, questions, or suggestions, please [add contact/support information here].
