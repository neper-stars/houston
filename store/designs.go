package store

import (
	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/data"
)

// DesignEntity represents a ship or starbase design.
type DesignEntity struct {
	meta EntityMeta

	// Identification
	DesignNumber int  // Slot 0-15
	Owner        int  // Player index
	IsStarbase   bool // True if starbase design

	// Design info
	Name   string
	HullId int // Hull type ID (see data.Hull* constants)

	// Raw block (preserved for re-encoding)
	designBlock *blocks.DesignBlock
}

// Meta returns the entity metadata.
func (d *DesignEntity) Meta() *EntityMeta {
	return &d.meta
}

// RawBlocks returns the original blocks.
func (d *DesignEntity) RawBlocks() []blocks.Block {
	if d.designBlock != nil {
		return []blocks.Block{*d.designBlock}
	}
	return nil
}

// SetDirty marks the entity as modified.
func (d *DesignEntity) SetDirty() {
	d.meta.Dirty = true
}

// newDesignEntityFromBlock creates a DesignEntity from a DesignBlock.
// The owner is taken from the source file's player index.
func newDesignEntityFromBlock(db *blocks.DesignBlock, source *FileSource) *DesignEntity {
	entityType := EntityTypeDesign
	if db.IsStarbase {
		entityType = EntityTypeStarbaseDesign
	}

	owner := source.PlayerIndex

	entity := &DesignEntity{
		meta: EntityMeta{
			Key: EntityKey{
				Type:   entityType,
				Owner:  owner,
				Number: db.DesignNumber,
			},
			BestSource: source,
			Quality:    QualityFull, // Designs are always full quality
			Turn:       source.Turn,
		},
		DesignNumber: db.DesignNumber,
		Owner:        owner,
		IsStarbase:   db.IsStarbase,
		Name:         db.Name,
		HullId:       db.HullId,
		designBlock:  db,
	}
	entity.meta.AddSource(source)
	return entity
}

// GetScannerRanges returns the best normal and penetrating scanner ranges
// from scanners equipped on this design.
//
// The Category field indicates the item type equipped (ItemCategoryScanner = 0x0002).
// ItemId is 0-indexed, so we add 1 to get the scanner constant (ScannerBat=1, etc.).
//
// Returns (0, 0) if no scanners are equipped.
func (d *DesignEntity) GetScannerRanges() (normal, penetrating int) {
	if d.designBlock == nil {
		return 0, 0
	}

	bestNormal := 0
	bestPen := 0

	for _, slot := range d.designBlock.Slots {
		if slot.Count == 0 || slot.Category != blocks.ItemCategoryScanner {
			continue
		}

		// ItemId is 0-indexed, scanner constants are 1-indexed
		scannerID := slot.ItemId + 1
		stats, found := data.ShipScannerStats[scannerID]

		if found {
			if stats.NormalRange > bestNormal {
				bestNormal = stats.NormalRange
			}
			if stats.PenetratingRange > bestPen {
				bestPen = stats.PenetratingRange
			}
		}
	}

	return bestNormal, bestPen
}

// GetNormalScannerRange returns the best normal scanner range from equipped scanners.
func (d *DesignEntity) GetNormalScannerRange() int {
	normal, _ := d.GetScannerRanges()
	return normal
}

// GetPenetratingScannerRange returns the best penetrating scanner range from equipped scanners.
func (d *DesignEntity) GetPenetratingScannerRange() int {
	_, pen := d.GetScannerRanges()
	return pen
}

// HasScanner returns true if this design has any scanner equipped.
func (d *DesignEntity) HasScanner() bool {
	normal, pen := d.GetScannerRanges()
	return normal > 0 || pen > 0
}

// DesignMap is a convenience type for looking up designs by slot.
type DesignMap map[int]*DesignEntity
