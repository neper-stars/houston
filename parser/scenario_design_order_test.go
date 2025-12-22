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

// Expected data structures for design-order scenario
type ExpectedDesignSlot struct {
	SlotType string `json:"slotType"`
	ItemName string `json:"itemName"`
	ItemId   int    `json:"itemId"`
	Count    int    `json:"count"`
}

type ExpectedDesignChange struct {
	Order        int                  `json:"order"`
	IsDelete     bool                 `json:"isDelete"`
	DesignNumber int                  `json:"designNumber"`
	Name         string               `json:"name"`
	HullId       int                  `json:"hullId"`
	HullName     string               `json:"hullName"`
	IsStarbase   bool                 `json:"isStarbase"`
	Armor        int                  `json:"armor"`
	SlotCount    int                  `json:"slotCount"`
	Slots        []ExpectedDesignSlot `json:"slots"`
}

type ExpectedDesignOrderData struct {
	Scenario      string                 `json:"scenario"`
	Description   string                 `json:"description"`
	DesignChanges []ExpectedDesignChange `json:"designChanges"`
}

func loadDesignOrderExpected(t *testing.T) *ExpectedDesignOrderData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-design-order", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedDesignOrderData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadDesignOrderFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-design-order", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

// slotCategoryToName maps slot category bitmask to human-readable name
func slotCategoryToName(category uint16) string {
	switch category {
	case 1:
		return "Engine"
	case 2:
		return "Scanner"
	case 16:
		return "General Purpose"
	default:
		return "Unknown"
	}
}

func TestScenarioDesignOrder_DesignChangeBlocks(t *testing.T) {
	expected := loadDesignOrderExpected(t)
	_, blockList := loadDesignOrderFile(t, "game.x1")

	// Collect design change blocks
	var designChanges []blocks.DesignChangeBlock
	for _, block := range blockList {
		if dcb, ok := block.(blocks.DesignChangeBlock); ok {
			designChanges = append(designChanges, dcb)
		}
	}

	require.Equal(t, len(expected.DesignChanges), len(designChanges),
		"Should have %d design changes", len(expected.DesignChanges))

	// Validate each design change
	for i, exp := range expected.DesignChanges {
		dcb := designChanges[i]
		t.Run(exp.Name, func(t *testing.T) {
			// Check if delete or create/modify
			assert.Equal(t, exp.IsDelete, dcb.IsDelete, "IsDelete should match")

			if !exp.IsDelete {
				require.NotNil(t, dcb.Design, "Design should not be nil for non-delete")

				// Design metadata
				assert.Equal(t, exp.DesignNumber, dcb.Design.DesignNumber, "Design number should match")
				assert.Equal(t, exp.Name, dcb.Design.Name, "Design name should match")
				assert.Equal(t, exp.HullId, dcb.Design.HullId, "Hull ID should match")
				assert.Equal(t, exp.IsStarbase, dcb.Design.IsStarbase, "IsStarbase should match")

				// Design stats
				assert.Equal(t, exp.Armor, dcb.Design.Armor, "Armor should match")
				assert.Equal(t, exp.SlotCount, dcb.Design.SlotCount, "Slot count should match")
				assert.True(t, dcb.Design.IsFullDesign, "Should be a full design")

				// Component slots
				require.Equal(t, len(exp.Slots), len(dcb.Design.Slots), "Should have %d slots", len(exp.Slots))
				for j, expSlot := range exp.Slots {
					slot := dcb.Design.Slots[j]
					slotType := slotCategoryToName(slot.Category)
					assert.Equal(t, expSlot.SlotType, slotType, "Slot %d type should match", j)
					assert.Equal(t, expSlot.ItemId, slot.ItemId, "Slot %d item ID should match", j)
					assert.Equal(t, expSlot.Count, slot.Count, "Slot %d count should match", j)
				}
			}
		})
	}
}
