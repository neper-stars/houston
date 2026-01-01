package blocks

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neper-stars/houston/crypto"
	"github.com/neper-stars/houston/encoding"
)

func TestBlockTypeName(t *testing.T) {
	tests := []struct {
		id   BlockTypeID
		want string
	}{
		{FileFooterBlockType, "FileFooter"},
		{FileHeaderBlockType, "FileHeader"},
		{PlayerBlockType, "Player"},
		{PlanetsBlockType, "Planets"},
		{PlanetBlockType, "Planet"},
		{PartialPlanetBlockType, "PartialPlanet"},
		{FleetBlockType, "Fleet"},
		{PartialFleetBlockType, "PartialFleet"},
		{DesignBlockType, "Design"},
		{DesignChangeBlockType, "DesignChange"},
		{WaypointBlockType, "Waypoint"},
		{WaypointAddBlockType, "WaypointAdd"},
		{WaypointDeleteBlockType, "WaypointDelete"},
		{WaypointChangeTaskBlockType, "WaypointChangeTask"},
		{ProductionQueueBlockType, "ProductionQueue"},
		{ProductionQueueChangeBlockType, "ProductionQueueChange"},
		{BattlePlanBlockType, "BattlePlan"},
		{BattleBlockType, "Battle"},
		{MessageBlockType, "Message"},
		{ObjectBlockType, "Object"},
		{CountersBlockType, "Counters"},
		{FleetSplitBlockType, "FleetSplit"},
		{FleetsMergeBlockType, "FleetsMerge"},
		{SaveAndSubmitBlockType, "SaveAndSubmit"},
		{BlockTypeID(99), "Unknown"}, // Unknown type
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := BlockTypeName(tt.id)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPRTName(t *testing.T) {
	tests := []struct {
		prt  int
		want string
	}{
		{PRTHyperExpansion, "HE"},
		{PRTSuperStealth, "SS"},
		{PRTWarMonger, "WM"},
		{PRTClaimAdjuster, "CA"},
		{PRTInnerStrength, "IS"},
		{PRTSpaceDemolition, "SD"},
		{PRTPacketPhysics, "PP"},
		{PRTInterstellarTraveler, "IT"},
		{PRTAlternateReality, "AR"},
		{PRTJackOfAllTrades, "JOAT"},
		{-1, "Unknown"},
		{99, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := PRTName(tt.prt)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPRTFullName(t *testing.T) {
	tests := []struct {
		prt  int
		want string
	}{
		{PRTHyperExpansion, "Hyper Expansion"},
		{PRTSuperStealth, "Super Stealth"},
		{PRTWarMonger, "War Monger"},
		{PRTClaimAdjuster, "Claim Adjuster"},
		{PRTInnerStrength, "Inner Strength"},
		{PRTSpaceDemolition, "Space Demolition"},
		{PRTPacketPhysics, "Packet Physics"},
		{PRTInterstellarTraveler, "Interstellar Traveler"},
		{PRTAlternateReality, "Alternate Reality"},
		{PRTJackOfAllTrades, "Jack of All Trades"},
		{-1, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := PRTFullName(tt.prt)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLRTNames(t *testing.T) {
	tests := []struct {
		name string
		lrt  uint16
		want []string
	}{
		{"None", 0, nil},
		{"IFE only", LRTImprovedFuelEfficiency, []string{"IFE"}},
		{"TT only", LRTTotalTerraforming, []string{"TT"}},
		{"IFE+TT", LRTImprovedFuelEfficiency | LRTTotalTerraforming, []string{"IFE", "TT"}},
		{"Multiple", LRTImprovedFuelEfficiency | LRTAdvancedRemoteMining | LRTUltimateRecycling,
			[]string{"IFE", "ARM", "UR"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LRTNames(tt.lrt)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBattlePlanTacticName(t *testing.T) {
	tests := []struct {
		tactic int
		want   string
	}{
		{TacticDisengage, "Disengage"},
		{TacticDisengageIfChallenged, "Disengage if Challenged"},
		{TacticMinimizeDamage, "Minimize Damage"},
		{TacticMaximizeNetDamage, "Maximize Net Damage"},
		{TacticMaximizeDamageRatio, "Maximize Damage Ratio"},
		{TacticMaximizeDamage, "Maximize Damage"},
		{99, "Unknown"},
	}

	bp := &BattlePlanBlock{}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			bp.Tactic = tt.tactic
			got := bp.TacticName()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBattlePlanTargetName(t *testing.T) {
	tests := []struct {
		target int
		want   string
	}{
		{TargetNone, "None/Disengage"},
		{TargetAny, "Any"},
		{TargetStarbase, "Starbase"},
		{TargetArmedShips, "Armed Ships"},
		{TargetBombers, "Bombers/Freighters"},
		{TargetUnarmedShips, "Unarmed Ships"},
		{TargetFuelTransports, "Fuel Transports"},
		{TargetFreighters, "Freighters"},
		{99, "Unknown"},
	}

	bp := &BattlePlanBlock{}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			bp.PrimaryTarget = tt.target
			got := bp.PrimaryTargetName()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBattlePlanAttackWhoName(t *testing.T) {
	tests := []struct {
		attackWho int
		want      string
	}{
		{AttackNobody, "Nobody"},
		{AttackEnemies, "Enemies"},
		{AttackNeutralAndEnemies, "Neutrals & Enemies"},
		{AttackEveryone, "Everyone"},
		{AttackPlayerBase + 5, "Specific Player"}, // Targeting player 5
	}

	bp := &BattlePlanBlock{}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			bp.AttackWho = tt.attackWho
			got := bp.AttackWhoName()
			assert.Equal(t, tt.want, got)
		})
	}
}

// parseBlocksFromFile parses all blocks from a game file and returns typed blocks
func parseBlocksFromFile(t *testing.T, filename string) ([]Block, *FileHeader) {
	t.Helper()

	data, err := os.ReadFile(filename)
	require.NoError(t, err, "failed to read file")

	var blocks []Block
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
			if h.Shareware() {
				sw = 1
			}
			decryptor.InitDecryption(h.Salt(), int(h.GameID), int(h.Turn), h.PlayerIndex(), sw)
			blocks = append(blocks, *h)

		case FileFooterBlockType:
			block.Decrypted = DecryptedData(block.Data)
			blocks = append(blocks, *NewFileFooterBlock(*block))

		case BattlePlanBlockType:
			block.Decrypted = decryptor.DecryptBytes(block.Data)
			blocks = append(blocks, *NewBattlePlanBlock(*block))

		case CountersBlockType:
			block.Decrypted = decryptor.DecryptBytes(block.Data)
			blocks = append(blocks, *NewCountersBlock(*block))

		case DesignBlockType:
			block.Decrypted = decryptor.DecryptBytes(block.Data)
			db, err := NewDesignBlock(*block)
			require.NoError(t, err)
			blocks = append(blocks, *db)

		case FleetBlockType:
			block.Decrypted = decryptor.DecryptBytes(block.Data)
			blocks = append(blocks, *NewFleetBlock(*block))

		case PartialFleetBlockType:
			block.Decrypted = decryptor.DecryptBytes(block.Data)
			blocks = append(blocks, *NewPartialFleetBlock(*block))

		case WaypointBlockType:
			block.Decrypted = decryptor.DecryptBytes(block.Data)
			blocks = append(blocks, *NewWaypointBlock(*block))

		case ProductionQueueBlockType:
			block.Decrypted = decryptor.DecryptBytes(block.Data)
			blocks = append(blocks, *NewProductionQueueBlock(*block))

		case MessageBlockType:
			block.Decrypted = decryptor.DecryptBytes(block.Data)
			blocks = append(blocks, *NewMessageBlock(*block))

		default:
			block.Decrypted = decryptor.DecryptBytes(block.Data)
			blocks = append(blocks, *block)
		}
	}

	return blocks, header
}

func TestBattlePlanBlockRoundTrip(t *testing.T) {
	blocks, _ := parseBlocksFromFile(t, "../testdata/scenario-battleplans/game.m2")

	var battlePlans []BattlePlanBlock
	for _, b := range blocks {
		if bp, ok := b.(BattlePlanBlock); ok {
			battlePlans = append(battlePlans, bp)
		}
	}

	require.NotEmpty(t, battlePlans, "no battle plans found")

	for i, bp := range battlePlans {
		t.Run(bp.Name, func(t *testing.T) {
			encoded := bp.Encode()
			original := []byte(bp.Decrypted)

			assert.Equal(t, original, encoded,
				"battle plan %d (%s) encode mismatch", i, bp.Name)
		})
	}
}

func TestDesignBlockRoundTrip(t *testing.T) {
	blocks, _ := parseBlocksFromFile(t, "../testdata/scenario-basic/game.m1")

	var designs []DesignBlock
	for _, b := range blocks {
		if db, ok := b.(DesignBlock); ok {
			designs = append(designs, db)
		}
	}

	require.NotEmpty(t, designs, "no design blocks found")

	for _, db := range designs {
		t.Run(db.Name, func(t *testing.T) {
			encoded := db.Encode()
			original := []byte(db.Decrypted)

			assert.Equal(t, original, encoded,
				"design %d (%s) encode mismatch", db.DesignNumber, db.Name)
		})
	}
}

func TestProductionQueueBlockRoundTrip(t *testing.T) {
	blocks, _ := parseBlocksFromFile(t, "../testdata/scenario-map/game.m1")

	var queues []ProductionQueueBlock
	for _, b := range blocks {
		if pq, ok := b.(ProductionQueueBlock); ok {
			queues = append(queues, pq)
		}
	}

	require.NotEmpty(t, queues, "no production queue blocks found")

	for i, pq := range queues {
		t.Run("ProductionQueue", func(t *testing.T) {
			encoded := pq.Encode()
			original := []byte(pq.Decrypted)

			assert.Equal(t, original, encoded,
				"production queue %d encode mismatch", i)
		})
	}
}

func TestWaypointBlockRoundTrip(t *testing.T) {
	blocks, _ := parseBlocksFromFile(t, "../testdata/scenario-basic/game.m1")

	var waypoints []WaypointBlock
	for _, b := range blocks {
		if wp, ok := b.(WaypointBlock); ok {
			waypoints = append(waypoints, wp)
		}
	}

	require.NotEmpty(t, waypoints, "no waypoints found")

	for i, wp := range waypoints {
		t.Run("Waypoint", func(t *testing.T) {
			encoded := wp.Encode()
			original := []byte(wp.Decrypted)

			assert.Equal(t, original, encoded,
				"waypoint %d encode mismatch", i)
		})
	}
}

func TestFleetBlockRoundTrip(t *testing.T) {
	blocks, _ := parseBlocksFromFile(t, "../testdata/scenario-basic/game.m1")

	var fleets []FleetBlock
	for _, b := range blocks {
		if fb, ok := b.(FleetBlock); ok {
			fleets = append(fleets, fb)
		}
	}

	require.NotEmpty(t, fleets, "no fleet blocks found")

	for _, fb := range fleets {
		t.Run("Fleet", func(t *testing.T) {
			encoded := fb.Encode()
			original := []byte(fb.Decrypted)

			assert.Equal(t, original, encoded,
				"fleet %d encode mismatch", fb.FleetNumber)
		})
	}
}

func TestMessageBlockRoundTrip(t *testing.T) {
	blocks, _ := parseBlocksFromFile(t, "../testdata/scenario-message/game.m1")

	var messages []MessageBlock
	for _, b := range blocks {
		if mb, ok := b.(MessageBlock); ok {
			messages = append(messages, mb)
		}
	}

	require.NotEmpty(t, messages, "no message blocks found")

	for i, mb := range messages {
		t.Run("Message", func(t *testing.T) {
			encoded := mb.Encode()
			original := []byte(mb.Decrypted)

			assert.Equal(t, original, encoded,
				"message %d encode mismatch", i)
		})
	}
}
