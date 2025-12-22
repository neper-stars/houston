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

// Expected data structures for mass driver destination scenario
type ExpectedMassDriverDest struct {
	PlanetName        string `json:"planetName"`
	PlanetNumber      int    `json:"planetNumber"`
	StarbaseDesign    string `json:"starbaseDesign"`
	DestinationName   string `json:"destinationName"`
	DestinationNumber int    `json:"destinationNumber"`
}

type ExpectedMassDriverData struct {
	Scenario              string                   `json:"scenario"`
	MassDriverDestinations []ExpectedMassDriverDest `json:"massDriverDestinations"`
}

func loadMassDriverExpected(t *testing.T) *ExpectedMassDriverData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-packet-driver-destination", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedMassDriverData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadMassDriverFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-packet-driver-destination", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

func TestScenarioMassDriver_Destinations(t *testing.T) {
	expected := loadMassDriverExpected(t)
	_, blockList := loadMassDriverFile(t, "game.m1")

	// Load planet names from XY file
	xyPath := filepath.Join("..", "testdata", "scenario-packet-driver-destination", "game.xy")
	xyData, err := os.ReadFile(xyPath)
	require.NoError(t, err)
	xyFd := FileData(xyData)
	xyBlocks, err := xyFd.BlockList()
	require.NoError(t, err)

	planetNames := make(map[int]string)
	for _, block := range xyBlocks {
		if pb, ok := block.(blocks.PlanetsBlock); ok {
			for _, p := range pb.Planets {
				planetNames[p.ID] = data.PlanetNames[p.NameID]
			}
		}
	}

	// Build starbase design name map
	starbaseDesigns := make(map[int]string)
	for _, block := range blockList {
		if db, ok := block.(blocks.DesignBlock); ok {
			if db.IsStarbase {
				starbaseDesigns[db.DesignNumber] = db.Name
			}
		}
	}

	// Collect planets with mass driver destinations
	type massDriverInfo struct {
		planetNumber      int
		planetName        string
		starbaseDesign    string
		destinationNumber int
		destinationName   string
	}
	var massDrivers []massDriverInfo

	for _, block := range blockList {
		if pb, ok := block.(blocks.PlanetBlock); ok {
			if pb.HasStarbase && pb.MassDriverDest > 0 {
				destID := pb.MassDriverDest - 1 // Convert display ID to internal ID
				massDrivers = append(massDrivers, massDriverInfo{
					planetNumber:      pb.PlanetNumber,
					planetName:        planetNames[pb.PlanetNumber],
					starbaseDesign:    starbaseDesigns[pb.StarbaseDesign],
					destinationNumber: destID,
					destinationName:   planetNames[destID],
				})
			}
		}
	}

	require.Equal(t, len(expected.MassDriverDestinations), len(massDrivers),
		"Should have %d mass driver destinations", len(expected.MassDriverDestinations))

	// Validate each mass driver destination
	for i, exp := range expected.MassDriverDestinations {
		md := massDrivers[i]
		t.Run(exp.PlanetName, func(t *testing.T) {
			// Validate planet
			assert.Equal(t, exp.PlanetNumber, md.planetNumber, "Planet number should match")
			assert.Equal(t, exp.PlanetName, md.planetName, "Planet name should match")

			// Validate starbase design
			assert.Equal(t, exp.StarbaseDesign, md.starbaseDesign, "Starbase design should match")

			// Validate destination
			assert.Equal(t, exp.DestinationNumber, md.destinationNumber, "Destination number should match")
			assert.Equal(t, exp.DestinationName, md.destinationName, "Destination name should match")
		})
	}
}
