package blockdetail

import (
	"github.com/neper-stars/houston/blocks"
)

// FormatGeneric provides a generic hex dump view for blocks without specific formatters
func FormatGeneric(block blocks.Block, index int) string {
	width := DefaultWidth
	data := block.DecryptedData()

	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(data, width)

	// Add note that detailed view is not implemented
	var fields []string
	fields = append(fields, "(detailed view not yet implemented for this block type)")

	fieldsSection := FormatFieldsSection(fields, width)

	return BuildOutput(header, hexSection, fieldsSection)
}
