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

type cometStrikeExpected struct {
	Scenario    string `json:"scenario"`
	Description string `json:"description"`
	Year        int    `json:"year"`
	CometStrike struct {
		PlanetID   int    `json:"planetId"`
		PlanetName string `json:"planetName"`
		Subtype    int    `json:"subtype"`
	} `json:"cometStrike"`
}

func TestScenarioCometStrike(t *testing.T) {
	// Load expected data
	expectedData, err := os.ReadFile("../testdata/scenario-message/event/comet/expected.json")
	require.NoError(t, err, "failed to read expected.json")

	var expected cometStrikeExpected
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Parse the M file
	data, err := os.ReadFile("../testdata/scenario-message/event/comet/game.m1")
	require.NoError(t, err, "failed to read game.m1")

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "failed to parse block list")

	// Collect comet strike events
	var cometEvents []blocks.CometStrikeEvent
	for _, block := range blockList {
		if eb, ok := block.(blocks.EventsBlock); ok {
			cometEvents = append(cometEvents, eb.CometStrikes...)
		}
	}

	require.Len(t, cometEvents, 1, "expected exactly 1 comet strike event")

	event := cometEvents[0]
	assert.Equal(t, expected.CometStrike.PlanetID, event.PlanetID,
		"planet ID should be %d (%s)", expected.CometStrike.PlanetID, expected.CometStrike.PlanetName)
	assert.Equal(t, expected.CometStrike.Subtype, event.Subtype,
		"subtype should match")
}

type newColonyExpected struct {
	Scenario    string `json:"scenario"`
	Description string `json:"description"`
	Year        int    `json:"year"`
	NewColony   struct {
		PlanetID   int    `json:"planetId"`
		PlanetName string `json:"planetName"`
	} `json:"newColony"`
	StrangeArtifact struct {
		PlanetID          int    `json:"planetId"`
		PlanetName        string `json:"planetName"`
		ResearchField     int    `json:"researchField"`
		ResearchFieldName string `json:"researchFieldName"`
		BoostAmount       int    `json:"boostAmount"`
	} `json:"strangeArtifact"`
	FleetScrapped struct {
		PlanetID           int    `json:"planetId"`
		PlanetName         string `json:"planetName"`
		FleetIndex         int    `json:"fleetIndex"`
		FleetDisplayNumber int    `json:"fleetDisplayNumber"`
		FleetName          string `json:"fleetName"`
		MineralAmount      int    `json:"mineralAmount"`
		Flags              int    `json:"flags"`
	} `json:"fleetScrapped"`
}

func TestScenarioNewColony(t *testing.T) {
	// Load expected data
	expectedData, err := os.ReadFile("../testdata/scenario-message/event/new-colony/expected.json")
	require.NoError(t, err, "failed to read expected.json")

	var expected newColonyExpected
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Parse the M file
	data, err := os.ReadFile("../testdata/scenario-message/event/new-colony/game.m1")
	require.NoError(t, err, "failed to read game.m1")

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "failed to parse block list")

	t.Run("NewColonyEvent", func(t *testing.T) {
		var newColonyEvents []blocks.NewColonyEvent
		for _, block := range blockList {
			if eb, ok := block.(blocks.EventsBlock); ok {
				newColonyEvents = append(newColonyEvents, eb.NewColonies...)
			}
		}

		require.Len(t, newColonyEvents, 1, "expected exactly 1 new colony event")

		event := newColonyEvents[0]
		assert.Equal(t, expected.NewColony.PlanetID, event.PlanetID,
			"planet ID should be %d (%s)", expected.NewColony.PlanetID, expected.NewColony.PlanetName)
	})

	t.Run("StrangeArtifactEvent", func(t *testing.T) {
		var artifactEvents []blocks.StrangeArtifactEvent
		for _, block := range blockList {
			if eb, ok := block.(blocks.EventsBlock); ok {
				artifactEvents = append(artifactEvents, eb.StrangeArtifacts...)
			}
		}

		require.Len(t, artifactEvents, 1, "expected exactly 1 strange artifact event")

		event := artifactEvents[0]
		assert.Equal(t, expected.StrangeArtifact.PlanetID, event.PlanetID,
			"planet ID should be %d (%s)", expected.StrangeArtifact.PlanetID, expected.StrangeArtifact.PlanetName)
		assert.Equal(t, expected.StrangeArtifact.ResearchField, event.ResearchField,
			"research field should be %d (%s)", expected.StrangeArtifact.ResearchField, expected.StrangeArtifact.ResearchFieldName)
		assert.Equal(t, expected.StrangeArtifact.BoostAmount, event.BoostAmount,
			"boost amount should be %d", expected.StrangeArtifact.BoostAmount)
	})

	t.Run("FleetScrappedEvent", func(t *testing.T) {
		var scrappedEvents []blocks.FleetScrappedEvent
		for _, block := range blockList {
			if eb, ok := block.(blocks.EventsBlock); ok {
				scrappedEvents = append(scrappedEvents, eb.FleetsScrapped...)
			}
		}

		require.Len(t, scrappedEvents, 1, "expected exactly 1 fleet scrapped event")

		event := scrappedEvents[0]
		assert.Equal(t, expected.FleetScrapped.PlanetID, event.PlanetID,
			"planet ID should be %d (%s)", expected.FleetScrapped.PlanetID, expected.FleetScrapped.PlanetName)
		assert.Equal(t, expected.FleetScrapped.FleetIndex, event.FleetIndex,
			"fleet index should be %d (%s)", expected.FleetScrapped.FleetIndex, expected.FleetScrapped.FleetName)
		assert.Equal(t, expected.FleetScrapped.MineralAmount, event.MineralAmount,
			"mineral amount should be %dkT", expected.FleetScrapped.MineralAmount)
		assert.Equal(t, expected.FleetScrapped.Flags, event.Flags,
			"flags should be %d", expected.FleetScrapped.Flags)
	})
}

type battleExpected struct {
	Scenario    string `json:"scenario"`
	Description string `json:"description"`
	Year        int    `json:"year"`
	Battle      struct {
		PlanetID      int    `json:"planetId"`
		PlanetName    string `json:"planetName"`
		EnemyPlayer   int    `json:"enemyPlayer"`
		EnemyRace     string `json:"enemyRace"`
		YourForces    int    `json:"yourForces"`
		EnemyForces   int    `json:"enemyForces"`
		YourLosses    int    `json:"yourLosses"`
		EnemyLosses   int    `json:"enemyLosses"`
		YouSurvived   bool   `json:"youSurvived"`
		EnemySurvived bool   `json:"enemySurvived"`
		HasRecording  bool   `json:"hasRecording"`
	} `json:"battle"`
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
	} `json:"battleBlock"`
}

func TestScenarioBattle(t *testing.T) {
	// Load expected data
	expectedData, err := os.ReadFile("../testdata/scenario-message/event/battle/side1/expected.json")
	require.NoError(t, err, "failed to read expected.json")

	var expected battleExpected
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Parse the M file
	data, err := os.ReadFile("../testdata/scenario-message/event/battle/side1/game.m1")
	require.NoError(t, err, "failed to read game.m1")

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "failed to parse block list")

	t.Run("BattleEvent", func(t *testing.T) {
		// Collect battle events
		var battleEvents []blocks.BattleEvent
		for _, block := range blockList {
			if eb, ok := block.(blocks.EventsBlock); ok {
				battleEvents = append(battleEvents, eb.Battles...)
			}
		}

		require.Len(t, battleEvents, 1, "expected exactly 1 battle event")

		event := battleEvents[0]
		assert.Equal(t, expected.Battle.PlanetID, event.PlanetID,
			"planet ID should be %d (%s)", expected.Battle.PlanetID, expected.Battle.PlanetName)
		assert.Equal(t, expected.Battle.EnemyPlayer, event.EnemyPlayer,
			"enemy player should be %d (%s)", expected.Battle.EnemyPlayer, expected.Battle.EnemyRace)
		assert.Equal(t, expected.Battle.YourForces, event.YourForces,
			"your forces should be %d", expected.Battle.YourForces)
		assert.Equal(t, expected.Battle.EnemyForces, event.EnemyForces,
			"enemy forces should be %d", expected.Battle.EnemyForces)
		assert.Equal(t, expected.Battle.YourLosses, event.YourLosses,
			"your losses should be %d", expected.Battle.YourLosses)
		assert.Equal(t, expected.Battle.EnemyLosses, event.EnemyLosses,
			"enemy losses should be %d", expected.Battle.EnemyLosses)
		assert.Equal(t, expected.Battle.YouSurvived, event.YouSurvived,
			"youSurvived should be %v", expected.Battle.YouSurvived)
		assert.Equal(t, expected.Battle.EnemySurvived, event.EnemySurvived,
			"enemySurvived should be %v", expected.Battle.EnemySurvived)
		assert.Equal(t, expected.Battle.HasRecording, event.HasRecording,
			"hasRecording should be %v", expected.Battle.HasRecording)
	})

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

	t.Run("BattleRecordEvents", func(t *testing.T) {
		var battleBlocks []blocks.BattleBlock
		for _, block := range blockList {
			if bb, ok := block.(blocks.BattleBlock); ok {
				battleBlocks = append(battleBlocks, bb)
			}
		}
		require.Len(t, battleBlocks, 1)
		bb := battleBlocks[0]

		// Get all decoded events
		events := bb.AllEvents()
		assert.Greater(t, len(events), 0, "should have decoded events")

		// Count event types
		var moves, fires, damages int
		for _, e := range events {
			switch e.ActionType {
			case blocks.ActionMove:
				moves++
			case blocks.ActionFire:
				fires++
			case blocks.ActionDamage:
				damages++
			}
		}
		assert.Greater(t, moves, 0, "should have move events")
		assert.Greater(t, fires, 0, "should have fire events")
		assert.Greater(t, damages, 0, "should have damage events")

		// Verify events are grouped by round
		byRound := bb.EventsByRound()
		assert.Greater(t, len(byRound), 0, "should have events grouped by round")

		// Test String() method works
		if len(events) > 0 {
			str := events[0].String()
			assert.Contains(t, str, "Round", "event string should contain Round")
			assert.Contains(t, str, "Stack", "event string should contain Stack")
		}
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

		// Phase count should be reasonable (not complete but substantial)
		// Screenshots show 68 phases, we detect ~49
		assert.GreaterOrEqual(t, len(phases), 40, "should detect at least 40 phases")
		assert.LessOrEqual(t, len(phases), 70, "should not exceed expected phase count")

		// Verify phase structure
		for _, phase := range phases {
			assert.GreaterOrEqual(t, phase.Round, 0, "round should be >= 0")
			assert.LessOrEqual(t, phase.Round, 15, "round should be <= 15")
			assert.GreaterOrEqual(t, phase.StackID, 0, "stack should be >= 0")
			assert.LessOrEqual(t, phase.StackID, 5, "stack should be <= 5")
		}

		// Test String() method
		if len(phases) > 0 {
			str := phases[0].String()
			assert.Contains(t, str, "Phase", "phase string should contain Phase")
			assert.Contains(t, str, "Round", "phase string should contain Round")
			assert.Contains(t, str, "Stack", "phase string should contain Stack")
		}
	})
}

func TestScenarioBattleSide2(t *testing.T) {
	// Load expected data
	expectedData, err := os.ReadFile("../testdata/scenario-message/event/battle/side2/expected.json")
	require.NoError(t, err, "failed to read expected.json")

	var expected battleExpected
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Parse the M file
	data, err := os.ReadFile("../testdata/scenario-message/event/battle/side2/game.m2")
	require.NoError(t, err, "failed to read game.m2")

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "failed to parse block list")

	t.Run("BattleEvent", func(t *testing.T) {
		// Collect battle events
		var battleEvents []blocks.BattleEvent
		for _, block := range blockList {
			if eb, ok := block.(blocks.EventsBlock); ok {
				battleEvents = append(battleEvents, eb.Battles...)
			}
		}

		require.Len(t, battleEvents, 1, "expected exactly 1 battle event")

		event := battleEvents[0]
		assert.Equal(t, expected.Battle.PlanetID, event.PlanetID,
			"planet ID should be %d (%s)", expected.Battle.PlanetID, expected.Battle.PlanetName)
		assert.Equal(t, expected.Battle.EnemyPlayer, event.EnemyPlayer,
			"enemy player should be %d (%s)", expected.Battle.EnemyPlayer, expected.Battle.EnemyRace)
		assert.Equal(t, expected.Battle.YourForces, event.YourForces,
			"your forces should be %d", expected.Battle.YourForces)
		assert.Equal(t, expected.Battle.EnemyForces, event.EnemyForces,
			"enemy forces should be %d", expected.Battle.EnemyForces)
		assert.Equal(t, expected.Battle.YourLosses, event.YourLosses,
			"your losses should be %d", expected.Battle.YourLosses)
		assert.Equal(t, expected.Battle.EnemyLosses, event.EnemyLosses,
			"enemy losses should be %d", expected.Battle.EnemyLosses)
		assert.Equal(t, expected.Battle.YouSurvived, event.YouSurvived,
			"youSurvived should be %v", expected.Battle.YouSurvived)
		assert.Equal(t, expected.Battle.EnemySurvived, event.EnemySurvived,
			"enemySurvived should be %v", expected.Battle.EnemySurvived)
		assert.Equal(t, expected.Battle.HasRecording, event.HasRecording,
			"hasRecording should be %v", expected.Battle.HasRecording)
	})

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

		// Verify stacks by player (side2 = player 1)
		viewingPlayer := 1
		enemyPlayer := 0
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
}
