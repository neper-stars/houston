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

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/tools/displayblocks"
)

type options struct {
	Args struct {
		File string `positional-arg-name:"file" description:"Stars! game file to read" required:"true"`
	} `positional-args:"yes"`
}

func main() {
	var opts options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "displayblocks"

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	info, err := displayblocks.ReadFile(opts.Args.File)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := displayblocks.WriteBlocks(os.Stdout, info); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}
}
