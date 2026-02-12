package awsctx

import (
	"os"
	"testing"
)

func TestPreviousCache(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", dir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	// Empty initially
	if v := readPrevious("profile"); v != "" {
		t.Errorf("expected empty, got %s", v)
	}

	savePrevious("profile", "dev")
	if v := readPrevious("profile"); v != "dev" {
		t.Errorf("expected dev, got %s", v)
	}

	// Overwrite
	savePrevious("profile", "staging")
	if v := readPrevious("profile"); v != "staging" {
		t.Errorf("expected staging, got %s", v)
	}
}

func TestStateCache(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", dir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	if v := readState("profile"); v != "" {
		t.Errorf("expected empty, got %s", v)
	}

	saveState("profile", "dev")
	if v := readState("profile"); v != "dev" {
		t.Errorf("expected dev, got %s", v)
	}

	saveState("region", "us-east-1")
	if v := readState("region"); v != "us-east-1" {
		t.Errorf("expected us-east-1, got %s", v)
	}

	// Profile state not affected by region
	if v := readState("profile"); v != "dev" {
		t.Errorf("expected dev, got %s", v)
	}
}
