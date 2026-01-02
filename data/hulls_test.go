package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHullNameToID(t *testing.T) {
	testCases := []struct {
		name     string
		expected int
	}{
		{"Scout", HullScout},
		{"Frigate", HullFrigate},
		{"Destroyer", HullDestroyer},
		{"Cruiser", HullCruiser},
		{"Colony Ship", HullColonyShip},
		{"Medium Freighter", HullMediumFreighter},
		{"Mini-Miner", HullMiniMiner},
		{"Battleship", HullBattleship},
		{"Dreadnought", HullDreadnought},
		{"Orbital Fort", HullOrbitalFort},
		{"Space Station", HullSpaceStation},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id, found := HullNameToID[tc.name]
			assert.True(t, found, "Hull '%s' should exist in HullNameToID", tc.name)
			assert.Equal(t, tc.expected, id, "Hull '%s' should have ID %d", tc.name, tc.expected)
		})
	}
}

func TestHullNames_ReverseMapping(t *testing.T) {
	// Verify that HullNames and HullNameToID are consistent
	for name, id := range HullNameToID {
		reverseName, found := HullNames[id]
		assert.True(t, found, "Hull ID %d should exist in HullNames", id)
		assert.Equal(t, name, reverseName, "Reverse mapping for ID %d should match", id)
	}
}

func TestIsStarbaseHull(t *testing.T) {
	assert.False(t, IsStarbaseHull(HullScout), "Scout should not be a starbase")
	assert.False(t, IsStarbaseHull(HullDestroyer), "Destroyer should not be a starbase")
	assert.False(t, IsStarbaseHull(HullMetaMorph), "Meta Morph should not be a starbase")
	assert.True(t, IsStarbaseHull(HullOrbitalFort), "Orbital Fort should be a starbase")
	assert.True(t, IsStarbaseHull(HullSpaceStation), "Space Station should be a starbase")
	assert.True(t, IsStarbaseHull(HullDeathStar), "Death Star should be a starbase")
}

func TestGetHull(t *testing.T) {
	// Test valid hull IDs
	testCases := []struct {
		id   int
		name string
	}{
		{HullScout, "Scout"},
		{HullFrigate, "Frigate"},
		{HullDestroyer, "Destroyer"},
		{HullSmallFreighter, "Small Freighter"},
		{HullColonyShip, "Colony Ship"},
		{HullBattleship, "Battleship"},
		{HullOrbitalFort, "Orbital Fort"},
		{HullSpaceStation, "Space Station"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hull := GetHull(tc.id)
			assert.NotNil(t, hull, "GetHull(%d) should not return nil", tc.id)
			assert.Equal(t, tc.name, hull.Name)
			assert.Equal(t, tc.id, hull.ID)
		})
	}

	// Test invalid hull ID
	assert.Nil(t, GetHull(9999), "GetHull(9999) should return nil")
	assert.Nil(t, GetHull(-1), "GetHull(-1) should return nil")
}

func TestHullSlotAccepts(t *testing.T) {
	// Engine slot should accept engines
	engineSlot := HullSlot{Category: SlotEngine, MaxItems: 1}
	assert.True(t, engineSlot.Accepts(SlotEngine), "Engine slot should accept engines")
	assert.False(t, engineSlot.Accepts(SlotWeapon), "Engine slot should not accept weapons")
	assert.False(t, engineSlot.Accepts(SlotShieldArmor), "Engine slot should not accept shields")

	// General purpose slot: Weapons, Devices, Armor, Mine Layers, Scanners, Shields
	// GP does NOT accept engines
	gpSlot := HullSlot{Category: SlotGeneralPurpose, MaxItems: 1}
	assert.False(t, gpSlot.Accepts(SlotEngine), "GP slot should NOT accept engines")
	assert.True(t, gpSlot.Accepts(SlotWeapon), "GP slot should accept weapons")
	assert.True(t, gpSlot.Accepts(SlotShieldArmor), "GP slot should accept shields/armor")

	// Weapon slot
	weaponSlot := HullSlot{Category: SlotWeapon, MaxItems: 1}
	assert.True(t, weaponSlot.Accepts(SlotWeapon), "Weapon slot should accept weapons")
	assert.False(t, weaponSlot.Accepts(SlotEngine), "Weapon slot should not accept engines")
}

func TestHullHasSlots(t *testing.T) {
	// Scout should have engine slot
	scout := GetHull(HullScout)
	assert.NotNil(t, scout)
	assert.Greater(t, len(scout.Slots), 0, "Scout should have slots")

	// Verify Scout has required engine slot
	hasEngine := false
	for _, slot := range scout.Slots {
		if slot.Accepts(SlotEngine) {
			hasEngine = true
			break
		}
	}
	assert.True(t, hasEngine, "Scout should have an engine slot")
}

func TestHullStarbaseFlag(t *testing.T) {
	// Ships should not be starbases
	scout := GetHull(HullScout)
	assert.NotNil(t, scout)
	assert.False(t, scout.IsStarbase, "Scout should not be a starbase")

	// Starbases should be starbases
	station := GetHull(HullSpaceStation)
	assert.NotNil(t, station)
	assert.True(t, station.IsStarbase, "Space Station should be a starbase")
}
