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

// Wormhole stability index values (iStable field, 2 bits)
// Lower values = more stable
const (
	WormholeStabilityIndexRockSolid    = 0 // Most stable
	WormholeStabilityIndexStable       = 1
	WormholeStabilityIndexVolatile     = 2
	WormholeStabilityIndexVeryVolatile = 3 // Least stable
)

// Mystery trader item bits
const (
	TraderItemMultiCargoPod          = 1 << 0
	TraderItemMultiFunctionPod       = 1 << 1
	TraderItemLangstonShield         = 1 << 2
	TraderItemMegaPolyShell          = 1 << 3
	TraderItemAlienMiner             = 1 << 4
	TraderItemHushABoom              = 1 << 5
	TraderItemAntiMatterTorpedo      = 1 << 6
	TraderItemMultiContainedMunition = 1 << 7
	TraderItemMiniMorph              = 1 << 8
	TraderItemEnigmaPulsar           = 1 << 9
	TraderItemGenesisDevice          = 1 << 10
	TraderItemJumpGate               = 1 << 11
	TraderItemShip                   = 1 << 12
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
	MineCount          int64  // Number of mines
	MineCanSeeBits     uint16 // Player visibility mask (historical, who has ever detected)
	MinefieldType      int    // 0=standard, 1=heavy, 2=speed bump
	Detonating         bool   // True if detonating
	MineCurrentSeeBits uint16 // Current turn visibility bitmask (grbitPlrNow)
	MineTurnNumber     int    // Turn number when minefield was last updated

	// Wormhole-specific fields (ObjectType == 2)
	WormholeId         int    // Wormhole ID
	TargetId           int    // Target wormhole ID
	CanSeeBits         uint16 // Player mask: who can see it (bytes 8-9)
	BeenThroughBits    uint16 // Player mask: who has been through (bytes 10-11)
	StabilityIndex     int    // Wormhole stability index (0-3, bits 0-1 of stability word)
	TurnsSinceMove     int    // Turns since last movement (0-1023, bits 2-11)
	DestKnown          bool   // Destination known to players (bit 12)
	IncludeInDisplay   bool   // Include in display flag (bit 13)
	WormholePadding    uint16 // Padding bytes 14-15 (unused, preserved for round-trip)
	WormholeTurnNumber int    // Turn number when wormhole was last updated

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
	PacketMaxWeight     int // Maximum weight/capacity in kT (bits 0-13 of bytes 14-15)
	PacketDecayRate     int // Decay rate index 0-3 (bits 14-15 of bytes 14-15)
	PacketTurnNumber    int // Turn number when packet was created/updated

	// Salvage-specific fields (ObjectType == 1, salvage variant)
	// Salvage is distinguished from packets by byte 6 == 0xFF
	IsSalvageObject    bool // True if this is salvage (not a packet)
	SourceFleetID      int  // Source fleet ID (low nibble of byte 7, 0-indexed)
	SalvageSourceFlags int  // Source flags (high nibble of byte 7)
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
	if len(data) < 14 {
		return
	}

	ob.MineCount = int64(encoding.Read32(data, 6))
	ob.MineCanSeeBits = encoding.Read16(data, 10)
	// Byte 12: Minefield type (0=standard, 1=heavy, 2=speed bump)
	// Byte 13: Detonation flag (1 if detonating)
	ob.MinefieldType = int(data[12] & 0xFF)
	ob.Detonating = data[13] == 1

	if len(data) >= 16 {
		// Bytes 14-15: Current turn visibility bitmask (grbitPlrNow)
		ob.MineCurrentSeeBits = encoding.Read16(data, 14)
	}
	if len(data) >= 18 {
		ob.MineTurnNumber = int(encoding.Read16(data, 16))
	}
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
	// Bytes 14-15: wtMax|iDecayRate (bits 0-13 = max weight, bits 14-15 = decay rate)
	// Bytes 16-17: Turn number

	// Common minerals
	ob.Ironium = int(encoding.Read16(data, 8))
	ob.Boranium = int(encoding.Read16(data, 10))
	ob.Germanium = int(encoding.Read16(data, 12))

	// Check if salvage (byte 6 == 0xFF)
	if data[6] == 0xFF {
		ob.IsSalvageObject = true
		// Byte 7: Low nibble = source fleet ID, high nibble = flags
		ob.SourceFleetID = int(data[7] & 0x0F)
		ob.SalvageSourceFlags = int((data[7] >> 4) & 0x0F)
	} else {
		// Mineral packet
		ob.DestinationPlanetID = int(data[6])
		ob.PacketSpeed = int(data[7])
	}

	// Bytes 14-15: wtMax|iDecayRate
	if len(data) >= 16 {
		wtMaxDecay := encoding.Read16(data, 14)
		ob.PacketMaxWeight = int(wtMaxDecay & 0x3FFF)       // Bits 0-13: max weight (14 bits)
		ob.PacketDecayRate = int((wtMaxDecay >> 14) & 0x03) // Bits 14-15: decay rate (2 bits)
	}
	if len(data) >= 18 {
		ob.PacketTurnNumber = int(encoding.Read16(data, 16))
	}
}

func (ob *ObjectBlock) decodeWormhole(data []byte) {
	if len(data) < 14 {
		return
	}

	ob.WormholeId = int(encoding.Read16(data, 0) & 0x0FFF) // Lower 12 bits

	// Bytes 6-7: Stability/movement word (THWORM structure)
	// Bits 0-1: iStable (stability index, 0-3)
	// Bits 2-11: cLastMove (turns since last movement, 0-1023)
	// Bit 12: fDestKnown (destination known to players)
	// Bit 13: fInclude (include in display flag)
	// Bits 14-15: unused
	stabilityWord := encoding.Read16(data, 6)
	ob.StabilityIndex = int(stabilityWord & 0x03)          // Bits 0-1
	ob.TurnsSinceMove = int((stabilityWord >> 2) & 0x03FF) // Bits 2-11 (10 bits)
	ob.DestKnown = (stabilityWord & (1 << 12)) != 0        // Bit 12
	ob.IncludeInDisplay = (stabilityWord & (1 << 13)) != 0 // Bit 13

	// Bytes 8-9: grbitPlr (visibility mask)
	ob.CanSeeBits = encoding.Read16(data, 8)
	// Bytes 10-11: grbitPlrTrav (traversal mask)
	ob.BeenThroughBits = encoding.Read16(data, 10)
	// Bytes 12-13: idPartner (target wormhole ID, lower 12 bits)
	ob.TargetId = int(encoding.Read16(data, 12) & 0x0FFF)

	if len(data) >= 16 {
		// Bytes 14-15: Padding (THWORM is 8 bytes, but THING union is 10 bytes)
		ob.WormholePadding = encoding.Read16(data, 14)
	}
	if len(data) >= 18 {
		ob.WormholeTurnNumber = int(encoding.Read16(data, 16))
	}
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
	data := make([]byte, 18)
	objectId := uint16(ob.Number&0x01FF) | uint16((ob.Owner&0x0F)<<9) | uint16(ObjectTypeMinefield<<13)
	encoding.Write16(data, 0, objectId)
	encoding.Write16(data, 2, uint16(ob.X))
	encoding.Write16(data, 4, uint16(ob.Y))
	encoding.Write32(data, 6, uint32(ob.MineCount))
	encoding.Write16(data, 10, ob.MineCanSeeBits)
	// Byte 12: Minefield type (0=standard, 1=heavy, 2=speed bump)
	// Byte 13: Detonation flag (1 if detonating)
	data[12] = byte(ob.MinefieldType & 0xFF)
	if ob.Detonating {
		data[13] = 1
	} else {
		data[13] = 0
	}
	encoding.Write16(data, 14, ob.MineCurrentSeeBits)
	encoding.Write16(data, 16, uint16(ob.MineTurnNumber))
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
		data[7] = byte(ob.SourceFleetID&0x0F) | byte((ob.SalvageSourceFlags&0x0F)<<4)
	} else {
		data[6] = byte(ob.DestinationPlanetID)
		data[7] = byte(ob.PacketSpeed)
	}

	encoding.Write16(data, 8, uint16(ob.Ironium))
	encoding.Write16(data, 10, uint16(ob.Boranium))
	encoding.Write16(data, 12, uint16(ob.Germanium))
	// Bytes 14-15: wtMax|iDecayRate
	wtMaxDecay := uint16(ob.PacketMaxWeight&0x3FFF) | uint16((ob.PacketDecayRate&0x03)<<14)
	encoding.Write16(data, 14, wtMaxDecay)
	encoding.Write16(data, 16, uint16(ob.PacketTurnNumber))
	return data
}

func (ob *ObjectBlock) encodeWormhole() []byte {
	data := make([]byte, 18)
	objectId := uint16(ob.Number&0x01FF) | uint16((ob.Owner&0x0F)<<9) | uint16(ObjectTypeWormhole<<13)
	encoding.Write16(data, 0, objectId)
	encoding.Write16(data, 2, uint16(ob.X))
	encoding.Write16(data, 4, uint16(ob.Y))
	// Bytes 6-7: Stability/movement word
	stabilityWord := uint16(ob.StabilityIndex&0x03) |
		uint16((ob.TurnsSinceMove&0x03FF)<<2)
	if ob.DestKnown {
		stabilityWord |= (1 << 12)
	}
	if ob.IncludeInDisplay {
		stabilityWord |= (1 << 13)
	}
	encoding.Write16(data, 6, stabilityWord)
	encoding.Write16(data, 8, ob.CanSeeBits)
	encoding.Write16(data, 10, ob.BeenThroughBits)
	encoding.Write16(data, 12, uint16(ob.TargetId))
	encoding.Write16(data, 14, ob.WormholePadding)
	encoding.Write16(data, 16, uint16(ob.WormholeTurnNumber))
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

// PlayerCanSeeMinefield returns true if the given player can see this minefield
func (ob *ObjectBlock) PlayerCanSeeMinefield(playerIndex int) bool {
	if playerIndex < 0 || playerIndex >= 16 {
		return false
	}
	return (ob.MineCanSeeBits & (1 << playerIndex)) != 0
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
// Based on the iStable index (0-3)
func (ob *ObjectBlock) StabilityName() string {
	switch ob.StabilityIndex {
	case WormholeStabilityIndexRockSolid:
		return "Rock Solid"
	case WormholeStabilityIndexStable:
		return "Stable"
	case WormholeStabilityIndexVolatile:
		return "Volatile"
	case WormholeStabilityIndexVeryVolatile:
		return "Very Volatile"
	default:
		return "Unknown"
	}
}
