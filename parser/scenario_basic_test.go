package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/data"
)

// Expected data structures for JSON parsing
type ExpectedFleet struct {
	Name      string `json:"name"`
	ShipCount int    `json:"shipCount"`
}

type ExpectedDesign struct {
	Name      string `json:"name"`
	Hull      string `json:"hull"`
	SlotCount int    `json:"slotCount"`
}

type ExpectedPlayer struct {
	RaceName       string `json:"raceName"`
	RacePluralName string `json:"racePluralName"`
	Homeworld      struct {
		Name string `json:"name"`
	} `json:"homeworld"`
	Fleets  []ExpectedFleet  `json:"fleets"`
	Designs []ExpectedDesign `json:"designs"`
}

type ExpectedData struct {
	Scenario string `json:"scenario"`
	Game     struct {
		Name string `json:"name"`
		Year int    `json:"year"`
	} `json:"game"`
	Player1  ExpectedPlayer `json:"player1"`
	Player2  ExpectedPlayer `json:"player2"`
	Universe struct {
		TotalPlanets string `json:"totalPlanets"`
	} `json:"universe"`
}

// ScenarioHelper provides utilities for loading test scenario data
type ScenarioHelper struct {
	t        *testing.T
	dir      string
	Expected *ExpectedData
}

// NewScenarioHelper creates a helper for the given scenario directory
func NewScenarioHelper(t *testing.T, scenarioName string) *ScenarioHelper {
	t.Helper()

	dir := filepath.Join("..", "testdata", scenarioName)
	expectedPath := filepath.Join(dir, "expected.json")

	// Skip if test files don't exist
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Skipf("Test files not found in testdata/%s/", scenarioName)
	}

	// Load expected data
	data, err := os.ReadFile(expectedPath)
	require.NoError(t, err, "Failed to read expected.json")

	var expected ExpectedData
	err = json.Unmarshal(data, &expected)
	require.NoError(t, err, "Failed to parse expected.json")

	return &ScenarioHelper{
		t:        t,
		dir:      dir,
		Expected: &expected,
	}
}

// LoadFile loads and parses a game file from the scenario directory
func (s *ScenarioHelper) LoadFile(filename string) (FileData, []blocks.Block) {
	s.t.Helper()

	path := filepath.Join(s.dir, filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(s.t, err, "Failed to read %s", filename)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(s.t, err, "BlockList() failed for %s", filename)

	return fd, blockList
}

// LoadFileHeader loads a file and returns its header
func (s *ScenarioHelper) LoadFileHeader(filename string) (*blocks.FileHeader, FileData) {
	s.t.Helper()

	path := filepath.Join(s.dir, filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(s.t, err, "Failed to read %s", filename)

	fd := FileData(fileBytes)
	header, err := fd.FileHeader()
	require.NoError(s.t, err, "FileHeader() failed for %s", filename)

	return header, fd
}

// LoadPlanets loads the .xy file and returns the PlanetsBlock with all planet data
func (s *ScenarioHelper) LoadPlanets(filename string) *blocks.PlanetsBlock {
	s.t.Helper()

	_, blockList := s.LoadFile(filename)

	for _, block := range blockList {
		if p, ok := block.(blocks.PlanetsBlock); ok {
			return &p
		}
	}

	s.t.Fatalf("No PlanetsBlock found in %s", filename)
	return nil
}

// GetPlanetName returns the name of a planet by its ID from the PlanetsBlock
func (s *ScenarioHelper) GetPlanetName(planets *blocks.PlanetsBlock, planetID int) string {
	s.t.Helper()

	if planetID < 0 || planetID >= len(planets.Planets) {
		s.t.Fatalf("Invalid planet ID: %d (max: %d)", planetID, len(planets.Planets)-1)
	}
	return planets.Planets[planetID].Name
}

func TestScenarioBasic_Player1(t *testing.T) {
	h := NewScenarioHelper(t, "scenario-basic")

	header, _ := h.LoadFileHeader("game.m1")
	_, blockList := h.LoadFile("game.m1")

	// Test FileHeader
	assert.Equal(t, uint16(0), header.Turn, "Turn should be 0")
	assert.Equal(t, h.Expected.Game.Year, header.Year(), "Year should match expected")
	assert.Equal(t, 0, header.PlayerIndex(), "Player index should be 0 for m1 file")

	// Find and validate PlayerBlock for player 0
	var player0 *blocks.PlayerBlock
	var fleetBlocks []blocks.FleetBlock
	var designBlocks []blocks.DesignBlock

	for _, block := range blockList {
		switch b := block.(type) {
		case blocks.PlayerBlock:
			if b.PlayerNumber == 0 {
				player0 = &b
			}
		case blocks.FleetBlock:
			if b.Owner == 0 {
				fleetBlocks = append(fleetBlocks, b)
			}
		case blocks.DesignBlock:
			designBlocks = append(designBlocks, b)
		}
	}

	// Validate Player block
	require.NotNil(t, player0, "Player 0 block should exist")
	assert.Equal(t, h.Expected.Player1.RaceName, player0.NameSingular, "Race singular name should match")
	assert.Equal(t, h.Expected.Player1.RacePluralName, player0.NamePlural, "Race plural name should match")
	assert.Equal(t, len(h.Expected.Player1.Fleets), player0.Fleets, "Fleet count should match")
	assert.Equal(t, 6, player0.ShipDesignCount, "Ship design count should be 6")
	assert.Equal(t, 1, player0.StarbaseDesignCount, "Starbase design count should be 1")

	// Validate fleet count matches expected
	assert.Equal(t, len(h.Expected.Player1.Fleets), len(fleetBlocks), "Number of fleet blocks should match expected")

	// Validate design count
	assert.GreaterOrEqual(t, len(designBlocks), 6, "Should have at least 6 ship designs")
}

func TestScenarioBasic_Player2(t *testing.T) {
	h := NewScenarioHelper(t, "scenario-basic")

	header, _ := h.LoadFileHeader("game.m2")
	_, blockList := h.LoadFile("game.m2")

	// Test FileHeader
	assert.Equal(t, uint16(0), header.Turn, "Turn should be 0")
	assert.Equal(t, h.Expected.Game.Year, header.Year(), "Year should match expected")
	assert.Equal(t, 1, header.PlayerIndex(), "Player index should be 1 for m2 file")

	// Find and validate PlayerBlock for player 1
	var player1 *blocks.PlayerBlock
	var fleetBlocks []blocks.FleetBlock

	for _, block := range blockList {
		switch b := block.(type) {
		case blocks.PlayerBlock:
			if b.PlayerNumber == 1 {
				player1 = &b
			}
		case blocks.FleetBlock:
			if b.Owner == 1 {
				fleetBlocks = append(fleetBlocks, b)
			}
		}
	}

	// Validate Player block
	require.NotNil(t, player1, "Player 1 block should exist")
	assert.Equal(t, h.Expected.Player2.RaceName, player1.NameSingular, "Race singular name should match")
	assert.Equal(t, h.Expected.Player2.RacePluralName, player1.NamePlural, "Race plural name should match")
	assert.Equal(t, len(h.Expected.Player2.Fleets), player1.Fleets, "Fleet count should match")
}

func TestScenarioBasic_Designs(t *testing.T) {
	h := NewScenarioHelper(t, "scenario-basic")

	_, blockList := h.LoadFile("game.m1")

	// Collect ship designs
	var designs []blocks.DesignBlock
	for _, block := range blockList {
		if d, ok := block.(blocks.DesignBlock); ok && !d.IsStarbase {
			designs = append(designs, d)
		}
	}

	// Build a map for easy lookup
	designByName := make(map[string]*blocks.DesignBlock)
	for i := range designs {
		designByName[designs[i].Name] = &designs[i]
	}

	require.GreaterOrEqual(t, len(designs), len(h.Expected.Player1.Designs),
		"Should have at least %d ship designs", len(h.Expected.Player1.Designs))

	// Validate each expected design
	for _, expected := range h.Expected.Player1.Designs {
		t.Run(expected.Name, func(t *testing.T) {
			design, found := designByName[expected.Name]
			require.True(t, found, "Design '%s' should exist", expected.Name)

			// Validate hull using data package mapping
			expectedHullID, hullFound := data.HullNameToID[expected.Hull]
			require.True(t, hullFound, "Hull '%s' should be a valid hull name", expected.Hull)
			assert.Equal(t, expectedHullID, design.HullId,
				"Design '%s' should have hull '%s' (ID %d), got ID %d",
				expected.Name, expected.Hull, expectedHullID, design.HullId)

			// Validate slot count
			assert.Equal(t, expected.SlotCount, design.SlotCount,
				"Design '%s' should have %d slots, got %d",
				expected.Name, expected.SlotCount, design.SlotCount)
		})
	}
}

func TestScenarioBasic_OwnedPlanets(t *testing.T) {
	h := NewScenarioHelper(t, "scenario-basic")

	_, blockList := h.LoadFile("game.m1")

	// Find owned planet blocks (PlanetsBlock with coordinates is in .xy file)
	var ownedPlanets []blocks.PlanetBlock
	var partialPlanets []blocks.PartialPlanetBlock

	for _, block := range blockList {
		switch b := block.(type) {
		case blocks.PlanetBlock:
			if b.Owner >= 0 {
				ownedPlanets = append(ownedPlanets, b)
			}
		case blocks.PartialPlanetBlock:
			partialPlanets = append(partialPlanets, b)
		}
	}

	// Player 1 should own at least their homeworld
	assert.GreaterOrEqual(t, len(ownedPlanets), 1, "Should have at least 1 owned planet (homeworld)")

	// The owned planet should belong to player 0
	for _, p := range ownedPlanets {
		assert.Equal(t, 0, p.Owner, "Owned planet should belong to player 0")
	}
}

func TestScenarioBasic_PlanetNames(t *testing.T) {
	h := NewScenarioHelper(t, "scenario-basic")

	// Load planet data from .xy file
	planets := h.LoadPlanets("game.xy")

	// Validate planet count matches expected
	assert.Equal(t, 32, planets.GetPlanetCount(), "Universe should have 32 planets")
	assert.Equal(t, 32, len(planets.Planets), "Should have parsed 32 planet entries")

	// Validate game settings from PlanetsBlock
	assert.Equal(t, uint16(2), planets.PlayerCount, "Should have 2 players")
	assert.Contains(t, planets.GameName, "TestBed01", "Game name should match")

	// Load M1 file to find player 1's homeworld by owner
	_, blockList := h.LoadFile("game.m1")
	var player1Homeworld *blocks.PlanetBlock
	for _, block := range blockList {
		if p, ok := block.(blocks.PlanetBlock); ok && p.Owner == 0 {
			player1Homeworld = &p
			break
		}
	}
	require.NotNil(t, player1Homeworld, "Player 1 should have an owned planet")

	// Get the name from the XY file using the planet ID
	homeworldName := h.GetPlanetName(planets, player1Homeworld.PlanetNumber)
	assert.Equal(t, h.Expected.Player1.Homeworld.Name, homeworldName, "Player 1 homeworld name should match")

	// Load M2 file to find player 2's homeworld
	_, blockList2 := h.LoadFile("game.m2")
	var player2Homeworld *blocks.PlanetBlock
	for _, block := range blockList2 {
		if p, ok := block.(blocks.PlanetBlock); ok && p.Owner == 1 {
			player2Homeworld = &p
			break
		}
	}
	require.NotNil(t, player2Homeworld, "Player 2 should have an owned planet")

	// Get the name from the XY file using the planet ID
	homeworld2Name := h.GetPlanetName(planets, player2Homeworld.PlanetNumber)
	assert.Equal(t, h.Expected.Player2.Homeworld.Name, homeworld2Name, "Player 2 homeworld name should match")
}
