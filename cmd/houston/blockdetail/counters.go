package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.CountersBlockType, FormatCounters)
}

// FormatCounters provides detailed view for CountersBlock (type 32)
func FormatCounters(block blocks.Block, index int) string {
	width := DefaultWidth
	cb, ok := block.(blocks.CountersBlock)
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

	// Bytes 0-1: Planet count
	planetCount := encoding.Read16(d, 0)
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "PlanetCount",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = %d", planetCount)))

	// Bytes 2-3: Fleet count
	fleetCount := encoding.Read16(d, 2)
	fields = append(fields, FormatFieldRaw(0x02, 0x03, "FleetCount",
		fmt.Sprintf("0x%02X%02X", d[3], d[2]),
		fmt.Sprintf("uint16 LE = %d", fleetCount)))

	// Additional bytes if present
	if len(d) > 4 {
		fields = append(fields, "")
		fields = append(fields, "── Additional Data ──")
		for i := 4; i < len(d); i += 2 {
			if i+2 <= len(d) {
				val := encoding.Read16(d, i)
				fields = append(fields, FormatFieldRaw(i, i+1, fmt.Sprintf("Unknown%d", (i-4)/2),
					fmt.Sprintf("0x%02X%02X", d[i+1], d[i]),
					fmt.Sprintf("uint16 LE = %d (TBD)", val)))
			} else if i < len(d) {
				fields = append(fields, FormatFieldRaw(i, i, fmt.Sprintf("Unknown%d", (i-4)/2),
					fmt.Sprintf("0x%02X", d[i]),
					"TBD"))
			}
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Planets: %d", cb.PlanetCount))
	fields = append(fields, fmt.Sprintf("  Fleets: %d", cb.FleetCount))

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
