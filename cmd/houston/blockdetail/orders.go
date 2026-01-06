package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.PlayersRelationChangeBlockType, FormatPlayersRelationChange)
	RegisterFormatter(blocks.SetFleetBattlePlanBlockType, FormatSetFleetBattlePlan)
	RegisterFormatter(blocks.RenameFleetBlockType, FormatRenameFleet)
}

// FormatPlayersRelationChange provides detailed view for PlayersRelationChangeBlock (type 38)
func FormatPlayersRelationChange(block blocks.Block, index int) string {
	width := DefaultWidth
	prc, ok := block.(blocks.PlayersRelationChangeBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := prc.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 2 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Byte 0: Relation type
	fields = append(fields, FormatFieldRaw(0x00, 0x00, "Relation",
		fmt.Sprintf("0x%02X", d[0]),
		fmt.Sprintf("%d = %s", prc.Relation, prc.RelationName())))

	// Byte 1: Target player
	fields = append(fields, FormatFieldRaw(0x01, 0x01, "TargetPlayer",
		fmt.Sprintf("0x%02X", d[1]),
		fmt.Sprintf("%d -> Player %d", prc.TargetPlayer, prc.TargetPlayer+1)))

	// Additional bytes if present
	if len(d) > 2 {
		fields = append(fields, "")
		fields = append(fields, "── Additional Data ──")
		for i := 2; i < len(d); i++ {
			fields = append(fields, FormatFieldRaw(i, i, fmt.Sprintf("Byte%d", i),
				fmt.Sprintf("0x%02X", d[i]),
				"TBD"))
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Set relation with Player %d to: %s",
		prc.TargetPlayer+1, prc.RelationName()))

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// FormatSetFleetBattlePlan provides detailed view for SetFleetBattlePlanBlock (type 42)
func FormatSetFleetBattlePlan(block blocks.Block, index int) string {
	width := DefaultWidth
	sfbp, ok := block.(blocks.SetFleetBattlePlanBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := sfbp.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 4 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: Fleet number (9 bits)
	fleetWord := encoding.Read16(d, 0)
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "FleetNumber",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = 0x%04X", fleetWord)))
	fields = append(fields, fmt.Sprintf("           %s bits 0-8: Fleet = %d -> Fleet #%d",
		TreeEnd, sfbp.FleetNumber, sfbp.FleetNumber+1))

	// Bytes 2-3: Battle plan index
	planWord := encoding.Read16(d, 2)
	planName := "Default"
	if sfbp.BattlePlanIndex > 0 {
		planName = fmt.Sprintf("Custom Plan #%d", sfbp.BattlePlanIndex)
	}
	fields = append(fields, FormatFieldRaw(0x02, 0x03, "BattlePlanIdx",
		fmt.Sprintf("0x%02X%02X", d[3], d[2]),
		fmt.Sprintf("uint16 LE = %d -> %s", planWord, planName)))

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Fleet #%d: Assign battle plan \"%s\"",
		sfbp.FleetNumber+1, planName))

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// FormatRenameFleet provides detailed view for RenameFleetBlock (type 44)
func FormatRenameFleet(block blocks.Block, index int) string {
	width := DefaultWidth
	rfb, ok := block.(blocks.RenameFleetBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := rfb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 5 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: Fleet number
	fleetWord := encoding.Read16(d, 0)
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "FleetNumber",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = %d -> Fleet #%d", fleetWord, rfb.FleetNumber+1)))

	// Bytes 2-3: Unknown (typically same as fleet number)
	unknownWord := encoding.Read16(d, 2)
	fields = append(fields, FormatFieldRaw(0x02, 0x03, "Unknown",
		fmt.Sprintf("0x%02X%02X", d[3], d[2]),
		fmt.Sprintf("uint16 LE = %d (typically matches fleet number)", unknownWord)))

	// Bytes 4+: Encoded name
	if len(d) > 4 {
		nameLen := int(d[4])
		fields = append(fields, "")
		fields = append(fields, "── Name Data ──")
		fields = append(fields, FormatFieldRaw(0x04, 0x04, "NameLength",
			fmt.Sprintf("0x%02X", d[4]),
			fmt.Sprintf("%d bytes", nameLen)))

		if len(d) > 5 && nameLen > 0 {
			endIdx := 5 + nameLen
			if endIdx > len(d) {
				endIdx = len(d)
			}
			fields = append(fields, FormatFieldRaw(0x05, endIdx-1, "NameData",
				fmt.Sprintf("0x%s", HexDumpSingleLine(d[5:endIdx])),
				fmt.Sprintf("%q (Stars! encoded)", rfb.NewName)))
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Rename Fleet #%d to: %q", rfb.FleetNumber+1, rfb.NewName))

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
