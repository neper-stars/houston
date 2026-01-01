package parser_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

type patrolWaypointExpected struct {
	WaypointNumber  int    `json:"waypointNumber"`
	X               int    `json:"x"`
	Y               int    `json:"y"`
	PatrolRange     int    `json:"patrolRange"`
	PatrolRangeLY   int    `json:"patrolRangeLY"`
	PatrolRangeName string `json:"patrolRangeName"`
}

type waypointPatrolRepeatExpected struct {
	Description  string `json:"description"`
	Year         int    `json:"year"`
	RepeatOrders struct {
		FleetNumber        int    `json:"fleetNumber"`
		FleetName          string `json:"fleetName"`
		RepeatFromWaypoint int    `json:"repeatFromWaypoint"`
	} `json:"repeatOrders"`
	PatrolWaypoints []patrolWaypointExpected `json:"patrolWaypoints"`
}

func TestScenarioWaypointPatrolRepeat(t *testing.T) {
	// Load expected data
	expectedData, err := os.ReadFile("../testdata/scenario-orders/waypoint-patrol-repeat/expected.json")
	require.NoError(t, err, "failed to read expected.json")

	var expected waypointPatrolRepeatExpected
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Parse the X file (orders file)
	data, err := os.ReadFile("../testdata/scenario-orders/waypoint-patrol-repeat/game.x1")
	require.NoError(t, err, "failed to read game.x1")

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "failed to parse block list")

	t.Run("WaypointRepeatOrdersBlock", func(t *testing.T) {
		// Find WaypointRepeatOrdersBlock
		var repeatBlocks []blocks.WaypointRepeatOrdersBlock
		for _, block := range blockList {
			if wrob, ok := block.(blocks.WaypointRepeatOrdersBlock); ok {
				repeatBlocks = append(repeatBlocks, wrob)
			}
		}

		require.Len(t, repeatBlocks, 1, "expected exactly 1 WaypointRepeatOrdersBlock")

		wrob := repeatBlocks[0]
		assert.Equal(t, expected.RepeatOrders.FleetNumber, wrob.FleetNumber,
			"fleet number should be %d (%s)", expected.RepeatOrders.FleetNumber, expected.RepeatOrders.FleetName)
		assert.Equal(t, expected.RepeatOrders.RepeatFromWaypoint, wrob.RepeatFromWaypoint,
			"repeat from waypoint should be %d", expected.RepeatOrders.RepeatFromWaypoint)
	})

	t.Run("PatrolWaypointsWithRange", func(t *testing.T) {
		// Find WaypointChangeTaskBlocks with Patrol task and non-zero range
		var patrolBlocks []blocks.WaypointChangeTaskBlock
		for _, block := range blockList {
			if wctb, ok := block.(blocks.WaypointChangeTaskBlock); ok {
				// Only include patrol tasks with explicit range (15-byte blocks)
				if wctb.WaypointTask == blocks.WaypointTaskPatrol && wctb.PatrolRange > 0 {
					patrolBlocks = append(patrolBlocks, wctb)
				}
			}
		}

		require.Len(t, patrolBlocks, len(expected.PatrolWaypoints),
			"expected %d patrol waypoints with explicit range", len(expected.PatrolWaypoints))

		for i, exp := range expected.PatrolWaypoints {
			wctb := patrolBlocks[i]
			t.Run(exp.PatrolRangeName, func(t *testing.T) {
				assert.Equal(t, exp.WaypointNumber, wctb.WaypointNumber,
					"waypoint number should be %d", exp.WaypointNumber)
				assert.Equal(t, exp.X, wctb.X, "X should be %d", exp.X)
				assert.Equal(t, exp.Y, wctb.Y, "Y should be %d", exp.Y)
				assert.Equal(t, exp.PatrolRange, wctb.PatrolRange,
					"patrol range value should be %d", exp.PatrolRange)
				assert.Equal(t, exp.PatrolRangeLY, blocks.PatrolRangeLY(wctb.PatrolRange),
					"patrol range should be %d ly", exp.PatrolRangeLY)
				assert.Equal(t, exp.PatrolRangeName, blocks.PatrolRangeName(wctb.PatrolRange),
					"patrol range name should be %s", exp.PatrolRangeName)
			})
		}
	})

	t.Run("PatrolRangeHelpers", func(t *testing.T) {
		// Test PatrolRangeLY helper
		assert.Equal(t, 50, blocks.PatrolRangeLY(0), "range 0 should be 50 ly")
		assert.Equal(t, 100, blocks.PatrolRangeLY(1), "range 1 should be 100 ly")
		assert.Equal(t, 150, blocks.PatrolRangeLY(2), "range 2 should be 150 ly")
		assert.Equal(t, 550, blocks.PatrolRangeLY(10), "range 10 should be 550 ly")
		assert.Equal(t, -1, blocks.PatrolRangeLY(11), "range 11 (any enemy) should be -1")

		// Test PatrolRangeName helper
		assert.Equal(t, "within 50 ly", blocks.PatrolRangeName(0))
		assert.Equal(t, "within 100 ly", blocks.PatrolRangeName(1))
		assert.Equal(t, "any enemy", blocks.PatrolRangeName(11))
	})
}
