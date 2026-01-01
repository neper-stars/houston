package blocks

import (
	"fmt"

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

// Special warp values
const (
	WarpStargate = 11 // Use stargate for travel
)

// Transport action flags (byte 15 of WaypointTaskBlock for Transport tasks)
const (
	TransportActionLoadAll   = 0x10 // Load All Available
	TransportActionUnloadAll = 0x20 // Unload All
)

// Transport task action types (stored in high nibble of action byte)
const (
	TransportTaskNoAction       = 0 // No action for this cargo type
	TransportTaskLoadAll        = 1 // Load All Available
	TransportTaskUnloadAll      = 2 // Unload All
	TransportTaskLoadExactly    = 3 // Load Exactly N kT
	TransportTaskUnloadExactly  = 4 // Unload Exactly N kT
	TransportTaskFillToPercent  = 5 // Fill Up to N%
	TransportTaskWaitForPercent = 6 // Wait for N%
	TransportTaskDropAndLoad    = 7 // Drop and Load (unload all, then load)
	TransportTaskSetAmountTo    = 8 // Set Amount To N kT
)

// Cargo type indices
const (
	CargoIronium   = 0
	CargoBoranium  = 1
	CargoGermanium = 2
	CargoColonists = 3
)

// TransportOrder represents a transport task for a single cargo type
type TransportOrder struct {
	Action int // Transport action type (0-8)
	Value  int // Amount in kT or percentage depending on action
}

// TransportTaskName returns the human-readable name for a transport action
func TransportTaskName(action int) string {
	names := []string{
		"No Action",
		"Load All Available",
		"Unload All",
		"Load Exactly",
		"Unload Exactly",
		"Fill Up to %",
		"Wait for %",
		"Drop and Load",
		"Set Amount To",
	}
	if action >= 0 && action < len(names) {
		return names[action]
	}
	return "Unknown"
}

// WaypointBlock represents a waypoint in a fleet's route (Type 20)
type WaypointBlock struct {
	GenericBlock

	X                  int // X coordinate
	Y                  int // Y coordinate
	PositionObject     int // Object ID at position
	Warp               int // Warp factor (0-15)
	WaypointTask       int // Task type (0-9)
	PositionObjectType int // Type of object at position
	TransportAction    int // Transport action (0x10=Load All, 0x20=Unload All) from byte 15

	// Transport task orders (when WaypointTask == WaypointTaskTransport)
	// Each cargo type has an action and value
	TransportOrders [4]TransportOrder // [0]=Ironium, [1]=Boranium, [2]=Germanium, [3]=Colonists

	// Patrol range (when WaypointTask == WaypointTaskPatrol)
	PatrolRange int

	AdditionalBytes []byte // Variable-length additional data for tasks
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
	wb.Warp = int(data[6]&0xFF) >> 4      // Upper nibble
	wb.WaypointTask = int(data[6] & 0x0F) // Lower nibble
	wb.PositionObjectType = int(data[7] & 0xFF)

	// Additional bytes for task data
	if len(data) > 8 {
		wb.AdditionalBytes = make([]byte, len(data)-8)
		copy(wb.AdditionalBytes, data[8:])
	}

	// Decode transport orders if this is a Transport task
	// Format: 2 bytes per cargo type starting at byte 8
	//   Byte 0: Value (amount in kT or percentage)
	//   Byte 1: Action type in high nibble (action << 4)
	if wb.WaypointTask == WaypointTaskTransport && len(data) >= 16 {
		for i := 0; i < 4; i++ {
			offset := 8 + (i * 2)
			if offset+1 < len(data) {
				wb.TransportOrders[i].Value = int(data[offset] & 0xFF)
				wb.TransportOrders[i].Action = int(data[offset+1]&0xFF) >> 4
			}
		}
		// TransportAction flag is in byte 15 (last byte of transport data)
		wb.TransportAction = int(data[15] & 0xFF)
	}

	// Decode patrol range if this is a Patrol task
	// The range value is typically at byte 8 (first additional byte)
	if wb.WaypointTask == WaypointTaskPatrol && len(data) > 8 {
		wb.PatrolRange = int(data[8] & 0xFF)
	}
}

// Encode returns the raw block data bytes (without the 2-byte block header).
func (wb *WaypointBlock) Encode() []byte {
	// Calculate size based on task type
	size := 8
	if wb.WaypointTask == WaypointTaskTransport {
		size = 16 // 8 base + 8 transport orders
	} else if len(wb.AdditionalBytes) > 0 {
		size = 8 + len(wb.AdditionalBytes)
	}

	data := make([]byte, size)

	encoding.Write16(data, 0, uint16(wb.X))
	encoding.Write16(data, 2, uint16(wb.Y))
	encoding.Write16(data, 4, uint16(wb.PositionObject))
	data[6] = byte((wb.Warp&0x0F)<<4) | byte(wb.WaypointTask&0x0F)
	data[7] = byte(wb.PositionObjectType)

	// Encode transport orders if this is a Transport task
	if wb.WaypointTask == WaypointTaskTransport && size >= 16 {
		for i := 0; i < 4; i++ {
			offset := 8 + (i * 2)
			data[offset] = byte(wb.TransportOrders[i].Value)
			data[offset+1] = byte((wb.TransportOrders[i].Action & 0x0F) << 4)
		}
		// Set transport action flag in byte 15
		data[15] |= byte(wb.TransportAction & 0x0F)
	} else if len(wb.AdditionalBytes) > 0 {
		copy(data[8:], wb.AdditionalBytes)
	}

	return data
}

// UsesStargate returns true if this waypoint uses stargate travel
func (wb *WaypointBlock) UsesStargate() bool {
	return wb.Warp == WarpStargate
}

// IsLoadAllTransport returns true if this is a "Load All Available" transport task
func (wb *WaypointBlock) IsLoadAllTransport() bool {
	return wb.WaypointTask == WaypointTaskTransport && (wb.TransportAction&TransportActionLoadAll) != 0
}

// IsUnloadAllTransport returns true if this is an "Unload All" transport task
func (wb *WaypointBlock) IsUnloadAllTransport() bool {
	return wb.WaypointTask == WaypointTaskTransport && (wb.TransportAction&TransportActionUnloadAll) != 0
}

// GetTransportOrder returns the transport order for a specific cargo type
// cargoType should be CargoIronium, CargoBoranium, CargoGermanium, or CargoColonists
func (wb *WaypointBlock) GetTransportOrder(cargoType int) TransportOrder {
	if cargoType >= 0 && cargoType < 4 {
		return wb.TransportOrders[cargoType]
	}
	return TransportOrder{}
}

// HasTransportOrders returns true if any cargo type has a non-zero transport action
func (wb *WaypointBlock) HasTransportOrders() bool {
	for i := 0; i < 4; i++ {
		if wb.TransportOrders[i].Action != TransportTaskNoAction {
			return true
		}
	}
	return false
}

// TransportOrderDescription returns a human-readable description of the transport order for a cargo type
func (wb *WaypointBlock) TransportOrderDescription(cargoType int) string {
	if cargoType < 0 || cargoType >= 4 {
		return ""
	}
	order := wb.TransportOrders[cargoType]
	if order.Action == TransportTaskNoAction {
		return "No action"
	}
	cargoNames := []string{"Ironium", "Boranium", "Germanium", "Colonists"}
	actionName := TransportTaskName(order.Action)
	switch order.Action {
	case TransportTaskLoadAll, TransportTaskUnloadAll, TransportTaskDropAndLoad:
		return fmt.Sprintf("%s: %s", cargoNames[cargoType], actionName)
	case TransportTaskFillToPercent, TransportTaskWaitForPercent:
		return fmt.Sprintf("%s: %s %d%%", cargoNames[cargoType], actionName, order.Value)
	default:
		return fmt.Sprintf("%s: %s %d kT", cargoNames[cargoType], actionName, order.Value)
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

// Encode returns the raw block data bytes (without the 2-byte block header).
func (wtb *WaypointTaskBlock) Encode() []byte {
	return wtb.WaypointBlock.Encode()
}

// Patrol range constants
const (
	PatrolRangeAnyEnemy = 11 // Special value meaning "any enemy" (no range limit)
)

// PatrolRangeLY converts a patrol range value to light years
// Returns -1 for "any enemy" (infinite range)
func PatrolRangeLY(value int) int {
	if value == PatrolRangeAnyEnemy {
		return -1 // Infinite
	}
	return (value + 1) * 50
}

// PatrolRangeName returns a human-readable name for a patrol range value
func PatrolRangeName(value int) string {
	if value == PatrolRangeAnyEnemy {
		return "any enemy"
	}
	return fmt.Sprintf("within %d ly", (value+1)*50)
}

// WaypointChangeTaskBlock represents a waypoint task modification (Type 5)
type WaypointChangeTaskBlock struct {
	GenericBlock

	FleetNumber               int // Fleet ID (9 bits)
	WaypointNumber            int // Waypoint index (0 for immediate move)
	UnknownByte3              int
	X                         int // X coordinate
	Y                         int // Y coordinate
	Target                    int // Target ID (fleet/planet number)
	Warp                      int // Warp factor
	WaypointTask              int // Task type
	UnknownBitsWithTargetType int // Upper nibble of target type byte
	TargetType                int // 1=planet, 2=fleet, 4=deep space, 8=wormhole
	SubTaskIndex              int // Optional sub-task index

	// Transport task orders (when WaypointTask == WaypointTaskTransport)
	// Each cargo type has an action and value
	TransportOrders [4]TransportOrder // [0]=Ironium, [1]=Boranium, [2]=Germanium, [3]=Colonists

	// Patrol range (when WaypointTask == WaypointTaskPatrol)
	// 0=50ly, 1=100ly, 2=150ly, ..., 10=550ly, 11=any enemy
	// Use PatrolRangeLY() to convert to light years
	PatrolRange int
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
	wctb.Warp = int(data[10]&0xFF) >> 4      // Upper nibble
	wctb.WaypointTask = int(data[10] & 0x0F) // Lower nibble
	wctb.UnknownBitsWithTargetType = int(data[11]&0xFF) >> 4
	wctb.TargetType = int(data[11] & 0x0F)

	if len(data) > 12 {
		wctb.SubTaskIndex = int(data[12] & 0xFF)
	}

	// Decode transport orders if this is a Transport task
	// Format: 2 bytes per cargo type starting at byte 12
	//   Byte 0: Value (amount in kT or percentage)
	//   Byte 1: Action type in high nibble (action << 4)
	if wctb.WaypointTask == WaypointTaskTransport && len(data) >= 14 {
		for i := 0; i < 4; i++ {
			offset := 12 + (i * 2)
			if offset+1 < len(data) {
				wctb.TransportOrders[i].Value = int(data[offset] & 0xFF)
				wctb.TransportOrders[i].Action = int(data[offset+1]&0xFF) >> 4
			}
		}
	}

	// Decode patrol range if this is a Patrol task
	// Patrol range is at byte 14 when present (15-byte block)
	// 0=50ly, 1=100ly, ..., 10=550ly, 11=any enemy
	if wctb.WaypointTask == WaypointTaskPatrol && len(data) >= 15 {
		wctb.PatrolRange = int(data[14] & 0xFF)
	}
}

// Encode returns the raw block data bytes (without the 2-byte block header).
func (wctb *WaypointChangeTaskBlock) Encode() []byte {
	// Determine size based on task type
	var size int
	switch {
	case wctb.WaypointTask == WaypointTaskTransport:
		size = 20 // 12 base + 8 for transport orders
	case wctb.WaypointTask == WaypointTaskPatrol:
		size = 15 // 12 base + 1 sub-task + 2 for patrol range
	case wctb.SubTaskIndex > 0:
		size = 13 // 12 base + 1 sub-task index
	default:
		size = 12
	}

	data := make([]byte, size)

	// Bytes 0-1: Fleet number (9 bits)
	data[0] = byte(wctb.FleetNumber & 0xFF)
	data[1] = byte((wctb.FleetNumber >> 8) & 0x01)

	// Byte 2: Waypoint number
	data[2] = byte(wctb.WaypointNumber)

	// Byte 3: Unknown
	data[3] = byte(wctb.UnknownByte3)

	// Bytes 4-5: X coordinate
	encoding.Write16(data, 4, uint16(wctb.X))

	// Bytes 6-7: Y coordinate
	encoding.Write16(data, 6, uint16(wctb.Y))

	// Bytes 8-9: Target (9 bits)
	data[8] = byte(wctb.Target & 0xFF)
	data[9] = byte((wctb.Target >> 8) & 0x01)

	// Byte 10: Warp (upper nibble) | WaypointTask (lower nibble)
	data[10] = byte((wctb.Warp&0x0F)<<4) | byte(wctb.WaypointTask&0x0F)

	// Byte 11: Unknown bits (upper nibble) | TargetType (lower nibble)
	data[11] = byte((wctb.UnknownBitsWithTargetType&0x0F)<<4) | byte(wctb.TargetType&0x0F)

	// Encode task-specific data
	switch {
	case wctb.WaypointTask == WaypointTaskTransport && size >= 20:
		// Transport orders: 2 bytes per cargo type
		for i := 0; i < 4; i++ {
			offset := 12 + (i * 2)
			data[offset] = byte(wctb.TransportOrders[i].Value)
			data[offset+1] = byte((wctb.TransportOrders[i].Action & 0x0F) << 4)
		}
	case wctb.WaypointTask == WaypointTaskPatrol && size >= 15:
		data[12] = byte(wctb.SubTaskIndex)
		data[13] = 0
		data[14] = byte(wctb.PatrolRange)
	case size >= 13:
		data[12] = byte(wctb.SubTaskIndex)
	}

	return data
}

// UsesStargate returns true if this waypoint uses stargate travel
func (wctb *WaypointChangeTaskBlock) UsesStargate() bool {
	return wctb.Warp == WarpStargate
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

// Encode returns the raw block data bytes (without the 2-byte block header).
func (wab *WaypointAddBlock) Encode() []byte {
	return wab.WaypointChangeTaskBlock.Encode()
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

// Encode returns the raw block data bytes (without the 2-byte block header).
func (wdb *WaypointDeleteBlock) Encode() []byte {
	data := make([]byte, 3)
	data[0] = byte(wdb.FleetNumber & 0xFF)
	data[1] = byte((wdb.FleetNumber >> 8) & 0x01)
	data[2] = byte(wdb.WaypointNumber)
	return data
}
