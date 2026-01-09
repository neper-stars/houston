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

type transportOrderExpected struct {
	Action     int    `json:"action"`
	ActionName string `json:"actionName"`
	Value      int    `json:"value"`
}

type waypointLoadExpected struct {
	Description        string `json:"description"`
	Year               int    `json:"year"`
	WaypointChangeTask struct {
		FleetNumber    int    `json:"fleetNumber"`
		FleetName      string `json:"fleetName"`
		WaypointIndex  int    `json:"waypointNumber"`
		X              int    `json:"x"`
		Y              int    `json:"y"`
		TargetType     int    `json:"targetType"`
		TargetTypeName string `json:"targetTypeName"`
		Warp           int    `json:"warp"`
		Task           int    `json:"task"`
		TaskName       string `json:"taskName"`
	} `json:"waypointChangeTask"`
	TransportOrders struct {
		Ironium   transportOrderExpected `json:"ironium"`
		Boranium  transportOrderExpected `json:"boranium"`
		Germanium transportOrderExpected `json:"germanium"`
		Colonists transportOrderExpected `json:"colonists"`
	} `json:"transportOrders"`
}

func TestScenarioWaypointLoad(t *testing.T) {
	// Load expected data
	expectedData, err := os.ReadFile("../testdata/scenario-orders/waypoint-load/expected.json")
	require.NoError(t, err, "failed to read expected.json")

	var expected waypointLoadExpected
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Parse the X file (orders file)
	data, err := os.ReadFile("../testdata/scenario-orders/waypoint-load/game.x1")
	require.NoError(t, err, "failed to read game.x1")

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "failed to parse block list")

	// Find WaypointChangeTaskBlock with Transport task
	var transportBlocks []blocks.WaypointChangeTaskBlock
	for _, block := range blockList {
		if wctb, ok := block.(blocks.WaypointChangeTaskBlock); ok {
			if wctb.WaypointTask == blocks.WaypointTaskTransport {
				transportBlocks = append(transportBlocks, wctb)
			}
		}
	}

	require.Len(t, transportBlocks, 1, "expected exactly 1 WaypointChangeTaskBlock with Transport task")

	wctb := transportBlocks[0]

	t.Run("WaypointChangeTask", func(t *testing.T) {
		assert.Equal(t, expected.WaypointChangeTask.FleetNumber, wctb.FleetNumber,
			"fleet number should be %d (%s)", expected.WaypointChangeTask.FleetNumber, expected.WaypointChangeTask.FleetName)
		assert.Equal(t, expected.WaypointChangeTask.WaypointIndex, wctb.WaypointIndex,
			"waypoint index should be %d", expected.WaypointChangeTask.WaypointIndex)
		assert.Equal(t, expected.WaypointChangeTask.X, wctb.X,
			"X should be %d", expected.WaypointChangeTask.X)
		assert.Equal(t, expected.WaypointChangeTask.Y, wctb.Y,
			"Y should be %d", expected.WaypointChangeTask.Y)
		assert.Equal(t, expected.WaypointChangeTask.TargetType, wctb.TargetType,
			"target type should be %d (%s)", expected.WaypointChangeTask.TargetType, expected.WaypointChangeTask.TargetTypeName)
		assert.Equal(t, expected.WaypointChangeTask.Warp, wctb.Warp,
			"warp should be %d", expected.WaypointChangeTask.Warp)
		assert.Equal(t, expected.WaypointChangeTask.Task, wctb.WaypointTask,
			"task should be %d (%s)", expected.WaypointChangeTask.Task, expected.WaypointChangeTask.TaskName)
	})

	t.Run("TransportOrders", func(t *testing.T) {
		// Ironium: Load Exactly 18 kT
		assert.Equal(t, expected.TransportOrders.Ironium.Action, wctb.TransportOrders[blocks.CargoIronium].Action,
			"ironium action should be %d (%s)", expected.TransportOrders.Ironium.Action, expected.TransportOrders.Ironium.ActionName)
		assert.Equal(t, expected.TransportOrders.Ironium.Value, wctb.TransportOrders[blocks.CargoIronium].Value,
			"ironium value should be %d", expected.TransportOrders.Ironium.Value)

		// Boranium: Load All Available
		assert.Equal(t, expected.TransportOrders.Boranium.Action, wctb.TransportOrders[blocks.CargoBoranium].Action,
			"boranium action should be %d (%s)", expected.TransportOrders.Boranium.Action, expected.TransportOrders.Boranium.ActionName)
		assert.Equal(t, expected.TransportOrders.Boranium.Value, wctb.TransportOrders[blocks.CargoBoranium].Value,
			"boranium value should be %d", expected.TransportOrders.Boranium.Value)

		// Germanium: Fill Up to 50%
		assert.Equal(t, expected.TransportOrders.Germanium.Action, wctb.TransportOrders[blocks.CargoGermanium].Action,
			"germanium action should be %d (%s)", expected.TransportOrders.Germanium.Action, expected.TransportOrders.Germanium.ActionName)
		assert.Equal(t, expected.TransportOrders.Germanium.Value, wctb.TransportOrders[blocks.CargoGermanium].Value,
			"germanium value should be %d%%", expected.TransportOrders.Germanium.Value)

		// Colonists: No Action
		assert.Equal(t, expected.TransportOrders.Colonists.Action, wctb.TransportOrders[blocks.CargoColonists].Action,
			"colonists action should be %d (%s)", expected.TransportOrders.Colonists.Action, expected.TransportOrders.Colonists.ActionName)
	})

	t.Run("TransportTaskName", func(t *testing.T) {
		// Verify the TransportTaskName helper function
		assert.Equal(t, "Load Exactly", blocks.TransportTaskName(blocks.TransportTaskLoadExactly))
		assert.Equal(t, "Load All Available", blocks.TransportTaskName(blocks.TransportTaskLoadAll))
		assert.Equal(t, "Fill Up to %", blocks.TransportTaskName(blocks.TransportTaskFillToPercent))
		assert.Equal(t, "No Action", blocks.TransportTaskName(blocks.TransportTaskNoAction))
	})
}
