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
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/tools/hfilemerger"
)

type options struct {
	NoBackup bool `short:"n" long:"no-backup" description:"Don't create backup files"`
	Args     struct {
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

	// Classify files by type
	var hFiles, mFiles []string
	for _, filename := range opts.Args.Files {
		ext := strings.ToLower(filepath.Ext(filename))
		if len(ext) >= 2 && ext[1] == 'h' {
			hFiles = append(hFiles, filename)
		} else if len(ext) >= 2 && ext[1] == 'm' {
			mFiles = append(mFiles, filename)
		} else {
			fmt.Fprintf(os.Stderr, "Unknown file type: %s\n", filename)
			os.Exit(1)
		}
	}

	merger := hfilemerger.New()

	// Read and add H files
	for _, filename := range hFiles {
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filename, err)
			os.Exit(1)
		}

		if err := merger.AddH(filename, data); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding H file %s: %v\n", filename, err)
			os.Exit(1)
		}
	}

	// Read and add M files
	for _, filename := range mFiles {
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filename, err)
			os.Exit(1)
		}

		if err := merger.AddM(filename, data); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding M file %s: %v\n", filename, err)
			os.Exit(1)
		}
	}

	// Perform merge
	result, err := merger.Merge()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error merging files: %v\n", err)
		os.Exit(1)
	}

	// Write back H files
	var backupFiles []string
	for _, filename := range hFiles {
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
	fmt.Printf("Successfully merged %d H files (with %d M files for design data)\n",
		result.HEntriesProcessed, result.MEntriesProcessed)
	fmt.Printf("  Planets: %d\n", result.PlanetsMerged)
	fmt.Printf("  Designs: %d\n", result.DesignsMerged)

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
	if len(ext) >= 2 && (ext[1] == 'h' || ext[1] == 'H') {
		return strings.TrimSuffix(filename, ext) + ".backup-" + ext[1:]
	}
	return filename + ".backup-h"
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
