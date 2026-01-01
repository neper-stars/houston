package parser_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

// TestScenarioFleetRename tests the RenameFleetBlock (Type 44) parsing in X files.
// Test data: Fleet renamed to "Scoutty"
func TestScenarioFleetRename(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-orders/fleetnames/orders/game.x1")
	require.NoError(t, err)

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err)

	var renameBlock *blocks.RenameFleetBlock
	for _, block := range blockList {
		if rb, ok := block.(blocks.RenameFleetBlock); ok {
			renameBlock = &rb
			break
		}
	}

	require.NotNil(t, renameBlock, "should find RenameFleetBlock")
	assert.Equal(t, 0, renameBlock.FleetNumber, "fleet number should be 0")
	assert.Equal(t, "Scoutty", renameBlock.NewName, "new name should be 'Scoutty'")
}

// TestScenarioFleetName tests the FleetNameBlock (Type 21) parsing in M files.
// After turn generation, the renamed fleet has a FleetNameBlock preceding its FleetBlock.
func TestScenarioFleetName(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-orders/fleetnames/results/game.m1")
	require.NoError(t, err)

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err)

	var nameBlock *blocks.FleetNameBlock
	for _, block := range blockList {
		if fnb, ok := block.(blocks.FleetNameBlock); ok {
			nameBlock = &fnb
			break
		}
	}

	require.NotNil(t, nameBlock, "should find FleetNameBlock")
	assert.Equal(t, "Scoutty", nameBlock.Name, "fleet name should be 'Scoutty'")
}

// TestExtractFleets tests the higher-level ExtractFleets API that associates
// FleetNameBlocks with their corresponding FleetBlocks.
func TestExtractFleets(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-orders/fleetnames/results/game.m1")
	require.NoError(t, err)

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err)

	// Use ExtractAllFleetInfo to get fleets with design names resolved
	fleets := parser.ExtractAllFleetInfo(blockList)
	require.NotEmpty(t, fleets, "should find fleets")

	// Find the fleet with custom name
	var customNamedFleet *parser.FleetInfo
	var defaultNamedFleet *parser.FleetInfo
	for _, fi := range fleets {
		if fi.HasCustomName {
			customNamedFleet = fi
		} else if defaultNamedFleet == nil {
			defaultNamedFleet = fi
		}
	}

	// Test custom-named fleet
	require.NotNil(t, customNamedFleet, "should have a fleet with custom name")
	assert.Equal(t, "Scoutty", customNamedFleet.CustomName)
	assert.Equal(t, "Scoutty", customNamedFleet.Name()) // Custom name is used

	// Test default-named fleet - name is auto-generated from design
	require.NotNil(t, defaultNamedFleet, "should have a fleet with default name")
	assert.False(t, defaultNamedFleet.HasCustomName)
	assert.Equal(t, "", defaultNamedFleet.CustomName)
	// Name() auto-resolves from PrimaryDesign
	name := defaultNamedFleet.Name()
	assert.Contains(t, name, "#") // Should have fleet number
	// If design was resolved, name should contain design name
	if defaultNamedFleet.PrimaryDesign != nil {
		assert.Contains(t, name, defaultNamedFleet.PrimaryDesign.Name)
	}
}

// TestExtractDesigns tests the design extraction API.
func TestExtractDesigns(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-orders/fleetnames/results/game.m1")
	require.NoError(t, err)

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err)

	designs := parser.ExtractDesigns(blockList)
	require.NotEmpty(t, designs, "should find designs")

	// Check that designs have names
	for slot, design := range designs {
		assert.Equal(t, slot, design.DesignNumber)
		assert.NotEmpty(t, design.Name, "design should have a name")
	}
}

// TestExtractFleetsMap tests the map-based fleet lookup.
func TestExtractFleetsMap(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-orders/fleetnames/results/game.m1")
	require.NoError(t, err)

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err)

	fleetsMap := parser.ExtractFleetsMap(blockList)
	require.NotEmpty(t, fleetsMap, "should find fleets")

	// Should be able to look up fleets by number
	for fleetNum, fi := range fleetsMap {
		assert.Equal(t, fleetNum, fi.Fleet.FleetNumber)
	}
}
