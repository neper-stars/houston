// Package race provides race building and point calculation for Stars!
//
// TODO: Implement predefined race templates from step 1 of the race wizard:
//   - Humanoid (default)
//   - Rabbitoid
//   - Insectoid
//   - Nucleotid
//   - Silicanoid
//   - Antetheral
//   - Random (randomize all settings)
// User will provide the specific values for each predefined race.
package race

// Primary Race Trait (PRT) constants for use with Builder.PRT()
const (
	PRTHyperExpansion       = 0  // HE - Hyper Expansion
	PRTSuperStealth         = 1  // SS - Super Stealth
	PRTWarMonger            = 2  // WM - War Monger
	PRTClaimAdjuster        = 3  // CA - Claim Adjuster
	PRTInnerStrength        = 4  // IS - Inner Strength
	PRTSpaceDemolition      = 5  // SD - Space Demolition
	PRTPacketPhysics        = 6  // PP - Packet Physics
	PRTInterstellarTraveler = 7  // IT - Interstellar Traveler
	PRTAlternateReality     = 8  // AR - Alternate Reality
	PRTJackOfAllTrades      = 9  // JOAT - Jack of All Trades
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

// Default returns a new Race with Humanoid defaults.
// This is the standard starting race with balanced settings.
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
