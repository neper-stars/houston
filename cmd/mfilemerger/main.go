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

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/tools/mfilemerger"
)

type options struct {
	Args struct {
		Files []string `positional-arg-name:"file" description:"M files to merge" required:"true"`
	} `positional-args:"yes"`
}

var description = `All M files supplied on the command line will have their data augmented
with the data on each planet, player, design, fleet, minefield, packet,
salvage, or wormhole from any of the files.

Backups of each input M file will be retained with suffix .backup-m#.`

func main() {
	var opts options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "mfilemerger"
	parser.LongDescription = description

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	merger := mfilemerger.New()

	// Add all files
	for _, filename := range opts.Args.Files {
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
