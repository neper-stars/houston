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

type waypointRemoteMiningExpected struct {
	Description          string `json:"description"`
	Year                 int    `json:"year"`
	WaypointRemoteMining struct {
		FleetNumber        int    `json:"fleetNumber"`
		FleetName          string `json:"fleetName"`
		WaypointNumber     int    `json:"waypointNumber"`
		X                  int    `json:"x"`
		Y                  int    `json:"y"`
		Warp               int    `json:"warp"`
		Task               int    `json:"task"`
		TaskName           string `json:"taskName"`
		TargetPlanetNumber int    `json:"targetPlanetNumber"`
		TargetPlanetName   string `json:"targetPlanetName"`
		TargetType         int    `json:"targetType"`
		TargetTypeName     string `json:"targetTypeName"`
	} `json:"waypointRemoteMining"`
}

func TestScenarioWaypointRemoteMining(t *testing.T) {
	// Load expected data
	expectedData, err := os.ReadFile("../testdata/scenario-orders/waypoint-remotemining/expected.json")
	require.NoError(t, err, "failed to read expected.json")

	var expected waypointRemoteMiningExpected
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Parse the X file (orders file)
	data, err := os.ReadFile("../testdata/scenario-orders/waypoint-remotemining/game.x1")
	require.NoError(t, err, "failed to read game.x1")

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "failed to parse block list")

	// Find WaypointChangeTaskBlock with Remote Mining task
	var remoteMiningBlocks []blocks.WaypointChangeTaskBlock
	for _, block := range blockList {
		if wctb, ok := block.(blocks.WaypointChangeTaskBlock); ok {
			if wctb.WaypointTask == blocks.WaypointTaskRemoteMining {
				remoteMiningBlocks = append(remoteMiningBlocks, wctb)
			}
		}
	}

	require.Len(t, remoteMiningBlocks, 1, "expected exactly 1 WaypointChangeTaskBlock with Remote Mining task")

	wctb := remoteMiningBlocks[0]

	assert.Equal(t, expected.WaypointRemoteMining.FleetNumber, wctb.FleetNumber,
		"fleet number should be %d (%s)", expected.WaypointRemoteMining.FleetNumber, expected.WaypointRemoteMining.FleetName)
	assert.Equal(t, expected.WaypointRemoteMining.WaypointNumber, wctb.WaypointNumber,
		"waypoint number should be %d", expected.WaypointRemoteMining.WaypointNumber)
	assert.Equal(t, expected.WaypointRemoteMining.X, wctb.X,
		"X should be %d", expected.WaypointRemoteMining.X)
	assert.Equal(t, expected.WaypointRemoteMining.Y, wctb.Y,
		"Y should be %d", expected.WaypointRemoteMining.Y)
	assert.Equal(t, expected.WaypointRemoteMining.Warp, wctb.Warp,
		"warp should be %d", expected.WaypointRemoteMining.Warp)
	assert.Equal(t, expected.WaypointRemoteMining.Task, wctb.WaypointTask,
		"task should be %d (%s)", expected.WaypointRemoteMining.Task, expected.WaypointRemoteMining.TaskName)
	assert.Equal(t, expected.WaypointRemoteMining.TargetPlanetNumber, wctb.Target,
		"target planet should be %d (%s)", expected.WaypointRemoteMining.TargetPlanetNumber, expected.WaypointRemoteMining.TargetPlanetName)
	assert.Equal(t, expected.WaypointRemoteMining.TargetType, wctb.TargetType,
		"target type should be %d (%s)", expected.WaypointRemoteMining.TargetType, expected.WaypointRemoteMining.TargetTypeName)
}
