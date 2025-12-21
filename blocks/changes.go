package blocks

// ResearchChangeBlock represents a research priority change (Type 34)
// Structure not fully documented - preserves raw data for analysis
type ResearchChangeBlock struct {
	GenericBlock
}

// NewResearchChangeBlock creates a ResearchChangeBlock from a GenericBlock
func NewResearchChangeBlock(b GenericBlock) *ResearchChangeBlock {
	return &ResearchChangeBlock{GenericBlock: b}
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
