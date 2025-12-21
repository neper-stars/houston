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

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

// Renderer holds the state for rendering galaxy maps.
type Renderer struct {
	planets    map[int]*PlanetData
	fleets     map[int]*FleetData
	minefields map[int]*MinefieldData
	wormholes  map[int]*WormholeData
	players    map[int]*PlayerData

	gameID uint32
	turn   uint16
	year   int

	// Map bounds
	minX, maxX int
	minY, maxY int
}

// PlanetData contains information about a planet for rendering.
type PlanetData struct {
	ID         int
	Name       string
	X, Y       int
	Owner      int
	IsHomeworld bool
	Population int
}

// FleetData contains information about a fleet for rendering.
type FleetData struct {
	ID      int
	Owner   int
	X, Y    int
	DeltaX  int
	DeltaY  int
	Mass    int
	Ships   int
}

// MinefieldData contains information about a minefield for rendering.
type MinefieldData struct {
	ID        int
	Owner     int
	X, Y      int
	MineCount int
	Type      int
}

// WormholeData contains information about a wormhole for rendering.
type WormholeData struct {
	ID       int
	X, Y     int
	TargetID int
}

// PlayerData contains information about a player for rendering.
type PlayerData struct {
	Number int
	Name   string
	Color  color.RGBA
}

// RenderOptions controls how the map is rendered.
type RenderOptions struct {
	Width       int  // Image width in pixels (default: 800)
	Height      int  // Image height in pixels (default: 600)
	ShowNames   bool // Show planet names
	ShowFleets  bool // Show fleet indicators
	ShowMines   bool // Show minefields
	ShowWormholes bool // Show wormholes
	ShowLegend  bool // Show player legend
	Padding     int  // Padding around the galaxy (default: 20)
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
		planets:    make(map[int]*PlanetData),
		fleets:     make(map[int]*FleetData),
		minefields: make(map[int]*MinefieldData),
		wormholes:  make(map[int]*WormholeData),
		players:    make(map[int]*PlayerData),
		minX:       math.MaxInt32,
		maxX:       math.MinInt32,
		minY:       math.MaxInt32,
		maxY:       math.MinInt32,
	}
}

// LoadFile loads game data from a Stars! file.
func (r *Renderer) LoadFile(filename string) error {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	return r.LoadBytes(fileBytes)
}

// LoadReader loads game data from an io.Reader.
func (r *Renderer) LoadReader(reader io.Reader) error {
	fileBytes, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}
	return r.LoadBytes(fileBytes)
}

// LoadBytes loads game data from file bytes.
func (r *Renderer) LoadBytes(fileBytes []byte) error {
	fd := parser.FileData(fileBytes)

	header, err := fd.FileHeader()
	if err != nil {
		return fmt.Errorf("failed to parse file header: %w", err)
	}

	r.gameID = header.GameID
	r.turn = header.Turn
	r.year = header.Year()

	blockList, err := fd.BlockList()
	if err != nil {
		return fmt.Errorf("failed to parse blocks: %w", err)
	}

	for _, block := range blockList {
		r.processBlock(block)
	}

	r.assignPlayerColors()
	return nil
}

func (r *Renderer) processBlock(block blocks.Block) {
	switch b := block.(type) {
	case blocks.PlanetsBlock:
		for _, planet := range b.Planets {
			r.updateBounds(int(planet.X), int(planet.Y))
			r.planets[planet.ID] = &PlanetData{
				ID:    planet.ID,
				Name:  planet.Name,
				X:     int(planet.X),
				Y:     int(planet.Y),
				Owner: -1,
			}
		}

	case blocks.PartialPlanetBlock:
		if pd, ok := r.planets[b.PlanetNumber]; ok {
			pd.Owner = b.Owner
			pd.IsHomeworld = b.IsHomeworld
		}
		// Note: PartialPlanetBlock doesn't have X/Y coordinates
		// They come from PlanetsBlock

	case blocks.PlanetBlock:
		if pd, ok := r.planets[b.PlanetNumber]; ok {
			pd.Owner = b.Owner
			pd.IsHomeworld = b.IsHomeworld
		}
		// Note: PlanetBlock doesn't have X/Y coordinates
		// They come from PlanetsBlock

	case blocks.PartialFleetBlock:
		ships := 0
		for _, count := range b.ShipCount {
			ships += count
		}
		r.fleets[b.FleetNumber] = &FleetData{
			ID:     b.FleetNumber,
			Owner:  b.Owner,
			X:      b.X,
			Y:      b.Y,
			DeltaX: b.DeltaX,
			DeltaY: b.DeltaY,
			Mass:   int(b.Mass),
			Ships:  ships,
		}

	case blocks.FleetBlock:
		ships := 0
		for _, count := range b.ShipCount {
			ships += count
		}
		r.fleets[b.FleetNumber] = &FleetData{
			ID:     b.FleetNumber,
			Owner:  b.Owner,
			X:      b.X,
			Y:      b.Y,
			DeltaX: b.DeltaX,
			DeltaY: b.DeltaY,
			Mass:   int(b.Mass),
			Ships:  ships,
		}

	case blocks.ObjectBlock:
		if b.IsMinefield() {
			r.minefields[b.Number] = &MinefieldData{
				ID:        b.Number,
				Owner:     b.Owner,
				X:         b.X,
				Y:         b.Y,
				MineCount: int(b.MineCount),
				Type:      b.MinefieldType,
			}
		} else if b.IsWormhole() {
			r.wormholes[b.Number] = &WormholeData{
				ID:       b.Number,
				X:        b.X,
				Y:        b.Y,
				TargetID: b.TargetId,
			}
		}

	case blocks.PlayerBlock:
		r.players[b.PlayerNumber] = &PlayerData{
			Number: b.PlayerNumber,
			Name:   b.NameSingular,
		}
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
	{255, 3, 3, 255},      // Red
	{0, 66, 255, 255},     // Blue
	{28, 230, 185, 255},   // Teal
	{84, 0, 129, 255},     // Purple
	{255, 252, 1, 255},    // Yellow
	{254, 138, 14, 255},   // Orange
	{32, 192, 0, 255},     // Green
	{229, 91, 176, 255},   // Pink
	{149, 150, 151, 255},  // Gray
	{126, 191, 241, 255},  // Light blue
	{16, 98, 70, 255},     // Dark green
	{78, 42, 4, 255},      // Brown
	{255, 255, 255, 255},  // White
	{187, 115, 20, 255},   // Gold
	{200, 100, 100, 255},  // Light red
	{100, 100, 200, 255},  // Light purple
}

func (r *Renderer) assignPlayerColors() {
	for playerNum, player := range r.players {
		if playerNum >= 0 && playerNum < len(playerColors) {
			player.Color = playerColors[playerNum]
		} else {
			player.Color = color.RGBA{128, 128, 128, 255}
		}
	}
}

// GetPlayerColor returns the color for a player.
func (r *Renderer) GetPlayerColor(playerNum int) color.RGBA {
	if player, ok := r.players[playerNum]; ok {
		return player.Color
	}
	if playerNum >= 0 && playerNum < len(playerColors) {
		return playerColors[playerNum]
	}
	return color.RGBA{128, 128, 128, 255}
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

	// Draw minefields first (background)
	if opts.ShowMines {
		for _, mf := range r.minefields {
			px, py := transform(mf.X, mf.Y)
			radius := int(math.Sqrt(float64(mf.MineCount)) * scale / 10)
			if radius < 2 {
				radius = 2
			}
			col := r.GetPlayerColor(mf.Owner)
			col.A = 64 // Make semi-transparent
			drawCircleOutline(img, px, py, radius, col)
		}
	}

	// Draw wormholes
	if opts.ShowWormholes {
		purple := color.RGBA{128, 0, 128, 255}
		for _, wh := range r.wormholes {
			px, py := transform(wh.X, wh.Y)
			drawFilledCircle(img, px, py, 4, purple)
			// Draw connection to target if known
			if target, ok := r.wormholes[wh.TargetID]; ok {
				tx, ty := transform(target.X, target.Y)
				drawLine(img, px, py, tx, ty, color.RGBA{255, 0, 255, 128})
			}
		}
	}

	// Draw planets
	for _, planet := range r.planets {
		px, py := transform(planet.X, planet.Y)

		var col color.RGBA
		radius := 2

		if planet.Owner >= 0 {
			col = r.GetPlayerColor(planet.Owner)
			radius = 3
			if planet.IsHomeworld {
				// Draw homeworld highlight
				hwCol := col
				hwCol.A = 64
				drawFilledCircle(img, px, py, 12, hwCol)
			}
		} else {
			col = color.RGBA{128, 128, 128, 255}
		}

		drawFilledCircle(img, px, py, radius, col)
	}

	// Draw fleets
	if opts.ShowFleets {
		for _, fleet := range r.fleets {
			px, py := transform(fleet.X, fleet.Y)
			col := r.GetPlayerColor(fleet.Owner)
			col.A = 200

			// Draw direction triangle
			dx := float64(fleet.DeltaX - 127)
			dy := -float64(fleet.DeltaY - 127) // Flip Y for screen coords
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
	// Sort players by number
	var playerNums []int
	for num := range r.players {
		playerNums = append(playerNums, num)
	}
	sort.Ints(playerNums)

	y := 10
	for _, num := range playerNums {
		player := r.players[num]
		// Draw color box
		for dy := 0; dy < 10; dy++ {
			for dx := 0; dx < 10; dx++ {
				img.Set(5+dx, y+dy, player.Color)
			}
		}
		// Draw player name (simplified - just colored line)
		for dx := 0; dx < 50; dx++ {
			img.Set(20+dx, y+5, player.Color)
		}
		y += 14
	}
}

func (r *Renderer) drawYear(img *image.RGBA, opts *RenderOptions) {
	// Draw year in bottom left corner
	// Simple representation with colored pixels
	yearStr := fmt.Sprintf("%d", r.year)
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
func (r *Renderer) WritePNG(w io.Writer, opts *RenderOptions) error {
	img := r.Render(opts)

	if err := png.Encode(w, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
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
	return r.year
}

// Turn returns the game turn number.
func (r *Renderer) Turn() uint16 {
	return r.turn
}

// GameID returns the game ID.
func (r *Renderer) GameID() uint32 {
	return r.gameID
}

// PlanetCount returns the number of planets.
func (r *Renderer) PlanetCount() int {
	return len(r.planets)
}

// FleetCount returns the number of fleets.
func (r *Renderer) FleetCount() int {
	return len(r.fleets)
}

// Animator creates animated GIFs from multiple game files.
type Animator struct {
	renderers []*Renderer
	opts      *RenderOptions
}

// NewAnimator creates a new Animator.
func NewAnimator() *Animator {
	return &Animator{
		opts: DefaultOptions(),
	}
}

// SetOptions sets the rendering options for all frames.
func (a *Animator) SetOptions(opts *RenderOptions) {
	a.opts = opts
}

// AddFile adds a game file as a frame.
func (a *Animator) AddFile(filename string) error {
	r := New()
	if err := r.LoadFile(filename); err != nil {
		return err
	}
	a.renderers = append(a.renderers, r)
	return nil
}

// AddBytes adds game data from bytes as a frame.
func (a *Animator) AddBytes(data []byte) error {
	r := New()
	if err := r.LoadBytes(data); err != nil {
		return err
	}
	a.renderers = append(a.renderers, r)
	return nil
}

// AddReader adds game data from an io.Reader as a frame.
func (a *Animator) AddReader(reader io.Reader) error {
	r := New()
	if err := r.LoadReader(reader); err != nil {
		return err
	}
	a.renderers = append(a.renderers, r)
	return nil
}

// SortByYear sorts frames by game year.
func (a *Animator) SortByYear() {
	sort.Slice(a.renderers, func(i, j int) bool {
		return a.renderers[i].year < a.renderers[j].year
	})
}

// FrameCount returns the number of frames.
func (a *Animator) FrameCount() int {
	return len(a.renderers)
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
func (a *Animator) WriteGIF(w io.Writer, delayMs int) error {
	if len(a.renderers) == 0 {
		return fmt.Errorf("no frames to save")
	}

	// delay is in 100ths of a second
	delay := delayMs / 10

	anim := gif.GIF{
		LoopCount: 0, // Loop forever
	}

	for _, r := range a.renderers {
		img := r.Render(a.opts)
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
	// Determine direction
	var points [][2]int

	if dy < 0 && math.Abs(dy)/math.Abs(dx) >= 2.0 { // Up
		points = [][2]int{{cx - 4, cy + 3}, {cx, cy - 3}, {cx + 4, cy + 3}}
	} else if dy > 0 && math.Abs(dy)/math.Abs(dx) >= 2.0 { // Down
		points = [][2]int{{cx - 4, cy - 3}, {cx, cy + 3}, {cx + 4, cy - 3}}
	} else if dx > 0 && math.Abs(dx)/math.Abs(dy) >= 2.0 { // East
		points = [][2]int{{cx - 3, cy - 4}, {cx + 3, cy}, {cx - 3, cy + 4}}
	} else if dx < 0 && math.Abs(dx)/math.Abs(dy) >= 2.0 { // West
		points = [][2]int{{cx + 3, cy - 4}, {cx - 3, cy}, {cx + 3, cy + 4}}
	} else if dx > 0 && dy > 0 { // South-east
		points = [][2]int{{cx - 3, cy + 3}, {cx + 3, cy - 3}, {cx + 3, cy + 3}}
	} else if dx > 0 && dy < 0 { // North-east
		points = [][2]int{{cx - 3, cy - 3}, {cx + 3, cy - 3}, {cx + 3, cy + 3}}
	} else if dx < 0 && dy < 0 { // North-west
		points = [][2]int{{cx - 3, cy - 3}, {cx - 3, cy + 3}, {cx + 3, cy - 3}}
	} else if dx < 0 && dy > 0 { // South-west
		points = [][2]int{{cx - 3, cy - 3}, {cx - 3, cy + 3}, {cx + 3, cy + 3}}
	} else { // Stationary
		points = [][2]int{{cx - 3, cy - 4}, {cx + 3, cy}, {cx - 3, cy + 4}}
	}

	// Draw triangle outline
	if len(points) == 3 {
		drawLine(img, points[0][0], points[0][1], points[1][0], points[1][1], col)
		drawLine(img, points[1][0], points[1][1], points[2][0], points[2][1], col)
		drawLine(img, points[2][0], points[2][1], points[0][0], points[0][1], col)
	}
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
