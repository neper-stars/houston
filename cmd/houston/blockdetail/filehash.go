package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
)

func init() {
	RegisterFormatter(blocks.FileHashBlockType, FormatFileHash)
}

// FormatFileHash provides detailed view for FileHashBlock (type 9)
// Contains registration serial number and hardware fingerprint for piracy detection.
func FormatFileHash(block blocks.Block, index int) string {
	width := DefaultWidth
	fhb, ok := block.(blocks.FileHashBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := fhb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 17 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-3: lSerial - registration serial number
	fields = append(fields, FormatFieldRaw(0x00, 0x03, "lSerial",
		fmt.Sprintf("0x%02X%02X%02X%02X", d[3], d[2], d[1], d[0]),
		fmt.Sprintf("uint32 LE = %d (registration serial)", fhb.SerialNumber)))

	// Serial validation info
	quotient := fhb.SerialNumber / 1679616 // 36^4
	validQuotients := []uint32{2, 4, 6, 18, 22}
	isValid := false
	for _, v := range validQuotients {
		if quotient == v {
			isValid = true
			break
		}
	}
	validStr := "INVALID"
	if isValid {
		validStr = "VALID"
	}
	fields = append(fields, fmt.Sprintf("           %s Validation: lSerial/36^4 = %d -> %s",
		TreeEnd, quotient, validStr))

	// Bytes 4-14: pbEnv (11 bytes - used in piracy detection)
	fields = append(fields, "")
	fields = append(fields, "── pbEnv (11 bytes, used in piracy detection) ──")
	fields = append(fields, FormatFieldRaw(0x04, 0x0E, "pbEnv",
		fmt.Sprintf("0x%s", HexDumpSingleLine(d[4:15])),
		"hardware fingerprint"))

	// pbEnv components breakdown
	fields = append(fields, "")
	fields = append(fields, "── pbEnv Components ──")

	// pbEnv bytes 0-3 (block offset 4-7): Label C:
	fields = append(fields, FormatFieldRaw(0x04, 0x07, "LabelC",
		fmt.Sprintf("0x%02X%02X%02X%02X", d[7], d[6], d[5], d[4]),
		fmt.Sprintf("%q (C: volume label)", fhb.LabelC)))

	// pbEnv bytes 4-5 (block offset 8-9): C: timestamp
	fields = append(fields, FormatFieldRaw(0x08, 0x09, "TimestampC",
		fmt.Sprintf("0x%02X%02X", d[9], d[8]),
		fmt.Sprintf("uint16 LE = 0x%04X (C: volume date/time)", fhb.TimestampC)))

	// pbEnv bytes 6-8 (block offset 10-12): Label D:
	fields = append(fields, FormatFieldRaw(0x0A, 0x0C, "LabelD",
		fmt.Sprintf("0x%02X%02X%02X", d[12], d[11], d[10]),
		fmt.Sprintf("%q (D: volume label)", fhb.LabelD)))

	// pbEnv byte 9 (block offset 13): D: timestamp
	fields = append(fields, FormatFieldRaw(0x0D, 0x0D, "TimestampD",
		fmt.Sprintf("0x%02X", d[13]),
		fmt.Sprintf("0x%02X (D: volume date/time)", fhb.TimestampD)))

	// pbEnv byte 10 (block offset 14): Drive sizes
	fields = append(fields, FormatFieldRaw(0x0E, 0x0E, "DriveSizesMB",
		fmt.Sprintf("0x%02X", d[14]),
		fmt.Sprintf("%d (combined drive sizes in 100s of MB)", fhb.DriveSizesMB)))

	// Bytes 15-16: pbEnv tail (NOT used in piracy detection)
	fields = append(fields, "")
	fields = append(fields, FormatFieldRaw(0x0F, 0x10, "pbEnv tail",
		fmt.Sprintf("0x%02X%02X", d[16], d[15]),
		"(NOT used in piracy detection)"))

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  lSerial: %d (0x%08X) - %s", fhb.SerialNumber, fhb.SerialNumber, validStr))
	fields = append(fields, fmt.Sprintf("  pbEnv[0:11]: %s", fhb.HardwareHashString()))
	fields = append(fields, "")
	fields = append(fields, "── Piracy Detection ──")
	fields = append(fields, "  Players flagged as cheaters (fCheater) if:")
	fields = append(fields, "    1. Their lSerial values match (same registration)")
	fields = append(fields, "    2. Their pbEnv[0:11] values differ (different hardware)")
	fields = append(fields, "  Consequence: Tech capped at 9, production -20%, random bad events")

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
