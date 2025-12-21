package encoding

import (
	"testing"
)

func TestDecodeEncodeStarsString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple alphanumeric", "abcdefABCDEF1234567890"},
		{"single char a", "a"},
		{"triple e", "eee"},
		{"five i", "iiiii"},
		{"seven o", "ooooooo"},
		{"nine r", "rrrrrrrrr"},
		{"default 2-nibble chars", " aehilnorstABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789bcdfgjkmpquvwxyz+-,!.?:;'*%$"},
		{"special chars forcing 3-nibble", "~@#^&*()_+"},
		{"mixed case letters", "HelloWorld"},
		{"with spaces", "Hello World Test"},
		{"numbers only", "1234567890"},
		{"ship name style", "Scout #1"},
		{"planet name", "Alpha Centauri"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode the string
			encoded := EncodeStarsString(tt.input)

			// Decode it back
			decoded, err := DecodeStarsString(encoded)
			if err != nil {
				t.Fatalf("DecodeStarsString failed: %v", err)
			}

			if decoded != tt.input {
				t.Errorf("Round-trip failed:\n  input:   %q\n  decoded: %q", tt.input, decoded)
			}
		})
	}
}

func TestDecodeHexStarsString(t *testing.T) {
	tests := []struct {
		name     string
		hexChars string
		byteSize int
		expected string
	}{
		// Single nibble characters: space=0, a=1, e=2, h=3, i=4, l=5, n=6, o=7, r=8, s=9, t=A
		{"single a", "1F", 1, "a"},
		{"space", "0F", 1, " "},
		{"aei", "124F", 2, "aei"},
		{"hello", "32557F", 3, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := DecodeHexStarsString(tt.hexChars, tt.byteSize)
			if err != nil {
				t.Fatalf("DecodeHexStarsString failed: %v", err)
			}

			if decoded != tt.expected {
				t.Errorf("DecodeHexStarsString(%q, %d) = %q, want %q",
					tt.hexChars, tt.byteSize, decoded, tt.expected)
			}
		})
	}
}

func TestEncodeHexStarsString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Single nibble: space=0, a=1, e=2, h=3, i=4, l=5, n=6, o=7, r=8, s=9, t=A
		{"single a", "a", "1"},
		{"space", " ", "0"},
		{"aei", "aei", "124"},
		{"hello", "hello", "32557"},
		// 2-nibble: B=A-P, C=Q-Z+0-5, D=6-9+bcdfgjkmpquv, E=wxyz+-,!.?:;'*%$
		{"uppercase A", "A", "B0"},
		{"uppercase Z", "Z", "C9"},
		{"digit 6", "6", "D0"},
		{"plus sign", "+", "E4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeHexStarsString(tt.input)

			if result != tt.expected {
				t.Errorf("EncodeHexStarsString(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestHexConversions(t *testing.T) {
	tests := []struct {
		name  string
		bytes []byte
		hex   string
	}{
		{"empty", []byte{}, ""},
		{"single byte", []byte{0x5A}, "5A"},
		{"two bytes", []byte{0x12, 0x34}, "1234"},
		{"all zeros", []byte{0x00, 0x00}, "0000"},
		{"all ones", []byte{0xFF, 0xFF}, "FFFF"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ByteArrayToHex
			hex := ByteArrayToHex(tt.bytes)
			if hex != tt.hex {
				t.Errorf("ByteArrayToHex(%v) = %q, want %q", tt.bytes, hex, tt.hex)
			}

			// Test HexToByteArray (round-trip)
			if len(tt.hex) > 0 {
				bytes := HexToByteArray(tt.hex)
				if len(bytes) != len(tt.bytes) {
					t.Errorf("HexToByteArray(%q) length = %d, want %d", tt.hex, len(bytes), len(tt.bytes))
				}
				for i := range bytes {
					if bytes[i] != tt.bytes[i] {
						t.Errorf("HexToByteArray(%q)[%d] = %02X, want %02X", tt.hex, i, bytes[i], tt.bytes[i])
					}
				}
			}
		})
	}
}

func TestByteToHex(t *testing.T) {
	tests := []struct {
		input    byte
		expected string
	}{
		{0x00, "00"},
		{0x0F, "0F"},
		{0x10, "10"},
		{0x5A, "5A"},
		{0xFF, "FF"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := ByteToHex(tt.input)
			if result != tt.expected {
				t.Errorf("ByteToHex(%02X) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
