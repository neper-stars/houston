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
	"github.com/neper-stars/houston/encoding"
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

// parseRaceNamesFromData extracts the singular and plural race names from decrypted PlayerBlock data.
// This matches the parsing logic in blocks.PlayerBlock.decode().
func parseRaceNamesFromData(decryptedData []byte) (singular, plural string, err error) {
	if len(decryptedData) < 8 {
		return "", "", fmt.Errorf("data too short: need at least 8 bytes, got %d", len(decryptedData))
	}

	// Check the required bits in byte 6 (bits 0-1 must be 0x03)
	if (decryptedData[6] & 0x03) != 0x03 {
		return "", "", fmt.Errorf("invalid PlayerBlock format: byte 6 bits 0-1 should be 0x03")
	}

	// FullDataFlag is bit 2 of byte 6
	fullDataFlag := (decryptedData[6] & 0x04) != 0

	index := 8
	if fullDataFlag {
		// For full data, names are after player relations at offset 0x70
		index = 0x70
		if index >= len(decryptedData) {
			return "", "", fmt.Errorf("data too short for full data: need at least %d bytes, got %d", index+1, len(decryptedData))
		}
		playerRelationsLength := int(decryptedData[index])
		index += 1 + playerRelationsLength
	}

	if index >= len(decryptedData) {
		return "", "", fmt.Errorf("data too short for names: index %d >= length %d", index, len(decryptedData))
	}

	// Decode singular name
	singularNameLength := int(decryptedData[index])
	singularEnd := index + 1 + singularNameLength
	if singularEnd > len(decryptedData) {
		return "", "", fmt.Errorf("singular name extends beyond data: end %d > length %d", singularEnd, len(decryptedData))
	}
	singularBytes := decryptedData[index:singularEnd]
	singular, err = encoding.DecodeStarsString(singularBytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode singular name: %w", err)
	}

	// Decode plural name
	pluralIndex := singularEnd
	if pluralIndex >= len(decryptedData) {
		return singular, "", nil
	}
	pluralNameLength := int(decryptedData[pluralIndex])
	pluralEnd := pluralIndex + 1 + pluralNameLength
	if pluralEnd > len(decryptedData) {
		pluralEnd = len(decryptedData)
	}
	pluralBytes := decryptedData[pluralIndex:pluralEnd]
	plural, err = encoding.DecodeStarsString(pluralBytes)
	if err != nil {
		return singular, "", fmt.Errorf("failed to decode plural name: %w", err)
	}

	return singular, plural, nil
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

	// Find header for salt
	var salt int
	var headerBlock blocks.FileHeader
	for _, block := range blockList {
		switch b := block.(type) {
		case blocks.FileHeader:
			salt = b.Salt()
			headerBlock = b
		case *blocks.FileHeader:
			salt = b.Salt()
			headerBlock = *b
		}
	}

	// Create output buffer
	repaired := make([]byte, len(data))
	copy(repaired, data)

	// Set up encryption/decryption
	decryptor := crypto.NewDecryptor()
	decryptor.InitDecryption(salt, 0, 0, 31, 0)

	encryptor := crypto.NewEncryptor()
	encryptor.InitEncryption(salt, 0, 0, 31, 0)

	// Track the decrypted data for checksum calculation
	var decryptedPlayerData []byte
	var singularName, pluralName string
	var playerBlockOffset, playerBlockSize int

	// Process blocks
	offset := 0
	for offset < len(repaired)-2 {
		header := uint16(repaired[offset]) | uint16(repaired[offset+1])<<8
		blockType := blocks.BlockTypeID(header >> 10)
		blockSize := int(header & 0x3FF)

		if blockType == blocks.PlayerBlockType {
			playerBlockOffset = offset
			playerBlockSize = blockSize

			// Decrypt the block
			encryptedData := repaired[offset+2 : offset+2+blockSize]
			decryptedData := decryptor.DecryptBytes(encryptedData)

			// Check if password exists
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

			// Parse race names for checksum
			singularName, pluralName, err = parseRaceNamesFromData(decryptedData)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse race names: %w", err)
			}

			decryptedPlayerData = decryptedData

			// Re-encrypt the modified data
			reEncrypted := encryptor.EncryptBytes(decryptedData)
			copy(repaired[offset+2:], reEncrypted)

			result.PasswordRemoved = true
		}

		offset += 2 + blockSize
	}

	if !result.PasswordRemoved {
		result.Message = "no PlayerBlock found"
		return data, result, nil
	}

	// Calculate new checksum
	newFooter := store.ComputeRaceFooter(decryptedPlayerData, singularName, pluralName)

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

	_ = headerBlock       // silence unused warning
	_ = playerBlockOffset // silence unused warning
	_ = playerBlockSize   // silence unused warning

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
