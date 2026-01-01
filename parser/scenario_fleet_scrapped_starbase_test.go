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

type fleetScrappedAtStarbaseExpected struct {
	Description             string `json:"description"`
	Year                    int    `json:"year"`
	FleetScrappedAtStarbase struct {
		PlanetID             int    `json:"planetID"`
		PlanetName           string `json:"planetName"`
		FleetIndex           int    `json:"fleetIndex"`
		FleetDisplayNumber   int    `json:"fleetDisplayNumber"`
		MineralAmountEncoded int    `json:"mineralAmountEncoded"`
		FleetMass            int    `json:"fleetMass"`
		MineralsRecovered    int    `json:"mineralsRecovered"`
		RecoveryRate         string `json:"recoveryRate"`
		Flags                int    `json:"flags"`
		Note                 string `json:"note"`
	} `json:"fleetScrappedAtStarbase"`
}

func TestScenarioFleetScrappedAtStarbase(t *testing.T) {
	// Load expected data
	expectedData, err := os.ReadFile("../testdata/scenario-message/event/fleet-dismantled-at-planet/expected.json")
	require.NoError(t, err, "failed to read expected.json")

	var expected fleetScrappedAtStarbaseExpected
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Parse the M file
	data, err := os.ReadFile("../testdata/scenario-message/event/fleet-dismantled-at-planet/game.m1")
	require.NoError(t, err, "failed to read game.m1")

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "failed to parse block list")

	// Find FleetScrappedAtStarbase events
	var scrappedEvents []blocks.FleetScrappedAtStarbaseEvent
	for _, block := range blockList {
		if eb, ok := block.(blocks.EventsBlock); ok {
			scrappedEvents = append(scrappedEvents, eb.FleetsScrappedAtStarbase...)
		}
	}

	require.Len(t, scrappedEvents, 1, "expected exactly 1 fleet scrapped at starbase event")

	event := scrappedEvents[0]

	assert.Equal(t, expected.FleetScrappedAtStarbase.PlanetID, event.PlanetID,
		"planet ID should be %d (%s)", expected.FleetScrappedAtStarbase.PlanetID, expected.FleetScrappedAtStarbase.PlanetName)
	assert.Equal(t, expected.FleetScrappedAtStarbase.FleetIndex, event.FleetIndex,
		"fleet index should be %d (displays as #%d)", expected.FleetScrappedAtStarbase.FleetIndex, expected.FleetScrappedAtStarbase.FleetDisplayNumber)
	assert.Equal(t, expected.FleetScrappedAtStarbase.FleetMass, event.FleetMass,
		"fleet mass should be %dkT (encoded as %d, minerals recovered: %dkT at %s)",
		expected.FleetScrappedAtStarbase.FleetMass, expected.FleetScrappedAtStarbase.MineralAmountEncoded,
		expected.FleetScrappedAtStarbase.MineralsRecovered, expected.FleetScrappedAtStarbase.RecoveryRate)
	assert.Equal(t, expected.FleetScrappedAtStarbase.Flags, event.Flags,
		"flags should be %d", expected.FleetScrappedAtStarbase.Flags)
}
