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
	CargoFuel      = 4 // Fuel (mg) - only in transport orders, not in cargo hold
)

// Number of cargo types in transport orders (includes Fuel)
const TransportCargoTypeCount = 5

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

	// Transport task orders (when WaypointTask == WaypointTaskTransport)
	// Each cargo type has an action and value
	// [0]=Ironium, [1]=Boranium, [2]=Germanium, [3]=Colonists, [4]=Fuel
	TransportOrders [TransportCargoTypeCount]TransportOrder

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
	//   Byte 0: Value (amount in kT, percentage, or mg for fuel)
	//   Byte 1: Action type in high nibble (action << 4)
	// Order: Ironium, Boranium, Germanium, Colonists, Fuel
	// Note: Format is variable-length - not all cargo types may be present
	if wb.WaypointTask == WaypointTaskTransport && len(data) > 8 {
		for i := 0; i < TransportCargoTypeCount; i++ {
			offset := 8 + (i * 2)
			if offset+1 < len(data) {
				wb.TransportOrders[i].Value = int(data[offset] & 0xFF)
				wb.TransportOrders[i].Action = int(data[offset+1]&0xFF) >> 4
			}
		}
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
		size = 18 // 8 base + 10 transport orders (5 types × 2 bytes)
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
	// 5 cargo types × 2 bytes = 10 bytes (Ironium, Boranium, Germanium, Colonists, Fuel)
	if wb.WaypointTask == WaypointTaskTransport && size >= 18 {
		for i := 0; i < TransportCargoTypeCount; i++ {
			offset := 8 + (i * 2)
			data[offset] = byte(wb.TransportOrders[i].Value)
			data[offset+1] = byte((wb.TransportOrders[i].Action & 0x0F) << 4)
		}
	} else if len(wb.AdditionalBytes) > 0 {
		copy(data[8:], wb.AdditionalBytes)
	}

	return data
}

// UsesStargate returns true if this waypoint uses stargate travel
func (wb *WaypointBlock) UsesStargate() bool {
	return wb.Warp == WarpStargate
}

// IsLoadAllTransport returns true if this is a "Load All Available" transport task for Colonists
// This is a legacy compatibility method - prefer checking individual transport orders directly
func (wb *WaypointBlock) IsLoadAllTransport() bool {
	return wb.WaypointTask == WaypointTaskTransport && wb.TransportOrders[CargoColonists].Action == TransportTaskLoadAll
}

// IsUnloadAllTransport returns true if this is an "Unload All" transport task for Colonists
// This is a legacy compatibility method - prefer checking individual transport orders directly
func (wb *WaypointBlock) IsUnloadAllTransport() bool {
	return wb.WaypointTask == WaypointTaskTransport && wb.TransportOrders[CargoColonists].Action == TransportTaskUnloadAll
}

// GetTransportOrder returns the transport order for a specific cargo type
// cargoType should be CargoIronium, CargoBoranium, CargoGermanium, CargoColonists, or CargoFuel
func (wb *WaypointBlock) GetTransportOrder(cargoType int) TransportOrder {
	if cargoType >= 0 && cargoType < TransportCargoTypeCount {
		return wb.TransportOrders[cargoType]
	}
	return TransportOrder{}
}

// HasTransportOrders returns true if any cargo type has a non-zero transport action
func (wb *WaypointBlock) HasTransportOrders() bool {
	for i := 0; i < TransportCargoTypeCount; i++ {
		if wb.TransportOrders[i].Action != TransportTaskNoAction {
			return true
		}
	}
	return false
}

// CargoTypeName returns the human-readable name for a cargo type index
func CargoTypeName(cargoType int) string {
	names := []string{"Ironium", "Boranium", "Germanium", "Colonists", "Fuel"}
	if cargoType >= 0 && cargoType < len(names) {
		return names[cargoType]
	}
	return fmt.Sprintf("Unknown(%d)", cargoType)
}

// CargoTypeUnit returns the unit for a cargo type (kT for minerals/colonists, mg for fuel)
func CargoTypeUnit(cargoType int) string {
	if cargoType == CargoFuel {
		return "mg"
	}
	return "kT"
}

// TransportOrderDescription returns a human-readable description of the transport order for a cargo type
func (wb *WaypointBlock) TransportOrderDescription(cargoType int) string {
	if cargoType < 0 || cargoType >= TransportCargoTypeCount {
		return ""
	}
	order := wb.TransportOrders[cargoType]
	if order.Action == TransportTaskNoAction {
		return "No action"
	}
	cargoName := CargoTypeName(cargoType)
	unit := CargoTypeUnit(cargoType)
	actionName := TransportTaskName(order.Action)
	switch order.Action {
	case TransportTaskLoadAll, TransportTaskUnloadAll, TransportTaskDropAndLoad:
		return fmt.Sprintf("%s: %s", cargoName, actionName)
	case TransportTaskFillToPercent, TransportTaskWaitForPercent:
		return fmt.Sprintf("%s: %s %d%%", cargoName, actionName, order.Value)
	default:
		return fmt.Sprintf("%s: %s %d %s", cargoName, actionName, order.Value, unit)
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

	FleetNumber   int  // Fleet ID (9 bits)
	WaypointIndex int  // Waypoint index (uint16 LE from bytes 2-3, 0 for immediate move)
	X             int  // X coordinate
	Y             int  // Y coordinate
	Target        int  // Target ID (fleet/planet number)
	Warp          int  // Warp factor
	WaypointTask  int  // Task type
	ValidTask     bool // fValidTask flag (bit 12 of flags word) - indicates task is valid
	NoAutoTrack   bool // fNoAutoTrack flag (bit 13 of flags word)
	TargetType    int  // 1=planet, 2=fleet, 4=deep space, 8=wormhole
	SubTaskIndex  int  // Optional sub-task index

	// Transport task orders (when WaypointTask == WaypointTaskTransport)
	// Each cargo type has an action and value
	// [0]=Ironium, [1]=Boranium, [2]=Germanium, [3]=Colonists, [4]=Fuel
	TransportOrders [TransportCargoTypeCount]TransportOrder

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
	wctb.WaypointIndex = int(encoding.Read16(data, 2)) // uint16 LE
	wctb.X = int(encoding.Read16(data, 4))
	wctb.Y = int(encoding.Read16(data, 6))
	wctb.Target = int(data[8]&0xFF) + (int(data[9]&0x01) << 8)
	wctb.Warp = int(data[10]&0xFF) >> 4      // Upper nibble
	wctb.WaypointTask = int(data[10] & 0x0F) // Lower nibble

	// Byte 11 contains: bits 0-3=TargetType, bit 4=fValidTask, bit 5=fNoAutoTrack, bits 6-7=unused
	wctb.TargetType = int(data[11] & 0x0F)
	wctb.ValidTask = (data[11] & 0x10) != 0   // bit 4
	wctb.NoAutoTrack = (data[11] & 0x20) != 0 // bit 5

	if len(data) > 12 {
		wctb.SubTaskIndex = int(data[12] & 0xFF)
	}

	// Decode transport orders if this is a Transport task
	// Format: 2 bytes per cargo type starting at byte 12
	//   Byte 0: Value (amount in kT, percentage, or mg for fuel)
	//   Byte 1: Action type in high nibble (action << 4)
	// Order: Ironium, Boranium, Germanium, Colonists, Fuel
	// Note: Format is variable-length - not all cargo types may be present
	if wctb.WaypointTask == WaypointTaskTransport && len(data) > 12 {
		for i := 0; i < TransportCargoTypeCount; i++ {
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
		size = 22 // 12 base + 10 for transport orders (5 types × 2 bytes)
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

	// Bytes 2-3: Waypoint index (uint16 LE)
	encoding.Write16(data, 2, uint16(wctb.WaypointIndex))

	// Bytes 4-5: X coordinate
	encoding.Write16(data, 4, uint16(wctb.X))

	// Bytes 6-7: Y coordinate
	encoding.Write16(data, 6, uint16(wctb.Y))

	// Bytes 8-9: Target (9 bits)
	data[8] = byte(wctb.Target & 0xFF)
	data[9] = byte((wctb.Target >> 8) & 0x01)

	// Byte 10: Warp (upper nibble) | WaypointTask (lower nibble)
	data[10] = byte((wctb.Warp&0x0F)<<4) | byte(wctb.WaypointTask&0x0F)

	// Byte 11: flags (bit4=fValidTask, bit5=fNoAutoTrack) | TargetType (lower nibble)
	data[11] = byte(wctb.TargetType & 0x0F)
	if wctb.ValidTask {
		data[11] |= 0x10
	}
	if wctb.NoAutoTrack {
		data[11] |= 0x20
	}

	// Encode task-specific data
	switch {
	case wctb.WaypointTask == WaypointTaskTransport && size >= 22:
		// Transport orders: 2 bytes per cargo type (5 types)
		for i := 0; i < TransportCargoTypeCount; i++ {
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

// WaypointTaskTypeChangeBlock represents a lightweight waypoint task type change (Type 11)
// This block modifies only the task type at a specific waypoint without changing
// task-specific parameters (unlike Type 5/19 which include full task data).
//
// Format: 6 bytes
//
//	Bytes 0-1: FleetID (uint16 LE)
//	Bytes 2-3: WaypointIndex (uint16 LE, 0-based index into fleet's waypoint array)
//	Bytes 4-5: TaskType (uint16 LE, 0-9)
//
// This is useful when:
//   - Changing a waypoint from one simple task to another
//   - Clearing a task (setting to 0 = None)
//   - The task-specific parameters should remain unchanged
type WaypointTaskTypeChangeBlock struct {
	GenericBlock

	FleetID       int // Fleet identifier
	WaypointIndex int // Index into fleet's waypoint array (0-based)
	TaskType      int // New task type (0-9, see WaypointTask* constants)
}

// NewWaypointTaskTypeChangeBlock creates a WaypointTaskTypeChangeBlock from a GenericBlock
func NewWaypointTaskTypeChangeBlock(b GenericBlock) *WaypointTaskTypeChangeBlock {
	wttcb := &WaypointTaskTypeChangeBlock{GenericBlock: b}
	wttcb.decode()
	return wttcb
}

func (wttcb *WaypointTaskTypeChangeBlock) decode() {
	data := wttcb.Decrypted
	if len(data) < 6 {
		return
	}

	wttcb.FleetID = int(encoding.Read16(data, 0))
	wttcb.WaypointIndex = int(encoding.Read16(data, 2))
	wttcb.TaskType = int(encoding.Read16(data, 4))
}

// Encode returns the raw block data bytes (without the 2-byte block header).
func (wttcb *WaypointTaskTypeChangeBlock) Encode() []byte {
	data := make([]byte, 6)
	encoding.Write16(data, 0, uint16(wttcb.FleetID))
	encoding.Write16(data, 2, uint16(wttcb.WaypointIndex))
	encoding.Write16(data, 4, uint16(wttcb.TaskType))
	return data
}

// TaskTypeName returns a human-readable name for the task type
func (wttcb *WaypointTaskTypeChangeBlock) TaskTypeName() string {
	return WaypointTaskName(wttcb.TaskType)
}

// WaypointTaskName returns a human-readable name for a waypoint task type
func WaypointTaskName(task int) string {
	names := map[int]string{
		WaypointTaskNone:         "None",
		WaypointTaskTransport:    "Transport",
		WaypointTaskColonize:     "Colonize",
		WaypointTaskRemoteMining: "Remote Mining",
		WaypointTaskMergeFleet:   "Merge Fleet",
		WaypointTaskScrapFleet:   "Scrap Fleet",
		WaypointTaskLayMines:     "Lay Mines",
		WaypointTaskPatrol:       "Patrol",
		WaypointTaskRoute:        "Route",
		WaypointTaskTransfer:     "Transfer",
	}
	if name, ok := names[task]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", task)
}
