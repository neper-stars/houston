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

// WriteVarLen writes a variable-length encoded integer and returns the new index.
// The number of bytes written is determined by ByteLengthForInt:
//   - 0 bytes if value is 0
//   - 1 byte if value <= 0xFF
//   - 2 bytes if value <= 0xFFFF
//   - 4 bytes otherwise
func WriteVarLen(data []byte, index int, value int64) int {
	byteLen := ByteLengthForInt(value)
	switch byteLen {
	case 0:
		// Zero uses no bytes
		return index
	case 1:
		data[index] = byte(value & 0xFF)
		return index + 1
	case 2:
		Write16(data, index, uint16(value))
		return index + 2
	case 3: // 3 means 4 bytes
		Write32(data, index, uint32(value))
		return index + 4
	}
	return index
}

// PackVarLenIndicators packs multiple 2-bit length indicators into a uint16.
// Each value's byte length is determined and packed at 2-bit intervals.
// Up to 8 values can be packed (16 bits / 2 bits per value).
func PackVarLenIndicators(values ...int64) uint16 {
	var result uint16
	for i, v := range values {
		if i >= 8 {
			break
		}
		result |= uint16(ByteLengthForInt(v)) << (i * 2)
	}
	return result
}

// WriteVarLenFixedSize writes a variable-length encoded integer using exactly the
// specified number of bytes. This is useful for in-place updates where the byte
// count must match the original encoding to preserve block layout.
// If the value doesn't fit in the specified byte count, it will be truncated.
func WriteVarLenFixedSize(data []byte, index int, value int64, byteCount int) {
	switch byteCount {
	case 0:
		// Zero bytes - nothing to write
	case 1:
		data[index] = byte(value & 0xFF)
	case 2:
		Write16(data, index, uint16(value&0xFFFF))
	case 4:
		Write32(data, index, uint32(value&0xFFFFFFFF))
	}
}
