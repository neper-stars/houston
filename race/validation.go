package race

// Validate checks a race configuration and returns any validation errors.
func Validate(r *Race) []ValidationError {
	var errors []ValidationError

	// Name validation
	if r.SingularName == "" {
		errors = append(errors, ValidationError{
			Field:   "SingularName",
			Message: "singular name is required",
		})
	} else if len(r.SingularName) > 32 {
		errors = append(errors, ValidationError{
			Field:   "SingularName",
			Message: "singular name must be at most 32 characters",
		})
	}

	if r.PluralName == "" {
		errors = append(errors, ValidationError{
			Field:   "PluralName",
			Message: "plural name is required",
		})
	} else if len(r.PluralName) > 32 {
		errors = append(errors, ValidationError{
			Field:   "PluralName",
			Message: "plural name must be at most 32 characters",
		})
	}

	// PRT validation
	if r.PRT < 0 || r.PRT > 9 {
		errors = append(errors, ValidationError{
			Field:   "PRT",
			Message: "PRT must be between 0 and 9",
		})
	}

	// LRT validation (bits 0-13 only)
	if r.LRT&^uint16(0x3FFF) != 0 {
		errors = append(errors, ValidationError{
			Field:   "LRT",
			Message: "LRT contains invalid bits (only bits 0-13 are valid)",
		})
	}

	// Habitability validation
	if !r.GravityImmune {
		if r.GravityCenter < 0 || r.GravityCenter > 100 {
			errors = append(errors, ValidationError{
				Field:   "GravityCenter",
				Message: "gravity center must be between 0 and 100",
			})
		}
		if r.GravityWidth < 0 || r.GravityWidth > 50 {
			errors = append(errors, ValidationError{
				Field:   "GravityWidth",
				Message: "gravity width must be between 0 and 50",
			})
		}
	}

	if !r.TemperatureImmune {
		if r.TemperatureCenter < 0 || r.TemperatureCenter > 100 {
			errors = append(errors, ValidationError{
				Field:   "TemperatureCenter",
				Message: "temperature center must be between 0 and 100",
			})
		}
		if r.TemperatureWidth < 0 || r.TemperatureWidth > 50 {
			errors = append(errors, ValidationError{
				Field:   "TemperatureWidth",
				Message: "temperature width must be between 0 and 50",
			})
		}
	}

	if !r.RadiationImmune {
		if r.RadiationCenter < 0 || r.RadiationCenter > 100 {
			errors = append(errors, ValidationError{
				Field:   "RadiationCenter",
				Message: "radiation center must be between 0 and 100",
			})
		}
		if r.RadiationWidth < 0 || r.RadiationWidth > 50 {
			errors = append(errors, ValidationError{
				Field:   "RadiationWidth",
				Message: "radiation width must be between 0 and 50",
			})
		}
	}

	// At least one habitability dimension must be non-immune
	if r.GravityImmune && r.TemperatureImmune && r.RadiationImmune {
		errors = append(errors, ValidationError{
			Field:   "Habitability",
			Message: "at least one habitability dimension must be non-immune",
		})
	}

	// Growth rate validation
	if r.GrowthRate < 1 || r.GrowthRate > 20 {
		errors = append(errors, ValidationError{
			Field:   "GrowthRate",
			Message: "growth rate must be between 1 and 20",
		})
	}

	// Economy validation
	if r.ColonistsPerResource < 700 || r.ColonistsPerResource > 2500 {
		errors = append(errors, ValidationError{
			Field:   "ColonistsPerResource",
			Message: "colonists per resource must be between 700 and 2500",
		})
	}

	if r.FactoryOutput < 5 || r.FactoryOutput > 25 {
		errors = append(errors, ValidationError{
			Field:   "FactoryOutput",
			Message: "factory output must be between 5 and 25",
		})
	}

	if r.FactoryCost < 5 || r.FactoryCost > 25 {
		errors = append(errors, ValidationError{
			Field:   "FactoryCost",
			Message: "factory cost must be between 5 and 25",
		})
	}

	if r.FactoryCount < 5 || r.FactoryCount > 25 {
		errors = append(errors, ValidationError{
			Field:   "FactoryCount",
			Message: "factory count must be between 5 and 25",
		})
	}

	if r.MineOutput < 5 || r.MineOutput > 25 {
		errors = append(errors, ValidationError{
			Field:   "MineOutput",
			Message: "mine output must be between 5 and 25",
		})
	}

	if r.MineCost < 2 || r.MineCost > 15 {
		errors = append(errors, ValidationError{
			Field:   "MineCost",
			Message: "mine cost must be between 2 and 15",
		})
	}

	if r.MineCount < 5 || r.MineCount > 25 {
		errors = append(errors, ValidationError{
			Field:   "MineCount",
			Message: "mine count must be between 5 and 25",
		})
	}

	// Research cost validation
	validateResearchCost := func(fieldName string, value int) {
		if value < 0 || value > 2 {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: "research cost must be 0 (Extra), 1 (Standard), or 2 (Less)",
			})
		}
	}

	validateResearchCost("ResearchEnergy", r.ResearchEnergy)
	validateResearchCost("ResearchWeapons", r.ResearchWeapons)
	validateResearchCost("ResearchPropulsion", r.ResearchPropulsion)
	validateResearchCost("ResearchConstruction", r.ResearchConstruction)
	validateResearchCost("ResearchElectronics", r.ResearchElectronics)
	validateResearchCost("ResearchBiotech", r.ResearchBiotech)

	// LeftoverPointsOn validation
	if r.LeftoverPointsOn < LeftoverSurfaceMinerals || r.LeftoverPointsOn > LeftoverMineralConcentration {
		errors = append(errors, ValidationError{
			Field:   "LeftoverPointsOn",
			Message: "invalid leftover points allocation option",
		})
	}

	return errors
}
