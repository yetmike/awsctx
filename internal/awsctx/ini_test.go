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

	ini, err := loadINI(path)
	if err != nil {
		t.Fatalf("loadINI failed: %v", err)
	}

	if len(ini.lines) != 11 {
		t.Errorf("expected 11 lines, got %d", len(ini.lines))
	}
}

func TestLoadINI_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent")
	ini, err := loadINI(path)
	if err != nil {
		t.Fatalf("loadINI should not fail for missing file: %v", err)
	}
	if len(ini.lines) != 0 {
		t.Errorf("expected 0 lines, got %d", len(ini.lines))
	}
}

func TestSectionRange(t *testing.T) {
	ini := &iniFile{lines: strings.Split(strings.TrimSuffix(testINI, "\n"), "\n")}

	tests := []struct {
		name      string
		wantStart int
		wantEnd   int
		wantFound bool
	}{
		{"default", 0, 3, true},
		{"profile dev", 4, 8, true},
		{"profile staging", 9, 10, true},
		{"nonexistent", -1, -1, false},
	}

	for _, tt := range tests {
		start, end, found := ini.sectionRange(tt.name)
		if found != tt.wantFound || start != tt.wantStart || end != tt.wantEnd {
			t.Errorf("sectionRange(%q) = (%v, %v, %v), want (%v, %v, %v)",
				tt.name, start, end, found, tt.wantStart, tt.wantEnd, tt.wantFound)
		}
	}
}

func TestGetKeys(t *testing.T) {
	ini := &iniFile{lines: strings.Split(strings.TrimSuffix(testINI, "\n"), "\n")}

	keys := ini.getKeys("profile dev")
	want := map[string]string{
		"region":   "us-west-2",
		"output":   "yaml",
		"role_arn": "arn:aws:iam::123456789:role/dev",
	}

	if !reflect.DeepEqual(keys, want) {
		t.Errorf("getKeys(\"profile dev\") = %v, want %v", keys, want)
	}
}

func TestSetKey(t *testing.T) {
	ini := &iniFile{lines: strings.Split(strings.TrimSuffix(testINI, "\n"), "\n")}

	// Update existing key
	ini.setKey("default", "region", "us-west-1")
	if keys := ini.getKeys("default"); keys["region"] != "us-west-1" {
		t.Errorf("expected us-west-1, got %s", keys["region"])
	}

	// Add new key to existing section
	ini.setKey("default", "newkey", "newval")
	if keys := ini.getKeys("default"); keys["newkey"] != "newval" {
		t.Errorf("expected newval, got %s", keys["newkey"])
	}

	// Add new section and key
	ini.setKey("newsection", "k", "v")
	if keys := ini.getKeys("newsection"); keys["k"] != "v" {
		t.Errorf("expected v, got %s", keys["k"])
	}
}

func TestReplaceSection(t *testing.T) {
	ini := &iniFile{lines: strings.Split(strings.TrimSuffix(testINI, "\n"), "\n")}

	newKeys := map[string]string{
		"region": "eu-central-1",
		"custom": "value",
	}
	ini.replaceSection("default", newKeys)

	keys := ini.getKeys("default")
	if !reflect.DeepEqual(keys, newKeys) {
		t.Errorf("replaceSection failed, got %v, want %v", keys, newKeys)
	}

	// Ensure other sections are still there
	if !ini.hasSection("profile dev") {
		t.Error("profile dev section disappeared")
	}
}

func TestCopySection(t *testing.T) {
	ini := &iniFile{lines: strings.Split(strings.TrimSuffix(testINI, "\n"), "\n")}

	ini.copySection("profile dev", "default")

	devKeys := ini.getKeys("profile dev")
	defaultKeys := ini.getKeys("default")

	if !reflect.DeepEqual(devKeys, defaultKeys) {
		t.Errorf("copySection failed, default has %v, dev has %v", defaultKeys, devKeys)
	}
}

func TestDeleteSection(t *testing.T) {
	ini := &iniFile{lines: strings.Split(strings.TrimSuffix(testINI, "\n"), "\n")}

	ini.deleteSection("profile staging")
	if ini.hasSection("profile staging") {
		t.Error("deleteSection failed, section still exists")
	}
}

func TestSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	ini := &iniFile{path: path, lines: strings.Split(strings.TrimSuffix(testINI, "\n"), "\n")}

	if err := ini.save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	if string(data) != testINI {
		t.Errorf("saved content mismatch\ngot:\n%s\nwant:\n%s", string(data), testINI)
	}
}

func TestPreservesComments(t *testing.T) {
	content := `[default]
# sync with production
region = us-east-1
`
	path := filepath.Join(t.TempDir(), "config")
	os.WriteFile(path, []byte(content), 0644)

	ini, _ := loadINI(path)
	ini.setKey("default", "region", "us-west-2")
	ini.save()

	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "# sync with production") {
		t.Error("comment was lost")
	}
}
