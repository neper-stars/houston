// Package blocks provides Stars! file block parsers.
package blocks

import (
	"encoding/binary"

	"github.com/neper-stars/houston/data"
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

	// StartingDistance affects how far apart players' homeworlds are placed.
	// Higher values mean players start further from each other.
	StartingDistance uint32

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

	// GameName is the name of the game, up to 32 characters.
	// Padded with null bytes if shorter.
	GameName string

	// Planets is the list of all planets in the universe.
	// Populated by ParsePlanetsData() after the main block is decoded.
	Planets []Planet
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

	// Bytes 0-3: Unknown/reserved (TODO: document if purpose discovered)
	// Bytes 4-5: Universe size
	block.UniverseSize = binary.LittleEndian.Uint16(data[4:6])

	// Bytes 6-7: Planet density
	block.Density = binary.LittleEndian.Uint16(data[6:8])

	// Bytes 8-9: Number of players
	block.PlayerCount = binary.LittleEndian.Uint16(data[8:10])

	// Bytes 10-11: Number of planets
	block.PlanetCount = binary.LittleEndian.Uint16(data[10:12])

	// Bytes 12-15: Starting distance between players
	block.StartingDistance = binary.LittleEndian.Uint32(data[12:16])

	// Bytes 16-17: Game settings bitmask
	block.GameSettings = binary.LittleEndian.Uint16(data[16:18])

	// Bytes 18-31: Unknown/reserved (TODO: document if purpose discovered)
	// These bytes likely contain additional game settings or victory conditions

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
