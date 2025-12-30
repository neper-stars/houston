package race

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
)

const (
	// MinHabWidth is the minimum habitability width allowed by the Stars! Race Wizard.
	// The narrowest range the GUI allows is 0.12g to 0.31g (internal 0-19), which is width 10.
	MinHabWidth = 10

	// MaxHabWidth is the maximum habitability width (half the full range)
	MaxHabWidth = 50
)

// Validate checks a race configuration and returns any validation errors.
// If finalize is true, also validates that advantage points are >= 0.
// Use finalize=true when validating a completed race (e.g., from a file or at submission).
// Use finalize=false (or omit) during incremental building where points may temporarily be negative.
func Validate(r *Race, finalize ...bool) []ValidationError {
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

	// Plural name is optional (Stars! allows empty plural names in predefined races)
	if len(r.PluralName) > 32 {
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
		if r.GravityWidth < MinHabWidth || r.GravityWidth > MaxHabWidth {
			errors = append(errors, ValidationError{
				Field:   "GravityWidth",
				Message: "gravity width must be between 10 and 50",
			})
		}
		// Range edge constraints: the habitable range [center-width, center+width] must stay within [0, 100]
		if lowEdge := r.GravityCenter - r.GravityWidth; lowEdge < 0 {
			errors = append(errors, ValidationError{
				Field:   "GravityCenter",
				Message: fmt.Sprintf("gravity range low edge would be %.2fg (below minimum 0.12g)", blocks.GravityToDisplay(lowEdge)),
			})
		}
		if highEdge := r.GravityCenter + r.GravityWidth; highEdge > 100 {
			errors = append(errors, ValidationError{
				Field:   "GravityCenter",
				Message: fmt.Sprintf("gravity range high edge would be %.2fg (above maximum 8.00g)", blocks.GravityToDisplay(highEdge)),
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
		if r.TemperatureWidth < MinHabWidth || r.TemperatureWidth > MaxHabWidth {
			errors = append(errors, ValidationError{
				Field:   "TemperatureWidth",
				Message: "temperature width must be between 10 and 50",
			})
		}
		// Range edge constraints: the habitable range [center-width, center+width] must stay within [0, 100]
		if lowEdge := r.TemperatureCenter - r.TemperatureWidth; lowEdge < 0 {
			errors = append(errors, ValidationError{
				Field:   "TemperatureCenter",
				Message: fmt.Sprintf("temperature range low edge would be %d°C (below minimum -200°C)", blocks.TemperatureToDisplay(lowEdge)),
			})
		}
		if highEdge := r.TemperatureCenter + r.TemperatureWidth; highEdge > 100 {
			errors = append(errors, ValidationError{
				Field:   "TemperatureCenter",
				Message: fmt.Sprintf("temperature range high edge would be %d°C (above maximum 200°C)", blocks.TemperatureToDisplay(highEdge)),
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
		if r.RadiationWidth < MinHabWidth || r.RadiationWidth > MaxHabWidth {
			errors = append(errors, ValidationError{
				Field:   "RadiationWidth",
				Message: "radiation width must be between 10 and 50",
			})
		}
		// Range edge constraints: the habitable range [center-width, center+width] must stay within [0, 100]
		if lowEdge := r.RadiationCenter - r.RadiationWidth; lowEdge < 0 {
			errors = append(errors, ValidationError{
				Field:   "RadiationCenter",
				Message: fmt.Sprintf("radiation range low edge would be %dmR (below minimum 0mR)", blocks.RadiationToDisplay(lowEdge)),
			})
		}
		if highEdge := r.RadiationCenter + r.RadiationWidth; highEdge > 100 {
			errors = append(errors, ValidationError{
				Field:   "RadiationCenter",
				Message: fmt.Sprintf("radiation range high edge would be %dmR (above maximum 100mR)", blocks.RadiationToDisplay(highEdge)),
			})
		}
	}

	// Note: All three immunities are allowed (e.g., Silicanoid predefined race)

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

	// Points validation (only when finalizing)
	if len(finalize) > 0 && finalize[0] {
		points := CalculatePoints(r)
		if points < 0 {
			errors = append(errors, ValidationError{
				Field:   "Points",
				Message: fmt.Sprintf("race has negative advantage points (%d)", points),
			})
		}
	}

	return errors
}
