package store

import "github.com/neper-stars/houston/blocks"

// PlanetEntity represents a planet with full context.
type PlanetEntity struct {
	meta EntityMeta

	// Core identification
	PlanetNumber int
	Owner        int  // -1 = unowned
	IsHomeworld  bool

	// Name (from PlanetsBlock)
	Name string
	X, Y int

	// Status flags
	HasStarbase      bool
	HasArtifact      bool
	IsTerraformed    bool
	HasInstallations bool

	// Environment data
	IroniumConc   int
	BoraniumConc  int
	GermaniumConc int
	Gravity       int
	Temperature   int
	Radiation     int

	// Original environment (if terraformed)
	OrigGravity     int
	OrigTemperature int
	OrigRadiation   int

	// Surface minerals
	Ironium   int64
	Boranium  int64
	Germanium int64

	// Installations
	Mines      int
	Factories  int
	Defenses   int
	ExcessPop  int
	HasScanner bool
	Population int64

	// Starbase design slot (if HasStarbase)
	StarbaseDesign int

	// Route (if has route)
	RouteTarget int

	// Raw block (preserved for re-encoding)
	planetBlock *blocks.PartialPlanetBlock
}

// Meta returns the entity metadata.
func (p *PlanetEntity) Meta() *EntityMeta {
	return &p.meta
}

// RawBlocks returns the original blocks.
func (p *PlanetEntity) RawBlocks() []blocks.Block {
	if p.planetBlock != nil {
		return []blocks.Block{*p.planetBlock}
	}
	return nil
}

// SetDirty marks the entity as modified.
func (p *PlanetEntity) SetDirty() {
	p.meta.Dirty = true
}

// IsOwned returns true if the planet has an owner.
func (p *PlanetEntity) IsOwned() bool {
	return p.Owner >= 0
}

// GetMinerals returns the current surface minerals as a Cargo struct.
func (p *PlanetEntity) GetMinerals() Cargo {
	return Cargo{
		Ironium:   p.Ironium,
		Boranium:  p.Boranium,
		Germanium: p.Germanium,
	}
}

// SetMinerals sets surface minerals (marks dirty).
func (p *PlanetEntity) SetMinerals(c Cargo) {
	p.Ironium = c.Ironium
	p.Boranium = c.Boranium
	p.Germanium = c.Germanium
	p.SetDirty()
}

// qualityFromPlanetBlock determines data quality from planet flags.
func qualityFromPlanetBlock(pb *blocks.PartialPlanetBlock) DataQuality {
	// Full planet data includes installations and population
	if pb.HasInstallations {
		return QualityFull
	}
	// Has environment info but not full data
	if pb.HasEnvironmentInfo {
		return QualityPartial
	}
	// Minimal - just position and owner
	return QualityMinimal
}

// newPlanetEntityFromBlock creates a PlanetEntity from a PartialPlanetBlock.
func newPlanetEntityFromBlock(pb *blocks.PartialPlanetBlock, source *FileSource) *PlanetEntity {
	entity := &PlanetEntity{
		meta: EntityMeta{
			Key: EntityKey{
				Type:   EntityTypePlanet,
				Owner:  pb.Owner,
				Number: pb.PlanetNumber,
			},
			BestSource: source,
			Quality:    qualityFromPlanetBlock(pb),
			Turn:       source.Turn,
		},
		PlanetNumber:    pb.PlanetNumber,
		Owner:           pb.Owner,
		IsHomeworld:     pb.IsHomeworld,
		HasStarbase:     pb.HasStarbase,
		HasArtifact:     pb.HasArtifact,
		IsTerraformed:   pb.IsTerraformed,
		HasInstallations: pb.HasInstallations,
		IroniumConc:     pb.IroniumConc,
		BoraniumConc:    pb.BoraniumConc,
		GermaniumConc:   pb.GermaniumConc,
		Gravity:         pb.Gravity,
		Temperature:     pb.Temperature,
		Radiation:       pb.Radiation,
		OrigGravity:     pb.OrigGravity,
		OrigTemperature: pb.OrigTemperature,
		OrigRadiation:   pb.OrigRadiation,
		Ironium:         pb.Ironium,
		Boranium:        pb.Boranium,
		Germanium:       pb.Germanium,
		Mines:      pb.Mines,
		Factories:  pb.Factories,
		Defenses:   pb.Defenses,
		ExcessPop:  pb.ExcessPop,
		HasScanner: pb.HasScanner,
		Population: pb.Population,
		StarbaseDesign:  pb.StarbaseDesign,
		RouteTarget:     pb.RouteTarget,
		planetBlock:     pb,
	}
	entity.meta.AddSource(source)
	return entity
}
