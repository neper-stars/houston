package blockdetail

import (
	"fmt"
	"strings"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/data"
)

func init() {
	RegisterFormatter(blocks.PlanetsBlockType, FormatPlanets)
}

// universeSizeName returns human-readable name for universe size
func universeSizeName(size uint16) string {
	names := []string{"Tiny", "Small", "Medium", "Large", "Huge"}
	if int(size) < len(names) {
		return names[size]
	}
	return fmt.Sprintf("Unknown(%d)", size)
}

// densityName returns human-readable name for planet density
func densityName(density uint16) string {
	names := []string{"Sparse", "Normal", "Dense", "Packed"}
	if int(density) < len(names) {
		return names[density]
	}
	return fmt.Sprintf("Unknown(%d)", density)
}

// FormatPlanets provides detailed view for PlanetsBlock (type 7)
func FormatPlanets(block blocks.Block, index int) string {
	width := DefaultWidth
	pb, ok := block.(blocks.PlanetsBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := pb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 64 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-3: Unknown/reserved
	fields = append(fields, FormatFieldRaw(0x00, 0x03, "Unknown",
		fmt.Sprintf("0x%02X%02X%02X%02X", d[3], d[2], d[1], d[0]),
		"TBD (reserved)"))

	// Bytes 4-5: Universe size
	fields = append(fields, FormatFieldRaw(0x04, 0x05, "UniverseSize",
		fmt.Sprintf("0x%02X%02X", d[5], d[4]),
		fmt.Sprintf("uint16 LE = %d (%s)", pb.UniverseSize, universeSizeName(pb.UniverseSize))))

	// Bytes 6-7: Density
	fields = append(fields, FormatFieldRaw(0x06, 0x07, "Density",
		fmt.Sprintf("0x%02X%02X", d[7], d[6]),
		fmt.Sprintf("uint16 LE = %d (%s)", pb.Density, densityName(pb.Density))))

	// Bytes 8-9: Player count
	fields = append(fields, FormatFieldRaw(0x08, 0x09, "PlayerCount",
		fmt.Sprintf("0x%02X%02X", d[9], d[8]),
		fmt.Sprintf("uint16 LE = %d", pb.PlayerCount)))

	// Bytes 10-11: Planet count
	fields = append(fields, FormatFieldRaw(0x0A, 0x0B, "PlanetCount",
		fmt.Sprintf("0x%02X%02X", d[11], d[10]),
		fmt.Sprintf("uint16 LE = %d", pb.PlanetCount)))

	// Bytes 12-15: Starting distance
	fields = append(fields, FormatFieldRaw(0x0C, 0x0F, "StartingDistance",
		fmt.Sprintf("0x%02X%02X%02X%02X", d[15], d[14], d[13], d[12]),
		fmt.Sprintf("uint32 LE = %d", pb.StartingDistance)))

	// Bytes 16-17: Game settings
	fields = append(fields, FormatFieldRaw(0x10, 0x11, "GameSettings",
		fmt.Sprintf("0x%02X%02X", d[17], d[16]),
		fmt.Sprintf("uint16 LE = 0x%04X", pb.GameSettings)))

	// Game settings breakdown
	fields = append(fields, fmt.Sprintf("           %s bit0: MaxMinerals = %v",
		TreeBranch, pb.HasGameSetting(data.GameSettingMaxMinerals)))
	fields = append(fields, fmt.Sprintf("           %s bit1: SlowTechAdvances = %v",
		TreeBranch, pb.HasGameSetting(data.GameSettingSlowTech)))
	fields = append(fields, fmt.Sprintf("           %s bit2: SinglePlayer = %v",
		TreeBranch, pb.HasGameSetting(data.GameSettingSinglePlayer)))
	fields = append(fields, fmt.Sprintf("           %s bit3: Unknown = %v",
		TreeBranch, (pb.GameSettings&0x08) != 0))
	fields = append(fields, fmt.Sprintf("           %s bit4: ComputerAlliances = %v",
		TreeBranch, pb.HasGameSetting(data.GameSettingComputerAlliances)))
	fields = append(fields, fmt.Sprintf("           %s bit5: PublicScores = %v",
		TreeBranch, pb.HasGameSetting(data.GameSettingPublicScores)))
	fields = append(fields, fmt.Sprintf("           %s bit6: AcceleratedBBSPlay = %v",
		TreeBranch, pb.HasGameSetting(data.GameSettingAcceleratedBBS)))
	fields = append(fields, fmt.Sprintf("           %s bit7: NoRandomEvents = %v",
		TreeBranch, pb.HasGameSetting(data.GameSettingNoRandomEvents)))
	fields = append(fields, fmt.Sprintf("           %s bit8: GalaxyClumping = %v",
		TreeEnd, pb.HasGameSetting(data.GameSettingGalaxyClumping)))

	// Bytes 18-19: Turn number
	fields = append(fields, FormatFieldRaw(0x12, 0x13, "Turn",
		fmt.Sprintf("0x%02X%02X", d[19], d[18]),
		fmt.Sprintf("uint16 LE = %d", pb.Turn)))

	// Bytes 20-31: Victory conditions (12 bytes)
	fields = append(fields, FormatFieldRaw(0x14, 0x1F, "VictoryConditions",
		fmt.Sprintf("0x%s", HexDumpSingleLine(d[20:32])),
		"12 bytes (see breakdown below)"))

	// Victory conditions breakdown
	vc := pb.GetVictoryConditions()
	fields = append(fields, fmt.Sprintf("           %s [0] Owns %%planets: enabled=%v, value=%d%%",
		TreeBranch, vc.OwnsPercentPlanetsEnabled, vc.OwnsPercentPlanetsValue))
	fields = append(fields, fmt.Sprintf("           %s [1] Tech level X in Y fields: enabled=%v, level=%d, fields=%d",
		TreeBranch, vc.AttainTechLevelEnabled, vc.AttainTechLevelValue, vc.AttainTechInYFields))
	fields = append(fields, fmt.Sprintf("           %s [3] Exceeds score: enabled=%v, value=%d",
		TreeBranch, vc.ExceedScoreEnabled, vc.ExceedScoreValue))
	fields = append(fields, fmt.Sprintf("           %s [4] Exceeds 2nd place by: enabled=%v, value=%d%%",
		TreeBranch, vc.ExceedSecondPlaceEnabled, vc.ExceedSecondPlaceValue))
	fields = append(fields, fmt.Sprintf("           %s [5] Production capacity: enabled=%v, value=%dk",
		TreeBranch, vc.ProductionCapacityEnabled, vc.ProductionCapacityValue))
	fields = append(fields, fmt.Sprintf("           %s [6] Own capital ships: enabled=%v, value=%d",
		TreeBranch, vc.OwnCapitalShipsEnabled, vc.OwnCapitalShipsValue))
	fields = append(fields, fmt.Sprintf("           %s [7] Highest score after: enabled=%v, value=%d years",
		TreeBranch, vc.HighestScoreYearsEnabled, vc.HighestScoreYearsValue))
	fields = append(fields, fmt.Sprintf("           %s [8] Must meet N criteria: value=%d",
		TreeBranch, vc.NumCriteriaMetValue))
	fields = append(fields, fmt.Sprintf("           %s [9] Min years before winner: value=%d",
		TreeEnd, vc.MinYearsBeforeWinValue))

	// Bytes 32-63: Game name
	gameName := strings.TrimRight(string(d[32:64]), "\x00")
	fields = append(fields, FormatFieldRaw(0x20, 0x3F, "GameName",
		fmt.Sprintf("0x%s...", HexDumpSingleLine(d[32:40])),
		fmt.Sprintf("%q (32 bytes, null-padded)", gameName)))

	// Planet data info
	fields = append(fields, "")
	fields = append(fields, "── Planet Data (follows block) ──")
	fields = append(fields, fmt.Sprintf("  %d planets × 4 bytes = %d trailing bytes",
		pb.PlanetCount, pb.PlanetCount*4))
	fields = append(fields, "  Format per planet: [nameID:10][Y:12][Xoffset:10]")

	// Show a few planets if parsed
	if len(pb.Planets) > 0 {
		fields = append(fields, "")
		fields = append(fields, "── Sample Planets ──")
		maxShow := 5
		if len(pb.Planets) < maxShow {
			maxShow = len(pb.Planets)
		}
		for i := 0; i < maxShow; i++ {
			p := pb.Planets[i]
			fields = append(fields, fmt.Sprintf("  #%d: %s @ (%d, %d)", p.DisplayId, p.Name, p.X, p.Y))
		}
		if len(pb.Planets) > maxShow {
			fields = append(fields, fmt.Sprintf("  ... and %d more planets", len(pb.Planets)-maxShow))
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Game: %q", gameName))
	fields = append(fields, fmt.Sprintf("  Universe: %s, %s density",
		universeSizeName(pb.UniverseSize), densityName(pb.Density)))
	fields = append(fields, fmt.Sprintf("  Players: %d, Planets: %d", pb.PlayerCount, pb.PlanetCount))

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
