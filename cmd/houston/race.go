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

type raceCommand struct {
	NoBackup bool `short:"n" long:"no-backup" description:"Don't create backup file"`
	Args     struct {
		File string `positional-arg-name:"file" description:"Race file to fix" required:"true"`
	} `positional-args:"yes"`
}

func (c *raceCommand) Execute(args []string) error {
	filename := c.Args.File

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if len(ext) < 2 || ext[1] != 'r' {
		return fmt.Errorf("%s does not appear to be a race file", filename)
	}

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Analyze the file
	info, err := racefixer.AnalyzeBytes(filename, data)
	if err != nil {
		return err
	}

	fmt.Printf("File: %s (%d bytes, %d blocks)\n", info.Filename, info.Size, info.BlockCount)

	if info.HasHashBlock {
		fmt.Println("Hash block found")
	} else {
		return fmt.Errorf("no hash block found - this may not be a race file")
	}

	// Create backup before repair
	if !c.NoBackup {
		backupFile := filename + ".backup"
		if err := copyFileRace(filename, backupFile); err != nil {
			return fmt.Errorf("error creating backup: %w", err)
		}
		fmt.Printf("Created backup: %s\n", backupFile)
	}

	// Attempt repair
	repaired, result, err := racefixer.RepairBytes(data)
	if err != nil {
		return fmt.Errorf("error during repair: %w", err)
	}

	if result != nil {
		fmt.Printf("Result: %s\n", result.Message)
	}

	// Write repaired data if successful
	if repaired != nil && result != nil && result.Success {
		if err := os.WriteFile(filename, repaired, 0644); err != nil {
			return fmt.Errorf("error writing repaired file: %w", err)
		}
		fmt.Println("File repaired successfully")
	}

	return nil
}

func copyFileRace(src, dst string) error {
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

func addRaceCommand(parser *flags.Parser) {
	_, err := parser.AddCommand("race",
		"Fix corrupted race files",
		"Fixes corrupted Stars! race files by recalculating checksums.\n\n"+
			"Stars! race files can become corrupted if edited improperly.\n"+
			"This tool recalculates and fixes the file checksum/hash.\n\n"+
			"A backup of the original file will be created unless --no-backup is specified.",
		&raceCommand{})
	if err != nil {
		panic(err)
	}
}
