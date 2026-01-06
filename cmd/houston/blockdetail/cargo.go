package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.ManualSmallLoadUnloadTaskBlockType, FormatManualSmallLoadUnload)
	RegisterFormatter(blocks.ManualMediumLoadUnloadTaskBlockType, FormatManualMediumLoadUnload)
	RegisterFormatter(blocks.ManualLargeLoadUnloadTaskBlockType, FormatManualLargeLoadUnload)
}

// FormatManualSmallLoadUnload provides detailed view for ManualSmallLoadUnloadTaskBlock (type 1)
func FormatManualSmallLoadUnload(block blocks.Block, index int) string {
	width := DefaultWidth
	cb, ok := block.(blocks.ManualSmallLoadUnloadTaskBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := cb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 10 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: Fleet number
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "FleetNumber",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = %d -> Fleet #%d", cb.FleetNumber, cb.FleetNumber+1)))

	// Bytes 2-3: Target number
	fields = append(fields, FormatFieldRaw(0x02, 0x03, "TargetNumber",
		fmt.Sprintf("0x%02X%02X", d[3], d[2]),
		fmt.Sprintf("uint16 LE = %d -> Target #%d", cb.TargetNumber, cb.TargetNumber+1)))

	// Byte 4: Task/flags byte
	direction := "Unload (fleet -> target)"
	if cb.IsLoad() {
		direction = "Load (target -> fleet)"
	}
	fields = append(fields, FormatFieldRaw(0x04, 0x04, "TaskByte",
		fmt.Sprintf("0x%02X", d[4]),
		fmt.Sprintf("0b%08b", d[4])))
	fields = append(fields, fmt.Sprintf("           %s bit4 (0x10): %s", TreeEnd, direction))

	// Byte 5: Cargo mask
	fields = append(fields, FormatFieldRaw(0x05, 0x05, "CargoMask",
		fmt.Sprintf("0x%02X", d[5]),
		fmt.Sprintf("0b%04b", d[5]&0x0F)))
	fields = append(fields, fmt.Sprintf("           %s bit0: Ironium = %v", TreeBranch, cb.HasIronium()))
	fields = append(fields, fmt.Sprintf("           %s bit1: Boranium = %v", TreeBranch, cb.HasBoranium()))
	fields = append(fields, fmt.Sprintf("           %s bit2: Germanium = %v", TreeBranch, cb.HasGermanium()))
	fields = append(fields, fmt.Sprintf("           %s bit3: Colonists = %v", TreeEnd, cb.HasColonists()))

	// Bytes 6-9: Cargo amounts (signed bytes)
	fields = append(fields, "")
	fields = append(fields, "── Cargo Amounts (signed bytes, -128 to 127 kT) ──")
	fields = append(fields, FormatFieldRaw(0x06, 0x06, "Ironium",
		fmt.Sprintf("0x%02X", d[6]),
		fmt.Sprintf("int8 = %d kT", cb.Ironium)))
	fields = append(fields, FormatFieldRaw(0x07, 0x07, "Boranium",
		fmt.Sprintf("0x%02X", d[7]),
		fmt.Sprintf("int8 = %d kT", cb.Boranium)))
	fields = append(fields, FormatFieldRaw(0x08, 0x08, "Germanium",
		fmt.Sprintf("0x%02X", d[8]),
		fmt.Sprintf("int8 = %d kT", cb.Germanium)))
	fields = append(fields, FormatFieldRaw(0x09, 0x09, "Colonists",
		fmt.Sprintf("0x%02X", d[9]),
		fmt.Sprintf("int8 = %d (×100 colonists)", cb.Colonists)))

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Fleet #%d -> Target #%d", cb.FleetNumber+1, cb.TargetNumber+1))
	fields = append(fields, fmt.Sprintf("  Direction: %s", direction))
	if cb.HasIronium() {
		fields = append(fields, fmt.Sprintf("  %s Ironium: %d kT", TreeBranch, cb.Ironium))
	}
	if cb.HasBoranium() {
		fields = append(fields, fmt.Sprintf("  %s Boranium: %d kT", TreeBranch, cb.Boranium))
	}
	if cb.HasGermanium() {
		fields = append(fields, fmt.Sprintf("  %s Germanium: %d kT", TreeBranch, cb.Germanium))
	}
	if cb.HasColonists() {
		fields = append(fields, fmt.Sprintf("  %s Colonists: %d00", TreeEnd, cb.Colonists))
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// FormatManualMediumLoadUnload provides detailed view for ManualMediumLoadUnloadTaskBlock (type 2)
func FormatManualMediumLoadUnload(block blocks.Block, index int) string {
	width := DefaultWidth
	cb, ok := block.(blocks.ManualMediumLoadUnloadTaskBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := cb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 4 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Structure not fully documented - show what we can infer
	fields = append(fields, "── Raw Data (structure partially documented) ──")

	// Bytes 0-1: Fleet number (likely same format as small)
	fleetNum := int(encoding.Read16(d, 0))
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "FleetNumber",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = %d -> Fleet #%d", fleetNum, fleetNum+1)))

	// Bytes 2-3: Target number (likely same format as small)
	if len(d) >= 4 {
		targetNum := int(encoding.Read16(d, 2))
		fields = append(fields, FormatFieldRaw(0x02, 0x03, "TargetNumber",
			fmt.Sprintf("0x%02X%02X", d[3], d[2]),
			fmt.Sprintf("uint16 LE = %d -> Target #%d", targetNum, targetNum+1)))
	}

	// Remaining bytes - show raw with annotations
	if len(d) > 4 {
		fields = append(fields, "")
		fields = append(fields, "── Additional Data ──")
		// Byte 4: likely task/flags
		if len(d) > 4 {
			fields = append(fields, FormatFieldRaw(0x04, 0x04, "TaskByte",
				fmt.Sprintf("0x%02X", d[4]),
				"likely flags/direction"))
		}
		// Byte 5: likely cargo mask
		if len(d) > 5 {
			fields = append(fields, FormatFieldRaw(0x05, 0x05, "CargoMask",
				fmt.Sprintf("0x%02X", d[5]),
				"likely cargo type bitmask"))
		}
		// Remaining bytes: cargo amounts (likely int16 each)
		for i := 6; i+1 < len(d); i += 2 {
			val := int(int16(encoding.Read16(d, i))) //nolint:gosec // intentional signed conversion for cargo values
			cargoName := "Unknown"
			switch (i - 6) / 2 {
			case 0:
				cargoName = "Ironium"
			case 1:
				cargoName = "Boranium"
			case 2:
				cargoName = "Germanium"
			case 3:
				cargoName = "Colonists"
			}
			fields = append(fields, FormatFieldRaw(i, i+1, cargoName,
				fmt.Sprintf("0x%02X%02X", d[i+1], d[i]),
				fmt.Sprintf("int16 LE = %d kT", val)))
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Info ──")
	fields = append(fields, "  Purpose: Medium cargo transfer (int16 amounts)")
	fields = append(fields, "  Structure: Partially documented")
	fields = append(fields, fmt.Sprintf("  Size: %d bytes", len(d)))

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// FormatManualLargeLoadUnload provides detailed view for ManualLargeLoadUnloadTaskBlock (type 25)
func FormatManualLargeLoadUnload(block blocks.Block, index int) string {
	width := DefaultWidth
	cb, ok := block.(blocks.ManualLargeLoadUnloadTaskBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := cb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 4 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Structure not fully documented - show what we can infer
	fields = append(fields, "── Raw Data (structure partially documented) ──")

	// Bytes 0-1: Fleet number (likely same format as small)
	fleetNum := int(encoding.Read16(d, 0))
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "FleetNumber",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = %d -> Fleet #%d", fleetNum, fleetNum+1)))

	// Bytes 2-3: Target number (likely same format as small)
	if len(d) >= 4 {
		targetNum := int(encoding.Read16(d, 2))
		fields = append(fields, FormatFieldRaw(0x02, 0x03, "TargetNumber",
			fmt.Sprintf("0x%02X%02X", d[3], d[2]),
			fmt.Sprintf("uint16 LE = %d -> Target #%d", targetNum, targetNum+1)))
	}

	// Remaining bytes - show raw with annotations
	if len(d) > 4 {
		fields = append(fields, "")
		fields = append(fields, "── Additional Data ──")
		// Byte 4: likely task/flags
		if len(d) > 4 {
			fields = append(fields, FormatFieldRaw(0x04, 0x04, "TaskByte",
				fmt.Sprintf("0x%02X", d[4]),
				"likely flags/direction"))
		}
		// Byte 5: likely cargo mask
		if len(d) > 5 {
			fields = append(fields, FormatFieldRaw(0x05, 0x05, "CargoMask",
				fmt.Sprintf("0x%02X", d[5]),
				"likely cargo type bitmask"))
		}
		// Remaining bytes: cargo amounts (likely int32 each for large transfers)
		for i := 6; i+3 < len(d); i += 4 {
			val := int(int32(encoding.Read32(d, i))) //nolint:gosec // intentional signed conversion for cargo values
			cargoName := "Unknown"
			switch (i - 6) / 4 {
			case 0:
				cargoName = "Ironium"
			case 1:
				cargoName = "Boranium"
			case 2:
				cargoName = "Germanium"
			case 3:
				cargoName = "Colonists"
			}
			fields = append(fields, FormatFieldRaw(i, i+3, cargoName,
				fmt.Sprintf("0x%02X%02X%02X%02X", d[i+3], d[i+2], d[i+1], d[i]),
				fmt.Sprintf("int32 LE = %d kT", val)))
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Info ──")
	fields = append(fields, "  Purpose: Large cargo transfer (int32 amounts)")
	fields = append(fields, "  Structure: Partially documented")
	fields = append(fields, fmt.Sprintf("  Size: %d bytes", len(d)))

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
