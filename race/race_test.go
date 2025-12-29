package race

import (
	"testing"
)

func TestDefault(t *testing.T) {
	r := Default()

	// Check basic defaults
	if r.SingularName != "Humanoid" {
		t.Errorf("expected SingularName 'Humanoid', got '%s'", r.SingularName)
	}
	if r.PluralName != "Humanoids" {
		t.Errorf("expected PluralName 'Humanoids', got '%s'", r.PluralName)
	}
	if r.PRT != 9 { // JOAT
		t.Errorf("expected PRT 9 (JOAT), got %d", r.PRT)
	}
	if r.LRT != 0 {
		t.Errorf("expected LRT 0, got %d", r.LRT)
	}

	// Check habitability defaults
	if r.GravityCenter != 50 || r.GravityWidth != 25 {
		t.Errorf("expected gravity 50±25, got %d±%d", r.GravityCenter, r.GravityWidth)
	}
	if r.TemperatureCenter != 50 || r.TemperatureWidth != 25 {
		t.Errorf("expected temperature 50±25, got %d±%d", r.TemperatureCenter, r.TemperatureWidth)
	}
	if r.RadiationCenter != 50 || r.RadiationWidth != 25 {
		t.Errorf("expected radiation 50±25, got %d±%d", r.RadiationCenter, r.RadiationWidth)
	}

	// Check economy defaults
	if r.GrowthRate != 15 {
		t.Errorf("expected growth rate 15, got %d", r.GrowthRate)
	}
	if r.ColonistsPerResource != 1000 {
		t.Errorf("expected colonists per resource 1000, got %d", r.ColonistsPerResource)
	}
}

func TestClone(t *testing.T) {
	r := Default()
	clone := r.Clone()

	// Modify clone
	clone.SingularName = "Klingon"
	clone.PRT = 2

	// Original should be unchanged
	if r.SingularName != "Humanoid" {
		t.Error("clone modified original SingularName")
	}
	if r.PRT != 9 {
		t.Error("clone modified original PRT")
	}
}

func TestHabLowHigh(t *testing.T) {
	r := Default()

	if r.GravityLow() != 25 || r.GravityHigh() != 75 {
		t.Errorf("expected gravity 25-75, got %d-%d", r.GravityLow(), r.GravityHigh())
	}
	if r.TemperatureLow() != 25 || r.TemperatureHigh() != 75 {
		t.Errorf("expected temperature 25-75, got %d-%d", r.TemperatureLow(), r.TemperatureHigh())
	}
	if r.RadiationLow() != 25 || r.RadiationHigh() != 75 {
		t.Errorf("expected radiation 25-75, got %d-%d", r.RadiationLow(), r.RadiationHigh())
	}
}

func TestNumImmunities(t *testing.T) {
	r := Default()
	if r.NumImmunities() != 0 {
		t.Errorf("expected 0 immunities, got %d", r.NumImmunities())
	}

	r.GravityImmune = true
	if r.NumImmunities() != 1 {
		t.Errorf("expected 1 immunity, got %d", r.NumImmunities())
	}

	r.TemperatureImmune = true
	r.RadiationImmune = true
	if r.NumImmunities() != 3 {
		t.Errorf("expected 3 immunities, got %d", r.NumImmunities())
	}
}

func TestHasLRT(t *testing.T) {
	r := Default()
	if r.HasLRT(1) {
		t.Error("default race should not have IFE")
	}

	r.LRT = 0x0001 // IFE
	if !r.HasLRT(1) {
		t.Error("race with LRT 1 should have IFE")
	}

	r.LRT = 0x0403 // IFE + TT + NAS
	if !r.HasLRT(1) || !r.HasLRT(2) || !r.HasLRT(0x400) {
		t.Error("race should have multiple LRTs")
	}
}
