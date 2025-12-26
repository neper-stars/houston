package blocks

import (
	"errors"

	"github.com/neper-stars/houston/encoding"
)

var ErrInvalidPlayerBlock = errors.New("invalid player block")

// Primary Race Traits (PRT)
const (
	PRTHyperExpansion       = 0  // HE
	PRTSuperStealth         = 1  // SS
	PRTWarMonger            = 2  // WM
	PRTClaimAdjuster        = 3  // CA
	PRTInnerStrength        = 4  // IS
	PRTSpaceDemolition      = 5  // SD
	PRTPacketPhysics        = 6  // PP
	PRTInterstellarTraveler = 7  // IT
	PRTAlternateReality     = 8  // AR
	PRTJackOfAllTrades      = 9  // JOAT
)

// PRTName returns the short name for a PRT
func PRTName(prt int) string {
	names := []string{"HE", "SS", "WM", "CA", "IS", "SD", "PP", "IT", "AR", "JOAT"}
	if prt >= 0 && prt < len(names) {
		return names[prt]
	}
	return "Unknown"
}

// PRTFullName returns the full name for a PRT
func PRTFullName(prt int) string {
	names := []string{
		"Hyper Expansion", "Super Stealth", "War Monger", "Claim Adjuster",
		"Inner Strength", "Space Demolition", "Packet Physics",
		"Interstellar Traveler", "Alternate Reality", "Jack of All Trades",
	}
	if prt >= 0 && prt < len(names) {
		return names[prt]
	}
	return "Unknown"
}

// Lesser Race Traits (LRT) bitmask values
const (
	LRTImprovedFuelEfficiency  = 1 << 0  // IFE
	LRTTotalTerraforming       = 1 << 1  // TT
	LRTAdvancedRemoteMining    = 1 << 2  // ARM
	LRTImprovedStarbases       = 1 << 3  // ISB
	LRTGeneralizedResearch     = 1 << 4  // GR
	LRTUltimateRecycling       = 1 << 5  // UR
	LRTMineralAlchemy          = 1 << 6  // MA
	LRTNoRamScoopEngines       = 1 << 7  // NRSE
	LRTCheapEngines            = 1 << 8  // CE
	LRTOnlyBasicRemoteMining   = 1 << 9  // OBRM
	LRTNoAdvancedScanners      = 1 << 10 // NAS
	LRTLowStartingPopulation   = 1 << 11 // LSP
	LRTBleedingEdgeTechnology  = 1 << 12 // BET
	LRTRegeneratingShields     = 1 << 13 // RS
)

// LRTNames returns the short names for all set LRT bits
func LRTNames(lrt uint16) []string {
	allLRTs := []struct {
		bit  uint16
		name string
	}{
		{LRTImprovedFuelEfficiency, "IFE"},
		{LRTTotalTerraforming, "TT"},
		{LRTAdvancedRemoteMining, "ARM"},
		{LRTImprovedStarbases, "ISB"},
		{LRTGeneralizedResearch, "GR"},
		{LRTUltimateRecycling, "UR"},
		{LRTMineralAlchemy, "MA"},
		{LRTNoRamScoopEngines, "NRSE"},
		{LRTCheapEngines, "CE"},
		{LRTOnlyBasicRemoteMining, "OBRM"},
		{LRTNoAdvancedScanners, "NAS"},
		{LRTLowStartingPopulation, "LSP"},
		{LRTBleedingEdgeTechnology, "BET"},
		{LRTRegeneratingShields, "RS"},
	}
	var names []string
	for _, l := range allLRTs {
		if (lrt & l.bit) != 0 {
			names = append(names, l.name)
		}
	}
	return names
}

// AI skill levels
const (
	AISkillEasy     = 0
	AISkillStandard = 1
	AISkillHarder   = 2
	AISkillExpert   = 3
)

// Research cost modifiers
const (
	ResearchCostExpensive = 0 // +75%
	ResearchCostNormal    = 1 // 0%
	ResearchCostCheap     = 2 // -50%
)

// Leftover points spending options
const (
	SpendLeftoverSurfaceMinerals = 0
	SpendLeftoverMining          = 1
	SpendLeftoverDefenses        = 2
	SpendLeftoverFactories       = 3
	SpendLeftoverMineralAlchemy  = 4
	SpendLeftoverResearch        = 5
)

// HashedPass is a 4-byte clump
type HashedPass []byte

func (h HashedPass) Uint32() uint32 {
	return encoding.Read32(h, 0)
}

// TechLevels holds all six technology field levels
type TechLevels struct {
	Energy       int
	Weapons      int
	Propulsion   int
	Construction int
	Electronics  int
	Biotech      int
}

// Habitability holds the three habitability parameters
type Habitability struct {
	GravityCenter    int // Base 65 (0.12g), 255 = immune
	TemperatureCenter int // Base 35 (-200°C), 255 = immune
	RadiationCenter  int // Base 0, 255 = immune
	GravityLow       int
	TemperatureLow   int
	RadiationLow     int
	GravityHigh      int
	TemperatureHigh  int
	RadiationHigh    int
}

// IsGravityImmune returns true if the race is immune to gravity
func (h *Habitability) IsGravityImmune() bool {
	return h.GravityCenter == 255
}

// IsTemperatureImmune returns true if the race is immune to temperature
func (h *Habitability) IsTemperatureImmune() bool {
	return h.TemperatureCenter == 255
}

// IsRadiationImmune returns true if the race is immune to radiation
func (h *Habitability) IsRadiationImmune() bool {
	return h.RadiationCenter == 255
}

// ProductionSettings holds factory and mine production parameters
type ProductionSettings struct {
	ResourcePerColonist int // Resources per 100 colonists
	FactoryProduction   int // Resources per 10 factories
	FactoryCost         int // Resources to build one factory
	FactoriesOperate    int // Factories per 10k colonists
	MineProduction      int // kT per 10 mines
	MineCost            int // Resources to build one mine
	MinesOperate        int // Mines per 10k colonists
}

// ResearchCosts holds the research cost modifier for each field
// Values: 0 = +75%, 1 = normal, 2 = -50%
type ResearchCosts struct {
	Energy       int
	Weapons      int
	Propulsion   int
	Construction int
	Electronics  int
	Biotech      int
}

// PlayerBlock represents player data (Type 6)
type PlayerBlock struct {
	GenericBlock
	Valid               bool
	PlayerNumber        int
	NamePlural          string
	NameSingular        string
	ShipDesignCount     int
	Planets             int
	Fleets              int
	StarbaseDesignCount int
	Logo                int
	FullDataFlag        bool
	Byte7               byte
	FullDataBytes       []byte
	PlayerRelations     []byte

	// AI settings (from Byte7)
	AIEnabled bool
	AISkill   int // 0=Easy, 1=Standard, 2=Harder, 3=Expert
	AIRace    int // PRT used when AI enabled

	// Race settings (from FullDataBytes when FullDataFlag is true)
	Homeworld    int // Homeworld planet ID
	Rank         int // Player rank
	Hab          Habitability
	GrowthRate   int // Max population growth rate percentage (1-20)
	Tech         TechLevels
	Production   ProductionSettings
	ResearchCost ResearchCosts
	PRT          int    // Primary Race Trait (0-9)
	LRT          uint16 // Lesser Race Traits bitmask

	// Special flags
	ExpensiveTechStartsAt3 bool // Expensive tech starts at level 3
	FactoriesCost1LessGerm bool // Factories cost 1 less germanium

	// Leftover points spending preference (0-5)
	SpendLeftoverPoints int

	// Research settings
	ResearchPercentage int // Default research budget percentage
	CurrentResearchField int // Current research priority
	NextResearchField    int // Next research priority

	// Mystery Trader items owned (bitmask, always 0 in race files)
	MTItems uint16
}

// HashedPass returns the hashed password from inside the PlayerBlock
// This can be used as a source for the GuessRacePassword function
// by doing hashed.Uint32()
func (p *PlayerBlock) HashedPass() HashedPass {
	// the hashed password is stored at offset 12
	// of the decrypted data, and it is 4 bytes long
	return []byte(p.DecryptedData()[12:16])
}

// HasPassword returns true if the player has a password set
func (p *PlayerBlock) HasPassword() bool {
	pass := p.HashedPass()
	return pass[0] != 0 || pass[1] != 0 || pass[2] != 0 || pass[3] != 0
}

// IsHumanInactive returns true if the player is set to Human(Inactive)
// When inactive, byte 7 is 227 (0xE3) and password bits are inverted
func (p *PlayerBlock) IsHumanInactive() bool {
	return p.Byte7 == 227
}

func (p *PlayerBlock) decode() error {
	// Ensure that there is enough data to decode
	if len(p.Decrypted) < 8 {
		return errors.New("unexpected player data size")
	}

	p.PlayerNumber = int(p.Decrypted[0])
	p.ShipDesignCount = int(p.Decrypted[1])
	p.Planets = int(p.Decrypted[2]) + (int(p.Decrypted[3]) & 0x03 << 8)

	if int(p.Decrypted[3])&0xFC != 0 {
		return errors.New("unexpected player values")
	}

	p.Fleets = int(p.Decrypted[4]) + (int(p.Decrypted[5]) & 0x03 << 8)
	p.StarbaseDesignCount = int(p.Decrypted[5]) >> 4

	if int(p.Decrypted[5])&0x0C != 0 {
		return errors.New("unexpected player values")
	}

	p.Logo = int(p.Decrypted[6]) >> 3
	p.FullDataFlag = (int(p.Decrypted[6]) & 0x04) != 0

	if int(p.Decrypted[6])&0x03 != 3 {
		return errors.New("unexpected player values")
	}

	p.Byte7 = p.Decrypted[7]

	// Decode AI settings from byte 7
	// Bit 0: always 1
	// Bit 1: AI enabled
	// Bits 2-3: AI skill
	// Bit 4: always 0
	// Bits 5-7: AI race (PRT)
	p.AIEnabled = (p.Byte7>>1)&0x01 == 1
	if p.AIEnabled {
		p.AISkill = int((p.Byte7 >> 2) & 0x03)
		p.AIRace = int((p.Byte7 >> 5) & 0x07)
	}

	index := 8
	if p.FullDataFlag {
		p.FullDataBytes = make([]byte, 0x68)
		copy(p.FullDataBytes, p.Decrypted[8:8+0x68])

		// Decode full race data from FullDataBytes
		// Offsets are relative to start of decrypted data (add 8 for FullDataBytes index)

		// Homeworld and rank (bytes 8-11, FDB 0-3)
		p.Homeworld = int(encoding.Read16(p.Decrypted, 8))
		p.Rank = int(encoding.Read16(p.Decrypted, 10))
		// Password at bytes 12-15 (handled by HashedPass())

		// Habitability (bytes 16-24, FDB 8-16)
		p.Hab.GravityCenter = int(p.Decrypted[16])
		p.Hab.TemperatureCenter = int(p.Decrypted[17])
		p.Hab.RadiationCenter = int(p.Decrypted[18])
		p.Hab.GravityLow = int(p.Decrypted[19])
		p.Hab.TemperatureLow = int(p.Decrypted[20])
		p.Hab.RadiationLow = int(p.Decrypted[21])
		p.Hab.GravityHigh = int(p.Decrypted[22])
		p.Hab.TemperatureHigh = int(p.Decrypted[23])
		p.Hab.RadiationHigh = int(p.Decrypted[24])

		// Growth rate (byte 25, FDB 17)
		p.GrowthRate = int(p.Decrypted[25])

		// Tech levels (bytes 26-31, FDB 18-23)
		p.Tech.Energy = int(p.Decrypted[26])
		p.Tech.Weapons = int(p.Decrypted[27])
		p.Tech.Propulsion = int(p.Decrypted[28])
		p.Tech.Construction = int(p.Decrypted[29])
		p.Tech.Electronics = int(p.Decrypted[30])
		p.Tech.Biotech = int(p.Decrypted[31])

		// Tech points since last level (bytes 32-55, 4 bytes each, FDB 24-47)
		// Skipping detailed points tracking for now

		// Research settings (bytes 56-57, FDB 48-49)
		p.ResearchPercentage = int(p.Decrypted[56])
		p.CurrentResearchField = int(p.Decrypted[57] >> 4)
		p.NextResearchField = int(p.Decrypted[57] & 0x0F)

		// Production settings (bytes 62-68, FDB 54-60)
		p.Production.ResourcePerColonist = int(p.Decrypted[62])
		p.Production.FactoryProduction = int(p.Decrypted[63])
		p.Production.FactoryCost = int(p.Decrypted[64])
		p.Production.FactoriesOperate = int(p.Decrypted[65])
		p.Production.MineProduction = int(p.Decrypted[66])
		p.Production.MineCost = int(p.Decrypted[67])
		p.Production.MinesOperate = int(p.Decrypted[68])

		// Leftover points (byte 69, FDB 61)
		p.SpendLeftoverPoints = int(p.Decrypted[69])

		// Research costs (bytes 70-75, FDB 62-67)
		p.ResearchCost.Energy = int(p.Decrypted[70])
		p.ResearchCost.Weapons = int(p.Decrypted[71])
		p.ResearchCost.Propulsion = int(p.Decrypted[72])
		p.ResearchCost.Construction = int(p.Decrypted[73])
		p.ResearchCost.Electronics = int(p.Decrypted[74])
		p.ResearchCost.Biotech = int(p.Decrypted[75])

		// PRT and LRT (bytes 76-79, FDB 68-71)
		p.PRT = int(p.Decrypted[76])
		// byte 77 is always 0
		p.LRT = encoding.Read16(p.Decrypted, 78)

		// Checkboxes (byte 81, FDB 73)
		checkBoxes := p.Decrypted[81]
		p.ExpensiveTechStartsAt3 = (checkBoxes & 0x20) != 0 // bit 5
		p.FactoriesCost1LessGerm = (checkBoxes & 0x80) != 0 // bit 7

		// MT Items (bytes 82-83, FDB 74-75)
		p.MTItems = encoding.Read16(p.Decrypted, 82)

		// Player relations
		index = 0x70
		playerRelationsLength := int(p.Decrypted[index]) & 0xFF
		p.PlayerRelations = make([]byte, playerRelationsLength)
		copy(p.PlayerRelations, p.Decrypted[index+1:index+1+playerRelationsLength])
		index += 1 + playerRelationsLength
	}

	// Decode the singular name
	singularNameLength := int(p.Decrypted[index]) & 0xFF
	nameBytesSingular := make([]byte, singularNameLength+1)
	copy(nameBytesSingular, p.Decrypted[index:index+singularNameLength+1])

	var err error
	p.NameSingular, err = encoding.DecodeStarsString(nameBytesSingular)
	if err != nil {
		return err
	}

	index += singularNameLength + 1

	// Decode plural name (if exist)
	pluralNameLength := int(p.Decrypted[index]) & 0xFF
	nameBytesPlural := make([]byte, pluralNameLength+1)
	copy(nameBytesPlural, p.Decrypted[index:index+pluralNameLength+1])

	p.NamePlural, err = encoding.DecodeStarsString(nameBytesPlural)
	if err != nil {
		return err
	}

	index += pluralNameLength + 1
	// If no plural name skip another byte because of 16-bit alignment
	if pluralNameLength == 0 {
		index++
	}
	return nil
}

func NewPlayerBlock(b GenericBlock) (*PlayerBlock, error) {
	p := &PlayerBlock{
		GenericBlock: b,
	}
	if len(b.DecryptedData()) >= 16 {
		p.Valid = true
	}

	if err := p.decode(); err != nil {
		return nil, err
	}

	return p, nil
}

// Stored relation values in M files (different from order file encoding)
const (
	StoredRelationNeutral = 0
	StoredRelationFriend  = 1
	StoredRelationEnemy   = 2
)

// GetRelationTo returns this player's diplomatic relation to another player.
// Returns the stored relation value (0=Neutral, 1=Friend, 2=Enemy).
// Relations beyond the stored array length default to Neutral (0).
// Returns -1 only if playerIndex is negative.
func (p *PlayerBlock) GetRelationTo(playerIndex int) int {
	if playerIndex < 0 {
		return -1
	}
	if len(p.PlayerRelations) == 0 || playerIndex >= len(p.PlayerRelations) {
		// Relations not explicitly stored default to Neutral
		return StoredRelationNeutral
	}
	return int(p.PlayerRelations[playerIndex])
}

// GetRelationName returns the human-readable name for a stored relation value
func GetRelationName(storedRelation int) string {
	switch storedRelation {
	case StoredRelationNeutral:
		return "Neutral"
	case StoredRelationFriend:
		return "Friend"
	case StoredRelationEnemy:
		return "Enemy"
	default:
		return "Unknown"
	}
}

// HasLRT returns true if the player has the specified Lesser Race Trait
func (p *PlayerBlock) HasLRT(lrt uint16) bool {
	return (p.LRT & lrt) != 0
}

// PRTName returns the short name of the player's Primary Race Trait
func (p *PlayerBlock) PRTName() string {
	return PRTName(p.PRT)
}

// PRTFullName returns the full name of the player's Primary Race Trait
func (p *PlayerBlock) PRTFullName() string {
	return PRTFullName(p.PRT)
}

// LRTNames returns a list of all LRT short names for this player
func (p *PlayerBlock) LRTNames() []string {
	return LRTNames(p.LRT)
}
