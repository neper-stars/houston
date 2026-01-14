// Package reporter provides functionality to generate analysis reports for Stars! games.
//
// It creates LibreOffice ODS spreadsheets with multi-turn tracking for both
// the player's own assets and opponent intelligence.
package reporter

import "github.com/neper-stars/houston/store"

// ShipCategoryCount holds ship counts by combat category.
type ShipCategoryCount struct {
	Unarmed int // Ships with combat power = 0
	Escort  int // Ships with 0 < power < 2000
	Capital int // Ships with power >= 2000
	Total   int // Sum of all ships
}

// PlanetMineralData holds mineral data for a single planet.
type PlanetMineralData struct {
	Number     int
	Name       string
	X, Y       int
	Ironium    int64
	Boranium   int64
	Germanium  int64
	Population int64

	// En route cargo (from fleets heading to this planet)
	EnRouteIronium   int64
	EnRouteBoranium  int64
	EnRouteGermanium int64
}

// MineralTotals holds aggregated mineral totals.
type MineralTotals struct {
	Ironium   int64
	Boranium  int64
	Germanium int64
}

// PlayerSnapshot holds a snapshot of a player's state at a given turn.
type PlayerSnapshot struct {
	PlayerNumber int
	Turn         int
	Year         int

	// Totals
	TotalPopulation int64
	Minerals        MineralTotals
	PlanetCount     int
	FleetCount      int

	// Ship counts by category
	Ships ShipCategoryCount

	// Score estimation
	EstimatedScore int

	// Tech levels (if known)
	TechEnergy       int
	TechWeapons      int
	TechPropulsion   int
	TechConstruction int
	TechElectronics  int
	TechBiotech      int
}

// DesignInfo holds information about a ship design.
type DesignInfo struct {
	Owner       int
	Slot        int
	Name        string
	HullName    string
	CombatPower int
	FirstSeen   int // Turn when first seen
	IsStarbase  bool

	// Inferred tech levels from components
	InferredTechEnergy       int
	InferredTechWeapons      int
	InferredTechPropulsion   int
	InferredTechConstruction int
	InferredTechElectronics  int
	InferredTechBiotech      int
}

// MineralNeed represents a planet that needs minerals.
type MineralNeed struct {
	Planet  *store.PlanetEntity
	Mineral string // "Ironium", "Boranium", or "Germanium"
	Current int64
	Needed  int64 // Threshold - Current (positive if below threshold)
}

// MineralSource represents a planet with surplus minerals.
type MineralSource struct {
	Planet  *store.PlanetEntity
	Mineral string // "Ironium", "Boranium", or "Germanium"
	Surplus int64  // Current - Threshold (positive if above threshold)
}

// ShuffleRecommendation suggests moving minerals between planets.
type ShuffleRecommendation struct {
	From     *store.PlanetEntity
	To       *store.PlanetEntity
	Mineral  string
	Amount   int64
	Distance float64
}

// ReportOptions controls report generation.
type ReportOptions struct {
	PlayerNumber       int   // Which player to report as (0-indexed)
	MineralThreshold   int64 // Threshold for mineral shuffle analysis (default: 500)
	IncludeAllPlanets  bool  // Include planets with 0 minerals in My Minerals sheet
	IncludeEmptyFleets bool  // Include fleets with 0 cargo
}

// DefaultOptions returns default report options.
func DefaultOptions() *ReportOptions {
	return &ReportOptions{
		PlayerNumber:       0,
		MineralThreshold:   500,
		IncludeAllPlanets:  false,
		IncludeEmptyFleets: false,
	}
}

// Sheet names in the ODS template.
const (
	SheetSummary         = "Summary"
	SheetMyMinerals      = "My Minerals"
	SheetMyMineralHist   = "My Minerals History"
	SheetMineralShuffle  = "Mineral Shuffle"
	SheetOpponentPop     = "Opponent Population"
	SheetOpponentPopHist = "Opponent Pop History"
	SheetOpponentShips   = "Opponent Ships"
	SheetOpponentFleets  = "Opponent Fleets"
	SheetNewDesigns      = "New Designs"
	SheetScoreEstimates  = "Score Estimates"
)
