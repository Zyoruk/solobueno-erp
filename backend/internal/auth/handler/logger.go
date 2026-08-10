package handler

import (
	"os"

	"github.com/solobueno/erp/internal/shared/observability"
)

// logger is the package-level structured logger used by writeInternalError.
// Defaults to a safe instance so tests and any caller that never invokes
// SetLogger still get real logging - same "safe default, override when
// something needs it" shape as service.Emailer/LogEmailer in this module.
var logger observability.Logger = observability.New(getEnv("APP_ENV", "dev"))

// SetLogger overrides the package logger. Called once from cmd/server/main.go
// at startup with the real instance, and by tests that need to capture log
// output.
func SetLogger(l observability.Logger) {
	logger = l
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
