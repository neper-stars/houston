package blocks

import "github.com/neper-stars/houston/encoding"

// EncodeBlockWithHeader wraps block data with the 2-byte header.
// The header format is: type (6 bits) << 10 | size (10 bits)
// Returns the raw bytes ready for encryption (for data blocks) or direct writing (for header/footer).
func EncodeBlockWithHeader(typeID BlockTypeID, data []byte) []byte {
	size := len(data)
	if size > 1023 {
		size = 1023 // Max 10 bits
	}

	// Header: type (6 bits) << 10 | size (10 bits)
	header := (uint16(typeID) << 10) | uint16(size)

	result := make([]byte, 2+len(data))
	encoding.Write16(result, 0, header)
	copy(result[2:], data)

	return result
}
