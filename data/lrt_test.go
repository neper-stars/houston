package data

import (
	"testing"
)

func TestAllLRTsCount(t *testing.T) {
	if len(AllLRTs) != 14 {
		t.Errorf("Expected 14 LRTs, got %d", len(AllLRTs))
	}
}

func TestAllLRTsIndexes(t *testing.T) {
	for i, lrt := range AllLRTs {
		if lrt.Index != i {
			t.Errorf("LRT %s has Index %d but is at position %d", lrt.Code, lrt.Index, i)
		}
		expectedBitmask := uint16(1 << i)
		if lrt.Bitmask != expectedBitmask {
			t.Errorf("LRT %s has Bitmask %d, expected %d", lrt.Code, lrt.Bitmask, expectedBitmask)
		}
	}
}

func TestGetLRT(t *testing.T) {
	tests := []struct {
		index int
		code  string
	}{
		{0, "IFE"},
		{1, "TT"},
		{2, "ARM"},
		{3, "ISB"},
		{4, "GR"},
		{5, "UR"},
		{6, "MA"},
		{7, "NRSE"},
		{8, "CE"},
		{9, "OBRM"},
		{10, "NAS"},
		{11, "LSP"},
		{12, "BET"},
		{13, "RS"},
	}

	for _, tt := range tests {
		lrt := GetLRT(tt.index)
		if lrt == nil {
			t.Errorf("GetLRT(%d) returned nil", tt.index)
			continue
		}
		if lrt.Code != tt.code {
			t.Errorf("GetLRT(%d) returned code %s, expected %s", tt.index, lrt.Code, tt.code)
		}
	}

	// Test out of range
	if GetLRT(-1) != nil {
		t.Error("GetLRT(-1) should return nil")
	}
	if GetLRT(14) != nil {
		t.Error("GetLRT(14) should return nil")
	}
}

func TestGetLRTByCode(t *testing.T) {
	lrt := GetLRTByCode("IFE")
	if lrt == nil {
		t.Fatal("GetLRTByCode(IFE) returned nil")
	}
	if lrt.Index != 0 {
		t.Errorf("IFE should have index 0, got %d", lrt.Index)
	}
	if lrt.PointCost != -235 {
		t.Errorf("IFE should have point cost -235, got %d", lrt.PointCost)
	}

	// Test non-existent code
	if GetLRTByCode("XXX") != nil {
		t.Error("GetLRTByCode(XXX) should return nil")
	}
}

func TestGetLRTByBitmask(t *testing.T) {
	lrt := GetLRTByBitmask(1 << 6)
	if lrt == nil {
		t.Fatal("GetLRTByBitmask(1 << 6) returned nil")
	}
	if lrt.Code != "MA" {
		t.Errorf("Bitmask 1<<6 should be MA, got %s", lrt.Code)
	}

	// Test non-existent bitmask
	if GetLRTByBitmask(0) != nil {
		t.Error("GetLRTByBitmask(0) should return nil")
	}
}

func TestGetLRTsFromBitmask(t *testing.T) {
	// IFE (1) + MA (64) + RS (8192) = 8257
	bitmask := uint16(1 | (1 << 6) | (1 << 13))
	lrts := GetLRTsFromBitmask(bitmask)

	if len(lrts) != 3 {
		t.Errorf("Expected 3 LRTs, got %d", len(lrts))
	}

	codes := make(map[string]bool)
	for _, lrt := range lrts {
		codes[lrt.Code] = true
	}

	if !codes["IFE"] {
		t.Error("Expected IFE in results")
	}
	if !codes["MA"] {
		t.Error("Expected MA in results")
	}
	if !codes["RS"] {
		t.Error("Expected RS in results")
	}
}

func TestHasLRT(t *testing.T) {
	// IFE (1) + TT (2) = 3
	bitmask := uint16(3)

	if !HasLRT(bitmask, LRTIndexIFE) {
		t.Error("Bitmask 3 should have IFE")
	}
	if !HasLRT(bitmask, LRTIndexTT) {
		t.Error("Bitmask 3 should have TT")
	}
	if HasLRT(bitmask, LRTIndexARM) {
		t.Error("Bitmask 3 should not have ARM")
	}

	// Out of range
	if HasLRT(bitmask, -1) {
		t.Error("HasLRT with index -1 should return false")
	}
	if HasLRT(bitmask, 14) {
		t.Error("HasLRT with index 14 should return false")
	}
}

func TestLRTGoodBad(t *testing.T) {
	// Good LRTs (cost points, negative PointCost)
	goodLRTs := []string{"IFE", "TT", "ARM", "ISB", "UR", "MA"}
	for _, code := range goodLRTs {
		lrt := GetLRTByCode(code)
		if lrt == nil {
			t.Errorf("LRT %s not found", code)
			continue
		}
		if !lrt.IsGoodLRT() {
			t.Errorf("LRT %s should be good (PointCost=%d)", code, lrt.PointCost)
		}
		if lrt.IsBadLRT() {
			t.Errorf("LRT %s should not be bad", code)
		}
	}

	// Bad LRTs (grant points, positive PointCost)
	badLRTs := []string{"GR", "NRSE", "CE", "OBRM", "NAS", "LSP", "BET", "RS"}
	for _, code := range badLRTs {
		lrt := GetLRTByCode(code)
		if lrt == nil {
			t.Errorf("LRT %s not found", code)
			continue
		}
		if !lrt.IsBadLRT() {
			t.Errorf("LRT %s should be bad (PointCost=%d)", code, lrt.PointCost)
		}
		if lrt.IsGoodLRT() {
			t.Errorf("LRT %s should not be good", code)
		}
	}
}

func TestLRTSpecificAbilities(t *testing.T) {
	// IFE abilities
	ife := GetLRTByCode("IFE")
	if ife.FuelEfficiencyBonus != 0.15 {
		t.Errorf("IFE FuelEfficiencyBonus = %f, want 0.15", ife.FuelEfficiencyBonus)
	}
	if !ife.UnlocksFuelMizer {
		t.Error("IFE should unlock Fuel Mizer")
	}
	if !ife.UnlocksGalaxyScoop {
		t.Error("IFE should unlock Galaxy Scoop")
	}
	if ife.StartingTechPropulsion != 1 {
		t.Errorf("IFE StartingTechPropulsion = %d, want 1", ife.StartingTechPropulsion)
	}

	// TT abilities
	tt := GetLRTByCode("TT")
	if tt.MaxTerraformPercent != 30 {
		t.Errorf("TT MaxTerraformPercent = %d, want 30", tt.MaxTerraformPercent)
	}
	if tt.TerraformingCostModifier != 0.70 {
		t.Errorf("TT TerraformingCostModifier = %f, want 0.70", tt.TerraformingCostModifier)
	}

	// NAS abilities
	nas := GetLRTByCode("NAS")
	if !nas.NoAdvancedScanners {
		t.Error("NAS should have NoAdvancedScanners")
	}
	if nas.NormalScannerMultiplier != 2 {
		t.Errorf("NAS NormalScannerMultiplier = %d, want 2", nas.NormalScannerMultiplier)
	}

	// RS abilities
	rs := GetLRTByCode("RS")
	if rs.ShieldStrengthMultiplier != 1.40 {
		t.Errorf("RS ShieldStrengthMultiplier = %f, want 1.40", rs.ShieldStrengthMultiplier)
	}
	if rs.ShieldRegenPerRound != 0.10 {
		t.Errorf("RS ShieldRegenPerRound = %f, want 0.10", rs.ShieldRegenPerRound)
	}
	if rs.ArmorStrengthMultiplier != 0.50 {
		t.Errorf("RS ArmorStrengthMultiplier = %f, want 0.50", rs.ArmorStrengthMultiplier)
	}
}

func TestLRTPointCosts(t *testing.T) {
	expectedCosts := map[string]int{
		"IFE":  -235,
		"TT":   -25,
		"ARM":  -159,
		"ISB":  -201,
		"GR":   40,
		"UR":   -240,
		"MA":   -155,
		"NRSE": 160,
		"CE":   240,
		"OBRM": 255,
		"NAS":  325,
		"LSP":  180,
		"BET":  70,
		"RS":   30,
	}

	for code, expectedCost := range expectedCosts {
		lrt := GetLRTByCode(code)
		if lrt == nil {
			t.Errorf("LRT %s not found", code)
			continue
		}
		if lrt.PointCost != expectedCost {
			t.Errorf("LRT %s has PointCost %d, expected %d", code, lrt.PointCost, expectedCost)
		}
	}
}

func TestNASARIntrinsicScannerMultiplier(t *testing.T) {
	nas := GetLRTByCode("NAS")
	if nas == nil {
		t.Fatal("NAS LRT not found")
	}

	// Should be √2
	expected := 1.4142135623730951
	if nas.ARIntrinsicScannerMultiplier != expected {
		t.Errorf("NAS ARIntrinsicScannerMultiplier = %f, want %f", nas.ARIntrinsicScannerMultiplier, expected)
	}
}

func TestARNASScannerRange(t *testing.T) {
	// Test data from in-game measurements
	// Formula: floor(floor(sqrt(pop/10)) × √2)
	tests := []struct {
		population int64
		expected   int
	}{
		{1000, 14},
		{1100, 14},
		{1200, 14},
		{1300, 15},
		{1400, 15},
		{1500, 16},
		{1600, 16},
		{1800, 18},
		{25000, 70},
		{28700, 74},
		{33000, 80},
		{38000, 86},
		{43700, 93},
		{50200, 98},
		{58800, 107},
		{896600, 422},
	}

	for _, tt := range tests {
		result := ARNASScannerRange(tt.population)
		if result != tt.expected {
			t.Errorf("ARNASScannerRange(%d) = %d, want %d", tt.population, result, tt.expected)
		}
	}
}

func TestARNASScannerRangeZeroPopulation(t *testing.T) {
	if ARNASScannerRange(0) != 0 {
		t.Error("ARNASScannerRange(0) should return 0")
	}
	if ARNASScannerRange(-100) != 0 {
		t.Error("ARNASScannerRange(-100) should return 0")
	}
}
