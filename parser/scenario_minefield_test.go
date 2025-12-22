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

// Expected data structures for minefield scenario
type ExpectedMinefield struct {
	Number    int    `json:"number"`
	Owner     string `json:"owner"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Type      string `json:"type"`
	MineCount int64  `json:"mineCount"`
}

type ExpectedMinefieldData struct {
	Scenario   string              `json:"scenario"`
	Minefields []ExpectedMinefield `json:"minefields"`
}

func loadMinefieldExpected(t *testing.T) *ExpectedMinefieldData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-minefield", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedMinefieldData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadMinefieldFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-minefield", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

func minefieldTypeName(t int) string {
	switch t {
	case blocks.MinefieldTypeStandard:
		return "Standard"
	case blocks.MinefieldTypeHeavy:
		return "Heavy"
	case blocks.MinefieldTypeSpeedBump:
		return "Speed Bump"
	default:
		return "Unknown"
	}
}

func TestScenarioMinefield_ObjectBlocks(t *testing.T) {
	expected := loadMinefieldExpected(t)
	_, blockList := loadMinefieldFile(t, "game.m1")

	// Build player ID to name map
	playerNames := make(map[int]string)
	for _, block := range blockList {
		if p, ok := block.(blocks.PlayerBlock); ok {
			playerNames[p.PlayerNumber] = p.NamePlural
		}
	}

	// Collect minefield objects
	var minefields []blocks.ObjectBlock
	for _, block := range blockList {
		if ob, ok := block.(blocks.ObjectBlock); ok {
			if !ob.IsCountObject && ob.IsMinefield() {
				minefields = append(minefields, ob)
			}
		}
	}

	require.Equal(t, len(expected.Minefields), len(minefields),
		"Should have %d minefields", len(expected.Minefields))

	// Validate each minefield
	for i, exp := range expected.Minefields {
		mf := minefields[i]
		t.Run(exp.Type, func(t *testing.T) {
			// Validate owner name
			actualOwnerName := playerNames[mf.Owner]
			assert.Equal(t, exp.Owner, actualOwnerName, "Owner should match")

			// Validate position
			assert.Equal(t, exp.X, mf.X, "X should match")
			assert.Equal(t, exp.Y, mf.Y, "Y should match")

			// Validate minefield type
			actualTypeName := minefieldTypeName(mf.MinefieldType)
			assert.Equal(t, exp.Type, actualTypeName, "Type should match")

			// Validate mine count
			assert.Equal(t, exp.MineCount, mf.MineCount, "Mine count should match")

			// Validate number
			assert.Equal(t, exp.Number, mf.Number, "Number should match")
		})
	}
}
