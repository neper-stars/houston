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

type setFleetBattlePlanExpected struct {
	Description        string `json:"description"`
	Year               int    `json:"year"`
	SetFleetBattlePlan struct {
		FleetNumber     int    `json:"fleetNumber"`
		FleetName       string `json:"fleetName"`
		BattlePlanIndex int    `json:"battlePlanIndex"`
		BattlePlanName  string `json:"battlePlanName"`
	} `json:"setFleetBattlePlan"`
}

func TestScenarioSetFleetBattlePlan(t *testing.T) {
	// Load expected data
	expectedData, err := os.ReadFile("../testdata/scenario-orders/set-fleet-battleplan/expected.json")
	require.NoError(t, err, "failed to read expected.json")

	var expected setFleetBattlePlanExpected
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Parse the X file (orders file)
	data, err := os.ReadFile("../testdata/scenario-orders/set-fleet-battleplan/game.x1")
	require.NoError(t, err, "failed to read game.x1")

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "failed to parse block list")

	// Find SetFleetBattlePlanBlock
	var setBlocks []blocks.SetFleetBattlePlanBlock
	for _, block := range blockList {
		if sfbp, ok := block.(blocks.SetFleetBattlePlanBlock); ok {
			setBlocks = append(setBlocks, sfbp)
		}
	}

	require.Len(t, setBlocks, 1, "expected exactly 1 SetFleetBattlePlanBlock")

	sfbp := setBlocks[0]
	assert.Equal(t, expected.SetFleetBattlePlan.FleetNumber, sfbp.FleetNumber,
		"fleet number should be %d (%s)", expected.SetFleetBattlePlan.FleetNumber, expected.SetFleetBattlePlan.FleetName)
	assert.Equal(t, expected.SetFleetBattlePlan.BattlePlanIndex, sfbp.BattlePlanIndex,
		"battle plan index should be %d (%s)", expected.SetFleetBattlePlan.BattlePlanIndex, expected.SetFleetBattlePlan.BattlePlanName)
}
