package blocks

import (
	"github.com/neper-stars/houston/encoding"
)

// Object types
const (
	ObjectTypeMinefield     = 0
	ObjectTypePacketSalvage = 1
	ObjectTypeWormhole      = 2
	ObjectTypeMysteryTrader = 3
)

// Minefield types
const (
	MinefieldTypeStandard  = 0
	MinefieldTypeHeavy     = 1
	MinefieldTypeSpeedBump = 2
)

// Wormhole stability thresholds (raw byte values)
// Stability decreases as the value increases
const (
	WormholeStabilityRockSolid         = 32  // Most stable
	WormholeStabilityStable            = 40
	WormholeStabilityMostlyStable      = 60
	WormholeStabilityAverage           = 80
	WormholeStabilitySlightlyVolatile  = 100
	WormholeStabilityVolatile          = 120
	WormholeStabilityExtremelyVolatile = 196 // Least stable
)

// Mystery trader item bits
const (
	TraderItemMultiCargoPod    = 1 << 0
	TraderItemMultiFunctionPod = 1 << 1
	TraderItemLangstonShield   = 1 << 2
	TraderItemMegaPolyShell    = 1 << 3
	TraderItemAlienMiner       = 1 << 4
	TraderItemHushABoom        = 1 << 5
	TraderItemAntiMatterTorpedo = 1 << 6
	TraderItemMultiContainedMunition = 1 << 7
	TraderItemMiniMorph        = 1 << 8
	TraderItemEnigmaPulsar     = 1 << 9
	TraderItemGenesisDevice    = 1 << 10
	TraderItemJumpGate         = 1 << 11
	TraderItemShip             = 1 << 12
)

// ObjectBlock represents various game objects (Type 43)
// This is a multipurpose block with different subtypes
type ObjectBlock struct {
	GenericBlock

	// Common fields
	IsCountObject bool // True if this is a count object (2 bytes)
	Count         int  // Count value (for count objects only)

	// Map object fields (when not count object)
	Number     int // Object number (9 bits)
	Owner      int // Owner player (5 bits)
	ObjectType int // Object type (3 bits): 0=minefield, 1=packet, 2=wormhole, 3=trader
	X          int // X coordinate
	Y          int // Y coordinate

	// Minefield-specific fields (ObjectType == 0)
	MineCount      int64  // Number of mines
	MinefieldType  int    // 0=standard, 1=heavy, 2=speed bump
	Detonating     bool   // True if detonating

	// Wormhole-specific fields (ObjectType == 2)
	WormholeId      int    // Wormhole ID
	TargetId        int    // Target wormhole ID
	BeenThroughBits uint16 // Player mask: who has been through
	CanSeeBits      uint16 // Player mask: who can see it
	Stability       int    // Wormhole stability (raw byte value)

	// Mystery trader-specific fields (ObjectType == 3)
	XDest    int    // Destination X
	YDest    int    // Destination Y
	Warp     int    // Warp factor
	MetBits  uint16 // Player mask: who trader has met
	ItemBits uint16 // Items the trader is carrying
	TurnNo   int    // Turn number

	// Mineral packet-specific fields (ObjectType == 1, not salvage)
	DestinationPlanetID int // Target planet for the packet
	Ironium             int // Ironium amount in kT
	Boranium            int // Boranium amount in kT
	Germanium           int // Germanium amount in kT
	PacketSpeed         int // Raw speed byte value

	// Salvage-specific fields (ObjectType == 1, salvage variant)
	// Salvage is distinguished from packets by byte 6 == 0xFF
	IsSalvageObject bool // True if this is salvage (not a packet)
	SourceFleetID   int  // Source fleet ID (0-indexed, display is +1)
}

// NewObjectBlock creates an ObjectBlock from a GenericBlock
func NewObjectBlock(b GenericBlock) *ObjectBlock {
	ob := &ObjectBlock{
		GenericBlock: b,
	}
	ob.decode()
	return ob
}

func (ob *ObjectBlock) decode() {
	data := ob.Decrypted

	// Count object (2 bytes)
	if len(data) == 2 {
		ob.IsCountObject = true
		ob.Count = int(encoding.Read16(data, 0))
		return
	}

	if len(data) < 6 {
		return
	}

	// Map object - parse common fields from first 16-bit word
	// Bits 0-8: object number (9 bits)
	// Bits 9-12: owner (4 bits)
	// Bits 13-15: object type (3 bits)
	objectId := encoding.Read16(data, 0)
	ob.Number = int(objectId & 0x01FF)
	ob.Owner = int((objectId >> 9) & 0x0F)
	ob.ObjectType = int(objectId >> 13)

	ob.X = int(encoding.Read16(data, 2))
	ob.Y = int(encoding.Read16(data, 4))

	// Decode type-specific fields
	switch ob.ObjectType {
	case ObjectTypeMinefield:
		ob.decodeMinefield(data)
	case ObjectTypePacketSalvage:
		ob.decodePacket(data)
	case ObjectTypeWormhole:
		ob.decodeWormhole(data)
	case ObjectTypeMysteryTrader:
		ob.decodeMysteryTrader(data)
	}
}


func (ob *ObjectBlock) decodeMinefield(data []byte) {
	if len(data) < 16 {
		return
	}

	ob.MineCount = int64(encoding.Read32(data, 6))
	// Bytes 10-11: unknown
	ob.MinefieldType = int(data[12] & 0xFF)
	ob.Detonating = data[13] != 0
	// Bytes 14-15: player visibility mask
}

func (ob *ObjectBlock) decodePacket(data []byte) {
	if len(data) < 14 {
		return
	}

	// ObjectType 1 can be either a mineral packet or salvage.
	// Distinguish by byte 6:
	//   - Mineral packet: byte 6 = destination planet ID (0-255)
	//   - Salvage: byte 6 = 0xFF (no destination)
	//
	// Mineral packet format (18 bytes total):
	// Bytes 0-1: Object ID (already decoded)
	// Bytes 2-3: X position (already decoded)
	// Bytes 4-5: Y position (already decoded)
	// Byte 6: Destination planet ID (or 0xFF for salvage)
	// Byte 7: Speed byte (packet) or source fleet info (salvage)
	// Bytes 8-9: Ironium in kT
	// Bytes 10-11: Boranium in kT
	// Bytes 12-13: Germanium in kT
	// Bytes 14-17: Unknown

	// Common minerals
	ob.Ironium = int(encoding.Read16(data, 8))
	ob.Boranium = int(encoding.Read16(data, 10))
	ob.Germanium = int(encoding.Read16(data, 12))

	// Check if salvage (byte 6 == 0xFF)
	if data[6] == 0xFF {
		ob.IsSalvageObject = true
		// Byte 7: Low nibble = source fleet ID, high nibble = flags
		ob.SourceFleetID = int(data[7] & 0x0F)
	} else {
		// Mineral packet
		ob.DestinationPlanetID = int(data[6])
		ob.PacketSpeed = int(data[7])
	}
}

func (ob *ObjectBlock) decodeWormhole(data []byte) {
	if len(data) < 16 {
		return
	}

	ob.WormholeId = int(encoding.Read16(data, 0) & 0x0FFF) // Lower 12 bits
	ob.Stability = int(data[6])                            // Stability byte
	ob.BeenThroughBits = encoding.Read16(data, 8)
	ob.CanSeeBits = encoding.Read16(data, 10)
	ob.TargetId = int(encoding.Read16(data, 12) & 0x0FFF) // Lower 12 bits
}

func (ob *ObjectBlock) decodeMysteryTrader(data []byte) {
	if len(data) < 18 {
		return
	}

	ob.XDest = int(encoding.Read16(data, 6))
	ob.YDest = int(encoding.Read16(data, 8))
	ob.Warp = int(data[10] & 0x0F) // Lower 4 bits
	ob.MetBits = encoding.Read16(data, 12)
	ob.ItemBits = encoding.Read16(data, 14)
	ob.TurnNo = int(encoding.Read16(data, 16))
}

// Encode returns the raw block data bytes (without the 2-byte block header).
func (ob *ObjectBlock) Encode() []byte {
	if ob.IsCountObject {
		data := make([]byte, 2)
		encoding.Write16(data, 0, uint16(ob.Count))
		return data
	}

	// Encode based on object type
	switch ob.ObjectType {
	case ObjectTypeMinefield:
		return ob.encodeMinefield()
	case ObjectTypePacketSalvage:
		return ob.encodePacket()
	case ObjectTypeWormhole:
		return ob.encodeWormhole()
	case ObjectTypeMysteryTrader:
		return ob.encodeMysteryTrader()
	default:
		// Fallback: return minimal object data
		data := make([]byte, 6)
		objectId := uint16(ob.Number&0x01FF) | uint16((ob.Owner&0x0F)<<9) | uint16(ob.ObjectType<<13)
		encoding.Write16(data, 0, objectId)
		encoding.Write16(data, 2, uint16(ob.X))
		encoding.Write16(data, 4, uint16(ob.Y))
		return data
	}
}

func (ob *ObjectBlock) encodeMinefield() []byte {
	data := make([]byte, 16)
	objectId := uint16(ob.Number&0x01FF) | uint16((ob.Owner&0x0F)<<9) | uint16(ObjectTypeMinefield<<13)
	encoding.Write16(data, 0, objectId)
	encoding.Write16(data, 2, uint16(ob.X))
	encoding.Write16(data, 4, uint16(ob.Y))
	encoding.Write32(data, 6, uint32(ob.MineCount))
	// Bytes 10-11: unknown (set to 0)
	data[12] = byte(ob.MinefieldType)
	if ob.Detonating {
		data[13] = 1
	}
	// Bytes 14-15: player visibility mask (preserve if available)
	return data
}

func (ob *ObjectBlock) encodePacket() []byte {
	data := make([]byte, 18)
	objectId := uint16(ob.Number&0x01FF) | uint16((ob.Owner&0x0F)<<9) | uint16(ObjectTypePacketSalvage<<13)
	encoding.Write16(data, 0, objectId)
	encoding.Write16(data, 2, uint16(ob.X))
	encoding.Write16(data, 4, uint16(ob.Y))

	if ob.IsSalvageObject {
		data[6] = 0xFF
		data[7] = byte(ob.SourceFleetID & 0x0F)
	} else {
		data[6] = byte(ob.DestinationPlanetID)
		data[7] = byte(ob.PacketSpeed)
	}

	encoding.Write16(data, 8, uint16(ob.Ironium))
	encoding.Write16(data, 10, uint16(ob.Boranium))
	encoding.Write16(data, 12, uint16(ob.Germanium))
	// Bytes 14-17: unknown
	return data
}

func (ob *ObjectBlock) encodeWormhole() []byte {
	data := make([]byte, 16)
	objectId := uint16(ob.Number&0x01FF) | uint16((ob.Owner&0x0F)<<9) | uint16(ObjectTypeWormhole<<13)
	encoding.Write16(data, 0, objectId)
	encoding.Write16(data, 2, uint16(ob.X))
	encoding.Write16(data, 4, uint16(ob.Y))
	data[6] = byte(ob.Stability)
	// Byte 7: unknown
	encoding.Write16(data, 8, ob.BeenThroughBits)
	encoding.Write16(data, 10, ob.CanSeeBits)
	encoding.Write16(data, 12, uint16(ob.TargetId&0x0FFF))
	// Bytes 14-15: unknown
	return data
}

func (ob *ObjectBlock) encodeMysteryTrader() []byte {
	data := make([]byte, 18)
	objectId := uint16(ob.Number&0x01FF) | uint16((ob.Owner&0x0F)<<9) | uint16(ObjectTypeMysteryTrader<<13)
	encoding.Write16(data, 0, objectId)
	encoding.Write16(data, 2, uint16(ob.X))
	encoding.Write16(data, 4, uint16(ob.Y))
	encoding.Write16(data, 6, uint16(ob.XDest))
	encoding.Write16(data, 8, uint16(ob.YDest))
	data[10] = byte(ob.Warp & 0x0F)
	// Byte 11: unknown
	encoding.Write16(data, 12, ob.MetBits)
	encoding.Write16(data, 14, ob.ItemBits)
	encoding.Write16(data, 16, uint16(ob.TurnNo))
	return data
}

// IsMinefield returns true if this is a minefield object
func (ob *ObjectBlock) IsMinefield() bool {
	return !ob.IsCountObject && ob.ObjectType == ObjectTypeMinefield
}

// IsWormhole returns true if this is a wormhole object
func (ob *ObjectBlock) IsWormhole() bool {
	return !ob.IsCountObject && ob.ObjectType == ObjectTypeWormhole
}

// IsMysteryTrader returns true if this is a mystery trader object
func (ob *ObjectBlock) IsMysteryTrader() bool {
	return !ob.IsCountObject && ob.ObjectType == ObjectTypeMysteryTrader
}

// IsPacket returns true if this is a mineral packet object (not salvage)
func (ob *ObjectBlock) IsPacket() bool {
	return !ob.IsCountObject && ob.ObjectType == ObjectTypePacketSalvage && !ob.IsSalvageObject
}

// IsSalvage returns true if this is a salvage object
func (ob *ObjectBlock) IsSalvage() bool {
	return !ob.IsCountObject && ob.ObjectType == ObjectTypePacketSalvage && ob.IsSalvageObject
}

// TotalMinerals returns the total mineral content of a packet
func (ob *ObjectBlock) TotalMinerals() int {
	return ob.Ironium + ob.Boranium + ob.Germanium
}

// WarpSpeed returns the decoded warp speed for a mineral packet
// The raw byte encodes warp as: rawByte = (warp - 5) * 4 + 196
// So warp = (rawByte >> 2) - 44
func (ob *ObjectBlock) WarpSpeed() int {
	return (ob.PacketSpeed >> 2) - 44
}

// PlayerCanSee returns true if the given player can see this wormhole
func (ob *ObjectBlock) PlayerCanSee(playerIndex int) bool {
	if playerIndex < 0 || playerIndex >= 16 {
		return false
	}
	return (ob.CanSeeBits & (1 << playerIndex)) != 0
}

// PlayerBeenThrough returns true if the given player has been through this wormhole
func (ob *ObjectBlock) PlayerBeenThrough(playerIndex int) bool {
	if playerIndex < 0 || playerIndex >= 16 {
		return false
	}
	return (ob.BeenThroughBits & (1 << playerIndex)) != 0
}

// TraderHasMet returns true if the trader has met the given player
func (ob *ObjectBlock) TraderHasMet(playerIndex int) bool {
	if playerIndex < 0 || playerIndex >= 16 {
		return false
	}
	return (ob.MetBits & (1 << playerIndex)) != 0
}

// TraderHasItem returns true if the trader has the given item
func (ob *ObjectBlock) TraderHasItem(itemBit uint16) bool {
	return (ob.ItemBits & itemBit) != 0
}

// StabilityName returns the human-readable stability name for a wormhole
func (ob *ObjectBlock) StabilityName() string {
	if ob.Stability <= WormholeStabilityRockSolid {
		return "Rock Solid"
	} else if ob.Stability <= WormholeStabilityStable {
		return "Stable"
	} else if ob.Stability <= WormholeStabilityMostlyStable {
		return "Mostly Stable"
	} else if ob.Stability <= WormholeStabilityAverage {
		return "Average"
	} else if ob.Stability <= WormholeStabilitySlightlyVolatile {
		return "Slightly Volatile"
	} else if ob.Stability <= WormholeStabilityVolatile {
		return "Volatile"
	}
	return "Extremely Volatile"
}
