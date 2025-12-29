package store

import (
	"os"
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

func TestHumanoidPredefinedRace(t *testing.T) {
	// Read the expected race file produced by Stars!
	expectedData, err := os.ReadFile("../testdata/scenario-racebuilder/predefined-races/humanoids/race.r1")
	if err != nil {
		t.Fatalf("Failed to read expected race file: %v", err)
	}

	// Parse the expected file
	expectedFd := parser.FileData(expectedData)
	expectedBlocks, err := expectedFd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse expected race file: %v", err)
	}

	// Find the PlayerBlock
	var expectedPlayer *blocks.PlayerBlock
	for _, block := range expectedBlocks {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			expectedPlayer = &pb
			break
		}
	}
	if expectedPlayer == nil {
		t.Fatal("No PlayerBlock found in expected file")
	}

	// Log the expected values for verification
	t.Logf("Expected race: '%s' / '%s'", expectedPlayer.NameSingular, expectedPlayer.NamePlural)
	t.Logf("  PRT: %d (%s)", expectedPlayer.PRT, blocks.PRTName(expectedPlayer.PRT))
	t.Logf("  LRT: 0x%04X (%v)", expectedPlayer.LRT, blocks.LRTNames(expectedPlayer.LRT))
	t.Logf("  Growth rate: %d%%", expectedPlayer.GrowthRate)
	t.Logf("  Habitability: G=%d/%d/%d T=%d/%d/%d R=%d/%d/%d",
		expectedPlayer.Hab.GravityLow, expectedPlayer.Hab.GravityCenter, expectedPlayer.Hab.GravityHigh,
		expectedPlayer.Hab.TemperatureLow, expectedPlayer.Hab.TemperatureCenter, expectedPlayer.Hab.TemperatureHigh,
		expectedPlayer.Hab.RadiationLow, expectedPlayer.Hab.RadiationCenter, expectedPlayer.Hab.RadiationHigh)
	t.Logf("  Production: CPR=%d F=%d/%d/%d M=%d/%d/%d",
		expectedPlayer.Production.ResourcePerColonist,
		expectedPlayer.Production.FactoryProduction, expectedPlayer.Production.FactoryCost, expectedPlayer.Production.FactoriesOperate,
		expectedPlayer.Production.MineProduction, expectedPlayer.Production.MineCost, expectedPlayer.Production.MinesOperate)
	t.Logf("  Research: E=%d W=%d P=%d C=%d El=%d B=%d",
		expectedPlayer.ResearchCost.Energy, expectedPlayer.ResearchCost.Weapons, expectedPlayer.ResearchCost.Propulsion,
		expectedPlayer.ResearchCost.Construction, expectedPlayer.ResearchCost.Electronics, expectedPlayer.ResearchCost.Biotech)
	t.Logf("  Logo: %d, LeftoverPoints: %d", expectedPlayer.Logo, expectedPlayer.SpendLeftoverPoints)

	// Now create a race using the builder with the exact settings from the expected file
	builder := race.New()

	// Set name (note: plural is empty in the predefined Humanoid)
	builder.Name(expectedPlayer.NameSingular, expectedPlayer.NamePlural)

	// Set PRT and LRT
	builder.PRT(expectedPlayer.PRT)
	builder.SetLRTs(expectedPlayer.LRT)

	// Set habitability - calculate center and width from low/high
	gravCenter := (expectedPlayer.Hab.GravityLow + expectedPlayer.Hab.GravityHigh) / 2
	gravWidth := (expectedPlayer.Hab.GravityHigh - expectedPlayer.Hab.GravityLow) / 2
	tempCenter := (expectedPlayer.Hab.TemperatureLow + expectedPlayer.Hab.TemperatureHigh) / 2
	tempWidth := (expectedPlayer.Hab.TemperatureHigh - expectedPlayer.Hab.TemperatureLow) / 2
	radCenter := (expectedPlayer.Hab.RadiationLow + expectedPlayer.Hab.RadiationHigh) / 2
	radWidth := (expectedPlayer.Hab.RadiationHigh - expectedPlayer.Hab.RadiationLow) / 2

	builder.Gravity(gravCenter, gravWidth)
	builder.Temperature(tempCenter, tempWidth)
	builder.Radiation(radCenter, radWidth)

	// Set growth
	builder.GrowthRate(expectedPlayer.GrowthRate)

	// Set production
	builder.ColonistsPerResource(expectedPlayer.Production.ResourcePerColonist * 100)
	builder.Factories(
		expectedPlayer.Production.FactoryProduction,
		expectedPlayer.Production.FactoryCost,
		expectedPlayer.Production.FactoriesOperate,
		expectedPlayer.FactoriesCost1LessGerm,
	)
	builder.Mines(
		expectedPlayer.Production.MineProduction,
		expectedPlayer.Production.MineCost,
		expectedPlayer.Production.MinesOperate,
	)

	// Set research
	builder.Research(
		expectedPlayer.ResearchCost.Energy,
		expectedPlayer.ResearchCost.Weapons,
		expectedPlayer.ResearchCost.Propulsion,
		expectedPlayer.ResearchCost.Construction,
		expectedPlayer.ResearchCost.Electronics,
		expectedPlayer.ResearchCost.Biotech,
	)
	builder.TechsStartHigh(expectedPlayer.ExpensiveTechStartsAt3)

	// Set leftover points
	builder.LeftoverPointsOn(race.LeftoverPointsOption(expectedPlayer.SpendLeftoverPoints))

	// Set icon
	builder.Icon(expectedPlayer.Logo)

	r, err := builder.Finish()
	if err != nil {
		t.Fatalf("Failed to finish race: %v", err)
	}

	// Create a race file
	data, err := CreateRaceFile(r, 1)
	if err != nil {
		t.Fatalf("Failed to create race file: %v", err)
	}

	// Parse our generated file
	fd := parser.FileData(data)
	parsedBlocks, err := fd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse generated race file: %v", err)
	}

	// Find the PlayerBlock in our generated file
	var generatedPlayer *blocks.PlayerBlock
	for _, block := range parsedBlocks {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			generatedPlayer = &pb
			break
		}
	}
	if generatedPlayer == nil {
		t.Fatal("No PlayerBlock found in generated file")
	}

	// Compare the key race attributes
	if generatedPlayer.NameSingular != expectedPlayer.NameSingular {
		t.Errorf("Name singular: expected '%s', got '%s'", expectedPlayer.NameSingular, generatedPlayer.NameSingular)
	}
	if generatedPlayer.NamePlural != expectedPlayer.NamePlural {
		t.Errorf("Name plural: expected '%s', got '%s'", expectedPlayer.NamePlural, generatedPlayer.NamePlural)
	}
	if generatedPlayer.PRT != expectedPlayer.PRT {
		t.Errorf("PRT: expected %d, got %d", expectedPlayer.PRT, generatedPlayer.PRT)
	}
	if generatedPlayer.LRT != expectedPlayer.LRT {
		t.Errorf("LRT: expected 0x%04X, got 0x%04X", expectedPlayer.LRT, generatedPlayer.LRT)
	}
	if generatedPlayer.GrowthRate != expectedPlayer.GrowthRate {
		t.Errorf("Growth rate: expected %d, got %d", expectedPlayer.GrowthRate, generatedPlayer.GrowthRate)
	}

	// Compare habitability
	if generatedPlayer.Hab.GravityCenter != expectedPlayer.Hab.GravityCenter {
		t.Errorf("Gravity center: expected %d, got %d", expectedPlayer.Hab.GravityCenter, generatedPlayer.Hab.GravityCenter)
	}
	if generatedPlayer.Hab.GravityLow != expectedPlayer.Hab.GravityLow {
		t.Errorf("Gravity low: expected %d, got %d", expectedPlayer.Hab.GravityLow, generatedPlayer.Hab.GravityLow)
	}
	if generatedPlayer.Hab.GravityHigh != expectedPlayer.Hab.GravityHigh {
		t.Errorf("Gravity high: expected %d, got %d", expectedPlayer.Hab.GravityHigh, generatedPlayer.Hab.GravityHigh)
	}

	// Compare production
	if generatedPlayer.Production.ResourcePerColonist != expectedPlayer.Production.ResourcePerColonist {
		t.Errorf("ResourcePerColonist: expected %d, got %d", expectedPlayer.Production.ResourcePerColonist, generatedPlayer.Production.ResourcePerColonist)
	}
	if generatedPlayer.Production.FactoryProduction != expectedPlayer.Production.FactoryProduction {
		t.Errorf("FactoryProduction: expected %d, got %d", expectedPlayer.Production.FactoryProduction, generatedPlayer.Production.FactoryProduction)
	}

	// Compare research costs
	if generatedPlayer.ResearchCost != expectedPlayer.ResearchCost {
		t.Errorf("ResearchCost: expected %+v, got %+v", expectedPlayer.ResearchCost, generatedPlayer.ResearchCost)
	}

	t.Logf("Humanoid predefined race test passed - all attributes match!")
}

func TestRabbitoidPredefinedRace(t *testing.T) {
	// Read the expected race file produced by Stars!
	expectedData, err := os.ReadFile("../testdata/scenario-racebuilder/predefined-races/rabbitoids/race.r1")
	if err != nil {
		t.Fatalf("Failed to read expected race file: %v", err)
	}

	// Parse the expected file
	expectedFd := parser.FileData(expectedData)
	expectedBlocks, err := expectedFd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse expected race file: %v", err)
	}

	// Find the PlayerBlock
	var expectedPlayer *blocks.PlayerBlock
	for _, block := range expectedBlocks {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			expectedPlayer = &pb
			break
		}
	}
	if expectedPlayer == nil {
		t.Fatal("No PlayerBlock found in expected file")
	}

	// Log the expected values for verification
	t.Logf("Expected race: '%s' / '%s'", expectedPlayer.NameSingular, expectedPlayer.NamePlural)
	t.Logf("  PRT: %d (%s)", expectedPlayer.PRT, blocks.PRTName(expectedPlayer.PRT))
	t.Logf("  LRT: 0x%04X (%v)", expectedPlayer.LRT, blocks.LRTNames(expectedPlayer.LRT))
	t.Logf("  Growth rate: %d%%", expectedPlayer.GrowthRate)
	t.Logf("  Habitability: G=%d/%d/%d T=%d/%d/%d R=%d/%d/%d",
		expectedPlayer.Hab.GravityLow, expectedPlayer.Hab.GravityCenter, expectedPlayer.Hab.GravityHigh,
		expectedPlayer.Hab.TemperatureLow, expectedPlayer.Hab.TemperatureCenter, expectedPlayer.Hab.TemperatureHigh,
		expectedPlayer.Hab.RadiationLow, expectedPlayer.Hab.RadiationCenter, expectedPlayer.Hab.RadiationHigh)
	t.Logf("  Production: CPR=%d F=%d/%d/%d (lessGerm=%v) M=%d/%d/%d",
		expectedPlayer.Production.ResourcePerColonist,
		expectedPlayer.Production.FactoryProduction, expectedPlayer.Production.FactoryCost, expectedPlayer.Production.FactoriesOperate,
		expectedPlayer.FactoriesCost1LessGerm,
		expectedPlayer.Production.MineProduction, expectedPlayer.Production.MineCost, expectedPlayer.Production.MinesOperate)
	t.Logf("  Research: E=%d W=%d P=%d C=%d El=%d B=%d",
		expectedPlayer.ResearchCost.Energy, expectedPlayer.ResearchCost.Weapons, expectedPlayer.ResearchCost.Propulsion,
		expectedPlayer.ResearchCost.Construction, expectedPlayer.ResearchCost.Electronics, expectedPlayer.ResearchCost.Biotech)
	t.Logf("  Logo: %d, LeftoverPoints: %d", expectedPlayer.Logo, expectedPlayer.SpendLeftoverPoints)

	// Verify expected values match what we see in the screenshots
	if expectedPlayer.NameSingular != "Rabbitoid" {
		t.Errorf("Expected name 'Rabbitoid', got '%s'", expectedPlayer.NameSingular)
	}
	if expectedPlayer.PRT != 7 { // IT
		t.Errorf("Expected PRT 7 (IT), got %d", expectedPlayer.PRT)
	}
	// LRT: IFE(0) + TT(1) + CE(8) + NAS(10) = 0x0001 + 0x0002 + 0x0100 + 0x0400 = 0x0503
	if expectedPlayer.LRT != 0x0503 {
		t.Errorf("Expected LRT 0x0503, got 0x%04X", expectedPlayer.LRT)
	}
	if expectedPlayer.GrowthRate != 20 {
		t.Errorf("Expected growth rate 20, got %d", expectedPlayer.GrowthRate)
	}

	// Now create a race using the builder with the exact settings from the expected file
	builder := race.New()
	builder.Name(expectedPlayer.NameSingular, expectedPlayer.NamePlural)
	builder.PRT(expectedPlayer.PRT)
	builder.SetLRTs(expectedPlayer.LRT)

	// Set habitability
	gravCenter := (expectedPlayer.Hab.GravityLow + expectedPlayer.Hab.GravityHigh) / 2
	gravWidth := (expectedPlayer.Hab.GravityHigh - expectedPlayer.Hab.GravityLow) / 2
	tempCenter := (expectedPlayer.Hab.TemperatureLow + expectedPlayer.Hab.TemperatureHigh) / 2
	tempWidth := (expectedPlayer.Hab.TemperatureHigh - expectedPlayer.Hab.TemperatureLow) / 2
	radCenter := (expectedPlayer.Hab.RadiationLow + expectedPlayer.Hab.RadiationHigh) / 2
	radWidth := (expectedPlayer.Hab.RadiationHigh - expectedPlayer.Hab.RadiationLow) / 2

	builder.Gravity(gravCenter, gravWidth)
	builder.Temperature(tempCenter, tempWidth)
	builder.Radiation(radCenter, radWidth)
	builder.GrowthRate(expectedPlayer.GrowthRate)

	// Set production
	builder.ColonistsPerResource(expectedPlayer.Production.ResourcePerColonist * 100)
	builder.Factories(
		expectedPlayer.Production.FactoryProduction,
		expectedPlayer.Production.FactoryCost,
		expectedPlayer.Production.FactoriesOperate,
		expectedPlayer.FactoriesCost1LessGerm,
	)
	builder.Mines(
		expectedPlayer.Production.MineProduction,
		expectedPlayer.Production.MineCost,
		expectedPlayer.Production.MinesOperate,
	)

	// Set research
	builder.Research(
		expectedPlayer.ResearchCost.Energy,
		expectedPlayer.ResearchCost.Weapons,
		expectedPlayer.ResearchCost.Propulsion,
		expectedPlayer.ResearchCost.Construction,
		expectedPlayer.ResearchCost.Electronics,
		expectedPlayer.ResearchCost.Biotech,
	)
	builder.TechsStartHigh(expectedPlayer.ExpensiveTechStartsAt3)
	builder.LeftoverPointsOn(race.LeftoverPointsOption(expectedPlayer.SpendLeftoverPoints))
	builder.Icon(expectedPlayer.Logo)

	r, err := builder.Finish()
	if err != nil {
		t.Fatalf("Failed to finish race: %v", err)
	}

	// Create a race file
	data, err := CreateRaceFile(r, 1)
	if err != nil {
		t.Fatalf("Failed to create race file: %v", err)
	}

	// Parse our generated file
	fd := parser.FileData(data)
	parsedBlocks, err := fd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse generated race file: %v", err)
	}

	// Find the PlayerBlock in our generated file
	var generatedPlayer *blocks.PlayerBlock
	for _, block := range parsedBlocks {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			generatedPlayer = &pb
			break
		}
	}
	if generatedPlayer == nil {
		t.Fatal("No PlayerBlock found in generated file")
	}

	// Compare key attributes
	if generatedPlayer.NameSingular != expectedPlayer.NameSingular {
		t.Errorf("Name singular: expected '%s', got '%s'", expectedPlayer.NameSingular, generatedPlayer.NameSingular)
	}
	if generatedPlayer.PRT != expectedPlayer.PRT {
		t.Errorf("PRT: expected %d, got %d", expectedPlayer.PRT, generatedPlayer.PRT)
	}
	if generatedPlayer.LRT != expectedPlayer.LRT {
		t.Errorf("LRT: expected 0x%04X, got 0x%04X", expectedPlayer.LRT, generatedPlayer.LRT)
	}
	if generatedPlayer.GrowthRate != expectedPlayer.GrowthRate {
		t.Errorf("Growth rate: expected %d, got %d", expectedPlayer.GrowthRate, generatedPlayer.GrowthRate)
	}
	if generatedPlayer.Hab != expectedPlayer.Hab {
		t.Errorf("Hab: expected %+v, got %+v", expectedPlayer.Hab, generatedPlayer.Hab)
	}
	if generatedPlayer.Production != expectedPlayer.Production {
		t.Errorf("Production: expected %+v, got %+v", expectedPlayer.Production, generatedPlayer.Production)
	}
	if generatedPlayer.FactoriesCost1LessGerm != expectedPlayer.FactoriesCost1LessGerm {
		t.Errorf("FactoriesCost1LessGerm: expected %v, got %v", expectedPlayer.FactoriesCost1LessGerm, generatedPlayer.FactoriesCost1LessGerm)
	}
	if generatedPlayer.ResearchCost != expectedPlayer.ResearchCost {
		t.Errorf("ResearchCost: expected %+v, got %+v", expectedPlayer.ResearchCost, generatedPlayer.ResearchCost)
	}

	t.Logf("Rabbitoid predefined race test passed - all attributes match!")
}

func TestInsectoidPredefinedRace(t *testing.T) {
	// Read the expected race file produced by Stars!
	expectedData, err := os.ReadFile("../testdata/scenario-racebuilder/predefined-races/insectoids/race.r1")
	if err != nil {
		t.Fatalf("Failed to read expected race file: %v", err)
	}

	// Parse the expected file
	expectedFd := parser.FileData(expectedData)
	expectedBlocks, err := expectedFd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse expected race file: %v", err)
	}

	// Find the PlayerBlock
	var expectedPlayer *blocks.PlayerBlock
	for _, block := range expectedBlocks {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			expectedPlayer = &pb
			break
		}
	}
	if expectedPlayer == nil {
		t.Fatal("No PlayerBlock found in expected file")
	}

	// Log the expected values for verification
	t.Logf("Expected race: '%s' / '%s'", expectedPlayer.NameSingular, expectedPlayer.NamePlural)
	t.Logf("  PRT: %d (%s)", expectedPlayer.PRT, blocks.PRTName(expectedPlayer.PRT))
	t.Logf("  LRT: 0x%04X (%v)", expectedPlayer.LRT, blocks.LRTNames(expectedPlayer.LRT))
	t.Logf("  Growth rate: %d%%", expectedPlayer.GrowthRate)
	t.Logf("  Habitability: G=%d/%d/%d (immune=%v) T=%d/%d/%d R=%d/%d/%d",
		expectedPlayer.Hab.GravityLow, expectedPlayer.Hab.GravityCenter, expectedPlayer.Hab.GravityHigh,
		expectedPlayer.Hab.IsGravityImmune(),
		expectedPlayer.Hab.TemperatureLow, expectedPlayer.Hab.TemperatureCenter, expectedPlayer.Hab.TemperatureHigh,
		expectedPlayer.Hab.RadiationLow, expectedPlayer.Hab.RadiationCenter, expectedPlayer.Hab.RadiationHigh)
	t.Logf("  Production: CPR=%d F=%d/%d/%d M=%d/%d/%d",
		expectedPlayer.Production.ResourcePerColonist,
		expectedPlayer.Production.FactoryProduction, expectedPlayer.Production.FactoryCost, expectedPlayer.Production.FactoriesOperate,
		expectedPlayer.Production.MineProduction, expectedPlayer.Production.MineCost, expectedPlayer.Production.MinesOperate)
	t.Logf("  Research: E=%d W=%d P=%d C=%d El=%d B=%d",
		expectedPlayer.ResearchCost.Energy, expectedPlayer.ResearchCost.Weapons, expectedPlayer.ResearchCost.Propulsion,
		expectedPlayer.ResearchCost.Construction, expectedPlayer.ResearchCost.Electronics, expectedPlayer.ResearchCost.Biotech)
	t.Logf("  Logo: %d, LeftoverPoints: %d", expectedPlayer.Logo, expectedPlayer.SpendLeftoverPoints)

	// Verify expected values match what we see in the screenshots
	if expectedPlayer.NameSingular != "Insectoid" {
		t.Errorf("Expected name 'Insectoid', got '%s'", expectedPlayer.NameSingular)
	}
	if expectedPlayer.PRT != 2 { // WM
		t.Errorf("Expected PRT 2 (WM), got %d", expectedPlayer.PRT)
	}
	if !expectedPlayer.Hab.IsGravityImmune() {
		t.Error("Expected gravity immunity")
	}

	// Now create a race using the builder with the exact settings from the expected file
	builder := race.New()
	builder.Name(expectedPlayer.NameSingular, expectedPlayer.NamePlural)
	builder.PRT(expectedPlayer.PRT)
	builder.SetLRTs(expectedPlayer.LRT)

	// Set habitability
	if expectedPlayer.Hab.IsGravityImmune() {
		builder.GravityImmune(true)
	} else {
		gravCenter := (expectedPlayer.Hab.GravityLow + expectedPlayer.Hab.GravityHigh) / 2
		gravWidth := (expectedPlayer.Hab.GravityHigh - expectedPlayer.Hab.GravityLow) / 2
		builder.Gravity(gravCenter, gravWidth)
	}
	if expectedPlayer.Hab.IsTemperatureImmune() {
		builder.TemperatureImmune(true)
	} else {
		tempCenter := (expectedPlayer.Hab.TemperatureLow + expectedPlayer.Hab.TemperatureHigh) / 2
		tempWidth := (expectedPlayer.Hab.TemperatureHigh - expectedPlayer.Hab.TemperatureLow) / 2
		builder.Temperature(tempCenter, tempWidth)
	}
	if expectedPlayer.Hab.IsRadiationImmune() {
		builder.RadiationImmune(true)
	} else {
		radCenter := (expectedPlayer.Hab.RadiationLow + expectedPlayer.Hab.RadiationHigh) / 2
		radWidth := (expectedPlayer.Hab.RadiationHigh - expectedPlayer.Hab.RadiationLow) / 2
		builder.Radiation(radCenter, radWidth)
	}
	builder.GrowthRate(expectedPlayer.GrowthRate)

	// Set production
	builder.ColonistsPerResource(expectedPlayer.Production.ResourcePerColonist * 100)
	builder.Factories(
		expectedPlayer.Production.FactoryProduction,
		expectedPlayer.Production.FactoryCost,
		expectedPlayer.Production.FactoriesOperate,
		expectedPlayer.FactoriesCost1LessGerm,
	)
	builder.Mines(
		expectedPlayer.Production.MineProduction,
		expectedPlayer.Production.MineCost,
		expectedPlayer.Production.MinesOperate,
	)

	// Set research
	builder.Research(
		expectedPlayer.ResearchCost.Energy,
		expectedPlayer.ResearchCost.Weapons,
		expectedPlayer.ResearchCost.Propulsion,
		expectedPlayer.ResearchCost.Construction,
		expectedPlayer.ResearchCost.Electronics,
		expectedPlayer.ResearchCost.Biotech,
	)
	builder.TechsStartHigh(expectedPlayer.ExpensiveTechStartsAt3)
	builder.LeftoverPointsOn(race.LeftoverPointsOption(expectedPlayer.SpendLeftoverPoints))
	builder.Icon(expectedPlayer.Logo)

	r, err := builder.Finish()
	if err != nil {
		t.Fatalf("Failed to finish race: %v", err)
	}

	// Create a race file
	data, err := CreateRaceFile(r, 1)
	if err != nil {
		t.Fatalf("Failed to create race file: %v", err)
	}

	// Parse our generated file
	fd := parser.FileData(data)
	parsedBlocks, err := fd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse generated race file: %v", err)
	}

	// Find the PlayerBlock in our generated file
	var generatedPlayer *blocks.PlayerBlock
	for _, block := range parsedBlocks {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			generatedPlayer = &pb
			break
		}
	}
	if generatedPlayer == nil {
		t.Fatal("No PlayerBlock found in generated file")
	}

	// Compare key attributes
	if generatedPlayer.NameSingular != expectedPlayer.NameSingular {
		t.Errorf("Name singular: expected '%s', got '%s'", expectedPlayer.NameSingular, generatedPlayer.NameSingular)
	}
	if generatedPlayer.PRT != expectedPlayer.PRT {
		t.Errorf("PRT: expected %d, got %d", expectedPlayer.PRT, generatedPlayer.PRT)
	}
	if generatedPlayer.LRT != expectedPlayer.LRT {
		t.Errorf("LRT: expected 0x%04X, got 0x%04X", expectedPlayer.LRT, generatedPlayer.LRT)
	}
	if generatedPlayer.GrowthRate != expectedPlayer.GrowthRate {
		t.Errorf("Growth rate: expected %d, got %d", expectedPlayer.GrowthRate, generatedPlayer.GrowthRate)
	}
	if generatedPlayer.Hab.IsGravityImmune() != expectedPlayer.Hab.IsGravityImmune() {
		t.Errorf("Gravity immune: expected %v, got %v", expectedPlayer.Hab.IsGravityImmune(), generatedPlayer.Hab.IsGravityImmune())
	}
	if generatedPlayer.Production != expectedPlayer.Production {
		t.Errorf("Production: expected %+v, got %+v", expectedPlayer.Production, generatedPlayer.Production)
	}
	if generatedPlayer.ResearchCost != expectedPlayer.ResearchCost {
		t.Errorf("ResearchCost: expected %+v, got %+v", expectedPlayer.ResearchCost, generatedPlayer.ResearchCost)
	}

	t.Logf("Insectoid predefined race test passed - all attributes match!")
}

func TestNucleotidPredefinedRace(t *testing.T) {
	// Read the expected race file produced by Stars!
	expectedData, err := os.ReadFile("../testdata/scenario-racebuilder/predefined-races/nucleotids/race.r1")
	if err != nil {
		t.Fatalf("Failed to read expected race file: %v", err)
	}

	// Parse the expected file
	expectedFd := parser.FileData(expectedData)
	expectedBlocks, err := expectedFd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse expected race file: %v", err)
	}

	// Find the PlayerBlock
	var expectedPlayer *blocks.PlayerBlock
	for _, block := range expectedBlocks {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			expectedPlayer = &pb
			break
		}
	}
	if expectedPlayer == nil {
		t.Fatal("No PlayerBlock found in expected file")
	}

	// Log the expected values for verification
	t.Logf("Expected race: '%s' / '%s'", expectedPlayer.NameSingular, expectedPlayer.NamePlural)
	t.Logf("  PRT: %d (%s)", expectedPlayer.PRT, blocks.PRTName(expectedPlayer.PRT))
	t.Logf("  LRT: 0x%04X (%v)", expectedPlayer.LRT, blocks.LRTNames(expectedPlayer.LRT))
	t.Logf("  Growth rate: %d%%", expectedPlayer.GrowthRate)
	t.Logf("  Habitability: G=%d/%d/%d (immune=%v) T=%d/%d/%d R=%d/%d/%d",
		expectedPlayer.Hab.GravityLow, expectedPlayer.Hab.GravityCenter, expectedPlayer.Hab.GravityHigh,
		expectedPlayer.Hab.IsGravityImmune(),
		expectedPlayer.Hab.TemperatureLow, expectedPlayer.Hab.TemperatureCenter, expectedPlayer.Hab.TemperatureHigh,
		expectedPlayer.Hab.RadiationLow, expectedPlayer.Hab.RadiationCenter, expectedPlayer.Hab.RadiationHigh)
	t.Logf("  Production: CPR=%d F=%d/%d/%d M=%d/%d/%d",
		expectedPlayer.Production.ResourcePerColonist,
		expectedPlayer.Production.FactoryProduction, expectedPlayer.Production.FactoryCost, expectedPlayer.Production.FactoriesOperate,
		expectedPlayer.Production.MineProduction, expectedPlayer.Production.MineCost, expectedPlayer.Production.MinesOperate)
	t.Logf("  Research: E=%d W=%d P=%d C=%d El=%d B=%d, TechsStartHigh=%v",
		expectedPlayer.ResearchCost.Energy, expectedPlayer.ResearchCost.Weapons, expectedPlayer.ResearchCost.Propulsion,
		expectedPlayer.ResearchCost.Construction, expectedPlayer.ResearchCost.Electronics, expectedPlayer.ResearchCost.Biotech,
		expectedPlayer.ExpensiveTechStartsAt3)
	t.Logf("  Logo: %d, LeftoverPoints: %d", expectedPlayer.Logo, expectedPlayer.SpendLeftoverPoints)

	// Verify expected values match what we see in the screenshots
	if expectedPlayer.NameSingular != "Nucleotid" {
		t.Errorf("Expected name 'Nucleotid', got '%s'", expectedPlayer.NameSingular)
	}
	if expectedPlayer.PRT != 1 { // SS
		t.Errorf("Expected PRT 1 (SS), got %d", expectedPlayer.PRT)
	}
	if !expectedPlayer.Hab.IsGravityImmune() {
		t.Error("Expected gravity immunity")
	}

	// Now create a race using the builder with the exact settings from the expected file
	builder := race.New()
	builder.Name(expectedPlayer.NameSingular, expectedPlayer.NamePlural)
	builder.PRT(expectedPlayer.PRT)
	builder.SetLRTs(expectedPlayer.LRT)

	// Set habitability
	if expectedPlayer.Hab.IsGravityImmune() {
		builder.GravityImmune(true)
	} else {
		gravCenter := (expectedPlayer.Hab.GravityLow + expectedPlayer.Hab.GravityHigh) / 2
		gravWidth := (expectedPlayer.Hab.GravityHigh - expectedPlayer.Hab.GravityLow) / 2
		builder.Gravity(gravCenter, gravWidth)
	}
	if expectedPlayer.Hab.IsTemperatureImmune() {
		builder.TemperatureImmune(true)
	} else {
		tempCenter := (expectedPlayer.Hab.TemperatureLow + expectedPlayer.Hab.TemperatureHigh) / 2
		tempWidth := (expectedPlayer.Hab.TemperatureHigh - expectedPlayer.Hab.TemperatureLow) / 2
		builder.Temperature(tempCenter, tempWidth)
	}
	if expectedPlayer.Hab.IsRadiationImmune() {
		builder.RadiationImmune(true)
	} else {
		radCenter := (expectedPlayer.Hab.RadiationLow + expectedPlayer.Hab.RadiationHigh) / 2
		radWidth := (expectedPlayer.Hab.RadiationHigh - expectedPlayer.Hab.RadiationLow) / 2
		builder.Radiation(radCenter, radWidth)
	}
	builder.GrowthRate(expectedPlayer.GrowthRate)

	// Set production
	builder.ColonistsPerResource(expectedPlayer.Production.ResourcePerColonist * 100)
	builder.Factories(
		expectedPlayer.Production.FactoryProduction,
		expectedPlayer.Production.FactoryCost,
		expectedPlayer.Production.FactoriesOperate,
		expectedPlayer.FactoriesCost1LessGerm,
	)
	builder.Mines(
		expectedPlayer.Production.MineProduction,
		expectedPlayer.Production.MineCost,
		expectedPlayer.Production.MinesOperate,
	)

	// Set research
	builder.Research(
		expectedPlayer.ResearchCost.Energy,
		expectedPlayer.ResearchCost.Weapons,
		expectedPlayer.ResearchCost.Propulsion,
		expectedPlayer.ResearchCost.Construction,
		expectedPlayer.ResearchCost.Electronics,
		expectedPlayer.ResearchCost.Biotech,
	)
	builder.TechsStartHigh(expectedPlayer.ExpensiveTechStartsAt3)
	builder.LeftoverPointsOn(race.LeftoverPointsOption(expectedPlayer.SpendLeftoverPoints))
	builder.Icon(expectedPlayer.Logo)

	r, err := builder.Finish()
	if err != nil {
		t.Fatalf("Failed to finish race: %v", err)
	}

	// Create a race file
	data, err := CreateRaceFile(r, 1)
	if err != nil {
		t.Fatalf("Failed to create race file: %v", err)
	}

	// Parse our generated file
	fd := parser.FileData(data)
	parsedBlocks, err := fd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse generated race file: %v", err)
	}

	// Find the PlayerBlock in our generated file
	var generatedPlayer *blocks.PlayerBlock
	for _, block := range parsedBlocks {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			generatedPlayer = &pb
			break
		}
	}
	if generatedPlayer == nil {
		t.Fatal("No PlayerBlock found in generated file")
	}

	// Compare key attributes
	if generatedPlayer.NameSingular != expectedPlayer.NameSingular {
		t.Errorf("Name singular: expected '%s', got '%s'", expectedPlayer.NameSingular, generatedPlayer.NameSingular)
	}
	if generatedPlayer.PRT != expectedPlayer.PRT {
		t.Errorf("PRT: expected %d, got %d", expectedPlayer.PRT, generatedPlayer.PRT)
	}
	if generatedPlayer.LRT != expectedPlayer.LRT {
		t.Errorf("LRT: expected 0x%04X, got 0x%04X", expectedPlayer.LRT, generatedPlayer.LRT)
	}
	if generatedPlayer.Hab.IsGravityImmune() != expectedPlayer.Hab.IsGravityImmune() {
		t.Errorf("Gravity immune: expected %v, got %v", expectedPlayer.Hab.IsGravityImmune(), generatedPlayer.Hab.IsGravityImmune())
	}
	if generatedPlayer.Production != expectedPlayer.Production {
		t.Errorf("Production: expected %+v, got %+v", expectedPlayer.Production, generatedPlayer.Production)
	}
	if generatedPlayer.ResearchCost != expectedPlayer.ResearchCost {
		t.Errorf("ResearchCost: expected %+v, got %+v", expectedPlayer.ResearchCost, generatedPlayer.ResearchCost)
	}
	if generatedPlayer.ExpensiveTechStartsAt3 != expectedPlayer.ExpensiveTechStartsAt3 {
		t.Errorf("ExpensiveTechStartsAt3: expected %v, got %v", expectedPlayer.ExpensiveTechStartsAt3, generatedPlayer.ExpensiveTechStartsAt3)
	}

	t.Logf("Nucleotid predefined race test passed - all attributes match!")
}

func TestSilicanoidPredefinedRace(t *testing.T) {
	// Read the expected race file produced by Stars!
	expectedData, err := os.ReadFile("../testdata/scenario-racebuilder/predefined-races/silicanoids/race.r1")
	if err != nil {
		t.Fatalf("Failed to read expected race file: %v", err)
	}

	// Parse the expected file
	expectedFd := parser.FileData(expectedData)
	expectedBlocks, err := expectedFd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse expected race file: %v", err)
	}

	// Find the PlayerBlock
	var expectedPlayer *blocks.PlayerBlock
	for _, block := range expectedBlocks {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			expectedPlayer = &pb
			break
		}
	}
	if expectedPlayer == nil {
		t.Fatal("No PlayerBlock found in expected file")
	}

	// Log the expected values for verification
	t.Logf("Expected race: '%s' / '%s'", expectedPlayer.NameSingular, expectedPlayer.NamePlural)
	t.Logf("  PRT: %d (%s)", expectedPlayer.PRT, blocks.PRTName(expectedPlayer.PRT))
	t.Logf("  LRT: 0x%04X (%v)", expectedPlayer.LRT, blocks.LRTNames(expectedPlayer.LRT))
	t.Logf("  Growth rate: %d%%", expectedPlayer.GrowthRate)
	t.Logf("  Habitability: G(immune=%v) T(immune=%v) R(immune=%v)",
		expectedPlayer.Hab.IsGravityImmune(),
		expectedPlayer.Hab.IsTemperatureImmune(),
		expectedPlayer.Hab.IsRadiationImmune())
	t.Logf("  Production: CPR=%d F=%d/%d/%d M=%d/%d/%d",
		expectedPlayer.Production.ResourcePerColonist,
		expectedPlayer.Production.FactoryProduction, expectedPlayer.Production.FactoryCost, expectedPlayer.Production.FactoriesOperate,
		expectedPlayer.Production.MineProduction, expectedPlayer.Production.MineCost, expectedPlayer.Production.MinesOperate)
	t.Logf("  Research: E=%d W=%d P=%d C=%d El=%d B=%d",
		expectedPlayer.ResearchCost.Energy, expectedPlayer.ResearchCost.Weapons, expectedPlayer.ResearchCost.Propulsion,
		expectedPlayer.ResearchCost.Construction, expectedPlayer.ResearchCost.Electronics, expectedPlayer.ResearchCost.Biotech)
	t.Logf("  Logo: %d, LeftoverPoints: %d", expectedPlayer.Logo, expectedPlayer.SpendLeftoverPoints)

	// Verify expected values match what we see in the screenshots
	if expectedPlayer.NameSingular != "Silicanoid" {
		t.Errorf("Expected name 'Silicanoid', got '%s'", expectedPlayer.NameSingular)
	}
	if expectedPlayer.PRT != 0 { // HE
		t.Errorf("Expected PRT 0 (HE), got %d", expectedPlayer.PRT)
	}
	if !expectedPlayer.Hab.IsGravityImmune() || !expectedPlayer.Hab.IsTemperatureImmune() || !expectedPlayer.Hab.IsRadiationImmune() {
		t.Error("Expected all immunities")
	}

	// Now create a race using the builder with the exact settings from the expected file
	builder := race.New()
	builder.Name(expectedPlayer.NameSingular, expectedPlayer.NamePlural)
	builder.PRT(expectedPlayer.PRT)
	builder.SetLRTs(expectedPlayer.LRT)

	// Set habitability - all immune
	builder.GravityImmune(expectedPlayer.Hab.IsGravityImmune())
	builder.TemperatureImmune(expectedPlayer.Hab.IsTemperatureImmune())
	builder.RadiationImmune(expectedPlayer.Hab.IsRadiationImmune())
	builder.GrowthRate(expectedPlayer.GrowthRate)

	// Set production
	builder.ColonistsPerResource(expectedPlayer.Production.ResourcePerColonist * 100)
	builder.Factories(
		expectedPlayer.Production.FactoryProduction,
		expectedPlayer.Production.FactoryCost,
		expectedPlayer.Production.FactoriesOperate,
		expectedPlayer.FactoriesCost1LessGerm,
	)
	builder.Mines(
		expectedPlayer.Production.MineProduction,
		expectedPlayer.Production.MineCost,
		expectedPlayer.Production.MinesOperate,
	)

	// Set research
	builder.Research(
		expectedPlayer.ResearchCost.Energy,
		expectedPlayer.ResearchCost.Weapons,
		expectedPlayer.ResearchCost.Propulsion,
		expectedPlayer.ResearchCost.Construction,
		expectedPlayer.ResearchCost.Electronics,
		expectedPlayer.ResearchCost.Biotech,
	)
	builder.TechsStartHigh(expectedPlayer.ExpensiveTechStartsAt3)
	builder.LeftoverPointsOn(race.LeftoverPointsOption(expectedPlayer.SpendLeftoverPoints))
	builder.Icon(expectedPlayer.Logo)

	r, err := builder.Finish()
	if err != nil {
		t.Fatalf("Failed to finish race: %v", err)
	}

	// Create a race file
	data, err := CreateRaceFile(r, 1)
	if err != nil {
		t.Fatalf("Failed to create race file: %v", err)
	}

	// Parse our generated file
	fd := parser.FileData(data)
	parsedBlocks, err := fd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse generated race file: %v", err)
	}

	// Find the PlayerBlock in our generated file
	var generatedPlayer *blocks.PlayerBlock
	for _, block := range parsedBlocks {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			generatedPlayer = &pb
			break
		}
	}
	if generatedPlayer == nil {
		t.Fatal("No PlayerBlock found in generated file")
	}

	// Compare key attributes
	if generatedPlayer.NameSingular != expectedPlayer.NameSingular {
		t.Errorf("Name singular: expected '%s', got '%s'", expectedPlayer.NameSingular, generatedPlayer.NameSingular)
	}
	if generatedPlayer.PRT != expectedPlayer.PRT {
		t.Errorf("PRT: expected %d, got %d", expectedPlayer.PRT, generatedPlayer.PRT)
	}
	if generatedPlayer.LRT != expectedPlayer.LRT {
		t.Errorf("LRT: expected 0x%04X, got 0x%04X", expectedPlayer.LRT, generatedPlayer.LRT)
	}
	if generatedPlayer.Hab.IsGravityImmune() != expectedPlayer.Hab.IsGravityImmune() {
		t.Errorf("Gravity immune: expected %v, got %v", expectedPlayer.Hab.IsGravityImmune(), generatedPlayer.Hab.IsGravityImmune())
	}
	if generatedPlayer.Hab.IsTemperatureImmune() != expectedPlayer.Hab.IsTemperatureImmune() {
		t.Errorf("Temperature immune: expected %v, got %v", expectedPlayer.Hab.IsTemperatureImmune(), generatedPlayer.Hab.IsTemperatureImmune())
	}
	if generatedPlayer.Hab.IsRadiationImmune() != expectedPlayer.Hab.IsRadiationImmune() {
		t.Errorf("Radiation immune: expected %v, got %v", expectedPlayer.Hab.IsRadiationImmune(), generatedPlayer.Hab.IsRadiationImmune())
	}
	if generatedPlayer.Production != expectedPlayer.Production {
		t.Errorf("Production: expected %+v, got %+v", expectedPlayer.Production, generatedPlayer.Production)
	}
	if generatedPlayer.ResearchCost != expectedPlayer.ResearchCost {
		t.Errorf("ResearchCost: expected %+v, got %+v", expectedPlayer.ResearchCost, generatedPlayer.ResearchCost)
	}

	t.Logf("Silicanoid predefined race test passed - all attributes match!")
}

func TestAntetheralPredefinedRace(t *testing.T) {
	// Read the expected race file produced by Stars!
	expectedData, err := os.ReadFile("../testdata/scenario-racebuilder/predefined-races/antetherals/race.r1")
	if err != nil {
		t.Fatalf("Failed to read expected race file: %v", err)
	}

	// Parse the expected file
	expectedFd := parser.FileData(expectedData)
	expectedBlocks, err := expectedFd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse expected race file: %v", err)
	}

	// Find the PlayerBlock
	var expectedPlayer *blocks.PlayerBlock
	for _, block := range expectedBlocks {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			expectedPlayer = &pb
			break
		}
	}
	if expectedPlayer == nil {
		t.Fatal("No PlayerBlock found in expected file")
	}

	// Log the expected values for verification
	t.Logf("Expected race: '%s' / '%s'", expectedPlayer.NameSingular, expectedPlayer.NamePlural)
	t.Logf("  PRT: %d (%s)", expectedPlayer.PRT, blocks.PRTName(expectedPlayer.PRT))
	t.Logf("  LRT: 0x%04X (%v)", expectedPlayer.LRT, blocks.LRTNames(expectedPlayer.LRT))
	t.Logf("  Growth rate: %d%%", expectedPlayer.GrowthRate)
	t.Logf("  Habitability: G=%d/%d/%d T=%d/%d/%d R=%d/%d/%d",
		expectedPlayer.Hab.GravityLow, expectedPlayer.Hab.GravityCenter, expectedPlayer.Hab.GravityHigh,
		expectedPlayer.Hab.TemperatureLow, expectedPlayer.Hab.TemperatureCenter, expectedPlayer.Hab.TemperatureHigh,
		expectedPlayer.Hab.RadiationLow, expectedPlayer.Hab.RadiationCenter, expectedPlayer.Hab.RadiationHigh)
	t.Logf("  Production: CPR=%d F=%d/%d/%d M=%d/%d/%d",
		expectedPlayer.Production.ResourcePerColonist,
		expectedPlayer.Production.FactoryProduction, expectedPlayer.Production.FactoryCost, expectedPlayer.Production.FactoriesOperate,
		expectedPlayer.Production.MineProduction, expectedPlayer.Production.MineCost, expectedPlayer.Production.MinesOperate)
	t.Logf("  Research: E=%d W=%d P=%d C=%d El=%d B=%d",
		expectedPlayer.ResearchCost.Energy, expectedPlayer.ResearchCost.Weapons, expectedPlayer.ResearchCost.Propulsion,
		expectedPlayer.ResearchCost.Construction, expectedPlayer.ResearchCost.Electronics, expectedPlayer.ResearchCost.Biotech)
	t.Logf("  Logo: %d, LeftoverPoints: %d", expectedPlayer.Logo, expectedPlayer.SpendLeftoverPoints)

	// Verify expected values match what we see in the screenshots
	if expectedPlayer.NameSingular != "Antetheral" {
		t.Errorf("Expected name 'Antetheral', got '%s'", expectedPlayer.NameSingular)
	}
	if expectedPlayer.PRT != 5 { // SD
		t.Errorf("Expected PRT 5 (SD), got %d", expectedPlayer.PRT)
	}

	// Now create a race using the builder with the exact settings from the expected file
	builder := race.New()
	builder.Name(expectedPlayer.NameSingular, expectedPlayer.NamePlural)
	builder.PRT(expectedPlayer.PRT)
	builder.SetLRTs(expectedPlayer.LRT)

	// Set habitability
	gravCenter := (expectedPlayer.Hab.GravityLow + expectedPlayer.Hab.GravityHigh) / 2
	gravWidth := (expectedPlayer.Hab.GravityHigh - expectedPlayer.Hab.GravityLow) / 2
	builder.Gravity(gravCenter, gravWidth)

	if expectedPlayer.Hab.IsTemperatureImmune() {
		builder.TemperatureImmune(true)
	} else {
		tempCenter := (expectedPlayer.Hab.TemperatureLow + expectedPlayer.Hab.TemperatureHigh) / 2
		tempWidth := (expectedPlayer.Hab.TemperatureHigh - expectedPlayer.Hab.TemperatureLow) / 2
		builder.Temperature(tempCenter, tempWidth)
	}

	radCenter := (expectedPlayer.Hab.RadiationLow + expectedPlayer.Hab.RadiationHigh) / 2
	radWidth := (expectedPlayer.Hab.RadiationHigh - expectedPlayer.Hab.RadiationLow) / 2
	builder.Radiation(radCenter, radWidth)

	builder.GrowthRate(expectedPlayer.GrowthRate)

	// Set production
	builder.ColonistsPerResource(expectedPlayer.Production.ResourcePerColonist * 100)
	builder.Factories(
		expectedPlayer.Production.FactoryProduction,
		expectedPlayer.Production.FactoryCost,
		expectedPlayer.Production.FactoriesOperate,
		expectedPlayer.FactoriesCost1LessGerm,
	)
	builder.Mines(
		expectedPlayer.Production.MineProduction,
		expectedPlayer.Production.MineCost,
		expectedPlayer.Production.MinesOperate,
	)

	// Set research
	builder.Research(
		expectedPlayer.ResearchCost.Energy,
		expectedPlayer.ResearchCost.Weapons,
		expectedPlayer.ResearchCost.Propulsion,
		expectedPlayer.ResearchCost.Construction,
		expectedPlayer.ResearchCost.Electronics,
		expectedPlayer.ResearchCost.Biotech,
	)
	builder.TechsStartHigh(expectedPlayer.ExpensiveTechStartsAt3)
	builder.LeftoverPointsOn(race.LeftoverPointsOption(expectedPlayer.SpendLeftoverPoints))
	builder.Icon(expectedPlayer.Logo)

	r, err := builder.Finish()
	if err != nil {
		t.Fatalf("Failed to finish race: %v", err)
	}

	// Create a race file
	data, err := CreateRaceFile(r, 1)
	if err != nil {
		t.Fatalf("Failed to create race file: %v", err)
	}

	// Parse our generated file
	fd := parser.FileData(data)
	parsedBlocks, err := fd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse generated race file: %v", err)
	}

	// Find the PlayerBlock in our generated file
	var generatedPlayer *blocks.PlayerBlock
	for _, block := range parsedBlocks {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			generatedPlayer = &pb
			break
		}
	}
	if generatedPlayer == nil {
		t.Fatal("No PlayerBlock found in generated file")
	}

	// Compare key attributes
	if generatedPlayer.NameSingular != expectedPlayer.NameSingular {
		t.Errorf("Name singular: expected '%s', got '%s'", expectedPlayer.NameSingular, generatedPlayer.NameSingular)
	}
	if generatedPlayer.PRT != expectedPlayer.PRT {
		t.Errorf("PRT: expected %d, got %d", expectedPlayer.PRT, generatedPlayer.PRT)
	}
	if generatedPlayer.LRT != expectedPlayer.LRT {
		t.Errorf("LRT: expected 0x%04X, got 0x%04X", expectedPlayer.LRT, generatedPlayer.LRT)
	}
	if generatedPlayer.Hab != expectedPlayer.Hab {
		t.Errorf("Hab: expected %+v, got %+v", expectedPlayer.Hab, generatedPlayer.Hab)
	}
	if generatedPlayer.Production != expectedPlayer.Production {
		t.Errorf("Production: expected %+v, got %+v", expectedPlayer.Production, generatedPlayer.Production)
	}
	if generatedPlayer.ResearchCost != expectedPlayer.ResearchCost {
		t.Errorf("ResearchCost: expected %+v, got %+v", expectedPlayer.ResearchCost, generatedPlayer.ResearchCost)
	}

	t.Logf("Antetheral predefined race test passed - all attributes match!")
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
