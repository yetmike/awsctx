package awsctx

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	ini "gopkg.in/ini.v1"
)

// iniFile represents an INI file backed by gopkg.in/ini.v1.
type iniFile struct {
	path string
	file *ini.File
}

// loadINI reads the file at path.
// If the file doesn't exist, it returns an empty iniFile and no error.
func loadINI(path string) (*iniFile, error) {
	opts := ini.LoadOptions{
		Insensitive:            false,
		IgnoreInlineComment:    true,
		AllowNonUniqueSections: false,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			f := ini.Empty(opts)
			return &iniFile{path: path, file: f}, nil
		}
		return nil, err
	}

	f, err := ini.LoadSources(opts, data)
	if err != nil {
		return nil, fmt.Errorf("parsing INI %s: %w", path, err)
	}

	return &iniFile{path: path, file: f}, nil
}

// save writes the INI file back to f.path with 0600 permissions.
func (f *iniFile) save() error {
	dir := filepath.Dir(f.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var buf bytes.Buffer
	if _, err := f.file.WriteTo(&buf); err != nil {
		return err
	}

	return os.WriteFile(f.path, buf.Bytes(), 0600)
}

// hasSection returns true if the named section exists.
func (f *iniFile) hasSection(name string) bool {
	return f.file.HasSection(name)
}

// getKeys returns a map of key-value pairs for the given section.
func (f *iniFile) getKeys(section string) map[string]string {
	keys := make(map[string]string)
	sec, err := f.file.GetSection(section)
	if err != nil {
		return keys
	}
	for _, k := range sec.Keys() {
		keys[k.Name()] = k.Value()
	}
	return keys
}

// setKey sets a key=value pair in the given section, creating the section if needed.
func (f *iniFile) setKey(section, key, value string) {
	f.file.Section(section).Key(key).SetValue(value)
}

// replaceSection replaces all keys in the section with the given map.
// Keys are written in sorted order.
func (f *iniFile) replaceSection(name string, keys map[string]string) {
	sec := f.file.Section(name)

	// Delete existing keys.
	for _, k := range sec.KeyStrings() {
		sec.DeleteKey(k)
	}

	// Add new keys in sorted order.
	var sortedKeys []string
	for k := range keys {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		sec.Key(k).SetValue(keys[k])
	}
}

// copySection copies keys from srcSection to dstSection.
func (f *iniFile) copySection(srcSection, dstSection string) {
	keys := f.getKeys(srcSection)
	f.replaceSection(dstSection, keys)
}

// isEmpty returns true if the file has no sections (besides the implicit default).
func (f *iniFile) isEmpty() bool {
	sections := f.file.SectionStrings()
	// ini.v1 always includes a "DEFAULT" section; check if there's anything beyond it
	for _, s := range sections {
		if s != "DEFAULT" {
			return false
		}
	}
	// Also check if the default section has any keys
	return len(f.file.Section("DEFAULT").Keys()) == 0
}

// deleteSection removes the entire section.
func (f *iniFile) deleteSection(name string) {
	f.file.DeleteSection(name)
}
