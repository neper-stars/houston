package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetItemInfo_Engines(t *testing.T) {
	testCases := []struct {
		name        string
		expectedCat ItemCategory
		expectedID  int
	}{
		{"Settler's Delight", CategoryEngine, EngineSettlersDelight},
		{"Long Hump 6", CategoryEngine, EngineLongHump6},
		{"Trans-Galactic Drive", CategoryEngine, EngineTransGalacticDrive},
		{"Fuel Mizer", CategoryEngine, EngineFuelMizer},
		{"Galaxy Scoop", CategoryEngine, EngineGalaxyScoop},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info, found := GetItemInfo(tc.name)
			assert.True(t, found, "Item '%s' should exist", tc.name)
			assert.Equal(t, tc.expectedCat, info.Category, "Category mismatch")
			assert.Equal(t, tc.expectedID, info.ItemID, "ItemID mismatch")
		})
	}
}

func TestGetItemInfo_Weapons(t *testing.T) {
	testCases := []struct {
		name        string
		expectedCat ItemCategory
		expectedID  int
	}{
		{"Laser", CategoryBeamWeapon, BeamLaser},
		{"X-Ray Laser", CategoryBeamWeapon, BeamXRayLaser},
		{"Mega Disruptor", CategoryBeamWeapon, BeamMegaDisruptor},
		{"Alpha Torpedo", CategoryTorpedo, TorpedoAlpha},
		{"Omega Torpedo", CategoryTorpedo, TorpedoOmega},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info, found := GetItemInfo(tc.name)
			assert.True(t, found, "Item '%s' should exist", tc.name)
			assert.Equal(t, tc.expectedCat, info.Category, "Category mismatch")
			assert.Equal(t, tc.expectedID, info.ItemID, "ItemID mismatch")
		})
	}
}

func TestGetItemInfo_Defenses(t *testing.T) {
	testCases := []struct {
		name        string
		expectedCat ItemCategory
		expectedID  int
	}{
		{"Mole-skin Shield", CategoryShield, ShieldMoleskin},
		{"Gorilla Delagator", CategoryShield, ShieldGorillaDelagator},
		{"Tritanium", CategoryArmor, ArmorTritanium},
		{"Neutronium", CategoryArmor, ArmorNeutronium},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info, found := GetItemInfo(tc.name)
			assert.True(t, found, "Item '%s' should exist", tc.name)
			assert.Equal(t, tc.expectedCat, info.Category, "Category mismatch")
			assert.Equal(t, tc.expectedID, info.ItemID, "ItemID mismatch")
		})
	}
}

func TestGetItemInfo_Electronics(t *testing.T) {
	testCases := []struct {
		name        string
		expectedCat ItemCategory
		expectedID  int
	}{
		{"Rhino Scanner", CategoryScanner, ScannerRhino},
		{"Ferret Scanner", CategoryScanner, ScannerFerret},
		{"Peerless Scanner", CategoryScanner, ScannerPeerless},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info, found := GetItemInfo(tc.name)
			assert.True(t, found, "Item '%s' should exist", tc.name)
			assert.Equal(t, tc.expectedCat, info.Category, "Category mismatch")
			assert.Equal(t, tc.expectedID, info.ItemID, "ItemID mismatch")
		})
	}
}

func TestGetItemInfo_Mechanical(t *testing.T) {
	testCases := []struct {
		name        string
		expectedCat ItemCategory
		expectedID  int
	}{
		{"Colonization Module", CategoryMechanical, MechColonizationModule},
		{"Fuel Tank", CategoryMechanical, MechFuelTank},
		{"Cargo Pod", CategoryMechanical, MechCargoPod},
		{"Robo-Miner", CategoryMiningRobo, MiningRobo},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info, found := GetItemInfo(tc.name)
			assert.True(t, found, "Item '%s' should exist", tc.name)
			assert.Equal(t, tc.expectedCat, info.Category, "Category mismatch")
			assert.Equal(t, tc.expectedID, info.ItemID, "ItemID mismatch")
		})
	}
}

func TestGetItemName_RoundTrip(t *testing.T) {
	// Verify round-trip: name -> info -> name
	for name, info := range ItemNameToInfo {
		t.Run(name, func(t *testing.T) {
			reverseName := GetItemName(info.Category, info.ItemID)
			assert.Equal(t, name, reverseName,
				"Round-trip for '%s' should return same name", name)
		})
	}
}

func TestCategoryNames(t *testing.T) {
	assert.Equal(t, "Engine", CategoryNames[CategoryEngine])
	assert.Equal(t, "Beam Weapon", CategoryNames[CategoryBeamWeapon])
	assert.Equal(t, "Torpedo", CategoryNames[CategoryTorpedo])
	assert.Equal(t, "Shield", CategoryNames[CategoryShield])
	assert.Equal(t, "Scanner", CategoryNames[CategoryScanner])
	assert.Equal(t, "Armor", CategoryNames[CategoryArmor])
	assert.Equal(t, "Mechanical", CategoryNames[CategoryMechanical])
}

func TestItemNameToInfo_AllCategories(t *testing.T) {
	// Verify we have items in all expected categories
	categories := make(map[ItemCategory]int)
	for _, info := range ItemNameToInfo {
		categories[info.Category]++
	}

	assert.Greater(t, categories[CategoryEngine], 0, "Should have engines")
	assert.Greater(t, categories[CategoryBeamWeapon], 0, "Should have beam weapons")
	assert.Greater(t, categories[CategoryTorpedo], 0, "Should have torpedoes")
	assert.Greater(t, categories[CategoryBomb], 0, "Should have bombs")
	assert.Greater(t, categories[CategoryShield], 0, "Should have shields")
	assert.Greater(t, categories[CategoryScanner], 0, "Should have scanners")
	assert.Greater(t, categories[CategoryArmor], 0, "Should have armor")
	assert.Greater(t, categories[CategoryMechanical], 0, "Should have mechanical")
	assert.Greater(t, categories[CategoryElectrical], 0, "Should have electrical")
	assert.Greater(t, categories[CategoryMineLayer], 0, "Should have mine layers")
	assert.Greater(t, categories[CategoryMiningRobo], 0, "Should have mining robots")
	assert.Greater(t, categories[CategoryOrbital], 0, "Should have orbitals")
	assert.Greater(t, categories[CategoryTerraform], 0, "Should have terraforming")
	assert.Greater(t, categories[CategoryPlanetary], 0, "Should have planetary")
}

func TestGetScanner(t *testing.T) {
	// Test valid scanners
	testCases := []struct {
		id          int
		name        string
		normalRange int
	}{
		// Bat Scanner: no range - only detects fleets at same location,
		// and planet env/composition while orbiting
		{ScannerBat, "Bat Scanner", 0},
		{ScannerRhino, "Rhino Scanner", 50},
		{ScannerFerret, "Ferret Scanner", 185},
		{ScannerPeerless, "Peerless Scanner", 500},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scanner := GetScanner(tc.id)
			assert.NotNil(t, scanner, "GetScanner(%d) should not return nil", tc.id)
			assert.Equal(t, tc.name, scanner.Name)
			assert.Equal(t, tc.normalRange, scanner.NormalRange)
		})
	}

	// Test invalid ID
	assert.Nil(t, GetScanner(9999), "GetScanner(9999) should return nil")
}

func TestGetPlanetaryScanner(t *testing.T) {
	testCases := []struct {
		id          int
		name        string
		normalRange int
	}{
		{PlanetaryScannerViewer50, "Viewer 50", 50},
		{PlanetaryScannerScoper150, "Scoper 150", 150},
		{PlanetaryScannerSnooper320, "Snooper 320X", 320},
		{PlanetaryScannerSnooper620, "Snooper 620X", 620},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scanner := GetPlanetaryScanner(tc.id)
			assert.NotNil(t, scanner, "GetPlanetaryScanner(%d) should not return nil", tc.id)
			assert.Equal(t, tc.name, scanner.Name)
			assert.Equal(t, tc.normalRange, scanner.NormalRange)
		})
	}

	// Test invalid ID
	assert.Nil(t, GetPlanetaryScanner(9999), "GetPlanetaryScanner(9999) should return nil")
}

func TestGetBestPlanetaryScanner(t *testing.T) {
	// No tech - should get Viewer 50
	scanner, id := GetBestPlanetaryScanner(TechRequirements{})
	assert.Equal(t, PlanetaryScannerViewer50, id)
	assert.Equal(t, "Viewer 50", scanner.Name)

	// Electronics 3 - should get Scoper 150
	scanner, id = GetBestPlanetaryScanner(TechRequirements{Electronics: 3})
	assert.Equal(t, PlanetaryScannerScoper150, id)
	assert.Equal(t, "Scoper 150", scanner.Name)

	// High tech - should get Snooper 620X
	scanner, id = GetBestPlanetaryScanner(TechRequirements{
		Energy: 7, Electronics: 23, Biotech: 9,
	})
	assert.Equal(t, PlanetaryScannerSnooper620, id)
	assert.Equal(t, "Snooper 620X", scanner.Name)
}

func TestGetShipScannerStats(t *testing.T) {
	// Valid scanner
	stats, found := GetShipScannerStats(ScannerFerret)
	assert.True(t, found)
	assert.Equal(t, 185, stats.NormalRange)
	assert.Equal(t, 50, stats.PenetratingRange)

	// Invalid scanner
	stats, found = GetShipScannerStats(9999)
	assert.False(t, found)
	assert.Equal(t, 0, stats.NormalRange)
}

func TestGetPlanetaryScannerStats(t *testing.T) {
	// Valid scanner with pen range
	stats, found := GetPlanetaryScannerStats(PlanetaryScannerSnooper320)
	assert.True(t, found)
	assert.Equal(t, 320, stats.NormalRange)
	assert.Equal(t, 160, stats.PenetratingRange)

	// Valid scanner without pen range
	stats, found = GetPlanetaryScannerStats(PlanetaryScannerViewer50)
	assert.True(t, found)
	assert.Equal(t, 50, stats.NormalRange)
	assert.Equal(t, 0, stats.PenetratingRange)

	// Invalid scanner
	stats, found = GetPlanetaryScannerStats(9999)
	assert.False(t, found)
	assert.Equal(t, 0, stats.NormalRange)
}

func TestTechRequirementsCanBuildWith(t *testing.T) {
	// Test various tech requirement scenarios
	noReq := TechRequirements{}
	assert.True(t, noReq.CanBuildWith(TechRequirements{}), "No requirements should always pass")
	assert.True(t, noReq.CanBuildWith(TechRequirements{Energy: 5}), "No requirements should pass with any tech")

	// Energy requirement
	energyReq := TechRequirements{Energy: 3}
	assert.False(t, energyReq.CanBuildWith(TechRequirements{Energy: 2}), "Should fail with lower energy")
	assert.True(t, energyReq.CanBuildWith(TechRequirements{Energy: 3}), "Should pass with exact energy")
	assert.True(t, energyReq.CanBuildWith(TechRequirements{Energy: 5}), "Should pass with higher energy")

	// Multiple requirements
	multiReq := TechRequirements{Energy: 3, Electronics: 5, Biotech: 2}
	assert.False(t, multiReq.CanBuildWith(TechRequirements{Energy: 3, Electronics: 5}), "Should fail missing biotech")
	assert.False(t, multiReq.CanBuildWith(TechRequirements{Energy: 3, Electronics: 4, Biotech: 2}), "Should fail low electronics")
	assert.True(t, multiReq.CanBuildWith(TechRequirements{Energy: 3, Electronics: 5, Biotech: 2}), "Should pass exact")
	assert.True(t, multiReq.CanBuildWith(TechRequirements{Energy: 10, Electronics: 10, Biotech: 10}), "Should pass high tech")
}

func TestGetEngine(t *testing.T) {
	engine := GetEngine(EngineSettlersDelight)
	assert.NotNil(t, engine)
	assert.Equal(t, "Settler's Delight", engine.Name)

	engine = GetEngine(EngineTransGalacticDrive)
	assert.NotNil(t, engine)
	assert.Equal(t, "Trans-Galactic Drive", engine.Name)

	// Invalid
	assert.Nil(t, GetEngine(9999))
}

func TestGetShield(t *testing.T) {
	shield := GetShield(ShieldMoleskin)
	assert.NotNil(t, shield)
	assert.Equal(t, "Mole-skin Shield", shield.Name)

	// Invalid
	assert.Nil(t, GetShield(9999))
}

func TestGetArmor(t *testing.T) {
	armor := GetArmor(ArmorTritanium)
	assert.NotNil(t, armor)
	assert.Equal(t, "Tritanium", armor.Name)

	// Invalid
	assert.Nil(t, GetArmor(9999))
}

func TestGetBeamWeapon(t *testing.T) {
	beam := GetBeamWeapon(BeamLaser)
	assert.NotNil(t, beam)
	assert.Equal(t, "Laser", beam.Name)

	// Invalid
	assert.Nil(t, GetBeamWeapon(9999))
}

func TestGetTorpedo(t *testing.T) {
	torp := GetTorpedo(TorpedoAlpha)
	assert.NotNil(t, torp)
	assert.Equal(t, "Alpha Torpedo", torp.Name)

	// Invalid
	assert.Nil(t, GetTorpedo(9999))
}

func TestGetBomb(t *testing.T) {
	// Test that invalid ID returns nil
	assert.Nil(t, GetBomb(9999))

	// Test first bomb if map is populated
	bomb := GetBomb(1)
	if bomb != nil {
		assert.Equal(t, 1, bomb.ID)
	}
}

func TestGetMineLayer(t *testing.T) {
	// Test that invalid ID returns nil
	assert.Nil(t, GetMineLayer(9999))

	// Test first mine layer if map is populated
	miner := GetMineLayer(1)
	if miner != nil {
		assert.Equal(t, 1, miner.ID)
	}
}

func TestGetMiningRobot(t *testing.T) {
	// Test with defined constant
	robot := GetMiningRobot(MiningRoboMidget)
	assert.NotNil(t, robot)
	assert.Equal(t, "Robo-Midget Miner", robot.Name)

	robot = GetMiningRobot(MiningRobo)
	assert.NotNil(t, robot)
	assert.Equal(t, "Robo-Miner", robot.Name)

	// Invalid
	assert.Nil(t, GetMiningRobot(9999))
}

func TestGetElectrical(t *testing.T) {
	// Test first electrical if map is populated
	elec := GetElectrical(1)
	if elec != nil {
		assert.Equal(t, 1, elec.ID)
	}

	// Invalid
	assert.Nil(t, GetElectrical(9999))
}

func TestGetMechanical(t *testing.T) {
	// Test with defined constant
	mech := GetMechanical(MechColonizationModule)
	assert.NotNil(t, mech)
	assert.Equal(t, "Colonization Module", mech.Name)

	// Invalid
	assert.Nil(t, GetMechanical(9999))
}

func TestGetOrbital(t *testing.T) {
	// Test first orbital if map is populated
	orb := GetOrbital(1)
	if orb != nil {
		assert.Equal(t, 1, orb.ID)
	}

	// Invalid
	assert.Nil(t, GetOrbital(9999))
}

func TestGetTerraformer(t *testing.T) {
	// Test first terraformer if map is populated
	terra := GetTerraformer(1)
	if terra != nil {
		assert.Equal(t, 1, terra.ID)
	}

	// Invalid
	assert.Nil(t, GetTerraformer(9999))
}

func TestGetPlanetaryDefense(t *testing.T) {
	// Test with defined constant
	def := GetPlanetaryDefense(DefensePlanetaryShield)
	assert.NotNil(t, def)
	assert.Equal(t, "Planetary Shield", def.Name)

	// Invalid
	assert.Nil(t, GetPlanetaryDefense(9999))
}
