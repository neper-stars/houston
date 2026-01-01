package main

import (
	"fmt"
	"io"
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/lib/tools/playerchanger"
)

type playerCommand struct {
	Player   int  `short:"p" long:"player" description:"Player number to modify (0-15)" default:"-1"`
	AI       bool `short:"a" long:"ai" description:"Change player to AI"`
	Human    bool `short:"u" long:"human" description:"Change player to human"`
	Info     bool `short:"i" long:"info" description:"Display player information only (no changes)"`
	NoBackup bool `short:"n" long:"no-backup" description:"Don't create backup file"`
	Args     struct {
		File string `positional-arg-name:"file" description:"Stars! game file" required:"true"`
	} `positional-args:"yes"`
}

func (c *playerCommand) Execute(args []string) error {
	filename := c.Args.File

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Read player information
	info, err := playerchanger.ReadPlayersFromBytes(filename, data)
	if err != nil {
		return err
	}

	fmt.Printf("File: %s (%d bytes, %d blocks)\n", info.Filename, info.Size, info.BlockCount)
	fmt.Printf("Game ID: %d, Turn: %d (Year %d)\n\n", info.GameID, info.Turn, info.Year)

	// Display players
	if len(info.Players) == 0 {
		return fmt.Errorf("no player blocks found")
	}

	fmt.Println("Players found:")
	for _, p := range info.Players {
		fmt.Printf("  Player %d: %s (%s)\n", p.Number, p.Name, p.PluralName)
		fmt.Printf("    Ships: %d designs, Starbases: %d designs\n",
			p.ShipDesignCount, p.StarbaseDesignCount)
		fmt.Printf("    Planets: %d, Fleets: %d\n", p.Planets, p.Fleets)
	}

	if c.Info {
		return nil
	}

	// Validate options
	if c.AI && c.Human {
		return fmt.Errorf("cannot specify both --ai and --human")
	}

	if !c.AI && !c.Human {
		fmt.Println("\nNo changes requested. Use --ai or --human to modify.")
		return nil
	}

	if c.Player < 0 || c.Player > 15 {
		return fmt.Errorf("invalid player number: %d (must be 0-15)", c.Player)
	}

	// Create backup before making changes
	if !c.NoBackup {
		backupFile := filename + ".backup"
		if err := copyFilePlayer(filename, backupFile); err != nil {
			return fmt.Errorf("error creating backup: %w", err)
		}
		fmt.Printf("\nCreated backup: %s\n", backupFile)
	}

	// Perform change
	var modified []byte
	var result *playerchanger.ChangeResult
	if c.AI {
		modified, result, err = playerchanger.ChangeToAIBytes(data, c.Player)
	} else {
		modified, result, err = playerchanger.ChangeToHumanBytes(data, c.Player)
	}

	if err != nil {
		return err
	}

	action := "AI"
	if c.Human {
		action = "human"
	}
	fmt.Printf("Changed player %d to %s.\n", c.Player, action)
	fmt.Printf("Note: %s\n", result.Message)

	// Write modified data if successful
	if modified != nil && result.Success {
		if err := os.WriteFile(filename, modified, 0644); err != nil {
			return fmt.Errorf("error writing file: %w", err)
		}
		fmt.Println("File updated successfully")
	}

	return nil
}

func copyFilePlayer(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = source.Close() }()

	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := dest.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close %s: %v\n", dst, cerr)
		}
	}()

	_, err = io.Copy(dest, source)
	return err
}

func addPlayerCommand(parser *flags.Parser) {
	_, err := parser.AddCommand("player",
		"View and modify player attributes",
		"Modifies player attributes in Stars! game files.\n\n"+
			"Use --info to view player information without making changes.\n"+
			"Use --ai or --human with --player to change a player's type.\n\n"+
			"A backup of the original file will be created when making changes\n"+
			"unless --no-backup is specified.",
		&playerCommand{})
	if err != nil {
		panic(err)
	}
}
