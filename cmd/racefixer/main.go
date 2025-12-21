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
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/tools/racefixer"
)

type options struct {
	NoBackup bool `short:"n" long:"no-backup" description:"Don't create backup file"`
	Args     struct {
		File string `positional-arg-name:"file" description:"Race file to fix" required:"true"`
	} `positional-args:"yes"`
}

var description = `Fixes corrupted Stars! race files by recalculating checksums.
A backup of the original file will be created unless --no-backup is specified.`

func main() {
	var opts options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "racefixer"
	parser.LongDescription = description

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	filename := opts.Args.File

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if len(ext) < 2 || ext[1] != 'r' {
		fmt.Fprintf(os.Stderr, "Error: %s does not appear to be a race file\n", filename)
		os.Exit(1)
	}

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Analyze the file
	info, err := racefixer.AnalyzeBytes(filename, data)
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

	// Create backup before repair
	var backupFile string
	if !opts.NoBackup {
		backupFile = filename + ".backup"
		if err := copyFile(filename, backupFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating backup: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created backup: %s\n", backupFile)
	}

	// Attempt repair
	repaired, result, err := racefixer.RepairBytes(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during repair: %v\n", err)
		os.Exit(1)
	}

	if result != nil {
		fmt.Printf("Result: %s\n", result.Message)
	}

	// Write repaired data if we have it
	if repaired != nil && result != nil && result.Success {
		if err := os.WriteFile(filename, repaired, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing repaired file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("File repaired successfully")
	}
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
