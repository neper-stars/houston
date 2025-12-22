package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neper-stars/houston/blocks"
)

// Expected data structures for history scenario
type ExpectedPlayerScore struct {
	Turn         int    `json:"turn"`
	Player       string `json:"player"`
	Score        int    `json:"score"`
	Resources    int64  `json:"resources"`
	Planets      int    `json:"planets"`
	Starbases    int    `json:"starbases"`
	UnarmedShips int    `json:"unarmedShips"`
	EscortShips  int    `json:"escortShips"`
	CapitalShips int    `json:"capitalShips"`
	TechLevels   int    `json:"techLevels"`
}

type ExpectedHistoryData struct {
	Scenario     string                `json:"scenario"`
	PlayerScores []ExpectedPlayerScore `json:"playerScores"`
}

func loadHistoryExpected(t *testing.T) *ExpectedHistoryData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-history", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedHistoryData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadHistoryFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-history", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

func TestScenarioHistory_PlayerScores(t *testing.T) {
	expected := loadHistoryExpected(t)
	_, hBlockList := loadHistoryFile(t, "game.h2")

	// Build player ID to name map from M file (H file only has enemy player info)
	_, mBlockList := loadHistoryFile(t, "game.m2")
	playerNames := make(map[int]string)
	for _, block := range mBlockList {
		if p, ok := block.(blocks.PlayerBlock); ok {
			playerNames[p.PlayerNumber] = p.NamePlural
		}
	}

	blockList := hBlockList

	// Collect player scores
	var scores []blocks.PlayerScoresBlock
	for _, block := range blockList {
		if ps, ok := block.(blocks.PlayerScoresBlock); ok {
			scores = append(scores, ps)
		}
	}

	require.Equal(t, len(expected.PlayerScores), len(scores),
		"Should have %d player score entries", len(expected.PlayerScores))

	// Validate each expected score entry
	for i, exp := range expected.PlayerScores {
		t.Run(fmt.Sprintf("Turn_%d_%s", exp.Turn, exp.Player), func(t *testing.T) {
			score := scores[i]

			// Validate player name
			actualPlayerName := playerNames[score.PlayerID]
			assert.Equal(t, exp.Player, actualPlayerName, "Player name should match")

			// Validate turn
			assert.Equal(t, exp.Turn, score.Turn, "Turn should match")

			// Validate score data
			assert.Equal(t, exp.Score, score.Score, "Score should match")
			assert.Equal(t, exp.Resources, score.Resources, "Resources should match")
			assert.Equal(t, exp.Planets, score.Planets, "Planets should match")
			assert.Equal(t, exp.Starbases, score.Starbases, "Starbases should match")
			assert.Equal(t, exp.UnarmedShips, score.UnarmedShips, "Unarmed ships should match")
			assert.Equal(t, exp.EscortShips, score.EscortShips, "Escort ships should match")
			assert.Equal(t, exp.CapitalShips, score.CapitalShips, "Capital ships should match")
			assert.Equal(t, exp.TechLevels, score.TechLevels, "Tech levels should match")
		})
	}
}

func TestScenarioHistory_FileHeader(t *testing.T) {
	_, blockList := loadHistoryFile(t, "game.h2")

	var header *blocks.FileHeader
	for _, block := range blockList {
		if h, ok := block.(blocks.FileHeader); ok {
			header = &h
			break
		}
	}

	require.NotNil(t, header, "FileHeader should exist")
	assert.Equal(t, 1, header.PlayerIndex(), "Should be player 2's file (index 1)")
	assert.Equal(t, uint16(7), header.Turn, "H file should be at turn 7")
	assert.Equal(t, 2407, header.Year(), "Year should be 2407")
}
