package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
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

	if len(d) < 18 {
		fields = append(fields, "(block too short for battle header)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Header (18 bytes)
	fields = append(fields, "── Battle Header (18 bytes) ──")

	// Byte 0: Battle ID
	fields = append(fields, FormatFieldRaw(0x00, 0x00, "BattleID",
		fmt.Sprintf("0x%02X", d[0]),
		fmt.Sprintf("%d", bb.BattleID)))

	// Byte 1: Rounds (stored as rounds-1)
	fields = append(fields, FormatFieldRaw(0x01, 0x01, "Rounds",
		fmt.Sprintf("0x%02X", d[1]),
		fmt.Sprintf("stored=%d -> actual=%d rounds", d[1], bb.Rounds)))

	// Byte 2: Side 1 stacks
	fields = append(fields, FormatFieldRaw(0x02, 0x02, "Side1Stacks",
		fmt.Sprintf("0x%02X", d[2]),
		fmt.Sprintf("%d (viewer's side)", bb.Side1Stacks)))

	// Byte 3: Total stacks
	fields = append(fields, FormatFieldRaw(0x03, 0x03, "TotalStacks",
		fmt.Sprintf("0x%02X", d[3]),
		fmt.Sprintf("%d (side2=%d)", bb.TotalStacks, bb.Side2Stacks())))

	// Bytes 4-5: Unknown
	fields = append(fields, FormatFieldRaw(0x04, 0x05, "Unknown",
		fmt.Sprintf("0x%02X%02X", d[5], d[4]),
		"TBD"))

	// Bytes 6-7: Recorded size
	fields = append(fields, FormatFieldRaw(0x06, 0x07, "RecordedSize",
		fmt.Sprintf("0x%02X%02X", d[7], d[6]),
		fmt.Sprintf("uint16 LE = %d bytes", bb.RecordedSize)))

	// Bytes 8-9: Planet ID
	fields = append(fields, FormatFieldRaw(0x08, 0x09, "PlanetID",
		fmt.Sprintf("0x%02X%02X", d[9], d[8]),
		fmt.Sprintf("uint16 LE = %d -> Planet #%d", bb.PlanetID, bb.PlanetID+1)))

	// Bytes 10-11: X coordinate
	fields = append(fields, FormatFieldRaw(0x0A, 0x0B, "X",
		fmt.Sprintf("0x%02X%02X", d[11], d[10]),
		fmt.Sprintf("uint16 LE = %d", bb.X)))

	// Bytes 12-13: Y coordinate
	fields = append(fields, FormatFieldRaw(0x0C, 0x0D, "Y",
		fmt.Sprintf("0x%02X%02X", d[13], d[12]),
		fmt.Sprintf("uint16 LE = %d", bb.Y)))

	// Byte 14: Attacker stacks
	fields = append(fields, FormatFieldRaw(0x0E, 0x0E, "AttackerStacks",
		fmt.Sprintf("0x%02X", d[14]),
		fmt.Sprintf("%d", bb.AttackerStacks)))

	// Byte 15: Defender stacks
	fields = append(fields, FormatFieldRaw(0x0F, 0x0F, "DefenderStacks",
		fmt.Sprintf("0x%02X", d[15]),
		fmt.Sprintf("%d", bb.DefenderStacks)))

	// Byte 16: Attacker losses
	fields = append(fields, FormatFieldRaw(0x10, 0x10, "AttackerLosses",
		fmt.Sprintf("0x%02X", d[16]),
		fmt.Sprintf("%d ships", bb.AttackerLosses)))

	// Byte 17: Unknown
	fields = append(fields, FormatFieldRaw(0x11, 0x11, "Unknown17",
		fmt.Sprintf("0x%02X", d[17]),
		fmt.Sprintf("%d (TBD)", bb.Unknown17)))

	// Stack definitions
	if len(bb.Stacks) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Stack Definitions (%d stacks, 29 bytes each) ──", len(bb.Stacks)))

		for i, stack := range bb.Stacks {
			offset := 18 + i*29
			side := "Side1"
			if stack.OwnerID == 1 {
				side = "Side2"
			}
			prefix := TreeBranch
			if i == len(bb.Stacks)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s Stack %d @ 0x%02X: %s, Design #%d, %d ships",
				prefix, i, offset, side, stack.DesignID, stack.ShipCount))
		}
	}

	// Action records
	if len(bb.Actions) > 0 {
		fields = append(fields, "")
		actionStart := 18 + bb.TotalStacks*29
		fields = append(fields, fmt.Sprintf("── Action Records (%d records, 22 bytes each) @ 0x%02X ──",
			len(bb.Actions), actionStart))

		// Show first few actions
		maxShow := 5
		if len(bb.Actions) < maxShow {
			maxShow = len(bb.Actions)
		}
		for i := 0; i < maxShow; i++ {
			action := bb.Actions[i]
			offset := actionStart + i*22
			prefix := TreeBranch
			if i == maxShow-1 && maxShow == len(bb.Actions) {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s Action %d @ 0x%02X: %d events",
				prefix, i, offset, len(action.Events)))
		}
		if len(bb.Actions) > maxShow {
			fields = append(fields, fmt.Sprintf("  %s ... and %d more action records", TreeEnd, len(bb.Actions)-maxShow))
		}
	}

	// Battle phases (decoded)
	phases := bb.Phases()
	if len(phases) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Detected Phases (%d phases) ──", len(phases)))

		maxShow := 8
		if len(phases) < maxShow {
			maxShow = len(phases)
		}
		for i := 0; i < maxShow; i++ {
			phase := phases[i]
			prefix := TreeBranch
			if i == maxShow-1 && maxShow == len(phases) {
				prefix = TreeEnd
			}

			action := "waits"
			if phase.Fired {
				if phase.Damage > 0 {
					action = fmt.Sprintf("fires at Stack %d for %d damage", phase.TargetID, phase.Damage)
				} else {
					action = "fires (no damage)"
				}
			}
			fields = append(fields, fmt.Sprintf("  %s Phase %d (R%d): Stack %d at (%d,%d) %s",
				prefix, phase.PhaseNum, phase.Round, phase.StackID, phase.GridX, phase.GridY, action))
		}
		if len(phases) > maxShow {
			fields = append(fields, fmt.Sprintf("  %s ... and %d more phases", TreeEnd, len(phases)-maxShow))
		}
	}

	// Warnings
	fields = append(fields, "")
	fields = append(fields, "── WARNING ──")
	fields = append(fields, "  Battle recording parsing is INCOMPLETE and INACCURATE!")
	fields = append(fields, "  Known limitations:")
	fields = append(fields, "    - Detects ~60% of phases shown in Battle VCR")
	fields = append(fields, "    - Early rounds (0-1) have lower detection rates")
	fields = append(fields, "    - Some phases use unknown encoding")
	fields = append(fields, "    - Stack/damage data may be incorrect")
	fields = append(fields, "  Use RawActionData() for custom analysis if needed.")

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Location: Planet #%d @ (%d, %d)", bb.PlanetID+1, bb.X, bb.Y))
	fields = append(fields, fmt.Sprintf("  Rounds: %d", bb.Rounds))
	fields = append(fields, fmt.Sprintf("  Forces: %d vs %d stacks", bb.Side1Stacks, bb.Side2Stacks()))
	fields = append(fields, fmt.Sprintf("  Attacker losses: %d ships", bb.AttackerLosses))
	fields = append(fields, fmt.Sprintf("  Total action data: %d bytes", len(bb.RawActionData())))

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
