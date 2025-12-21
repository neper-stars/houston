// Command xfilereader validates and displays X (turn order) file contents.
//
// Usage:
//
//	xfilereader <file.x1>
//
// This tool reads a Stars! X file (player turn orders) and displays its contents.
// It can be used to validate X files before submitting them to the host.
package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/tools/xfilereader"
)

type options struct {
	Args struct {
		File string `positional-arg-name:"file" description:"X file to read" required:"true"`
	} `positional-args:"yes"`
}

var description = `Reads and validates a Stars! X file (turn order file).
Displays the orders contained in the file.`

func main() {
	var opts options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "xfilereader"
	parser.LongDescription = description

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	info, err := xfilereader.ReadFile(opts.Args.File)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print file info
	fmt.Printf("File: %s\n", info.Filename)
	fmt.Printf("Game ID: %d\n", info.GameID)
	fmt.Printf("Turn: %d (Year %d)\n", info.Turn, info.Year)
	fmt.Printf("Player: %d\n", info.PlayerIndex)
	fmt.Println()

	// Print orders
	if len(info.Orders) > 0 {
		fmt.Println("Orders:")
		for _, order := range info.Orders {
			fmt.Printf("  %s\n", order.Description)
		}
		fmt.Println()
	}

	// Print block summary
	fmt.Println("Block Summary:")
	for blockType, count := range info.BlockCounts {
		fmt.Printf("  %s: %d\n", blockType, count)
	}

	fmt.Printf("\nTotal blocks: %d\n", info.BlockCount)

	if info.IsSubmitted {
		fmt.Println("Status: Turn submitted")
	} else {
		fmt.Println("Status: Turn not submitted")
	}

	fmt.Println("\nX file is valid.")
}
