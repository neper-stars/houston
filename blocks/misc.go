package blocks

import "github.com/neper-stars/houston/encoding"

// PlayerScoresBlock represents player score history data (Type 45)
// Found in H (history) files, one block per turn per player
type PlayerScoresBlock struct {
	GenericBlock

	PlayerID     int   // Player ID (0-15)
	Turn         int   // Turn number (1-based)
	Score        int   // Player's score for this turn
	Resources    int64 // Resources available
	Planets      int   // Number of planets owned
	Starbases    int   // Number of starbases
	UnarmedShips int   // Number of unarmed ships
	EscortShips  int   // Number of escort ships
	CapitalShips int   // Number of capital ships
	TechLevels   int   // Sum of tech levels
	Rank         int   // Player's rank (derived from Score)
}

// NewPlayerScoresBlock creates a PlayerScoresBlock from a GenericBlock
func NewPlayerScoresBlock(b GenericBlock) *PlayerScoresBlock {
	psb := &PlayerScoresBlock{GenericBlock: b}
	psb.decode()
	return psb
}

func (psb *PlayerScoresBlock) decode() {
	data := psb.Decrypted
	if len(data) < 24 {
		return
	}

	// Bytes 0-1: Player ID and flags
	word0 := encoding.Read16(data, 0)
	psb.PlayerID = int(word0 & 0x0F)

	// Bytes 2-3: Turn number
	psb.Turn = int(encoding.Read16(data, 2))

	// Bytes 4-5: Score
	psb.Score = int(encoding.Read16(data, 4))

	// Bytes 6-7: Padding (always 0)

	// Bytes 8-11: Resources (32-bit)
	psb.Resources = int64(encoding.Read32(data, 8))

	// Bytes 12-13: Planets
	psb.Planets = int(encoding.Read16(data, 12))

	// Bytes 14-15: Starbases
	psb.Starbases = int(encoding.Read16(data, 14))

	// Bytes 16-17: Unarmed ships
	psb.UnarmedShips = int(encoding.Read16(data, 16))

	// Bytes 18-19: Escort ships
	psb.EscortShips = int(encoding.Read16(data, 18))

	// Bytes 20-21: Capital ships
	psb.CapitalShips = int(encoding.Read16(data, 20))

	// Bytes 22-23: Tech levels (sum)
	psb.TechLevels = int(encoding.Read16(data, 22))
}

// SaveAndSubmitBlock represents save and submit action (Type 46)
// Structure not fully documented - preserves raw data for analysis
type SaveAndSubmitBlock struct {
	GenericBlock
}

// NewSaveAndSubmitBlock creates a SaveAndSubmitBlock from a GenericBlock
func NewSaveAndSubmitBlock(b GenericBlock) *SaveAndSubmitBlock {
	return &SaveAndSubmitBlock{GenericBlock: b}
}

// FileHashBlock represents player identification data (Type 9)
// Contains serial number and hardware fingerprint used to detect
// multi-accounting (same serial on different machines).
//
// Format (17 bytes decrypted):
//
//	Bytes 0-1: Unknown (possibly flags or player ID)
//	Bytes 2-5: Serial number (32-bit LE) - from Stars! registration
//	Bytes 6-16: Hardware hash (11 bytes) - machine fingerprint
//
// Hardware hash breakdown:
//
//	Bytes 0-3: Label C: (volume label)
//	Bytes 4-5: C: date/time of volume
//	Bytes 6-8: Label D: (volume label)
//	Byte 9: D: date/time of volume
//	Byte 10: C: and D: drive size in 100's of MB
//
// Used to detect when the same serial number is used
// on different computers (different hardware hash = likely cheating).
type FileHashBlock struct {
	GenericBlock

	// Unknown bytes at start (possibly flags or player ID)
	Unknown uint16

	// Serial number from Stars! registration
	SerialNumber uint32

	// Hardware hash - machine fingerprint (11 bytes)
	HardwareHash []byte

	// Parsed hardware hash components
	LabelC       string // Volume label of C: drive (4 bytes)
	TimestampC   uint16 // C: volume date/time
	LabelD       string // Volume label of D: drive (3 bytes)
	TimestampD   uint8  // D: volume date/time
	DriveSizesMB uint8  // Combined drive sizes in 100s of MB
}

// NewFileHashBlock creates a FileHashBlock from a GenericBlock
func NewFileHashBlock(b GenericBlock) *FileHashBlock {
	fhb := &FileHashBlock{GenericBlock: b}
	fhb.decode()
	return fhb
}

func (fhb *FileHashBlock) decode() {
	data := fhb.Decrypted
	if len(data) < 17 {
		return
	}

	// Bytes 0-1: Unknown
	fhb.Unknown = encoding.Read16(data, 0)

	// Bytes 2-5: Serial number (32-bit LE)
	fhb.SerialNumber = encoding.Read32(data, 2)

	// Bytes 6-16: Hardware hash (11 bytes)
	fhb.HardwareHash = make([]byte, 11)
	copy(fhb.HardwareHash, data[6:17])

	// Parse hardware hash components
	// Bytes 0-3 of hash: Label C: (volume label, null-terminated string)
	fhb.LabelC = string(trimNullBytes(data[6:10]))

	// Bytes 4-5 of hash: C: date/time
	fhb.TimestampC = encoding.Read16(data, 10)

	// Bytes 6-8 of hash: Label D: (volume label, null-terminated string)
	fhb.LabelD = string(trimNullBytes(data[12:15]))

	// Byte 9 of hash: D: date/time
	fhb.TimestampD = data[15]

	// Byte 10 of hash: Drive sizes in 100s of MB
	fhb.DriveSizesMB = data[16]
}

// trimNullBytes removes trailing null bytes from a byte slice
func trimNullBytes(b []byte) []byte {
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] != 0 {
			return b[:i+1]
		}
	}
	return nil
}

// HardwareHashString returns the hardware hash as a hex string for comparison
func (fhb *FileHashBlock) HardwareHashString() string {
	if len(fhb.HardwareHash) == 0 {
		return ""
	}
	result := make([]byte, len(fhb.HardwareHash)*2)
	const hex = "0123456789abcdef"
	for i, b := range fhb.HardwareHash {
		result[i*2] = hex[b>>4]
		result[i*2+1] = hex[b&0x0f]
	}
	return string(result)
}

// WaypointRepeatOrdersBlock represents waypoint repeat orders (Type 10)
// Structure not fully documented - preserves raw data for analysis
type WaypointRepeatOrdersBlock struct {
	GenericBlock
}

// NewWaypointRepeatOrdersBlock creates a WaypointRepeatOrdersBlock from a GenericBlock
func NewWaypointRepeatOrdersBlock(b GenericBlock) *WaypointRepeatOrdersBlock {
	return &WaypointRepeatOrdersBlock{GenericBlock: b}
}

// EventsBlock represents game events (Type 12)
// Structure not fully documented - preserves raw data for analysis
type EventsBlock struct {
	GenericBlock
}

// NewEventsBlock creates an EventsBlock from a GenericBlock
func NewEventsBlock(b GenericBlock) *EventsBlock {
	return &EventsBlock{GenericBlock: b}
}

// MessagesFilterBlock represents message filter settings (Type 33)
// Structure not fully documented - preserves raw data for analysis
type MessagesFilterBlock struct {
	GenericBlock
}

// NewMessagesFilterBlock creates a MessagesFilterBlock from a GenericBlock
func NewMessagesFilterBlock(b GenericBlock) *MessagesFilterBlock {
	return &MessagesFilterBlock{GenericBlock: b}
}

// AiHFileRecordBlock represents AI host file record (Type 41)
// Structure not fully documented - preserves raw data for analysis
type AiHFileRecordBlock struct {
	GenericBlock
}

// NewAiHFileRecordBlock creates an AiHFileRecordBlock from a GenericBlock
func NewAiHFileRecordBlock(b GenericBlock) *AiHFileRecordBlock {
	return &AiHFileRecordBlock{GenericBlock: b}
}

// ManualSmallLoadUnloadTaskBlock represents small load/unload task (Type 1)
// Structure not fully documented - preserves raw data for analysis
type ManualSmallLoadUnloadTaskBlock struct {
	GenericBlock
}

// NewManualSmallLoadUnloadTaskBlock creates a ManualSmallLoadUnloadTaskBlock from a GenericBlock
func NewManualSmallLoadUnloadTaskBlock(b GenericBlock) *ManualSmallLoadUnloadTaskBlock {
	return &ManualSmallLoadUnloadTaskBlock{GenericBlock: b}
}

// ManualMediumLoadUnloadTaskBlock represents medium load/unload task (Type 2)
// Structure not fully documented - preserves raw data for analysis
type ManualMediumLoadUnloadTaskBlock struct {
	GenericBlock
}

// NewManualMediumLoadUnloadTaskBlock creates a ManualMediumLoadUnloadTaskBlock from a GenericBlock
func NewManualMediumLoadUnloadTaskBlock(b GenericBlock) *ManualMediumLoadUnloadTaskBlock {
	return &ManualMediumLoadUnloadTaskBlock{GenericBlock: b}
}

// ManualLargeLoadUnloadTaskBlock represents large load/unload task (Type 25)
// Structure not fully documented - preserves raw data for analysis
type ManualLargeLoadUnloadTaskBlock struct {
	GenericBlock
}

// NewManualLargeLoadUnloadTaskBlock creates a ManualLargeLoadUnloadTaskBlock from a GenericBlock
func NewManualLargeLoadUnloadTaskBlock(b GenericBlock) *ManualLargeLoadUnloadTaskBlock {
	return &ManualLargeLoadUnloadTaskBlock{GenericBlock: b}
}
