// Package houston provides tools for reading and parsing Stars! game files.
// This file serves as a facade for backwards compatibility, re-exporting
// types and functions from the sub-packages.
package houston

import (
	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/data"
	"github.com/neper-stars/houston/parser"
	"github.com/neper-stars/houston/password"
)

// Block type aliases for backwards compatibility
type (
	Block         = blocks.Block
	BlockTypeID   = blocks.BlockTypeID
	BlockSize     = blocks.BlockSize
	BlockData     = blocks.BlockData
	DecryptedData = blocks.DecryptedData
	GenericBlock  = blocks.GenericBlock
	FileHeader    = blocks.FileHeader
	PlayerBlock   = blocks.PlayerBlock
	PlanetsBlock  = blocks.PlanetsBlock
	Planet        = blocks.Planet
	HashedPass    = blocks.HashedPass

	// New block types
	FileFooterBlock    = blocks.FileFooterBlock
	PartialPlanetBlock = blocks.PartialPlanetBlock
	PlanetBlock        = blocks.PlanetBlock
	PartialFleetBlock  = blocks.PartialFleetBlock
	FleetBlock         = blocks.FleetBlock
	DesignBlock        = blocks.DesignBlock
	DesignSlot         = blocks.DesignSlot

	// Waypoint blocks
	WaypointBlock           = blocks.WaypointBlock
	WaypointTaskBlock       = blocks.WaypointTaskBlock
	WaypointChangeTaskBlock = blocks.WaypointChangeTaskBlock
	WaypointAddBlock        = blocks.WaypointAddBlock
	WaypointDeleteBlock     = blocks.WaypointDeleteBlock

	// Production & battle blocks
	ProductionQueueBlock = blocks.ProductionQueueBlock
	QueueItem            = blocks.QueueItem
	BattlePlanBlock      = blocks.BattlePlanBlock

	// Object & message blocks
	ObjectBlock  = blocks.ObjectBlock
	MessageBlock = blocks.MessageBlock

	// Fleet operation blocks
	FleetSplitBlock  = blocks.FleetSplitBlock
	FleetsMergeBlock = blocks.FleetsMergeBlock

	// Change blocks
	DesignChangeBlock          = blocks.DesignChangeBlock
	ProductionQueueChangeBlock = blocks.ProductionQueueChangeBlock

	// Info blocks
	CountersBlock = blocks.CountersBlock

	// Fleet stub blocks
	FleetNameBlock  = blocks.FleetNameBlock
	MoveShipsBlock  = blocks.MoveShipsBlock
	RenameFleetBlock = blocks.RenameFleetBlock

	// Battle stub blocks
	BattleBlock             = blocks.BattleBlock
	BattleContinuationBlock = blocks.BattleContinuationBlock
	SetFleetBattlePlanBlock = blocks.SetFleetBattlePlanBlock

	// Change stub blocks
	ResearchChangeBlock        = blocks.ResearchChangeBlock
	PlanetChangeBlock          = blocks.PlanetChangeBlock
	ChangePasswordBlock        = blocks.ChangePasswordBlock
	PlayersRelationChangeBlock = blocks.PlayersRelationChangeBlock

	// Misc stub blocks
	PlayerScoresBlock            = blocks.PlayerScoresBlock
	SaveAndSubmitBlock           = blocks.SaveAndSubmitBlock
	FileHashBlock                = blocks.FileHashBlock
	WaypointRepeatOrdersBlock    = blocks.WaypointRepeatOrdersBlock
	EventsBlock                  = blocks.EventsBlock
	MessagesFilterBlock          = blocks.MessagesFilterBlock
	AiHFileRecordBlock           = blocks.AiHFileRecordBlock
	ManualSmallLoadUnloadTaskBlock  = blocks.ManualSmallLoadUnloadTaskBlock
	ManualMediumLoadUnloadTaskBlock = blocks.ManualMediumLoadUnloadTaskBlock
	ManualLargeLoadUnloadTaskBlock  = blocks.ManualLargeLoadUnloadTaskBlock
)

// Parser type aliases
type FileData = parser.FileData

// Re-exported block type constants
const (
	FileFooterBlockType                 = blocks.FileFooterBlockType
	ManualSmallLoadUnloadTaskBlockType  = blocks.ManualSmallLoadUnloadTaskBlockType
	ManualMediumLoadUnloadTaskBlockType = blocks.ManualMediumLoadUnloadTaskBlockType
	WaypointDeleteBlockType             = blocks.WaypointDeleteBlockType
	WaypointAddBlockType                = blocks.WaypointAddBlockType
	WaypointChangeTaskBlockType         = blocks.WaypointChangeTaskBlockType
	PlayerBlockType                     = blocks.PlayerBlockType
	PlanetsBlockType                    = blocks.PlanetsBlockType
	FileHeaderBlockType                 = blocks.FileHeaderBlockType
	FileHashBlockType                   = blocks.FileHashBlockType
	WaypointRepeatOrdersBlockType       = blocks.WaypointRepeatOrdersBlockType
	UnknownBlock11BlockType             = blocks.UnknownBlock11BlockType
	EventsBlockType                     = blocks.EventsBlockType
	PlanetBlockType                     = blocks.PlanetBlockType
	PartialPlanetBlockType              = blocks.PartialPlanetBlockType
	UnknownBlock15BlockType             = blocks.UnknownBlock15BlockType
	FleetBlockType                      = blocks.FleetBlockType
	PartialFleetBlockType               = blocks.PartialFleetBlockType
	UnknownBlock18BlockType             = blocks.UnknownBlock18BlockType
	WaypointTaskBlockType               = blocks.WaypointTaskBlockType
	WaypointBlockType                   = blocks.WaypointBlockType
	FleetNameBlockType                  = blocks.FleetNameBlockType
	UnknownBlock22BlockType             = blocks.UnknownBlock22BlockType
	MoveShipsBlockType                  = blocks.MoveShipsBlockType
	FleetSplitBlockType                 = blocks.FleetSplitBlockType
	ManualLargeLoadUnloadTaskBlockType  = blocks.ManualLargeLoadUnloadTaskBlockType
	DesignBlockType                     = blocks.DesignBlockType
	DesignChangeBlockType               = blocks.DesignChangeBlockType
	ProductionQueueBlockType            = blocks.ProductionQueueBlockType
	ProductionQueueChangeBlockType      = blocks.ProductionQueueChangeBlockType
	BattlePlanBlockType                 = blocks.BattlePlanBlockType
	BattleBlockType                     = blocks.BattleBlockType
	CountersBlockType                   = blocks.CountersBlockType
	MessagesFilterBlockType             = blocks.MessagesFilterBlockType
	ResearchChangeBlockType             = blocks.ResearchChangeBlockType
	PlanetChangeBlockType               = blocks.PlanetChangeBlockType
	ChangePasswordBlockType             = blocks.ChangePasswordBlockType
	FleetsMergeBlockType                = blocks.FleetsMergeBlockType
	PlayersRelationChangeBlockType      = blocks.PlayersRelationChangeBlockType
	BattleContinuationBlockType         = blocks.BattleContinuationBlockType
	MessageBlockType                    = blocks.MessageBlockType
	AiHFileRecordBlockType              = blocks.AiHFileRecordBlockType
	SetFleetBattlePlanBlockType         = blocks.SetFleetBattlePlanBlockType
	ObjectBlockType                     = blocks.ObjectBlockType
	RenameFleetBlockType                = blocks.RenameFleetBlockType
	PlayerScoresBlockType               = blocks.PlayerScoresBlockType
	SaveAndSubmitBlockType              = blocks.SaveAndSubmitBlockType
)

// Fleet kind constants
const (
	FleetKindPartial    = blocks.FleetKindPartial
	FleetKindPickPocket = blocks.FleetKindPickPocket
	FleetKindFull       = blocks.FleetKindFull
)

// Game settings flags (used in PlanetsBlock.GameSettings bitmask)
const (
	GameSettingMaxMinerals       = data.GameSettingMaxMinerals
	GameSettingSlowTech          = data.GameSettingSlowTech
	GameSettingSinglePlayer      = data.GameSettingSinglePlayer
	GameSettingComputerAlliances = data.GameSettingComputerAlliances
	GameSettingPublicScores      = data.GameSettingPublicScores
	GameSettingAcceleratedBBS    = data.GameSettingAcceleratedBBS
	GameSettingNoRandomEvents    = data.GameSettingNoRandomEvents
	GameSettingGalaxyClumping    = data.GameSettingGalaxyClumping
)

// Universe size constants
type UniverseSize = data.UniverseSize

const (
	UniverseSizeTiny   = data.UniverseSizeTiny
	UniverseSizeSmall  = data.UniverseSizeSmall
	UniverseSizeMedium = data.UniverseSizeMedium
	UniverseSizeLarge  = data.UniverseSizeLarge
	UniverseSizeHuge   = data.UniverseSizeHuge
)

// Universe density constants
type UniverseDensity = data.UniverseDensity

const (
	UniverseDensitySparse = data.UniverseDensitySparse
	UniverseDensityNormal = data.UniverseDensityNormal
	UniverseDensityDense  = data.UniverseDensityDense
	UniverseDensityPacked = data.UniverseDensityPacked
)

// Re-exported error variables
var (
	ErrNoFileHeaderFound      = parser.ErrNoFileHeaderFound
	ErrInvalidFileHeaderBlock = blocks.ErrInvalidFileHeaderBlock
	ErrInvalidPlayerBlock     = blocks.ErrInvalidPlayerBlock
)

// Re-exported data
var PlanetNames = data.PlanetNames

// Re-exported password functions
var (
	AsciiString                = password.AsciiString
	HashRacePassword           = password.HashRacePassword
	HashRacePasswordBytes      = password.HashRacePasswordBytes
	GuessRacePassword          = password.GuessRacePassword
	GuessRacePasswordParallel  = password.GuessRacePasswordParallel
)

// ProgressCallback is called periodically during parallel password search
type ProgressCallback = password.ProgressCallback

// ReadRawFile reads an entire file into a FileData struct
func ReadRawFile(fName string, fileData *FileData) error {
	return parser.ReadRawFile(fName, (*parser.FileData)(fileData))
}
