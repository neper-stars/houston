package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type expectedMinefield struct {
	Owner          string `json:"owner"`
	X              int    `json:"x"`
	Y              int    `json:"y"`
	MineCount      int64  `json:"mineCount"`
	ExpectedRadius int    `json:"expectedRadius"`
}

type expectedData struct {
	Minefields []expectedMinefield `json:"minefields"`
}

func TestObjectEntity_Radius(t *testing.T) {
	testdataDir := filepath.Join("..", "testdata", "scenario-map", "minefields")

	// Load expected data
	expectedFile := filepath.Join(testdataDir, "expected.json")
	expectedBytes, err := os.ReadFile(expectedFile)
	require.NoError(t, err, "failed to read expected.json")

	var expected expectedData
	err = json.Unmarshal(expectedBytes, &expected)
	require.NoError(t, err, "failed to parse expected.json")

	// Load game file
	gameFile := filepath.Join(testdataDir, "game.m1")
	gs := New()
	err = gs.AddFileWithXY(gameFile)
	require.NoError(t, err, "failed to load game file")

	// Get minefields from store
	minefields := gs.Minefields()
	require.NotEmpty(t, minefields, "no minefields found in game file")

	// Build a map of expected minefields by position for easy lookup
	expectedByPos := make(map[[2]int]expectedMinefield)
	for _, mf := range expected.Minefields {
		key := [2]int{mf.X, mf.Y}
		expectedByPos[key] = mf
	}

	// Verify each minefield's radius
	for _, mf := range minefields {
		key := [2]int{mf.X, mf.Y}
		exp, found := expectedByPos[key]
		if !found {
			continue // Skip minefields not in expected data
		}

		// Radius() returns float64, round down to match GUI display
		actualRadius := int(mf.Radius())

		assert.Equal(t, exp.MineCount, mf.MineCount,
			"mine count mismatch for minefield at (%d, %d)", mf.X, mf.Y)
		assert.Equal(t, exp.ExpectedRadius, actualRadius,
			"radius mismatch for minefield at (%d, %d): expected %d ly, got %d ly (from %d mines)",
			mf.X, mf.Y, exp.ExpectedRadius, actualRadius, mf.MineCount)

		// Mark as verified by removing from expected map
		delete(expectedByPos, key)
	}

	// Ensure all expected minefields were found
	assert.Empty(t, expectedByPos, "some expected minefields were not found in game data")
}
