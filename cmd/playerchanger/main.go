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
	"io"
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/tools/playerchanger"
)

type options struct {
	Player   int  `short:"p" long:"player" description:"Player number to modify (0-15)" default:"-1"`
	AI       bool `short:"a" long:"ai" description:"Change player to AI"`
	Human    bool `short:"u" long:"human" description:"Change player to human"`
	Info     bool `short:"i" long:"info" description:"Display player information only (no changes)"`
	NoBackup bool `short:"n" long:"no-backup" description:"Don't create backup file"`
	Args     struct {
		File string `positional-arg-name:"file" description:"Stars! game file" required:"true"`
	} `positional-args:"yes"`
}

var description = `Modifies player attributes in Stars! game files.
A backup of the original file will be created when making changes unless --no-backup is specified.`

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

	filename := opts.Args.File

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Read player information
	info, err := playerchanger.ReadPlayersFromBytes(filename, data)
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

	// Create backup before making changes
	if !opts.NoBackup {
		backupFile := filename + ".backup"
		if err := copyFile(filename, backupFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating backup: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\nCreated backup: %s\n", backupFile)
	}

	// Perform change
	var modified []byte
	var result *playerchanger.ChangeResult
	if opts.AI {
		modified, result, err = playerchanger.ChangeToAIBytes(data, opts.Player)
	} else {
		modified, result, err = playerchanger.ChangeToHumanBytes(data, opts.Player)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	action := "AI"
	if opts.Human {
		action = "human"
	}
	fmt.Printf("Would change player %d to %s.\n", opts.Player, action)
	fmt.Printf("Note: %s\n", result.Message)

	// Write modified data if successful
	if modified != nil && result.Success {
		if err := os.WriteFile(filename, modified, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("File updated successfully")
	}
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	return err
}
