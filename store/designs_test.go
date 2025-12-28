package store_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
