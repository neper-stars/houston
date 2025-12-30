package store

import (
	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/password"
	"github.com/neper-stars/houston/race"
)

// CreateRaceFile generates a complete .r1-.r16 file from a race configuration.
// This creates the file from scratch (not from an existing source).
// The playerSlot should be 1-16 (the extension suffix).
func CreateRaceFile(r *race.Race, playerSlot int) ([]byte, error) {
	_ = playerSlot // Player slot is only used for filename, not stored in header
	writer := NewFileWriter()

	// 1. Create FileHeader for race file using block encoder
	header := blocks.NewFileHeaderForRaceFile()
	result := writer.WriteHeader(header)

	// Get salt from header for encryption
	salt := header.Salt()

	// 2. Init encryption (gameId=0, turn=0, playerIndex=31)
	writer.InitEncryption(salt, 0, 0, 31, 0)

	// 3. Build PlayerBlock from Race and encode it
	playerBlock := raceToPlayerBlock(r)
	playerBlockData, err := playerBlock.Encode()
	if err != nil {
		return nil, err
	}
	result = append(result, writer.WriteEncryptedBlock(blocks.PlayerBlockType, playerBlockData)...)

	// 4. Compute and write footer
	footerData := blocks.ComputeRaceFooter(playerBlockData, r.SingularName, r.PluralName)
	result = append(result, writer.WriteFooter(true, footerData)...)

	return result, nil
}

// raceToPlayerBlock converts a race configuration to a PlayerBlock.
// Race files always have FullDataFlag set and PlayerNumber=255.
// The returned PlayerBlock can be encoded using its Encode() method.
func raceToPlayerBlock(r *race.Race) *blocks.PlayerBlock {
	pb := &blocks.PlayerBlock{
		PlayerNumber:    255, // Not assigned to a player yet
		FullDataFlag:    true,
		Logo:            r.Icon & 0x1F,
		Byte7:           0x00, // No AI for race files
		NameSingular:    r.SingularName,
		NamePlural:      r.PluralName,
		PlayerRelations: nil, // Empty for race files

		// Race data
		Homeworld: 0,
		Rank:      0,
		Hab: blocks.Habitability{
			GravityCenter:     habCenter(r.GravityCenter, r.GravityImmune),
			TemperatureCenter: habCenter(r.TemperatureCenter, r.TemperatureImmune),
			RadiationCenter:   habCenter(r.RadiationCenter, r.RadiationImmune),
			GravityLow:        habLow(r.GravityCenter, r.GravityWidth, r.GravityImmune),
			TemperatureLow:    habLow(r.TemperatureCenter, r.TemperatureWidth, r.TemperatureImmune),
			RadiationLow:      habLow(r.RadiationCenter, r.RadiationWidth, r.RadiationImmune),
			GravityHigh:       habHigh(r.GravityCenter, r.GravityWidth, r.GravityImmune),
			TemperatureHigh:   habHigh(r.TemperatureCenter, r.TemperatureWidth, r.TemperatureImmune),
			RadiationHigh:     habHigh(r.RadiationCenter, r.RadiationWidth, r.RadiationImmune),
		},
		GrowthRate: r.GrowthRate,

		// Tech levels (0 for race files)
		Tech: blocks.TechLevels{},

		// Research settings
		ResearchPercentage:   15, // Default 15% like Stars!
		CurrentResearchField: 0,
		NextResearchField:    0,

		// Production settings
		Production: blocks.ProductionSettings{
			ResourcePerColonist: r.ColonistsPerResource / 100,
			FactoryProduction:   r.FactoryOutput,
			FactoryCost:         r.FactoryCost,
			FactoriesOperate:    r.FactoryCount,
			MineProduction:      r.MineOutput,
			MineCost:            r.MineCost,
			MinesOperate:        r.MineCount,
		},

		SpendLeftoverPoints: int(r.LeftoverPointsOn),

		// Research costs
		ResearchCost: blocks.ResearchCosts{
			Energy:       r.ResearchEnergy,
			Weapons:      r.ResearchWeapons,
			Propulsion:   r.ResearchPropulsion,
			Construction: r.ResearchConstruction,
			Electronics:  r.ResearchElectronics,
			Biotech:      r.ResearchBiotech,
		},

		PRT: r.PRT,
		LRT: r.LRT,

		ExpensiveTechStartsAt3: r.TechsStartHigh,
		FactoriesCost1LessGerm: r.FactoriesUseLessGerm,

		MTItems: 0, // Always 0 for race files
	}

	// Set password hash if provided
	if r.Password != "" {
		pb.PasswordHash = password.HashRacePassword(r.Password)
	}

	return pb
}

// PlayerBlockToRace converts a PlayerBlock back to a Race configuration.
// This is the reverse of raceToPlayerBlock and is useful for validating
// race files loaded from disk.
// Note: Password cannot be recovered (only the hash is stored).
func PlayerBlockToRace(pb *blocks.PlayerBlock) *race.Race {
	gravImmune := pb.Hab.GravityCenter == 255
	tempImmune := pb.Hab.TemperatureCenter == 255
	radImmune := pb.Hab.RadiationCenter == 255

	r := &race.Race{
		SingularName: pb.NameSingular,
		PluralName:   pb.NamePlural,
		Icon:         pb.Logo,
		// Password cannot be recovered from hash

		PRT: pb.PRT,
		LRT: pb.LRT,

		GravityImmune:     gravImmune,
		GravityCenter:     habCenterFromBlock(pb.Hab.GravityCenter, pb.Hab.GravityLow, pb.Hab.GravityHigh, gravImmune),
		GravityWidth:      habWidthFromBlock(pb.Hab.GravityLow, pb.Hab.GravityHigh, gravImmune),
		TemperatureImmune: tempImmune,
		TemperatureCenter: habCenterFromBlock(pb.Hab.TemperatureCenter, pb.Hab.TemperatureLow, pb.Hab.TemperatureHigh, tempImmune),
		TemperatureWidth:  habWidthFromBlock(pb.Hab.TemperatureLow, pb.Hab.TemperatureHigh, tempImmune),
		RadiationImmune:   radImmune,
		RadiationCenter:   habCenterFromBlock(pb.Hab.RadiationCenter, pb.Hab.RadiationLow, pb.Hab.RadiationHigh, radImmune),
		RadiationWidth:    habWidthFromBlock(pb.Hab.RadiationLow, pb.Hab.RadiationHigh, radImmune),

		GrowthRate:           pb.GrowthRate,
		ColonistsPerResource: pb.Production.ResourcePerColonist * 100,

		FactoryOutput:        pb.Production.FactoryProduction,
		FactoryCost:          pb.Production.FactoryCost,
		FactoryCount:         pb.Production.FactoriesOperate,
		FactoriesUseLessGerm: pb.FactoriesCost1LessGerm,

		MineOutput: pb.Production.MineProduction,
		MineCost:   pb.Production.MineCost,
		MineCount:  pb.Production.MinesOperate,

		ResearchEnergy:       pb.ResearchCost.Energy,
		ResearchWeapons:      pb.ResearchCost.Weapons,
		ResearchPropulsion:   pb.ResearchCost.Propulsion,
		ResearchConstruction: pb.ResearchCost.Construction,
		ResearchElectronics:  pb.ResearchCost.Electronics,
		ResearchBiotech:      pb.ResearchCost.Biotech,
		TechsStartHigh:       pb.ExpensiveTechStartsAt3,

		LeftoverPointsOn: race.LeftoverPointsOption(pb.SpendLeftoverPoints),
	}

	return r
}

// habCenterFromBlock extracts the center value from block data.
// If immune, returns a default value (50).
func habCenterFromBlock(blockCenter, blockLow, blockHigh int, immune bool) int {
	if immune {
		return 50 // Default center for immune (doesn't matter for gameplay)
	}
	// The block stores the actual center
	return blockCenter
}

// habWidthFromBlock calculates the width from low and high values.
// Width = (high - low) / 2
func habWidthFromBlock(blockLow, blockHigh int, immune bool) int {
	if immune {
		return 50 // Default width for immune (doesn't matter for gameplay)
	}
	return (blockHigh - blockLow) / 2
}

// habCenter returns 255 for immune, otherwise the center value.
func habCenter(center int, immune bool) int {
	if immune {
		return 255
	}
	return center
}

// habLow returns 255 for immune, otherwise center - width (clamped to 0).
func habLow(center, width int, immune bool) int {
	if immune {
		return 255
	}
	low := center - width
	if low < 0 {
		return 0
	}
	return low
}

// habHigh returns 255 for immune, otherwise center + width (clamped to 100).
func habHigh(center, width int, immune bool) int {
	if immune {
		return 255
	}
	high := center + width
	if high > 100 {
		return 100
	}
	return high
}
