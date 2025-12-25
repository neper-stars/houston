package store

import "github.com/neper-stars/houston/blocks"

// ResourceType identifies a resource/cargo type for the fluent builder.
type ResourceType int

const (
	Ironium ResourceType = iota
	Boranium
	Germanium
	Population
	Fuel
)

// Cargo represents resource amounts for named struct style cargo manipulation.
type Cargo struct {
	Ironium    int64
	Boranium   int64
	Germanium  int64
	Population int64
	Fuel       int64
}

// DataQuality represents the completeness/quality of entity data.
// Higher values indicate more complete data.
type DataQuality int

const (
	QualityUnknown    DataQuality = iota // Unknown quality
	QualityMinimal                       // Just ID and position
	QualityPartial                       // Some fields (enemy scan)
	QualityPickPocket                    // Has cargo info (Robber Baron)
	QualityFull                          // Complete data (owned entity)
)

// String returns a human-readable quality name.
func (q DataQuality) String() string {
	switch q {
	case QualityUnknown:
		return "Unknown"
	case QualityMinimal:
		return "Minimal"
	case QualityPartial:
		return "Partial"
	case QualityPickPocket:
		return "PickPocket"
	case QualityFull:
		return "Full"
	default:
		return "Invalid"
	}
}

// QualityFromFleetKind maps fleet KindByte to DataQuality.
func QualityFromFleetKind(kind byte) DataQuality {
	switch kind {
	case blocks.FleetKindPartial:
		return QualityPartial
	case blocks.FleetKindPickPocket:
		return QualityPickPocket
	case blocks.FleetKindFull:
		return QualityFull
	default:
		return QualityUnknown
	}
}

// ConflictResolver decides which data wins when merging overlapping entities.
type ConflictResolver interface {
	// ShouldReplace returns true if incoming should replace existing.
	ShouldReplace(existing, incoming Entity) bool
}

// DefaultResolver implements "best data wins" logic.
type DefaultResolver struct{}

// ShouldReplace implements ConflictResolver with these rules:
// 1. Higher quality wins
// 2. Same quality: prefer owner's perspective (if applicable)
// 3. Same quality + ownership: later turn wins
func (r *DefaultResolver) ShouldReplace(existing, incoming Entity) bool {
	existMeta := existing.Meta()
	incomeMeta := incoming.Meta()

	// Rule 1: Higher quality wins
	if incomeMeta.Quality > existMeta.Quality {
		return true
	}
	if existMeta.Quality > incomeMeta.Quality {
		return false
	}

	// Rule 2: Same quality - prefer owner's perspective
	if incomeMeta.BestSource != nil && existMeta.BestSource != nil {
		inOwned := incomeMeta.BestSource.PlayerIndex == incomeMeta.Key.Owner
		exOwned := existMeta.BestSource.PlayerIndex == existMeta.Key.Owner
		if inOwned && !exOwned {
			return true
		}
		if exOwned && !inOwned {
			return false
		}
	}

	// Rule 3: Later turn wins
	if incomeMeta.Turn > existMeta.Turn {
		return true
	}

	return false
}
