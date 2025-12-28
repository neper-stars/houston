package visibility

import (
	"math"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/data"
	"github.com/neper-stars/houston/store"
)

// DetectionResult provides detailed visibility information between two fleets.
type DetectionResult struct {
	// CanSee indicates whether the observer can detect the target.
	CanSee bool

	// Distance is the distance in light-years between the fleets.
	Distance float64

	// BaseCloakPercent is the target's cloaking before Tachyon reduction.
	BaseCloakPercent float64

	// EffectiveCloakPercent is the target's cloaking after Tachyon reduction.
	EffectiveCloakPercent float64

	// NormalScannerRange is the observer's normal scanner range.
	NormalScannerRange int

	// PenetratingScannerRange is the observer's penetrating scanner range.
	PenetratingScannerRange int

	// EffectiveNormalRange is the effective normal range after cloaking reduction.
	EffectiveNormalRange float64

	// EffectivePenRange is the effective penetrating range after cloaking reduction.
	EffectivePenRange float64

	// TachyonCount is the number of Tachyon Detectors on the observer.
	TachyonCount int

	// DetectedBy indicates which scanner type made the detection ("normal", "penetrating", or "none").
	DetectedBy string
}

// CanDetect returns true if the observer fleet can detect the target fleet.
// This is a convenience function that wraps GetDetectionDetails.
func CanDetect(observer, target *store.FleetEntity, gs *store.GameStore) bool {
	result := GetDetectionDetails(observer, target, gs)
	return result.CanSee
}

// GetDetectionDetails calculates detailed visibility information between two fleets.
func GetDetectionDetails(observer, target *store.FleetEntity, gs *store.GameStore) DetectionResult {
	result := DetectionResult{
		DetectedBy: "none",
	}

	// Calculate distance
	result.Distance = Distance(observer.X, observer.Y, target.X, target.Y)

	// Get observer's scanner ranges
	result.NormalScannerRange, result.PenetratingScannerRange = observer.GetFleetScannerRanges(gs)

	// Get observer's Tachyon detector count
	result.TachyonCount = observer.GetFleetTachyonCount(gs)

	// Calculate target's cloaking
	result.BaseCloakPercent = FleetCloaking(target, gs)
	result.EffectiveCloakPercent = EffectiveCloaking(result.BaseCloakPercent, result.TachyonCount)

	// Calculate effective scanner ranges
	result.EffectiveNormalRange = EffectiveScannerRange(result.NormalScannerRange, result.EffectiveCloakPercent)
	result.EffectivePenRange = EffectiveScannerRange(result.PenetratingScannerRange, result.EffectiveCloakPercent)

	// Check detection
	if result.Distance <= result.EffectiveNormalRange {
		result.CanSee = true
		result.DetectedBy = "normal"
	} else if result.Distance <= result.EffectivePenRange {
		result.CanSee = true
		result.DetectedBy = "penetrating"
	}

	return result
}

// SSIntrinsicCloakPercent is the intrinsic cloaking percentage for Super Stealth PRT ships.
// All SS ships have a minimum 75% cloaking built-in.
const SSIntrinsicCloakPercent = 0.75

// FleetCloaking calculates a fleet's cloaking percentage.
// This implements the Stars! cloaking formula:
//  1. Calculate total cloak units (sum of ship cloak units × ship count)
//  2. Divide by total fleet mass to get cloak/kT
//  3. Apply piecewise linear curve to get percentage (max 98%)
//
// Special handling for Super Stealth (SS) PRT:
//   - All ships have intrinsic 75% cloaking (not 75 units)
//   - Additional cloaking devices can increase this further
//   - Cargo does NOT count toward mass for cloaking calculations
func FleetCloaking(fleet *store.FleetEntity, gs *store.GameStore) float64 {
	// Check if owner is Super Stealth PRT
	isSS := false
	if player, ok := gs.Player(fleet.Owner); ok {
		isSS = player.PRT == blocks.PRTSuperStealth
	}

	// Get total cloak units from equipped devices
	totalCloakUnits := fleet.GetFleetCloakUnits(gs)

	// Calculate equipment-based cloaking
	var equipmentCloak float64
	if totalCloakUnits > 0 {
		var fleetMass int64
		if isSS {
			// SS PRT: cargo doesn't count toward cloaking mass
			fleetMass = fleetMassWithoutCargo(fleet, gs)
		} else {
			fleetMass = fleet.GetTotalMass(gs)
		}

		if fleetMass > 0 {
			cloakPerKT := float64(totalCloakUnits) / float64(fleetMass)
			equipmentCloak = CloakPerKTToPercent(cloakPerKT)
		}
	}

	// For SS PRT, the minimum cloaking is 75%
	// Additional equipment cloaking stacks: combined = 1 - (1 - base) * (1 - equip)
	if isSS {
		if equipmentCloak > 0 {
			// Stack SS intrinsic with equipment cloaking
			return 1.0 - (1.0-SSIntrinsicCloakPercent)*(1.0-equipmentCloak)
		}
		return SSIntrinsicCloakPercent
	}

	return equipmentCloak
}

// fleetMassWithoutCargo calculates fleet mass excluding cargo.
// Used for SS PRT where cargo doesn't affect cloaking.
func fleetMassWithoutCargo(fleet *store.FleetEntity, gs *store.GameStore) int64 {
	var total int64
	designs := fleet.GetDesigns(gs)
	for _, info := range designs {
		if info.Design != nil {
			hull := info.Design.Hull()
			if hull != nil {
				total += int64(hull.Mass) * int64(info.Count)
			}
		}
	}
	return total
}

// FleetScannerRange returns the best normal and penetrating scanner ranges for a fleet.
func FleetScannerRange(fleet *store.FleetEntity, gs *store.GameStore) (normal, penetrating int) {
	return fleet.GetFleetScannerRanges(gs)
}

// TachyonCount returns the total number of Tachyon Detectors in a fleet.
func TachyonCount(fleet *store.FleetEntity, gs *store.GameStore) int {
	return fleet.GetFleetTachyonCount(gs)
}

// CanFleetSeePosition returns true if the fleet can see a given position.
// This doesn't account for cloaking (used for seeing planets, minefields, etc.).
func CanFleetSeePosition(fleet *store.FleetEntity, x, y int, gs *store.GameStore) bool {
	normalRange, _ := fleet.GetFleetScannerRanges(gs)
	distance := Distance(fleet.X, fleet.Y, x, y)
	return distance <= float64(normalRange)
}

// CanFleetSeePenetrating returns true if the fleet can see a given position
// with penetrating scanners (for seeing through planets).
func CanFleetSeePenetrating(fleet *store.FleetEntity, x, y int, gs *store.GameStore) bool {
	_, penRange := fleet.GetFleetScannerRanges(gs)
	distance := Distance(fleet.X, fleet.Y, x, y)
	return distance <= float64(penRange)
}

// ============================================================================
// Planet Observer Functions
// ============================================================================

// PlanetScannerRanges returns the scanner ranges for a planet.
// This considers:
// - Planetary scanner (based on owner's tech level)
// - Starbase scanner (if starbase exists)
// - AR PRT intrinsic scanner (sqrt(population/10) for AR starbases)
//
// Returns (normal, penetrating) ranges in light-years.
func PlanetScannerRanges(planet *store.PlanetEntity, gs *store.GameStore) (int, int) {
	if !planet.IsOwned() {
		return 0, 0
	}

	bestNormal := 0
	bestPen := 0

	// 1. Planetary scanner (if planet has scanner)
	if planet.HasScanner {
		player, ok := gs.Player(planet.Owner)
		if ok {
			scanner, _ := data.GetBestPlanetaryScanner(player.Tech)
			if scanner != nil {
				if scanner.NormalRange > bestNormal {
					bestNormal = scanner.NormalRange
				}
				if scanner.PenetratingRange > bestPen {
					bestPen = scanner.PenetratingRange
				}
			}
		}
	}

	// 2. Starbase scanner (if starbase exists)
	if planet.HasStarbase {
		starbase, ok := gs.StarbaseDesign(planet.Owner, planet.StarbaseDesign)
		if ok {
			sbNormal, sbPen := starbase.GetScannerRanges()
			if sbNormal > bestNormal {
				bestNormal = sbNormal
			}
			if sbPen > bestPen {
				bestPen = sbPen
			}
		}

		// 3. AR PRT intrinsic scanner: range = sqrt(population/10)
		// Note: For AR, population location (starbase vs planet) needs verification
		// with real data. For now, we use planet population.
		player, ok := gs.Player(planet.Owner)
		if ok && player.PRT == blocks.PRTAlternateReality {
			arRange := ARIntrinsicScannerRange(planet.Population)
			if arRange > bestNormal {
				bestNormal = arRange
			}
		}
	}

	return bestNormal, bestPen
}

// ARIntrinsicScannerRange calculates the intrinsic scanner range for
// Alternate Reality (AR) PRT starbases.
// Formula: range = sqrt(population / 10)
func ARIntrinsicScannerRange(population int64) int {
	if population <= 0 {
		return 0
	}
	return int(math.Sqrt(float64(population) / 10.0))
}

// CanPlanetDetectFleet returns true if a planet can detect a fleet.
func CanPlanetDetectFleet(planet *store.PlanetEntity, fleet *store.FleetEntity, gs *store.GameStore) bool {
	result := GetPlanetDetectionDetails(planet, fleet, gs)
	return result.CanSee
}

// GetPlanetDetectionDetails calculates visibility from a planet to a fleet.
func GetPlanetDetectionDetails(planet *store.PlanetEntity, target *store.FleetEntity, gs *store.GameStore) DetectionResult {
	result := DetectionResult{
		DetectedBy: "none",
	}

	// Calculate distance
	result.Distance = Distance(planet.X, planet.Y, target.X, target.Y)

	// Get planet's scanner ranges
	result.NormalScannerRange, result.PenetratingScannerRange = PlanetScannerRanges(planet, gs)

	// Planets don't have Tachyon detectors (only ships do)
	// TODO: Starbases might have Tachyon detectors - need to verify with real data
	result.TachyonCount = 0

	// If starbase exists, check for Tachyon detectors
	if planet.HasStarbase {
		starbase, ok := gs.StarbaseDesign(planet.Owner, planet.StarbaseDesign)
		if ok {
			result.TachyonCount = starbase.GetTachyonCount()
		}
	}

	// Calculate target's cloaking
	result.BaseCloakPercent = FleetCloaking(target, gs)
	result.EffectiveCloakPercent = EffectiveCloaking(result.BaseCloakPercent, result.TachyonCount)

	// Calculate effective scanner ranges
	result.EffectiveNormalRange = EffectiveScannerRange(result.NormalScannerRange, result.EffectiveCloakPercent)
	result.EffectivePenRange = EffectiveScannerRange(result.PenetratingScannerRange, result.EffectiveCloakPercent)

	// Check detection
	if result.Distance <= result.EffectiveNormalRange {
		result.CanSee = true
		result.DetectedBy = "normal"
	} else if result.Distance <= result.EffectivePenRange {
		result.CanSee = true
		result.DetectedBy = "penetrating"
	}

	return result
}

// CanPlanetSeePosition returns true if a planet can see a given position.
func CanPlanetSeePosition(planet *store.PlanetEntity, x, y int, gs *store.GameStore) bool {
	normalRange, _ := PlanetScannerRanges(planet, gs)
	distance := Distance(planet.X, planet.Y, x, y)
	return distance <= float64(normalRange)
}
