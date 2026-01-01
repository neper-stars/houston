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

type waypointMergeExpected struct {
	Description   string `json:"description"`
	Year          int    `json:"year"`
	WaypointMerge struct {
		FleetNumber       int    `json:"fleetNumber"`
		FleetName         string `json:"fleetName"`
		WaypointNumber    int    `json:"waypointNumber"`
		X                 int    `json:"x"`
		Y                 int    `json:"y"`
		DestinationName   string `json:"destinationName"`
		Warp              int    `json:"warp"`
		UsesStargate      bool   `json:"usesStargate"`
		Task              int    `json:"task"`
		TaskName          string `json:"taskName"`
		TargetType        int    `json:"targetType"`
		TargetTypeName    string `json:"targetTypeName"`
		TargetFleetNumber int    `json:"targetFleetNumber"`
		TargetFleetName   string `json:"targetFleetName"`
	} `json:"waypointMerge"`
}

func TestScenarioWaypointMerge(t *testing.T) {
	// Load expected data
	expectedData, err := os.ReadFile("../testdata/scenario-orders/waypoint-merge/expected.json")
	require.NoError(t, err, "failed to read expected.json")

	var expected waypointMergeExpected
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Parse the X file (orders file)
	data, err := os.ReadFile("../testdata/scenario-orders/waypoint-merge/game.x1")
	require.NoError(t, err, "failed to read game.x1")

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "failed to parse block list")

	// Find WaypointChangeTaskBlock with Merge Fleet task
	var mergeBlocks []blocks.WaypointChangeTaskBlock
	for _, block := range blockList {
		if wctb, ok := block.(blocks.WaypointChangeTaskBlock); ok {
			if wctb.WaypointTask == blocks.WaypointTaskMergeFleet {
				mergeBlocks = append(mergeBlocks, wctb)
			}
		}
	}

	require.Len(t, mergeBlocks, 1, "expected exactly 1 WaypointChangeTaskBlock with Merge Fleet task")

	wctb := mergeBlocks[0]

	assert.Equal(t, expected.WaypointMerge.FleetNumber, wctb.FleetNumber,
		"fleet number should be %d (%s)", expected.WaypointMerge.FleetNumber, expected.WaypointMerge.FleetName)
	assert.Equal(t, expected.WaypointMerge.WaypointNumber, wctb.WaypointNumber,
		"waypoint number should be %d", expected.WaypointMerge.WaypointNumber)
	assert.Equal(t, expected.WaypointMerge.X, wctb.X,
		"X should be %d", expected.WaypointMerge.X)
	assert.Equal(t, expected.WaypointMerge.Y, wctb.Y,
		"Y should be %d", expected.WaypointMerge.Y)
	assert.Equal(t, expected.WaypointMerge.Warp, wctb.Warp,
		"warp should be %d", expected.WaypointMerge.Warp)
	assert.Equal(t, expected.WaypointMerge.UsesStargate, wctb.UsesStargate(),
		"UsesStargate() should be %v", expected.WaypointMerge.UsesStargate)
	assert.Equal(t, expected.WaypointMerge.Task, wctb.WaypointTask,
		"task should be %d (%s)", expected.WaypointMerge.Task, expected.WaypointMerge.TaskName)
	assert.Equal(t, expected.WaypointMerge.TargetType, wctb.TargetType,
		"target type should be %d (%s)", expected.WaypointMerge.TargetType, expected.WaypointMerge.TargetTypeName)
	assert.Equal(t, expected.WaypointMerge.TargetFleetNumber, wctb.Target,
		"target fleet should be %d (%s)", expected.WaypointMerge.TargetFleetNumber, expected.WaypointMerge.TargetFleetName)
}
