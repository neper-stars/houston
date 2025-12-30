package race

import (
	"strings"
	"testing"
)

func TestValidateHabitabilityEdgeConstraints(t *testing.T) {
	// Helper to create a valid base race for testing
	validBase := func() *Race {
		r := Default()
		return r
	}

	tests := []struct {
		name        string
		modify      func(*Race)
		wantErr     bool
		errContains string
	}{
		// Valid edge cases - edges exactly at bounds
		{
			name: "gravity center at minimum valid (low edge = 0)",
			modify: func(r *Race) {
				r.GravityWidth = 15
				r.GravityCenter = 15 // low edge = 15-15 = 0, high edge = 15+15 = 30
			},
			wantErr: false,
		},
		{
			name: "gravity center at maximum valid (high edge = 100)",
			modify: func(r *Race) {
				r.GravityWidth = 15
				r.GravityCenter = 85 // low edge = 85-15 = 70, high edge = 85+15 = 100
			},
			wantErr: false,
		},
		{
			name: "gravity maximum width centered",
			modify: func(r *Race) {
				r.GravityWidth = 50
				r.GravityCenter = 50 // low edge = 0, high edge = 100
			},
			wantErr: false,
		},
		{
			name: "temperature center at minimum valid (low edge = 0)",
			modify: func(r *Race) {
				r.TemperatureWidth = 20
				r.TemperatureCenter = 20 // low edge = 0
			},
			wantErr: false,
		},
		{
			name: "temperature center at maximum valid (high edge = 100)",
			modify: func(r *Race) {
				r.TemperatureWidth = 20
				r.TemperatureCenter = 80 // high edge = 100
			},
			wantErr: false,
		},
		{
			name: "radiation center at minimum valid (low edge = 0)",
			modify: func(r *Race) {
				r.RadiationWidth = 25
				r.RadiationCenter = 25 // low edge = 0
			},
			wantErr: false,
		},
		{
			name: "radiation center at maximum valid (high edge = 100)",
			modify: func(r *Race) {
				r.RadiationWidth = 25
				r.RadiationCenter = 75 // high edge = 100
			},
			wantErr: false,
		},

		// Invalid cases - low edge below 0
		{
			name: "gravity low edge below 0",
			modify: func(r *Race) {
				r.GravityWidth = 30
				r.GravityCenter = 20 // low edge = 20-30 = -10
			},
			wantErr:     true,
			errContains: "gravity range low edge",
		},
		{
			name: "temperature low edge below 0",
			modify: func(r *Race) {
				r.TemperatureWidth = 40
				r.TemperatureCenter = 30 // low edge = 30-40 = -10
			},
			wantErr:     true,
			errContains: "temperature range low edge",
		},
		{
			name: "radiation low edge below 0",
			modify: func(r *Race) {
				r.RadiationWidth = 50
				r.RadiationCenter = 40 // low edge = 40-50 = -10
			},
			wantErr:     true,
			errContains: "radiation range low edge",
		},

		// Invalid cases - high edge above 100
		{
			name: "gravity high edge above 100",
			modify: func(r *Race) {
				r.GravityWidth = 30
				r.GravityCenter = 80 // high edge = 80+30 = 110
			},
			wantErr:     true,
			errContains: "gravity range high edge",
		},
		{
			name: "temperature high edge above 100",
			modify: func(r *Race) {
				r.TemperatureWidth = 40
				r.TemperatureCenter = 70 // high edge = 70+40 = 110
			},
			wantErr:     true,
			errContains: "temperature range high edge",
		},
		{
			name: "radiation high edge above 100",
			modify: func(r *Race) {
				r.RadiationWidth = 50
				r.RadiationCenter = 60 // high edge = 60+50 = 110
			},
			wantErr:     true,
			errContains: "radiation range high edge",
		},

		// Immune dimensions should skip edge validation
		{
			name: "gravity immune skips edge validation",
			modify: func(r *Race) {
				r.GravityImmune = true
				r.GravityWidth = 50
				r.GravityCenter = 10 // would be invalid if not immune
			},
			wantErr: false,
		},
		{
			name: "temperature immune skips edge validation",
			modify: func(r *Race) {
				r.TemperatureImmune = true
				r.TemperatureWidth = 50
				r.TemperatureCenter = 90 // would be invalid if not immune
			},
			wantErr: false,
		},
		{
			name: "radiation immune skips edge validation",
			modify: func(r *Race) {
				r.RadiationImmune = true
				r.RadiationWidth = 50
				r.RadiationCenter = 5 // would be invalid if not immune
			},
			wantErr: false,
		},

		// Multiple violations
		{
			name: "both edges invalid for gravity",
			modify: func(r *Race) {
				// This shouldn't happen in practice since width is capped at 50,
				// but test the validation logic
				r.GravityWidth = 50
				r.GravityCenter = 10 // low = -40, high = 60
			},
			wantErr:     true,
			errContains: "gravity range low edge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := validBase()
			tt.modify(r)

			errs := Validate(r)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Error("expected validation error, got none")
					return
				}
				// Check that at least one error contains the expected substring
				found := false
				for _, err := range errs {
					if strings.Contains(err.Message, tt.errContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, got %v", tt.errContains, errs)
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("expected no errors, got %v", errs)
				}
			}
		})
	}
}

func TestValidateHabitabilityEdgeDisplayValues(t *testing.T) {
	// Test that error messages contain proper display values with units

	tests := []struct {
		name        string
		modify      func(*Race)
		errContains []string // all strings must be present in error message
	}{
		{
			name: "gravity high edge error shows computed value and maximum in g",
			modify: func(r *Race) {
				r.GravityWidth = 30
				r.GravityCenter = 80 // high edge = 110 internal → 8.00g (clamped)
			},
			errContains: []string{"8.00g", "above maximum 8.00g"},
		},
		{
			name: "gravity low edge error shows computed value and minimum in g",
			modify: func(r *Race) {
				r.GravityWidth = 30
				r.GravityCenter = 20 // low edge = -10 internal → 0.12g (clamped)
			},
			errContains: []string{"0.12g", "below minimum 0.12g"},
		},
		{
			name: "temperature high edge error shows computed value in °C",
			modify: func(r *Race) {
				r.TemperatureWidth = 30
				r.TemperatureCenter = 80 // high edge = 110 internal → 240°C
			},
			errContains: []string{"240°C", "above maximum 200°C"},
		},
		{
			name: "temperature low edge error shows computed value in °C",
			modify: func(r *Race) {
				r.TemperatureWidth = 30
				r.TemperatureCenter = 20 // low edge = -10 internal → -240°C
			},
			errContains: []string{"-240°C", "below minimum -200°C"},
		},
		{
			name: "radiation high edge error shows computed value in mR",
			modify: func(r *Race) {
				r.RadiationWidth = 30
				r.RadiationCenter = 80 // high edge = 110 internal → 110mR
			},
			errContains: []string{"110mR", "above maximum 100mR"},
		},
		{
			name: "radiation low edge error shows computed value in mR",
			modify: func(r *Race) {
				r.RadiationWidth = 30
				r.RadiationCenter = 20 // low edge = -10 internal → -10mR
			},
			errContains: []string{"-10mR", "below minimum 0mR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Default()
			tt.modify(r)

			errs := Validate(r)

			if len(errs) == 0 {
				t.Error("expected validation error, got none")
				return
			}

			// Concatenate all error messages for easier searching
			var allMessages string
			for _, err := range errs {
				allMessages += err.Message + " "
			}

			for _, substr := range tt.errContains {
				if !strings.Contains(allMessages, substr) {
					t.Errorf("expected error message to contain %q, got: %v", substr, errs)
				}
			}
		})
	}
}
