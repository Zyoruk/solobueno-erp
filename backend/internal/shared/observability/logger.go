// Package observability provides the structured logging adapter used across
// backend modules, per the project constitution's Logger Adapter Pattern
// (Principle XI).
package observability

// Logger is the structured logging interface. Implementations can be
// swapped (e.g. zerolog in Phase 1) without touching call sites.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger // Create child logger with context
}

// Field represents a structured log field.
type Field struct {
	Key   string
	Value any
}
