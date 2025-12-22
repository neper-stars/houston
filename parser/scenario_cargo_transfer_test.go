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

// Expected data structures for cargo-transfer scenario
type ExpectedTransferCargo struct {
	Ironium   int `json:"ironium"`
	Boranium  int `json:"boranium"`
	Germanium int `json:"germanium"`
	Colonists int `json:"colonists"`
}

type ExpectedCargoTransfer struct {
	FleetNumber     int                   `json:"fleetNumber"`
	FleetDisplayId  int                   `json:"fleetDisplayId"`
	FleetName       string                `json:"fleetName"`
	TargetNumber    int                   `json:"targetNumber"`
	TargetDisplayId int                   `json:"targetDisplayId"`
	TargetName      string                `json:"targetName"`
	Direction       string                `json:"direction"`
	Cargo           ExpectedTransferCargo `json:"cargo"`
}

type ExpectedCargoTransferData struct {
	Scenario       string                  `json:"scenario"`
	Description    string                  `json:"description"`
	CargoTransfers []ExpectedCargoTransfer `json:"cargoTransfers"`
}

func loadCargoTransferExpected(t *testing.T) *ExpectedCargoTransferData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-cargo-transfer", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedCargoTransferData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadCargoTransferFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-cargo-transfer", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

func TestScenarioCargoTransfer_SmallLoadUnload(t *testing.T) {
	expected := loadCargoTransferExpected(t)
	_, blockList := loadCargoTransferFile(t, "game.x1")

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
		t.Run(exp.FleetName+"->"+exp.TargetName, func(t *testing.T) {
			// Fleet and target identification
			assert.Equal(t, exp.FleetNumber, transfer.FleetNumber, "Fleet number should match")
			assert.Equal(t, exp.TargetNumber, transfer.TargetNumber, "Target number should match")

			// Direction
			if exp.Direction == "load" {
				assert.True(t, transfer.IsLoad(), "Should be a load operation (target -> fleet)")
				assert.False(t, transfer.IsUnload(), "Should not be an unload operation")
			} else {
				assert.True(t, transfer.IsUnload(), "Should be an unload operation (fleet -> target)")
				assert.False(t, transfer.IsLoad(), "Should not be a load operation")
			}

			// Cargo amounts
			assert.Equal(t, exp.Cargo.Ironium, transfer.Ironium, "Ironium should match")
			assert.Equal(t, exp.Cargo.Boranium, transfer.Boranium, "Boranium should match")
			assert.Equal(t, exp.Cargo.Germanium, transfer.Germanium, "Germanium should match")
			assert.Equal(t, exp.Cargo.Colonists, transfer.Colonists, "Colonists should match")

			// Cargo mask should have all 4 bits set (0x0F) since all cargo types are transferred
			if exp.Cargo.Ironium > 0 {
				assert.True(t, transfer.HasIronium(), "Should have ironium flag set")
			}
			if exp.Cargo.Boranium > 0 {
				assert.True(t, transfer.HasBoranium(), "Should have boranium flag set")
			}
			if exp.Cargo.Germanium > 0 {
				assert.True(t, transfer.HasGermanium(), "Should have germanium flag set")
			}
			if exp.Cargo.Colonists > 0 {
				assert.True(t, transfer.HasColonists(), "Should have colonists flag set")
			}
		})
	}
}
