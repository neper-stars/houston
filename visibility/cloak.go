// Package visibility provides fleet detection and cloaking calculations.
package visibility

import (
	"math"
)

// CloakPerKTToPercent converts cloak units per kiloton to a cloaking percentage.
// This implements the piecewise linear curve from Stars!:
//   - 0-100 cloak/kT: 0% to 50%
//   - 100-300 cloak/kT: 50% to 75%
//   - 300-600 cloak/kT: 75% to 87.5%
//   - 600-1000 cloak/kT: 87.5% to 93.75%
//   - 1000+ cloak/kT: 93.75% to 98% (max)
func CloakPerKTToPercent(cloakPerKT float64) float64 {
	var percent float64

	switch {
	case cloakPerKT <= 0:
		percent = 0
	case cloakPerKT <= 100:
		percent = cloakPerKT / 200.0
	case cloakPerKT <= 300:
		percent = 0.5 + 0.25*(cloakPerKT-100)/200.0
	case cloakPerKT <= 600:
		percent = 0.75 + 0.125*(cloakPerKT-300)/300.0
	case cloakPerKT <= 1000:
		percent = 0.875 + 0.0625*(cloakPerKT-600)/400.0
	default:
		percent = 0.9375 + 0.03125*(cloakPerKT-1000)/500.0
	}

	// Cap at 98%
	if percent > 0.98 {
		percent = 0.98
	}

	return percent
}

// TachyonReduction calculates the cloaking reduction from Tachyon Detectors.
// The formula is: effective_cloak = base_cloak × 0.95^sqrt(num_tachyons)
// Returns the multiplier to apply to base cloaking (0.0 to 1.0).
func TachyonReduction(numTachyons int) float64 {
	if numTachyons <= 0 {
		return 1.0
	}
	return math.Pow(0.95, math.Sqrt(float64(numTachyons)))
}

// EffectiveCloaking calculates the effective cloaking percentage after Tachyon reduction.
func EffectiveCloaking(baseCloakPercent float64, numTachyons int) float64 {
	return baseCloakPercent * TachyonReduction(numTachyons)
}

// EffectiveScannerRange calculates the effective scanner range against a cloaked target.
// Formula: effective_range = scanner_range × (1 - cloak_percent)
func EffectiveScannerRange(scannerRange int, cloakPercent float64) float64 {
	return float64(scannerRange) * (1.0 - cloakPercent)
}

// Distance calculates the Euclidean distance between two points.
func Distance(x1, y1, x2, y2 int) float64 {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	return math.Sqrt(dx*dx + dy*dy)
}
