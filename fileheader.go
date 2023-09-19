package houston

import (
	"fmt"
	"errors"
)

var ErrInvalidFileHeaderBlock = errors.New("invalid file header")

// FileHeader is a specialized block that implements the
// Block interface and a more specialized one dedicated to
// data you only find in a file header
type FileHeader struct {
	GenericBlock
	magic       [4]byte
	GameID      uint32
	VersionData uint16
	Turn        uint16
	PlayerData  uint16
	Flags       uint8
}

// NewFileHeader is constructor that takes a GenericBlock and returns
// a pointer to a FileHeader
func NewFileHeader(b GenericBlock) (*FileHeader, error) {
	data := b.BlockData()
	if len(data) < 15 {
		return nil, ErrInvalidFileHeaderBlock
	}
	fh := FileHeader{
		GenericBlock: b,
		magic:        [4]byte(data[0:4]), // +4
		GameID:       read32(data, 4),    // +4
		VersionData:  read16(data, 8),    // +2
		Turn:         read16(data, 10),   // +2
		PlayerData:   read16(data, 12),   // +2
		Flags:        data[15],           // +1
	}
	return &fh, nil
}

func (fh *FileHeader) Magic() string {
	return string([]byte{fh.magic[0], fh.magic[1], fh.magic[2], fh.magic[3]})
}

func (fh *FileHeader) VersionMajor() int {
	return int(fh.VersionData >> 12) // first 4 bits
}

func (fh *FileHeader) VersionMinor() int {
	return int((fh.VersionData >> 5) & 0x7F) // middle 7 bits
}

func (fh *FileHeader) VersionString() string {
	return fmt.Sprintf("%d.%d.%d", fh.VersionMajor(), fh.VersionMinor(), fh.VersionIncrement())
}

func (fh *FileHeader) VersionIncrement() int {
	return int(fh.VersionData & 0x1F) // last 5 bits
}

func (fh *FileHeader) Year() int {
	return 2400 + int(fh.Turn)
}

func (fh *FileHeader) Salt() int {
	return int(fh.PlayerData >> 5) // first 11 bits
}

func (fh *FileHeader) PlayerIndex() int {
	return int(fh.PlayerData & 0x1F) // last 5 bits
}

func (fh *FileHeader) TurnSubmitted() bool {
	return (fh.Flags & (1 << 0)) == 1
}

func (fh *FileHeader) HostUsing() bool {
	return (fh.Flags & (1 << 1)) == 1
}

func (fh *FileHeader) MultipleTurns() bool {
	return (fh.Flags & (1 << 2)) == 1
}

func (fh *FileHeader) GameOver() bool {
	return (fh.Flags & (1 << 3)) == 1
}

func (fh *FileHeader) Shareware() bool {
	return (fh.Flags & (1 << 4)) == 1
}
