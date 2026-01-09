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

// TestFileHeaderHelperMethods tests all FileHeader helper methods
func TestFileHeaderHelperMethods(t *testing.T) {
	_, header := parseBlocksFromFile(t, "../testdata/scenario-basic/game.m1")
	require.NotNil(t, header, "file header should be parsed")

	t.Run("Magic", func(t *testing.T) {
		magic := header.Magic()
		assert.Len(t, magic, 4, "magic should be 4 bytes")
		assert.Equal(t, "J3J3", magic, "Stars! files should have J3J3 magic")
	})

	t.Run("Version", func(t *testing.T) {
		// Version string should be formatted as major.minor.increment
		vs := header.VersionString()
		assert.NotEmpty(t, vs, "version string should not be empty")
		assert.Contains(t, vs, ".", "version should contain dots")

		// Individual version components
		assert.GreaterOrEqual(t, header.VersionMajor(), 0, "major version should be non-negative")
		assert.GreaterOrEqual(t, header.VersionMinor(), 0, "minor version should be non-negative")
		assert.GreaterOrEqual(t, header.VersionIncrement(), 0, "increment should be non-negative")
	})

	t.Run("Year", func(t *testing.T) {
		year := header.Year()
		// Year = 2400 + Turn, should be at least 2400
		assert.GreaterOrEqual(t, year, StarsBaseYear, "year should be at least base year")
	})

	t.Run("PlayerIndex", func(t *testing.T) {
		idx := header.PlayerIndex()
		// Player index is 0-based, so player 1 files have index 0
		assert.Equal(t, 0, idx, "game.m1 should be player index 0")
	})

	t.Run("Salt", func(t *testing.T) {
		salt := header.Salt()
		// Salt is 11 bits, so max is 2047
		assert.GreaterOrEqual(t, salt, 0, "salt should be non-negative")
		assert.LessOrEqual(t, salt, MaxSaltValue-1, "salt should fit in 11 bits")
	})

	t.Run("FileTypeName", func(t *testing.T) {
		typeName := header.FileTypeName()
		assert.Equal(t, "M", typeName, ".m1 file should have type M")
	})

	t.Run("Flags", func(t *testing.T) {
		// Just test that flag methods don't panic
		_ = header.TurnSubmitted()
		_ = header.HostUsing()
		_ = header.MultipleTurns()
		_ = header.GameOver()
		_ = header.Shareware()
	})
}

// TestFileHeaderEncode tests FileHeader round-trip encoding
func TestFileHeaderEncode(t *testing.T) {
	_, header := parseBlocksFromFile(t, "../testdata/scenario-basic/game.m1")
	require.NotNil(t, header, "file header should be parsed")

	encoded := header.Encode()
	original := []byte(header.Data)

	assert.Equal(t, original, encoded, "FileHeader encode should match original")
}

// TestVersionEncoding tests the version encoding/decoding functions
func TestVersionEncoding(t *testing.T) {
	t.Run("EncodeVersion", func(t *testing.T) {
		// Test known version 2.83.0
		encoded := EncodeVersion(2, 83, 0)
		major := int(encoded >> 12)
		minor := int((encoded >> 5) & 0x7F)
		increment := int(encoded & 0x1F)

		assert.Equal(t, 2, major, "major version should be 2")
		assert.Equal(t, 83, minor, "minor version should be 83")
		assert.Equal(t, 0, increment, "increment should be 0")
	})

	t.Run("StarsVersionData", func(t *testing.T) {
		v := StarsVersionData()
		// Parse back the version
		major := int(v >> 12)
		minor := int((v >> 5) & 0x7F)
		increment := int(v & 0x1F)

		assert.Equal(t, StarsVersionMajor, major)
		assert.Equal(t, StarsVersionMinor, minor)
		assert.Equal(t, StarsVersionIncrement, increment)
	})
}

// TestNewFileHeaderForRaceFile tests race file header creation
func TestNewFileHeaderForRaceFile(t *testing.T) {
	header := NewFileHeaderForRaceFile()
	require.NotNil(t, header)

	assert.Equal(t, "J3J3", header.Magic())
	assert.Equal(t, uint32(0), header.GameID, "race files have GameID 0")
	assert.Equal(t, uint16(0), header.Turn, "race files have Turn 0")
	assert.Equal(t, RaceFilePlayerIndex, header.PlayerIndex(), "race files use player index 31")
	assert.Equal(t, FileTypeRace, int(header.FileType), "should be race file type")

	// Salt should be random but valid
	salt := header.Salt()
	assert.GreaterOrEqual(t, salt, 0)
	assert.Less(t, salt, MaxSaltValue)
}

// TestFileHeaderSetters tests SetSalt and SetPlayerIndex
func TestFileHeaderSetters(t *testing.T) {
	header := NewFileHeaderForRaceFile()

	t.Run("SetSalt", func(t *testing.T) {
		originalPlayer := header.PlayerIndex()
		header.SetSalt(1234)
		assert.Equal(t, 1234, header.Salt())
		assert.Equal(t, originalPlayer, header.PlayerIndex(), "player index should be preserved")
	})

	t.Run("SetPlayerIndex", func(t *testing.T) {
		originalSalt := header.Salt()
		header.SetPlayerIndex(5)
		assert.Equal(t, 5, header.PlayerIndex())
		assert.Equal(t, originalSalt, header.Salt(), "salt should be preserved")
	})
}

// TestCountersBlock tests the CountersBlock parsing
func TestCountersBlock(t *testing.T) {
	// Use scenario-map which has a counters block
	blocks, _ := parseBlocksFromFile(t, "../testdata/scenario-map/game.m1")

	var counters *CountersBlock
	for _, b := range blocks {
		if cb, ok := b.(CountersBlock); ok {
			counters = &cb
			break
		}
	}

	if counters != nil {
		assert.Greater(t, counters.PlanetCount, 0, "should have planets")
		assert.GreaterOrEqual(t, counters.FleetCount, 0, "fleet count should be non-negative")
	}
}

// TestCountersBlockDirect tests CountersBlock parsing directly
func TestCountersBlockDirect(t *testing.T) {
	// Create a synthetic counters block
	decrypted := []byte{0x64, 0x00, 0x0A, 0x00} // 100 planets, 10 fleets
	block := GenericBlock{
		Type:      CountersBlockType,
		Size:      4,
		Data:      BlockData(decrypted),
		Decrypted: DecryptedData(decrypted),
	}

	cb := NewCountersBlock(block)
	assert.Equal(t, 100, cb.PlanetCount)
	assert.Equal(t, 10, cb.FleetCount)
}

// TestPartialFleetBlockRoundTrip tests PartialFleetBlock encoding
func TestPartialFleetBlockRoundTrip(t *testing.T) {
	blocks, _ := parseBlocksFromFile(t, "../testdata/scenario-basic/game.m1")

	var partialFleets []PartialFleetBlock
	for _, b := range blocks {
		if pf, ok := b.(PartialFleetBlock); ok {
			partialFleets = append(partialFleets, pf)
		}
	}

	// Note: The scenario may not have partial fleets, so we don't require them
	for i, pf := range partialFleets {
		t.Run("PartialFleet", func(t *testing.T) {
			encoded := pf.Encode()
			original := []byte(pf.Decrypted)

			assert.Equal(t, original, encoded,
				"partial fleet %d encode mismatch", i)
		})
	}
}

// TestPartialFleetBlockHelpers tests PartialFleetBlock helper methods
func TestPartialFleetBlockHelpers(t *testing.T) {
	blocks, _ := parseBlocksFromFile(t, "../testdata/scenario-basic/game.m1")

	var fleet *FleetBlock
	for _, b := range blocks {
		if fb, ok := b.(FleetBlock); ok {
			fleet = &fb
			break
		}
	}

	require.NotNil(t, fleet, "should have at least one fleet")

	t.Run("GetFleetIdAndOwner", func(t *testing.T) {
		id := fleet.GetFleetIdAndOwner()
		// Should combine fleet number (9 bits) and owner (shifted by 9)
		assert.Equal(t, fleet.FleetNumber|(fleet.Owner<<9), id)
	})

	t.Run("IsFullFleet", func(t *testing.T) {
		// FleetBlock is a full fleet (kind byte = 7)
		assert.True(t, fleet.IsFullFleet())
	})

	t.Run("HasCargo", func(t *testing.T) {
		// Full fleets have cargo
		assert.True(t, fleet.HasCargo())
	})

	t.Run("TotalShips", func(t *testing.T) {
		total := fleet.TotalShips()
		assert.GreaterOrEqual(t, total, 0)
	})
}

// TestObjectBlockMinefield tests ObjectBlock parsing for minefields
func TestObjectBlockMinefield(t *testing.T) {
	blocks, _ := parseBlocksFromFileWithObjects(t, "../testdata/scenario-map/minefields/game.m1")

	var minefields []ObjectBlock
	for _, b := range blocks {
		if ob, ok := b.(ObjectBlock); ok && ob.IsMinefield() {
			minefields = append(minefields, ob)
		}
	}

	require.NotEmpty(t, minefields, "should have minefields in minefield scenario")

	for i, mf := range minefields {
		t.Run("Minefield", func(t *testing.T) {
			// Test helper methods
			assert.True(t, mf.IsMinefield())
			assert.False(t, mf.IsWormhole())
			assert.False(t, mf.IsMysteryTrader())
			assert.False(t, mf.IsPacket())
			assert.False(t, mf.IsSalvage())

			// Test round-trip encoding
			encoded := mf.Encode()
			original := []byte(mf.Decrypted)
			assert.Equal(t, original, encoded, "minefield %d encode mismatch", i)
		})
	}
}

// TestObjectBlockCountObject tests count object parsing
func TestObjectBlockCountObject(t *testing.T) {
	blocks, _ := parseBlocksFromFileWithObjects(t, "../testdata/scenario-map/minefields/game.m1")

	var countObj *ObjectBlock
	for _, b := range blocks {
		if ob, ok := b.(ObjectBlock); ok && ob.IsCountObject {
			countObj = &ob
			break
		}
	}

	if countObj != nil {
		assert.True(t, countObj.IsCountObject)
		assert.GreaterOrEqual(t, countObj.Count, 0)

		// Test round-trip
		encoded := countObj.Encode()
		original := []byte(countObj.Decrypted)
		assert.Equal(t, original, encoded, "count object encode mismatch")
	}
}

// TestObjectBlockStabilityName tests wormhole stability name helper
func TestObjectBlockStabilityName(t *testing.T) {
	ob := &ObjectBlock{}

	testCases := []struct {
		stabilityIndex int
		expected       string
	}{
		{WormholeStabilityIndexRockSolid, "Rock Solid"},
		{WormholeStabilityIndexStable, "Stable"},
		{WormholeStabilityIndexVolatile, "Volatile"},
		{WormholeStabilityIndexVeryVolatile, "Very Volatile"},
		{5, "Unknown"}, // Invalid index
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			ob.StabilityIndex = tc.stabilityIndex
			assert.Equal(t, tc.expected, ob.StabilityName())
		})
	}
}

// TestObjectBlockPlayerVisibility tests player visibility helpers
func TestObjectBlockPlayerVisibility(t *testing.T) {
	ob := &ObjectBlock{
		MineCanSeeBits:  0x0005, // Players 0 and 2
		CanSeeBits:      0x0003, // Players 0 and 1
		BeenThroughBits: 0x0002, // Player 1
		MetBits:         0x0001, // Player 0
	}

	t.Run("PlayerCanSeeMinefield", func(t *testing.T) {
		assert.True(t, ob.PlayerCanSeeMinefield(0))
		assert.False(t, ob.PlayerCanSeeMinefield(1))
		assert.True(t, ob.PlayerCanSeeMinefield(2))
		assert.False(t, ob.PlayerCanSeeMinefield(-1), "invalid player should return false")
		assert.False(t, ob.PlayerCanSeeMinefield(16), "invalid player should return false")
	})

	t.Run("PlayerCanSee", func(t *testing.T) {
		assert.True(t, ob.PlayerCanSee(0))
		assert.True(t, ob.PlayerCanSee(1))
		assert.False(t, ob.PlayerCanSee(2))
		assert.False(t, ob.PlayerCanSee(-1))
	})

	t.Run("PlayerBeenThrough", func(t *testing.T) {
		assert.False(t, ob.PlayerBeenThrough(0))
		assert.True(t, ob.PlayerBeenThrough(1))
		assert.False(t, ob.PlayerBeenThrough(2))
	})

	t.Run("TraderHasMet", func(t *testing.T) {
		assert.True(t, ob.TraderHasMet(0))
		assert.False(t, ob.TraderHasMet(1))
	})
}

// TestObjectBlockTraderItems tests trader item helper
func TestObjectBlockTraderItems(t *testing.T) {
	ob := &ObjectBlock{
		ItemBits: TraderItemMultiCargoPod | TraderItemGenesisDevice,
	}

	assert.True(t, ob.TraderHasItem(TraderItemMultiCargoPod))
	assert.True(t, ob.TraderHasItem(TraderItemGenesisDevice))
	assert.False(t, ob.TraderHasItem(TraderItemLangstonShield))
	assert.False(t, ob.TraderHasItem(TraderItemShip))
}

// TestObjectBlockMinerals tests packet mineral helpers
func TestObjectBlockMinerals(t *testing.T) {
	ob := &ObjectBlock{
		Ironium:   100,
		Boranium:  200,
		Germanium: 300,
	}

	assert.Equal(t, 600, ob.TotalMinerals())
}

// TestObjectBlockWarpSpeed tests warp speed decoding for packets
func TestObjectBlockWarpSpeed(t *testing.T) {
	// The formula is: warp = (rawByte >> 2) - 44
	// So rawByte = (warp + 44) * 4 for warp 5: (5+44)*4 = 196
	ob := &ObjectBlock{PacketSpeed: 196}
	assert.Equal(t, 5, ob.WarpSpeed())

	// For warp 10: (10+44)*4 = 216
	ob.PacketSpeed = 216
	assert.Equal(t, 10, ob.WarpSpeed())
}

// TestFleetSplitBlock tests FleetSplitBlock parsing and encoding
func TestFleetSplitBlock(t *testing.T) {
	// Create a synthetic split block for fleet 123
	decrypted := []byte{0x7B, 0x00} // Fleet 123 (0x7B = 123)
	block := GenericBlock{
		Type:      FleetSplitBlockType,
		Size:      2,
		Data:      BlockData(decrypted),
		Decrypted: DecryptedData(decrypted),
	}

	fsb := NewFleetSplitBlock(block)
	assert.Equal(t, 123, fsb.FleetNumber)

	// Test round-trip
	encoded := fsb.Encode()
	assert.Equal(t, decrypted, encoded)
}

// TestFleetsMergeBlock tests FleetsMergeBlock parsing and encoding
func TestFleetsMergeBlock(t *testing.T) {
	// Create a synthetic merge block: fleet 1 merging fleets 2 and 3
	decrypted := []byte{0x01, 0x00, 0x02, 0x00, 0x03, 0x00}
	block := GenericBlock{
		Type:      FleetsMergeBlockType,
		Size:      6,
		Data:      BlockData(decrypted),
		Decrypted: DecryptedData(decrypted),
	}

	fmb := NewFleetsMergeBlock(block)
	assert.Equal(t, 1, fmb.FleetNumber)
	assert.Equal(t, []int{2, 3}, fmb.FleetsToMerge)

	// Test round-trip
	encoded := fmb.Encode()
	assert.Equal(t, decrypted, encoded)
}

// TestFleetNameBlock tests FleetNameBlock parsing and encoding
func TestFleetNameBlock(t *testing.T) {
	// Create a synthetic fleet name block with Stars! string encoding
	// Stars! string format: length byte followed by encoded characters
	name := "Test Fleet"
	encodedName := encoding.EncodeStarsString(name)

	block := GenericBlock{
		Type:      FleetNameBlockType,
		Size:      BlockSize(len(encodedName)),
		Data:      BlockData(encodedName),
		Decrypted: DecryptedData(encodedName),
	}

	fnb := NewFleetNameBlock(block)
	assert.Equal(t, name, fnb.Name)

	// Test round-trip
	encoded := fnb.Encode()
	assert.Equal(t, encodedName, encoded)
}

// TestBattlePlanTargetPlayer tests specific player targeting
func TestBattlePlanTargetPlayer(t *testing.T) {
	bp := &BattlePlanBlock{AttackWho: AttackPlayerBase + 5}
	assert.Equal(t, 5, bp.TargetPlayer())

	// Non-specific targeting
	bp.AttackWho = AttackEnemies
	assert.Equal(t, -1, bp.TargetPlayer())
}

// TestBattlePlanSecondaryTarget tests secondary target name
func TestBattlePlanSecondaryTarget(t *testing.T) {
	bp := &BattlePlanBlock{SecondaryTarget: TargetStarbase}
	assert.Equal(t, "Starbase", bp.SecondaryTargetName())
}

// TestProductionQueueItemHelpers tests QueueItem helper methods
func TestProductionQueueItemHelpers(t *testing.T) {
	// Auto items are standard items with ItemId <= 6
	autoItem := QueueItem{ItemType: ProductionItemTypeStandard, ItemId: ProductionItemAutoFactories}
	assert.True(t, autoItem.IsAutoItem())
	assert.False(t, autoItem.IsShipDesign())

	// Ship design items have ItemType == 4 (custom)
	shipItem := QueueItem{ItemType: ProductionItemTypeCustom, ItemId: 5}
	assert.False(t, shipItem.IsAutoItem())
	assert.True(t, shipItem.IsShipDesign())

	// Standard non-auto items
	mineItem := QueueItem{ItemType: ProductionItemTypeStandard, ItemId: ProductionItemMine}
	assert.False(t, mineItem.IsAutoItem())
	assert.False(t, mineItem.IsShipDesign())
}

// TestMessageBlockHelpers tests MessageBlock helper methods
func TestMessageBlockHelpers(t *testing.T) {
	t.Run("IsBroadcast", func(t *testing.T) {
		// Broadcast: ReceiverId == 0 (sent to everyone)
		mb := &MessageBlock{ReceiverId: 0}
		assert.True(t, mb.IsBroadcast())

		mb.ReceiverId = 1
		assert.False(t, mb.IsBroadcast())
	})

	t.Run("IsReply", func(t *testing.T) {
		// Reply: InReplyTo > 0 (message ID of the message being replied to)
		mb := &MessageBlock{InReplyTo: 3}
		assert.True(t, mb.IsReply())

		mb.InReplyTo = 0 // Not a reply
		assert.False(t, mb.IsReply())
	})

	t.Run("DisplayIds", func(t *testing.T) {
		mb := &MessageBlock{SenderId: 0, ReceiverId: 1}
		// SenderDisplayId is 1-indexed (SenderId + 1)
		assert.Equal(t, 1, mb.SenderDisplayId())
		// ReceiverDisplayId returns ReceiverId directly
		assert.Equal(t, 1, mb.ReceiverDisplayId())
	})
}

// TestWaypointHelpers tests WaypointBlock helper methods
func TestWaypointHelpers(t *testing.T) {
	t.Run("UsesStargate", func(t *testing.T) {
		wp := &WaypointBlock{}
		wp.Warp = WarpStargate // Warp 11 means stargate
		assert.True(t, wp.UsesStargate())

		wp.Warp = 5
		assert.False(t, wp.UsesStargate())
	})

	t.Run("TransportOrders", func(t *testing.T) {
		wp := &WaypointBlock{WaypointTask: WaypointTaskTransport}

		// Set load all action for ironium
		wp.TransportOrders[CargoIronium] = TransportOrder{Action: TransportTaskLoadAll}
		assert.True(t, wp.HasTransportOrders())

		// Get transport order
		order := wp.GetTransportOrder(CargoIronium)
		assert.Equal(t, TransportTaskLoadAll, order.Action)

		// Invalid cargo type returns empty order
		invalidOrder := wp.GetTransportOrder(99)
		assert.Equal(t, 0, invalidOrder.Action)
	})

	t.Run("LoadUnloadAll", func(t *testing.T) {
		wp := &WaypointBlock{WaypointTask: WaypointTaskTransport}

		// Set Colonists to Load All - should make IsLoadAllTransport return true
		wp.TransportOrders[CargoColonists].Action = TransportTaskLoadAll
		assert.True(t, wp.IsLoadAllTransport())
		assert.False(t, wp.IsUnloadAllTransport())

		// Set Colonists to Unload All - should make IsUnloadAllTransport return true
		wp.TransportOrders[CargoColonists].Action = TransportTaskUnloadAll
		assert.True(t, wp.IsUnloadAllTransport())
		assert.False(t, wp.IsLoadAllTransport())

		// Non-transport task - should return false regardless of TransportOrders
		wp.WaypointTask = WaypointTaskColonize
		assert.False(t, wp.IsLoadAllTransport())
		assert.False(t, wp.IsUnloadAllTransport())
	})
}

// TestTransportTaskName tests transport task naming
func TestTransportTaskName(t *testing.T) {
	tests := []struct {
		task int
		name string
	}{
		{TransportTaskNoAction, "No Action"},
		{TransportTaskLoadAll, "Load All Available"},
		{TransportTaskUnloadAll, "Unload All"},
		{TransportTaskLoadExactly, "Load Exactly"},
		{TransportTaskUnloadExactly, "Unload Exactly"},
		{TransportTaskFillToPercent, "Fill Up to %"},
		{TransportTaskWaitForPercent, "Wait for %"},
		{TransportTaskDropAndLoad, "Drop and Load"},
		{TransportTaskSetAmountTo, "Set Amount To"},
		{99, "Unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.name, TransportTaskName(tc.task))
		})
	}
}

// TestDesignBlockHelpers tests DesignBlock helper methods
func TestDesignBlockHelpers(t *testing.T) {
	blocks, _ := parseBlocksFromFile(t, "../testdata/scenario-basic/game.m1")

	var design *DesignBlock
	for _, b := range blocks {
		if db, ok := b.(DesignBlock); ok {
			design = &db
			break
		}
	}

	require.NotNil(t, design)

	t.Run("GetSlot", func(t *testing.T) {
		if len(design.Slots) > 0 {
			slot := design.GetSlot(0)
			assert.NotNil(t, slot)
		}

		// Invalid slot
		slot := design.GetSlot(99)
		assert.Nil(t, slot)
	})

	t.Run("ShipCount", func(t *testing.T) {
		count := design.ShipCount()
		assert.GreaterOrEqual(t, count, int64(0))
	})
}

// TestFileFooterHelpers tests FileFooterBlock helper methods
func TestFileFooterHelpers(t *testing.T) {
	// Footer with checksum (4 bytes)
	withChecksum := GenericBlock{
		Type:      FileFooterBlockType,
		Size:      4,
		Data:      BlockData([]byte{0x12, 0x34, 0x56, 0x78}),
		Decrypted: DecryptedData([]byte{0x12, 0x34, 0x56, 0x78}),
	}
	fb1 := NewFileFooterBlock(withChecksum)
	assert.True(t, fb1.HasChecksum())

	// Footer without checksum (0 bytes)
	withoutChecksum := GenericBlock{
		Type:      FileFooterBlockType,
		Size:      0,
		Data:      BlockData([]byte{}),
		Decrypted: DecryptedData([]byte{}),
	}
	fb2 := NewFileFooterBlock(withoutChecksum)
	assert.False(t, fb2.HasChecksum())
}

// TestFileFooterEncode tests FileFooterBlock encoding
func TestFileFooterEncode(t *testing.T) {
	// Footer with checksum 0x3412 (little-endian in first 2 bytes)
	data := []byte{0x12, 0x34}
	block := GenericBlock{
		Type:      FileFooterBlockType,
		Size:      2,
		Data:      BlockData(data),
		Decrypted: DecryptedData(data),
	}
	fb := NewFileFooterBlock(block)

	// Encode returns only the 2-byte checksum
	encoded := fb.Encode()
	assert.Equal(t, data, encoded)
	assert.Equal(t, uint16(0x3412), fb.Checksum)

	// Test H file (no checksum)
	emptyBlock := GenericBlock{
		Type:      FileFooterBlockType,
		Size:      0,
		Data:      BlockData([]byte{}),
		Decrypted: DecryptedData([]byte{}),
	}
	hFileFooter := NewFileFooterBlock(emptyBlock)
	assert.Equal(t, []byte{}, hFileFooter.Encode())
}

// TestFileTypeNames tests all file type names
func TestFileTypeNames(t *testing.T) {
	tests := []struct {
		fileType uint8
		expected string
	}{
		{FileTypeXY, "XY"},      // XY files use type 0
		{FileTypeUnknown, "XY"}, // Alias for XY
		{FileTypeX, "X"},
		{FileTypeHST, "HST"},
		{FileTypeM, "M"},
		{FileTypeH, "H"},
		{FileTypeRace, "Race"},
		{99, "Unknown(99)"}, // True unknown type
	}

	header := &FileHeader{}
	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			header.FileType = tc.fileType
			assert.Equal(t, tc.expected, header.FileTypeName())
		})
	}
}

// TestEncodeBlockWithHeader tests the block header encoding
func TestEncodeBlockWithHeader(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}
	result := EncodeBlockWithHeader(CountersBlockType, data)

	// Should have 2 bytes header + data
	assert.Len(t, result, 6)

	// Decode header
	header := encoding.Read16(result, 0)
	typeID := BlockTypeID(header >> 10)
	size := int(header & 0x3FF)

	assert.Equal(t, CountersBlockType, typeID)
	assert.Equal(t, 4, size)
	assert.Equal(t, data, result[2:])
}

// parseBlocksFromFileWithObjects extends parseBlocksFromFile to also handle ObjectBlocks
func parseBlocksFromFileWithObjects(t *testing.T, filename string) ([]Block, *FileHeader) {
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

		case ObjectBlockType:
			block.Decrypted = decryptor.DecryptBytes(block.Data)
			blocks = append(blocks, *NewObjectBlock(*block))

		default:
			block.Decrypted = decryptor.DecryptBytes(block.Data)
			blocks = append(blocks, *block)
		}
	}

	return blocks, header
}
