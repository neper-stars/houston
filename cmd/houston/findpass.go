package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/jessevdk/go-flags"

	hs "github.com/neper-stars/houston"
)

type findpassCommand struct {
	MaxLength int    `short:"l" long:"length" description:"Maximum password length to try" default:"8"`
	Charset   string `short:"c" long:"charset" description:"Characters to use for brute force" default:"abcdefghijklmnopqrstuvwxyz"`
	Matches   int    `short:"m" long:"matches" description:"Stop after this many matches" default:"1"`
	Workers   int    `short:"w" long:"workers" description:"Number of parallel workers (0 = all CPUs)" default:"0"`
	Progress  bool   `short:"p" long:"progress" description:"Show progress while searching"`
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

	workers := c.Workers
	if workers <= 0 {
		workers = runtime.NumCPU()
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
		fmt.Printf("Using %d workers\n", workers)

		// Progress callback
		var progressCb hs.ProgressCallback
		var lastUpdate time.Time
		if c.Progress {
			lastUpdate = time.Now()
			progressCb = func(tried uint64) {
				now := time.Now()
				if now.Sub(lastUpdate) >= 500*time.Millisecond {
					fmt.Printf("\rTried: %d passwords...", tried)
					lastUpdate = now
				}
			}
		}

		start := time.Now()
		matches := hs.GuessRacePasswordParallel(
			pb.HashedPass().Uint32(),
			c.MaxLength,
			c.Matches,
			c.Charset,
			workers,
			progressCb,
		)
		elapsed := time.Since(start)

		if c.Progress {
			fmt.Println() // newline after progress
		}

		if len(matches) > 0 {
			fmt.Println("Found passwords:")
			for _, m := range matches {
				fmt.Printf("  %s\n", m)
			}
		} else {
			fmt.Println("No passwords found")
		}
		fmt.Printf("Time: %v\n", elapsed)
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
