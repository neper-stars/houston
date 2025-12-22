package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neper-stars/houston/blocks"
)

// Expected data structures for mystery trader scenario
type ExpectedMysteryTrader struct {
	Number int    `json:"number"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	DestX  int    `json:"destX"`
	DestY  int    `json:"destY"`
	Warp   int    `json:"warp"`
	Item   string `json:"item"`
}

type ExpectedMysteryTraderData struct {
	Scenario       string                  `json:"scenario"`
	MysteryTraders []ExpectedMysteryTrader `json:"mysteryTraders"`
}

func loadMysteryTraderExpected(t *testing.T) *ExpectedMysteryTraderData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-mysterytrader", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedMysteryTraderData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadMysteryTraderFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-mysterytrader", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

func traderItemName(itemBits uint16) string {
	switch {
	case itemBits == 0:
		return "Research"
	case (itemBits & blocks.TraderItemMultiCargoPod) != 0:
		return "Multi Cargo Pod"
	case (itemBits & blocks.TraderItemMultiFunctionPod) != 0:
		return "Multi Function Pod"
	case (itemBits & blocks.TraderItemLangstonShield) != 0:
		return "Langston Shield"
	case (itemBits & blocks.TraderItemMegaPolyShell) != 0:
		return "Mega Poly Shell"
	case (itemBits & blocks.TraderItemAlienMiner) != 0:
		return "Alien Miner"
	case (itemBits & blocks.TraderItemHushABoom) != 0:
		return "Hush-a-Boom"
	case (itemBits & blocks.TraderItemAntiMatterTorpedo) != 0:
		return "Anti Matter Torpedo"
	case (itemBits & blocks.TraderItemMultiContainedMunition) != 0:
		return "Multi Contained Munition"
	case (itemBits & blocks.TraderItemMiniMorph) != 0:
		return "Mini Morph"
	case (itemBits & blocks.TraderItemEnigmaPulsar) != 0:
		return "Enigma Pulsar"
	case (itemBits & blocks.TraderItemGenesisDevice) != 0:
		return "Genesis Device"
	case (itemBits & blocks.TraderItemJumpGate) != 0:
		return "Jump Gate"
	case (itemBits & blocks.TraderItemShip) != 0:
		return "Ship"
	default:
		return "Unknown"
	}
}

func TestScenarioMysteryTrader_ObjectBlocks(t *testing.T) {
	expected := loadMysteryTraderExpected(t)
	_, blockList := loadMysteryTraderFile(t, "game.m1")

	// Collect mystery traders
	var traders []blocks.ObjectBlock
	for _, block := range blockList {
		if ob, ok := block.(blocks.ObjectBlock); ok {
			if !ob.IsCountObject && ob.IsMysteryTrader() {
				traders = append(traders, ob)
			}
		}
	}

	require.Equal(t, len(expected.MysteryTraders), len(traders),
		"Should have %d mystery traders", len(expected.MysteryTraders))

	// Validate each mystery trader
	for i, exp := range expected.MysteryTraders {
		mt := traders[i]
		t.Run(exp.Item, func(t *testing.T) {
			// Validate number
			assert.Equal(t, exp.Number, mt.Number, "Number should match")

			// Validate position
			assert.Equal(t, exp.X, mt.X, "X should match")
			assert.Equal(t, exp.Y, mt.Y, "Y should match")

			// Validate destination
			assert.Equal(t, exp.DestX, mt.XDest, "DestX should match")
			assert.Equal(t, exp.DestY, mt.YDest, "DestY should match")

			// Validate warp
			assert.Equal(t, exp.Warp, mt.Warp, "Warp should match")

			// Validate item
			actualItem := traderItemName(mt.ItemBits)
			assert.Equal(t, exp.Item, actualItem, "Item should match")
		})
	}
}
