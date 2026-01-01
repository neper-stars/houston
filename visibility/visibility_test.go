package visibility_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/store"
	"github.com/neper-stars/houston/visibility"
)

func TestCloakingScenario_Turn2430to2435(t *testing.T) {
	// Test scenario from testdata/scenario-cloaking-visibility/game01/README.md
	// Player 1 (index 0): Super Stealth (SS) - has cloaked fleet "Smaugarian Peeping Tom #7"
	// Player 2 (index 1): Alternate Reality (AR) - has starbase with scanner
	//
	// Turn 2430-2434: Player 2 does NOT see the fleet
	// Turn 2435: Player 2 finally sees the fleet

	basePath := "../testdata/scenario-cloaking-visibility/game01/historic-backup"

	// Test each turn
	testCases := []struct {
		turn        int
		shouldSee   bool
		description string
	}{
		{2430, false, "Fleet entering perceived range - should NOT be visible"},
		{2431, false, "Fleet approaching - should NOT be visible"},
		{2432, false, "Fleet closer - should NOT be visible"},
		{2433, false, "Fleet even closer - should NOT be visible"},
		{2434, false, "Fleet very close - should NOT be visible"},
		{2435, true, "Fleet finally visible!"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Turn_%d", tc.turn), func(t *testing.T) {
			gs := store.New()

			// Load XY file first (contains planet coordinates)
			xyPath := fmt.Sprintf("%s/game-%d.xy", basePath, tc.turn)
			xyData, err := os.ReadFile(xyPath)
			require.NoError(t, err, "Failed to read xy file")

			err = gs.AddFile(fmt.Sprintf("game-%d.xy", tc.turn), xyData)
			require.NoError(t, err, "Failed to add xy file")

			// Load both player files
			m1Path := fmt.Sprintf("%s/game-%d.m1", basePath, tc.turn)
			m2Path := fmt.Sprintf("%s/game-%d.m2", basePath, tc.turn)

			m1Data, err := os.ReadFile(m1Path)
			require.NoError(t, err, "Failed to read m1 file")

			m2Data, err := os.ReadFile(m2Path)
			require.NoError(t, err, "Failed to read m2 file")

			err = gs.AddFile(fmt.Sprintf("game-%d.m1", tc.turn), m1Data)
			require.NoError(t, err, "Failed to add m1 file")

			err = gs.AddFile(fmt.Sprintf("game-%d.m2", tc.turn), m2Data)
			require.NoError(t, err, "Failed to add m2 file")

			// Get players and verify PRTs
			player1, ok := gs.Player(0)
			require.True(t, ok, "Player 1 should exist")
			t.Logf("Player 1: %s (PRT=%d=%s)", player1.NamePlural, player1.PRT, blocks.PRTName(player1.PRT))

			player2, ok := gs.Player(1)
			require.True(t, ok, "Player 2 should exist")
			t.Logf("Player 2: %s (PRT=%d=%s)", player2.NamePlural, player2.PRT, blocks.PRTName(player2.PRT))

			// Find the "Smaugarian Peeping Tom #7" fleet (owned by player 0)
			var targetFleet *store.FleetEntity
			for _, fleet := range gs.FleetsByOwner(0) {
				if fleet.Name() == "Smaugarian Peeping Tom #7" {
					targetFleet = fleet
					break
				}
			}
			require.NotNil(t, targetFleet, "Target fleet 'Smaugarian Peeping Tom #7' should exist")

			t.Logf("Target fleet at (%d, %d), cloaking: %.2f%%",
				targetFleet.X, targetFleet.Y, visibility.FleetCloaking(targetFleet, gs)*100)

			// Check if ANY of Player 2's scanners can see the target fleet
			canSeeTarget := false
			var detectionSource string

			// Check all planets with starbases
			for _, planet := range gs.PlanetsByOwner(1) {
				if !planet.HasStarbase {
					continue
				}
				result := visibility.GetPlanetDetectionDetails(planet, targetFleet, gs)
				if result.CanSee {
					canSeeTarget = true
					detectionSource = fmt.Sprintf("planet %s (dist=%.2f, effectiveRange=%.2f)",
						planet.Name, result.Distance, result.EffectiveNormalRange)
					break
				}
			}

			// Check all fleets if not already detected
			if !canSeeTarget {
				for _, fleet := range gs.FleetsByOwner(1) {
					result := visibility.GetDetectionDetails(fleet, targetFleet, gs)
					if result.CanSee {
						canSeeTarget = true
						detectionSource = fmt.Sprintf("fleet %s (dist=%.2f, effectiveRange=%.2f)",
							fleet.Name(), result.Distance, result.EffectiveNormalRange)
						break
					}
				}
			}

			if canSeeTarget {
				t.Logf("Target fleet DETECTED by %s", detectionSource)
			} else {
				t.Logf("Target fleet NOT detected by any scanner")
			}

			// Verify expected visibility for "Smaugarian Peeping Tom #7"
			assert.Equal(t, tc.shouldSee, canSeeTarget, tc.description)
		})
	}
}

func TestCloakPerKTCurve_KnownValues(t *testing.T) {
	// Test known values from the Stars! spreadsheet
	tests := []struct {
		cloakPerKT float64
		expected   float64
	}{
		{0, 0},
		{100, 0.5},
		{300, 0.75},
		{600, 0.875},
		{1000, 0.9375},
	}

	for _, tt := range tests {
		got := visibility.CloakPerKTToPercent(tt.cloakPerKT)
		assert.InDelta(t, tt.expected, got, 0.001,
			"CloakPerKTToPercent(%v) should be %v", tt.cloakPerKT, tt.expected)
	}
}
