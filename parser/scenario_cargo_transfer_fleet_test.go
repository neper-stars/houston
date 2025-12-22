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

// Expected data structures for fleet-to-fleet cargo transfer
type ExpectedFleetTransferCargo struct {
	Ironium   int `json:"ironium"`
	Boranium  int `json:"boranium"`
	Germanium int `json:"germanium"`
	Colonists int `json:"colonists"`
}

type ExpectedFleetCargoTransfer struct {
	FleetNumber     int                        `json:"fleetNumber"`
	FleetDisplayId  int                        `json:"fleetDisplayId"`
	FleetName       string                     `json:"fleetName"`
	TargetNumber    int                        `json:"targetNumber"`
	TargetDisplayId int                        `json:"targetDisplayId"`
	TargetName      string                     `json:"targetName"`
	TargetType      string                     `json:"targetType"`
	Cargo           ExpectedFleetTransferCargo `json:"cargo"`
}

type ExpectedFleetCargoTransferData struct {
	Scenario       string                       `json:"scenario"`
	Description    string                       `json:"description"`
	CargoTransfers []ExpectedFleetCargoTransfer `json:"cargoTransfers"`
}

func loadCargoTransferFleetExpected(t *testing.T) *ExpectedFleetCargoTransferData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-cargo-transfer-fleet", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedFleetCargoTransferData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadCargoTransferFleetFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-cargo-transfer-fleet", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

func TestScenarioCargoTransferFleet_BidirectionalTransfer(t *testing.T) {
	expected := loadCargoTransferFleetExpected(t)
	_, blockList := loadCargoTransferFleetFile(t, "game.x1")

	// Collect cargo transfer blocks
	var transfers []blocks.ManualSmallLoadUnloadTaskBlock
	for _, block := range blockList {
		if cb, ok := block.(blocks.ManualSmallLoadUnloadTaskBlock); ok {
			transfers = append(transfers, cb)
		}
	}

	require.Equal(t, len(expected.CargoTransfers), len(transfers),
		"Should have %d cargo transfers", len(expected.CargoTransfers))

	// Validate each transfer
	for i, exp := range expected.CargoTransfers {
		transfer := transfers[i]
		t.Run(exp.FleetName+"<->"+exp.TargetName, func(t *testing.T) {
			// Fleet and target identification
			assert.Equal(t, exp.FleetNumber, transfer.FleetNumber, "Fleet number should match")
			assert.Equal(t, exp.TargetNumber, transfer.TargetNumber, "Target number should match")

			// Cargo amounts (signed for bidirectional transfer)
			assert.Equal(t, exp.Cargo.Ironium, transfer.Ironium,
				"Ironium should match (negative = unload, positive = load)")
			assert.Equal(t, exp.Cargo.Boranium, transfer.Boranium,
				"Boranium should match")
			assert.Equal(t, exp.Cargo.Germanium, transfer.Germanium,
				"Germanium should match")
			assert.Equal(t, exp.Cargo.Colonists, transfer.Colonists,
				"Colonists should match")

			// Verify bidirectional nature - has both positive and negative values
			hasPositive := transfer.Ironium > 0 || transfer.Boranium > 0 ||
				transfer.Germanium > 0 || transfer.Colonists > 0
			hasNegative := transfer.Ironium < 0 || transfer.Boranium < 0 ||
				transfer.Germanium < 0 || transfer.Colonists < 0

			assert.True(t, hasPositive, "Should have at least one positive cargo (load from target)")
			assert.True(t, hasNegative, "Should have at least one negative cargo (unload to target)")
		})
	}
}
