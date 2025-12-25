package store

import "github.com/neper-stars/houston/blocks"

// PlayerEntity represents a player in the game.
type PlayerEntity struct {
	meta EntityMeta

	// Core identification
	PlayerNumber int

	// Display info
	NamePlural   string
	NameSingular string
	Logo         int

	// Counts
	ShipDesignCount     int
	StarbaseDesignCount int
	PlanetCount         int
	FleetCount          int

	// Race info (if full data available)
	GrowthRate   int
	HasFullData  bool

	// Diplomatic relations (from file owner's perspective)
	PlayerRelations []byte

	// Raw block (preserved for re-encoding)
	playerBlock *blocks.PlayerBlock
}

// Meta returns the entity metadata.
func (p *PlayerEntity) Meta() *EntityMeta {
	return &p.meta
}

// RawBlocks returns the original blocks.
func (p *PlayerEntity) RawBlocks() []blocks.Block {
	if p.playerBlock != nil {
		return []blocks.Block{*p.playerBlock}
	}
	return nil
}

// SetDirty marks the entity as modified.
func (p *PlayerEntity) SetDirty() {
	p.meta.Dirty = true
}

// GetRelationTo returns the relation to another player.
// Returns: 0=Neutral, 1=Friend, 2=Enemy, -1=invalid
func (p *PlayerEntity) GetRelationTo(playerIndex int) int {
	if playerIndex < 0 || playerIndex >= len(p.PlayerRelations) {
		return 0 // Default to Neutral
	}
	return int(p.PlayerRelations[playerIndex])
}

// newPlayerEntityFromBlock creates a PlayerEntity from a PlayerBlock.
func newPlayerEntityFromBlock(pb *blocks.PlayerBlock, source *FileSource) *PlayerEntity {
	entity := &PlayerEntity{
		meta: EntityMeta{
			Key: EntityKey{
				Type:   EntityTypePlayer,
				Owner:  pb.PlayerNumber,
				Number: pb.PlayerNumber,
			},
			BestSource: source,
			Quality:    QualityFull, // Player blocks are always full quality
			Turn:       source.Turn,
		},
		PlayerNumber:        pb.PlayerNumber,
		NamePlural:          pb.NamePlural,
		NameSingular:        pb.NameSingular,
		Logo:                pb.Logo,
		ShipDesignCount:     pb.ShipDesignCount,
		StarbaseDesignCount: pb.StarbaseDesignCount,
		PlanetCount:         pb.Planets,
		FleetCount:          pb.Fleets,
		GrowthRate:          pb.GrowthRate,
		HasFullData:         pb.FullDataFlag,
		PlayerRelations:     pb.PlayerRelations,
		playerBlock:         pb,
	}
	entity.meta.AddSource(source)
	return entity
}
