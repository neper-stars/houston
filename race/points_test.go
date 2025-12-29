package race

import (
	"testing"
)

func TestCalculatePointsHumanoid(t *testing.T) {
	r := Default()
	points := CalculatePoints(r)

	// Humanoid (JOAT with all defaults) should have positive points
	// The exact value depends on the algorithm port accuracy
	if points < 0 {
		t.Errorf("Humanoid should have non-negative points, got %d", points)
	}

	// Log actual value for reference
	t.Logf("Humanoid default points: %d", points)
}

func TestCalculatePointsPRTImpact(t *testing.T) {
	// Test that different PRTs produce different point values
	// In Stars!: negative PRT cost = powerful advantage = costs points (lower total)
	//            positive PRT cost = disadvantage = gives points (higher total)
	base := Default()
	basePoints := CalculatePoints(base)

	// War Monger (PRT 2) has -45 cost = powerful, costs points
	wm := Default()
	wm.PRT = 2
	wmPoints := CalculatePoints(wm)

	// Inner Strength (PRT 4) has +100 cost = weak, gives points
	is := Default()
	is.PRT = 4
	isPoints := CalculatePoints(is)

	// IS (weak trait, +100) should have MORE points than WM (powerful trait, -45)
	if isPoints <= wmPoints {
		t.Errorf("Inner Strength (%d) should have more points than War Monger (%d)", isPoints, wmPoints)
	}

	// Verify the points changed from base
	if wmPoints == basePoints && isPoints == basePoints {
		t.Error("PRT changes should affect point calculation")
	}
}

func TestCalculatePointsLRTImpact(t *testing.T) {
	// In Stars!: negative LRT cost = powerful advantage = costs points (lower total)
	//            positive LRT cost = disadvantage = gives points (higher total)
	base := Default()
	basePoints := CalculatePoints(base)

	// IFE (index 0, bitmask 0x01) has negative cost (-235) = powerful advantage = costs points
	withIFE := Default()
	withIFE.LRT = 0x0001 // IFE at index 0
	ifePoints := CalculatePoints(withIFE)

	// NRSE (index 7, bitmask 0x80) has positive cost (+160) = disadvantage = gives points
	withNRSE := Default()
	withNRSE.LRT = 0x0080 // NRSE at index 7 (1 << 7 = 128 = 0x80)
	nrsePoints := CalculatePoints(withNRSE)

	// IFE (powerful) should DECREASE points
	if ifePoints >= basePoints {
		t.Errorf("IFE should decrease points (base %d, with IFE %d)", basePoints, ifePoints)
	}

	// NRSE (weak) should INCREASE points
	if nrsePoints <= basePoints {
		t.Errorf("NRSE should increase points (base %d, with NRSE %d)", basePoints, nrsePoints)
	}
}

func TestCalculatePointsHabitability(t *testing.T) {
	// Wider habitability should give fewer points (more planets habitable)
	narrow := Default()
	narrow.GravityWidth = 10
	narrow.TemperatureWidth = 10
	narrow.RadiationWidth = 10
	narrowPoints := CalculatePoints(narrow)

	wide := Default()
	wide.GravityWidth = 50
	wide.TemperatureWidth = 50
	wide.RadiationWidth = 50
	widePoints := CalculatePoints(wide)

	if narrowPoints <= widePoints {
		t.Errorf("Narrow hab range (%d) should give more points than wide (%d)", narrowPoints, widePoints)
	}
}

func TestCalculatePointsImmunity(t *testing.T) {
	base := Default()
	basePoints := CalculatePoints(base)

	// One immunity should reduce points
	oneImmune := Default()
	oneImmune.GravityImmune = true
	oneImmunePoints := CalculatePoints(oneImmune)

	// Two immunities should reduce points more (includes -150 penalty)
	twoImmune := Default()
	twoImmune.GravityImmune = true
	twoImmune.TemperatureImmune = true
	twoImmunePoints := CalculatePoints(twoImmune)

	if oneImmunePoints >= basePoints {
		t.Errorf("One immunity (%d) should cost points vs base (%d)", oneImmunePoints, basePoints)
	}

	if twoImmunePoints >= oneImmunePoints {
		t.Errorf("Two immunities (%d) should cost more than one (%d)", twoImmunePoints, oneImmunePoints)
	}
}

func TestCalculatePointsGrowthRate(t *testing.T) {
	// Higher growth rate should cost points
	lowGrowth := Default()
	lowGrowth.GrowthRate = 5
	lowPoints := CalculatePoints(lowGrowth)

	highGrowth := Default()
	highGrowth.GrowthRate = 20
	highPoints := CalculatePoints(highGrowth)

	if lowPoints <= highPoints {
		t.Errorf("Low growth rate (%d) should give more points than high (%d)", lowPoints, highPoints)
	}
}

func TestCalculatePointsResearch(t *testing.T) {
	// Extra cost (0) should give points
	extraCost := Default()
	extraCost.ResearchEnergy = 0
	extraCost.ResearchWeapons = 0
	extraCost.ResearchPropulsion = 0
	extraCost.ResearchConstruction = 0
	extraCost.ResearchElectronics = 0
	extraCost.ResearchBiotech = 0
	extraPoints := CalculatePoints(extraCost)

	// Less cost (2) should cost points
	lessCost := Default()
	lessCost.ResearchEnergy = 2
	lessCost.ResearchWeapons = 2
	lessCost.ResearchPropulsion = 2
	lessCost.ResearchConstruction = 2
	lessCost.ResearchElectronics = 2
	lessCost.ResearchBiotech = 2
	lessPoints := CalculatePoints(lessCost)

	if extraPoints <= lessPoints {
		t.Errorf("Extra research cost (%d) should give more points than less cost (%d)", extraPoints, lessPoints)
	}
}

func TestCalculatePointsTechsStartHigh(t *testing.T) {
	base := Default()
	basePoints := CalculatePoints(base)

	withTechStart := Default()
	withTechStart.TechsStartHigh = true
	techStartPoints := CalculatePoints(withTechStart)

	// TechsStartHigh costs 180 points (60 after /3)
	if techStartPoints >= basePoints {
		t.Errorf("Techs start high (%d) should cost points vs base (%d)", techStartPoints, basePoints)
	}
}

func TestCalculatePointsEconomy(t *testing.T) {
	// Better factory efficiency should cost points
	efficientFactories := Default()
	efficientFactories.FactoryOutput = 15
	efficientFactories.FactoryCost = 5
	efficientFactories.FactoryCount = 25
	efficientPoints := CalculatePoints(efficientFactories)

	inefficientFactories := Default()
	inefficientFactories.FactoryOutput = 5
	inefficientFactories.FactoryCost = 25
	inefficientFactories.FactoryCount = 5
	inefficientPoints := CalculatePoints(inefficientFactories)

	if efficientPoints >= inefficientPoints {
		t.Errorf("Efficient factories (%d) should cost more than inefficient (%d)", efficientPoints, inefficientPoints)
	}
}

func TestCalculatePointsColonistsPerResource(t *testing.T) {
	// Fewer colonists per resource = more efficient = costs points
	efficient := Default()
	efficient.ColonistsPerResource = 700
	efficientPoints := CalculatePoints(efficient)

	inefficient := Default()
	inefficient.ColonistsPerResource = 2500
	inefficientPoints := CalculatePoints(inefficient)

	if efficientPoints >= inefficientPoints {
		t.Errorf("Efficient pop (%d) should cost more than inefficient (%d)", efficientPoints, inefficientPoints)
	}
}

func TestCalculatePointsARSpecialCase(t *testing.T) {
	// AR (PRT 8) has special handling
	ar := Default()
	ar.PRT = 8
	arPoints := CalculatePoints(ar)

	// AR with cheap energy costs extra 100 points (33 after /3)
	arCheapEnergy := Default()
	arCheapEnergy.PRT = 8
	arCheapEnergy.ResearchEnergy = 2 // Less cost
	arCheapEnergyPoints := CalculatePoints(arCheapEnergy)

	if arCheapEnergyPoints >= arPoints {
		t.Errorf("AR with cheap energy (%d) should cost more than plain AR (%d)", arCheapEnergyPoints, arPoints)
	}
}

func TestGetPlanetHabitability(t *testing.T) {
	r := Default()

	// Perfect planet (center of all hab ranges)
	perfectHab := getPlanetHabitability(r, 50, 50, 50)
	if perfectHab <= 0 {
		t.Errorf("Perfect planet should have positive habitability, got %d", perfectHab)
	}

	// Planet at edge of hab range
	edgeHab := getPlanetHabitability(r, 25, 50, 50) // At low edge of gravity range
	if edgeHab >= perfectHab {
		t.Errorf("Edge planet (%d) should have lower hab than perfect (%d)", edgeHab, perfectHab)
	}

	// Planet outside hab range
	outsideHab := getPlanetHabitability(r, 0, 50, 50) // Outside gravity range
	if outsideHab >= 0 {
		t.Errorf("Planet outside hab range should have negative habitability, got %d", outsideHab)
	}
}

func TestGetPlanetHabitabilityImmune(t *testing.T) {
	r := Default()
	r.GravityImmune = true

	// With gravity immunity, gravity value shouldn't matter
	hab1 := getPlanetHabitability(r, 0, 50, 50)
	hab2 := getPlanetHabitability(r, 100, 50, 50)

	if hab1 != hab2 {
		t.Errorf("With gravity immunity, gravity shouldn't affect hab: %d vs %d", hab1, hab2)
	}
}

func TestLRTBalancePenalty(t *testing.T) {
	// Test that having too many LRTs causes balance penalties
	base := Default()
	basePoints := CalculatePoints(base)

	// Multiple disadvantage LRTs (positive costs) should give points
	// NRSE (index 7, 0x80, +160), CE (index 8, 0x100, +240), OBRM (index 9, 0x200, +255)
	manyBadLRTs := Default()
	manyBadLRTs.LRT = 0x0080 | 0x0100 | 0x0200 // NRSE, CE, OBRM
	manyBadPoints := CalculatePoints(manyBadLRTs)

	// These disadvantage LRTs should INCREASE points (give more advantage points to spend)
	if manyBadPoints <= basePoints {
		t.Errorf("Multiple disadvantage LRTs (%d) should increase points vs base (%d)", manyBadPoints, basePoints)
	}

	// Test that powerful LRTs cost points
	// IFE (index 0, 0x01, -235), TT (index 1, 0x02, -25), ISB (index 3, 0x08, -201)
	manyGoodLRTs := Default()
	manyGoodLRTs.LRT = 0x0001 | 0x0002 | 0x0008 // IFE, TT, ISB (all negative costs = advantages)
	manyGoodPoints := CalculatePoints(manyGoodLRTs)

	// Powerful advantages should DECREASE points
	if manyGoodPoints >= basePoints {
		t.Errorf("Multiple advantage LRTs (%d) should decrease points vs base (%d)", manyGoodPoints, basePoints)
	}
}
