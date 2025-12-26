// Command houston is a unified CLI for Stars! file operations.
//
// Usage:
//
//	houston <command> [options]
//
// Commands:
//
//	blocks     Display blocks in a Stars! file
//	xfile      Read and validate X (turn order) files
//	findpass   Find race passwords by brute force
//	race       Fix corrupted race files
//	race-password  Remove password from race files
//	player     View and modify player attributes
//	merge-m    Merge M files between allied players
//	merge-h    Merge H (history) files
//	map        Render galaxy maps as PNG or animated GIF
//	exploits   Detect and fix known exploits
package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

var version = "dev"

type globalOptions struct {
	Version func() `short:"V" long:"version" description:"Print version and exit"`
}

func main() {
	var globals globalOptions
	globals.Version = func() {
		fmt.Printf("houston %s\n", version)
		os.Exit(0)
	}

	parser := flags.NewParser(&globals, flags.Default)
	parser.Name = "houston"
	parser.LongDescription = "A toolkit for working with Stars! game files"

	// Add subcommands
	addBlocksCommand(parser)
	addXFileCommand(parser)
	addFindPassCommand(parser)
	addRaceCommand(parser)
	addRacePasswordCommand(parser)
	addPlayerCommand(parser)
	addMergeMCommand(parser)
	addMergeHCommand(parser)
	addMapCommand(parser)
	addExploitsCommand(parser)

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok {
			if flagsErr.Type == flags.ErrHelp {
				os.Exit(0)
			}
			if flagsErr.Type == flags.ErrCommandRequired {
				parser.WriteHelp(os.Stderr)
				os.Exit(1)
			}
		}
		os.Exit(1)
	}
}
