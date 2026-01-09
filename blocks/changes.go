package blocks

// ResearchChangeBlock represents a research priority change (Type 34)
//
// Format (2 bytes):
//
//	Byte 0: Research budget percentage (0-100)
//	Byte 1: (next_field << 4) | current_field
//	        High nibble = next research field (0-5)
//	        Low nibble = current research field (0-5)
type ResearchChangeBlock struct {
	GenericBlock

	BudgetPercent int // Research budget percentage (0-100)
	CurrentField  int // Current research field (0-5)
	NextField     int // Next research field after current completes (0-5)
}

// NewResearchChangeBlock creates a ResearchChangeBlock from a GenericBlock
func NewResearchChangeBlock(b GenericBlock) *ResearchChangeBlock {
	rcb := &ResearchChangeBlock{GenericBlock: b}
	rcb.decode()
	return rcb
}

func (rcb *ResearchChangeBlock) decode() {
	data := rcb.Decrypted
	if len(data) < 2 {
		return
	}

	rcb.BudgetPercent = int(data[0])
	rcb.NextField = int((data[1] >> 4) & 0x0F)
	rcb.CurrentField = int(data[1] & 0x0F)
}

// Encode returns the raw block data bytes (without the 2-byte block header).
func (rcb *ResearchChangeBlock) Encode() []byte {
	data := make([]byte, 2)
	data[0] = byte(rcb.BudgetPercent)
	data[1] = byte((rcb.NextField&0x0F)<<4) | byte(rcb.CurrentField&0x0F)
	return data
}

// PlanetChangeBlock represents planet settings changes (Type 35)
//
// Format (6 bytes):
//
//	Bytes 0-1: Planet ID (uint16 LE)
//	Bytes 2-3: Packed settings (uint16 LE)
//	           Bit 0:      fNoResearch - Contribute leftover to research (1=yes)
//	           Bits 1-10:  pctDp - Driver/packet percentage (10 bits, 0-1023)
//	           Bits 11-14: iWarpFling - Packet fling warp speed (0-15)
//	           Bit 15:     unused
//	Bytes 4-5: Route destination (uint16 LE, left-shifted by 1)
//	           Actual planet ID = value >> 1
//	           0 = no route
type PlanetChangeBlock struct {
	GenericBlock

	PlanetId            int  // Planet ID
	ContributeLeftover  bool // "Contribute only leftover resources to research" (fNoResearch bit 0)
	DriverPacketPercent int  // Mass driver/packet percentage (0-1023, typically 0-100)
	PacketWarpSpeed     int  // Packet fling warp speed (0-15)
	RouteDestinationId  int  // Route destination planet ID (0 = no route)
}

// NewPlanetChangeBlock creates a PlanetChangeBlock from a GenericBlock
func NewPlanetChangeBlock(b GenericBlock) *PlanetChangeBlock {
	pcb := &PlanetChangeBlock{GenericBlock: b}
	pcb.decode()
	return pcb
}

func (pcb *PlanetChangeBlock) decode() {
	data := pcb.Decrypted
	if len(data) < 6 {
		return
	}

	// Bytes 0-1: Planet ID (uint16 LE)
	pcb.PlanetId = int(data[0]) | (int(data[1]) << 8)

	// Bytes 2-3: Packed settings (uint16 LE)
	settings := int(data[2]) | (int(data[3]) << 8)
	pcb.ContributeLeftover = (settings & 0x01) != 0   // Bit 0
	pcb.DriverPacketPercent = (settings >> 1) & 0x3FF // Bits 1-10 (10 bits)
	pcb.PacketWarpSpeed = (settings >> 11) & 0x0F     // Bits 11-14 (4 bits)

	// Bytes 4-5: Route destination (uint16 LE, left-shifted by 1)
	routeRaw := int(data[4]) | (int(data[5]) << 8)
	pcb.RouteDestinationId = routeRaw >> 1 // Actual planet ID
}

// Encode returns the raw block data bytes (without the 2-byte block header).
func (pcb *PlanetChangeBlock) Encode() []byte {
	data := make([]byte, 6)

	// Bytes 0-1: Planet ID (uint16 LE)
	data[0] = byte(pcb.PlanetId & 0xFF)
	data[1] = byte((pcb.PlanetId >> 8) & 0xFF)

	// Bytes 2-3: Packed settings (uint16 LE)
	var settings int
	if pcb.ContributeLeftover {
		settings |= 0x01 // Bit 0
	}
	settings |= (pcb.DriverPacketPercent & 0x3FF) << 1 // Bits 1-10
	settings |= (pcb.PacketWarpSpeed & 0x0F) << 11     // Bits 11-14
	data[2] = byte(settings & 0xFF)
	data[3] = byte((settings >> 8) & 0xFF)

	// Bytes 4-5: Route destination (uint16 LE, left-shifted by 1)
	routeRaw := pcb.RouteDestinationId << 1
	data[4] = byte(routeRaw & 0xFF)
	data[5] = byte((routeRaw >> 8) & 0xFF)

	return data
}

// ChangePasswordBlock represents a password change request (Type 36)
//
// Format (4 bytes):
//
//	Bytes 0-3: New password hash (uint32 little-endian)
//	           Hash 0 = no password / remove password
//
// The hash is computed using HashRacePassword() from the password package.
type ChangePasswordBlock struct {
	GenericBlock

	NewPasswordHash uint32 // New password hash (0 = no password)
}

// NewChangePasswordBlock creates a ChangePasswordBlock from a GenericBlock
func NewChangePasswordBlock(b GenericBlock) *ChangePasswordBlock {
	cpb := &ChangePasswordBlock{GenericBlock: b}
	cpb.decode()
	return cpb
}

func (cpb *ChangePasswordBlock) decode() {
	data := cpb.Decrypted
	if len(data) < 4 {
		return
	}

	cpb.NewPasswordHash = uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
}

// Encode returns the raw block data bytes (without the 2-byte block header).
func (cpb *ChangePasswordBlock) Encode() []byte {
	data := make([]byte, 4)
	data[0] = byte(cpb.NewPasswordHash & 0xFF)
	data[1] = byte((cpb.NewPasswordHash >> 8) & 0xFF)
	data[2] = byte((cpb.NewPasswordHash >> 16) & 0xFF)
	data[3] = byte((cpb.NewPasswordHash >> 24) & 0xFF)
	return data
}

// HasPassword returns true if this sets a password (hash != 0)
func (cpb *ChangePasswordBlock) HasPassword() bool {
	return cpb.NewPasswordHash != 0
}

// Diplomatic relation types
const (
	RelationFriend  = 0
	RelationNeutral = 1
	RelationEnemy   = 2
)

// PlayersRelationChangeBlock represents diplomatic relation changes (Type 38)
//
// Format (2 bytes):
//
//	Byte 0: Relation type (0=Friend, 1=Neutral, 2=Enemy)
//	Byte 1: Target player index (0-15)
type PlayersRelationChangeBlock struct {
	GenericBlock

	Relation     int // Diplomatic relation: 0=Friend, 1=Neutral, 2=Enemy
	TargetPlayer int // Target player index (0-15)
}

// NewPlayersRelationChangeBlock creates a PlayersRelationChangeBlock from a GenericBlock
func NewPlayersRelationChangeBlock(b GenericBlock) *PlayersRelationChangeBlock {
	prc := &PlayersRelationChangeBlock{GenericBlock: b}
	prc.decode()
	return prc
}

func (prc *PlayersRelationChangeBlock) decode() {
	data := prc.Decrypted
	if len(data) < 2 {
		return
	}

	prc.Relation = int(data[0])
	prc.TargetPlayer = int(data[1])
}

// Encode returns the raw block data bytes (without the 2-byte block header).
func (prc *PlayersRelationChangeBlock) Encode() []byte {
	data := make([]byte, 2)
	data[0] = byte(prc.Relation)
	data[1] = byte(prc.TargetPlayer)
	return data
}

// RelationName returns the human-readable name of the relation
func (prc *PlayersRelationChangeBlock) RelationName() string {
	switch prc.Relation {
	case RelationFriend:
		return "Friend"
	case RelationNeutral:
		return "Neutral"
	case RelationEnemy:
		return "Enemy"
	default:
		return "Unknown"
	}
}
