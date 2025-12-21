package encoding

import (
	"testing"
)

func TestRead16(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		offset   int
		expected uint16
	}{
		{"zero", []byte{0x00, 0x00}, 0, 0x0000},
		{"little endian 0x1234", []byte{0x34, 0x12}, 0, 0x1234},
		{"max value", []byte{0xFF, 0xFF}, 0, 0xFFFF},
		{"with offset", []byte{0x00, 0x34, 0x12, 0x00}, 1, 0x1234},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Read16(tt.data, tt.offset)
			if result != tt.expected {
				t.Errorf("Read16(%v, %d) = %04X, want %04X", tt.data, tt.offset, result, tt.expected)
			}
		})
	}
}

func TestRead32(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		offset   int
		expected uint32
	}{
		{"zero", []byte{0x00, 0x00, 0x00, 0x00}, 0, 0x00000000},
		{"little endian 0x12345678", []byte{0x78, 0x56, 0x34, 0x12}, 0, 0x12345678},
		{"max value", []byte{0xFF, 0xFF, 0xFF, 0xFF}, 0, 0xFFFFFFFF},
		{"with offset", []byte{0x00, 0x78, 0x56, 0x34, 0x12, 0x00}, 1, 0x12345678},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Read32(tt.data, tt.offset)
			if result != tt.expected {
				t.Errorf("Read32(%v, %d) = %08X, want %08X", tt.data, tt.offset, result, tt.expected)
			}
		})
	}
}

func TestVarLenByteCount(t *testing.T) {
	tests := []struct {
		encoded  int
		expected int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{3, 4},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := VarLenByteCount(tt.encoded)
			if result != tt.expected {
				t.Errorf("VarLenByteCount(%d) = %d, want %d", tt.encoded, result, tt.expected)
			}
		})
	}
}

func TestExtractVarLenField16(t *testing.T) {
	tests := []struct {
		name      string
		value     uint16
		bitOffset int
		expected  int
	}{
		{"bits 0-1, value 0", 0x0000, 0, 0},
		{"bits 0-1, value 3", 0x0003, 0, 3},
		{"bits 2-3, value 2", 0x0008, 2, 2}, // 0b1000 >> 2 & 3 = 2
		{"bits 4-5, value 1", 0x0010, 4, 1}, // 0b10000 >> 4 & 3 = 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractVarLenField16(tt.value, tt.bitOffset)
			if result != tt.expected {
				t.Errorf("ExtractVarLenField16(%04X, %d) = %d, want %d", tt.value, tt.bitOffset, result, tt.expected)
			}
		})
	}
}

func TestReadVarLen(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		offset    int
		byteCount int
		expected  int64
	}{
		{"0 bytes", []byte{}, 0, 0, 0},
		{"1 byte", []byte{0x42}, 0, 1, 0x42},
		{"2 bytes", []byte{0x34, 0x12}, 0, 2, 0x1234},
		{"4 bytes", []byte{0x78, 0x56, 0x34, 0x12}, 0, 4, 0x12345678},
		{"with offset", []byte{0x00, 0x42}, 1, 1, 0x42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := ReadVarLen(tt.data, tt.offset, tt.byteCount)
			if result != tt.expected {
				t.Errorf("ReadVarLen(%v, %d, %d) = %d, want %d", tt.data, tt.offset, tt.byteCount, result, tt.expected)
			}
		})
	}
}

func TestRoundTripRead16(t *testing.T) {
	// Test that we can write and read back values correctly
	testValues := []uint16{0, 1, 255, 256, 1000, 65535}

	for _, val := range testValues {
		data := make([]byte, 2)
		// Write little-endian
		data[0] = byte(val & 0xFF)
		data[1] = byte((val >> 8) & 0xFF)

		result := Read16(data, 0)
		if result != val {
			t.Errorf("Round-trip failed for %d: got %d", val, result)
		}
	}
}

func TestRoundTripRead32(t *testing.T) {
	testValues := []uint32{0, 1, 255, 256, 65535, 65536, 0x12345678, 0xFFFFFFFF}

	for _, val := range testValues {
		data := make([]byte, 4)
		// Write little-endian
		data[0] = byte(val & 0xFF)
		data[1] = byte((val >> 8) & 0xFF)
		data[2] = byte((val >> 16) & 0xFF)
		data[3] = byte((val >> 24) & 0xFF)

		result := Read32(data, 0)
		if result != val {
			t.Errorf("Round-trip failed for %d: got %d", val, result)
		}
	}
}
