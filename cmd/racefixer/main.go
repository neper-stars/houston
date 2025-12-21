// Command racefixer repairs corrupted race files by fixing their checksums.
//
// Usage:
//
//	racefixer <file.r1>
//
// Stars! race files can become corrupted if edited improperly. This tool
// recalculates and fixes the file checksum/hash to make the file valid again.
package main

import (
	"fmt"
	"os"

	"github.com/neper-stars/houston/tools/racefixer"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		printUsage()
		os.Exit(0)
	}

	filename := os.Args[1]

	// Analyze the file
	info, err := racefixer.Analyze(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("File: %s (%d bytes, %d blocks)\n", info.Filename, info.Size, info.BlockCount)

	if info.HasHashBlock {
		fmt.Println("Hash block found")
	} else {
		fmt.Println("No hash block found - this may not be a race file")
		os.Exit(1)
	}

	// Attempt repair
	result, err := racefixer.RepairWithResult(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during repair: %v\n", err)
		os.Exit(1)
	}

	if result.BackupFile != "" {
		fmt.Printf("Created backup: %s\n", result.BackupFile)
	}

	fmt.Printf("Result: %s\n", result.Message)
}

func printUsage() {
	fmt.Println("Usage: racefixer <file>")
	fmt.Println()
	fmt.Println("Fixes corrupted Stars! race files by recalculating checksums.")
	fmt.Println("A backup of the original file will be created.")
}
