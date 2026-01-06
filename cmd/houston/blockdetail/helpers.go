package blockdetail

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Box drawing characters
const (
	BoxHoriz       = "─"
	BoxVert        = "│"
	BoxTopLeft     = "┌"
	BoxTopRight    = "┐"
	BoxBottomLeft  = "└"
	BoxBottomRight = "┘"
	BoxVertRight   = "├"
	BoxVertLeft    = "┤"
	BoxHorizDown   = "┬"
	BoxHorizUp     = "┴"
	BoxCross       = "┼"

	// Tree drawing
	TreeBranch = "├─"
	TreeEnd    = "└─"
	TreeVert   = "│ "
	TreeSpace  = "  "
)

// DefaultWidth is the default box width for formatted output
const DefaultWidth = 100

// BoxLine creates a horizontal line with optional title
func BoxLine(width int, left, right, fill string, title string) string {
	if title == "" {
		return left + strings.Repeat(fill, width-2) + right
	}
	titlePart := " " + title + " "
	fillLen := width - 2 - len(titlePart)
	if fillLen < 0 {
		fillLen = 0
	}
	return left + titlePart + strings.Repeat(fill, fillLen) + right
}

// BoxTop creates the top border of a box
func BoxTop(width int, title string) string {
	return BoxLine(width, BoxTopLeft, BoxTopRight, BoxHoriz, title)
}

// BoxBottom creates the bottom border of a box
func BoxBottom(width int) string {
	return BoxLine(width, BoxBottomLeft, BoxBottomRight, BoxHoriz, "")
}

// BoxSeparator creates a horizontal separator within a box
func BoxSeparator(width int, title string) string {
	return BoxLine(width, BoxVertRight, BoxVertLeft, BoxHoriz, title)
}

// BoxContent creates a content line with vertical borders
func BoxContent(width int, content string) string {
	// Pad or truncate content to fit width
	// Use RuneCountInString for proper Unicode character counting
	contentWidth := width - 4 // Account for "│ " on each side
	runeCount := utf8.RuneCountInString(content)
	if runeCount > contentWidth {
		// Truncate by runes, not bytes
		runes := []rune(content)
		content = string(runes[:contentWidth])
		runeCount = contentWidth
	}
	return BoxVert + " " + content + strings.Repeat(" ", contentWidth-runeCount) + " " + BoxVert
}

// HexDump formats bytes as a hex dump with offsets
func HexDump(data []byte, startOffset int, bytesPerLine int) []string {
	var lines []string
	for i := 0; i < len(data); i += bytesPerLine {
		end := i + bytesPerLine
		if end > len(data) {
			end = len(data)
		}

		// Offset
		offset := fmt.Sprintf("0x%02X:", startOffset+i)

		// Hex bytes with space in middle
		hexPart := ""
		for j, b := range data[i:end] {
			if j == bytesPerLine/2 {
				hexPart += " "
			}
			hexPart += fmt.Sprintf("%02X ", b)
		}

		// Pad if line is short
		expectedLen := bytesPerLine*3 + 1 // 3 chars per byte + middle space
		if len(hexPart) < expectedLen {
			hexPart += strings.Repeat(" ", expectedLen-len(hexPart))
		}

		lines = append(lines, fmt.Sprintf("%s %s", offset, hexPart))
	}
	return lines
}

// HexDumpSingleLine formats bytes as a single hex string
func HexDumpSingleLine(data []byte) string {
	return strings.ToUpper(hex.EncodeToString(data))
}

// FormatField formats a field with offset range and description
func FormatField(startOffset, endOffset int, name string, value interface{}) string {
	if startOffset == endOffset {
		return fmt.Sprintf("0x%02X:      %s = %v", startOffset, name, value)
	}
	return fmt.Sprintf("0x%02X-0x%02X: %s = %v", startOffset, endOffset, name, value)
}

// FormatFieldRaw formats a field with offset, raw hex value, and decoded interpretation
func FormatFieldRaw(startOffset, endOffset int, name string, rawHex string, decoded string) string {
	if startOffset == endOffset {
		return fmt.Sprintf("0x%02X:      %s = %s -> %s", startOffset, name, rawHex, decoded)
	}
	return fmt.Sprintf("0x%02X-0x%02X: %s = %s -> %s", startOffset, endOffset, name, rawHex, decoded)
}

// FormatUnknown formats an unknown field
func FormatUnknown(startOffset, endOffset int, data []byte) string {
	hexStr := HexDumpSingleLine(data)
	if len(hexStr) > 16 {
		hexStr = hexStr[:16] + "..."
	}
	if startOffset == endOffset {
		return fmt.Sprintf("0x%02X:      ??? = 0x%s (TBD)", startOffset, hexStr)
	}
	return fmt.Sprintf("0x%02X-0x%02X: ??? = 0x%s (TBD)", startOffset, endOffset, hexStr)
}

// TreeItem formats an item with tree prefix
func TreeItem(prefix string, isLast bool, content string) string {
	if isLast {
		return prefix + TreeEnd + content
	}
	return prefix + TreeBranch + content
}

// TreeChildPrefix returns the prefix for child items
func TreeChildPrefix(prefix string, isLast bool) string {
	if isLast {
		return prefix + TreeSpace
	}
	return prefix + TreeVert
}
