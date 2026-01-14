package main

import (
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/lib/tools/reporter"
)

//go:embed resources/empty.ods
var embeddedTemplate []byte

type reportCommand struct {
	Output    string `short:"o" long:"output" description:"Output ODS filename" default:"report.ods"`
	Template  string `short:"t" long:"template" description:"Template ODS file (uses embedded template by default)"`
	Player    int    `short:"p" long:"player" description:"Player number (1-16, auto-detected from M-file if not specified)"`
	Threshold int64  `long:"threshold" description:"Mineral threshold for shuffle analysis" default:"500"`
	Args      struct {
		Files []string `positional-arg-name:"file" description:"Stars! game files (.m, .h, .xy)" required:"true"`
	} `positional-args:"yes"`
}

func (c *reportCommand) Execute(args []string) error {
	startTime := time.Now()
	defer func() {
		fmt.Printf("  Generated in: %v\n", time.Since(startTime))
	}()

	// Check we have input
	if len(c.Args.Files) == 0 {
		return fmt.Errorf("no input files specified")
	}

	// Create reporter
	rep := reporter.New()

	// Load template (use embedded by default, or from file if specified)
	if c.Template != "" {
		if err := rep.SetTemplateFile(c.Template); err != nil {
			return fmt.Errorf("failed to load template: %w", err)
		}
	} else {
		rep.SetTemplateBytes(embeddedTemplate)
	}

	// Try to load existing report (for history preservation)
	if _, err := os.Stat(c.Output); err == nil {
		if err := rep.SetExistingReportFile(c.Output); err != nil {
			fmt.Printf("Warning: could not load existing report for history: %v\n", err)
		} else {
			fmt.Printf("Loaded existing report for history preservation\n")
		}
	}

	// Load all input files
	for _, filename := range c.Args.Files {
		fmt.Printf("Loading %s...\n", filename)
		if err := rep.LoadFileWithXY(filename); err != nil {
			return fmt.Errorf("failed to load %s: %w", filename, err)
		}
	}

	// Determine player number (auto-detect or use override)
	playerNumber := c.Player - 1 // Convert to 0-indexed (-1 if not specified)
	if c.Player == 0 {
		// Auto-detect from M-file
		detected := rep.DetectedPlayerNumber()
		if detected < 0 {
			return fmt.Errorf("could not auto-detect player number: no M-file loaded")
		}
		playerNumber = detected
		fmt.Printf("Auto-detected player %d from M-file\n", playerNumber+1)
	}

	// Generate report
	opts := &reporter.ReportOptions{
		PlayerNumber:     playerNumber,
		MineralThreshold: c.Threshold,
	}

	if err := rep.GenerateReportToFile(c.Output, opts); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	fmt.Printf("Created %s\n", c.Output)
	fmt.Printf("  Game ID: %d\n", rep.GameID())
	fmt.Printf("  Year: %d (Turn %d)\n", rep.Year(), rep.Turn())
	fmt.Printf("  Player: %d\n", playerNumber+1)

	return nil
}

func addReportCommand(parser *flags.Parser) {
	_, err := parser.AddCommand("report",
		"Generate analysis report as ODS spreadsheet",
		"Creates a LibreOffice spreadsheet with multi-turn analysis.\n\n"+
			"The report includes:\n"+
			"  - Summary of all players (planets, population, minerals, ships, score)\n"+
			"  - Your minerals per planet (including cargo en route)\n"+
			"  - Historical tracking of mineral totals across turns\n"+
			"  - Mineral shuffle recommendations (planets needing minerals)\n"+
			"  - Opponent population tracking and growth rates\n"+
			"  - Opponent ship counts by category (unarmed/escort/capital)\n"+
			"  - Enemy ship designs detected\n"+
			"  - Score estimates for all players\n\n"+
			"The player number is automatically detected from the M-file.\n"+
			"If the output file already exists, it will be updated with the new turn's data\n"+
			"while preserving historical information.\n\n"+
			"Example:\n"+
			"  houston report game.m1 -o game-report.ods\n"+
			"  houston report game.m1 game.h1 -o game-report.ods",
		&reportCommand{})
	if err != nil {
		panic(err)
	}
}
