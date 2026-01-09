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

// Expected data structures for stargate scenario
type ExpectedStargateOrder struct {
	FleetNumber   int  `json:"fleetNumber"`
	WaypointIndex int  `json:"waypointNumber"`
	DestinationX  int  `json:"destinationX"`
	DestinationY  int  `json:"destinationY"`
	TargetPlanet  int  `json:"targetPlanet"`
	UsesStargate  bool `json:"usesStargate"`
}

type ExpectedStargateData struct {
	Scenario       string                  `json:"scenario"`
	StargateOrders []ExpectedStargateOrder `json:"stargateOrders"`
}

func loadStargateExpected(t *testing.T) *ExpectedStargateData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-stargate", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedStargateData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadStargateFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-stargate", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

func TestScenarioStargate_WaypointAddBlock(t *testing.T) {
	expected := loadStargateExpected(t)
	_, blockList := loadStargateFile(t, "game.x1")

	// Collect stargate waypoint orders
	var stargateOrders []blocks.WaypointAddBlock
	for _, block := range blockList {
		if wab, ok := block.(blocks.WaypointAddBlock); ok {
			if wab.UsesStargate() {
				stargateOrders = append(stargateOrders, wab)
			}
		}
	}

	require.Equal(t, len(expected.StargateOrders), len(stargateOrders),
		"Should have %d stargate orders", len(expected.StargateOrders))

	// Validate each stargate order
	for i, exp := range expected.StargateOrders {
		order := stargateOrders[i]
		t.Run("StargateOrder", func(t *testing.T) {
			// Validate fleet
			assert.Equal(t, exp.FleetNumber, order.FleetNumber, "Fleet number should match")
			assert.Equal(t, exp.WaypointIndex, order.WaypointIndex, "Waypoint index should match")

			// Validate destination
			assert.Equal(t, exp.DestinationX, order.X, "Destination X should match")
			assert.Equal(t, exp.DestinationY, order.Y, "Destination Y should match")

			// Validate target planet
			assert.Equal(t, exp.TargetPlanet, order.Target, "Target planet should match")

			// Validate stargate usage
			assert.Equal(t, exp.UsesStargate, order.UsesStargate(), "UsesStargate should match")
			assert.Equal(t, blocks.WarpStargate, order.Warp, "Warp should be stargate value (11)")
		})
	}
}
