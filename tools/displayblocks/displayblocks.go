// Package displayblocks provides functionality to read and display Stars! game file blocks.
//
// This package decrypts and formats block data for debugging and analysis purposes.
// It can be used programmatically to inspect game file contents.
//
// Example usage:
//
//	blocks, err := displayblocks.ReadFile("Game.m1")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, info := range blocks {
//	    fmt.Println(info.Summary())
//	}
package displayblocks

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

// BlockInfo contains information about a single parsed block.
type BlockInfo struct {
	Index     int               // Block index in the file
	TypeID    blocks.BlockTypeID // Block type identifier
	TypeName  string            // Human-readable type name
	Size      blocks.BlockSize  // Block size in bytes
	Data      []byte            // Decrypted block data
	Block     blocks.Block      // The parsed block
	Details   map[string]any    // Type-specific parsed details
}

// Summary returns a one-line summary of the block.
func (bi *BlockInfo) Summary() string {
	return fmt.Sprintf("[%d] Block Type %d (%s), Size: %d",
		bi.Index, bi.TypeID, bi.TypeName, bi.Size)
}

// HexData returns the decrypted data as a hex string.
func (bi *BlockInfo) HexData() string {
	if len(bi.Data) == 0 {
		return "(empty)"
	}

	var sb strings.Builder
	for i, b := range bi.Data {
		if i > 0 && i%32 == 0 {
			sb.WriteString("\n")
		}
		fmt.Fprintf(&sb, "%02X ", b)
	}
	return sb.String()
}

// FileInfo contains information about a parsed Stars! file.
type FileInfo struct {
	Filename   string
	Size       int
	BlockCount int
	Blocks     []BlockInfo
	Header     *HeaderInfo
}

// HeaderInfo contains file header information.
type HeaderInfo struct {
	Magic       string
	GameID      uint32
	Version     string
	Turn        uint16
	Year        int
	PlayerIndex int
	Salt        int
	Flags       uint8
}

// ReadFile reads a Stars! file and returns information about all blocks.
func ReadFile(filename string) (*FileInfo, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ReadBytes(filename, fileBytes)
}

// ReadBytes parses Stars! file data and returns information about all blocks.
func ReadBytes(filename string, fileBytes []byte) (*FileInfo, error) {
	fd := parser.FileData(fileBytes)

	blockList, err := fd.BlockList()
	if err != nil {
		return nil, fmt.Errorf("failed to parse blocks: %w", err)
	}

	info := &FileInfo{
		Filename:   filename,
		Size:       len(fileBytes),
		BlockCount: len(blockList),
		Blocks:     make([]BlockInfo, 0, len(blockList)),
	}

	for i, block := range blockList {
		bi := extractBlockInfo(i, block)
		info.Blocks = append(info.Blocks, bi)

		// Extract header info from FileHeader block
		if h, ok := block.(blocks.FileHeader); ok {
			info.Header = &HeaderInfo{
				Magic:       h.Magic(),
				GameID:      h.GameID,
				Version:     h.VersionString(),
				Turn:        h.Turn,
				Year:        h.Year(),
				PlayerIndex: h.PlayerIndex(),
				Salt:        h.Salt(),
				Flags:       h.Flags,
			}
		}
	}

	return info, nil
}

// extractBlockInfo extracts information from a parsed block.
func extractBlockInfo(index int, block blocks.Block) BlockInfo {
	bi := BlockInfo{
		Index:   index,
		Data:    block.DecryptedData(),
		Block:   block,
		Details: make(map[string]any),
	}

	switch b := block.(type) {
	case blocks.FileHeader:
		bi.TypeID = b.Type
		bi.Size = b.Size
		bi.TypeName = blocks.BlockTypeName(b.Type)
		bi.Details["Magic"] = b.Magic()
		bi.Details["GameID"] = b.GameID
		bi.Details["Version"] = b.VersionString()
		bi.Details["Turn"] = b.Turn
		bi.Details["Year"] = b.Year()
		bi.Details["PlayerIndex"] = b.PlayerIndex()
		bi.Details["Salt"] = b.Salt()
		bi.Details["Flags"] = b.Flags

	case blocks.PlanetsBlock:
		bi.TypeID = b.Type
		bi.Size = b.Size
		bi.TypeName = blocks.BlockTypeName(b.Type)
		bi.Details["UniverseSize"] = b.UniverseSize
		bi.Details["Density"] = b.Density
		bi.Details["PlayerCount"] = b.PlayerCount
		bi.Details["PlanetCount"] = b.PlanetCount
		bi.Details["GameName"] = b.GameName
		bi.Details["GameSettings"] = b.GameSettings
		bi.Details["Planets"] = b.Planets

	case blocks.PlayerBlock:
		bi.TypeID = b.Type
		bi.Size = b.Size
		bi.TypeName = blocks.BlockTypeName(b.Type)
		bi.Details["PlayerNumber"] = b.PlayerNumber
		bi.Details["NameSingular"] = b.NameSingular
		bi.Details["NamePlural"] = b.NamePlural
		bi.Details["ShipDesignCount"] = b.ShipDesignCount
		bi.Details["StarbaseDesignCount"] = b.StarbaseDesignCount
		bi.Details["Planets"] = b.Planets
		bi.Details["Fleets"] = b.Fleets

	case blocks.PartialPlanetBlock:
		bi.TypeID = b.Type
		bi.Size = b.Size
		bi.TypeName = blocks.BlockTypeName(b.Type)
		bi.Details["PlanetNumber"] = b.PlanetNumber
		bi.Details["Owner"] = b.Owner

	case blocks.PlanetBlock:
		bi.TypeID = b.Type
		bi.Size = b.Size
		bi.TypeName = blocks.BlockTypeName(b.Type)
		bi.Details["PlanetNumber"] = b.PlanetNumber
		bi.Details["Owner"] = b.Owner

	case blocks.PartialFleetBlock:
		bi.TypeID = b.Type
		bi.Size = b.Size
		bi.TypeName = blocks.BlockTypeName(b.Type)
		bi.Details["FleetNumber"] = b.FleetNumber
		bi.Details["Owner"] = b.Owner
		bi.Details["X"] = b.X
		bi.Details["Y"] = b.Y

	case blocks.FleetBlock:
		bi.TypeID = b.Type
		bi.Size = b.Size
		bi.TypeName = blocks.BlockTypeName(b.Type)
		bi.Details["FleetNumber"] = b.FleetNumber
		bi.Details["Owner"] = b.Owner
		bi.Details["X"] = b.X
		bi.Details["Y"] = b.Y

	case blocks.DesignBlock:
		bi.TypeID = b.Type
		bi.Size = b.Size
		bi.TypeName = blocks.BlockTypeName(b.Type)
		bi.Details["DesignNumber"] = b.DesignNumber
		bi.Details["IsStarbase"] = b.IsStarbase
		bi.Details["HullId"] = b.HullId
		bi.Details["Name"] = b.Name
		bi.Details["SlotCount"] = len(b.Slots)

	case blocks.ProductionQueueBlock:
		bi.TypeID = b.Type
		bi.Size = b.Size
		bi.TypeName = blocks.BlockTypeName(b.Type)
		bi.Details["ItemCount"] = len(b.Items)

	case blocks.FileFooterBlock:
		bi.TypeID = b.Type
		bi.Size = b.Size
		bi.TypeName = blocks.BlockTypeName(b.Type)

	case blocks.GenericBlock:
		bi.TypeID = b.Type
		bi.Size = b.Size
		bi.TypeName = blocks.BlockTypeName(b.Type)

	default:
		// For other block types, get basic info via interface
		bi.TypeID = block.BlockTypeID()
		bi.Size = block.BlockSize()
		bi.TypeName = fmt.Sprintf("%T", block)
	}

	return bi
}

// FormatBlock formats a BlockInfo for display.
func FormatBlock(bi *BlockInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n[%d] Block Type %d (%s), Size: %d\n",
		bi.Index, bi.TypeID, bi.TypeName, bi.Size))

	// Add type-specific details
	for key, value := range bi.Details {
		sb.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
	}

	// Add hex data
	if len(bi.Data) > 0 {
		sb.WriteString("  Data: ")
		for i, b := range bi.Data {
			if i > 0 && i%32 == 0 {
				sb.WriteString("\n        ")
			}
			fmt.Fprintf(&sb, "%02X ", b)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// WriteBlocks writes formatted block information to a writer.
func WriteBlocks(w io.Writer, info *FileInfo) error {
	fmt.Fprintf(w, "File: %s (%d bytes, %d blocks)\n", info.Filename, info.Size, info.BlockCount)
	fmt.Fprintln(w, strings.Repeat("=", 60))

	for _, bi := range info.Blocks {
		fmt.Fprint(w, FormatBlock(&bi))
	}

	return nil
}
