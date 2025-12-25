package store

import "github.com/neper-stars/houston/blocks"

// DesignEntity represents a ship or starbase design.
type DesignEntity struct {
	meta EntityMeta

	// Identification
	DesignNumber int  // Slot 0-15
	Owner        int  // Player index
	IsStarbase   bool // True if starbase design

	// Design info
	Name string

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
		designBlock:  db,
	}
	entity.meta.AddSource(source)
	return entity
}

// DesignMap is a convenience type for looking up designs by slot.
type DesignMap map[int]*DesignEntity
