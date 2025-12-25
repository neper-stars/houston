package store

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
)

// FleetEntity represents a fleet with full context and modification support.
type FleetEntity struct {
	meta EntityMeta

	// Core identification
	FleetNumber int
	Owner       int

	// Position
	X, Y             int
	PositionObjectId int

	// Ship composition
	ShipTypes  uint16  // 16-bit bitmask
	ShipCounts [16]int // Count per design slot

	// Cargo
	ironium    int64
	boranium   int64
	germanium  int64
	population int64
	fuel       int64

	// Movement (partial fleet data)
	DeltaX int
	DeltaY int
	Warp   int
	Mass   int64

	// Full fleet data
	DamagedShipTypes uint16
	DamagedShipInfo  [16]uint16
	BattlePlan       int
	WaypointCount    int
	RepeatOrders     bool

	// Linked data
	CustomName    string
	HasCustomName bool
	PrimaryDesign *DesignEntity

	// Raw blocks (preserved for re-encoding)
	fleetBlock *blocks.PartialFleetBlock
	nameBlock  *blocks.FleetNameBlock
}

// Meta returns the entity metadata.
func (f *FleetEntity) Meta() *EntityMeta {
	return &f.meta
}

// RawBlocks returns the original blocks.
func (f *FleetEntity) RawBlocks() []blocks.Block {
	var result []blocks.Block
	if f.nameBlock != nil {
		result = append(result, *f.nameBlock)
	}
	if f.fleetBlock != nil {
		result = append(result, *f.fleetBlock)
	}
	return result
}

// SetDirty marks the entity as modified.
func (f *FleetEntity) SetDirty() {
	f.meta.Dirty = true
}

// Name returns the fleet's display name.
// If the fleet has a custom name, returns that.
// Otherwise generates "{designName} #{fleetNumber+1}" or "Fleet #{fleetNumber+1}".
func (f *FleetEntity) Name() string {
	if f.HasCustomName {
		return f.CustomName
	}
	designName := ""
	if f.PrimaryDesign != nil {
		designName = f.PrimaryDesign.Name
	}
	if designName == "" {
		return fmt.Sprintf("Fleet #%d", f.FleetNumber+1)
	}
	return fmt.Sprintf("%s #%d", designName, f.FleetNumber+1)
}

// TotalShips returns the total number of ships in the fleet.
func (f *FleetEntity) TotalShips() int {
	total := 0
	for i := 0; i < 16; i++ {
		if (f.ShipTypes & (1 << i)) != 0 {
			total += f.ShipCounts[i]
		}
	}
	return total
}

// GetCargo returns the current cargo as a Cargo struct.
func (f *FleetEntity) GetCargo() Cargo {
	return Cargo{
		Ironium:    f.ironium,
		Boranium:   f.boranium,
		Germanium:  f.germanium,
		Population: f.population,
		Fuel:       f.fuel,
	}
}

// SetCargo sets all cargo values at once (named struct style).
func (f *FleetEntity) SetCargo(c Cargo) {
	f.ironium = c.Ironium
	f.boranium = c.Boranium
	f.germanium = c.Germanium
	f.population = c.Population
	f.fuel = c.Fuel
	f.SetDirty()
}

// Cargo returns a CargoBuilder for fluent cargo manipulation.
func (f *FleetEntity) Cargo() *CargoBuilder {
	return &CargoBuilder{
		fleet:      f,
		ironium:    f.ironium,
		boranium:   f.boranium,
		germanium:  f.germanium,
		population: f.population,
		fuel:       f.fuel,
	}
}

// CargoBuilder provides fluent cargo manipulation.
type CargoBuilder struct {
	fleet      *FleetEntity
	ironium    int64
	boranium   int64
	germanium  int64
	population int64
	fuel       int64
}

// Set sets a specific resource amount.
func (b *CargoBuilder) Set(r ResourceType, amount int64) *CargoBuilder {
	switch r {
	case Ironium:
		b.ironium = amount
	case Boranium:
		b.boranium = amount
	case Germanium:
		b.germanium = amount
	case Population:
		b.population = amount
	case Fuel:
		b.fuel = amount
	}
	return b
}

// Add adds to a specific resource.
func (b *CargoBuilder) Add(r ResourceType, amount int64) *CargoBuilder {
	switch r {
	case Ironium:
		b.ironium += amount
	case Boranium:
		b.boranium += amount
	case Germanium:
		b.germanium += amount
	case Population:
		b.population += amount
	case Fuel:
		b.fuel += amount
	}
	return b
}

// Apply writes the changes to the fleet and marks it dirty.
func (b *CargoBuilder) Apply() {
	b.fleet.ironium = b.ironium
	b.fleet.boranium = b.boranium
	b.fleet.germanium = b.germanium
	b.fleet.population = b.population
	b.fleet.fuel = b.fuel
	b.fleet.SetDirty()
}

// Get returns the current value for a resource.
func (b *CargoBuilder) Get(r ResourceType) int64 {
	switch r {
	case Ironium:
		return b.ironium
	case Boranium:
		return b.boranium
	case Germanium:
		return b.germanium
	case Population:
		return b.population
	case Fuel:
		return b.fuel
	default:
		return 0
	}
}

// newFleetEntityFromBlock creates a FleetEntity from a PartialFleetBlock.
func newFleetEntityFromBlock(fb *blocks.PartialFleetBlock, source *FileSource) *FleetEntity {
	entity := &FleetEntity{
		meta: EntityMeta{
			Key: EntityKey{
				Type:   EntityTypeFleet,
				Owner:  fb.Owner,
				Number: fb.FleetNumber,
			},
			BestSource: source,
			Quality:    QualityFromFleetKind(fb.KindByte),
			Turn:       source.Turn,
		},
		FleetNumber:      fb.FleetNumber,
		Owner:            fb.Owner,
		X:                fb.X,
		Y:                fb.Y,
		PositionObjectId: fb.PositionObjectId,
		ShipTypes:        fb.ShipTypes,
		ShipCounts:       fb.ShipCount,
		ironium:          fb.Ironium,
		boranium:         fb.Boranium,
		germanium:        fb.Germanium,
		population:       fb.Population,
		fuel:             fb.Fuel,
		DeltaX:           fb.DeltaX,
		DeltaY:           fb.DeltaY,
		Warp:             fb.Warp,
		Mass:             fb.Mass,
		DamagedShipTypes: fb.DamagedShipTypes,
		DamagedShipInfo:  fb.DamagedShipInfo,
		BattlePlan:       fb.BattlePlan,
		WaypointCount:    fb.WaypointCount,
		RepeatOrders:     fb.RepeatOrders,
		fleetBlock:       fb,
	}
	entity.meta.AddSource(source)
	return entity
}

// getPrimaryDesignSlot returns the first (lowest) design slot set in the ShipTypes bitmask.
func getPrimaryDesignSlot(shipTypes uint16) int {
	for i := 0; i < 16; i++ {
		if (shipTypes & (1 << i)) != 0 {
			return i
		}
	}
	return -1
}
