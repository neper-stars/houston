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
		// Summary of filtered messages
		fields = append(fields, "── Filter Summary ──")
		fields = append(fields, fmt.Sprintf("  Bitmap size: %d bytes (%d bits)", len(mfb.FilterBitmap), len(mfb.FilterBitmap)*8))
		fields = append(fields, fmt.Sprintf("  Filtered message types: %d", mfb.FilteredCount))

		// Show which message IDs are filtered (if any)
		filteredIds := mfb.GetFilteredMessageIds()
		if len(filteredIds) > 0 {
			fields = append(fields, "")
			fields = append(fields, "── Filtered Message IDs ──")

			// Group by category for cleaner display
			categoryCount := make(map[string]int)
			for _, id := range filteredIds {
				cat := blocks.MessageCategory(blocks.MessageTypeID(id))
				if cat == "" {
					cat = "Unknown"
				}
				categoryCount[cat]++
			}

			// Show category summary first
			fields = append(fields, "  By category:")
			for cat, count := range categoryCount {
				fields = append(fields, fmt.Sprintf("    %s: %d", cat, count))
			}

			// Show individual IDs (up to 30)
			fields = append(fields, "")
			fields = append(fields, "  Individual IDs:")
			maxShow := 30
			for i, id := range filteredIds {
				if i >= maxShow {
					fields = append(fields, fmt.Sprintf("    ... and %d more", len(filteredIds)-maxShow))
					break
				}
				prefix := TreeBranch
				if i == len(filteredIds)-1 || i == maxShow-1 {
					prefix = TreeEnd
				}
				cat := blocks.MessageCategory(blocks.MessageTypeID(id))
				if cat != "" {
					fields = append(fields, fmt.Sprintf("  %s 0x%03X (%3d) - %s", prefix, id, id, cat))
				} else {
					fields = append(fields, fmt.Sprintf("  %s 0x%03X (%3d)", prefix, id, id))
				}
			}
		}

		// Show bitmap as hex for reference
		fields = append(fields, "")
		fields = append(fields, "── Bitmap (hex) ──")
		// Show in rows of 16 bytes
		for row := 0; row < len(mfb.FilterBitmap); row += 16 {
			end := row + 16
			if end > len(mfb.FilterBitmap) {
				end = len(mfb.FilterBitmap)
			}
			hexStr := ""
			for i := row; i < end; i++ {
				hexStr += fmt.Sprintf("%02X ", mfb.FilterBitmap[i])
			}
			fields = append(fields, fmt.Sprintf("  %02X: %s", row, hexStr))
		}
	}

	// Info section
	fields = append(fields, "")
	fields = append(fields, "── Info ──")
	fields = append(fields, "  Purpose: Message filter preferences bitmap")
	fields = append(fields, "  Bit addressing: byte[msgId/8] & (1 << (msgId%8))")
	fields = append(fields, fmt.Sprintf("  Size: %d bytes", len(d)))

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
