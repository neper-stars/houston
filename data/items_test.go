package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetItemInfo_Engines(t *testing.T) {
	testCases := []struct {
		name         string
		expectedCat  ItemCategory
		expectedID   int
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
		name         string
		expectedCat  ItemCategory
		expectedID   int
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
		name         string
		expectedCat  ItemCategory
		expectedID   int
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
		name         string
		expectedCat  ItemCategory
		expectedID   int
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
		name         string
		expectedCat  ItemCategory
		expectedID   int
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
