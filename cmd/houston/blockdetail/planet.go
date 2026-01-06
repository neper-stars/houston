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

	// Bytes 2-3: Flags
	flags := encoding.Read16(d, 2)
	fields = append(fields, FormatFieldRaw(0x02, 0x03, "Flags",
		fmt.Sprintf("0x%02X%02X", d[3], d[2]),
		fmt.Sprintf("uint16 LE = 0x%04X", flags)))

	// Flag bits breakdown
	fields = append(fields, fmt.Sprintf("           %s bit0: RemoteMining/RobberBaron flag = %v", TreeBranch, pb.BitWhichIsOffForRemoteMiningAndRobberBaron))
	fields = append(fields, fmt.Sprintf("           %s bit1: HasEnvironmentInfo = %v", TreeBranch, pb.HasEnvironmentInfo))
	fields = append(fields, fmt.Sprintf("           %s bit2: IsInUseOrRobberBaron = %v", TreeBranch, pb.IsInUseOrRobberBaron))
	fields = append(fields, fmt.Sprintf("           %s bits3-6: ??? (unknown) = 0x%X", TreeBranch, (flags>>3)&0x0F))
	fields = append(fields, fmt.Sprintf("           %s bit7: IsHomeworld = %v", TreeBranch, pb.IsHomeworld))
	fields = append(fields, fmt.Sprintf("           %s bit8: ??? (unknown) = %v", TreeBranch, (flags&0x0100) != 0))
	fields = append(fields, fmt.Sprintf("           %s bit9: HasStarbase = %v", TreeBranch, pb.HasStarbase))
	fields = append(fields, fmt.Sprintf("           %s bit10: IsTerraformed = %v", TreeBranch, pb.IsTerraformed))
	fields = append(fields, fmt.Sprintf("           %s bit11: HasInstallations = %v", TreeBranch, pb.HasInstallations))
	fields = append(fields, fmt.Sprintf("           %s bit12: HasArtifact = %v", TreeBranch, pb.HasArtifact))
	fields = append(fields, fmt.Sprintf("           %s bit13: HasSurfaceMinerals = %v", TreeBranch, pb.HasSurfaceMinerals))
	fields = append(fields, fmt.Sprintf("           %s bit14: HasRoute = %v", TreeBranch, pb.HasRoute))
	fields = append(fields, fmt.Sprintf("           %s bit15: WeirdBit = %v", TreeEnd, pb.WeirdBit))

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

	// Installations section
	if pb.HasInstallations && idx+8 <= len(d) {
		fields = append(fields, "")
		fields = append(fields, "── Installations ──")

		fields = append(fields, FormatFieldRaw(idx, idx, "ExcessPop",
			fmt.Sprintf("0x%02X", d[idx]),
			fmt.Sprintf("%d", pb.ExcessPop)))

		// Mines and Factories are packed in 3 bytes (12 bits each)
		// Layout: [byte1][byte2][byte3] where byte2's nibbles split between Mines and Factories
		//   Mines    = byte1 | (byte2 & 0x0F) << 8     (byte1 is bits 0-7, lo_nibble is bits 8-11)
		//   Factories = (byte2 >> 4) | byte3 << 4      (hi_nibble is bits 0-3, byte3 is bits 4-11)
		fields = append(fields, FormatFieldRaw(idx+1, idx+3, "Mines/Factories",
			fmt.Sprintf("0x%02X %02X %02X", d[idx+1], d[idx+2], d[idx+3]),
			"packed 12-bit values"))
		// Show the nibble split of the middle byte
		loNibble := int(d[idx+2] & 0x0F)
		hiNibble := int((d[idx+2] >> 4) & 0x0F)
		byte3 := int(d[idx+3])
		fields = append(fields, fmt.Sprintf("           %s byte[0x%02X]=0x%02X splits: lo_nibble=0x%X, hi_nibble=0x%X",
			TreeBranch, idx+2, d[idx+2], loNibble, hiNibble))
		fields = append(fields, fmt.Sprintf("           %s Mines:     byte[0x%02X] | (lo_nibble << 8) = 0x%02X | (0x%X << 8) = 0x%02X | 0x%03X = 0x%03X = %d",
			TreeBranch, idx+1, d[idx+1], loNibble, d[idx+1], loNibble<<8, pb.Mines, pb.Mines))
		fields = append(fields, fmt.Sprintf("           %s Factories: hi_nibble | (byte[0x%02X] << 4) = 0x%X | (0x%02X << 4) = 0x%X | 0x%03X = 0x%03X = %d",
			TreeEnd, idx+3, hiNibble, d[idx+3], hiNibble, byte3<<4, pb.Factories, pb.Factories))

		fields = append(fields, FormatFieldRaw(idx+4, idx+4, "Defenses",
			fmt.Sprintf("0x%02X", d[idx+4]),
			fmt.Sprintf("%d", pb.Defenses)))

		fields = append(fields, FormatFieldRaw(idx+5, idx+5, "Unknown",
			fmt.Sprintf("0x%02X", d[idx+5]),
			"TBD"))

		fields = append(fields, FormatFieldRaw(idx+6, idx+6, "InstallFlags",
			fmt.Sprintf("0x%02X", d[idx+6]),
			fmt.Sprintf("0b%08b", d[idx+6])))
		fields = append(fields, fmt.Sprintf("           %s bit0: HasScanner = %v (0=has, 1=no)", TreeBranch, pb.HasScanner))
		fields = append(fields, fmt.Sprintf("           %s bit7: LeftoverResearch = %v", TreeEnd, pb.ContributeOnlyLeftoverResourcesToResearch))

		fields = append(fields, FormatFieldRaw(idx+7, idx+7, "Unknown",
			fmt.Sprintf("0x%02X", d[idx+7]),
			"TBD"))

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
