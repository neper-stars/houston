package log

import (
	"github.com/rs/zerolog"
)

// zerologAdapter wraps a zerolog.Logger to implement the Logger interface.
type zerologAdapter struct {
	logger zerolog.Logger
}

// NewZerologAdapter creates a Logger that wraps a zerolog.Logger.
//
// Example:
//
//	import (
//	    "os"
//	    "github.com/rs/zerolog"
//	    "github.com/neper-stars/houston/log"
//	)
//
//	func main() {
//	    // Create a zerolog logger
//	    zlog := zerolog.New(os.Stderr).With().Timestamp().Logger()
//
//	    // Use it with houston
//	    log.SetLogger(log.NewZerologAdapter(zlog))
//	}
func NewZerologAdapter(logger zerolog.Logger) Logger {
	return &zerologAdapter{logger: logger}
}

func (l *zerologAdapter) Debug(msg string, fields ...Field) {
	event := l.logger.Debug()
	for _, f := range fields {
		event = addField(event, f)
	}
	event.Msg(msg)
}

func (l *zerologAdapter) Info(msg string, fields ...Field) {
	event := l.logger.Info()
	for _, f := range fields {
		event = addField(event, f)
	}
	event.Msg(msg)
}

func (l *zerologAdapter) Warn(msg string, fields ...Field) {
	event := l.logger.Warn()
	for _, f := range fields {
		event = addField(event, f)
	}
	event.Msg(msg)
}

func (l *zerologAdapter) Error(msg string, fields ...Field) {
	event := l.logger.Error()
	for _, f := range fields {
		event = addField(event, f)
	}
	event.Msg(msg)
}

// addField adds a Field to a zerolog event with proper type handling.
func addField(event *zerolog.Event, f Field) *zerolog.Event {
	switch v := f.Value.(type) {
	case string:
		return event.Str(f.Key, v)
	case int:
		return event.Int(f.Key, v)
	case int8:
		return event.Int8(f.Key, v)
	case int16:
		return event.Int16(f.Key, v)
	case int32:
		return event.Int32(f.Key, v)
	case int64:
		return event.Int64(f.Key, v)
	case uint:
		return event.Uint(f.Key, v)
	case uint8:
		return event.Uint8(f.Key, v)
	case uint16:
		return event.Uint16(f.Key, v)
	case uint32:
		return event.Uint32(f.Key, v)
	case uint64:
		return event.Uint64(f.Key, v)
	case float32:
		return event.Float32(f.Key, v)
	case float64:
		return event.Float64(f.Key, v)
	case bool:
		return event.Bool(f.Key, v)
	case error:
		return event.AnErr(f.Key, v)
	case []byte:
		return event.Bytes(f.Key, v)
	default:
		return event.Interface(f.Key, v)
	}
}
