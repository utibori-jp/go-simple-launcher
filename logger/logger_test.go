package logger

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"time"
)

// TestLoggerInfo tests the Info logging method
func TestLoggerInfo(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a test logger that writes to the buffer
	testLogger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	// Log an info message
	testLogger.Info("Test info message")

	// Verify the output
	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("Expected log to contain 'INFO', got: %s", output)
	}
	if !strings.Contains(output, "Test info message") {
		t.Errorf("Expected log to contain 'Test info message', got: %s", output)
	}

	// Verify timestamp format (YYYY-MM-DD HH:MM:SS)
	if !strings.Contains(output, time.Now().Format("2006-01-02")) {
		t.Errorf("Expected log to contain today's date, got: %s", output)
	}
}

// TestLoggerError tests the Error logging method
func TestLoggerError(t *testing.T) {
	var buf bytes.Buffer
	testLogger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	testLogger.Error("Test error message")

	output := buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Errorf("Expected log to contain 'ERROR', got: %s", output)
	}
	if !strings.Contains(output, "Test error message") {
		t.Errorf("Expected log to contain 'Test error message', got: %s", output)
	}
}

// TestLoggerWarn tests the Warn logging method
func TestLoggerWarn(t *testing.T) {
	var buf bytes.Buffer
	testLogger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	testLogger.Warn("Test warning message")

	output := buf.String()
	if !strings.Contains(output, "WARN") {
		t.Errorf("Expected log to contain 'WARN', got: %s", output)
	}
	if !strings.Contains(output, "Test warning message") {
		t.Errorf("Expected log to contain 'Test warning message', got: %s", output)
	}
}

// TestLoggerWithFormatting tests logging with format strings
func TestLoggerWithFormatting(t *testing.T) {
	var buf bytes.Buffer
	testLogger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	testLogger.Info("User %s logged in with ID %d", "Alice", 123)

	output := buf.String()
	if !strings.Contains(output, "User Alice logged in with ID 123") {
		t.Errorf("Expected formatted message, got: %s", output)
	}
}

// TestPackageLevelFunctions tests that package-level functions exist and can be called
// Note: We don't capture output here because the default logger is initialized at package init
// and capturing stderr in tests is unreliable. The Logger method tests already verify the logic.
func TestPackageLevelFunctions(t *testing.T) {
	// Just verify the functions can be called without panicking
	// The actual logging behavior is tested via Logger methods
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Package level functions panicked: %v", r)
		}
	}()

	// These will write to stderr but we're just checking they don't panic
	Info("Package level info test")
	Error("Package level error test")
	Warn("Package level warn test")
}

// TestTimestampFormat tests that the timestamp is in the correct format
func TestTimestampFormat(t *testing.T) {
	var buf bytes.Buffer
	testLogger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	testLogger.Info("Timestamp test")

	output := buf.String()

	// Extract timestamp from output (format: [YYYY-MM-DD HH:MM:SS])
	if !strings.HasPrefix(output, "[") {
		t.Errorf("Expected log to start with '[', got: %s", output)
	}

	// Check for date format YYYY-MM-DD
	now := time.Now()
	expectedDate := now.Format("2006-01-02")
	if !strings.Contains(output, expectedDate) {
		t.Errorf("Expected log to contain date %s, got: %s", expectedDate, output)
	}

	// Check for time format HH:MM:SS (at least the hour part)
	expectedHour := now.Format("15")
	if !strings.Contains(output, expectedHour) {
		t.Errorf("Expected log to contain hour %s, got: %s", expectedHour, output)
	}
}

// TestMultipleLogLevels tests logging at different levels
func TestMultipleLogLevels(t *testing.T) {
	var buf bytes.Buffer
	testLogger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	testLogger.Info("Info message")
	testLogger.Warn("Warn message")
	testLogger.Error("Error message")

	output := buf.String()

	// Verify all three log levels are present
	if !strings.Contains(output, "INFO") {
		t.Error("Expected output to contain INFO level")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("Expected output to contain WARN level")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Expected output to contain ERROR level")
	}

	// Verify all three messages are present
	if !strings.Contains(output, "Info message") {
		t.Error("Expected output to contain 'Info message'")
	}
	if !strings.Contains(output, "Warn message") {
		t.Error("Expected output to contain 'Warn message'")
	}
	if !strings.Contains(output, "Error message") {
		t.Error("Expected output to contain 'Error message'")
	}
}

// TestEmptyMessage tests logging with an empty message
func TestEmptyMessage(t *testing.T) {
	var buf bytes.Buffer
	testLogger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	testLogger.Info("")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("Expected log to contain 'INFO' even with empty message, got: %s", output)
	}
}

// TestSpecialCharacters tests logging with special characters
func TestSpecialCharacters(t *testing.T) {
	var buf bytes.Buffer
	testLogger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	specialMsg := "Test with special chars: !@#$%&*()_+-=[]{}|;',./<>?"
	testLogger.Info(specialMsg)

	output := buf.String()
	// Check for the key parts of the message
	if !strings.Contains(output, "Test with special chars") {
		t.Errorf("Expected log to contain 'Test with special chars', got: %s", output)
	}
	if !strings.Contains(output, "INFO") {
		t.Errorf("Expected log to contain 'INFO', got: %s", output)
	}
}

// TestUnicodeCharacters tests logging with Unicode characters
func TestUnicodeCharacters(t *testing.T) {
	var buf bytes.Buffer
	testLogger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	unicodeMsg := "Unicode test: 你好世界 🚀 Привет мир"
	testLogger.Info(unicodeMsg)

	output := buf.String()
	if !strings.Contains(output, unicodeMsg) {
		t.Errorf("Expected log to contain Unicode characters, got: %s", output)
	}
}

// TestLongMessage tests logging with a very long message
func TestLongMessage(t *testing.T) {
	var buf bytes.Buffer
	testLogger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	longMsg := strings.Repeat("A", 1000)
	testLogger.Info(longMsg)

	output := buf.String()
	if !strings.Contains(output, longMsg) {
		t.Error("Expected log to contain the full long message")
	}
}

// TestMultipleFormatArguments tests logging with multiple format arguments
func TestMultipleFormatArguments(t *testing.T) {
	var buf bytes.Buffer
	testLogger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	testLogger.Info("String: %s, Int: %d, Float: %.2f, Bool: %t", "test", 42, 3.14, true)

	output := buf.String()
	if !strings.Contains(output, "String: test") {
		t.Error("Expected formatted string")
	}
	if !strings.Contains(output, "Int: 42") {
		t.Error("Expected formatted int")
	}
	if !strings.Contains(output, "Float: 3.14") {
		t.Error("Expected formatted float")
	}
	if !strings.Contains(output, "Bool: true") {
		t.Error("Expected formatted bool")
	}
}

// TestDefaultLoggerInitialization tests that the default logger is initialized
func TestDefaultLoggerInitialization(t *testing.T) {
	if defaultLogger == nil {
		t.Error("Expected defaultLogger to be initialized")
	}
	if defaultLogger.logger == nil {
		t.Error("Expected defaultLogger.logger to be initialized")
	}
}

// BenchmarkLoggerInfo benchmarks the Info logging method
func BenchmarkLoggerInfo(b *testing.B) {
	var buf bytes.Buffer
	testLogger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testLogger.Info("Benchmark test message %d", i)
	}
}

// BenchmarkLoggerError benchmarks the Error logging method
func BenchmarkLoggerError(b *testing.B) {
	var buf bytes.Buffer
	testLogger := &Logger{
		logger: log.New(&buf, "", 0),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testLogger.Error("Benchmark error message %d", i)
	}
}
