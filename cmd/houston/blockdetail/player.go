package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.PlayerBlockType, FormatPlayer)
}

// Auto-build item names for ZipProd display
var autoBuildNames = []string{
	"AutoMines",
	"AutoFactories",
	"AutoDefenses",
	"AutoAlchemy",
	"AutoMinTerraform",
	"AutoMaxTerraform",
	"AutoPackets",
}

// PRT names
var prtNames = []string{
	"HE (Hyper-Expansion)",
	"SS (Super Stealth)",
	"WM (War Monger)",
	"CA (Claim Adjuster)",
	"IS (Inner Strength)",
	"SD (Space Demolition)",
	"PP (Packet Physics)",
	"IT (Interstellar Traveler)",
	"AR (Alternate Reality)",
	"JoaT (Jack of all Trades)",
}

// FormatPlayer provides detailed view for PlayerBlock
func FormatPlayer(block blocks.Block, index int) string {
	width := DefaultWidth
	pb, ok := block.(blocks.PlayerBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	data := pb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(data, width)

	// Decode fields
	var fields []string

	// Bytes 0-7: Basic player info
	fields = append(fields, FormatFieldRaw(0x00, 0x00, "Player Number",
		fmt.Sprintf("0x%02X", data[0]),
		fmt.Sprintf("%d", pb.PlayerNumber)))

	fields = append(fields, FormatFieldRaw(0x01, 0x01, "Ship Design Count",
		fmt.Sprintf("0x%02X", data[1]),
		fmt.Sprintf("%d", pb.ShipDesignCount)))

	fields = append(fields, FormatFieldRaw(0x02, 0x03, "Planets",
		fmt.Sprintf("0x%02X%02X", data[3], data[2]),
		fmt.Sprintf("(d[2] + (d[3] & 0x03) << 8) = %d", pb.Planets)))

	fields = append(fields, FormatFieldRaw(0x04, 0x05, "Fleets",
		fmt.Sprintf("0x%02X%02X", data[5], data[4]),
		fmt.Sprintf("(d[4] + (d[5] & 0x03) << 8) = %d", pb.Fleets)))

	fields = append(fields, FormatFieldRaw(0x05, 0x05, "Starbase Design Count",
		fmt.Sprintf("0x%02X", data[5]),
		fmt.Sprintf("(d[5] >> 4) = %d", pb.StarbaseDesignCount)))

	fields = append(fields, FormatFieldRaw(0x06, 0x06, "Logo/Flags",
		fmt.Sprintf("0x%02X", data[6]),
		fmt.Sprintf("Logo=(d[6]>>3)=%d, FullData=(d[6]&0x04)=%v", pb.Logo, pb.FullDataFlag)))

	fields = append(fields, FormatFieldRaw(0x07, 0x07, "Byte7 (AI settings)",
		fmt.Sprintf("0x%02X", pb.Byte7),
		decodeAIByte(pb)))

	// Bytes 8-11: idPlanetHome and Rank (header fields)
	if len(data) > 11 {
		// idPlanetHome: Home planet ID (bytes 0x08-0x09, uint16 LE)
		homePlanetID := encoding.Read16(data, 8)
		fields = append(fields, FormatFieldRaw(0x08, 0x09, "idPlanetHome",
			fmt.Sprintf("0x%02X%02X", data[9], data[8]),
			fmt.Sprintf("uint16 LE = %d (Home planet ID)", homePlanetID)))

		// Rank: Player ranking position (bytes 0x0A-0x0B, uint16 LE)
		// NOTE: Decompiled source calls this "wScore" but it's actually the Rank in the UI
		rank := encoding.Read16(data, 10)
		fields = append(fields, FormatFieldRaw(0x0A, 0x0B, "Rank (wScore)",
			fmt.Sprintf("0x%02X%02X", data[11], data[10]),
			fmt.Sprintf("uint16 LE = %d (Ranking position)", rank)))
	}

	// Bytes 12-15: Password hash
	if len(data) >= 16 {
		passHash := pb.HashedPass().Uint32()
		rawHex := fmt.Sprintf("0x%02X%02X%02X%02X", data[15], data[14], data[13], data[12])
		if passHash == 0 {
			fields = append(fields, FormatFieldRaw(0x0C, 0x0F, "Password Hash", rawHex, "uint32 LE = 0 (no password)"))
		} else {
			fields = append(fields, FormatFieldRaw(0x0C, 0x0F, "Password Hash", rawHex, fmt.Sprintf("uint32 LE = 0x%08X", passHash)))
		}
	}

	// Full data section (if present)
	if pb.FullDataFlag && len(data) > 112 {
		fields = append(fields, "")
		fields = append(fields, "── Full Data Section (starts at offset 0x08) ──")

		// Habitability (bytes 8-16 relative to full data start)
		// Full data offset 0 = block offset 8
		fields = append(fields, FormatFieldRaw(0x10, 0x18, "Habitability",
			fmt.Sprintf("0x%s", HexDumpSingleLine(data[16:25])),
			fmt.Sprintf("Grav=%d-%d, Temp=%d-%d, Rad=%d-%d",
				pb.Hab.GravityLow, pb.Hab.GravityHigh,
				pb.Hab.TemperatureLow, pb.Hab.TemperatureHigh,
				pb.Hab.RadiationLow, pb.Hab.RadiationHigh)))

		// Growth rate (byte 17 in full data = offset 0x19)
		fields = append(fields, FormatFieldRaw(0x19, 0x19, "Growth Rate",
			fmt.Sprintf("0x%02X", data[25]),
			fmt.Sprintf("%d%%", pb.GrowthRate)))

		// Tech levels (bytes 18-23 in full data = offset 0x1A-0x1F)
		fields = append(fields, FormatFieldRaw(0x1A, 0x1F, "Tech Levels",
			fmt.Sprintf("0x%s", HexDumpSingleLine(data[26:32])),
			fmt.Sprintf("E=%d W=%d P=%d C=%d El=%d B=%d",
				pb.Tech.Energy, pb.Tech.Weapons, pb.Tech.Propulsion,
				pb.Tech.Construction, pb.Tech.Electronics, pb.Tech.Biotech)))

		// Research settings (bytes 56-57 = offset 0x38-0x39)
		fields = append(fields, FormatFieldRaw(0x38, 0x38, "Research %",
			fmt.Sprintf("0x%02X", data[56]),
			fmt.Sprintf("%d%%", pb.ResearchPercentage)))
		fields = append(fields, FormatFieldRaw(0x39, 0x39, "Research Fields",
			fmt.Sprintf("0x%02X", data[57]),
			fmt.Sprintf("0b%08b", data[57])))
		fields = append(fields, fmt.Sprintf("           %s bits4-7: currentField = %d (%s)",
			TreeBranch, pb.CurrentResearchField, blocks.ResearchFieldName(pb.CurrentResearchField)))
		fields = append(fields, fmt.Sprintf("           %s bits0-3: nextField = %d (%s)",
			TreeEnd, pb.NextResearchField, blocks.ResearchFieldName(pb.NextResearchField)))

		// PRT (offset 0x4C = byte 76)
		prtName := "Unknown"
		if pb.PRT >= 0 && pb.PRT < len(prtNames) {
			prtName = prtNames[pb.PRT]
		}
		fields = append(fields, FormatFieldRaw(0x4C, 0x4C, "PRT",
			fmt.Sprintf("0x%02X", data[76]),
			fmt.Sprintf("%d = %s", pb.PRT, prtName)))

		// LRT (offset 0x4E-0x4F = bytes 78-79)
		lrtRaw := encoding.Read16(data, 78)
		fields = append(fields, FormatFieldRaw(0x4E, 0x4F, "LRT Bitmask",
			fmt.Sprintf("0x%02X%02X", data[79], data[78]),
			fmt.Sprintf("uint16 LE = 0x%04X", lrtRaw)))
		if pb.LRT != 0 {
			lrtList := decodeLRT(pb.LRT)
			for _, lrt := range lrtList {
				fields = append(fields, fmt.Sprintf("           %s %s", TreeBranch, lrt))
			}
		}

		// Player Flags (offset 0x54 = byte 84)
		fields = append(fields, "")
		flagsRaw := encoding.Read16(data, 84)
		fields = append(fields, FormatFieldRaw(0x54, 0x55, "Player Flags (wFlags)",
			fmt.Sprintf("0x%02X%02X", data[85], data[84]),
			fmt.Sprintf("uint16 LE = 0x%04X", flagsRaw)))
		fields = append(fields, formatPlayerFlagsDetailed(pb.Flags)...)

		// ZipProd Queue (offset 0x56 = byte 86)
		fields = append(fields, "")
		zipStart := 86
		fields = append(fields, FormatFieldRaw(0x56, 0x6F, "ZipProd Default Queue",
			fmt.Sprintf("0x%s", HexDumpSingleLine(data[zipStart:zipStart+26])),
			""))
		noResearchStr := "Contribute to Research"
		if pb.ZipProdDefault.NoResearch {
			noResearchStr = "Don't contribute to Research"
		}
		fields = append(fields, fmt.Sprintf("           %s Byte 0 (fNoResearch): 0x%02X -> %s", TreeBranch, data[zipStart], noResearchStr))
		fields = append(fields, fmt.Sprintf("           %s Byte 1 (Count): 0x%02X -> %d items", TreeBranch, data[zipStart+1], len(pb.ZipProdDefault.Items)))

		for i, item := range pb.ZipProdDefault.Items {
			itemName := "Unknown"
			if int(item.ItemType) < len(autoBuildNames) {
				itemName = autoBuildNames[item.ItemType]
			}
			// Raw uint16 value
			rawVal := encoding.Read16(data, zipStart+2+i*2)
			prefix := TreeBranch
			if i == len(pb.ZipProdDefault.Items)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("           %s Item %d: 0x%04X -> (val & 0x3F)=%d (%s), (val >> 6)=%d",
				prefix, i, rawVal, item.ItemType, itemName, item.Quantity))
		}

		// Player Relations (offset 0x70 = byte 112)
		if len(pb.PlayerRelations) > 0 {
			fields = append(fields, "")
			relStart := 112
			fields = append(fields, FormatFieldRaw(0x70, 0x70+len(pb.PlayerRelations),
				"Player Relations",
				fmt.Sprintf("0x%s", HexDumpSingleLine(data[relStart:relStart+1+len(pb.PlayerRelations)])),
				fmt.Sprintf("len=%d", len(pb.PlayerRelations))))
			for i, rel := range pb.PlayerRelations {
				relName := "Neutral"
				switch rel {
				case 1:
					relName = "Friend"
				case 2:
					relName = "Enemy"
				}
				prefix := TreeBranch
				if i == len(pb.PlayerRelations)-1 {
					prefix = TreeEnd
				}
				fields = append(fields, fmt.Sprintf("           %s [%d]: 0x%02X -> Player %d = %s", prefix, i, rel, i, relName))
			}
		}
	}

	// Race names (at end of block)
	fields = append(fields, "")
	fields = append(fields, "── Race Names (nibble-encoded at end of block) ──")
	fields = append(fields, fmt.Sprintf("  Singular: %q", pb.NameSingular))
	fields = append(fields, fmt.Sprintf("  Plural:   %q", pb.NamePlural))

	fieldsSection := FormatFieldsSection(fields, width)
	unknownSection := FormatUnknownSection(nil, width)

	return BuildOutput(header, hexSection, fieldsSection, unknownSection)
}

// decodeAIByte decodes byte 7 AI settings to human readable string
func decodeAIByte(pb blocks.PlayerBlock) string {
	if pb.AIEnabled {
		skillNames := []string{"Easy", "Standard", "Harder", "Expert"}
		skill := "Unknown"
		if pb.AISkill >= 0 && pb.AISkill < len(skillNames) {
			skill = skillNames[pb.AISkill]
		}
		return fmt.Sprintf("AI=%v, Skill=(d[7]>>2&3)=%d (%s)", pb.AIEnabled, pb.AISkill, skill)
	}
	if pb.IsHumanInactive() {
		return "Human (Inactive) - byte=0xE3"
	}
	return "Human player"
}

// formatPlayerFlagsDetailed formats player flags with detailed consequences
func formatPlayerFlagsDetailed(flags blocks.PlayerFlags) []string {
	var lines []string

	// Check if any flags are set
	hasFlags := flags.Dead || flags.Crippled || flags.Cheater || flags.Learned || flags.Hacker
	if !hasFlags {
		lines = append(lines, fmt.Sprintf("           %s (no flags set)", TreeEnd))
		return lines
	}

	// Collect active flags for proper tree formatting
	type flagInfo struct {
		name    string
		details []string
	}
	var activeFlags []flagInfo

	if flags.Dead {
		activeFlags = append(activeFlags, flagInfo{
			name:    "bit0: fDead - Player eliminated",
			details: []string{"Player has been removed from the game"},
		})
	}
	if flags.Crippled {
		activeFlags = append(activeFlags, flagInfo{
			name:    "bit1: fCrippled - Unregistered shareware",
			details: []string{"Tech levels capped at 9 (vs 25 normal)"},
		})
	}
	if flags.Cheater {
		activeFlags = append(activeFlags, flagInfo{
			name: "bit2: fCheater - Serial piracy detected",
			details: []string{
				"Same registration used on different hardware",
				"Tech levels capped at 9",
				"Production reduced to 80%",
				"~75% chance of random negative events",
			},
		})
	}
	if flags.Learned {
		activeFlags = append(activeFlags, flagInfo{
			name:    "bit3: fLearned - Deprecated (cleared on load)",
			details: nil,
		})
	}
	if flags.Hacker {
		activeFlags = append(activeFlags, flagInfo{
			name: "bit4: fHacker - Race values corrected",
			details: []string{
				"Growth rate degraded until race value >= 500",
				"Tech levels may be zeroed if still invalid",
				"Does NOT cap tech levels",
			},
		})
	}

	// Format with tree structure
	for i, f := range activeFlags {
		prefix := TreeBranch
		if i == len(activeFlags)-1 && len(f.details) == 0 {
			prefix = TreeEnd
		}
		lines = append(lines, fmt.Sprintf("           %s %s", prefix, f.name))

		for j, detail := range f.details {
			detailPrefix := "│  " + TreeBranch
			if i == len(activeFlags)-1 {
				detailPrefix = "   " + TreeBranch
			}
			if j == len(f.details)-1 {
				if i == len(activeFlags)-1 {
					detailPrefix = "   " + TreeEnd
				} else {
					detailPrefix = "│  " + TreeEnd
				}
			}
			lines = append(lines, fmt.Sprintf("           %s %s", detailPrefix, detail))
		}
	}

	return lines
}

// decodeLRT decodes LRT bitmask to a list of trait names
func decodeLRT(lrt uint16) []string {
	lrtNames := []string{
		"bit0: IFE (Improved Fuel Efficiency)",
		"bit1: TT (Total Terraforming)",
		"bit2: ARM (Advanced Remote Mining)",
		"bit3: ISB (Improved Starbases)",
		"bit4: GR (Generalised Research)",
		"bit5: UR (Ultimate Recycling)",
		"bit6: MA (Mineral Alchemy)",
		"bit7: NRSE (No Ram Scoop Engines)",
		"bit8: CE (Cheap Engines)",
		"bit9: OBRM (Only Basic Remote Mining)",
		"bit10: NAS (No Advanced Scanners)",
		"bit11: LSP (Low Starting Population)",
		"bit12: BET (Bleeding Edge Technology)",
		"bit13: RS (Regenerating Shields)",
	}

	var result []string
	for i, name := range lrtNames {
		if lrt&(1<<i) != 0 {
			result = append(result, name)
		}
	}
	return result
}
