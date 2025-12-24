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

type fleetScrappedSpaceExpected struct {
	Description   string `json:"description"`
	Year          int    `json:"year"`
	SalvageObject struct {
		X               int    `json:"x"`
		Y               int    `json:"y"`
		Ironium         int    `json:"ironium"`
		Boranium        int    `json:"boranium"`
		Germanium       int    `json:"germanium"`
		SourceFleetName string `json:"sourceFleetName"`
	} `json:"salvageObject"`
	Battle struct {
		PlanetID    int `json:"planetID"`
		EnemyPlayer int `json:"enemyPlayer"`
		YourForces  int `json:"yourForces"`
		EnemyForces int `json:"enemyForces"`
		YourLosses  int `json:"yourLosses"`
		EnemyLosses int `json:"enemyLosses"`
	} `json:"battle"`
}

func TestScenarioFleetScrappedInSpace(t *testing.T) {
	// Load expected data
	expectedData, err := os.ReadFile("../testdata/scenario-message/event/fleet-scrapped/expected.json")
	require.NoError(t, err, "failed to read expected.json")

	var expected fleetScrappedSpaceExpected
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Parse the M file
	data, err := os.ReadFile("../testdata/scenario-message/event/fleet-scrapped/game.m1")
	require.NoError(t, err, "failed to read game.m1")

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "failed to parse block list")

	t.Run("FleetScrappedInSpaceEvent", func(t *testing.T) {
		// Collect fleet scrapped in space events
		var scrappedEvents []blocks.FleetScrappedInSpaceEvent
		for _, block := range blockList {
			if eb, ok := block.(blocks.EventsBlock); ok {
				scrappedEvents = append(scrappedEvents, eb.FleetsScrappedInSpace...)
			}
		}

		require.Len(t, scrappedEvents, 1, "expected exactly 1 fleet scrapped in space event")

		// The event just tells us a salvage was created; the salvage object has the details
		event := scrappedEvents[0]
		assert.Greater(t, event.SalvageObjectID, 0, "salvage object ID should be set")
	})

	t.Run("SalvageObject", func(t *testing.T) {
		// Find the salvage object
		var salvageObjects []blocks.ObjectBlock
		for _, block := range blockList {
			if ob, ok := block.(blocks.ObjectBlock); ok && ob.IsSalvage() {
				salvageObjects = append(salvageObjects, ob)
			}
		}

		require.Len(t, salvageObjects, 1, "expected exactly 1 salvage object")

		salvage := salvageObjects[0]

		// Verify location
		assert.Equal(t, expected.SalvageObject.X, salvage.X,
			"salvage X coordinate should be %d", expected.SalvageObject.X)
		assert.Equal(t, expected.SalvageObject.Y, salvage.Y,
			"salvage Y coordinate should be %d", expected.SalvageObject.Y)

		// Verify minerals
		assert.Equal(t, expected.SalvageObject.Ironium, salvage.Ironium,
			"ironium should be %dkT", expected.SalvageObject.Ironium)
		assert.Equal(t, expected.SalvageObject.Boranium, salvage.Boranium,
			"boranium should be %dkT", expected.SalvageObject.Boranium)
		assert.Equal(t, expected.SalvageObject.Germanium, salvage.Germanium,
			"germanium should be %dkT", expected.SalvageObject.Germanium)

		// Verify source fleet - fleet index 3 displays as "Teamster #4"
		// (Fleet indices are 0-based internally, displayed as 1-based)
		assert.Equal(t, 3, salvage.SourceFleetID,
			"source fleet ID should be 3 (displayed as #4)")
	})

	t.Run("BattleEvent", func(t *testing.T) {
		// This scenario also contains a battle event
		var battleEvents []blocks.BattleEvent
		for _, block := range blockList {
			if eb, ok := block.(blocks.EventsBlock); ok {
				battleEvents = append(battleEvents, eb.Battles...)
			}
		}

		require.Len(t, battleEvents, 1, "expected exactly 1 battle event")

		event := battleEvents[0]
		assert.Equal(t, expected.Battle.PlanetID, event.PlanetID,
			"planet ID should be %d", expected.Battle.PlanetID)
		assert.Equal(t, expected.Battle.EnemyPlayer, event.EnemyPlayer,
			"enemy player should be %d", expected.Battle.EnemyPlayer)
		assert.Equal(t, expected.Battle.YourForces, event.YourForces,
			"your forces should be %d", expected.Battle.YourForces)
		assert.Equal(t, expected.Battle.EnemyForces, event.EnemyForces,
			"enemy forces should be %d", expected.Battle.EnemyForces)
		assert.Equal(t, expected.Battle.YourLosses, event.YourLosses,
			"your losses should be %d", expected.Battle.YourLosses)
		assert.Equal(t, expected.Battle.EnemyLosses, event.EnemyLosses,
			"enemy losses should be %d", expected.Battle.EnemyLosses)
	})
}
