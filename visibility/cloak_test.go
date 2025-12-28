package visibility

import (
	"math"
	"testing"
)

func TestCloakPerKTToPercent(t *testing.T) {
	tests := []struct {
		name       string
		cloakPerKT float64
		want       float64
		tolerance  float64
	}{
		{"zero cloak", 0, 0, 0.001},
		{"50 cloak/kT", 50, 0.25, 0.001},
		{"100 cloak/kT", 100, 0.5, 0.001},
		{"200 cloak/kT", 200, 0.625, 0.001},
		{"300 cloak/kT", 300, 0.75, 0.001},
		{"450 cloak/kT", 450, 0.8125, 0.001},
		{"600 cloak/kT", 600, 0.875, 0.001},
		{"800 cloak/kT", 800, 0.90625, 0.001},
		{"1000 cloak/kT", 1000, 0.9375, 0.001},
		{"1500 cloak/kT", 1500, 0.96875, 0.001},
		{"2000 cloak/kT", 2000, 0.98, 0.001}, // Capped at 98%
		{"5000 cloak/kT", 5000, 0.98, 0.001}, // Capped at 98%
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CloakPerKTToPercent(tt.cloakPerKT)
			if math.Abs(got-tt.want) > tt.tolerance {
				t.Errorf("CloakPerKTToPercent(%v) = %v, want %v", tt.cloakPerKT, got, tt.want)
			}
		})
	}
}

func TestTachyonReduction(t *testing.T) {
	tests := []struct {
		name        string
		numTachyons int
		want        float64
		tolerance   float64
	}{
		{"no tachyons", 0, 1.0, 0.001},
		{"1 tachyon", 1, 0.95, 0.001},
		{"4 tachyons", 4, 0.9025, 0.001}, // 0.95^sqrt(4) = 0.95^2
		{"9 tachyons", 9, 0.857375, 0.001}, // 0.95^sqrt(9) = 0.95^3
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TachyonReduction(tt.numTachyons)
			if math.Abs(got-tt.want) > tt.tolerance {
				t.Errorf("TachyonReduction(%v) = %v, want %v", tt.numTachyons, got, tt.want)
			}
		})
	}
}

func TestEffectiveCloaking(t *testing.T) {
	tests := []struct {
		name        string
		baseCloak   float64
		numTachyons int
		want        float64
		tolerance   float64
	}{
		{"no tachyons", 0.875, 0, 0.875, 0.001},
		{"1 tachyon", 0.875, 1, 0.83125, 0.001}, // 0.875 × 0.95
		{"4 tachyons", 0.875, 4, 0.7896875, 0.001}, // 0.875 × 0.9025
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EffectiveCloaking(tt.baseCloak, tt.numTachyons)
			if math.Abs(got-tt.want) > tt.tolerance {
				t.Errorf("EffectiveCloaking(%v, %v) = %v, want %v", tt.baseCloak, tt.numTachyons, got, tt.want)
			}
		})
	}
}

func TestEffectiveScannerRange(t *testing.T) {
	tests := []struct {
		name         string
		scannerRange int
		cloakPercent float64
		want         float64
		tolerance    float64
	}{
		{"no cloak", 100, 0, 100, 0.001},
		{"50% cloak", 100, 0.5, 50, 0.001},
		{"87.5% cloak", 100, 0.875, 12.5, 0.001},
		{"98% cloak", 100, 0.98, 2, 0.001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EffectiveScannerRange(tt.scannerRange, tt.cloakPercent)
			if math.Abs(got-tt.want) > tt.tolerance {
				t.Errorf("EffectiveScannerRange(%v, %v) = %v, want %v", tt.scannerRange, tt.cloakPercent, got, tt.want)
			}
		})
	}
}

func TestDistance(t *testing.T) {
	tests := []struct {
		name      string
		x1, y1    int
		x2, y2    int
		want      float64
		tolerance float64
	}{
		{"same point", 0, 0, 0, 0, 0, 0.001},
		{"horizontal", 0, 0, 10, 0, 10, 0.001},
		{"vertical", 0, 0, 0, 10, 10, 0.001},
		{"diagonal 3-4-5", 0, 0, 3, 4, 5, 0.001},
		{"large distance", 100, 100, 200, 200, 141.42, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Distance(tt.x1, tt.y1, tt.x2, tt.y2)
			if math.Abs(got-tt.want) > tt.tolerance {
				t.Errorf("Distance(%v, %v, %v, %v) = %v, want %v", tt.x1, tt.y1, tt.x2, tt.y2, got, tt.want)
			}
		})
	}
}

// TestCloakPerKTCurve verifies the curve is continuous and monotonically increasing.
func TestCloakPerKTCurve(t *testing.T) {
	prev := 0.0
	for cloakPerKT := 0.0; cloakPerKT <= 2000; cloakPerKT += 10 {
		curr := CloakPerKTToPercent(cloakPerKT)
		if curr < prev {
			t.Errorf("Curve not monotonic at cloak/kT=%v: prev=%v, curr=%v", cloakPerKT, prev, curr)
		}
		if curr < 0 || curr > 0.98 {
			t.Errorf("Curve out of bounds at cloak/kT=%v: %v", cloakPerKT, curr)
		}
		prev = curr
	}
}
