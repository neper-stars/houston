package store

import "github.com/neper-stars/houston/blocks"

// BattleRecord provides high-level analysis of a battle block with
// damage calculations and design resolution.
type BattleRecord struct {
	Block *blocks.BattleBlock

	// Resolved stack information (needs GameStore for design lookup)
	Stacks []*BattleStackInfo
}

// BattleStackInfo extends BattleStack with resolved design information.
type BattleStackInfo struct {
	blocks.BattleStack

	// Resolved from GameStore
	Design    *DesignEntity
	BaseArmor int // Total armor per ship from design
}

// CombatEvent represents a single damage event with full details.
type CombatEvent struct {
	Phase        int       // Phase number (1-based, as shown in VCR)
	Round        int       // Round number (0-indexed)
	AttackerID   int       // Attacking stack ID
	TargetID     int       // Target stack ID
	WeaponFlags  int       // Weapon type flags
	ShipsKilled  int       // Ships destroyed
	ShieldDamage int       // Shield damage dealt
	ArmorDamage  int       // Armor damage dealt (calculated from DV change)
	DVBefore     blocks.DV // Target's damage state before attack
	DVAfter      blocks.DV // Target's damage state after attack
}

// NewBattleRecord creates a BattleRecord from a BattleBlock.
// Call ResolveDesigns() to populate design information from a GameStore.
func NewBattleRecord(bb *blocks.BattleBlock) *BattleRecord {
	br := &BattleRecord{
		Block:  bb,
		Stacks: make([]*BattleStackInfo, len(bb.Stacks)),
	}

	for i, stack := range bb.Stacks {
		br.Stacks[i] = &BattleStackInfo{
			BattleStack: stack,
		}
	}

	return br
}

// ResolveDesigns looks up ship designs from the GameStore and calculates
// base armor values. This enables armor damage calculations.
func (br *BattleRecord) ResolveDesigns(gs *GameStore) {
	for _, stack := range br.Stacks {
		// Look up design by owner and design ID
		if design, ok := gs.Design(stack.OwnerPlayerID, stack.DesignID); ok {
			stack.Design = design
			stack.BaseArmor = design.GetTotalArmorValue()
		}
	}
}

// ArmorRemainingHP calculates remaining armor HP from a DV value.
// This implements the LdpFromItokDv formula from Stars! decompilation.
//
// Formula (from stars26jrc3.exe @ 10e8:07a8):
//
//	totalArmor = baseArmor * shipCount
//	if dv != 0:
//	    affectedShips = (shipCount * pctSh) / 100
//	    if affectedShips < 1: affectedShips = 1
//	    dmg = (baseArmor * pctDp) / 10 * affectedShips / 50
//	    totalArmor -= dmg
func ArmorRemainingHP(dv blocks.DV, baseArmor, shipCount int) int {
	totalArmor := baseArmor * shipCount

	if dv.IsZero() {
		return totalArmor
	}

	pctSh := dv.PctSh()
	pctDp := dv.PctDp()

	// Ships with distributed damage
	affectedShips := (shipCount * pctSh) / 100
	if affectedShips < 1 {
		affectedShips = 1
	}

	// Damage to subtract
	dmg := (baseArmor * pctDp) / 10 * affectedShips / 50
	totalArmor -= dmg

	if totalArmor < 0 {
		totalArmor = 0
	}

	return totalArmor
}

// ArmorDamageDealt calculates the armor damage dealt between two DV states.
// This is the difference in remaining armor HP before and after an attack.
func ArmorDamageDealt(dvBefore, dvAfter blocks.DV, baseArmor, shipCountBefore, shipCountAfter int) int {
	armorBefore := ArmorRemainingHP(dvBefore, baseArmor, shipCountBefore)
	armorAfter := ArmorRemainingHP(dvAfter, baseArmor, shipCountAfter)

	// Armor damage = what was lost (before - after)
	damage := armorBefore - armorAfter
	if damage < 0 {
		damage = 0
	}

	return damage
}

// CombatEvents extracts all combat events with calculated armor damage.
// This requires designs to be resolved first via ResolveDesigns().
func (br *BattleRecord) CombatEvents() []CombatEvent {
	if br.Block == nil {
		return nil
	}

	var events []CombatEvent

	// Track DV state for each stack through the battle
	dvState := make([]blocks.DV, len(br.Stacks))
	shipCount := make([]int, len(br.Stacks))

	// Initialize from stack starting states
	for i, stack := range br.Stacks {
		dvState[i] = stack.DamageState
		shipCount[i] = stack.ShipCount
	}

	// Process each action
	for actionIdx, action := range br.Block.Actions {
		phase := actionIdx + 2 // Phase = Action + 2 (Phase 1 is setup)

		// Get round from first event
		round := 0
		if len(action.Events) > 0 {
			round = action.Events[0].Round
		}

		// Get attacker from first event
		attackerID := -1
		if len(action.Events) > 0 {
			attackerID = action.Events[0].StackID
		}

		// Process kill records (damage events)
		for _, kill := range action.Kills {
			targetID := kill.StackID
			if targetID < 0 || targetID >= len(br.Stacks) {
				continue
			}

			dvBefore := dvState[targetID]
			dvAfter := kill.TargetDV
			shipsBefore := shipCount[targetID]
			shipsAfter := shipsBefore - kill.ShipsKilled

			// Calculate armor damage if we have design info
			armorDmg := 0
			if br.Stacks[targetID].BaseArmor > 0 {
				armorDmg = ArmorDamageDealt(dvBefore, dvAfter,
					br.Stacks[targetID].BaseArmor, shipsBefore, shipsAfter)
			}

			event := CombatEvent{
				Phase:        phase,
				Round:        round,
				AttackerID:   attackerID,
				TargetID:     targetID,
				WeaponFlags:  kill.WeaponFlags,
				ShipsKilled:  kill.ShipsKilled,
				ShieldDamage: kill.ShieldDamage,
				ArmorDamage:  armorDmg,
				DVBefore:     dvBefore,
				DVAfter:      dvAfter,
			}
			events = append(events, event)

			// Update tracking state
			dvState[targetID] = dvAfter
			shipCount[targetID] = shipsAfter
		}
	}

	return events
}

// TotalArmorDamageByPlayer calculates total armor damage dealt TO each player.
func (br *BattleRecord) TotalArmorDamageByPlayer() map[int]int {
	result := make(map[int]int)

	for _, event := range br.CombatEvents() {
		if event.TargetID >= 0 && event.TargetID < len(br.Stacks) {
			playerID := br.Stacks[event.TargetID].OwnerPlayerID
			result[playerID] += event.ArmorDamage
		}
	}

	return result
}

// TotalShieldDamageByPlayer calculates total shield damage dealt TO each player.
func (br *BattleRecord) TotalShieldDamageByPlayer() map[int]int {
	result := make(map[int]int)

	for _, event := range br.CombatEvents() {
		if event.TargetID >= 0 && event.TargetID < len(br.Stacks) {
			playerID := br.Stacks[event.TargetID].OwnerPlayerID
			result[playerID] += event.ShieldDamage
		}
	}

	return result
}
