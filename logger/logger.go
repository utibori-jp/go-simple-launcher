package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger provides structured logging with timestamps to stderr
type Logger struct {
	logger *log.Logger
}

var defaultLogger *Logger

func init() {
	// Initialize default logger that writes to stderr with timestamps
	defaultLogger = &Logger{
		logger: log.New(os.Stderr, "", 0), // We'll add our own timestamp format
	}
}

// logWithTimestamp formats a log message with timestamp
func (l *Logger) logWithTimestamp(level, format string, v ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, v...)
	l.logger.Printf("[%s] %s: %s", timestamp, level, message)
}

// Info logs an informational message
func (l *Logger) Info(format string, v ...interface{}) {
	l.logWithTimestamp("INFO", format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.logWithTimestamp("ERROR", format, v...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	l.logWithTimestamp("WARN", format, v...)
}

// Fatal logs an error message and exits the program
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.logWithTimestamp("FATAL", format, v...)
	os.Exit(1)
}

// Package-level convenience functions using the default logger

// Info logs an informational message using the default logger
func Info(format string, v ...interface{}) {
	defaultLogger.Info(format, v...)
}

// Error logs an error message using the default logger
func Error(format string, v ...interface{}) {
	defaultLogger.Error(format, v...)
}

// Warn logs a warning message using the default logger
func Warn(format string, v ...interface{}) {
	defaultLogger.Warn(format, v...)
}

// Fatal logs an error message and exits the program using the default logger
func Fatal(format string, v ...interface{}) {
	defaultLogger.Fatal(format, v...)
}
