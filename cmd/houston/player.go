package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/lib/tools/playerchanger"
	"github.com/neper-stars/houston/store"
)

type playerCommand struct {
	Player   int    `short:"p" long:"player" description:"Player number to modify (0-15)" default:"-1"`
	AI       string `short:"a" long:"ai" description:"Change player to AI with specified expert type (HE, SS, IS, CA, PP, AR)"`
	Human    bool   `short:"u" long:"human" description:"Change player to human"`
	Inactive bool   `short:"x" long:"inactive" description:"Change player to Human (Inactive)"`
	Info     bool   `short:"i" long:"info" description:"Display player information only (no changes)"`
	NoBackup bool   `short:"n" long:"no-backup" description:"Don't create backup file"`
	Args     struct {
		File string `positional-arg-name:"file" description:"Stars! game file (.hst)" required:"true"`
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
		fmt.Printf("  Player %d: %s (%s) - %s\n", p.Number, p.Name, p.PluralName, p.Status)
		fmt.Printf("    Ships: %d designs, Starbases: %d designs\n",
			p.ShipDesignCount, p.StarbaseDesignCount)
		fmt.Printf("    Planets: %d, Fleets: %d\n", p.OwnedPlanets, p.Fleets)
	}

	if c.Info {
		return nil
	}

	// Count how many change options are specified
	changeCount := 0
	if c.AI != "" {
		changeCount++
	}
	if c.Human {
		changeCount++
	}
	if c.Inactive {
		changeCount++
	}

	// Validate options
	if changeCount > 1 {
		return fmt.Errorf("cannot specify multiple change options (--ai, --human, --inactive)")
	}

	if changeCount == 0 {
		fmt.Println("\nNo changes requested. Use --ai, --human, or --inactive to modify.")
		fmt.Println("\nAvailable AI expert types:")
		for _, aiType := range store.AllAIExpertTypes() {
			fmt.Printf("  %-2s  %-18s  %s\n", aiType.ShortName(), aiType.FullName(), aiType.Description())
		}
		return nil
	}

	if c.Player < 0 || c.Player > 15 {
		return fmt.Errorf("invalid player number: %d (must be 0-15)", c.Player)
	}

	// Parse AI type if specified
	var aiType store.AIExpertType
	if c.AI != "" {
		var parseErr error
		aiType, parseErr = store.ParseAIExpertType(c.AI)
		if parseErr != nil {
			return parseErr
		}
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
	switch {
	case c.AI != "":
		modified, result, err = playerchanger.ChangeToAIBytes(data, c.Player, aiType)
	case c.Human:
		modified, result, err = playerchanger.ChangeToHumanBytes(data, c.Player)
	case c.Inactive:
		modified, result, err = playerchanger.ChangeToInactiveBytes(data, c.Player)
	}

	if err != nil {
		return err
	}

	fmt.Printf("\n%s\n", result.Message)

	// Write modified data if successful
	if modified != nil && result.Success {
		if err := os.WriteFile(filename, modified, 0644); err != nil {
			return fmt.Errorf("error writing file: %w", err)
		}
		fmt.Println("File updated successfully.")

		// Show note about AI password if changing to AI
		if c.AI != "" {
			fmt.Println("\nNote: The password to view AI turn files is \"viewai\"")
		}
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
	// Build AI types help text with full descriptions
	var aiHelp strings.Builder
	aiHelp.WriteString("Available AI expert types:\n")
	for _, t := range store.AllAIExpertTypes() {
		aiHelp.WriteString(fmt.Sprintf("  %-2s  %-18s  %s\n", t.ShortName(), t.FullName(), t.Description()))
	}

	_, err := parser.AddCommand("player",
		"View and modify player attributes",
		"Modifies player attributes in Stars! HST game files.\n\n"+
			"Use --info to view player information without making changes.\n"+
			"Use --ai with a type to change a player to AI control.\n"+
			"Use --human to change a player to human control.\n"+
			"Use --inactive to change a player to Human (Inactive).\n\n"+
			aiHelp.String()+"\n"+
			"Example:\n"+
			"  houston player --player 1 --ai CA game.hst\n"+
			"  houston player --player 2 --human game.hst\n\n"+
			"A backup of the original file will be created when making changes\n"+
			"unless --no-backup is specified.\n\n"+
			"Note: The password to view AI turn files is \"viewai\"",
		&playerCommand{})
	if err != nil {
		panic(err)
	}
}
