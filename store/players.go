package store

import (
	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/data"
)

// TechLevels is an alias for data.TechRequirements for backward compatibility.
type TechLevels = data.TechRequirements

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

	// Player state (from header, bytes 0x08-0x0B)
	HomePlanetID int // idPlanetHome: ID of player's homeworld planet
	Rank         int // Player ranking position (1=1st, 2=2nd, etc.)
	//               // NOTE: Decompiled source names this "wScore" but it's actually Rank in the UI

	// Race info (if full data available)
	GrowthRate  int
	HasFullData bool
	Tech        TechLevels // Current tech levels
	PRT         int        // Primary Race Trait (0-9, see blocks.PRT* constants)
	LRT         uint16     // Lesser Race Traits bitmask (see blocks.LRT* constants)

	// Production settings (economy parameters)
	Production blocks.ProductionSettings

	// Habitability settings (environment preferences)
	Hab blocks.Habitability

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

// HasLRT returns true if the player has the specified Lesser Race Trait.
// The lrtBitmask should be one of the blocks.LRT* constants.
func (p *PlayerEntity) HasLRT(lrtBitmask uint16) bool {
	return (p.LRT & lrtBitmask) != 0
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
		HomePlanetID:        pb.HomePlanetID,
		Rank:                pb.Rank,
		GrowthRate:          pb.GrowthRate,
		HasFullData:         pb.FullDataFlag,
		Tech: TechLevels{
			Energy:       pb.Tech.Energy,
			Weapons:      pb.Tech.Weapons,
			Propulsion:   pb.Tech.Propulsion,
			Construction: pb.Tech.Construction,
			Electronics:  pb.Tech.Electronics,
			Biotech:      pb.Tech.Biotech,
		},
		PRT:             pb.PRT,
		LRT:             pb.LRT,
		Production:      pb.Production,
		Hab:             pb.Hab,
		PlayerRelations: pb.PlayerRelations,
		playerBlock:     pb,
	}
	entity.meta.AddSource(source)
	return entity
}
