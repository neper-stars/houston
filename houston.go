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
	AsciiString       = password.AsciiString
	HashRacePassword  = password.HashRacePassword
	GuessRacePassword = password.GuessRacePassword
)

// ReadRawFile reads an entire file into a FileData struct
func ReadRawFile(fName string, fileData *FileData) error {
	return parser.ReadRawFile(fName, (*parser.FileData)(fileData))
}
