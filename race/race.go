// Package race provides race building and point calculation for Stars!
//
// Predefined race templates from step 1 of the race wizard:
//   - Humanoid() - balanced race with Jack of All Trades
//   - Rabbitoid() - fast-breeding Interstellar Traveler
//   - Insectoid() - gravity-immune War Monger
//   - Nucleotid() - stealthy Super Stealth with all research expensive but starting at Tech 3
//   - Silicanoid() - all-immune Hyper Expansion with efficient factories
//   - Antetheral() - mine-laying Space Demolition with narrow gravity
//   - Random() / RandomWithSeed() - randomize all settings
package race

import "math/rand"

// Primary Race Trait (PRT) constants for use with Builder.PRT()
const (
	PRTHyperExpansion       = 0 // HE - Hyper Expansion
	PRTSuperStealth         = 1 // SS - Super Stealth
	PRTWarMonger            = 2 // WM - War Monger
	PRTClaimAdjuster        = 3 // CA - Claim Adjuster
	PRTInnerStrength        = 4 // IS - Inner Strength
	PRTSpaceDemolition      = 5 // SD - Space Demolition
	PRTPacketPhysics        = 6 // PP - Packet Physics
	PRTInterstellarTraveler = 7 // IT - Interstellar Traveler
	PRTAlternateReality     = 8 // AR - Alternate Reality
	PRTJackOfAllTrades      = 9 // JOAT - Jack of All Trades
)

// Lesser Race Trait (LRT) index constants for use with Builder.AddLRT()/RemoveLRT()
const (
	LRTImprovedFuelEfficiency = 0  // IFE
	LRTTotalTerraforming      = 1  // TT
	LRTAdvancedRemoteMining   = 2  // ARM
	LRTImprovedStarbases      = 3  // ISB
	LRTGeneralizedResearch    = 4  // GR
	LRTUltimateRecycling      = 5  // UR
	LRTMineralAlchemy         = 6  // MA
	LRTNoRamScoopEngines      = 7  // NRSE
	LRTCheapEngines           = 8  // CE
	LRTOnlyBasicRemoteMining  = 9  // OBRM
	LRTNoAdvancedScanners     = 10 // NAS
	LRTLowStartingPopulation  = 11 // LSP
	LRTBleedingEdgeTechnology = 12 // BET
	LRTRegeneratingShields    = 13 // RS
)

// LRTs combines multiple LRT indices into a single bitmask.
// This provides a cleaner way to specify LRTs without bit manipulation.
//
// Example:
//
//	LRT: LRTs(LRTImprovedFuelEfficiency, LRTTotalTerraforming, LRTCheapEngines)
func LRTs(lrts ...int) uint16 {
	var result uint16
	for _, lrt := range lrts {
		result |= (1 << lrt)
	}
	return result
}

// Race represents a complete Stars! race configuration.
type Race struct {
	// Identity
	SingularName string
	PluralName   string
	Password     string
	Icon         int // Race icon/logo (0-31)

	// Traits
	PRT int    // Primary Race Trait (0-9)
	LRT uint16 // Lesser Race Traits bitmask

	// Habitability (0-100 scale)
	// Center is the ideal value, Width is half the acceptable range
	GravityCenter     int
	GravityWidth      int
	GravityImmune     bool
	TemperatureCenter int
	TemperatureWidth  int
	TemperatureImmune bool
	RadiationCenter   int
	RadiationWidth    int
	RadiationImmune   bool

	// Growth
	GrowthRate int // 1-20 (represents 1%-20%)

	// Economy - Population
	ColonistsPerResource int // 700-2500

	// Economy - Factories
	FactoryOutput        int  // 5-25 (resources per 10 factories)
	FactoryCost          int  // 5-25 (resources to build 1 factory)
	FactoryCount         int  // 5-25 (max factories per 10k colonists)
	FactoriesUseLessGerm bool // Factories cost 1kT less germanium

	// Economy - Mines
	MineOutput int // 5-25 (kT per 10 mines)
	MineCost   int // 2-15 (resources to build 1 mine)
	MineCount  int // 5-25 (max mines per 10k colonists)

	// Research costs (0=Extra, 1=Standard, 2=Less)
	ResearchEnergy       int
	ResearchWeapons      int
	ResearchPropulsion   int
	ResearchConstruction int
	ResearchElectronics  int
	ResearchBiotech      int
	TechsStartHigh       bool // Expensive fields start at Tech 3

	// Leftover points allocation
	LeftoverPointsOn LeftoverPointsOption
}

// LeftoverPointsOption specifies where to spend leftover advantage points.
type LeftoverPointsOption int

const (
	LeftoverSurfaceMinerals LeftoverPointsOption = iota
	LeftoverMines
	LeftoverFactories
	LeftoverDefenses
	LeftoverMineralConcentration
)

// Research cost levels
const (
	ResearchCostExtra    = 0 // Costs 75% more
	ResearchCostStandard = 1 // Normal cost
	ResearchCostLess     = 2 // Costs 50% less
)

// Default returns a new Race with sensible defaults for the race builder.
// This provides a good starting point for creating custom races.
// For the exact Stars! predefined Humanoid race, use Humanoid() instead.
func Default() *Race {
	return &Race{
		SingularName: "Humanoid",
		PluralName:   "Humanoids",

		PRT: PRTJackOfAllTrades,
		LRT: 0, // No LRTs

		// Habitability - centered at 50, range 25-75
		GravityCenter:     50,
		GravityWidth:      25,
		TemperatureCenter: 50,
		TemperatureWidth:  25,
		RadiationCenter:   50,
		RadiationWidth:    25,

		GrowthRate:           15,   // 15%
		ColonistsPerResource: 1000, // 1000 colonists per resource

		FactoryOutput:        10, // 10 resources per 10 factories
		FactoryCost:          10, // 10 resources to build 1 factory
		FactoryCount:         10, // 10 factories per 10k colonists
		FactoriesUseLessGerm: false,

		MineOutput: 10, // 10 kT per 10 mines
		MineCost:   5,  // 5 resources to build 1 mine
		MineCount:  10, // 10 mines per 10k colonists

		ResearchEnergy:       ResearchCostStandard,
		ResearchWeapons:      ResearchCostStandard,
		ResearchPropulsion:   ResearchCostStandard,
		ResearchConstruction: ResearchCostStandard,
		ResearchElectronics:  ResearchCostStandard,
		ResearchBiotech:      ResearchCostStandard,
		TechsStartHigh:       false,

		LeftoverPointsOn: LeftoverSurfaceMinerals,
	}
}

// Humanoid returns the predefined Humanoid race from Stars! race wizard.
// This is the exact configuration produced by selecting "Humanoid" in step 1.
//
// Characteristics:
//   - PRT: Jack of All Trades (starts with Tech 3 in all areas, 20% more max pop)
//   - No LRTs
//   - Wide habitability: 1 in 2 planets habitable (0.22g-4.40g, -140°C-140°C, 15mR-85mR)
//   - 15% growth rate
//   - Balanced economy (10/10/10/10/5/10)
//   - Standard research costs
//   - 25 advantage points to spend on surface minerals
func Humanoid() *Race {
	return &Race{
		SingularName: "Humanoid",
		PluralName:   "Humanoids",
		Icon:         1, // Icon 1 in Stars!

		PRT: PRTJackOfAllTrades,
		LRT: 0, // No LRTs

		// Habitability - centered at 50, range 15-85 (internal 0-100 scale)
		// Display values: 0.22g-4.40g, -140°C-140°C, 15mR-85mR
		GravityCenter:     50,
		GravityWidth:      35,
		TemperatureCenter: 50,
		TemperatureWidth:  35,
		RadiationCenter:   50,
		RadiationWidth:    35,

		GrowthRate:           15,   // 15%
		ColonistsPerResource: 1000, // 1000 colonists per resource

		FactoryOutput:        10, // 10 resources per 10 factories
		FactoryCost:          10, // 10 resources to build 1 factory
		FactoryCount:         10, // 10 factories per 10k colonists
		FactoriesUseLessGerm: false,

		MineOutput: 10, // 10 kT per 10 mines
		MineCost:   5,  // 5 resources to build 1 mine
		MineCount:  10, // 10 mines per 10k colonists

		ResearchEnergy:       ResearchCostStandard,
		ResearchWeapons:      ResearchCostStandard,
		ResearchPropulsion:   ResearchCostStandard,
		ResearchConstruction: ResearchCostStandard,
		ResearchElectronics:  ResearchCostStandard,
		ResearchBiotech:      ResearchCostStandard,
		TechsStartHigh:       false,

		LeftoverPointsOn: LeftoverSurfaceMinerals,
	}
}

// Rabbitoid returns the predefined Rabbitoid race from Stars! race wizard.
// This is the exact configuration produced by selecting "Rabbitoid" in step 1.
//
// Characteristics:
//   - PRT: Interstellar Traveler (starts with Tech 5 in Prop/Const, 2nd planet if not tiny)
//   - LRTs: IFE, TT, CE, NAS
//   - Narrow habitability: 1 in 9 planets habitable (0.17g-1.24g, -60°C-124°C, 13mR-53mR)
//   - 20% growth rate (fast breeding)
//   - Efficient factories (10/9/17, less germanium), standard mines (10/9/10)
//   - Cheap propulsion/biotech research, expensive energy/weapons
//   - 32 advantage points to spend on mineral concentration
func Rabbitoid() *Race {
	return &Race{
		SingularName: "Rabbitoid",
		PluralName:   "Rabbitoids",
		Icon:         12, // Icon 12 in Stars!

		PRT: PRTInterstellarTraveler,
		LRT: LRTs(LRTImprovedFuelEfficiency, LRTTotalTerraforming, LRTCheapEngines, LRTNoAdvancedScanners),

		// Habitability - narrow ranges (1 in 9 planets habitable)
		// Display values: 0.17g-1.24g, -60°C-124°C, 13mR-53mR
		GravityCenter:     33, // Range 10-56
		GravityWidth:      23,
		TemperatureCenter: 58, // Range 35-81
		TemperatureWidth:  23,
		RadiationCenter:   33, // Range 13-53
		RadiationWidth:    20,

		GrowthRate:           20,   // 20% - fast breeding
		ColonistsPerResource: 1000, // 1000 colonists per resource

		FactoryOutput:        10, // 10 resources per 10 factories
		FactoryCost:          9,  // 9 resources to build 1 factory
		FactoryCount:         17, // 17 factories per 10k colonists
		FactoriesUseLessGerm: true,

		MineOutput: 10, // 10 kT per 10 mines
		MineCost:   9,  // 9 resources to build 1 mine
		MineCount:  10, // 10 mines per 10k colonists

		ResearchEnergy:       ResearchCostExtra,    // 75% more expensive
		ResearchWeapons:      ResearchCostExtra,    // 75% more expensive
		ResearchPropulsion:   ResearchCostLess,     // 50% less expensive
		ResearchConstruction: ResearchCostStandard, // Standard
		ResearchElectronics:  ResearchCostStandard, // Standard
		ResearchBiotech:      ResearchCostLess,     // 50% less expensive
		TechsStartHigh:       false,

		LeftoverPointsOn: LeftoverMineralConcentration,
	}
}

// Insectoid returns the predefined Insectoid race from Stars! race wizard.
// This is the exact configuration produced by selecting "Insectoid" in step 1.
//
// Characteristics:
//   - PRT: War Monger (colonists attack better, weapons 25% cheaper, Tech 6 weapons)
//   - LRTs: ISB, CE, RS
//   - Gravity immune, wide temperature range, narrow radiation (1 in 3 planets)
//   - 10% growth rate
//   - Standard factories, efficient mines (9 output, 10 cost, 6 count)
//   - Cheap energy/weapons/propulsion/construction, expensive biotech
//   - 43 advantage points to spend on mining
func Insectoid() *Race {
	return &Race{
		SingularName: "Insectoid",
		PluralName:   "Insectoids",
		Icon:         4, // Icon 4 in Stars!

		PRT: PRTWarMonger,
		LRT: LRTs(LRTImprovedStarbases, LRTCheapEngines, LRTRegeneratingShields),

		// Habitability - gravity immune, wide temp, narrow radiation
		// 1 in 3 planets habitable
		GravityImmune:     true,
		GravityCenter:     50, // Ignored when immune
		GravityWidth:      25,
		TemperatureCenter: 50, // Range 0-100 (full range: -200°C to 200°C)
		TemperatureWidth:  50,
		RadiationCenter:   85, // Range 70-100 (70mR to 100mR)
		RadiationWidth:    15,

		GrowthRate:           10,   // 10%
		ColonistsPerResource: 1000, // 1000 colonists per resource

		FactoryOutput:        10, // 10 resources per 10 factories
		FactoryCost:          10, // 10 resources to build 1 factory
		FactoryCount:         10, // 10 factories per 10k colonists
		FactoriesUseLessGerm: false,

		MineOutput: 9,  // 9 kT per 10 mines
		MineCost:   10, // 10 resources to build 1 mine
		MineCount:  6,  // 6 mines per 10k colonists

		ResearchEnergy:       ResearchCostLess,     // 50% less expensive
		ResearchWeapons:      ResearchCostLess,     // 50% less expensive
		ResearchPropulsion:   ResearchCostLess,     // 50% less expensive
		ResearchConstruction: ResearchCostLess,     // 50% less expensive
		ResearchElectronics:  ResearchCostStandard, // Standard
		ResearchBiotech:      ResearchCostExtra,    // 75% more expensive
		TechsStartHigh:       false,

		LeftoverPointsOn: LeftoverMines,
	}
}

// Nucleotid returns the predefined Nucleotid race from Stars! race wizard.
// This is the exact configuration produced by selecting "Nucleotid" in step 1.
//
// Characteristics:
//   - PRT: Super Stealth (75% cloaking, stealth ships/scanners, steal minerals)
//   - LRTs: ARM, ISB
//   - Gravity immune, wide temperature/radiation (virtually all planets habitable)
//   - 10% growth rate
//   - Efficient colonists (900/resource), expensive mines (10/15/5)
//   - All research 75% extra but starts at Tech 3
//   - 11 advantage points to spend on factories
func Nucleotid() *Race {
	return &Race{
		SingularName: "Nucleotid",
		PluralName:   "Nucleotids",
		Icon:         25, // Icon 25 in Stars!

		PRT: PRTSuperStealth,
		LRT: LRTs(LRTAdvancedRemoteMining, LRTImprovedStarbases),

		// Habitability - gravity immune, wide temp/rad
		// "Virtually all planets will be habitable"
		GravityImmune:     true,
		GravityCenter:     50, // Ignored when immune
		GravityWidth:      25,
		TemperatureCenter: 50, // Range 12-88 (-152°C to 152°C)
		TemperatureWidth:  38,
		RadiationCenter:   50, // Range 0-100 (full range)
		RadiationWidth:    50,

		GrowthRate:           10,  // 10%
		ColonistsPerResource: 900, // 900 colonists per resource (efficient)

		FactoryOutput:        10, // 10 resources per 10 factories
		FactoryCost:          10, // 10 resources to build 1 factory
		FactoryCount:         10, // 10 factories per 10k colonists
		FactoriesUseLessGerm: false,

		MineOutput: 10, // 10 kT per 10 mines
		MineCost:   15, // 15 resources to build 1 mine (expensive)
		MineCount:  5,  // 5 mines per 10k colonists

		ResearchEnergy:       ResearchCostExtra, // 75% more expensive
		ResearchWeapons:      ResearchCostExtra, // 75% more expensive
		ResearchPropulsion:   ResearchCostExtra, // 75% more expensive
		ResearchConstruction: ResearchCostExtra, // 75% more expensive
		ResearchElectronics:  ResearchCostExtra, // 75% more expensive
		ResearchBiotech:      ResearchCostExtra, // 75% more expensive
		TechsStartHigh:       true,              // All expensive fields start at Tech 3

		LeftoverPointsOn: LeftoverFactories,
	}
}

// Silicanoid returns the predefined Silicanoid race from Stars! race wizard.
// This is the exact configuration produced by selecting "Silicanoid" in step 1.
//
// Characteristics:
//   - PRT: Hyper Expansion (cheap colony ships, 2x growth, half max pop)
//   - LRTs: IFE, UR, OBRM, BET
//   - All three immunities (gravity, temperature, radiation) - all planets habitable
//   - 6% growth rate (doubled to 12% by HE)
//   - Efficient colonists (800/resource), efficient factories (12/12/15)
//   - Cheap propulsion/construction, expensive biotech
//   - 9 advantage points to spend on factories
func Silicanoid() *Race {
	return &Race{
		SingularName: "Silicanoid",
		PluralName:   "Silicanoids",
		Icon:         5, // Icon 5 in Stars!

		PRT: PRTHyperExpansion,
		LRT: LRTs(LRTImprovedFuelEfficiency, LRTUltimateRecycling, LRTOnlyBasicRemoteMining, LRTBleedingEdgeTechnology),

		// Habitability - all three immunities
		// "All planets will be habitable"
		GravityImmune:     true,
		GravityCenter:     50, // Ignored when immune
		GravityWidth:      25,
		TemperatureImmune: true,
		TemperatureCenter: 50, // Ignored when immune
		TemperatureWidth:  25,
		RadiationImmune:   true,
		RadiationCenter:   50, // Ignored when immune
		RadiationWidth:    25,

		GrowthRate:           6,   // 6% (doubled to 12% by HE)
		ColonistsPerResource: 800, // 800 colonists per resource (efficient)

		FactoryOutput:        12, // 12 resources per 10 factories
		FactoryCost:          12, // 12 resources to build 1 factory
		FactoryCount:         15, // 15 factories per 10k colonists
		FactoriesUseLessGerm: false,

		MineOutput: 10, // 10 kT per 10 mines
		MineCost:   9,  // 9 resources to build 1 mine
		MineCount:  10, // 10 mines per 10k colonists

		ResearchEnergy:       ResearchCostStandard, // Standard
		ResearchWeapons:      ResearchCostStandard, // Standard
		ResearchPropulsion:   ResearchCostLess,     // 50% less expensive
		ResearchConstruction: ResearchCostLess,     // 50% less expensive
		ResearchElectronics:  ResearchCostStandard, // Standard
		ResearchBiotech:      ResearchCostExtra,    // 75% more expensive
		TechsStartHigh:       false,

		LeftoverPointsOn: LeftoverFactories,
	}
}

// Antetheral returns the predefined Antetheral race from Stars! race wizard.
// This is the exact configuration produced by selecting "Antetheral" in step 1.
//
// Characteristics:
//   - PRT: Space Demolition (mine expert, 2 mine-laying ships, Tech 2 in Prop/Bio)
//   - LRTs: ARM, MA, NRSE, CE, NAS
//   - Narrow gravity (0.12g-0.55g), full temperature, narrow radiation (1 in 12 planets)
//   - 7% growth rate
//   - Very efficient colonists (700/resource), efficient factories (11/10/18)
//   - Cheap energy/propulsion/construction/electronics/biotech, expensive weapons
//   - 7 advantage points to spend on surface minerals
func Antetheral() *Race {
	return &Race{
		SingularName: "Antetheral",
		PluralName:   "Antetherals",
		Icon:         18, // Icon 18 in Stars!

		PRT: PRTSpaceDemolition,
		LRT: LRTs(LRTAdvancedRemoteMining, LRTMineralAlchemy, LRTNoRamScoopEngines, LRTCheapEngines, LRTNoAdvancedScanners),

		// Habitability - narrow gravity, full temp, narrow radiation
		// 1 in 12 planets habitable
		GravityCenter:     15, // Range 0-30 (0.12g to 0.55g)
		GravityWidth:      15,
		TemperatureCenter: 50, // Range 0-100 (full range: -200°C to 200°C)
		TemperatureWidth:  50,
		RadiationCenter:   85, // Range 70-100 (70mR to 100mR)
		RadiationWidth:    15,

		GrowthRate:           7,   // 7%
		ColonistsPerResource: 700, // 700 colonists per resource (very efficient)

		FactoryOutput:        11, // 11 resources per 10 factories
		FactoryCost:          10, // 10 resources to build 1 factory
		FactoryCount:         18, // 18 factories per 10k colonists
		FactoriesUseLessGerm: false,

		MineOutput: 10, // 10 kT per 10 mines
		MineCost:   10, // 10 resources to build 1 mine
		MineCount:  10, // 10 mines per 10k colonists

		ResearchEnergy:       ResearchCostLess,     // 50% less expensive
		ResearchWeapons:      ResearchCostExtra,    // 75% more expensive
		ResearchPropulsion:   ResearchCostLess,     // 50% less expensive
		ResearchConstruction: ResearchCostLess,     // 50% less expensive
		ResearchElectronics:  ResearchCostLess,     // 50% less expensive
		ResearchBiotech:      ResearchCostLess,     // 50% less expensive
		TechsStartHigh:       false,

		LeftoverPointsOn: LeftoverSurfaceMinerals,
	}
}

// Random returns a race with randomized settings using the default random source.
// All generated values are within valid ranges and will pass validation.
// The race name is set to "Random" - callers should set their own name.
func Random() *Race {
	return RandomWithSeed(rand.Int63())
}

// RandomWithSeed returns a race with randomized settings using a seeded random source.
// Using the same seed will produce identical race configurations.
// All generated values are within valid ranges and will pass validation.
func RandomWithSeed(seed int64) *Race {
	rng := rand.New(rand.NewSource(seed))

	// Helper to generate random int in range [min, max]
	randRange := func(min, max int) int {
		return min + rng.Intn(max-min+1)
	}

	// Random LRT selection - each of the 14 LRTs has 50% chance
	var lrt uint16
	for i := 0; i < 14; i++ {
		if rng.Intn(2) == 1 {
			lrt |= (1 << i)
		}
	}

	// Random habitability - each dimension has 20% chance of immunity
	gravityImmune := rng.Intn(5) == 0
	temperatureImmune := rng.Intn(5) == 0
	radiationImmune := rng.Intn(5) == 0

	return &Race{
		SingularName: "Random",
		PluralName:   "Randoms",
		Icon:         rng.Intn(32), // Icons 0-31

		PRT: rng.Intn(10), // 0-9
		LRT: lrt,

		GravityImmune:     gravityImmune,
		GravityCenter:     randRange(0, 100),
		GravityWidth:      randRange(0, 50),
		TemperatureImmune: temperatureImmune,
		TemperatureCenter: randRange(0, 100),
		TemperatureWidth:  randRange(0, 50),
		RadiationImmune:   radiationImmune,
		RadiationCenter:   randRange(0, 100),
		RadiationWidth:    randRange(0, 50),

		GrowthRate:           randRange(1, 20),
		ColonistsPerResource: randRange(700, 2500),

		FactoryOutput:        randRange(5, 25),
		FactoryCost:          randRange(5, 25),
		FactoryCount:         randRange(5, 25),
		FactoriesUseLessGerm: rng.Intn(2) == 1,

		MineOutput: randRange(5, 25),
		MineCost:   randRange(2, 15),
		MineCount:  randRange(5, 25),

		ResearchEnergy:       rng.Intn(3), // 0-2
		ResearchWeapons:      rng.Intn(3),
		ResearchPropulsion:   rng.Intn(3),
		ResearchConstruction: rng.Intn(3),
		ResearchElectronics:  rng.Intn(3),
		ResearchBiotech:      rng.Intn(3),
		TechsStartHigh:       rng.Intn(2) == 1,

		LeftoverPointsOn: LeftoverPointsOption(rng.Intn(5)), // 0-4
	}
}

// Clone creates a deep copy of the race.
func (r *Race) Clone() *Race {
	clone := *r
	return &clone
}

// GravityLow returns the low end of the gravity habitability range.
func (r *Race) GravityLow() int {
	return r.GravityCenter - r.GravityWidth
}

// GravityHigh returns the high end of the gravity habitability range.
func (r *Race) GravityHigh() int {
	return r.GravityCenter + r.GravityWidth
}

// TemperatureLow returns the low end of the temperature habitability range.
func (r *Race) TemperatureLow() int {
	return r.TemperatureCenter - r.TemperatureWidth
}

// TemperatureHigh returns the high end of the temperature habitability range.
func (r *Race) TemperatureHigh() int {
	return r.TemperatureCenter + r.TemperatureWidth
}

// RadiationLow returns the low end of the radiation habitability range.
func (r *Race) RadiationLow() int {
	return r.RadiationCenter - r.RadiationWidth
}

// RadiationHigh returns the high end of the radiation habitability range.
func (r *Race) RadiationHigh() int {
	return r.RadiationCenter + r.RadiationWidth
}

// HasLRT returns true if the race has the specified Lesser Race Trait.
func (r *Race) HasLRT(lrtBitmask uint16) bool {
	return (r.LRT & lrtBitmask) != 0
}

// NumImmunities returns the count of immune habitability dimensions.
func (r *Race) NumImmunities() int {
	count := 0
	if r.GravityImmune {
		count++
	}
	if r.TemperatureImmune {
		count++
	}
	if r.RadiationImmune {
		count++
	}
	return count
}
