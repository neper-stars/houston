package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.FleetNameBlockType, FormatFleetName)
	RegisterFormatter(blocks.MoveShipsBlockType, FormatMoveShips)
	RegisterFormatter(blocks.FleetSplitBlockType, FormatFleetSplit)
	RegisterFormatter(blocks.FleetsMergeBlockType, FormatFleetsMerge)
}

// FormatFleetName provides detailed view for FleetNameBlock (type 21)
func FormatFleetName(block blocks.Block, index int) string {
	width := DefaultWidth
	fnb, ok := block.(blocks.FleetNameBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := fnb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 1 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Byte 0: Name length
	nameLen := int(d[0])
	fields = append(fields, FormatFieldRaw(0x00, 0x00, "NameLength",
		fmt.Sprintf("0x%02X", d[0]),
		fmt.Sprintf("%d bytes", nameLen)))

	// Bytes 1+: Encoded name
	if len(d) > 1 && nameLen > 0 {
		endIdx := 1 + nameLen
		if endIdx > len(d) {
			endIdx = len(d)
		}
		fields = append(fields, FormatFieldRaw(0x01, endIdx-1, "NameData",
			fmt.Sprintf("0x%s", HexDumpSingleLine(d[1:endIdx])),
			fmt.Sprintf("%q (Stars! encoded)", fnb.Name)))
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Fleet name: %q", fnb.Name))
	fields = append(fields, "  Note: This block precedes the FleetBlock it names")

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// FormatMoveShips provides detailed view for MoveShipsBlock (type 23)
func FormatMoveShips(block blocks.Block, index int) string {
	width := DefaultWidth
	msb, ok := block.(blocks.MoveShipsBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := msb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 4 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: Destination fleet number
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "DestFleet",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = %d -> Fleet #%d", msb.DestFleetNumber, msb.DestFleetNumber+1)))

	// Bytes 2-3: Source fleet number
	fields = append(fields, FormatFieldRaw(0x02, 0x03, "SourceFleet",
		fmt.Sprintf("0x%02X%02X", d[3], d[2]),
		fmt.Sprintf("uint16 LE = %d -> Fleet #%d", msb.SourceFleetNumber, msb.SourceFleetNumber+1)))

	// Transfer info
	if len(msb.TransferInfo) > 0 {
		fields = append(fields, "")
		fields = append(fields, "── Transfer Info ──")

		// Byte 0: Flags
		if len(msb.TransferInfo) > 0 {
			fields = append(fields, FormatFieldRaw(0x04, 0x04, "Flags",
				fmt.Sprintf("0x%02X", msb.TransferInfo[0]),
				"transfer flags"))
		}

		// Bytes 1-2: Ship types bitmask
		if len(msb.TransferInfo) >= 3 {
			fields = append(fields, FormatFieldRaw(0x05, 0x06, "ShipTypeMask",
				fmt.Sprintf("0x%02X%02X", msb.TransferInfo[2], msb.TransferInfo[1]),
				fmt.Sprintf("bitmask = 0x%04X", msb.ShipTypeMask)))

			// Show which design slots are involved
			var activeSlots []string
			for i := 0; i < 16; i++ {
				if msb.ShipTypeMask&(1<<i) != 0 {
					activeSlots = append(activeSlots, fmt.Sprintf("D%d", i))
				}
			}
			if len(activeSlots) > 0 {
				fields = append(fields, fmt.Sprintf("           %s Designs: %v", TreeEnd, activeSlots))
			}
		}

		// Ship transfers
		if len(msb.ShipTransfers) > 0 {
			fields = append(fields, "")
			fields = append(fields, fmt.Sprintf("── Ship Transfers (%d design slots) ──", len(msb.ShipTransfers)))

			idx := 7 // Start of ship counts in raw data
			for i, transfer := range msb.ShipTransfers {
				prefix := TreeBranch
				if i == len(msb.ShipTransfers)-1 {
					prefix = TreeEnd
				}

				direction := "arriving at dest"
				if transfer.Count < 0 {
					direction = "leaving dest"
				}

				// Show raw bytes if available
				if idx+1 < len(d) {
					fields = append(fields, FormatFieldRaw(idx, idx+1, fmt.Sprintf("Design%d", transfer.DesignSlot),
						fmt.Sprintf("0x%02X%02X", d[idx+1], d[idx]),
						fmt.Sprintf("int16 LE = %d ships (%s)", transfer.Count, direction)))
				} else {
					fields = append(fields, fmt.Sprintf("  %s Design #%d: %d ships (%s)",
						prefix, transfer.DesignSlot, transfer.Count, direction))
				}
				idx += 2
			}
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Source: Fleet #%d", msb.SourceFleetNumber+1))
	fields = append(fields, fmt.Sprintf("  Destination: Fleet #%d", msb.DestFleetNumber+1))
	if len(msb.ShipTransfers) > 0 {
		fields = append(fields, "  Transfers:")
		for i, transfer := range msb.ShipTransfers {
			prefix := TreeBranch
			if i == len(msb.ShipTransfers)-1 {
				prefix = TreeEnd
			}
			absCount := transfer.Count
			if absCount < 0 {
				absCount = -absCount
			}
			if transfer.Count > 0 {
				fields = append(fields, fmt.Sprintf("  %s Design #%d: +%d ships to dest", prefix, transfer.DesignSlot, absCount))
			} else {
				fields = append(fields, fmt.Sprintf("  %s Design #%d: -%d ships from dest", prefix, transfer.DesignSlot, absCount))
			}
		}
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// FormatFleetSplit provides detailed view for FleetSplitBlock (type 24)
func FormatFleetSplit(block blocks.Block, index int) string {
	width := DefaultWidth
	fsb, ok := block.(blocks.FleetSplitBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := fsb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 2 {
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
		TreeEnd, fsb.FleetNumber, fsb.FleetNumber+1))

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
	fields = append(fields, fmt.Sprintf("  Split Fleet #%d into new fleet", fsb.FleetNumber+1))
	fields = append(fields, "  Note: Split details in subsequent MoveShips block")

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// FormatFleetsMerge provides detailed view for FleetsMergeBlock (type 37)
func FormatFleetsMerge(block blocks.Block, index int) string {
	width := DefaultWidth
	fmb, ok := block.(blocks.FleetsMergeBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := fmb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 2 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: Target fleet number (9 bits)
	fleetWord := encoding.Read16(d, 0)
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "TargetFleet",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = 0x%04X", fleetWord)))
	fields = append(fields, fmt.Sprintf("           %s bits 0-8: Fleet = %d -> Fleet #%d",
		TreeEnd, fmb.FleetNumber, fmb.FleetNumber+1))

	// Fleets to merge (2 bytes each)
	if len(fmb.FleetsToMerge) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Fleets to Merge (%d fleets) ──", len(fmb.FleetsToMerge)))

		for i, fleet := range fmb.FleetsToMerge {
			offset := 2 + i*2
			if offset+2 <= len(d) {
				fields = append(fields, FormatFieldRaw(offset, offset+1, fmt.Sprintf("MergeFleet%d", i),
					fmt.Sprintf("0x%02X%02X", d[offset+1], d[offset]),
					fmt.Sprintf("(d & 0x1FF) = %d -> Fleet #%d", fleet, fleet+1)))
			}
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Target: Fleet #%d", fmb.FleetNumber+1))
	if len(fmb.FleetsToMerge) > 0 {
		fields = append(fields, "  Merging:")
		for i, fleet := range fmb.FleetsToMerge {
			prefix := TreeBranch
			if i == len(fmb.FleetsToMerge)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s Fleet #%d", prefix, fleet+1))
		}
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
