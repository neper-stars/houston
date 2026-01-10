// Package blocks provides Stars! file block parsers.
package blocks

import (
	"encoding/binary"

	"github.com/neper-stars/houston/data"
	"github.com/neper-stars/houston/encoding"
)

// Planet represents a single planet in the Stars! universe.
//
// Each planet has a unique ID, a name (from a predefined list of 999 names),
// and coordinates in the game universe. The coordinate system uses light-years
// as units, with the origin typically in the upper-left corner of the map.
//
// Planet coordinates in the file are stored in a compressed format:
//   - Y coordinate: absolute value (12 bits, range 0-4095)
//   - X coordinate: stored as an OFFSET from the previous planet's X position
//     (10 bits, range 0-1023). The first planet uses 1000 as the base X.
//
// This delta encoding for X coordinates saves space in the file format since
// planets are typically sorted and nearby planets have similar X values.
type Planet struct {
	// ID is the zero-based internal planet identifier (0 to PlanetCount-1).
	ID int

	// DisplayId is the one-based display identifier shown to players (1 to PlanetCount).
	DisplayId int

	// NameID is the index into the planet names table (0 to 998).
	// See data.PlanetNames for the complete list of planet names.
	NameID uint32

	// Name is the human-readable planet name resolved from NameID.
	Name string

	// Y is the absolute Y coordinate in light-years (0 to ~4095 depending on universe size).
	Y uint32

	// X is the absolute X coordinate in light-years (calculated from delta encoding).
	// The first planet's X is always 1000 + its offset value.
	X uint32
}

// PlanetsBlock represents the universe configuration and planet list (Block Type 7).
//
// This block appears once per game file and contains:
//   - Universe configuration (size, density, player count)
//   - Game settings flags (see data.GameSetting* constants)
//   - The complete list of all planets in the universe
//
// The block has a special structure: after the main block data, there are
// 4 additional bytes per planet that are NOT included in the block size.
// These trailing bytes contain the planet coordinates and name IDs.
//
// Planet Data Encoding (4 bytes per planet, little-endian uint32):
//
//	Bits 31-22 (10 bits): Planet name ID (index into planet names table)
//	Bits 21-10 (12 bits): Y coordinate (absolute)
//	Bits  9-0  (10 bits): X offset from previous planet (first planet uses base 1000)
//
// Example: If bytes are [0x45, 0x23, 0x01, 0x80], the uint32 is 0x80012345:
//   - NameID = 0x80012345 >> 22 = 512
//   - Y = (0x80012345 >> 10) & 0xFFF = 73
//   - X offset = 0x80012345 & 0x3FF = 837
type PlanetsBlock struct {
	GenericBlock

	// Valid indicates whether the block was successfully parsed.
	Valid bool

	// UniverseSize indicates the size of the game universe.
	// Values: 0=Tiny, 1=Small, 2=Medium, 3=Large, 4=Huge
	// See data.UniverseSize* constants.
	UniverseSize uint16

	// Density indicates how many planets are in the universe relative to its size.
	// Values: 0=Sparse, 1=Normal, 2=Dense, 3=Packed
	// See data.UniverseDensity* constants.
	Density uint16

	// PlayerCount is the number of players in the game (1-16).
	PlayerCount uint16

	// PlanetCount is the total number of planets in the universe.
	// This determines how many 4-byte planet entries follow the block.
	PlanetCount uint16

	// StartingDistance is the starting distance mode (mdStartDist).
	// An index indicating how far apart players' homeworlds are placed.
	// Higher values mean greater initial separation between players.
	// See reversing_notes/planets-block.md for details.
	StartingDistance uint16

	// GameSettings is a bitmask of game configuration options.
	// See data.GameSetting* constants for individual bit meanings:
	//   Bit 0: Max Minerals
	//   Bit 1: Slow Tech Advances
	//   Bit 2: Single Player (?)
	//   Bit 3: Unknown
	//   Bit 4: Computer Alliances
	//   Bit 5: Public Scores
	//   Bit 6: Accelerated BBS Play
	//   Bit 7: No Random Events
	//   Bit 8: Galaxy Clumping
	GameSettings uint16

	// Turn is the current game year/turn number.
	// In XY files (initial game state), this is always 0.
	// In M/HST files, this reflects the current turn.
	Turn uint16

	// VictoryConditions contains the 12-byte victory condition settings.
	// Each byte encodes: bit 7 (0x80) = enabled, bits 0-6 (0x7F) = threshold index.
	//
	// Index mapping (see data.VictoryIdx* constants):
	//   [0]: Owns % of planets (idx*5+20 → 20-100%)
	//   [1]: Attain tech level X (idx+8 → 8-26)
	//   [2]: In Y tech fields (idx+2 → 2-6, 2nd part of tech condition)
	//   [3]: Exceeds score (idx*1000+1000 → 1k-20k)
	//   [4]: Exceeds 2nd place by % (idx*10+20 → 20-300%)
	//   [5]: Production capacity thousands (idx*10+10 → 10-500)
	//   [6]: Owns capital ships (idx*10+10 → 10-300)
	//   [7]: Highest score after N years (idx*10+30 → 30-900)
	//   [8]: Meets N of above criteria (special: counts enabled, 1-7)
	//   [9]: Min years before winner declared (idx*10+30 → 30-500)
	//   [10-11]: Reserved
	VictoryConditions [12]byte

	// GameName is the name of the game, up to 32 characters.
	// Padded with null bytes if shorter.
	GameName string

	// Planets is the list of all planets in the universe.
	// Populated by ParsePlanetsData() after the main block is decoded.
	Planets []Planet

	// RawPlanetsData contains the original trailing planet bytes (4 bytes per planet).
	// Preserved for re-encoding when writing files.
	RawPlanetsData []byte
}

// GetPlanetCount returns the number of planets in the universe.
// This returns PlanetCount from the block header, which is available
// immediately after NewPlanetsBlock() is called.
// After ParsePlanetsData() is called, len(Planets) will match this value.
func (p *PlanetsBlock) GetPlanetCount() int {
	return int(p.PlanetCount)
}

// NewPlanetsBlock creates a PlanetsBlock from a GenericBlock.
// This parses the main block data but NOT the trailing planet coordinate data.
// Call ParsePlanetsData() separately with the trailing bytes.
func NewPlanetsBlock(b GenericBlock) *PlanetsBlock {
	block := PlanetsBlock{
		GenericBlock: b,
	}
	data := b.DecryptedData()
	if len(data) < 64 {
		// not enough data to form a specialized block
		return &block
	}

	block.Valid = true

	// Bytes 0-3: Game ID (lid) - unique identifier for this game instance
	// Bytes 4-5: Universe size
	block.UniverseSize = binary.LittleEndian.Uint16(data[4:6])

	// Bytes 6-7: Planet density
	block.Density = binary.LittleEndian.Uint16(data[6:8])

	// Bytes 8-9: Number of players
	block.PlayerCount = binary.LittleEndian.Uint16(data[8:10])

	// Bytes 10-11: Number of planets
	block.PlanetCount = binary.LittleEndian.Uint16(data[10:12])

	// Bytes 12-13: Starting distance mode (mdStartDist)
	block.StartingDistance = binary.LittleEndian.Uint16(data[12:14])
	// Bytes 14-15: fDirty - runtime flag, ignored on read (not meaningful in files)

	// Bytes 16-17: Game settings bitmask
	block.GameSettings = binary.LittleEndian.Uint16(data[16:18])

	// Bytes 18-19: Turn number (always 0 in XY files)
	block.Turn = binary.LittleEndian.Uint16(data[18:20])

	// Bytes 20-31: Victory condition settings (12 bytes)
	copy(block.VictoryConditions[:], data[20:32])

	// Bytes 32-63: Game name (32 bytes, null-padded)
	block.GameName = string(data[32:64])

	return &block
}

// ParsePlanetsData parses the trailing planet coordinate data.
//
// This data comes AFTER the main block and is NOT included in the block size.
// Each planet is encoded as 4 bytes (little-endian uint32) containing:
//   - Name ID (10 bits): Index into data.PlanetNames
//   - Y coordinate (12 bits): Absolute Y position in light-years
//   - X offset (10 bits): Delta from previous planet's X (or from base 1000)
//
// The X coordinate uses delta encoding for compression:
// - First planet: X = 1000 + offset
// - Subsequent planets: X = previous_X + offset
//
// This encoding takes advantage of the fact that planets are sorted and
// nearby planets tend to have similar X coordinates, making the deltas small.
//
// Parameters:
//   - d: Raw bytes containing 4 bytes per planet (len must be PlanetCount * 4)
func (p *PlanetsBlock) ParsePlanetsData(d []byte) {
	// Store raw data for re-encoding
	p.RawPlanetsData = make([]byte, len(d))
	copy(p.RawPlanetsData, d)

	// Base X coordinate - all planet X values are relative to this
	x := uint32(1000)

	for i := 0; i < int(p.PlanetCount); i++ {
		// Read 4 bytes as little-endian uint32
		planetData := binary.LittleEndian.Uint32(d[i*4 : (i+1)*4])

		// Extract fields from the packed 32-bit value:
		// [31:22] = nameID (10 bits, 0-1023, but only 0-998 are valid names)
		// [21:10] = y coordinate (12 bits, 0-4095)
		// [9:0]   = x offset (10 bits, 0-1023)
		nameID := planetData >> 22
		y := (planetData >> 10) & 0xFFF
		xOffset := planetData & 0x3FF

		// Accumulate X offset to get absolute position
		x += xOffset

		planet := Planet{
			ID:        i,
			DisplayId: i + 1,
			NameID:    nameID,
			Name:      data.PlanetNames[nameID],
			Y:         y,
			X:         x,
		}

		p.Planets = append(p.Planets, planet)
	}
}

// GetPlanets returns the list of all planets in the universe.
// Returns nil if ParsePlanetsData() has not been called.
func (p *PlanetsBlock) GetPlanets() []Planet {
	return p.Planets
}

// HasGameSetting checks if a specific game setting flag is enabled.
// Use with data.GameSetting* constants.
//
// Example:
//
//	if block.HasGameSetting(data.GameSettingPublicScores) {
//	    // Scores are visible to all players
//	}
func (p *PlanetsBlock) HasGameSetting(flag int) bool {
	return (int(p.GameSettings) & flag) != 0
}

// DecodedVictoryConditions contains the parsed victory condition settings
// with enabled flags and actual computed values.
type DecodedVictoryConditions struct {
	// OwnsPercentPlanets: owns X% of all planets (20-100%)
	OwnsPercentPlanetsEnabled bool
	OwnsPercentPlanetsValue   int

	// AttainTechLevel: attain tech level X in Y fields
	AttainTechLevelEnabled bool
	AttainTechLevelValue   int // Tech level (8-26)
	AttainTechInYFields    int // Number of fields (2-6)

	// ExceedScore: exceed score of X points (1k-20k)
	ExceedScoreEnabled bool
	ExceedScoreValue   int

	// ExceedSecondPlace: exceed 2nd place score by X% (20-300%)
	ExceedSecondPlaceEnabled bool
	ExceedSecondPlaceValue   int

	// ProductionCapacity: production capacity of X thousand (10-500)
	ProductionCapacityEnabled bool
	ProductionCapacityValue   int

	// OwnCapitalShips: own X capital ships (10-300)
	OwnCapitalShipsEnabled bool
	OwnCapitalShipsValue   int

	// HighestScoreYears: highest score after X years (30-900)
	HighestScoreYearsEnabled bool
	HighestScoreYearsValue   int

	// NumCriteriaMet: meet N of the above criteria (1-7)
	NumCriteriaMetValue int

	// MinYearsBeforeWin: minimum years before winner can be declared (30-500)
	MinYearsBeforeWinValue int
}

// GetVictoryConditions decodes the raw VictoryConditions bytes into a structured format.
// Uses the formulas from data.GetVictoryValue() to convert threshold indices to actual values.
func (p *PlanetsBlock) GetVictoryConditions() DecodedVictoryConditions {
	vc := DecodedVictoryConditions{}

	// Helper to extract enabled flag and threshold index from a byte
	decode := func(b byte) (enabled bool, idx int) {
		return (b & data.VictoryConditionEnabledBit) != 0, int(b & data.VictoryConditionIndexMask)
	}

	// [0] Owns % of planets
	enabled, idx := decode(p.VictoryConditions[data.VictoryIdxOwnsPercentPlanets])
	vc.OwnsPercentPlanetsEnabled = enabled
	vc.OwnsPercentPlanetsValue, _ = data.GetVictoryValue(data.VictoryIdxOwnsPercentPlanets, idx)

	// [1] Attain tech level X
	enabled, idx = decode(p.VictoryConditions[data.VictoryIdxAttainsTechLevel])
	vc.AttainTechLevelEnabled = enabled
	vc.AttainTechLevelValue, _ = data.GetVictoryValue(data.VictoryIdxAttainsTechLevel, idx)

	// [2] In Y tech fields (2nd part of tech condition, shares enabled with [1])
	_, idx = decode(p.VictoryConditions[data.VictoryIdxTechInYFields])
	vc.AttainTechInYFields, _ = data.GetVictoryValue(data.VictoryIdxTechInYFields, idx)

	// [3] Exceeds score
	enabled, idx = decode(p.VictoryConditions[data.VictoryIdxExceedScore])
	vc.ExceedScoreEnabled = enabled
	vc.ExceedScoreValue, _ = data.GetVictoryValue(data.VictoryIdxExceedScore, idx)

	// [4] Exceeds 2nd place by %
	enabled, idx = decode(p.VictoryConditions[data.VictoryIdxExceedSecondPlace])
	vc.ExceedSecondPlaceEnabled = enabled
	vc.ExceedSecondPlaceValue, _ = data.GetVictoryValue(data.VictoryIdxExceedSecondPlace, idx)

	// [5] Production capacity (thousands)
	enabled, idx = decode(p.VictoryConditions[data.VictoryIdxProductionCapacity])
	vc.ProductionCapacityEnabled = enabled
	vc.ProductionCapacityValue, _ = data.GetVictoryValue(data.VictoryIdxProductionCapacity, idx)

	// [6] Owns capital ships
	enabled, idx = decode(p.VictoryConditions[data.VictoryIdxOwnCapitalShips])
	vc.OwnCapitalShipsEnabled = enabled
	vc.OwnCapitalShipsValue, _ = data.GetVictoryValue(data.VictoryIdxOwnCapitalShips, idx)

	// [7] Highest score after N years
	enabled, idx = decode(p.VictoryConditions[data.VictoryIdxHighestScoreYears])
	vc.HighestScoreYearsEnabled = enabled
	vc.HighestScoreYearsValue, _ = data.GetVictoryValue(data.VictoryIdxHighestScoreYears, idx)

	// [8] Meets N criteria (special: value is directly the count, 1-7)
	_, idx = decode(p.VictoryConditions[data.VictoryIdxNumCriteriaMet])
	vc.NumCriteriaMetValue, _ = data.GetVictoryValue(data.VictoryIdxNumCriteriaMet, idx)

	// [9] Min years before winner declared
	_, idx = decode(p.VictoryConditions[data.VictoryIdxMinYearsBeforeWin])
	vc.MinYearsBeforeWinValue, _ = data.GetVictoryValue(data.VictoryIdxMinYearsBeforeWin, idx)

	return vc
}

// SetVictoryConditions encodes the DecodedVictoryConditions struct back to raw bytes.
// This converts actual game values back to threshold indices using inverse formulas.
func (p *PlanetsBlock) SetVictoryConditions(vc DecodedVictoryConditions) {
	// Helper to encode enabled flag and value back to a byte
	encode := func(enabled bool, conditionIdx, value int) byte {
		idx := valueToIndex(conditionIdx, value)
		b := byte(idx & data.VictoryConditionIndexMask)
		if enabled {
			b |= data.VictoryConditionEnabledBit
		}
		return b
	}

	// [0] Owns % of planets
	p.VictoryConditions[data.VictoryIdxOwnsPercentPlanets] = encode(
		vc.OwnsPercentPlanetsEnabled,
		data.VictoryIdxOwnsPercentPlanets,
		vc.OwnsPercentPlanetsValue,
	)

	// [1] Attain tech level X
	p.VictoryConditions[data.VictoryIdxAttainsTechLevel] = encode(
		vc.AttainTechLevelEnabled,
		data.VictoryIdxAttainsTechLevel,
		vc.AttainTechLevelValue,
	)

	// [2] In Y tech fields (shares enabled with [1], but we don't set enabled here)
	p.VictoryConditions[data.VictoryIdxTechInYFields] = encode(
		false, // enabled flag not used for this index
		data.VictoryIdxTechInYFields,
		vc.AttainTechInYFields,
	)

	// [3] Exceeds score
	p.VictoryConditions[data.VictoryIdxExceedScore] = encode(
		vc.ExceedScoreEnabled,
		data.VictoryIdxExceedScore,
		vc.ExceedScoreValue,
	)

	// [4] Exceeds 2nd place by %
	p.VictoryConditions[data.VictoryIdxExceedSecondPlace] = encode(
		vc.ExceedSecondPlaceEnabled,
		data.VictoryIdxExceedSecondPlace,
		vc.ExceedSecondPlaceValue,
	)

	// [5] Production capacity (thousands)
	p.VictoryConditions[data.VictoryIdxProductionCapacity] = encode(
		vc.ProductionCapacityEnabled,
		data.VictoryIdxProductionCapacity,
		vc.ProductionCapacityValue,
	)

	// [6] Owns capital ships
	p.VictoryConditions[data.VictoryIdxOwnCapitalShips] = encode(
		vc.OwnCapitalShipsEnabled,
		data.VictoryIdxOwnCapitalShips,
		vc.OwnCapitalShipsValue,
	)

	// [7] Highest score after N years
	p.VictoryConditions[data.VictoryIdxHighestScoreYears] = encode(
		vc.HighestScoreYearsEnabled,
		data.VictoryIdxHighestScoreYears,
		vc.HighestScoreYearsValue,
	)

	// [8] Meets N criteria (no enabled flag, just the count)
	p.VictoryConditions[data.VictoryIdxNumCriteriaMet] = byte(vc.NumCriteriaMetValue & data.VictoryConditionIndexMask)

	// [9] Min years before winner declared (no enabled flag)
	p.VictoryConditions[data.VictoryIdxMinYearsBeforeWin] = encode(
		false,
		data.VictoryIdxMinYearsBeforeWin,
		vc.MinYearsBeforeWinValue,
	)

	// [10-11] Reserved
	p.VictoryConditions[data.VictoryIdxReserved10] = 0
	p.VictoryConditions[data.VictoryIdxReserved11] = 0
}

// valueToIndex converts an actual game value back to a threshold index.
// This is the inverse of data.GetVictoryValue().
func valueToIndex(conditionIdx, value int) int {
	switch conditionIdx {
	case data.VictoryIdxOwnsPercentPlanets:
		return (value - 20) / 5 // idx*5+20 → (value-20)/5
	case data.VictoryIdxAttainsTechLevel:
		return value - 8 // idx+8 → value-8
	case data.VictoryIdxTechInYFields:
		return value - 2 // idx+2 → value-2
	case data.VictoryIdxExceedScore:
		return (value - 1000) / 1000 // idx*1000+1000 → (value-1000)/1000
	case data.VictoryIdxExceedSecondPlace:
		return (value - 20) / 10 // idx*10+20 → (value-20)/10
	case data.VictoryIdxProductionCapacity:
		return (value - 10) / 10 // idx*10+10 → (value-10)/10
	case data.VictoryIdxOwnCapitalShips:
		return (value - 10) / 10 // idx*10+10 → (value-10)/10
	case data.VictoryIdxHighestScoreYears:
		return (value - 30) / 10 // idx*10+30 → (value-30)/10
	case data.VictoryIdxNumCriteriaMet:
		return value // direct value
	case data.VictoryIdxMinYearsBeforeWin:
		return (value - 30) / 10 // idx*10+30 → (value-30)/10
	default:
		return 0
	}
}

// Encode returns the raw 64-byte block data for the PlanetsBlock.
// Note: This only encodes the main block data, not the trailing planet coordinates.
// Use EncodePlanetsData() to encode the planet coordinate data separately.
func (p *PlanetsBlock) Encode() []byte {
	buf := make([]byte, 64)

	// Bytes 0-3: Game ID (lid) - unique identifier, preserved from original
	if len(p.DecryptedData()) >= 4 {
		copy(buf[0:4], p.DecryptedData()[0:4])
	}

	// Bytes 4-5: Universe size
	encoding.Write16(buf, 4, p.UniverseSize)

	// Bytes 6-7: Density
	encoding.Write16(buf, 6, p.Density)

	// Bytes 8-9: Player count
	encoding.Write16(buf, 8, p.PlayerCount)

	// Bytes 10-11: Planet count
	encoding.Write16(buf, 10, p.PlanetCount)

	// Bytes 12-13: Starting distance mode (mdStartDist)
	encoding.Write16(buf, 12, p.StartingDistance)
	// Bytes 14-15: fDirty - runtime flag, write as 0 (not meaningful in files)
	encoding.Write16(buf, 14, 0)

	// Bytes 16-17: Game settings
	encoding.Write16(buf, 16, p.GameSettings)

	// Bytes 18-19: Turn
	encoding.Write16(buf, 18, p.Turn)

	// Bytes 20-31: Victory conditions (12 bytes)
	copy(buf[20:32], p.VictoryConditions[:])

	// Bytes 32-63: Game name (32 bytes, null-padded)
	gameName := []byte(p.GameName)
	if len(gameName) > 32 {
		gameName = gameName[:32]
	}
	copy(buf[32:64], gameName)

	return buf
}

// EncodePlanetsData returns the raw trailing planet coordinate data.
// This data follows the main block and is NOT encrypted.
// Returns 4 bytes per planet (PlanetCount * 4 bytes total).
func (p *PlanetsBlock) EncodePlanetsData() []byte {
	// If we have raw data preserved, return it
	if len(p.RawPlanetsData) > 0 {
		return p.RawPlanetsData
	}

	// Otherwise, encode from the Planets slice
	if len(p.Planets) == 0 {
		return nil
	}

	result := make([]byte, len(p.Planets)*4)
	prevX := uint32(1000)

	for i, planet := range p.Planets {
		// Calculate X offset from previous planet
		xOffset := planet.X - prevX
		prevX = planet.X

		// Pack: [31:22]=nameID, [21:10]=Y, [9:0]=xOffset
		packed := (planet.NameID << 22) | (planet.Y << 10) | (xOffset & 0x3FF)
		binary.LittleEndian.PutUint32(result[i*4:], packed)
	}

	return result
}
