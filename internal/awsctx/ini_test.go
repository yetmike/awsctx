package awsctx

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const testINI = `[default]
region = us-east-1
output = json

[profile dev]
region = us-west-2
output = yaml
role_arn = arn:aws:iam::123456789:role/dev

[profile staging]
region = eu-west-1
`

func TestLoadINI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	os.WriteFile(path, []byte(testINI), 0644)

	f, err := loadINI(path)
	if err != nil {
		t.Fatalf("loadINI failed: %v", err)
	}

	if !f.hasSection("default") {
		t.Error("expected default section")
	}
	if !f.hasSection("profile dev") {
		t.Error("expected profile dev section")
	}
	if !f.hasSection("profile staging") {
		t.Error("expected profile staging section")
	}
}

func TestLoadINI_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent")
	f, err := loadINI(path)
	if err != nil {
		t.Fatalf("loadINI should not fail for missing file: %v", err)
	}
	if f.hasSection("default") {
		t.Error("expected no sections in empty file")
	}
}

func TestHasSection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	os.WriteFile(path, []byte(testINI), 0644)
	f, _ := loadINI(path)

	if !f.hasSection("default") {
		t.Error("expected default section to exist")
	}
	if !f.hasSection("profile dev") {
		t.Error("expected profile dev section to exist")
	}
	if f.hasSection("nonexistent") {
		t.Error("expected nonexistent section to not exist")
	}
}

func TestGetKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	os.WriteFile(path, []byte(testINI), 0644)
	f, _ := loadINI(path)

	keys := f.getKeys("profile dev")
	want := map[string]string{
		"region":   "us-west-2",
		"output":   "yaml",
		"role_arn": "arn:aws:iam::123456789:role/dev",
	}

	if !reflect.DeepEqual(keys, want) {
		t.Errorf("getKeys(\"profile dev\") = %v, want %v", keys, want)
	}
}

func TestGetKeys_NonexistentSection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	os.WriteFile(path, []byte(testINI), 0644)
	f, _ := loadINI(path)

	keys := f.getKeys("nonexistent")
	if len(keys) != 0 {
		t.Errorf("expected empty map, got %v", keys)
	}
}

func TestSetKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	os.WriteFile(path, []byte(testINI), 0644)
	f, _ := loadINI(path)

	// Update existing key
	f.setKey("default", "region", "us-west-1")
	if keys := f.getKeys("default"); keys["region"] != "us-west-1" {
		t.Errorf("expected us-west-1, got %s", keys["region"])
	}

	// Add new key to existing section
	f.setKey("default", "newkey", "newval")
	if keys := f.getKeys("default"); keys["newkey"] != "newval" {
		t.Errorf("expected newval, got %s", keys["newkey"])
	}

	// Add new section and key
	f.setKey("newsection", "k", "v")
	if keys := f.getKeys("newsection"); keys["k"] != "v" {
		t.Errorf("expected v, got %s", keys["k"])
	}
}

func TestReplaceSection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	os.WriteFile(path, []byte(testINI), 0644)
	f, _ := loadINI(path)

	newKeys := map[string]string{
		"region": "eu-central-1",
		"custom": "value",
	}
	f.replaceSection("default", newKeys)

	keys := f.getKeys("default")
	if !reflect.DeepEqual(keys, newKeys) {
		t.Errorf("replaceSection failed, got %v, want %v", keys, newKeys)
	}

	// Ensure other sections are still there
	if !f.hasSection("profile dev") {
		t.Error("profile dev section disappeared")
	}
}

func TestReplaceSection_MoreKeysThanOriginal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	os.WriteFile(path, []byte(testINI), 0644)
	f, _ := loadINI(path)

	// default has 2 keys; replace with 5 — this triggered the original slice aliasing bug
	bigKeys := map[string]string{
		"region":      "ap-southeast-1",
		"output":      "table",
		"role_arn":    "arn:aws:iam::999:role/big",
		"mfa_serial":  "arn:aws:iam::999:mfa/user",
		"source_profile": "base",
	}
	f.replaceSection("default", bigKeys)

	got := f.getKeys("default")
	if !reflect.DeepEqual(got, bigKeys) {
		t.Errorf("replaceSection with more keys failed, got %v, want %v", got, bigKeys)
	}

	// Verify sections below are intact
	devKeys := f.getKeys("profile dev")
	wantDev := map[string]string{
		"region":   "us-west-2",
		"output":   "yaml",
		"role_arn": "arn:aws:iam::123456789:role/dev",
	}
	if !reflect.DeepEqual(devKeys, wantDev) {
		t.Errorf("profile dev corrupted after replaceSection, got %v, want %v", devKeys, wantDev)
	}

	stagingKeys := f.getKeys("profile staging")
	wantStaging := map[string]string{
		"region": "eu-west-1",
	}
	if !reflect.DeepEqual(stagingKeys, wantStaging) {
		t.Errorf("profile staging corrupted after replaceSection, got %v, want %v", stagingKeys, wantStaging)
	}
}

func TestReplaceSection_NewSection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	os.WriteFile(path, []byte(testINI), 0644)
	f, _ := loadINI(path)

	newKeys := map[string]string{"key": "val"}
	f.replaceSection("brand-new", newKeys)

	if !f.hasSection("brand-new") {
		t.Error("new section was not created")
	}
	if got := f.getKeys("brand-new"); got["key"] != "val" {
		t.Errorf("expected val, got %s", got["key"])
	}
}

func TestCopySection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	os.WriteFile(path, []byte(testINI), 0644)
	f, _ := loadINI(path)

	f.copySection("profile dev", "default")

	devKeys := f.getKeys("profile dev")
	defaultKeys := f.getKeys("default")

	if !reflect.DeepEqual(devKeys, defaultKeys) {
		t.Errorf("copySection failed, default has %v, dev has %v", defaultKeys, devKeys)
	}
}

func TestDeleteSection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	os.WriteFile(path, []byte(testINI), 0644)
	f, _ := loadINI(path)

	f.deleteSection("profile staging")
	if f.hasSection("profile staging") {
		t.Error("deleteSection failed, section still exists")
	}

	// Other sections should still exist
	if !f.hasSection("default") {
		t.Error("default section disappeared after deleting staging")
	}
	if !f.hasSection("profile dev") {
		t.Error("profile dev section disappeared after deleting staging")
	}
}

func TestSaveRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	os.WriteFile(path, []byte(testINI), 0644)

	f, _ := loadINI(path)
	if err := f.save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Reload and verify keys survived
	f2, _ := loadINI(path)
	for _, sec := range []string{"default", "profile dev", "profile staging"} {
		if !f2.hasSection(sec) {
			t.Errorf("section %q lost after save round-trip", sec)
		}
	}

	devKeys := f2.getKeys("profile dev")
	want := map[string]string{
		"region":   "us-west-2",
		"output":   "yaml",
		"role_arn": "arn:aws:iam::123456789:role/dev",
	}
	if !reflect.DeepEqual(devKeys, want) {
		t.Errorf("profile dev keys changed after save, got %v, want %v", devKeys, want)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "subdir", "config")
	f, _ := loadINI(path) // missing file → empty
	f.setKey("default", "region", "us-east-1")

	if err := f.save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	f2, _ := loadINI(path)
	if keys := f2.getKeys("default"); keys["region"] != "us-east-1" {
		t.Errorf("expected us-east-1 after save to new dir, got %s", keys["region"])
	}
}

func TestPreservesComments(t *testing.T) {
	content := `[default]
# sync with production
region = us-east-1
`
	path := filepath.Join(t.TempDir(), "config")
	os.WriteFile(path, []byte(content), 0644)

	f, _ := loadINI(path)
	f.setKey("default", "region", "us-west-2")
	f.save()

	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "# sync with production") {
		t.Error("comment was lost")
	}
}

func TestSave_FilePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	f, _ := loadINI(path)
	f.setKey("default", "region", "us-east-1")
	f.save()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected 0600 permissions, got %04o", perm)
	}
}
