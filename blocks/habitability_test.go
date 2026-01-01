package blocks

import (
	"testing"
)

func TestGravityConversions(t *testing.T) {
	// Test against complete Stars! calibration data (101-entry lookup table)
	// Selected key points from the full table
	calibrationData := []struct {
		internal int
		display  float64
	}{
		{0, 0.12},
		{1, 0.12},
		{2, 0.13},
		{5, 0.14},
		{10, 0.17},
		{15, 0.22},
		{20, 0.31},
		{25, 0.50},
		{30, 0.55},
		{35, 0.62},
		{40, 0.71},
		{45, 0.83},
		{50, 1.00},
		{55, 1.20},
		{60, 1.40},
		{65, 1.60},
		{70, 1.80},
		{75, 2.00},
		{80, 3.20},
		{85, 4.40},
		{90, 5.60},
		{95, 6.80},
		{100, 8.00},
	}

	for _, tc := range calibrationData {
		got := GravityToDisplay(tc.internal)
		if got != tc.display {
			t.Errorf("GravityToDisplay(%d) = %.2f, want %.2f", tc.internal, got, tc.display)
		}
	}
}

func TestGravityRoundTrip(t *testing.T) {
	// Test that converting to display and back gives a value that produces the same display
	// Note: Multiple internal values can map to the same display value (e.g., 0 and 1 both → 0.12g)
	// So we test that the round-trip produces the same display value, not necessarily the same internal value
	for i := 0; i <= 100; i++ {
		display := GravityToDisplay(i)
		back := GravityFromDisplay(display)
		backDisplay := GravityToDisplay(back)
		if display != backDisplay {
			t.Errorf("GravityRoundTrip(%d): display=%.2f, back=%d, backDisplay=%.2f", i, display, back, backDisplay)
		}
	}
}

func TestTemperatureConversions(t *testing.T) {
	testCases := []struct {
		internal int
		display  int
	}{
		{0, -200},
		{50, 0},
		{100, 200},
		{15, -140}, // Humanoid low
		{85, 140},  // Humanoid high
	}

	for _, tc := range testCases {
		got := TemperatureToDisplay(tc.internal)
		if got != tc.display {
			t.Errorf("TemperatureToDisplay(%d) = %d, want %d", tc.internal, got, tc.display)
		}
	}
}

func TestTemperatureRoundTrip(t *testing.T) {
	// Test that converting to display and back gives the same value
	for i := 0; i <= 100; i++ {
		display := TemperatureToDisplay(i)
		back := TemperatureFromDisplay(display)
		if back != i {
			t.Errorf("TemperatureRoundTrip(%d): display=%d, back=%d", i, display, back)
		}
	}
}

func TestRadiationConversions(t *testing.T) {
	testCases := []struct {
		internal int
		display  int
	}{
		{0, 0},
		{50, 50},
		{100, 100},
		{15, 15}, // Humanoid low
		{85, 85}, // Humanoid high
	}

	for _, tc := range testCases {
		got := RadiationToDisplay(tc.internal)
		if got != tc.display {
			t.Errorf("RadiationToDisplay(%d) = %d, want %d", tc.internal, got, tc.display)
		}
	}
}

func TestHabitabilityToDisplay(t *testing.T) {
	// Test with Humanoid-like values (center 50, width 35)
	h := Habitability{
		GravityLow:        15,
		GravityCenter:     50,
		GravityHigh:       85,
		TemperatureLow:    15,
		TemperatureCenter: 50,
		TemperatureHigh:   85,
		RadiationLow:      15,
		RadiationCenter:   50,
		RadiationHigh:     85,
	}

	display := h.ToDisplay()

	// Check gravity - exact values from Stars! calibration
	// Internal 15 = 0.22g, Internal 85 = 4.40g
	if display.GravityLow != 0.22 {
		t.Errorf("GravityLow = %.2f, want 0.22", display.GravityLow)
	}
	if display.GravityHigh != 4.40 {
		t.Errorf("GravityHigh = %.2f, want 4.40", display.GravityHigh)
	}
	t.Logf("Gravity: %.2fg to %.2fg", display.GravityLow, display.GravityHigh)

	// Check temperature (linear)
	if display.TemperatureLow != -140 {
		t.Errorf("TemperatureLow = %d, want -140", display.TemperatureLow)
	}
	if display.TemperatureHigh != 140 {
		t.Errorf("TemperatureHigh = %d, want 140", display.TemperatureHigh)
	}

	// Check radiation (linear)
	if display.RadiationLow != 15 {
		t.Errorf("RadiationLow = %d, want 15", display.RadiationLow)
	}
	if display.RadiationHigh != 85 {
		t.Errorf("RadiationHigh = %d, want 85", display.RadiationHigh)
	}
}

func TestHabitabilityRangeStrings(t *testing.T) {
	h := Habitability{
		GravityLow:        15,
		GravityCenter:     50,
		GravityHigh:       85,
		TemperatureLow:    15,
		TemperatureCenter: 50,
		TemperatureHigh:   85,
		RadiationLow:      15,
		RadiationCenter:   50,
		RadiationHigh:     85,
	}

	gravStr := h.GravityRangeString()
	t.Logf("Gravity range: %s", gravStr)

	tempStr := h.TemperatureRangeString()
	if tempStr != "-140°C to 140°C" {
		t.Errorf("TemperatureRangeString = %q, want %q", tempStr, "-140°C to 140°C")
	}

	radStr := h.RadiationRangeString()
	if radStr != "15mR to 85mR" {
		t.Errorf("RadiationRangeString = %q, want %q", radStr, "15mR to 85mR")
	}
}

func TestHabitabilityImmune(t *testing.T) {
	h := Habitability{
		GravityCenter:     255, // Immune
		TemperatureCenter: 255, // Immune
		RadiationCenter:   255, // Immune
	}

	if h.GravityRangeString() != "Immune" {
		t.Errorf("GravityRangeString = %q, want %q", h.GravityRangeString(), "Immune")
	}
	if h.TemperatureRangeString() != "Immune" {
		t.Errorf("TemperatureRangeString = %q, want %q", h.TemperatureRangeString(), "Immune")
	}
	if h.RadiationRangeString() != "Immune" {
		t.Errorf("RadiationRangeString = %q, want %q", h.RadiationRangeString(), "Immune")
	}

	display := h.ToDisplay()
	if !display.GravityImmune {
		t.Error("Expected GravityImmune to be true")
	}
	if !display.TemperatureImmune {
		t.Error("Expected TemperatureImmune to be true")
	}
	if !display.RadiationImmune {
		t.Error("Expected RadiationImmune to be true")
	}
}
