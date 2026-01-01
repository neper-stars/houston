package blocks

import (
	"github.com/neper-stars/houston/encoding"
)

// PartialPlanetBlock represents planet data with conditional fields (Type 14)
// The fields present depend on various flag bits
type PartialPlanetBlock struct {
	GenericBlock

	// Planet identification
	PlanetNumber int  // 0-511 (9 bits)
	Owner        int  // -1 = no owner, 0-15 = player ID
	IsHomeworld  bool // True if this is a homeworld

	// Status flags (from 2-byte flag word at offset 2)
	IsInUseOrRobberBaron                       bool
	HasEnvironmentInfo                         bool
	BitWhichIsOffForRemoteMiningAndRobberBaron bool
	WeirdBit                                   bool
	HasRoute                                   bool
	HasSurfaceMinerals                         bool
	HasArtifact                                bool
	HasInstallations                           bool
	IsTerraformed                              bool
	HasStarbase                                bool

	// Environment data (if HasEnvironmentInfo or can see environment)
	IroniumConc   int // 0-100
	BoraniumConc  int // 0-100
	GermaniumConc int // 0-100
	Gravity       int // 0-100
	Temperature   int // 0-100
	Radiation     int // 0-100

	// Original environment values (if terraformed)
	OrigGravity     int
	OrigTemperature int
	OrigRadiation   int

	// Estimates (if owner >= 0)
	DefensesEstimate int // 0-15 (in sixteenths of 100%)
	PopEstimate      int // in 400s, up to 4090

	// Surface minerals (if HasSurfaceMinerals)
	Ironium    int64
	Boranium   int64
	Germanium  int64
	Population int64

	// Installations (if HasInstallations)
	ExcessPop                                 int
	Mines                                     int // 12-bit value
	Factories                                 int // 12-bit value
	Defenses                                  int
	UnknownInstallationsByte                  byte
	ContributeOnlyLeftoverResourcesToResearch bool
	HasScanner                                bool

	// Starbase (if HasStarbase)
	StarbaseDesign     int    // Design number 0-15
	StarbaseBytes      []byte // Full starbase data (4 bytes for full planet)
	MassDriverDest     int    // Mass driver destination planet (display ID, 0 = none)
	MassDriverDestZero bool   // True if destination is explicitly planet 0 vs no destination

	// Route (if HasRoute)
	RouteTarget int // Route destination

	// Turn number (optional, last 2 bytes if present)
	Turn int
}

// NewPartialPlanetBlock creates a PartialPlanetBlock from a GenericBlock
func NewPartialPlanetBlock(b GenericBlock) *PartialPlanetBlock {
	pb := &PartialPlanetBlock{
		GenericBlock: b,
		Owner:        -1,
	}
	pb.decode(false)
	return pb
}

// CanSeeEnvironment returns true if environment data should be present
func (pb *PartialPlanetBlock) CanSeeEnvironment() bool {
	return pb.HasEnvironmentInfo || pb.IsInUseOrRobberBaron
}

func (pb *PartialPlanetBlock) decode(isPlanet bool) {
	data := pb.Decrypted
	if len(data) < 4 {
		return
	}

	// Bytes 0-1: Planet number and owner
	pb.PlanetNumber = int(data[0]&0xFF) + (int(data[1]&0x07) << 8)
	ownerBits := (int(data[1]) & 0xF8) >> 3
	if ownerBits == 31 {
		pb.Owner = -1 // No owner
	} else {
		pb.Owner = ownerBits
	}

	// Bytes 2-3: 16-bit flag word
	flags := encoding.Read16(data, 2)
	pb.IsInUseOrRobberBaron = (flags & 0x04) != 0
	pb.HasEnvironmentInfo = (flags & 0x02) != 0
	pb.BitWhichIsOffForRemoteMiningAndRobberBaron = (flags & 0x01) != 0
	pb.WeirdBit = (flags & 0x8000) != 0
	pb.HasRoute = (flags & 0x4000) != 0
	pb.HasSurfaceMinerals = (flags & 0x2000) != 0
	pb.HasArtifact = (flags & 0x1000) != 0
	pb.HasInstallations = (flags & 0x0800) != 0
	pb.IsTerraformed = (flags & 0x0400) != 0
	pb.HasStarbase = (flags & 0x0200) != 0
	pb.IsHomeworld = (flags & 0x0080) != 0 // Bit 7 of low byte (flag1)

	index := 4

	// Variable-length environment section
	if pb.CanSeeEnvironment() && index < len(data) {
		// Byte 4 encodes the length of fractional mineral concentration bytes
		// Length = 1 + (bits 0-1) + (bits 2-3) + (bits 4-5)
		// Bits 6-7 must be 0
		preEnvLengthByte := int(data[index] & 0xFF)
		preEnvLength := 1
		preEnvLength += preEnvLengthByte & 0x03
		preEnvLength += (preEnvLengthByte & 0x0C) >> 2
		preEnvLength += (preEnvLengthByte & 0x30) >> 4
		index += preEnvLength

		if index+3 <= len(data) {
			// Mineral concentrations (3 bytes)
			pb.IroniumConc = int(data[index] & 0xFF)
			pb.BoraniumConc = int(data[index+1] & 0xFF)
			pb.GermaniumConc = int(data[index+2] & 0xFF)
			index += 3
		}

		if index+3 <= len(data) {
			// Gravity, temperature, radiation (3 bytes)
			pb.Gravity = int(data[index] & 0xFF)
			pb.Temperature = int(data[index+1] & 0xFF)
			pb.Radiation = int(data[index+2] & 0xFF)
			index += 3
		}

		// Original values if terraformed
		if pb.IsTerraformed && index+3 <= len(data) {
			pb.OrigGravity = int(data[index] & 0xFF)
			pb.OrigTemperature = int(data[index+1] & 0xFF)
			pb.OrigRadiation = int(data[index+2] & 0xFF)
			index += 3
		}

		// Estimates if owned
		if pb.Owner >= 0 && index+2 <= len(data) {
			estimateWord := encoding.Read16(data, index)
			pb.DefensesEstimate = int(estimateWord >> 12)   // Upper 4 bits
			pb.PopEstimate = int(estimateWord&0x0FFF) * 400 // Lower 12 bits * 400
			index += 2
		}
	}

	// Variable-length surface minerals
	if pb.HasSurfaceMinerals && index < len(data) {
		contentsLengths := data[index]
		index++

		// Extract length encodings (2 bits each)
		ironLen := encoding.VarLenByteCount(encoding.ExtractVarLenField(contentsLengths, 0))
		boraLen := encoding.VarLenByteCount(encoding.ExtractVarLenField(contentsLengths, 2))
		germLen := encoding.VarLenByteCount(encoding.ExtractVarLenField(contentsLengths, 4))
		popLen := encoding.VarLenByteCount(encoding.ExtractVarLenField(contentsLengths, 6))

		if index+ironLen <= len(data) {
			pb.Ironium, _ = encoding.ReadVarLen(data, index, ironLen)
			index += ironLen
		}
		if index+boraLen <= len(data) {
			pb.Boranium, _ = encoding.ReadVarLen(data, index, boraLen)
			index += boraLen
		}
		if index+germLen <= len(data) {
			pb.Germanium, _ = encoding.ReadVarLen(data, index, germLen)
			index += germLen
		}
		if index+popLen <= len(data) {
			pb.Population, _ = encoding.ReadVarLen(data, index, popLen)
			index += popLen
		}
	}

	// Fixed 8-byte installations block
	if pb.HasInstallations && index+8 <= len(data) {
		pb.ExcessPop = int(data[index] & 0xFF)
		pb.Mines = int(data[index+1]&0xFF) | (int(data[index+2]&0x0F) << 8)
		pb.Factories = (int(data[index+2]&0xF0) >> 4) | (int(data[index+3]&0xFF) << 4)
		pb.Defenses = int(data[index+4] & 0xFF)
		pb.UnknownInstallationsByte = data[index+5]
		installFlags := data[index+6]
		pb.ContributeOnlyLeftoverResourcesToResearch = (installFlags & 0x80) != 0
		pb.HasScanner = (installFlags & 0x01) == 0 // Note: bit 0 = 0 means HAS scanner
		index += 8
	}

	// Starbase data
	if pb.HasStarbase && index < len(data) {
		if isPlanet && index+4 <= len(data) {
			// Full planet has 4 bytes of starbase data
			pb.StarbaseBytes = make([]byte, 4)
			copy(pb.StarbaseBytes, data[index:index+4])
			pb.StarbaseDesign = int(data[index] & 0x0F)
			// Byte 2 contains mass driver destination (display ID)
			pb.MassDriverDest = int(data[index+2])
			pb.MassDriverDestZero = pb.MassDriverDest == 0 && data[index+2] == 0
			index += 4
		} else {
			// Partial planet has 1 byte with design number
			pb.StarbaseDesign = int(data[index] & 0x0F)
			index++
		}
	}

	// Route target (only for full planet blocks)
	if pb.HasRoute && isPlanet && index+2 <= len(data) {
		pb.RouteTarget = int(encoding.Read16(data, index))
		index += 2
	}

	// Optional turn number (last 2 bytes if present)
	if index+2 == len(data) {
		pb.Turn = int(encoding.Read16(data, index))
	}
}

// Encode returns the raw block data bytes (without the 2-byte block header).
// Note: Planet blocks have complex variable-length encoding. This preserves raw data
// for blocks that were decoded, or builds from scratch for new blocks.
func (pb *PartialPlanetBlock) Encode() []byte {
	// If we have the original decrypted data, preserve it
	if len(pb.Decrypted) > 0 {
		return pb.Decrypted
	}

	// Build minimal planet header (requires implementing full encoding)
	// For now, return the original data if available
	if pb.Data != nil {
		return pb.Data
	}

	// Minimal encoding: just the 4-byte header
	data := make([]byte, 4)

	// Bytes 0-1: Planet number and owner
	data[0] = byte(pb.PlanetNumber & 0xFF)
	ownerBits := 31 // No owner
	if pb.Owner >= 0 && pb.Owner <= 15 {
		ownerBits = pb.Owner
	}
	data[1] = byte((pb.PlanetNumber>>8)&0x07) | byte(ownerBits<<3)

	// Bytes 2-3: Flags
	var flags uint16
	if pb.IsInUseOrRobberBaron {
		flags |= 0x04
	}
	if pb.HasEnvironmentInfo {
		flags |= 0x02
	}
	if pb.BitWhichIsOffForRemoteMiningAndRobberBaron {
		flags |= 0x01
	}
	if pb.WeirdBit {
		flags |= 0x8000
	}
	if pb.HasRoute {
		flags |= 0x4000
	}
	if pb.HasSurfaceMinerals {
		flags |= 0x2000
	}
	if pb.HasArtifact {
		flags |= 0x1000
	}
	if pb.HasInstallations {
		flags |= 0x0800
	}
	if pb.IsTerraformed {
		flags |= 0x0400
	}
	if pb.HasStarbase {
		flags |= 0x0200
	}
	if pb.IsHomeworld {
		flags |= 0x0080
	}
	encoding.Write16(data, 2, flags)

	return data
}

// PlanetBlock represents a full planet with all information (Type 13)
// It extends PartialPlanetBlock with additional data
type PlanetBlock struct {
	PartialPlanetBlock
}

// Encode returns the raw block data bytes (without the 2-byte block header).
func (pb *PlanetBlock) Encode() []byte {
	return pb.PartialPlanetBlock.Encode()
}

// NewPlanetBlock creates a PlanetBlock from a GenericBlock
func NewPlanetBlock(b GenericBlock) *PlanetBlock {
	pb := &PlanetBlock{
		PartialPlanetBlock: PartialPlanetBlock{
			GenericBlock: b,
			Owner:        -1,
		},
	}
	pb.decode(true) // isPlanet = true for full planet parsing
	return pb
}
