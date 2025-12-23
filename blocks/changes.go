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
// Structure not fully documented - preserves raw data for analysis
type PlanetChangeBlock struct {
	GenericBlock
}

// NewPlanetChangeBlock creates a PlanetChangeBlock from a GenericBlock
func NewPlanetChangeBlock(b GenericBlock) *PlanetChangeBlock {
	return &PlanetChangeBlock{GenericBlock: b}
}

// ChangePasswordBlock represents a password change request (Type 36)
// Structure not fully documented - preserves raw data for analysis
type ChangePasswordBlock struct {
	GenericBlock
}

// NewChangePasswordBlock creates a ChangePasswordBlock from a GenericBlock
func NewChangePasswordBlock(b GenericBlock) *ChangePasswordBlock {
	return &ChangePasswordBlock{GenericBlock: b}
}

// PlayersRelationChangeBlock represents diplomatic relation changes (Type 38)
// Structure not fully documented - preserves raw data for analysis
type PlayersRelationChangeBlock struct {
	GenericBlock
}

// NewPlayersRelationChangeBlock creates a PlayersRelationChangeBlock from a GenericBlock
func NewPlayersRelationChangeBlock(b GenericBlock) *PlayersRelationChangeBlock {
	return &PlayersRelationChangeBlock{GenericBlock: b}
}
