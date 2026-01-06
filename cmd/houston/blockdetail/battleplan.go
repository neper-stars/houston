package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.BattlePlanBlockType, FormatBattlePlan)
}

// FormatBattlePlan provides detailed view for BattlePlanBlock (type 30)
func FormatBattlePlan(block blocks.Block, index int) string {
	width := DefaultWidth
	bpb, ok := block.(blocks.BattlePlanBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := bpb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 4 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: Word 0 - Owner, PlanID, Tactic, DumpCargo
	word0 := encoding.Read16(d, 0)
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "Word0",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = 0x%04X", word0)))

	fields = append(fields, fmt.Sprintf("           %s bits0-3: OwnerPlayerId = %d -> Player %d",
		TreeBranch, bpb.OwnerPlayerId, bpb.OwnerPlayerId+1))
	fields = append(fields, fmt.Sprintf("           %s bits4-7: PlanId = %d",
		TreeBranch, bpb.PlanId))
	fields = append(fields, fmt.Sprintf("           %s bits8-11: Tactic = %d (%s)",
		TreeBranch, bpb.Tactic, bpb.TacticName()))
	fields = append(fields, fmt.Sprintf("           %s bits12-14: reserved",
		TreeBranch))
	fields = append(fields, fmt.Sprintf("           %s bit15: DumpCargo = %v",
		TreeEnd, bpb.DumpCargo))

	// Bytes 2-3: Word 1 - Targets and AttackWho
	word1 := encoding.Read16(d, 2)
	fields = append(fields, FormatFieldRaw(0x02, 0x03, "Word1",
		fmt.Sprintf("0x%02X%02X", d[3], d[2]),
		fmt.Sprintf("uint16 LE = 0x%04X", word1)))

	fields = append(fields, fmt.Sprintf("           %s bits0-3: PrimaryTarget = %d (%s)",
		TreeBranch, bpb.PrimaryTarget, bpb.PrimaryTargetName()))
	fields = append(fields, fmt.Sprintf("           %s bits4-7: SecondaryTarget = %d (%s)",
		TreeBranch, bpb.SecondaryTarget, bpb.SecondaryTargetName()))

	attackWhoStr := bpb.AttackWhoName()
	if bpb.TargetPlayer() >= 0 {
		attackWhoStr = fmt.Sprintf("Player %d", bpb.TargetPlayer()+1)
	}
	fields = append(fields, fmt.Sprintf("           %s bits8-15: AttackWho = %d (%s)",
		TreeEnd, bpb.AttackWho, attackWhoStr))

	// Check if deleted (4-byte block)
	if bpb.Deleted {
		fields = append(fields, "")
		fields = append(fields, "── Status ──")
		fields = append(fields, "  DELETED (4-byte block = plan deleted)")
	} else if len(d) > 4 {
		// Name section
		fields = append(fields, "")
		fields = append(fields, "── Plan Name ──")

		nameBytes := d[4:]
		if len(nameBytes) > 0 {
			// First byte is length
			nameLen := int(nameBytes[0])
			fields = append(fields, FormatFieldRaw(0x04, 0x04, "NameLength",
				fmt.Sprintf("0x%02X", nameBytes[0]),
				fmt.Sprintf("%d bytes", nameLen)))

			if len(nameBytes) > 1 {
				endOffset := 4 + len(nameBytes) - 1
				fields = append(fields, FormatFieldRaw(0x05, endOffset, "NameData",
					fmt.Sprintf("0x%s", HexDumpSingleLine(nameBytes[1:])),
					fmt.Sprintf("%q (Stars! encoded)", bpb.Name)))
			}
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	if bpb.Deleted {
		fields = append(fields, fmt.Sprintf("  Player %d, Plan #%d: DELETED",
			bpb.OwnerPlayerId+1, bpb.PlanId))
	} else {
		fields = append(fields, fmt.Sprintf("  Player %d, Plan #%d: %q",
			bpb.OwnerPlayerId+1, bpb.PlanId, bpb.Name))
		fields = append(fields, fmt.Sprintf("  Tactic: %s", bpb.TacticName()))
		fields = append(fields, fmt.Sprintf("  Targets: %s / %s",
			bpb.PrimaryTargetName(), bpb.SecondaryTargetName()))
		fields = append(fields, fmt.Sprintf("  Attack: %s", attackWhoStr))
		if bpb.DumpCargo {
			fields = append(fields, "  Dump Cargo: Yes")
		}
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
