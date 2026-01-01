package race

import (
	"testing"
)

func TestBuilderNew(t *testing.T) {
	b := New()
	result := b.Get()

	if result.Race.SingularName != "Humanoid" {
		t.Errorf("expected SingularName 'Humanoid', got '%s'", result.Race.SingularName)
	}

	// Default race should be valid
	if !result.IsValid {
		t.Errorf("default race should be valid, errors: %v", result.Errors)
	}
}

func TestBuilderName(t *testing.T) {
	b := New()
	result := b.Name("Klingon", "Klingons")

	if result.Race.SingularName != "Klingon" {
		t.Errorf("expected SingularName 'Klingon', got '%s'", result.Race.SingularName)
	}
	if result.Race.PluralName != "Klingons" {
		t.Errorf("expected PluralName 'Klingons', got '%s'", result.Race.PluralName)
	}
}

func TestBuilderPRT(t *testing.T) {
	b := New()
	result := b.PRT(2) // War Monger

	if result.Race.PRT != 2 {
		t.Errorf("expected PRT 2, got %d", result.Race.PRT)
	}
}

func TestBuilderLRTs(t *testing.T) {
	b := New()

	// Add IFE (index 0)
	result := b.AddLRT(0)
	if result.Race.LRT != 1 {
		t.Errorf("expected LRT 1, got %d", result.Race.LRT)
	}

	// Add TT (index 1)
	result = b.AddLRT(1)
	if result.Race.LRT != 3 {
		t.Errorf("expected LRT 3, got %d", result.Race.LRT)
	}

	// Remove IFE
	result = b.RemoveLRT(0)
	if result.Race.LRT != 2 {
		t.Errorf("expected LRT 2, got %d", result.Race.LRT)
	}

	// Set all LRTs at once
	result = b.SetLRTs(0x0005) // IFE + ARM
	if result.Race.LRT != 5 {
		t.Errorf("expected LRT 5, got %d", result.Race.LRT)
	}
}

func TestBuilderHabitability(t *testing.T) {
	b := New()

	result := b.Gravity(60, 30)
	if result.Race.GravityCenter != 60 || result.Race.GravityWidth != 30 {
		t.Errorf("expected gravity 60±30, got %d±%d", result.Race.GravityCenter, result.Race.GravityWidth)
	}

	result = b.GravityImmune(true)
	if !result.Race.GravityImmune {
		t.Error("expected gravity immune")
	}

	result = b.Temperature(40, 20)
	if result.Race.TemperatureCenter != 40 || result.Race.TemperatureWidth != 20 {
		t.Errorf("expected temperature 40±20, got %d±%d", result.Race.TemperatureCenter, result.Race.TemperatureWidth)
	}

	result = b.Radiation(70, 15)
	if result.Race.RadiationCenter != 70 || result.Race.RadiationWidth != 15 {
		t.Errorf("expected radiation 70±15, got %d±%d", result.Race.RadiationCenter, result.Race.RadiationWidth)
	}
}

func TestBuilderGrowthRate(t *testing.T) {
	b := New()
	result := b.GrowthRate(18)

	if result.Race.GrowthRate != 18 {
		t.Errorf("expected growth rate 18, got %d", result.Race.GrowthRate)
	}
}

func TestBuilderEconomy(t *testing.T) {
	b := New()

	result := b.ColonistsPerResource(800)
	if result.Race.ColonistsPerResource != 800 {
		t.Errorf("expected colonists per resource 800, got %d", result.Race.ColonistsPerResource)
	}

	result = b.Factories(15, 8, 20, true)
	if result.Race.FactoryOutput != 15 {
		t.Errorf("expected factory output 15, got %d", result.Race.FactoryOutput)
	}
	if result.Race.FactoryCost != 8 {
		t.Errorf("expected factory cost 8, got %d", result.Race.FactoryCost)
	}
	if result.Race.FactoryCount != 20 {
		t.Errorf("expected factory count 20, got %d", result.Race.FactoryCount)
	}
	if !result.Race.FactoriesUseLessGerm {
		t.Error("expected factories use less germ")
	}

	result = b.Mines(12, 4, 15)
	if result.Race.MineOutput != 12 {
		t.Errorf("expected mine output 12, got %d", result.Race.MineOutput)
	}
	if result.Race.MineCost != 4 {
		t.Errorf("expected mine cost 4, got %d", result.Race.MineCost)
	}
	if result.Race.MineCount != 15 {
		t.Errorf("expected mine count 15, got %d", result.Race.MineCount)
	}
}

func TestBuilderResearch(t *testing.T) {
	b := New()
	result := b.Research(0, 2, 1, 1, 0, 2) // Extra, Less, Standard, Standard, Extra, Less

	if result.Race.ResearchEnergy != 0 {
		t.Errorf("expected research energy 0, got %d", result.Race.ResearchEnergy)
	}
	if result.Race.ResearchWeapons != 2 {
		t.Errorf("expected research weapons 2, got %d", result.Race.ResearchWeapons)
	}
	if result.Race.ResearchPropulsion != 1 {
		t.Errorf("expected research propulsion 1, got %d", result.Race.ResearchPropulsion)
	}
	if result.Race.ResearchConstruction != 1 {
		t.Errorf("expected research construction 1, got %d", result.Race.ResearchConstruction)
	}
	if result.Race.ResearchElectronics != 0 {
		t.Errorf("expected research electronics 0, got %d", result.Race.ResearchElectronics)
	}
	if result.Race.ResearchBiotech != 2 {
		t.Errorf("expected research biotech 2, got %d", result.Race.ResearchBiotech)
	}

	result = b.TechsStartHigh(true)
	if !result.Race.TechsStartHigh {
		t.Error("expected techs start high")
	}
}

func TestBuilderLeftoverPoints(t *testing.T) {
	b := New()
	result := b.LeftoverPointsOn(LeftoverFactories)

	if result.Race.LeftoverPointsOn != LeftoverFactories {
		t.Errorf("expected leftover points on factories, got %d", result.Race.LeftoverPointsOn)
	}
}

func TestBuilderFinish(t *testing.T) {
	b := New()
	race, err := b.Finish()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if race == nil {
		t.Error("expected race, got nil")
	}
}

func TestBuilderFinishInvalid(t *testing.T) {
	b := New()
	b.Name("", "") // Invalid: empty names

	_, err := b.Finish()
	if err == nil {
		t.Error("expected error for invalid race")
	}
}

func TestBuilderChaining(t *testing.T) {
	b := New()

	// Test that we can chain operations
	b.Name("Warrior", "Warriors")
	b.PRT(PRTWarMonger)
	result := b.GrowthRate(18)

	if result.Race.SingularName != "Warrior" {
		t.Error("chaining didn't preserve name")
	}
	if result.Race.PRT != PRTWarMonger {
		t.Error("chaining didn't preserve PRT")
	}
	if result.Race.GrowthRate != 18 {
		t.Error("chaining didn't preserve growth rate")
	}
}
