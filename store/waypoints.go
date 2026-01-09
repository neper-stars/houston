package store

import "github.com/neper-stars/houston/blocks"

// Waypoint task constants
const (
	WaypointTaskNone         = 0
	WaypointTaskTransport    = 1
	WaypointTaskColonize     = 2
	WaypointTaskRemoteMining = 3
	WaypointTaskMerge        = 4
	WaypointTaskScrapFleet   = 5
	WaypointTaskLayMines     = 6
	WaypointTaskPatrol       = 7
	WaypointTaskRoute        = 8
	WaypointTaskTransfer     = 9
)

// WaypointEntity represents a fleet waypoint.
type WaypointEntity struct {
	meta EntityMeta

	// Position
	X, Y           int
	PositionObject int // Object ID at position

	// Movement
	Warp int // Warp factor (0-15)

	// Task
	Task int // Task type (0-9)

	// Transport orders (when Task == WaypointTaskTransport)
	// [0]=Ironium, [1]=Boranium, [2]=Germanium, [3]=Colonists, [4]=Fuel
	TransportOrders [blocks.TransportCargoTypeCount]blocks.TransportOrder

	// Additional task data
	AdditionalBytes []byte

	// Raw block (preserved for re-encoding)
	waypointBlock *blocks.WaypointBlock
	taskBlock     *blocks.WaypointTaskBlock
}

// Meta returns the entity metadata.
func (w *WaypointEntity) Meta() *EntityMeta {
	return &w.meta
}

// RawBlocks returns the original blocks.
func (w *WaypointEntity) RawBlocks() []blocks.Block {
	var result []blocks.Block
	if w.waypointBlock != nil {
		result = append(result, *w.waypointBlock)
	}
	if w.taskBlock != nil {
		result = append(result, *w.taskBlock)
	}
	return result
}

// SetDirty marks the entity as modified.
func (w *WaypointEntity) SetDirty() {
	w.meta.Dirty = true
}

// TaskName returns a human-readable task name.
func (w *WaypointEntity) TaskName() string {
	switch w.Task {
	case WaypointTaskNone:
		return "None"
	case WaypointTaskTransport:
		return "Transport"
	case WaypointTaskColonize:
		return "Colonize"
	case WaypointTaskRemoteMining:
		return "Remote Mining"
	case WaypointTaskMerge:
		return "Merge"
	case WaypointTaskScrapFleet:
		return "Scrap Fleet"
	case WaypointTaskLayMines:
		return "Lay Mines"
	case WaypointTaskPatrol:
		return "Patrol"
	case WaypointTaskRoute:
		return "Route"
	case WaypointTaskTransfer:
		return "Transfer"
	default:
		return "Unknown"
	}
}

// newWaypointEntityFromBlock creates a WaypointEntity from a WaypointBlock.
func newWaypointEntityFromBlock(wb *blocks.WaypointBlock, fleetOwner, fleetNumber, waypointIndex int, source *FileSource) *WaypointEntity {
	entity := &WaypointEntity{
		meta: EntityMeta{
			Key: EntityKey{
				Type:   EntityTypeWaypoint,
				Owner:  fleetOwner,
				Number: fleetNumber*100 + waypointIndex, // Composite key
			},
			BestSource: source,
			Quality:    QualityFull,
			Turn:       source.Turn,
		},
		X:               wb.X,
		Y:               wb.Y,
		PositionObject:  wb.PositionObject,
		Warp:            wb.Warp,
		Task:            wb.WaypointTask,
		TransportOrders: wb.TransportOrders,
		AdditionalBytes: wb.AdditionalBytes,
		waypointBlock:   wb,
	}
	entity.meta.AddSource(source)
	return entity
}

// newWaypointEntityFromTaskBlock creates a WaypointEntity from a WaypointTaskBlock.
func newWaypointEntityFromTaskBlock(wtb *blocks.WaypointTaskBlock, fleetOwner, fleetNumber, waypointIndex int, source *FileSource) *WaypointEntity {
	entity := &WaypointEntity{
		meta: EntityMeta{
			Key: EntityKey{
				Type:   EntityTypeWaypoint,
				Owner:  fleetOwner,
				Number: fleetNumber*100 + waypointIndex,
			},
			BestSource: source,
			Quality:    QualityFull,
			Turn:       source.Turn,
		},
		X:               wtb.X,
		Y:               wtb.Y,
		PositionObject:  wtb.PositionObject,
		Warp:            wtb.Warp,
		Task:            wtb.WaypointTask,
		TransportOrders: wtb.TransportOrders,
		AdditionalBytes: wtb.AdditionalBytes,
		taskBlock:       wtb,
	}
	entity.meta.AddSource(source)
	return entity
}
