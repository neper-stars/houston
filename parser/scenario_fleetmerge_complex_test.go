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

// Expected data structures for complex fleetmerge scenario
type ExpectedComplexTransfer struct {
	DesignName string `json:"designName"`
	Count      int    `json:"count"`
}

type ExpectedComplexFleetMerge struct {
	DestFleetNumber      int                       `json:"destFleetNumber"`
	DestFleetDisplayId   int                       `json:"destFleetDisplayId"`
	DestFleetName        string                    `json:"destFleetName"`
	SourceFleetNumber    int                       `json:"sourceFleetNumber"`
	SourceFleetDisplayId int                       `json:"sourceFleetDisplayId"`
	SourceFleetName      string                    `json:"sourceFleetName"`
	Transfers            []ExpectedComplexTransfer `json:"transfers"`
}

type ExpectedComplexFleetMergeData struct {
	Scenario      string                      `json:"scenario"`
	Description   string                      `json:"description"`
	ShipTransfers []ExpectedComplexFleetMerge `json:"shipTransfers"`
}

func loadFleetMergeComplexExpected(t *testing.T) *ExpectedComplexFleetMergeData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-fleetmerge-complex", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedComplexFleetMergeData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadFleetMergeComplexFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-fleetmerge-complex", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

func TestScenarioFleetMergeComplex_MoveShipsBlock(t *testing.T) {
	expected := loadFleetMergeComplexExpected(t)
	_, blockList := loadFleetMergeComplexFile(t, "game.x1")

	// Collect move ships blocks
	var moveBlocks []blocks.MoveShipsBlock
	for _, block := range blockList {
		if mb, ok := block.(blocks.MoveShipsBlock); ok {
			moveBlocks = append(moveBlocks, mb)
		}
	}

	require.Equal(t, len(expected.ShipTransfers), len(moveBlocks),
		"Should have %d ship transfer blocks", len(expected.ShipTransfers))

	// Validate each transfer
	for i, exp := range expected.ShipTransfers {
		mb := moveBlocks[i]
		t.Run(exp.DestFleetName+"<->"+exp.SourceFleetName, func(t *testing.T) {
			// Fleet identification
			assert.Equal(t, exp.DestFleetNumber, mb.DestFleetNumber,
				"Destination fleet number should match")
			assert.Equal(t, exp.SourceFleetNumber, mb.SourceFleetNumber,
				"Source fleet number should match")

			// Ship transfers should match
			require.Equal(t, len(exp.Transfers), len(mb.ShipTransfers),
				"Number of ship transfers should match")

			for j, expTransfer := range exp.Transfers {
				actualTransfer := mb.ShipTransfers[j]
				assert.Equal(t, expTransfer.Count, actualTransfer.Count,
					"Ship count for %s should match", expTransfer.DesignName)
			}
		})
	}
}

func TestScenarioFleetMergeComplex_BidirectionalTransfer(t *testing.T) {
	_, blockList := loadFleetMergeComplexFile(t, "game.x1")

	// Find the move ships block
	var mb *blocks.MoveShipsBlock
	for _, block := range blockList {
		if msb, ok := block.(blocks.MoveShipsBlock); ok {
			mb = &msb
			break
		}
	}
	require.NotNil(t, mb, "Should have a MoveShipsBlock")

	// Verify bidirectional transfer: some positive, some negative counts
	hasPositive := false
	hasNegative := false
	for _, transfer := range mb.ShipTransfers {
		if transfer.Count > 0 {
			hasPositive = true
		}
		if transfer.Count < 0 {
			hasNegative = true
		}
	}

	assert.True(t, hasPositive, "Should have at least one positive transfer (ships arriving at dest)")
	assert.True(t, hasNegative, "Should have at least one negative transfer (ships leaving dest)")
}
