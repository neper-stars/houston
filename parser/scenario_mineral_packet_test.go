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

type mineralPacketExpected struct {
	Scenario      string `json:"scenario"`
	Description   string `json:"description"`
	Year          int    `json:"year"`
	TotalMessages int    `json:"totalMessages"`
	PacketEvent   struct {
		Type                  string `json:"type"`
		SourcePlanetName      string `json:"sourcePlanetName"`
		SourcePlanetID        int    `json:"sourcePlanetId"`
		DestinationPlanetName string `json:"destinationPlanetName"`
		DestinationPlanetID   int    `json:"destinationPlanetId"`
	} `json:"packetEvent"`
	PacketObject struct {
		X                   int `json:"x"`
		Y                   int `json:"y"`
		DestinationPlanetID int `json:"destinationPlanetId"`
		Ironium             int `json:"ironium"`
		Boranium            int `json:"boranium"`
		Germanium           int `json:"germanium"`
		TotalMinerals       int `json:"totalMinerals"`
		WarpSpeed           int `json:"warpSpeed"`
	} `json:"packetObject"`
}

func TestScenarioMineralPacket(t *testing.T) {
	// Load expected data
	expectedData, err := os.ReadFile("../testdata/scenario-mineral-packet/expected.json")
	require.NoError(t, err, "failed to read expected.json")

	var expected mineralPacketExpected
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Parse the M file
	data, err := os.ReadFile("../testdata/scenario-mineral-packet/game.m1")
	require.NoError(t, err, "failed to read game.m1")

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "failed to parse block list")

	t.Run("PacketProducedEvent", func(t *testing.T) {
		var packetEvents []blocks.MineralPacketProducedEvent

		for _, block := range blockList {
			if eb, ok := block.(blocks.EventsBlock); ok {
				packetEvents = append(packetEvents, eb.PacketsProduced...)
			}
		}

		require.Len(t, packetEvents, 1, "expected exactly 1 packet produced event")

		event := packetEvents[0]
		assert.Equal(t, expected.PacketEvent.SourcePlanetID, event.SourcePlanetID,
			"source planet ID should be %d (%s)", expected.PacketEvent.SourcePlanetID, expected.PacketEvent.SourcePlanetName)
		assert.Equal(t, expected.PacketEvent.DestinationPlanetID, event.DestinationPlanetID,
			"destination planet ID should be %d (%s)", expected.PacketEvent.DestinationPlanetID, expected.PacketEvent.DestinationPlanetName)
	})

	t.Run("PacketObject", func(t *testing.T) {
		var packetObject *blocks.ObjectBlock

		for _, block := range blockList {
			if ob, ok := block.(blocks.ObjectBlock); ok {
				if ob.IsPacket() {
					packetObject = &ob
					break
				}
			}
		}

		require.NotNil(t, packetObject, "expected to find a mineral packet object")

		assert.Equal(t, expected.PacketObject.X, packetObject.X, "packet X position")
		assert.Equal(t, expected.PacketObject.Y, packetObject.Y, "packet Y position")
		assert.Equal(t, expected.PacketObject.DestinationPlanetID, packetObject.DestinationPlanetID,
			"destination planet ID")
		assert.Equal(t, expected.PacketObject.Ironium, packetObject.Ironium, "ironium amount")
		assert.Equal(t, expected.PacketObject.Boranium, packetObject.Boranium, "boranium amount")
		assert.Equal(t, expected.PacketObject.Germanium, packetObject.Germanium, "germanium amount")
		assert.Equal(t, expected.PacketObject.TotalMinerals, packetObject.TotalMinerals(), "total minerals")

		// Warp speed encoding: raw byte is 0xCC (204), expected warp 7
		// The encoding is not fully understood - byte 7 contains speed info but
		// the exact formula is unclear. Java starsapi also has this as TODO.
		// For now, test that we can access the raw value.
		// TODO: implement proper warp speed decoding once encoding is understood
		assert.NotZero(t, packetObject.PacketSpeed, "packet speed byte should be populated")
	})
}
