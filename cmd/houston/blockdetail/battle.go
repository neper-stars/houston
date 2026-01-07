package blockdetail

import (
	"fmt"
	"strings"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.BattleBlockType, FormatBattle)
}

// FormatBattle provides detailed view for BattleBlock (type 31)
func FormatBattle(block blocks.Block, index int) string {
	width := DefaultWidth
	bb, ok := block.(blocks.BattleBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := bb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 14 {
		fields = append(fields, "(block too short for battle header)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Header (14 bytes from BTLDATA structure)
	fields = append(fields, "── Battle Header (14 bytes, BTLDATA) ──")

	// Bytes 0-1: Battle ID (uint16)
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "BattleID",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = %d", bb.BattleID)))

	// Byte 2: Player count
	fields = append(fields, FormatFieldRaw(0x02, 0x02, "PlayerCount",
		fmt.Sprintf("0x%02X", d[2]),
		fmt.Sprintf("%d players involved", bb.PlayerCount)))

	// Byte 3: Total stacks
	fields = append(fields, FormatFieldRaw(0x03, 0x03, "TotalStacks",
		fmt.Sprintf("0x%02X", d[3]),
		fmt.Sprintf("%d stacks", bb.TotalStacks)))

	// Bytes 4-5: Player bitmask
	fields = append(fields, FormatFieldRaw(0x04, 0x05, "PlayerBitmask",
		fmt.Sprintf("0x%02X%02X", d[5], d[4]),
		fmt.Sprintf("0x%04X (bit N = player N involved)", bb.PlayerBitmask)))

	// Bytes 6-7: Recorded size
	fields = append(fields, FormatFieldRaw(0x06, 0x07, "RecordedSize",
		fmt.Sprintf("0x%02X%02X", d[7], d[6]),
		fmt.Sprintf("uint16 LE = %d bytes", bb.RecordedSize)))

	// Bytes 8-9: Planet ID
	planetStr := fmt.Sprintf("Planet #%d", bb.PlanetID+1)
	if bb.PlanetID == -1 {
		planetStr = "Deep space (no planet)"
	}
	fields = append(fields, FormatFieldRaw(0x08, 0x09, "PlanetID",
		fmt.Sprintf("0x%02X%02X", d[9], d[8]),
		fmt.Sprintf("int16 LE = %d -> %s", bb.PlanetID, planetStr)))

	// Bytes 10-11: X coordinate
	fields = append(fields, FormatFieldRaw(0x0A, 0x0B, "X",
		fmt.Sprintf("0x%02X%02X", d[11], d[10]),
		fmt.Sprintf("uint16 LE = %d", bb.X)))

	// Bytes 12-13: Y coordinate
	fields = append(fields, FormatFieldRaw(0x0C, 0x0D, "Y",
		fmt.Sprintf("0x%02X%02X", d[13], d[12]),
		fmt.Sprintf("uint16 LE = %d", bb.Y)))

	// Stack definitions (TOK structures)
	if len(bb.Stacks) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Stacks (%d total, TOK 29 bytes each) ──", len(bb.Stacks)))

		// Group stacks by player
		stacksByPlayer := bb.StackCountByPlayer()
		var playerInfo []string
		for playerID, count := range stacksByPlayer {
			playerInfo = append(playerInfo, fmt.Sprintf("Player %d: %d", playerID, count))
		}
		fields = append(fields, fmt.Sprintf("  By player: %s", strings.Join(playerInfo, ", ")))
		fields = append(fields, "")

		for i, stack := range bb.Stacks {
			offset := 14 + i*29
			objType := "Fleet"
			if stack.IsStarbase {
				objType = "Starbase"
			}
			prefix := TreeBranch
			if i == len(bb.Stacks)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s Stack %d @ 0x%02X: Player %d, %s, Design #%d, %d ship(s)",
				prefix, i, offset, stack.OwnerPlayerID, objType, stack.DesignID, stack.ShipCount))
			fields = append(fields, fmt.Sprintf("      Position: %d, Init: %d-%d (base %d)",
				stack.GridPosition, stack.InitiativeMin, stack.InitiativeMax, stack.InitiativeBase))
			dmgState := "undamaged"
			if !stack.DamageState.IsZero() {
				dmgState = fmt.Sprintf("%.1f%% armor (%d%% ships)",
					stack.DamageState.ArmorDamagePercent(), stack.DamageState.PctSh())
			}
			fields = append(fields, fmt.Sprintf("      Shields: %d HP, State: %s, Mass: %d",
				stack.ShieldHP, dmgState, stack.Mass))
			if stack.CloakPercent > 0 || stack.JammerPercent > 0 || stack.BattleCompPercent > 0 || stack.BeamDeflectPercent > 0 {
				fields = append(fields, fmt.Sprintf("      Modifiers: Cloak %d%%, Jammer %d%%, BC %d%%, Cap %d%%, Deflect %d%%",
					stack.CloakPercent, stack.JammerPercent, stack.BattleCompPercent,
					stack.CapacitorPercent, stack.BeamDeflectPercent))
			}
		}
	}

	// Action records with kill details
	if len(bb.Actions) > 0 {
		fields = append(fields, "")
		actionStart := 14 + bb.TotalStacks*29
		fields = append(fields, fmt.Sprintf("── Actions (%d records, BTLREC) @ 0x%02X ──",
			len(bb.Actions), actionStart))
		fields = append(fields, fmt.Sprintf("  Total phases: %d (actions + 1 setup phase)", bb.TotalPhases()))
		fields = append(fields, fmt.Sprintf("  Rounds: %d", bb.Rounds))
		fields = append(fields, "")

		// Show actions with firing/kill events
		fields = append(fields, "  Combat events:")
		combatCount := 0
		maxShow := 10
		for i, action := range bb.Actions {
			if len(action.Kills) > 0 || hasFireEvent(action) {
				if combatCount >= maxShow {
					remaining := countCombatActions(bb.Actions[i:])
					if remaining > 0 {
						fields = append(fields, fmt.Sprintf("    %s ... and %d more combat actions", TreeEnd, remaining))
					}
					break
				}
				lines := formatActionDetail(action, i, bb.TotalStacks, d, actionStart)
				fields = append(fields, lines...)
				combatCount++
			}
		}
		if combatCount == 0 {
			fields = append(fields, "    (no combat events detected)")
		}
	}

	// Casualties summary
	casualties := bb.CasualtiesByPlayer()
	if len(casualties) > 0 {
		fields = append(fields, "")
		fields = append(fields, "── Casualties ──")
		for playerID, count := range casualties {
			fields = append(fields, fmt.Sprintf("  Player %d: %d ship(s) destroyed", playerID, count))
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	if bb.PlanetID == -1 {
		fields = append(fields, fmt.Sprintf("  Location: Deep Space @ (%d, %d)", bb.X, bb.Y))
	} else {
		fields = append(fields, fmt.Sprintf("  Location: Planet #%d @ (%d, %d)", bb.PlanetID+1, bb.X, bb.Y))
	}
	fields = append(fields, fmt.Sprintf("  Players: %d involved (bitmask: 0x%04X)", bb.PlayerCount, bb.PlayerBitmask))
	fields = append(fields, fmt.Sprintf("  Stacks: %d total", bb.TotalStacks))
	fields = append(fields, fmt.Sprintf("  Rounds: %d, Phases: %d", bb.Rounds, bb.TotalPhases()))
	fields = append(fields, fmt.Sprintf("  Actions: %d records (%d bytes)", len(bb.Actions), len(bb.RawActionData())))

	// Notes about parsing
	fields = append(fields, "")
	fields = append(fields, "── Notes ──")
	fields = append(fields, "  Phase N = Action (N-2), since Phase 1 is setup")
	fields = append(fields, "  KILL record: shipsKilled, shieldDamage verified against VCR")
	fields = append(fields, "  DV field = target's damage STATE after attack (not dmg dealt)")
	fields = append(fields, "  DV: pctDp (bits 7-15) = armor dmg %, pctSh (bits 0-6) = % ships")

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// hasFireEvent checks if an action has a fire event
func hasFireEvent(action blocks.BattleAction) bool {
	for _, event := range action.Events {
		if event.ActionType == blocks.ActionFire {
			return true
		}
	}
	return false
}

// countCombatActions counts remaining actions with combat events
func countCombatActions(actions []blocks.BattleAction) int {
	count := 0
	for _, action := range actions {
		if len(action.Kills) > 0 || hasFireEvent(action) {
			count++
		}
	}
	return count
}

// formatActionDetail formats a single action with combat details
// Returns multiple lines as separate strings to add to fields slice
func formatActionDetail(action blocks.BattleAction, idx int, totalStacks int, _ []byte, _ int) []string {
	raw := action.RawData
	if len(raw) < 6 {
		return []string{fmt.Sprintf("    %s Action %d: (too short)", TreeBranch, idx)}
	}

	itok := int(raw[0])
	ctok := int(int16(encoding.Read16(raw, 2))) //nolint:gosec // intentional signed conversion
	bitfield := encoding.Read16(raw, 4)
	iRound := int(bitfield & 0x0F)
	itokAttack := int((bitfield >> 8) & 0xFF)

	phase := idx + 2 // Phase = Action + 2 (Phase 1 is setup)

	var lines []string

	// Describe the action
	actionDesc := "moves/waits"
	if itokAttack < totalStacks && itokAttack != itok {
		actionDesc = fmt.Sprintf("fires at Stack %d", itokAttack)
	}
	lines = append(lines, fmt.Sprintf("    %s Phase %d (R%d): Stack %d %s",
		TreeBranch, phase, iRound, itok, actionDesc))

	// Show kill records with DV state
	if ctok > 0 && len(raw) >= 6+ctok*8 {
		for k := 0; k < ctok; k++ {
			killOffset := 6 + k*8
			killTarget := int(raw[killOffset])
			weaponFlags := raw[killOffset+1]
			shipsKilled := int(encoding.Read16(raw, killOffset+2))
			shieldDmg := int(encoding.Read16(raw, killOffset+4))
			targetDV := blocks.DV(encoding.Read16(raw, killOffset+6))

			var dmgParts []string
			if shieldDmg > 0 {
				dmgParts = append(dmgParts, fmt.Sprintf("%d shield", shieldDmg))
			}
			if shipsKilled > 0 {
				dmgParts = append(dmgParts, fmt.Sprintf("%d killed", shipsKilled))
			}

			dmgStr := "hit"
			if len(dmgParts) > 0 {
				dmgStr = strings.Join(dmgParts, ", ")
			}

			// Show target's damage state after attack
			stateStr := fmt.Sprintf("-> %.1f%% armor", targetDV.ArmorDamagePercent())

			lines = append(lines, fmt.Sprintf("        -> Stack %d: %s %s (0x%02X)",
				killTarget, dmgStr, stateStr, weaponFlags))
		}
	}

	return lines
}
