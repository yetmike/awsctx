package awsctx

import "testing"

func TestIsValidRegion(t *testing.T) {
	valid := []string{"us-east-1", "eu-west-1", "ap-southeast-1", "af-south-1"}
	for _, r := range valid {
		if !isValidRegion(r) {
			t.Errorf("%s should be valid", r)
		}
	}

	invalid := []string{"us-east-99", "fake-region", "", "US-EAST-1"}
	for _, r := range invalid {
		if isValidRegion(r) {
			t.Errorf("%s should be invalid", r)
		}
	}
}

func TestRegionListNotEmpty(t *testing.T) {
	if len(awsRegions) == 0 {
		t.Error("region list should not be empty")
	}
}

func TestRegionListNoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	for _, r := range awsRegions {
		if seen[r] {
			t.Errorf("duplicate region: %s", r)
		}
		seen[r] = true
	}
}
