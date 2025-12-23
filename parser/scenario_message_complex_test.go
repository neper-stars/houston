package parser_test

import (
	"os"
	"testing"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

func TestScenarioMessageComplex(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-message-complex/game.m1")
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

	t.Run("ProductionEventsCount", func(t *testing.T) {
		if len(eventsBlock.ProductionEvents) != 32 {
			t.Errorf("Expected 32 production events, got %d", len(eventsBlock.ProductionEvents))
		}
	})

	t.Run("MungleFactoriesBuilt", func(t *testing.T) {
		// First event: "You have built 3 factories on Mungle" (planet 14)
		if len(eventsBlock.ProductionEvents) < 1 {
			t.Fatal("No production events")
		}
		event := eventsBlock.ProductionEvents[0]
		if event.EventType != blocks.EventTypeFactoriesBuilt {
			t.Errorf("Expected EventTypeFactoriesBuilt (0x36), got 0x%02X", event.EventType)
		}
		if event.PlanetID != 14 {
			t.Errorf("Expected planet 14 (Mungle), got %d", event.PlanetID)
		}
		if event.Count != 3 {
			t.Errorf("Expected count 3, got %d", event.Count)
		}
	})

	t.Run("MungleMinesBuilt", func(t *testing.T) {
		// Second event: "You have built 3 mines on Mungle" (planet 14)
		if len(eventsBlock.ProductionEvents) < 2 {
			t.Fatal("Not enough production events")
		}
		event := eventsBlock.ProductionEvents[1]
		if event.EventType != blocks.EventTypeMinesBuilt {
			t.Errorf("Expected EventTypeMinesBuilt (0x38), got 0x%02X", event.EventType)
		}
		if event.PlanetID != 14 {
			t.Errorf("Expected planet 14 (Mungle), got %d", event.PlanetID)
		}
		if event.Count != 3 {
			t.Errorf("Expected count 3, got %d", event.Count)
		}
	})

	t.Run("MungleQueueEmpty", func(t *testing.T) {
		// Third event: "Mungle has completed its orders. The production queue is empty."
		if len(eventsBlock.ProductionEvents) < 3 {
			t.Fatal("Not enough production events")
		}
		event := eventsBlock.ProductionEvents[2]
		if event.EventType != blocks.EventTypeQueueEmpty {
			t.Errorf("Expected EventTypeQueueEmpty (0x3E), got 0x%02X", event.EventType)
		}
		if event.PlanetID != 14 {
			t.Errorf("Expected planet 14 (Mungle), got %d", event.PlanetID)
		}
	})

	t.Run("PurgatoryQueueEmpty", func(t *testing.T) {
		// Event 20: Queue empty on Purgatory (planet 109)
		if len(eventsBlock.ProductionEvents) < 20 {
			t.Fatal("Not enough production events")
		}
		event := eventsBlock.ProductionEvents[19] // 0-indexed
		if event.EventType != blocks.EventTypeQueueEmpty {
			t.Errorf("Expected EventTypeQueueEmpty (0x3E), got 0x%02X", event.EventType)
		}
		if event.PlanetID != 109 {
			t.Errorf("Expected planet 109 (Purgatory), got %d", event.PlanetID)
		}
	})

	t.Run("VerifyEventTypes", func(t *testing.T) {
		// Count different event types
		counts := make(map[int]int)
		for _, event := range eventsBlock.ProductionEvents {
			counts[event.EventType]++
		}

		// Should have factories, mines, defenses, items, and queue empty events
		if counts[blocks.EventTypeFactoriesBuilt] == 0 {
			t.Error("No factory events found")
		}
		if counts[blocks.EventTypeMinesBuilt] == 0 {
			t.Error("No mine events found")
		}
		if counts[blocks.EventTypeQueueEmpty] == 0 {
			t.Error("No queue empty events found")
		}
		if counts[blocks.EventTypeDefensesBuilt] == 0 {
			t.Error("No defense events found")
		}
	})
}
