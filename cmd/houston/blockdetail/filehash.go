package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.FileHashBlockType, FormatFileHash)
}

// FormatFileHash provides detailed view for FileHashBlock (type 9)
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

	// Bytes 0-1: Unknown
	unknown := encoding.Read16(d, 0)
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "Unknown",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = 0x%04X (flags/playerID?)", unknown)))

	// Bytes 2-5: Serial number
	fields = append(fields, FormatFieldRaw(0x02, 0x05, "SerialNumber",
		fmt.Sprintf("0x%02X%02X%02X%02X", d[5], d[4], d[3], d[2]),
		fmt.Sprintf("uint32 LE = %d", fhb.SerialNumber)))

	// Bytes 6-16: Hardware hash (11 bytes)
	fields = append(fields, "")
	fields = append(fields, "── Hardware Hash (11 bytes) ──")
	fields = append(fields, FormatFieldRaw(0x06, 0x10, "HardwareHash",
		fmt.Sprintf("0x%s", HexDumpSingleLine(d[6:17])),
		"machine fingerprint"))

	// Hardware hash breakdown
	fields = append(fields, "")
	fields = append(fields, "── Hash Components ──")

	// Bytes 0-3 of hash (6-9): Label C:
	fields = append(fields, FormatFieldRaw(0x06, 0x09, "LabelC",
		fmt.Sprintf("0x%02X%02X%02X%02X", d[9], d[8], d[7], d[6]),
		fmt.Sprintf("%q (C: volume label)", fhb.LabelC)))

	// Bytes 4-5 of hash (10-11): C: timestamp
	fields = append(fields, FormatFieldRaw(0x0A, 0x0B, "TimestampC",
		fmt.Sprintf("0x%02X%02X", d[11], d[10]),
		fmt.Sprintf("uint16 LE = 0x%04X (C: volume date/time)", fhb.TimestampC)))

	// Bytes 6-8 of hash (12-14): Label D:
	fields = append(fields, FormatFieldRaw(0x0C, 0x0E, "LabelD",
		fmt.Sprintf("0x%02X%02X%02X", d[14], d[13], d[12]),
		fmt.Sprintf("%q (D: volume label)", fhb.LabelD)))

	// Byte 9 of hash (15): D: timestamp
	fields = append(fields, FormatFieldRaw(0x0F, 0x0F, "TimestampD",
		fmt.Sprintf("0x%02X", d[15]),
		fmt.Sprintf("0x%02X (D: volume date/time)", fhb.TimestampD)))

	// Byte 10 of hash (16): Drive sizes
	fields = append(fields, FormatFieldRaw(0x10, 0x10, "DriveSizesMB",
		fmt.Sprintf("0x%02X", d[16]),
		fmt.Sprintf("%d (combined drive sizes in 100s of MB)", fhb.DriveSizesMB)))

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Serial Number: %d", fhb.SerialNumber))
	fields = append(fields, fmt.Sprintf("  Hardware Hash: %s", fhb.HardwareHashString()))
	fields = append(fields, "")
	fields = append(fields, "  Purpose: Detects multi-accounting (same serial")
	fields = append(fields, "  number on different machines = different hash)")

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
