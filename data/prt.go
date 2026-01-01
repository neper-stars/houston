package data

import "math"

// PRT represents a Primary Racial Trait with all its specific abilities.
type PRT struct {
	// Basic identification
	Index int    // 0-9 index matching blocks.PRTxxx constants
	Code  string // Short code: "HE", "SS", etc.
	Name  string // Full name: "Hyper Expansion", etc.
	Desc  string // Description text

	// Race builder point cost (negative = advantage, positive = costs points)
	PointCost int

	// Population modifiers
	GrowthRateModifier    float64 // Multiplier for growth rate (HE: 2.0)
	MaxPopulationModifier float64 // Multiplier for max population (HE: 0.5, JOAT: 1.2)

	// Cost modifiers (1.0 = normal, 0.75 = 25% cheaper, 1.25 = 25% more expensive)
	WeaponsCostModifier  float64 // WM: 0.75, IS: 1.25
	DefensesCostModifier float64 // IS: 0.60
	StarbaseCostModifier float64 // AR: 0.80, IT: 0.75 (stargates only)

	// Cloaking
	IntrinsicCloakPercent float64 // SS: 0.75 (75% base cloaking for all ships)
	CargoAffectsCloak     bool    // true for most, false for SS

	// Scanning - Starbase intrinsic scanners (AR)
	HasIntrinsicScanner       bool                // AR: true (starbases have intrinsic scanner)
	IntrinsicScannerRangeFunc func(pop int64) int // AR: sqrt(pop/10)

	// Scanning - Fleet intrinsic scanners (JOAT)
	HasFleetIntrinsicScanner       bool                            // JOAT: true (some hulls have intrinsic scanners)
	FleetIntrinsicScannerHulls     []int                           // Hull IDs with intrinsic scanners (Scout, Frigate, Destroyer for JOAT)
	FleetIntrinsicScannerRangeFunc func(elecTech int) ScannerStats // JOAT: Electronics × 20 normal, × 10 pen

	// Scanning - Other
	MineFieldsActAsScanners bool // SD: true
	CanScanEnemyStargates   bool // IT: true
	PacketsHavePenScanner   bool // PP: true

	// Mine field interaction
	MineTravelBonus int // Warp speed bonus in mine fields (SS: +1, SD: +2)

	// Special restrictions
	CanBuildMineFields       bool // WM: false
	CanBuildAdvancedDefenses bool // WM: false
	CanBuildSmartBombs       bool // IS: false
	CanLiveOnPlanets         bool // AR: false (live on starbases)

	// Special abilities
	ColonistsReproduceDuringTransport bool    // IS: true
	FreeTerraforming                  bool    // CA: true
	TerraformingImprovementChance     float64 // CA: 0.10 (10% per year)
	PlanetsRevertWhenAbandoned        bool    // CA: true
	CanRemoteDetonateMines            bool    // SD: true
	MaxPacketWarp                     int     // PP: 13 (default is lower)
	StargateSafetyBonus               bool    // IT: true (safer stargate exceeding)

	// Starting tech bonuses (added to base starting tech)
	StartingTechEnergy       int
	StartingTechWeapons      int
	StartingTechPropulsion   int
	StartingTechConstruction int
	StartingTechElectronics  int
	StartingTechBiotech      int
}

// arIntrinsicScannerRange calculates the intrinsic scanner range for AR starbases.
// Formula: range = sqrt(population / 10)
func arIntrinsicScannerRange(population int64) int {
	if population <= 0 {
		return 0
	}
	return int(math.Sqrt(float64(population) / 10.0))
}

// joatFleetIntrinsicScanner calculates the intrinsic scanner range for JOAT fleet ships.
// Formula: Normal range = Electronics × 20, Penetrating range = Electronics × 10
// Minimum ranges are 60 ly normal and 30 ly penetrating (equivalent to Electronics 3).
func joatFleetIntrinsicScanner(electronicsLevel int) ScannerStats {
	normalRange := electronicsLevel * 20
	penRange := electronicsLevel * 10

	// Minimum ranges (JOAT starts with Electronics 3)
	if normalRange < 60 {
		normalRange = 60
	}
	if penRange < 30 {
		penRange = 30
	}

	return ScannerStats{
		NormalRange:      normalRange,
		PenetratingRange: penRange,
	}
}

// AllPRTs contains data for all 10 Primary Racial Traits.
var AllPRTs = []PRT{
	// 0: HE - Hyper Expansion
	{
		Index:                    0,
		Code:                     "HE",
		Name:                     "Hyper-Expansion",
		Desc:                     "You must expand to survive. You are given a small and cheap colony hull and an engine which travels at Warp 6 using no fuel. Your race will grow at twice the growth rate you select in step four; however, the maximum population for a given planet is cut in half. The completely flexible Meta Morph hull will be available only to your race.",
		PointCost:                -40,
		GrowthRateModifier:       2.0,
		MaxPopulationModifier:    0.5,
		WeaponsCostModifier:      1.0,
		DefensesCostModifier:     1.0,
		StarbaseCostModifier:     1.0,
		IntrinsicCloakPercent:    0.0,
		CargoAffectsCloak:        true,
		MineTravelBonus:          0,
		CanBuildMineFields:       true,
		CanBuildAdvancedDefenses: true,
		CanBuildSmartBombs:       true,
		CanLiveOnPlanets:         true,
	},

	// 1: SS - Super Stealth
	{
		Index:                    1,
		Code:                     "SS",
		Name:                     "Super Stealth",
		Desc:                     "You can sneak through enemy territory and execute stunning surprise attacks. You are given top-drawer cloaks and all your ships have 75% cloaking built in. Cargo does not decrease your cloaking abilities. The Stealth Bomber and Rogue are at your disposal, as are a scanner, shield and armor with stealthy properties. Two scanners which allow you to steal minerals from enemy fleets and planets are also available. You may safely travel through mine fields at one warp speed faster than the limits.",
		PointCost:                -95,
		GrowthRateModifier:       1.0,
		MaxPopulationModifier:    1.0,
		WeaponsCostModifier:      1.0,
		DefensesCostModifier:     1.0,
		StarbaseCostModifier:     1.0,
		IntrinsicCloakPercent:    0.75, // 75% base cloaking for ALL ships
		CargoAffectsCloak:        false,
		MineTravelBonus:          1, // +1 warp speed in mine fields
		CanBuildMineFields:       true,
		CanBuildAdvancedDefenses: true,
		CanBuildSmartBombs:       true,
		CanLiveOnPlanets:         true,
	},

	// 2: WM - War Monger
	{
		Index:                    2,
		Code:                     "WM",
		Name:                     "War Monger",
		Desc:                     "You rule the battle field. Your colonists attack better, your ships are faster in battle, and you build weapons 25% cheaper than other races. You start the game with a knowledge of Tech 6 weapons and Tech 1 in energy and propulsion. Unfortunately, your race doesn't understand the necessity of building any but the most basic planetary defenses and no mine fields.",
		PointCost:                -45,
		GrowthRateModifier:       1.0,
		MaxPopulationModifier:    1.0,
		WeaponsCostModifier:      0.75, // 25% cheaper
		DefensesCostModifier:     1.0,
		StarbaseCostModifier:     1.0,
		IntrinsicCloakPercent:    0.0,
		CargoAffectsCloak:        true,
		MineTravelBonus:          0,
		CanBuildMineFields:       false,
		CanBuildAdvancedDefenses: false,
		CanBuildSmartBombs:       true,
		CanLiveOnPlanets:         true,
		StartingTechEnergy:       1,
		StartingTechWeapons:      6,
		StartingTechPropulsion:   1,
	},

	// 3: CA - Claim Adjuster
	{
		Index:                         3,
		Code:                          "CA",
		Name:                          "Claim Adjuster",
		Desc:                          "You are an expert at fiddling with planetary environments. You start the game with Tech 6 in Biotechnology and a ship capable of terraforming planets from orbit. You can arm your ships with bombs that unterraform enemy worlds. Terraforming costs you nothing and planets you leave revert to their original environments. Planets you own have up to a 10% chance of permanently improving an environment variable by 1% per year.",
		PointCost:                     -10,
		GrowthRateModifier:            1.0,
		MaxPopulationModifier:         1.0,
		WeaponsCostModifier:           1.0,
		DefensesCostModifier:          1.0,
		StarbaseCostModifier:          1.0,
		IntrinsicCloakPercent:         0.0,
		CargoAffectsCloak:             true,
		MineTravelBonus:               0,
		CanBuildMineFields:            true,
		CanBuildAdvancedDefenses:      true,
		CanBuildSmartBombs:            true,
		CanLiveOnPlanets:              true,
		FreeTerraforming:              true,
		TerraformingImprovementChance: 0.10, // 10% per year
		PlanetsRevertWhenAbandoned:    true,
		StartingTechBiotech:           6,
	},

	// 4: IS - Inner-Strength
	{
		Index:                             4,
		Code:                              "IS",
		Name:                              "Inner-Strength",
		Desc:                              "You are strong and hard to defeat. Your colonists repel attacks better, your ships heal faster, you have special battle devices that protect your ships and can lay Speed Trap mine fields. You have a device that acts as both a shield and armor. Your peace-loving people refuse to build Smart Bombs. Planetary defenses cost you 40% less, though weapons cost you 25% more. Your colonists are able to reproduce while being transported by your fleets.",
		PointCost:                         100,
		GrowthRateModifier:                1.0,
		MaxPopulationModifier:             1.0,
		WeaponsCostModifier:               1.25, // 25% more expensive
		DefensesCostModifier:              0.60, // 40% cheaper
		StarbaseCostModifier:              1.0,
		IntrinsicCloakPercent:             0.0,
		CargoAffectsCloak:                 true,
		MineTravelBonus:                   0,
		CanBuildMineFields:                true,
		CanBuildAdvancedDefenses:          true,
		CanBuildSmartBombs:                false,
		CanLiveOnPlanets:                  true,
		ColonistsReproduceDuringTransport: true,
	},

	// 5: SD - Space Demolition
	{
		Index:                    5,
		Code:                     "SD",
		Name:                     "Space Demolition",
		Desc:                     "You are an expert in laying mine fields. You have a vast array of mine types at your disposal and two unique hull design which are made for mine dispersal. Your mine fields act as scanners and you have the ability to remote detonate your own Standard mine fields. You may safely travel two warps speeds faster than the stated limits through enemy mine fields. You start the game with 2 mine laying ships and Tech 2 in Propulsion and BioTech.",
		PointCost:                150,
		GrowthRateModifier:       1.0,
		MaxPopulationModifier:    1.0,
		WeaponsCostModifier:      1.0,
		DefensesCostModifier:     1.0,
		StarbaseCostModifier:     1.0,
		IntrinsicCloakPercent:    0.0,
		CargoAffectsCloak:        true,
		MineTravelBonus:          2, // +2 warp in enemy mine fields
		MineFieldsActAsScanners:  true,
		CanRemoteDetonateMines:   true,
		CanBuildMineFields:       true,
		CanBuildAdvancedDefenses: true,
		CanBuildSmartBombs:       true,
		CanLiveOnPlanets:         true,
		StartingTechPropulsion:   2,
		StartingTechBiotech:      2,
	},

	// 6: PP - Packet Physics
	{
		Index:                    6,
		Code:                     "PP",
		Name:                     "Packet Physics",
		Desc:                     "Your race excels at accelerating mineral packets to distant planets. You start with a Warp 5 accelerator at your home starbase and Tech 4 Energy. You will eventually be able to fling packets at the mind numbing speed of Warp 13. You can fling smaller packets and all of your packets have penetrating scanners embedded in them. You will start the game owning a second planet some distance away if the universe size isn't tiny. Packets you fling that aren't fully caught have a chance of terraforming the target planet.",
		PointCost:                -120,
		GrowthRateModifier:       1.0,
		MaxPopulationModifier:    1.0,
		WeaponsCostModifier:      1.0,
		DefensesCostModifier:     1.0,
		StarbaseCostModifier:     1.0,
		IntrinsicCloakPercent:    0.0,
		CargoAffectsCloak:        true,
		MineTravelBonus:          0,
		PacketsHavePenScanner:    true,
		MaxPacketWarp:            13,
		CanBuildMineFields:       true,
		CanBuildAdvancedDefenses: true,
		CanBuildSmartBombs:       true,
		CanLiveOnPlanets:         true,
		StartingTechEnergy:       4,
	},

	// 7: IT - Interstellar Traveler
	{
		Index:                    7,
		Code:                     "IT",
		Name:                     "Interstellar Traveler",
		Desc:                     "Your race excels in building stargates. You start with Tech 5 in Propulsion and Construction. You start the game with a second planet if the universe size isn't tiny. Both planets have stargates. Eventually you may build stargates which have unlimited capabilities. Stargates cost you 25% less to build. Your race can automatically scan any enemy planet with a stargate which is in range of one of your stargates. Exceeding the safety limits of stargates is less likely to kill your ships.",
		PointCost:                -180,
		GrowthRateModifier:       1.0,
		MaxPopulationModifier:    1.0,
		WeaponsCostModifier:      1.0,
		DefensesCostModifier:     1.0,
		StarbaseCostModifier:     0.75, // 25% cheaper (for stargates)
		IntrinsicCloakPercent:    0.0,
		CargoAffectsCloak:        true,
		MineTravelBonus:          0,
		CanScanEnemyStargates:    true,
		StargateSafetyBonus:      true,
		CanBuildMineFields:       true,
		CanBuildAdvancedDefenses: true,
		CanBuildSmartBombs:       true,
		CanLiveOnPlanets:         true,
		StartingTechPropulsion:   5,
		StartingTechConstruction: 5,
	},

	// 8: AR - Alternate Reality
	{
		Index:                     8,
		Code:                      "AR",
		Name:                      "Alternate Reality",
		Desc:                      "Your race developed in an alternate plane. Your people cannot survive on planets and live in orbit on your starbases, which are 20% cheaper to build. You cannot build planetary installations, but your people have an intrinsic ability to mine and scan for enemy fleets. You can remote mine your own worlds. If a starbase is destroyed, all your colonists orbiting that world are killed. Your population maximums are determined by the type of starbase you have. You will eventually be able to build the Death Star.",
		PointCost:                 -90,
		GrowthRateModifier:        1.0,
		MaxPopulationModifier:     1.0, // Max pop determined by starbase type
		WeaponsCostModifier:       1.0,
		DefensesCostModifier:      1.0,
		StarbaseCostModifier:      0.80, // 20% cheaper
		IntrinsicCloakPercent:     0.0,
		CargoAffectsCloak:         true,
		MineTravelBonus:           0,
		HasIntrinsicScanner:       true,
		IntrinsicScannerRangeFunc: arIntrinsicScannerRange, // sqrt(pop/10)
		CanBuildMineFields:        true,
		CanBuildAdvancedDefenses:  true,
		CanBuildSmartBombs:        true,
		CanLiveOnPlanets:          false, // Must live on starbases
	},

	// 9: JOAT - Jack of All Trades
	{
		Index:                          9,
		Code:                           "JOAT",
		Name:                           "Jack of All Trades",
		Desc:                           "Your race does not specialize in a single area. You start the game with Tech 3 in all areas and an assortment of ships. Your Scout, Destroyer, and Frigate hulls have a built-in penetrating scanner which grows more powerful as your Electronics tech increases. Your maximum planetary population is 20% greater than other races.",
		PointCost:                      66,
		GrowthRateModifier:             1.0,
		MaxPopulationModifier:          1.20, // 20% greater
		WeaponsCostModifier:            1.0,
		DefensesCostModifier:           1.0,
		StarbaseCostModifier:           1.0,
		IntrinsicCloakPercent:          0.0,
		CargoAffectsCloak:              true,
		MineTravelBonus:                0,
		HasFleetIntrinsicScanner:       true,
		FleetIntrinsicScannerHulls:     []int{HullScout, HullFrigate, HullDestroyer},
		FleetIntrinsicScannerRangeFunc: joatFleetIntrinsicScanner,
		CanBuildMineFields:             true,
		CanBuildAdvancedDefenses:       true,
		CanBuildSmartBombs:             true,
		CanLiveOnPlanets:               true,
		StartingTechEnergy:             3,
		StartingTechWeapons:            3,
		StartingTechPropulsion:         3,
		StartingTechConstruction:       3,
		StartingTechElectronics:        3,
		StartingTechBiotech:            3,
	},
}

// GetPRT returns the PRT data for the given PRT index.
// Returns nil if the index is out of range.
func GetPRT(prtIndex int) *PRT {
	if prtIndex < 0 || prtIndex >= len(AllPRTs) {
		return nil
	}
	return &AllPRTs[prtIndex]
}

// GetPRTByCode returns the PRT data for the given code (e.g., "SS", "AR").
// Returns nil if not found.
func GetPRTByCode(code string) *PRT {
	for i := range AllPRTs {
		if AllPRTs[i].Code == code {
			return &AllPRTs[i]
		}
	}
	return nil
}

// IntrinsicScannerRange returns the intrinsic scanner range for a PRT at the given population.
// Returns 0 if the PRT has no intrinsic scanner. Used for AR starbases.
func (p *PRT) IntrinsicScannerRange(population int64) int {
	if !p.HasIntrinsicScanner || p.IntrinsicScannerRangeFunc == nil {
		return 0
	}
	return p.IntrinsicScannerRangeFunc(population)
}

// HasFleetIntrinsicScannerForHull returns true if this PRT has intrinsic scanners
// for the given hull type.
func (p *PRT) HasFleetIntrinsicScannerForHull(hullId int) bool {
	if !p.HasFleetIntrinsicScanner {
		return false
	}
	for _, h := range p.FleetIntrinsicScannerHulls {
		if h == hullId {
			return true
		}
	}
	return false
}

// FleetIntrinsicScannerRange returns the intrinsic scanner stats for a fleet ship
// based on the owner's tech level. Returns zero stats if the PRT doesn't have
// fleet intrinsic scanners.
func (p *PRT) FleetIntrinsicScannerRange(electronicsLevel int) ScannerStats {
	if !p.HasFleetIntrinsicScanner || p.FleetIntrinsicScannerRangeFunc == nil {
		return ScannerStats{}
	}
	return p.FleetIntrinsicScannerRangeFunc(electronicsLevel)
}
