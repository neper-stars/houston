package blocks

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/neper-stars/houston/encoding"
)

var ErrInvalidFileHeaderBlock = errors.New("invalid file header")

// Stars! version constants (Stars! 2.60j RC4 writes version 2.83.0)
const (
	StarsVersionMajor     = 2
	StarsVersionMinor     = 83
	StarsVersionIncrement = 0
)

// StarsVersionData returns the encoded version for Stars! 2.60j RC4 (reports as 2.83.0)
func StarsVersionData() uint16 {
	return EncodeVersion(StarsVersionMajor, StarsVersionMinor, StarsVersionIncrement)
}

// EncodeVersion encodes major, minor, increment into Stars! version format.
// Format: major (4 bits) << 12 | minor (7 bits) << 5 | increment (5 bits)
func EncodeVersion(major, minor, increment int) uint16 {
	return uint16((major&0x0F)<<12 | (minor&0x7F)<<5 | (increment & 0x1F))
}

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
		magic:        [4]byte(data[0:4]),          // +4
		GameID:       encoding.Read32(data, 4),    // +4
		VersionData:  encoding.Read16(data, 8),    // +2
		Turn:         encoding.Read16(data, 10),   // +2
		PlayerData:   encoding.Read16(data, 12),   // +2
		Flags:        data[15],                    // +1
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

// Encode returns the raw 16-byte file header data.
func (fh *FileHeader) Encode() []byte {
	data := make([]byte, 16)

	// Magic
	copy(data[0:4], fh.magic[:])

	// GameID
	encoding.Write32(data, 4, fh.GameID)

	// VersionData
	encoding.Write16(data, 8, fh.VersionData)

	// Turn
	encoding.Write16(data, 10, fh.Turn)

	// PlayerData
	encoding.Write16(data, 12, fh.PlayerData)

	// Flags (bytes 14-15, only byte 15 used)
	data[14] = 0
	data[15] = fh.Flags

	return data
}

// NewFileHeaderForRaceFile creates a FileHeader configured for race files.
// Race files use GameID=0, Turn=0, playerIndex=31 for encryption.
// A random salt is generated for the encryption.
func NewFileHeaderForRaceFile() *FileHeader {
	// Generate random salt (11 bits)
	salt := uint16(rand.Intn(2048))
	playerData := (salt << 5) | uint16(31) // Always playerIndex=31 for race files

	return &FileHeader{
		magic:       [4]byte{'J', '3', 'J', '3'},
		GameID:      0,
		VersionData: StarsVersionData(),
		Turn:        0,
		PlayerData:  playerData,
		Flags:       0,
	}
}

// SetSalt sets the encryption salt (11 bits) while preserving playerIndex.
func (fh *FileHeader) SetSalt(salt int) {
	playerIndex := fh.PlayerData & 0x1F
	fh.PlayerData = (uint16(salt&0x7FF) << 5) | playerIndex
}

// SetPlayerIndex sets the player index (5 bits) while preserving salt.
func (fh *FileHeader) SetPlayerIndex(playerIndex int) {
	salt := fh.PlayerData >> 5
	fh.PlayerData = (salt << 5) | uint16(playerIndex&0x1F)
}
