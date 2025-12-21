package encoding

import (
	"encoding/binary"
)

// Read16 reads a little-endian uint16 from bytes at the given offset
func Read16(bytes []byte, offset int) uint16 {
	return binary.LittleEndian.Uint16(bytes[offset:])
}

// Read32 reads a little-endian uint32 from bytes at the given offset
func Read32(bytes []byte, offset int) uint32 {
	return binary.LittleEndian.Uint32(bytes[offset:])
}

// SubArray returns a slice of the input array from startIdx to endIdx (inclusive)
func SubArray(input []byte, startIdx int, endIdx int) []byte {
	size := endIdx - startIdx + 1
	output := make([]byte, size)
	copy(output, input[startIdx:endIdx+1])
	return output
}

// SubArrayFromStart returns a slice from startIdx to the end of the array
func SubArrayFromStart(input []byte, startIdx int) []byte {
	return SubArray(input, startIdx, len(input)-1)
}
