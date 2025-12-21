package log

// noopLogger is a logger that discards all output.
// This is the default logger used when no logger is configured.
type noopLogger struct{}

// Noop returns a no-op logger that discards all output.
// This is useful for explicitly disabling logging or for testing.
func Noop() Logger {
	return &noopLogger{}
}

func (l *noopLogger) Debug(msg string, fields ...Field) {}
func (l *noopLogger) Info(msg string, fields ...Field)  {}
func (l *noopLogger) Warn(msg string, fields ...Field)  {}
func (l *noopLogger) Error(msg string, fields ...Field) {}
