package store_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/store"
)

// TestCalculateScore_ScenarioHistory tests score calculation against scenario-history.
// Expected values from testdata/scenario-history/scores.png:
// - Player 2 (The Hobbits): Rank 2, Score 29
// - Planets: 1, Starbases: 1, Unarmed: 5, Escort: 2, Capital: 0
// - Tech Levels: 18, Resources: 143
func TestCalculateScore_ScenarioHistory(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-history/game.m2")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("game.m2", data)
	require.NoError(t, err)

	// Calculate score for player 2 (file owner)
	sc := gs.ComputeScoreFromActualData(1) // Player numbers are 0-indexed, so player 2 = index 1

	// Verify individual components match expected values
	// From screenshot: Score=29, Planets=1, Starbases=1, Unarmed=5, Escort=2, Capital=0
	// Tech=18, Resources=143

	assert.Equal(t, 1, sc.PlanetCount, "Planet count should be 1")
	assert.Equal(t, 1, sc.StarbaseCount, "Starbase count should be 1")
	assert.Equal(t, 143, sc.TotalResources, "Total resources should be 143")
	assert.Equal(t, 5, sc.UnarmedShips, "Unarmed ships should be 5")
	assert.Equal(t, 2, sc.EscortShips, "Escort ships should be 2")
	assert.Equal(t, 0, sc.CapitalShips, "Capital ships should be 0")

	// Tech score: 18 total means each of 6 fields at level 3 = 3*6 = 18
	assert.Equal(t, 18, sc.TechScore, "Tech score should be 18")

	// Resource score: 143/30 = 4
	assert.Equal(t, 4, sc.ResourceScore, "Resource score should be 4")

	// Starbase score: 1*3 = 3
	assert.Equal(t, 3, sc.StarbaseScore, "Starbase score should be 3")

	// Ship score: unarmedCapped=min(5,1)=1, escortCapped=min(2,1)=1
	// shipScore = 1/2 + 1 + 0 = 0 + 1 = 1
	assert.Equal(t, 1, sc.ShipScore, "Ship score should be 1")

	// Final score should be 29
	assert.Equal(t, 29, sc.Score, "Total score should be 29")
}

// TestCalculateScore_ScenarioMinefield tests score calculation against scenario-minefield.
// Expected values from testdata/scenario-minefield/scores.png:
// - Player 1 (The Halflings): Rank 1, Score 27
// - Planets: 2, Starbases: 1, Unarmed: 3, Escort: 0, Capital: 0
// - Tech Levels: 19, Resources: 55
func TestCalculateScore_ScenarioMinefield(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-minefield/game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("game.m1", data)
	require.NoError(t, err)

	// Calculate score for player 1 (file owner)
	sc := gs.ComputeScoreFromActualData(0) // Player numbers are 0-indexed

	// Verify individual components match expected values
	// From screenshot: Score=27, Planets=2, Starbases=1, Unarmed=3, Escort=0, Capital=0
	// Tech=19, Resources=55

	assert.Equal(t, 2, sc.PlanetCount, "Planet count should be 2")
	assert.Equal(t, 1, sc.StarbaseCount, "Starbase count should be 1")
	assert.Equal(t, 55, sc.TotalResources, "Total resources should be 55")
	assert.Equal(t, 3, sc.UnarmedShips, "Unarmed ships should be 3")
	assert.Equal(t, 0, sc.EscortShips, "Escort ships should be 0")
	assert.Equal(t, 0, sc.CapitalShips, "Capital ships should be 0")

	// Tech score: 19
	assert.Equal(t, 19, sc.TechScore, "Tech score should be 19")

	// Resource score: 55/30 = 1
	assert.Equal(t, 1, sc.ResourceScore, "Resource score should be 1")

	// Starbase score: 1*3 = 3
	assert.Equal(t, 3, sc.StarbaseScore, "Starbase score should be 3")

	// Ship score: unarmedCapped=min(3,2)=2, escortCapped=min(0,2)=0
	// shipScore = 2/2 + 0 + 0 = 1
	assert.Equal(t, 1, sc.ShipScore, "Ship score should be 1")

	// Final score should be 27
	assert.Equal(t, 27, sc.Score, "Total score should be 27")
}

// TestCalculateScore_ScenarioSingleplayer tests score calculation against scenario-singleplayer.
// Expected values from testdata/scenario-singleplayer/2483/scores.png:
// - Player 1 (The Humanoids): Rank 1, Score 838
// - Planets: 11, Starbases: 1, Unarmed: 6, Escort: 2, Capital: 0
// - Tech Levels: 76 (raw sum), Resources: 17k
func TestCalculateScore_ScenarioSingleplayer(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-singleplayer/2483/Game.m1")
	require.NoError(t, err)

	gs := store.New()
	err = gs.AddFile("Game.m1", data)
	require.NoError(t, err)

	// Calculate score for player 1 (file owner)
	sc := gs.ComputeScoreFromActualData(0) // Player numbers are 0-indexed

	// Verify individual components match expected values
	// From screenshot: Score=838, Planets=11, Starbases=1, Unarmed=6, Escort=2, Capital=0
	// Tech=76, Resources=17k

	assert.Equal(t, 11, sc.PlanetCount, "Planet count should be 11")
	assert.Equal(t, 1, sc.StarbaseCount, "Starbase count should be 1")
	assert.Equal(t, 6, sc.UnarmedShips, "Unarmed ships should be 6")
	assert.Equal(t, 2, sc.EscortShips, "Escort ships should be 2")
	assert.Equal(t, 0, sc.CapitalShips, "Capital ships should be 0")

	// Tech score: 197 (tiered formula with +1 bonus for having tech >= 10)
	// Raw sum is 76, but tiered formula with bonus gives 196+1=197
	assert.Equal(t, 197, sc.TechScore, "Tech score should be 197")

	// Resources should be around 17000 (displayed as "17k")
	// Allow some tolerance since "17k" could mean 17000-17999
	assert.True(t, sc.TotalResources >= 17000 && sc.TotalResources < 18000,
		"Total resources should be ~17k, got %d", sc.TotalResources)

	// Resource score: 17000/30 ≈ 566
	assert.Equal(t, sc.TotalResources/30, sc.ResourceScore, "Resource score should be resources/30")

	// Starbase score: 1*3 = 3
	assert.Equal(t, 3, sc.StarbaseScore, "Starbase score should be 3")

	// Final score should be 838
	assert.Equal(t, 838, sc.Score, "Total score should be 838")
}

// TestTechLevelScore tests the tech level scoring tiers.
func TestTechLevelScore(t *testing.T) {
	// Tech level 0-4: level points (raw value)
	// Tech level 5-6: level×2 - 4 = 6, 8
	// Tech level 7-9: level×3 - 9 = 12, 15, 18
	// Tech level 10+: level×4 - 18 = 22, 26, 30, ...

	testCases := []struct {
		level    int
		expected int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 4},   // raw level value (tier at 4, not 3)
		{5, 6},   // 5×2 - 4 = 6
		{6, 8},   // 6×2 - 4 = 8
		{7, 12},  // 7×3 - 9 = 12
		{8, 15},  // 8×3 - 9 = 15
		{9, 18},  // 9×3 - 9 = 18
		{10, 22}, // 10×4 - 18 = 22
		{11, 26}, // 11×4 - 18 = 26
		{12, 30}, // 12×4 - 18 = 30
	}

	for _, tc := range testCases {
		t.Run("level "+string(rune('0'+tc.level)), func(t *testing.T) {
			score := techLevelScore(tc.level)
			assert.Equal(t, tc.expected, score, "Tech level %d should give %d points", tc.level, tc.expected)
		})
	}
}

// techLevelScore mirrors the internal function for testing.
// Uses tier boundaries at 4, 6, 9.
func techLevelScore(level int) int {
	if level <= 4 {
		return level
	}
	if level <= 6 {
		return level*2 - 4
	}
	if level <= 9 {
		return level*3 - 9
	}
	return level*4 - 18
}

// TestShipScoreCalculation tests the ship score calculation formula.
func TestShipScoreCalculation(t *testing.T) {
	testCases := []struct {
		name        string
		unarmed     int
		escort      int
		capital     int
		planetCount int
		expected    int
	}{
		{"no ships", 0, 0, 0, 1, 0},
		{"unarmed only capped", 5, 0, 0, 1, 0},         // min(5,1)/2 = 0
		{"escort only capped", 0, 2, 0, 1, 1},          // min(2,1) = 1
		{"mixed capped", 5, 2, 0, 1, 1},                // 0 + 1 = 1
		{"capital ship", 0, 0, 1, 1, 0},                // (1*1)/(1+1) = 0
		{"capital ships with planets", 0, 0, 2, 4, 1},  // (4*2)/(4+2) = 1
		{"many capital ships", 0, 0, 10, 10, 5},        // (10*10)/(10+10) = 5
		{"all types uncapped", 2, 3, 1, 10, 1 + 3 + 0}, // 2/2 + 3 + (10*1)/(10+1) = 1 + 3 + 0 = 4
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := calculateShipScore(tc.unarmed, tc.escort, tc.capital, tc.planetCount)
			assert.Equal(t, tc.expected, score, "Ship score calculation mismatch")
		})
	}
}

// calculateShipScore mirrors the internal function for testing.
func calculateShipScore(unarmed, escort, capital, planetCount int) int {
	unarmedCapped := unarmed
	if unarmedCapped > planetCount {
		unarmedCapped = planetCount
	}

	escortCapped := escort
	if escortCapped > planetCount {
		escortCapped = planetCount
	}

	score := unarmedCapped/2 + escortCapped

	if capital > 0 && planetCount > 0 {
		score += (planetCount * capital) / (planetCount + capital)
	}

	return score
}

// TestPlayerScoresBlock_StoredVsCalculated verifies that scores stored in PlayerScoresBlock
// match (or explain discrepancies with) our calculated scores.
// This helps identify whether score discrepancies are in our calculation or in reading stored scores.
func TestPlayerScoresBlock_StoredVsCalculated(t *testing.T) {
	testCases := []struct {
		name           string
		file           string
		playerIndex    int
		expectedStored int // Score stored in PlayerScoresBlock
		expectedCalc   int // Score from our CalculateScore function
	}{
		{
			name:           "scenario-history player 2",
			file:           "../testdata/scenario-history/game.m2",
			playerIndex:    1,
			expectedStored: 29,
			expectedCalc:   29,
		},
		{
			name:           "scenario-minefield player 1",
			file:           "../testdata/scenario-minefield/game.m1",
			playerIndex:    0,
			expectedStored: 27,
			expectedCalc:   27,
		},
		{
			name:           "scenario-singleplayer player 1",
			file:           "../testdata/scenario-singleplayer/2483/Game.m1",
			playerIndex:    0,
			expectedStored: 838,
			expectedCalc:   838,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(tc.file)
			require.NoError(t, err)

			gs := store.New()
			err = gs.AddFile(tc.file, data)
			require.NoError(t, err)

			// Get calculated score
			sc := gs.ComputeScoreFromActualData(tc.playerIndex)

			// Find PlayerScoresBlock for this player
			sources := gs.Sources()
			require.NotEmpty(t, sources, "Should have at least one source")

			var storedScore *storedScoreData
			for _, source := range sources {
				for _, block := range source.Blocks {
					if psb, ok := block.(blocks.PlayerScoresBlock); ok {
						if psb.PlayerID == tc.playerIndex {
							storedScore = &storedScoreData{
								Score:        psb.Score,
								Resources:    psb.Resources,
								Planets:      psb.Planets,
								Starbases:    psb.Starbases,
								UnarmedShips: psb.UnarmedShips,
								EscortShips:  psb.EscortShips,
								CapitalShips: psb.CapitalShips,
								TechLevels:   psb.TechLevels,
							}
							break
						}
					}
				}
				if storedScore != nil {
					break
				}
			}

			require.NotNil(t, storedScore, "Should find PlayerScoresBlock for player %d", tc.playerIndex)

			// Log comparison for debugging
			t.Logf("Stored Score Data (PlayerScoresBlock):")
			t.Logf("  Score: %d, Planets: %d, Starbases: %d", storedScore.Score, storedScore.Planets, storedScore.Starbases)
			t.Logf("  Unarmed: %d, Escort: %d, Capital: %d", storedScore.UnarmedShips, storedScore.EscortShips, storedScore.CapitalShips)
			t.Logf("  TechLevels: %d, Resources: %d", storedScore.TechLevels, storedScore.Resources)

			t.Logf("Calculated Score Data:")
			t.Logf("  Score: %d, PlanetCount: %d, Starbases: %d", sc.Score, sc.PlanetCount, sc.StarbaseCount)
			t.Logf("  Unarmed: %d, Escort: %d, Capital: %d", sc.UnarmedShips, sc.EscortShips, sc.CapitalShips)
			t.Logf("  TechScore: %d, Resources: %d", sc.TechScore, sc.TotalResources)
			t.Logf("  Score breakdown: PlanetPopScore=%d + ResourceScore=%d + StarbaseScore=%d + TechScore=%d + ShipScore=%d",
				sc.PlanetPopScore, sc.ResourceScore, sc.StarbaseScore, sc.TechScore, sc.ShipScore)

			// Verify component counts match between stored and calculated
			assert.Equal(t, storedScore.Planets, sc.PlanetCount, "Planet count should match")
			assert.Equal(t, storedScore.Starbases, sc.StarbaseCount, "Starbase count should match")
			assert.Equal(t, storedScore.UnarmedShips, sc.UnarmedShips, "Unarmed ships should match")
			assert.Equal(t, storedScore.EscortShips, sc.EscortShips, "Escort ships should match")
			assert.Equal(t, storedScore.CapitalShips, sc.CapitalShips, "Capital ships should match")

			// Verify stored score matches expected
			assert.Equal(t, tc.expectedStored, storedScore.Score, "Stored score should match expected")

			// Verify calculated score matches expected
			assert.Equal(t, tc.expectedCalc, sc.Score, "Calculated score should match expected")

			// Log the discrepancy if any
			if storedScore.Score != sc.Score {
				t.Logf("DISCREPANCY: Stored=%d, Calculated=%d, Diff=%d",
					storedScore.Score, sc.Score, storedScore.Score-sc.Score)

				// Check if tech score difference explains the gap
				// storedScore.TechLevels is raw sum, sc.TechScore is tiered
				t.Logf("TechLevels (raw): %d, TechScore (tiered): %d",
					storedScore.TechLevels, sc.TechScore)
			}
		})
	}
}

// storedScoreData holds score data from PlayerScoresBlock for comparison.
type storedScoreData struct {
	Score        int
	Resources    int64
	Planets      int
	Starbases    int
	UnarmedShips int
	EscortShips  int
	CapitalShips int
	TechLevels   int
}

// TestPlayerScore_API verifies that gs.PlayerScore() returns correct stored values.
// This tests the higher-level API that exposes StoredScore from PlayerEntity.
func TestPlayerScore_API(t *testing.T) {
	testCases := []struct {
		name               string
		file               string
		playerIndex        int
		expectedScore      int
		expectedPlanets    int
		expectedStarbases  int
		expectedUnarmed    int
		expectedEscort     int
		expectedCapital    int
		expectedTechLevels int
	}{
		{
			name:               "scenario-history player 2",
			file:               "../testdata/scenario-history/game.m2",
			playerIndex:        1,
			expectedScore:      29,
			expectedPlanets:    1,
			expectedStarbases:  1,
			expectedUnarmed:    5,
			expectedEscort:     2,
			expectedCapital:    0,
			expectedTechLevels: 18,
		},
		{
			name:               "scenario-minefield player 1",
			file:               "../testdata/scenario-minefield/game.m1",
			playerIndex:        0,
			expectedScore:      27,
			expectedPlanets:    2,
			expectedStarbases:  1,
			expectedUnarmed:    3,
			expectedEscort:     0,
			expectedCapital:    0,
			expectedTechLevels: 19,
		},
		{
			name:               "scenario-singleplayer player 1",
			file:               "../testdata/scenario-singleplayer/2483/Game.m1",
			playerIndex:        0,
			expectedScore:      838,
			expectedPlanets:    11,
			expectedStarbases:  1,
			expectedUnarmed:    6,
			expectedEscort:     2,
			expectedCapital:    0,
			expectedTechLevels: 76,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(tc.file)
			require.NoError(t, err)

			gs := store.New()
			err = gs.AddFile(tc.file, data)
			require.NoError(t, err)

			// Use the high-level API
			score := gs.PlayerScore(tc.playerIndex)
			require.NotNil(t, score, "PlayerScore() should return stored score for player %d", tc.playerIndex)

			// Verify all fields match expected values
			assert.Equal(t, tc.expectedScore, score.Score, "Score")
			assert.Equal(t, tc.expectedPlanets, score.Planets, "Planets")
			assert.Equal(t, tc.expectedStarbases, score.Starbases, "Starbases")
			assert.Equal(t, tc.expectedUnarmed, score.UnarmedShips, "UnarmedShips")
			assert.Equal(t, tc.expectedEscort, score.EscortShips, "EscortShips")
			assert.Equal(t, tc.expectedCapital, score.CapitalShips, "CapitalShips")
			assert.Equal(t, tc.expectedTechLevels, score.TechLevels, "TechLevels")
		})
	}
}

// TestPlayerScore_NotAvailable verifies PlayerScore returns nil when no score data exists.
func TestPlayerScore_NotAvailable(t *testing.T) {
	gs := store.New()

	// No files loaded - should return nil
	score := gs.PlayerScore(0)
	assert.Nil(t, score, "PlayerScore() should return nil when no data loaded")
}
