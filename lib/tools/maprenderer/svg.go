package maprenderer

import (
	"fmt"
	"image/color"
	"math"
	"strings"
)

// SVGBuilder provides a fluent interface for building SVG documents.
type SVGBuilder struct {
	width, height    int
	elements         []string
	defs             []string
	forRasterization bool // If true, skip markers and patterns during element creation
}

// NewSVGBuilder creates a new SVG builder with the given dimensions.
// Pre-allocates slices for typical map rendering (500+ elements).
func NewSVGBuilder(width, height int) *SVGBuilder {
	return &SVGBuilder{
		width:    width,
		height:   height,
		elements: make([]string, 0, 512), // Pre-allocate for typical map
		defs:     make([]string, 0, 16),
	}
}

// NewSVGBuilderForRasterization creates an SVG builder optimized for rasterization.
// It skips adding markers and patterns that aren't supported by rasterizers.
func NewSVGBuilderForRasterization(width, height int) *SVGBuilder {
	return &SVGBuilder{
		width:            width,
		height:           height,
		elements:         make([]string, 0, 512),
		defs:             make([]string, 0, 4),
		forRasterization: true,
	}
}

// AddDef adds a definition (pattern, gradient, marker, etc.) to the defs section.
// Skipped when forRasterization is true.
func (b *SVGBuilder) AddDef(def string) *SVGBuilder {
	if b.forRasterization {
		return b
	}
	b.defs = append(b.defs, def)
	return b
}

// AddMinefieldHatchPattern adds the standard minefield hatching pattern.
// Skipped when forRasterization is true (patterns not supported by rasterizers).
func (b *SVGBuilder) AddMinefieldHatchPattern() *SVGBuilder {
	if b.forRasterization {
		return b
	}
	b.defs = append(b.defs, `<pattern id="minefield-hatch" patternUnits="userSpaceOnUse" width="6" height="6" patternTransform="rotate(45)">
    <line x1="0" y1="0" x2="0" y2="6" stroke="currentColor" stroke-width="1" stroke-opacity="0.7"/>
  </pattern>`)
	return b
}

// AddArrowMarker adds an arrow marker for fleet speed lines.
// Skipped when forRasterization is true (markers not supported by rasterizers).
func (b *SVGBuilder) AddArrowMarker(id string, col color.RGBA) *SVGBuilder {
	if b.forRasterization {
		return b
	}
	b.defs = append(b.defs, fmt.Sprintf(`<marker id="%s" markerWidth="6" markerHeight="6" refX="3" refY="3" orient="auto" markerUnits="strokeWidth">
    <path d="M0,0 L0,6 L6,3 z" fill="rgba(%d,%d,%d,0.6)"/>
  </marker>`, id, col.R, col.G, col.B))
	return b
}

// Circle adds a circle element.
func (b *SVGBuilder) Circle(cx, cy, r float64, fill, stroke string, strokeWidth float64) *SVGBuilder {
	var s strings.Builder
	s.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="%.1f"`, cx, cy, r))
	if fill != "" {
		s.WriteString(fmt.Sprintf(` fill="%s"`, fill))
	}
	if stroke != "" {
		s.WriteString(fmt.Sprintf(` stroke="%s"`, stroke))
	}
	if strokeWidth > 0 {
		s.WriteString(fmt.Sprintf(` stroke-width="%.1f"`, strokeWidth))
	}
	s.WriteString("/>")
	b.elements = append(b.elements, s.String())
	return b
}

// CircleRGBA adds a circle with RGBA color.
func (b *SVGBuilder) CircleRGBA(cx, cy, r float64, col color.RGBA) *SVGBuilder {
	fill := fmt.Sprintf("rgb(%d,%d,%d)", col.R, col.G, col.B)
	return b.Circle(cx, cy, r, fill, "", 0)
}

// CircleOutline adds an unfilled circle outline.
func (b *SVGBuilder) CircleOutline(cx, cy, r float64, stroke string, strokeWidth float64) *SVGBuilder {
	return b.Circle(cx, cy, r, "none", stroke, strokeWidth)
}

// Minefield adds a minefield with semi-transparent fill and hatching.
// Hatching overlay is skipped when forRasterization is true.
func (b *SVGBuilder) Minefield(cx, cy, r float64, col color.RGBA) *SVGBuilder {
	// Semi-transparent fill (use integer alpha for rasterization compatibility)
	if b.forRasterization {
		const alphaFill = 38   // 0.15 * 255
		const alphaStroke = 102 // 0.4 * 255
		b.elements = append(b.elements, fmt.Sprintf(
			`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="rgba(%d,%d,%d,%d)" stroke="rgba(%d,%d,%d,%d)" stroke-width="1"/>`,
			cx, cy, r, col.R, col.G, col.B, alphaFill, col.R, col.G, col.B, alphaStroke))
	} else {
		b.elements = append(b.elements, fmt.Sprintf(
			`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="rgba(%d,%d,%d,0.15)" stroke="rgba(%d,%d,%d,0.4)" stroke-width="1"/>`,
			cx, cy, r, col.R, col.G, col.B, col.R, col.G, col.B))
		// Hatching overlay - only for non-rasterization output
		b.elements = append(b.elements, fmt.Sprintf(
			`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="url(#minefield-hatch)" style="color:rgba(%d,%d,%d,0.5)"/>`,
			cx, cy, r, col.R, col.G, col.B))
	}
	return b
}

// Rect adds a rectangle element.
func (b *SVGBuilder) Rect(x, y, width, height float64, fill string) *SVGBuilder {
	b.elements = append(b.elements, fmt.Sprintf(
		`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		x, y, width, height, fill))
	return b
}

// Text adds a text element.
func (b *SVGBuilder) Text(x, y float64, text string, col color.RGBA, fontSize int) *SVGBuilder {
	b.elements = append(b.elements, fmt.Sprintf(
		`<text x="%.1f" y="%.1f" fill="rgb(%d,%d,%d)" font-size="%d" font-family="monospace">%s</text>`,
		x, y, col.R, col.G, col.B, fontSize, text))
	return b
}

// Polygon adds a polygon element.
func (b *SVGBuilder) Polygon(points [][2]float64, fill, stroke string, strokeWidth float64) *SVGBuilder {
	var pointsStr strings.Builder
	for i, p := range points {
		if i > 0 {
			pointsStr.WriteString(" ")
		}
		pointsStr.WriteString(fmt.Sprintf("%.1f,%.1f", p[0], p[1]))
	}
	var s strings.Builder
	s.WriteString(fmt.Sprintf(`<polygon points="%s"`, pointsStr.String()))
	if fill != "" {
		s.WriteString(fmt.Sprintf(` fill="%s"`, fill))
	}
	if stroke != "" {
		s.WriteString(fmt.Sprintf(` stroke="%s"`, stroke))
	}
	if strokeWidth > 0 {
		s.WriteString(fmt.Sprintf(` stroke-width="%.1f"`, strokeWidth))
	}
	s.WriteString("/>")
	b.elements = append(b.elements, s.String())
	return b
}

// Diamond adds a diamond shape (for stationary fleets).
func (b *SVGBuilder) Diamond(cx, cy, size float64, col color.RGBA) *SVGBuilder {
	points := [][2]float64{
		{cx, cy - size},
		{cx + size, cy},
		{cx, cy + size},
		{cx - size, cy},
	}
	var stroke string
	if b.forRasterization {
		const alpha = 204 // 0.8 * 255
		stroke = fmt.Sprintf("rgba(%d,%d,%d,%d)", col.R, col.G, col.B, alpha)
	} else {
		stroke = fmt.Sprintf("rgba(%d,%d,%d,0.8)", col.R, col.G, col.B)
	}
	return b.Polygon(points, "none", stroke, 1)
}

// Triangle adds a triangle pointing in a direction (for moving fleets).
func (b *SVGBuilder) Triangle(cx, cy, size, angle float64, col color.RGBA) *SVGBuilder {
	tipX := cx + math.Cos(angle)*size
	tipY := cy + math.Sin(angle)*size
	base1X := cx + math.Cos(angle+math.Pi*2/3)*size
	base1Y := cy + math.Sin(angle+math.Pi*2/3)*size
	base2X := cx + math.Cos(angle-math.Pi*2/3)*size
	base2Y := cy + math.Sin(angle-math.Pi*2/3)*size

	points := [][2]float64{
		{tipX, tipY},
		{base1X, base1Y},
		{base2X, base2Y},
	}
	var stroke string
	if b.forRasterization {
		const alpha = 204 // 0.8 * 255
		stroke = fmt.Sprintf("rgba(%d,%d,%d,%d)", col.R, col.G, col.B, alpha)
	} else {
		stroke = fmt.Sprintf("rgba(%d,%d,%d,0.8)", col.R, col.G, col.B)
	}
	return b.Polygon(points, "none", stroke, 1)
}

// Line adds a line element.
func (b *SVGBuilder) Line(x1, y1, x2, y2 float64, stroke string, strokeWidth float64) *SVGBuilder {
	b.elements = append(b.elements, fmt.Sprintf(
		`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%.1f"/>`,
		x1, y1, x2, y2, stroke, strokeWidth))
	return b
}

// LineWithMarker adds a line with an arrow marker.
// Marker is skipped when forRasterization is true.
func (b *SVGBuilder) LineWithMarker(x1, y1, x2, y2 float64, stroke string, strokeWidth float64, markerID string) *SVGBuilder {
	if b.forRasterization {
		// Just draw the line without marker
		return b.Line(x1, y1, x2, y2, stroke, strokeWidth)
	}
	b.elements = append(b.elements, fmt.Sprintf(
		`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%.1f" marker-end="url(#%s)"/>`,
		x1, y1, x2, y2, stroke, strokeWidth, markerID))
	return b
}

// Path adds a path element with optional marker.
// Markers are skipped when forRasterization is true.
func (b *SVGBuilder) Path(d string, stroke string, strokeWidth float64, fill string, markerMid, markerEnd string) *SVGBuilder {
	var s strings.Builder
	s.WriteString(fmt.Sprintf(`<path d="%s"`, d))
	if fill != "" {
		s.WriteString(fmt.Sprintf(` fill="%s"`, fill))
	} else {
		s.WriteString(` fill="none"`)
	}
	if stroke != "" {
		// Convert decimal alpha to integer for rasterization compatibility
		if b.forRasterization {
			stroke = convertRGBAToIntAlpha(stroke)
		}
		s.WriteString(fmt.Sprintf(` stroke="%s"`, stroke))
	}
	if strokeWidth > 0 {
		s.WriteString(fmt.Sprintf(` stroke-width="%.1f"`, strokeWidth))
	}
	// Skip markers for rasterization
	if !b.forRasterization {
		if markerMid != "" {
			s.WriteString(fmt.Sprintf(` marker-mid="url(#%s)"`, markerMid))
		}
		if markerEnd != "" {
			s.WriteString(fmt.Sprintf(` marker-end="url(#%s)"`, markerEnd))
		}
	}
	s.WriteString("/>")
	b.elements = append(b.elements, s.String())
	return b
}

// FleetSpeedLine draws a fleet's projected path with arrow markers at each year.
func (b *SVGBuilder) FleetSpeedLine(startX, startY, dx, dy float64, years int, col color.RGBA, markerID string) *SVGBuilder {
	if years <= 0 || (math.Abs(dx) < 0.5 && math.Abs(dy) < 0.5) {
		return b // No movement or no years to show
	}

	// Build path with points at each year
	var pathD strings.Builder
	pathD.WriteString(fmt.Sprintf("M%.1f,%.1f", startX, startY))

	for i := 1; i <= years; i++ {
		x := startX + dx*float64(i)
		y := startY + dy*float64(i)
		pathD.WriteString(fmt.Sprintf(" L%.1f,%.1f", x, y))
	}

	stroke := fmt.Sprintf("rgba(%d,%d,%d,0.4)", col.R, col.G, col.B)
	return b.Path(pathD.String(), stroke, 1, "", markerID, markerID)
}

// WaypointPath draws a fleet's waypoint path with lines connecting each waypoint.
func (b *SVGBuilder) WaypointPath(points [][2]float64, col color.RGBA, markerID string) *SVGBuilder {
	if len(points) < 2 {
		return b // Need at least 2 points to draw a path
	}

	// Build path through all waypoints
	var pathD strings.Builder
	pathD.WriteString(fmt.Sprintf("M%.1f,%.1f", points[0][0], points[0][1]))

	for i := 1; i < len(points); i++ {
		pathD.WriteString(fmt.Sprintf(" L%.1f,%.1f", points[i][0], points[i][1]))
	}

	stroke := fmt.Sprintf("rgba(%d,%d,%d,0.5)", col.R, col.G, col.B)
	return b.Path(pathD.String(), stroke, 1.5, "", markerID, markerID)
}

// Starbase adds a starbase indicator (white circle + yellow satellite).
func (b *SVGBuilder) Starbase(cx, cy float64) *SVGBuilder {
	b.CircleOutline(cx, cy, 6, "white", 1)
	b.CircleRGBA(cx+5, cy-5, 2, color.RGBA{255, 255, 0, 255})
	return b
}

// Planet adds a planet circle.
func (b *SVGBuilder) Planet(cx, cy, radius float64, col color.RGBA, hasStarbase bool, name string, showName bool) *SVGBuilder {
	if hasStarbase {
		b.Starbase(cx, cy)
	}
	b.CircleRGBA(cx, cy, radius, col)
	if showName && name != "" {
		b.Text(cx+5, cy-5, name, col, 10)
	}
	return b
}

// Wormhole adds a wormhole indicator.
func (b *SVGBuilder) Wormhole(cx, cy float64) *SVGBuilder {
	return b.CircleOutline(cx, cy, 5, "purple", 1.5)
}

// ScannerCoverage adds a semi-transparent scanner coverage circle.
func (b *SVGBuilder) ScannerCoverage(cx, cy, radius float64, col color.RGBA) *SVGBuilder {
	// Draw a very faint filled circle for scanner coverage
	if b.forRasterization {
		const alphaFill = 20   // 0.08 * 255
		const alphaStroke = 51 // 0.2 * 255
		b.elements = append(b.elements, fmt.Sprintf(
			`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="rgba(%d,%d,%d,%d)" stroke="rgba(%d,%d,%d,%d)" stroke-width="0.5"/>`,
			cx, cy, radius, col.R, col.G, col.B, alphaFill, col.R, col.G, col.B, alphaStroke))
	} else {
		b.elements = append(b.elements, fmt.Sprintf(
			`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="rgba(%d,%d,%d,0.08)" stroke="rgba(%d,%d,%d,0.2)" stroke-width="0.5"/>`,
			cx, cy, radius, col.R, col.G, col.B, col.R, col.G, col.B))
	}
	return b
}

// LegendItem adds a legend entry.
func (b *SVGBuilder) LegendItem(x, y float64, name string, col color.RGBA) *SVGBuilder {
	b.Rect(x, y, 10, 10, fmt.Sprintf("rgb(%d,%d,%d)", col.R, col.G, col.B))
	b.Text(x+15, y+9, name, col, 10)
	return b
}

// String generates the final SVG document.
func (b *SVGBuilder) String() string {
	return b.buildSVG(false)
}

// StringForRasterization generates an SVG compatible with oksvg rasterization.
// It omits unsupported elements like <pattern>, <marker>, and their references.
// Note: For best performance, use NewSVGBuilderForRasterization() which skips
// markers/patterns at creation time, avoiding string replacements here.
func (b *SVGBuilder) StringForRasterization() string {
	return b.buildSVG(true)
}

// buildSVG generates the SVG document, optionally simplifying for rasterization.
func (b *SVGBuilder) buildSVG(forRasterization bool) string {
	// Pre-allocate builder with estimated capacity
	estimatedSize := 200 + len(b.elements)*100 + len(b.defs)*200
	var svg strings.Builder
	svg.Grow(estimatedSize)

	// Header (use absolute values for rect to support oksvg rasterization)
	svg.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">
<rect width="%d" height="%d" fill="black"/>
`, b.width, b.height, b.width, b.height, b.width, b.height))

	// Defs section - skip for rasterization (oksvg doesn't support pattern/marker)
	if !forRasterization && len(b.defs) > 0 {
		svg.WriteString("<defs>\n")
		for _, def := range b.defs {
			svg.WriteString("  ")
			svg.WriteString(def)
			svg.WriteString("\n")
		}
		svg.WriteString("</defs>\n")
	}

	// Elements - if builder was created with forRasterization, elements are already clean
	if b.forRasterization || !forRasterization {
		// Fast path: no string processing needed
		for _, elem := range b.elements {
			svg.WriteString(elem)
			svg.WriteString("\n")
		}
	} else {
		// Slow path: builder wasn't created for rasterization, need to clean elements
		for _, elem := range b.elements {
			// Remove marker references for rasterization
			elem = strings.ReplaceAll(elem, ` marker-mid="url(#arrow-0)"`, "")
			elem = strings.ReplaceAll(elem, ` marker-mid="url(#arrow-1)"`, "")
			elem = strings.ReplaceAll(elem, ` marker-end="url(#arrow-0)"`, "")
			elem = strings.ReplaceAll(elem, ` marker-end="url(#arrow-1)"`, "")
			// Remove pattern fill references
			if strings.Contains(elem, `fill="url(#minefield-hatch)"`) {
				continue // Skip hatching overlay circles
			}
			// Convert rgba with decimal alpha to integer alpha (0-255)
			elem = convertRGBAToIntAlpha(elem)
			svg.WriteString(elem)
			svg.WriteString("\n")
		}
	}

	svg.WriteString("</svg>")
	return svg.String()
}

// convertRGBAToIntAlpha converts rgba(r,g,b,0.x) to rgba(r,g,b,xxx) where xxx is 0-255.
func convertRGBAToIntAlpha(s string) string {
	// Pattern: rgba(R,G,B,0.X) -> rgba(R,G,B,XXX)
	result := s
	searchStart := 0
	for {
		idx := strings.Index(result[searchStart:], "rgba(")
		if idx == -1 {
			break
		}
		idx += searchStart

		endIdx := strings.Index(result[idx:], ")")
		if endIdx == -1 {
			break
		}
		endIdx += idx

		// Parse rgba(r,g,b,a)
		inner := result[idx+5 : endIdx]
		parts := strings.Split(inner, ",")
		if len(parts) == 4 {
			alpha := strings.TrimSpace(parts[3])
			// Check if alpha is a decimal (0.x format)
			if strings.HasPrefix(alpha, "0.") || alpha == "1" || alpha == "0" {
				var alphaFloat float64
				fmt.Sscanf(alpha, "%f", &alphaFloat)
				alphaInt := int(alphaFloat * 255)
				if alphaInt > 255 {
					alphaInt = 255
				}
				newRGBA := fmt.Sprintf("rgba(%s,%s,%s,%d)",
					strings.TrimSpace(parts[0]),
					strings.TrimSpace(parts[1]),
					strings.TrimSpace(parts[2]),
					alphaInt)
				result = result[:idx] + newRGBA + result[endIdx+1:]
				// Continue searching after the replacement
				searchStart = idx + len(newRGBA)
			} else {
				// Already integer format, continue to next occurrence
				searchStart = endIdx + 1
			}
		} else {
			searchStart = endIdx + 1
		}
	}
	return result
}
