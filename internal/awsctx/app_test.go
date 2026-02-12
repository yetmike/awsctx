package awsctx

import (
	"os"
	"os/exec"
	"testing"
)

func TestRun_NoArgs(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	err := Run([]string{"awsctx"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestRun_Help(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	err := Run([]string{"awsctx", "--help"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestRun_Version(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	err := Run([]string{"awsctx", "--version"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	err := Run([]string{"awsctx", "garbage"})
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

func TestRun_ProfileCurrent(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	err := Run([]string{"awsctx", "profile", "-c"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Short form
	err = Run([]string{"awsctx", "p", "-c"})
	if err != nil {
		t.Errorf("expected no error for short form, got %v", err)
	}
}

func TestRun_RegionCurrent(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	err := Run([]string{"awsctx", "region", "-c"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	err = Run([]string{"awsctx", "r", "-c"})
	if err != nil {
		t.Errorf("expected no error for short form, got %v", err)
	}
}

func TestRun_NoAWSCLI(t *testing.T) {
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", origPath)

	if _, err := exec.LookPath("aws"); err != nil {
		err := Run([]string{"awsctx"})
		if err == nil {
			t.Error("expected error when aws CLI is missing")
		}
	}
}

func TestRun_ProfileSwitch(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, testCredentials)
	defer cleanup()

	err := Run([]string{"awsctx", "p", "dev"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// State file should reflect the switch
	if p := readState("profile"); p != "dev" {
		t.Errorf("state: expected dev, got %s", p)
	}

	// Config file [default] should have dev's region
	ini, _ := loadINI(awsConfigPath())
	keys := ini.getKeys("default")
	if keys["region"] != "us-west-2" {
		t.Errorf("config [default] region: expected us-west-2, got %s", keys["region"])
	}

	// Credentials file [default] should have dev's keys
	cini, _ := loadINI(awsCredentialsPath())
	ckeys := cini.getKeys("default")
	if ckeys["aws_access_key_id"] != "AKIADEV" {
		t.Errorf("credentials [default] key: expected AKIADEV, got %s", ckeys["aws_access_key_id"])
	}
}

func TestRun_RegionSwitch(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, testCredentials)
	defer cleanup()

	err := Run([]string{"awsctx", "r", "us-east-1"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Config file [default] should have new region
	ini, _ := loadINI(awsConfigPath())
	keys := ini.getKeys("default")
	if keys["region"] != "us-east-1" {
		t.Errorf("config [default] region: expected us-east-1, got %s", keys["region"])
	}
}

func TestRun_ProfileSwap(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	Run([]string{"awsctx", "p", "dev"})
	Run([]string{"awsctx", "p", "staging"})

	err := Run([]string{"awsctx", "p", "-"})
	if err != nil {
		t.Fatalf("swap error: %v", err)
	}
	if p := currentProfile(); p != "dev" {
		t.Errorf("expected dev after swap, got %s", p)
	}
}

func TestRun_RegionSwap(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	Run([]string{"awsctx", "r", "us-east-1"})
	Run([]string{"awsctx", "r", "eu-west-1"})

	err := Run([]string{"awsctx", "r", "-"})
	if err != nil {
		t.Fatalf("swap error: %v", err)
	}
	if r := readState("region"); r != "us-east-1" {
		t.Errorf("expected us-east-1 after swap, got %s", r)
	}
}

func TestRun_InvalidProfile(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	err := Run([]string{"awsctx", "p", "nonexistent"})
	if err == nil {
		t.Error("expected error for invalid profile")
	}
}

func TestRun_InvalidRegion(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	err := Run([]string{"awsctx", "r", "fake-region"})
	if err == nil {
		t.Error("expected error for invalid region")
	}
}

func TestRun_SwapNoHistory(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	err := Run([]string{"awsctx", "p", "-"})
	if err == nil {
		t.Error("expected error for swap with no history")
	}

	err = Run([]string{"awsctx", "r", "-"})
	if err == nil {
		t.Error("expected error for swap with no history")
	}
}

func TestRun_FzfList(t *testing.T) {
	cleanup := setupTestAWS(t, testConfig, "")
	defer cleanup()

	err := Run([]string{"awsctx", "--fzf-list", "profile"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	err = Run([]string{"awsctx", "--fzf-list", "region"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
