package store_test

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neper-stars/houston/store"
)

// ExpectedPlanet represents expected planet data from Game.pla TSV file.
type ExpectedPlanet struct {
	Name       string
	Owner      string // "Humanoid" or empty
	Starbase   string // "Starbase" or empty
	Population int64
	Mines      int
	Factories  int
	Resources  int
	SIron      int64 // Surface Ironium
	SBora      int64 // Surface Boranium
	SGerm      int64 // Surface Germanium
}

// ExpectedFleet represents expected fleet data from Game.fle TSV file.
type ExpectedFleet struct {
	Name    string
	X       int
	Y       int
	Planet  string // Planet name or empty
	ShipCnt int
	Iron    int64
	Bora    int64
	Germ    int64
	Col     int64 // Colonists
	Fuel    int64
}

// parseGamePla parses the Game.pla TSV file and returns expected planet data.
func parseGamePla(t *testing.T, path string) []ExpectedPlanet {
	t.Helper()

	file, err := os.Open(path)
	require.NoError(t, err, "Failed to open Game.pla")
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	reader.Comma = '\t'
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	require.NoError(t, err, "Failed to parse Game.pla")
	require.Greater(t, len(records), 1, "Game.pla should have header and data rows")

	var planets []ExpectedPlanet
	// Skip header row (index 0)
	for i, row := range records[1:] {
		if len(row) < 20 {
			t.Logf("Skipping row %d: not enough columns (%d)", i+1, len(row))
			continue
		}

		pop, _ := strconv.ParseInt(row[4], 10, 64)
		mines, _ := strconv.Atoi(row[7])
		factories, _ := strconv.Atoi(row[8])
		resources, _ := strconv.Atoi(row[19])
		sIron, _ := strconv.ParseInt(row[10], 10, 64)
		sBora, _ := strconv.ParseInt(row[11], 10, 64)
		sGerm, _ := strconv.ParseInt(row[12], 10, 64)

		planets = append(planets, ExpectedPlanet{
			Name:       row[0],
			Owner:      row[1],
			Starbase:   row[2],
			Population: pop,
			Mines:      mines,
			Factories:  factories,
			Resources:  resources,
			SIron:      sIron,
			SBora:      sBora,
			SGerm:      sGerm,
		})
	}

	return planets
}

// parseGameFle parses the Game.fle TSV file and returns expected fleet data.
func parseGameFle(t *testing.T, path string) []ExpectedFleet {
	t.Helper()

	file, err := os.Open(path)
	require.NoError(t, err, "Failed to open Game.fle")
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	reader.Comma = '\t'
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	require.NoError(t, err, "Failed to parse Game.fle")
	require.Greater(t, len(records), 1, "Game.fle should have header and data rows")

	var fleets []ExpectedFleet
	// Skip header row (index 0)
	for i, row := range records[1:] {
		if len(row) < 12 {
			t.Logf("Skipping row %d: not enough columns (%d)", i+1, len(row))
			continue
		}

		x, _ := strconv.Atoi(row[1])
		y, _ := strconv.Atoi(row[2])
		shipCnt, _ := strconv.Atoi(row[6])
		iron, _ := strconv.ParseInt(row[7], 10, 64)
		bora, _ := strconv.ParseInt(row[8], 10, 64)
		germ, _ := strconv.ParseInt(row[9], 10, 64)
		col, _ := strconv.ParseInt(row[10], 10, 64)
		fuel, _ := strconv.ParseInt(row[11], 10, 64)

		fleets = append(fleets, ExpectedFleet{
			Name:    row[0],
			X:       x,
			Y:       y,
			Planet:  row[3],
			ShipCnt: shipCnt,
			Iron:    iron,
			Bora:    bora,
			Germ:    germ,
			Col:     col,
			Fuel:    fuel,
		})
	}

	return fleets
}

// TestPlanetParsing_ScenarioSingleplayer verifies that planet data parsed from
// the binary .m1 file matches the Game.pla dump file.
func TestPlanetParsing_ScenarioSingleplayer(t *testing.T) {
	// Load expected data from Game.pla
	expectedPlanets := parseGamePla(t, "../testdata/scenario-singleplayer/2483/Game.pla")
	require.NotEmpty(t, expectedPlanets, "Should have parsed planets from Game.pla")

	// Load binary game files - need .xy for planet names, .m1 for planet data
	gs := store.New()

	xyData, err := os.ReadFile("../testdata/scenario-singleplayer/2483/Game.xy")
	require.NoError(t, err, "Failed to read Game.xy")
	err = gs.AddFile("Game.xy", xyData)
	require.NoError(t, err, "Failed to parse Game.xy")

	m1Data, err := os.ReadFile("../testdata/scenario-singleplayer/2483/Game.m1")
	require.NoError(t, err, "Failed to read Game.m1")
	err = gs.AddFile("Game.m1", m1Data)
	require.NoError(t, err, "Failed to parse Game.m1")

	// Filter to only owned planets (Owner = "Humanoid")
	var ownedExpected []ExpectedPlanet
	for _, p := range expectedPlanets {
		if p.Owner == "Humanoid" {
			ownedExpected = append(ownedExpected, p)
		}
	}
	require.NotEmpty(t, ownedExpected, "Should have owned planets in Game.pla")

	t.Logf("Found %d owned planets in Game.pla", len(ownedExpected))

	// Verify each owned planet
	for _, expected := range ownedExpected {
		t.Run(expected.Name, func(t *testing.T) {
			planet, ok := gs.PlanetByName(expected.Name)
			require.True(t, ok, "Planet %q should exist in parsed data", expected.Name)

			// Verify ownership
			assert.GreaterOrEqual(t, planet.Owner, 0, "Planet should be owned")

			// Verify population
			assert.Equal(t, expected.Population, planet.Population,
				"Population mismatch for %s", expected.Name)

			// Verify installations
			assert.Equal(t, expected.Mines, planet.Mines,
				"Mines mismatch for %s", expected.Name)
			assert.Equal(t, expected.Factories, planet.Factories,
				"Factories mismatch for %s", expected.Name)

			// Verify starbase
			hasStarbase := expected.Starbase != ""
			assert.Equal(t, hasStarbase, planet.HasStarbase,
				"HasStarbase mismatch for %s", expected.Name)

			// Verify surface minerals
			assert.Equal(t, expected.SIron, planet.Ironium,
				"Surface Ironium mismatch for %s", expected.Name)
			assert.Equal(t, expected.SBora, planet.Boranium,
				"Surface Boranium mismatch for %s", expected.Name)
			assert.Equal(t, expected.SGerm, planet.Germanium,
				"Surface Germanium mismatch for %s", expected.Name)
		})
	}

	// Verify total owned planet count
	parsedOwned := gs.PlanetsByOwner(0) // Player 0 = "Humanoid"
	assert.Equal(t, len(ownedExpected), len(parsedOwned),
		"Should have same number of owned planets")
}

// TestFleetParsing_ScenarioSingleplayer verifies that fleet data parsed from
// the binary .m1 file matches the Game.fle dump file.
func TestFleetParsing_ScenarioSingleplayer(t *testing.T) {
	// Load expected data from Game.fle
	expectedFleets := parseGameFle(t, "../testdata/scenario-singleplayer/2483/Game.fle")
	require.NotEmpty(t, expectedFleets, "Should have parsed fleets from Game.fle")

	// Load binary game files
	gs := store.New()

	xyData, err := os.ReadFile("../testdata/scenario-singleplayer/2483/Game.xy")
	require.NoError(t, err, "Failed to read Game.xy")
	err = gs.AddFile("Game.xy", xyData)
	require.NoError(t, err, "Failed to parse Game.xy")

	m1Data, err := os.ReadFile("../testdata/scenario-singleplayer/2483/Game.m1")
	require.NoError(t, err, "Failed to read Game.m1")
	err = gs.AddFile("Game.m1", m1Data)
	require.NoError(t, err, "Failed to parse Game.m1")

	t.Logf("Found %d fleets in Game.fle", len(expectedFleets))

	// Get all parsed fleets for player 0 (Humanoid)
	parsedFleets := gs.FleetsByOwner(0)
	t.Logf("Found %d fleets in parsed data", len(parsedFleets))

	// Verify fleet count
	assert.Equal(t, len(expectedFleets), len(parsedFleets),
		"Should have same number of fleets")

	// Create a map of parsed fleets by name for easier lookup
	parsedByName := make(map[string]*store.FleetEntity)
	for _, f := range parsedFleets {
		parsedByName[f.Name()] = f
	}

	// Verify each expected fleet
	for _, expected := range expectedFleets {
		// Game.fle includes race name prefix (e.g., "Humanoid Armed Probe #1")
		// Our parser generates just "Armed Probe #1", so strip the race prefix
		fleetName := strings.TrimPrefix(expected.Name, "Humanoid ")

		t.Run(expected.Name, func(t *testing.T) {
			fleet, ok := parsedByName[fleetName]
			if !ok {
				// Try to find by partial name match
				t.Logf("Fleet %q (lookup: %q) not found, available fleets:", expected.Name, fleetName)
				for name := range parsedByName {
					t.Logf("  - %s", name)
				}
				t.Fatalf("Fleet %q not found in parsed data", fleetName)
			}

			// Verify position
			assert.Equal(t, expected.X, fleet.X,
				"X position mismatch for %s", expected.Name)
			assert.Equal(t, expected.Y, fleet.Y,
				"Y position mismatch for %s", expected.Name)

			// Verify ship count
			assert.Equal(t, expected.ShipCnt, fleet.TotalShips(),
				"Ship count mismatch for %s", expected.Name)

			// Verify cargo
			cargo := fleet.GetCargo()
			assert.Equal(t, expected.Iron, cargo.Ironium,
				"Ironium cargo mismatch for %s", expected.Name)
			assert.Equal(t, expected.Bora, cargo.Boranium,
				"Boranium cargo mismatch for %s", expected.Name)
			assert.Equal(t, expected.Germ, cargo.Germanium,
				"Germanium cargo mismatch for %s", expected.Name)
			assert.Equal(t, expected.Col, cargo.Population,
				"Colonist cargo mismatch for %s", expected.Name)
			assert.Equal(t, expected.Fuel, cargo.Fuel,
				"Fuel mismatch for %s", expected.Name)

			// Verify planet location (if fleet is at a planet)
			if expected.Planet != "" {
				planet, planetOk := gs.PlanetByName(expected.Planet)
				if assert.True(t, planetOk, "Planet %q should exist", expected.Planet) {
					// Fleet position should match planet position
					assert.Equal(t, planet.X, fleet.X,
						"Fleet X should match planet X for %s at %s", expected.Name, expected.Planet)
					assert.Equal(t, planet.Y, fleet.Y,
						"Fleet Y should match planet Y for %s at %s", expected.Name, expected.Planet)
				}
			}
		})
	}
}

// TestPlanetResourceCalculation_ScenarioSingleplayer verifies that resource
// calculation matches the "Resources" column in Game.pla.
func TestPlanetResourceCalculation_ScenarioSingleplayer(t *testing.T) {
	// Load expected data from Game.pla
	expectedPlanets := parseGamePla(t, "../testdata/scenario-singleplayer/2483/Game.pla")
	require.NotEmpty(t, expectedPlanets, "Should have parsed planets from Game.pla")

	// Load binary game files
	gs := store.New()

	xyData, err := os.ReadFile("../testdata/scenario-singleplayer/2483/Game.xy")
	require.NoError(t, err, "Failed to read Game.xy")
	err = gs.AddFile("Game.xy", xyData)
	require.NoError(t, err, "Failed to parse Game.xy")

	m1Data, err := os.ReadFile("../testdata/scenario-singleplayer/2483/Game.m1")
	require.NoError(t, err, "Failed to read Game.m1")
	err = gs.AddFile("Game.m1", m1Data)
	require.NoError(t, err, "Failed to parse Game.m1")

	player, ok := gs.Player(0)
	require.True(t, ok, "Player 0 should exist")

	// Filter to only owned planets with resources
	for _, expected := range expectedPlanets {
		if expected.Owner != "Humanoid" || expected.Resources == 0 {
			continue
		}

		t.Run(expected.Name, func(t *testing.T) {
			planet, ok := gs.PlanetByName(expected.Name)
			require.True(t, ok, "Planet %q should exist", expected.Name)

			calculatedResources := gs.CResourcesAtPlanet(planet, player)
			assert.Equal(t, expected.Resources, calculatedResources,
				"Resource calculation mismatch for %s (pop=%d, mines=%d, factories=%d)",
				expected.Name, planet.Population, planet.Mines, planet.Factories)
		})
	}
}

// ExpectedMapPlanet represents a planet from Game.map TSV file.
type ExpectedMapPlanet struct {
	Number int // 1-indexed planet number
	X      int
	Y      int
	Name   string
}

// parseGameMap parses the Game.map TSV file and returns expected planet map data.
func parseGameMap(t *testing.T, path string) []ExpectedMapPlanet {
	t.Helper()

	file, err := os.Open(path)
	require.NoError(t, err, "Failed to open Game.map")
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	reader.Comma = '\t'
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	require.NoError(t, err, "Failed to parse Game.map")
	require.Greater(t, len(records), 1, "Game.map should have header and data rows")

	var planets []ExpectedMapPlanet
	// Skip header row (index 0), header is: #	X	Y	Name
	for i, row := range records[1:] {
		if len(row) < 4 {
			t.Logf("Skipping row %d: not enough columns (%d)", i+1, len(row))
			continue
		}

		num, _ := strconv.Atoi(row[0])
		x, _ := strconv.Atoi(row[1])
		y, _ := strconv.Atoi(row[2])

		planets = append(planets, ExpectedMapPlanet{
			Number: num,
			X:      x,
			Y:      y,
			Name:   row[3],
		})
	}

	return planets
}

// TestPlanetMapVisibility_ScenarioSingleplayer verifies that all planets in the
// universe are visible and have correct coordinates when parsing the binary files.
// Uses Game.map as source of truth for the complete universe map.
func TestPlanetMapVisibility_ScenarioSingleplayer(t *testing.T) {
	// Load expected data from Game.map
	expectedPlanets := parseGameMap(t, "../testdata/scenario-singleplayer/2483/Game.map")
	require.NotEmpty(t, expectedPlanets, "Should have parsed planets from Game.map")
	t.Logf("Found %d planets in Game.map", len(expectedPlanets))

	// Load binary game files
	gs := store.New()

	xyData, err := os.ReadFile("../testdata/scenario-singleplayer/2483/Game.xy")
	require.NoError(t, err, "Failed to read Game.xy")
	err = gs.AddFile("Game.xy", xyData)
	require.NoError(t, err, "Failed to parse Game.xy")

	m1Data, err := os.ReadFile("../testdata/scenario-singleplayer/2483/Game.m1")
	require.NoError(t, err, "Failed to read Game.m1")
	err = gs.AddFile("Game.m1", m1Data)
	require.NoError(t, err, "Failed to parse Game.m1")

	// Check total planet count
	allPlanets := gs.AllPlanets()
	t.Logf("Found %d planets in parsed data", len(allPlanets))

	// Build map of parsed planets by number for quick lookup
	parsedByNumber := make(map[int]*store.PlanetEntity)
	for _, p := range allPlanets {
		parsedByNumber[p.PlanetNumber] = p
	}

	// Verify each expected planet exists with correct coordinates
	var missingPlanets []string
	var coordMismatches []string

	for _, expected := range expectedPlanets {
		// Game.map uses 1-indexed, our parser uses 0-indexed
		planetNum := expected.Number - 1

		parsed, ok := parsedByNumber[planetNum]
		if !ok {
			missingPlanets = append(missingPlanets,
				fmt.Sprintf("#%d %s", expected.Number, expected.Name))
			continue
		}

		// Verify coordinates
		if parsed.X != expected.X || parsed.Y != expected.Y {
			coordMismatches = append(coordMismatches,
				fmt.Sprintf("#%d %s: expected (%d,%d), got (%d,%d)",
					expected.Number, expected.Name,
					expected.X, expected.Y, parsed.X, parsed.Y))
		}

		// Verify name matches
		parsedName := gs.PlanetName(planetNum)
		if parsedName != expected.Name {
			t.Errorf("Planet #%d name mismatch: expected %q, got %q",
				expected.Number, expected.Name, parsedName)
		}
	}

	// Report missing planets
	if len(missingPlanets) > 0 {
		t.Errorf("Missing %d planets:", len(missingPlanets))
		for _, p := range missingPlanets[:min(10, len(missingPlanets))] {
			t.Errorf("  - %s", p)
		}
		if len(missingPlanets) > 10 {
			t.Errorf("  ... and %d more", len(missingPlanets)-10)
		}
	}

	// Report coordinate mismatches
	if len(coordMismatches) > 0 {
		t.Errorf("Coordinate mismatches for %d planets:", len(coordMismatches))
		for _, m := range coordMismatches[:min(10, len(coordMismatches))] {
			t.Errorf("  - %s", m)
		}
		if len(coordMismatches) > 10 {
			t.Errorf("  ... and %d more", len(coordMismatches)-10)
		}
	}

	// Summary assertion
	assert.Equal(t, len(expectedPlanets), len(allPlanets),
		"Should have same number of planets as Game.map")
	assert.Empty(t, missingPlanets, "All planets should be found")
	assert.Empty(t, coordMismatches, "All coordinates should match")
}

// TestAllPlanetNames_ScenarioSingleplayer verifies all planet names are parsed correctly.
func TestAllPlanetNames_ScenarioSingleplayer(t *testing.T) {
	// Load expected data from Game.pla
	expectedPlanets := parseGamePla(t, "../testdata/scenario-singleplayer/2483/Game.pla")
	require.NotEmpty(t, expectedPlanets, "Should have parsed planets from Game.pla")

	// Load binary game file - need both .m1 and .xy for full planet names
	gs := store.New()

	xyData, err := os.ReadFile("../testdata/scenario-singleplayer/2483/Game.xy")
	require.NoError(t, err, "Failed to read Game.xy")
	err = gs.AddFile("Game.xy", xyData)
	require.NoError(t, err, "Failed to parse Game.xy")

	m1Data, err := os.ReadFile("../testdata/scenario-singleplayer/2483/Game.m1")
	require.NoError(t, err, "Failed to read Game.m1")
	err = gs.AddFile("Game.m1", m1Data)
	require.NoError(t, err, "Failed to parse Game.m1")

	// Check that all expected planet names exist
	missingPlanets := []string{}
	for _, expected := range expectedPlanets {
		// Skip empty names
		if strings.TrimSpace(expected.Name) == "" {
			continue
		}

		_, ok := gs.PlanetByName(expected.Name)
		if !ok {
			missingPlanets = append(missingPlanets, expected.Name)
		}
	}

	if len(missingPlanets) > 0 {
		t.Logf("Missing %d planets:", len(missingPlanets))
		for _, name := range missingPlanets {
			t.Logf("  - %q", name)
		}
		// Log available planets for debugging
		allPlanets := gs.AllPlanets()
		t.Logf("Available planets in parsed data (%d total):", len(allPlanets))
		for i, p := range allPlanets {
			if i < 20 { // Only show first 20
				t.Logf("  %d: %q", p.PlanetNumber, gs.PlanetName(p.PlanetNumber))
			}
		}
	}

	assert.Empty(t, missingPlanets, "All planet names should be found")
}

// TestVictoryConditions_ScenarioSingleplayer verifies victory conditions are parsed correctly.
// Expected values are from testdata/scenario-singleplayer/2483/victory-conditions.png
func TestVictoryConditions_ScenarioSingleplayer(t *testing.T) {
	// Load Game.xy which contains the PlanetsBlock with victory conditions
	gs := store.New()

	xyData, err := os.ReadFile("../testdata/scenario-singleplayer/2483/Game.xy")
	require.NoError(t, err, "Failed to read Game.xy")
	err = gs.AddFile("Game.xy", xyData)
	require.NoError(t, err, "Failed to parse Game.xy")

	// Verify universe settings
	assert.Equal(t, uint16(2), gs.UniverseSize, "UniverseSize should be 2 (Medium)")
	assert.Equal(t, "Medium", gs.UniverseSizeName(), "UniverseSizeName should be Medium")
	assert.Equal(t, uint16(3), gs.Density, "Density should be 3 (Packed)")
	assert.Equal(t, "Packed", gs.DensityName(), "DensityName should be Packed")
	assert.Equal(t, uint16(1), gs.PlayerCount, "PlayerCount should be 1")
	assert.Equal(t, uint16(540), gs.PlanetCount, "PlanetCount should be 540")

	// Verify victory conditions from victory-conditions.png:
	// ☑ Owns 60% of all planets
	// ☑ Attains Tech 22 in 4 fields
	// ☐ Exceeds a score of 11000
	// ☑ Exceeds second place score by 100%
	// ☐ Has a production capacity of 100 thousand
	// ☐ Owns 100 capital ships
	// ☐ Has the highest score after 100 years
	// Winner must meet 1 of the above selected criteria
	// At least 70 years must pass before a winner is declared

	vc := gs.VictoryConditions

	// [0] Owns % of planets: enabled=true, value=60%
	assert.True(t, vc.OwnsPercentPlanetsEnabled, "Owns % planets should be enabled")
	assert.Equal(t, 60, vc.OwnsPercentPlanetsValue, "Owns % planets value should be 60")

	// [1] Attains Tech X in Y fields: enabled=true, level=22, fields=4
	assert.True(t, vc.AttainTechLevelEnabled, "Attain tech level should be enabled")
	assert.Equal(t, 22, vc.AttainTechLevelValue, "Tech level should be 22")
	assert.Equal(t, 4, vc.AttainTechInYFields, "Tech fields should be 4")

	// [3] Exceeds score: enabled=false, value=11000
	assert.False(t, vc.ExceedScoreEnabled, "Exceeds score should be disabled")
	assert.Equal(t, 11000, vc.ExceedScoreValue, "Exceeds score value should be 11000")

	// [4] Exceeds 2nd place by %: enabled=true, value=100%
	assert.True(t, vc.ExceedSecondPlaceEnabled, "Exceeds 2nd place should be enabled")
	assert.Equal(t, 100, vc.ExceedSecondPlaceValue, "Exceeds 2nd place value should be 100")

	// [5] Production capacity: enabled=false, value=100k
	assert.False(t, vc.ProductionCapacityEnabled, "Production capacity should be disabled")
	assert.Equal(t, 100, vc.ProductionCapacityValue, "Production capacity value should be 100")

	// [6] Own capital ships: enabled=false, value=100
	assert.False(t, vc.OwnCapitalShipsEnabled, "Own capital ships should be disabled")
	assert.Equal(t, 100, vc.OwnCapitalShipsValue, "Own capital ships value should be 100")

	// [7] Highest score after N years: enabled=false, value=100
	assert.False(t, vc.HighestScoreYearsEnabled, "Highest score years should be disabled")
	assert.Equal(t, 100, vc.HighestScoreYearsValue, "Highest score years value should be 100")

	// [8] Must meet N criteria: value=1
	assert.Equal(t, 1, vc.NumCriteriaMetValue, "Must meet N criteria value should be 1")

	// [9] Min years before winner: value=70
	assert.Equal(t, 70, vc.MinYearsBeforeWinValue, "Min years before winner value should be 70")
}
