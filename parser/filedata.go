package parser

import (
	"errors"
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/crypto"
	"github.com/neper-stars/houston/encoding"
)

var ErrNoFileHeaderFound = errors.New("no file header found")

type ErrMalformedBlock struct {
	Msg string
}

func (e ErrMalformedBlock) Error() string {
	return e.Msg
}

// FileData represents the raw bytes of a Stars! file
type FileData []byte

// FileHeader extracts the file header from the raw file data
func (fd FileData) FileHeader() (*blocks.FileHeader, error) {
	b, err := fd.ParseBlock(0)
	if err != nil {
		return nil, err
	}

	// type 8
	if b.Type == blocks.FileHeaderBlockType {
		// we have found ourselves the file header. Parse content
		return blocks.NewFileHeader(*b)
	}

	return nil, ErrNoFileHeaderFound
}

// ParseBlock parses a single block at the given offset
func (fd FileData) ParseBlock(offset int) (*blocks.GenericBlock, error) {
	blockHeader := encoding.Read16(fd, offset)
	// typeID is the first 6 bits of the header
	typeID := blocks.BlockTypeID(blockHeader >> 10)
	// size is the last 10 bits
	size := blocks.BlockSize(blockHeader & 0x3FF)
	// data is the file data without the 2 first bytes (16 bits) of header
	var data blocks.BlockData
	// if size == 0 (like in a FileFooterBlock, then we do not need to read data)
	if int(size) > 0 {
		// make sure we have enough data to read
		if len(fd) >= offset+2+int(size) {
			data = blocks.BlockData(fd[offset+2 : offset+2+int(size)])
		} else {
			// report some kind of error
			wholeDataLen := len(fd)
			lowerBound := offset + 2
			upperBound := offset + 2 + int(size)
			return nil, &ErrMalformedBlock{
				Msg: fmt.Sprintf(
					"malformed block, id: %d, size: %d, whole data len: %d, lowerBound: %d, upperBound: %d",
					typeID, size, wholeDataLen, lowerBound, upperBound,
				)}
		}
	}
	return &blocks.GenericBlock{
		Type: typeID,
		Size: size,
		Data: data,
	}, nil
}

// BlockList parses all blocks in the file data and returns them as a list
func (fd FileData) BlockList() ([]blocks.Block, error) {
	var blockList []blocks.Block
	decryptor := crypto.NewDecryptor()

	offset := 0
	for offset < len(fd) {
		block, err := fd.ParseBlock(offset)
		if err != nil {
			return nil, err
		}

		offset += int(block.Size) + 2
		var item blocks.Block

		// type 8
		if block.Type == blocks.FileHeaderBlockType {
			header, err := blocks.NewFileHeader(*block)
			if err != nil {
				return nil, err
			}
			var sw int
			if header.Shareware() {
				sw = 1
			}
			decryptor.InitDecryption(header.Salt(), int(header.GameID), int(header.Turn), header.PlayerIndex(), sw)
			item = *header
		} else if block.Type == blocks.FileFooterBlockType {
			// File footer is NOT encrypted
			block.Decrypted = blocks.DecryptedData(block.Data)
			item = *blocks.NewFileFooterBlock(*block)
		} else {
			block.Decrypted = decryptor.DecryptBytes(block.Data)

			switch block.Type {
			case blocks.PlanetsBlockType:
				// PlanetsBlock is an exception in that it has more data tacked onto the end
				planetBlock := blocks.NewPlanetsBlock(*block)

				// A bunch of planets data is tacked onto the end of this block
				// We need to determine how much and parse it
				// 4 bytes per planet
				length := planetBlock.GetPlanetCount() * 4
				planetBlock.ParsePlanetsData(fd[offset : offset+length])
				// Adjust our offset to after the planet data
				offset += length
				item = *planetBlock

			case blocks.PlayerBlockType:
				playerBlock, err := blocks.NewPlayerBlock(*block)
				if err != nil {
					return nil, err
				}
				item = *playerBlock

			case blocks.PlanetBlockType:
				// Full planet block (Type 13)
				item = *blocks.NewPlanetBlock(*block)

			case blocks.PartialPlanetBlockType:
				// Partial planet block (Type 14)
				item = *blocks.NewPartialPlanetBlock(*block)

			case blocks.FleetBlockType:
				// Full fleet block (Type 16)
				item = *blocks.NewFleetBlock(*block)

			case blocks.PartialFleetBlockType:
				// Partial fleet block (Type 17)
				item = *blocks.NewPartialFleetBlock(*block)

			case blocks.DesignBlockType:
				// Design block (Type 26)
				designBlock, err := blocks.NewDesignBlock(*block)
				if err != nil {
					return nil, err
				}
				item = *designBlock

			case blocks.WaypointDeleteBlockType:
				// Waypoint delete block (Type 3)
				item = *blocks.NewWaypointDeleteBlock(*block)

			case blocks.WaypointAddBlockType:
				// Waypoint add block (Type 4)
				item = *blocks.NewWaypointAddBlock(*block)

			case blocks.WaypointChangeTaskBlockType:
				// Waypoint change task block (Type 5)
				item = *blocks.NewWaypointChangeTaskBlock(*block)

			case blocks.WaypointTaskBlockType:
				// Waypoint task block (Type 19)
				item = *blocks.NewWaypointTaskBlock(*block)

			case blocks.WaypointBlockType:
				// Waypoint block (Type 20)
				item = *blocks.NewWaypointBlock(*block)

			case blocks.ProductionQueueBlockType:
				// Production queue block (Type 28)
				item = *blocks.NewProductionQueueBlock(*block)

			case blocks.BattlePlanBlockType:
				// Battle plan block (Type 30)
				item = *blocks.NewBattlePlanBlock(*block)

			case blocks.ObjectBlockType:
				// Object block (Type 43) - minefields, wormholes, etc.
				item = *blocks.NewObjectBlock(*block)

			case blocks.MessageBlockType:
				// Message block (Type 40)
				item = *blocks.NewMessageBlock(*block)

			case blocks.FleetSplitBlockType:
				// Fleet split block (Type 24)
				item = *blocks.NewFleetSplitBlock(*block)

			case blocks.FleetsMergeBlockType:
				// Fleets merge block (Type 37)
				item = *blocks.NewFleetsMergeBlock(*block)

			case blocks.DesignChangeBlockType:
				// Design change block (Type 27)
				designChangeBlock, err := blocks.NewDesignChangeBlock(*block)
				if err != nil {
					return nil, err
				}
				item = *designChangeBlock

			case blocks.ProductionQueueChangeBlockType:
				// Production queue change block (Type 29)
				item = *blocks.NewProductionQueueChangeBlock(*block)

			case blocks.CountersBlockType:
				// Counters block (Type 32)
				item = *blocks.NewCountersBlock(*block)

			// Fleet-related stub blocks
			case blocks.FleetNameBlockType:
				item = *blocks.NewFleetNameBlock(*block)

			case blocks.MoveShipsBlockType:
				item = *blocks.NewMoveShipsBlock(*block)

			case blocks.RenameFleetBlockType:
				item = *blocks.NewRenameFleetBlock(*block)

			// Battle-related stub blocks
			case blocks.BattleBlockType:
				item = *blocks.NewBattleBlock(*block)

			case blocks.BattleContinuationBlockType:
				item = *blocks.NewBattleContinuationBlock(*block)

			case blocks.SetFleetBattlePlanBlockType:
				item = *blocks.NewSetFleetBattlePlanBlock(*block)

			// Change blocks
			case blocks.ResearchChangeBlockType:
				item = *blocks.NewResearchChangeBlock(*block)

			case blocks.PlanetChangeBlockType:
				item = *blocks.NewPlanetChangeBlock(*block)

			case blocks.ChangePasswordBlockType:
				item = *blocks.NewChangePasswordBlock(*block)

			case blocks.PlayersRelationChangeBlockType:
				item = *blocks.NewPlayersRelationChangeBlock(*block)

			// Misc stub blocks
			case blocks.PlayerScoresBlockType:
				item = *blocks.NewPlayerScoresBlock(*block)

			case blocks.SaveAndSubmitBlockType:
				item = *blocks.NewSaveAndSubmitBlock(*block)

			case blocks.FileHashBlockType:
				item = *blocks.NewFileHashBlock(*block)

			case blocks.WaypointRepeatOrdersBlockType:
				item = *blocks.NewWaypointRepeatOrdersBlock(*block)

			case blocks.EventsBlockType:
				item = *blocks.NewEventsBlock(*block)

			case blocks.MessagesFilterBlockType:
				item = *blocks.NewMessagesFilterBlock(*block)

			case blocks.AiHFileRecordBlockType:
				item = *blocks.NewAiHFileRecordBlock(*block)

			case blocks.ManualSmallLoadUnloadTaskBlockType:
				item = *blocks.NewManualSmallLoadUnloadTaskBlock(*block)

			case blocks.ManualMediumLoadUnloadTaskBlockType:
				item = *blocks.NewManualMediumLoadUnloadTaskBlock(*block)

			case blocks.ManualLargeLoadUnloadTaskBlockType:
				item = *blocks.NewManualLargeLoadUnloadTaskBlock(*block)

			default:
				// by default return the most basic kind of block
				item = *block
			}
		}

		blockList = append(blockList, item)
	}

	return blockList, nil
}
