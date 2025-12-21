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
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/tools/mfilemerger"
)

type options struct {
	NoBackup bool `short:"n" long:"no-backup" description:"Don't create backup files"`
	Args     struct {
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

	// Validate file extensions
	for _, filename := range opts.Args.Files {
		ext := strings.ToLower(filepath.Ext(filename))
		if len(ext) < 2 || ext[1] != 'm' {
			fmt.Fprintf(os.Stderr, "Error: %s does not appear to be an M file\n", filename)
			os.Exit(1)
		}
	}

	merger := mfilemerger.New()

	// Read all files into memory
	for _, filename := range opts.Args.Files {
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filename, err)
			os.Exit(1)
		}

		if err := merger.Add(filename, data); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding %s: %v\n", filename, err)
			os.Exit(1)
		}
	}

	// Perform merge
	result, err := merger.Merge()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error merging files: %v\n", err)
		os.Exit(1)
	}

	// Write back merged files
	var backupFiles []string
	for _, filename := range opts.Args.Files {
		// Create backup if requested
		if !opts.NoBackup {
			backupName := backupFilename(filename)
			if err := copyFile(filename, backupName); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating backup for %s: %v\n", filename, err)
				os.Exit(1)
			}
			backupFiles = append(backupFiles, backupName)
		}

		// Get merged data and write it
		mergedData, err := merger.GetMergedData(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting merged data for %s: %v\n", filename, err)
			os.Exit(1)
		}

		if err := os.WriteFile(filename, mergedData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", filename, err)
			os.Exit(1)
		}
	}

	// Print results
	fmt.Printf("Successfully merged %d files\n", result.EntriesProcessed)
	fmt.Printf("  Planets: %d\n", result.PlanetsMerged)
	fmt.Printf("  Fleets: %d\n", result.FleetsMerged)
	fmt.Printf("  Designs: %d\n", result.DesignsMerged)
	fmt.Printf("  Objects: %d\n", result.ObjectsMerged)

	if len(backupFiles) > 0 {
		fmt.Println("\nBackups created:")
		for _, backup := range backupFiles {
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

func backupFilename(filename string) string {
	ext := filepath.Ext(filename)
	if len(ext) >= 2 && (ext[1] == 'm' || ext[1] == 'M') {
		return strings.TrimSuffix(filename, ext) + ".backup-" + ext[1:]
	}
	return filename + ".backup-m"
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
