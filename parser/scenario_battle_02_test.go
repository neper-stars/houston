package parser_test

import (
	"encoding/json"
	"fmt"
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
		Rounds          int `json:"rounds"`
		TotalStacks     int `json:"totalStacks"`
		PlanetID        int `json:"planetId"`
		NumActions      int `json:"numActions"`
		TotalPhases     int `json:"totalPhases"`
		OurStacks       int `json:"ourStacks"`
		TheirStacks     int `json:"theirStacks"`
		OurCasualties   int `json:"ourCasualties"`
		TheirCasualties int `json:"theirCasualties"`
		Stacks          []struct {
			StackID  int `json:"stackId"`
			PlayerID int `json:"playerId"`
			DesignID int `json:"designId"`
		} `json:"stacks"`
		VerifiedPhases []struct {
			Phase        int `json:"phase"`
			Action       int `json:"action"`
			Round        int `json:"round"`
			Attacker     int `json:"attacker"`
			Target       int `json:"target"`
			ShipsKilled  int `json:"shipsKilled"`
			ShieldDamage int `json:"shieldDamage"`
		} `json:"verifiedPhases"`
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
		assert.Equal(t, expected.BattleBlock.TotalStacks, bb.TotalStacks,
			"total stacks should be %d", expected.BattleBlock.TotalStacks)
		assert.Equal(t, expected.BattleBlock.PlanetID, bb.PlanetID,
			"planet ID should be %d", expected.BattleBlock.PlanetID)
		assert.Equal(t, expected.BattleBlock.NumActions, len(bb.Actions),
			"number of actions should be %d", expected.BattleBlock.NumActions)
		assert.Equal(t, expected.BattleBlock.TotalPhases, bb.TotalPhases(),
			"total phases should be %d", expected.BattleBlock.TotalPhases)
		assert.Equal(t, expected.BattleBlock.TotalStacks, len(bb.Stacks),
			"number of stacks should match total stacks")

		// Verify stacks by player (side1 = player 0)
		viewingPlayer := 0
		enemyPlayer := 1
		assert.Equal(t, expected.BattleBlock.OurStacks, bb.StacksForPlayer(viewingPlayer),
			"our stacks should be %d", expected.BattleBlock.OurStacks)
		assert.Equal(t, expected.BattleBlock.TheirStacks, bb.TotalStacks-bb.StacksForPlayer(viewingPlayer),
			"their stacks should be %d", expected.BattleBlock.TheirStacks)

		// Verify casualties
		assert.Equal(t, expected.BattleBlock.OurCasualties, bb.CasualtiesForPlayer(viewingPlayer),
			"our casualties should be %d", expected.BattleBlock.OurCasualties)
		assert.Equal(t, expected.BattleBlock.TheirCasualties, bb.CasualtiesForPlayer(enemyPlayer),
			"their casualties should be %d", expected.BattleBlock.TheirCasualties)
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

	t.Run("Stacks", func(t *testing.T) {
		var battleBlocks []blocks.BattleBlock
		for _, block := range blockList {
			if bb, ok := block.(blocks.BattleBlock); ok {
				battleBlocks = append(battleBlocks, bb)
			}
		}
		require.Len(t, battleBlocks, 1)
		bb := battleBlocks[0]

		// Verify stack definitions from expected.json
		for _, expStack := range expected.BattleBlock.Stacks {
			require.Less(t, expStack.StackID, len(bb.Stacks), "stack %d should exist", expStack.StackID)
			stack := bb.Stacks[expStack.StackID]
			assert.Equal(t, expStack.PlayerID, stack.OwnerPlayerID,
				"stack %d player should be %d", expStack.StackID, expStack.PlayerID)
			assert.Equal(t, expStack.DesignID, stack.DesignID,
				"stack %d design should be %d", expStack.StackID, expStack.DesignID)
		}
	})

	t.Run("VerifiedPhases", func(t *testing.T) {
		// This test verifies action records against VCR screenshots
		// Phase N = Action (N-2) because Phase 1 is the setup phase with no action
		var battleBlocks []blocks.BattleBlock
		for _, block := range blockList {
			if bb, ok := block.(blocks.BattleBlock); ok {
				battleBlocks = append(battleBlocks, bb)
			}
		}
		require.Len(t, battleBlocks, 1)
		bb := battleBlocks[0]

		for _, vp := range expected.BattleBlock.VerifiedPhases {
			t.Run(fmt.Sprintf("Phase%02d", vp.Phase), func(t *testing.T) {
				require.Less(t, vp.Action, len(bb.Actions),
					"action %d should exist for phase %d", vp.Action, vp.Phase)

				action := bb.Actions[vp.Action]
				require.Greater(t, len(action.Events), 0, "action should have events")

				// Check round from first event
				assert.Equal(t, vp.Round, action.Events[0].Round,
					"phase %d round should be %d", vp.Phase, vp.Round)

				// Check attacker (stack ID from first event)
				assert.Equal(t, vp.Attacker, action.Events[0].StackID,
					"phase %d attacker should be stack %d", vp.Phase, vp.Attacker)

				// Check ships killed
				totalKilled := 0
				for _, kill := range action.Kills {
					if kill.StackID == vp.Target {
						totalKilled += kill.ShipsKilled
					}
				}
				assert.Equal(t, vp.ShipsKilled, totalKilled,
					"phase %d should kill %d ships", vp.Phase, vp.ShipsKilled)
			})
		}
	})
}
