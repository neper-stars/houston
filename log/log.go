// Package log provides a simple logging abstraction for the houston library.
//
// By default, the library uses a no-op logger that discards all output.
// Users can configure logging by calling SetLogger with their preferred
// implementation.
//
// The package provides built-in support for zerolog via NewZerologAdapter,
// but any logger implementing the Logger interface can be used.
//
// Example with zerolog:
//
//	import (
//	    "os"
//	    "github.com/rs/zerolog"
//	    "github.com/neper-stars/houston/log"
//	)
//
//	func main() {
//	    zlog := zerolog.New(os.Stderr).With().Timestamp().Logger()
//	    log.SetLogger(log.NewZerologAdapter(zlog))
//	    // ... use houston library
//	}
//
// Example with custom logger:
//
//	type MyLogger struct{}
//
//	func (l *MyLogger) Debug(msg string, fields ...log.Field) { /* ... */ }
//	func (l *MyLogger) Info(msg string, fields ...log.Field)  { /* ... */ }
//	func (l *MyLogger) Warn(msg string, fields ...log.Field)  { /* ... */ }
//	func (l *MyLogger) Error(msg string, fields ...log.Field) { /* ... */ }
//
//	func main() {
//	    log.SetLogger(&MyLogger{})
//	}
package log

import (
	"sync"
)

// Field represents a key-value pair for structured logging.
type Field struct {
	Key   string
	Value any
}

// F creates a Field with the given key and value.
// This is a convenience function for creating structured log fields.
//
// Example:
//
//	log.Debug("processing file", log.F("filename", "game.m1"), log.F("size", 1024))
func F(key string, value any) Field {
	return Field{Key: key, Value: value}
}

// Logger defines the interface for logging in the houston library.
// Implementations should handle structured logging with key-value fields.
type Logger interface {
	// Debug logs a message at debug level with optional structured fields.
	Debug(msg string, fields ...Field)

	// Info logs a message at info level with optional structured fields.
	Info(msg string, fields ...Field)

	// Warn logs a message at warn level with optional structured fields.
	Warn(msg string, fields ...Field)

	// Error logs a message at error level with optional structured fields.
	Error(msg string, fields ...Field)
}

var (
	globalLogger Logger = &noopLogger{}
	mu           sync.RWMutex
)

// SetLogger sets the global logger used by the houston library.
// Pass nil to disable logging (uses no-op logger).
//
// This function is safe to call from multiple goroutines.
func SetLogger(l Logger) {
	mu.Lock()
	defer mu.Unlock()
	if l == nil {
		globalLogger = &noopLogger{}
	} else {
		globalLogger = l
	}
}

// GetLogger returns the current global logger.
// This function is safe to call from multiple goroutines.
func GetLogger() Logger {
	mu.RLock()
	defer mu.RUnlock()
	return globalLogger
}

// Debug logs a message at debug level using the global logger.
func Debug(msg string, fields ...Field) {
	GetLogger().Debug(msg, fields...)
}

// Info logs a message at info level using the global logger.
func Info(msg string, fields ...Field) {
	GetLogger().Info(msg, fields...)
}

// Warn logs a message at warn level using the global logger.
func Warn(msg string, fields ...Field) {
	GetLogger().Warn(msg, fields...)
}

// Error logs a message at error level using the global logger.
func Error(msg string, fields ...Field) {
	GetLogger().Error(msg, fields...)
}
