// Command hfilemerger merges data from multiple H (history) files.
//
// Usage:
//
//	hfilemerger <file1.h1> <file2.h2> [file.m1] [...]
//
// All H files supplied will have their data replaced with the newest data
// on each planet, player, and design from any of the files.
// M files will have their data incorporated but will not be changed.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/tools/hfilemerger"
)

type options struct {
	Args struct {
		Files []string `positional-arg-name:"file" description:"H and M files to process" required:"true"`
	} `positional-args:"yes"`
}

var description = `All H files supplied on the command line will have their data replaced
with the newest data on each planet, player, and design from any of the files.

M files supplied on the command line will have their data incorporated
but will not be changed. M files are needed for accurately determining
the latest ship designs.

Backups of each input H file will be retained with suffix .backup-h#.`

func main() {
	var opts options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "hfilemerger"
	parser.LongDescription = description

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	merger := hfilemerger.New()

	// Add files based on extension
	for _, filename := range opts.Args.Files {
		ext := strings.ToLower(filepath.Ext(filename))
		if len(ext) >= 2 && ext[1] == 'h' {
			if err := merger.AddHFile(filename); err != nil {
				fmt.Fprintf(os.Stderr, "Error adding H file %s: %v\n", filename, err)
				os.Exit(1)
			}
		} else if len(ext) >= 2 && ext[1] == 'm' {
			if err := merger.AddMFile(filename); err != nil {
				fmt.Fprintf(os.Stderr, "Error adding M file %s: %v\n", filename, err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Unknown file type: %s\n", filename)
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
	fmt.Printf("Successfully merged %d H files (with %d M files for design data)\n",
		result.HFilesProcessed, result.MFilesProcessed)
	fmt.Printf("  Planets: %d\n", result.PlanetsMerged)
	fmt.Printf("  Designs: %d\n", result.DesignsMerged)

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
