package hotkey

import (
	"app-launcher/logger"
	"fmt"
	"strings"

	"github.com/moutend/go-hook/pkg/keyboard"
	"github.com/moutend/go-hook/pkg/types"
)

// HotkeyManager manages global keyboard shortcuts
type HotkeyManager struct {
	callback  func()
	modifiers []types.VKCode
	key       types.VKCode
	stopChan  chan struct{}
	keydownCh chan types.KeyboardEvent
	isRunning bool
}

// NewHotkeyManager creates a new hotkey manager with the specified callback
func NewHotkeyManager(callback func()) (*HotkeyManager, error) {
	if callback == nil {
		err := fmt.Errorf("callback function cannot be nil")
		logger.Error("Failed to create HotkeyManager: %v", err)
		return nil, err
	}

	logger.Info("Creating HotkeyManager")
	keydownCh := make(chan types.KeyboardEvent, 100)

	return &HotkeyManager{
		callback:  callback,
		stopChan:  make(chan struct{}),
		keydownCh: keydownCh,
		isRunning: false,
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

	h.modifiers = modifiers
	h.key = key

	logger.Info("Successfully registered hotkey: %s", hotkeyStr)
	return nil
}

// Start begins listening for hotkey events
// This function blocks until Stop() is called
func (h *HotkeyManager) Start() error {
	if h.key == 0 {
		err := fmt.Errorf("hotkey not registered, call Register() first")
		logger.Error("Failed to start hotkey listener: %v", err)
		return err
	}

	logger.Info("Starting hotkey listener")
	h.isRunning = true

	// Install keyboard hook
	if err := keyboard.Install(nil, h.keydownCh); err != nil {
		logger.Error("Failed to install keyboard hook: %v", err)
		return fmt.Errorf("failed to install keyboard hook: %w", err)
	}

	// Listen for hotkey events in a goroutine
	go func() {
		pressedKeys := make(map[types.VKCode]bool)

		for {
			select {
			case event := <-h.keydownCh:
				// Track key state
				if event.Message == types.WM_KEYDOWN || event.Message == types.WM_SYSKEYDOWN {
					pressedKeys[event.VKCode] = true

					// Check if hotkey combination is pressed
					if h.isHotkeyPressed(pressedKeys) {
						logger.Info("Hotkey pressed, invoking callback")
						if h.callback != nil {
							h.callback()
						}
					}
				} else if event.Message == types.WM_KEYUP || event.Message == types.WM_SYSKEYUP {
					delete(pressedKeys, event.VKCode)
				}

			case <-h.stopChan:
				// Stop signal received
				logger.Info("Hotkey listener stopped")
				keyboard.Uninstall()
				return
			}
		}
	}()

	return nil
}

// isHotkeyPressed checks if the registered hotkey combination is currently pressed
func (h *HotkeyManager) isHotkeyPressed(pressedKeys map[types.VKCode]bool) bool {
	// Check if the main key is pressed
	if !pressedKeys[h.key] {
		return false
	}

	// Check if all required modifiers are pressed
	for _, mod := range h.modifiers {
		if !pressedKeys[mod] {
			return false
		}
	}

	// Check that no extra modifiers are pressed
	modifierSet := make(map[types.VKCode]bool)
	for _, mod := range h.modifiers {
		modifierSet[mod] = true
	}

	// Common modifier keys
	allModifiers := []types.VKCode{
		types.VK_LCONTROL, types.VK_RCONTROL,
		types.VK_LMENU, types.VK_RMENU,
		types.VK_LSHIFT, types.VK_RSHIFT,
		types.VK_LWIN, types.VK_RWIN,
	}

	for _, mod := range allModifiers {
		if pressedKeys[mod] && !modifierSet[mod] {
			return false
		}
	}

	return true
}

// Stop unregisters the hotkey and stops listening for events
func (h *HotkeyManager) Stop() {
	logger.Info("Stopping hotkey manager")
	if h.isRunning {
		// Only close the channel if it's not already closed
		select {
		case <-h.stopChan:
			// Already closed
		default:
			close(h.stopChan)
		}
		h.isRunning = false
		logger.Info("Hotkey unregistered")
	}
}

// parseHotkey parses a hotkey string into modifiers and key
// Supported format: "Modifier+Modifier+Key" (e.g., "Ctrl+Alt+L", "Alt+Space")
func parseHotkey(hotkeyStr string) ([]types.VKCode, types.VKCode, error) {
	parts := strings.Split(hotkeyStr, "+")
	if len(parts) < 2 {
		return nil, 0, fmt.Errorf("hotkey must contain at least one modifier and a key")
	}

	var modifiers []types.VKCode
	var key types.VKCode
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

// parseModifier converts a string to a VKCode modifier
func parseModifier(mod string) (types.VKCode, bool) {
	switch strings.ToLower(mod) {
	case "ctrl", "control":
		return types.VK_LCONTROL, true
	case "alt":
		return types.VK_LMENU, true
	case "shift":
		return types.VK_LSHIFT, true
	case "win", "windows", "super", "cmd", "command":
		return types.VK_LWIN, true
	default:
		return 0, false
	}
}

// parseKey converts a string to a VKCode
func parseKey(keyStr string) (types.VKCode, bool) {
	keyStr = strings.ToLower(keyStr)

	// Special keys
	switch keyStr {
	case "space":
		return types.VK_SPACE, true
	case "enter", "return":
		return types.VK_RETURN, true
	case "tab":
		return types.VK_TAB, true
	case "escape", "esc":
		return types.VK_ESCAPE, true
	case "up":
		return types.VK_UP, true
	case "down":
		return types.VK_DOWN, true
	case "left":
		return types.VK_LEFT, true
	case "right":
		return types.VK_RIGHT, true
	}

	// Function keys
	if len(keyStr) >= 2 && keyStr[0] == 'f' {
		switch keyStr {
		case "f1":
			return types.VK_F1, true
		case "f2":
			return types.VK_F2, true
		case "f3":
			return types.VK_F3, true
		case "f4":
			return types.VK_F4, true
		case "f5":
			return types.VK_F5, true
		case "f6":
			return types.VK_F6, true
		case "f7":
			return types.VK_F7, true
		case "f8":
			return types.VK_F8, true
		case "f9":
			return types.VK_F9, true
		case "f10":
			return types.VK_F10, true
		case "f11":
			return types.VK_F11, true
		case "f12":
			return types.VK_F12, true
		}
	}

	// Number keys (0-9)
	if len(keyStr) == 1 && keyStr[0] >= '0' && keyStr[0] <= '9' {
		return types.VKCode('0' + (keyStr[0] - '0')), true
	}

	// Letter keys (A-Z)
	if len(keyStr) == 1 && keyStr[0] >= 'a' && keyStr[0] <= 'z' {
		return types.VKCode('A' + (keyStr[0] - 'a')), true
	}

	return 0, false
}
