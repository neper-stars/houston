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

// Slot category bitmasks (what item types a slot accepts)
// These can be combined (OR'd) for multi-type slots
const (
	SlotEngine     uint16 = 0x0001
	SlotScanner    uint16 = 0x0002
	SlotShield     uint16 = 0x0004
	SlotArmor      uint16 = 0x0008
	SlotBeamWeapon uint16 = 0x0010
	SlotTorpedo    uint16 = 0x0020
	SlotBomb       uint16 = 0x0040
	SlotMining     uint16 = 0x0080
	SlotMineLayer  uint16 = 0x0100
	SlotOrbital    uint16 = 0x0200
	SlotPlanetary  uint16 = 0x0400
	SlotElectrical uint16 = 0x0800
	SlotMechanical uint16 = 0x1000

	// Common slot combinations
	SlotWeapon          = SlotBeamWeapon | SlotTorpedo                  // 0x0030
	SlotShieldArmor     = SlotShield | SlotArmor                        // 0x000C
	SlotScannerElecMech = SlotScanner | SlotElectrical | SlotMechanical // 0x1802

	// GeneralPurpose accepts most item types (0x193E)
	SlotGeneralPurpose = SlotScanner | SlotShield | SlotArmor | SlotBeamWeapon |
		SlotTorpedo | SlotMineLayer | SlotElectrical | SlotMechanical
)

// HullSlot defines a slot on a hull
type HullSlot struct {
	Category uint16 // Bitmask of accepted item types
	MaxItems int    // Maximum number of items in this slot
}

// Hull defines a ship or starbase hull type
type Hull struct {
	ID            int
	Name          string
	Mass          int
	Armor         int
	FuelCapacity  int
	CargoCapacity int
	IsStarbase    bool
	Slots         []HullSlot
}

// Accepts returns true if this slot accepts the given item category
func (s HullSlot) Accepts(itemCategory uint16) bool {
	return (s.Category & itemCategory) != 0
}

// Hulls contains all hull definitions indexed by hull ID
var Hulls = map[int]*Hull{
	HullSmallFreighter: {
		ID: HullSmallFreighter, Name: "Small Freighter",
		Mass: 25, Armor: 20, FuelCapacity: 130, CargoCapacity: 70,
		Slots: []HullSlot{
			{SlotEngine, 1},
			{SlotShieldArmor, 1},
			{SlotScannerElecMech, 1},
		},
	},
	HullMediumFreighter: {
		ID: HullMediumFreighter, Name: "Medium Freighter",
		Mass: 60, Armor: 50, FuelCapacity: 450, CargoCapacity: 210,
		Slots: []HullSlot{
			{SlotEngine, 1},
			{SlotShieldArmor, 1},
			{SlotScannerElecMech, 1},
		},
	},
	HullLargeFreighter: {
		ID: HullLargeFreighter, Name: "Large Freighter",
		Mass: 125, Armor: 150, FuelCapacity: 2600, CargoCapacity: 1200,
		Slots: []HullSlot{
			{SlotEngine, 2},
			{SlotShieldArmor, 2},
			{SlotGeneralPurpose, 2},
		},
	},
	HullSuperFreighter: {
		ID: HullSuperFreighter, Name: "Super Freighter",
		Mass: 175, Armor: 400, FuelCapacity: 8000, CargoCapacity: 3000,
		Slots: []HullSlot{
			{SlotEngine, 3},
			{SlotShieldArmor, 5},
			{SlotElectrical, 2},
		},
	},
	HullScout: {
		ID: HullScout, Name: "Scout",
		Mass: 8, Armor: 20, FuelCapacity: 50, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 1},
			{SlotScanner, 1},
			{SlotGeneralPurpose, 1},
		},
	},
	HullFrigate: {
		ID: HullFrigate, Name: "Frigate",
		Mass: 8, Armor: 45, FuelCapacity: 125, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 1},
			{SlotScanner, 2},
			{SlotGeneralPurpose, 3},
			{SlotShieldArmor, 2},
		},
	},
	HullDestroyer: {
		ID: HullDestroyer, Name: "Destroyer",
		Mass: 30, Armor: 200, FuelCapacity: 280, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 1},
			{SlotWeapon, 1},
			{SlotWeapon, 1},
			{SlotGeneralPurpose, 1},
			{SlotArmor, 2},
			{SlotMechanical, 1},
			{SlotElectrical, 1},
		},
	},
	HullCruiser: {
		ID: HullCruiser, Name: "Cruiser",
		Mass: 90, Armor: 700, FuelCapacity: 600, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 2},
			{SlotShieldArmor, 1},
			{SlotShieldArmor, 1},
			{SlotWeapon, 2},
			{SlotWeapon, 2},
			{SlotGeneralPurpose, 2},
			{SlotShieldArmor, 2},
		},
	},
	HullBattleCruiser: {
		ID: HullBattleCruiser, Name: "Battle Cruiser",
		Mass: 120, Armor: 1000, FuelCapacity: 1400, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 2},
			{SlotShieldArmor, 2},
			{SlotShieldArmor, 2},
			{SlotWeapon, 3},
			{SlotWeapon, 3},
			{SlotGeneralPurpose, 3},
			{SlotShieldArmor, 4},
		},
	},
	HullBattleship: {
		ID: HullBattleship, Name: "Battleship",
		Mass: 222, Armor: 2000, FuelCapacity: 2800, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 4},
			{SlotShieldArmor, 1},
			{SlotShield, 8},
			{SlotWeapon, 6},
			{SlotWeapon, 6},
			{SlotWeapon, 2},
			{SlotWeapon, 2},
			{SlotWeapon, 4},
			{SlotArmor, 6},
			{SlotElectrical, 3},
			{SlotElectrical, 3},
		},
	},
	HullDreadnought: {
		ID: HullDreadnought, Name: "Dreadnought",
		Mass: 250, Armor: 4500, FuelCapacity: 4500, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 5},
			{SlotShieldArmor, 4},
			{SlotShieldArmor, 4},
			{SlotWeapon, 6},
			{SlotWeapon, 6},
			{SlotElectrical, 4},
			{SlotElectrical, 4},
			{SlotWeapon, 8},
			{SlotWeapon, 8},
			{SlotArmor, 8},
			{SlotWeapon | SlotShield, 5},
			{SlotWeapon | SlotShield, 5},
			{SlotGeneralPurpose, 2},
		},
	},
	HullPrivateer: {
		ID: HullPrivateer, Name: "Privateer",
		Mass: 65, Armor: 150, FuelCapacity: 650, CargoCapacity: 250,
		Slots: []HullSlot{
			{SlotEngine, 1},
			{SlotShieldArmor, 2},
			{SlotShieldArmor, 1},
			{SlotGeneralPurpose, 1},
			{SlotGeneralPurpose, 1},
		},
	},
	HullRogue: {
		ID: HullRogue, Name: "Rogue",
		Mass: 75, Armor: 450, FuelCapacity: 2250, CargoCapacity: 500,
		Slots: []HullSlot{
			{SlotEngine, 2},
			{SlotShieldArmor, 3},
			{SlotScanner | SlotElectrical, 2},
			{SlotScanner, 1},
			{SlotGeneralPurpose, 2},
			{SlotGeneralPurpose, 2},
			{SlotScanner | SlotElectrical, 2},
			{SlotElectrical, 1},
			{SlotElectrical, 1},
		},
	},
	HullGalleon: {
		ID: HullGalleon, Name: "Galleon",
		Mass: 125, Armor: 900, FuelCapacity: 2500, CargoCapacity: 1000,
		Slots: []HullSlot{
			{SlotEngine, 4},
			{SlotShieldArmor, 2},
			{SlotShieldArmor, 2},
			{SlotGeneralPurpose, 3},
			{SlotGeneralPurpose, 3},
			{SlotScanner | SlotElectrical, 2},
			{SlotScanner, 2},
			{SlotScanner, 2},
		},
	},
	HullMiniColonyShip: {
		ID: HullMiniColonyShip, Name: "Mini-Colony Ship",
		Mass: 8, Armor: 10, FuelCapacity: 150, CargoCapacity: 10,
		Slots: []HullSlot{
			{SlotEngine, 1},
			{SlotMechanical, 1},
		},
	},
	HullColonyShip: {
		ID: HullColonyShip, Name: "Colony Ship",
		Mass: 20, Armor: 20, FuelCapacity: 200, CargoCapacity: 25,
		Slots: []HullSlot{
			{SlotEngine, 1},
			{SlotMechanical, 1},
		},
	},
	HullMiniBomber: {
		ID: HullMiniBomber, Name: "Mini Bomber",
		Mass: 28, Armor: 50, FuelCapacity: 120, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 1},
			{SlotBomb, 2},
		},
	},
	HullB17Bomber: {
		ID: HullB17Bomber, Name: "B-17 Bomber",
		Mass: 69, Armor: 175, FuelCapacity: 400, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 2},
			{SlotBomb, 4},
			{SlotBomb, 4},
			{SlotShieldArmor, 1},
		},
	},
	HullStealthBomber: {
		ID: HullStealthBomber, Name: "Stealth Bomber",
		Mass: 70, Armor: 225, FuelCapacity: 750, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 2},
			{SlotBomb, 4},
			{SlotBomb, 4},
			{SlotShieldArmor, 1},
			{SlotElectrical, 3},
		},
	},
	HullB52Bomber: {
		ID: HullB52Bomber, Name: "B-52 Bomber",
		Mass: 110, Armor: 450, FuelCapacity: 750, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 3},
			{SlotBomb, 4},
			{SlotBomb, 4},
			{SlotBomb, 4},
			{SlotBomb, 4},
			{SlotShieldArmor, 2},
			{SlotShield, 2},
		},
	},
	HullMidgetMiner: {
		ID: HullMidgetMiner, Name: "Midget Miner",
		Mass: 10, Armor: 100, FuelCapacity: 210, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 1},
			{SlotMining, 2},
		},
	},
	HullMiniMiner: {
		ID: HullMiniMiner, Name: "Mini-Miner",
		Mass: 80, Armor: 130, FuelCapacity: 210, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 1},
			{SlotShieldArmor, 1},
			{SlotMining, 1},
			{SlotMining, 1},
		},
	},
	HullMiner: {
		ID: HullMiner, Name: "Miner",
		Mass: 110, Armor: 475, FuelCapacity: 500, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 2},
			{SlotShieldArmor | SlotScanner, 2},
			{SlotMining, 2},
			{SlotMining, 1},
			{SlotMining, 2},
			{SlotMining, 1},
		},
	},
	HullMaxiMiner: {
		ID: HullMaxiMiner, Name: "Maxi-Miner",
		Mass: 110, Armor: 1400, FuelCapacity: 850, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 3},
			{SlotShieldArmor | SlotScanner, 2},
			{SlotMining, 4},
			{SlotMining, 1},
			{SlotMining, 4},
			{SlotMining, 1},
		},
	},
	HullUltraMiner: {
		ID: HullUltraMiner, Name: "Ultra-Miner",
		Mass: 100, Armor: 1500, FuelCapacity: 1300, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 2},
			{SlotShieldArmor | SlotScanner, 3},
			{SlotMining, 4},
			{SlotMining, 2},
			{SlotMining, 4},
			{SlotMining, 2},
		},
	},
	HullFuelTransport: {
		ID: HullFuelTransport, Name: "Fuel Transport",
		Mass: 12, Armor: 5, FuelCapacity: 750, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 1},
			{SlotShield, 1},
		},
	},
	HullSuperFuelXport: {
		ID: HullSuperFuelXport, Name: "Super-Fuel Xport",
		Mass: 111, Armor: 12, FuelCapacity: 2250, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 2},
			{SlotShield, 2},
			{SlotScanner, 1},
		},
	},
	HullMiniMineLayer: {
		ID: HullMiniMineLayer, Name: "Mini Mine Layer",
		Mass: 10, Armor: 60, FuelCapacity: 400, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 1},
			{SlotMineLayer, 2},
			{SlotMineLayer, 2},
			{SlotShieldArmor, 1},
		},
	},
	HullSuperMineLayer: {
		ID: HullSuperMineLayer, Name: "Super Mine Layer",
		Mass: 30, Armor: 1200, FuelCapacity: 2200, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 3},
			{SlotMineLayer, 8},
			{SlotMineLayer, 8},
			{SlotShieldArmor, 3},
			{SlotShieldArmor, 3},
			{SlotScanner | SlotElectrical, 3},
		},
	},
	HullNubian: {
		ID: HullNubian, Name: "Nubian",
		Mass: 100, Armor: 5000, FuelCapacity: 5000, CargoCapacity: 0,
		Slots: []HullSlot{
			{SlotEngine, 3},
			{SlotGeneralPurpose, 3},
			{SlotGeneralPurpose, 3},
			{SlotGeneralPurpose, 3},
			{SlotGeneralPurpose, 3},
			{SlotGeneralPurpose, 3},
			{SlotGeneralPurpose, 3},
			{SlotGeneralPurpose, 3},
			{SlotGeneralPurpose, 3},
			{SlotGeneralPurpose, 3},
			{SlotGeneralPurpose, 3},
			{SlotGeneralPurpose, 3},
			{SlotGeneralPurpose, 3},
		},
	},
	HullMiniMorph: {
		ID: HullMiniMorph, Name: "Mini Morph",
		Mass: 70, Armor: 250, FuelCapacity: 400, CargoCapacity: 150,
		Slots: []HullSlot{
			{SlotEngine, 2},
			{SlotGeneralPurpose, 3},
			{SlotGeneralPurpose, 1},
			{SlotGeneralPurpose, 1},
			{SlotGeneralPurpose, 1},
			{SlotGeneralPurpose, 2},
			{SlotGeneralPurpose, 2},
		},
	},
	HullMetaMorph: {
		ID: HullMetaMorph, Name: "Meta Morph",
		Mass: 85, Armor: 500, FuelCapacity: 700, CargoCapacity: 300,
		Slots: []HullSlot{
			{SlotEngine, 3},
			{SlotGeneralPurpose, 8},
			{SlotGeneralPurpose, 2},
			{SlotGeneralPurpose, 2},
			{SlotGeneralPurpose, 1},
			{SlotGeneralPurpose, 2},
			{SlotGeneralPurpose, 2},
		},
	},

	// Starbases
	HullOrbitalFort: {
		ID: HullOrbitalFort, Name: "Orbital Fort",
		Mass: 0, Armor: 100, FuelCapacity: 0, CargoCapacity: 0,
		IsStarbase: true,
		Slots: []HullSlot{
			{SlotOrbital, 1},
			{SlotWeapon, 12},
			{SlotShieldArmor, 12},
			{SlotWeapon, 12},
			{SlotShieldArmor, 12},
		},
	},
	HullSpaceDock: {
		ID: HullSpaceDock, Name: "Space Dock",
		Mass: 0, Armor: 250, FuelCapacity: 0, CargoCapacity: 200,
		IsStarbase: true,
		Slots: []HullSlot{
			{SlotOrbital, 1},
			{SlotWeapon, 16},
			{SlotShieldArmor, 24},
			{SlotWeapon, 16},
			{SlotShield, 24},
			{SlotElectrical, 2},
			{SlotElectrical, 2},
			{SlotWeapon, 16},
		},
	},
	HullSpaceStation: {
		ID: HullSpaceStation, Name: "Space Station",
		Mass: 0, Armor: 500, FuelCapacity: 0, CargoCapacity: 65535,
		IsStarbase: true,
		Slots: []HullSlot{
			{SlotOrbital, 1},
			{SlotWeapon, 16},
			{SlotShield, 16},
			{SlotWeapon, 16},
			{SlotShieldArmor, 16},
			{SlotShield, 16},
			{SlotElectrical, 3},
			{SlotWeapon, 16},
			{SlotElectrical, 3},
			{SlotWeapon, 16},
			{SlotOrbital, 1},
			{SlotShieldArmor, 16},
		},
	},
	HullUltraStation: {
		ID: HullUltraStation, Name: "Ultra Station",
		Mass: 0, Armor: 1000, FuelCapacity: 0, CargoCapacity: 65535,
		IsStarbase: true,
		Slots: []HullSlot{
			{SlotOrbital, 1},
			{SlotWeapon, 16},
			{SlotElectrical, 3},
			{SlotWeapon, 16},
			{SlotShield, 20},
			{SlotShield, 20},
			{SlotElectrical, 3},
			{SlotWeapon, 16},
			{SlotElectrical, 3},
			{SlotWeapon, 16},
			{SlotOrbital, 1},
			{SlotShieldArmor, 20},
			{SlotWeapon, 16},
			{SlotShieldArmor, 20},
			{SlotElectrical, 3},
			{SlotWeapon, 16},
		},
	},
	HullDeathStar: {
		ID: HullDeathStar, Name: "Death Star",
		Mass: 0, Armor: 1500, FuelCapacity: 0, CargoCapacity: 65535,
		IsStarbase: true,
		Slots: []HullSlot{
			{SlotOrbital, 1},
			{SlotWeapon, 32},
			{SlotElectrical, 4},
			{SlotElectrical, 4},
			{SlotShield, 30},
			{SlotShield, 30},
			{SlotElectrical, 4},
			{SlotWeapon, 32},
			{SlotElectrical, 4},
			{SlotWeapon, 32},
			{SlotOrbital, 1},
			{SlotShieldArmor, 20},
			{SlotElectrical, 4},
			{SlotShieldArmor, 20},
			{SlotElectrical, 4},
			{SlotWeapon, 32},
		},
	},
}

// GetHull returns the hull definition for the given hull ID, or nil if not found
func GetHull(hullID int) *Hull {
	return Hulls[hullID]
}
