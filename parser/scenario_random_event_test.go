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
	Scenario        string `json:"scenario"`
	Description     string `json:"description"`
	Year            int    `json:"year"`
	NewColony       struct {
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
