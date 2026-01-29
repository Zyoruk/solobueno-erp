package version

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestAppName(t *testing.T) {
	if AppName == "" {
		t.Error("AppName should not be empty")
	}
	if AppName != "Solobueno ERP" {
		t.Errorf("AppName should be 'Solobueno ERP', got '%s'", AppName)
	}
}

func TestInfo(t *testing.T) {
	info := Info()
	if info == "" {
		t.Error("Info() should not return empty string")
	}
	if !strings.Contains(info, Version) {
		t.Errorf("Info() should contain version, got '%s'", info)
	}
	if !strings.Contains(info, AppName) {
		t.Errorf("Info() should contain app name, got '%s'", info)
	}
}

func TestIsPreRelease(t *testing.T) {
	// Since version starts with 0, it should be pre-release
	if !IsPreRelease() {
		t.Error("IsPreRelease() should return true for version starting with 0")
	}
}
