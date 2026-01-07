package store

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/neper-stars/houston/blocks"
)

func TestArmorRemainingHP(t *testing.T) {
	tests := []struct {
		name      string
		dv        blocks.DV
		baseArmor int
		shipCount int
		wantArmor int
	}{
		{
			name:      "undamaged single ship",
			dv:        0,
			baseArmor: 100,
			shipCount: 1,
			wantArmor: 100,
		},
		{
			name:      "undamaged fleet of 5",
			dv:        0,
			baseArmor: 100,
			shipCount: 5,
			wantArmor: 500,
		},
		{
			name:      "50% damage single ship",
			dv:        blocks.NewDV(50.0, 100), // 50% armor damage, 100% ships affected
			baseArmor: 100,
			shipCount: 1,
			wantArmor: 50,
		},
		{
			name:      "99.8% damage single ship (nearly dead)",
			dv:        blocks.NewDV(99.8, 100), // 99.8% armor damage, 100% ships affected
			baseArmor: 100,
			shipCount: 1,
			wantArmor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ArmorRemainingHP(tt.dv, tt.baseArmor, tt.shipCount)
			// Allow some variance due to integer math
			assert.InDelta(t, tt.wantArmor, got, 5, "armor remaining should be approximately %d", tt.wantArmor)
		})
	}
}

func TestArmorDamageDealt(t *testing.T) {
	tests := []struct {
		name            string
		dvBefore        blocks.DV
		dvAfter         blocks.DV
		baseArmor       int
		shipCountBefore int
		shipCountAfter  int
		wantDamage      int
	}{
		{
			name:            "no damage dealt",
			dvBefore:        0,
			dvAfter:         0,
			baseArmor:       100,
			shipCountBefore: 1,
			shipCountAfter:  1,
			wantDamage:      0,
		},
		{
			name:            "damage to previously undamaged ship",
			dvBefore:        0,
			dvAfter:         blocks.NewDV(10.0, 100), // 10% armor damage
			baseArmor:       100,
			shipCountBefore: 1,
			shipCountAfter:  1,
			wantDamage:      10,
		},
		{
			name:            "kill shot - remaining armor consumed",
			dvBefore:        blocks.NewDV(90.0, 100), // 90% damage (10 HP left)
			dvAfter:         blocks.NewDV(99.8, 100), // nearly dead
			baseArmor:       100,
			shipCountBefore: 1,
			shipCountAfter:  0,  // ship destroyed
			wantDamage:      10, // final 10 HP
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ArmorDamageDealt(tt.dvBefore, tt.dvAfter, tt.baseArmor, tt.shipCountBefore, tt.shipCountAfter)
			// Allow some variance due to integer math
			assert.InDelta(t, tt.wantDamage, got, 5, "damage dealt should be approximately %d", tt.wantDamage)
		})
	}
}

func TestDVEncoding(t *testing.T) {
	// Verify our understanding of DV encoding from reversing notes
	tests := []struct {
		raw   uint16
		pctSh int
		pctDp int
		pct   float64
	}{
		{868, 100, 6, 1.2},      // Phase 9 value from battle-02
		{15076, 100, 117, 23.4}, // Phase 12 value
		{63972, 100, 499, 99.8}, // Phase 61 value (nearly dead)
	}

	for _, tt := range tests {
		dv := blocks.DV(tt.raw)
		assert.Equal(t, tt.pctSh, dv.PctSh(), "pctSh for DV=%d", tt.raw)
		assert.Equal(t, tt.pctDp, dv.PctDp(), "pctDp for DV=%d", tt.raw)
		assert.InDelta(t, tt.pct, dv.ArmorDamagePercent(), 0.1, "percent for DV=%d", tt.raw)
	}
}

func TestBattleRecordCombatEvents(t *testing.T) {
	// Create a minimal battle block for testing
	bb := &blocks.BattleBlock{
		TotalStacks: 2,
		Stacks: []blocks.BattleStack{
			{OwnerPlayerID: 0, DesignID: 0, ShipCount: 1, DamageState: 0},
			{OwnerPlayerID: 1, DesignID: 0, ShipCount: 1, DamageState: 0},
		},
		Actions: []blocks.BattleAction{
			{
				Events: []blocks.BattleRecordEvent{{Round: 0, StackID: 0}},
				Kills: []blocks.KillRecord{
					{
						StackID:      1,
						ShipsKilled:  0,
						ShieldDamage: 10,
						TargetDV:     blocks.NewDV(10.0, 100), // 10% armor damage, 100% ships
					},
				},
			},
		},
	}

	br := NewBattleRecord(bb)

	// Set base armor manually for testing (normally from ResolveDesigns)
	br.Stacks[0].BaseArmor = 100
	br.Stacks[1].BaseArmor = 100

	events := br.CombatEvents()
	assert.Len(t, events, 1, "should have 1 combat event")

	event := events[0]
	assert.Equal(t, 2, event.Phase, "phase should be 2 (action 0 + 2)")
	assert.Equal(t, 0, event.Round, "round should be 0")
	assert.Equal(t, 0, event.AttackerID, "attacker should be stack 0")
	assert.Equal(t, 1, event.TargetID, "target should be stack 1")
	assert.Equal(t, 10, event.ShieldDamage, "shield damage should be 10")
	assert.Greater(t, event.ArmorDamage, 0, "armor damage should be calculated")
}
