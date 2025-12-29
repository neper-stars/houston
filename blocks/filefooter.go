package blocks

import (
	"github.com/neper-stars/houston/encoding"
)

// FileFooterBlock represents the end-of-file marker block (Type 0)
// It contains a checksum for the file, except for .h# files which have no checksum
type FileFooterBlock struct {
	GenericBlock
	Checksum uint16
}

// NewFileFooterBlock creates a FileFooterBlock from a GenericBlock
func NewFileFooterBlock(b GenericBlock) *FileFooterBlock {
	fb := &FileFooterBlock{
		GenericBlock: b,
	}

	// .h# files have no checksum in their footer (size 0)
	// Other files have a 2-byte checksum
	if len(b.Data) >= 2 {
		fb.Checksum = encoding.Read16(b.Data, 0)
	}

	return fb
}

// HasChecksum returns true if the footer contains a checksum
func (fb *FileFooterBlock) HasChecksum() bool {
	return len(fb.Data) >= 2
}

// Encode returns the raw footer data bytes (without the 2-byte block header).
// Returns empty slice for H files (no checksum) or 2 bytes for other file types.
func (fb *FileFooterBlock) Encode() []byte {
	if fb.Checksum == 0 && len(fb.Data) == 0 {
		// H files have no checksum
		return []byte{}
	}
	data := make([]byte, 2)
	encoding.Write16(data, 0, fb.Checksum)
	return data
}

// ComputeRaceFooter calculates the correct footer checksum for decrypted race data.
//
// The algorithm:
//  1. Take decrypted PlayerBlock data up to (but not including) the nibble-packed names
//  2. Decode the singular and plural race names to ASCII
//  3. Pad each name to 15 characters with a leading 0
//  4. Interleave the name bytes: singular[0:2], plural[0:2], singular[2:4], plural[2:4], ...
//  5. checkSum1 = XOR of all even-indexed bytes
//  6. checkSum2 = XOR of all odd-indexed bytes
//  7. Return (checkSum2 << 8) | checkSum1
func ComputeRaceFooter(decryptedData []byte, singularName, pluralName string) uint16 {
	// Find where the name data starts
	// Structure: first 8 bytes are header, then if fullDataFlag is set at byte 6 bit 2,
	// there's 0x68 bytes of full data, then playerRelations, then names
	fullDataFlag := (decryptedData[6] & 0x04) != 0
	index := 8
	if fullDataFlag {
		index = 0x70 // 112 bytes: 8 header + 0x68 (104) full data
		playerRelationsLength := int(decryptedData[index])
		index += 1 + playerRelationsLength
	}

	// Data length is everything before the names section
	dataLength := index

	// Build the checksum data array
	var dData []byte
	dData = append(dData, decryptedData[:dataLength]...)

	// Prepare singular name: leading 0, ASCII bytes, padded to 16 total
	singularOrd := make([]byte, 16)
	singularOrd[0] = 0
	for i, c := range singularName {
		if i < 15 {
			singularOrd[i+1] = byte(c)
		}
	}

	// Prepare plural name: leading 0, ASCII bytes, padded to 16 total
	pluralOrd := make([]byte, 16)
	pluralOrd[0] = 0
	for i, c := range pluralName {
		if i < 15 {
			pluralOrd[i+1] = byte(c)
		}
	}

	// Interleave: add pairs from singular, then pairs from plural
	for i := 0; i < 16; i += 2 {
		dData = append(dData, singularOrd[i], singularOrd[i+1])
		dData = append(dData, pluralOrd[i], pluralOrd[i+1])
	}

	// Compute checksums
	var checkSum1, checkSum2 byte
	for i := 0; i < len(dData); i += 2 {
		checkSum1 ^= dData[i]
	}
	for i := 1; i < len(dData); i += 2 {
		checkSum2 ^= dData[i]
	}

	return uint16(checkSum1) | uint16(checkSum2)<<8
}
