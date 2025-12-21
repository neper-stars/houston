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
		} else {
			block.Decrypted = decryptor.DecryptBytes(block.Data)

			if block.Type == blocks.PlanetsBlockType {
				// if type = 7
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
			} else if block.Type == blocks.PlayerBlockType {
				// if type == 6
				playerBlock, err := blocks.NewPlayerBlock(*block)
				if err != nil {
					return nil, err
				}
				item = *playerBlock
			} else {
				// by default return the most basic kind of block
				item = *block
			}
		}

		blockList = append(blockList, item)
	}

	return blockList, nil
}
