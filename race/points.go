package race

import (
	"math"

	"github.com/neper-stars/houston/data"
)

// Starting advantage points before adjustments.
const raceStartingPoints = 1650

// LRT index for Total Terraforming (used in habitability calculation).
const lrtTotalTerraforming = 1

// LRT index for No Advanced Scanners.
const lrtNAS = 10

// PRT indices.
const (
	prtHE   = 0
	prtSS   = 1
	prtPP   = 6
	prtAR   = 8
	prtJoaT = 9
)

// CalculatePoints calculates the advantage points for a race.
// This is a direct port of starsapi's RacePointsCalculator.java.
// Returns negative points if the race is invalid.
func CalculatePoints(r *Race) int {
	points := raceStartingPoints

	// 1. Habitability range points
	habPoints := getHabRangePoints(r) / 2000

	// 2. Growth rate adjustment
	growthRateFactor := r.GrowthRate
	grRate := float64(r.GrowthRate)

	switch {
	case growthRateFactor <= 5:
		points += (6 - growthRateFactor) * 4200
	case growthRateFactor <= 13:
		switch growthRateFactor {
		case 6:
			points += 3600
		case 7:
			points += 2250
		case 8:
			points += 600
		case 9:
			points += 225
		}
		growthRateFactor = growthRateFactor*2 - 5
	case growthRateFactor < 20:
		growthRateFactor = (growthRateFactor - 6) * 3
	default:
		growthRateFactor = 45
	}

	points -= (habPoints * growthRateFactor) / 24

	// 3. Off-center habitability bonus
	numImmunities := 0
	if r.GravityImmune {
		numImmunities++
	} else {
		points += abs(r.GravityCenter-50) * 4
	}
	if r.TemperatureImmune {
		numImmunities++
	} else {
		points += abs(r.TemperatureCenter-50) * 4
	}
	if r.RadiationImmune {
		numImmunities++
	} else {
		points += abs(r.RadiationCenter-50) * 4
	}

	// 4. Multiple immunity penalty
	if numImmunities > 1 {
		points -= 150
	}

	// 5. Factory efficiency penalty (depends on growth rate)
	operationPoints := r.FactoryCount
	productionPoints := r.FactoryOutput

	if operationPoints > 10 || productionPoints > 10 {
		operationPoints -= 9
		if operationPoints < 1 {
			operationPoints = 1
		}
		productionPoints -= 9
		if productionPoints < 1 {
			productionPoints = 1
		}

		// HE penalty: 3 for HE, 2 for others
		factoryProductionCost := 2
		if r.PRT == prtHE {
			factoryProductionCost = 3
		}

		productionPoints *= factoryProductionCost

		// Additional penalty for 2+ immunities
		if numImmunities >= 2 {
			points -= int(float64(productionPoints*operationPoints) * grRate / 2)
		} else {
			points -= int(float64(productionPoints*operationPoints) * grRate / 9)
		}
	}

	// 6. Population efficiency
	popEfficiency := r.ColonistsPerResource / 100
	if popEfficiency > 25 {
		popEfficiency = 25
	}

	switch {
	case popEfficiency <= 7:
		points -= 2400
	case popEfficiency == 8:
		points -= 1260
	case popEfficiency == 9:
		points -= 600
	case popEfficiency > 10:
		points += (popEfficiency - 10) * 120
	}

	// 7. Factory/Mine production points
	if r.PRT == prtAR {
		// AR races have very simple factory points
		points += 210
	} else {
		// Factory points
		productionPoints = 10 - r.FactoryOutput
		costPoints := 10 - r.FactoryCost
		operationPoints = 10 - r.FactoryCount
		tmpPoints := 0

		if productionPoints > 0 {
			tmpPoints = productionPoints * 100
		} else {
			tmpPoints = productionPoints * 121
		}

		if costPoints > 0 {
			tmpPoints += costPoints * costPoints * -60
		} else {
			tmpPoints += costPoints * -55
		}

		if operationPoints > 0 {
			tmpPoints += operationPoints * 40
		} else {
			tmpPoints += operationPoints * 35
		}

		// Limit low factory points
		llfp := 700
		if tmpPoints > llfp {
			tmpPoints = (tmpPoints-llfp)/3 + llfp
		}

		if operationPoints <= -7 {
			if operationPoints < -11 {
				if operationPoints < -14 {
					tmpPoints -= 360
				} else {
					tmpPoints += (operationPoints + 7) * 45
				}
			} else {
				tmpPoints += (operationPoints + 6) * 30
			}
		}

		if productionPoints <= -3 {
			tmpPoints += (productionPoints + 2) * 60
		}

		points += tmpPoints

		if r.FactoriesUseLessGerm {
			points -= 175
		}

		// Mine points
		productionPoints = 10 - r.MineOutput
		costPoints = 3 - r.MineCost
		operationPoints = 10 - r.MineCount

		if productionPoints > 0 {
			tmpPoints = productionPoints * 100
		} else {
			tmpPoints = productionPoints * 169
		}

		if costPoints > 0 {
			tmpPoints -= 360
		} else {
			tmpPoints += costPoints*-65 + 80
		}

		if operationPoints > 0 {
			tmpPoints += operationPoints * 40
		} else {
			tmpPoints += operationPoints * 35
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

	for i := 0; i < 14; i++ {
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
	if totalLRTs > 4 {
		points -= totalLRTs * (totalLRTs - 4) * 10
	}

	// Imbalance penalty
	if badLRTs-goodLRTs > 3 {
		points -= (badLRTs - goodLRTs - 3) * 60
	}
	if goodLRTs-badLRTs > 3 {
		points -= (goodLRTs - badLRTs - 3) * 40
	}

	// 10. NAS penalty by PRT
	if (r.LRT & (1 << lrtNAS)) != 0 {
		switch r.PRT {
		case prtPP:
			points -= 280
		case prtSS:
			points -= 200
		case prtJoaT:
			points -= 40
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
		points -= techCosts * techCosts * 130
		if techCosts >= 6 {
			points += 1430 // Already paid 4680 so true cost is 3250
		} else if techCosts == 5 {
			points += 520 // Already paid 3250 so true cost is 2730
		}
	} else if techCosts < 0 {
		// More "Extra" than "Less" - gives points
		scienceCost := []int{150, 330, 540, 780, 1050, 1380}
		points += scienceCost[-techCosts-1]
		if techCosts < -4 && r.ColonistsPerResource < 1000 {
			points -= 190
		}
	}

	// 12. Techs start high penalty
	if r.TechsStartHigh {
		points -= 180
	}

	// 13. AR + cheap energy penalty
	if r.PRT == prtAR && r.ResearchEnergy == ResearchCostLess {
		points -= 100
	}

	return points / 3
}

// getHabRangePoints calculates habitability range advantage points.
// This uses a 3-loop simulation testing planets across the hab range.
func getHabRangePoints(r *Race) int {
	hasTT := (r.LRT & (1 << lrtTotalTerraforming)) != 0

	points := 0.0

	// Determine iteration counts (1 for immune, 11 for non-immune)
	numIterGrav := 11
	numIterTemp := 11
	numIterRad := 11
	if r.GravityImmune {
		numIterGrav = 1
	}
	if r.TemperatureImmune {
		numIterTemp = 1
	}
	if r.RadiationImmune {
		numIterRad = 1
	}

	// Three main loops with different terraforming correction factors
	for loopIndex := 0; loopIndex < 3; loopIndex++ {
		var ttCorrectionFactor int
		switch loopIndex {
		case 0:
			ttCorrectionFactor = 0
		case 1:
			if hasTT {
				ttCorrectionFactor = 8
			} else {
				ttCorrectionFactor = 5
			}
		case 2:
			if hasTT {
				ttCorrectionFactor = 17
			} else {
				ttCorrectionFactor = 15
			}
		}

		// Calculate test hab starts and widths for this loop
		testHabStart := [3]int{}
		testHabWidth := [3]int{}

		// Gravity
		if r.GravityImmune {
			testHabStart[0] = 50
			testHabWidth[0] = 11
		} else {
			testHabStart[0] = r.GravityLow() - ttCorrectionFactor
			if testHabStart[0] < 0 {
				testHabStart[0] = 0
			}
			tmpHab := r.GravityHigh() + ttCorrectionFactor
			if tmpHab > 100 {
				tmpHab = 100
			}
			testHabWidth[0] = tmpHab - testHabStart[0]
		}

		// Temperature
		if r.TemperatureImmune {
			testHabStart[1] = 50
			testHabWidth[1] = 11
		} else {
			testHabStart[1] = r.TemperatureLow() - ttCorrectionFactor
			if testHabStart[1] < 0 {
				testHabStart[1] = 0
			}
			tmpHab := r.TemperatureHigh() + ttCorrectionFactor
			if tmpHab > 100 {
				tmpHab = 100
			}
			testHabWidth[1] = tmpHab - testHabStart[1]
		}

		// Radiation
		if r.RadiationImmune {
			testHabStart[2] = 50
			testHabWidth[2] = 11
		} else {
			testHabStart[2] = r.RadiationLow() - ttCorrectionFactor
			if testHabStart[2] < 0 {
				testHabStart[2] = 0
			}
			tmpHab := r.RadiationHigh() + ttCorrectionFactor
			if tmpHab > 100 {
				tmpHab = 100
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
						planetDesirability *= 7
					case 1:
						planetDesirability *= 5
					default:
						planetDesirability *= 6
					}

					radiationSum += planetDesirability
				}

				if !r.RadiationImmune {
					radiationSum = (radiationSum * int64(testHabWidth[2])) / 100
				} else {
					radiationSum *= 11
				}

				temperatureSum += float64(radiationSum)
			}

			if !r.TemperatureImmune {
				temperatureSum = (temperatureSum * float64(testHabWidth[1])) / 100
			} else {
				temperatureSum *= 11
			}

			gravitySum += temperatureSum
		}

		if !r.GravityImmune {
			gravitySum = (gravitySum * float64(testHabWidth[0])) / 100
		} else {
			gravitySum *= 11
		}

		points += gravitySum
	}

	return int(points/10.0 + 0.5)
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
	var planetValuePoints, redValue, ideality int64 = 0, 0, 10000

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
			planetValuePoints += 10000
		} else {
			if habLower <= habValue && habUpper >= habValue {
				// Green planet
				fromIdeal := abs(habValue-habCenter) * 100
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
				fromIdeal = 100 - fromIdeal
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

				if habRed > 15 {
					habRed = 15
				}

				redValue += int64(habRed)
			}
		}
	}

	if redValue != 0 {
		return -redValue
	}

	planetValuePoints = int64(math.Sqrt(float64(planetValuePoints)/3) + 0.9)
	planetValuePoints = planetValuePoints * ideality / 10000

	return planetValuePoints
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
