package store_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neper-stars/houston/store"
)

func TestDetectFileType(t *testing.T) {
	tests := []struct {
		filename string
		expected store.FileSourceType
	}{
		// M files
		{"game.m1", store.SourceTypeMFile},
		{"game.M1", store.SourceTypeMFile},
		{"Game.m16", store.SourceTypeMFile},

		// X files
		{"game.x1", store.SourceTypeXFile},
		{"Game.X16", store.SourceTypeXFile},

		// H files (history)
		{"game.h1", store.SourceTypeHFile},
		{"Game.H16", store.SourceTypeHFile},

		// HST files (host)
		{"game.hst", store.SourceTypeHSTFile},
		{"Game.HST", store.SourceTypeHSTFile},
		{"game-2400.hst", store.SourceTypeHSTFile},

		// XY files
		{"game.xy", store.SourceTypeXYFile},
		{"Game.XY", store.SourceTypeXYFile},

		// R files (race)
		{"player.r1", store.SourceTypeRFile},
		{"Player.R16", store.SourceTypeRFile},

		// Unknown
		{"game.txt", store.SourceTypeUnknown},
		{"game", store.SourceTypeUnknown},
	}

	for _, tc := range tests {
		t.Run(tc.filename, func(t *testing.T) {
			result := store.DetectFileType(tc.filename)
			assert.Equal(t, tc.expected, result, "DetectFileType(%q)", tc.filename)
		})
	}
}

func TestFileSourceType_String(t *testing.T) {
	tests := []struct {
		fileType store.FileSourceType
		expected string
	}{
		{store.SourceTypeMFile, "M-File"},
		{store.SourceTypeXFile, "X-File"},
		{store.SourceTypeHFile, "H-File"},
		{store.SourceTypeXYFile, "XY-File"},
		{store.SourceTypeRFile, "R-File"},
		{store.SourceTypeHSTFile, "HST-File"},
		{store.SourceTypeUnknown, "Unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.fileType.String())
		})
	}
}

func TestHSTFile_ParseAndRegenerate(t *testing.T) {
	// Load an HST file from testdata
	data, err := os.ReadFile("../testdata/scenario-cloaking-visibility/game01/historic-backup/game-2400.hst")
	if err != nil {
		t.Skip("HST test file not found")
	}

	gs := store.New()
	err = gs.AddFile("game-2400.hst", data)
	require.NoError(t, err)

	// Verify the source was detected as HST
	sources := gs.Sources()
	require.Len(t, sources, 1)
	assert.Equal(t, store.SourceTypeHSTFile, sources[0].Type)

	// Regenerate the HST file
	regenerated, err := gs.RegenerateHSTFile()
	require.NoError(t, err)
	assert.NotEmpty(t, regenerated)

	// Parse the regenerated file to verify it's valid
	gs2 := store.New()
	err = gs2.AddFile("regenerated.hst", regenerated)
	require.NoError(t, err)

	// Verify block count matches (basic round-trip check)
	sources2 := gs2.Sources()
	require.Len(t, sources2, 1)
	assert.Equal(t, len(sources[0].Blocks), len(sources2[0].Blocks),
		"regenerated file should have the same number of blocks")
}
