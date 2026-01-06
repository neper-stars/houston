package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
)

func init() {
	RegisterFormatter(blocks.MessagesFilterBlockType, FormatMessagesFilter)
}

// FormatMessagesFilter provides detailed view for MessagesFilterBlock (type 33)
func FormatMessagesFilter(block blocks.Block, index int) string {
	width := DefaultWidth
	mfb, ok := block.(blocks.MessagesFilterBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := mfb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) == 0 {
		fields = append(fields, "(no data)")
	} else {
		// Structure not fully documented - show raw bytes with potential interpretations
		fields = append(fields, "── Raw Data (structure TBD) ──")

		// Messages filter likely contains bitmasks for filtering message types
		// Show each byte with potential filter flag interpretation
		for i := 0; i < len(d); i++ {
			annotation := "TBD"
			if i == 0 {
				annotation = fmt.Sprintf("0b%08b (possibly filter flags)", d[i])
			} else if i < 4 {
				annotation = fmt.Sprintf("0b%08b (possibly more flags)", d[i])
			}
			fields = append(fields, FormatFieldRaw(i, i, fmt.Sprintf("Byte%d", i),
				fmt.Sprintf("0x%02X", d[i]),
				annotation))
		}

		// If we have at least 2 bytes, show as potential bitmask
		if len(d) >= 2 {
			fields = append(fields, "")
			fields = append(fields, "── Potential Filter Flags ──")
			fields = append(fields, "  (interpretation speculative)")

			// Common message types that might be filterable
			filterNames := []string{
				"Battles",
				"Production",
				"Research",
				"Planets",
				"Fleets",
				"Diplomacy",
				"System",
				"Other",
			}

			for i, name := range filterNames {
				if i < len(d)*8 {
					byteIdx := i / 8
					bitIdx := i % 8
					if byteIdx < len(d) {
						enabled := (d[byteIdx] & (1 << bitIdx)) != 0
						prefix := TreeBranch
						if i == len(filterNames)-1 || i == len(d)*8-1 {
							prefix = TreeEnd
						}
						fields = append(fields, fmt.Sprintf("  %s bit%d: %s = %v", prefix, i, name, enabled))
					}
				}
			}
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Info ──")
	fields = append(fields, "  Purpose: Message filter settings")
	fields = append(fields, "  Structure: Not fully documented")
	fields = append(fields, fmt.Sprintf("  Size: %d bytes", len(d)))

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
