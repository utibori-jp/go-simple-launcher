package gui

import (
	"app-launcher/executor"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// GUIManager manages the Fyne-based graphical user interface
type GUIManager struct {
	app        fyne.App
	window     fyne.Window
	entry      *widget.Entry
	errorLabel *widget.Label
	executor   *executor.Executor
	visible    bool
}

// NewGUIManager creates a new GUIManager with the specified executor
func NewGUIManager(exec *executor.Executor) *GUIManager {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(("Enter command..."))

	errorLabel := widget.NewLabel("")
	errorLabel.Hide()

	return &GUIManager{
		executor:   exec,
		visible:    false,
		entry:      entry,
		errorLabel: errorLabel,
	}
}

// Initialize creates the Fyne window with text entry widget and configures it
func (g *GUIManager) Initialize() {
	// Create Fyne application
	g.app = app.New()

	// Create window
	g.window = g.app.NewWindow("Launcher")

	// Set up Enter key handler to execute commands
	g.entry.OnSubmitted = func(text string) {
		g.handleCommandSubmit(text)
	}

	// Set up key event handler for Escape
	g.window.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyEscape {
			g.Hide()
		}
	})

	// Create container with entry and error label
	content := container.NewVBox(
		g.entry,
		g.errorLabel,
	)

	g.window.SetContent(content)

	// Configure window to be always on top and centered
	g.window.Resize(fyne.NewSize(400, 100))
	g.window.CenterOnScreen()
	g.window.SetFixedSize(true)

	// Don't show window initially
	g.visible = false
}

// Show displays the window and focuses the input field
func (g *GUIManager) Show() {
	if !g.visible {
		g.window.Show()
		g.visible = true

		// Clear previous input and error
		g.entry.SetText("")
		g.errorLabel.Hide()

		// Focus the input field
		g.window.Canvas().Focus(g.entry)
	}
}

// Hide hides the window
func (g *GUIManager) Hide() {
	if g.visible {
		g.window.Hide()
		g.visible = false
	}
}

// Toggle toggles window visibility (for hotkey activation)
func (g *GUIManager) Toggle() {
	if g.visible {
		g.Hide()
	} else {
		g.Show()
	}
}

// ShowError displays an error message in the GUI
func (g *GUIManager) ShowError(message string) {
	g.errorLabel.SetText(message)
	g.errorLabel.Show()
}

// handleCommandSubmit processes command submission when Enter is pressed
func (g *GUIManager) handleCommandSubmit(commandName string) {
	// Clear any previous error
	g.errorLabel.Hide()

	// Execute the command
	err := g.executor.Execute(commandName)

	if err != nil {
		// Show error message and keep window visible
		g.ShowError(err.Error())
	} else {
		// Successful launch - hide the window
		g.Hide()
	}
}

// Run starts the Fyne application event loop
func (g *GUIManager) Run() {
	g.app.Run()
}
