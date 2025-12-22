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

// Expected data structures for fleetmerge scenario
type ExpectedShipTransfer struct {
	DesignName string `json:"designName"`
	Count      int    `json:"count"`
}

type ExpectedFleetMerge struct {
	DestFleetNumber     int                    `json:"destFleetNumber"`
	DestFleetDisplayId  int                    `json:"destFleetDisplayId"`
	DestFleetName       string                 `json:"destFleetName"`
	SourceFleetNumber   int                    `json:"sourceFleetNumber"`
	SourceFleetDisplayId int                   `json:"sourceFleetDisplayId"`
	SourceFleetName     string                 `json:"sourceFleetName"`
	ShipsTransferred    []ExpectedShipTransfer `json:"shipsTransferred"`
}

type ExpectedFleetMergeData struct {
	Scenario      string               `json:"scenario"`
	Description   string               `json:"description"`
	ShipTransfers []ExpectedFleetMerge `json:"shipTransfers"`
}

func loadFleetMergeExpected(t *testing.T) *ExpectedFleetMergeData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-fleetmerge", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedFleetMergeData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadFleetMergeFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-fleetmerge", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

func TestScenarioFleetMerge_MoveShipsBlock(t *testing.T) {
	expected := loadFleetMergeExpected(t)
	_, blockList := loadFleetMergeFile(t, "game.x1")

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
		t.Run(exp.SourceFleetName+"->"+exp.DestFleetName, func(t *testing.T) {
			// Fleet identification
			assert.Equal(t, exp.DestFleetNumber, mb.DestFleetNumber,
				"Destination fleet number should match")
			assert.Equal(t, exp.SourceFleetNumber, mb.SourceFleetNumber,
				"Source fleet number should match")

			// Ship transfers should match
			require.Equal(t, len(exp.ShipsTransferred), len(mb.ShipTransfers),
				"Number of ship transfers should match")

			for j, expTransfer := range exp.ShipsTransferred {
				assert.Equal(t, expTransfer.Count, mb.ShipTransfers[j].Count,
					"Ship count for transfer %d should match", j)
			}
		})
	}
}
