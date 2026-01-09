package store

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/data"
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

	// Warp byte upper bits - movement/status flags from dirLong union
	DirectionValid     bool // Fleet has a valid destination
	CompositionChanged bool // Fleet merged/split this turn
	Targeted           bool // Fleet is targeted by another fleet
	Skipped            bool // Fleet was skipped this turn

	// Full fleet data
	DamagedShipTypes uint16
	DamagedShipInfo  [16]uint16
	BattlePlan       int
	WaypointCount    int

	// Flags from Byte5 (see reversing_notes/fleet-block.md)
	Include      bool // fInclude - Include in reports/selection
	RepeatOrders bool // fRepOrders - Repeat waypoint orders when complete
	IsDead       bool // fDead - Fleet has been destroyed

	// Linked data
	CustomName    string
	HasCustomName bool
	PrimaryDesign *DesignEntity
	Waypoints     []*WaypointEntity // Associated waypoints in order

	// Raw blocks (preserved for re-encoding)
	fleetBlock *blocks.PartialFleetBlock
	nameBlock  *blocks.FleetNameBlock
}

// Meta returns the entity metadata.
func (f *FleetEntity) Meta() *EntityMeta {
	return &f.meta
}

// RawBlocks returns the original blocks including waypoints.
func (f *FleetEntity) RawBlocks() []blocks.Block {
	var result []blocks.Block
	if f.nameBlock != nil {
		result = append(result, *f.nameBlock)
	}
	if f.fleetBlock != nil {
		result = append(result, *f.fleetBlock)
	}
	// Include waypoint blocks
	for _, wp := range f.Waypoints {
		result = append(result, wp.RawBlocks()...)
	}
	return result
}

// AddWaypoint adds a waypoint to the fleet.
func (f *FleetEntity) AddWaypoint(wp *WaypointEntity) {
	f.Waypoints = append(f.Waypoints, wp)
}

// ClearWaypoints removes all waypoints from the fleet.
func (f *FleetEntity) ClearWaypoints() {
	f.Waypoints = nil
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
		// Population in Stars! files is stored in 100s of colonists
		population:         fb.Population * 100,
		fuel:               fb.Fuel,
		DeltaX:             fb.DeltaX,
		DeltaY:             fb.DeltaY,
		Warp:               fb.Warp,
		Mass:               fb.Mass,
		DirectionValid:     fb.DirectionValid,
		CompositionChanged: fb.CompositionChanged,
		Targeted:           fb.Targeted,
		Skipped:            fb.Skipped,
		DamagedShipTypes:   fb.DamagedShipTypes,
		DamagedShipInfo:    fb.DamagedShipInfo,
		BattlePlan:         fb.BattlePlan,
		WaypointCount:      fb.WaypointCount,
		Include:            fb.Include,
		RepeatOrders:       fb.RepeatOrders,
		IsDead:             fb.IsDead,
		fleetBlock:         fb,
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

// GetDesigns returns all designs present in this fleet, with their ship counts.
// Returns a map of design slot -> (design, count). Requires GameStore to look up designs.
func (f *FleetEntity) GetDesigns(gs *GameStore) map[int]struct {
	Design *DesignEntity
	Count  int
} {
	result := make(map[int]struct {
		Design *DesignEntity
		Count  int
	})

	for i := 0; i < 16; i++ {
		if (f.ShipTypes & (1 << i)) != 0 {
			count := f.ShipCounts[i]
			if count > 0 {
				design, ok := gs.Design(f.Owner, i)
				if ok {
					result[i] = struct {
						Design *DesignEntity
						Count  int
					}{design, count}
				}
			}
		}
	}
	return result
}

// GetTotalMass returns the total mass of the fleet including cargo.
// Requires GameStore to look up design masses from hulls.
func (f *FleetEntity) GetTotalMass(gs *GameStore) int64 {
	// If we have Mass from the block, use that
	if f.Mass > 0 {
		return f.Mass
	}

	// Otherwise calculate from designs
	var total int64
	designs := f.GetDesigns(gs)
	for _, info := range designs {
		if info.Design != nil {
			hull := info.Design.Hull()
			if hull != nil {
				total += int64(hull.Mass) * int64(info.Count)
			}
		}
	}

	// Add cargo mass (minerals + population)
	total += f.ironium + f.boranium + f.germanium
	// Population mass: 1 kT per 100 colonists
	total += f.population / 100

	return total
}

// GetScannerRanges returns the best scanner ranges from all ships in the fleet.
// This handles:
// - Equipped scanners on all ship designs
// - PRT intrinsic scanners (e.g., JOAT Scout/Frigate/Destroyer)
// - NAS LRT effects: 2× normal scanner range, no penetrating scanners
// Returns (normal, penetrating) ranges.
func (f *FleetEntity) GetScannerRanges(gs *GameStore) (int, int) {
	bestNormal := 0
	bestPen := 0

	// Get player info for PRT and LRT
	player, hasPlayer := gs.Player(f.Owner)
	var prt *data.PRT
	var hasNAS bool
	var nasLRT *data.LRT

	if hasPlayer {
		prt = data.GetPRT(player.PRT)
		nasLRT = data.GetLRTByCode("NAS")
		hasNAS = nasLRT != nil && player.HasLRT(nasLRT.Bitmask)
	}

	designs := f.GetDesigns(gs)
	for _, info := range designs {
		if info.Design == nil || info.Count <= 0 {
			continue
		}

		var normal, pen int

		// Check if this PRT has intrinsic scanner for this hull type (e.g., JOAT)
		if prt != nil && prt.HasFleetIntrinsicScannerForHull(info.Design.HullId) {
			// Use intrinsic scanner for this hull
			intrinsic := prt.FleetIntrinsicScannerRange(player.Tech.Electronics)
			normal = intrinsic.NormalRange
			pen = intrinsic.PenetratingRange
		} else {
			// Use equipped scanner
			normal, pen = info.Design.GetScannerRanges()
		}

		// Apply NAS effects: 2× normal range, no penetrating
		if hasNAS && nasLRT != nil {
			normal *= nasLRT.NormalScannerMultiplier
			pen = 0
		}

		if normal > bestNormal {
			bestNormal = normal
		}
		if pen > bestPen {
			bestPen = pen
		}
	}
	return bestNormal, bestPen
}

// GetTachyonCount returns the total number of Tachyon Detectors in the fleet.
func (f *FleetEntity) GetTachyonCount(gs *GameStore) int {
	total := 0
	designs := f.GetDesigns(gs)
	for _, info := range designs {
		if info.Design != nil {
			total += info.Design.GetTachyonCount() * info.Count
		}
	}
	return total
}

// GetCloakUnits returns the total cloaking units for the fleet.
// This is the sum of (ship_cloak_units × ship_count) for all ships in the fleet.
func (f *FleetEntity) GetCloakUnits(gs *GameStore) int {
	total := 0
	designs := f.GetDesigns(gs)
	for _, info := range designs {
		if info.Design != nil {
			cloakUnits := info.Design.GetCloakPercent()
			total += cloakUnits * info.Count
		}
	}
	return total
}
