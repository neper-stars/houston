package blockdetail

import "github.com/neper-stars/houston/blocks"

// DesignInfo holds information about a ship design
type DesignInfo struct {
	Name   string
	HullID int
}

// FormatterContext holds contextual information for formatting blocks.
// This allows formatters to resolve cross-references (e.g., design names in fleets).
type FormatterContext struct {
	// FileOwner is the player index from the FileHeader
	FileOwner int

	// Designs maps design slot number to design info for each player
	// Key: (playerIndex << 4) | designSlot
	Designs map[int]DesignInfo
}

// global context instance
var globalContext *FormatterContext

// SetContext sets the formatter context for subsequent FormatDetailed calls
func SetContext(ctx *FormatterContext) {
	globalContext = ctx
}

// GetContext returns the current formatter context (may be nil)
func GetContext() *FormatterContext {
	return globalContext
}

// ClearContext clears the formatter context
func ClearContext() {
	globalContext = nil
}

// BuildContextFromBlocks builds a FormatterContext from a list of blocks
func BuildContextFromBlocks(blockList []blocks.Block) *FormatterContext {
	ctx := &FormatterContext{
		Designs:   make(map[int]DesignInfo),
		FileOwner: -1,
	}

	// First pass: find the file owner from FileHeader
	for _, block := range blockList {
		if fh, ok := block.(blocks.FileHeader); ok {
			ctx.FileOwner = fh.PlayerIndex()
			break
		}
	}

	// Second pass: collect designs
	// In M files, designs belong to the file owner unless marked as transferred
	for _, block := range blockList {
		if b, ok := block.(blocks.DesignBlock); ok {
			// Designs in M files are for the file owner
			owner := ctx.FileOwner
			if owner < 0 {
				owner = 0 // Fallback if no FileHeader found
			}

			// Key: (owner << 4) | designSlot
			// This allows 16 players with 16 designs each
			key := (owner << 4) | b.DesignNumber
			ctx.Designs[key] = DesignInfo{
				Name:   b.Name,
				HullID: b.HullId,
			}
		}
	}

	return ctx
}

// GetDesignName returns the design name for a given player and design slot
func (ctx *FormatterContext) GetDesignName(owner, designSlot int) string {
	if ctx == nil {
		return ""
	}
	key := (owner << 4) | designSlot
	if info, ok := ctx.Designs[key]; ok {
		return info.Name
	}
	return ""
}
