// Command playerchanger modifies player attributes in Stars! game files.
//
// Usage:
//
//	playerchanger <file> [options]
//
// Options:
//
//	-p, --player <n>  Player number to modify (0-15)
//	-a, --ai          Change player to AI
//	-u, --human       Change player to human
//	-i, --info        Display player information only (no changes)
package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/tools/playerchanger"
)

type options struct {
	Player int  `short:"p" long:"player" description:"Player number to modify (0-15)" default:"-1"`
	AI     bool `short:"a" long:"ai" description:"Change player to AI"`
	Human  bool `short:"u" long:"human" description:"Change player to human"`
	Info   bool `short:"i" long:"info" description:"Display player information only (no changes)"`
	Args   struct {
		File string `positional-arg-name:"file" description:"Stars! game file" required:"true"`
	} `positional-args:"yes"`
}

var description = `Modifies player attributes in Stars! game files.
A backup of the original file will be created when making changes.`

func main() {
	var opts options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "playerchanger"
	parser.LongDescription = description

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	// Read player information
	info, err := playerchanger.ReadPlayers(opts.Args.File)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("File: %s (%d bytes, %d blocks)\n", info.Filename, info.Size, info.BlockCount)
	fmt.Printf("Game ID: %d, Turn: %d (Year %d)\n\n", info.GameID, info.Turn, info.Year)

	// Display players
	if len(info.Players) == 0 {
		fmt.Println("No player blocks found")
		os.Exit(1)
	}

	fmt.Println("Players found:")
	for _, p := range info.Players {
		fmt.Printf("  Player %d: %s (%s)\n", p.Number, p.Name, p.PluralName)
		fmt.Printf("    Ships: %d designs, Starbases: %d designs\n",
			p.ShipDesignCount, p.StarbaseDesignCount)
		fmt.Printf("    Planets: %d, Fleets: %d\n", p.Planets, p.Fleets)
	}

	if opts.Info {
		return
	}

	// Validate options
	if opts.AI && opts.Human {
		fmt.Fprintln(os.Stderr, "Error: cannot specify both --ai and --human")
		os.Exit(1)
	}

	if !opts.AI && !opts.Human {
		fmt.Println("\nNo changes requested. Use --ai or --human to modify.")
		return
	}

	if opts.Player < 0 || opts.Player > 15 {
		fmt.Fprintf(os.Stderr, "Error: invalid player number: %d (must be 0-15)\n", opts.Player)
		os.Exit(1)
	}

	// Perform change
	var result *playerchanger.ChangeResult
	if opts.AI {
		result, err = playerchanger.ChangeToAI(opts.Args.File, opts.Player)
	} else {
		result, err = playerchanger.ChangeToHuman(opts.Args.File, opts.Player)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if result.BackupFile != "" {
		fmt.Printf("\nCreated backup: %s\n", result.BackupFile)
	}

	action := "AI"
	if opts.Human {
		action = "human"
	}
	fmt.Printf("Would change player %d to %s.\n", opts.Player, action)
	fmt.Printf("Note: %s\n", result.Message)
}
