package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"app-launcher/config"
	"app-launcher/executor"
	"app-launcher/gui"
	"app-launcher/hotkey"
	"app-launcher/logger"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
)

// App coordinates all components of the launcher
type App struct {
	config   *config.ConfigManager
	executor *executor.Executor
	gui      *gui.GUIManager
	hotkey   *hotkey.HotkeyManager
}

// NewApp creates and initializes a new App with all components
func NewApp(configPath, hotkeyStr string) (*App, error) {
	logger.Info("Initializing application launcher")
	logger.Info("Configuration path: %s", configPath)
	logger.Info("Hotkey: %s", hotkeyStr)

	// Initialize ConfigManager
	configManager, err := config.NewConfigManager(configPath)
	if err != nil {
		logger.Error("Failed to create config manager: %v", err)
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}

	// Load configuration
	if err := configManager.Load(); err != nil {
		logger.Error("Failed to load configuration: %v", err)
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize Executor
	exec := executor.NewExecutor(configManager)
	logger.Info("Executor initialized")

	// Create Fyne application
	fyneApp := app.New()

	// Initialize GUIManager
	guiManager := gui.NewGUIManager(exec, fyneApp)
	guiManager.Initialize()

	// Initialize HotkeyManager with toggle callback
	hotkeyManager, err := hotkey.NewHotkeyManager(func() {
		guiManager.Toggle()
	})
	if err != nil {
		logger.Error("Failed to create hotkey manager: %v", err)
		return nil, fmt.Errorf("failed to create hotkey manager: %w", err)
	}

	// Register the hotkey
	if err := hotkeyManager.Register(hotkeyStr); err != nil {
		logger.Error("Failed to register hotkey: %v", err)
		return nil, fmt.Errorf("failed to register hotkey: %w", err)
	}

	logger.Info("Application launcher initialized successfully")
	return &App{
		config:   configManager,
		executor: exec,
		gui:      guiManager,
		hotkey:   hotkeyManager,
	}, nil
}

// Run starts the hotkey listener and Fyne application
func (a *App) Run() error {
	logger.Info("Starting application")

	// Start hotkey listener
	if err := a.hotkey.Start(); err != nil {
		logger.Error("Failed to start hotkey listener: %v", err)
		return fmt.Errorf("failed to start hotkey listener: %w", err)
	}

	logger.Info("Application running, waiting for hotkey events")
	// Run the GUI (this blocks until the app is closed)
	a.gui.Run()

	logger.Info("Application shutting down")
	return nil
}

// Shutdown performs graceful cleanup of all components
func (a *App) Shutdown() {
	logger.Info("Performing graceful shutdown")
	if a.hotkey != nil {
		a.hotkey.Stop()
	}
	logger.Info("Shutdown complete")
}

// getDefaultConfigPath returns the default configuration file path
// %APPDATA%\launcher\config.json on Windows
func getDefaultConfigPath() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		// Fallback to current directory if APPDATA is not set
		return "config.json"
	}
	return filepath.Join(appData, "launcher", "config.json")
}

func main() {
	// Parse command-line flags
	//
	// Available flags:
	//   --config: Path to the JSON configuration file
	//             Default: %APPDATA%\launcher\config.json
	//             Example: --config="C:\custom\config.json"
	//
	//   --hotkey: Global hotkey to activate the launcher
	//             Default: Alt+Space
	//             Supported formats: "Alt+Space", "Ctrl+Space", "Ctrl+Alt+L", etc.
	//             Example: --hotkey="Ctrl+Alt+L"
	configPath := flag.String("config", getDefaultConfigPath(), "Path to configuration file")
	hotkeyStr := flag.String("hotkey", "Alt+Space", "Hotkey to activate launcher (e.g., 'Ctrl+Space', 'Alt+Space')")
	flag.Parse()

	logger.Info("Application launcher starting")
	logger.Info("Command-line arguments: config=%s, hotkey=%s", *configPath, *hotkeyStr)

	// Create the app
	app, err := NewApp(*configPath, *hotkeyStr)
	if err != nil {
		// Log detailed error information
		logger.Fatal("Failed to initialize launcher: %v", err)

		// Try to show GUI error if possible
		if app != nil && app.gui != nil {
			dialog.ShowError(err, nil)
		}

		os.Exit(1)
	}

	// Ensure cleanup on exit
	defer app.Shutdown()

	// Run the application
	if err := app.Run(); err != nil {
		logger.Fatal("Application error: %v", err)
		os.Exit(1)
	}
}
