package awsctx

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// iniFile represents an INI file as a slice of raw lines (preserves formatting).
type iniFile struct {
	path  string
	lines []string
}

// loadINI reads the file at path into lines.
// If the file doesn't exist, it returns an empty iniFile and no error.
func loadINI(path string) (*iniFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &iniFile{path: path, lines: []string{}}, nil
		}
		return nil, err
	}

	content := string(data)
	if content == "" {
		return &iniFile{path: path, lines: []string{}}, nil
	}

	// Split lines, removing trailing newline to avoid empty last string
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	return &iniFile{path: path, lines: lines}, nil
}

// save writes lines back to the file at f.path.
func (f *iniFile) save() error {
	dir := filepath.Dir(f.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	content := strings.Join(f.lines, "\n")
	if len(f.lines) > 0 {
		content += "\n"
	}

	// Use 0600 as credentials may contain secrets
	return os.WriteFile(f.path, []byte(content), 0600)
}

// sectionRange finds the line index of the [name] header (start) and the
// last line of that section's body (end, inclusive).
func (f *iniFile) sectionRange(name string) (start, end int, found bool) {
	header := "[" + name + "]"
	start = -1
	for i, line := range f.lines {
		if strings.TrimSpace(line) == header {
			start = i
			break
		}
	}

	if start == -1 {
		return -1, -1, false
	}

	end = start
	for i := start + 1; i < len(f.lines); i++ {
		line := strings.TrimSpace(f.lines[i])
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			break
		}
		end = i
	}

	return start, end, true
}

// hasSection returns true if sectionRange finds the section.
func (f *iniFile) hasSection(name string) bool {
	_, _, found := f.sectionRange(name)
	return found
}

// getKeys returns a map of key-value pairs for the given section.
func (f *iniFile) getKeys(section string) map[string]string {
	keys := make(map[string]string)
	start, end, found := f.sectionRange(section)
	if !found {
		return keys
	}

	for i := start + 1; i <= end; i++ {
		line := strings.TrimSpace(f.lines[i])
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			keys[k] = v
		}
	}

	return keys
}

// setKey replaces or appends a key=value pair in the given section.
func (f *iniFile) setKey(section, key, value string) {
	start, end, found := f.sectionRange(section)
	newLine := fmt.Sprintf("%s = %s", key, value)

	if found {
		for i := start + 1; i <= end; i++ {
			line := strings.TrimSpace(f.lines[i])
			if strings.HasPrefix(line, key) {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
					f.lines[i] = newLine
					return
				}
			}
		}
		// Key not found in section, append it
		f.lines = append(f.lines[:end+1], append([]string{newLine}, f.lines[end+1:]...)...)
	} else {
		// Section not found, append it
		if len(f.lines) > 0 && f.lines[len(f.lines)-1] != "" {
			f.lines = append(f.lines, "")
		}
		f.lines = append(f.lines, "["+section+"]", newLine)
	}
}

// replaceSection replaces all keys in the section with the given map.
func (f *iniFile) replaceSection(name string, keys map[string]string) {
	start, end, found := f.sectionRange(name)

	var newBody []string
	var sortedKeys []string
	for k := range keys {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	for _, k := range sortedKeys {
		newBody = append(newBody, fmt.Sprintf("%s = %s", k, keys[k]))
	}

	if found {
		// Replace body lines
		newLines := append(f.lines[:start+1], newBody...)
		f.lines = append(newLines, f.lines[end+1:]...)
	} else {
		// Create section at the end
		if len(f.lines) > 0 && f.lines[len(f.lines)-1] != "" {
			f.lines = append(f.lines, "")
		}
		f.lines = append(f.lines, "["+name+"]")
		f.lines = append(f.lines, newBody...)
	}
}

// copySection copies keys from srcSection to dstSection.
func (f *iniFile) copySection(srcSection, dstSection string) {
	keys := f.getKeys(srcSection)
	f.replaceSection(dstSection, keys)
}

// deleteSection removes the entire section.
func (f *iniFile) deleteSection(name string) {
	start, end, found := f.sectionRange(name)
	if !found {
		return
	}

	// Also remove leading blank lines
	for start > 0 && strings.TrimSpace(f.lines[start-1]) == "" {
		start--
	}

	f.lines = append(f.lines[:start], f.lines[end+1:]...)
}
