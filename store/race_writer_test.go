package store

import (
	"testing"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
	"github.com/neper-stars/houston/race"
)

func TestCreateRaceFile(t *testing.T) {
	// Create a race using the builder - use default Humanoid which should be valid
	builder := race.New()
	builder.Name("Klingon", "Klingons")
	// Don't add expensive traits, just use defaults

	r, err := builder.Finish()
	if err != nil {
		t.Fatalf("Failed to finish race: %v", err)
	}

	// Create a race file
	data, err := CreateRaceFile(r, 1)
	if err != nil {
		t.Fatalf("Failed to create race file: %v", err)
	}

	// Verify the file has reasonable size
	if len(data) < 50 {
		t.Errorf("Race file too small: %d bytes", len(data))
	}

	t.Logf("Created race file: %d bytes", len(data))

	// Try parsing it to see what happens
	source, err := ParseSource("test.r1", data)
	if err != nil {
		t.Logf("Parse error: %v", err)
		// Dump first few blocks for debugging
		fd := parser.FileData(data)
		allBlocks, listErr := fd.BlockList()
		t.Logf("BlockList returned %d blocks, err: %v", len(allBlocks), listErr)
		for i, block := range allBlocks {
			t.Logf("Block %d: type=%d size=%d", i, block.BlockTypeID(), block.BlockSize())
			decrypted := block.DecryptedData()
			t.Logf("  Decrypted len=%d", len(decrypted))
			if len(decrypted) >= 8 {
				t.Logf("  bytes 0-7: %02x %02x %02x %02x %02x %02x %02x %02x",
					decrypted[0], decrypted[1], decrypted[2], decrypted[3],
					decrypted[4], decrypted[5], decrypted[6], decrypted[7])
				if block.BlockTypeID() == blocks.PlayerBlockType {
					t.Logf("  byte3 & 0xFC = 0x%02x (should be 0)", decrypted[3]&0xFC)
					t.Logf("  byte5 & 0x0C = 0x%02x (should be 0)", decrypted[5]&0x0C)
					t.Logf("  byte6 & 0x03 = 0x%02x (should be 3)", decrypted[6]&0x03)
				}
			}
		}
	} else {
		t.Logf("Parse succeeded, header: magic=%s player=%d", source.Header.Magic(), source.PlayerIndex)
	}
}

func TestCreateRaceFileRoundTrip(t *testing.T) {
	// Create a race using the builder
	builder := race.New()
	builder.Name("TestRace", "TestRaces")
	builder.PRT(9) // JOAT
	builder.GrowthRate(15)
	builder.Gravity(50, 25)
	builder.Temperature(50, 25)
	builder.Radiation(50, 25)
	builder.ColonistsPerResource(1000)
	builder.Factories(10, 10, 10, false)
	builder.Mines(10, 5, 10)
	builder.Research(1, 1, 1, 1, 1, 1) // All standard
	builder.LeftoverPointsOn(race.LeftoverSurfaceMinerals)

	r, err := builder.Finish()
	if err != nil {
		t.Fatalf("Failed to finish race: %v", err)
	}

	// Create a race file
	data, err := CreateRaceFile(r, 1)
	if err != nil {
		t.Fatalf("Failed to create race file: %v", err)
	}

	// Parse the file back
	fd := parser.FileData(data)
	parsedBlocks, err := fd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse race file: %v", err)
	}

	// Verify we got the expected blocks
	if len(parsedBlocks) < 3 {
		t.Fatalf("Expected at least 3 blocks (header, player, footer), got %d", len(parsedBlocks))
	}

	// Find and verify the header
	var header *blocks.FileHeader
	for _, block := range parsedBlocks {
		if fh, ok := block.(blocks.FileHeader); ok {
			header = &fh
			break
		}
	}
	if header == nil {
		t.Fatal("No FileHeader found in parsed blocks")
	}

	// Verify header properties
	if header.Magic() != "J3J3" {
		t.Errorf("Expected magic 'J3J3', got '%s'", header.Magic())
	}
	if header.GameID != 0 {
		t.Errorf("Expected GameID 0, got %d", header.GameID)
	}
	if header.Turn != 0 {
		t.Errorf("Expected Turn 0, got %d", header.Turn)
	}
	// Race files always use playerIndex=31 in header (player slot is in filename only)
	if header.PlayerIndex() != 31 {
		t.Errorf("Expected PlayerIndex 31 (race file convention), got %d", header.PlayerIndex())
	}

	// Find and verify the PlayerBlock
	var playerBlock *blocks.PlayerBlock
	for _, block := range parsedBlocks {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			playerBlock = &pb
			break
		}
	}
	if playerBlock == nil {
		t.Fatal("No PlayerBlock found in parsed blocks")
	}

	// Verify race data round-tripped correctly
	if playerBlock.NameSingular != "TestRace" {
		t.Errorf("Expected singular name 'TestRace', got '%s'", playerBlock.NameSingular)
	}
	if playerBlock.NamePlural != "TestRaces" {
		t.Errorf("Expected plural name 'TestRaces', got '%s'", playerBlock.NamePlural)
	}
	if playerBlock.PRT != 9 {
		t.Errorf("Expected PRT 9 (JOAT), got %d", playerBlock.PRT)
	}
	if playerBlock.GrowthRate != 15 {
		t.Errorf("Expected growth rate 15, got %d", playerBlock.GrowthRate)
	}
	if playerBlock.Production.FactoryProduction != 10 {
		t.Errorf("Expected factory production 10, got %d", playerBlock.Production.FactoryProduction)
	}
	if playerBlock.Production.FactoryCost != 10 {
		t.Errorf("Expected factory cost 10, got %d", playerBlock.Production.FactoryCost)
	}
	if playerBlock.Production.MineProduction != 10 {
		t.Errorf("Expected mine production 10, got %d", playerBlock.Production.MineProduction)
	}
	if playerBlock.Production.MineCost != 5 {
		t.Errorf("Expected mine cost 5, got %d", playerBlock.Production.MineCost)
	}

	t.Logf("Round-trip successful: race '%s' (%s)", playerBlock.NameSingular, blocks.PRTName(playerBlock.PRT))
}

func TestCreateRaceFileWithLRTs(t *testing.T) {
	// Create a race with LRTs
	builder := race.New()
	builder.Name("Warrior", "Warriors")
	builder.PRT(2) // WM
	builder.AddLRT(0)  // IFE
	builder.AddLRT(7)  // NRSE
	builder.GrowthRate(18)

	r, err := builder.Finish()
	if err != nil {
		t.Fatalf("Failed to finish race: %v", err)
	}

	// Create and parse the race file
	data, err := CreateRaceFile(r, 2)
	if err != nil {
		t.Fatalf("Failed to create race file: %v", err)
	}

	fd := parser.FileData(data)
	parsedBlocks, err := fd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse race file: %v", err)
	}

	// Find the PlayerBlock
	var playerBlock *blocks.PlayerBlock
	for _, block := range parsedBlocks {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			playerBlock = &pb
			break
		}
	}
	if playerBlock == nil {
		t.Fatal("No PlayerBlock found")
	}

	// Verify LRTs
	if playerBlock.LRT != r.LRT {
		t.Errorf("Expected LRT 0x%04X, got 0x%04X", r.LRT, playerBlock.LRT)
	}
	if !playerBlock.HasLRT(blocks.LRTImprovedFuelEfficiency) {
		t.Error("Expected IFE LRT")
	}
	if !playerBlock.HasLRT(blocks.LRTNoRamScoopEngines) {
		t.Error("Expected NRSE LRT")
	}

	t.Logf("LRTs verified: %v", blocks.LRTNames(playerBlock.LRT))
}

func TestCreateRaceFileWithImmunity(t *testing.T) {
	// Create a race with immunity - balance with narrow ranges
	builder := race.New()
	builder.Name("Immune", "Immunes")
	builder.PRT(9) // JOAT
	builder.GravityImmune(true)
	builder.Temperature(50, 10) // Narrow range to save points
	builder.Radiation(50, 10)   // Narrow range to save points

	r, err := builder.Finish()
	if err != nil {
		t.Fatalf("Failed to finish race: %v", err)
	}

	// Create and parse the race file
	data, err := CreateRaceFile(r, 3)
	if err != nil {
		t.Fatalf("Failed to create race file: %v", err)
	}

	fd := parser.FileData(data)
	parsedBlocks, err := fd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse race file: %v", err)
	}

	// Find the PlayerBlock
	var playerBlock *blocks.PlayerBlock
	for _, block := range parsedBlocks {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			playerBlock = &pb
			break
		}
	}
	if playerBlock == nil {
		t.Fatal("No PlayerBlock found")
	}

	// Verify immunity
	if !playerBlock.Hab.IsGravityImmune() {
		t.Error("Expected gravity immunity")
	}
	if playerBlock.Hab.IsTemperatureImmune() {
		t.Error("Did not expect temperature immunity")
	}
	if playerBlock.Hab.IsRadiationImmune() {
		t.Error("Did not expect radiation immunity")
	}

	t.Logf("Immunity verified: gravity=%v, temp=%v, rad=%v",
		playerBlock.Hab.IsGravityImmune(),
		playerBlock.Hab.IsTemperatureImmune(),
		playerBlock.Hab.IsRadiationImmune())
}
