package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/lib/tools/hfilemerger"
)

type mergeHCommand struct {
	NoBackup bool `short:"n" long:"no-backup" description:"Don't create backup files"`
	Args     struct {
		Files []string `positional-arg-name:"file" description:"H and M files to process" required:"true"`
	} `positional-args:"yes"`
}

func (c *mergeHCommand) Execute(args []string) error {
	// Classify files by type
	var hFiles, mFiles []string
	for _, filename := range c.Args.Files {
		ext := strings.ToLower(filepath.Ext(filename))
		if len(ext) >= 2 && ext[1] == 'h' {
			hFiles = append(hFiles, filename)
		} else if len(ext) >= 2 && ext[1] == 'm' {
			mFiles = append(mFiles, filename)
		} else {
			return fmt.Errorf("unknown file type: %s", filename)
		}
	}

	merger := hfilemerger.New()

	// Read and add H files
	for _, filename := range hFiles {
		data, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", filename, err)
		}

		if err := merger.AddH(filename, data); err != nil {
			return fmt.Errorf("error adding H file %s: %w", filename, err)
		}
	}

	// Read and add M files
	for _, filename := range mFiles {
		data, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", filename, err)
		}

		if err := merger.AddM(filename, data); err != nil {
			return fmt.Errorf("error adding M file %s: %w", filename, err)
		}
	}

	// Perform merge
	result, err := merger.Merge()
	if err != nil {
		return fmt.Errorf("error merging files: %w", err)
	}

	// Write back H files
	var backupFiles []string
	for _, filename := range hFiles {
		// Create backup if requested
		if !c.NoBackup {
			backupName := backupFilenameMergeH(filename)
			if err := copyFileMergeH(filename, backupName); err != nil {
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

	return nil
}

func backupFilenameMergeH(filename string) string {
	ext := filepath.Ext(filename)
	if len(ext) >= 2 && (ext[1] == 'h' || ext[1] == 'H') {
		return strings.TrimSuffix(filename, ext) + ".backup-" + ext[1:]
	}
	return filename + ".backup-h"
}

func copyFileMergeH(src, dst string) error {
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

func addMergeHCommand(parser *flags.Parser) {
	_, err := parser.AddCommand("merge-h",
		"Merge H (history) files",
		"All H files supplied on the command line will have their data replaced\n"+
			"with the newest data on each planet, player, and design from any of the files.\n\n"+
			"M files supplied on the command line will have their data incorporated\n"+
			"but will not be changed. M files are needed for accurately determining\n"+
			"the latest ship designs.\n\n"+
			"Backups of each input H file will be retained with suffix .backup-h#.",
		&mergeHCommand{})
	if err != nil {
		panic(err)
	}
}
