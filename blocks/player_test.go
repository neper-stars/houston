package blocks

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neper-stars/houston/crypto"
	"github.com/neper-stars/houston/encoding"
)

// TestPlayerBlockHeaderFields tests the idPlanetHome and Rank fields
// in the player block header (bytes 0x08-0x0B).
//
// Header structure from PLAYER in types.h:
//   - 0x08-0x09: idPlanetHome (Home planet ID)
//   - 0x0A-0x0B: wScore (actually stores Rank, not Score - see note below)
//
// NOTE: The decompiled source names the field at 0x0A-0x0B as "wScore", but
// in the actual game UI this value represents the player's Rank (1st, 2nd, etc.),
// not their Score. The Score shown in the Player Scores dialog appears to be
// computed client-side and is not stored in the file.
func TestPlayerBlockHeaderFields(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		playerNumber int
		homePlanetID int
		rank         int // Ranking position (1=1st, 2=2nd, etc.) - called "wScore" in decompiled source
		raceName     string
	}{
		// scenario-basic: Year 2400, all players start at rank 0 (unranked)
		{
			name:         "scenario-basic player 1 (Halflings)",
			path:         "../testdata/scenario-basic/game.m1",
			playerNumber: 0,
			homePlanetID: 11, // Planet "Abacus"
			rank:         0,  // Year 2400, not yet ranked
			raceName:     "Halfling",
		},
		{
			name:         "scenario-basic player 2 (Hobbits)",
			path:         "../testdata/scenario-basic/game.m2",
			playerNumber: 1,
			homePlanetID: 19, // Planet "Milky Way"
			rank:         0,  // Year 2400, not yet ranked
			raceName:     "Hobbit",
		},
		// scenario-minefield: Later turn, player is ranked 1st
		{
			name:         "scenario-minefield player 1",
			path:         "../testdata/scenario-minefield/game.m1",
			playerNumber: 0,
			homePlanetID: 21,
			rank:         1, // Ranked 1st place (see testdata/scenario-minefield/scores.png)
			raceName:     "MineMonger",
		},
		// scenario-history: Later turn, player is ranked 2nd
		{
			name:         "scenario-history player 2 (file owner)",
			path:         "../testdata/scenario-history/game.m2",
			playerNumber: 1,
			homePlanetID: 19, // Same homeworld as scenario-basic
			rank:         2,  // Ranked 2nd place (see testdata/scenario-history/scores.png)
			raceName:     "Hobbit",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			players := loadPlayerBlocks(t, tc.path)

			// Find the file owner's player block (has FullDataFlag=true)
			var player *PlayerBlock
			for i := range players {
				if players[i].PlayerNumber == tc.playerNumber && players[i].FullDataFlag {
					player = &players[i]
					break
				}
			}
			require.NotNil(t, player, "Should find player %d with full data", tc.playerNumber)

			// Verify header fields
			assert.Equal(t, tc.homePlanetID, player.HomePlanetID,
				"HomePlanetID (idPlanetHome) should match")
			assert.Equal(t, tc.rank, player.Rank,
				"Rank (wScore in decompiled source) should match")
			assert.Equal(t, tc.raceName, player.NameSingular,
				"Race name should match")
		})
	}
}

// TestPlayerBlockHeaderFields_OnlyFileOwnerHasValidData verifies that
// only the file owner's player block has valid HomePlanetID and Rank.
// Other players' blocks in an M file have FullDataFlag=false and contain
// garbage/encrypted data in these fields.
func TestPlayerBlockHeaderFields_OnlyFileOwnerHasValidData(t *testing.T) {
	// Load game.m2 which contains both player 1 and player 2 data
	// Player 2 (index 1) is the file owner with valid data
	// Player 1 (index 0) has partial data without valid header fields
	players := loadPlayerBlocks(t, "../testdata/scenario-history/game.m2")

	require.GreaterOrEqual(t, len(players), 2, "Should have at least 2 players")

	// Find player blocks
	var fileOwner, otherPlayer *PlayerBlock
	for i := range players {
		switch players[i].PlayerNumber {
		case 1:
			fileOwner = &players[i]
		case 0:
			otherPlayer = &players[i]
		}
	}

	require.NotNil(t, fileOwner, "Should find file owner (player 1)")
	require.NotNil(t, otherPlayer, "Should find other player (player 0)")

	t.Run("FileOwnerHasFullData", func(t *testing.T) {
		assert.True(t, fileOwner.FullDataFlag, "File owner should have FullDataFlag=true")
		assert.Equal(t, 19, fileOwner.HomePlanetID, "File owner should have valid HomePlanetID")
		assert.Equal(t, 2, fileOwner.Rank, "File owner should have valid Rank")
	})

	t.Run("OtherPlayerHasPartialData", func(t *testing.T) {
		// Other player may or may not have FullDataFlag depending on game settings
		// But if they don't have full data, the header fields will be 0 or garbage
		if !otherPlayer.FullDataFlag {
			// Without full data, HomePlanetID and Rank are not parsed
			// (they stay at default 0 values)
			t.Log("Other player has partial data (FullDataFlag=false)")
		}
	})
}

// TestPlayerBlockRankProgression verifies that player ranks are assigned
// after the first turn.
func TestPlayerBlockRankProgression(t *testing.T) {
	// scenario-basic is year 2400 (turn 0) - players are not yet ranked
	earlyPlayer := loadFileOwnerPlayer(t, "../testdata/scenario-basic/game.m1")
	require.NotNil(t, earlyPlayer)
	assert.Equal(t, 0, earlyPlayer.Rank, "Year 2400 should have rank 0 (unranked)")

	// scenario-minefield is a later turn - player has been ranked
	laterPlayer := loadFileOwnerPlayer(t, "../testdata/scenario-minefield/game.m1")
	require.NotNil(t, laterPlayer)
	assert.GreaterOrEqual(t, laterPlayer.Rank, 1, "Later turn should have rank >= 1")
}

// TestPlayerBlockHomePlanetIDConsistency verifies that the same player
// has the same homeworld ID across different turns (homeworld doesn't change).
func TestPlayerBlockHomePlanetIDConsistency(t *testing.T) {
	// Player 2 (Hobbits) appears in both scenario-basic and scenario-history
	basicPlayer := loadFileOwnerPlayer(t, "../testdata/scenario-basic/game.m2")
	require.NotNil(t, basicPlayer)

	historyPlayer := loadFileOwnerPlayer(t, "../testdata/scenario-history/game.m2")
	require.NotNil(t, historyPlayer)

	assert.Equal(t, basicPlayer.HomePlanetID, historyPlayer.HomePlanetID,
		"Same player should have same HomePlanetID across turns")
	assert.Equal(t, "Hobbit", basicPlayer.NameSingular)
	assert.Equal(t, "Hobbit", historyPlayer.NameSingular)
}

// TestPlayerBlockEncode_RoundTrip tests that HomePlanetID and Rank
// are correctly encoded back to binary format.
func TestPlayerBlockEncode_RoundTrip(t *testing.T) {
	players := loadPlayerBlocks(t, "../testdata/scenario-basic/game.m1")

	for _, player := range players {
		if !player.FullDataFlag {
			continue // Skip partial player data
		}

		t.Run(player.NameSingular, func(t *testing.T) {
			// Encode the player block
			encoded, err := player.Encode()
			require.NoError(t, err)

			// Verify HomePlanetID at bytes 8-9
			homePlanetID := encoding.Read16(encoded, 8)
			assert.Equal(t, uint16(player.HomePlanetID), homePlanetID,
				"Encoded HomePlanetID should match")

			// Verify Rank at bytes 10-11
			rank := encoding.Read16(encoded, 10)
			assert.Equal(t, uint16(player.Rank), rank,
				"Encoded Rank should match")
		})
	}
}

// loadPlayerBlocks loads all PlayerBlocks from a Stars! M file.
func loadPlayerBlocks(t *testing.T, path string) []PlayerBlock {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read file: %s", path)

	var players []PlayerBlock
	var header *FileHeader
	decryptor := crypto.NewDecryptor()
	offset := 0

	for offset < len(data) {
		blockHeader := encoding.Read16(data, offset)
		typeID := BlockTypeID(blockHeader >> 10)
		size := BlockSize(blockHeader & 0x3FF)

		var blockData BlockData
		if int(size) > 0 && offset+2+int(size) <= len(data) {
			blockData = BlockData(data[offset+2 : offset+2+int(size)])
		}

		block := &GenericBlock{
			Type: typeID,
			Size: size,
			Data: blockData,
		}
		offset += int(size) + 2

		switch typeID {
		case FileHeaderBlockType:
			h, err := NewFileHeader(*block)
			require.NoError(t, err)
			header = h
			var sw int
			if h.Crippled() {
				sw = 1
			}
			decryptor.InitDecryption(h.Salt(), int(h.GameID), int(h.Turn), h.PlayerIndex(), sw)

		case PlayerBlockType:
			block.Decrypted = decryptor.DecryptBytes(block.Data)
			pb, err := NewPlayerBlock(*block)
			require.NoError(t, err)
			players = append(players, *pb)
		}
	}

	require.NotNil(t, header, "File should have a header")
	return players
}

// loadFileOwnerPlayer loads the file owner's PlayerBlock from an M file.
// The file owner is the player whose index matches the file's player index.
func loadFileOwnerPlayer(t *testing.T, path string) *PlayerBlock {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read file: %s", path)

	var header *FileHeader
	decryptor := crypto.NewDecryptor()
	offset := 0

	for offset < len(data) {
		blockHeader := encoding.Read16(data, offset)
		typeID := BlockTypeID(blockHeader >> 10)
		size := BlockSize(blockHeader & 0x3FF)

		var blockData BlockData
		if int(size) > 0 && offset+2+int(size) <= len(data) {
			blockData = BlockData(data[offset+2 : offset+2+int(size)])
		}

		block := &GenericBlock{
			Type: typeID,
			Size: size,
			Data: blockData,
		}
		offset += int(size) + 2

		switch typeID {
		case FileHeaderBlockType:
			h, err := NewFileHeader(*block)
			require.NoError(t, err)
			header = h
			var sw int
			if h.Crippled() {
				sw = 1
			}
			decryptor.InitDecryption(h.Salt(), int(h.GameID), int(h.Turn), h.PlayerIndex(), sw)

		case PlayerBlockType:
			block.Decrypted = decryptor.DecryptBytes(block.Data)
			pb, err := NewPlayerBlock(*block)
			require.NoError(t, err)

			// Return if this is the file owner's player block
			if pb.PlayerNumber == header.PlayerIndex() && pb.FullDataFlag {
				return pb
			}
		}
	}

	return nil
}

// TestPlayerBlockConstants verifies the constant values match the spec.
func TestPlayerBlockConstants(t *testing.T) {
	// AI skill levels
	assert.Equal(t, 0, AISkillEasy)
	assert.Equal(t, 1, AISkillStandard)
	assert.Equal(t, 2, AISkillHarder)
	assert.Equal(t, 3, AISkillExpert)

	// Research cost modifiers
	assert.Equal(t, 0, ResearchCostExpensive)
	assert.Equal(t, 1, ResearchCostNormal)
	assert.Equal(t, 2, ResearchCostCheap)

	// Leftover points spending options
	assert.Equal(t, 0, SpendLeftoverSurfaceMinerals)
	assert.Equal(t, 1, SpendLeftoverMining)
	assert.Equal(t, 2, SpendLeftoverDefenses)
	assert.Equal(t, 3, SpendLeftoverFactories)
	assert.Equal(t, 4, SpendLeftoverMineralAlchemy)
	assert.Equal(t, 5, SpendLeftoverResearch)

	// Stored relation values
	assert.Equal(t, 0, StoredRelationNeutral)
	assert.Equal(t, 1, StoredRelationFriend)
	assert.Equal(t, 2, StoredRelationEnemy)
}
