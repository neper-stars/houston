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

// File type constants (byte 14 of file header).
// Source: dt field in RTBOF structure (types.h)
const (
	FileTypeUnknown = 0 // Unknown/Invalid - also used by XY (universe) files
	FileTypeXY      = 0 // Universe definition file (.xy) - same value as Unknown
	FileTypeX       = 1 // Turn order file (.x1-.x16) - wGen is validated for this type!
	FileTypeHST     = 2 // Host file (.hst)
	FileTypeM       = 3 // Player turn file (.m1-.m16)
	FileTypeH       = 4 // History file (.h1-.h16)
	FileTypeRace    = 5 // Race file (.r1-.r16)
)

// Flags byte constants (byte 15 of file header).
// Source: RTBOF structure (types.h)
const (
	FlagDone        = 0x01 // fDone: Turn has been submitted
	FlagInUse       = 0x02 // fInUse: File is currently in use by host
	FlagMulti       = 0x04 // fMulti: Multiplayer game / multiple turns
	FlagGameOverMan = 0x08 // fGameOverMan: Game has ended
	FlagCrippled    = 0x10 // fCrippled: Crippled/demo mode
	FlagGenMask     = 0xE0 // wGen: Generation counter (bits 5-7)
	FlagGenShift    = 5    // Shift for wGen field
)

// Game year and encryption constants
const (
	StarsBaseYear       = 2400 // Base year for turn calculation (Year = 2400 + Turn)
	MaxSaltValue        = 2048 // Maximum salt value (11 bits)
	RaceFilePlayerIndex = 31   // Player index used for race file encryption
	PlayerIndexBits     = 5    // Number of bits for player index
	SaltBits            = 11   // Number of bits for salt
	PlayerIndexMask     = 0x1F // Mask for 5-bit player index
	SaltShift           = 5    // Bits to shift for salt in PlayerData
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

// FileHeader is a specialized block that implements the Block interface
// for the file header (RTBOF structure in types.h).
//
// Structure (16 bytes):
//
//	Bytes 0-3:   Magic number ("J3D1" or "J3J3")
//	Bytes 4-7:   Game ID (32-bit)
//	Bytes 8-9:   Version data (encoded: major<<12 | minor<<5 | increment)
//	Bytes 10-11: Turn number
//	Bytes 12-13: Player data (salt<<5 | playerIndex)
//	Byte 14:     File type (dt)
//	Byte 15:     Flags (fDone, fInUse, fMulti, fGameOverMan, fCrippled, wGen)
type FileHeader struct {
	GenericBlock
	magic       [4]byte
	GameID      uint32
	VersionData uint16
	Turn        uint16
	PlayerData  uint16
	FileType    uint8 // byte 14: file type (dt) - see FileType* constants
	Flags       uint8 // byte 15: flags + wGen - see Flag* constants
}

// NewFileHeader is constructor that takes a GenericBlock and returns
// a pointer to a FileHeader
func NewFileHeader(b GenericBlock) (*FileHeader, error) {
	data := b.BlockData()
	if len(data) < 16 {
		return nil, ErrInvalidFileHeaderBlock
	}
	fh := FileHeader{
		GenericBlock: b,
		magic:        [4]byte(data[0:4]),        // +4 (bytes 0-3)
		GameID:       encoding.Read32(data, 4),  // +4 (bytes 4-7)
		VersionData:  encoding.Read16(data, 8),  // +2 (bytes 8-9)
		Turn:         encoding.Read16(data, 10), // +2 (bytes 10-11)
		PlayerData:   encoding.Read16(data, 12), // +2 (bytes 12-13)
		FileType:     data[14],                  // +1 (byte 14)
		Flags:        data[15],                  // +1 (byte 15)
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
	return StarsBaseYear + int(fh.Turn)
}

func (fh *FileHeader) Salt() int {
	return int(fh.PlayerData >> SaltShift)
}

func (fh *FileHeader) PlayerIndex() int {
	return int(fh.PlayerData & PlayerIndexMask)
}

// TurnSubmitted returns true if fDone flag is set (turn has been submitted).
func (fh *FileHeader) TurnSubmitted() bool {
	return (fh.Flags & FlagDone) != 0
}

// HostUsing returns true if fInUse flag is set (file is currently in use by host).
func (fh *FileHeader) HostUsing() bool {
	return (fh.Flags & FlagInUse) != 0
}

// MultipleTurns returns true if fMulti flag is set (multiplayer game).
func (fh *FileHeader) MultipleTurns() bool {
	return (fh.Flags & FlagMulti) != 0
}

// GameOver returns true if fGameOverMan flag is set (game has ended).
func (fh *FileHeader) GameOver() bool {
	return (fh.Flags & FlagGameOverMan) != 0
}

// Crippled returns true if fCrippled flag is set (demo/crippled mode).
func (fh *FileHeader) Crippled() bool {
	return (fh.Flags & FlagCrippled) != 0
}

// Shareware is an alias for Crippled() for backward compatibility.
//
// Deprecated: Use Crippled() instead.
func (fh *FileHeader) Shareware() bool {
	return fh.Crippled()
}

// Generation returns the wGen field (3-bit generation counter, bits 5-7).
// This tracks file versioning for consistency checks.
//
// IMPORTANT: wGen validation depends on file type:
//   - X files (dt=1): wGen MUST match the host's game state when loaded
//   - M files (dt=3): wGen is NOT validated, any value 0-7 is accepted
//   - HST files (dt=2): wGen is NOT validated
//
// Source: In WriteBOF, this value comes from game.wCrap >> 9.
func (fh *FileHeader) Generation() int {
	return int((fh.Flags & FlagGenMask) >> FlagGenShift)
}

// IsGenerationValidated returns true if wGen is validated for this file type.
// Only X files (dt=1) have wGen validation when loaded by the host.
func (fh *FileHeader) IsGenerationValidated() bool {
	return fh.FileType == FileTypeX
}

// FileTypeName returns a human-readable name for the file type.
func (fh *FileHeader) FileTypeName() string {
	switch fh.FileType {
	case FileTypeXY: // Also FileTypeUnknown (0) - XY files use this value
		return "XY"
	case FileTypeX:
		return "X"
	case FileTypeHST:
		return "HST"
	case FileTypeM:
		return "M"
	case FileTypeH:
		return "H"
	case FileTypeRace:
		return "Race"
	default:
		return fmt.Sprintf("Unknown(%d)", fh.FileType)
	}
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

	// FileType (byte 14) and Flags (byte 15)
	data[14] = fh.FileType
	data[15] = fh.Flags

	return data
}

// NewFileHeaderForRaceFile creates a FileHeader configured for race files.
// Race files use GameID=0, Turn=0, playerIndex=31 for encryption.
// A random salt is generated for the encryption.
func NewFileHeaderForRaceFile() *FileHeader {
	// Generate random salt (11 bits)
	salt := uint16(rand.Intn(MaxSaltValue))
	playerData := (salt << SaltShift) | uint16(RaceFilePlayerIndex)

	return &FileHeader{
		magic:       [4]byte{'J', '3', 'J', '3'},
		GameID:      0,
		VersionData: StarsVersionData(),
		Turn:        0,
		PlayerData:  playerData,
		FileType:    FileTypeRace,
		Flags:       0,
	}
}

// SetSalt sets the encryption salt (11 bits) while preserving playerIndex.
func (fh *FileHeader) SetSalt(salt int) {
	playerIndex := fh.PlayerData & PlayerIndexMask
	fh.PlayerData = (uint16(salt&0x7FF) << SaltShift) | playerIndex
}

// SetPlayerIndex sets the player index (5 bits) while preserving salt.
func (fh *FileHeader) SetPlayerIndex(playerIndex int) {
	salt := fh.PlayerData >> SaltShift
	fh.PlayerData = (salt << SaltShift) | uint16(playerIndex&PlayerIndexMask)
}
