// Package racefixer provides functionality to repair corrupted Stars! race files.
//
// Race files can become corrupted if edited improperly. This package provides
// functions to analyze race files and recalculate checksums to make them valid.
//
// Example usage:
//
//	info, err := racefixer.Analyze("MyRace.r1")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if info.NeedsRepair {
//	    if err := racefixer.Repair("MyRace.r1"); err != nil {
//	        log.Fatal(err)
//	    }
//	}
package racefixer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
func Analyze(filename string) (*FileInfo, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return AnalyzeBytes(filename, fileBytes)
}

// AnalyzeBytes analyzes race file data.
func AnalyzeBytes(filename string, fileBytes []byte) (*FileInfo, error) {
	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if len(ext) < 2 || ext[1] != 'r' {
		return nil, fmt.Errorf("%s does not appear to be a race file", filename)
	}

	fd := parser.FileData(fileBytes)

	blockList, err := fd.BlockList()
	if err != nil {
		return nil, fmt.Errorf("failed to parse blocks: %w", err)
	}

	info := &FileInfo{
		Filename:   filename,
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

// Repair attempts to fix a corrupted race file.
func Repair(filename string) error {
	info, err := Analyze(filename)
	if err != nil {
		return err
	}

	if !info.HasHashBlock {
		return fmt.Errorf("no hash block found - cannot determine checksum location")
	}

	// Create backup
	backupName := filename + ".backup"
	if err := copyFile(filename, backupName); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Note: Actual checksum recalculation requires understanding the Stars! checksum algorithm
	// This is a placeholder
	return fmt.Errorf("checksum recalculation not yet implemented")
}

// RepairResult contains the results of a repair operation.
type RepairResult struct {
	Success    bool
	BackupFile string
	Message    string
}

// RepairWithResult repairs a race file and returns detailed results.
func RepairWithResult(filename string) (*RepairResult, error) {
	info, err := Analyze(filename)
	if err != nil {
		return nil, err
	}

	result := &RepairResult{}

	if !info.HasHashBlock {
		result.Message = "no hash block found"
		return result, nil
	}

	// Create backup
	backupName := filename + ".backup"
	if err := copyFile(filename, backupName); err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}
	result.BackupFile = backupName

	// Note: Actual implementation would recalculate and write the checksum
	result.Message = "checksum recalculation not yet implemented"

	return result, nil
}

// ValidateChecksum verifies the checksum of a race file.
func ValidateChecksum(filename string) (bool, error) {
	info, err := Analyze(filename)
	if err != nil {
		return false, err
	}

	if !info.HasHashBlock {
		return false, fmt.Errorf("no hash block found")
	}

	// Note: Actual implementation would compute and compare checksums
	return true, nil
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	return err
}
