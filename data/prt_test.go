package data

import (
	"testing"
)

func TestAllPRTsCount(t *testing.T) {
	if len(AllPRTs) != 10 {
		t.Errorf("Expected 10 PRTs, got %d", len(AllPRTs))
	}
}

func TestAllPRTsIndexes(t *testing.T) {
	for i, prt := range AllPRTs {
		if prt.Index != i {
			t.Errorf("PRT %s has Index %d but is at position %d", prt.Code, prt.Index, i)
		}
	}
}

func TestGetPRT(t *testing.T) {
	tests := []struct {
		index int
		code  string
		name  string
	}{
		{0, "HE", "Hyper-Expansion"},
		{1, "SS", "Super Stealth"},
		{2, "WM", "War Monger"},
		{3, "CA", "Claim Adjuster"},
		{4, "IS", "Inner-Strength"},
		{5, "SD", "Space Demolition"},
		{6, "PP", "Packet Physics"},
		{7, "IT", "Interstellar Traveler"},
		{8, "AR", "Alternate Reality"},
		{9, "JOAT", "Jack of All Trades"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			prt := GetPRT(tt.index)
			if prt == nil {
				t.Fatalf("GetPRT(%d) returned nil", tt.index)
			}
			if prt.Code != tt.code {
				t.Errorf("GetPRT(%d) returned code %s, expected %s", tt.index, prt.Code, tt.code)
			}
			if prt.Name != tt.name {
				t.Errorf("GetPRT(%d) returned name %s, expected %s", tt.index, prt.Name, tt.name)
			}
		})
	}

	// Test out of range
	if GetPRT(-1) != nil {
		t.Error("GetPRT(-1) should return nil")
	}
	if GetPRT(10) != nil {
		t.Error("GetPRT(10) should return nil")
	}
	if GetPRT(100) != nil {
		t.Error("GetPRT(100) should return nil")
	}
}

func TestGetPRTByCode(t *testing.T) {
	codes := []string{"HE", "SS", "WM", "CA", "IS", "SD", "PP", "IT", "AR", "JOAT"}

	for i, code := range codes {
		t.Run(code, func(t *testing.T) {
			prt := GetPRTByCode(code)
			if prt == nil {
				t.Fatalf("GetPRTByCode(%s) returned nil", code)
			}
			if prt.Index != i {
				t.Errorf("GetPRTByCode(%s) returned index %d, expected %d", code, prt.Index, i)
			}
		})
	}

	// Test non-existent codes
	if GetPRTByCode("XXX") != nil {
		t.Error("GetPRTByCode(XXX) should return nil")
	}
	if GetPRTByCode("") != nil {
		t.Error("GetPRTByCode('') should return nil")
	}
	if GetPRTByCode("he") != nil {
		t.Error("GetPRTByCode('he') should return nil (case sensitive)")
	}
}

func TestPRTPointCosts(t *testing.T) {
	expectedCosts := map[string]int{
		"HE":   -40,
		"SS":   -95,
		"WM":   -45,
		"CA":   -10,
		"IS":   100,
		"SD":   150,
		"PP":   -120,
		"IT":   -180,
		"AR":   -90,
		"JOAT": 66,
	}

	for code, expectedCost := range expectedCosts {
		prt := GetPRTByCode(code)
		if prt == nil {
			t.Errorf("PRT %s not found", code)
			continue
		}
		if prt.PointCost != expectedCost {
			t.Errorf("PRT %s has PointCost %d, expected %d", code, prt.PointCost, expectedCost)
		}
	}
}

func TestHEAbilities(t *testing.T) {
	he := GetPRTByCode("HE")
	if he == nil {
		t.Fatal("HE PRT not found")
	}

	if he.GrowthRateModifier != 2.0 {
		t.Errorf("HE GrowthRateModifier = %f, want 2.0", he.GrowthRateModifier)
	}
	if he.MaxPopulationModifier != 0.5 {
		t.Errorf("HE MaxPopulationModifier = %f, want 0.5", he.MaxPopulationModifier)
	}
	if !he.CanLiveOnPlanets {
		t.Error("HE should be able to live on planets")
	}
}

func TestSSAbilities(t *testing.T) {
	ss := GetPRTByCode("SS")
	if ss == nil {
		t.Fatal("SS PRT not found")
	}

	if ss.IntrinsicCloakPercent != 0.75 {
		t.Errorf("SS IntrinsicCloakPercent = %f, want 0.75", ss.IntrinsicCloakPercent)
	}
	if ss.CargoAffectsCloak {
		t.Error("SS cargo should not affect cloak")
	}
	if ss.MineTravelBonus != 1 {
		t.Errorf("SS MineTravelBonus = %d, want 1", ss.MineTravelBonus)
	}
}

func TestWMAbilities(t *testing.T) {
	wm := GetPRTByCode("WM")
	if wm == nil {
		t.Fatal("WM PRT not found")
	}

	if wm.WeaponsCostModifier != 0.75 {
		t.Errorf("WM WeaponsCostModifier = %f, want 0.75", wm.WeaponsCostModifier)
	}
	if wm.CanBuildMineFields {
		t.Error("WM should not be able to build mine fields")
	}
	if wm.CanBuildAdvancedDefenses {
		t.Error("WM should not be able to build advanced defenses")
	}
	if wm.StartingTechWeapons != 6 {
		t.Errorf("WM StartingTechWeapons = %d, want 6", wm.StartingTechWeapons)
	}
	if wm.StartingTechEnergy != 1 {
		t.Errorf("WM StartingTechEnergy = %d, want 1", wm.StartingTechEnergy)
	}
	if wm.StartingTechPropulsion != 1 {
		t.Errorf("WM StartingTechPropulsion = %d, want 1", wm.StartingTechPropulsion)
	}
}

func TestCAAbilities(t *testing.T) {
	ca := GetPRTByCode("CA")
	if ca == nil {
		t.Fatal("CA PRT not found")
	}

	if !ca.FreeTerraforming {
		t.Error("CA should have free terraforming")
	}
	if ca.TerraformingImprovementChance != 0.10 {
		t.Errorf("CA TerraformingImprovementChance = %f, want 0.10", ca.TerraformingImprovementChance)
	}
	if !ca.PlanetsRevertWhenAbandoned {
		t.Error("CA planets should revert when abandoned")
	}
	if ca.StartingTechBiotech != 6 {
		t.Errorf("CA StartingTechBiotech = %d, want 6", ca.StartingTechBiotech)
	}
}

func TestISAbilities(t *testing.T) {
	is := GetPRTByCode("IS")
	if is == nil {
		t.Fatal("IS PRT not found")
	}

	if is.WeaponsCostModifier != 1.25 {
		t.Errorf("IS WeaponsCostModifier = %f, want 1.25", is.WeaponsCostModifier)
	}
	if is.DefensesCostModifier != 0.60 {
		t.Errorf("IS DefensesCostModifier = %f, want 0.60", is.DefensesCostModifier)
	}
	if is.CanBuildSmartBombs {
		t.Error("IS should not be able to build smart bombs")
	}
	if !is.ColonistsReproduceDuringTransport {
		t.Error("IS colonists should reproduce during transport")
	}
}

func TestSDAbilities(t *testing.T) {
	sd := GetPRTByCode("SD")
	if sd == nil {
		t.Fatal("SD PRT not found")
	}

	if sd.MineTravelBonus != 2 {
		t.Errorf("SD MineTravelBonus = %d, want 2", sd.MineTravelBonus)
	}
	if !sd.MineFieldsActAsScanners {
		t.Error("SD mine fields should act as scanners")
	}
	if !sd.CanRemoteDetonateMines {
		t.Error("SD should be able to remote detonate mines")
	}
	if sd.StartingTechPropulsion != 2 {
		t.Errorf("SD StartingTechPropulsion = %d, want 2", sd.StartingTechPropulsion)
	}
	if sd.StartingTechBiotech != 2 {
		t.Errorf("SD StartingTechBiotech = %d, want 2", sd.StartingTechBiotech)
	}
}

func TestPPAbilities(t *testing.T) {
	pp := GetPRTByCode("PP")
	if pp == nil {
		t.Fatal("PP PRT not found")
	}

	if !pp.PacketsHavePenScanner {
		t.Error("PP packets should have penetrating scanner")
	}
	if pp.MaxPacketWarp != 13 {
		t.Errorf("PP MaxPacketWarp = %d, want 13", pp.MaxPacketWarp)
	}
	if pp.StartingTechEnergy != 4 {
		t.Errorf("PP StartingTechEnergy = %d, want 4", pp.StartingTechEnergy)
	}
}

func TestITAbilities(t *testing.T) {
	it := GetPRTByCode("IT")
	if it == nil {
		t.Fatal("IT PRT not found")
	}

	if it.StarbaseCostModifier != 0.75 {
		t.Errorf("IT StarbaseCostModifier = %f, want 0.75", it.StarbaseCostModifier)
	}
	if !it.CanScanEnemyStargates {
		t.Error("IT should be able to scan enemy stargates")
	}
	if !it.StargateSafetyBonus {
		t.Error("IT should have stargate safety bonus")
	}
	if it.StartingTechPropulsion != 5 {
		t.Errorf("IT StartingTechPropulsion = %d, want 5", it.StartingTechPropulsion)
	}
	if it.StartingTechConstruction != 5 {
		t.Errorf("IT StartingTechConstruction = %d, want 5", it.StartingTechConstruction)
	}
}

func TestARAbilities(t *testing.T) {
	ar := GetPRTByCode("AR")
	if ar == nil {
		t.Fatal("AR PRT not found")
	}

	if ar.StarbaseCostModifier != 0.80 {
		t.Errorf("AR StarbaseCostModifier = %f, want 0.80", ar.StarbaseCostModifier)
	}
	if !ar.HasIntrinsicScanner {
		t.Error("AR should have intrinsic scanner")
	}
	if ar.IntrinsicScannerRangeFunc == nil {
		t.Error("AR IntrinsicScannerRangeFunc should not be nil")
	}
	if ar.CanLiveOnPlanets {
		t.Error("AR should not be able to live on planets")
	}
}

func TestJOATAbilities(t *testing.T) {
	joat := GetPRTByCode("JOAT")
	if joat == nil {
		t.Fatal("JOAT PRT not found")
	}

	if joat.MaxPopulationModifier != 1.20 {
		t.Errorf("JOAT MaxPopulationModifier = %f, want 1.20", joat.MaxPopulationModifier)
	}
	if !joat.HasFleetIntrinsicScanner {
		t.Error("JOAT should have fleet intrinsic scanner")
	}
	if joat.FleetIntrinsicScannerRangeFunc == nil {
		t.Error("JOAT FleetIntrinsicScannerRangeFunc should not be nil")
	}

	// Check starting tech levels (all 3)
	if joat.StartingTechEnergy != 3 {
		t.Errorf("JOAT StartingTechEnergy = %d, want 3", joat.StartingTechEnergy)
	}
	if joat.StartingTechWeapons != 3 {
		t.Errorf("JOAT StartingTechWeapons = %d, want 3", joat.StartingTechWeapons)
	}
	if joat.StartingTechPropulsion != 3 {
		t.Errorf("JOAT StartingTechPropulsion = %d, want 3", joat.StartingTechPropulsion)
	}
	if joat.StartingTechConstruction != 3 {
		t.Errorf("JOAT StartingTechConstruction = %d, want 3", joat.StartingTechConstruction)
	}
	if joat.StartingTechElectronics != 3 {
		t.Errorf("JOAT StartingTechElectronics = %d, want 3", joat.StartingTechElectronics)
	}
	if joat.StartingTechBiotech != 3 {
		t.Errorf("JOAT StartingTechBiotech = %d, want 3", joat.StartingTechBiotech)
	}

	// Check JOAT scanner hulls
	expectedHulls := []int{HullScout, HullFrigate, HullDestroyer}
	if len(joat.FleetIntrinsicScannerHulls) != len(expectedHulls) {
		t.Errorf("JOAT FleetIntrinsicScannerHulls has %d hulls, want %d",
			len(joat.FleetIntrinsicScannerHulls), len(expectedHulls))
	}
	for _, h := range expectedHulls {
		found := false
		for _, jh := range joat.FleetIntrinsicScannerHulls {
			if jh == h {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("JOAT FleetIntrinsicScannerHulls should contain hull %d", h)
		}
	}
}

func TestArIntrinsicScannerRange(t *testing.T) {
	// Formula: range = sqrt(population / 10)
	tests := []struct {
		population int64
		expected   int
	}{
		{0, 0},
		{-100, 0},
		{10, 1},
		{100, 3},
		{1000, 10},
		{10000, 31},
		{100000, 100},
		{250000, 158},
		{1000000, 316},
	}

	for _, tt := range tests {
		result := arIntrinsicScannerRange(tt.population)
		if result != tt.expected {
			t.Errorf("arIntrinsicScannerRange(%d) = %d, want %d", tt.population, result, tt.expected)
		}
	}
}

func TestJoatFleetIntrinsicScanner(t *testing.T) {
	// Formula: Normal = Electronics × 20, Pen = Electronics × 10
	// Minimum: 60 ly normal, 30 ly pen (equiv to Electronics 3)
	tests := []struct {
		electronics    int
		expectedNormal int
		expectedPen    int
	}{
		{0, 60, 30},  // Below minimum, use minimums
		{1, 60, 30},  // Below minimum
		{2, 60, 30},  // Below minimum
		{3, 60, 30},  // At minimum
		{4, 80, 40},  // Above minimum
		{5, 100, 50}, // Above minimum
		{10, 200, 100},
		{15, 300, 150},
		{26, 520, 260}, // Max electronics
	}

	for _, tt := range tests {
		result := joatFleetIntrinsicScanner(tt.electronics)
		if result.NormalRange != tt.expectedNormal {
			t.Errorf("joatFleetIntrinsicScanner(%d).NormalRange = %d, want %d",
				tt.electronics, result.NormalRange, tt.expectedNormal)
		}
		if result.PenetratingRange != tt.expectedPen {
			t.Errorf("joatFleetIntrinsicScanner(%d).PenetratingRange = %d, want %d",
				tt.electronics, result.PenetratingRange, tt.expectedPen)
		}
	}
}

func TestPRTIntrinsicScannerRange(t *testing.T) {
	// Test AR's IntrinsicScannerRange method
	ar := GetPRTByCode("AR")
	if ar == nil {
		t.Fatal("AR PRT not found")
	}

	// AR should return scanner range
	range1000 := ar.IntrinsicScannerRange(1000)
	if range1000 != 10 {
		t.Errorf("AR.IntrinsicScannerRange(1000) = %d, want 10", range1000)
	}

	range0 := ar.IntrinsicScannerRange(0)
	if range0 != 0 {
		t.Errorf("AR.IntrinsicScannerRange(0) = %d, want 0", range0)
	}

	// Non-AR PRTs should return 0
	he := GetPRTByCode("HE")
	if he == nil {
		t.Fatal("HE PRT not found")
	}
	heRange := he.IntrinsicScannerRange(1000)
	if heRange != 0 {
		t.Errorf("HE.IntrinsicScannerRange(1000) = %d, want 0", heRange)
	}
}

func TestPRTHasFleetIntrinsicScannerForHull(t *testing.T) {
	joat := GetPRTByCode("JOAT")
	if joat == nil {
		t.Fatal("JOAT PRT not found")
	}

	// JOAT should have scanners for Scout, Frigate, Destroyer
	if !joat.HasFleetIntrinsicScannerForHull(HullScout) {
		t.Error("JOAT should have intrinsic scanner for Scout")
	}
	if !joat.HasFleetIntrinsicScannerForHull(HullFrigate) {
		t.Error("JOAT should have intrinsic scanner for Frigate")
	}
	if !joat.HasFleetIntrinsicScannerForHull(HullDestroyer) {
		t.Error("JOAT should have intrinsic scanner for Destroyer")
	}

	// JOAT should NOT have scanners for other hulls
	if joat.HasFleetIntrinsicScannerForHull(HullCruiser) {
		t.Error("JOAT should not have intrinsic scanner for Cruiser")
	}
	if joat.HasFleetIntrinsicScannerForHull(HullBattleship) {
		t.Error("JOAT should not have intrinsic scanner for Battleship")
	}

	// Non-JOAT PRTs should not have fleet scanners
	he := GetPRTByCode("HE")
	if he == nil {
		t.Fatal("HE PRT not found")
	}
	if he.HasFleetIntrinsicScannerForHull(HullScout) {
		t.Error("HE should not have intrinsic scanner for any hull")
	}
}

func TestPRTFleetIntrinsicScannerRange(t *testing.T) {
	joat := GetPRTByCode("JOAT")
	if joat == nil {
		t.Fatal("JOAT PRT not found")
	}

	// JOAT with Electronics 5
	stats := joat.FleetIntrinsicScannerRange(5)
	if stats.NormalRange != 100 {
		t.Errorf("JOAT.FleetIntrinsicScannerRange(5).NormalRange = %d, want 100", stats.NormalRange)
	}
	if stats.PenetratingRange != 50 {
		t.Errorf("JOAT.FleetIntrinsicScannerRange(5).PenetratingRange = %d, want 50", stats.PenetratingRange)
	}

	// Non-JOAT should return zero stats
	he := GetPRTByCode("HE")
	if he == nil {
		t.Fatal("HE PRT not found")
	}
	heStats := he.FleetIntrinsicScannerRange(10)
	if heStats.NormalRange != 0 || heStats.PenetratingRange != 0 {
		t.Errorf("HE.FleetIntrinsicScannerRange should return zero stats, got %+v", heStats)
	}
}

func TestAllPRTsHaveRequiredFields(t *testing.T) {
	for _, prt := range AllPRTs {
		t.Run(prt.Code, func(t *testing.T) {
			if prt.Code == "" {
				t.Error("PRT Code should not be empty")
			}
			if prt.Name == "" {
				t.Error("PRT Name should not be empty")
			}
			if prt.Desc == "" {
				t.Error("PRT Desc should not be empty")
			}
			// GrowthRateModifier and MaxPopulationModifier should be set (default 1.0)
			if prt.GrowthRateModifier == 0 {
				t.Error("PRT GrowthRateModifier should not be 0")
			}
			if prt.MaxPopulationModifier == 0 {
				t.Error("PRT MaxPopulationModifier should not be 0")
			}
		})
	}
}

func TestAllPRTsDefaultAbilities(t *testing.T) {
	// Most PRTs should have these abilities enabled by default
	defaultEnabled := []string{"HE", "SS", "CA", "SD", "PP", "IT", "JOAT"}

	for _, code := range defaultEnabled {
		prt := GetPRTByCode(code)
		if prt == nil {
			t.Errorf("PRT %s not found", code)
			continue
		}
		t.Run(code, func(t *testing.T) {
			if !prt.CanBuildMineFields {
				t.Errorf("%s should be able to build mine fields", code)
			}
			if !prt.CanBuildAdvancedDefenses {
				t.Errorf("%s should be able to build advanced defenses", code)
			}
			if !prt.CanBuildSmartBombs {
				t.Errorf("%s should be able to build smart bombs", code)
			}
			if !prt.CanLiveOnPlanets {
				t.Errorf("%s should be able to live on planets", code)
			}
		})
	}
}

func TestCostModifiersDefaults(t *testing.T) {
	// PRTs without special cost modifiers should have 1.0
	standardCost := []string{"HE", "SS", "CA", "SD", "PP", "JOAT"}

	for _, code := range standardCost {
		prt := GetPRTByCode(code)
		if prt == nil {
			t.Errorf("PRT %s not found", code)
			continue
		}
		t.Run(code, func(t *testing.T) {
			if prt.WeaponsCostModifier != 1.0 {
				t.Errorf("%s WeaponsCostModifier = %f, want 1.0", code, prt.WeaponsCostModifier)
			}
			if prt.DefensesCostModifier != 1.0 {
				t.Errorf("%s DefensesCostModifier = %f, want 1.0", code, prt.DefensesCostModifier)
			}
		})
	}
}
