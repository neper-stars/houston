package store_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/store"
)

func TestDesignEntity_GetScannerRanges(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-map/joat-spread-fleets/Game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("Game.m1", data)
	require.NoError(t, err)

	tests := []struct {
		name         string
		designNum    int
		wantNormal   int
		wantPen      int
		wantHas      bool
	}{
		{
			name:       "Teamster with Rhino Scanner",
			designNum:  3,
			wantNormal: 50,
			wantPen:    0,
			wantHas:    true,
		},
		{
			name:       "Cotton Picker with Rhino Scanner",
			designNum:  5,
			wantNormal: 50,
			wantPen:    0,
			wantHas:    true,
		},
		{
			name:       "Long Range Scout with Rhino Scanner",
			designNum:  1,
			wantNormal: 50,
			wantPen:    0,
			wantHas:    true,
		},
		{
			name:       "Stalwart Defender with Rhino Scanner",
			designNum:  4,
			wantNormal: 50,
			wantPen:    0,
			wantHas:    true,
		},
		{
			name:       "Nubian test 1 with Dolphin Scanner",
			designNum:  6,
			wantNormal: 220,
			wantPen:    100,
			wantHas:    true,
		},
		{
			name:       "Santa Maria (colony ship, no scanner)",
			designNum:  2,
			wantNormal: 0,
			wantPen:    0,
			wantHas:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			design, ok := gs.Design(0, tt.designNum)
			require.True(t, ok, "design %d should exist", tt.designNum)

			gotNormal, gotPen := design.GetScannerRanges()
			assert.Equal(t, tt.wantNormal, gotNormal, "normal range")
			assert.Equal(t, tt.wantPen, gotPen, "penetrating range")
			assert.Equal(t, tt.wantHas, design.HasScanner(), "has scanner")
		})
	}
}

func TestDesignEntity_GetScannerRanges_EmptyDesign(t *testing.T) {
	// Create a design entity without a design block
	// This tests that the methods handle edge cases gracefully
	gs := store.New()

	// Try to get a non-existent design
	design, ok := gs.Design(0, 99)
	assert.False(t, ok, "design 99 should not exist")
	assert.Nil(t, design)
}

func TestDesignEntity_Hull(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-map/joat-spread-fleets/Game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("Game.m1", data)
	require.NoError(t, err)

	// Long Range Scout - should be Scout hull
	design, ok := gs.Design(0, 1)
	require.True(t, ok)
	hull := design.Hull()
	require.NotNil(t, hull, "hull should be found")
	assert.Equal(t, "Scout", hull.Name)
	assert.Equal(t, 50, hull.FuelCapacity)
	assert.Len(t, hull.Slots, 3) // Scout has 3 slots

	// Santa Maria - should be Colony Ship hull
	design, ok = gs.Design(0, 2)
	require.True(t, ok)
	hull = design.Hull()
	require.NotNil(t, hull)
	assert.Equal(t, "Colony Ship", hull.Name)

	// Teamster - should be Medium Freighter hull
	design, ok = gs.Design(0, 3)
	require.True(t, ok)
	hull = design.Hull()
	require.NotNil(t, hull)
	assert.Equal(t, "Medium Freighter", hull.Name)
	assert.Equal(t, 210, hull.CargoCapacity)
}

func TestDesignEntity_Capabilities(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-map/joat-spread-fleets/Game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("Game.m1", data)
	require.NoError(t, err)

	// Long Range Scout - Scout hull with scanner
	t.Run("Scout capabilities", func(t *testing.T) {
		design, ok := gs.Design(0, 1)
		require.True(t, ok)

		// Should have scanner
		assert.True(t, design.HasScanner())
		normal, pen := design.GetScannerRanges()
		assert.Equal(t, 50, normal)
		assert.Equal(t, 0, pen)

		// Scout hull has 50 fuel capacity
		assert.Equal(t, 50, design.GetFuelCapacity())

		// Should have engine
		engine := design.GetEngine()
		require.NotNil(t, engine)

		// Scout has no cargo
		assert.Equal(t, 0, design.GetCargoCapacity())
	})

	// Santa Maria - Colony Ship
	t.Run("Colony Ship capabilities", func(t *testing.T) {
		design, ok := gs.Design(0, 2)
		require.True(t, ok)

		// Colony ships can colonize
		assert.True(t, design.CanColonize())

		// No scanner on colony ship
		assert.False(t, design.HasScanner())

		// Colony ship hull has 25 cargo capacity
		assert.Equal(t, 25, design.GetCargoCapacity())
	})

	// Teamster - Medium Freighter
	t.Run("Freighter capabilities", func(t *testing.T) {
		design, ok := gs.Design(0, 3)
		require.True(t, ok)

		// Medium Freighter has 210 cargo
		assert.Equal(t, 210, design.GetCargoCapacity())

		// Has scanner
		assert.True(t, design.HasScanner())
	})

	// Stalwart Defender - Destroyer
	t.Run("Destroyer capabilities", func(t *testing.T) {
		design, ok := gs.Design(0, 4)
		require.True(t, ok)

		// Should have minesweep capability (has beam weapons)
		sweepRate := design.GetMinesweepRate()
		assert.Greater(t, sweepRate, 0, "Destroyer should be able to sweep mines")

		// Should have shields
		shields := design.GetTotalShieldValue()
		assert.Greater(t, shields, 0, "Destroyer should have shields")

		// Should have armor (hull + equipped)
		armor := design.GetTotalArmorValue()
		assert.Greater(t, armor, 0, "Destroyer should have armor")
	})

	// Test EquippedItems enumeration
	t.Run("EquippedItems", func(t *testing.T) {
		design, ok := gs.Design(0, 1) // Scout
		require.True(t, ok)

		items := design.EquippedItems()
		assert.NotEmpty(t, items, "Scout should have equipped items")

		// Should have at least engine and scanner
		hasEngine := false
		hasScanner := false
		for _, item := range items {
			if item.Category == blocks.ItemCategoryEngine {
				hasEngine = true
			}
			if item.Category == blocks.ItemCategoryScanner {
				hasScanner = true
			}
		}
		assert.True(t, hasEngine, "Scout should have engine")
		assert.True(t, hasScanner, "Scout should have scanner")
	})
}
