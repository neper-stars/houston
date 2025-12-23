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

type starbaseConstructionExpected struct {
	Scenario     string `json:"scenario"`
	Description  string `json:"description"`
	Year         int    `json:"year"`
	StarbaseBuilt struct {
		PlanetID   int    `json:"planetId"`
		PlanetName string `json:"planetName"`
		DesignInfo int    `json:"designInfo"`
	} `json:"starbaseBuilt"`
}

func TestScenarioStarbaseConstruction(t *testing.T) {
	// Load expected data
	expectedData, err := os.ReadFile("../testdata/scenario-message/construction/starbase/expected.json")
	require.NoError(t, err, "failed to read expected.json")

	var expected starbaseConstructionExpected
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Parse the M file
	data, err := os.ReadFile("../testdata/scenario-message/construction/starbase/game.m1")
	require.NoError(t, err, "failed to read game.m1")

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "failed to parse block list")

	// Collect starbase built events
	var starbaseEvents []blocks.StarbaseBuiltEvent
	for _, block := range blockList {
		if eb, ok := block.(blocks.EventsBlock); ok {
			starbaseEvents = append(starbaseEvents, eb.StarbasesBuilt...)
		}
	}

	require.Len(t, starbaseEvents, 1, "expected exactly 1 starbase built event")

	event := starbaseEvents[0]
	assert.Equal(t, expected.StarbaseBuilt.PlanetID, event.PlanetID,
		"planet ID should be %d (%s)", expected.StarbaseBuilt.PlanetID, expected.StarbaseBuilt.PlanetName)
	assert.Equal(t, expected.StarbaseBuilt.DesignInfo, event.DesignInfo,
		"design info should match")
}
