package observability

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func init() {
	// Constitution XI requires the field name "timestamp"; zerolog defaults to "time".
	zerolog.TimestampFieldName = "timestamp"
}

// zerologLogger implements Logger via zerolog (constitution's named Phase 1
// logging implementation).
type zerologLogger struct {
	z zerolog.Logger
}

// New creates a Logger configured for the given environment ("dev", "test",
// "staging", "prod"). dev/test get a human-readable console writer;
// everything else gets structured JSON. Level follows the constitution's
// table: dev=debug, test=warn, staging/prod=info, default=info.
func New(env string) Logger {
	var w io.Writer = os.Stdout
	if env == "dev" || env == "test" {
		w = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}

	level := zerolog.InfoLevel
	switch env {
	case "dev":
		level = zerolog.DebugLevel
	case "test":
		level = zerolog.WarnLevel
	}

	z := zerolog.New(w).Level(level).With().Timestamp().Logger()
	return &zerologLogger{z: z}
}

func (l *zerologLogger) Debug(msg string, fields ...Field) { withFields(l.z.Debug(), fields).Msg(msg) }
func (l *zerologLogger) Info(msg string, fields ...Field)  { withFields(l.z.Info(), fields).Msg(msg) }
func (l *zerologLogger) Warn(msg string, fields ...Field)  { withFields(l.z.Warn(), fields).Msg(msg) }
func (l *zerologLogger) Error(msg string, fields ...Field) { withFields(l.z.Error(), fields).Msg(msg) }

func (l *zerologLogger) With(fields ...Field) Logger {
	ctx := l.z.With()
	for _, f := range fields {
		ctx = ctx.Interface(f.Key, f.Value)
	}
	return &zerologLogger{z: ctx.Logger()}
}

func withFields(e *zerolog.Event, fields []Field) *zerolog.Event {
	for _, f := range fields {
		e = e.Interface(f.Key, f.Value)
	}
	return e
}

var _ Logger = (*zerologLogger)(nil)
