package parser_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

func TestScenarioDiplomacy(t *testing.T) {
	t.Run("Side1_HobbitSetsFriend", func(t *testing.T) {
		// Hobbits (player 0) set Halflings (player 1) as Friend
		data, err := os.ReadFile("../testdata/scenario-diplomacy/1/side1/game.x1")
		require.NoError(t, err)

		fd := parser.FileData(data)
		blockList, err := fd.BlockList()
		require.NoError(t, err)

		var found bool
		for _, block := range blockList {
			if prc, ok := block.(blocks.PlayersRelationChangeBlock); ok {
				found = true
				assert.Equal(t, blocks.RelationFriend, prc.Relation, "relation should be Friend")
				assert.Equal(t, 1, prc.TargetPlayer, "target player should be 1 (Halflings)")
				assert.Equal(t, "Friend", prc.RelationName())
				break
			}
		}
		require.True(t, found, "should find PlayersRelationChangeBlock")
	})

	t.Run("Side2_HalflingSetsEnemy", func(t *testing.T) {
		// Halflings (player 1) set Hobbits (player 0) as Enemy
		data, err := os.ReadFile("../testdata/scenario-diplomacy/1/side2/game.x2")
		require.NoError(t, err)

		fd := parser.FileData(data)
		blockList, err := fd.BlockList()
		require.NoError(t, err)

		var found bool
		for _, block := range blockList {
			if prc, ok := block.(blocks.PlayersRelationChangeBlock); ok {
				found = true
				assert.Equal(t, blocks.RelationEnemy, prc.Relation, "relation should be Enemy")
				assert.Equal(t, 0, prc.TargetPlayer, "target player should be 0 (Hobbits)")
				assert.Equal(t, "Enemy", prc.RelationName())
				break
			}
		}
		require.True(t, found, "should find PlayersRelationChangeBlock")
	})
}

// TestScenarioDiplomacy3Way tests reading stored relations from M files
// in a 3-player game after turn generation.
//
// Player relations set:
//   - P0 (Hobbits):   P1=Friend, P2=Neutral
//   - P1 (Halflings): P0=Neutral, P2=Enemy
//   - P2 (Orcs):      P0=Friend, P1=Enemy
func TestScenarioDiplomacy3Way(t *testing.T) {
	t.Run("Player0_Relations", func(t *testing.T) {
		// Player 0 (Hobbits) set: P1=Friend, P2=Neutral
		data, err := os.ReadFile("../testdata/scenario-diplomacy-3way/1/side1/game.m1")
		require.NoError(t, err)

		fd := parser.FileData(data)
		blockList, err := fd.BlockList()
		require.NoError(t, err)

		var ownPlayer *blocks.PlayerBlock
		for _, block := range blockList {
			if pb, ok := block.(blocks.PlayerBlock); ok && pb.PlayerNumber == 0 {
				ownPlayer = &pb
				break
			}
		}

		require.NotNil(t, ownPlayer, "should find own PlayerBlock")
		require.True(t, len(ownPlayer.PlayerRelations) >= 2, "should have relation data")

		// Storage encoding: 0=Neutral, 1=Friend, 2=Enemy
		// P0 set P1 as Friend -> stored as 1
		// P0 set P2 as Neutral -> stored as 0
		// Note: index 0 is self (P0), so we check indices 1 and 2
		assert.Equal(t, blocks.StoredRelationFriend, ownPlayer.GetRelationTo(1), "relation to P1 should be Friend")
		assert.Equal(t, blocks.StoredRelationNeutral, ownPlayer.GetRelationTo(2), "relation to P2 should be Neutral")
	})

	t.Run("Player1_Relations", func(t *testing.T) {
		// Player 1 (Halflings) set: P0=Neutral, P2=Enemy
		data, err := os.ReadFile("../testdata/scenario-diplomacy-3way/1/side2/game.m2")
		require.NoError(t, err)

		fd := parser.FileData(data)
		blockList, err := fd.BlockList()
		require.NoError(t, err)

		var ownPlayer *blocks.PlayerBlock
		for _, block := range blockList {
			if pb, ok := block.(blocks.PlayerBlock); ok && pb.PlayerNumber == 1 {
				ownPlayer = &pb
				break
			}
		}

		require.NotNil(t, ownPlayer, "should find own PlayerBlock")
		require.True(t, len(ownPlayer.PlayerRelations) >= 3, "should have relation data")

		// P1 set P0 as Neutral -> stored as 0
		// P1 set P2 as Enemy -> stored as 2
		assert.Equal(t, blocks.StoredRelationNeutral, ownPlayer.GetRelationTo(0), "relation to P0 should be Neutral")
		assert.Equal(t, blocks.StoredRelationEnemy, ownPlayer.GetRelationTo(2), "relation to P2 should be Enemy")
	})

	t.Run("Player2_Relations", func(t *testing.T) {
		// Player 2 (Orcs) set: P0=Friend, P1=Enemy
		data, err := os.ReadFile("../testdata/scenario-diplomacy-3way/1/side3/game.m3")
		require.NoError(t, err)

		fd := parser.FileData(data)
		blockList, err := fd.BlockList()
		require.NoError(t, err)

		var ownPlayer *blocks.PlayerBlock
		for _, block := range blockList {
			if pb, ok := block.(blocks.PlayerBlock); ok && pb.PlayerNumber == 2 {
				ownPlayer = &pb
				break
			}
		}

		require.NotNil(t, ownPlayer, "should find own PlayerBlock")
		require.True(t, len(ownPlayer.PlayerRelations) >= 2, "should have relation data")

		// P2 set P0 as Friend -> stored as 1
		// P2 set P1 as Enemy -> stored as 2
		assert.Equal(t, blocks.StoredRelationFriend, ownPlayer.GetRelationTo(0), "relation to P0 should be Friend")
		assert.Equal(t, blocks.StoredRelationEnemy, ownPlayer.GetRelationTo(1), "relation to P1 should be Enemy")
	})
}
