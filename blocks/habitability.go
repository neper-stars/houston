package blocks

import (
	"fmt"
)

// Habitability display conversion functions
//
// Stars! uses internal values 0-100 for habitability parameters.
// These functions convert between internal values and display values.
//
// Gravity: piecewise linear (Stars! uses a lookup table)
//   - Internal 0 = 0.12g, Internal 100 = 8.00g
//   - Values derived from Stars! Race Wizard calibration
//
// Temperature: linear scale
//   - Internal 0 = -200°C, Internal 100 = 200°C
//   - Formula: temp = -200 + val * 4
//
// Radiation: linear scale (direct mapping)
//   - Internal 0 = 0mR, Internal 100 = 100mR
//   - Formula: rad = val

const (
	// Gravity constants
	GravityMin = 0.12 // Minimum gravity in g (internal 0)
	GravityMax = 8.00 // Maximum gravity in g (internal 100)

	// Temperature constants
	TemperatureMin = -200 // Minimum temperature in °C (internal 0)
	TemperatureMax = 200  // Maximum temperature in °C (internal 100)

	// Radiation constants
	RadiationMin = 0   // Minimum radiation in mR (internal 0)
	RadiationMax = 100 // Maximum radiation in mR (internal 100)
)

// gravityTable is the complete 101-entry lookup table for gravity display values.
// Derived from Stars! Race Wizard by iterating through all positions.
// Index is the internal value (0-100), value is the display gravity in g.
var gravityTable = [101]float64{
	0.12, // 0
	0.12, // 1
	0.13, // 2
	0.13, // 3
	0.14, // 4
	0.14, // 5
	0.15, // 6
	0.15, // 7
	0.16, // 8
	0.17, // 9
	0.17, // 10
	0.18, // 11
	0.19, // 12
	0.20, // 13
	0.21, // 14
	0.22, // 15
	0.24, // 16
	0.25, // 17
	0.27, // 18
	0.29, // 19
	0.31, // 20
	0.33, // 21
	0.36, // 22
	0.40, // 23
	0.44, // 24
	0.50, // 25
	0.51, // 26
	0.52, // 27
	0.53, // 28
	0.54, // 29
	0.55, // 30
	0.56, // 31
	0.58, // 32
	0.59, // 33
	0.60, // 34
	0.62, // 35
	0.64, // 36
	0.65, // 37
	0.67, // 38
	0.69, // 39
	0.71, // 40
	0.73, // 41
	0.75, // 42
	0.78, // 43
	0.80, // 44
	0.83, // 45
	0.86, // 46
	0.89, // 47
	0.92, // 48
	0.96, // 49
	1.00, // 50
	1.04, // 51
	1.08, // 52
	1.12, // 53
	1.16, // 54
	1.20, // 55
	1.24, // 56
	1.28, // 57
	1.32, // 58
	1.36, // 59
	1.40, // 60
	1.44, // 61
	1.48, // 62
	1.52, // 63
	1.56, // 64
	1.60, // 65
	1.64, // 66
	1.68, // 67
	1.72, // 68
	1.76, // 69
	1.80, // 70
	1.84, // 71
	1.88, // 72
	1.92, // 73
	1.96, // 74
	2.00, // 75
	2.24, // 76
	2.48, // 77
	2.72, // 78
	2.96, // 79
	3.20, // 80
	3.44, // 81
	3.68, // 82
	3.92, // 83
	4.16, // 84
	4.40, // 85
	4.64, // 86
	4.88, // 87
	5.12, // 88
	5.36, // 89
	5.60, // 90
	5.84, // 91
	6.08, // 92
	6.32, // 93
	6.56, // 94
	6.80, // 95
	7.04, // 96
	7.28, // 97
	7.52, // 98
	7.76, // 99
	8.00, // 100
}

// GravityToDisplay converts an internal gravity value (0-100) to display value in g.
// Returns the gravity in g units (e.g., 0.12 to 8.00).
// Uses direct lookup from the Stars! gravity table.
func GravityToDisplay(internal int) float64 {
	if internal <= 0 {
		return gravityTable[0]
	}
	if internal >= 100 {
		return gravityTable[100]
	}
	return gravityTable[internal]
}

// GravityFromDisplay converts a display gravity value in g to internal value (0-100).
// Input should be in range 0.12 to 8.00.
// Returns the internal value that produces the closest display value.
func GravityFromDisplay(g float64) int {
	if g <= gravityTable[0] {
		return 0
	}
	if g >= gravityTable[100] {
		return 100
	}

	// Find the closest matching entry
	closest := 0
	minDiff := g - gravityTable[0]
	if minDiff < 0 {
		minDiff = -minDiff
	}

	for i := 1; i <= 100; i++ {
		diff := g - gravityTable[i]
		if diff < 0 {
			diff = -diff
		}
		if diff < minDiff {
			minDiff = diff
			closest = i
		}
	}
	return closest
}

// GravityDisplayValues returns all 101 display gravity values from the lookup table.
// Useful for building UI displays that match Stars! exactly.
func GravityDisplayValues() []float64 {
	values := make([]float64, 101)
	copy(values, gravityTable[:])
	return values
}

// GravityDisplayString returns a formatted gravity string (e.g., "1.00g").
func GravityDisplayString(internal int) string {
	return fmt.Sprintf("%.2fg", GravityToDisplay(internal))
}

// TemperatureToDisplay converts an internal temperature value (0-100) to display value in °C.
// Returns the temperature in Celsius (e.g., -200 to 200).
func TemperatureToDisplay(internal int) int {
	// Linear: temp = -200 + val * 4
	return TemperatureMin + internal*4
}

// TemperatureFromDisplay converts a display temperature value in °C to internal value (0-100).
// Input should be in range -200 to 200.
func TemperatureFromDisplay(temp int) int {
	if temp <= TemperatureMin {
		return 0
	}
	if temp >= TemperatureMax {
		return 100
	}
	// val = (temp + 200) / 4
	return (temp - TemperatureMin) / 4
}

// TemperatureDisplayString returns a formatted temperature string (e.g., "-140°C").
func TemperatureDisplayString(internal int) string {
	return fmt.Sprintf("%d°C", TemperatureToDisplay(internal))
}

// RadiationToDisplay converts an internal radiation value (0-100) to display value in mR.
// This is a direct mapping (internal value = display value).
func RadiationToDisplay(internal int) int {
	return internal
}

// RadiationFromDisplay converts a display radiation value in mR to internal value (0-100).
// This is a direct mapping (display value = internal value).
func RadiationFromDisplay(mR int) int {
	if mR < RadiationMin {
		return RadiationMin
	}
	if mR > RadiationMax {
		return RadiationMax
	}
	return mR
}

// RadiationDisplayString returns a formatted radiation string (e.g., "50mR").
func RadiationDisplayString(internal int) string {
	return fmt.Sprintf("%dmR", RadiationToDisplay(internal))
}

// HabitabilityDisplay holds display-friendly habitability values.
type HabitabilityDisplay struct {
	GravityLow    float64 // Low gravity in g
	GravityCenter float64 // Center gravity in g
	GravityHigh   float64 // High gravity in g
	GravityImmune bool

	TemperatureLow    int // Low temperature in °C
	TemperatureCenter int // Center temperature in °C
	TemperatureHigh   int // High temperature in °C
	TemperatureImmune bool

	RadiationLow    int // Low radiation in mR
	RadiationCenter int // Center radiation in mR
	RadiationHigh   int // High radiation in mR
	RadiationImmune bool
}

// ToDisplay converts internal Habitability values to display values.
func (h *Habitability) ToDisplay() HabitabilityDisplay {
	return HabitabilityDisplay{
		GravityLow:    GravityToDisplay(h.GravityLow),
		GravityCenter: GravityToDisplay(h.GravityCenter),
		GravityHigh:   GravityToDisplay(h.GravityHigh),
		GravityImmune: h.IsGravityImmune(),

		TemperatureLow:    TemperatureToDisplay(h.TemperatureLow),
		TemperatureCenter: TemperatureToDisplay(h.TemperatureCenter),
		TemperatureHigh:   TemperatureToDisplay(h.TemperatureHigh),
		TemperatureImmune: h.IsTemperatureImmune(),

		RadiationLow:    RadiationToDisplay(h.RadiationLow),
		RadiationCenter: RadiationToDisplay(h.RadiationCenter),
		RadiationHigh:   RadiationToDisplay(h.RadiationHigh),
		RadiationImmune: h.IsRadiationImmune(),
	}
}

// GravityRangeString returns a formatted gravity range string (e.g., "0.22g to 4.40g" or "Immune").
func (h *Habitability) GravityRangeString() string {
	if h.IsGravityImmune() {
		return "Immune"
	}
	return fmt.Sprintf("%.2fg to %.2fg", GravityToDisplay(h.GravityLow), GravityToDisplay(h.GravityHigh))
}

// TemperatureRangeString returns a formatted temperature range string (e.g., "-140°C to 140°C" or "Immune").
func (h *Habitability) TemperatureRangeString() string {
	if h.IsTemperatureImmune() {
		return "Immune"
	}
	return fmt.Sprintf("%d°C to %d°C", TemperatureToDisplay(h.TemperatureLow), TemperatureToDisplay(h.TemperatureHigh))
}

// RadiationRangeString returns a formatted radiation range string (e.g., "15mR to 85mR" or "Immune").
func (h *Habitability) RadiationRangeString() string {
	if h.IsRadiationImmune() {
		return "Immune"
	}
	return fmt.Sprintf("%dmR to %dmR", RadiationToDisplay(h.RadiationLow), RadiationToDisplay(h.RadiationHigh))
}

// String returns a summary of all habitability parameters.
func (h *Habitability) String() string {
	return fmt.Sprintf("Gravity: %s, Temperature: %s, Radiation: %s",
		h.GravityRangeString(), h.TemperatureRangeString(), h.RadiationRangeString())
}
