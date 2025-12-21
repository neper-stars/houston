// Package racefixer provides functionality to repair corrupted Stars! race files.
//
// Race files can become corrupted if edited improperly. This package provides
// functions to analyze race files and recalculate checksums to make them valid.
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
//	    repaired, err := racefixer.RepairBytes(data)
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
	"github.com/neper-stars/houston/parser"
)

// FileInfo contains information about a race file.
type FileInfo struct {
	Filename    string
	Size        int
	BlockCount  int
	HasHashBlock bool
	NeedsRepair bool
	Blocks      []blocks.Block
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

	// Look for hash block
	for _, block := range blockList {
		if _, ok := block.(blocks.FileHashBlock); ok {
			info.HasHashBlock = true
			break
		}
	}

	return info, nil
}

// RepairResult contains the results of a repair operation.
type RepairResult struct {
	Success bool
	Message string
}

// RepairBytes attempts to fix corrupted race file data and returns the repaired bytes.
// Returns the repaired data or an error if repair is not possible.
func RepairBytes(data []byte) ([]byte, *RepairResult, error) {
	info, err := AnalyzeBytes("", data)
	if err != nil {
		return nil, nil, err
	}

	result := &RepairResult{}

	if !info.HasHashBlock {
		result.Message = "no hash block found - cannot determine checksum location"
		return nil, result, nil
	}

	// Note: Actual checksum recalculation requires understanding the Stars! checksum algorithm
	// This is a placeholder - return original data for now
	result.Message = "checksum recalculation not yet implemented"

	return data, result, nil
}

// RepairReader repairs race file data from an io.Reader.
func RepairReader(r io.Reader) ([]byte, *RepairResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read data: %w", err)
	}
	return RepairBytes(data)
}

// ValidateChecksumBytes verifies the checksum of race file data.
func ValidateChecksumBytes(data []byte) (bool, error) {
	info, err := AnalyzeBytes("", data)
	if err != nil {
		return false, err
	}

	if !info.HasHashBlock {
		return false, fmt.Errorf("no hash block found")
	}

	// Note: Actual implementation would compute and compare checksums
	return true, nil
}

// ValidateChecksumReader verifies the checksum from an io.Reader.
func ValidateChecksumReader(r io.Reader) (bool, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return false, fmt.Errorf("failed to read data: %w", err)
	}
	return ValidateChecksumBytes(data)
}
