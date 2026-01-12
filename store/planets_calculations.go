package store

import "github.com/neper-stars/houston/blocks"

// This file contains planet-related calculations for population, factories, mines, and defenses.
// These calculations replicate the original Stars! game formulas.

// MaxPopulation calculates the maximum population a planet can support for a given race.
// This replicates PLANET::CalcPlanetMaxPop at MEMORY_PLANET:0x7096.
func (gs *GameStore) MaxPopulation(planet *PlanetEntity, player *PlayerEntity) int {
	prt := player.PRT

	// AR races can only have population at planets with their own starbases
	if prt == blocks.PRTAlternateReality {
		if planet.Owner != player.PlayerNumber || !planet.HasStarbase {
			return 0
		}
		// Max pop = starbase hull capacity × 4
		// We need the starbase design to get hull capacity
		if design, ok := gs.StarbaseDesign(player.PlayerNumber, planet.StarbaseDesign); ok {
			if hull := design.Hull(); hull != nil {
				// Hull BaseCapacity is stored with an offset of 0x20 in the original
				// The formula is: (hull.baseCapacity - 0x20) * 4
				// For simplicity, we use the hull's cargo capacity as a proxy
				// Note: This may need adjustment based on actual hull data structure
				return hull.CargoCapacity * 4
			}
		}
		return 0
	}

	// Standard race calculation
	pctDesire := gs.PctPlanetDesirability(planet, player)

	var maxPop int
	if pctDesire < 5 {
		maxPop = 500 // Minimum for barely habitable
	} else {
		maxPop = pctDesire * 100 // Base: 100 colonists per % desirability
	}

	// PRT Modifiers
	switch prt {
	case blocks.PRTHyperExpansion:
		// HE: -50% capacity
		maxPop -= maxPop / 2
	case blocks.PRTJackOfAllTrades:
		// JOAT: +20% capacity
		maxPop += maxPop / 5
	}

	// LRT Modifier: OBRM (Only Basic Remote Mining) = +10%
	if player.HasLRT(blocks.LRTOnlyBasicRemoteMining) {
		maxPop += maxPop / 10
	}

	return maxPop
}

// MaxFactories calculates the maximum number of factories a planet can support.
// This replicates PLANET::CMaxFactories at MEMORY_PLANET:0x755c.
//
// Formula: maxFactories = max(10, (MaxPopulation × FactoriesOperate) / 100)
func (gs *GameStore) MaxFactories(planet *PlanetEntity, player *PlayerEntity) int {
	// AR races can't have factories
	if player.PRT == blocks.PRTAlternateReality {
		return 0
	}

	// Get max population for this planet
	maxPop := gs.MaxPopulation(planet, player)

	// Get race's "factories per 100 colonists" setting
	factoriesOperate := player.Production.FactoriesOperate

	// Calculate max factories based on max population
	maxFactories := maxPop * factoriesOperate / 100

	// Minimum of 10 factories
	if maxFactories < 10 {
		maxFactories = 10
	}

	return maxFactories
}

// MaxDefenses calculates the absolute maximum defenses based on planet habitability.
// This replicates CMaxDefenses (FUN_1048_5714) from the original game.
//
// Formula: clamp(habitability% * 4, 10, 100)
// AR races return 0 (no planetary defenses).
func (gs *GameStore) MaxDefenses(planet *PlanetEntity, player *PlayerEntity) int {
	// AR races can't have planetary defenses
	if player.PRT == blocks.PRTAlternateReality {
		return 0
	}

	// Get habitability percentage for this planet/race combination
	habitability := gs.PctPlanetDesirability(planet, player)

	// Calculate max defenses: habitability * 4, clamped to [10, 100]
	maxDef := habitability * 4
	if maxDef < 10 {
		maxDef = 10
	}
	if maxDef > 100 {
		maxDef = 100
	}

	return maxDef
}

// MaxOperableDefenses calculates the population-limited defenses that can actually operate.
// This replicates CMaxOperableDefenses (FUN_1048_5768) from the original game.
//
// Formula: min(MaxDefenses, min(1000, (population_in_file_units + 24) / 25))
// Where population_in_file_units = actual_colonists / 100
//
// AR races return 0 (no planetary defenses).
func (gs *GameStore) MaxOperableDefenses(planet *PlanetEntity, player *PlayerEntity) int {
	// AR races can't have planetary defenses
	if player.PRT == blocks.PRTAlternateReality {
		return 0
	}

	// Get habitability-based maximum
	maxDef := gs.MaxDefenses(planet, player)

	// Calculate population-based limit
	// Population in entity is actual colonists, file uses 100s of colonists
	popInFileUnits := int(planet.Population / 100)
	popLimit := (popInFileUnits + 24) / 25

	// Cap population limit at 1000
	if popLimit > 1000 {
		popLimit = 1000
	}

	// Return the minimum of habitability limit and population limit
	if maxDef < popLimit {
		return maxDef
	}
	return popLimit
}
