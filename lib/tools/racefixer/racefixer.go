// Package racefixer provides functionality to repair corrupted Stars! race files
// and remove passwords from race files.
//
// Race files (.r1-.r16) can become corrupted if edited improperly. This package
// provides functions to analyze race files, recalculate their integrity checksum,
// and remove passwords.
//
// Race file structure:
//   - FileHeader (Type 8): Standard 16-byte header with salt value
//   - PlayerBlock (Type 6): Encrypted race data
//   - FileFooter (Type 0): 2-byte checksum derived from race settings
//
// The checksum algorithm:
//  1. Take decrypted PlayerBlock data up to (but not including) the nibble-packed names
//  2. Decode the singular and plural race names to ASCII
//  3. Pad each name to 15 characters with a leading 0
//  4. Interleave the name bytes: singular[0:2], plural[0:2], singular[2:4], plural[2:4], ...
//  5. checkSum1 = XOR of all even-indexed bytes
//  6. checkSum2 = XOR of all odd-indexed bytes
//
// The library operates entirely in memory - callers are responsible for reading files
// from and writing files to their storage (disk, database, etc.).
//
// Example usage:
//
//	data, _ := os.ReadFile("MyRace.r1")
//	info, err := racefixer.AnalyzeBytes("MyRace.r1", data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if info.NeedsRepair {
//	    repaired, result, err := racefixer.RepairBytes(data)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    os.WriteFile("MyRace.r1", repaired, 0644)
//	}
package racefixer

import (
	"fmt"
	"io"
	"os"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/crypto"
	"github.com/neper-stars/houston/parser"
	"github.com/neper-stars/houston/store"
)

// FileInfo contains information about a race file.
type FileInfo struct {
	Filename       string
	Size           int
	BlockCount     int
	HasHashBlock   bool
	NeedsRepair    bool
	HasPassword    bool
	Blocks         []blocks.Block
	CurrentFooter  uint16 // The footer value currently in the file
	ExpectedFooter uint16 // The computed correct footer value
	SingularName   string // Race singular name (e.g., "Humanoid")
	PluralName     string // Race plural name (e.g., "Humanoids")
}

// Analyze reads a race file and determines if it needs repair.
// This is a convenience function that reads from disk.
func Analyze(filename string) (*FileInfo, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return AnalyzeBytes(filename, fileBytes)
}

// AnalyzeReader analyzes race file data from an io.Reader.
func AnalyzeReader(name string, r io.Reader) (*FileInfo, error) {
	fileBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	return AnalyzeBytes(name, fileBytes)
}

// AnalyzeBytes analyzes race file data.
// The name parameter is used for display purposes only.
func AnalyzeBytes(name string, fileBytes []byte) (*FileInfo, error) {
	fd := parser.FileData(fileBytes)

	blockList, err := fd.BlockList()
	if err != nil {
		return nil, fmt.Errorf("failed to parse blocks: %w", err)
	}

	info := &FileInfo{
		Filename:   name,
		Size:       len(fileBytes),
		BlockCount: len(blockList),
		Blocks:     blockList,
	}

	// Look for hash block, footer, and player block
	var footerData []byte
	var playerBlock *blocks.PlayerBlock

	for _, block := range blockList {
		if _, ok := block.(blocks.FileHashBlock); ok {
			info.HasHashBlock = true
		}

		if block.BlockTypeID() == blocks.FileFooterBlockType {
			footerData = block.BlockData()
		}

		// Get PlayerBlock
		if pb, ok := block.(*blocks.PlayerBlock); ok {
			playerBlock = pb
		} else if pb, ok := block.(blocks.PlayerBlock); ok {
			playerBlock = &pb
		}
	}

	// If we have the required data, compute the expected footer
	if playerBlock != nil && len(footerData) >= 2 {
		// Get decrypted data from the PlayerBlock
		decryptedData := playerBlock.DecryptedData()

		// Check for password (bytes 12-15)
		if len(decryptedData) >= 16 {
			hasPass := decryptedData[12] != 0 || decryptedData[13] != 0 ||
				decryptedData[14] != 0 || decryptedData[15] != 0
			info.HasPassword = hasPass
		}

		// Get race names from the already-parsed PlayerBlock
		info.SingularName = playerBlock.NameSingular
		info.PluralName = playerBlock.NamePlural

		// Compute expected footer
		info.ExpectedFooter = store.ComputeRaceFooter(decryptedData, info.SingularName, info.PluralName)

		// Get current footer from file
		info.CurrentFooter = uint16(footerData[0]) | uint16(footerData[1])<<8

		// Check if repair is needed
		info.NeedsRepair = info.CurrentFooter != info.ExpectedFooter
	}

	return info, nil
}

// RepairResult contains the results of a repair operation.
type RepairResult struct {
	Success         bool
	Message         string
	OldFooter       uint16
	NewFooter       uint16
	FooterChanged   bool
	PasswordRemoved bool
}

// RepairBytes attempts to fix corrupted race file data and returns the repaired bytes.
// Returns the repaired data or an error if repair is not possible.
func RepairBytes(data []byte) ([]byte, *RepairResult, error) {
	info, err := AnalyzeBytes("", data)
	if err != nil {
		return nil, nil, err
	}

	result := &RepairResult{
		OldFooter: info.CurrentFooter,
		NewFooter: info.ExpectedFooter,
	}

	if info.HasHashBlock {
		// Files with FileHashBlock (like X files) use a different algorithm
		result.Message = "file uses FileHashBlock - not a race file format"
		return data, result, nil
	}

	if !info.NeedsRepair {
		result.Success = true
		result.Message = "file checksum is already correct"
		return data, result, nil
	}

	// Find and update the footer in the file
	repaired := make([]byte, len(data))
	copy(repaired, data)

	// Find footer block offset
	offset := 0
	for offset < len(repaired)-2 {
		header := uint16(repaired[offset]) | uint16(repaired[offset+1])<<8
		blockType := blocks.BlockTypeID(header >> 10)
		blockSize := int(header & 0x3FF)

		if blockType == blocks.FileFooterBlockType {
			// Found footer block - update the data
			if offset+2+blockSize <= len(repaired) && blockSize == 2 {
				repaired[offset+2] = byte(info.ExpectedFooter & 0xFF)
				repaired[offset+3] = byte(info.ExpectedFooter >> 8)
				result.Success = true
				result.FooterChanged = true
				result.Message = fmt.Sprintf("footer updated from 0x%04X to 0x%04X",
					info.CurrentFooter, info.ExpectedFooter)
				return repaired, result, nil
			}
		}

		offset += 2 + blockSize
	}

	result.Message = "could not locate footer block in file"
	return data, result, nil
}

// RemovePasswordBytes removes the password from a race file.
// Returns the modified data with password removed and checksum updated.
func RemovePasswordBytes(data []byte) ([]byte, *RepairResult, error) {
	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse blocks: %w", err)
	}

	result := &RepairResult{}

	// Find header and player block from already-parsed blocks
	var headerBlock *blocks.FileHeader
	var playerBlock *blocks.PlayerBlock
	for _, block := range blockList {
		switch b := block.(type) {
		case blocks.FileHeader:
			headerBlock = &b
		case *blocks.FileHeader:
			headerBlock = b
		case blocks.PlayerBlock:
			playerBlock = &b
		case *blocks.PlayerBlock:
			playerBlock = b
		}
	}

	if headerBlock == nil {
		return nil, nil, fmt.Errorf("no FileHeader found in file")
	}
	if playerBlock == nil {
		result.Message = "no PlayerBlock found"
		return data, result, nil
	}

	// Use already-decrypted data from the parsed PlayerBlock
	decryptedData := make([]byte, len(playerBlock.DecryptedData()))
	copy(decryptedData, playerBlock.DecryptedData())

	// Check if password exists (bytes 12-15)
	hasPassword := len(decryptedData) >= 16 &&
		(decryptedData[12] != 0 || decryptedData[13] != 0 ||
			decryptedData[14] != 0 || decryptedData[15] != 0)

	if !hasPassword {
		result.Success = true
		result.Message = "file has no password"
		return data, result, nil
	}

	// Remove password (bytes 12-15)
	decryptedData[12] = 0
	decryptedData[13] = 0
	decryptedData[14] = 0
	decryptedData[15] = 0

	// Get race names from already-parsed PlayerBlock
	singularName := playerBlock.NameSingular
	pluralName := playerBlock.NamePlural

	// Create output buffer
	repaired := make([]byte, len(data))
	copy(repaired, data)

	// Set up encryptor using header values (same source of truth as parser)
	encryptor := crypto.NewEncryptor()
	encryptor.InitEncryption(
		headerBlock.Salt(),
		0, // shareware flag
		headerBlock.PlayerIndex(),
		int(headerBlock.Turn),
		int(headerBlock.GameID),
	)

	// Find PlayerBlock offset in raw data and re-encrypt
	offset := 0
	for offset < len(repaired)-2 {
		header := uint16(repaired[offset]) | uint16(repaired[offset+1])<<8
		blockType := blocks.BlockTypeID(header >> 10)
		blockSize := int(header & 0x3FF)

		if blockType == blocks.PlayerBlockType {
			// Re-encrypt the modified data
			reEncrypted := encryptor.EncryptBytes(decryptedData)
			copy(repaired[offset+2:offset+2+blockSize], reEncrypted)
			result.PasswordRemoved = true
			break
		}

		offset += 2 + blockSize
	}

	// Calculate new checksum
	newFooter := store.ComputeRaceFooter(decryptedData, singularName, pluralName)

	// Find and update footer
	offset = 0
	for offset < len(repaired)-2 {
		header := uint16(repaired[offset]) | uint16(repaired[offset+1])<<8
		blockType := blocks.BlockTypeID(header >> 10)
		blockSize := int(header & 0x3FF)

		if blockType == blocks.FileFooterBlockType && blockSize == 2 {
			oldFooter := uint16(repaired[offset+2]) | uint16(repaired[offset+3])<<8
			repaired[offset+2] = byte(newFooter & 0xFF)
			repaired[offset+3] = byte(newFooter >> 8)
			result.OldFooter = oldFooter
			result.NewFooter = newFooter
			result.FooterChanged = oldFooter != newFooter
			break
		}

		offset += 2 + blockSize
	}

	result.Success = true
	result.Message = fmt.Sprintf("password removed, footer updated from 0x%04X to 0x%04X",
		result.OldFooter, result.NewFooter)

	return repaired, result, nil
}

// RemovePassword removes the password from a race file on disk.
func RemovePassword(filename string) (*RepairResult, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	repaired, result, err := RemovePasswordBytes(data)
	if err != nil {
		return nil, err
	}

	if result.Success && result.PasswordRemoved {
		if err := os.WriteFile(filename, repaired, 0644); err != nil {
			return nil, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return result, nil
}

// RepairReader repairs race file data from an io.Reader.
func RepairReader(r io.Reader) ([]byte, *RepairResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read data: %w", err)
	}
	return RepairBytes(data)
}

// ValidateHashBytes verifies the integrity checksum of race file data.
// Returns true if the checksum is valid, false if it needs repair.
func ValidateHashBytes(data []byte) (bool, error) {
	info, err := AnalyzeBytes("", data)
	if err != nil {
		return false, err
	}

	return !info.NeedsRepair, nil
}

// ValidateHashReader verifies the integrity checksum from an io.Reader.
func ValidateHashReader(r io.Reader) (bool, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return false, fmt.Errorf("failed to read data: %w", err)
	}
	return ValidateHashBytes(data)
}
