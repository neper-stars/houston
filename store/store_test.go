package store_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/neper-stars/houston/store"
)

func TestGameStore_AddFile(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-orders/fleetnames/results/game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("game.m1", data)
	require.NoError(t, err)

	assert.Equal(t, 1, gs.SourceCount())
	assert.NotZero(t, gs.GameID)
}

func TestGameStore_Fleets(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-orders/fleetnames/results/game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("game.m1", data)
	require.NoError(t, err)

	// Should have fleets
	fleets := gs.AllFleets()
	require.NotEmpty(t, fleets)

	// Find the custom-named fleet
	var customNamedFleet *store.FleetEntity
	for _, f := range fleets {
		if f.HasCustomName {
			customNamedFleet = f
			break
		}
	}

	require.NotNil(t, customNamedFleet, "should find a fleet with custom name")
	assert.Equal(t, "Scoutty", customNamedFleet.CustomName)
	assert.Equal(t, "Scoutty", customNamedFleet.Name())
}

func TestGameStore_Designs(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-orders/fleetnames/results/game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("game.m1", data)
	require.NoError(t, err)

	// Should have designs
	designs := gs.AllDesigns()
	require.NotEmpty(t, designs)

	// All designs should have names
	for _, d := range designs {
		assert.NotEmpty(t, d.Name, "design should have a name")
	}
}

func TestGameStore_FleetWithDesignName(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-orders/fleetnames/results/game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("game.m1", data)
	require.NoError(t, err)

	fleets := gs.AllFleets()

	// Find a fleet without custom name
	var defaultNamedFleet *store.FleetEntity
	for _, f := range fleets {
		if !f.HasCustomName {
			defaultNamedFleet = f
			break
		}
	}

	require.NotNil(t, defaultNamedFleet, "should find a fleet with default name")

	// Name should include the design name and fleet number
	name := defaultNamedFleet.Name()
	assert.Contains(t, name, "#")

	// If design was resolved, name should contain design name
	if defaultNamedFleet.PrimaryDesign != nil {
		assert.Contains(t, name, defaultNamedFleet.PrimaryDesign.Name)
	}
}

func TestGameStore_CargoManipulation(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-orders/fleetnames/results/game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("game.m1", data)
	require.NoError(t, err)

	fleets := gs.AllFleets()
	require.NotEmpty(t, fleets)

	fleet := fleets[0]

	// Test named struct style
	fleet.SetCargo(store.Cargo{Ironium: 100, Boranium: 50, Germanium: 75})
	cargo := fleet.GetCargo()
	assert.Equal(t, int64(100), cargo.Ironium)
	assert.Equal(t, int64(50), cargo.Boranium)
	assert.Equal(t, int64(75), cargo.Germanium)
	assert.True(t, fleet.Meta().Dirty)

	// Reset for next test
	fleet.Meta().Dirty = false

	// Test fluent builder style
	fleet.Cargo().
		Set(store.Ironium, 200).
		Set(store.Germanium, 150).
		Apply()

	cargo = fleet.GetCargo()
	assert.Equal(t, int64(200), cargo.Ironium)
	assert.Equal(t, int64(50), cargo.Boranium) // Unchanged from before
	assert.Equal(t, int64(150), cargo.Germanium)
	assert.True(t, fleet.Meta().Dirty)
}

func TestGameStore_HasChanges(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-orders/fleetnames/results/game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("game.m1", data)
	require.NoError(t, err)

	// Initially no changes
	assert.False(t, gs.HasChanges())

	// Modify a fleet
	fleets := gs.AllFleets()
	require.NotEmpty(t, fleets)
	fleets[0].SetCargo(store.Cargo{Ironium: 100})

	// Now has changes
	assert.True(t, gs.HasChanges())

	// Reset dirty flags
	gs.ResetDirtyFlags()
	assert.False(t, gs.HasChanges())
}

func TestGameStore_Planets(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-orders/fleetnames/results/game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("game.m1", data)
	require.NoError(t, err)

	// Should have planets
	planets := gs.AllPlanets()
	require.NotEmpty(t, planets)

	// Check planet has data
	planet := planets[0]
	assert.GreaterOrEqual(t, planet.PlanetNumber, 0)

	// Planet names should be resolved from PlanetsBlock
	// (if the file contains a PlanetsBlock)
	if gs.PlanetName(planet.PlanetNumber) != "" {
		assert.NotEmpty(t, planet.Name)
	}
}

func TestGameStore_MultipleFiles(t *testing.T) {
	// Load two M files from the same game (battle scenario has two sides)
	data1, err := os.ReadFile("../testdata/scenario-message/event/battle/side1/game.m1")
	require.NoError(t, err)
	data2, err := os.ReadFile("../testdata/scenario-message/event/battle/side2/game.m2")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("game.m1", data1)
	require.NoError(t, err)
	err = gs.AddFile("game.m2", data2)
	require.NoError(t, err)

	assert.Equal(t, 2, gs.SourceCount())

	// Should have fleets from both players
	fleets := gs.AllFleets()
	assert.NotEmpty(t, fleets)

	// Check we have fleets from different owners (both players have fleets)
	owners := make(map[int]bool)
	for _, f := range fleets {
		owners[f.Owner] = true
	}
	assert.GreaterOrEqual(t, len(owners), 2, "should have fleets from multiple owners")
}

func TestGameStore_GenerateMFile(t *testing.T) {
	// Load an M file
	originalData, err := os.ReadFile("../testdata/scenario-orders/fleetnames/results/game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("game.m1", originalData)
	require.NoError(t, err)

	// Generate a new M file
	regeneratedData, err := gs.GenerateMFile(0) // Player 0 (the m1 file)
	require.NoError(t, err)
	require.NotEmpty(t, regeneratedData)

	// The regenerated file should be parseable
	gs2 := store.New()
	err = gs2.AddFile("regenerated.m1", regeneratedData)
	require.NoError(t, err)

	// Should have the same number of fleets
	assert.Equal(t, len(gs.AllFleets()), len(gs2.AllFleets()))

	// Should have the same number of planets
	assert.Equal(t, len(gs.AllPlanets()), len(gs2.AllPlanets()))
}

func TestGameStore_RegenerateMFile(t *testing.T) {
	// Load an M file
	originalData, err := os.ReadFile("../testdata/scenario-orders/fleetnames/results/game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("game.m1", originalData)
	require.NoError(t, err)

	// Modify a fleet's cargo
	fleets := gs.AllFleets()
	require.NotEmpty(t, fleets)

	fleet := fleets[0]
	fleet.SetCargo(store.Cargo{Ironium: 999, Boranium: 888, Germanium: 777})

	// Regenerate with changes
	regeneratedData, err := gs.RegenerateMFile(0)
	require.NoError(t, err)
	require.NotEmpty(t, regeneratedData)

	// The regenerated file should be parseable
	gs2 := store.New()
	err = gs2.AddFile("regenerated.m1", regeneratedData)
	require.NoError(t, err)

	// Should still have same number of entities
	assert.Equal(t, len(gs.AllFleets()), len(gs2.AllFleets()))
}

func TestFileWriter_EncodeBlock(t *testing.T) {
	// Test block encoding creates valid block structure
	encoder := store.NewBlockEncoder()

	// Test with some sample data
	testData := []byte{0x01, 0x02, 0x03, 0x04}
	typeID := uint16(17) // PartialFleetBlockType

	encoded := encoder.EncodeBlock(17, testData)

	// Should be header (2 bytes) + data (4 bytes) = 6 bytes
	assert.Equal(t, 6, len(encoded))

	// Verify header: type (6 bits) << 10 | size (10 bits)
	// Type 17 << 10 = 17408, size 4 -> 17412 = 0x4404
	header := uint16(encoded[0]) | (uint16(encoded[1]) << 8)
	extractedType := header >> 10
	extractedSize := header & 0x3FF

	assert.Equal(t, typeID, extractedType)
	assert.Equal(t, uint16(4), extractedSize)
}

func TestBlockEncoder_EncodeFleet_RoundTrip(t *testing.T) {
	// Load an M file
	data, err := os.ReadFile("../testdata/scenario-orders/fleetnames/results/game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("game.m1", data)
	require.NoError(t, err)

	fleets := gs.AllFleets()
	require.NotEmpty(t, fleets)

	// Find a fleet with cargo capability
	encoder := store.NewBlockEncoder()
	for _, fleet := range fleets {
		encoded, err := encoder.EncodeFleetBlock(fleet)
		if err != nil {
			continue // Skip fleets without raw block data
		}

		// The encoded data should not be empty
		rawBlocks := fleet.RawBlocks()
		if len(rawBlocks) > 0 {
			// Check that we can encode without error
			assert.NotEmpty(t, encoded)
		}
		break
	}
}

func TestMFileRoundTrip_ByteForByte(t *testing.T) {
	// Test files that should produce identical output
	testFiles := []struct {
		path        string
		playerIndex int
	}{
		{"../testdata/Game.m1", 0},
		{"../testdata/scenario-basic/game.m1", 0},
		{"../testdata/scenario-basic/game.m2", 1},
		{"../testdata/scenario-orders/fleetnames/results/game.m1", 0},
		{"../testdata/scenario-diplomacy/1/side1/game.m1", 0},
		{"../testdata/scenario-diplomacy/1/side2/game.m2", 1},
		{"../testdata/scenario-diplomacy-3way/1/side3/game.m3", 2},
		{"../testdata/scenario-message/event/battle/side1/game.m1", 0},
		{"../testdata/scenario-minefield/game.m1", 0},
		{"../testdata/scenario-wormhole/game.m1", 0},
	}

	for _, tc := range testFiles {
		t.Run(tc.path, func(t *testing.T) {
			original, err := os.ReadFile(tc.path)
			require.NoError(t, err)

			gs := store.New()
			err = gs.AddFile("game.m"+string('1'+rune(tc.playerIndex)), original)
			require.NoError(t, err)

			regenerated, err := gs.GenerateMFile(tc.playerIndex)
			require.NoError(t, err)

			assert.Equal(t, original, regenerated,
				"regenerated file should match original byte-for-byte")
		})
	}
}

func TestXFileRoundTrip_ByteForByte(t *testing.T) {
	// Test X file round-trip using X file as source
	testFiles := []struct {
		xFilePath   string
		playerIndex int
	}{
		{"../testdata/scenario-orders/fleetnames/orders/game.x1", 0},
	}

	for _, tc := range testFiles {
		t.Run(tc.xFilePath, func(t *testing.T) {
			// Load original X file
			originalX, err := os.ReadFile(tc.xFilePath)
			require.NoError(t, err)

			gs := store.New()
			// Add the X file as source
			err = gs.AddFile("game.x1", originalX)
			require.NoError(t, err)

			regenerated, err := gs.GenerateXFile(tc.playerIndex)
			require.NoError(t, err)

			// Check byte-for-byte match
			assert.Equal(t, originalX, regenerated,
				"regenerated X file should match original byte-for-byte")
		})
	}
}

func TestMFileRoundTrip_AllTestdata(t *testing.T) {
	// Walk all M files in testdata and verify round-trip
	err := filepath.Walk("../testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !strings.HasPrefix(ext, ".m") {
			return nil
		}
		// Parse player number from extension (.m1 -> 1, .m16 -> 16)
		playerNum := 0
		if _, err := fmt.Sscanf(ext, ".m%d", &playerNum); err != nil || playerNum < 1 || playerNum > 16 {
			return nil
		}

		t.Run(path, func(t *testing.T) {
			original, err := os.ReadFile(path)
			require.NoError(t, err)

			gs := store.New()
			err = gs.AddFile(filepath.Base(path), original)
			if err != nil {
				t.Skipf("cannot parse file: %v", err)
				return
			}

			// Player index is 0-based (.m1 -> 0, .m16 -> 15)
			playerIndex := playerNum - 1

			regenerated, err := gs.GenerateMFile(playerIndex)
			require.NoError(t, err)

			assert.Equal(t, original, regenerated,
				"regenerated file should match original byte-for-byte")
		})
		return nil
	})
	require.NoError(t, err)
}

func TestXFileRoundTrip_AllTestdata(t *testing.T) {
	// Walk all X files in testdata and verify round-trip
	err := filepath.Walk("../testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !strings.HasPrefix(ext, ".x") {
			return nil
		}
		// Skip .xy files (universe files)
		if ext == ".xy" {
			return nil
		}
		// Parse player number from extension (.x1 -> 1, .x16 -> 16)
		playerNum := 0
		if _, err := fmt.Sscanf(ext, ".x%d", &playerNum); err != nil || playerNum < 1 || playerNum > 16 {
			return nil
		}

		t.Run(path, func(t *testing.T) {
			original, err := os.ReadFile(path)
			require.NoError(t, err)

			gs := store.New()
			err = gs.AddFile(filepath.Base(path), original)
			if err != nil {
				t.Skipf("cannot parse file: %v", err)
				return
			}

			// Player index is 0-based (.x1 -> 0, .x16 -> 15)
			playerIndex := playerNum - 1

			regenerated, err := gs.GenerateXFile(playerIndex)
			require.NoError(t, err)

			assert.Equal(t, original, regenerated,
				"regenerated X file should match original byte-for-byte")
		})
		return nil
	})
	require.NoError(t, err)
}

func TestXYFileRoundTrip_AllTestdata(t *testing.T) {
	// Walk all XY files in testdata and verify round-trip
	err := filepath.Walk("../testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".xy" {
			return nil
		}

		t.Run(path, func(t *testing.T) {
			original, err := os.ReadFile(path)
			require.NoError(t, err)

			gs := store.New()
			err = gs.AddFile(filepath.Base(path), original)
			if err != nil {
				t.Skipf("cannot parse file: %v", err)
				return
			}

			regenerated, err := gs.GenerateXYFile()
			require.NoError(t, err)

			assert.Equal(t, original, regenerated,
				"regenerated XY file should match original byte-for-byte")
		})
		return nil
	})
	require.NoError(t, err)
}

func TestHFileRoundTrip_AllTestdata(t *testing.T) {
	// Walk all H files in testdata and verify round-trip
	err := filepath.Walk("../testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !strings.HasPrefix(ext, ".h") {
			return nil
		}
		// Parse player number from extension (.h1 -> 1, .h16 -> 16)
		playerNum := 0
		if _, err := fmt.Sscanf(ext, ".h%d", &playerNum); err != nil || playerNum < 1 || playerNum > 16 {
			return nil
		}

		t.Run(path, func(t *testing.T) {
			original, err := os.ReadFile(path)
			require.NoError(t, err)

			gs := store.New()
			err = gs.AddFile(filepath.Base(path), original)
			if err != nil {
				t.Skipf("cannot parse file: %v", err)
				return
			}

			// Player index is 0-based (.h1 -> 0, .h16 -> 15)
			playerIndex := playerNum - 1

			regenerated, err := gs.GenerateHFile(playerIndex)
			require.NoError(t, err)

			assert.Equal(t, original, regenerated,
				"regenerated H file should match original byte-for-byte")
		})
		return nil
	})
	require.NoError(t, err)
}

func TestFleetWaypoints(t *testing.T) {
	// Use a file that has fleets with waypoints
	data, err := os.ReadFile("../testdata/scenario-orders/waypoint-merge/game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("game.m1", data)
	require.NoError(t, err)

	// Find fleets with waypoints
	fleetsWithWaypoints := 0
	for _, fleet := range gs.AllFleets() {
		if len(fleet.Waypoints) > 0 {
			fleetsWithWaypoints++
			t.Logf("Fleet %d (owner %d) has %d waypoints", fleet.FleetNumber, fleet.Owner, len(fleet.Waypoints))
			for i, wp := range fleet.Waypoints {
				t.Logf("  Waypoint %d: (%d, %d) task=%s", i, wp.X, wp.Y, wp.TaskName())
			}
		}
	}

	// We expect at least some fleets to have waypoints in this scenario
	assert.Greater(t, fleetsWithWaypoints, 0, "expected at least one fleet with waypoints")
}

func TestDirtyFleetRegeneration(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-orders/fleetnames/results/game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("game.m1", data)
	require.NoError(t, err)

	// Find a fleet with cargo
	var fleetWithCargo *store.FleetEntity
	for _, fleet := range gs.AllFleets() {
		cargo := fleet.GetCargo()
		if cargo.Fuel > 0 {
			fleetWithCargo = fleet
			break
		}
	}

	if fleetWithCargo == nil {
		t.Skip("No fleet with fuel cargo found")
	}

	// Record original cargo
	originalCargo := fleetWithCargo.GetCargo()
	t.Logf("Original cargo: Ironium=%d, Boranium=%d, Germanium=%d, Fuel=%d",
		originalCargo.Ironium, originalCargo.Boranium, originalCargo.Germanium, originalCargo.Fuel)

	// Modify the cargo
	fleetWithCargo.SetCargo(store.Cargo{
		Ironium:    originalCargo.Ironium + 100,
		Boranium:   originalCargo.Boranium + 50,
		Germanium:  originalCargo.Germanium + 25,
		Population: originalCargo.Population,
		Fuel:       originalCargo.Fuel,
	})

	assert.True(t, fleetWithCargo.Meta().Dirty, "fleet should be marked dirty after cargo change")
	assert.True(t, gs.HasChanges(), "store should have changes")

	// Generate modified M file
	regenerated, err := gs.GenerateMFile(0)
	require.NoError(t, err)

	// Reload the regenerated file
	gs2 := store.New()
	err = gs2.AddFile("regenerated.m1", regenerated)
	require.NoError(t, err)

	// Find the same fleet and verify cargo was updated
	fleet2, ok := gs2.Fleet(fleetWithCargo.Owner, fleetWithCargo.FleetNumber)
	require.True(t, ok, "should find fleet in regenerated file")

	newCargo := fleet2.GetCargo()
	t.Logf("Regenerated cargo: Ironium=%d, Boranium=%d, Germanium=%d, Fuel=%d",
		newCargo.Ironium, newCargo.Boranium, newCargo.Germanium, newCargo.Fuel)

	// Verify the cargo was modified
	assert.Equal(t, originalCargo.Ironium+100, newCargo.Ironium, "ironium should be updated")
	assert.Equal(t, originalCargo.Boranium+50, newCargo.Boranium, "boranium should be updated")
	assert.Equal(t, originalCargo.Germanium+25, newCargo.Germanium, "germanium should be updated")
}
