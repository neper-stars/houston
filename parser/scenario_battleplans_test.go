package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neper-stars/houston/blocks"
)

// Expected data structures for battle plans scenario
type ExpectedBattlePlan struct {
	Name            string `json:"name"`
	PrimaryTarget   string `json:"primaryTarget"`
	SecondaryTarget string `json:"secondaryTarget"`
	Tactic          string `json:"tactic"`
	AttackWho       string `json:"attackWho"`
	DumpCargo       bool   `json:"dumpCargo"`
}

type ExpectedBattlePlanData struct {
	Scenario    string               `json:"scenario"`
	BattlePlans []ExpectedBattlePlan `json:"battlePlans"`
}

func loadBattlePlanExpected(t *testing.T) *ExpectedBattlePlanData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-battleplans", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedBattlePlanData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadBattlePlanFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-battleplans", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

func TestScenarioBattlePlans(t *testing.T) {
	expected := loadBattlePlanExpected(t)
	_, blockList := loadBattlePlanFile(t, "game.m2")

	// Collect battle plans
	var battlePlans []blocks.BattlePlanBlock
	for _, block := range blockList {
		if bp, ok := block.(blocks.BattlePlanBlock); ok {
			if !bp.Deleted {
				battlePlans = append(battlePlans, bp)
			}
		}
	}

	require.Equal(t, len(expected.BattlePlans), len(battlePlans),
		"Should have %d battle plans", len(expected.BattlePlans))

	// Create a map by name for easier lookup
	planByName := make(map[string]*blocks.BattlePlanBlock)
	for i := range battlePlans {
		planByName[battlePlans[i].Name] = &battlePlans[i]
	}

	// Validate each expected battle plan
	for _, exp := range expected.BattlePlans {
		t.Run(exp.Name, func(t *testing.T) {
			plan, found := planByName[exp.Name]
			require.True(t, found, "Battle plan '%s' should exist", exp.Name)

			assert.Equal(t, exp.Name, plan.Name, "Name should match")
			assert.Equal(t, exp.PrimaryTarget, plan.PrimaryTargetName(), "Primary target should match")
			assert.Equal(t, exp.SecondaryTarget, plan.SecondaryTargetName(), "Secondary target should match")
			assert.Equal(t, exp.Tactic, plan.TacticName(), "Tactic should match")
			assert.Equal(t, exp.AttackWho, plan.AttackWhoName(), "Attack who should match")
			assert.Equal(t, exp.DumpCargo, plan.DumpCargo, "Dump cargo should match")
		})
	}
}
