package houston

import "encoding/binary"

func read16(bytes []byte, offset int) uint16 {
	return binary.LittleEndian.Uint16(bytes[offset:])
}

func read32(bytes []byte, offset int) uint32 {
	return binary.LittleEndian.Uint32(bytes[offset:])
}
