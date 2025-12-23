package parser_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

type ProductionQueueChangeExpected struct {
	Scenario    string `json:"scenario"`
	Description string `json:"description"`
	Year        int    `json:"year"`
	PlanetId    int    `json:"planetId"`
	PlanetName  string `json:"planetName"`

	ProductionQueueChange struct {
		ItemCount  int `json:"itemCount"`
		QueueItems []struct {
			Index  int    `json:"index"`
			ItemId int    `json:"itemId"`
			Name   string `json:"name"`
			Count  int    `json:"count"`
		} `json:"queueItems"`
	} `json:"productionQueueChange"`

	PlanetChanges []struct {
		Setting string `json:"setting"`
		Value   bool   `json:"value"`
	} `json:"planetChanges"`
}

func TestScenarioProductionQueueChange(t *testing.T) {
	// Load expected values
	expectedData, err := os.ReadFile("../testdata/scenario-production-queue-change/expected.json")
	if err != nil {
		t.Fatalf("Failed to read expected.json: %v", err)
	}
	var expected ProductionQueueChangeExpected
	if err := json.Unmarshal(expectedData, &expected); err != nil {
		t.Fatalf("Failed to parse expected.json: %v", err)
	}

	// Load X file
	data, err := os.ReadFile("../testdata/scenario-production-queue-change/game.x1")
	if err != nil {
		t.Fatalf("Failed to read X file: %v", err)
	}

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse blocks: %v", err)
	}

	// Find ProductionQueueChangeBlock and PlanetChangeBlock
	var queueChange *blocks.ProductionQueueChangeBlock
	var planetChange *blocks.PlanetChangeBlock
	for _, block := range blockList {
		if pqcb, ok := block.(blocks.ProductionQueueChangeBlock); ok {
			queueChange = &pqcb
		}
		if pcb, ok := block.(blocks.PlanetChangeBlock); ok {
			planetChange = &pcb
		}
	}

	t.Run("ProductionQueueChange", func(t *testing.T) {
		if queueChange == nil {
			t.Fatal("ProductionQueueChangeBlock not found")
		}

		t.Run("PlanetId", func(t *testing.T) {
			if queueChange.PlanetId != expected.PlanetId {
				t.Errorf("Expected planet ID %d (%s), got %d",
					expected.PlanetId, expected.PlanetName, queueChange.PlanetId)
			}
		})

		t.Run("QueueLength", func(t *testing.T) {
			if queueChange.QueueLength() != expected.ProductionQueueChange.ItemCount {
				t.Errorf("Expected %d queue items, got %d",
					expected.ProductionQueueChange.ItemCount, queueChange.QueueLength())
			}
		})

		t.Run("QueueItems", func(t *testing.T) {
			for _, expectedItem := range expected.ProductionQueueChange.QueueItems {
				idx := expectedItem.Index - 1 // expected.json uses 1-based index
				item := queueChange.GetItem(idx)
				if item == nil {
					t.Errorf("Item %d not found", expectedItem.Index)
					continue
				}
				if item.ItemId != expectedItem.ItemId {
					t.Errorf("Item %d: expected %s (%d), got %d",
						expectedItem.Index, expectedItem.Name, expectedItem.ItemId, item.ItemId)
				}
				if item.Count != expectedItem.Count {
					t.Errorf("Item %d (%s): expected count %d, got %d",
						expectedItem.Index, expectedItem.Name, expectedItem.Count, item.Count)
				}
			}
		})
	})

	t.Run("PlanetChange", func(t *testing.T) {
		if planetChange == nil {
			t.Fatal("PlanetChangeBlock not found")
		}

		t.Run("PlanetId", func(t *testing.T) {
			if planetChange.PlanetId != expected.PlanetId {
				t.Errorf("Expected planet ID %d (%s), got %d",
					expected.PlanetId, expected.PlanetName, planetChange.PlanetId)
			}
		})

		t.Run("Settings", func(t *testing.T) {
			for _, change := range expected.PlanetChanges {
				switch change.Setting {
				case "ContributeOnlyLeftover":
					if planetChange.ContributeOnlyLeftover != change.Value {
						t.Errorf("ContributeOnlyLeftover: expected %v, got %v",
							change.Value, planetChange.ContributeOnlyLeftover)
					}
				default:
					t.Logf("Unknown setting: %s", change.Setting)
				}
			}
		})
	})
}
