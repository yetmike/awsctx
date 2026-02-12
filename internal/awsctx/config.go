package awsctx

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	profileHeaderRe = regexp.MustCompile(`^\[profile\s+(.+)\]$`)
	defaultHeaderRe = regexp.MustCompile(`^\[default\]$`)
)

func awsConfigPath() string {
	if p := os.Getenv("AWS_CONFIG_FILE"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aws", "config")
}

func awsCredentialsPath() string {
	if p := os.Getenv("AWS_SHARED_CREDENTIALS_FILE"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aws", "credentials")
}

// getProfiles parses ~/.aws/config and returns profile names.
func getProfiles() ([]string, error) {
	f, err := os.Open(awsConfigPath())
	if err != nil {
		return nil, fmt.Errorf("cannot read AWS config: %w", err)
	}
	defer f.Close()

	var profiles []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if defaultHeaderRe.MatchString(line) {
			profiles = append(profiles, "default")
		} else if m := profileHeaderRe.FindStringSubmatch(line); m != nil {
			profiles = append(profiles, m[1])
		}
	}
	return profiles, scanner.Err()
}

// profileExists checks whether a profile is defined in AWS config.
func profileExists(name string) bool {
	profiles, err := getProfiles()
	if err != nil {
		return false
	}
	for _, p := range profiles {
		if p == name {
			return true
		}
	}
	return false
}

// currentProfile returns the currently active AWS profile.
// Checks: env var > state file > "default".
func currentProfile() string {
	if p := os.Getenv("AWS_PROFILE"); p != "" {
		return p
	}
	if p := readState("profile"); p != "" {
		return p
	}
	return "default"
}

// currentRegion returns the currently active AWS region.
// Checks: env var > state file > config file > "(none)".
func currentRegion() string {
	if r := os.Getenv("AWS_REGION"); r != "" {
		return r
	}
	if r := os.Getenv("AWS_DEFAULT_REGION"); r != "" {
		return r
	}
	if r := readState("region"); r != "" {
		return r
	}
	// Fall back to config file region for current profile
	if r := getProfileRegion(currentProfile()); r != "" {
		return r
	}
	return "(none)"
}

// getProfileRegion returns the region configured for a specific profile in ~/.aws/config.
func getProfileRegion(name string) string {
	f, err := os.Open(awsConfigPath())
	if err != nil {
		return ""
	}
	defer f.Close()

	inTarget := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if name == "default" && defaultHeaderRe.MatchString(line) {
			inTarget = true
			continue
		}
		if m := profileHeaderRe.FindStringSubmatch(line); m != nil {
			inTarget = (m[1] == name)
			continue
		}
		if defaultHeaderRe.MatchString(line) && name != "default" {
			inTarget = false
			continue
		}

		if inTarget && strings.HasPrefix(line, "region") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}
func switchProfileInConfig(name string) error {
	ini, err := loadINI(awsConfigPath())
	if err != nil {
		return err
	}

	// One-time backup of original [default]
	if !ini.hasSection("_awsctx_original_default") {
		if ini.hasSection("default") {
			ini.copySection("default", "_awsctx_original_default")
		}
	}

	if name == "default" {
		// Restore original
		if ini.hasSection("_awsctx_original_default") {
			ini.copySection("_awsctx_original_default", "default")
			ini.deleteSection("_awsctx_original_default")
		}
	} else {
		// Copy [profile <name>] → [default]
		srcSection := "profile " + name
		if !ini.hasSection(srcSection) {
			return fmt.Errorf("profile %q not found in %s", name, awsConfigPath())
		}
		ini.copySection(srcSection, "default")
	}

	return ini.save()
}

func switchProfileInCredentials(name string) error {
	ini, err := loadINI(awsCredentialsPath())
	if err != nil {
		return err
	}

	// If credentials file is empty/missing, skip silently
	if len(ini.lines) == 0 {
		return nil
	}

	// One-time backup of original [default]
	if ini.hasSection("default") && !ini.hasSection("_awsctx_original_default") {
		ini.copySection("default", "_awsctx_original_default")
	}

	if name == "default" {
		if ini.hasSection("_awsctx_original_default") {
			ini.copySection("_awsctx_original_default", "default")
			ini.deleteSection("_awsctx_original_default")
		}
	} else {
		// In credentials file, profiles are [name] not [profile name]
		if !ini.hasSection(name) {
			// Some profiles (SSO, role-based) have no credentials entry — skip silently
			return nil
		}
		ini.copySection(name, "default")
	}

	return ini.save()
}

func switchRegionInConfig(region string) error {
	ini, err := loadINI(awsConfigPath())
	if err != nil {
		return err
	}
	ini.setKey("default", "region", region)
	return ini.save()
}
