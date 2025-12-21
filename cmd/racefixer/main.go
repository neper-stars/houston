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

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/tools/racefixer"
)

type options struct {
	Args struct {
		File string `positional-arg-name:"file" description:"Race file to fix" required:"true"`
	} `positional-args:"yes"`
}

var description = `Fixes corrupted Stars! race files by recalculating checksums.
A backup of the original file will be created.`

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

	// Analyze the file
	info, err := racefixer.Analyze(opts.Args.File)
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
	result, err := racefixer.RepairWithResult(opts.Args.File)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during repair: %v\n", err)
		os.Exit(1)
	}

	if result.BackupFile != "" {
		fmt.Printf("Created backup: %s\n", result.BackupFile)
	}

	fmt.Printf("Result: %s\n", result.Message)
}
