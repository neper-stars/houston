package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
)

func init() {
	RegisterFormatter(blocks.FileHeaderBlockType, FormatFileHeader)
}

// FormatFileHeader provides detailed view for FileHeader block
func FormatFileHeader(block blocks.Block, index int) string {
	width := DefaultWidth
	fh, ok := block.(blocks.FileHeader)
	if !ok {
		return FormatGeneric(block, index)
	}

	data := fh.BlockData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(data, width)

	// Decode fields
	var fields []string

	// Bytes 0-3: Magic
	fields = append(fields, FormatFieldRaw(0x00, 0x03, "Magic",
		fmt.Sprintf("0x%02X%02X%02X%02X", data[0], data[1], data[2], data[3]),
		fmt.Sprintf("%q", fh.Magic())))

	// Bytes 4-7: GameID (uint32 LE)
	fields = append(fields, FormatFieldRaw(0x04, 0x07, "GameID",
		fmt.Sprintf("0x%02X%02X%02X%02X", data[7], data[6], data[5], data[4]),
		fmt.Sprintf("uint32 LE = %d", fh.GameID)))

	// Bytes 8-9: VersionData (uint16 LE)
	fields = append(fields, FormatFieldRaw(0x08, 0x09, "VersionData",
		fmt.Sprintf("0x%02X%02X", data[9], data[8]),
		fmt.Sprintf("uint16 LE = 0x%04X -> version %s", fh.VersionData, fh.VersionString())))
	fields = append(fields, fmt.Sprintf("           %s (d>>12)=major=%d, ((d>>5)&0x7F)=minor=%d, (d&0x1F)=inc=%d",
		TreeEnd, fh.VersionMajor(), fh.VersionMinor(), fh.VersionIncrement()))

	// Bytes 10-11: Turn (uint16 LE)
	fields = append(fields, FormatFieldRaw(0x0A, 0x0B, "Turn",
		fmt.Sprintf("0x%02X%02X", data[11], data[10]),
		fmt.Sprintf("uint16 LE = %d -> Year = 2400 + %d = %d", fh.Turn, fh.Turn, fh.Year())))

	// Bytes 12-13: PlayerData (uint16 LE)
	fields = append(fields, FormatFieldRaw(0x0C, 0x0D, "PlayerData",
		fmt.Sprintf("0x%02X%02X", data[13], data[12]),
		fmt.Sprintf("uint16 LE = 0x%04X -> (d>>5)=salt=%d, (d&0x1F)=playerIdx=%d",
			fh.PlayerData, fh.Salt(), fh.PlayerIndex())))

	// Byte 14: FileType
	fields = append(fields, FormatFieldRaw(0x0E, 0x0E, "FileType",
		fmt.Sprintf("0x%02X", data[14]),
		fmt.Sprintf("%d = %s", fh.FileType, fh.FileTypeName())))

	// Byte 15: Flags + wGen
	fields = append(fields, FormatFieldRaw(0x0F, 0x0F, "Flags",
		fmt.Sprintf("0x%02X", data[15]),
		fmt.Sprintf("0b%08b", fh.Flags)))
	// Decode individual flag bits
	fields = append(fields, fmt.Sprintf("           %s bit0: fDone (TurnSubmitted) = %v", TreeBranch, fh.TurnSubmitted()))
	fields = append(fields, fmt.Sprintf("           %s bit1: fInUse (HostUsing) = %v", TreeBranch, fh.HostUsing()))
	fields = append(fields, fmt.Sprintf("           %s bit2: fMulti (MultipleTurns) = %v", TreeBranch, fh.MultipleTurns()))
	fields = append(fields, fmt.Sprintf("           %s bit3: fGameOverMan (GameOver) = %v", TreeBranch, fh.GameOver()))
	fields = append(fields, fmt.Sprintf("           %s bit4: fCrippled = %v", TreeBranch, fh.Crippled()))
	// wGen field (bits 5-7)
	genValidated := ""
	if fh.IsGenerationValidated() {
		genValidated = " (VALIDATED for X files)"
	}
	fields = append(fields, fmt.Sprintf("           %s bits5-7: wGen (generation) = %d%s", TreeEnd, fh.Generation(), genValidated))

	fieldsSection := FormatFieldsSection(fields, width)

	return BuildOutput(header, hexSection, fieldsSection)
}
