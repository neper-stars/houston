package encoding

// Variable-length integer encoding used in Stars! files
// The encoding uses 2 bits to indicate the length:
//   00 = 0 bytes (value is 0)
//   01 = 1 byte
//   10 = 2 bytes
//   11 = 4 bytes

// VarLenByteCount returns the number of bytes needed for a variable-length encoded value
// based on the 2-bit encoding value (0-3)
func VarLenByteCount(encoded int) int {
	// Formula: 4 >> (3 - encoded), but handle 0 specially
	switch encoded {
	case 0:
		return 0
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 4
	default:
		return 0
	}
}

// ReadVarLen reads a variable-length encoded integer from bytes at the given offset
// Returns the value and the number of bytes consumed
func ReadVarLen(bytes []byte, offset int, byteCount int) (int64, int) {
	switch byteCount {
	case 0:
		return 0, 0
	case 1:
		return int64(bytes[offset] & 0xFF), 1
	case 2:
		return int64(Read16(bytes, offset)), 2
	case 4:
		return int64(Read32(bytes, offset)), 4
	default:
		return 0, 0
	}
}

// ByteLengthForInt returns the encoding value (0-3) for the number of bytes
// needed to store the given integer value
func ByteLengthForInt(value int64) int {
	if value == 0 {
		return 0
	}
	if value <= 0xFF {
		return 1
	}
	if value <= 0xFFFF {
		return 2
	}
	return 3
}

// ExtractVarLenField extracts variable-length encoded values from a contents byte
// The contents byte packs multiple 2-bit length indicators
// bitOffset is the starting bit position (0, 2, 4, or 6)
func ExtractVarLenField(contentsByte byte, bitOffset int) int {
	return int((contentsByte >> bitOffset) & 0x03)
}

// ExtractVarLenField16 extracts variable-length encoded values from a 2-byte contents word
// The contents word packs multiple 2-bit length indicators
// bitOffset is the starting bit position (0, 2, 4, 6, 8, 10, 12, or 14)
func ExtractVarLenField16(contentsWord uint16, bitOffset int) int {
	return int((contentsWord >> bitOffset) & 0x03)
}
