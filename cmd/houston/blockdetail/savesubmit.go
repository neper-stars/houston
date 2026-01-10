package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
)

func init() {
	RegisterFormatter(blocks.SaveAndSubmitBlockType, FormatSaveAndSubmit)
}

// autoBuildItemName returns the name of an auto-build production item
func autoBuildItemName(itemID uint16) string {
	names := map[uint16]string{
		0: "AutoMines",
		1: "AutoFactories",
		2: "AutoDefenses",
		3: "AutoAlchemy",
		4: "AutoMinTerraform",
		5: "AutoMaxTerraform",
		6: "AutoPackets",
	}
	if name, ok := names[itemID]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", itemID)
}

// FormatSaveAndSubmit provides detailed view for SaveAndSubmitBlock (type 46)
// Contains the default ZipProd template (ZIPPRODQ1) for production queues.
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

	if len(d) < 2 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Byte 0: fNoResearch flag
	noResearchStr := "Contribute to Research"
	if ssb.ZipProd.NoResearch {
		noResearchStr = "Don't contribute to Research"
	}
	fields = append(fields, FormatFieldRaw(0x00, 0x00, "fNoResearch",
		fmt.Sprintf("0x%02X", d[0]),
		fmt.Sprintf("%d (%s)", d[0], noResearchStr)))

	// Byte 1: cpq (item count)
	fields = append(fields, FormatFieldRaw(0x01, 0x01, "cpq (item count)",
		fmt.Sprintf("0x%02X", d[1]),
		fmt.Sprintf("%d items", d[1])))

	// Bytes 2+: rgpq items (2 bytes each)
	if len(ssb.ZipProd.Items) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Production Items (rgpq[%d]) ──", len(ssb.ZipProd.Items)))
		for i, item := range ssb.ZipProd.Items {
			offset := 2 + i*2
			prefix := TreeBranch
			if i == len(ssb.ZipProd.Items)-1 {
				prefix = TreeEnd
			}

			// Show raw bytes
			if offset+1 < len(d) {
				rawVal := uint16(d[offset]) | (uint16(d[offset+1]) << 8)
				fields = append(fields, FormatFieldRaw(offset, offset+1, fmt.Sprintf("Item[%d]", i),
					fmt.Sprintf("0x%04X", rawVal),
					fmt.Sprintf("(Count<<6)|ItemId")))
			}

			// Show decoded values
			quantityStr := fmt.Sprintf("%d", item.Quantity)
			if item.Quantity == 0 && item.ItemType == 3 { // AutoAlchemy with 0 = unlimited
				quantityStr = "unlimited"
			}
			fields = append(fields, fmt.Sprintf("           %s ItemType=%d (%s), Quantity=%s",
				prefix, item.ItemType, autoBuildItemName(item.ItemType), quantityStr))
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Research mode: %s", noResearchStr))
	fields = append(fields, fmt.Sprintf("  Queue items: %d", len(ssb.ZipProd.Items)))
	if len(ssb.ZipProd.Items) > 0 {
		for _, item := range ssb.ZipProd.Items {
			quantityStr := fmt.Sprintf("%d", item.Quantity)
			if item.Quantity == 0 && item.ItemType == 3 {
				quantityStr = "unlimited"
			}
			fields = append(fields, fmt.Sprintf("    - %s × %s", autoBuildItemName(item.ItemType), quantityStr))
		}
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
