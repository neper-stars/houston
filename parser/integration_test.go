package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/data"
)

// TestParseGameM1 tests parsing a real Stars! multiplayer file.
// This file is from the starsapi-python test suite.
func TestParseGameM1(t *testing.T) {
	// Find the testdata directory relative to the test file
	testFile := filepath.Join("..", "testdata", "Game.m1")

	// Check if file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file Game.m1 not found in testdata/")
	}

	// Read the file
	fileBytes, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	fd := FileData(fileBytes)

	// Parse file header
	header, err := fd.FileHeader()
	if err != nil {
		t.Fatalf("FileHeader() failed: %v", err)
	}

	// Verify it's a valid file header
	if header.Type != blocks.FileHeaderBlockType {
		t.Errorf("Expected FileHeader block type %d, got %d",
			blocks.FileHeaderBlockType, header.Type)
	}

	// Parse all blocks
	blockList, err := fd.BlockList()
	if err != nil {
		t.Fatalf("BlockList() failed: %v", err)
	}

	// Should have at least a few blocks
	if len(blockList) < 3 {
		t.Errorf("Expected at least 3 blocks, got %d", len(blockList))
	}

	// Log block types found (useful for debugging)
	blockTypes := make(map[blocks.BlockTypeID]int)
	for _, block := range blockList {
		switch b := block.(type) {
		case blocks.FileHeader:
			blockTypes[b.Type]++
		case blocks.GenericBlock:
			blockTypes[b.Type]++
		case blocks.PlanetsBlock:
			blockTypes[b.Type]++
			// Verify planets were parsed
			if len(b.Planets) > 0 {
				t.Logf("Found %d planets, first: %s at (%d, %d)",
					len(b.Planets), b.Planets[0].Name, b.Planets[0].X, b.Planets[0].Y)
			}
			// Test HasGameSetting helper
			if b.HasGameSetting(data.GameSettingPublicScores) {
				t.Log("Game has public scores enabled")
			}
		case blocks.FileFooterBlock:
			blockTypes[b.Type]++
		case blocks.FileHashBlock:
			blockTypes[b.Type]++
		default:
			// For other block types, try to get the type
			if gb, ok := block.(interface{ GetType() blocks.BlockTypeID }); ok {
				blockTypes[gb.GetType()]++
			}
		}
	}

	t.Logf("Found %d blocks with types: %v", len(blockList), blockTypes)

	// Verify we found expected block types
	if _, ok := blockTypes[blocks.FileHeaderBlockType]; !ok {
		t.Error("Expected to find FileHeader block")
	}
	if _, ok := blockTypes[blocks.FileFooterBlockType]; !ok {
		t.Error("Expected to find FileFooter block")
	}
}

// TestPlanetsBlockGameSettings tests the game settings helper function.
func TestPlanetsBlockGameSettings(t *testing.T) {
	// Create a mock PlanetsBlock with known settings
	pb := &blocks.PlanetsBlock{
		GameSettings: uint16(data.GameSettingMaxMinerals | data.GameSettingPublicScores),
	}

	// Test positive cases
	if !pb.HasGameSetting(data.GameSettingMaxMinerals) {
		t.Error("Expected MaxMinerals to be set")
	}
	if !pb.HasGameSetting(data.GameSettingPublicScores) {
		t.Error("Expected PublicScores to be set")
	}

	// Test negative cases
	if pb.HasGameSetting(data.GameSettingSlowTech) {
		t.Error("Expected SlowTech to NOT be set")
	}
	if pb.HasGameSetting(data.GameSettingNoRandomEvents) {
		t.Error("Expected NoRandomEvents to NOT be set")
	}
}
