package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/data"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.DesignBlockType, FormatDesign)
	RegisterFormatter(blocks.DesignChangeBlockType, FormatDesignChange)
}

// itemCategoryNames maps category bitmasks to their names
var itemCategoryNames = map[uint16]string{
	blocks.ItemCategoryEmpty:       "Empty",
	blocks.ItemCategoryEngine:      "Engine",
	blocks.ItemCategoryScanner:     "Scanner",
	blocks.ItemCategoryShield:      "Shield",
	blocks.ItemCategoryArmor:       "Armor",
	blocks.ItemCategoryBeamWeapon:  "BeamWeapon",
	blocks.ItemCategoryTorpedo:     "Torpedo",
	blocks.ItemCategoryBomb:        "Bomb",
	blocks.ItemCategoryMiningRobot: "MiningRobot",
	blocks.ItemCategoryMineLayer:   "MineLayer",
	blocks.ItemCategoryOrbital:     "Orbital",
	blocks.ItemCategoryPlanetary:   "Planetary",
	blocks.ItemCategoryElectrical:  "Electrical",
	blocks.ItemCategoryMechanical:  "Mechanical",
}

// FormatDesign provides detailed view for DesignBlock
func FormatDesign(block blocks.Block, index int) string {
	width := DefaultWidth
	db, ok := block.(blocks.DesignBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := db.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 4 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Byte 0: Control byte 1
	fields = append(fields, FormatFieldRaw(0x00, 0x00, "Control Byte 1",
		fmt.Sprintf("0x%02X", d[0]),
		fmt.Sprintf("0b%08b", d[0])))
	fields = append(fields, fmt.Sprintf("           %s bits0-1: fixed = %d (expected 3)", TreeBranch, d[0]&0x03))
	fields = append(fields, fmt.Sprintf("           %s bit2: isFullDesign = %v", TreeBranch, db.IsFullDesign))
	fields = append(fields, fmt.Sprintf("           %s bits3-7: unused = %d", TreeEnd, (d[0]>>3)&0x1F))

	// Byte 1: Control byte 2
	fields = append(fields, FormatFieldRaw(0x01, 0x01, "Control Byte 2",
		fmt.Sprintf("0x%02X", d[1]),
		fmt.Sprintf("0b%08b", d[1])))
	fields = append(fields, fmt.Sprintf("           %s bit0: fixed = %d (expected 1)", TreeBranch, d[1]&0x01))
	fields = append(fields, fmt.Sprintf("           %s bit1: unused = %d", TreeBranch, (d[1]>>1)&0x01))
	fields = append(fields, fmt.Sprintf("           %s bits2-5: designNumber = %d", TreeBranch, db.DesignNumber))
	fields = append(fields, fmt.Sprintf("           %s bit6: isStarbase = %v", TreeBranch, db.IsStarbase))
	fields = append(fields, fmt.Sprintf("           %s bit7: isTransferred = %v", TreeEnd, db.IsTransferred))

	// Byte 2: Hull ID
	hullName := data.HullNames[db.HullId]
	if hullName == "" {
		hullName = "Unknown"
	}
	fields = append(fields, FormatFieldRaw(0x02, 0x02, "Hull ID",
		fmt.Sprintf("0x%02X", d[2]),
		fmt.Sprintf("%d = %s", db.HullId, hullName)))

	// Byte 3: Picture ID
	fields = append(fields, FormatFieldRaw(0x03, 0x03, "Picture ID",
		fmt.Sprintf("0x%02X", d[3]),
		fmt.Sprintf("%d", db.Pic)))

	if db.IsFullDesign {
		// Full design format
		if len(d) < 17 {
			fields = append(fields, "(full design block too short)")
			fieldsSection := FormatFieldsSection(fields, width)
			return BuildOutput(header, hexSection, fieldsSection)
		}

		fields = append(fields, "")
		fields = append(fields, "── Full Design Data ──")

		// Bytes 4-5: Armor
		fields = append(fields, FormatFieldRaw(0x04, 0x05, "Armor",
			fmt.Sprintf("0x%02X%02X", d[5], d[4]),
			fmt.Sprintf("uint16 LE = %d", db.Armor)))

		// Byte 6: Slot count
		fields = append(fields, FormatFieldRaw(0x06, 0x06, "Slot Count",
			fmt.Sprintf("0x%02X", d[6]),
			fmt.Sprintf("%d", db.SlotCount)))

		// Bytes 7-8: Turn designed
		fields = append(fields, FormatFieldRaw(0x07, 0x08, "Turn Designed",
			fmt.Sprintf("0x%02X%02X", d[8], d[7]),
			fmt.Sprintf("uint16 LE = %d -> Year %d", db.TurnDesigned, 2400+db.TurnDesigned)))

		// Bytes 9-12: Total built
		fields = append(fields, FormatFieldRaw(0x09, 0x0C, "Total Built",
			fmt.Sprintf("0x%02X%02X%02X%02X", d[12], d[11], d[10], d[9]),
			fmt.Sprintf("uint32 LE = %d", db.TotalBuilt)))

		// Bytes 13-16: Total remaining
		fields = append(fields, FormatFieldRaw(0x0D, 0x10, "Total Remaining",
			fmt.Sprintf("0x%02X%02X%02X%02X", d[16], d[15], d[14], d[13]),
			fmt.Sprintf("uint32 LE = %d", db.TotalRemaining)))

		// Component slots
		if len(db.Slots) > 0 {
			fields = append(fields, "")
			slotStart := 17
			slotEnd := slotStart + len(db.Slots)*4 - 1
			fields = append(fields, fmt.Sprintf("0x%02X-0x%02X: Component Slots (%d slots, 4 bytes each)",
				slotStart, slotEnd, len(db.Slots)))

			for i, slot := range db.Slots {
				offset := 17 + i*4
				if offset+4 > len(d) {
					break
				}
				rawCat := encoding.Read16(d, offset)
				rawItemId := d[offset+2]
				rawCount := d[offset+3]

				catName := getCategoryName(slot.Category)
				prefix := TreeBranch
				if i == len(db.Slots)-1 {
					prefix = TreeEnd
				}
				fields = append(fields, fmt.Sprintf("           %s Slot %d @ 0x%02X: cat=0x%04X (%s), itemId=0x%02X -> %d, count=0x%02X -> %d",
					prefix, i, offset, rawCat, catName, rawItemId, rawItemId, rawCount, rawCount))
			}
		}

		// Name starts after slots
		nameOffset := 17 + db.SlotCount*4
		if nameOffset < len(d) {
			fields = append(fields, "")
			nameLen := int(d[nameOffset])
			fields = append(fields, FormatFieldRaw(nameOffset, nameOffset, "Name Length",
				fmt.Sprintf("0x%02X", d[nameOffset]),
				fmt.Sprintf("%d bytes", nameLen)))
			if nameOffset+1+nameLen <= len(d) {
				fields = append(fields, FormatFieldRaw(nameOffset+1, nameOffset+nameLen, "Name Data",
					fmt.Sprintf("0x%s", HexDumpSingleLine(d[nameOffset+1:nameOffset+1+nameLen])),
					fmt.Sprintf("%q (Stars! encoded)", db.Name)))
			}
		}
	} else {
		// Brief design format
		fields = append(fields, "")
		fields = append(fields, "── Brief Design Data ──")

		// Bytes 4-5: Mass
		if len(d) >= 6 {
			fields = append(fields, FormatFieldRaw(0x04, 0x05, "Mass",
				fmt.Sprintf("0x%02X%02X", d[5], d[4]),
				fmt.Sprintf("uint16 LE = %d", db.Mass)))

			// Name starts at offset 6
			if len(d) > 6 {
				nameLen := int(d[6])
				fields = append(fields, FormatFieldRaw(0x06, 0x06, "Name Length",
					fmt.Sprintf("0x%02X", d[6]),
					fmt.Sprintf("%d bytes", nameLen)))
				if 7+nameLen <= len(d) {
					fields = append(fields, FormatFieldRaw(0x07, 0x06+nameLen, "Name Data",
						fmt.Sprintf("0x%s", HexDumpSingleLine(d[7:7+nameLen])),
						fmt.Sprintf("%q (Stars! encoded)", db.Name)))
				}
			}
		}
	}

	// Bug flags
	if db.ColonizerModuleBug || db.SpaceDocBug {
		fields = append(fields, "")
		fields = append(fields, "── Bug Detection ──")
		if db.ColonizerModuleBug {
			fields = append(fields, "  WARNING: ColonizerModuleBug detected")
		}
		if db.SpaceDocBug {
			fields = append(fields, "  WARNING: SpaceDocBug detected")
		}
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// getCategoryName returns a human-readable name for a slot category
func getCategoryName(cat uint16) string {
	if name, ok := itemCategoryNames[cat]; ok {
		return name
	}
	// For combined categories, list the components
	var parts []string
	for mask, name := range itemCategoryNames {
		if mask != 0 && (cat&mask) == mask {
			parts = append(parts, name)
		}
	}
	if len(parts) > 0 {
		result := parts[0]
		for i := 1; i < len(parts); i++ {
			result += "|" + parts[i]
		}
		return result
	}
	return fmt.Sprintf("0x%04X", cat)
}

// FormatDesignChange provides detailed view for DesignChangeBlock (type 27)
func FormatDesignChange(block blocks.Block, index int) string {
	width := DefaultWidth
	dcb, ok := block.(blocks.DesignChangeBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := dcb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 2 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Byte 0: Control byte - low nibble determines if deletion
	fields = append(fields, FormatFieldRaw(0x00, 0x00, "Control",
		fmt.Sprintf("0x%02X", d[0]),
		fmt.Sprintf("low nibble = %d (0=delete, other=modify)", d[0]&0x0F)))

	// Byte 1: Design info
	fields = append(fields, FormatFieldRaw(0x01, 0x01, "DesignInfo",
		fmt.Sprintf("0x%02X", d[1]),
		fmt.Sprintf("0b%08b", d[1])))

	if dcb.IsDelete {
		// Deletion
		fields = append(fields, fmt.Sprintf("           %s bits0-3: designNumber = %d",
			TreeBranch, dcb.DesignToDelete))
		fields = append(fields, fmt.Sprintf("           %s bit4: isStarbase = %v",
			TreeEnd, dcb.IsStarbase))

		fields = append(fields, "")
		fields = append(fields, "── Summary ──")
		designType := "ship"
		if dcb.IsStarbase {
			designType = "starbase"
		}
		fields = append(fields, fmt.Sprintf("  Action: DELETE %s design #%d", designType, dcb.DesignToDelete))
	} else {
		// Modification - show the embedded design
		fields = append(fields, "")
		fields = append(fields, "── Design Data (starts at offset 0x02) ──")

		if dcb.Design != nil {
			hullName := data.HullNames[dcb.Design.HullId]
			if hullName == "" {
				hullName = "Unknown"
			}

			fields = append(fields, fmt.Sprintf("  Design #%d: %q", dcb.Design.DesignNumber, dcb.Design.Name))
			fields = append(fields, fmt.Sprintf("  Hull: %s (ID=%d)", hullName, dcb.Design.HullId))
			fields = append(fields, fmt.Sprintf("  IsStarbase: %v", dcb.Design.IsStarbase))
			fields = append(fields, fmt.Sprintf("  IsFullDesign: %v", dcb.Design.IsFullDesign))

			if dcb.Design.IsFullDesign {
				fields = append(fields, fmt.Sprintf("  Armor: %d", dcb.Design.Armor))
				fields = append(fields, fmt.Sprintf("  Slots: %d", dcb.Design.SlotCount))
				fields = append(fields, fmt.Sprintf("  Turn Designed: %d (Year %d)", dcb.Design.TurnDesigned, 2400+dcb.Design.TurnDesigned))
				fields = append(fields, fmt.Sprintf("  Built: %d, Remaining: %d", dcb.Design.TotalBuilt, dcb.Design.TotalRemaining))
			} else {
				fields = append(fields, fmt.Sprintf("  Mass: %d", dcb.Design.Mass))
			}
		} else {
			fields = append(fields, "  (design data not parsed)")
		}

		fields = append(fields, "")
		fields = append(fields, "── Summary ──")
		if dcb.Design != nil {
			fields = append(fields, fmt.Sprintf("  Action: MODIFY design #%d (%q)", dcb.Design.DesignNumber, dcb.Design.Name))
		} else {
			fields = append(fields, "  Action: MODIFY design (parse error)")
		}
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
