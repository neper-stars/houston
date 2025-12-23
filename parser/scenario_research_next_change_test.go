package parser_test

import (
	"os"
	"testing"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

// ResearchChangeBlock format (2 bytes):
//   Byte 0: Research budget percentage (0-100)
//   Byte 1: (next_field << 4) | current_field

func TestScenarioResearchNextChange(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-research-next-change/game.x1")
	if err != nil {
		t.Fatalf("Failed to read X file: %v", err)
	}

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse blocks: %v", err)
	}

	// Find ResearchChangeBlock
	var researchChange *blocks.ResearchChangeBlock
	for _, block := range blockList {
		if rcb, ok := block.(blocks.ResearchChangeBlock); ok {
			researchChange = &rcb
			break
		}
	}

	if researchChange == nil {
		t.Fatal("ResearchChangeBlock not found")
	}

	t.Run("BudgetPercent", func(t *testing.T) {
		// Screenshot shows: "Resources budgeted for research: 15%"
		if researchChange.BudgetPercent != 15 {
			t.Errorf("Expected budget 15%%, got %d%%", researchChange.BudgetPercent)
		}
	})

	t.Run("CurrentField", func(t *testing.T) {
		// Currently researching Biotechnology (5)
		if researchChange.CurrentField != blocks.ResearchFieldBiotechnology {
			t.Errorf("Expected current field Biotechnology (5), got %d", researchChange.CurrentField)
		}
	})

	t.Run("NextField", func(t *testing.T) {
		// Changed next field to Propulsion (2)
		if researchChange.NextField != blocks.ResearchFieldPropulsion {
			t.Errorf("Expected next field Propulsion (2), got %d", researchChange.NextField)
		}
	})
}
