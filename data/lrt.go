package data

import "math"

// LRT represents a Lesser Race Trait with all its specific abilities.
// LRTs are optional traits that can be selected during race creation to
// provide advantages at the cost of race points, or disadvantages that
// grant additional race points.
type LRT struct {
	// Basic identification
	Index   int    // 0-13 index matching blocks.LRTxxx bitmask positions
	Bitmask uint16 // Bitmask value (1 << Index)
	Code    string // Short code: "IFE", "TT", etc.
	Name    string // Full name: "Improved Fuel Efficiency", etc.
	Desc    string // Description text

	// Race point cost (negative = advantage/costs points, positive = disadvantage/grants points)
	PointCost int

	// Fuel efficiency (IFE)
	FuelEfficiencyBonus float64 // IFE: 0.15 (15% less fuel)
	UnlocksFuelMizer    bool    // IFE: true
	UnlocksGalaxyScoop  bool    // IFE: true

	// Terraforming (TT)
	MaxTerraformPercent      int     // TT: 30 (can terraform up to 30%)
	TerraformingCostModifier float64 // TT: 0.70 (30% cheaper)

	// Remote mining (ARM, OBRM)
	UnlocksAdvancedMiningHulls bool    // ARM: true
	StartingMidgetMiners       int     // ARM: 2
	OnlyBasicMining            bool    // OBRM: true (only Mini-Miner available)
	MaxPopulationBonus         float64 // OBRM: 0.10 (10% more max pop)

	// Starbases (ISB)
	UnlocksStardock             bool    // ISB: true
	UnlocksUltraStation         bool    // ISB: true
	StarbaseCostModifier        float64 // ISB: 0.80 (20% cheaper)
	StarbaseCloakPercent        float64 // ISB: 0.20 (20% cloaked)
	FactoriesUseOneKTLessGerm   bool    // ISB: true (factories cost 1kT less Germanium)

	// Research (GR)
	ResearchToCurrentField float64 // GR: 0.50 (50% to current field)
	ResearchToOtherFields  float64 // GR: 0.15 (15% to each other field)

	// Recycling (UR)
	StarbaseRecyclingPercent float64 // UR: 0.90 (90% minerals recovered at starbase)
	PlanetRecyclingPercent   float64 // UR: 0.45 (45% minerals recovered at planet)

	// Mineral alchemy (MA)
	MineralAlchemyCostDivisor int // MA: 4 (costs 25 resources instead of 100)

	// Engines (NRSE, CE)
	NoRamScoopEngines       bool    // NRSE: true
	UnlocksInterspace10     bool    // NRSE: true
	CheapEngines            bool    // CE: true
	EngineCostModifier      float64 // CE: 0.50 (half cost)
	EngineFailureChance     float64 // CE: 0.10 (10% failure above warp 6)
	EngineFailureMinWarp    int     // CE: 7 (failure starts at warp 7+)

	// Scanners (NAS)
	NoAdvancedScanners            bool    // NAS: true (no penetrating scanners)
	NormalScannerMultiplier       int     // NAS: 2 (2× normal scanner range for equipped scanners)
	ARIntrinsicScannerMultiplier  float64 // NAS: √2 ≈ 1.414 (multiplier for AR intrinsic scanner)

	// Starting population (LSP)
	StartingPopulationModifier float64 // LSP: 0.70 (30% fewer colonists)

	// Technology (BET)
	NewTechCostMultiplier      float64 // BET: 2.0 (new techs cost 2×)
	MiniaturizationPerLevel    float64 // BET: 0.05 (5% per level)
	MaxMiniaturizationPercent  float64 // BET: 0.80 (max 80%)

	// Shields/Armor (RS)
	ShieldStrengthMultiplier float64 // RS: 1.40 (40% stronger)
	ShieldRegenPerRound      float64 // RS: 0.10 (10% regeneration per round)
	ArmorStrengthMultiplier  float64 // RS: 0.50 (50% weaker)

	// Starting tech bonuses
	StartingTechPropulsion int // IFE, CE: +1
}

// AllLRTs contains data for all 14 Lesser Race Traits.
var AllLRTs = []LRT{
	// 0: IFE - Improved Fuel Efficiency
	{
		Index:               0,
		Bitmask:             1 << 0,
		Code:                "IFE",
		Name:                "Improved Fuel Efficiency",
		Desc:                "All engines use 15% less fuel. Gives Fuel Mizer and Galaxy Scoop engines. +1 starting propulsion.",
		PointCost:           -235,
		FuelEfficiencyBonus: 0.15,
		UnlocksFuelMizer:    true,
		UnlocksGalaxyScoop:  true,
		StartingTechPropulsion: 1,
	},

	// 1: TT - Total Terraforming
	{
		Index:                    1,
		Bitmask:                  1 << 1,
		Code:                     "TT",
		Name:                     "Total Terraforming",
		Desc:                     "Allows terraforming by investing solely in Biotech. May terraform up to 30%. Terraforming costs 30% less.",
		PointCost:                -25,
		MaxTerraformPercent:      30,
		TerraformingCostModifier: 0.70,
	},

	// 2: ARM - Advanced Remote Mining
	{
		Index:                      2,
		Bitmask:                    1 << 2,
		Code:                       "ARM",
		Name:                       "Advanced Remote Mining",
		Desc:                       "Gives three additional mining hulls and two new robots. Start with two Midget Miners.",
		PointCost:                  -159,
		UnlocksAdvancedMiningHulls: true,
		StartingMidgetMiners:       2,
	},

	// 3: ISB - Improved Starbases
	{
		Index:                       3,
		Bitmask:                     1 << 3,
		Code:                        "ISB",
		Name:                        "Improved Starbases",
		Desc:                        "Gives Stardock and Ultra Station starbase designs. Starbases cost 20% less and are 20% cloaked.",
		PointCost:                   -201,
		UnlocksStardock:             true,
		UnlocksUltraStation:         true,
		StarbaseCostModifier:        0.80,
		StarbaseCloakPercent:        0.20,
		FactoriesUseOneKTLessGerm:   true,
	},

	// 4: GR - Generalized Research
	{
		Index:                  4,
		Bitmask:                1 << 4,
		Code:                   "GR",
		Name:                   "Generalized Research",
		Desc:                   "Only 50% of resources go to current research field, but 15% goes to each other field.",
		PointCost:              40,
		ResearchToCurrentField: 0.50,
		ResearchToOtherFields:  0.15,
	},

	// 5: UR - Ultimate Recycling
	{
		Index:                    5,
		Bitmask:                  1 << 5,
		Code:                     "UR",
		Name:                     "Ultimate Recycling",
		Desc:                     "When scrapping at starbase, recover 90% of minerals and some resources. At planet, recover 45%.",
		PointCost:                -240,
		StarbaseRecyclingPercent: 0.90,
		PlanetRecyclingPercent:   0.45,
	},

	// 6: MA - Mineral Alchemy
	{
		Index:                     6,
		Bitmask:                   1 << 6,
		Code:                      "MA",
		Name:                      "Mineral Alchemy",
		Desc:                      "Turn resources into minerals 4× more efficiently (25 resources instead of 100).",
		PointCost:                 -155,
		MineralAlchemyCostDivisor: 4,
	},

	// 7: NRSE - No Ram Scoop Engines
	{
		Index:               7,
		Bitmask:             1 << 7,
		Code:                "NRSE",
		Name:                "No Ram Scoop Engines",
		Desc:                "No engines that travel at warp 5+ burning no fuel. However, Interspace 10 engine is available.",
		PointCost:           160,
		NoRamScoopEngines:   true,
		UnlocksInterspace10: true,
	},

	// 8: CE - Cheap Engines
	{
		Index:                  8,
		Bitmask:                1 << 8,
		Code:                   "CE",
		Name:                   "Cheap Engines",
		Desc:                   "Engines cost half price, but 10% chance of failure above warp 6. +1 starting propulsion.",
		PointCost:              240,
		CheapEngines:           true,
		EngineCostModifier:     0.50,
		EngineFailureChance:    0.10,
		EngineFailureMinWarp:   7,
		StartingTechPropulsion: 1,
	},

	// 9: OBRM - Only Basic Remote Mining
	{
		Index:              9,
		Bitmask:            1 << 9,
		Code:               "OBRM",
		Name:               "Only Basic Remote Mining",
		Desc:               "Only Mini-Miner available. Overrides ARM. Max population per planet increased by 10%.",
		PointCost:          255,
		OnlyBasicMining:    true,
		MaxPopulationBonus: 0.10,
	},

	// 10: NAS - No Advanced Scanners
	{
		Index:                        10,
		Bitmask:                      1 << 10,
		Code:                         "NAS",
		Name:                         "No Advanced Scanners",
		Desc:                         "No planet-penetrating scanners available. Conventional scanners have double range. AR intrinsic scanner is multiplied by √2.",
		PointCost:                    325,
		NoAdvancedScanners:           true,
		NormalScannerMultiplier:      2,
		ARIntrinsicScannerMultiplier: 1.4142135623730951, // math.Sqrt2
	},

	// 11: LSP - Low Starting Population
	{
		Index:                      11,
		Bitmask:                    1 << 11,
		Code:                       "LSP",
		Name:                       "Low Starting Population",
		Desc:                       "Start with 30% fewer colonists.",
		PointCost:                  180,
		StartingPopulationModifier: 0.70,
	},

	// 12: BET - Bleeding Edge Technology
	{
		Index:                     12,
		Bitmask:                   1 << 12,
		Code:                      "BET",
		Name:                      "Bleeding Edge Technology",
		Desc:                      "New techs cost 2× until you exceed all requirements by one level. Miniaturization is 5% per level (max 80%).",
		PointCost:                 70,
		NewTechCostMultiplier:     2.0,
		MiniaturizationPerLevel:   0.05,
		MaxMiniaturizationPercent: 0.80,
	},

	// 13: RS - Regenerating Shields
	{
		Index:                    13,
		Bitmask:                  1 << 13,
		Code:                     "RS",
		Name:                     "Regenerating Shields",
		Desc:                     "Shields are 40% stronger and regenerate 10% per round. Armor is only 50% strength.",
		PointCost:                30,
		ShieldStrengthMultiplier: 1.40,
		ShieldRegenPerRound:      0.10,
		ArmorStrengthMultiplier:  0.50,
	},
}

// GetLRT returns the LRT data for the given LRT index (0-13).
// Returns nil if the index is out of range.
func GetLRT(lrtIndex int) *LRT {
	if lrtIndex < 0 || lrtIndex >= len(AllLRTs) {
		return nil
	}
	return &AllLRTs[lrtIndex]
}

// GetLRTByCode returns the LRT data for the given code (e.g., "IFE", "TT").
// Returns nil if not found.
func GetLRTByCode(code string) *LRT {
	for i := range AllLRTs {
		if AllLRTs[i].Code == code {
			return &AllLRTs[i]
		}
	}
	return nil
}

// GetLRTByBitmask returns the LRT data for the given bitmask value.
// Returns nil if not found.
func GetLRTByBitmask(bitmask uint16) *LRT {
	for i := range AllLRTs {
		if AllLRTs[i].Bitmask == bitmask {
			return &AllLRTs[i]
		}
	}
	return nil
}

// GetLRTsFromBitmask returns all LRTs that are set in the given bitmask.
func GetLRTsFromBitmask(bitmask uint16) []*LRT {
	var lrts []*LRT
	for i := range AllLRTs {
		if (bitmask & AllLRTs[i].Bitmask) != 0 {
			lrts = append(lrts, &AllLRTs[i])
		}
	}
	return lrts
}

// HasLRT checks if a specific LRT is set in the given bitmask.
func HasLRT(bitmask uint16, lrtIndex int) bool {
	if lrtIndex < 0 || lrtIndex >= len(AllLRTs) {
		return false
	}
	return (bitmask & AllLRTs[lrtIndex].Bitmask) != 0
}

// IsGoodLRT returns true if this LRT costs race points (i.e., it's an advantage).
func (l *LRT) IsGoodLRT() bool {
	return l.PointCost < 0
}

// IsBadLRT returns true if this LRT grants race points (i.e., it's a disadvantage).
func (l *LRT) IsBadLRT() bool {
	return l.PointCost >= 0
}

// LRT index constants for convenience
const (
	LRTIndexIFE  = 0
	LRTIndexTT   = 1
	LRTIndexARM  = 2
	LRTIndexISB  = 3
	LRTIndexGR   = 4
	LRTIndexUR   = 5
	LRTIndexMA   = 6
	LRTIndexNRSE = 7
	LRTIndexCE   = 8
	LRTIndexOBRM = 9
	LRTIndexNAS  = 10
	LRTIndexLSP  = 11
	LRTIndexBET  = 12
	LRTIndexRS   = 13
)

// ARNASScannerRange calculates the scanner range for an AR player with NAS LRT.
// Formula: floor(floor(sqrt(pop/10)) × √2)
// The base AR intrinsic scanner range is first truncated to an integer,
// then multiplied by √2, and truncated again.
func ARNASScannerRange(population int64) int {
	if population <= 0 {
		return 0
	}
	// Base AR range (truncated)
	baseRange := int(math.Sqrt(float64(population) / 10.0))
	// Apply NAS √2 multiplier
	nas := GetLRTByCode("NAS")
	if nas == nil || nas.ARIntrinsicScannerMultiplier == 0 {
		return baseRange
	}
	return int(float64(baseRange) * nas.ARIntrinsicScannerMultiplier)
}
