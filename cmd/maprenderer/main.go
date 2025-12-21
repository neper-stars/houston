// Command maprenderer renders Stars! galaxy maps as PNG or animated GIF images.
//
// Usage:
//
//	maprenderer [OPTIONS] <file>...
//
// Examples:
//
//	maprenderer game.m1                    # Render single file to PNG
//	maprenderer -o galaxy.png game.m1      # Specify output filename
//	maprenderer --gif -o anim.gif *.m1     # Create animated GIF from multiple files
//	maprenderer --dir ./backups -o anim.gif # Create GIF from directory
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/tools/maprenderer"
)

type options struct {
	Output     string `short:"o" long:"output" description:"Output filename (default: input.png or animation.gif)"`
	Width      int    `short:"W" long:"width" description:"Image width in pixels" default:"800"`
	Height     int    `short:"H" long:"height" description:"Image height in pixels" default:"600"`
	GIF        bool   `short:"g" long:"gif" description:"Create animated GIF from multiple files"`
	Dir        string `short:"d" long:"dir" description:"Load all M files from directory for animation"`
	Delay      int    `long:"delay" description:"Delay between frames in milliseconds" default:"1000"`
	ShowNames  bool   `short:"n" long:"names" description:"Show planet names"`
	ShowFleets bool   `short:"f" long:"fleets" description:"Show fleet indicators"`
	ShowMines  bool   `short:"m" long:"mines" description:"Show minefields"`
	ShowWH     bool   `short:"w" long:"wormholes" description:"Show wormholes"`
	ShowLegend bool   `short:"l" long:"legend" description:"Show player legend"`
	Args       struct {
		Files []string `positional-arg-name:"file" description:"Stars! game files to render"`
	} `positional-args:"yes"`
}

var description = `Renders Stars! galaxy maps as PNG images or animated GIFs.

For single files, creates a PNG image showing planets, fleets, and other objects.
For multiple files or with --gif, creates an animated GIF showing the galaxy
over multiple turns.

Player colors are automatically assigned. Owned planets are shown in player colors,
while unowned planets are gray. Fleets are shown as directional triangles.`

func main() {
	var opts options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "maprenderer"
	parser.LongDescription = description

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	// Check we have input
	if len(opts.Args.Files) == 0 && opts.Dir == "" {
		fmt.Fprintln(os.Stderr, "Error: no input files specified")
		fmt.Fprintln(os.Stderr, "Use --help for usage information")
		os.Exit(1)
	}

	// Set defaults for boolean options not explicitly set
	showFleets := opts.ShowFleets
	showWH := opts.ShowWH
	showLegend := opts.ShowLegend
	// If none of the display options are set, use sensible defaults
	if !opts.ShowFleets && !opts.ShowMines && !opts.ShowWH && !opts.ShowLegend && !opts.ShowNames {
		showFleets = true
		showWH = true
		showLegend = true
	}

	renderOpts := &maprenderer.RenderOptions{
		Width:         opts.Width,
		Height:        opts.Height,
		ShowNames:     opts.ShowNames,
		ShowFleets:    showFleets,
		ShowMines:     opts.ShowMines,
		ShowWormholes: showWH,
		ShowLegend:    showLegend,
		Padding:       20,
	}

	// Determine if we're creating a GIF
	createGIF := opts.GIF || opts.Dir != "" || len(opts.Args.Files) > 1

	if createGIF {
		if err := createAnimation(&opts, renderOpts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := createSingleImage(&opts, renderOpts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func createSingleImage(opts *options, renderOpts *maprenderer.RenderOptions) error {
	if len(opts.Args.Files) == 0 {
		return fmt.Errorf("no input file specified")
	}

	filename := opts.Args.Files[0]
	renderer := maprenderer.New()

	if err := renderer.LoadFile(filename); err != nil {
		return fmt.Errorf("failed to load %s: %w", filename, err)
	}

	output := opts.Output
	if output == "" {
		// Default to input filename with .png extension
		output = filename + ".png"
	}

	if err := renderer.SavePNG(output, renderOpts); err != nil {
		return fmt.Errorf("failed to save PNG: %w", err)
	}

	fmt.Printf("Created %s\n", output)
	fmt.Printf("  Game ID: %d\n", renderer.GameID())
	fmt.Printf("  Year: %d (Turn %d)\n", renderer.Year(), renderer.Turn())
	fmt.Printf("  Planets: %d\n", renderer.PlanetCount())
	fmt.Printf("  Fleets: %d\n", renderer.FleetCount())

	return nil
}

func createAnimation(opts *options, renderOpts *maprenderer.RenderOptions) error {
	animator := maprenderer.NewAnimator()
	animator.SetOptions(renderOpts)

	// Load files from directory if specified
	if opts.Dir != "" {
		fmt.Printf("Loading M files from %s...\n", opts.Dir)
		files, err := findMFiles(opts.Dir)
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
	for _, file := range opts.Args.Files {
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

	output := opts.Output
	if output == "" {
		output = "animation.gif"
	}

	fmt.Printf("Creating animation with %d frames...\n", animator.FrameCount())

	if err := animator.SaveGIF(output, opts.Delay); err != nil {
		return fmt.Errorf("failed to save GIF: %w", err)
	}

	fmt.Printf("Created %s\n", output)
	fmt.Printf("  Frames: %d\n", animator.FrameCount())
	fmt.Printf("  Delay: %d ms\n", opts.Delay)

	return nil
}

// findMFiles finds all M files in a directory.
func findMFiles(dir string) ([]string, error) {
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
