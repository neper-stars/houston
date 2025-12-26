// Package maprenderer provides functionality to render Stars! galaxy maps as images.
//
// It can create static PNG images or animated GIF files showing the galaxy
// state over multiple turns. The renderer displays planets, fleets, minefields,
// and wormholes with player-specific colors.
//
// Example usage:
//
//	renderer := maprenderer.New()
//	if err := renderer.LoadFile("game.m1"); err != nil {
//	    log.Fatal(err)
//	}
//	if err := renderer.SavePNG("galaxy.png", nil); err != nil {
//	    log.Fatal(err)
//	}
package maprenderer

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"
	"github.com/neper-stars/houston/store"
)

// Renderer holds the state for rendering galaxy maps.
type Renderer struct {
	store *store.GameStore

	// Map bounds (computed from entities)
	minX, maxX int
	minY, maxY int
}

// RenderOptions controls how the map is rendered.
type RenderOptions struct {
	Width          int  // Image width in pixels (default: 800)
	Height         int  // Image height in pixels (default: 600)
	ShowNames      bool // Show planet names
	ShowFleets     bool // Show fleet indicators
	ShowFleetPaths int  // Show fleet projected paths (0=off, N=years to project)
	ShowMines      bool // Show minefields
	ShowWormholes  bool // Show wormholes
	ShowLegend     bool // Show player legend
	Padding        int  // Padding around the galaxy (default: 20)
}

// DefaultOptions returns default rendering options.
func DefaultOptions() *RenderOptions {
	return &RenderOptions{
		Width:       800,
		Height:      600,
		ShowNames:   false,
		ShowFleets:  true,
		ShowMines:   false,
		ShowWormholes: true,
		ShowLegend:  true,
		Padding:     20,
	}
}

// New creates a new Renderer.
func New() *Renderer {
	return &Renderer{
		store: store.New(),
		minX:  math.MaxInt32,
		maxX:  math.MinInt32,
		minY:  math.MaxInt32,
		maxY:  math.MinInt32,
	}
}

// NewFromStore creates a Renderer from an existing GameStore.
func NewFromStore(gs *store.GameStore) *Renderer {
	r := &Renderer{
		store: gs,
		minX:  math.MaxInt32,
		maxX:  math.MinInt32,
		minY:  math.MaxInt32,
		maxY:  math.MinInt32,
	}
	r.computeBounds()
	return r
}

// Store returns the underlying GameStore.
func (r *Renderer) Store() *store.GameStore {
	return r.store
}

// LoadFile loads game data from a Stars! file.
func (r *Renderer) LoadFile(filename string) error {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	return r.LoadBytes(filename, fileBytes)
}

// LoadReader loads game data from an io.Reader.
func (r *Renderer) LoadReader(name string, reader io.Reader) error {
	fileBytes, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}
	return r.LoadBytes(name, fileBytes)
}

// LoadBytes loads game data from file bytes.
func (r *Renderer) LoadBytes(name string, fileBytes []byte) error {
	if err := r.store.AddFile(name, fileBytes); err != nil {
		return err
	}
	r.computeBounds()
	return nil
}

// LoadFileWithXY loads a game file and automatically loads the companion XY file
// if the input is an M or H file.
func (r *Renderer) LoadFileWithXY(filename string) error {
	// Use GameStore's auto-loading with XY file support
	if err := r.store.AddFileWithXY(filename); err != nil {
		return err
	}
	r.computeBounds()
	return nil
}

// computeBounds calculates the map bounds from all entities.
func (r *Renderer) computeBounds() {
	r.minX = math.MaxInt32
	r.maxX = math.MinInt32
	r.minY = math.MaxInt32
	r.maxY = math.MinInt32

	// Bounds from planets
	for _, planet := range r.store.AllPlanets() {
		r.updateBounds(planet.X, planet.Y)
	}

	// Bounds from fleets
	for _, fleet := range r.store.AllFleets() {
		r.updateBounds(fleet.X, fleet.Y)
	}

	// Bounds from minefields
	for _, mf := range r.store.Minefields() {
		r.updateBounds(mf.X, mf.Y)
	}

	// Bounds from wormholes
	for _, wh := range r.store.Wormholes() {
		r.updateBounds(wh.X, wh.Y)
	}
}

func (r *Renderer) updateBounds(x, y int) {
	if x < r.minX {
		r.minX = x
	}
	if x > r.maxX {
		r.maxX = x
	}
	if y < r.minY {
		r.minY = y
	}
	if y > r.maxY {
		r.maxY = y
	}
}

// Player colors - same as Java version
var playerColors = []color.RGBA{
	{255, 3, 3, 255},     // Red
	{0, 66, 255, 255},    // Blue
	{28, 230, 185, 255},  // Teal
	{84, 0, 129, 255},    // Purple
	{255, 252, 1, 255},   // Yellow
	{254, 138, 14, 255},  // Orange
	{32, 192, 0, 255},    // Green
	{229, 91, 176, 255},  // Pink
	{149, 150, 151, 255}, // Gray
	{126, 191, 241, 255}, // Light blue
	{16, 98, 70, 255},    // Dark green
	{78, 42, 4, 255},     // Brown
	{255, 255, 255, 255}, // White
	{187, 115, 20, 255},  // Gold
	{200, 100, 100, 255}, // Light red
	{100, 100, 200, 255}, // Light purple
}

// GetPlayerColor returns the color for a player.
func (r *Renderer) GetPlayerColor(playerNum int) color.RGBA {
	if playerNum >= 0 && playerNum < len(playerColors) {
		return playerColors[playerNum]
	}
	return color.RGBA{128, 128, 128, 255}
}

// getPlayerName returns the name for a player (from GameStore).
func (r *Renderer) getPlayerName(playerNum int) string {
	if player, ok := r.store.Player(playerNum); ok {
		if player.NameSingular != "" {
			return player.NameSingular
		}
	}
	return fmt.Sprintf("Player %d", playerNum+1)
}

// Render creates an image of the galaxy map.
func (r *Renderer) Render(opts *RenderOptions) *image.RGBA {
	if opts == nil {
		opts = DefaultOptions()
	}

	img := image.NewRGBA(image.Rect(0, 0, opts.Width, opts.Height))

	// Fill background with black
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	// Calculate scaling
	galaxyWidth := float64(r.maxX - r.minX)
	galaxyHeight := float64(r.maxY - r.minY)
	if galaxyWidth == 0 {
		galaxyWidth = 1
	}
	if galaxyHeight == 0 {
		galaxyHeight = 1
	}

	availWidth := float64(opts.Width - 2*opts.Padding)
	availHeight := float64(opts.Height - 2*opts.Padding)

	scaleX := availWidth / galaxyWidth
	scaleY := availHeight / galaxyHeight
	scale := math.Min(scaleX, scaleY)

	// Calculate offsets to center the galaxy
	offsetX := float64(opts.Padding) + (availWidth-galaxyWidth*scale)/2
	offsetY := float64(opts.Padding) + (availHeight-galaxyHeight*scale)/2

	// Transform function
	transform := func(x, y int) (int, int) {
		px := int(offsetX + float64(x-r.minX)*scale)
		py := int(offsetY + float64(r.maxY-y)*scale) // Flip Y axis
		return px, py
	}

	// Draw minefields first (background) as cloud of dots
	if opts.ShowMines {
		for _, mf := range r.store.Minefields() {
			px, py := transform(mf.X, mf.Y)
			radius := int(math.Sqrt(float64(mf.MineCount)) * scale / 10)
			if radius < 2 {
				radius = 2
			}
			col := r.GetPlayerColor(mf.Owner)
			col.A = 180 // Semi-transparent
			drawMinefieldCloud(img, px, py, radius, col, mf.Number)
		}
	}

	// Draw wormholes
	if opts.ShowWormholes {
		purple := color.RGBA{128, 0, 128, 255}
		wormholes := r.store.Wormholes()
		// Build lookup map for wormhole connections
		whByID := make(map[int]*store.ObjectEntity)
		for _, wh := range wormholes {
			whByID[wh.WormholeId] = wh
		}
		for _, wh := range wormholes {
			px, py := transform(wh.X, wh.Y)
			drawFilledCircle(img, px, py, 4, purple)
			// Draw connection to target if known
			if target, ok := whByID[wh.TargetId]; ok {
				tx, ty := transform(target.X, target.Y)
				drawLine(img, px, py, tx, ty, color.RGBA{255, 0, 255, 128})
			}
		}
	}

	// Draw planets
	for _, planet := range r.store.AllPlanets() {
		px, py := transform(planet.X, planet.Y)

		var col color.RGBA
		radius := 2

		if planet.Owner >= 0 {
			col = r.GetPlayerColor(planet.Owner)
			radius = 3
		} else {
			col = color.RGBA{128, 128, 128, 255}
		}

		// Draw starbase if present (white circle + yellow satellite)
		if planet.HasStarbase {
			// White circle around planet
			drawCircleOutline(img, px, py, 6, color.RGBA{255, 255, 255, 255})
			// Yellow satellite dot offset to upper-right
			drawFilledCircle(img, px+5, py-5, 1, color.RGBA{255, 255, 0, 255})
		}

		drawFilledCircle(img, px, py, radius, col)
	}

	// Draw fleets
	if opts.ShowFleets {
		for _, fleet := range r.store.AllFleets() {
			px, py := transform(fleet.X, fleet.Y)
			col := r.GetPlayerColor(fleet.Owner)
			col.A = 200

			// Draw direction triangle
			// DeltaX/DeltaY are signed velocity components (-128 to 127)
			dx := float64(fleet.DeltaX)
			dy := -float64(fleet.DeltaY) // Flip Y for screen coords
			drawFleetTriangle(img, px, py, dx, dy, col)
		}
	}

	// Draw legend
	if opts.ShowLegend {
		r.drawLegend(img, opts)
	}

	// Draw year
	r.drawYear(img, opts)

	return img
}

func (r *Renderer) drawLegend(img *image.RGBA, opts *RenderOptions) {
	// Get players from store and sort by number
	players := r.store.AllPlayers()
	sort.Slice(players, func(i, j int) bool {
		return players[i].PlayerNumber < players[j].PlayerNumber
	})

	y := 10
	for _, player := range players {
		col := r.GetPlayerColor(player.PlayerNumber)
		// Draw color box
		for dy := 0; dy < 10; dy++ {
			for dx := 0; dx < 10; dx++ {
				img.Set(5+dx, y+dy, col)
			}
		}
		// Draw player name
		name := player.NameSingular
		if name == "" {
			name = fmt.Sprintf("Player %d", player.PlayerNumber+1)
		}
		drawText(img, 20, y+2, name, col)
		y += 14
	}
}

func (r *Renderer) drawYear(img *image.RGBA, opts *RenderOptions) {
	// Draw year in bottom left corner
	// Simple representation with colored pixels
	yearStr := fmt.Sprintf("%d", r.Year())
	x := 10
	y := opts.Height - 20

	// Draw each digit
	for _, ch := range yearStr {
		digit := int(ch - '0')
		drawDigit(img, x, y, digit, color.RGBA{0, 128, 255, 255})
		x += 8
	}
}

// SavePNG saves the rendered map as a PNG file.
func (r *Renderer) SavePNG(filename string, opts *RenderOptions) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	return r.WritePNG(f, opts)
}

// WritePNG writes the rendered map as PNG to an io.Writer.
// Uses SVG-based rendering for higher quality anti-aliased output.
func (r *Renderer) WritePNG(w io.Writer, opts *RenderOptions) error {
	// Use SVG-based rendering for better quality
	img, err := r.RenderSVGToImage(opts)
	if err != nil {
		// Fall back to basic rendering if SVG fails
		img = r.Render(opts)
	}

	if err := png.Encode(w, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

// SaveSVG saves the rendered map as SVG to a file.
func (r *Renderer) SaveSVG(filename string, opts *RenderOptions) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	return r.WriteSVG(f, opts)
}

// WriteSVG writes the rendered map as SVG to an io.Writer.
func (r *Renderer) WriteSVG(w io.Writer, opts *RenderOptions) error {
	svg := r.RenderSVG(opts)
	_, err := w.Write([]byte(svg))
	return err
}

// RenderSVG renders the map as an SVG string.
func (r *Renderer) RenderSVG(opts *RenderOptions) string {
	svg := r.buildSVG(opts)
	return svg.String()
}

// renderSVGForRasterization renders an SVG compatible with oksvg rasterization.
func (r *Renderer) renderSVGForRasterization(opts *RenderOptions) string {
	svg := r.buildSVG(opts)
	return svg.StringForRasterization()
}

// buildSVG builds the SVG structure (shared between RenderSVG and rasterization).
func (r *Renderer) buildSVG(opts *RenderOptions) *SVGBuilder {
	if opts == nil {
		opts = DefaultOptions()
	}

	svg := NewSVGBuilder(opts.Width, opts.Height)

	// Add patterns and markers
	svg.AddMinefieldHatchPattern()

	// Calculate scale and transform
	if r.minX == math.MaxInt32 || r.maxX == math.MinInt32 {
		return svg
	}

	rangeX := float64(r.maxX - r.minX)
	rangeY := float64(r.maxY - r.minY)
	if rangeX == 0 {
		rangeX = 1
	}
	if rangeY == 0 {
		rangeY = 1
	}

	padding := float64(opts.Padding)
	availWidth := float64(opts.Width) - 2*padding
	availHeight := float64(opts.Height) - 2*padding

	scaleX := availWidth / rangeX
	scaleY := availHeight / rangeY
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	offsetX := padding + (availWidth-rangeX*scale)/2
	offsetY := padding + (availHeight-rangeY*scale)/2

	transform := func(x, y int) (float64, float64) {
		px := offsetX + float64(x-r.minX)*scale
		py := offsetY + float64(r.maxY-y)*scale // Flip Y axis
		return px, py
	}

	// Add arrow markers for fleet paths (one per player color)
	if opts.ShowFleetPaths > 0 {
		for _, player := range r.store.AllPlayers() {
			markerID := fmt.Sprintf("arrow-%d", player.PlayerNumber)
			col := r.GetPlayerColor(player.PlayerNumber)
			svg.AddArrowMarker(markerID, col)
		}
	}

	// Draw minefields
	if opts.ShowMines {
		for _, mf := range r.store.Minefields() {
			px, py := transform(mf.X, mf.Y)
			radius := math.Sqrt(float64(mf.MineCount)) * scale / 10
			if radius < 2 {
				radius = 2
			}
			col := r.GetPlayerColor(mf.Owner)
			svg.Minefield(px, py, radius, col)
		}
	}

	// Draw wormholes
	if opts.ShowWormholes {
		for _, wh := range r.store.Wormholes() {
			px, py := transform(wh.X, wh.Y)
			svg.Wormhole(px, py)
		}
	}

	// Draw fleet projected paths (before fleets so paths are behind)
	if opts.ShowFleetPaths > 0 {
		for _, fleet := range r.store.AllFleets() {
			col := r.GetPlayerColor(fleet.Owner)
			markerID := fmt.Sprintf("arrow-%d", fleet.Owner)

			// Check if fleet has waypoints (owned fleets)
			if len(fleet.Waypoints) > 0 {
				// Build path from current position through all waypoints
				var points [][2]float64
				px, py := transform(fleet.X, fleet.Y)
				points = append(points, [2]float64{px, py})

				for _, wp := range fleet.Waypoints {
					// Skip waypoints at current fleet position
					if wp.X == fleet.X && wp.Y == fleet.Y {
						continue
					}
					wpx, wpy := transform(wp.X, wp.Y)
					// Avoid duplicate consecutive points
					if len(points) > 0 {
						last := points[len(points)-1]
						if math.Abs(wpx-last[0]) < 0.5 && math.Abs(wpy-last[1]) < 0.5 {
							continue
						}
					}
					points = append(points, [2]float64{wpx, wpy})
				}

				svg.WaypointPath(points, col, markerID)
			} else {
				// Use DeltaX/DeltaY for enemy fleets
				dx := float64(fleet.DeltaX)
				dy := float64(fleet.DeltaY)
				if math.Abs(dx) < 0.5 && math.Abs(dy) < 0.5 {
					continue // Stationary
				}

				px, py := transform(fleet.X, fleet.Y)

				// Scale the delta to screen coordinates
				// DeltaX/DeltaY are in game units per turn
				screenDx := dx * scale
				screenDy := -dy * scale // Flip Y

				svg.FleetSpeedLine(px, py, screenDx, screenDy, opts.ShowFleetPaths, col, markerID)
			}
		}
	}

	// Draw planets
	for _, planet := range r.store.AllPlanets() {
		px, py := transform(planet.X, planet.Y)

		var col color.RGBA
		radius := 2.0

		if planet.Owner >= 0 {
			col = r.GetPlayerColor(planet.Owner)
			radius = 3.0
		} else {
			col = color.RGBA{128, 128, 128, 255}
		}

		svg.Planet(px, py, radius, col, planet.HasStarbase, planet.Name, opts.ShowNames)
	}

	// Draw fleets
	if opts.ShowFleets {
		for _, fleet := range r.store.AllFleets() {
			px, py := transform(fleet.X, fleet.Y)
			col := r.GetPlayerColor(fleet.Owner)

			var dx, dy float64
			isMoving := false

			// Check for waypoints first (owned fleets moving to waypoint)
			if len(fleet.Waypoints) > 0 {
				// Find first waypoint that's not at current position
				for _, wp := range fleet.Waypoints {
					if wp.X != fleet.X || wp.Y != fleet.Y {
						wpx, wpy := transform(wp.X, wp.Y)
						dx = wpx - px
						dy = wpy - py
						isMoving = true
						break
					}
				}
			}

			// If no waypoint movement, check DeltaX/DeltaY (enemy fleets)
			if !isMoving {
				dx = float64(fleet.DeltaX)
				dy = -float64(fleet.DeltaY) // Flip Y for screen coords
				isMoving = math.Abs(dx) >= 0.5 || math.Abs(dy) >= 0.5
			}

			if !isMoving {
				svg.Diamond(px, py, 3, col)
			} else {
				angle := math.Atan2(dy, dx)
				svg.Triangle(px, py, 4, angle, col)
			}
		}
	}

	// Draw legend
	if opts.ShowLegend {
		players := r.store.AllPlayers()
		sort.Slice(players, func(i, j int) bool {
			return players[i].PlayerNumber < players[j].PlayerNumber
		})

		y := 10.0
		for _, player := range players {
			col := r.GetPlayerColor(player.PlayerNumber)
			name := player.NameSingular
			if name == "" {
				name = fmt.Sprintf("Player %d", player.PlayerNumber+1)
			}
			svg.LegendItem(5, y, name, col)
			y += 14
		}
	}

	// Draw year
	svg.Text(10, float64(opts.Height-10), fmt.Sprintf("%d", r.Year()), color.RGBA{0, 128, 255, 255}, 12)

	return svg
}

// RenderSVGToImage renders the map as SVG and rasterizes it to an RGBA image.
// This produces higher quality output with anti-aliased circles and lines.
func (r *Renderer) RenderSVGToImage(opts *RenderOptions) (*image.RGBA, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Generate SVG (use rasterization-compatible version without markers/patterns
	// that may contain unsupported color syntax)
	svgStr := r.renderSVGForRasterization(opts)

	// Parse SVG using tdewolff/canvas
	c, err := canvas.ParseSVG(strings.NewReader(svgStr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SVG: %w", err)
	}

	// Calculate DPI to get the exact requested pixel size
	// The SVG is sized in pixels, canvas treats them as points (1/72 inch)
	// To get exact pixel output: dpi = targetWidth / canvasWidth * 72
	w, h := opts.Width, opts.Height
	canvasW := c.W // canvas width in mm
	if canvasW <= 0 {
		canvasW = float64(w) // fallback
	}
	// DPMM = dots per mm, we want w pixels for canvasW mm
	dpmm := float64(w) / canvasW

	img := rasterizer.Draw(c, canvas.DPMM(dpmm), canvas.DefaultColorSpace)

	// If the image is still not the right size, resize it
	bounds := img.Bounds()
	if bounds.Dx() != w || bounds.Dy() != h {
		// Create properly sized image and scale
		rgba := image.NewRGBA(image.Rect(0, 0, w, h))
		// Simple nearest-neighbor scaling for now
		scaleX := float64(bounds.Dx()) / float64(w)
		scaleY := float64(bounds.Dy()) / float64(h)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				srcX := int(float64(x) * scaleX)
				srcY := int(float64(y) * scaleY)
				rgba.Set(x, y, img.At(srcX, srcY))
			}
		}
		return rgba, nil
	}

	// Convert to RGBA if needed
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	return rgba, nil
}

// RenderBytes returns the rendered map as PNG bytes.
func (r *Renderer) RenderBytes(opts *RenderOptions) ([]byte, error) {
	var buf bytes.Buffer
	if err := r.WritePNG(&buf, opts); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Year returns the game year.
func (r *Renderer) Year() int {
	return int(r.store.Turn) + 2400
}

// Turn returns the game turn number.
func (r *Renderer) Turn() uint16 {
	return r.store.Turn
}

// GameID returns the game ID.
func (r *Renderer) GameID() uint32 {
	return r.store.GameID
}

// PlanetCount returns the number of planets.
func (r *Renderer) PlanetCount() int {
	return len(r.store.AllPlanets())
}

// FleetCount returns the number of fleets.
func (r *Renderer) FleetCount() int {
	return len(r.store.AllFleets())
}

// Animator creates animated GIFs from multiple game files.
// Files from the same year are automatically merged into a single frame.
type Animator struct {
	// framesByYear maps year to renderer, merging multiple files per year
	framesByYear map[int]*Renderer
	// renderers is the sorted list of frames (built from framesByYear)
	renderers []*Renderer
	opts      *RenderOptions
}

// NewAnimator creates a new Animator.
func NewAnimator() *Animator {
	return &Animator{
		framesByYear: make(map[int]*Renderer),
		opts:         DefaultOptions(),
	}
}

// SetOptions sets the rendering options for all frames.
func (a *Animator) SetOptions(opts *RenderOptions) {
	a.opts = opts
}

// AddFile adds a game file. Files from the same year are merged into a single frame.
func (a *Animator) AddFile(filename string) error {
	// First load the file to get its year
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Create a temporary renderer to get the year
	tempR := New()
	if err := tempR.LoadBytes(filename, data); err != nil {
		return err
	}
	year := tempR.Year()

	// Check if we already have a frame for this year
	if existingR, ok := a.framesByYear[year]; ok {
		// Merge into existing renderer's store
		if err := existingR.store.AddFile(filename, data); err != nil {
			return err
		}
		existingR.computeBounds()
	} else {
		// Use LoadFileWithXY to also load companion XY file
		r := New()
		if err := r.LoadFileWithXY(filename); err != nil {
			return err
		}
		a.framesByYear[year] = r
	}

	return nil
}

// AddBytes adds game data from bytes. Files from the same year are merged into a single frame.
func (a *Animator) AddBytes(name string, data []byte) error {
	// Create a temporary renderer to get the year
	tempR := New()
	if err := tempR.LoadBytes(name, data); err != nil {
		return err
	}
	year := tempR.Year()

	// Check if we already have a frame for this year
	if existingR, ok := a.framesByYear[year]; ok {
		// Merge into existing renderer's store
		if err := existingR.store.AddFile(name, data); err != nil {
			return err
		}
		existingR.computeBounds()
	} else {
		a.framesByYear[year] = tempR
	}

	return nil
}

// AddReader adds game data from an io.Reader. Files from the same year are merged into a single frame.
func (a *Animator) AddReader(name string, reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}
	return a.AddBytes(name, data)
}

// SortByYear builds the sorted renderers list from framesByYear.
func (a *Animator) SortByYear() {
	// Build sorted list from map
	a.renderers = make([]*Renderer, 0, len(a.framesByYear))
	for _, r := range a.framesByYear {
		a.renderers = append(a.renderers, r)
	}
	sort.Slice(a.renderers, func(i, j int) bool {
		return a.renderers[i].Year() < a.renderers[j].Year()
	})
}

// FrameCount returns the number of frames (unique years).
func (a *Animator) FrameCount() int {
	if len(a.renderers) > 0 {
		return len(a.renderers)
	}
	return len(a.framesByYear)
}

// SaveGIF saves all frames as an animated GIF.
func (a *Animator) SaveGIF(filename string, delayMs int) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	return a.WriteGIF(f, delayMs)
}

// WriteGIF writes all frames as an animated GIF to an io.Writer.
// Uses SVG-based rendering for higher quality anti-aliased output.
func (a *Animator) WriteGIF(w io.Writer, delayMs int) error {
	if len(a.renderers) == 0 {
		return fmt.Errorf("no frames to save")
	}

	// delay is in 100ths of a second
	delay := delayMs / 10

	anim := gif.GIF{
		LoopCount: 0, // Loop forever
	}

	for i, r := range a.renderers {
		// Use SVG-based rendering for better quality
		img, err := r.RenderSVGToImage(a.opts)
		if err != nil {
			// Log the error and fall back to basic rendering
			fmt.Fprintf(os.Stderr, "Warning: SVG rendering failed for frame %d (year %d): %v, using bitmap fallback\n", i, r.Year(), err)
			img = r.Render(a.opts)
		}
		paletted := imageToPaletted(img)
		anim.Image = append(anim.Image, paletted)
		anim.Delay = append(anim.Delay, delay)
	}

	if err := gif.EncodeAll(w, &anim); err != nil {
		return fmt.Errorf("failed to encode GIF: %w", err)
	}

	return nil
}

// RenderGIFBytes returns all frames as an animated GIF in bytes.
func (a *Animator) RenderGIFBytes(delayMs int) ([]byte, error) {
	var buf bytes.Buffer
	if err := a.WriteGIF(&buf, delayMs); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// imageToPaletted converts an RGBA image to a paletted image.
func imageToPaletted(img *image.RGBA) *image.Paletted {
	bounds := img.Bounds()

	// Create a color palette with the most common colors
	colorMap := make(map[color.RGBA]int)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			c := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
			colorMap[c]++
		}
	}

	// Sort colors by frequency and take top 256
	type colorCount struct {
		c     color.RGBA
		count int
	}
	var colors []colorCount
	for c, count := range colorMap {
		colors = append(colors, colorCount{c, count})
	}
	sort.Slice(colors, func(i, j int) bool {
		return colors[i].count > colors[j].count
	})

	palette := make(color.Palette, 0, 256)
	for i := 0; i < len(colors) && i < 256; i++ {
		palette = append(palette, colors[i].c)
	}

	// Create paletted image
	paletted := image.NewPaletted(bounds, palette)
	draw.FloydSteinberg.Draw(paletted, bounds, img, bounds.Min)

	return paletted
}

// Drawing helper functions

func drawFilledCircle(img *image.RGBA, cx, cy, radius int, col color.RGBA) {
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			if x*x+y*y <= radius*radius {
				img.Set(cx+x, cy+y, col)
			}
		}
	}
}

func drawCircleOutline(img *image.RGBA, cx, cy, radius int, col color.RGBA) {
	for angle := 0.0; angle < 2*math.Pi; angle += 0.1 {
		x := cx + int(float64(radius)*math.Cos(angle))
		y := cy + int(float64(radius)*math.Sin(angle))
		img.Set(x, y, col)
	}
}

func drawLine(img *image.RGBA, x0, y0, x1, y1 int, col color.RGBA) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx - dy

	for {
		img.Set(x0, y0, col)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func drawFleetTriangle(img *image.RGBA, cx, cy int, dx, dy float64, col color.RGBA) {
	// Determine direction using angle
	var points [][2]int

	// Check if stationary (both dx and dy near zero)
	if math.Abs(dx) < 0.5 && math.Abs(dy) < 0.5 {
		// Stationary - draw a diamond
		points = [][2]int{{cx, cy - 3}, {cx + 3, cy}, {cx, cy + 3}, {cx - 3, cy}}
		// Draw diamond
		drawLine(img, points[0][0], points[0][1], points[1][0], points[1][1], col)
		drawLine(img, points[1][0], points[1][1], points[2][0], points[2][1], col)
		drawLine(img, points[2][0], points[2][1], points[3][0], points[3][1], col)
		drawLine(img, points[3][0], points[3][1], points[0][0], points[0][1], col)
		return
	}

	// Calculate angle and draw triangle pointing in direction of movement
	angle := math.Atan2(dy, dx)
	size := 4.0

	// Triangle tip points in direction of movement
	tipX := cx + int(math.Cos(angle)*size)
	tipY := cy + int(math.Sin(angle)*size)

	// Base points perpendicular to direction
	baseAngle1 := angle + math.Pi*2/3
	baseAngle2 := angle - math.Pi*2/3

	base1X := cx + int(math.Cos(baseAngle1)*size)
	base1Y := cy + int(math.Sin(baseAngle1)*size)
	base2X := cx + int(math.Cos(baseAngle2)*size)
	base2Y := cy + int(math.Sin(baseAngle2)*size)

	points = [][2]int{{tipX, tipY}, {base1X, base1Y}, {base2X, base2Y}}

	// Draw triangle outline
	drawLine(img, points[0][0], points[0][1], points[1][0], points[1][1], col)
	drawLine(img, points[1][0], points[1][1], points[2][0], points[2][1], col)
	drawLine(img, points[2][0], points[2][1], points[0][0], points[0][1], col)
}

// Simple 3x5 digit font for year display
var digitPatterns = [10][5][3]bool{
	{ // 0
		{true, true, true},
		{true, false, true},
		{true, false, true},
		{true, false, true},
		{true, true, true},
	},
	{ // 1
		{false, true, false},
		{true, true, false},
		{false, true, false},
		{false, true, false},
		{true, true, true},
	},
	{ // 2
		{true, true, true},
		{false, false, true},
		{true, true, true},
		{true, false, false},
		{true, true, true},
	},
	{ // 3
		{true, true, true},
		{false, false, true},
		{true, true, true},
		{false, false, true},
		{true, true, true},
	},
	{ // 4
		{true, false, true},
		{true, false, true},
		{true, true, true},
		{false, false, true},
		{false, false, true},
	},
	{ // 5
		{true, true, true},
		{true, false, false},
		{true, true, true},
		{false, false, true},
		{true, true, true},
	},
	{ // 6
		{true, true, true},
		{true, false, false},
		{true, true, true},
		{true, false, true},
		{true, true, true},
	},
	{ // 7
		{true, true, true},
		{false, false, true},
		{false, false, true},
		{false, false, true},
		{false, false, true},
	},
	{ // 8
		{true, true, true},
		{true, false, true},
		{true, true, true},
		{true, false, true},
		{true, true, true},
	},
	{ // 9
		{true, true, true},
		{true, false, true},
		{true, true, true},
		{false, false, true},
		{true, true, true},
	},
}

func drawDigit(img *image.RGBA, x, y, digit int, col color.RGBA) {
	if digit < 0 || digit > 9 {
		return
	}
	pattern := digitPatterns[digit]
	for dy := 0; dy < 5; dy++ {
		for dx := 0; dx < 3; dx++ {
			if pattern[dy][dx] {
				// Draw 2x2 pixels for each "pixel" in the pattern
				for py := 0; py < 2; py++ {
					for px := 0; px < 2; px++ {
						img.Set(x+dx*2+px, y+dy*2+py, col)
					}
				}
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Letter patterns for A-Z (3x5 bitmap font)
var letterPatterns = map[rune][5][3]bool{
	'A': {{false, true, false}, {true, false, true}, {true, true, true}, {true, false, true}, {true, false, true}},
	'B': {{true, true, false}, {true, false, true}, {true, true, false}, {true, false, true}, {true, true, false}},
	'C': {{false, true, true}, {true, false, false}, {true, false, false}, {true, false, false}, {false, true, true}},
	'D': {{true, true, false}, {true, false, true}, {true, false, true}, {true, false, true}, {true, true, false}},
	'E': {{true, true, true}, {true, false, false}, {true, true, false}, {true, false, false}, {true, true, true}},
	'F': {{true, true, true}, {true, false, false}, {true, true, false}, {true, false, false}, {true, false, false}},
	'G': {{false, true, true}, {true, false, false}, {true, false, true}, {true, false, true}, {false, true, true}},
	'H': {{true, false, true}, {true, false, true}, {true, true, true}, {true, false, true}, {true, false, true}},
	'I': {{true, true, true}, {false, true, false}, {false, true, false}, {false, true, false}, {true, true, true}},
	'J': {{false, false, true}, {false, false, true}, {false, false, true}, {true, false, true}, {false, true, false}},
	'K': {{true, false, true}, {true, false, true}, {true, true, false}, {true, false, true}, {true, false, true}},
	'L': {{true, false, false}, {true, false, false}, {true, false, false}, {true, false, false}, {true, true, true}},
	'M': {{true, false, true}, {true, true, true}, {true, false, true}, {true, false, true}, {true, false, true}},
	'N': {{true, false, true}, {true, true, true}, {true, true, true}, {true, false, true}, {true, false, true}},
	'O': {{false, true, false}, {true, false, true}, {true, false, true}, {true, false, true}, {false, true, false}},
	'P': {{true, true, false}, {true, false, true}, {true, true, false}, {true, false, false}, {true, false, false}},
	'Q': {{false, true, false}, {true, false, true}, {true, false, true}, {true, true, true}, {false, true, true}},
	'R': {{true, true, false}, {true, false, true}, {true, true, false}, {true, false, true}, {true, false, true}},
	'S': {{false, true, true}, {true, false, false}, {false, true, false}, {false, false, true}, {true, true, false}},
	'T': {{true, true, true}, {false, true, false}, {false, true, false}, {false, true, false}, {false, true, false}},
	'U': {{true, false, true}, {true, false, true}, {true, false, true}, {true, false, true}, {false, true, false}},
	'V': {{true, false, true}, {true, false, true}, {true, false, true}, {false, true, false}, {false, true, false}},
	'W': {{true, false, true}, {true, false, true}, {true, false, true}, {true, true, true}, {true, false, true}},
	'X': {{true, false, true}, {true, false, true}, {false, true, false}, {true, false, true}, {true, false, true}},
	'Y': {{true, false, true}, {true, false, true}, {false, true, false}, {false, true, false}, {false, true, false}},
	'Z': {{true, true, true}, {false, false, true}, {false, true, false}, {true, false, false}, {true, true, true}},
	// Lowercase (same as uppercase for simplicity)
	'a': {{false, true, false}, {true, false, true}, {true, true, true}, {true, false, true}, {true, false, true}},
	'b': {{true, true, false}, {true, false, true}, {true, true, false}, {true, false, true}, {true, true, false}},
	'c': {{false, true, true}, {true, false, false}, {true, false, false}, {true, false, false}, {false, true, true}},
	'd': {{true, true, false}, {true, false, true}, {true, false, true}, {true, false, true}, {true, true, false}},
	'e': {{true, true, true}, {true, false, false}, {true, true, false}, {true, false, false}, {true, true, true}},
	'f': {{true, true, true}, {true, false, false}, {true, true, false}, {true, false, false}, {true, false, false}},
	'g': {{false, true, true}, {true, false, false}, {true, false, true}, {true, false, true}, {false, true, true}},
	'h': {{true, false, true}, {true, false, true}, {true, true, true}, {true, false, true}, {true, false, true}},
	'i': {{true, true, true}, {false, true, false}, {false, true, false}, {false, true, false}, {true, true, true}},
	'j': {{false, false, true}, {false, false, true}, {false, false, true}, {true, false, true}, {false, true, false}},
	'k': {{true, false, true}, {true, false, true}, {true, true, false}, {true, false, true}, {true, false, true}},
	'l': {{true, false, false}, {true, false, false}, {true, false, false}, {true, false, false}, {true, true, true}},
	'm': {{true, false, true}, {true, true, true}, {true, false, true}, {true, false, true}, {true, false, true}},
	'n': {{true, false, true}, {true, true, true}, {true, true, true}, {true, false, true}, {true, false, true}},
	'o': {{false, true, false}, {true, false, true}, {true, false, true}, {true, false, true}, {false, true, false}},
	'p': {{true, true, false}, {true, false, true}, {true, true, false}, {true, false, false}, {true, false, false}},
	'q': {{false, true, false}, {true, false, true}, {true, false, true}, {true, true, true}, {false, true, true}},
	'r': {{true, true, false}, {true, false, true}, {true, true, false}, {true, false, true}, {true, false, true}},
	's': {{false, true, true}, {true, false, false}, {false, true, false}, {false, false, true}, {true, true, false}},
	't': {{true, true, true}, {false, true, false}, {false, true, false}, {false, true, false}, {false, true, false}},
	'u': {{true, false, true}, {true, false, true}, {true, false, true}, {true, false, true}, {false, true, false}},
	'v': {{true, false, true}, {true, false, true}, {true, false, true}, {false, true, false}, {false, true, false}},
	'w': {{true, false, true}, {true, false, true}, {true, false, true}, {true, true, true}, {true, false, true}},
	'x': {{true, false, true}, {true, false, true}, {false, true, false}, {true, false, true}, {true, false, true}},
	'y': {{true, false, true}, {true, false, true}, {false, true, false}, {false, true, false}, {false, true, false}},
	'z': {{true, true, true}, {false, false, true}, {false, true, false}, {true, false, false}, {true, true, true}},
	' ': {{false, false, false}, {false, false, false}, {false, false, false}, {false, false, false}, {false, false, false}},
}

// drawText draws a string using the bitmap font
func drawText(img *image.RGBA, x, y int, text string, col color.RGBA) {
	startX := x
	for _, ch := range text {
		if ch == '\n' {
			x = startX
			y += 12
			continue
		}
		// Try letter patterns first
		if pattern, ok := letterPatterns[ch]; ok {
			drawPattern(img, x, y, pattern, col)
			x += 8
		} else if ch >= '0' && ch <= '9' {
			// Use digit patterns
			drawDigit(img, x, y, int(ch-'0'), col)
			x += 8
		} else {
			// Unknown character - skip space
			x += 8
		}
	}
}

// drawPattern draws a 3x5 pattern at the given position
func drawPattern(img *image.RGBA, x, y int, pattern [5][3]bool, col color.RGBA) {
	for dy := 0; dy < 5; dy++ {
		for dx := 0; dx < 3; dx++ {
			if pattern[dy][dx] {
				// Draw 2x2 pixels for each "pixel" in the pattern
				for py := 0; py < 2; py++ {
					for px := 0; px < 2; px++ {
						img.Set(x+dx*2+px, y+dy*2+py, col)
					}
				}
			}
		}
	}
}

// drawMinefieldCloud draws a minefield with diagonal line hatching
func drawMinefieldCloud(img *image.RGBA, cx, cy, radius int, col color.RGBA, seed int) {
	radiusSq := radius * radius
	spacing := 3 // Spacing between diagonal lines

	// Draw diagonal lines (going from top-left to bottom-right)
	// For each diagonal line offset
	for offset := -radius * 2; offset <= radius*2; offset += spacing {
		// Draw points along the diagonal where x + y = offset (relative to center)
		for dy := -radius; dy <= radius; dy++ {
			dx := offset - dy
			// Check if within circle
			if dx >= -radius && dx <= radius && dx*dx+dy*dy <= radiusSq {
				img.Set(cx+dx, cy+dy, col)
			}
		}
	}
}
