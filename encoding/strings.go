package encoding

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// All of these characters are found in the stars26jrc4 binary at offset
	// 000B:DD8A
	encodesOneNibble = " aehilnorst" // 0-A indexed
	encodesB         = "ABCDEFGHIJKLMNOP"
	encodesC         = "QRSTUVWXYZ012345"
	encodesD         = "6789bcdfgjkmpquv"
	encodesE         = "wxyz+-,!.?:;'*%$"
)

// DecodeHexStarsString decodes a hex-encoded Stars! string
func DecodeHexStarsString(hexChars string, byteSize int) (string, error) {
	var result strings.Builder
	// Keep track of what byte we're at for certain checks
	atByteIndex := -1

	// Loop through each hex character and decode the text depending on
	// what the hex value is. 1 Nibble (4 bits) is represented by one char
	for t := 0; t < 2*byteSize; t++ {
		// Every 2 nibbles is the start of a new byte
		// Integer division expected
		atByteIndex = t / 2
		thisNibble := hexChars[t]

		// 0-A is 1-Nibble (4-bits) encoded text
		if thisNibble <= 'A' { // ascii math FTW
			thisNibbleStr := string(thisNibble)
			charIndex, err := strconv.ParseInt(thisNibbleStr, 16, 0)
			if err != nil {
				return "", err
			}
			// This nibble is just an index in a char array
			result.WriteByte(encodesOneNibble[charIndex])
		} else if thisNibble == 'F' {
			// Three-nibble encoded text starts with an 'F'
			// We've already hit the last byte, no decodeable 3-nibble
			// chars are left (probably just junk remaining)
			if atByteIndex >= byteSize-1 {
				continue
			}

			nextNibble := hexChars[t+1]
			nextNextNibble := hexChars[t+2]

			// The encoded text is the direct ASCII value of the swapped
			// nibbles
			combinedNibbles := string(nextNextNibble) + string(nextNibble)
			parsed, err := strconv.ParseInt(combinedNibbles, 16, 0)
			if err != nil {
				return "", err
			}

			theChar := byte(parsed & 0xff)
			result.WriteByte(theChar)
			// Advance passed the two characters we decoded
			t += 2
		} else {
			// Otherwise, the next hex value is B,C,D, or E, and text is
			// 2-nibble encoded
			nextNibble := hexChars[t+1]
			nextNibbleStr := string(nextNibble)
			charIndex, err := strconv.ParseInt(nextNibbleStr, 16, 0)
			if err != nil {
				return "", err
			}

			switch thisNibble {
			case 'B':
				result.WriteByte(encodesB[charIndex])
			case 'C':
				result.WriteByte(encodesC[charIndex])
			case 'D':
				result.WriteByte(encodesD[charIndex])
			case 'E':
				result.WriteByte(encodesE[charIndex])
			}

			// Advance passed the character we decoded
			t++
		}
	}

	return result.String(), nil
}

// DecodeStarsString decodes a Stars! encoded string from raw bytes
func DecodeStarsString(res []byte) (string, error) {
	byteSize := int(res[0])
	textBytes := res[1:]
	hexChars := ByteArrayToHex(textBytes)
	decoded, err := DecodeHexStarsString(hexChars, byteSize)
	if err != nil {
		return "", err
	}
	return decoded, nil
}

// ByteToHex converts a single byte to a hex string
func ByteToHex(b byte) string {
	i := int(b & 0xff)
	hex := fmt.Sprintf("%02X", i)
	return hex
}

// ByteArrayToHex converts a byte array to a hex string
func ByteArrayToHex(bytes []byte) string {
	var hexChars strings.Builder

	for _, b := range bytes {
		hexChars.WriteString(ByteToHex(b))
	}

	return hexChars.String()
}

// HexToByteArray converts a hex string to a byte array
func HexToByteArray(hexChars string) []byte {
	res := make([]byte, len(hexChars)/2)

	for i := 0; i < len(res); i++ {
		firstChar := hexChars[2*i]
		secondChar := hexChars[2*i+1]

		highNibble := charToNibble(firstChar)
		lowNibble := charToNibble(secondChar)

		res[i] = byte(highNibble<<4 | lowNibble)
	}

	return res
}

func charToNibble(ch byte) byte {
	if ch >= '0' && ch <= '9' {
		return ch - '0'
	}
	if ch >= 'A' && ch <= 'F' {
		return ch - 'A' + 10
	}
	if ch >= 'a' && ch <= 'f' {
		return ch - 'a' + 10
	}
	panic("Invalid hex character")
}

// hexDigits for encoding
const hexDigits = "0123456789ABCDEF"

// EncodeHexStarsString encodes a string using Stars! text encoding and returns the hex-encoded string
func EncodeHexStarsString(text string) string {
	var hexChars strings.Builder

	for i := 0; i < len(text); i++ {
		thisChar := text[i]

		// Check for bad value (> 255)
		if thisChar > 255 {
			thisChar = '?'
		}

		// Check if this character is one that will be encoded with 1 nibble
		index := strings.IndexByte(encodesOneNibble, thisChar)
		if index >= 0 {
			hexChars.WriteByte(hexDigits[index])
			continue
		}

		// Check 2-nibble encodings
		index = strings.IndexByte(encodesB, thisChar)
		if index >= 0 {
			hexChars.WriteByte('B')
			hexChars.WriteByte(hexDigits[index])
			continue
		}

		index = strings.IndexByte(encodesC, thisChar)
		if index >= 0 {
			hexChars.WriteByte('C')
			hexChars.WriteByte(hexDigits[index])
			continue
		}

		index = strings.IndexByte(encodesD, thisChar)
		if index >= 0 {
			hexChars.WriteByte('D')
			hexChars.WriteByte(hexDigits[index])
			continue
		}

		index = strings.IndexByte(encodesE, thisChar)
		if index >= 0 {
			hexChars.WriteByte('E')
			hexChars.WriteByte(hexDigits[index])
			continue
		}

		// Otherwise, 3-nibble encoded (direct ASCII with swapped nibbles)
		hexChars.WriteByte('F')
		hexChars.WriteByte(hexDigits[thisChar&0x0F])
		hexChars.WriteByte(hexDigits[(thisChar&0xF0)>>4])
	}

	return hexChars.String()
}

// EncodeStarsString encodes a string using Stars! encoding and returns the byte array
func EncodeStarsString(s string) []byte {
	hexChars := EncodeHexStarsString(s)

	// Require multiple of 2 bytes and append an 'F' to make it so
	if len(hexChars)%2 != 0 {
		hexChars = hexChars + "F"
	}

	// Convert byte size to a hex string
	byteSizeHex := ByteToHex(byte(len(hexChars) / 2))

	// Add the byte size as a header to the data
	hexChars = byteSizeHex + hexChars

	return HexToByteArray(hexChars)
}
