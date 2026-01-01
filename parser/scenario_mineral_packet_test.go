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

		assert.Equal(t, expected.PacketObject.WarpSpeed, packetObject.WarpSpeed(), "warp speed")
	})
}

type mineralPacketBatch2Expected struct {
	Scenario    string `json:"scenario"`
	Description string `json:"description"`
	Year        int    `json:"year"`
	Packets     []struct {
		X                     int    `json:"x"`
		Y                     int    `json:"y"`
		DestinationPlanetID   int    `json:"destinationPlanetId"`
		DestinationPlanetName string `json:"destinationPlanetName"`
		Ironium               int    `json:"ironium"`
		Boranium              int    `json:"boranium"`
		Germanium             int    `json:"germanium"`
		TotalMinerals         int    `json:"totalMinerals"`
		WarpSpeed             int    `json:"warpSpeed"`
	} `json:"packets"`
}

func TestScenarioMineralPacketBatch2(t *testing.T) {
	// Load expected data
	expectedData, err := os.ReadFile("../testdata/scenario-mineral-packet/batch2/expected.json")
	require.NoError(t, err, "failed to read expected.json")

	var expected mineralPacketBatch2Expected
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Parse the M file
	data, err := os.ReadFile("../testdata/scenario-mineral-packet/batch2/game.m1")
	require.NoError(t, err, "failed to read game.m1")

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "failed to parse block list")

	// Collect all packet objects
	var packets []blocks.ObjectBlock
	for _, block := range blockList {
		if ob, ok := block.(blocks.ObjectBlock); ok {
			if ob.IsPacket() {
				packets = append(packets, ob)
			}
		}
	}

	require.Len(t, packets, len(expected.Packets), "expected %d packets", len(expected.Packets))

	// Create a map of packets by position for easier lookup
	packetsByPos := make(map[string]blocks.ObjectBlock)
	for _, p := range packets {
		key := fmt.Sprintf("%d,%d", p.X, p.Y)
		packetsByPos[key] = p
	}

	for _, exp := range expected.Packets {
		key := fmt.Sprintf("%d,%d", exp.X, exp.Y)
		t.Run(fmt.Sprintf("Packet_%s_Warp%d", key, exp.WarpSpeed), func(t *testing.T) {
			packet, found := packetsByPos[key]
			require.True(t, found, "packet at position %s not found", key)

			assert.Equal(t, exp.X, packet.X, "X position")
			assert.Equal(t, exp.Y, packet.Y, "Y position")
			assert.Equal(t, exp.DestinationPlanetID, packet.DestinationPlanetID, "destination planet ID")
			assert.Equal(t, exp.Ironium, packet.Ironium, "ironium")
			assert.Equal(t, exp.Boranium, packet.Boranium, "boranium")
			assert.Equal(t, exp.Germanium, packet.Germanium, "germanium")
			assert.Equal(t, exp.TotalMinerals, packet.TotalMinerals(), "total minerals")
			assert.Equal(t, exp.WarpSpeed, packet.WarpSpeed(), "warp speed")
		})
	}
}

type bombardmentExpected struct {
	Scenario     string `json:"scenario"`
	Description  string `json:"description"`
	Year         int    `json:"year"`
	Bombardments []struct {
		PlanetID        int    `json:"planetId"`
		PlanetName      string `json:"planetName"`
		MineralAmount   int    `json:"mineralAmount"`
		ColonistsKilled int    `json:"colonistsKilled"`
	} `json:"bombardments"`
}

func TestScenarioMineralPacketBombardment(t *testing.T) {
	// Load expected data
	expectedData, err := os.ReadFile("../testdata/scenario-mineral-packet/bombardment-message/expected.json")
	require.NoError(t, err, "failed to read expected.json")

	var expected bombardmentExpected
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Parse the M file
	data, err := os.ReadFile("../testdata/scenario-mineral-packet/bombardment-message/game.m1")
	require.NoError(t, err, "failed to read game.m1")

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "failed to parse block list")

	// Collect bombardment events
	var bombardments []blocks.PacketBombardmentEvent
	for _, block := range blockList {
		if eb, ok := block.(blocks.EventsBlock); ok {
			bombardments = append(bombardments, eb.PacketBombardments...)
		}
	}

	require.Len(t, bombardments, len(expected.Bombardments), "expected %d bombardment events", len(expected.Bombardments))

	for i, exp := range expected.Bombardments {
		t.Run(fmt.Sprintf("Bombardment_%d_%dkT", i+1, exp.MineralAmount), func(t *testing.T) {
			event := bombardments[i]
			assert.Equal(t, exp.PlanetID, event.PlanetID, "planet ID")
			assert.Equal(t, exp.MineralAmount, event.MineralAmount, "mineral amount")
			assert.Equal(t, exp.ColonistsKilled, event.ColonistsKilled, "colonists killed")
		})
	}
}
