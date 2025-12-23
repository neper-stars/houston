package parser_test

import (
	"os"
	"testing"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

// Research event parsing tests.
//
// Research Complete Event Format (7 bytes):
//   50 00 FE FF LL CF NF
//   - 0x50 = EventTypeResearchComplete
//   - 0x00 = flags
//   - 0xFFFE = "no planet" marker (research is global, not planet-specific)
//   - LL = level achieved (1-26)
//   - CF = completed field (0-5)
//   - NF = next research field (0-5)
//
// Format confirmed by cross-referencing:
//   - Player 1 (game.m1): Has population events before research
//   - Player 2 (other-player/game.m2): NO population events before research
// Both use identical structure, proving 0x50 is the research event type.

func TestScenarioMessageResearchDone(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-message-research-done/game.m1")
	if err != nil {
		t.Fatalf("Failed to read M file: %v", err)
	}

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse blocks: %v", err)
	}

	// Find EventsBlock
	var eventsBlock *blocks.EventsBlock
	for _, block := range blockList {
		if eb, ok := block.(blocks.EventsBlock); ok {
			eventsBlock = &eb
			break
		}
	}

	if eventsBlock == nil {
		t.Fatal("EventsBlock not found")
	}

	t.Run("ResearchEventsCount", func(t *testing.T) {
		if len(eventsBlock.ResearchEvents) != 3 {
			t.Errorf("Expected 3 research events, got %d", len(eventsBlock.ResearchEvents))
		}
	})

	t.Run("BiotechLevel4", func(t *testing.T) {
		if len(eventsBlock.ResearchEvents) < 1 {
			t.Fatal("No research events")
		}
		event := eventsBlock.ResearchEvents[0]
		if event.Level != 4 {
			t.Errorf("Expected level 4, got %d", event.Level)
		}
		if event.Field != blocks.ResearchFieldBiotechnology {
			t.Errorf("Expected field Biotechnology (5), got %d", event.Field)
		}
		// NextField should also be Biotechnology (player continued in same field)
		if event.NextField != blocks.ResearchFieldBiotechnology {
			t.Errorf("Expected next field Biotechnology (5), got %d", event.NextField)
		}
	})

	t.Run("BiotechLevel5", func(t *testing.T) {
		if len(eventsBlock.ResearchEvents) < 2 {
			t.Fatal("Not enough research events")
		}
		event := eventsBlock.ResearchEvents[1]
		if event.Level != 5 {
			t.Errorf("Expected level 5, got %d", event.Level)
		}
		if event.Field != blocks.ResearchFieldBiotechnology {
			t.Errorf("Expected field Biotechnology (5), got %d", event.Field)
		}
	})

	t.Run("BiotechLevel6", func(t *testing.T) {
		if len(eventsBlock.ResearchEvents) < 3 {
			t.Fatal("Not enough research events")
		}
		event := eventsBlock.ResearchEvents[2]
		if event.Level != 6 {
			t.Errorf("Expected level 6, got %d", event.Level)
		}
		if event.Field != blocks.ResearchFieldBiotechnology {
			t.Errorf("Expected field Biotechnology (5), got %d", event.Field)
		}
	})

	t.Run("TechBenefitsCount", func(t *testing.T) {
		if len(eventsBlock.TechBenefits) != 4 {
			t.Errorf("Expected 4 tech benefits, got %d", len(eventsBlock.TechBenefits))
		}
	})

	t.Run("TechBenefitItemIDs", func(t *testing.T) {
		// Expected item IDs from screenshots:
		// - 1473 (Dolphin Scanner, category 8)
		// - 1475 (Carbonic Armor, category 2)
		// - 1480 (Mine Dispenser 50, category 1)
		// - 1473 (DNA Scanner, category 3)
		expectedItemIDs := []int{1473, 1475, 1480, 1473}
		expectedCategories := []int{8, 2, 1, 3}

		for i, expected := range expectedItemIDs {
			if i >= len(eventsBlock.TechBenefits) {
				t.Errorf("Missing tech benefit %d", i)
				continue
			}
			benefit := eventsBlock.TechBenefits[i]
			if benefit.ItemID != expected {
				t.Errorf("Benefit %d: expected itemID %d, got %d", i, expected, benefit.ItemID)
			}
			if benefit.Category != expectedCategories[i] {
				t.Errorf("Benefit %d: expected category %d, got %d", i, expectedCategories[i], benefit.Category)
			}
		}
	})

	t.Run("ResearchFieldName", func(t *testing.T) {
		if blocks.ResearchFieldName(blocks.ResearchFieldBiotechnology) != "Biotechnology" {
			t.Error("ResearchFieldName for Biotechnology incorrect")
		}
		if blocks.ResearchFieldName(blocks.ResearchFieldWeapons) != "Weapons" {
			t.Error("ResearchFieldName for Weapons incorrect")
		}
	})
}

func TestScenarioMessageResearchDonePlayer2(t *testing.T) {
	// Player 2 (Halflings) - critical test case: NO population events before research
	// This confirms that 0x50 is the research event type, not population data
	data, err := os.ReadFile("../testdata/scenario-message-research-done/other-player/game.m2")
	if err != nil {
		t.Fatalf("Failed to read M file: %v", err)
	}

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse blocks: %v", err)
	}

	var eventsBlock *blocks.EventsBlock
	for _, block := range blockList {
		if eb, ok := block.(blocks.EventsBlock); ok {
			eventsBlock = &eb
			break
		}
	}

	if eventsBlock == nil {
		t.Fatal("EventsBlock not found")
	}

	t.Run("ElectronicsResearchCount", func(t *testing.T) {
		// Player 2 has 2 research events: Level 4 and Level 5 Electronics
		if len(eventsBlock.ResearchEvents) != 2 {
			t.Errorf("Expected 2 research events, got %d", len(eventsBlock.ResearchEvents))
		}
	})

	t.Run("ElectronicsLevel4", func(t *testing.T) {
		if len(eventsBlock.ResearchEvents) < 1 {
			t.Fatal("No research events")
		}
		event := eventsBlock.ResearchEvents[0]
		if event.Level != 4 {
			t.Errorf("Expected level 4, got %d", event.Level)
		}
		if event.Field != blocks.ResearchFieldElectronics {
			t.Errorf("Expected field Electronics (4), got %d", event.Field)
		}
		// NextField should also be Electronics (player continued in same field)
		if event.NextField != blocks.ResearchFieldElectronics {
			t.Errorf("Expected next field Electronics (4), got %d", event.NextField)
		}
	})

	t.Run("ElectronicsLevel5", func(t *testing.T) {
		if len(eventsBlock.ResearchEvents) < 2 {
			t.Fatal("Not enough research events")
		}
		event := eventsBlock.ResearchEvents[1]
		if event.Level != 5 {
			t.Errorf("Expected level 5, got %d", event.Level)
		}
		if event.Field != blocks.ResearchFieldElectronics {
			t.Errorf("Expected field Electronics (4), got %d", event.Field)
		}
	})

	t.Run("TechBenefitsCount", func(t *testing.T) {
		// Player 2 has 5 tech benefits from Electronics
		// msg-18: Mole Scanner, msg-19: Robo-Maxi-Miner, msg-20: Energy Capacitor
		// msg-22: Possum Scanner, msg-23: Stealth Cloak
		if len(eventsBlock.TechBenefits) != 5 {
			t.Errorf("Expected 5 tech benefits, got %d", len(eventsBlock.TechBenefits))
		}
	})
}

func TestScenarioMessageComplexResearch(t *testing.T) {
	// Cross-reference with message-complex scenario for Weapons research
	data, err := os.ReadFile("../testdata/scenario-message-complex/game.m1")
	if err != nil {
		t.Fatalf("Failed to read M file: %v", err)
	}

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse blocks: %v", err)
	}

	var eventsBlock *blocks.EventsBlock
	for _, block := range blockList {
		if eb, ok := block.(blocks.EventsBlock); ok {
			eventsBlock = &eb
			break
		}
	}

	if eventsBlock == nil {
		t.Fatal("EventsBlock not found")
	}

	t.Run("WeaponsLevel12", func(t *testing.T) {
		if len(eventsBlock.ResearchEvents) != 1 {
			t.Errorf("Expected 1 research event, got %d", len(eventsBlock.ResearchEvents))
			return
		}
		event := eventsBlock.ResearchEvents[0]
		if event.Level != 12 {
			t.Errorf("Expected level 12, got %d", event.Level)
		}
		if event.Field != blocks.ResearchFieldWeapons {
			t.Errorf("Expected field Weapons (1), got %d", event.Field)
		}
	})

	t.Run("WeaponsBenefits", func(t *testing.T) {
		// Message 45: Mini Blaster (itemID 452)
		// Message 46: Jihad Missile (itemID 453)
		if len(eventsBlock.TechBenefits) != 2 {
			t.Errorf("Expected 2 tech benefits, got %d", len(eventsBlock.TechBenefits))
			return
		}
		if eventsBlock.TechBenefits[0].ItemID != 452 {
			t.Errorf("Expected Mini Blaster itemID 452, got %d", eventsBlock.TechBenefits[0].ItemID)
		}
		if eventsBlock.TechBenefits[1].ItemID != 453 {
			t.Errorf("Expected Jihad Missile itemID 453, got %d", eventsBlock.TechBenefits[1].ItemID)
		}
	})
}
