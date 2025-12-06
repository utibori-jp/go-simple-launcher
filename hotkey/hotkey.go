package hotkey

import (
	"app-launcher/logger"
	"fmt"
	"strings"

	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

// HotkeyManager manages global keyboard shortcuts
type HotkeyManager struct {
	callback func()
	hk       *hotkey.Hotkey
	stopChan chan struct{}
}

// NewHotkeyManager creates a new hotkey manager with the specified callback
func NewHotkeyManager(callback func()) (*HotkeyManager, error) {
	if callback == nil {
		err := fmt.Errorf("callback function cannot be nil")
		logger.Error("Failed to create HotkeyManager: %v", err)
		return nil, err
	}

	logger.Info("Creating HotkeyManager")
	return &HotkeyManager{
		callback: callback,
		stopChan: make(chan struct{}),
	}, nil
}

// Register registers a global hotkey with the specified key combination
// Supported formats: "Ctrl+Space", "Alt+Space", "Ctrl+Alt+L", etc.
func (h *HotkeyManager) Register(hotkeyStr string) error {
	logger.Info("Attempting to register hotkey: %s", hotkeyStr)

	if hotkeyStr == "" {
		err := fmt.Errorf("hotkey string cannot be empty")
		logger.Error("Hotkey registration failed: %v", err)
		return err
	}

	// Parse the hotkey string
	modifiers, key, err := parseHotkey(hotkeyStr)
	if err != nil {
		detailedErr := fmt.Errorf("invalid hotkey format: %w", err)
		logger.Error("Hotkey registration failed for '%s': %v", hotkeyStr, detailedErr)
		return detailedErr
	}

	// Create the hotkey
	h.hk = hotkey.New(modifiers, key)

	// Try to register the hotkey
	if err := h.hk.Register(); err != nil {
		detailedErr := fmt.Errorf("failed to register hotkey %s: %w (hotkey may already be in use)", hotkeyStr, err)
		logger.Error("Hotkey registration failed: %v", detailedErr)
		return detailedErr
	}

	logger.Info("Successfully registered hotkey: %s", hotkeyStr)
	return nil
}

// Start begins listening for hotkey events
// This function blocks until Stop() is called
func (h *HotkeyManager) Start() error {
	if h.hk == nil {
		err := fmt.Errorf("hotkey not registered, call Register() first")
		logger.Error("Failed to start hotkey listener: %v", err)
		return err
	}

	logger.Info("Starting hotkey listener")

	// Listen for hotkey events in a goroutine
	go func() {
		for {
			select {
			case <-h.hk.Keydown():
				// Hotkey was pressed, invoke callback
				logger.Info("Hotkey pressed, invoking callback")
				if h.callback != nil {
					h.callback()
				}
			case <-h.stopChan:
				// Stop signal received
				logger.Info("Hotkey listener stopped")
				return
			}
		}
	}()

	return nil
}

// Stop unregisters the hotkey and stops listening for events
func (h *HotkeyManager) Stop() {
	logger.Info("Stopping hotkey manager")
	if h.hk != nil {
		h.hk.Unregister()
		logger.Info("Hotkey unregistered")
	}
	// Only close the channel if it's not already closed
	select {
	case <-h.stopChan:
		// Already closed
	default:
		close(h.stopChan)
	}
}

// parseHotkey parses a hotkey string into modifiers and key
// Supported format: "Modifier+Modifier+Key" (e.g., "Ctrl+Alt+L", "Alt+Space")
func parseHotkey(hotkeyStr string) ([]hotkey.Modifier, hotkey.Key, error) {
	parts := strings.Split(hotkeyStr, "+")
	if len(parts) < 2 {
		return nil, 0, fmt.Errorf("hotkey must contain at least one modifier and a key")
	}

	var modifiers []hotkey.Modifier
	var key hotkey.Key
	var keyFound bool

	for i, part := range parts {
		part = strings.TrimSpace(part)
		isLastPart := i == len(parts)-1

		// Try to parse as modifier first
		if mod, ok := parseModifier(part); ok && !isLastPart {
			modifiers = append(modifiers, mod)
			continue
		}

		// Last part should be the key
		if isLastPart {
			parsedKey, ok := parseKey(part)
			if !ok {
				return nil, 0, fmt.Errorf("unknown key: %s", part)
			}
			key = parsedKey
			keyFound = true
		} else {
			return nil, 0, fmt.Errorf("unknown modifier: %s", part)
		}
	}

	if !keyFound {
		return nil, 0, fmt.Errorf("no key specified")
	}

	if len(modifiers) == 0 {
		return nil, 0, fmt.Errorf("at least one modifier is required")
	}

	return modifiers, key, nil
}

// parseModifier converts a string to a hotkey.Modifier
func parseModifier(mod string) (hotkey.Modifier, bool) {
	switch strings.ToLower(mod) {
	case "ctrl", "control":
		return hotkey.ModCtrl, true
	case "alt":
		return hotkey.ModAlt, true
	case "shift":
		return hotkey.ModShift, true
	case "win", "windows", "super", "cmd", "command":
		return hotkey.ModWin, true
	default:
		return 0, false
	}
}

// parseKey converts a string to a hotkey.Key
func parseKey(keyStr string) (hotkey.Key, bool) {
	keyStr = strings.ToLower(keyStr)

	// Special keys
	switch keyStr {
	case "space":
		return hotkey.KeySpace, true
	case "enter", "return":
		return hotkey.KeyReturn, true
	case "tab":
		return hotkey.KeyTab, true
	case "escape", "esc":
		return hotkey.KeyEscape, true
	case "up":
		return hotkey.KeyUp, true
	case "down":
		return hotkey.KeyDown, true
	case "left":
		return hotkey.KeyLeft, true
	case "right":
		return hotkey.KeyRight, true
	}

	// Function keys
	if len(keyStr) >= 2 && keyStr[0] == 'f' {
		switch keyStr {
		case "f1":
			return hotkey.KeyF1, true
		case "f2":
			return hotkey.KeyF2, true
		case "f3":
			return hotkey.KeyF3, true
		case "f4":
			return hotkey.KeyF4, true
		case "f5":
			return hotkey.KeyF5, true
		case "f6":
			return hotkey.KeyF6, true
		case "f7":
			return hotkey.KeyF7, true
		case "f8":
			return hotkey.KeyF8, true
		case "f9":
			return hotkey.KeyF9, true
		case "f10":
			return hotkey.KeyF10, true
		case "f11":
			return hotkey.KeyF11, true
		case "f12":
			return hotkey.KeyF12, true
		}
	}

	// Number keys
	if len(keyStr) == 1 && keyStr[0] >= '0' && keyStr[0] <= '9' {
		return hotkey.Key(keyStr[0]), true
	}

	// Letter keys (A-Z)
	if len(keyStr) == 1 && keyStr[0] >= 'a' && keyStr[0] <= 'z' {
		return hotkey.Key(keyStr[0] - 'a' + 'A'), true
	}

	return 0, false
}

// Init must be called from the main function to initialize the hotkey system
// This is required by the golang.design/x/hotkey library
func Init(fn func()) {
	mainthread.Init(fn)
}
