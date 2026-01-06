package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.ProductionQueueBlockType, FormatProductionQueue)
	RegisterFormatter(blocks.ProductionQueueChangeBlockType, FormatProductionQueueChange)
}

// productionItemName returns human-readable name for production item
func productionItemName(itemId int, itemType int) string {
	if itemType == blocks.ProductionItemTypeCustom {
		return fmt.Sprintf("Ship Design #%d", itemId)
	}

	// Standard items
	names := map[int]string{
		blocks.ProductionItemAutoMines:        "Auto Mines",
		blocks.ProductionItemAutoFactories:    "Auto Factories",
		blocks.ProductionItemAutoDefenses:     "Auto Defenses",
		blocks.ProductionItemAutoAlchemy:      "Auto Alchemy",
		blocks.ProductionItemAutoMinTerraform: "Auto Min Terraform",
		blocks.ProductionItemAutoMaxTerraform: "Auto Max Terraform",
		blocks.ProductionItemAutoPackets:      "Auto Packets",
		blocks.ProductionItemFactory:          "Factory",
		blocks.ProductionItemMine:             "Mine",
		blocks.ProductionItemDefense:          "Defense",
		blocks.ProductionItemMineralAlchemy:   "Mineral Alchemy",
		blocks.ProductionItemPacketIronium:    "Packet (Ironium)",
		blocks.ProductionItemPacketBoranium:   "Packet (Boranium)",
		blocks.ProductionItemPacketGermanium:  "Packet (Germanium)",
		blocks.ProductionItemPacketMixed:      "Packet (Mixed)",
		blocks.ProductionItemScanner:          "Scanner",
	}
	if name, ok := names[itemId]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", itemId)
}

// itemTypeName returns human-readable name for item type
func itemTypeName(itemType int) string {
	switch itemType {
	case blocks.ProductionItemTypeStandard:
		return "Standard"
	case blocks.ProductionItemTypeCustom:
		return "Custom"
	default:
		return fmt.Sprintf("Unknown(%d)", itemType)
	}
}

// FormatProductionQueue provides detailed view for ProductionQueueBlock (type 28)
func FormatProductionQueue(block blocks.Block, index int) string {
	width := DefaultWidth
	pqb, ok := block.(blocks.ProductionQueueBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := pqb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) == 0 {
		fields = append(fields, "(empty queue)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	fields = append(fields, fmt.Sprintf("── Queue Items (%d items, 4 bytes each) ──", len(pqb.Items)))

	for i, item := range pqb.Items {
		offset := i * 4
		if offset+4 > len(d) {
			break
		}

		chunk1 := encoding.Read16(d, offset)
		chunk2 := encoding.Read16(d, offset+2)

		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("Item %d @ 0x%02X:", i, offset))
		fields = append(fields, FormatFieldRaw(offset, offset+1, "Chunk1",
			fmt.Sprintf("0x%02X%02X", d[offset+1], d[offset]),
			fmt.Sprintf("uint16 LE = 0x%04X", chunk1)))
		fields = append(fields, fmt.Sprintf("           %s itemId = (d >> 10) = %d",
			TreeBranch, item.ItemId))
		fields = append(fields, fmt.Sprintf("           %s count = (d & 0x3FF) = %d",
			TreeEnd, item.Count))

		fields = append(fields, FormatFieldRaw(offset+2, offset+3, "Chunk2",
			fmt.Sprintf("0x%02X%02X", d[offset+3], d[offset+2]),
			fmt.Sprintf("uint16 LE = 0x%04X", chunk2)))
		fields = append(fields, fmt.Sprintf("           %s completePercent = (d >> 4) = %d (%.1f%%)",
			TreeBranch, item.CompletePercent, float64(item.CompletePercent)/40.95))
		fields = append(fields, fmt.Sprintf("           %s itemType = (d & 0x0F) = %d (%s)",
			TreeEnd, item.ItemType, itemTypeName(item.ItemType)))

		// Item name
		itemName := productionItemName(item.ItemId, item.ItemType)
		fields = append(fields, fmt.Sprintf("           -> %s × %d", itemName, item.Count))
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Queue length: %d items", len(pqb.Items)))
	for i, item := range pqb.Items {
		itemName := productionItemName(item.ItemId, item.ItemType)
		prefix := TreeBranch
		if i == len(pqb.Items)-1 {
			prefix = TreeEnd
		}
		fields = append(fields, fmt.Sprintf("  %s %s × %d", prefix, itemName, item.Count))
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// FormatProductionQueueChange provides detailed view for ProductionQueueChangeBlock (type 29)
func FormatProductionQueueChange(block blocks.Block, index int) string {
	width := DefaultWidth
	pqcb, ok := block.(blocks.ProductionQueueChangeBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := pqcb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 2 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: Planet ID (11 bits)
	planetWord := encoding.Read16(d, 0)
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "PlanetId",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("(d & 0x7FF) = %d -> Planet #%d", pqcb.PlanetId, pqcb.PlanetId+1)))

	// Upper bits of planet word
	upperBits := planetWord >> 11
	if upperBits != 0 {
		fields = append(fields, fmt.Sprintf("           %s upper bits = 0x%X (unknown)", TreeEnd, upperBits))
	}

	// Queue items
	if len(d) > 2 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Queue Items (%d items, 4 bytes each) ──", len(pqcb.Items)))

		for i, item := range pqcb.Items {
			offset := 2 + i*4
			if offset+4 > len(d) {
				break
			}

			chunk1 := encoding.Read16(d, offset)
			chunk2 := encoding.Read16(d, offset+2)

			fields = append(fields, "")
			fields = append(fields, fmt.Sprintf("Item %d @ 0x%02X:", i, offset))
			fields = append(fields, FormatFieldRaw(offset, offset+1, "Chunk1",
				fmt.Sprintf("0x%02X%02X", d[offset+1], d[offset]),
				fmt.Sprintf("uint16 LE = 0x%04X", chunk1)))
			fields = append(fields, fmt.Sprintf("           %s itemId = (d >> 10) = %d",
				TreeBranch, item.ItemId))
			fields = append(fields, fmt.Sprintf("           %s count = (d & 0x3FF) = %d",
				TreeEnd, item.Count))

			fields = append(fields, FormatFieldRaw(offset+2, offset+3, "Chunk2",
				fmt.Sprintf("0x%02X%02X", d[offset+3], d[offset+2]),
				fmt.Sprintf("uint16 LE = 0x%04X", chunk2)))
			fields = append(fields, fmt.Sprintf("           %s completePercent = (d >> 4) = %d (%.1f%%)",
				TreeBranch, item.CompletePercent, float64(item.CompletePercent)/40.95))
			fields = append(fields, fmt.Sprintf("           %s itemType = (d & 0x0F) = %d (%s)",
				TreeEnd, item.ItemType, itemTypeName(item.ItemType)))

			// Item name
			itemName := productionItemName(item.ItemId, item.ItemType)
			fields = append(fields, fmt.Sprintf("           -> %s × %d", itemName, item.Count))
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Planet #%d: %d queue items", pqcb.PlanetId+1, len(pqcb.Items)))
	for i, item := range pqcb.Items {
		itemName := productionItemName(item.ItemId, item.ItemType)
		prefix := TreeBranch
		if i == len(pqcb.Items)-1 {
			prefix = TreeEnd
		}
		fields = append(fields, fmt.Sprintf("  %s %s × %d", prefix, itemName, item.Count))
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
