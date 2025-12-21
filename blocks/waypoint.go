package blocks

import (
	"github.com/neper-stars/houston/encoding"
)

// Waypoint task types
const (
	WaypointTaskNone         = 0
	WaypointTaskTransport    = 1
	WaypointTaskColonize     = 2
	WaypointTaskRemoteMining = 3
	WaypointTaskMergeFleet   = 4
	WaypointTaskScrapFleet   = 5
	WaypointTaskLayMines     = 6
	WaypointTaskPatrol       = 7
	WaypointTaskRoute        = 8
	WaypointTaskTransfer     = 9
)

// Waypoint target types
const (
	WaypointTargetPlanet    = 1
	WaypointTargetFleet     = 2
	WaypointTargetDeepSpace = 4
	WaypointTargetWormhole  = 8 // Also mystery trader, minefield
)

// WaypointBlock represents a waypoint in a fleet's route (Type 20)
type WaypointBlock struct {
	GenericBlock

	X                  int // X coordinate
	Y                  int // Y coordinate
	PositionObject     int // Object ID at position
	Warp               int // Warp factor (0-15)
	WaypointTask       int // Task type (0-9)
	PositionObjectType int // Type of object at position
	AdditionalBytes    []byte // Variable-length additional data for tasks
}

// NewWaypointBlock creates a WaypointBlock from a GenericBlock
func NewWaypointBlock(b GenericBlock) *WaypointBlock {
	wb := &WaypointBlock{
		GenericBlock: b,
	}
	wb.decode()
	return wb
}

func (wb *WaypointBlock) decode() {
	data := wb.Decrypted
	if len(data) < 8 {
		return
	}

	wb.X = int(encoding.Read16(data, 0))
	wb.Y = int(encoding.Read16(data, 2))
	wb.PositionObject = int(encoding.Read16(data, 4))
	wb.Warp = int(data[6]&0xFF) >> 4        // Upper nibble
	wb.WaypointTask = int(data[6] & 0x0F)   // Lower nibble
	wb.PositionObjectType = int(data[7] & 0xFF)

	// Additional bytes for task data
	if len(data) > 8 {
		wb.AdditionalBytes = make([]byte, len(data)-8)
		copy(wb.AdditionalBytes, data[8:])
	}
}

// WaypointTaskBlock represents a waypoint with task information (Type 19)
type WaypointTaskBlock struct {
	WaypointBlock
}

// NewWaypointTaskBlock creates a WaypointTaskBlock from a GenericBlock
func NewWaypointTaskBlock(b GenericBlock) *WaypointTaskBlock {
	wtb := &WaypointTaskBlock{
		WaypointBlock: WaypointBlock{
			GenericBlock: b,
		},
	}
	wtb.decode()
	return wtb
}

// WaypointChangeTaskBlock represents a waypoint task modification (Type 5)
type WaypointChangeTaskBlock struct {
	GenericBlock

	FleetNumber             int // Fleet ID (9 bits)
	WaypointNumber          int // Waypoint index (0 for immediate move)
	UnknownByte3            int
	X                       int // X coordinate
	Y                       int // Y coordinate
	Target                  int // Target ID (fleet/planet number)
	Warp                    int // Warp factor
	WaypointTask            int // Task type
	UnknownBitsWithTargetType int // Upper nibble of target type byte
	TargetType              int // 1=planet, 2=fleet, 4=deep space, 8=wormhole
	SubTaskIndex            int // Optional sub-task index
}

// NewWaypointChangeTaskBlock creates a WaypointChangeTaskBlock from a GenericBlock
func NewWaypointChangeTaskBlock(b GenericBlock) *WaypointChangeTaskBlock {
	wctb := &WaypointChangeTaskBlock{
		GenericBlock: b,
	}
	wctb.decode()
	return wctb
}

func (wctb *WaypointChangeTaskBlock) decode() {
	data := wctb.Decrypted
	if len(data) < 12 {
		return
	}

	wctb.FleetNumber = int(data[0]&0xFF) + (int(data[1]&0x01) << 8)
	wctb.WaypointNumber = int(data[2] & 0xFF)
	wctb.UnknownByte3 = int(data[3] & 0xFF)
	wctb.X = int(encoding.Read16(data, 4))
	wctb.Y = int(encoding.Read16(data, 6))
	wctb.Target = int(data[8]&0xFF) + (int(data[9]&0x01) << 8)
	wctb.Warp = int(data[10]&0xFF) >> 4       // Upper nibble
	wctb.WaypointTask = int(data[10] & 0x0F)  // Lower nibble
	wctb.UnknownBitsWithTargetType = int(data[11]&0xFF) >> 4
	wctb.TargetType = int(data[11] & 0x0F)

	if len(data) > 12 {
		wctb.SubTaskIndex = int(data[12] & 0xFF)
	}
}

// WaypointAddBlock represents adding a waypoint to a fleet (Type 4)
type WaypointAddBlock struct {
	WaypointChangeTaskBlock
}

// NewWaypointAddBlock creates a WaypointAddBlock from a GenericBlock
func NewWaypointAddBlock(b GenericBlock) *WaypointAddBlock {
	wab := &WaypointAddBlock{
		WaypointChangeTaskBlock: WaypointChangeTaskBlock{
			GenericBlock: b,
		},
	}
	wab.decode()
	return wab
}

// WaypointDeleteBlock represents deleting a waypoint from a fleet (Type 3)
type WaypointDeleteBlock struct {
	GenericBlock

	FleetNumber    int // Fleet ID (9 bits)
	WaypointNumber int // Waypoint index to delete
}

// NewWaypointDeleteBlock creates a WaypointDeleteBlock from a GenericBlock
func NewWaypointDeleteBlock(b GenericBlock) *WaypointDeleteBlock {
	wdb := &WaypointDeleteBlock{
		GenericBlock: b,
	}
	wdb.decode()
	return wdb
}

func (wdb *WaypointDeleteBlock) decode() {
	data := wdb.Decrypted
	if len(data) < 3 {
		return
	}

	wdb.FleetNumber = int(data[0]&0xFF) + (int(data[1]&0x01) << 8)
	wdb.WaypointNumber = int(data[2] & 0xFF)
}
