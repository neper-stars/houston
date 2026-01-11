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

	// Fleet identification (see reversing_notes/fleet-block.md)
	FleetNumber int // 0-511 (9 bits from bytes 0-1)
	Owner       int // 0-15 (4 bits from bytes 0-1)

	// Header bytes (preserved for encoding)
	// Bytes 2-3: iPlayer (int16) - Owner player index (redundant with Owner from bytes 0-1)
	Byte2    byte
	Byte3    byte
	KindByte byte // Byte 4: det - detection level (3=partial, 4=pick-pocket, 7=full)
	Byte5    byte // Byte 5: flags

	// Position
	PositionObjectId int // Referenced object
	X                int // X coordinate
	Y                int // Y coordinate

	// Ship composition
	ShipTypes         uint16  // 16-bit bitmask, one bit per design
	ShipCount         [16]int // Ship counts per design slot
	ShipCountTwoBytes bool    // True if ship counts are 2 bytes each (bit 3 of Byte5 clear)

	// Flag bits from Byte5 (see reversing_notes/fleet-block.md)
	// Bit 0: fInclude - Include in reports/selection
	Include bool
	// Bit 1: fRepOrders - Repeat waypoint orders when complete
	RepeatOrders bool
	// Bit 2: fDead - Fleet has been destroyed
	IsDead bool
	// Bit 3: fByteCsh - handled by ShipCountTwoBytes field
	// Bits 4-7: NOT PERSISTED - these are runtime-only flags (fDone, fBombed, fHereAllTurn, fNoHeal)
	//           They are zeroed when writing to file and recalculated each turn.

	// Resources/Cargo (if full or pick-pocket)
	Ironium    int64
	Boranium   int64
	Germanium  int64
	Population int64
	Fuel       int64

	// Partial fleet movement data (if kindByte != FleetKindFull)
	// See reversing_notes/fleet-block.md for warp byte breakdown
	DeltaX int
	DeltaY int
	Warp   int // 0-15 (bits 0-3 of warp byte)
	Mass   int64

	// Warp byte upper bits (bits 4-7) - movement/status flags from dirLong union
	DirectionValid     bool // Bit 4 (0x10): fdirValid - Direction is valid (fleet has a destination)
	CompositionChanged bool // Bit 5 (0x20): fCompChg - Composition changed (fleet merged/split)
	Targeted           bool // Bit 6 (0x40): fTargeted - Fleet is targeted by another fleet
	Skipped            bool // Bit 7 (0x80): fSkipped - Fleet was skipped this turn

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

	// Extract flag bits from Byte5 (see reversing_notes/fleet-block.md)
	fb.Include = (fb.Byte5 & 0x01) != 0      // Bit 0: fInclude - Include in reports/selection
	fb.RepeatOrders = (fb.Byte5 & 0x02) != 0 // Bit 1: fRepOrders - Repeat waypoint orders
	fb.IsDead = (fb.Byte5 & 0x04) != 0       // Bit 2: fDead - Fleet has been destroyed
	// Bits 4-7 are NOT persisted - runtime-only flags (fDone, fBombed, fHereAllTurn, fNoHeal)

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
		}
	} else if index+8 <= len(data) {
		// Partial fleet movement data
		// DeltaX/DeltaY are unsigned bytes (0-255) centered around 127
		// 127 means no movement, values above/below indicate direction
		// We convert to actual delta by subtracting 127
		fb.DeltaX = int(data[index]) - 127
		fb.DeltaY = int(data[index+1]) - 127
		warpByte := data[index+2]
		fb.Warp = int(warpByte & 0x0F)
		// Extract upper bits as movement/status flags (see reversing_notes/fleet-block.md)
		fb.DirectionValid = (warpByte & 0x10) != 0     // Bit 4: fdirValid
		fb.CompositionChanged = (warpByte & 0x20) != 0 // Bit 5: fCompChg
		fb.Targeted = (warpByte & 0x40) != 0           // Bit 6: fTargeted
		fb.Skipped = (warpByte & 0x80) != 0            // Bit 7: fSkipped
		// index+3 is padding (should be 0)
		fb.Mass = int64(encoding.Read32(data, index+4))
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

	FleetNumber   int   // The target fleet (9 bits)
	FleetsToMerge []int // List of fleet numbers to merge into target
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
// Found in M files, follows the FleetBlock it names.
//
// Format:
//
//	LL [encoded name bytes...]
//	│  └─────────────────────── Stars! encoded string (LL bytes)
//	└────────────────────────── Name length
//
// Note: The fleet association is positional - this block immediately
// follows the FleetBlock whose name it contains.
type FleetNameBlock struct {
	GenericBlock

	Name string // Decoded fleet name
}

// NewFleetNameBlock creates a FleetNameBlock from a GenericBlock
func NewFleetNameBlock(b GenericBlock) *FleetNameBlock {
	fnb := &FleetNameBlock{GenericBlock: b}
	fnb.decode()
	return fnb
}

func (fnb *FleetNameBlock) decode() {
	data := fnb.Decrypted
	if len(data) < 1 {
		return
	}

	name, err := encoding.DecodeStarsString(data)
	if err == nil {
		fnb.Name = name
	}
}

// ShipTransfer represents a single design's ship transfer in a merge operation
type ShipTransfer struct {
	DesignSlot int // Design slot number (0-15)
	Count      int // Signed count: positive = ships arriving at dest, negative = ships leaving dest
}

// MoveShipsBlock represents ship movement between fleets (Type 23)
// Used when transferring ships from one fleet to another (merge operation)
//
// TransferInfo structure:
//   - Byte 0: Flags (typically 0x22)
//   - Byte 1: Ship types bitmask (low byte) - bit N set means design slot N is involved
//   - Byte 2: Ship types bitmask (high byte) - for design slots 8-15
//   - Bytes 3+: Signed int16 counts for each design slot with bit set
//
// Counts are from dest fleet's perspective:
//   - Positive count = ships arriving at dest fleet (from source)
//   - Negative count = ships leaving dest fleet (to source)
type MoveShipsBlock struct {
	GenericBlock

	DestFleetNumber   int            // Destination fleet number (0-indexed)
	SourceFleetNumber int            // Source fleet number (0-indexed)
	ShipTypeMask      uint16         // Bitmask of design slots involved in transfer
	ShipTransfers     []ShipTransfer // List of transfers per design slot
	TransferInfo      []byte         // Raw transfer data (preserved for analysis)
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

		if len(b.TransferInfo) >= 3 {
			// Byte 0: flags (typically 0x22)
			// Bytes 1-2: ship types bitmask (16-bit, little-endian)
			b.ShipTypeMask = uint16(b.TransferInfo[1]) | (uint16(b.TransferInfo[2]) << 8)

			// Parse ship counts for each design slot with bit set
			idx := 3
			for slot := 0; slot < 16; slot++ {
				if b.ShipTypeMask&(1<<slot) != 0 {
					if idx+1 < len(b.TransferInfo) {
						// Signed 16-bit count (little-endian)
						count := int(int16(b.TransferInfo[idx]) | (int16(b.TransferInfo[idx+1]) << 8))
						b.ShipTransfers = append(b.ShipTransfers, ShipTransfer{
							DesignSlot: slot,
							Count:      count,
						})
						idx += 2
					}
				}
			}
		}
	}
}

// RenameFleetBlock represents a fleet rename operation (Type 44)
type RenameFleetBlock struct {
	GenericBlock

	FleetNumber int    // Fleet being renamed (0-indexed)
	NewName     string // The new name for the fleet
}

// NewRenameFleetBlock creates a RenameFleetBlock from a GenericBlock
func NewRenameFleetBlock(b GenericBlock) *RenameFleetBlock {
	block := &RenameFleetBlock{GenericBlock: b}
	block.decode()
	return block
}

func (b *RenameFleetBlock) decode() {
	data := b.Decrypted
	if len(data) < 5 {
		return
	}

	// Bytes 0-1: Fleet number (16-bit)
	b.FleetNumber = int(encoding.Read16(data, 0))

	// Bytes 2-3: Unknown (typically matches fleet number)
	// Bytes 4+: Encoded name (Stars! string format)
	if len(data) > 4 {
		name, err := encoding.DecodeStarsString(data[4:])
		if err == nil {
			b.NewName = name
		}
	}
}

// Encode returns the raw block data bytes for FleetSplitBlock.
func (fsb *FleetSplitBlock) Encode() []byte {
	data := make([]byte, 2)
	data[0] = byte(fsb.FleetNumber & 0xFF)
	data[1] = byte((fsb.FleetNumber >> 8) & 0x01)
	return data
}

// Encode returns the raw block data bytes for FleetsMergeBlock.
func (fmb *FleetsMergeBlock) Encode() []byte {
	data := make([]byte, 2+len(fmb.FleetsToMerge)*2)
	data[0] = byte(fmb.FleetNumber & 0xFF)
	data[1] = byte((fmb.FleetNumber >> 8) & 0x01)

	for i, fleet := range fmb.FleetsToMerge {
		offset := 2 + i*2
		data[offset] = byte(fleet & 0xFF)
		data[offset+1] = byte((fleet >> 8) & 0x01)
	}
	return data
}

// Encode returns the raw block data bytes for FleetNameBlock.
func (fnb *FleetNameBlock) Encode() []byte {
	return encoding.EncodeStarsString(fnb.Name)
}

// Encode returns the raw block data bytes for MoveShipsBlock.
func (b *MoveShipsBlock) Encode() []byte {
	// Calculate size: 4 bytes header + transfer info
	size := 4 + len(b.TransferInfo)
	data := make([]byte, size)

	encoding.Write16(data, 0, uint16(b.DestFleetNumber))
	encoding.Write16(data, 2, uint16(b.SourceFleetNumber))
	copy(data[4:], b.TransferInfo)

	return data
}

// Encode returns the raw block data bytes for RenameFleetBlock.
func (b *RenameFleetBlock) Encode() []byte {
	encodedName := encoding.EncodeStarsString(b.NewName)
	data := make([]byte, 4+len(encodedName))

	encoding.Write16(data, 0, uint16(b.FleetNumber))
	encoding.Write16(data, 2, uint16(b.FleetNumber)) // Unknown field, typically matches fleet number
	copy(data[4:], encodedName)

	return data
}

// Encode returns the raw block data bytes for PartialFleetBlock.
// This handles the complex variable-length encoding of ship counts and cargo.
func (fb *PartialFleetBlock) Encode() []byte {
	// Calculate the size needed
	size := 14 // Fixed header

	// Ship counts (variable length)
	shipCountBytes := 0
	for bit := 0; bit < 16; bit++ {
		if (fb.ShipTypes & (1 << bit)) != 0 {
			if fb.ShipCountTwoBytes {
				shipCountBytes += 2
			} else {
				shipCountBytes++
			}
		}
	}
	size += shipCountBytes

	// Cargo (if applicable)
	if fb.HasCargo() {
		size += 2 // Contents lengths word
		size += encoding.VarLenByteCount(encoding.ByteLengthForInt(fb.Ironium))
		size += encoding.VarLenByteCount(encoding.ByteLengthForInt(fb.Boranium))
		size += encoding.VarLenByteCount(encoding.ByteLengthForInt(fb.Germanium))
		size += encoding.VarLenByteCount(encoding.ByteLengthForInt(fb.Population))
		size += encoding.VarLenByteCount(encoding.ByteLengthForInt(fb.Fuel))
	}

	// Full fleet vs partial fleet specific data
	if fb.KindByte == FleetKindFull {
		size += 2 // Damaged ship types
		for bit := 0; bit < 16; bit++ {
			if (fb.DamagedShipTypes & (1 << bit)) != 0 {
				size += 2
			}
		}
		size += 2 // Battle plan and waypoint count
	} else {
		size += 8 // Delta X, Delta Y, Warp, padding, Mass
	}

	data := make([]byte, size)
	index := 0

	// Bytes 0-1: Fleet number and owner
	data[index] = byte(fb.FleetNumber & 0xFF)
	index++
	data[index] = byte(((fb.FleetNumber >> 8) & 0x01) | (fb.Owner << 1))
	index++

	// Bytes 2-5: Control bytes
	data[index] = fb.Byte2
	index++
	data[index] = fb.Byte3
	index++
	data[index] = fb.KindByte
	index++
	// Reconstruct Byte5 from individual flag fields (see reversing_notes/fleet-block.md)
	byte5 := byte(0)
	if fb.Include {
		byte5 |= 0x01 // Bit 0: fInclude
	}
	if fb.RepeatOrders {
		byte5 |= 0x02 // Bit 1: fRepOrders
	}
	if fb.IsDead {
		byte5 |= 0x04 // Bit 2: fDead
	}
	if !fb.ShipCountTwoBytes {
		byte5 |= 0x08 // Bit 3: fByteCsh - clear means 2-byte counts
	}
	// Bits 4-7 are NOT persisted - they are runtime-only flags
	data[index] = byte5
	index++

	// Bytes 6-13: Position and ship types
	encoding.Write16(data, index, uint16(fb.PositionObjectId))
	index += 2
	encoding.Write16(data, index, uint16(fb.X))
	index += 2
	encoding.Write16(data, index, uint16(fb.Y))
	index += 2
	encoding.Write16(data, index, fb.ShipTypes)
	index += 2

	// Variable-length ship counts
	for bit := 0; bit < 16; bit++ {
		if (fb.ShipTypes & (1 << bit)) != 0 {
			if fb.ShipCountTwoBytes {
				encoding.Write16(data, index, uint16(fb.ShipCount[bit]))
				index += 2
			} else {
				data[index] = byte(fb.ShipCount[bit])
				index++
			}
		}
	}

	// Cargo section
	if fb.HasCargo() {
		contentsLengths := encoding.PackVarLenIndicators(
			fb.Ironium, fb.Boranium, fb.Germanium, fb.Population, fb.Fuel,
		)
		encoding.Write16(data, index, contentsLengths)
		index += 2

		index = encoding.WriteVarLen(data, index, fb.Ironium)
		index = encoding.WriteVarLen(data, index, fb.Boranium)
		index = encoding.WriteVarLen(data, index, fb.Germanium)
		index = encoding.WriteVarLen(data, index, fb.Population)
		index = encoding.WriteVarLen(data, index, fb.Fuel)
	}

	// Full fleet vs partial fleet specific data
	if fb.KindByte == FleetKindFull {
		encoding.Write16(data, index, fb.DamagedShipTypes)
		index += 2

		for bit := 0; bit < 16; bit++ {
			if (fb.DamagedShipTypes & (1 << bit)) != 0 {
				encoding.Write16(data, index, fb.DamagedShipInfo[bit])
				index += 2
			}
		}

		data[index] = byte(fb.BattlePlan)
		index++
		data[index] = byte(fb.WaypointCount)
		index++
	} else {
		data[index] = byte(int8(fb.DeltaX))
		index++
		data[index] = byte(int8(fb.DeltaY))
		index++
		// Reconstruct warp byte from warp speed and movement/status flags
		warpByte := byte(fb.Warp & 0x0F)
		if fb.DirectionValid {
			warpByte |= 0x10 // Bit 4: fdirValid
		}
		if fb.CompositionChanged {
			warpByte |= 0x20 // Bit 5: fCompChg
		}
		if fb.Targeted {
			warpByte |= 0x40 // Bit 6: fTargeted
		}
		if fb.Skipped {
			warpByte |= 0x80 // Bit 7: fSkipped
		}
		data[index] = warpByte
		index++
		data[index] = 0 // Padding
		index++
		encoding.Write32(data, index, uint32(fb.Mass))
		index += 4
	}

	return data[:index]
}

// Encode returns the raw block data bytes for FleetBlock.
func (fb *FleetBlock) Encode() []byte {
	return fb.PartialFleetBlock.Encode()
}
