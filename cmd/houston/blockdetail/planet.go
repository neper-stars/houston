package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.PlanetBlockType, FormatPlanet)
	RegisterFormatter(blocks.PartialPlanetBlockType, FormatPartialPlanet)
}

// FormatPlanet provides detailed view for PlanetBlock (type 13)
func FormatPlanet(block blocks.Block, index int) string {
	pb, ok := block.(blocks.PlanetBlock)
	if !ok {
		return FormatGeneric(block, index)
	}
	return formatPlanetCommon(&pb.PartialPlanetBlock, block, index, true)
}

// FormatPartialPlanet provides detailed view for PartialPlanetBlock (type 14)
func FormatPartialPlanet(block blocks.Block, index int) string {
	ppb, ok := block.(blocks.PartialPlanetBlock)
	if !ok {
		return FormatGeneric(block, index)
	}
	return formatPlanetCommon(&ppb, block, index, false)
}

// formatPlanetCommon handles the common formatting for both PlanetBlock and PartialPlanetBlock
func formatPlanetCommon(pb *blocks.PartialPlanetBlock, block blocks.Block, index int, isPlanet bool) string {
	width := DefaultWidth
	d := pb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 4 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: Planet number and owner
	ownerBits := (int(d[1]) & 0xF8) >> 3
	ownerStr := fmt.Sprintf("%d", pb.Owner)
	if pb.Owner == -1 {
		ownerStr = "-1 (no owner)"
	}
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "Planet ID",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("planetNum=(d[0]+(d[1]&0x07)<<8)=%d, owner=((d[1]>>3)&0x1F)=%d -> %s",
			pb.PlanetNumber, ownerBits, ownerStr)))

	// Bytes 2-3: Flags word
	// Bits 0-6: det (detection level), Bits 7-15: various flags
	flags := encoding.Read16(d, 2)
	fields = append(fields, FormatFieldRaw(0x02, 0x03, "Flags",
		fmt.Sprintf("0x%02X%02X", d[3], d[2]),
		fmt.Sprintf("uint16 LE = 0x%04X", flags)))

	// Detection level description using constant names
	detLevel := pb.DetectionLevel
	detName, detDesc := detectionLevelInfo(detLevel)

	// Flag bits breakdown
	fields = append(fields, fmt.Sprintf("           %s bits0-6: det = %d (%s: %s)", TreeBranch, detLevel, detName, detDesc))
	fields = append(fields, fmt.Sprintf("           %s bit7: fHomeworld = %v", TreeBranch, pb.IsHomeworld))
	fields = append(fields, fmt.Sprintf("           %s bit8: fInclude = %v", TreeBranch, pb.Include))
	fields = append(fields, fmt.Sprintf("           %s bit9: fStarbase = %v", TreeBranch, pb.HasStarbase))
	fields = append(fields, fmt.Sprintf("           %s bit10: fIncEVO (terraformed) = %v", TreeBranch, pb.IsTerraformed))
	fields = append(fields, fmt.Sprintf("           %s bit11: fIncImp (installations) = %v", TreeBranch, pb.HasInstallations))
	fields = append(fields, fmt.Sprintf("           %s bit12: fIsArtifact = %v", TreeBranch, pb.HasArtifact))
	fields = append(fields, fmt.Sprintf("           %s bit13: fIncSurfMin = %v", TreeBranch, pb.HasSurfaceMinerals))
	fields = append(fields, fmt.Sprintf("           %s bit14: fRouting = %v", TreeBranch, pb.HasRoute))
	fields = append(fields, fmt.Sprintf("           %s bit15: fFirstYear = %v", TreeEnd, pb.FirstYear))

	idx := 4

	// Environment section
	if pb.CanSeeEnvironment() && idx < len(d) {
		fields = append(fields, "")
		fields = append(fields, "── Environment Section ──")

		// Pre-environment length byte
		preEnvLengthByte := int(d[idx])
		preEnvLength := 1
		preEnvLength += preEnvLengthByte & 0x03
		preEnvLength += (preEnvLengthByte & 0x0C) >> 2
		preEnvLength += (preEnvLengthByte & 0x30) >> 4

		fields = append(fields, FormatFieldRaw(idx, idx, "PreEnvLength",
			fmt.Sprintf("0x%02X", d[idx]),
			fmt.Sprintf("1+(d&0x03)+((d>>2)&0x03)+((d>>4)&0x03) = %d bytes", preEnvLength)))
		idx += preEnvLength

		// Mineral concentrations
		if idx+3 <= len(d) {
			fields = append(fields, FormatFieldRaw(idx, idx+2, "Mineral Conc",
				fmt.Sprintf("0x%02X%02X%02X", d[idx], d[idx+1], d[idx+2]),
				fmt.Sprintf("Iron=%d%%, Bora=%d%%, Germ=%d%%", pb.IroniumConc, pb.BoraniumConc, pb.GermaniumConc)))
			idx += 3
		}

		// Gravity, temperature, radiation
		if idx+3 <= len(d) {
			fields = append(fields, FormatFieldRaw(idx, idx+2, "Environment",
				fmt.Sprintf("0x%02X%02X%02X", d[idx], d[idx+1], d[idx+2]),
				fmt.Sprintf("Grav=%d, Temp=%d, Rad=%d", pb.Gravity, pb.Temperature, pb.Radiation)))
			idx += 3
		}

		// Original values if terraformed
		if pb.IsTerraformed && idx+3 <= len(d) {
			fields = append(fields, FormatFieldRaw(idx, idx+2, "Original Env",
				fmt.Sprintf("0x%02X%02X%02X", d[idx], d[idx+1], d[idx+2]),
				fmt.Sprintf("OrigGrav=%d, OrigTemp=%d, OrigRad=%d", pb.OrigGravity, pb.OrigTemperature, pb.OrigRadiation)))
			idx += 3
		}

		// Estimates if owned
		if pb.Owner >= 0 && idx+2 <= len(d) {
			estimateWord := encoding.Read16(d, idx)
			fields = append(fields, FormatFieldRaw(idx, idx+1, "Estimates",
				fmt.Sprintf("0x%02X%02X", d[idx+1], d[idx]),
				fmt.Sprintf("uint16 LE=0x%04X -> defEst=(d>>12)=%d, popEst=(d&0xFFF)*400=%d",
					estimateWord, pb.DefensesEstimate, pb.PopEstimate)))
			idx += 2
		}
	}

	// Surface minerals section
	if pb.HasSurfaceMinerals && idx < len(d) {
		fields = append(fields, "")
		fields = append(fields, "── Surface Minerals ──")

		contentsLengths := d[idx]
		fields = append(fields, FormatFieldRaw(idx, idx, "ContentsLengths",
			fmt.Sprintf("0x%02X", contentsLengths),
			fmt.Sprintf("0b%08b (var-len indicators)", contentsLengths)))
		idx++

		// Calculate and skip variable-length minerals data
		ironLen := encoding.VarLenByteCount(encoding.ExtractVarLenField(contentsLengths, 0))
		boraLen := encoding.VarLenByteCount(encoding.ExtractVarLenField(contentsLengths, 2))
		germLen := encoding.VarLenByteCount(encoding.ExtractVarLenField(contentsLengths, 4))
		popLen := encoding.VarLenByteCount(encoding.ExtractVarLenField(contentsLengths, 6))

		fields = append(fields, fmt.Sprintf("           %s Ironium = %d (%d bytes)", TreeBranch, pb.Ironium, ironLen))
		fields = append(fields, fmt.Sprintf("           %s Boranium = %d (%d bytes)", TreeBranch, pb.Boranium, boraLen))
		fields = append(fields, fmt.Sprintf("           %s Germanium = %d (%d bytes)", TreeBranch, pb.Germanium, germLen))
		fields = append(fields, fmt.Sprintf("           %s Population = %d (%d bytes)", TreeEnd, pb.Population, popLen))

		idx += ironLen + boraLen + germLen + popLen
	}

	// Installations section (8 bytes)
	// Bytes 0-3 (32-bit): iDeltaPop(8) + cMines(12) + cFactories(12)
	// Bytes 4-7 (32-bit): cDefenses(12) + iScanner(5) + unused5(5) + fArtifact(1) + fNoResearch(1) + unused(8)
	if pb.HasInstallations && idx+8 <= len(d) {
		fields = append(fields, "")
		fields = append(fields, "── Installations ──")

		// First dword
		dword1 := encoding.Read32(d, idx)
		fields = append(fields, FormatFieldRaw(idx, idx+3, "DWord1",
			fmt.Sprintf("0x%02X%02X%02X%02X", d[idx+3], d[idx+2], d[idx+1], d[idx]),
			fmt.Sprintf("uint32 LE = 0x%08X", dword1)))
		fields = append(fields, fmt.Sprintf("           %s bits0-7: iDeltaPop = %d", TreeBranch, pb.DeltaPop))
		fields = append(fields, fmt.Sprintf("           %s bits8-19: cMines = %d", TreeBranch, pb.Mines))
		fields = append(fields, fmt.Sprintf("           %s bits20-31: cFactories = %d", TreeEnd, pb.Factories))

		// Second dword
		dword2 := encoding.Read32(d, idx+4)
		fields = append(fields, FormatFieldRaw(idx+4, idx+7, "DWord2",
			fmt.Sprintf("0x%02X%02X%02X%02X", d[idx+7], d[idx+6], d[idx+5], d[idx+4]),
			fmt.Sprintf("uint32 LE = 0x%08X", dword2)))
		fields = append(fields, fmt.Sprintf("           %s bits0-11: cDefenses = %d", TreeBranch, pb.Defenses))
		fields = append(fields, fmt.Sprintf("           %s bits12-16: iScanner = %d (%s)", TreeBranch, pb.ScannerID, scannerInfo(pb.ScannerID)))
		fields = append(fields, fmt.Sprintf("           %s bits17-21: unused5 = %d", TreeBranch, (dword2>>17)&0x1F))
		fields = append(fields, fmt.Sprintf("           %s bit22: fArtifact = %v", TreeBranch, pb.InstArtifact))
		fields = append(fields, fmt.Sprintf("           %s bit23: fNoResearch = %v", TreeBranch, pb.NoResearch))
		fields = append(fields, fmt.Sprintf("           %s bits24-31: unused2 = %d", TreeEnd, (dword2>>24)&0xFF))

		idx += 8
	}

	// Starbase section
	if pb.HasStarbase && idx < len(d) {
		fields = append(fields, "")
		fields = append(fields, "── Starbase ──")

		if isPlanet && idx+4 <= len(d) {
			fields = append(fields, FormatFieldRaw(idx, idx+3, "StarbaseData",
				fmt.Sprintf("0x%02X%02X%02X%02X", d[idx], d[idx+1], d[idx+2], d[idx+3]),
				fmt.Sprintf("design=(d[0]&0x0F)=%d, massDriverDest=d[2]=%d", pb.StarbaseDesign, pb.MassDriverDest)))
			idx += 4
		} else {
			fields = append(fields, FormatFieldRaw(idx, idx, "StarbaseDesign",
				fmt.Sprintf("0x%02X", d[idx]),
				fmt.Sprintf("(d&0x0F) = %d", pb.StarbaseDesign)))
			idx++
		}
	}

	// Route target
	if pb.HasRoute && isPlanet && idx+2 <= len(d) {
		fields = append(fields, "")
		fields = append(fields, FormatFieldRaw(idx, idx+1, "RouteTarget",
			fmt.Sprintf("0x%02X%02X", d[idx+1], d[idx]),
			fmt.Sprintf("uint16 LE = %d", pb.RouteTarget)))
		idx += 2
	}

	// Turn number (if present at end)
	if idx+2 == len(d) {
		fields = append(fields, "")
		fields = append(fields, FormatFieldRaw(idx, idx+1, "Turn",
			fmt.Sprintf("0x%02X%02X", d[idx+1], d[idx]),
			fmt.Sprintf("uint16 LE = %d -> Year %d", pb.Turn, 2400+pb.Turn)))
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// detectionLevelInfo returns the constant name and description for a detection level.
func detectionLevelInfo(level int) (name, desc string) {
	switch level {
	case blocks.DetNotVisible:
		return "DetNotVisible", "planet not visible"
	case blocks.DetPenScan:
		return "DetPenScan", "basic visibility"
	case blocks.DetSpecial:
		return "DetSpecial", "special (avoids starbase updates)"
	case blocks.DetNormalScan:
		return "DetNormalScan", "standard scan"
	case blocks.DetFull:
		return "DetFull", "full planet details"
	case blocks.DetMaximum:
		return "DetMaximum", "complete information"
	default:
		if level > blocks.DetFull && level < blocks.DetMaximum {
			return "DetFull+", "full planet details"
		}
		return "Unknown", fmt.Sprintf("unknown level %d", level)
	}
}

// scannerInfo returns a description for a scanner ID.
func scannerInfo(scannerID int) string {
	switch scannerID {
	case 0:
		return "None (no scanner installed)"
	case 31:
		return "NoScanner (special value)"
	default:
		return fmt.Sprintf("Scanner type %d", scannerID)
	}
}
