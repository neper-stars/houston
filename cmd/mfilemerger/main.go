// Command mfilemerger merges data between allied players' M files.
//
// Usage:
//
//	mfilemerger <file1.m1> <file2.m2> [...]
//
// This tool takes multiple M files from allied players and augments each file
// with planet, fleet, design, and object data from the other files.
package main

import (
	"fmt"
	"os"

	"github.com/neper-stars/houston/tools/mfilemerger"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		printUsage()
		os.Exit(0)
	}

	filenames := os.Args[1:]

	merger := mfilemerger.New()

	// Add all files
	for _, filename := range filenames {
		if err := merger.AddFile(filename); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding file %s: %v\n", filename, err)
			os.Exit(1)
		}
	}

	// Perform merge
	result, err := merger.Merge()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error merging files: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Printf("Successfully merged %d files\n", result.FilesProcessed)
	fmt.Printf("  Planets: %d\n", result.PlanetsMerged)
	fmt.Printf("  Fleets: %d\n", result.FleetsMerged)
	fmt.Printf("  Designs: %d\n", result.DesignsMerged)
	fmt.Printf("  Objects: %d\n", result.ObjectsMerged)

	if len(result.BackupFiles) > 0 {
		fmt.Println("\nBackups created:")
		for _, backup := range result.BackupFiles {
			fmt.Printf("  %s\n", backup)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("  %s\n", warning)
		}
	}
}

func printUsage() {
	fmt.Println("Usage: mfilemerger file...")
	fmt.Println()
	fmt.Println("All M files supplied on the command line will have their data augmented")
	fmt.Println("with the data on each planet, player, design, fleet, minefield, packet,")
	fmt.Println("salvage, or wormhole from any of the files.")
	fmt.Println()
	fmt.Println("Backups of each input M file will be retained with suffix .backup-m#.")
}
