package race

import (
	"math"

	"github.com/neper-stars/houston/data"
)

// =============================================================================
// Race Point Calculation Constants
// These values are from the Stars! game engine and define the point costs
// and bonuses for various race attributes.
// =============================================================================

// Starting advantage points before adjustments.
const raceStartingPoints = 1650

// LRT indices for point calculation.
const (
	lrtTotalTerraforming = 1  // TT - Total Terraforming
	lrtNAS               = 10 // NAS - No Advanced Scanners
	lrtCount             = 14 // Total number of LRTs
)

// PRT indices for point calculation.
const (
	prtHE   = 0 // Hyper Expansion
	prtSS   = 1 // Super Stealth
	prtPP   = 6 // Packet Physics
	prtAR   = 8 // Alternate Reality
	prtJoaT = 9 // Jack of All Trades
)

// Habitability calculation constants.
const (
	habPointsDivisor           = 2000 // Divisor for hab range points
	habGrowthPenaltyDivisor    = 24   // Divisor for growth rate penalty on hab
	habCenterIdeal             = 50   // Ideal center value for off-center bonus
	habOffCenterBonusPerPoint  = 4    // Points per unit away from center
	habMultipleImmunityPenalty = 150  // Penalty for 2+ immunities
	habMaxValue                = 100  // Maximum habitability value
	habImmuneIterations        = 1    // Iteration count for immune dimensions
	habNonImmuneIterations     = 11   // Iteration count for non-immune dimensions
	habImmuneTestWidth         = 11   // Test width for immune dimensions
	habImmuneTestCenter        = 50   // Test center for immune dimensions
	habPerfectValue            = 10000
	habRedMaxPenalty           = 15 // Maximum red planet penalty per dimension
)

// Terraforming correction factors.
const (
	ttCorrectionLoop1WithTT    = 8  // TT correction factor for loop 1 with TT
	ttCorrectionLoop1WithoutTT = 5  // TT correction factor for loop 1 without TT
	ttCorrectionLoop2WithTT    = 17 // TT correction factor for loop 2 with TT
	ttCorrectionLoop2WithoutTT = 15 // TT correction factor for loop 2 without TT
)

// Loop multipliers for hab range calculation.
const (
	habLoopMultiplier0 = 7 // Multiplier for loop index 0
	habLoopMultiplier1 = 5 // Multiplier for loop index 1
	habLoopMultiplier2 = 6 // Multiplier for loop index 2
)

// Growth rate point values.
const (
	growthRateBaseMultiplier = 4200 // Points per growth rate below 6
	growthRateBonus6         = 3600 // Bonus points for growth rate 6
	growthRateBonus7         = 2250 // Bonus points for growth rate 7
	growthRateBonus8         = 600  // Bonus points for growth rate 8
	growthRateBonus9         = 225  // Bonus points for growth rate 9
	growthRateMaxFactor      = 45   // Maximum growth factor for penalty calc
)

// Population efficiency thresholds and values.
const (
	popEfficiencyDivisor    = 100  // Divisor for colonists per resource
	popEfficiencyMax        = 25   // Maximum effective pop efficiency
	popEfficiencyPenalty7   = 2400 // Penalty for efficiency <= 7
	popEfficiencyPenalty8   = 1260 // Penalty for efficiency == 8
	popEfficiencyPenalty9   = 600  // Penalty for efficiency == 9
	popEfficiencyBonusPer   = 120  // Bonus per efficiency above 10
	popEfficiencyBonusStart = 10   // Efficiency level where bonus starts
)

// Factory/mine production point calculation.
const (
	productionBaseline = 10 // Baseline for production calculations

	// Factory output multipliers
	factoryOutputMultPositive = 100 // Multiplier when output > baseline
	factoryOutputMultNegative = 121 // Multiplier when output <= baseline

	// Factory cost multipliers
	factoryCostMultPositive = -60 // Squared multiplier when cost > baseline
	factoryCostMultNegative = -55 // Multiplier when cost <= baseline

	// Factory/mine count multipliers
	factoryCountMultPositive = 40 // Multiplier when count > baseline
	factoryCountMultNegative = 35 // Multiplier when count <= baseline

	// Low factory points limit and cap
	lowFactoryPointsLimit   = 700 // Limit for factory point capping
	factoryPointsCapDivisor = 3   // Divisor for capped factory points

	// Factory operation penalties
	factoryOpPenaltyThreshold1 = -7  // First threshold for operation penalty
	factoryOpPenaltyThreshold2 = -11 // Second threshold for operation penalty
	factoryOpPenaltyThreshold3 = -14 // Third threshold for operation penalty
	factoryOpPenaltyValue      = 360 // Penalty value at threshold 3
	factoryOpPenaltyMult1      = 45  // Multiplier between thresholds 2 and 3
	factoryOpPenaltyMult2      = 30  // Multiplier between thresholds 1 and 2
	factoryOpPenaltyOffset1    = 7   // Offset for penalty calc at threshold 2-3
	factoryOpPenaltyOffset2    = 6   // Offset for penalty calc at threshold 1-2

	// Factory production penalty
	factoryProdPenaltyThreshold = -3 // Threshold for production penalty
	factoryProdPenaltyMult      = 60 // Multiplier for production penalty
	factoryProdPenaltyOffset    = 2  // Offset for production penalty

	// Less germanium cost
	factoryLessGermaniumCost = 175 // Cost for factories use less germanium

	// AR race factory points
	arFactoryPoints = 210 // Fixed factory points for AR races

	// Mine-specific values
	mineOutputMultNegative = 169 // Multiplier for mine output <= baseline
	mineCostBaseline       = 3   // Baseline for mine cost calculation
	mineCostPenalty        = 360 // Penalty for mine cost > baseline
	mineCostMultNegative   = -65 // Multiplier for mine cost <= baseline
	mineCostOffset         = 80  // Offset for mine cost penalty
)

// Factory efficiency penalty (when factory output/count exceed baseline).
const (
	factoryEfficiencyThreshold = 10 // Threshold for efficiency penalty
	factoryEfficiencyOffset    = 9  // Offset for efficiency penalty calc
	factoryEfficiencyMinValue  = 1  // Minimum value after offset
	factoryProductionCostStd   = 2  // Production cost for non-HE races
	factoryProductionCostHE    = 3  // Production cost for HE races
	immunityPenaltyDivMulti    = 2  // Divisor for 2+ immunity penalty
	immunityPenaltyDivSingle   = 9  // Divisor for single/no immunity penalty
)

// LRT balance constants.
const (
	lrtMaxBeforePenalty   = 4  // Max LRTs before penalty applies
	lrtExcessPenaltyMult  = 10 // Multiplier for excess LRT penalty
	lrtImbalanceThreshold = 3  // Threshold for imbalance penalty
	lrtBadImbalanceMult   = 60 // Multiplier for too many bad LRTs
	lrtGoodImbalanceMult  = 40 // Multiplier for too many good LRTs
)

// NAS (No Advanced Scanners) penalties by PRT.
const (
	nasPenaltyPP   = 280 // NAS penalty for Packet Physics
	nasPenaltySS   = 200 // NAS penalty for Super Stealth
	nasPenaltyJoaT = 40  // NAS penalty for Jack of All Trades
)

// Research cost calculation.
const (
	researchCostSquaredMult     = 130  // Squared multiplier for "less" research costs
	researchCostAdj6Less        = 1430 // Adjustment when all 6 are "less"
	researchCostAdj5Less        = 520  // Adjustment when 5 are "less"
	researchCostLowPopThreshold = 1000 // Colonists threshold for extra penalty
	researchCostLowPopPenalty   = 190  // Penalty for low pop with 5+ extra research
)

// Science cost table for extra research costs (indexed by count - 1).
var scienceCostTable = [6]int{150, 330, 540, 780, 1050, 1380}

// Miscellaneous penalties.
const (
	techsStartHighPenalty = 180 // Penalty for techs start at level 3
	arCheapEnergyPenalty  = 100 // Penalty for AR with cheap energy
	pointsFinalDivisor    = 3   // Final divisor for calculated points
	habPointsFinalDivisor = 10.0
	habPointsRoundAdjust  = 0.5
)

// CalculatePoints calculates the advantage points for a race.
// This is a direct port of starsapi's RacePointsCalculator.java.
// Returns negative points if the race is invalid.
func CalculatePoints(r *Race) int {
	points := raceStartingPoints

	// 1. Habitability range points
	habPoints := getHabRangePoints(r) / habPointsDivisor

	// 2. Growth rate adjustment
	growthRateFactor := r.GrowthRate
	grRate := float64(r.GrowthRate)

	switch {
	case growthRateFactor <= 5:
		points += (6 - growthRateFactor) * growthRateBaseMultiplier
	case growthRateFactor <= 13:
		switch growthRateFactor {
		case 6:
			points += growthRateBonus6
		case 7:
			points += growthRateBonus7
		case 8:
			points += growthRateBonus8
		case 9:
			points += growthRateBonus9
		}
		growthRateFactor = growthRateFactor*2 - 5
	case growthRateFactor < 20:
		growthRateFactor = (growthRateFactor - 6) * 3
	default:
		growthRateFactor = growthRateMaxFactor
	}

	points -= (habPoints * growthRateFactor) / habGrowthPenaltyDivisor

	// 3. Off-center habitability bonus
	numImmunities := 0
	if r.GravityImmune {
		numImmunities++
	} else {
		points += abs(r.GravityCenter-habCenterIdeal) * habOffCenterBonusPerPoint
	}
	if r.TemperatureImmune {
		numImmunities++
	} else {
		points += abs(r.TemperatureCenter-habCenterIdeal) * habOffCenterBonusPerPoint
	}
	if r.RadiationImmune {
		numImmunities++
	} else {
		points += abs(r.RadiationCenter-habCenterIdeal) * habOffCenterBonusPerPoint
	}

	// 4. Multiple immunity penalty
	if numImmunities > 1 {
		points -= habMultipleImmunityPenalty
	}

	// 5. Factory efficiency penalty (depends on growth rate)
	operationPoints := r.FactoryCount
	productionPoints := r.FactoryOutput

	if operationPoints > factoryEfficiencyThreshold || productionPoints > factoryEfficiencyThreshold {
		operationPoints -= factoryEfficiencyOffset
		if operationPoints < factoryEfficiencyMinValue {
			operationPoints = factoryEfficiencyMinValue
		}
		productionPoints -= factoryEfficiencyOffset
		if productionPoints < factoryEfficiencyMinValue {
			productionPoints = factoryEfficiencyMinValue
		}

		// HE penalty: 3 for HE, 2 for others
		factoryProductionCost := factoryProductionCostStd
		if r.PRT == prtHE {
			factoryProductionCost = factoryProductionCostHE
		}

		productionPoints *= factoryProductionCost

		// Additional penalty for 2+ immunities
		if numImmunities >= 2 {
			points -= int(float64(productionPoints*operationPoints) * grRate / immunityPenaltyDivMulti)
		} else {
			points -= int(float64(productionPoints*operationPoints) * grRate / immunityPenaltyDivSingle)
		}
	}

	// 6. Population efficiency
	popEfficiency := r.ColonistsPerResource / popEfficiencyDivisor
	if popEfficiency > popEfficiencyMax {
		popEfficiency = popEfficiencyMax
	}

	switch {
	case popEfficiency <= 7:
		points -= popEfficiencyPenalty7
	case popEfficiency == 8:
		points -= popEfficiencyPenalty8
	case popEfficiency == 9:
		points -= popEfficiencyPenalty9
	case popEfficiency > popEfficiencyBonusStart:
		points += (popEfficiency - popEfficiencyBonusStart) * popEfficiencyBonusPer
	}

	// 7. Factory/Mine production points
	if r.PRT == prtAR {
		// AR races have very simple factory points
		points += arFactoryPoints
	} else {
		// Factory points
		productionPoints = productionBaseline - r.FactoryOutput
		costPoints := productionBaseline - r.FactoryCost
		operationPoints = productionBaseline - r.FactoryCount
		tmpPoints := 0

		if productionPoints > 0 {
			tmpPoints = productionPoints * factoryOutputMultPositive
		} else {
			tmpPoints = productionPoints * factoryOutputMultNegative
		}

		if costPoints > 0 {
			tmpPoints += costPoints * costPoints * factoryCostMultPositive
		} else {
			tmpPoints += costPoints * factoryCostMultNegative
		}

		if operationPoints > 0 {
			tmpPoints += operationPoints * factoryCountMultPositive
		} else {
			tmpPoints += operationPoints * factoryCountMultNegative
		}

		// Limit low factory points
		if tmpPoints > lowFactoryPointsLimit {
			tmpPoints = (tmpPoints-lowFactoryPointsLimit)/factoryPointsCapDivisor + lowFactoryPointsLimit
		}

		if operationPoints <= factoryOpPenaltyThreshold1 {
			if operationPoints < factoryOpPenaltyThreshold2 {
				if operationPoints < factoryOpPenaltyThreshold3 {
					tmpPoints -= factoryOpPenaltyValue
				} else {
					tmpPoints += (operationPoints + factoryOpPenaltyOffset1) * factoryOpPenaltyMult1
				}
			} else {
				tmpPoints += (operationPoints + factoryOpPenaltyOffset2) * factoryOpPenaltyMult2
			}
		}

		if productionPoints <= factoryProdPenaltyThreshold {
			tmpPoints += (productionPoints + factoryProdPenaltyOffset) * factoryProdPenaltyMult
		}

		points += tmpPoints

		if r.FactoriesUseLessGerm {
			points -= factoryLessGermaniumCost
		}

		// Mine points
		productionPoints = productionBaseline - r.MineOutput
		costPoints = mineCostBaseline - r.MineCost
		operationPoints = productionBaseline - r.MineCount

		if productionPoints > 0 {
			tmpPoints = productionPoints * factoryOutputMultPositive
		} else {
			tmpPoints = productionPoints * mineOutputMultNegative
		}

		if costPoints > 0 {
			tmpPoints -= mineCostPenalty
		} else {
			tmpPoints += costPoints*mineCostMultNegative + mineCostOffset
		}

		if operationPoints > 0 {
			tmpPoints += operationPoints * factoryCountMultPositive
		} else {
			tmpPoints += operationPoints * factoryCountMultNegative
		}

		points += tmpPoints
	}

	// 8. PRT points
	if prt := data.GetPRT(r.PRT); prt != nil {
		points += prt.PointCost
	}

	// 9. LRT points and balance penalties
	badLRTs := 0
	goodLRTs := 0

	for i := 0; i < lrtCount; i++ {
		if (r.LRT & (1 << i)) != 0 {
			if lrt := data.GetLRT(i); lrt != nil {
				points += lrt.PointCost
				if lrt.PointCost >= 0 {
					badLRTs++
				} else {
					goodLRTs++
				}
			}
		}
	}

	// Too many LRTs penalty
	totalLRTs := goodLRTs + badLRTs
	if totalLRTs > lrtMaxBeforePenalty {
		points -= totalLRTs * (totalLRTs - lrtMaxBeforePenalty) * lrtExcessPenaltyMult
	}

	// Imbalance penalty
	if badLRTs-goodLRTs > lrtImbalanceThreshold {
		points -= (badLRTs - goodLRTs - lrtImbalanceThreshold) * lrtBadImbalanceMult
	}
	if goodLRTs-badLRTs > lrtImbalanceThreshold {
		points -= (goodLRTs - badLRTs - lrtImbalanceThreshold) * lrtGoodImbalanceMult
	}

	// 10. NAS penalty by PRT
	if (r.LRT & (1 << lrtNAS)) != 0 {
		switch r.PRT {
		case prtPP:
			points -= nasPenaltyPP
		case prtSS:
			points -= nasPenaltySS
		case prtJoaT:
			points -= nasPenaltyJoaT
		}
	}

	// 11. Research cost points
	techCosts := 0
	researchCosts := []int{
		r.ResearchEnergy,
		r.ResearchWeapons,
		r.ResearchPropulsion,
		r.ResearchConstruction,
		r.ResearchElectronics,
		r.ResearchBiotech,
	}

	for _, rc := range researchCosts {
		switch rc {
		case ResearchCostExtra:
			techCosts--
		case ResearchCostLess:
			techCosts++
		}
	}

	if techCosts > 0 {
		// More "Less" than "Extra" - costs points
		points -= techCosts * techCosts * researchCostSquaredMult
		if techCosts >= 6 {
			points += researchCostAdj6Less // Already paid 4680 so true cost is 3250
		} else if techCosts == 5 {
			points += researchCostAdj5Less // Already paid 3250 so true cost is 2730
		}
	} else if techCosts < 0 {
		// More "Extra" than "Less" - gives points
		points += scienceCostTable[-techCosts-1]
		if techCosts < -4 && r.ColonistsPerResource < researchCostLowPopThreshold {
			points -= researchCostLowPopPenalty
		}
	}

	// 12. Techs start high penalty
	if r.TechsStartHigh {
		points -= techsStartHighPenalty
	}

	// 13. AR + cheap energy penalty
	if r.PRT == prtAR && r.ResearchEnergy == ResearchCostLess {
		points -= arCheapEnergyPenalty
	}

	return points / pointsFinalDivisor
}

// getHabRangePoints calculates habitability range advantage points.
// This uses a 3-loop simulation testing planets across the hab range.
func getHabRangePoints(r *Race) int {
	hasTT := (r.LRT & (1 << lrtTotalTerraforming)) != 0

	points := 0.0

	// Determine iteration counts (1 for immune, 11 for non-immune)
	numIterGrav := habNonImmuneIterations
	numIterTemp := habNonImmuneIterations
	numIterRad := habNonImmuneIterations
	if r.GravityImmune {
		numIterGrav = habImmuneIterations
	}
	if r.TemperatureImmune {
		numIterTemp = habImmuneIterations
	}
	if r.RadiationImmune {
		numIterRad = habImmuneIterations
	}

	// Three main loops with different terraforming correction factors
	for loopIndex := 0; loopIndex < 3; loopIndex++ {
		var ttCorrectionFactor int
		switch loopIndex {
		case 0:
			ttCorrectionFactor = 0
		case 1:
			if hasTT {
				ttCorrectionFactor = ttCorrectionLoop1WithTT
			} else {
				ttCorrectionFactor = ttCorrectionLoop1WithoutTT
			}
		case 2:
			if hasTT {
				ttCorrectionFactor = ttCorrectionLoop2WithTT
			} else {
				ttCorrectionFactor = ttCorrectionLoop2WithoutTT
			}
		}

		// Calculate test hab starts and widths for this loop
		testHabStart := [3]int{}
		testHabWidth := [3]int{}

		// Gravity
		if r.GravityImmune {
			testHabStart[0] = habImmuneTestCenter
			testHabWidth[0] = habImmuneTestWidth
		} else {
			testHabStart[0] = r.GravityLow() - ttCorrectionFactor
			if testHabStart[0] < 0 {
				testHabStart[0] = 0
			}
			tmpHab := r.GravityHigh() + ttCorrectionFactor
			if tmpHab > habMaxValue {
				tmpHab = habMaxValue
			}
			testHabWidth[0] = tmpHab - testHabStart[0]
		}

		// Temperature
		if r.TemperatureImmune {
			testHabStart[1] = habImmuneTestCenter
			testHabWidth[1] = habImmuneTestWidth
		} else {
			testHabStart[1] = r.TemperatureLow() - ttCorrectionFactor
			if testHabStart[1] < 0 {
				testHabStart[1] = 0
			}
			tmpHab := r.TemperatureHigh() + ttCorrectionFactor
			if tmpHab > habMaxValue {
				tmpHab = habMaxValue
			}
			testHabWidth[1] = tmpHab - testHabStart[1]
		}

		// Radiation
		if r.RadiationImmune {
			testHabStart[2] = habImmuneTestCenter
			testHabWidth[2] = habImmuneTestWidth
		} else {
			testHabStart[2] = r.RadiationLow() - ttCorrectionFactor
			if testHabStart[2] < 0 {
				testHabStart[2] = 0
			}
			tmpHab := r.RadiationHigh() + ttCorrectionFactor
			if tmpHab > habMaxValue {
				tmpHab = habMaxValue
			}
			testHabWidth[2] = tmpHab - testHabStart[2]
		}

		// Nested iteration over all hab dimensions
		gravitySum := 0.0
		for iterGrav := 0; iterGrav < numIterGrav; iterGrav++ {
			terraformOffset := [3]int{}
			testGrav := getPlanetHabForIndex(iterGrav, 0, loopIndex, numIterGrav,
				testHabStart[0], testHabWidth[0], r.GravityCenter, r.GravityImmune, ttCorrectionFactor, &terraformOffset)

			temperatureSum := 0.0
			for iterTemp := 0; iterTemp < numIterTemp; iterTemp++ {
				testTemp := getPlanetHabForIndex(iterTemp, 1, loopIndex, numIterTemp,
					testHabStart[1], testHabWidth[1], r.TemperatureCenter, r.TemperatureImmune, ttCorrectionFactor, &terraformOffset)

				radiationSum := int64(0)
				for iterRad := 0; iterRad < numIterRad; iterRad++ {
					testRad := getPlanetHabForIndex(iterRad, 2, loopIndex, numIterRad,
						testHabStart[2], testHabWidth[2], r.RadiationCenter, r.RadiationImmune, ttCorrectionFactor, &terraformOffset)

					// Calculate planet desirability
					planetDesirability := getPlanetHabitability(r, testGrav, testTemp, testRad)

					terraformOffsetSum := terraformOffset[0] + terraformOffset[1] + terraformOffset[2]
					if terraformOffsetSum > ttCorrectionFactor {
						planetDesirability -= int64(terraformOffsetSum - ttCorrectionFactor)
						if planetDesirability < 0 {
							planetDesirability = 0
						}
					}
					planetDesirability *= planetDesirability

					// Modify by loop index factor
					switch loopIndex {
					case 0:
						planetDesirability *= habLoopMultiplier0
					case 1:
						planetDesirability *= habLoopMultiplier1
					default:
						planetDesirability *= habLoopMultiplier2
					}

					radiationSum += planetDesirability
				}

				if !r.RadiationImmune {
					radiationSum = (radiationSum * int64(testHabWidth[2])) / habMaxValue
				} else {
					radiationSum *= habImmuneTestWidth
				}

				temperatureSum += float64(radiationSum)
			}

			if !r.TemperatureImmune {
				temperatureSum = (temperatureSum * float64(testHabWidth[1])) / habMaxValue
			} else {
				temperatureSum *= habImmuneTestWidth
			}

			gravitySum += temperatureSum
		}

		if !r.GravityImmune {
			gravitySum = (gravitySum * float64(testHabWidth[0])) / habMaxValue
		} else {
			gravitySum *= habImmuneTestWidth
		}

		points += gravitySum
	}

	return int(points/habPointsFinalDivisor + habPointsRoundAdjust)
}

// getPlanetHabForIndex calculates the hab value for a test planet at a specific iteration.
func getPlanetHabForIndex(iterIndex, habType, loopIndex, numIterations, testHabStart, testHabWidth, habCenter int,
	isImmune bool, ttCorrectionFactor int, terraformOffset *[3]int) int {
	var tmpHab int
	if iterIndex == 0 || numIterations <= 1 {
		tmpHab = testHabStart
	} else {
		tmpHab = (testHabWidth*iterIndex)/(numIterations-1) + testHabStart
	}

	if loopIndex != 0 && !isImmune {
		offset := habCenter - tmpHab
		switch {
		case abs(offset) <= ttCorrectionFactor:
			offset = 0
		case offset < 0:
			offset += ttCorrectionFactor
		default:
			offset -= ttCorrectionFactor
		}

		terraformOffset[habType] = offset
		tmpHab = habCenter - offset
	}

	return tmpHab
}

// getPlanetHabitability calculates the habitability of a planet for this race.
// Returns a value from 0-100, or negative if uninhabitable.
func getPlanetHabitability(r *Race, grav, temp, rad int) int64 {
	var planetValuePoints, redValue, ideality int64 = 0, 0, habPerfectValue

	habValues := [3]int{grav, temp, rad}
	habCenters := [3]int{r.GravityCenter, r.TemperatureCenter, r.RadiationCenter}
	habLows := [3]int{r.GravityLow(), r.TemperatureLow(), r.RadiationLow()}
	habHighs := [3]int{r.GravityHigh(), r.TemperatureHigh(), r.RadiationHigh()}
	isImmune := [3]bool{r.GravityImmune, r.TemperatureImmune, r.RadiationImmune}

	for habType := 0; habType < 3; habType++ {
		habValue := habValues[habType]
		habCenter := habCenters[habType]
		habLower := habLows[habType]
		habUpper := habHighs[habType]

		if isImmune[habType] {
			planetValuePoints += habPerfectValue
		} else {
			if habLower <= habValue && habUpper >= habValue {
				// Green planet
				fromIdeal := abs(habValue-habCenter) * habMaxValue
				var habRadius, tmp int
				if habCenter > habValue {
					habRadius = habCenter - habLower
					fromIdeal /= habRadius
					tmp = habCenter - habValue
				} else {
					habRadius = habUpper - habCenter
					fromIdeal /= habRadius
					tmp = habValue - habCenter
				}
				poorPlanetMod := tmp*2 - habRadius
				fromIdeal = habMaxValue - fromIdeal
				planetValuePoints += int64(fromIdeal * fromIdeal)
				if poorPlanetMod > 0 {
					ideality *= int64(habRadius*2 - poorPlanetMod)
					ideality /= int64(habRadius * 2)
				}
			} else {
				// Red planet
				var habRed int
				if habLower <= habValue {
					habRed = habValue - habUpper
				} else {
					habRed = habLower - habValue
				}

				if habRed > habRedMaxPenalty {
					habRed = habRedMaxPenalty
				}

				redValue += int64(habRed)
			}
		}
	}

	if redValue != 0 {
		return -redValue
	}

	planetValuePoints = int64(math.Sqrt(float64(planetValuePoints)/3) + 0.9)
	planetValuePoints = planetValuePoints * ideality / habPerfectValue

	return planetValuePoints
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
