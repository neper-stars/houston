package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.FileFooterBlockType, FormatFileFooter)
}

// FormatFileFooter provides detailed view for FileFooterBlock (type 0)
func FormatFileFooter(block blocks.Block, index int) string {
	width := DefaultWidth
	ffb, ok := block.(blocks.FileFooterBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := ffb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) == 0 {
		// H files have no checksum
		fields = append(fields, "(no checksum - H file footer)")
	} else if len(d) >= 2 {
		// Standard footer with checksum
		checksum := encoding.Read16(d, 0)
		fields = append(fields, FormatFieldRaw(0x00, 0x01, "Checksum",
			fmt.Sprintf("0x%02X%02X", d[1], d[0]),
			fmt.Sprintf("uint16 LE = 0x%04X (%d)", checksum, checksum)))

		fields = append(fields, "")
		fields = append(fields, "── Info ──")
		fields = append(fields, "  This block marks the end of the file.")
		fields = append(fields, "  The checksum is used to validate file integrity.")
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	if ffb.HasChecksum() {
		fields = append(fields, fmt.Sprintf("  Checksum: 0x%04X", ffb.Checksum))
	} else {
		fields = append(fields, "  No checksum (H file)")
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
