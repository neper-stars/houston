package blocks

import (
	"errors"

	"github.com/neper-stars/houston/encoding"
)

var ErrInvalidPlayerBlock = errors.New("invalid player block")

// Primary Race Traits (PRT)
const (
	PRTHyperExpansion       = 0 // HE
	PRTSuperStealth         = 1 // SS
	PRTWarMonger            = 2 // WM
	PRTClaimAdjuster        = 3 // CA
	PRTInnerStrength        = 4 // IS
	PRTSpaceDemolition      = 5 // SD
	PRTPacketPhysics        = 6 // PP
	PRTInterstellarTraveler = 7 // IT
	PRTAlternateReality     = 8 // AR
	PRTJackOfAllTrades      = 9 // JOAT
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
	LRTImprovedFuelEfficiency = 1 << 0  // IFE
	LRTTotalTerraforming      = 1 << 1  // TT
	LRTAdvancedRemoteMining   = 1 << 2  // ARM
	LRTImprovedStarbases      = 1 << 3  // ISB
	LRTGeneralizedResearch    = 1 << 4  // GR
	LRTUltimateRecycling      = 1 << 5  // UR
	LRTMineralAlchemy         = 1 << 6  // MA
	LRTNoRamScoopEngines      = 1 << 7  // NRSE
	LRTCheapEngines           = 1 << 8  // CE
	LRTOnlyBasicRemoteMining  = 1 << 9  // OBRM
	LRTNoAdvancedScanners     = 1 << 10 // NAS
	LRTLowStartingPopulation  = 1 << 11 // LSP
	LRTBleedingEdgeTechnology = 1 << 12 // BET
	LRTRegeneratingShields    = 1 << 13 // RS
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

// TechPoints holds accumulated research points for each technology field.
// These track progress toward the next tech level in each field.
type TechPoints struct {
	Energy       uint32
	Weapons      uint32
	Propulsion   uint32
	Construction uint32
	Electronics  uint32
	Biotech      uint32
}

// Habitability holds the three habitability parameters
type Habitability struct {
	GravityCenter     int // Base 65 (0.12g), 255 = immune
	TemperatureCenter int // Base 35 (-200°C), 255 = immune
	RadiationCenter   int // Base 0, 255 = immune
	GravityLow        int
	TemperatureLow    int
	RadiationLow      int
	GravityHigh       int
	TemperatureHigh   int
	RadiationHigh     int
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

// PlayerFlags holds player state flags (wFlags at offset 0x54)
type PlayerFlags struct {
	Dead     bool // Player has been eliminated
	Crippled bool // Player is crippled
	Cheater  bool // Cheater flag
	Learned  bool // Unknown purpose
	Hacker   bool // Hacker flag
}

// ZipProdQueueItem represents a single item in a zip production template.
// All zip prod items are Auto Build items (IDs 0-6):
//   - 0: AutoMines
//   - 1: AutoFactories
//   - 2: AutoDefenses
//   - 3: AutoAlchemy
//   - 4: AutoMinTerraform
//   - 5: AutoMaxTerraform
//   - 6: AutoPackets
type ZipProdQueueItem struct {
	ItemType uint16 // Auto build item ID (0-6)
	Quantity uint16 // Build limit (0 = unlimited for Alchemy)
}

// ZipProdQueue represents a "zip production" template (quick production orders).
// These are production templates that can be quickly applied to any planet.
// The default template (Q1) is auto-applied to newly conquered planets.
// Players can define up to 4 templates (Default + 3 custom named ones).
//
// Binary format (variable size in .x order file, 26 bytes with padding in player block):
//   - Byte 0: Flags (purpose TBD, usually 0x00)
//   - Byte 1: Number of items (can exceed 7 since items can repeat)
//   - Bytes 2+: Item data, 2 bytes per item as uint16 LE:
//   - Low 6 bits: Item ID (0-6 for auto-build items)
//   - High 10 bits: Count (0-1023, max settable in GUI is 1020)
//
// NOTE: The same item type can appear multiple times with different counts
// (e.g., AutoMines(1) followed by AutoMines(2)). Maximum 12 items per queue.
// In the GUI, zip queues are populated by importing from a planet's production queue.
//
// This format differs from ProductionQueueBlock which uses (ItemId << 10) | Count.
// ZipProd uses the reverse: (Count << 6) | ItemId.
//
// The zip prod data also appears in SaveAndSubmitBlockType (46) in .x order files,
// which is the source before being copied into the player block.
// See constants.go for SaveAndSubmitBlockType documentation.
type ZipProdQueue struct {
	Items    []ZipProdQueueItem // Production items in the template
	Flags    byte               // Flags byte (purpose TBD)
	RawBytes []byte             // Raw bytes - preserved for round-trip encoding
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
	TechProgress TechPoints // Accumulated research points toward next level
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
	ResearchPercentage     int    // Default research budget percentage
	CurrentResearchField   int    // Current research priority
	NextResearchField      int    // Next research priority
	ResearchPointsPrevYear uint32 // Research points spent in the previous year (bytes 58-61)

	// Mystery Trader items owned (bitmask, always 0 in race files)
	MTItems uint16

	// Player state flags (offset 0x54, bytes 84-85)
	Flags PlayerFlags

	// Default zip production queue template (offset 0x56, bytes 86-111)
	// This is auto-applied to newly conquered planets
	ZipProdDefault ZipProdQueue

	// PasswordHash stores the hashed password for encoding.
	// When decoding, use HashedPass() to read from raw data.
	// When encoding, set this field before calling Encode().
	PasswordHash uint32
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
		// These track accumulated research points toward the next tech level
		p.TechProgress.Energy = encoding.Read32(p.Decrypted, 32)
		p.TechProgress.Weapons = encoding.Read32(p.Decrypted, 36)
		p.TechProgress.Propulsion = encoding.Read32(p.Decrypted, 40)
		p.TechProgress.Construction = encoding.Read32(p.Decrypted, 44)
		p.TechProgress.Electronics = encoding.Read32(p.Decrypted, 48)
		p.TechProgress.Biotech = encoding.Read32(p.Decrypted, 52)

		// Research settings (bytes 56-57, FDB 48-49)
		p.ResearchPercentage = int(p.Decrypted[56])
		p.CurrentResearchField = int(p.Decrypted[57] >> 4)
		p.NextResearchField = int(p.Decrypted[57] & 0x0F)

		// Research points spent in previous year (bytes 58-61, FDB 50-53)
		p.ResearchPointsPrevYear = encoding.Read32(p.Decrypted, 58)

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

		// Player state flags (bytes 84-85, FDB 76-77)
		flags := encoding.Read16(p.Decrypted, 84)
		p.Flags.Dead = (flags & 0x01) != 0
		p.Flags.Crippled = (flags & 0x02) != 0
		p.Flags.Cheater = (flags & 0x04) != 0
		p.Flags.Learned = (flags & 0x08) != 0
		p.Flags.Hacker = (flags & 0x10) != 0

		// Default zip production queue (bytes 86-111, FDB 78-103)
		// This is the production template auto-applied to newly conquered planets
		p.ZipProdDefault.RawBytes = make([]byte, 26)
		copy(p.ZipProdDefault.RawBytes, p.Decrypted[86:112])
		p.ZipProdDefault.Flags = p.Decrypted[86]
		itemCount := int(p.Decrypted[87])
		p.ZipProdDefault.Items = make([]ZipProdQueueItem, 0, itemCount)
		for i := 0; i < itemCount && 88+i*2+1 < len(p.Decrypted); i++ {
			val := encoding.Read16(p.Decrypted, 88+i*2)
			// Format: (Count << 6) | ItemId
			// Low 6 bits: Item ID, High 10 bits: Count
			p.ZipProdDefault.Items = append(p.ZipProdDefault.Items, ZipProdQueueItem{
				ItemType: val & 0x3F,
				Quantity: val >> 6,
			})
		}

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
		index++ //nolint:ineffassign // documents binary format position for future extension
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

// Encode returns the raw block data bytes (without the 2-byte block header).
// This encodes all PlayerBlock fields back to the binary format.
func (p *PlayerBlock) Encode() ([]byte, error) {
	singularEncoded := encoding.EncodeStarsString(p.NameSingular)
	pluralEncoded := encoding.EncodeStarsString(p.NamePlural)

	// Calculate total size
	var totalSize int
	if p.FullDataFlag {
		// 8 header + 104 full data + 1 relations length + relations + names
		totalSize = 8 + 104 + 1 + len(p.PlayerRelations) + len(singularEncoded) + len(pluralEncoded)
	} else {
		// 8 header + names
		totalSize = 8 + len(singularEncoded) + len(pluralEncoded)
	}
	// Add padding byte if plural name is empty (for 16-bit alignment)
	if len(pluralEncoded) == 1 {
		totalSize++
	}

	data := make([]byte, totalSize)
	index := 0

	// Byte 0: Player number
	data[index] = byte(p.PlayerNumber)
	index++

	// Byte 1: Ship design count
	data[index] = byte(p.ShipDesignCount)
	index++

	// Bytes 2-3: Planets (10 bits)
	data[index] = byte(p.Planets & 0xFF)
	index++
	data[index] = byte((p.Planets >> 8) & 0x03)
	index++

	// Bytes 4-5: Fleets (10 bits) + Starbase design count (4 bits in high nibble of byte 5)
	data[index] = byte(p.Fleets & 0xFF)
	index++
	data[index] = byte((p.Fleets>>8)&0x03) | byte((p.StarbaseDesignCount&0x0F)<<4)
	index++

	// Byte 6: Logo (5 bits) << 3 | FullDataFlag (bit 2) | fixed bits (bits 0-1 = 3)
	byte6 := byte((p.Logo&0x1F)<<3) | 0x03
	if p.FullDataFlag {
		byte6 |= 0x04
	}
	data[index] = byte6
	index++

	// Byte 7: AI settings
	data[index] = p.Byte7
	index++

	if p.FullDataFlag {
		// Full data section (104 bytes starting at index 8)
		fullDataStart := index

		// Bytes 8-9: Homeworld
		encoding.Write16(data, fullDataStart, uint16(p.Homeworld))
		// Bytes 10-11: Rank
		encoding.Write16(data, fullDataStart+2, uint16(p.Rank))
		// Bytes 12-15: Password hash
		encoding.Write32(data, fullDataStart+4, p.PasswordHash)

		// Bytes 16-24: Habitability (9 bytes)
		data[fullDataStart+8] = byte(p.Hab.GravityCenter)
		data[fullDataStart+9] = byte(p.Hab.TemperatureCenter)
		data[fullDataStart+10] = byte(p.Hab.RadiationCenter)
		data[fullDataStart+11] = byte(p.Hab.GravityLow)
		data[fullDataStart+12] = byte(p.Hab.TemperatureLow)
		data[fullDataStart+13] = byte(p.Hab.RadiationLow)
		data[fullDataStart+14] = byte(p.Hab.GravityHigh)
		data[fullDataStart+15] = byte(p.Hab.TemperatureHigh)
		data[fullDataStart+16] = byte(p.Hab.RadiationHigh)

		// Byte 25: Growth rate
		data[fullDataStart+17] = byte(p.GrowthRate)

		// Bytes 26-31: Tech levels
		data[fullDataStart+18] = byte(p.Tech.Energy)
		data[fullDataStart+19] = byte(p.Tech.Weapons)
		data[fullDataStart+20] = byte(p.Tech.Propulsion)
		data[fullDataStart+21] = byte(p.Tech.Construction)
		data[fullDataStart+22] = byte(p.Tech.Electronics)
		data[fullDataStart+23] = byte(p.Tech.Biotech)

		// Bytes 32-55: Tech points (24 bytes, 4 bytes per field)
		encoding.Write32(data, fullDataStart+24, p.TechProgress.Energy)
		encoding.Write32(data, fullDataStart+28, p.TechProgress.Weapons)
		encoding.Write32(data, fullDataStart+32, p.TechProgress.Propulsion)
		encoding.Write32(data, fullDataStart+36, p.TechProgress.Construction)
		encoding.Write32(data, fullDataStart+40, p.TechProgress.Electronics)
		encoding.Write32(data, fullDataStart+44, p.TechProgress.Biotech)

		// Bytes 56-57: Research settings
		data[fullDataStart+48] = byte(p.ResearchPercentage)
		data[fullDataStart+49] = byte((p.CurrentResearchField&0x0F)<<4) | byte(p.NextResearchField&0x0F)

		// Bytes 58-61: Research points spent in previous year
		encoding.Write32(data, fullDataStart+50, p.ResearchPointsPrevYear)

		// Bytes 62-68: Production settings
		data[fullDataStart+54] = byte(p.Production.ResourcePerColonist)
		data[fullDataStart+55] = byte(p.Production.FactoryProduction)
		data[fullDataStart+56] = byte(p.Production.FactoryCost)
		data[fullDataStart+57] = byte(p.Production.FactoriesOperate)
		data[fullDataStart+58] = byte(p.Production.MineProduction)
		data[fullDataStart+59] = byte(p.Production.MineCost)
		data[fullDataStart+60] = byte(p.Production.MinesOperate)

		// Byte 69: Leftover points
		data[fullDataStart+61] = byte(p.SpendLeftoverPoints)

		// Bytes 70-75: Research costs
		data[fullDataStart+62] = byte(p.ResearchCost.Energy)
		data[fullDataStart+63] = byte(p.ResearchCost.Weapons)
		data[fullDataStart+64] = byte(p.ResearchCost.Propulsion)
		data[fullDataStart+65] = byte(p.ResearchCost.Construction)
		data[fullDataStart+66] = byte(p.ResearchCost.Electronics)
		data[fullDataStart+67] = byte(p.ResearchCost.Biotech)

		// Bytes 76-77: PRT
		data[fullDataStart+68] = byte(p.PRT)
		data[fullDataStart+69] = 0

		// Bytes 78-79: LRT
		encoding.Write16(data, fullDataStart+70, p.LRT)

		// Byte 80: Reserved
		data[fullDataStart+72] = 0

		// Byte 81: Checkboxes
		var checkBoxes byte
		if p.ExpensiveTechStartsAt3 {
			checkBoxes |= 0x20 // bit 5
		}
		if p.FactoriesCost1LessGerm {
			checkBoxes |= 0x80 // bit 7
		}
		data[fullDataStart+73] = checkBoxes

		// Bytes 82-83: MT Items
		encoding.Write16(data, fullDataStart+74, p.MTItems)

		// Bytes 84-85: Player state flags
		var flags uint16
		if p.Flags.Dead {
			flags |= 0x01
		}
		if p.Flags.Crippled {
			flags |= 0x02
		}
		if p.Flags.Cheater {
			flags |= 0x04
		}
		if p.Flags.Learned {
			flags |= 0x08
		}
		if p.Flags.Hacker {
			flags |= 0x10
		}
		encoding.Write16(data, fullDataStart+76, flags)

		// Bytes 86-111: Default zip production queue (26 bytes)
		switch {
		case len(p.ZipProdDefault.Items) > 0:
			// Encode from parsed items
			data[fullDataStart+78] = p.ZipProdDefault.Flags
			data[fullDataStart+79] = byte(len(p.ZipProdDefault.Items))
			for i, item := range p.ZipProdDefault.Items {
				if fullDataStart+80+i*2+1 < len(data) {
					// Format: (Count << 6) | ItemId
					val := (item.Quantity << 6) | (item.ItemType & 0x3F)
					encoding.Write16(data, fullDataStart+80+i*2, val)
				}
			}
		case len(p.ZipProdDefault.RawBytes) >= 26:
			// Fallback to raw bytes if items not parsed
			copy(data[fullDataStart+78:fullDataStart+104], p.ZipProdDefault.RawBytes)
		case len(p.FullDataBytes) >= 96:
			// Fallback to FullDataBytes if RawBytes not set
			copy(data[fullDataStart+78:fullDataStart+104], p.FullDataBytes[78:104])
		}

		index = fullDataStart + 104

		// Player relations
		data[index] = byte(len(p.PlayerRelations))
		index++
		copy(data[index:], p.PlayerRelations)
		index += len(p.PlayerRelations)
	}

	// Singular name
	copy(data[index:], singularEncoded)
	index += len(singularEncoded)

	// Plural name
	copy(data[index:], pluralEncoded)
	index += len(pluralEncoded)

	// Padding if plural name is empty
	if len(pluralEncoded) == 1 {
		data[index] = 0
		index++
	}

	return data[:index], nil
}
