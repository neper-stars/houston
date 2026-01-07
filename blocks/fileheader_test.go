package blocks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neper-stars/houston/encoding"
)

// TestFileHeaderDocumentTypes tests that different file extensions have the correct
// document type (dt) in their file header.
//
// File types from RTBOF structure:
//   - dt=0: XY files (universe definition)
//   - dt=1: X files (turn order/submitted)
//   - dt=2: HST files (host file)
//   - dt=3: M files (player turn)
//   - dt=4: H files (history)
//   - dt=5: R files (race)
func TestFileHeaderDocumentTypes(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		expectedType uint8
		expectedName string
	}{
		// XY file (universe definition) - dt=0
		{
			name:         "XY file",
			path:         "../testdata/scenario-minefield/game.xy",
			expectedType: FileTypeXY,
			expectedName: "XY",
		},
		// X file (turn order) - dt=1
		{
			name:         "X file",
			path:         "../testdata/scenario-minefield/game.x1",
			expectedType: FileTypeX,
			expectedName: "X",
		},
		// M file (player turn) - dt=3
		{
			name:         "M file",
			path:         "../testdata/scenario-minefield/game.m1",
			expectedType: FileTypeM,
			expectedName: "M",
		},
		// H file (history) - dt=4
		{
			name:         "H file",
			path:         "../testdata/scenario-minefield/game.h1",
			expectedType: FileTypeH,
			expectedName: "H",
		},
		// R file (race) - dt=5
		{
			name:         "R file",
			path:         "../testdata/scenario-minefield/game.r1",
			expectedType: FileTypeRace,
			expectedName: "Race",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			header := loadFileHeader(t, tc.path)

			assert.Equal(t, tc.expectedType, header.FileType,
				"FileType should be %d for %s", tc.expectedType, tc.name)
			assert.Equal(t, tc.expectedName, header.FileTypeName(),
				"FileTypeName should be %q for %s", tc.expectedName, tc.name)
		})
	}

	// TODO: Add HST file test case once we have .hst test data files
	// HST files should have dt=2 (FileTypeHST)
	// Example test case:
	// {
	//     name:         "HST file",
	//     path:         "../testdata/scenario-xxx/game.hst",
	//     expectedType: FileTypeHST,
	//     expectedName: "HST",
	// },
}

// TestFileHeaderDocumentTypes_MultipleScenarios tests document types across different
// scenarios to ensure consistency.
func TestFileHeaderDocumentTypes_MultipleScenarios(t *testing.T) {
	scenarios := []string{
		"scenario-minefield",
		"scenario-basic",
		"scenario-map",
		"scenario-history",
	}

	for _, scenario := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			basePath := filepath.Join("..", "testdata", scenario)

			// Check for common file types in each scenario
			fileTests := []struct {
				pattern      string
				expectedType uint8
			}{
				{"game.xy", FileTypeXY},
				{"game.m1", FileTypeM},
				{"game.m2", FileTypeM},
				{"game.x1", FileTypeX},
				{"game.h1", FileTypeH},
				{"game.h2", FileTypeH},
				{"game.r1", FileTypeRace},
			}

			for _, ft := range fileTests {
				path := filepath.Join(basePath, ft.pattern)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					continue // Skip files that don't exist in this scenario
				}

				t.Run(ft.pattern, func(t *testing.T) {
					header := loadFileHeader(t, path)
					assert.Equal(t, ft.expectedType, header.FileType,
						"FileType mismatch for %s in %s", ft.pattern, scenario)
				})
			}
		})
	}
}

// TestFileHeaderDocumentTypes_XFilesHaveWGenValidation verifies that X files (dt=1)
// are the only file type with wGen validation.
func TestFileHeaderDocumentTypes_XFilesHaveWGenValidation(t *testing.T) {
	testCases := []struct {
		name              string
		path              string
		shouldBeValidated bool
	}{
		{"XY file", "../testdata/scenario-minefield/game.xy", false},
		{"X file", "../testdata/scenario-minefield/game.x1", true}, // Only X files!
		{"M file", "../testdata/scenario-minefield/game.m1", false},
		{"H file", "../testdata/scenario-minefield/game.h1", false},
		{"R file", "../testdata/scenario-minefield/game.r1", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			header := loadFileHeader(t, tc.path)
			assert.Equal(t, tc.shouldBeValidated, header.IsGenerationValidated(),
				"IsGenerationValidated should be %v for %s (dt=%d)",
				tc.shouldBeValidated, tc.name, header.FileType)
		})
	}

	// TODO: Add HST file test - should NOT have wGen validation (dt=2)
}

// TestFileHeaderDocumentTypes_PlayerIndexByFileType tests the expected player index
// for different file types.
func TestFileHeaderDocumentTypes_PlayerIndexByFileType(t *testing.T) {
	t.Run("XY file has playerIndex 31", func(t *testing.T) {
		header := loadFileHeader(t, "../testdata/scenario-minefield/game.xy")
		assert.Equal(t, RaceFilePlayerIndex, header.PlayerIndex(),
			"XY files should use player index 31 (like race files)")
	})

	t.Run("M file has player-specific index", func(t *testing.T) {
		// game.m1 is for player 1 (index 0)
		header := loadFileHeader(t, "../testdata/scenario-minefield/game.m1")
		assert.Equal(t, 0, header.PlayerIndex(),
			"game.m1 should be for player index 0")
	})

	t.Run("R file has playerIndex 31", func(t *testing.T) {
		header := loadFileHeader(t, "../testdata/scenario-minefield/game.r1")
		assert.Equal(t, RaceFilePlayerIndex, header.PlayerIndex(),
			"Race files should use player index 31")
	})
}

// TestFileTypeConstants verifies the file type constant values match the RTBOF spec.
func TestFileTypeConstants(t *testing.T) {
	// Verify constants match the spec from types.h
	assert.Equal(t, uint8(0), uint8(FileTypeUnknown), "FileTypeUnknown should be 0")
	assert.Equal(t, uint8(0), uint8(FileTypeXY), "FileTypeXY should be 0")
	assert.Equal(t, uint8(1), uint8(FileTypeX), "FileTypeX should be 1")
	assert.Equal(t, uint8(2), uint8(FileTypeHST), "FileTypeHST should be 2")
	assert.Equal(t, uint8(3), uint8(FileTypeM), "FileTypeM should be 3")
	assert.Equal(t, uint8(4), uint8(FileTypeH), "FileTypeH should be 4")
	assert.Equal(t, uint8(5), uint8(FileTypeRace), "FileTypeRace should be 5")

	// Verify aliases
	assert.Equal(t, FileTypeUnknown, FileTypeXY,
		"FileTypeUnknown and FileTypeXY should be the same value")
}

// TestFileHeaderGeneration tests the wGen field parsing.
func TestFileHeaderGeneration(t *testing.T) {
	testFiles := []string{
		"../testdata/scenario-minefield/game.m1",
		"../testdata/scenario-minefield/game.x1",
		"../testdata/scenario-basic/game.m1",
	}

	for _, path := range testFiles {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		t.Run(filepath.Base(path), func(t *testing.T) {
			header := loadFileHeader(t, path)

			// wGen is 3 bits, so value should be 0-7
			gen := header.Generation()
			assert.GreaterOrEqual(t, gen, 0, "Generation should be >= 0")
			assert.LessOrEqual(t, gen, 7, "Generation should be <= 7 (3 bits)")

			// Verify extraction from flags byte
			expectedGen := int((header.Flags & FlagGenMask) >> FlagGenShift)
			assert.Equal(t, expectedGen, gen, "Generation extraction should match")
		})
	}
}

// loadFileHeader reads a Stars! file and returns its FileHeader.
func loadFileHeader(t *testing.T, path string) *FileHeader {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read file: %s", path)
	require.GreaterOrEqual(t, len(data), 18, "file too small for header: %s", path)

	// Parse block header
	blockHeader := encoding.Read16(data, 0)
	typeID := BlockTypeID(blockHeader >> 10)
	size := BlockSize(blockHeader & 0x3FF)

	require.Equal(t, FileHeaderBlockType, typeID,
		"first block should be FileHeader, got type %d", typeID)
	require.Equal(t, BlockSize(16), size,
		"FileHeader block should be 16 bytes, got %d", size)

	block := GenericBlock{
		Type: typeID,
		Size: size,
		Data: BlockData(data[2:18]),
	}

	header, err := NewFileHeader(block)
	require.NoError(t, err, "failed to parse FileHeader")

	return header
}
