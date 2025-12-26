package racefixer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemovePasswordBytes(t *testing.T) {
	// Test with a password-protected race file
	passFile := "../../../testdata/scenario-racefiles/race1-password.r2"
	passData, err := os.ReadFile(passFile)
	if err != nil {
		t.Fatalf("Failed to read password-protected file: %v", err)
	}

	// Analyze the file first
	info, err := AnalyzeBytes(passFile, passData)
	if err != nil {
		t.Fatalf("Failed to analyze: %v", err)
	}

	if !info.HasPassword {
		t.Fatal("Expected file to have password")
	}
	if info.SingularName != "race1" {
		t.Errorf("Expected singular name 'race1', got %q", info.SingularName)
	}
	if info.PluralName != "race1s" {
		t.Errorf("Expected plural name 'race1s', got %q", info.PluralName)
	}

	// Remove the password
	repaired, result, err := RemovePasswordBytes(passData)
	if err != nil {
		t.Fatalf("RemovePasswordBytes failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got: %s", result.Message)
	}
	if !result.PasswordRemoved {
		t.Error("Expected PasswordRemoved to be true")
	}

	// Verify the repaired file has no password
	repairedInfo, err := AnalyzeBytes("repaired", repaired)
	if err != nil {
		t.Fatalf("Failed to analyze repaired file: %v", err)
	}

	if repairedInfo.HasPassword {
		t.Error("Expected repaired file to have no password")
	}
	if repairedInfo.NeedsRepair {
		t.Error("Expected repaired file to have correct checksum")
	}
}

func TestRemovePasswordBytes_NoPassword(t *testing.T) {
	// Test with a file that has no password
	noPassFile := "../../../testdata/scenario-racefiles/race1-nopassword.r2"
	noPassData, err := os.ReadFile(noPassFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Remove password (should be a no-op)
	_, result, err := RemovePasswordBytes(noPassData)
	if err != nil {
		t.Fatalf("RemovePasswordBytes failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got: %s", result.Message)
	}
	if result.PasswordRemoved {
		t.Error("Expected PasswordRemoved to be false for file without password")
	}
}

func TestAnalyzeBytes_ScenarioRacefixer(t *testing.T) {
	// Test game.r1 which needs repair - verify we can detect and fix it
	t.Run("game.r1_needs_repair", func(t *testing.T) {
		data, err := os.ReadFile("../../../testdata/scenario-racefixer/game.r1")
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		// Analyze - should need repair
		info, err := AnalyzeBytes("game.r1", data)
		if err != nil {
			t.Fatalf("Failed to analyze: %v", err)
		}

		if info.SingularName != "Fool" {
			t.Errorf("Expected singular name 'Fool', got %q", info.SingularName)
		}
		if info.PluralName != "Fools" {
			t.Errorf("Expected plural name 'Fools', got %q", info.PluralName)
		}

		// Assert it needs repair
		if !info.NeedsRepair {
			t.Fatal("Expected game.r1 to need repair")
		}
		t.Logf("game.r1 needs repair: current=0x%04X, expected=0x%04X",
			info.CurrentFooter, info.ExpectedFooter)

		// Fix it in memory
		repaired, result, err := RepairBytes(data)
		if err != nil {
			t.Fatalf("RepairBytes failed: %v", err)
		}
		if !result.Success {
			t.Fatalf("Expected repair success, got: %s", result.Message)
		}
		if !result.FooterChanged {
			t.Error("Expected footer to be changed")
		}

		// Re-analyze the repaired data - should NOT need repair anymore
		repairedInfo, err := AnalyzeBytes("game.r1.repaired", repaired)
		if err != nil {
			t.Fatalf("Failed to analyze repaired file: %v", err)
		}

		if repairedInfo.NeedsRepair {
			t.Errorf("Repaired file still needs repair: current=0x%04X, expected=0x%04X",
				repairedInfo.CurrentFooter, repairedInfo.ExpectedFooter)
		}

		// Names should be preserved
		if repairedInfo.SingularName != "Fool" {
			t.Errorf("Expected singular name 'Fool' after repair, got %q", repairedInfo.SingularName)
		}
	})

	// Test game.r2 which should be valid
	t.Run("game.r2_valid", func(t *testing.T) {
		data, err := os.ReadFile("../../../testdata/scenario-racefixer/game.r2")
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		info, err := AnalyzeBytes("game.r2", data)
		if err != nil {
			t.Fatalf("Failed to analyze: %v", err)
		}

		if info.SingularName != "Halfling" {
			t.Errorf("Expected singular name 'Halfling', got %q", info.SingularName)
		}
		if info.PluralName != "Halflings" {
			t.Errorf("Expected plural name 'Halflings', got %q", info.PluralName)
		}

		// Should not need repair
		if info.NeedsRepair {
			t.Errorf("Expected game.r2 to be valid, but needs repair: current=0x%04X, expected=0x%04X",
				info.CurrentFooter, info.ExpectedFooter)
		}

		// RemovePasswordBytes should succeed (no password to remove)
		_, result, err := RemovePasswordBytes(data)
		if err != nil {
			t.Fatalf("RemovePasswordBytes failed: %v", err)
		}
		if !result.Success {
			t.Errorf("Expected success, got: %s", result.Message)
		}
	})
}

func TestAnalyzeBytes_AllRaceFiles(t *testing.T) {
	// Test that we can analyze all race files in testdata
	patterns := []string{
		"../../../testdata/scenario-racefiles/*.r*",
		"../../../testdata/scenario-racefiles/**/*.r*",
	}

	var files []string
	for _, p := range patterns {
		matches, _ := filepath.Glob(p)
		files = append(files, matches...)
	}

	if len(files) == 0 {
		t.Fatal("No race files found")
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			info, err := AnalyzeBytes(file, data)
			if err != nil {
				t.Fatalf("Failed to analyze: %v", err)
			}

			// Verify basic properties
			if info.Size != len(data) {
				t.Errorf("Size mismatch: expected %d, got %d", len(data), info.Size)
			}
			if info.SingularName == "" {
				t.Error("Expected non-empty singular name")
			}

			// Valid race files should not need repair
			if info.NeedsRepair {
				t.Errorf("File needs repair: current=0x%04X, expected=0x%04X",
					info.CurrentFooter, info.ExpectedFooter)
			}
		})
	}
}
