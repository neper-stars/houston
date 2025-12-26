package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/lib/tools/maprenderer"
)

type mapCommand struct {
	Output     string `short:"o" long:"output" description:"Output filename (default: input.png or animation.gif)"`
	Width      int    `short:"W" long:"width" description:"Image width in pixels" default:"800"`
	Height     int    `short:"H" long:"height" description:"Image height in pixels" default:"600"`
	SVG        bool   `short:"s" long:"svg" description:"Output as SVG instead of PNG"`
	GIF        bool   `short:"g" long:"gif" description:"Create animated GIF from multiple files"`
	Dir        string `short:"d" long:"dir" description:"Load all M files from directory for animation"`
	Delay      int    `long:"delay" description:"Delay between frames in milliseconds" default:"1000"`
	ShowNames  bool   `short:"n" long:"names" description:"Show planet names"`
	ShowFleets bool   `short:"f" long:"fleets" description:"Show fleet indicators"`
	FleetPaths int    `short:"p" long:"fleet-paths" description:"Show fleet projected paths (number of years)" default:"0"`
	ShowMines  bool   `short:"m" long:"mines" description:"Show minefields"`
	ShowWH     bool   `short:"w" long:"wormholes" description:"Show wormholes"`
	ShowLegend bool   `short:"l" long:"legend" description:"Show player legend"`
	Args       struct {
		Files []string `positional-arg-name:"file" description:"Stars! game files to render"`
	} `positional-args:"yes"`
}

func (c *mapCommand) Execute(args []string) error {
	// Check we have input
	if len(c.Args.Files) == 0 && c.Dir == "" {
		return fmt.Errorf("no input files specified")
	}

	// Set defaults for boolean options not explicitly set
	showFleets := c.ShowFleets
	showWH := c.ShowWH
	showLegend := c.ShowLegend
	// If none of the display options are set, use sensible defaults
	if !c.ShowFleets && !c.ShowMines && !c.ShowWH && !c.ShowLegend && !c.ShowNames {
		showFleets = true
		showWH = true
		showLegend = true
	}

	renderOpts := &maprenderer.RenderOptions{
		Width:          c.Width,
		Height:         c.Height,
		ShowNames:      c.ShowNames,
		ShowFleets:     showFleets,
		ShowFleetPaths: c.FleetPaths,
		ShowMines:      c.ShowMines,
		ShowWormholes:  showWH,
		ShowLegend:     showLegend,
		Padding:        20,
	}

	// Determine if we're creating a GIF
	createGIF := c.GIF || c.Dir != "" || len(c.Args.Files) > 1

	if createGIF {
		return c.createAnimation(renderOpts)
	}
	return c.createSingleImage(renderOpts)
}

func (c *mapCommand) createSingleImage(renderOpts *maprenderer.RenderOptions) error {
	if len(c.Args.Files) == 0 {
		return fmt.Errorf("no input file specified")
	}

	filename := c.Args.Files[0]
	renderer := maprenderer.New()

	// Use LoadFileWithXY to automatically load companion XY file for M/H files
	if err := renderer.LoadFileWithXY(filename); err != nil {
		return fmt.Errorf("failed to load %s: %w", filename, err)
	}

	output := c.Output
	if c.SVG {
		if output == "" {
			output = filename + ".svg"
		}
		if err := renderer.SaveSVG(output, renderOpts); err != nil {
			return fmt.Errorf("failed to save SVG: %w", err)
		}
	} else {
		if output == "" {
			output = filename + ".png"
		}
		if err := renderer.SavePNG(output, renderOpts); err != nil {
			return fmt.Errorf("failed to save PNG: %w", err)
		}
	}

	fmt.Printf("Created %s\n", output)
	fmt.Printf("  Game ID: %d\n", renderer.GameID())
	fmt.Printf("  Year: %d (Turn %d)\n", renderer.Year(), renderer.Turn())
	fmt.Printf("  Planets: %d\n", renderer.PlanetCount())
	fmt.Printf("  Fleets: %d\n", renderer.FleetCount())

	return nil
}

func (c *mapCommand) createAnimation(renderOpts *maprenderer.RenderOptions) error {
	animator := maprenderer.NewAnimator()
	animator.SetOptions(renderOpts)

	// Load files from directory if specified
	if c.Dir != "" {
		fmt.Printf("Loading M files from %s...\n", c.Dir)
		files, err := findMFilesMap(c.Dir)
		if err != nil {
			return fmt.Errorf("failed to scan directory: %w", err)
		}
		for _, file := range files {
			fmt.Printf("Loading %s...\n", file)
			if err := animator.AddFile(file); err != nil {
				return fmt.Errorf("failed to load %s: %w", file, err)
			}
		}
	}

	// Load explicitly specified files
	for _, file := range c.Args.Files {
		fmt.Printf("Loading %s...\n", file)
		if err := animator.AddFile(file); err != nil {
			return fmt.Errorf("failed to load %s: %w", file, err)
		}
	}

	if animator.FrameCount() == 0 {
		return fmt.Errorf("no frames to animate")
	}

	// Sort frames by year
	animator.SortByYear()

	output := c.Output
	if output == "" {
		output = "animation.gif"
	}

	fmt.Printf("Creating animation with %d frames...\n", animator.FrameCount())

	if err := animator.SaveGIF(output, c.Delay); err != nil {
		return fmt.Errorf("failed to save GIF: %w", err)
	}

	fmt.Printf("Created %s\n", output)
	fmt.Printf("  Frames: %d\n", animator.FrameCount())
	fmt.Printf("  Delay: %d ms\n", c.Delay)

	return nil
}

func findMFilesMap(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if len(ext) >= 2 && ext[1] == 'm' {
			files = append(files, filepath.Join(dir, name))
		}
	}

	return files, nil
}

func addMapCommand(parser *flags.Parser) {
	_, err := parser.AddCommand("map",
		"Render galaxy maps as PNG or animated GIF",
		"Renders Stars! galaxy maps as PNG images or animated GIFs.\n\n"+
			"For single files, creates a PNG image showing planets, fleets, and other objects.\n"+
			"For multiple files or with --gif, creates an animated GIF showing the galaxy\n"+
			"over multiple turns.\n\n"+
			"Player colors are automatically assigned. Owned planets are shown in player colors,\n"+
			"while unowned planets are gray. Fleets are shown as directional triangles.",
		&mapCommand{})
	if err != nil {
		panic(err)
	}
}
