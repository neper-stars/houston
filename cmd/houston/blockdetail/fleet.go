package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.FleetBlockType, FormatFleet)
	RegisterFormatter(blocks.PartialFleetBlockType, FormatPartialFleet)
}

// FormatFleet provides detailed view for FleetBlock (type 16)
func FormatFleet(block blocks.Block, index int) string {
	fb, ok := block.(blocks.FleetBlock)
	if !ok {
		return FormatGeneric(block, index)
	}
	return formatFleetCommon(&fb.PartialFleetBlock, block, index, "Fleet")
}

// FormatPartialFleet provides detailed view for PartialFleetBlock (type 17)
func FormatPartialFleet(block blocks.Block, index int) string {
	pfb, ok := block.(blocks.PartialFleetBlock)
	if !ok {
		return FormatGeneric(block, index)
	}
	return formatFleetCommon(&pfb, block, index, "PartialFleet")
}

// formatFleetCommon handles the common formatting for both FleetBlock and PartialFleetBlock
func formatFleetCommon(fb *blocks.PartialFleetBlock, block blocks.Block, index int, blockName string) string {
	width := DefaultWidth
	d := fb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 14 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: Fleet number and owner
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "Fleet ID",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("fleetNum=(d[0]+(d[1]&0x01)<<8)=%d, owner=(d[1]>>1)=%d", fb.FleetNumber, fb.Owner)))

	// Bytes 2-3: iPlayer (int16) - Owner player index (redundant with owner from bytes 0-1)
	iPlayer := int(d[2]) | (int(d[3]) << 8)
	fields = append(fields, FormatFieldRaw(0x02, 0x03, "iPlayer",
		fmt.Sprintf("0x%02X%02X", d[3], d[2]),
		fmt.Sprintf("int16 LE = %d (owner player, redundant)", iPlayer)))

	// Byte 4: Kind byte
	kindName := "Unknown"
	switch fb.KindByte {
	case blocks.FleetKindPartial:
		kindName = "Partial"
	case blocks.FleetKindPickPocket:
		kindName = "PickPocket"
	case blocks.FleetKindFull:
		kindName = "Full"
	}
	fields = append(fields, FormatFieldRaw(0x04, 0x04, "KindByte",
		fmt.Sprintf("0x%02X", d[4]),
		fmt.Sprintf("%d = %s", fb.KindByte, kindName)))

	// Byte 5: Flags
	fields = append(fields, FormatFieldRaw(0x05, 0x05, "Byte5 (flags)",
		fmt.Sprintf("0x%02X", d[5]),
		fmt.Sprintf("0b%08b", d[5])))
	fields = append(fields, fmt.Sprintf("           %s bit0: fInclude (in reports) = %v", TreeBranch, fb.Include))
	fields = append(fields, fmt.Sprintf("           %s bit1: fRepOrders (repeat waypoints) = %v", TreeBranch, fb.RepeatOrders))
	fields = append(fields, fmt.Sprintf("           %s bit2: fDead (fleet destroyed) = %v", TreeBranch, fb.IsDead))
	fields = append(fields, fmt.Sprintf("           %s bit3: fByteCsh (1=1byte counts, 0=2bytes) = %v", TreeEnd, !fb.ShipCountTwoBytes))
	// Bits 4-7 are NOT persisted - runtime-only flags (fDone, fBombed, fHereAllTurn, fNoHeal)

	// Bytes 6-7: Position object ID
	fields = append(fields, FormatFieldRaw(0x06, 0x07, "PositionObjectId",
		fmt.Sprintf("0x%02X%02X", d[7], d[6]),
		fmt.Sprintf("uint16 LE = %d", fb.PositionObjectId)))

	// Bytes 8-9: X coordinate
	fields = append(fields, FormatFieldRaw(0x08, 0x09, "X",
		fmt.Sprintf("0x%02X%02X", d[9], d[8]),
		fmt.Sprintf("uint16 LE = %d", fb.X)))

	// Bytes 10-11: Y coordinate
	fields = append(fields, FormatFieldRaw(0x0A, 0x0B, "Y",
		fmt.Sprintf("0x%02X%02X", d[11], d[10]),
		fmt.Sprintf("uint16 LE = %d", fb.Y)))

	// Bytes 12-13: Ship types bitmask
	fields = append(fields, FormatFieldRaw(0x0C, 0x0D, "ShipTypes",
		fmt.Sprintf("0x%02X%02X", d[13], d[12]),
		fmt.Sprintf("uint16 LE = 0x%04X (bitmask)", fb.ShipTypes)))

	// Ship counts (variable position)
	idx := 14
	shipCount := 0
	for bit := 0; bit < 16; bit++ {
		if (fb.ShipTypes & (1 << bit)) != 0 {
			shipCount++
		}
	}

	if shipCount > 0 {
		fields = append(fields, "")
		bytesPerCount := 1
		if fb.ShipCountTwoBytes {
			bytesPerCount = 2
		}
		shipCountEnd := idx + shipCount*bytesPerCount - 1
		fields = append(fields, fmt.Sprintf("0x%02X-0x%02X: Ship Counts (%d designs, %d byte(s) each)",
			idx, shipCountEnd, shipCount, bytesPerCount))

		designIdx := 0
		for bit := 0; bit < 16; bit++ {
			if (fb.ShipTypes & (1 << bit)) != 0 {
				prefix := TreeBranch
				designIdx++
				if designIdx == shipCount {
					prefix = TreeEnd
				}
				if fb.ShipCountTwoBytes {
					if idx+2 <= len(d) {
						raw := encoding.Read16(d, idx)
						fields = append(fields, fmt.Sprintf("           %s Design %d: 0x%04X -> %d ships",
							prefix, bit, raw, fb.ShipCount[bit]))
						idx += 2
					}
				} else {
					if idx < len(d) {
						fields = append(fields, fmt.Sprintf("           %s Design %d: 0x%02X -> %d ships",
							prefix, bit, d[idx], fb.ShipCount[bit]))
						idx++
					}
				}
			}
		}
	}

	// Cargo section (if Full or PickPocket)
	if fb.HasCargo() && idx+2 <= len(d) {
		fields = append(fields, "")
		fields = append(fields, "── Cargo Section ──")

		contentsLengths := encoding.Read16(d, idx)
		fields = append(fields, FormatFieldRaw(idx, idx+1, "ContentsLengths",
			fmt.Sprintf("0x%02X%02X", d[idx+1], d[idx]),
			fmt.Sprintf("uint16 LE = 0x%04X (var-len indicators)", contentsLengths)))
		idx += 2

		// Show cargo values
		fields = append(fields, fmt.Sprintf("           %s Ironium = %d", TreeBranch, fb.Ironium))
		fields = append(fields, fmt.Sprintf("           %s Boranium = %d", TreeBranch, fb.Boranium))
		fields = append(fields, fmt.Sprintf("           %s Germanium = %d", TreeBranch, fb.Germanium))
		fields = append(fields, fmt.Sprintf("           %s Population = %d", TreeBranch, fb.Population))
		fields = append(fields, fmt.Sprintf("           %s Fuel = %d", TreeEnd, fb.Fuel))
	}

	// Full fleet specific data
	if fb.KindByte == blocks.FleetKindFull {
		fields = append(fields, "")
		fields = append(fields, "── Full Fleet Data ──")

		// Skip cargo bytes to find damaged ships section
		// This is complex due to variable length encoding, so we use decoded values
		fields = append(fields, fmt.Sprintf("  DamagedShipTypes = 0x%04X (bitmask)", fb.DamagedShipTypes))

		if fb.DamagedShipTypes != 0 {
			for bit := 0; bit < 16; bit++ {
				if (fb.DamagedShipTypes & (1 << bit)) != 0 {
					fields = append(fields, fmt.Sprintf("           %s Design %d damage info = 0x%04X",
						TreeBranch, bit, fb.DamagedShipInfo[bit]))
				}
			}
		}

		fields = append(fields, fmt.Sprintf("  BattlePlan = %d", fb.BattlePlan))
		fields = append(fields, fmt.Sprintf("  WaypointCount = %d", fb.WaypointCount))
	} else if len(d) >= idx+8 {
		// Partial fleet movement data
		fields = append(fields, "")
		fields = append(fields, "── Movement Data (Partial Fleet) ──")

		fields = append(fields, FormatFieldRaw(idx, idx, "DeltaX (raw)",
			fmt.Sprintf("0x%02X", d[idx]),
			fmt.Sprintf("(%d - 127) = %d", d[idx], fb.DeltaX)))

		fields = append(fields, FormatFieldRaw(idx+1, idx+1, "DeltaY (raw)",
			fmt.Sprintf("0x%02X", d[idx+1]),
			fmt.Sprintf("(%d - 127) = %d", d[idx+1], fb.DeltaY)))

		fields = append(fields, FormatFieldRaw(idx+2, idx+2, "Warp + Unknown",
			fmt.Sprintf("0x%02X", d[idx+2]),
			fmt.Sprintf("warp=(d&0x0F)=%d, unknown=(d&0xF0)=0x%02X", fb.Warp, fb.UnknownBitsWithWarp)))

		fields = append(fields, FormatFieldRaw(idx+3, idx+3, "Padding",
			fmt.Sprintf("0x%02X", d[idx+3]),
			"(should be 0)"))

		if idx+8 <= len(d) {
			fields = append(fields, FormatFieldRaw(idx+4, idx+7, "Mass",
				fmt.Sprintf("0x%02X%02X%02X%02X", d[idx+7], d[idx+6], d[idx+5], d[idx+4]),
				fmt.Sprintf("uint32 LE = %d", fb.Mass)))
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Total Ships: %d", fb.TotalShips()))

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
