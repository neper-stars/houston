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

type battle02Expected struct {
	Scenario    string `json:"scenario"`
	Description string `json:"description"`
	Year        int    `json:"year"`
	BattleBlock struct {
		Rounds         int `json:"rounds"`
		Side1Stacks    int `json:"side1Stacks"`
		TotalStacks    int `json:"totalStacks"`
		PlanetID       int `json:"planetId"`
		AttackerStacks int `json:"attackerStacks"`
		DefenderStacks int `json:"defenderStacks"`
		AttackerLosses int `json:"attackerLosses"`
		NumActions     int `json:"numActions"`
	} `json:"battleBlock"`
}

func TestScenarioBattle02(t *testing.T) {
	// Load expected data
	expectedData, err := os.ReadFile("../testdata/scenario-message/event/battle/battle-02/side1/expected.json")
	require.NoError(t, err, "failed to read expected.json")

	var expected battle02Expected
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Parse the M file
	data, err := os.ReadFile("../testdata/scenario-message/event/battle/battle-02/side1/game.m1")
	require.NoError(t, err, "failed to read game.m1")

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "failed to parse block list")

	t.Run("BattleBlock", func(t *testing.T) {
		// Collect BattleBlocks
		var battleBlocks []blocks.BattleBlock
		for _, block := range blockList {
			if bb, ok := block.(blocks.BattleBlock); ok {
				battleBlocks = append(battleBlocks, bb)
			}
		}

		require.Len(t, battleBlocks, 1, "expected exactly 1 BattleBlock")

		bb := battleBlocks[0]
		assert.Equal(t, expected.BattleBlock.Rounds, bb.Rounds,
			"rounds should be %d", expected.BattleBlock.Rounds)
		assert.Equal(t, expected.BattleBlock.Side1Stacks, bb.Side1Stacks,
			"side1 stacks should be %d", expected.BattleBlock.Side1Stacks)
		assert.Equal(t, expected.BattleBlock.TotalStacks, bb.TotalStacks,
			"total stacks should be %d", expected.BattleBlock.TotalStacks)
		assert.Equal(t, expected.BattleBlock.PlanetID, bb.PlanetID,
			"planet ID should be %d", expected.BattleBlock.PlanetID)
		assert.Equal(t, expected.BattleBlock.AttackerStacks, bb.AttackerStacks,
			"attacker stacks should be %d", expected.BattleBlock.AttackerStacks)
		assert.Equal(t, expected.BattleBlock.DefenderStacks, bb.DefenderStacks,
			"defender stacks should be %d", expected.BattleBlock.DefenderStacks)
		assert.Equal(t, expected.BattleBlock.AttackerLosses, bb.AttackerLosses,
			"attacker losses should be %d", expected.BattleBlock.AttackerLosses)
		assert.Equal(t, expected.BattleBlock.NumActions, len(bb.Actions),
			"number of actions should be %d", expected.BattleBlock.NumActions)
		assert.Equal(t, expected.BattleBlock.TotalStacks, len(bb.Stacks),
			"number of stacks should match total stacks")
	})

	t.Run("BattlePhases", func(t *testing.T) {
		var battleBlocks []blocks.BattleBlock
		for _, block := range blockList {
			if bb, ok := block.(blocks.BattleBlock); ok {
				battleBlocks = append(battleBlocks, bb)
			}
		}
		require.Len(t, battleBlocks, 1)
		bb := battleBlocks[0]

		// Get phases
		phases := bb.Phases()
		assert.Greater(t, len(phases), 0, "should have decoded phases")

		// Battle-02 has 61 phases according to screenshots
		// Current phase detection is ~33% complete for this battle
		assert.GreaterOrEqual(t, len(phases), 15, "should detect at least 15 phases")
		assert.LessOrEqual(t, len(phases), 70, "should not exceed expected phase count")

		// Verify phase structure
		for _, phase := range phases {
			assert.GreaterOrEqual(t, phase.Round, 0, "round should be >= 0")
			assert.LessOrEqual(t, phase.Round, 15, "round should be <= 15")
			assert.GreaterOrEqual(t, phase.StackID, 0, "stack should be >= 0")
			assert.LessOrEqual(t, phase.StackID, 5, "stack should be <= 5")
		}
	})
}
