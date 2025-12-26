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

// PlanetChangeBlock represents planet settings changes (Type 35)
//
// Format (6 bytes):
//
//	Bytes 0-1: Planet ID (11 bits)
//	Byte 2: Flags
//	        Bit 7 (0x80): Contribute only leftover resources to research
//	        Other bits: TBD
//	Byte 3: Additional settings (TBD)
//	Bytes 4-5: Additional data (TBD)
type PlanetChangeBlock struct {
	GenericBlock

	PlanetId                  int  // Planet ID (0-2047)
	ContributeOnlyLeftover    bool // "Contribute only leftover resources to research"
	RouteDestinationPlanetId  int  // Route destination planet (if routing is set)
	Flags                     int  // Raw flags byte for analysis
}

// NewPlanetChangeBlock creates a PlanetChangeBlock from a GenericBlock
func NewPlanetChangeBlock(b GenericBlock) *PlanetChangeBlock {
	pcb := &PlanetChangeBlock{GenericBlock: b}
	pcb.decode()
	return pcb
}

func (pcb *PlanetChangeBlock) decode() {
	data := pcb.Decrypted
	if len(data) < 4 {
		return
	}

	pcb.PlanetId = int(data[0]) | ((int(data[1]) & 0x07) << 8)
	pcb.Flags = int(data[2])
	pcb.ContributeOnlyLeftover = (data[2] & 0x80) != 0

	// Bytes 3-5 contain additional settings, possibly route destination
	if len(data) >= 6 {
		// Route destination might be encoded in remaining bytes
		pcb.RouteDestinationPlanetId = int(data[3]) | (int(data[4]) << 8)
	}
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
