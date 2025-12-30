package store

import (
	"os"
	"testing"

	"github.com/neper-stars/houston/race"
)

func TestParseRaceData(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-racebuilder/predefined-races/humanoids/race.r1")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	r, err := ParseRaceData(data)
	if err != nil {
		t.Fatalf("ParseRaceData failed: %v", err)
	}

	if r.SingularName != "Humanoid" {
		t.Errorf("SingularName: got %q, want %q", r.SingularName, "Humanoid")
	}
	if r.PRT != race.PRTJackOfAllTrades {
		t.Errorf("PRT: got %d, want %d (JOAT)", r.PRT, race.PRTJackOfAllTrades)
	}
}

func TestParseRaceDataInvalid(t *testing.T) {
	// Test with invalid data
	_, err := ParseRaceData([]byte{0x00, 0x01, 0x02})
	if err == nil {
		t.Error("Expected error for invalid data")
	}
}

func TestValidateRaceData(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-racebuilder/predefined-races/humanoids/race.r1")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	r, errs := ValidateRaceData(data)
	if r == nil {
		t.Fatal("Expected race to be returned")
	}
	if len(errs) > 0 {
		t.Errorf("Expected no validation errors, got: %v", errs)
	}
}

func TestValidateRaceDataInvalid(t *testing.T) {
	// Test with invalid data
	r, errs := ValidateRaceData([]byte{0x00, 0x01, 0x02})
	if r != nil {
		t.Error("Expected nil race for invalid data")
	}
	if len(errs) == 0 {
		t.Error("Expected validation errors for invalid data")
	}
}

func TestParseRaceFile(t *testing.T) {
	r, err := ParseRaceFile("../testdata/scenario-racebuilder/predefined-races/humanoids/race.r1")
	if err != nil {
		t.Fatalf("ParseRaceFile failed: %v", err)
	}

	if r.SingularName != "Humanoid" {
		t.Errorf("SingularName: got %q, want %q", r.SingularName, "Humanoid")
	}
}

func TestParseRaceFileNotRaceFile(t *testing.T) {
	// Test with non-race file extension
	_, err := ParseRaceFile("game.m1")
	if err != ErrNotRaceFile {
		t.Errorf("Expected ErrNotRaceFile, got: %v", err)
	}
}

func TestValidateRaceFile(t *testing.T) {
	r, errs := ValidateRaceFile("../testdata/scenario-racebuilder/predefined-races/humanoids/race.r1")
	if r == nil {
		t.Fatal("Expected race to be returned")
	}
	if len(errs) > 0 {
		t.Errorf("Expected no validation errors, got: %v", errs)
	}
}

func TestValidateRaceFileNotRaceFile(t *testing.T) {
	r, errs := ValidateRaceFile("game.m1")
	if r != nil {
		t.Error("Expected nil race for non-race file")
	}
	if len(errs) == 0 {
		t.Error("Expected validation errors for non-race file")
	}
	if errs[0].Message != ErrNotRaceFile.Error() {
		t.Errorf("Expected ErrNotRaceFile message, got: %s", errs[0].Message)
	}
}

func TestValidateAllPredefinedRaceFiles(t *testing.T) {
	testFiles := []struct {
		path string
		name string
	}{
		{"../testdata/scenario-racebuilder/predefined-races/humanoids/race.r1", "Humanoid"},
		{"../testdata/scenario-racebuilder/predefined-races/rabbitoids/race.r1", "Rabbitoid"},
		{"../testdata/scenario-racebuilder/predefined-races/insectoids/race.r1", "Insectoid"},
		{"../testdata/scenario-racebuilder/predefined-races/nucleotids/race.r1", "Nucleotid"},
		{"../testdata/scenario-racebuilder/predefined-races/silicanoids/race.r1", "Silicanoid"},
		{"../testdata/scenario-racebuilder/predefined-races/antetherals/race.r1", "Antetheral"},
	}

	for _, tf := range testFiles {
		t.Run(tf.name, func(t *testing.T) {
			r, errs := ValidateRaceFile(tf.path)
			if r == nil {
				t.Fatalf("Expected race to be returned for %s", tf.path)
			}
			if len(errs) > 0 {
				t.Errorf("Validation errors for %s: %v", tf.path, errs)
			}
			if r.SingularName != tf.name {
				t.Errorf("SingularName: got %q, want %q", r.SingularName, tf.name)
			}
		})
	}
}
