// Package data contains static game data and constants for Stars! file parsing.
package data

// GameSettings represents the bit flags found in the PlanetsBlock (Type 7)
// that describe the game configuration options selected when the game was created.
//
// These settings are stored as a 16-bit bitmask in the game files.
// Each bit corresponds to a specific game option that was enabled or disabled
// by the host when creating the game.
//
// Example usage:
//
//	if settings & GameSettingMaxMinerals != 0 {
//	    // Max minerals is enabled
//	}
const (
	// GameSettingMaxMinerals indicates that planets start with maximum mineral
	// concentrations. When enabled, all planets begin with higher mineral
	// availability, making early game expansion easier.
	GameSettingMaxMinerals = 1 << 0 // Bit 0

	// GameSettingSlowTech indicates that technology research advances more slowly.
	// When enabled, the cost to research each technology level is increased,
	// resulting in a longer game with more focus on early-game strategies.
	GameSettingSlowTech = 1 << 1 // Bit 1

	// GameSettingSinglePlayer indicates the game is configured for single player
	// mode against AI opponents only. This affects certain game mechanics and
	// AI behavior.
	GameSettingSinglePlayer = 1 << 2 // Bit 2

	// GameSettingBit3 is reserved/unknown. Its purpose has not been documented.
	GameSettingBit3 = 1 << 3 // Bit 3

	// GameSettingComputerAlliances allows AI players to form alliances with
	// each other. When enabled, computer players may coordinate their actions
	// against human players.
	GameSettingComputerAlliances = 1 << 4 // Bit 4

	// GameSettingPublicScores makes all player scores visible to everyone.
	// When enabled, players can see each other's scores throughout the game
	// rather than only at the end or through espionage.
	GameSettingPublicScores = 1 << 5 // Bit 5

	// GameSettingAcceleratedBBS enables accelerated play mode designed for
	// BBS (Bulletin Board System) games. This mode has faster turn processing
	// and was originally designed for play-by-email or BBS-style games.
	GameSettingAcceleratedBBS = 1 << 6 // Bit 6

	// GameSettingNoRandomEvents disables random game events such as comet
	// strikes, mystery traders, and other random occurrences. When enabled,
	// the game becomes more predictable and strategic.
	GameSettingNoRandomEvents = 1 << 7 // Bit 7

	// GameSettingGalaxyClumping affects how stars are distributed in the
	// galaxy. When enabled, stars tend to cluster together rather than
	// being evenly distributed, creating more strategic chokepoints.
	GameSettingGalaxyClumping = 1 << 8 // Bit 8
)

// GameSettingNames maps game setting flags to human-readable names.
// Useful for debugging and displaying game configuration.
var GameSettingNames = map[int]string{
	GameSettingMaxMinerals:       "Max Minerals",
	GameSettingSlowTech:          "Slow Tech Advances",
	GameSettingSinglePlayer:      "Single Player",
	GameSettingBit3:              "Unknown (Bit 3)",
	GameSettingComputerAlliances: "Computer Alliances",
	GameSettingPublicScores:      "Public Scores",
	GameSettingAcceleratedBBS:    "Accelerated BBS Play",
	GameSettingNoRandomEvents:    "No Random Events",
	GameSettingGalaxyClumping:    "Galaxy Clumping",
}

// UniverseSize represents the size of the game universe.
// The universe size affects the number of planets and the distances between them.
type UniverseSize int

const (
	// UniverseSizeTiny is the smallest universe option (about 200x200 ly).
	UniverseSizeTiny UniverseSize = iota

	// UniverseSizeSmall is a small universe (about 400x400 ly).
	UniverseSizeSmall

	// UniverseSizeMedium is the default medium universe (about 600x600 ly).
	UniverseSizeMedium

	// UniverseSizeLarge is a large universe (about 800x800 ly).
	UniverseSizeLarge

	// UniverseSizeHuge is the largest universe option (about 1000x1000 ly).
	UniverseSizeHuge
)

// UniverseSizeNames maps universe size values to human-readable names.
var UniverseSizeNames = map[UniverseSize]string{
	UniverseSizeTiny:   "Tiny",
	UniverseSizeSmall:  "Small",
	UniverseSizeMedium: "Medium",
	UniverseSizeLarge:  "Large",
	UniverseSizeHuge:   "Huge",
}

// UniverseDensity represents the density of planets in the universe.
type UniverseDensity int

const (
	// UniverseDensitySparse has fewer planets spread further apart.
	UniverseDensitySparse UniverseDensity = iota

	// UniverseDensityNormal is the default planet density.
	UniverseDensityNormal

	// UniverseDensityDense has more planets packed closer together.
	UniverseDensityDense

	// UniverseDensityPacked has maximum planet density.
	UniverseDensityPacked
)

// UniverseDensityNames maps density values to human-readable names.
var UniverseDensityNames = map[UniverseDensity]string{
	UniverseDensitySparse: "Sparse",
	UniverseDensityNormal: "Normal",
	UniverseDensityDense:  "Dense",
	UniverseDensityPacked: "Packed",
}

// PlayerRelation represents the diplomatic relationship between two players.
type PlayerRelation int

const (
	// RelationNeutral is the default relationship - neither friend nor enemy.
	RelationNeutral PlayerRelation = iota

	// RelationFriend indicates an allied relationship.
	RelationFriend

	// RelationEnemy indicates a hostile relationship.
	RelationEnemy
)

// PlayerRelationNames maps relation values to human-readable names.
var PlayerRelationNames = map[PlayerRelation]string{
	RelationNeutral: "Neutral",
	RelationFriend:  "Friend",
	RelationEnemy:   "Enemy",
}

// VictoryCondition represents the different ways a game can be won.
type VictoryCondition int

const (
	// VictoryOwnsPercentPlanets - player owns a percentage of all planets.
	VictoryOwnsPercentPlanets VictoryCondition = 1 << 0

	// VictoryAttainsPercentTechLevels - player reaches a percentage of max tech.
	VictoryAttainsPercentTechLevels VictoryCondition = 1 << 1

	// VictoryExceedScoreCondition - player's score exceeds a threshold.
	VictoryExceedScoreCondition VictoryCondition = 1 << 2

	// VictoryExceedSecondPlace - player's score exceeds second place by percentage.
	VictoryExceedSecondPlace VictoryCondition = 1 << 3

	// VictoryProductionCapacity - player achieves production capacity threshold.
	VictoryProductionCapacity VictoryCondition = 1 << 4

	// VictoryOwnCapitalShips - player owns threshold of capital ships.
	VictoryOwnCapitalShips VictoryCondition = 1 << 5

	// VictoryHighestScoreYears - player has highest score after N years.
	VictoryHighestScoreYears VictoryCondition = 1 << 6

	// VictoryNumCriteriaMet - player meets N of the above criteria.
	VictoryNumCriteriaMet VictoryCondition = 1 << 7
)

// VictoryConditionNames maps victory conditions to human-readable descriptions.
var VictoryConditionNames = map[VictoryCondition]string{
	VictoryOwnsPercentPlanets:       "Owns % of Planets",
	VictoryAttainsPercentTechLevels: "Attains % of Tech Levels",
	VictoryExceedScoreCondition:     "Exceeds Score",
	VictoryExceedSecondPlace:        "Exceeds Second Place by %",
	VictoryProductionCapacity:       "Production Capacity",
	VictoryOwnCapitalShips:          "Owns Capital Ships",
	VictoryHighestScoreYears:        "Highest Score After Years",
	VictoryNumCriteriaMet:           "Meets N Criteria",
}

// VictoryConditionIndex represents the array indices for victory condition thresholds.
// The VictoryConditions array in PlanetsBlock is 12 bytes (indices 0-11).
//
// Each byte encodes: bit 7 (0x80) = enabled flag, bits 0-6 (0x7F) = threshold index.
//
// Source: Decompiled from stars26jrc3.exe GetVCVal() at 1078:b710
const (
	VictoryIdxOwnsPercentPlanets = 0  // Formula: idx*5+20, range 20-100%
	VictoryIdxAttainsTechLevel   = 1  // Formula: idx+8, range 8-26 (attain tech level X)
	VictoryIdxTechInYFields      = 2  // Formula: idx+2, range 2-6 (in Y tech fields, 2nd part of tech condition)
	VictoryIdxExceedScore        = 3  // Formula: idx*1000+1000, range 1k-20k
	VictoryIdxExceedSecondPlace  = 4  // Formula: idx*10+20, range 20-300%
	VictoryIdxProductionCapacity = 5  // Formula: idx*10+10, range 10-500 (thousands)
	VictoryIdxOwnCapitalShips    = 6  // Formula: idx*10+10, range 10-300
	VictoryIdxHighestScoreYears  = 7  // Formula: idx*10+30, range 30-900 (highest score after N years)
	VictoryIdxNumCriteriaMet     = 8  // Special: counts enabled conditions from 0-7 (excl. 2), range 1-7
	VictoryIdxMinYearsBeforeWin  = 9  // Formula: idx*10+30, range 30-500 (min years before winner declared)
	VictoryIdxReserved10         = 10 // Reserved
	VictoryIdxReserved11         = 11 // Reserved
)

// VictoryConditionThresholds defines the maximum threshold index for each
// victory condition. These values determine the valid range of settings
// for each victory condition in the game setup.
//
// Source: Decompiled from stars26jrc3.exe vrgvcMax[] array at 1078:b5a8
type VictoryConditionThresholds struct {
	OwnsPercentPlanets int // Max idx 16: formula idx*5+20, values 20-100%
	AttainsTechLevel   int // Max idx 18: formula idx+8, values 8-26 (attain tech level X)
	TechInYFields      int // Max idx 4: formula idx+2, values 2-6 (in Y tech fields)
	ExceedScore        int // Max idx 19: formula idx*1000+1000, values 1k-20k
	ExceedSecondPlace  int // Max idx 28: formula idx*10+20, values 20-300%
	ProductionCapacity int // Max idx 49: formula idx*10+10, values 10-500 (thousands)
	OwnCapitalShips    int // Max idx 29: formula idx*10+10, values 10-300
	HighestScoreYears  int // Max idx 87: formula idx*10+30, values 30-900
	NumCriteriaMet     int // Special: counts enabled conditions, values 1-7
	MinYearsBeforeWin  int // Max idx 47: formula idx*10+30, values 30-500
	Reserved10         int // Reserved
	Reserved11         int // Reserved
}

// ToArray converts the thresholds to the file format array (12 bytes).
func (v VictoryConditionThresholds) ToArray() [12]int {
	return [12]int{
		v.OwnsPercentPlanets,
		v.AttainsTechLevel,
		v.TechInYFields,
		v.ExceedScore,
		v.ExceedSecondPlace,
		v.ProductionCapacity,
		v.OwnCapitalShips,
		v.HighestScoreYears,
		v.NumCriteriaMet,
		v.MinYearsBeforeWin,
		v.Reserved10,
		v.Reserved11,
	}
}

// DefaultVictoryThresholds contains the standard max index values from Stars!
var DefaultVictoryThresholds = VictoryConditionThresholds{
	OwnsPercentPlanets: 16, // idx*5+20 → 20-100%
	AttainsTechLevel:   18, // idx+8 → 8-26 (attain tech level X)
	TechInYFields:      4,  // idx+2 → 2-6 (in Y tech fields)
	ExceedScore:        19, // idx*1000+1000 → 1k-20k
	ExceedSecondPlace:  28, // idx*10+20 → 20-300%
	ProductionCapacity: 49, // idx*10+10 → 10-500 (thousands)
	OwnCapitalShips:    29, // idx*10+10 → 10-300
	HighestScoreYears:  87, // idx*10+30 → 30-900 (highest score after N years)
	NumCriteriaMet:     7,  // counts enabled (max 7 conditions)
	MinYearsBeforeWin:  47, // idx*10+30 → 30-500 (min years before winner declared)
	Reserved10:         0,
	Reserved11:         0,
}

// VictoryCondition byte format constants.
const (
	VictoryConditionEnabledBit = 0x80 // Bit 7: condition is enabled
	VictoryConditionIndexMask  = 0x7F // Bits 0-6: threshold index value
)

// GetVictoryValue converts a threshold index to the actual game value.
// Returns the computed value and true if valid, or 0 and false if invalid.
//
// Source: Decompiled from stars26jrc3.exe GetVCVal() at 1078:b710
func GetVictoryValue(conditionIndex, thresholdIndex int) (int, bool) {
	switch conditionIndex {
	case VictoryIdxOwnsPercentPlanets:
		return thresholdIndex*5 + 20, true // 20-100%
	case VictoryIdxAttainsTechLevel:
		return thresholdIndex + 8, true // 8-26 (attain tech level X)
	case VictoryIdxTechInYFields:
		return thresholdIndex + 2, true // 2-6 (in Y tech fields)
	case VictoryIdxExceedScore:
		return thresholdIndex*1000 + 1000, true // 1k-20k
	case VictoryIdxExceedSecondPlace:
		return thresholdIndex*10 + 20, true // 20-300%
	case VictoryIdxProductionCapacity:
		return thresholdIndex*10 + 10, true // 10-500 (thousands)
	case VictoryIdxOwnCapitalShips:
		return thresholdIndex*10 + 10, true // 10-300
	case VictoryIdxHighestScoreYears:
		return thresholdIndex*10 + 30, true // 30-900 (highest score after N years)
	case VictoryIdxNumCriteriaMet:
		// Special case: value is 1-7 (count of enabled conditions)
		return thresholdIndex, true
	case VictoryIdxMinYearsBeforeWin:
		return thresholdIndex*10 + 30, true // 30-500 (min years before winner)
	default:
		return 0, false
	}
}
