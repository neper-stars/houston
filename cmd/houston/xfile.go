package main

import (
	"fmt"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/tools/xfilereader"
)

type xfileCommand struct {
	Args struct {
		File string `positional-arg-name:"file" description:"X file to read" required:"true"`
	} `positional-args:"yes"`
}

func (c *xfileCommand) Execute(args []string) error {
	info, err := xfilereader.ReadFile(c.Args.File)
	if err != nil {
		return err
	}

	fmt.Printf("File: %s\n", info.Filename)
	fmt.Printf("Game ID: %d\n", info.GameID)
	fmt.Printf("Turn: %d (Year %d)\n", info.Turn, info.Year)
	fmt.Printf("Player: %d\n", info.PlayerIndex)
	fmt.Println()

	if len(info.Orders) > 0 {
		fmt.Println("Orders:")
		for _, order := range info.Orders {
			fmt.Printf("  %s\n", order.Description)
		}
		fmt.Println()
	}

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
	return nil
}

func addXFileCommand(parser *flags.Parser) {
	_, err := parser.AddCommand("xfile",
		"Read and validate X (turn order) files",
		"Reads a Stars! X file (player turn orders) and displays its contents.\n"+
			"Can be used to validate X files before submitting them to the host.",
		&xfileCommand{})
	if err != nil {
		panic(err)
	}
}
