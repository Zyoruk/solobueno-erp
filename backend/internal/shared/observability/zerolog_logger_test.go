package observability

import "testing"

func TestNew_AllEnvironmentsProduceAWorkingLogger(t *testing.T) {
	for _, env := range []string{"dev", "test", "staging", "prod", ""} {
		logger := New(env)
		if logger == nil {
			t.Fatalf("New(%q) returned nil", env)
		}
		// Exercise every level and With() - must not panic.
		logger.Debug("debug", Field{Key: "env", Value: env})
		logger.Info("info", Field{Key: "env", Value: env})
		logger.Warn("warn", Field{Key: "env", Value: env})
		logger.Error("error", Field{Key: "env", Value: env})
		child := logger.With(Field{Key: "request_id", Value: "abc"})
		if child == nil {
			t.Fatalf("With() returned nil for env %q", env)
		}
		child.Info("child logger works")
	}
}
