package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
)

func init() {
	RegisterFormatter(blocks.SaveAndSubmitBlockType, FormatSaveAndSubmit)
}

// FormatSaveAndSubmit provides detailed view for SaveAndSubmitBlock (type 46)
func FormatSaveAndSubmit(block blocks.Block, index int) string {
	width := DefaultWidth
	ssb, ok := block.(blocks.SaveAndSubmitBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := ssb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) == 0 {
		fields = append(fields, "(no data)")
	} else {
		// Structure not fully documented - show raw bytes
		fields = append(fields, "── Raw Data (structure TBD) ──")

		// Show bytes with annotations for what we might know
		for i := 0; i < len(d); i++ {
			annotation := "TBD"
			if i == 0 {
				annotation = "possibly flags or action type"
			}
			fields = append(fields, FormatFieldRaw(i, i, fmt.Sprintf("Byte%d", i),
				fmt.Sprintf("0x%02X", d[i]),
				annotation))
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Info ──")
	fields = append(fields, "  Purpose: Marks turn as saved and submitted")
	fields = append(fields, "  Structure: Not fully documented")
	fields = append(fields, fmt.Sprintf("  Size: %d bytes", len(d)))

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
