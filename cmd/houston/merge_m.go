package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/lib/tools/mfilemerger"
)

type mergeMCommand struct {
	NoBackup bool `short:"n" long:"no-backup" description:"Don't create backup files"`
	Args     struct {
		Files []string `positional-arg-name:"file" description:"M files to merge" required:"true"`
	} `positional-args:"yes"`
}

func (c *mergeMCommand) Execute(args []string) error {
	// Validate file extensions
	for _, filename := range c.Args.Files {
		ext := strings.ToLower(filepath.Ext(filename))
		if len(ext) < 2 || ext[1] != 'm' {
			return fmt.Errorf("%s does not appear to be an M file", filename)
		}
	}

	merger := mfilemerger.New()

	// Read all files into memory
	for _, filename := range c.Args.Files {
		data, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", filename, err)
		}

		if err := merger.Add(filename, data); err != nil {
			return fmt.Errorf("error adding %s: %w", filename, err)
		}
	}

	// Perform merge
	result, err := merger.Merge()
	if err != nil {
		return fmt.Errorf("error merging files: %w", err)
	}

	// Write back merged files
	var backupFiles []string
	for _, filename := range c.Args.Files {
		// Create backup if requested
		if !c.NoBackup {
			backupName := backupFilenameMergeM(filename)
			if err := copyFileMergeM(filename, backupName); err != nil {
				return fmt.Errorf("error creating backup for %s: %w", filename, err)
			}
			backupFiles = append(backupFiles, backupName)
		}

		// Get merged data and write it
		mergedData, err := merger.GetMergedData(filename)
		if err != nil {
			return fmt.Errorf("error getting merged data for %s: %w", filename, err)
		}

		if err := os.WriteFile(filename, mergedData, 0644); err != nil {
			return fmt.Errorf("error writing %s: %w", filename, err)
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

	return nil
}

func backupFilenameMergeM(filename string) string {
	ext := filepath.Ext(filename)
	if len(ext) >= 2 && (ext[1] == 'm' || ext[1] == 'M') {
		return strings.TrimSuffix(filename, ext) + ".backup-" + ext[1:]
	}
	return filename + ".backup-m"
}

func copyFileMergeM(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = source.Close() }()

	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := dest.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close %s: %v\n", dst, cerr)
		}
	}()

	_, err = io.Copy(dest, source)
	return err
}

func addMergeMCommand(parser *flags.Parser) {
	_, err := parser.AddCommand("merge-m",
		"Merge M files between allied players",
		"All M files supplied on the command line will have their data augmented\n"+
			"with the data on each planet, player, design, fleet, minefield, packet,\n"+
			"salvage, or wormhole from any of the files.\n\n"+
			"Backups of each input M file will be retained with suffix .backup-m#.",
		&mergeMCommand{})
	if err != nil {
		panic(err)
	}
}
