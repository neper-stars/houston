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

// Expected data structures for wormhole scenario
type ExpectedWormhole struct {
	Number    int    `json:"number"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	TargetId  int    `json:"targetId"`
	Stability string `json:"stability"`
}

type ExpectedWormholeData struct {
	Scenario  string             `json:"scenario"`
	Wormholes []ExpectedWormhole `json:"wormholes"`
}

func loadWormholeExpected(t *testing.T) *ExpectedWormholeData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-wormhole", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedWormholeData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadWormholeFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-wormhole", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

func TestScenarioWormhole_ObjectBlocks(t *testing.T) {
	expected := loadWormholeExpected(t)
	_, blockList := loadWormholeFile(t, "game.m1")

	// Collect wormhole objects
	var wormholes []blocks.ObjectBlock
	for _, block := range blockList {
		if ob, ok := block.(blocks.ObjectBlock); ok {
			if !ob.IsCountObject && ob.IsWormhole() {
				wormholes = append(wormholes, ob)
			}
		}
	}

	require.Equal(t, len(expected.Wormholes), len(wormholes),
		"Should have %d wormholes", len(expected.Wormholes))

	// Validate each wormhole
	for i, exp := range expected.Wormholes {
		wh := wormholes[i]
		t.Run(exp.Stability, func(t *testing.T) {
			// Validate number
			assert.Equal(t, exp.Number, wh.Number, "Number should match")

			// Validate position
			assert.Equal(t, exp.X, wh.X, "X should match")
			assert.Equal(t, exp.Y, wh.Y, "Y should match")

			// Validate target
			assert.Equal(t, exp.TargetId, wh.TargetId, "TargetId should match")

			// Validate stability
			assert.Equal(t, exp.Stability, wh.StabilityName(), "Stability should match")
		})
	}
}

func TestScenarioWormhole_Connectivity(t *testing.T) {
	_, blockList := loadWormholeFile(t, "game.m1")

	// Collect wormholes into a map by number
	wormholeMap := make(map[int]blocks.ObjectBlock)
	for _, block := range blockList {
		if ob, ok := block.(blocks.ObjectBlock); ok {
			if !ob.IsCountObject && ob.IsWormhole() {
				wormholeMap[ob.Number] = ob
			}
		}
	}

	// Verify each wormhole's target exists and points back
	for num, wh := range wormholeMap {
		target, exists := wormholeMap[wh.TargetId]
		assert.True(t, exists, "Wormhole #%d's target #%d should exist", num, wh.TargetId)
		if exists {
			assert.Equal(t, num, target.TargetId,
				"Wormhole #%d and #%d should connect to each other", num, wh.TargetId)
		}
	}
}
