package store

import (
	"math"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/data"
)

// ScoreComponents contains the breakdown of a player's score.
type ScoreComponents struct {
	// Final computed score
	Score int

	// Individual components
	PlanetPopScore int // Points from planet populations
	ResourceScore  int // Points from resources (total/30)
	StarbaseScore  int // Points from starbases (count × 3)
	TechScore      int // Points from tech levels
	ShipScore      int // Points from ships (unarmed/escort/capital)

	// Raw counts used in calculation
	TotalResources int // Sum of CResourcesAtPlanet for all owned planets
	StarbaseCount  int // Number of starbases with non-zero hull cost
	PlanetCount    int // Number of owned planets (for ship score capping)

	// Ship counts by category
	UnarmedShips int // Ships with combat power = 0
	EscortShips  int // Ships with 0 < power < 2000
	CapitalShips int // Ships with power >= 2000
}

// ComputeScoreFromActualData computes a player's score from current game data.
//
// This method replicates the game's scoring formula. It can be used to compute
// scores from game data or to analyze score component breakdowns.
//
// Formula: Score = PlanetPopScore + Resources/30 + Starbases×3 + TechScore + ShipScore
//
// Source: Decompiled from UTIL::CalcPlayerScore at MEMORY_UTIL:0x58a6
func (gs *GameStore) ComputeScoreFromActualData(playerNumber int) ScoreComponents {
	var sc ScoreComponents

	player, ok := gs.Player(playerNumber)
	if !ok {
		return sc
	}

	// Get all planets owned by this player
	ownedPlanets := gs.PlanetsByOwner(playerNumber)
	sc.PlanetCount = len(ownedPlanets)

	// 1. Planet Population Score
	// Decompiled formula: popScore = sum(min(6, (population + 999) / 1000))
	// where population is in file units (100s of colonists).
	//
	// DEVIATION FROM DECOMPILED SOURCE:
	// Test data consistently shows expected popScore is +1 higher than the formula produces.
	// Adding a base +1 to match observed game behavior. The source of this discrepancy
	// is unknown - could be version differences or a different code path for score display.
	// See reversing_notes/player-block.md "MYSTERY - Off-by-One Discrepancy" for details.
	if len(ownedPlanets) > 0 {
		sc.PlanetPopScore = 1 // Base +1 bonus (not in decompiled source, but matches observed data)
	}
	for _, planet := range ownedPlanets {
		// Use population in file units (100s of colonists), not actual colonists
		popFileUnits := int(planet.Population / 100)
		popScore := (popFileUnits + 999) / 1000
		if popScore > 6 {
			popScore = 6
		}
		sc.PlanetPopScore += popScore
	}

	// 2. Resource Score
	// Sum CResourcesAtPlanet for all owned planets, divide by 30
	for _, planet := range ownedPlanets {
		sc.TotalResources += gs.CResourcesAtPlanet(planet, player)
	}
	sc.ResourceScore = sc.TotalResources / 30

	// 3. Starbase Score
	// Count starbases with non-zero hull cost × 3
	sc.StarbaseCount = gs.countStarbases(playerNumber)
	sc.StarbaseScore = sc.StarbaseCount * 3

	// 4. Tech Level Score
	// Uses tiered formula from decompiled source - confirmed correct by decompiler team.
	// At low tech levels (sum ~19), tiered ≈ raw sum.
	// At high tech levels the difference is significant:
	//   - Tech sum 19 → tiered ~20
	//   - Tech sum 76 → tiered ~196 (vs raw 76!)
	// Example for tech levels [13,13,13,13,12,12] (sum=76):
	//   Tiered: 34+34+34+34+30+30 = 196 points
	sc.TechScore = calculateTechScore(player.Tech)

	// 5. Ship Score
	sc.UnarmedShips, sc.EscortShips, sc.CapitalShips = gs.countShipsByCategory(playerNumber)
	sc.ShipScore = calculateShipScore(sc.UnarmedShips, sc.EscortShips, sc.CapitalShips, sc.PlanetCount)

	// Total score
	sc.Score = sc.PlanetPopScore + sc.ResourceScore + sc.StarbaseScore + sc.TechScore + sc.ShipScore

	return sc
}

// CResourcesAtPlanet calculates the resources produced at a planet.
// This replicates PLANET::CResourcesAtPlanet at MEMORY_PLANET:0x788e.
//
// Source: Decompiled from stars-decompile/decompiled/io_loadgame.c:526-604
//
// IMPORTANT: The original game stores population in units of 100 colonists.
// Since PlanetEntity.Population is stored as actual colonists (multiplied by 100
// during parsing), we convert back to file units (divide by 100) for calculations.
func (gs *GameStore) CResourcesAtPlanet(planet *PlanetEntity, player *PlayerEntity) int {
	// Step 1: Zero Population Check
	if planet.Population == 0 {
		return 0
	}

	// Step 2: Get Race Stats
	// popEfficiency = rgAttr[0] - Resources per 100 colonists (divisor)
	// factEfficiency = rgAttr[1] - Factory output multiplier
	// factoriesOperate = rgAttr[3] - Factories operable per 100 colonists
	popEfficiency := player.Production.ResourcePerColonist
	factEfficiency := player.Production.FactoryProduction
	prt := player.PRT

	// Convert population to file units (100s of colonists) for calculation
	// The original game stores and calculates with this scale
	popFileUnits := int(planet.Population / 100)

	// Step 3: Overcrowding Adjustment
	// If population exceeds max capacity, excess contributes at 50% efficiency
	// MaxPopulation returns actual colonists, convert to file units for this calculation
	maxPopFileUnits := gs.MaxPopulation(planet, player) / 100
	effectivePop := popFileUnits
	if popFileUnits > maxPopFileUnits && maxPopFileUnits > 0 {
		effectivePop = (popFileUnits-maxPopFileUnits)/2 + maxPopFileUnits
	}

	var resources int

	// Step 4: Resource Calculation (Two Paths)
	if prt == blocks.PRTAlternateReality {
		// Path A: Alternate Reality (AR) Race
		// AR races don't use factories - they use orbital bases instead
		// resources = floor(sqrt((energyTech × population) / popEfficiency))
		energyTech := player.Tech.Energy
		if energyTech < 1 {
			energyTech = 1
		}
		if popEfficiency > 0 {
			resources = int(math.Sqrt(float64(energyTech) * float64(effectivePop) / float64(popEfficiency)))
		}
	} else {
		// Path B: Standard Races (All Other PRTs)
		// popContribution = population / popEfficiency
		// factories = min(actualFactories, maxOperableFactories)
		// factoryContribution = (factories × factEfficiency + 9) / 10
		// resources = popContribution + factoryContribution

		popContribution := 0
		if popEfficiency > 0 {
			popContribution = effectivePop / popEfficiency
		}

		// Calculate operable factories
		maxOperable := gs.CMaxOperableFactories(planet, player)
		factories := planet.Factories
		if factories > maxOperable {
			factories = maxOperable
		}

		// Factory contribution: (factories × factEfficiency + 9) / 10
		// Note: Despite documentation suggesting factories², testing against
		// actual game data shows the formula uses factories (not squared)
		factoryContribution := (factories*factEfficiency + 9) / 10

		resources = popContribution + factoryContribution
	}

	// Step 5: Minimum Guarantee
	// Every inhabited planet produces at least 1 resource
	if resources == 0 {
		resources = 1
	}

	return resources
}

// CMaxOperableFactories calculates how many factories the current population can operate.
// This replicates PLANET::CMaxOperableFactories at MEMORY_PLANET:0x7618.
//
// Formula: maxOperable = min(MaxFactories, (Population × FactoriesOperate) / 100)
// Note: Population is in file units (100s of colonists) for this calculation.
func (gs *GameStore) CMaxOperableFactories(planet *PlanetEntity, player *PlayerEntity) int {
	// AR races can't operate factories
	if player.PRT == blocks.PRTAlternateReality {
		return 0
	}

	// Get race's "factories per 100 colonists" setting
	factoriesOperate := player.Production.FactoriesOperate

	// Convert population to file units (100s of colonists)
	popFileUnits := int(planet.Population / 100)

	// Calculate operable factories based on current population
	// factoriesOperate is per 10k colonists (100 file units)
	maxOperable := popFileUnits * factoriesOperate / 100

	// Cap at planet's max factories
	maxFactories := gs.MaxFactories(planet, player)
	if maxOperable > maxFactories {
		maxOperable = maxFactories
	}

	// Minimum of 1
	if maxOperable < 1 {
		maxOperable = 1
	}

	return maxOperable
}

// PctPlanetDesirability calculates how desirable a planet is for a race.
// This replicates PLANET::PctPlanetDesirability at MEMORY_PLANET:0x6e1e.
//
// Returns:
//   - > 0: Habitable, percentage desirability (0-100+)
//   - = 0: Marginal habitability
//   - < 0: Uninhabitable, negative = penalty points (up to -45)
func (gs *GameStore) PctPlanetDesirability(planet *PlanetEntity, player *PlayerEntity) int {
	// Three environment factors: Gravity, Temperature, Radiation
	planetValues := [3]int{planet.Gravity, planet.Temperature, planet.Radiation}
	raceCenters := [3]int{player.Hab.GravityCenter, player.Hab.TemperatureCenter, player.Hab.RadiationCenter}
	raceLows := [3]int{player.Hab.GravityLow, player.Hab.TemperatureLow, player.Hab.RadiationLow}
	raceHighs := [3]int{player.Hab.GravityHigh, player.Hab.TemperatureHigh, player.Hab.RadiationHigh}

	var pctPos int64
	var pctNeg int64
	var pctMod int64 = 100

	for i := 0; i < 3; i++ {
		planetValue := planetValues[i]
		raceCenter := raceCenters[i]
		raceLow := raceLows[i]
		raceHigh := raceHighs[i]

		// Check for immunity (center value 255 or high < 0 indicates immunity)
		if raceCenter == 255 || raceHigh < 0 {
			// Immune to this factor - full contribution
			pctPos += 10000
			continue
		}

		// Check if planet is outside habitable range
		if planetValue < raceLow || planetValue > raceHigh {
			// Uninhabitable for this factor
			var penalty int
			if planetValue < raceLow {
				penalty = raceLow - planetValue
			} else {
				penalty = planetValue - raceHigh
			}
			if penalty > 15 {
				penalty = 15
			}
			pctNeg += int64(penalty)
			continue
		}

		// Planet is within habitable range
		// Calculate distance from ideal
		var d int // distance from ideal to boundary
		if raceCenter > planetValue {
			d = raceCenter - raceLow
		} else {
			d = raceHigh - raceCenter
		}

		if d == 0 {
			d = 1 // Prevent division by zero
		}

		pctVar := abs(planetValue-raceCenter) * 100 / d
		contribution := (100 - pctVar) * (100 - pctVar)
		pctPos += int64(contribution)

		// Additional penalty if beyond race's "preferred" zone
		distFromIdeal := abs(planetValue - raceCenter)
		dPenalty := distFromIdeal*2 - d
		if dPenalty > 0 {
			pctMod = pctMod * int64(d*2-dPenalty) / int64(d*2)
		}
	}

	if pctNeg > 0 {
		// Uninhabitable: return negative penalty
		return -int(pctNeg)
	}

	// Habitable: return positive percentage
	// result = sqrt(pctPos / 3) * pctMod / 100
	result := int(math.Sqrt(float64(pctPos)/3.0)) * int(pctMod) / 100
	return result
}

// countStarbases counts starbases with non-zero cargo capacity for a player.
// Only starbases with wtCargoMax != 0 count towards the score.
// This means Orbital Fort (cargo capacity = 0) does NOT count,
// only actual starbases with docking capability.
//
// | Hull ID | Name          | Cargo Capacity | Counts? |
// |---------|---------------|----------------|---------|
// | 32      | Orbital Fort  | 0              | No      |
// | 33      | Space Dock    | 200            | Yes     |
// | 34      | Space Station | 65535          | Yes     |
// | 35      | Ultra Station | 65535          | Yes     |
// | 36      | Death Star    | 65535          | Yes     |
func (gs *GameStore) countStarbases(playerNumber int) int {
	count := 0
	ownedPlanets := gs.PlanetsByOwner(playerNumber)

	for _, planet := range ownedPlanets {
		if planet.HasStarbase {
			// Check if starbase has non-zero cargo capacity
			if design, ok := gs.StarbaseDesign(playerNumber, planet.StarbaseDesign); ok {
				if hull := design.Hull(); hull != nil && hull.CargoCapacity > 0 {
					count++
				}
			} else {
				// If we can't find the design, assume it counts
				count++
			}
		}
	}

	return count
}

// calculateTechScoreRawSum computes the score contribution from tech levels
// using raw sum of levels.
//
// NOTE: This was initially used based on early test scenarios where tiered ≈ raw sum.
// Decompiler team confirmed the tiered formula is correct - at high tech levels
// the difference is significant. Use calculateTechScore instead.
//
//nolint:unused // Kept for reference, tiered formula is the correct one
func calculateTechScoreRawSum(tech TechLevels) int {
	return tech.Energy + tech.Weapons + tech.Propulsion +
		tech.Construction + tech.Electronics + tech.Biotech
}

// calculateTechScore computes the score contribution from tech levels
// using a tiered formula with a bonus for advanced technology.
//
// Tech scoring uses tiered rates:
//   - Levels 0-4: +level points per level (1, 2, 3, 4)
//   - Levels 5-6: +level×2 - 4 points (6, 8)
//   - Levels 7-9: +level×3 - 9 points (12, 15, 18)
//   - Levels 10+: +level×4 - 18 points (22, 26, 30, ...)
//
// Additionally, a +1 bonus is awarded if any tech level reaches 10 or higher,
// representing "advanced technology" status.
func calculateTechScore(tech TechLevels) int {
	total := 0
	maxLevel := 0
	levels := []int{
		tech.Energy,
		tech.Weapons,
		tech.Propulsion,
		tech.Construction,
		tech.Electronics,
		tech.Biotech,
	}

	for _, level := range levels {
		total += techLevelScore(level)
		if level > maxLevel {
			maxLevel = level
		}
	}

	// +1 bonus for having any tech at level 10 or higher
	if maxLevel >= 10 {
		total++
	}

	return total
}

// techLevelScore returns the score contribution for a single tech level.
// Uses tier boundaries at 4, 6, 9:
//   - Levels 0-4: raw level value (0, 1, 2, 3, 4)
//   - Levels 5-6: level×2 - 4 (6, 8)
//   - Levels 7-9: level×3 - 9 (12, 15, 18)
//   - Levels 10+: level×4 - 18 (22, 26, 30, 34, ...)
func techLevelScore(level int) int {
	if level <= 4 {
		return level
	}
	if level <= 6 {
		return level*2 - 4
	}
	if level <= 9 {
		return level*3 - 9
	}
	return level*4 - 18
}

// countShipsByCategory counts ships by combat power category.
// Returns (unarmed, escort, capital) counts.
//   - Unarmed: power = 0
//   - Escort: 0 < power < 2000
//   - Capital: power >= 2000
func (gs *GameStore) countShipsByCategory(playerNumber int) (unarmed, escort, capital int) {
	fleets := gs.FleetsByOwner(playerNumber)

	for _, fleet := range fleets {
		designs := fleet.GetDesigns(gs)
		for _, info := range designs {
			if info.Design == nil || info.Count <= 0 {
				continue
			}

			power := info.Design.GetCombatPower()
			count := info.Count

			switch {
			case power == 0:
				unarmed += count
			case power < 2000:
				escort += count
			default:
				capital += count
			}
		}
	}

	return unarmed, escort, capital
}

// calculateShipScore computes the score contribution from ships.
// Ships are capped based on planet count and use different scoring:
//   - Unarmed: capped to planetCount, contributes count/2
//   - Escort: capped to planetCount, contributes count
//   - Capital: uses diminishing returns formula
func calculateShipScore(unarmed, escort, capital, planetCount int) int {
	// Cap unarmed and escort to planet count
	unarmedCapped := unarmed
	if unarmedCapped > planetCount {
		unarmedCapped = planetCount
	}

	escortCapped := escort
	if escortCapped > planetCount {
		escortCapped = planetCount
	}

	// Unarmed contribute half, escort contribute full
	score := unarmedCapped/2 + escortCapped

	// Capital ships use diminishing returns:
	// capitalScore = (planetCount × capitalCount) / (planetCount + capitalCount)
	if capital > 0 && planetCount > 0 {
		score += (planetCount * capital) / (planetCount + capital)
	}

	return score
}

// GetCombatPower calculates the combat power of a design.
// This replicates the LComputePower() function from the original game.
//
// Source: Decompiled from UTIL::LComputePower at MEMORY_UTIL:0x0b32
//
// The function calculates damage potential from:
// - Beam weapons (grhst = 0x10): dp × count × (range + 3), halved for sappers
// - Torpedoes (grhst = 0x20): dp × count × (range - 2)
// - Bombs (grhst = 0x40): (killRate + structureKill) × count × 2
// - Capacitors multiply beam damage (capped at 255%)
// - Speed bonus: dpBeams × (speed - 4)
func (d *DesignEntity) GetCombatPower() int {
	if d.designBlock == nil {
		return 0
	}

	var dpBeams, dpTorps, dpBombs int
	var engineID, engineCount int
	var thrusters, halfThrusters int
	var pctCap = 1000 // Start at 100.0% (scaled by 10 for precision)

	// First pass: collect all weapon damage and equipment info
	for _, slot := range d.designBlock.Slots {
		if slot.Count == 0 {
			continue
		}

		itemID := slot.ItemId + 1 // Convert 0-indexed slot to 1-indexed data lookup
		count := slot.Count

		switch slot.Category {
		case blocks.ItemCategoryEngine:
			engineID = itemID
			engineCount += count
			// Interspace-10 gives +0.5 speed per engine
			if itemID == data.EngineInterspace10 {
				halfThrusters += count
			}

		case blocks.ItemCategoryBeamWeapon:
			if beam := data.GetBeamWeapon(itemID); beam != nil {
				// beamPower = dp × count × (range + 3)
				beamPower := beam.Power * count * (beam.Range + 3)
				// Sappers (IsSapper or IsGatling) have power halved
				if beam.IsSapper || beam.IsGatling {
					beamPower /= 2
				}
				dpBeams += beamPower
			}

		case blocks.ItemCategoryTorpedo:
			if torpedo := data.GetTorpedo(itemID); torpedo != nil {
				// torpPower = dp × count × (range - 2)
				torpPower := torpedo.Power * count * (torpedo.Range - 2)
				if torpPower > 0 {
					dpTorps += torpPower
				}
			}

		case blocks.ItemCategoryBomb:
			if bomb := data.GetBomb(itemID); bomb != nil {
				// bombPower = (killRate + structureKill) × count × 2
				dpBombs += (bomb.KillRate + bomb.StructureKill) * count * 2
			}

		case blocks.ItemCategoryElectrical:
			// PRESERVED BUG: Code checks items 12 (Jammer 50) and 13 (Energy Capacitor)
			// but NOT 14 (Flux Capacitor). Including Jammer 50 appears to be a bug.
			if itemID == data.ElecJammer50 || itemID == data.ElecEnergyCapacitor {
				if elec := data.GetElectrical(itemID); elec != nil {
					// Capacitor bonus stacks multiplicatively
					// pctCap = pctCap × (100 + bonus) / 100
					bonus := elec.CapacitorBonus
					if itemID == data.ElecJammer50 {
						// Jammer 50 uses BeamDeflection as the "bonus" (bug)
						bonus = elec.BeamDeflection
					}
					for i := 0; i < count; i++ {
						pctCap = pctCap * (100 + bonus) / 100
					}
				}
			}
			// Thruster: +1 speed each
			if itemID == data.ElecThruster {
				thrusters += count
			}

		case blocks.ItemCategoryMiningRobot:
			// Robo-Ultra-Miner: +0.5 speed each (hardcoded in original SpdOfShip)
			if itemID == data.MiningRoboUltra {
				halfThrusters += count
			}

		case blocks.ItemCategoryMechanical:
			// Maneuver Jet: +1 speed each
			// Overthruster: +2 speed each
			switch itemID {
			case data.MechManeuveringJet:
				thrusters += count
			case data.MechOverthruster:
				thrusters += count * 2
			}
		}
	}

	// Cap capacitor bonus at 255%
	if pctCap > 2550 {
		pctCap = 2550
	}

	// Apply capacitor bonus to beam damage
	dpBeams = dpBeams * pctCap / 1000

	// Calculate ship speed for speed bonus
	speed := d.calculateSpeed(engineID, engineCount, thrusters, halfThrusters)

	// Speed bonus: dpBeams × (speed - 4)
	// Speed 4 is baseline (no bonus)
	speedBonus := dpBeams * (speed - 4)

	// Total power
	totalPower := dpBombs + dpBeams + speedBonus + dpTorps

	return totalPower
}

// calculateSpeed calculates combat speed for a ship design.
// This replicates SpdOfShip() when called from LComputePower (fleet=NULL).
//
// Source: Decompiled from BATTLE::SpdOfShip
func (d *DesignEntity) calculateSpeed(engineID, engineCount, thrusters, halfThrusters int) int {
	if engineCount == 0 {
		return 0
	}

	// Determine base warp
	// Hardcoded warp-10 engines: IDs 7, 8, 9, 14, 15
	// - 7: Trans-Galactic Drive
	// - 8: Interspace-10
	// - 9: Enigma Pulsar
	// - 14: Trans-Galactic Super Scoop
	// - 15: Trans-Galactic Mizer Scoop
	var baseWarp int
	switch engineID {
	case data.EngineTransGalacticDrive,
		data.EngineInterspace10,
		data.EngineEnigmaPulsar,
		data.EngineTransGalacticSuperScoop,
		data.EngineTransGalacticMizerScoop:
		baseWarp = 10
	default:
		// Find highest warp where fuel usage ≤ 120mg
		if engine := data.GetEngine(engineID); engine != nil {
			baseWarp = 9
			for baseWarp > 0 && engine.FuelPerMg[baseWarp] > 120 {
				baseWarp--
			}
		}
	}

	// Calculate base speed
	// speed = baseWarp - 4 + thrusters + (halfThrusters + 1) / 2
	speed := baseWarp - 4 + thrusters + (halfThrusters+1)/2

	// Mass penalty: (mass / 70) / engineCount
	// In score context, we only use hull empty weight
	if hull := d.Hull(); hull != nil {
		massPenalty := (hull.Mass / 70) / engineCount
		speed -= massPenalty
	}

	// Clamp to 0-8
	if speed > 8 {
		speed = 8
	}
	if speed < 0 {
		speed = 0
	}

	return speed
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
