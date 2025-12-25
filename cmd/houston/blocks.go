package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/lib/tools/displayblocks"
)

type blocksCommand struct {
	Args struct {
		File string `positional-arg-name:"file" description:"Stars! game file to read" required:"true"`
	} `positional-args:"yes"`
}

func (c *blocksCommand) Execute(args []string) error {
	info, err := displayblocks.ReadFile(c.Args.File)
	if err != nil {
		return err
	}

	if err := displayblocks.WriteBlocks(os.Stdout, info); err != nil {
		return fmt.Errorf("error writing output: %w", err)
	}

	return nil
}

func addBlocksCommand(parser *flags.Parser) {
	_, err := parser.AddCommand("blocks",
		"Display blocks in a Stars! file",
		"Reads a Stars! game file and displays its decrypted blocks.\n\n"+
			"This tool is useful for debugging and understanding Stars! file structure.\n"+
			"It displays each block with its type ID and hex-encoded decrypted data.\n"+
			"For certain block types (FileHeader, Planets, Planet, Fleet, Design),\n"+
			"it also shows the parsed structure.",
		&blocksCommand{})
	if err != nil {
		panic(err)
	}
}
