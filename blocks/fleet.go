package blocks

import (
	"github.com/neper-stars/houston/encoding"
)

// Fleet kind constants
const (
	FleetKindPartial    = 3 // Partial fleet data
	FleetKindPickPocket = 4 // Robber baron, has minerals
	FleetKindFull       = 7 // Complete fleet data
)

// PartialFleetBlock represents fleet data with variable fields (Type 17)
type PartialFleetBlock struct {
	GenericBlock

	// Fleet identification
	FleetNumber int // 0-511 (9 bits)
	Owner       int // 0-15

	// Control bytes (preserved for encoding)
	Byte2    byte
	Byte3    byte
	KindByte byte // 3=partial, 4=pick-pocket, 7=full
	Byte5    byte

	// Position
	PositionObjectId int // Referenced object
	X                int // X coordinate
	Y                int // Y coordinate

	// Ship composition
	ShipTypes          uint16   // 16-bit bitmask, one bit per design
	ShipCount          [16]int  // Ship counts per design slot
	ShipCountTwoBytes  bool     // True if ship counts are 2 bytes each

	// Resources/Cargo (if full or pick-pocket)
	Ironium    int64
	Boranium   int64
	Germanium  int64
	Population int64
	Fuel       int64

	// Partial fleet movement data (if kindByte != FleetKindFull)
	DeltaX              int
	DeltaY              int
	Warp                int  // 0-15
	UnknownBitsWithWarp int  // Upper 4 bits of warp byte
	Mass                int64

	// Full fleet data (if kindByte == FleetKindFull)
	DamagedShipTypes uint16     // 16-bit bitmask
	DamagedShipInfo  [16]uint16 // Damage info per type (2 bytes each)
	BattlePlan       int
	WaypointCount    int
}

// NewPartialFleetBlock creates a PartialFleetBlock from a GenericBlock
func NewPartialFleetBlock(b GenericBlock) *PartialFleetBlock {
	fb := &PartialFleetBlock{
		GenericBlock: b,
	}
	fb.decode()
	return fb
}

// GetFleetIdAndOwner returns a composite ID combining fleet number and owner
func (fb *PartialFleetBlock) GetFleetIdAndOwner() int {
	return fb.FleetNumber | (fb.Owner << 9)
}

// IsFullFleet returns true if this is a full fleet block
func (fb *PartialFleetBlock) IsFullFleet() bool {
	return fb.KindByte == FleetKindFull
}

// HasCargo returns true if this fleet has cargo data
func (fb *PartialFleetBlock) HasCargo() bool {
	return fb.KindByte == FleetKindFull || fb.KindByte == FleetKindPickPocket
}

// TotalShips returns the total number of ships in the fleet
func (fb *PartialFleetBlock) TotalShips() int {
	total := 0
	for i := 0; i < 16; i++ {
		if (fb.ShipTypes & (1 << i)) != 0 {
			total += fb.ShipCount[i]
		}
	}
	return total
}

func (fb *PartialFleetBlock) decode() {
	data := fb.Decrypted
	if len(data) < 14 {
		return
	}

	// Bytes 0-1: Fleet number and owner
	fb.FleetNumber = int(data[0]&0xFF) + (int(data[1]&0x01) << 8)
	fb.Owner = int(data[1]) >> 1

	// Bytes 2-5: Control bytes
	fb.Byte2 = data[2]
	fb.Byte3 = data[3]
	fb.KindByte = data[4]
	fb.Byte5 = data[5]

	// Determine if ship counts are 2 bytes (bit 3 of byte5 clear)
	fb.ShipCountTwoBytes = (fb.Byte5 & 0x08) == 0

	// Bytes 6-13: Position and ship types
	fb.PositionObjectId = int(encoding.Read16(data, 6))
	fb.X = int(encoding.Read16(data, 8))
	fb.Y = int(encoding.Read16(data, 10))
	fb.ShipTypes = encoding.Read16(data, 12)

	index := 14

	// Variable-length ship counts
	for bit := 0; bit < 16; bit++ {
		if (fb.ShipTypes & (1 << bit)) != 0 {
			if fb.ShipCountTwoBytes {
				if index+2 <= len(data) {
					fb.ShipCount[bit] = int(encoding.Read16(data, index))
					index += 2
				}
			} else {
				if index < len(data) {
					fb.ShipCount[bit] = int(data[index] & 0xFF)
					index++
				}
			}
		}
	}

	// Resources section (if FULL or PICK_POCKET kind)
	if fb.HasCargo() && index+2 <= len(data) {
		contentsLengths := encoding.Read16(data, index)
		index += 2

		// Extract length encodings (2 bits each)
		ironLen := encoding.VarLenByteCount(encoding.ExtractVarLenField16(contentsLengths, 0))
		boraLen := encoding.VarLenByteCount(encoding.ExtractVarLenField16(contentsLengths, 2))
		germLen := encoding.VarLenByteCount(encoding.ExtractVarLenField16(contentsLengths, 4))
		popLen := encoding.VarLenByteCount(encoding.ExtractVarLenField16(contentsLengths, 6))
		fuelLen := encoding.VarLenByteCount(encoding.ExtractVarLenField16(contentsLengths, 8))

		if index+ironLen <= len(data) {
			fb.Ironium, _ = encoding.ReadVarLen(data, index, ironLen)
			index += ironLen
		}
		if index+boraLen <= len(data) {
			fb.Boranium, _ = encoding.ReadVarLen(data, index, boraLen)
			index += boraLen
		}
		if index+germLen <= len(data) {
			fb.Germanium, _ = encoding.ReadVarLen(data, index, germLen)
			index += germLen
		}
		if index+popLen <= len(data) {
			fb.Population, _ = encoding.ReadVarLen(data, index, popLen)
			index += popLen
		}
		if index+fuelLen <= len(data) {
			fb.Fuel, _ = encoding.ReadVarLen(data, index, fuelLen)
			index += fuelLen
		}
	}

	// Full Fleet vs Partial Fleet specific data
	if fb.KindByte == FleetKindFull {
		// Damaged ship data
		if index+2 <= len(data) {
			fb.DamagedShipTypes = encoding.Read16(data, index)
			index += 2

			for bit := 0; bit < 16; bit++ {
				if (fb.DamagedShipTypes & (1 << bit)) != 0 {
					if index+2 <= len(data) {
						fb.DamagedShipInfo[bit] = encoding.Read16(data, index)
						index += 2
					}
				}
			}
		}

		if index+2 <= len(data) {
			fb.BattlePlan = int(data[index] & 0xFF)
			fb.WaypointCount = int(data[index+1] & 0xFF)
			index += 2
		}
	} else {
		// Partial fleet movement data
		if index+8 <= len(data) {
			fb.DeltaX = int(int8(data[index])) // Signed byte
			fb.DeltaY = int(int8(data[index+1]))
			fb.Warp = int(data[index+2] & 0x0F)
			fb.UnknownBitsWithWarp = int(data[index+2] & 0xF0)
			// index+3 is padding (should be 0)
			fb.Mass = int64(encoding.Read32(data, index+4))
			index += 8
		}
	}
}

// FleetBlock represents a full fleet with all information (Type 16)
type FleetBlock struct {
	PartialFleetBlock
}

// NewFleetBlock creates a FleetBlock from a GenericBlock
func NewFleetBlock(b GenericBlock) *FleetBlock {
	fb := &FleetBlock{
		PartialFleetBlock: PartialFleetBlock{
			GenericBlock: b,
		},
	}
	fb.decode()
	return fb
}

// FleetSplitBlock represents a fleet split operation (Type 24)
type FleetSplitBlock struct {
	GenericBlock

	FleetNumber int // The fleet being split (9 bits)
}

// NewFleetSplitBlock creates a FleetSplitBlock from a GenericBlock
func NewFleetSplitBlock(b GenericBlock) *FleetSplitBlock {
	fsb := &FleetSplitBlock{
		GenericBlock: b,
	}
	fsb.decode()
	return fsb
}

func (fsb *FleetSplitBlock) decode() {
	data := fsb.Decrypted
	if len(data) < 2 {
		return
	}

	fsb.FleetNumber = int(data[0]&0xFF) + (int(data[1]&0x01) << 8)
}

// FleetsMergeBlock represents a fleet merge operation (Type 37)
type FleetsMergeBlock struct {
	GenericBlock

	FleetNumber    int   // The target fleet (9 bits)
	FleetsToMerge  []int // List of fleet numbers to merge into target
}

// NewFleetsMergeBlock creates a FleetsMergeBlock from a GenericBlock
func NewFleetsMergeBlock(b GenericBlock) *FleetsMergeBlock {
	fmb := &FleetsMergeBlock{
		GenericBlock: b,
	}
	fmb.decode()
	return fmb
}

func (fmb *FleetsMergeBlock) decode() {
	data := fmb.Decrypted
	if len(data) < 2 {
		return
	}

	fmb.FleetNumber = int(data[0]&0xFF) + (int(data[1]&0x01) << 8)

	// Remaining bytes are fleet numbers to merge (2 bytes each)
	for i := 2; i+1 < len(data); i += 2 {
		mergeFleet := int(data[i]&0xFF) + (int(data[i+1]&0x01) << 8)
		fmb.FleetsToMerge = append(fmb.FleetsToMerge, mergeFleet)
	}
}

// FleetNameBlock represents a fleet name record (Type 21)
// Structure not fully documented - preserves raw data for analysis
type FleetNameBlock struct {
	GenericBlock
}

// NewFleetNameBlock creates a FleetNameBlock from a GenericBlock
func NewFleetNameBlock(b GenericBlock) *FleetNameBlock {
	return &FleetNameBlock{GenericBlock: b}
}

// MoveShipsBlock represents ship movement between fleets (Type 23)
// Used when transferring ships from one fleet to another (merge operation)
//
// TransferInfo structure (partial decoding):
//   - Byte 0: Flags/control byte
//   - Byte 1: Ship types bitmask (low byte) - bit N set means design slot N
//   - Byte 2: Ship types bitmask (high byte) or padding
//   - Byte 3: Ship count for first design in bitmask
//   - Byte 4+: Additional ship counts for other designs (if multiple)
//
// Note: Full TransferInfo structure needs more test samples to decode completely.
type MoveShipsBlock struct {
	GenericBlock

	DestFleetNumber   int    // Destination fleet number (0-indexed)
	SourceFleetNumber int    // Source fleet number (0-indexed)
	ShipCount         int    // Number of ships transferred (from TransferInfo byte 3)
	TransferInfo      []byte // Raw transfer data (preserved for analysis)
}

// NewMoveShipsBlock creates a MoveShipsBlock from a GenericBlock
func NewMoveShipsBlock(b GenericBlock) *MoveShipsBlock {
	block := &MoveShipsBlock{GenericBlock: b}
	block.decode()
	return block
}

func (b *MoveShipsBlock) decode() {
	data := b.Decrypted
	if len(data) < 4 {
		return
	}

	// Bytes 0-1: Destination fleet number (16-bit)
	b.DestFleetNumber = int(encoding.Read16(data, 0))

	// Bytes 2-3: Source fleet number (16-bit)
	b.SourceFleetNumber = int(encoding.Read16(data, 2))

	// Remaining bytes: Transfer info
	if len(data) > 4 {
		b.TransferInfo = make([]byte, len(data)-4)
		copy(b.TransferInfo, data[4:])

		// Extract ship count from byte 3 of TransferInfo
		if len(b.TransferInfo) > 3 {
			b.ShipCount = int(b.TransferInfo[3])
		}
	}
}

// RenameFleetBlock represents a fleet rename operation (Type 44)
// Structure not fully documented - preserves raw data for analysis
type RenameFleetBlock struct {
	GenericBlock
}

// NewRenameFleetBlock creates a RenameFleetBlock from a GenericBlock
func NewRenameFleetBlock(b GenericBlock) *RenameFleetBlock {
	return &RenameFleetBlock{GenericBlock: b}
}
