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

	"github.com/neper-stars/houston/tools/xfilereader"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		printUsage()
		os.Exit(0)
	}

	filename := os.Args[1]

	info, err := xfilereader.ReadFile(filename)
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

func printUsage() {
	fmt.Println("Usage: xfilereader <file>")
	fmt.Println()
	fmt.Println("Reads and validates a Stars! X file (turn order file).")
	fmt.Println("Displays the orders contained in the file.")
}
