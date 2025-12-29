package race

import (
	"testing"
)

func TestLRTs(t *testing.T) {
	// Test empty
	if LRTs() != 0 {
		t.Errorf("expected LRTs() to return 0, got %d", LRTs())
	}

	// Test single LRT
	if LRTs(LRTImprovedFuelEfficiency) != 0x0001 {
		t.Errorf("expected LRTs(IFE) to return 0x0001, got 0x%04X", LRTs(LRTImprovedFuelEfficiency))
	}

	// Test multiple LRTs - Rabbitoid: IFE + TT + CE + NAS = 0x0503
	result := LRTs(LRTImprovedFuelEfficiency, LRTTotalTerraforming, LRTCheapEngines, LRTNoAdvancedScanners)
	expected := uint16(0x0503)
	if result != expected {
		t.Errorf("expected LRTs(IFE, TT, CE, NAS) to return 0x%04X, got 0x%04X", expected, result)
	}

	// Test Insectoid: ISB + CE + RS = 0x2108
	result = LRTs(LRTImprovedStarbases, LRTCheapEngines, LRTRegeneratingShields)
	expected = uint16(0x2108)
	if result != expected {
		t.Errorf("expected LRTs(ISB, CE, RS) to return 0x%04X, got 0x%04X", expected, result)
	}

	// Test Silicanoid: IFE + UR + OBRM + BET = 0x1221
	result = LRTs(LRTImprovedFuelEfficiency, LRTUltimateRecycling, LRTOnlyBasicRemoteMining, LRTBleedingEdgeTechnology)
	expected = uint16(0x1221)
	if result != expected {
		t.Errorf("expected LRTs(IFE, UR, OBRM, BET) to return 0x%04X, got 0x%04X", expected, result)
	}
}

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

	// Check habitability defaults (Humanoid values: 50±35, range 15-85)
	if r.GravityCenter != 50 || r.GravityWidth != 35 {
		t.Errorf("expected gravity 50±35, got %d±%d", r.GravityCenter, r.GravityWidth)
	}
	if r.TemperatureCenter != 50 || r.TemperatureWidth != 35 {
		t.Errorf("expected temperature 50±35, got %d±%d", r.TemperatureCenter, r.TemperatureWidth)
	}
	if r.RadiationCenter != 50 || r.RadiationWidth != 35 {
		t.Errorf("expected radiation 50±35, got %d±%d", r.RadiationCenter, r.RadiationWidth)
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

	// Default (Humanoid) has range 15-85 (center 50, width 35)
	if r.GravityLow() != 15 || r.GravityHigh() != 85 {
		t.Errorf("expected gravity 15-85, got %d-%d", r.GravityLow(), r.GravityHigh())
	}
	if r.TemperatureLow() != 15 || r.TemperatureHigh() != 85 {
		t.Errorf("expected temperature 15-85, got %d-%d", r.TemperatureLow(), r.TemperatureHigh())
	}
	if r.RadiationLow() != 15 || r.RadiationHigh() != 85 {
		t.Errorf("expected radiation 15-85, got %d-%d", r.RadiationLow(), r.RadiationHigh())
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

func TestHumanoid(t *testing.T) {
	r := Humanoid()

	// Check identity
	if r.SingularName != "Humanoid" {
		t.Errorf("expected SingularName 'Humanoid', got '%s'", r.SingularName)
	}
	if r.PluralName != "Humanoids" {
		t.Errorf("expected empty PluralName, got '%s'", r.PluralName)
	}
	if r.Icon != 1 {
		t.Errorf("expected Icon 1, got %d", r.Icon)
	}

	// Check traits
	if r.PRT != PRTJackOfAllTrades {
		t.Errorf("expected PRT %d (JOAT), got %d", PRTJackOfAllTrades, r.PRT)
	}
	if r.LRT != 0 {
		t.Errorf("expected LRT 0, got %d", r.LRT)
	}

	// Check habitability - should be range 15-85
	if r.GravityCenter != 50 || r.GravityWidth != 35 {
		t.Errorf("expected gravity 50±35, got %d±%d", r.GravityCenter, r.GravityWidth)
	}
	if r.GravityLow() != 15 || r.GravityHigh() != 85 {
		t.Errorf("expected gravity range 15-85, got %d-%d", r.GravityLow(), r.GravityHigh())
	}
	if r.TemperatureCenter != 50 || r.TemperatureWidth != 35 {
		t.Errorf("expected temperature 50±35, got %d±%d", r.TemperatureCenter, r.TemperatureWidth)
	}
	if r.RadiationCenter != 50 || r.RadiationWidth != 35 {
		t.Errorf("expected radiation 50±35, got %d±%d", r.RadiationCenter, r.RadiationWidth)
	}

	// Check economy
	if r.GrowthRate != 15 {
		t.Errorf("expected growth rate 15, got %d", r.GrowthRate)
	}
	if r.ColonistsPerResource != 1000 {
		t.Errorf("expected colonists per resource 1000, got %d", r.ColonistsPerResource)
	}
	if r.FactoryOutput != 10 || r.FactoryCost != 10 || r.FactoryCount != 10 {
		t.Errorf("expected factory 10/10/10, got %d/%d/%d", r.FactoryOutput, r.FactoryCost, r.FactoryCount)
	}
	if r.MineOutput != 10 || r.MineCost != 5 || r.MineCount != 10 {
		t.Errorf("expected mine 10/5/10, got %d/%d/%d", r.MineOutput, r.MineCost, r.MineCount)
	}

	// Check research
	if r.ResearchEnergy != ResearchCostStandard {
		t.Errorf("expected standard research cost for energy, got %d", r.ResearchEnergy)
	}
}

func TestRabbitoid(t *testing.T) {
	r := Rabbitoid()

	// Check identity
	if r.SingularName != "Rabbitoid" {
		t.Errorf("expected SingularName 'Rabbitoid', got '%s'", r.SingularName)
	}
	if r.PluralName != "Rabbitoids" {
		t.Errorf("expected PluralName 'Rabbitoids', got '%s'", r.PluralName)
	}
	if r.Icon != 12 {
		t.Errorf("expected Icon 12, got %d", r.Icon)
	}

	// Check traits
	if r.PRT != PRTInterstellarTraveler {
		t.Errorf("expected PRT %d (IT), got %d", PRTInterstellarTraveler, r.PRT)
	}
	// LRT: IFE + TT + CE + NAS = 0x0503
	expectedLRT := uint16(0x0503)
	if r.LRT != expectedLRT {
		t.Errorf("expected LRT 0x%04X, got 0x%04X", expectedLRT, r.LRT)
	}

	// Check habitability - should be range 10-56, 35-81, 13-53
	if r.GravityCenter != 33 || r.GravityWidth != 23 {
		t.Errorf("expected gravity 33±23, got %d±%d", r.GravityCenter, r.GravityWidth)
	}
	if r.GravityLow() != 10 || r.GravityHigh() != 56 {
		t.Errorf("expected gravity range 10-56, got %d-%d", r.GravityLow(), r.GravityHigh())
	}
	if r.TemperatureCenter != 58 || r.TemperatureWidth != 23 {
		t.Errorf("expected temperature 58±23, got %d±%d", r.TemperatureCenter, r.TemperatureWidth)
	}
	if r.RadiationCenter != 33 || r.RadiationWidth != 20 {
		t.Errorf("expected radiation 33±20, got %d±%d", r.RadiationCenter, r.RadiationWidth)
	}

	// Check economy
	if r.GrowthRate != 20 {
		t.Errorf("expected growth rate 20, got %d", r.GrowthRate)
	}
	if r.ColonistsPerResource != 1000 {
		t.Errorf("expected colonists per resource 1000, got %d", r.ColonistsPerResource)
	}
	if r.FactoryOutput != 10 || r.FactoryCost != 9 || r.FactoryCount != 17 {
		t.Errorf("expected factory 10/9/17, got %d/%d/%d", r.FactoryOutput, r.FactoryCost, r.FactoryCount)
	}
	if !r.FactoriesUseLessGerm {
		t.Error("expected FactoriesUseLessGerm to be true")
	}
	if r.MineOutput != 10 || r.MineCost != 9 || r.MineCount != 10 {
		t.Errorf("expected mine 10/9/10, got %d/%d/%d", r.MineOutput, r.MineCost, r.MineCount)
	}

	// Check research - energy/weapons extra, propulsion/biotech less
	if r.ResearchEnergy != ResearchCostExtra {
		t.Errorf("expected extra research cost for energy, got %d", r.ResearchEnergy)
	}
	if r.ResearchWeapons != ResearchCostExtra {
		t.Errorf("expected extra research cost for weapons, got %d", r.ResearchWeapons)
	}
	if r.ResearchPropulsion != ResearchCostLess {
		t.Errorf("expected less research cost for propulsion, got %d", r.ResearchPropulsion)
	}
	if r.ResearchBiotech != ResearchCostLess {
		t.Errorf("expected less research cost for biotech, got %d", r.ResearchBiotech)
	}

	// Check leftover points
	if r.LeftoverPointsOn != LeftoverMineralConcentration {
		t.Errorf("expected LeftoverMineralConcentration, got %d", r.LeftoverPointsOn)
	}
}

func TestInsectoid(t *testing.T) {
	r := Insectoid()

	// Check identity
	if r.SingularName != "Insectoid" {
		t.Errorf("expected SingularName 'Insectoid', got '%s'", r.SingularName)
	}
	if r.PluralName != "Insectoids" {
		t.Errorf("expected PluralName 'Insectoids', got '%s'", r.PluralName)
	}
	if r.Icon != 4 {
		t.Errorf("expected Icon 4, got %d", r.Icon)
	}

	// Check traits
	if r.PRT != PRTWarMonger {
		t.Errorf("expected PRT %d (WM), got %d", PRTWarMonger, r.PRT)
	}
	// LRT: ISB + CE + RS = 0x2108
	expectedLRT := uint16(0x2108)
	if r.LRT != expectedLRT {
		t.Errorf("expected LRT 0x%04X, got 0x%04X", expectedLRT, r.LRT)
	}

	// Check habitability - gravity immune
	if !r.GravityImmune {
		t.Error("expected gravity immunity")
	}
	if r.TemperatureCenter != 50 || r.TemperatureWidth != 50 {
		t.Errorf("expected temperature 50±50, got %d±%d", r.TemperatureCenter, r.TemperatureWidth)
	}
	if r.RadiationCenter != 85 || r.RadiationWidth != 15 {
		t.Errorf("expected radiation 85±15, got %d±%d", r.RadiationCenter, r.RadiationWidth)
	}

	// Check economy
	if r.GrowthRate != 10 {
		t.Errorf("expected growth rate 10, got %d", r.GrowthRate)
	}
	if r.MineOutput != 9 || r.MineCost != 10 || r.MineCount != 6 {
		t.Errorf("expected mine 9/10/6, got %d/%d/%d", r.MineOutput, r.MineCost, r.MineCount)
	}

	// Check research - energy/weapons/propulsion/construction less, biotech extra
	if r.ResearchEnergy != ResearchCostLess {
		t.Errorf("expected less research cost for energy, got %d", r.ResearchEnergy)
	}
	if r.ResearchBiotech != ResearchCostExtra {
		t.Errorf("expected extra research cost for biotech, got %d", r.ResearchBiotech)
	}

	// Check leftover points
	if r.LeftoverPointsOn != LeftoverMines {
		t.Errorf("expected LeftoverMines, got %d", r.LeftoverPointsOn)
	}
}

func TestNucleotid(t *testing.T) {
	r := Nucleotid()

	// Check identity
	if r.SingularName != "Nucleotid" {
		t.Errorf("expected SingularName 'Nucleotid', got '%s'", r.SingularName)
	}
	if r.PluralName != "Nucleotids" {
		t.Errorf("expected PluralName 'Nucleotids', got '%s'", r.PluralName)
	}
	if r.Icon != 25 {
		t.Errorf("expected Icon 25, got %d", r.Icon)
	}

	// Check traits
	if r.PRT != PRTSuperStealth {
		t.Errorf("expected PRT %d (SS), got %d", PRTSuperStealth, r.PRT)
	}
	// LRT: ARM + ISB = 0x000C
	expectedLRT := uint16(0x000C)
	if r.LRT != expectedLRT {
		t.Errorf("expected LRT 0x%04X, got 0x%04X", expectedLRT, r.LRT)
	}

	// Check habitability - gravity immune, wide temp/rad
	if !r.GravityImmune {
		t.Error("expected gravity immunity")
	}
	if r.TemperatureCenter != 50 || r.TemperatureWidth != 38 {
		t.Errorf("expected temperature 50±38, got %d±%d", r.TemperatureCenter, r.TemperatureWidth)
	}
	if r.RadiationCenter != 50 || r.RadiationWidth != 50 {
		t.Errorf("expected radiation 50±50, got %d±%d", r.RadiationCenter, r.RadiationWidth)
	}

	// Check economy
	if r.GrowthRate != 10 {
		t.Errorf("expected growth rate 10, got %d", r.GrowthRate)
	}
	if r.ColonistsPerResource != 900 {
		t.Errorf("expected colonists per resource 900, got %d", r.ColonistsPerResource)
	}
	if r.MineOutput != 10 || r.MineCost != 15 || r.MineCount != 5 {
		t.Errorf("expected mine 10/15/5, got %d/%d/%d", r.MineOutput, r.MineCost, r.MineCount)
	}

	// Check research - all extra, techs start high
	if r.ResearchEnergy != ResearchCostExtra {
		t.Errorf("expected extra research cost for energy, got %d", r.ResearchEnergy)
	}
	if r.ResearchBiotech != ResearchCostExtra {
		t.Errorf("expected extra research cost for biotech, got %d", r.ResearchBiotech)
	}
	if !r.TechsStartHigh {
		t.Error("expected TechsStartHigh to be true")
	}

	// Check leftover points
	if r.LeftoverPointsOn != LeftoverFactories {
		t.Errorf("expected LeftoverFactories, got %d", r.LeftoverPointsOn)
	}
}

func TestSilicanoid(t *testing.T) {
	r := Silicanoid()

	// Check identity
	if r.SingularName != "Silicanoid" {
		t.Errorf("expected SingularName 'Silicanoid', got '%s'", r.SingularName)
	}
	if r.PluralName != "Silicanoids" {
		t.Errorf("expected PluralName 'Silicanoids', got '%s'", r.PluralName)
	}
	if r.Icon != 5 {
		t.Errorf("expected Icon 5, got %d", r.Icon)
	}

	// Check traits
	if r.PRT != PRTHyperExpansion {
		t.Errorf("expected PRT %d (HE), got %d", PRTHyperExpansion, r.PRT)
	}
	// LRT: IFE + UR + OBRM + BET = 0x1221
	expectedLRT := uint16(0x1221)
	if r.LRT != expectedLRT {
		t.Errorf("expected LRT 0x%04X, got 0x%04X", expectedLRT, r.LRT)
	}

	// Check habitability - all immune
	if !r.GravityImmune {
		t.Error("expected gravity immunity")
	}
	if !r.TemperatureImmune {
		t.Error("expected temperature immunity")
	}
	if !r.RadiationImmune {
		t.Error("expected radiation immunity")
	}

	// Check economy
	if r.GrowthRate != 6 {
		t.Errorf("expected growth rate 6, got %d", r.GrowthRate)
	}
	if r.ColonistsPerResource != 800 {
		t.Errorf("expected colonists per resource 800, got %d", r.ColonistsPerResource)
	}
	if r.FactoryOutput != 12 || r.FactoryCost != 12 || r.FactoryCount != 15 {
		t.Errorf("expected factory 12/12/15, got %d/%d/%d", r.FactoryOutput, r.FactoryCost, r.FactoryCount)
	}
	if r.MineOutput != 10 || r.MineCost != 9 || r.MineCount != 10 {
		t.Errorf("expected mine 10/9/10, got %d/%d/%d", r.MineOutput, r.MineCost, r.MineCount)
	}

	// Check research - propulsion/construction less, biotech extra
	if r.ResearchPropulsion != ResearchCostLess {
		t.Errorf("expected less research cost for propulsion, got %d", r.ResearchPropulsion)
	}
	if r.ResearchConstruction != ResearchCostLess {
		t.Errorf("expected less research cost for construction, got %d", r.ResearchConstruction)
	}
	if r.ResearchBiotech != ResearchCostExtra {
		t.Errorf("expected extra research cost for biotech, got %d", r.ResearchBiotech)
	}

	// Check leftover points
	if r.LeftoverPointsOn != LeftoverFactories {
		t.Errorf("expected LeftoverFactories, got %d", r.LeftoverPointsOn)
	}
}

func TestAntetheral(t *testing.T) {
	r := Antetheral()

	// Check identity
	if r.SingularName != "Antetheral" {
		t.Errorf("expected SingularName 'Antetheral', got '%s'", r.SingularName)
	}
	if r.PluralName != "Antetherals" {
		t.Errorf("expected PluralName 'Antetherals', got '%s'", r.PluralName)
	}
	if r.Icon != 18 {
		t.Errorf("expected Icon 18, got %d", r.Icon)
	}

	// Check traits
	if r.PRT != PRTSpaceDemolition {
		t.Errorf("expected PRT %d (SD), got %d", PRTSpaceDemolition, r.PRT)
	}
	// LRT: ARM + MA + NRSE + CE + NAS = 0x05C4
	expectedLRT := uint16(0x05C4)
	if r.LRT != expectedLRT {
		t.Errorf("expected LRT 0x%04X, got 0x%04X", expectedLRT, r.LRT)
	}

	// Check habitability - narrow gravity, full temp, narrow radiation
	if r.GravityCenter != 15 || r.GravityWidth != 15 {
		t.Errorf("expected gravity 15±15, got %d±%d", r.GravityCenter, r.GravityWidth)
	}
	if r.GravityLow() != 0 || r.GravityHigh() != 30 {
		t.Errorf("expected gravity range 0-30, got %d-%d", r.GravityLow(), r.GravityHigh())
	}
	if r.TemperatureCenter != 50 || r.TemperatureWidth != 50 {
		t.Errorf("expected temperature 50±50, got %d±%d", r.TemperatureCenter, r.TemperatureWidth)
	}
	if r.RadiationCenter != 85 || r.RadiationWidth != 15 {
		t.Errorf("expected radiation 85±15, got %d±%d", r.RadiationCenter, r.RadiationWidth)
	}

	// Check economy
	if r.GrowthRate != 7 {
		t.Errorf("expected growth rate 7, got %d", r.GrowthRate)
	}
	if r.ColonistsPerResource != 700 {
		t.Errorf("expected colonists per resource 700, got %d", r.ColonistsPerResource)
	}
	if r.FactoryOutput != 11 || r.FactoryCost != 10 || r.FactoryCount != 18 {
		t.Errorf("expected factory 11/10/18, got %d/%d/%d", r.FactoryOutput, r.FactoryCost, r.FactoryCount)
	}
	if r.MineOutput != 10 || r.MineCost != 10 || r.MineCount != 10 {
		t.Errorf("expected mine 10/10/10, got %d/%d/%d", r.MineOutput, r.MineCost, r.MineCount)
	}

	// Check research - most less, weapons extra
	if r.ResearchEnergy != ResearchCostLess {
		t.Errorf("expected less research cost for energy, got %d", r.ResearchEnergy)
	}
	if r.ResearchWeapons != ResearchCostExtra {
		t.Errorf("expected extra research cost for weapons, got %d", r.ResearchWeapons)
	}
	if r.ResearchBiotech != ResearchCostLess {
		t.Errorf("expected less research cost for biotech, got %d", r.ResearchBiotech)
	}

	// Check leftover points
	if r.LeftoverPointsOn != LeftoverSurfaceMinerals {
		t.Errorf("expected LeftoverSurfaceMinerals, got %d", r.LeftoverPointsOn)
	}
}

func TestRandom(t *testing.T) {
	// Test that Random() generates valid races
	for i := 0; i < 10; i++ {
		r := Random()

		// Check identity
		if r.SingularName != "Random" {
			t.Errorf("expected SingularName 'Random', got '%s'", r.SingularName)
		}
		if r.PluralName != "Randoms" {
			t.Errorf("expected PluralName 'Randoms', got '%s'", r.PluralName)
		}

		// Check traits are within valid ranges
		if r.PRT < 0 || r.PRT > 9 {
			t.Errorf("PRT %d out of range [0,9]", r.PRT)
		}
		if r.LRT&^uint16(0x3FFF) != 0 {
			t.Errorf("LRT 0x%04X has invalid bits", r.LRT)
		}

		// Check icon is valid
		if r.Icon < 0 || r.Icon > 31 {
			t.Errorf("Icon %d out of range [0,31]", r.Icon)
		}

		// Check habitability ranges (when not immune)
		if !r.GravityImmune {
			if r.GravityCenter < 0 || r.GravityCenter > 100 {
				t.Errorf("GravityCenter %d out of range [0,100]", r.GravityCenter)
			}
			if r.GravityWidth < 0 || r.GravityWidth > 50 {
				t.Errorf("GravityWidth %d out of range [0,50]", r.GravityWidth)
			}
		}
		if !r.TemperatureImmune {
			if r.TemperatureCenter < 0 || r.TemperatureCenter > 100 {
				t.Errorf("TemperatureCenter %d out of range [0,100]", r.TemperatureCenter)
			}
			if r.TemperatureWidth < 0 || r.TemperatureWidth > 50 {
				t.Errorf("TemperatureWidth %d out of range [0,50]", r.TemperatureWidth)
			}
		}
		if !r.RadiationImmune {
			if r.RadiationCenter < 0 || r.RadiationCenter > 100 {
				t.Errorf("RadiationCenter %d out of range [0,100]", r.RadiationCenter)
			}
			if r.RadiationWidth < 0 || r.RadiationWidth > 50 {
				t.Errorf("RadiationWidth %d out of range [0,50]", r.RadiationWidth)
			}
		}

		// Check growth rate
		if r.GrowthRate < 1 || r.GrowthRate > 20 {
			t.Errorf("GrowthRate %d out of range [1,20]", r.GrowthRate)
		}

		// Check economy
		if r.ColonistsPerResource < 700 || r.ColonistsPerResource > 2500 {
			t.Errorf("ColonistsPerResource %d out of range [700,2500]", r.ColonistsPerResource)
		}
		if r.FactoryOutput < 5 || r.FactoryOutput > 25 {
			t.Errorf("FactoryOutput %d out of range [5,25]", r.FactoryOutput)
		}
		if r.FactoryCost < 5 || r.FactoryCost > 25 {
			t.Errorf("FactoryCost %d out of range [5,25]", r.FactoryCost)
		}
		if r.FactoryCount < 5 || r.FactoryCount > 25 {
			t.Errorf("FactoryCount %d out of range [5,25]", r.FactoryCount)
		}
		if r.MineOutput < 5 || r.MineOutput > 25 {
			t.Errorf("MineOutput %d out of range [5,25]", r.MineOutput)
		}
		if r.MineCost < 2 || r.MineCost > 15 {
			t.Errorf("MineCost %d out of range [2,15]", r.MineCost)
		}
		if r.MineCount < 5 || r.MineCount > 25 {
			t.Errorf("MineCount %d out of range [5,25]", r.MineCount)
		}

		// Check research costs
		if r.ResearchEnergy < 0 || r.ResearchEnergy > 2 {
			t.Errorf("ResearchEnergy %d out of range [0,2]", r.ResearchEnergy)
		}
		if r.ResearchWeapons < 0 || r.ResearchWeapons > 2 {
			t.Errorf("ResearchWeapons %d out of range [0,2]", r.ResearchWeapons)
		}
		if r.ResearchPropulsion < 0 || r.ResearchPropulsion > 2 {
			t.Errorf("ResearchPropulsion %d out of range [0,2]", r.ResearchPropulsion)
		}
		if r.ResearchConstruction < 0 || r.ResearchConstruction > 2 {
			t.Errorf("ResearchConstruction %d out of range [0,2]", r.ResearchConstruction)
		}
		if r.ResearchElectronics < 0 || r.ResearchElectronics > 2 {
			t.Errorf("ResearchElectronics %d out of range [0,2]", r.ResearchElectronics)
		}
		if r.ResearchBiotech < 0 || r.ResearchBiotech > 2 {
			t.Errorf("ResearchBiotech %d out of range [0,2]", r.ResearchBiotech)
		}

		// Check leftover points
		if r.LeftoverPointsOn < LeftoverSurfaceMinerals || r.LeftoverPointsOn > LeftoverMineralConcentration {
			t.Errorf("LeftoverPointsOn %d out of range", r.LeftoverPointsOn)
		}

		// Validate the race passes validation
		errors := Validate(r)
		if len(errors) > 0 {
			t.Errorf("Random race failed validation: %v", errors)
		}
	}
}

func TestRandomWithSeed(t *testing.T) {
	// Test that RandomWithSeed produces deterministic results
	seed := int64(12345)

	r1 := RandomWithSeed(seed)
	r2 := RandomWithSeed(seed)

	// Same seed should produce identical races
	if r1.PRT != r2.PRT {
		t.Errorf("same seed produced different PRT: %d vs %d", r1.PRT, r2.PRT)
	}
	if r1.LRT != r2.LRT {
		t.Errorf("same seed produced different LRT: 0x%04X vs 0x%04X", r1.LRT, r2.LRT)
	}
	if r1.GravityImmune != r2.GravityImmune {
		t.Errorf("same seed produced different GravityImmune: %v vs %v", r1.GravityImmune, r2.GravityImmune)
	}
	if r1.GravityCenter != r2.GravityCenter {
		t.Errorf("same seed produced different GravityCenter: %d vs %d", r1.GravityCenter, r2.GravityCenter)
	}
	if r1.GrowthRate != r2.GrowthRate {
		t.Errorf("same seed produced different GrowthRate: %d vs %d", r1.GrowthRate, r2.GrowthRate)
	}
	if r1.ColonistsPerResource != r2.ColonistsPerResource {
		t.Errorf("same seed produced different ColonistsPerResource: %d vs %d", r1.ColonistsPerResource, r2.ColonistsPerResource)
	}
	if r1.Icon != r2.Icon {
		t.Errorf("same seed produced different Icon: %d vs %d", r1.Icon, r2.Icon)
	}
	if r1.FactoryOutput != r2.FactoryOutput {
		t.Errorf("same seed produced different FactoryOutput: %d vs %d", r1.FactoryOutput, r2.FactoryOutput)
	}
	if r1.ResearchEnergy != r2.ResearchEnergy {
		t.Errorf("same seed produced different ResearchEnergy: %d vs %d", r1.ResearchEnergy, r2.ResearchEnergy)
	}
	if r1.TechsStartHigh != r2.TechsStartHigh {
		t.Errorf("same seed produced different TechsStartHigh: %v vs %v", r1.TechsStartHigh, r2.TechsStartHigh)
	}
	if r1.LeftoverPointsOn != r2.LeftoverPointsOn {
		t.Errorf("same seed produced different LeftoverPointsOn: %d vs %d", r1.LeftoverPointsOn, r2.LeftoverPointsOn)
	}

	// Different seed should produce different races (with very high probability)
	r3 := RandomWithSeed(seed + 1)
	sameCount := 0
	if r1.PRT == r3.PRT {
		sameCount++
	}
	if r1.LRT == r3.LRT {
		sameCount++
	}
	if r1.GrowthRate == r3.GrowthRate {
		sameCount++
	}
	if r1.ColonistsPerResource == r3.ColonistsPerResource {
		sameCount++
	}
	// If more than half the values are the same, something is likely wrong
	if sameCount > 2 {
		t.Errorf("different seeds produced suspiciously similar races")
	}
}
