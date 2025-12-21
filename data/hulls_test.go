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
