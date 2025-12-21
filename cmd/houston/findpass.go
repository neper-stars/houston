package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"

	hs "github.com/neper-stars/houston"
)

type findpassCommand struct {
	MaxLength int    `short:"l" long:"length" description:"Maximum password length to try" default:"8"`
	Charset   string `short:"c" long:"charset" description:"Characters to use for brute force" default:"abcdefghijklmnopqrstuvwxyz"`
	Matches   int    `short:"m" long:"matches" description:"Stop after this many matches" default:"1"`
	Verbose   bool   `short:"v" long:"verbose" description:"Show progress"`
	Args      struct {
		File string `positional-arg-name:"file" description:"Stars! file containing player data" required:"true"`
	} `positional-args:"yes"`
}

func (c *findpassCommand) Execute(args []string) error {
	var fd hs.FileData
	if err := hs.ReadRawFile(c.Args.File, &fd); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	bl, err := fd.BlockList()
	if err != nil {
		return fmt.Errorf("failed to iterate over blocks: %w", err)
	}

	for _, b := range bl {
		if b.BlockTypeID() != hs.PlayerBlockType {
			continue
		}

		pb, ok := b.(hs.PlayerBlock)
		if !ok {
			fmt.Fprintln(os.Stderr, "Warning: failed to assert player block")
			continue
		}

		if !pb.Valid {
			continue
		}

		fmt.Printf("Player Block found: %s\n", pb.NameSingular)
		fmt.Printf("Hashed password: %d\n", pb.HashedPass().Uint32())

		matches := hs.GuessRacePassword(
			pb.HashedPass().Uint32(),
			c.MaxLength,
			c.Matches,
			c.Charset,
			c.Verbose,
		)

		if len(matches) > 0 {
			fmt.Println("Found passwords:")
			for _, m := range matches {
				fmt.Printf("  %s\n", m)
			}
		} else {
			fmt.Println("No passwords found")
		}
	}

	return nil
}

func addFindPassCommand(parser *flags.Parser) {
	_, err := parser.AddCommand("findpass",
		"Find race passwords by brute force",
		"Attempts to find race passwords by brute force hashing.\n\n"+
			"Because the hashing algorithm is weak, this will often find\n"+
			"alternative strings that work instead of the original password.",
		&findpassCommand{})
	if err != nil {
		panic(err)
	}
}
