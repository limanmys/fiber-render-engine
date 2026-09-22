package logger

import (
	"strings"
	"testing"
)

func TestInitLoggerRejectsInvalidDebugValue(t *testing.T) {
	t.Setenv("APP_DEBUG", "not-a-boolean")

	err := InitLogger()
	if err == nil {
		t.Fatal("InitLogger() error = nil, want invalid APP_DEBUG error")
	}
	if !strings.Contains(err.Error(), "invalid APP_DEBUG value") {
		t.Fatalf("InitLogger() error = %q, want invalid APP_DEBUG context", err)
	}
}
