package store

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/data"
)

// PlayerStatus represents the AI/Human status of a player.
type PlayerStatus int

const (
	// PlayerStatusHuman is an active human player.
	PlayerStatusHuman PlayerStatus = iota
	// PlayerStatusInactive is a human player marked as inactive.
	PlayerStatusInactive
	// PlayerStatusAI is an AI-controlled player.
	PlayerStatusAI
)

// AIExpertType represents the AI expert type (which PRT AI to use).
type AIExpertType int

const (
	AIExpertHE AIExpertType = iota // Robotoids (HE AI)
	AIExpertSS                     // Turindromes (SS AI)
	AIExpertIS                     // Automitrons (IS AI)
	AIExpertCA                     // Rototills (CA AI)
	AIExpertPP                     // Cybertrons (PP AI)
	AIExpertAR                     // Macinti (AR AI)
)

// String returns the human-readable name for the AI expert type.
func (a AIExpertType) String() string {
	names := []string{"HE (Robotoids)", "SS (Turindromes)", "IS (Automitrons)",
		"CA (Rototills)", "PP (Cybertrons)", "AR (Macinti)"}
	if int(a) >= 0 && int(a) < len(names) {
		return names[a]
	}
	return "Unknown"
}

// ShortName returns the short PRT code for the AI expert type.
func (a AIExpertType) ShortName() string {
	names := []string{"HE", "SS", "IS", "CA", "PP", "AR"}
	if int(a) >= 0 && int(a) < len(names) {
		return names[a]
	}
	return "Unknown"
}

// FullName returns the full PRT name for the AI expert type.
func (a AIExpertType) FullName() string {
	names := []string{
		"Hyper Expansion",
		"Super Stealth",
		"Inner Strength",
		"Claim Adjuster",
		"Packet Physics",
		"Alternate Reality",
	}
	if int(a) >= 0 && int(a) < len(names) {
		return names[a]
	}
	return "Unknown"
}

// Description returns a brief description of the AI expert type behavior.
func (a AIExpertType) Description() string {
	descriptions := []string{
		"Aggressive expansion AI. Expects IFE/OBRM.",
		"Stealth-focused AI. Expects IFE/ARM.",
		"Defensive/growth AI. Expects OBRM/NAS, not IFE.",
		"Terraforming AI. Expects TT/OBRM/NAS, not IFE.",
		"Mass driver AI. Expects IFE/TT/OBRM/NAS. May struggle with non-PP races.",
		"Orbital-based AI. Expects IFE/TT/ARM/ISB.",
	}
	if int(a) >= 0 && int(a) < len(descriptions) {
		return descriptions[a]
	}
	return "Unknown"
}

// ParseAIExpertType parses a string into an AIExpertType.
// Accepts: "HE", "SS", "IS", "CA", "PP", "AR" (case-insensitive)
// Also accepts full names: "Robotoids", "Turindromes", etc.
func ParseAIExpertType(s string) (AIExpertType, error) {
	switch s {
	case "HE", "he", "Robotoids", "robotoids":
		return AIExpertHE, nil
	case "SS", "ss", "Turindromes", "turindromes":
		return AIExpertSS, nil
	case "IS", "is", "Automitrons", "automitrons":
		return AIExpertIS, nil
	case "CA", "ca", "Rototills", "rototills":
		return AIExpertCA, nil
	case "PP", "pp", "Cybertrons", "cybertrons":
		return AIExpertPP, nil
	case "AR", "ar", "Macinti", "macinti":
		return AIExpertAR, nil
	default:
		return 0, fmt.Errorf("unknown AI expert type: %s (valid: HE, SS, IS, CA, PP, AR)", s)
	}
}

// AllAIExpertTypes returns all valid AI expert types.
func AllAIExpertTypes() []AIExpertType {
	return []AIExpertType{AIExpertHE, AIExpertSS, AIExpertIS, AIExpertCA, AIExpertPP, AIExpertAR}
}

// TechLevels is an alias for data.TechRequirements for backward compatibility.
type TechLevels = data.TechRequirements

// StoredScore contains score data read from PlayerScoresBlock (block type 45).
// This is the authoritative score as calculated by the game itself.
type StoredScore struct {
	Score        int   // Player's total score
	Resources    int64 // Resources available (displayed in score screen)
	Planets      int   // Number of planets owned
	Starbases    int   // Number of starbases
	UnarmedShips int   // Number of unarmed ships
	EscortShips  int   // Number of escort ships
	CapitalShips int   // Number of capital ships
	TechLevels   int   // Raw sum of tech levels (not the tiered score)
	Rank         int   // Player's rank position
	Turn         int   // Turn number this score was recorded
}

// PlayerEntity represents a player in the game.
type PlayerEntity struct {
	meta EntityMeta

	// Core identification
	PlayerNumber int

	// Display info
	NamePlural   string
	NameSingular string
	Logo         int

	// Counts
	ShipDesignCount     int
	StarbaseDesignCount int
	PlanetCount         int
	FleetCount          int

	// Player state (from header, bytes 0x08-0x0B)
	HomePlanetID int // idPlanetHome: ID of player's homeworld planet
	Rank         int // Player ranking position (1=1st, 2=2nd, etc.)
	//               // NOTE: Decompiled source names this "wScore" but it's actually Rank in the UI

	// Race info (if full data available)
	GrowthRate  int
	HasFullData bool
	Tech        TechLevels // Current tech levels
	PRT         int        // Primary Race Trait (0-9, see blocks.PRT* constants)
	LRT         uint16     // Lesser Race Traits bitmask (see blocks.LRT* constants)

	// Production settings (economy parameters)
	Production blocks.ProductionSettings

	// Habitability settings (environment preferences)
	Hab blocks.Habitability

	// Diplomatic relations (from file owner's perspective)
	PlayerRelations []byte

	// Score data from PlayerScoresBlock (authoritative game-calculated values)
	// This is nil if no PlayerScoresBlock was found for this player.
	StoredScore *StoredScore

	// Raw block (preserved for re-encoding)
	playerBlock *blocks.PlayerBlock
}

// Meta returns the entity metadata.
func (p *PlayerEntity) Meta() *EntityMeta {
	return &p.meta
}

// RawBlocks returns the original blocks.
func (p *PlayerEntity) RawBlocks() []blocks.Block {
	if p.playerBlock != nil {
		return []blocks.Block{*p.playerBlock}
	}
	return nil
}

// SetDirty marks the entity as modified.
func (p *PlayerEntity) SetDirty() {
	p.meta.Dirty = true
}

// GetRelationTo returns the relation to another player.
// Returns: 0=Neutral, 1=Friend, 2=Enemy, -1=invalid
func (p *PlayerEntity) GetRelationTo(playerIndex int) int {
	if playerIndex < 0 || playerIndex >= len(p.PlayerRelations) {
		return 0 // Default to Neutral
	}
	return int(p.PlayerRelations[playerIndex])
}

// HasLRT returns true if the player has the specified Lesser Race Trait.
// The lrtBitmask should be one of the blocks.LRT* constants.
func (p *PlayerEntity) HasLRT(lrtBitmask uint16) bool {
	return (p.LRT & lrtBitmask) != 0
}

// Byte7 values for player status.
// These are derived from TotalHost's StarsAI.pl implementation.
const (
	byte7HumanActive      byte = 1   // Active human player
	byte7HumanWasInactive byte = 225 // Human player that was previously inactive
	byte7HumanInactive    byte = 227 // Human (Inactive) / Expansion player
	byte7AIExpertHE       byte = 15  // HE AI Expert
	byte7AIExpertSS       byte = 47  // SS AI Expert
	byte7AIExpertIS       byte = 79  // IS AI Expert
	byte7AIExpertCA       byte = 111 // CA AI Expert
	byte7AIExpertPP       byte = 143 // PP AI Expert
	byte7AIExpertAR       byte = 175 // AR AI Expert
)

// Standard AI password hash: "viewai" hashed = 0x094DABEEb (little-endian: 238, 171, 77, 9)
var aiPasswordHash = uint32(0x094DABEE)

// GetStatus returns the current player status (Human, Inactive, or AI).
func (p *PlayerEntity) GetStatus() PlayerStatus {
	if p.playerBlock == nil {
		return PlayerStatusHuman
	}
	b7 := p.playerBlock.Byte7
	if b7 == byte7HumanActive || b7 == byte7HumanWasInactive {
		return PlayerStatusHuman
	}
	if b7 == byte7HumanInactive {
		return PlayerStatusInactive
	}
	// Any other value with AI bit set is AI
	if (b7>>1)&0x01 == 1 {
		return PlayerStatusAI
	}
	return PlayerStatusHuman
}

// GetStatusString returns a human-readable status string.
func (p *PlayerEntity) GetStatusString() string {
	switch p.GetStatus() {
	case PlayerStatusHuman:
		return "Human"
	case PlayerStatusInactive:
		return "Human (Inactive)"
	case PlayerStatusAI:
		return fmt.Sprintf("AI (%s)", p.GetAIExpertType().ShortName())
	default:
		return "Unknown"
	}
}

// GetAIExpertType returns the AI expert type if the player is AI-controlled.
func (p *PlayerEntity) GetAIExpertType() AIExpertType {
	if p.playerBlock == nil {
		return AIExpertHE
	}
	// Bits 5-7 of byte 7 define the AI race
	aiRace := int((p.playerBlock.Byte7 >> 5) & 0x07)
	if aiRace < 6 {
		return AIExpertType(aiRace)
	}
	return AIExpertHE // Default
}

// ChangeToHuman changes the player to human control.
// If the player was inactive, the password bits are flipped.
// If the player was AI, the password is cleared.
func (p *PlayerEntity) ChangeToHuman() error {
	if p.playerBlock == nil {
		return fmt.Errorf("no player block available")
	}

	currentStatus := p.GetStatus()

	switch currentStatus {
	case PlayerStatusHuman:
		// Already human, nothing to do
		return nil

	case PlayerStatusInactive:
		// Changing from Inactive to Human: flip password bits
		p.playerBlock.Byte7 = byte7HumanWasInactive
		// Flip the password bits
		currentHash := p.playerBlock.HashedPass().Uint32()
		p.playerBlock.PasswordHash = ^currentHash

	case PlayerStatusAI:
		// Changing from AI to Human: clear password
		p.playerBlock.Byte7 = byte7HumanWasInactive
		p.playerBlock.PasswordHash = 0
	}

	p.SetDirty()
	return nil
}

// ChangeToInactive changes the player to Human (Inactive).
// The password bits are flipped for inactive players.
func (p *PlayerEntity) ChangeToInactive() error {
	if p.playerBlock == nil {
		return fmt.Errorf("no player block available")
	}

	currentStatus := p.GetStatus()

	switch currentStatus {
	case PlayerStatusInactive:
		// Already inactive, nothing to do
		return nil

	case PlayerStatusHuman:
		// Changing from Human to Inactive: flip password bits
		p.playerBlock.Byte7 = byte7HumanInactive
		currentHash := p.playerBlock.HashedPass().Uint32()
		p.playerBlock.PasswordHash = ^currentHash

	case PlayerStatusAI:
		// Changing from AI to Inactive: set inverted blank password
		p.playerBlock.Byte7 = byte7HumanInactive
		p.playerBlock.PasswordHash = 0xFFFFFFFF // Inverted blank password
	}

	p.SetDirty()
	return nil
}

// ChangeToAI changes the player to AI control with the specified expert type.
// The standard AI password "viewai" is set.
func (p *PlayerEntity) ChangeToAI(expertType AIExpertType) error {
	if p.playerBlock == nil {
		return fmt.Errorf("no player block available")
	}

	// Determine the byte7 value for this AI type
	var byte7Value byte
	switch expertType {
	case AIExpertHE:
		byte7Value = byte7AIExpertHE
	case AIExpertSS:
		byte7Value = byte7AIExpertSS
	case AIExpertIS:
		byte7Value = byte7AIExpertIS
	case AIExpertCA:
		byte7Value = byte7AIExpertCA
	case AIExpertPP:
		byte7Value = byte7AIExpertPP
	case AIExpertAR:
		byte7Value = byte7AIExpertAR
	default:
		return fmt.Errorf("invalid AI expert type: %d", expertType)
	}

	p.playerBlock.Byte7 = byte7Value
	p.playerBlock.PasswordHash = aiPasswordHash

	p.SetDirty()
	return nil
}

// newPlayerEntityFromBlock creates a PlayerEntity from a PlayerBlock.
func newPlayerEntityFromBlock(pb *blocks.PlayerBlock, source *FileSource) *PlayerEntity {
	entity := &PlayerEntity{
		meta: EntityMeta{
			Key: EntityKey{
				Type:   EntityTypePlayer,
				Owner:  pb.PlayerNumber,
				Number: pb.PlayerNumber,
			},
			BestSource: source,
			Quality:    QualityFull, // Player blocks are always full quality
			Turn:       source.Turn,
		},
		PlayerNumber:        pb.PlayerNumber,
		NamePlural:          pb.NamePlural,
		NameSingular:        pb.NameSingular,
		Logo:                pb.Logo,
		ShipDesignCount:     pb.ShipDesignCount,
		StarbaseDesignCount: pb.StarbaseDesignCount,
		PlanetCount:         pb.Planets,
		FleetCount:          pb.Fleets,
		HomePlanetID:        pb.HomePlanetID,
		Rank:                pb.Rank,
		GrowthRate:          pb.GrowthRate,
		HasFullData:         pb.FullDataFlag,
		Tech: TechLevels{
			Energy:       pb.Tech.Energy,
			Weapons:      pb.Tech.Weapons,
			Propulsion:   pb.Tech.Propulsion,
			Construction: pb.Tech.Construction,
			Electronics:  pb.Tech.Electronics,
			Biotech:      pb.Tech.Biotech,
		},
		PRT:             pb.PRT,
		LRT:             pb.LRT,
		Production:      pb.Production,
		Hab:             pb.Hab,
		PlayerRelations: pb.PlayerRelations,
		playerBlock:     pb,
	}
	entity.meta.AddSource(source)
	return entity
}
