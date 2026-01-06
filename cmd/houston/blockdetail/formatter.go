package blockdetail

import (
	"fmt"
	"strings"

	"github.com/neper-stars/houston/blocks"
)

// DetailFormatter is a function that formats a block in detailed view
type DetailFormatter func(block blocks.Block, index int) string

// formatters maps block type IDs to their detailed formatters
var formatters = make(map[blocks.BlockTypeID]DetailFormatter)

// RegisterFormatter registers a formatter for a block type
func RegisterFormatter(typeID blocks.BlockTypeID, formatter DetailFormatter) {
	formatters[typeID] = formatter
}

// FormatDetailed formats a block with detailed ASCII view
func FormatDetailed(block blocks.Block, index int) string {
	typeID := block.BlockTypeID()

	if formatter, ok := formatters[typeID]; ok {
		return formatter(block, index)
	}

	// Fallback to generic formatter
	return FormatGeneric(block, index)
}

// FormatBlockHeader creates the standard block header
func FormatBlockHeader(block blocks.Block, index int, width int) []string {
	typeID := block.BlockTypeID()
	typeName := blocks.BlockTypeName(typeID)
	size := block.BlockSize()

	var lines []string
	lines = append(lines, BoxTop(width, fmt.Sprintf("Block %d: %s", index, typeName)))
	lines = append(lines, BoxContent(width, fmt.Sprintf("Type: %d (%s), Size: %d bytes", typeID, typeName, size)))
	return lines
}

// FormatHexSection creates a hex dump section
func FormatHexSection(data []byte, width int) []string {
	var lines []string
	lines = append(lines, BoxSeparator(width, "Hex Dump"))

	if len(data) == 0 {
		lines = append(lines, BoxContent(width, "(no data)"))
		return lines
	}

	hexLines := HexDump(data, 0, 16)
	for _, hexLine := range hexLines {
		lines = append(lines, BoxContent(width, hexLine))
	}
	return lines
}

// FormatFieldsSection creates a decoded fields section
func FormatFieldsSection(fields []string, width int) []string {
	var lines []string
	lines = append(lines, BoxSeparator(width, "Decoded Fields"))

	for _, field := range fields {
		lines = append(lines, BoxContent(width, field))
	}
	return lines
}

// FormatUnknownSection creates an unknown regions section
func FormatUnknownSection(unknowns []string, width int) []string {
	if len(unknowns) == 0 {
		return nil
	}

	var lines []string
	lines = append(lines, BoxSeparator(width, "Unknown Regions"))

	for _, unknown := range unknowns {
		lines = append(lines, BoxContent(width, unknown))
	}
	return lines
}

// BuildOutput combines all sections and adds the bottom border
func BuildOutput(sections ...[]string) string {
	var allLines []string
	for _, section := range sections {
		allLines = append(allLines, section...)
	}
	allLines = append(allLines, BoxBottom(DefaultWidth))
	return strings.Join(allLines, "\n") + "\n"
}
