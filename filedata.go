package houston

import (
	"errors"
	"fmt"
)

var ErrNoFileHeaderFound = errors.New("no file header found")

type ErrMalformedBlock struct {
	Msg string
}

func (e ErrMalformedBlock) Error() string {
	return e.Msg
}

type FileData []byte

func (fd FileData) FileHeader() (*FileHeader, error) {
	b, err := fd.ParseBlock(0)
	if err != nil {
		return nil, err
	}

	// type 8
	if b.Type == FileHeaderBlockType {
		// we have found ourselves the file header. Parse content
		return NewFileHeader(*b)
	}

	return nil, ErrNoFileHeaderFound
}

func (fd FileData) ParseBlock(offset int) (*GenericBlock, error) {
	// fmt.Println("fd dump:\n", hex.EncodeToString(fd))
	/*
		blockHeader = util.read16(fileBytes, offset)
		typeId = blockHeader >> 10   # First 6 bits
		size = blockHeader & 0x3FF # Last 10 bits
		data = fileBytes[offset+2:offset+2+size]
	*/
	blockHeader := read16(fd, offset)
	// fmt.Printf("header @offset %d: %x\n", offset, blockHeader)
	// typeID is the first 6 bits of the header
	typeID := BlockTypeID(blockHeader >> 10)
	// size is the last 10 bits
	size := BlockSize(blockHeader & 0x3FF)
	// fmt.Printf("typeID: %d, size: %d, blockHeader: %d\n", typeID, size, blockHeader)
	// data is the file data without the 2 first bytes (16 bits) of header
	var data BlockData
	// if size == 0 (like in a FileFooterBlock, then we do not need to read data)
	if int(size) > 0 {
		// make sure we have enough data to read
		if len(fd) >= offset+2+int(size) {
			data = BlockData(fd[offset+2 : offset+2+int(size)])
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
	return &GenericBlock{
		Type: typeID,
		Size: size,
		Data: data,
	}, nil
}

func (fd FileData) BlockList() ([]Block, error) {
	var blockList []Block
	decryptor := NewDecryptor()

	offset := 0
	for offset < len(fd) {
		// fmt.Println("will parse block @", offset)
		block, err := fd.ParseBlock(offset)
		if err != nil {
			return nil, err
		}

		// fmt.Println("BLOCK:", block.BlockTypeID(), block.BlockSize(), block.BlockData())

		offset += int(block.Size) + 2
		var item Block

		// type 8
		if block.Type == FileHeaderBlockType {
			header, err := NewFileHeader(*block)
			if err != nil {
				return nil, err
			}
			var sw int
			if header.Shareware() {
				sw = 1
			}
			decryptor.InitDecryption(header.Salt(), int(header.GameID), int(header.Turn), header.PlayerIndex(), sw)
			item = *header
			// fmt.Printf("%+v\n", item)
		} else {
			block.Decrypted = decryptor.DecryptBytes(block.Data)

			if block.Type == PlanetsBlockType {
				// if type = 7
				// PlanetsBlock is an exception in that it has more data tacked onto the end
				planetBlock := NewPlanetsBlock(*block)

				// A bunch of planets data is tacked onto the end of this block
				// We need to determine how much and parse it
				// 4 bytes per planet
				length := planetBlock.GetPlanetCount() * 4
				planetBlock.ParsePlanetsData(fd[offset : offset+length])
				// Adjust our offset to after the planet data
				offset += length
				item = *planetBlock
			} else if block.Type == PlayerBlockType {
				// if type == 6
				playerBlock, err := NewPlayerBlock(*block)
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
