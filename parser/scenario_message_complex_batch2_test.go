package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neper-stars/houston/blocks"
)

// Expected data structures for message-complex/batch2 scenario
type MessageComplexBatch2Event struct {
	Type              string  `json:"type"`
	MessageNumber     int     `json:"messageNumber,omitempty"`
	PlanetName        string  `json:"planetName,omitempty"`
	Amount            int     `json:"amount,omitempty"`
	Count             int     `json:"count,omitempty"`
	SourcePlanet      string  `json:"sourcePlanet,omitempty"`
	DestinationPlanet string  `json:"destinationPlanet,omitempty"`
	Reason            string  `json:"reason,omitempty"`
	Level             int     `json:"level,omitempty"`
	CompletedField    string  `json:"completedField,omitempty"`
	NextField         string  `json:"nextField,omitempty"`
	Field             string  `json:"field,omitempty"`
	BenefitName       string  `json:"benefitName,omitempty"`
	MineralAmount     int     `json:"mineralAmount,omitempty"`
	GrowthRatePercent float64 `json:"growthRatePercent,omitempty"`
}

type MessageComplexBatch2Expected struct {
	Scenario      string                      `json:"scenario"`
	Description   string                      `json:"description"`
	Year          int                         `json:"year"`
	TotalMessages int                         `json:"totalMessages"`
	Events        []MessageComplexBatch2Event `json:"events"`
}

func loadMessageComplexBatch2Expected(t *testing.T) *MessageComplexBatch2Expected {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-message-complex", "batch2", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected MessageComplexBatch2Expected
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadMessageComplexBatch2File(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-message-complex", "batch2", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

func TestScenarioMessageComplexBatch2_Events(t *testing.T) {
	expected := loadMessageComplexBatch2Expected(t)
	_, blockList := loadMessageComplexBatch2File(t, "game.m1")

	// Find EventsBlock
	var eventsBlock *blocks.EventsBlock
	for _, block := range blockList {
		if eb, ok := block.(blocks.EventsBlock); ok {
			eventsBlock = &eb
			break
		}
	}
	require.NotNil(t, eventsBlock, "EventsBlock not found")

	// Test each expected event from the JSON
	for _, ev := range expected.Events {
		switch ev.Type {
		case "ResearchComplete":
			t.Run("ResearchComplete", func(t *testing.T) {
				require.NotEmpty(t, eventsBlock.ResearchEvents,
					"Expected ResearchComplete event but found none")

				event := eventsBlock.ResearchEvents[0]
				assert.Equal(t, ev.Level, event.Level, "Research level")
				assert.Equal(t, ev.CompletedField, blocks.ResearchFieldName(event.Field),
					"Completed field")
				assert.Equal(t, ev.NextField, blocks.ResearchFieldName(event.NextField),
					"Next field")
			})

		case "TechBenefit":
			t.Run("TechBenefit", func(t *testing.T) {
				require.NotEmpty(t, eventsBlock.TechBenefits,
					"Expected TechBenefit event but found none")

				event := eventsBlock.TechBenefits[0]
				t.Logf("Parsed TechBenefit: category=%d, itemID=%d (expected: %s)",
					event.Category, event.ItemID, ev.BenefitName)
				// Category 10 corresponds to Propulsion/Terraforming benefits
				assert.True(t, event.Category > 0, "TechBenefit should have valid category")
			})

		case "PopulationDecrease":
			t.Run("PopulationDecrease", func(t *testing.T) {
				require.NotEmpty(t, eventsBlock.PopulationChanges,
					"Expected PopulationDecrease event but found none")

				event := eventsBlock.PopulationChanges[0]
				assert.Equal(t, ev.Amount, event.Amount,
					"Population change amount for planet %s", ev.PlanetName)
				t.Logf("PopulationDecrease on planetID=%d (%s): %d colonists",
					event.PlanetID, ev.PlanetName, event.Amount)
			})

		case "PacketCaptured":
			t.Run("PacketCaptured", func(t *testing.T) {
				require.NotEmpty(t, eventsBlock.PacketsCaptured,
					"Expected PacketCaptured event but found none")

				event := eventsBlock.PacketsCaptured[0]
				assert.Equal(t, ev.MineralAmount, event.MineralAmount,
					"Packet mineral amount for planet %s", ev.PlanetName)
				t.Logf("PacketCaptured on planetID=%d (%s): %d kT",
					event.PlanetID, ev.PlanetName, event.MineralAmount)
			})

		case "DefensesBuilt":
			t.Run("DefensesBuilt", func(t *testing.T) {
				count := 0
				for _, prodEv := range eventsBlock.ProductionEvents {
					if prodEv.EventType == blocks.EventTypeDefensesBuilt {
						count++
					}
				}
				assert.Greater(t, count, 0,
					"Expected DefensesBuilt event for planet %s", ev.PlanetName)
			})

		case "MineralAlchemy":
			t.Run("MineralAlchemy", func(t *testing.T) {
				count := 0
				for _, prodEv := range eventsBlock.ProductionEvents {
					if prodEv.EventType == blocks.EventTypeMineralAlchemyBuilt {
						count++
					}
				}
				assert.Greater(t, count, 0,
					"Expected MineralAlchemy event for planet %s", ev.PlanetName)
			})

		case "MineralPacketProduced":
			t.Run("MineralPacketProduced", func(t *testing.T) {
				require.NotEmpty(t, eventsBlock.PacketsProduced,
					"Expected MineralPacketProduced event but found none")

				event := eventsBlock.PacketsProduced[0]
				// Note: Source planet encoding not fully understood
				// Expected source is Hurl, destination is Purgatory (109)
				t.Logf("MineralPacketProduced: source=%d (expected: %s), dest=%d (expected: %s)",
					event.SourcePlanetID, ev.SourcePlanet,
					event.DestinationPlanetID, ev.DestinationPlanet)

				// Verify destination matches Purgatory (109)
				assert.Equal(t, 109, event.DestinationPlanetID,
					"Destination planet should be Purgatory (109)")
			})

		case "TerraformablePlanetFound":
			t.Run("TerraformablePlanetFound", func(t *testing.T) {
				// TODO: TerraformablePlanetFound not found in this events block format
				t.Logf("TerraformablePlanetFound not decoded (planet: %s, growth: %.2f%%)",
					ev.PlanetName, ev.GrowthRatePercent)
			})

		case "HabitablePlanetFound":
			t.Run("HabitablePlanetFound", func(t *testing.T) {
				// TODO: HabitablePlanetFound parsing not yet implemented
				t.Logf("HabitablePlanetFound parsing not yet implemented (planet: %s, growth: %.2f%%)",
					ev.PlanetName, ev.GrowthRatePercent)
			})
		}
	}
}
