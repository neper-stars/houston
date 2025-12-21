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
