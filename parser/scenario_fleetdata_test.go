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

// Expected data structures for fleet scenario
type ExpectedCargo struct {
	Ironium   int `json:"ironium"`
	Boranium  int `json:"boranium"`
	Germanium int `json:"germanium"`
	Colonists int `json:"colonists"`
}

type ExpectedOwnFleet struct {
	Name          string        `json:"name"`
	FleetID       int           `json:"fleetId"`
	X             int           `json:"x"`
	Y             int           `json:"y"`
	ShipCount     int           `json:"shipCount"`
	Cargo         ExpectedCargo `json:"cargo"`
	Fuel          int           `json:"fuel"`
	WaypointCount int           `json:"waypointCount"`
}

type ExpectedEnemyFleet struct {
	FleetID   int    `json:"fleetId"`
	Owner     string `json:"owner"`
	Hull      string `json:"hull"`
	ShipCount int    `json:"shipCount"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	WarpSpeed int    `json:"warpSpeed"`
	EstMass   int    `json:"estMass"`
}

type ExpectedFleetData struct {
	Scenario    string               `json:"scenario"`
	OwnFleets   []ExpectedOwnFleet   `json:"ownFleets"`
	EnemyFleets []ExpectedEnemyFleet `json:"enemyFleets"`
}

func loadFleetExpected(t *testing.T) *ExpectedFleetData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-fleetdata", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedFleetData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadFleetFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-fleetdata", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

func TestScenarioFleetData_OwnFleet(t *testing.T) {
	expected := loadFleetExpected(t)
	_, blockList := loadFleetFile(t, "game.m2")

	// Find the file owner (player index from header)
	var header *blocks.FileHeader
	for _, block := range blockList {
		if h, ok := block.(blocks.FileHeader); ok {
			header = &h
			break
		}
	}
	require.NotNil(t, header, "FileHeader should exist")
	ownerPlayerIndex := header.PlayerIndex()

	// Collect own fleets (FleetBlock where owner matches file owner)
	var ownFleets []blocks.FleetBlock
	for _, block := range blockList {
		if f, ok := block.(blocks.FleetBlock); ok {
			if f.Owner == ownerPlayerIndex {
				ownFleets = append(ownFleets, f)
			}
		}
	}

	require.GreaterOrEqual(t, len(ownFleets), len(expected.OwnFleets),
		"Should have at least %d own fleets", len(expected.OwnFleets))

	// Validate each expected own fleet
	for _, exp := range expected.OwnFleets {
		t.Run(exp.Name, func(t *testing.T) {
			// Find fleet by position (most reliable match)
			var fleet *blocks.FleetBlock
			for i := range ownFleets {
				if ownFleets[i].X == exp.X && ownFleets[i].Y == exp.Y {
					fleet = &ownFleets[i]
					break
				}
			}
			require.NotNil(t, fleet, "Fleet at position (%d, %d) should exist", exp.X, exp.Y)

			// Validate position
			assert.Equal(t, exp.X, fleet.X, "X position should match")
			assert.Equal(t, exp.Y, fleet.Y, "Y position should match")

			// Validate ship count
			assert.Equal(t, exp.ShipCount, fleet.TotalShips(), "Ship count should match")

			// Validate cargo
			assert.Equal(t, int64(exp.Cargo.Ironium), fleet.Ironium, "Ironium should match")
			assert.Equal(t, int64(exp.Cargo.Boranium), fleet.Boranium, "Boranium should match")
			assert.Equal(t, int64(exp.Cargo.Germanium), fleet.Germanium, "Germanium should match")
			assert.Equal(t, int64(exp.Cargo.Colonists), fleet.Population, "Colonists should match")

			// Validate fuel
			assert.Equal(t, int64(exp.Fuel), fleet.Fuel, "Fuel should match")

			// Validate waypoint count (stored in FleetBlock for full fleets)
			assert.Equal(t, exp.WaypointCount, fleet.WaypointCount, "Waypoint count should match")
		})
	}
}

func TestScenarioFleetData_EnemyFleet(t *testing.T) {
	expected := loadFleetExpected(t)
	_, blockList := loadFleetFile(t, "game.m2")

	// Find the file owner (player index from header)
	var header *blocks.FileHeader
	for _, block := range blockList {
		if h, ok := block.(blocks.FileHeader); ok {
			header = &h
			break
		}
	}
	require.NotNil(t, header, "FileHeader should exist")
	ownerPlayerIndex := header.PlayerIndex()

	// Build player name to ID map
	playerNameToID := make(map[string]int)
	for _, block := range blockList {
		if p, ok := block.(blocks.PlayerBlock); ok {
			playerNameToID[p.NameSingular] = p.PlayerNumber
			playerNameToID[p.NamePlural] = p.PlayerNumber
		}
	}

	// Collect enemy fleets (PartialFleetBlock where owner != file owner)
	var enemyFleets []blocks.PartialFleetBlock
	for _, block := range blockList {
		if f, ok := block.(blocks.PartialFleetBlock); ok {
			if f.Owner != ownerPlayerIndex {
				enemyFleets = append(enemyFleets, f)
			}
		}
	}

	require.GreaterOrEqual(t, len(enemyFleets), len(expected.EnemyFleets),
		"Should have at least %d enemy fleets", len(expected.EnemyFleets))

	// Validate each expected enemy fleet
	for _, exp := range expected.EnemyFleets {
		t.Run(exp.Owner, func(t *testing.T) {
			// Find fleet by position
			var fleet *blocks.PartialFleetBlock
			for i := range enemyFleets {
				if enemyFleets[i].X == exp.X && enemyFleets[i].Y == exp.Y {
					fleet = &enemyFleets[i]
					break
				}
			}
			require.NotNil(t, fleet, "Enemy fleet at position (%d, %d) should exist", exp.X, exp.Y)

			// Validate owner by race name
			expectedOwnerID, found := playerNameToID[exp.Owner]
			require.True(t, found, "Owner race '%s' should exist in player list", exp.Owner)
			assert.Equal(t, expectedOwnerID, fleet.Owner, "Owner should match")

			// Validate position
			assert.Equal(t, exp.X, fleet.X, "X position should match")
			assert.Equal(t, exp.Y, fleet.Y, "Y position should match")

			// Validate warp speed
			assert.Equal(t, exp.WarpSpeed, fleet.Warp, "Warp speed should match")

			// Validate estimated mass
			assert.Equal(t, int64(exp.EstMass), fleet.Mass, "Estimated mass should match")

			// Validate ship count
			assert.Equal(t, exp.ShipCount, fleet.TotalShips(), "Ship count should match")

			// Validate hull type if specified
			if exp.Hull != "" {
				expectedHullID, hullFound := data.HullNameToID[exp.Hull]
				require.True(t, hullFound, "Hull '%s' should be valid", exp.Hull)

				// For partial fleets, we need to look up the design from the enemy player's designs
				// The ShipTypes bitmask tells us which design slot is used
				var designHullID int
				for _, block := range blockList {
					if d, ok := block.(blocks.DesignBlock); ok {
						// Check if this design is used by the fleet
						designBit := 1 << d.DesignNumber
						if (int(fleet.ShipTypes) & designBit) != 0 {
							designHullID = d.HullId
							break
						}
					}
				}
				assert.Equal(t, expectedHullID, designHullID,
					"Hull should be '%s' (ID %d)", exp.Hull, expectedHullID)
			}
		})
	}
}

func TestScenarioFleetData_FileHeader(t *testing.T) {
	_, blockList := loadFleetFile(t, "game.m2")

	var header *blocks.FileHeader
	for _, block := range blockList {
		if h, ok := block.(blocks.FileHeader); ok {
			header = &h
			break
		}
	}

	require.NotNil(t, header, "FileHeader should exist")
	assert.Equal(t, 1, header.PlayerIndex(), "Should be player 2's file (index 1)")
	assert.Equal(t, uint16(8), header.Turn, "Should be turn 8")
	assert.Equal(t, 2408, header.Year(), "Year should be 2408")
}
