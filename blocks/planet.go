package blocks

import (
	"github.com/neper-stars/houston/encoding"
)

// Detection level constants for planet visibility.
// The det field (bits 0-6 of flags word) controls what information is visible.
// Source: IO::MarkPlanet function and WritePlanet logic in Stars! decompilation.
const (
	DetNotVisible = 0 // Planet not visible to player
	DetPenScan    = 1 // Pen Scan - basic visibility (planet exists, maybe position)
	DetSpecial    = 2 // Special - used for special cases (avoids starbase updates)
	DetNormalScan = 3 // Normal Scan - standard scan, can see starbase, some details
	DetFull       = 4 // Full - owner can see all planet details (4+ values)
	DetMaximum    = 7 // Maximum - complete information
)

// PartialPlanetBlock represents planet data with conditional fields (Type 14)
// The fields present depend on various flag bits.
//
// Header format (RTPLANET, 4 bytes):
//
//	Bytes 0-1: Planet ID (bits 0-10) + Owner (bits 11-15)
//	Bytes 2-3: Flags word with det field (bits 0-6) and various flags (bits 7-15)
//
// Source: RTPLANET structure in types.h and IO::WritePlanet function
type PartialPlanetBlock struct {
	GenericBlock

	// Planet identification
	PlanetNumber int  // 0-2047 (11 bits)
	Owner        int  // -1 = no owner (31), 0-15 = player ID
	IsHomeworld  bool // Bit 7: fHomeworld - this is a player's homeworld

	// Detection level (bits 0-6 of flags word)
	// 1=Pen Scan, 2=Special, 3=Normal Scan, 4+=Full, 7=Maximum
	DetectionLevel int

	// Status flags (from 2-byte flag word at offset 2)
	Include            bool // Bit 8: fInclude - planet included in scans/reports
	HasStarbase        bool // Bit 9: fStarbase - planet has a starbase
	IsTerraformed      bool // Bit 10: fIncEVO - original environment values included
	HasInstallations   bool // Bit 11: fIncImp - 8 bytes of installations data included
	HasArtifact        bool // Bit 12: fIsArtifact - planet has an ancient artifact
	HasSurfaceMinerals bool // Bit 13: fIncSurfMin - surface minerals data included
	HasRoute           bool // Bit 14: fRouting - fleet route destination is set
	FirstYear          bool // Bit 15: fFirstYear - first year this planet is visible to player

	// Environment data (if DetectionLevel >= 2)
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
	// Bytes 0-3: iDeltaPop(8) + cMines(12) + cFactories(12)
	// Bytes 4-7: cDefenses(12) + iScanner(5) + unused5(5) + fArtifact(1) + fNoResearch(1) + unused(8)
	DeltaPop     int  // 8-bit population change indicator
	Mines        int  // 12-bit mine count (0-4095)
	Factories    int  // 12-bit factory count (0-4095)
	Defenses     int  // 12-bit defense count (0-4095)
	ScannerID    int  // 5-bit planetary scanner ID (0=none, 31=no scanner)
	InstArtifact bool // Bit 22: fArtifact in installations (also in header)
	NoResearch   bool // Bit 23: fNoResearch - don't contribute to research

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

// CanSeeEnvironment returns true if environment data should be present.
// Environment data is present when detection level >= DetSpecial (2).
func (pb *PartialPlanetBlock) CanSeeEnvironment() bool {
	return pb.DetectionLevel >= DetSpecial
}

func (pb *PartialPlanetBlock) decode(isPlanet bool) {
	data := pb.Decrypted
	if len(data) < 4 {
		return
	}

	// Bytes 0-1: Planet ID (bits 0-10) + Owner (bits 11-15)
	pb.PlanetNumber = int(data[0]&0xFF) + (int(data[1]&0x07) << 8)
	ownerBits := (int(data[1]) & 0xF8) >> 3
	if ownerBits == 31 {
		pb.Owner = -1 // No owner
	} else {
		pb.Owner = ownerBits
	}

	// Bytes 2-3: 16-bit flag word
	// Bits 0-6: det (detection level, 7 bits)
	// Bit 7: fHomeworld
	// Bit 8: fInclude
	// Bit 9: fStarbase
	// Bit 10: fIncEVO (terraformed)
	// Bit 11: fIncImp (installations)
	// Bit 12: fIsArtifact
	// Bit 13: fIncSurfMin (surface minerals)
	// Bit 14: fRouting (route)
	// Bit 15: fFirstYear
	flags := encoding.Read16(data, 2)
	pb.DetectionLevel = int(flags & 0x7F)         // Bits 0-6
	pb.IsHomeworld = (flags & 0x0080) != 0        // Bit 7
	pb.Include = (flags & 0x0100) != 0            // Bit 8
	pb.HasStarbase = (flags & 0x0200) != 0        // Bit 9
	pb.IsTerraformed = (flags & 0x0400) != 0      // Bit 10
	pb.HasInstallations = (flags & 0x0800) != 0   // Bit 11
	pb.HasArtifact = (flags & 0x1000) != 0        // Bit 12
	pb.HasSurfaceMinerals = (flags & 0x2000) != 0 // Bit 13
	pb.HasRoute = (flags & 0x4000) != 0           // Bit 14
	pb.FirstYear = (flags & 0x8000) != 0          // Bit 15

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
	// Bytes 0-3 (32-bit packed): iDeltaPop(8) + cMines(12) + cFactories(12)
	// Bytes 4-7 (32-bit packed): cDefenses(12) + iScanner(5) + unused5(5) + fArtifact(1) + fNoResearch(1) + unused2(8)
	if pb.HasInstallations && index+8 <= len(data) {
		// First dword: population change + mines + factories
		dword1 := encoding.Read32(data, index)
		pb.DeltaPop = int(dword1 & 0xFF)           // Bits 0-7
		pb.Mines = int((dword1 >> 8) & 0xFFF)      // Bits 8-19
		pb.Factories = int((dword1 >> 20) & 0xFFF) // Bits 20-31

		// Second dword: defenses + scanner + flags
		dword2 := encoding.Read32(data, index+4)
		pb.Defenses = int(dword2 & 0xFFF)         // Bits 0-11
		pb.ScannerID = int((dword2 >> 12) & 0x1F) // Bits 12-16
		// Bits 17-21: unused5 (5 bits)
		pb.InstArtifact = (dword2 & (1 << 22)) != 0 // Bit 22
		pb.NoResearch = (dword2 & (1 << 23)) != 0   // Bit 23
		// Bits 24-31: unused2 (8 bits)
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

	// Bytes 2-3: Flags word
	// Bits 0-6: det (detection level)
	// Bits 7-15: various flags
	var flags uint16
	flags = uint16(pb.DetectionLevel & 0x7F) // Bits 0-6
	if pb.IsHomeworld {
		flags |= 0x0080 // Bit 7
	}
	if pb.Include {
		flags |= 0x0100 // Bit 8
	}
	if pb.HasStarbase {
		flags |= 0x0200 // Bit 9
	}
	if pb.IsTerraformed {
		flags |= 0x0400 // Bit 10
	}
	if pb.HasInstallations {
		flags |= 0x0800 // Bit 11
	}
	if pb.HasArtifact {
		flags |= 0x1000 // Bit 12
	}
	if pb.HasSurfaceMinerals {
		flags |= 0x2000 // Bit 13
	}
	if pb.HasRoute {
		flags |= 0x4000 // Bit 14
	}
	if pb.FirstYear {
		flags |= 0x8000 // Bit 15
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
