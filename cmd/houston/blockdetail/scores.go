package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.PlayerScoresBlockType, FormatPlayerScores)
}

// FormatPlayerScores provides detailed view for PlayerScoresBlock (type 45)
func FormatPlayerScores(block blocks.Block, index int) string {
	width := DefaultWidth
	psb, ok := block.(blocks.PlayerScoresBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := psb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 24 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: Player ID and flags
	word0 := encoding.Read16(d, 0)
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "PlayerID/Flags",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = 0x%04X, playerID=(d&0x0F)=%d -> Player %d",
			word0, psb.PlayerID, psb.PlayerID+1)))

	// Bytes 2-3: Turn number
	fields = append(fields, FormatFieldRaw(0x02, 0x03, "Turn",
		fmt.Sprintf("0x%02X%02X", d[3], d[2]),
		fmt.Sprintf("uint16 LE = %d -> Year %d", psb.Turn, 2400+psb.Turn)))

	// Bytes 4-5: Score
	fields = append(fields, FormatFieldRaw(0x04, 0x05, "Score",
		fmt.Sprintf("0x%02X%02X", d[5], d[4]),
		fmt.Sprintf("uint16 LE = %d", psb.Score)))

	// Bytes 6-7: Padding
	fields = append(fields, FormatFieldRaw(0x06, 0x07, "Padding",
		fmt.Sprintf("0x%02X%02X", d[7], d[6]),
		"(always 0)"))

	// Bytes 8-11: Resources (32-bit)
	fields = append(fields, FormatFieldRaw(0x08, 0x0B, "Resources",
		fmt.Sprintf("0x%02X%02X%02X%02X", d[11], d[10], d[9], d[8]),
		fmt.Sprintf("uint32 LE = %d", psb.Resources)))

	// Bytes 12-13: Planets
	fields = append(fields, FormatFieldRaw(0x0C, 0x0D, "Planets",
		fmt.Sprintf("0x%02X%02X", d[13], d[12]),
		fmt.Sprintf("uint16 LE = %d", psb.Planets)))

	// Bytes 14-15: Starbases
	fields = append(fields, FormatFieldRaw(0x0E, 0x0F, "Starbases",
		fmt.Sprintf("0x%02X%02X", d[15], d[14]),
		fmt.Sprintf("uint16 LE = %d", psb.Starbases)))

	// Bytes 16-17: Unarmed ships
	fields = append(fields, FormatFieldRaw(0x10, 0x11, "UnarmedShips",
		fmt.Sprintf("0x%02X%02X", d[17], d[16]),
		fmt.Sprintf("uint16 LE = %d", psb.UnarmedShips)))

	// Bytes 18-19: Escort ships
	fields = append(fields, FormatFieldRaw(0x12, 0x13, "EscortShips",
		fmt.Sprintf("0x%02X%02X", d[19], d[18]),
		fmt.Sprintf("uint16 LE = %d", psb.EscortShips)))

	// Bytes 20-21: Capital ships
	fields = append(fields, FormatFieldRaw(0x14, 0x15, "CapitalShips",
		fmt.Sprintf("0x%02X%02X", d[21], d[20]),
		fmt.Sprintf("uint16 LE = %d", psb.CapitalShips)))

	// Bytes 22-23: Tech levels (sum)
	fields = append(fields, FormatFieldRaw(0x16, 0x17, "TechLevels",
		fmt.Sprintf("0x%02X%02X", d[23], d[22]),
		fmt.Sprintf("uint16 LE = %d (sum of all tech levels)", psb.TechLevels)))

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Player %d, Year %d, Score: %d",
		psb.PlayerID+1, 2400+psb.Turn, psb.Score))
	fields = append(fields, fmt.Sprintf("  Resources: %d, Planets: %d, Starbases: %d",
		psb.Resources, psb.Planets, psb.Starbases))
	fields = append(fields, fmt.Sprintf("  Ships: %d unarmed, %d escort, %d capital",
		psb.UnarmedShips, psb.EscortShips, psb.CapitalShips))

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
