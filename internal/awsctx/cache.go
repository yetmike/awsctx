package awsctx

import (
	"os"
	"path/filepath"
	"strings"
)

func cacheDir() string {
	if d := os.Getenv("XDG_CACHE_HOME"); d != "" {
		return filepath.Join(d, "awsctx")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "awsctx")
}

// readPrevious reads the previous value for swap (-).
func readPrevious(key string) string {
	data, err := os.ReadFile(filepath.Join(cacheDir(), "previous_"+key))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// savePrevious saves the current value before switching, for swap (-).
func savePrevious(key, value string) {
	dir := cacheDir()
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "previous_"+key), []byte(value), 0o644)
}

// readState reads the current active value (for global mode tracking).
func readState(key string) string {
	data, err := os.ReadFile(filepath.Join(cacheDir(), "current_"+key))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// saveState saves the current active value (for global mode tracking).
func saveState(key, value string) {
	dir := cacheDir()
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "current_"+key), []byte(value), 0o644)
}
