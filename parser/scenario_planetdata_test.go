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

// Expected data structures for planet scenario
type ExpectedEnvironment struct {
	Gravity     int `json:"gravity"`
	Temperature int `json:"temperature"`
	Radiation   int `json:"radiation"`
}

type ExpectedConcentrations struct {
	Ironium   int `json:"ironium"`
	Boranium  int `json:"boranium"`
	Germanium int `json:"germanium"`
}

type ExpectedSurfaceMinerals struct {
	Ironium   int64 `json:"ironium"`
	Boranium  int64 `json:"boranium"`
	Germanium int64 `json:"germanium"`
}

type ExpectedInstallations struct {
	Mines     int `json:"mines"`
	Factories int `json:"factories"`
	Defenses  int `json:"defenses"`
}

type ExpectedOwnedPlanet struct {
	PlanetID        int                     `json:"planetId"`
	Name            string                  `json:"name"`
	Owner           int                     `json:"owner"`
	IsHomeworld     bool                    `json:"isHomeworld"`
	Environment     ExpectedEnvironment     `json:"environment"`
	Concentrations  ExpectedConcentrations  `json:"concentrations"`
	SurfaceMinerals ExpectedSurfaceMinerals `json:"surfaceMinerals"`
	Population      int64                   `json:"population"`
	Installations   ExpectedInstallations   `json:"installations"`
	ScannerID       int                     `json:"scannerId"` // 0=none, 31=no scanner, 1-30=scanner type
	HasStarbase     bool                    `json:"hasStarbase"`
	StarbaseDesign  int                     `json:"starbaseDesign"`
}

type ExpectedScannedPlanet struct {
	PlanetID        int                      `json:"planetId"`
	Name            string                   `json:"name"`
	Owner           int                      `json:"owner"`
	Environment     ExpectedEnvironment      `json:"environment"`
	Concentrations  ExpectedConcentrations   `json:"concentrations"`
	SurfaceMinerals *ExpectedSurfaceMinerals `json:"surfaceMinerals,omitempty"`
}

type ExpectedPlanetData struct {
	Scenario       string                  `json:"scenario"`
	OwnedPlanets   []ExpectedOwnedPlanet   `json:"ownedPlanets"`
	ScannedPlanets []ExpectedScannedPlanet `json:"scannedPlanets"`
}

func loadPlanetExpected(t *testing.T) *ExpectedPlanetData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-planetdata", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedPlanetData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadPlanetFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-planetdata", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

// loadPlanetNames loads planet ID to name mapping from the XY file
func loadPlanetNames(t *testing.T) map[int]string {
	t.Helper()
	_, blockList := loadPlanetFile(t, "game.xy")

	planetNames := make(map[int]string)
	for _, block := range blockList {
		if pb, ok := block.(blocks.PlanetsBlock); ok {
			for _, p := range pb.Planets {
				planetNames[p.ID] = p.Name
			}
			break
		}
	}
	return planetNames
}

func TestScenarioPlanetData_OwnedPlanet(t *testing.T) {
	expected := loadPlanetExpected(t)
	_, blockList := loadPlanetFile(t, "game.m2")
	planetNames := loadPlanetNames(t)

	// Collect owned planets (PlanetBlock Type 13)
	var ownedPlanets []blocks.PlanetBlock
	for _, block := range blockList {
		if p, ok := block.(blocks.PlanetBlock); ok {
			ownedPlanets = append(ownedPlanets, p)
		}
	}

	require.GreaterOrEqual(t, len(ownedPlanets), len(expected.OwnedPlanets),
		"Should have at least %d owned planets", len(expected.OwnedPlanets))

	// Validate each expected owned planet
	for _, exp := range expected.OwnedPlanets {
		t.Run(exp.Name, func(t *testing.T) {
			// Find planet by ID
			var planet *blocks.PlanetBlock
			for i := range ownedPlanets {
				if ownedPlanets[i].PlanetNumber == exp.PlanetID {
					planet = &ownedPlanets[i]
					break
				}
			}
			require.NotNil(t, planet, "Planet ID %d should exist", exp.PlanetID)

			// Validate planet name from XY file
			actualName := planetNames[planet.PlanetNumber]
			assert.Equal(t, exp.Name, actualName, "Planet name should match")

			// Validate owner
			assert.Equal(t, exp.Owner, planet.Owner, "Owner should match")
			assert.Equal(t, exp.IsHomeworld, planet.IsHomeworld, "IsHomeworld should match")

			// Validate environment
			assert.Equal(t, exp.Environment.Gravity, planet.Gravity, "Gravity should match")
			assert.Equal(t, exp.Environment.Temperature, planet.Temperature, "Temperature should match")
			assert.Equal(t, exp.Environment.Radiation, planet.Radiation, "Radiation should match")

			// Validate concentrations
			assert.Equal(t, exp.Concentrations.Ironium, planet.IroniumConc, "Ironium concentration should match")
			assert.Equal(t, exp.Concentrations.Boranium, planet.BoraniumConc, "Boranium concentration should match")
			assert.Equal(t, exp.Concentrations.Germanium, planet.GermaniumConc, "Germanium concentration should match")

			// Validate surface minerals
			assert.Equal(t, exp.SurfaceMinerals.Ironium, planet.Ironium, "Surface ironium should match")
			assert.Equal(t, exp.SurfaceMinerals.Boranium, planet.Boranium, "Surface boranium should match")
			assert.Equal(t, exp.SurfaceMinerals.Germanium, planet.Germanium, "Surface germanium should match")

			// Validate population
			assert.Equal(t, exp.Population, planet.Population, "Population should match")

			// Validate installations
			assert.Equal(t, exp.Installations.Mines, planet.Mines, "Mines should match")
			assert.Equal(t, exp.Installations.Factories, planet.Factories, "Factories should match")
			assert.Equal(t, exp.Installations.Defenses, planet.Defenses, "Defenses should match")

			// Validate scanner and starbase
			assert.Equal(t, exp.ScannerID, planet.ScannerID, "ScannerID should match")
			assert.Equal(t, exp.HasStarbase, planet.HasStarbase, "HasStarbase should match")
			if exp.HasStarbase {
				assert.Equal(t, exp.StarbaseDesign, planet.StarbaseDesign, "StarbaseDesign should match")
			}
		})
	}
}

func TestScenarioPlanetData_ScannedPlanet(t *testing.T) {
	expected := loadPlanetExpected(t)
	_, blockList := loadPlanetFile(t, "game.m2")
	planetNames := loadPlanetNames(t)

	// Collect scanned planets (PartialPlanetBlock Type 14)
	var scannedPlanets []blocks.PartialPlanetBlock
	for _, block := range blockList {
		if p, ok := block.(blocks.PartialPlanetBlock); ok {
			scannedPlanets = append(scannedPlanets, p)
		}
	}

	require.GreaterOrEqual(t, len(scannedPlanets), len(expected.ScannedPlanets),
		"Should have at least %d scanned planets", len(expected.ScannedPlanets))

	// Validate each expected scanned planet
	for _, exp := range expected.ScannedPlanets {
		t.Run(exp.Name, func(t *testing.T) {
			// Find planet by ID
			var planet *blocks.PartialPlanetBlock
			for i := range scannedPlanets {
				if scannedPlanets[i].PlanetNumber == exp.PlanetID {
					planet = &scannedPlanets[i]
					break
				}
			}
			require.NotNil(t, planet, "Planet ID %d should exist", exp.PlanetID)

			// Validate planet name from XY file
			actualName := planetNames[planet.PlanetNumber]
			assert.Equal(t, exp.Name, actualName, "Planet name should match")

			// Validate owner
			assert.Equal(t, exp.Owner, planet.Owner, "Owner should match")

			// Validate environment
			assert.Equal(t, exp.Environment.Gravity, planet.Gravity, "Gravity should match")
			assert.Equal(t, exp.Environment.Temperature, planet.Temperature, "Temperature should match")
			assert.Equal(t, exp.Environment.Radiation, planet.Radiation, "Radiation should match")

			// Validate concentrations
			assert.Equal(t, exp.Concentrations.Ironium, planet.IroniumConc, "Ironium concentration should match")
			assert.Equal(t, exp.Concentrations.Boranium, planet.BoraniumConc, "Boranium concentration should match")
			assert.Equal(t, exp.Concentrations.Germanium, planet.GermaniumConc, "Germanium concentration should match")

			// Validate surface minerals if present
			if exp.SurfaceMinerals != nil {
				assert.Equal(t, exp.SurfaceMinerals.Ironium, planet.Ironium, "Surface ironium should match")
				assert.Equal(t, exp.SurfaceMinerals.Boranium, planet.Boranium, "Surface boranium should match")
				assert.Equal(t, exp.SurfaceMinerals.Germanium, planet.Germanium, "Surface germanium should match")
			}
		})
	}
}
