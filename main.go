package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"app-launcher/config"
	"app-launcher/executor"
	"app-launcher/gui"
	"app-launcher/hotkey"

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
	// Initialize ConfigManager
	configManager, err := config.NewConfigManager(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}

	// Load configuration
	if err := configManager.Load(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize Executor
	exec := executor.NewExecutor(configManager)

	// Initialize GUIManager
	guiManager := gui.NewGUIManager(exec)
	guiManager.Initialize()

	// Initialize HotkeyManager with toggle callback
	hotkeyManager, err := hotkey.NewHotkeyManager(func() {
		guiManager.Toggle()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create hotkey manager: %w", err)
	}

	// Register the hotkey
	if err := hotkeyManager.Register(hotkeyStr); err != nil {
		return nil, fmt.Errorf("failed to register hotkey: %w", err)
	}

	return &App{
		config:   configManager,
		executor: exec,
		gui:      guiManager,
		hotkey:   hotkeyManager,
	}, nil
}

// Run starts the hotkey listener and Fyne application
func (a *App) Run() error {
	// Start hotkey listener
	if err := a.hotkey.Start(); err != nil {
		return fmt.Errorf("failed to start hotkey listener: %w", err)
	}

	// Run the GUI (this blocks until the app is closed)
	a.gui.Run()

	return nil
}

// Shutdown performs graceful cleanup of all components
func (a *App) Shutdown() {
	if a.hotkey != nil {
		a.hotkey.Stop()
	}
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
	configPath := flag.String("config", getDefaultConfigPath(), "Path to configuration file")
	hotkeyStr := flag.String("hotkey", "Alt+Space", "Hotkey to activate launcher (e.g., 'Ctrl+Space', 'Alt+Space')")
	flag.Parse()

	// The hotkey library requires mainthread initialization
	hotkey.Init(func() {
		// Create the app
		app, err := NewApp(*configPath, *hotkeyStr)
		if err != nil {
			// Show error dialog and exit
			log.Printf("Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "Failed to initialize launcher: %v\n", err)

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
			log.Printf("Error running application: %v\n", err)
			fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
			os.Exit(1)
		}
	})
}
