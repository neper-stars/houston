// Command playerchanger modifies player attributes in Stars! game files.
//
// Usage:
//
//	playerchanger <file> [options]
//
// Options:
//
//	-player <n>     Player number to modify (0-15)
//	-ai             Change player to AI
//	-human          Change player to human
//	-info           Display player information only (no changes)
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/neper-stars/houston/tools/playerchanger"
)

func main() {
	playerNum := flag.Int("player", -1, "Player number to modify (0-15)")
	toAI := flag.Bool("ai", false, "Change player to AI")
	toHuman := flag.Bool("human", false, "Change player to human")
	infoOnly := flag.Bool("info", false, "Display player information only")
	flag.Parse()

	if flag.NArg() != 1 {
		printUsage()
		os.Exit(0)
	}

	filename := flag.Arg(0)

	// Read player information
	info, err := playerchanger.ReadPlayers(filename)
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

	if *infoOnly {
		return
	}

	// Validate options
	if *toAI && *toHuman {
		fmt.Fprintln(os.Stderr, "Error: cannot specify both -ai and -human")
		os.Exit(1)
	}

	if !*toAI && !*toHuman {
		fmt.Println("\nNo changes requested. Use -ai or -human to modify.")
		return
	}

	if *playerNum < 0 || *playerNum > 15 {
		fmt.Fprintf(os.Stderr, "Error: invalid player number: %d (must be 0-15)\n", *playerNum)
		os.Exit(1)
	}

	// Perform change
	var result *playerchanger.ChangeResult
	if *toAI {
		result, err = playerchanger.ChangeToAI(filename, *playerNum)
	} else {
		result, err = playerchanger.ChangeToHuman(filename, *playerNum)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if result.BackupFile != "" {
		fmt.Printf("\nCreated backup: %s\n", result.BackupFile)
	}

	action := "AI"
	if *toHuman {
		action = "human"
	}
	fmt.Printf("Would change player %d to %s.\n", *playerNum, action)
	fmt.Printf("Note: %s\n", result.Message)
}

func printUsage() {
	fmt.Println("Usage: playerchanger <file> [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -player <n>   Player number to modify (0-15)")
	fmt.Println("  -ai           Change player to AI")
	fmt.Println("  -human        Change player to human")
	fmt.Println("  -info         Display player information only (no changes)")
	fmt.Println()
	fmt.Println("Modifies player attributes in Stars! game files.")
	fmt.Println("A backup of the original file will be created when making changes.")
}
