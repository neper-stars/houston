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

// Expected data structures for waypoints.json
type ExpectedWaypoint struct {
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Task string `json:"task"`
}

type ExpectedFleetWaypoints struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Waypoints []ExpectedWaypoint `json:"waypoints"`
}

type ExpectedWaypointsData struct {
	Scenario string                   `json:"scenario"`
	Files    []string                 `json:"files"`
	Fleets   []ExpectedFleetWaypoints `json:"fleets"`
}

// Expected data structures for production.json
type ExpectedQueueItem struct {
	Type  string `json:"type"`
	Name  string `json:"name,omitempty"`
	Count int    `json:"count"`
}

type ExpectedProductionData struct {
	Scenario string `json:"scenario"`
	Planet   struct {
		ID    string              `json:"id"`
		Name  string              `json:"name"`
		Queue []ExpectedQueueItem `json:"queue"`
	} `json:"planet"`
}

// OrdersScenarioHelper provides utilities for loading X file test data
type OrdersScenarioHelper struct {
	t   *testing.T
	dir string
}

// NewOrdersScenarioHelper creates a helper for the scenario-orders directory
func NewOrdersScenarioHelper(t *testing.T) *OrdersScenarioHelper {
	t.Helper()
	dir := filepath.Join("..", "testdata", "scenario-orders")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skipf("Test files not found in testdata/scenario-orders/")
	}
	return &OrdersScenarioHelper{t: t, dir: dir}
}

// LoadWaypointsExpected loads the waypoints.json expected data
func (h *OrdersScenarioHelper) LoadWaypointsExpected() *ExpectedWaypointsData {
	h.t.Helper()
	path := filepath.Join(h.dir, "waypoints.json")
	data, err := os.ReadFile(path)
	require.NoError(h.t, err, "Failed to read waypoints.json")

	var expected ExpectedWaypointsData
	err = json.Unmarshal(data, &expected)
	require.NoError(h.t, err, "Failed to parse waypoints.json")
	return &expected
}

// LoadProductionExpected loads the production.json expected data
func (h *OrdersScenarioHelper) LoadProductionExpected() *ExpectedProductionData {
	h.t.Helper()
	path := filepath.Join(h.dir, "production.json")
	data, err := os.ReadFile(path)
	require.NoError(h.t, err, "Failed to read production.json")

	var expected ExpectedProductionData
	err = json.Unmarshal(data, &expected)
	require.NoError(h.t, err, "Failed to parse production.json")
	return &expected
}

// LoadXFile loads and parses an X file from the scenario directory
func (h *OrdersScenarioHelper) LoadXFile(filename string) (FileData, []blocks.Block) {
	h.t.Helper()
	path := filepath.Join(h.dir, filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(h.t, err, "Failed to read %s", filename)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(h.t, err, "BlockList() failed for %s", filename)

	return fd, blockList
}

// LoadDesignsFromBasicScenario loads ship design names from scenario-basic/game.m1
// This is needed to correlate design IDs in production queues with design names
func (h *OrdersScenarioHelper) LoadDesignsFromBasicScenario() map[int]string {
	h.t.Helper()
	path := filepath.Join("..", "testdata", "scenario-basic", "game.m1")
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		h.t.Skipf("scenario-basic/game.m1 not found, skipping design name validation")
		return nil
	}

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(h.t, err, "Failed to parse game.m1")

	designs := make(map[int]string)
	for _, block := range blockList {
		if d, ok := block.(blocks.DesignBlock); ok && !d.IsStarbase {
			designs[d.DesignNumber] = d.Name
		}
	}
	return designs
}

func TestScenarioOrders_WaypointAdd(t *testing.T) {
	h := NewOrdersScenarioHelper(t)
	expected := h.LoadWaypointsExpected()
	_, blockList := h.LoadXFile("game.x1")

	// Find WaypointAddBlock and WaypointChangeTaskBlock
	var waypointAddBlocks []blocks.WaypointAddBlock
	var waypointChangeBlocks []blocks.WaypointChangeTaskBlock

	for _, block := range blockList {
		switch b := block.(type) {
		case blocks.WaypointAddBlock:
			waypointAddBlocks = append(waypointAddBlocks, b)
		case blocks.WaypointChangeTaskBlock:
			waypointChangeBlocks = append(waypointChangeBlocks, b)
		}
	}

	// We expect at least one waypoint add block
	require.GreaterOrEqual(t, len(waypointAddBlocks), 1, "Should have at least 1 WaypointAddBlock")

	// Validate the first waypoint add block coordinates match expected
	// Note: Fleet IDs in file are 0-indexed (Fleet 2 = game's Fleet #3 "Santa Maria")
	firstFleet := expected.Fleets[0]
	expectedX := firstFleet.Waypoints[0].X
	expectedY := firstFleet.Waypoints[0].Y

	assert.Equal(t, expectedX, waypointAddBlocks[0].X, "Waypoint X should match")
	assert.Equal(t, expectedY, waypointAddBlocks[0].Y, "Waypoint Y should match")

	// Check that a colonize task was set (WaypointTaskColonize = 2)
	foundColonizeTask := false
	for _, b := range waypointChangeBlocks {
		if b.WaypointTask == blocks.WaypointTaskColonize {
			foundColonizeTask = true
			assert.Equal(t, expectedX, b.X, "Colonize waypoint X should match")
			assert.Equal(t, expectedY, b.Y, "Colonize waypoint Y should match")
		}
	}
	assert.True(t, foundColonizeTask, "Should have a colonize task")
}

func TestScenarioOrders_ProductionQueue(t *testing.T) {
	h := NewOrdersScenarioHelper(t)
	expected := h.LoadProductionExpected()
	designs := h.LoadDesignsFromBasicScenario()
	_, blockList := h.LoadXFile("game.x1")

	// Find ProductionQueueChangeBlock
	var prodBlocks []blocks.ProductionQueueChangeBlock
	for _, block := range blockList {
		if b, ok := block.(blocks.ProductionQueueChangeBlock); ok {
			prodBlocks = append(prodBlocks, b)
		}
	}

	require.Len(t, prodBlocks, 1, "Should have exactly 1 ProductionQueueChangeBlock")
	prodBlock := prodBlocks[0]

	// Planet ID in file is 0-indexed (11 in file = planet #12 in 1-indexed game display)
	assert.Equal(t, 11, prodBlock.PlanetId, "Planet ID should be 11 (0-indexed)")

	// Check we have the right number of queue items
	assert.Len(t, prodBlock.Items, len(expected.Planet.Queue), "Queue should have %d items", len(expected.Planet.Queue))

	// Validate each item type
	for i, expectedItem := range expected.Planet.Queue {
		if i >= len(prodBlock.Items) {
			break
		}
		item := prodBlock.Items[i]

		switch expectedItem.Type {
		case "ship":
			assert.True(t, item.IsShipDesign(), "Item %d should be a ship design", i)
			assert.Equal(t, blocks.ProductionItemTypeCustom, item.ItemType, "Ship should have ItemType=4")
			// Validate ship design name by looking up the design ID
			if designs != nil {
				designName, ok := designs[item.ItemId]
				if assert.True(t, ok, "Design ID %d should exist", item.ItemId) {
					assert.Equal(t, expectedItem.Name, designName, "Ship design name should match")
				}
			}
		case "factory":
			assert.Equal(t, blocks.ProductionItemFactory, item.ItemId, "Factory should have ItemId=7")
			assert.Equal(t, expectedItem.Count, item.Count, "Factory count should match")
		case "mine":
			assert.Equal(t, blocks.ProductionItemMine, item.ItemId, "Mine should have ItemId=8")
			assert.Equal(t, expectedItem.Count, item.Count, "Mine count should match")
		}
	}
}

func TestScenarioOrders_XFileHeader(t *testing.T) {
	h := NewOrdersScenarioHelper(t)
	_, blockList := h.LoadXFile("game.x1")

	// First block should be FileHeader
	require.Greater(t, len(blockList), 0, "Should have at least one block")

	header, ok := blockList[0].(blocks.FileHeader)
	require.True(t, ok, "First block should be FileHeader")

	// Validate it's the same game as scenario-basic
	assert.Equal(t, uint16(0), header.Turn, "Turn should be 0")
	assert.Equal(t, 2400, header.Year(), "Year should be 2400")
	assert.Equal(t, 0, header.PlayerIndex(), "Player index should be 0 for x1 file")
}

func TestScenarioOrders_BlockTypes(t *testing.T) {
	h := NewOrdersScenarioHelper(t)
	_, blockList := h.LoadXFile("game.x1")

	// Count block types
	blockCounts := make(map[string]int)
	for _, block := range blockList {
		typeName := blocks.BlockTypeName(block.BlockTypeID())
		blockCounts[typeName]++
	}

	// We expect these block types in an X file with waypoints and production orders
	assert.Equal(t, 1, blockCounts["FileHeader"], "Should have 1 FileHeader")
	assert.Equal(t, 1, blockCounts["FileFooter"], "Should have 1 FileFooter")
	assert.GreaterOrEqual(t, blockCounts["WaypointAdd"], 1, "Should have at least 1 WaypointAdd")
	assert.Equal(t, 1, blockCounts["ProductionQueueChange"], "Should have 1 ProductionQueueChange")
}
