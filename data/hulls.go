// Package data contains static game data and constants for Stars! file parsing.
package data

// Hull ID constants for ship hulls.
// These values correspond to the HullId field in DesignBlock.
const (
	HullSmallFreighter  = 0
	HullMediumFreighter = 1
	HullLargeFreighter  = 2
	HullSuperFreighter  = 3
	HullScout           = 4
	HullFrigate         = 5
	HullDestroyer       = 6
	HullCruiser         = 7
	HullBattleCruiser   = 8
	HullBattleship      = 9
	HullDreadnought     = 10
	HullPrivateer       = 11
	HullRogue           = 12
	HullGalleon         = 13
	HullMiniColonyShip  = 14
	HullColonyShip      = 15
	HullMiniBomber      = 16
	HullB17Bomber       = 17
	HullStealthBomber   = 18
	HullB52Bomber       = 19
	HullMidgetMiner     = 20
	HullMiniMiner       = 21
	HullMiner           = 22
	HullMaxiMiner       = 23
	HullUltraMiner      = 24
	HullFuelTransport   = 25
	HullSuperFuelXport  = 26
	HullMiniMineLayer   = 27
	HullSuperMineLayer  = 28
	HullNubian          = 29
	HullMiniMorph       = 30
	HullMetaMorph       = 31
	// Starbase hulls
	HullOrbitalFort  = 32
	HullSpaceDock    = 33
	HullSpaceStation = 34
	HullUltraStation = 35
	HullDeathStar    = 36
)

// HullNames maps hull IDs to their display names.
var HullNames = map[int]string{
	HullSmallFreighter:  "Small Freighter",
	HullMediumFreighter: "Medium Freighter",
	HullLargeFreighter:  "Large Freighter",
	HullSuperFreighter:  "Super Freighter",
	HullScout:           "Scout",
	HullFrigate:         "Frigate",
	HullDestroyer:       "Destroyer",
	HullCruiser:         "Cruiser",
	HullBattleCruiser:   "Battle Cruiser",
	HullBattleship:      "Battleship",
	HullDreadnought:     "Dreadnought",
	HullPrivateer:       "Privateer",
	HullRogue:           "Rogue",
	HullGalleon:         "Galleon",
	HullMiniColonyShip:  "Mini-Colony Ship",
	HullColonyShip:      "Colony Ship",
	HullMiniBomber:      "Mini Bomber",
	HullB17Bomber:       "B-17 Bomber",
	HullStealthBomber:   "Stealth Bomber",
	HullB52Bomber:       "B-52 Bomber",
	HullMidgetMiner:     "Midget Miner",
	HullMiniMiner:       "Mini-Miner",
	HullMiner:           "Miner",
	HullMaxiMiner:       "Maxi-Miner",
	HullUltraMiner:      "Ultra-Miner",
	HullFuelTransport:   "Fuel Transport",
	HullSuperFuelXport:  "Super-Fuel Xport",
	HullMiniMineLayer:   "Mini Mine Layer",
	HullSuperMineLayer:  "Super Mine Layer",
	HullNubian:          "Nubian",
	HullMiniMorph:       "Mini Morph",
	HullMetaMorph:       "Meta Morph",
	HullOrbitalFort:     "Orbital Fort",
	HullSpaceDock:       "Space Dock",
	HullSpaceStation:    "Space Station",
	HullUltraStation:    "Ultra Station",
	HullDeathStar:       "Death Star",
}

// HullNameToID maps hull names to their IDs.
// Names are stored with their canonical spelling from the game.
var HullNameToID = map[string]int{
	"Small Freighter":  HullSmallFreighter,
	"Medium Freighter": HullMediumFreighter,
	"Large Freighter":  HullLargeFreighter,
	"Super Freighter":  HullSuperFreighter,
	"Scout":            HullScout,
	"Frigate":          HullFrigate,
	"Destroyer":        HullDestroyer,
	"Cruiser":          HullCruiser,
	"Battle Cruiser":   HullBattleCruiser,
	"Battleship":       HullBattleship,
	"Dreadnought":      HullDreadnought,
	"Privateer":        HullPrivateer,
	"Rogue":            HullRogue,
	"Galleon":          HullGalleon,
	"Mini-Colony Ship": HullMiniColonyShip,
	"Colony Ship":      HullColonyShip,
	"Mini Bomber":      HullMiniBomber,
	"B-17 Bomber":      HullB17Bomber,
	"Stealth Bomber":   HullStealthBomber,
	"B-52 Bomber":      HullB52Bomber,
	"Midget Miner":     HullMidgetMiner,
	"Mini-Miner":       HullMiniMiner,
	"Miner":            HullMiner,
	"Maxi-Miner":       HullMaxiMiner,
	"Ultra-Miner":      HullUltraMiner,
	"Fuel Transport":   HullFuelTransport,
	"Super-Fuel Xport": HullSuperFuelXport,
	"Mini Mine Layer":  HullMiniMineLayer,
	"Super Mine Layer": HullSuperMineLayer,
	"Nubian":           HullNubian,
	"Mini Morph":       HullMiniMorph,
	"Meta Morph":       HullMetaMorph,
	"Orbital Fort":     HullOrbitalFort,
	"Space Dock":       HullSpaceDock,
	"Space Station":    HullSpaceStation,
	"Ultra Station":    HullUltraStation,
	"Death Star":       HullDeathStar,
}

// IsStarbaseHull returns true if the hull ID is a starbase hull.
func IsStarbaseHull(hullID int) bool {
	return hullID >= HullOrbitalFort
}
