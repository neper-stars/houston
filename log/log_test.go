package log

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testLogger captures log messages for testing
type testLogger struct {
	messages []testMessage
}

type testMessage struct {
	level  string
	msg    string
	fields []Field
}

func (l *testLogger) Debug(msg string, fields ...Field) {
	l.messages = append(l.messages, testMessage{"debug", msg, fields})
}

func (l *testLogger) Info(msg string, fields ...Field) {
	l.messages = append(l.messages, testMessage{"info", msg, fields})
}

func (l *testLogger) Warn(msg string, fields ...Field) {
	l.messages = append(l.messages, testMessage{"warn", msg, fields})
}

func (l *testLogger) Error(msg string, fields ...Field) {
	l.messages = append(l.messages, testMessage{"error", msg, fields})
}

func TestSetLogger(t *testing.T) {
	// Save original logger
	original := GetLogger()
	defer SetLogger(original)

	// Test custom logger
	custom := &testLogger{}
	SetLogger(custom)
	assert.Equal(t, custom, GetLogger())

	// Test nil resets to noop
	SetLogger(nil)
	_, ok := GetLogger().(*noopLogger)
	assert.True(t, ok, "nil should set noop logger")
}

func TestGlobalLogFunctions(t *testing.T) {
	// Save original logger
	original := GetLogger()
	defer SetLogger(original)

	custom := &testLogger{}
	SetLogger(custom)

	Debug("debug msg", F("key", "value"))
	Info("info msg", F("count", 42))
	Warn("warn msg")
	Error("error msg", F("err", "something failed"))

	require.Len(t, custom.messages, 4)

	assert.Equal(t, "debug", custom.messages[0].level)
	assert.Equal(t, "debug msg", custom.messages[0].msg)
	assert.Equal(t, "key", custom.messages[0].fields[0].Key)
	assert.Equal(t, "value", custom.messages[0].fields[0].Value)

	assert.Equal(t, "info", custom.messages[1].level)
	assert.Equal(t, "info msg", custom.messages[1].msg)
	assert.Equal(t, 42, custom.messages[1].fields[0].Value)

	assert.Equal(t, "warn", custom.messages[2].level)
	assert.Equal(t, "warn msg", custom.messages[2].msg)

	assert.Equal(t, "error", custom.messages[3].level)
	assert.Equal(t, "error msg", custom.messages[3].msg)
}

func TestNoopLogger(t *testing.T) {
	// Noop logger should not panic
	noop := Noop()
	noop.Debug("test", F("key", "value"))
	noop.Info("test")
	noop.Warn("test")
	noop.Error("test")
}

func TestFieldHelper(t *testing.T) {
	f := F("key", "value")
	assert.Equal(t, "key", f.Key)
	assert.Equal(t, "value", f.Value)

	f2 := F("count", 42)
	assert.Equal(t, "count", f2.Key)
	assert.Equal(t, 42, f2.Value)
}

func TestZerologAdapter(t *testing.T) {
	var buf bytes.Buffer
	zlog := zerolog.New(&buf).Level(zerolog.DebugLevel)
	adapter := NewZerologAdapter(zlog)

	adapter.Debug("debug message", F("str", "value"), F("num", 42))
	output := buf.String()

	assert.Contains(t, output, "debug message")
	assert.Contains(t, output, `"str":"value"`)
	assert.Contains(t, output, `"num":42`)

	buf.Reset()
	adapter.Info("info message")
	assert.Contains(t, buf.String(), "info message")

	buf.Reset()
	adapter.Warn("warn message")
	assert.Contains(t, buf.String(), "warn message")

	buf.Reset()
	adapter.Error("error message")
	assert.Contains(t, buf.String(), "error message")
}

func TestZerologFieldTypes(t *testing.T) {
	var buf bytes.Buffer
	zlog := zerolog.New(&buf).Level(zerolog.DebugLevel)
	adapter := NewZerologAdapter(zlog)

	// Test various field types
	adapter.Debug("test",
		F("str", "hello"),
		F("int", 42),
		F("int64", int64(100)),
		F("uint", uint(50)),
		F("float64", 3.14),
		F("bool", true),
		F("bytes", []byte{0x01, 0x02}),
		F("error", os.ErrNotExist),
	)

	output := buf.String()
	assert.Contains(t, output, `"str":"hello"`)
	assert.Contains(t, output, `"int":42`)
	assert.Contains(t, output, `"bool":true`)
}

func TestConcurrentSetLogger(t *testing.T) {
	// Save original logger
	original := GetLogger()
	defer SetLogger(original)

	// Test concurrent access doesn't panic
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				SetLogger(&testLogger{})
				GetLogger().Debug("test")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestZerologIntegration(t *testing.T) {
	// Test that the zerolog adapter works correctly in a real scenario
	var buf bytes.Buffer
	zlog := zerolog.New(&buf).With().Timestamp().Logger().Level(zerolog.DebugLevel)

	// Save original logger
	original := GetLogger()
	defer SetLogger(original)

	SetLogger(NewZerologAdapter(zlog))

	Debug("processing file", F("filename", "game.m1"), F("size", 1024))

	output := buf.String()
	assert.True(t, strings.Contains(output, "processing file"))
	assert.True(t, strings.Contains(output, "game.m1"))
	assert.True(t, strings.Contains(output, "1024"))
}
