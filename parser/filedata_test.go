package parser

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

// testXFileHex is a .x file with blocks 8,9,29,0 (FileHeader, FileHash, ProductionQueueChange, FileFooter)
const testXFileHex = "10204a334a33314e6b0b602a0100a0d00140112433f1044513bf612675ad0732" +
	"f3ccb3aca2427437b4aa5a408947e2ed85eea601782cad8786032a95fb7ccae6" +
	"63181e5be020eab3301fc5c036f5e9c3afe4936a3d0625b09f748ef373f920e2" +
	"4c60a38e577be2d14f0000"

// TestBlockParsing tests parsing blocks from a known .x file
func TestBlockParsing(t *testing.T) {
	fileBytes := encoding.HexToByteArray(testXFileHex)
	fd := FileData(fileBytes)

	blockList, err := fd.BlockList()
	if err != nil {
		t.Fatalf("BlockList() failed: %v", err)
	}

	// Verify we got 4 blocks
	if len(blockList) != 4 {
		t.Errorf("Expected 4 blocks, got %d", len(blockList))
	}

	// Verify block types
	expectedTypes := []blocks.BlockTypeID{
		blocks.FileHeaderBlockType,            // Type 8
		blocks.FileHashBlockType,              // Type 9
		blocks.ProductionQueueChangeBlockType, // Type 29
		blocks.FileFooterBlockType,            // Type 0
	}

	for i, expectedType := range expectedTypes {
		if i >= len(blockList) {
			break
		}
		block := blockList[i]
		if gb, ok := block.(blocks.GenericBlock); ok {
			if gb.Type != expectedType {
				t.Errorf("Block %d: expected type %d, got %d", i, expectedType, gb.Type)
			}
		} else if fh, ok := block.(blocks.FileHeader); ok {
			if fh.Type != expectedType {
				t.Errorf("Block %d: expected type %d, got %d", i, expectedType, fh.Type)
			}
		} else if ff, ok := block.(blocks.FileFooterBlock); ok {
			if ff.Type != expectedType {
				t.Errorf("Block %d: expected type %d, got %d", i, expectedType, ff.Type)
			}
		} else if pqc, ok := block.(blocks.ProductionQueueChangeBlock); ok {
			if pqc.Type != expectedType {
				t.Errorf("Block %d: expected type %d, got %d", i, expectedType, pqc.Type)
			}
		} else if fhb, ok := block.(blocks.FileHashBlock); ok {
			if fhb.Type != expectedType {
				t.Errorf("Block %d: expected type %d, got %d", i, expectedType, fhb.Type)
			}
		}
	}
}

// TestFileHeader tests parsing a file header block
func TestFileHeader(t *testing.T) {
	fileBytes := encoding.HexToByteArray(testXFileHex)
	fd := FileData(fileBytes)

	header, err := fd.FileHeader()
	if err != nil {
		t.Fatalf("FileHeader() failed: %v", err)
	}

	// Verify it's a FileHeader block
	if header.Type != blocks.FileHeaderBlockType {
		t.Errorf("Expected type %d, got %d", blocks.FileHeaderBlockType, header.Type)
	}

	// Verify basic properties can be accessed without error
	_ = header.GameID
	_ = header.Turn
	_ = header.PlayerIndex()
	_ = header.Salt()
	_ = header.Crippled()
}

// TestParseBlock tests parsing individual blocks at offsets
func TestParseBlock(t *testing.T) {
	// Simple test data: block type 8, size 4, 4 bytes of data
	// Header: 0x2004 = type 8 (8 << 10) + size 4
	hexChars := "04200102030400"
	fileBytes := encoding.HexToByteArray(hexChars)
	fd := FileData(fileBytes)

	block, err := fd.ParseBlock(0)
	if err != nil {
		t.Fatalf("ParseBlock(0) failed: %v", err)
	}

	if block.Type != blocks.FileHeaderBlockType {
		t.Errorf("Expected type %d, got %d", blocks.FileHeaderBlockType, block.Type)
	}

	if block.Size != 4 {
		t.Errorf("Expected size 4, got %d", block.Size)
	}

	if len(block.Data) != 4 {
		t.Errorf("Expected data length 4, got %d", len(block.Data))
	}
}

// TestMalformedBlock tests handling of malformed block data
func TestMalformedBlock(t *testing.T) {
	// Block claims to have more data than available
	// Header: 0x20FF = type 8 + size 255 (but we only have 2 bytes)
	hexChars := "FF20"
	fileBytes := encoding.HexToByteArray(hexChars)
	fd := FileData(fileBytes)

	_, err := fd.ParseBlock(0)
	require.Error(t, err, "Expected error for malformed block")
	var errMalformedBlock *ErrMalformedBlock
	require.ErrorAs(t, err, &errMalformedBlock, "Expected ErrMalformedBlock error type")
}

// TestEmptyFileData tests handling of empty file data
func TestEmptyFileData(t *testing.T) {
	fd := FileData([]byte{})

	blockList, err := fd.BlockList()
	if err != nil {
		t.Fatalf("BlockList() on empty data failed: %v", err)
	}

	if len(blockList) != 0 {
		t.Errorf("Expected 0 blocks, got %d", len(blockList))
	}
}
