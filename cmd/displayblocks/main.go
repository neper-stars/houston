// Command displayblocks reads a Stars! game file and displays its decrypted blocks.
//
// Usage:
//
//	displayblocks <file>
//
// This tool is useful for debugging and understanding Stars! file structure.
// It displays each block with its type ID and hex-encoded decrypted data.
// For certain block types (FileHeader, Planets, Planet, Fleet, Design),
// it also shows the parsed structure.
package main

import (
	"fmt"
	"os"

	"github.com/neper-stars/houston/tools/displayblocks"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Println("Usage: displayblocks <file>")
		fmt.Println()
		fmt.Println("Displays the decrypted blocks of a Stars! game file.")
		os.Exit(1)
	}

	filename := os.Args[1]

	info, err := displayblocks.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := displayblocks.WriteBlocks(os.Stdout, info); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}
}
