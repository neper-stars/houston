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

// Expected data structures for fleetsplit scenario
type ExpectedFleetSplitTransfer struct {
	DesignName string `json:"designName"`
	Count      int    `json:"count"`
}

type ExpectedFleetSplitInfo struct {
	FleetNumber    int    `json:"fleetNumber"`
	FleetDisplayId int    `json:"fleetDisplayId"`
	FleetName      string `json:"fleetName"`
}

type ExpectedShipTransferInfo struct {
	DestFleetNumber      int                          `json:"destFleetNumber"`
	DestFleetDisplayId   int                          `json:"destFleetDisplayId"`
	DestFleetName        string                       `json:"destFleetName"`
	SourceFleetNumber    int                          `json:"sourceFleetNumber"`
	SourceFleetDisplayId int                          `json:"sourceFleetDisplayId"`
	SourceFleetName      string                       `json:"sourceFleetName"`
	Transfers            []ExpectedFleetSplitTransfer `json:"transfers"`
}

type ExpectedFleetRenameInfo struct {
	FleetNumber    int    `json:"fleetNumber"`
	FleetDisplayId int    `json:"fleetDisplayId"`
	OldFleetName   string `json:"oldFleetName"`
	NewFleetName   string `json:"newFleetName"`
}

type ExpectedFleetSplitData struct {
	Scenario     string                   `json:"scenario"`
	Description  string                   `json:"description"`
	FleetSplit   ExpectedFleetSplitInfo   `json:"fleetSplit"`
	ShipTransfer ExpectedShipTransferInfo `json:"shipTransfer"`
	FleetRename  ExpectedFleetRenameInfo  `json:"fleetRename"`
}

func loadFleetSplitExpected(t *testing.T) *ExpectedFleetSplitData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-fleetsplit", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedFleetSplitData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadFleetSplitFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-fleetsplit", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

func TestScenarioFleetSplit_FleetSplitBlock(t *testing.T) {
	expected := loadFleetSplitExpected(t)
	_, blockList := loadFleetSplitFile(t, "game.x1")

	// Find the FleetSplitBlock
	var splitBlock *blocks.FleetSplitBlock
	for _, block := range blockList {
		if fsb, ok := block.(blocks.FleetSplitBlock); ok {
			splitBlock = &fsb
			break
		}
	}

	require.NotNil(t, splitBlock, "Should have a FleetSplitBlock")

	assert.Equal(t, expected.FleetSplit.FleetNumber, splitBlock.FleetNumber,
		"Fleet number should match")
}

func TestScenarioFleetSplit_MoveShipsBlock(t *testing.T) {
	expected := loadFleetSplitExpected(t)
	_, blockList := loadFleetSplitFile(t, "game.x1")

	// Find the MoveShipsBlock
	var moveBlock *blocks.MoveShipsBlock
	for _, block := range blockList {
		if msb, ok := block.(blocks.MoveShipsBlock); ok {
			moveBlock = &msb
			break
		}
	}

	require.NotNil(t, moveBlock, "Should have a MoveShipsBlock")

	// Verify fleet numbers
	assert.Equal(t, expected.ShipTransfer.DestFleetNumber, moveBlock.DestFleetNumber,
		"Destination fleet number should match")
	assert.Equal(t, expected.ShipTransfer.SourceFleetNumber, moveBlock.SourceFleetNumber,
		"Source fleet number should match")

	// Verify ship transfers
	require.Equal(t, len(expected.ShipTransfer.Transfers), len(moveBlock.ShipTransfers),
		"Number of ship transfers should match")

	for i, expTransfer := range expected.ShipTransfer.Transfers {
		actualTransfer := moveBlock.ShipTransfers[i]
		assert.Equal(t, expTransfer.Count, actualTransfer.Count,
			"Ship count for %s should match", expTransfer.DesignName)
	}
}

func TestScenarioFleetSplit_RenameFleetBlock(t *testing.T) {
	expected := loadFleetSplitExpected(t)
	_, blockList := loadFleetSplitFile(t, "game.x1")

	// Find the RenameFleetBlock
	var renameBlock *blocks.RenameFleetBlock
	for _, block := range blockList {
		if rfb, ok := block.(blocks.RenameFleetBlock); ok {
			renameBlock = &rfb
			break
		}
	}

	require.NotNil(t, renameBlock, "Should have a RenameFleetBlock")

	assert.Equal(t, expected.FleetRename.FleetNumber, renameBlock.FleetNumber,
		"Fleet number should match")
	assert.Equal(t, expected.FleetRename.NewFleetName, renameBlock.NewName,
		"New fleet name should match")
}
