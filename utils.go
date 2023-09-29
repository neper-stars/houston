package houston

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

func read16(bytes []byte, offset int) uint16 {
	return binary.LittleEndian.Uint16(bytes[offset:])
}

func read32(bytes []byte, offset int) uint32 {
	return binary.LittleEndian.Uint32(bytes[offset:])
}

const (
	encodesOneNibble = " aehilnorst"
	encodesB         = "ABCDEFGHIJKLMNOP"
	encodesC         = "QRSTUVWXYZ012345"
	encodesD         = "6789bcdfgjkmpquv"
	encodesE         = "wxyz+-,!.?:;'*%$"
)

func decodeHexStarsString(hexChars string, byteSize int) (string, error) {
	var result strings.Builder
	atByteIndex := -1

	for t := 0; t < 2*byteSize; t++ {
		atByteIndex = t / 2
		thisNibble := hexChars[t]

		if thisNibble <= 'A' { // ascii math FTW
			thisNibbleStr := string(thisNibble)
			charIndex, err := strconv.ParseInt(thisNibbleStr, 16, 0)
			if err != nil {
				return "", err
			}
			result.WriteByte(encodesOneNibble[charIndex])
		} else if thisNibble == 'F' {
			if atByteIndex >= byteSize-1 {
				continue
			}

			nextNibble := hexChars[t+1]
			nextNextNibble := hexChars[t+2]

			combinedNibbles := string(nextNibble) + string(nextNextNibble)
			parsed, err := strconv.ParseInt(combinedNibbles, 16, 0)
			if err != nil {
				return "", err
			}

			theChar := byte(parsed & 0xff)
			result.WriteByte(theChar)
			t += 2
		} else {
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

			t++
		}
	}

	return result.String(), nil
}

func decodeStarsString(res []byte) (string, error) {
	byteSize := int(res[0])
	textBytes := res[1:]
	hexChars := byteArrayToHex(textBytes)
	decoded, err := decodeHexStarsString(hexChars, byteSize)
	if err != nil {
		return "", err
	}
	return decoded, nil
}

func byteToHex(b byte) string {
	i := int(b & 0xff)
	hex := fmt.Sprintf("%02X", i)
	return hex
}

func byteArrayToHex(bytes []byte) string {
	var hexChars strings.Builder

	for _, b := range bytes {
		hexChars.WriteString(byteToHex(b))
	}

	return hexChars.String()
}

func hexToByteArray(hexChars string) []byte {
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
	panic("Invalid hex character")
}

func subArray(input []byte, startIdx int, endIdx int) []byte {
	size := endIdx - startIdx + 1
	output := make([]byte, size)
	copy(output, input[startIdx:endIdx+1])
	return output
}

func subArrayFromStart(input []byte, startIdx int) []byte {
	return subArray(input, startIdx, len(input)-1)
}
