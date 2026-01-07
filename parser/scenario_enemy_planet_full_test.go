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

// Expected data structures for enemy-planets-full scenario
type ExpectedEnemyPlanetEnvironment struct {
	GravityInternal     int    `json:"gravityInternal"`
	TemperatureInternal int    `json:"temperatureInternal"`
	RadiationInternal   int    `json:"radiationInternal"`
	GravityDisplay      string `json:"gravityDisplay"`
	TemperatureDisplay  string `json:"temperatureDisplay"`
	RadiationDisplay    string `json:"radiationDisplay"`
}

type ExpectedEnemyPlanetConcentrations struct {
	Ironium   int `json:"ironium"`
	Boranium  int `json:"boranium"`
	Germanium int `json:"germanium"`
}

type ExpectedEnemyPlanetEstimates struct {
	Population int `json:"population"`
	Defenses   int `json:"defenses"`
}

type ExpectedEnemyPlanet struct {
	Name           string                            `json:"name"`
	DisplayID      int                               `json:"displayId"`
	PlanetNumber   int                               `json:"planetNumber"`
	Owner          int                               `json:"owner"`
	OwnerName      string                            `json:"ownerName"`
	IsHomeworld    bool                              `json:"isHomeworld"`
	DetectionLevel string                            `json:"detectionLevel"` // "NotVisible", "PenScan", "Special", "NormalScan", "Full", "Maximum"
	HasStarbase    bool                              `json:"hasStarbase"`
	StarbaseDesign int                               `json:"starbaseDesign"`
	Environment    ExpectedEnemyPlanetEnvironment    `json:"environment"`
	Concentrations ExpectedEnemyPlanetConcentrations `json:"concentrations"`
	Estimates      ExpectedEnemyPlanetEstimates      `json:"estimates"`
}

// parseDetectionLevel converts a detection level string to its numeric value.
func parseDetectionLevel(s string) int {
	switch s {
	case "NotVisible":
		return blocks.DetNotVisible
	case "PenScan":
		return blocks.DetPenScan
	case "Special":
		return blocks.DetSpecial
	case "NormalScan":
		return blocks.DetNormalScan
	case "Full":
		return blocks.DetFull
	case "Maximum":
		return blocks.DetMaximum
	default:
		return 0
	}
}

type ExpectedEnemyDesign struct {
	DesignNumber int    `json:"designNumber"`
	Name         string `json:"name"`
	HullId       int    `json:"hullId"`
	HullName     string `json:"hullName"`
	IsStarbase   bool   `json:"isStarbase"`
	Mass         int    `json:"mass"`
}

type ExpectedEnemyPlanetFullData struct {
	Scenario     string                `json:"scenario"`
	Description  string                `json:"description"`
	EnemyPlanets []ExpectedEnemyPlanet `json:"enemyPlanets"`
	EnemyDesigns []ExpectedEnemyDesign `json:"enemyDesigns"`
}

func loadEnemyPlanetFullExpected(t *testing.T) *ExpectedEnemyPlanetFullData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-enemy-planets-full", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedEnemyPlanetFullData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadEnemyPlanetFullFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-enemy-planets-full", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

func TestScenarioEnemyPlanetFull_PartialPlanetBlocks(t *testing.T) {
	expected := loadEnemyPlanetFullExpected(t)
	_, blockList := loadEnemyPlanetFullFile(t, "game.m1")

	// Build map of enemy planets (owner >= 0 in PartialPlanetBlock)
	enemyPlanets := make(map[int]blocks.PartialPlanetBlock)
	for _, block := range blockList {
		if ppb, ok := block.(blocks.PartialPlanetBlock); ok {
			if ppb.Owner >= 0 {
				enemyPlanets[ppb.PlanetNumber] = ppb
			}
		}
	}

	require.Equal(t, len(expected.EnemyPlanets), len(enemyPlanets),
		"Should have %d enemy planets with owner info", len(expected.EnemyPlanets))

	// Validate each expected enemy planet
	for _, exp := range expected.EnemyPlanets {
		planet, found := enemyPlanets[exp.PlanetNumber]
		require.True(t, found, "Should find planet number %d (%s)", exp.PlanetNumber, exp.Name)

		t.Run(exp.Name, func(t *testing.T) {
			// Basic identification
			assert.Equal(t, exp.PlanetNumber, planet.PlanetNumber, "Planet number should match")
			assert.Equal(t, exp.Owner, planet.Owner, "Owner should match")
			assert.Equal(t, exp.IsHomeworld, planet.IsHomeworld, "IsHomeworld should match")

			// Detection level (determines what info is visible)
			assert.Equal(t, parseDetectionLevel(exp.DetectionLevel), planet.DetectionLevel,
				"DetectionLevel should match (expected=%s)", exp.DetectionLevel)

			// Starbase
			assert.Equal(t, exp.HasStarbase, planet.HasStarbase, "HasStarbase should match")
			if exp.HasStarbase {
				assert.Equal(t, exp.StarbaseDesign, planet.StarbaseDesign, "StarbaseDesign should match")
			}

			// Environment values (internal representation)
			assert.Equal(t, exp.Environment.GravityInternal, planet.Gravity, "Gravity internal should match")
			assert.Equal(t, exp.Environment.TemperatureInternal, planet.Temperature, "Temperature internal should match")
			assert.Equal(t, exp.Environment.RadiationInternal, planet.Radiation, "Radiation internal should match")

			// Mineral concentrations (penetrating scanner data)
			assert.Equal(t, exp.Concentrations.Ironium, planet.IroniumConc, "Ironium concentration should match")
			assert.Equal(t, exp.Concentrations.Boranium, planet.BoraniumConc, "Boranium concentration should match")
			assert.Equal(t, exp.Concentrations.Germanium, planet.GermaniumConc, "Germanium concentration should match")

			// Population and defense estimates
			assert.Equal(t, exp.Estimates.Population, planet.PopEstimate, "Population estimate should match")
			assert.Equal(t, exp.Estimates.Defenses, planet.DefensesEstimate, "Defenses estimate should match")
		})
	}
}

// TestScenarioEnemyPlanetFull_EnvironmentConversions tests the conversion formulas
// from internal values to display values
func TestScenarioEnemyPlanetFull_EnvironmentConversions(t *testing.T) {
	expected := loadEnemyPlanetFullExpected(t)
	_, blockList := loadEnemyPlanetFullFile(t, "game.m1")

	// Find the enemy planet
	var planet *blocks.PartialPlanetBlock
	for _, block := range blockList {
		if ppb, ok := block.(blocks.PartialPlanetBlock); ok {
			if ppb.PlanetNumber == expected.EnemyPlanets[0].PlanetNumber {
				planet = &ppb
				break
			}
		}
	}
	require.NotNil(t, planet, "Should find the enemy planet")

	// Temperature conversion: display = (internal - 50) * 4
	// Internal 56 -> (56-50)*4 = 24°C
	tempCelsius := (planet.Temperature - 50) * 4
	assert.Equal(t, 24, tempCelsius, "Temperature should convert to 24°C")

	// Radiation: direct mapping (internal value = mR)
	assert.Equal(t, 63, planet.Radiation, "Radiation should be 63mR")
}

// TestScenarioEnemyPlanetFull_EnemyDesigns tests parsing of enemy (partial) designs
// Enemy designs only show name, hull, and mass - not component slots
func TestScenarioEnemyPlanetFull_EnemyDesigns(t *testing.T) {
	expected := loadEnemyPlanetFullExpected(t)
	_, blockList := loadEnemyPlanetFullFile(t, "game.m1")

	// Collect enemy designs (IsFullDesign = false)
	var enemyDesigns []blocks.DesignBlock
	for _, block := range blockList {
		if db, ok := block.(blocks.DesignBlock); ok {
			if !db.IsFullDesign {
				enemyDesigns = append(enemyDesigns, db)
			}
		}
	}

	require.Equal(t, len(expected.EnemyDesigns), len(enemyDesigns),
		"Should have %d enemy designs", len(expected.EnemyDesigns))

	// Validate each expected enemy design
	for i, exp := range expected.EnemyDesigns {
		design := enemyDesigns[i]
		t.Run(exp.Name, func(t *testing.T) {
			// Basic identification
			assert.Equal(t, exp.DesignNumber, design.DesignNumber, "Design number should match")
			assert.Equal(t, exp.Name, design.Name, "Design name should match")
			assert.Equal(t, exp.HullId, design.HullId, "Hull ID should match")
			assert.Equal(t, exp.IsStarbase, design.IsStarbase, "IsStarbase should match")

			// Enemy designs have mass but no component details
			assert.Equal(t, exp.Mass, design.Mass, "Mass should match")
			assert.False(t, design.IsFullDesign, "Should be a partial design")
			assert.Equal(t, 0, design.SlotCount, "Partial design should have no slot count")
			assert.Empty(t, design.Slots, "Partial design should have no slots")
		})
	}
}
