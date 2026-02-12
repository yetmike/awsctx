package awsctx

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestAWS creates a temp AWS config file and isolated cache dir.
// Returns a cleanup function that restores original env vars.
func setupTestAWS(t *testing.T, config, credentials string) func() {
	t.Helper()
	dir := t.TempDir()

	configPath := filepath.Join(dir, "config")
	credentialsPath := filepath.Join(dir, "credentials")
	cacheDir := filepath.Join(dir, "cache")

	os.WriteFile(configPath, []byte(config), 0o644)
	if credentials != "" {
		os.WriteFile(credentialsPath, []byte(credentials), 0o644)
	}
	os.MkdirAll(cacheDir, 0o755)

	origConfig := os.Getenv("AWS_CONFIG_FILE")
	origCredentials := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	origCache := os.Getenv("XDG_CACHE_HOME")
	origProfile := os.Getenv("AWS_PROFILE")
	origRegion := os.Getenv("AWS_REGION")
	origDefaultRegion := os.Getenv("AWS_DEFAULT_REGION")

	os.Setenv("AWS_CONFIG_FILE", configPath)
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", credentialsPath)
	os.Setenv("XDG_CACHE_HOME", cacheDir)
	os.Unsetenv("AWS_PROFILE")
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_DEFAULT_REGION")

	return func() {
		os.Setenv("AWS_CONFIG_FILE", origConfig)
		os.Setenv("AWS_SHARED_CREDENTIALS_FILE", origCredentials)
		os.Setenv("XDG_CACHE_HOME", origCache)
		if origProfile != "" {
			os.Setenv("AWS_PROFILE", origProfile)
		} else {
			os.Unsetenv("AWS_PROFILE")
		}
		if origRegion != "" {
			os.Setenv("AWS_REGION", origRegion)
		} else {
			os.Unsetenv("AWS_REGION")
		}
		if origDefaultRegion != "" {
			os.Setenv("AWS_DEFAULT_REGION", origDefaultRegion)
		} else {
			os.Unsetenv("AWS_DEFAULT_REGION")
		}
	}
}

// Tests use isolated temp files and never touch real ~/.aws/.
const testConfig = `[default]
region = eu-west-1
output = json

[profile dev]
region = us-west-2
output = yaml

[profile staging]
region = eu-west-1
output = json
`

const testCredentials = `[default]
aws_access_key_id = AKIADEFAULT
aws_secret_access_key = default-secret

[dev]
aws_access_key_id = AKIADEV
aws_secret_access_key = dev-secret

[staging]
aws_access_key_id = AKIASTAGING
aws_secret_access_key = staging-secret
`

func TestGetProfiles(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	profiles, err := getProfiles()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"default", "dev", "staging"}
	if len(profiles) != len(expected) {
		t.Fatalf("expected %d profiles, got %d: %v", len(expected), len(profiles), profiles)
	}
}

func TestProfileExists(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	if !profileExists("dev") {
		t.Error("dev should exist")
	}
}

func TestCurrentProfile_Default(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	if p := currentProfile(); p != "default" {
		t.Errorf("expected 'default', got %s", p)
	}
}

func TestCurrentRegion_FromConfig(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	if r := currentRegion(); r != "eu-west-1" {
		t.Errorf("expected eu-west-1, got %s", r)
	}
}

func TestSwitchProfileInConfig(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	err := switchProfileInConfig("dev")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ini, _ := loadINI(awsConfigPath())
	keys := ini.getKeys("default")
	if keys["region"] != "us-west-2" {
		t.Errorf("expected us-west-2, got %s", keys["region"])
	}

	if !ini.hasSection("_awsctx_original_default") {
		t.Error("backup section _awsctx_original_default not found")
	}
}

func TestSwitchProfileInConfig_RestoreDefault(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	switchProfileInConfig("dev")
	switchProfileInConfig("default")

	ini, _ := loadINI(awsConfigPath())
	keys := ini.getKeys("default")
	if keys["region"] != "eu-west-1" {
		t.Errorf("expected eu-west-1, got %s", keys["region"])
	}

	if ini.hasSection("_awsctx_original_default") {
		t.Error("backup section _awsctx_original_default should be removed")
	}
}

func TestSwitchProfileInCredentials(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, testCredentials)
	defer cleanup()

	err := switchProfileInCredentials("dev")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ini, _ := loadINI(awsCredentialsPath())
	keys := ini.getKeys("default")
	if keys["aws_access_key_id"] != "AKIADEV" {
		t.Errorf("expected AKIADEV, got %s", keys["aws_access_key_id"])
	}
}

func TestSwitchProfileInCredentials_MissingFile(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	err := switchProfileInCredentials("dev")
	if err != nil {
		t.Errorf("expected no error when credentials file missing, got %v", err)
	}
}

func TestSwitchRegionInConfig(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	err := switchRegionInConfig("us-east-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ini, _ := loadINI(awsConfigPath())
	keys := ini.getKeys("default")
	if keys["region"] != "us-east-1" {
		t.Errorf("expected us-east-1, got %s", keys["region"])
	}
}

func TestSwitchProfileInConfig_BackupOnlyOnce(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	switchProfileInConfig("dev")
	switchProfileInConfig("staging")

	ini, _ := loadINI(awsConfigPath())
	backupKeys := ini.getKeys("_awsctx_original_default")
	if backupKeys["region"] != "eu-west-1" {
		t.Errorf("backup should have original region eu-west-1, got %s", backupKeys["region"])
	}
}
