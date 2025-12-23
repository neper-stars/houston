package parser_test

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

type TerraformableFoundExpected struct {
	Scenario    string `json:"scenario"`
	Description string `json:"description"`
	Year        int    `json:"year"`

	Player struct {
		RaceMaxGrowthRate int `json:"raceMaxGrowthRate"`
	} `json:"player"`

	Events []struct {
		Type              string  `json:"type"`
		PlanetName        string  `json:"planetName,omitempty"`
		PlanetId          int     `json:"planetId,omitempty"`
		GrowthRatePercent float64 `json:"growthRatePercent,omitempty"`
		Level             int     `json:"level,omitempty"`
		CompletedField    string  `json:"completedField,omitempty"`
		NextField         string  `json:"nextField,omitempty"`
	} `json:"events"`

	PartialPlanets []struct {
		PlanetId         int    `json:"planetId"`
		Name             string `json:"name"`
		HasTerraformData bool   `json:"hasTerraformData"`
	} `json:"partialPlanets"`
}

func TestScenarioNewTerraformableFound(t *testing.T) {
	// Load expected values
	expectedData, err := os.ReadFile("../testdata/scenario-new-terraformable-found/expected.json")
	if err != nil {
		t.Fatalf("Failed to read expected.json: %v", err)
	}
	var expected TerraformableFoundExpected
	if err := json.Unmarshal(expectedData, &expected); err != nil {
		t.Fatalf("Failed to parse expected.json: %v", err)
	}

	// Load M file
	data, err := os.ReadFile("../testdata/scenario-new-terraformable-found/game.m2")
	if err != nil {
		t.Fatalf("Failed to read M file: %v", err)
	}

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse blocks: %v", err)
	}

	// Find PlayerBlock, EventsBlock
	var playerBlock *blocks.PlayerBlock
	var eventsBlock *blocks.EventsBlock
	for _, block := range blockList {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			playerBlock = &pb
		}
		if eb, ok := block.(blocks.EventsBlock); ok {
			eventsBlock = &eb
		}
	}

	t.Run("PlayerBlock", func(t *testing.T) {
		if playerBlock == nil {
			t.Fatal("PlayerBlock not found")
		}

		t.Run("GrowthRate", func(t *testing.T) {
			if playerBlock.GrowthRate != expected.Player.RaceMaxGrowthRate {
				t.Errorf("Expected growth rate %d%%, got %d%%",
					expected.Player.RaceMaxGrowthRate, playerBlock.GrowthRate)
			}
		})
	})

	t.Run("EventsBlock", func(t *testing.T) {
		if eventsBlock == nil {
			t.Fatal("EventsBlock not found")
		}

		// Find expected events by type
		for _, expectedEvent := range expected.Events {
			switch expectedEvent.Type {
			case "TerraformablePlanetFound":
				t.Run("TerraformablePlanetFound", func(t *testing.T) {
					if len(eventsBlock.TerraformablePlanets) == 0 {
						t.Fatal("No TerraformablePlanetFound events found")
					}
					event := eventsBlock.TerraformablePlanets[0]
					// Compare with tolerance for floating point
					if math.Abs(event.GrowthRatePercent-expectedEvent.GrowthRatePercent) > 0.01 {
						t.Errorf("Expected growth rate %.2f%%, got %.2f%%",
							expectedEvent.GrowthRatePercent, event.GrowthRatePercent)
					}
				})

			case "ResearchComplete":
				t.Run("ResearchComplete", func(t *testing.T) {
					if len(eventsBlock.ResearchEvents) == 0 {
						t.Fatal("No ResearchComplete events found")
					}
					event := eventsBlock.ResearchEvents[0]
					if event.Level != expectedEvent.Level {
						t.Errorf("Expected level %d, got %d",
							expectedEvent.Level, event.Level)
					}
					completedFieldName := blocks.ResearchFieldName(event.Field)
					if completedFieldName != expectedEvent.CompletedField {
						t.Errorf("Expected completed field %s, got %s",
							expectedEvent.CompletedField, completedFieldName)
					}
					nextFieldName := blocks.ResearchFieldName(event.NextField)
					if nextFieldName != expectedEvent.NextField {
						t.Errorf("Expected next field %s, got %s",
							expectedEvent.NextField, nextFieldName)
					}
				})
			}
		}
	})
}
