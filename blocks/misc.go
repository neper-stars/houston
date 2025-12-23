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

// Event type constants for production-related events
const (
	EventTypeDefensesBuilt            = 0x35 // Defenses built on planet
	EventTypeFactoriesBuilt           = 0x36 // Factories built on planet
	EventTypeMineralAlchemyBuilt      = 0x37 // Mineral alchemy or similar
	EventTypeMinesBuilt               = 0x38 // Mines built on planet
	EventTypeQueueEmpty               = 0x3E // Production queue empty
	EventTypePopulationChange         = 0x26 // Population changed (decrease due to overcrowding, etc.)
	EventTypeResearchComplete         = 0x50 // Research level completed
	EventTypeTerraformablePlanetFound = 0x57 // Terraformable planet found
	EventTypeTechBenefit              = 0x5F // Tech benefit gained
	EventTypePacketCaptured           = 0xD5 // Mineral packet captured at planet
	EventTypeMineralPacketProduced    = 0xD3 // Mineral packet produced (launched from mass driver)
	EventTypePacketBombardment        = 0xD8 // Mineral packet bombardment (uncaught packet hit planet)
)

// Research field IDs
const (
	ResearchFieldEnergy        = 0
	ResearchFieldWeapons       = 1
	ResearchFieldPropulsion    = 2
	ResearchFieldConstruction  = 3
	ResearchFieldElectronics   = 4
	ResearchFieldBiotechnology = 5
)

// ResearchFieldName returns the human-readable name for a research field
func ResearchFieldName(field int) string {
	names := []string{"Energy", "Weapons", "Propulsion", "Construction", "Electronics", "Biotechnology"}
	if field >= 0 && field < len(names) {
		return names[field]
	}
	return "Unknown"
}

// ProductionEvent represents a single production-related event
type ProductionEvent struct {
	EventType int // Event type (0x35-0x3E for production)
	PlanetID  int // Planet where event occurred
	Count     int // Count (for factories/mines built)
}

// ResearchCompleteEvent represents a research level completion
type ResearchCompleteEvent struct {
	Level     int // Tech level achieved (1-26)
	Field     int // Research field completed (0-5)
	NextField int // Next research field (where research continues)
}

// TechBenefitEvent represents gaining a tech benefit from research
type TechBenefitEvent struct {
	ItemID   int // Item ID gained
	Category int // Category/index within field
}

// TerraformablePlanetFoundEvent represents finding a terraformable planet
type TerraformablePlanetFoundEvent struct {
	GrowthRateEncoded int     // Raw encoded growth rate value
	GrowthRatePercent float64 // Calculated growth rate (encoded / 332)
}

// PopulationChangeEvent represents population change on a planet
// (e.g., decrease due to overcrowding)
type PopulationChangeEvent struct {
	PlanetID int // Planet where population changed
	Amount   int // Amount of change in colonists (e.g., 200 = 200 colonists died)
}

// PacketCapturedEvent represents capturing a mineral packet at a planet
type PacketCapturedEvent struct {
	PlanetID      int // Planet where packet was captured
	MineralAmount int // Total mineral amount in kT
}

// MineralPacketProducedEvent represents a mineral packet launched from a mass driver
// Note: Source planet encoding is not fully understood - the parsed SourcePlanetID
// may not match the actual planet name shown in the game message
type MineralPacketProducedEvent struct {
	SourcePlanetID      int // Planet with mass driver (encoding not fully confirmed)
	DestinationPlanetID int // Target planet for the packet
}

// PacketBombardmentEvent represents a mineral packet hitting a planet (uncaught)
// This occurs when a packet arrives at a planet that can't catch it
type PacketBombardmentEvent struct {
	PlanetID        int // Planet that was bombarded
	MineralAmount   int // Total mineral amount in kT
	ColonistsKilled int // Number of colonists killed
}

// EventsBlock represents game events (Type 12)
type EventsBlock struct {
	GenericBlock

	ProductionEvents         []ProductionEvent               // Decoded production events
	ResearchEvents           []ResearchCompleteEvent         // Research completion events
	TechBenefits             []TechBenefitEvent              // Tech benefits gained
	TerraformablePlanets     []TerraformablePlanetFoundEvent // Terraformable planets found
	PopulationChanges        []PopulationChangeEvent         // Population change events
	PacketsCaptured          []PacketCapturedEvent           // Packet captured events
	PacketsProduced          []MineralPacketProducedEvent    // Packet produced events
	PacketBombardments       []PacketBombardmentEvent        // Packet bombardment events
}

// NewEventsBlock creates an EventsBlock from a GenericBlock
func NewEventsBlock(b GenericBlock) *EventsBlock {
	eb := &EventsBlock{
		GenericBlock: b,
	}
	eb.decode()
	return eb
}

func (eb *EventsBlock) decode() {
	data := eb.Decrypted
	if len(data) < 5 {
		return
	}

	// Parse production events sequentially to maintain order
	// Then parse other event types by scanning
	eb.parseProductionEvents(data)
	eb.parseResearchEvents(data)
}

// parseProductionEvents parses production events sequentially while maintaining order
func (eb *EventsBlock) parseProductionEvents(data []byte) {
	// Production events (types 0x35-0x3E)
	// Format: type flags planetID[2] [count] checksum
	// - Types 0x35, 0x37, 0x3E: 5 bytes (no count)
	// - Types 0x36, 0x38: 6 bytes (with count in byte 4)
	//
	// Events with flags=0x00 are simple production events
	// Events with other flags may have different formats

	i := 0
	for i < len(data) {
		if i+5 > len(data) {
			break
		}

		eventType := int(data[i])
		flags := data[i+1]

		// Only process production events with flags=0x00
		if flags != 0x00 {
			// Skip this byte and continue looking for more events
			i++
			continue
		}

		planetID := int(data[i+2]) | (int(data[i+3]) << 8)

		var eventLen int
		var count int

		switch eventType {
		case EventTypeDefensesBuilt, EventTypeMineralAlchemyBuilt, EventTypeQueueEmpty:
			eventLen = 5
			count = 1 // Default count for DefensesBuilt
			if eventType != EventTypeDefensesBuilt {
				count = 0
			}
		case EventTypeFactoriesBuilt, EventTypeMinesBuilt:
			if i+6 > len(data) {
				i++
				continue
			}
			eventLen = 6
			count = int(data[i+4])
		default:
			// Unknown event type, skip this byte and continue
			i++
			continue
		}

		eb.ProductionEvents = append(eb.ProductionEvents, ProductionEvent{
			EventType: eventType,
			PlanetID:  planetID,
			Count:     count,
		})

		i += eventLen
	}
}

func (eb *EventsBlock) parseResearchEvents(data []byte) {
	// Research Complete Event Format (7 bytes):
	//   Byte 0: 0x50 (EventTypeResearchComplete)
	//   Byte 1: 0x00 (flags)
	//   Bytes 2-3: 0xFFFE = "no planet" marker (research is global, not planet-specific)
	//              Production events have planet IDs here; research uses -2/0xFFFE instead
	//   Byte 4: Level achieved (1-26)
	//   Byte 5: Field completed (0-5)
	//   Byte 6: Next research field (where research will continue, 0-5)
	//
	// Format confirmed by cross-referencing Player 1 (with population events)
	// and Player 2 (without population events) - both use identical structure.

	// Search for research complete events (0x50)
	for i := 0; i < len(data)-6; i++ {
		if data[i] == EventTypeResearchComplete && data[i+1] == 0x00 &&
			data[i+2] == 0xFE && data[i+3] == 0xFF {
			level := int(data[i+4])
			field := int(data[i+5])
			nextField := int(data[i+6])
			// Validate: level should be 1-26, fields should be 0-5
			if level >= 1 && level <= 26 && field <= 5 && nextField <= 5 {
				eb.ResearchEvents = append(eb.ResearchEvents, ResearchCompleteEvent{
					Level:     level,
					Field:     field,
					NextField: nextField,
				})
			}
		}
	}

	// Search for tech benefit events (0x5F)
	// Format: 5F flags category itemID[2] extra[2] (7 bytes total)
	for i := 0; i < len(data)-6; i++ {
		if data[i] == EventTypeTechBenefit {
			category := int(data[i+2])
			itemID := int(data[i+3]) | (int(data[i+4]) << 8)
			eb.TechBenefits = append(eb.TechBenefits, TechBenefitEvent{
				ItemID:   itemID,
				Category: category,
			})
		}
	}

	// Search for terraformable planet found events (0x57)
	// Format: 57 flags ?? ?? ?? ?? GG GG (8 bytes total)
	// Bytes 6-7: Growth rate encoded (16-bit LE), divide by 332 for percentage
	for i := 0; i < len(data)-7; i++ {
		if data[i] == EventTypeTerraformablePlanetFound {
			growthEncoded := int(data[i+6]) | (int(data[i+7]) << 8)
			growthPercent := float64(growthEncoded) / 332.0
			eb.TerraformablePlanets = append(eb.TerraformablePlanets, TerraformablePlanetFoundEvent{
				GrowthRateEncoded: growthEncoded,
				GrowthRatePercent: growthPercent,
			})
		}
	}

	// Search for population change events (0x26)
	// Format: 26 00 PP PP CC AA AA (7 bytes total)
	//   Bytes 2-3: Planet ID (16-bit LE)
	//   Byte 4: Checksum/repeat of planet low byte
	//   Bytes 5-6: Amount in hundreds of colonists (e.g., 2 = 200 colonists)
	for i := 0; i < len(data)-6; i++ {
		if data[i] == EventTypePopulationChange && data[i+1] == 0x00 {
			planetID := int(data[i+2]) | (int(data[i+3]) << 8)
			amountHundreds := int(data[i+5]) | (int(data[i+6]) << 8)
			amount := amountHundreds * 100
			eb.PopulationChanges = append(eb.PopulationChanges, PopulationChangeEvent{
				PlanetID: planetID,
				Amount:   amount,
			})
		}
	}

	// Search for packet captured events (0xD5)
	// Format: D5 00 PP PP PP PP MM MM (8 bytes total)
	//   Bytes 2-3: Planet ID (16-bit LE)
	//   Bytes 4-5: Planet ID repeated
	//   Bytes 6-7: Mineral amount in kT (16-bit LE)
	for i := 0; i < len(data)-7; i++ {
		if data[i] == EventTypePacketCaptured && data[i+1] == 0x00 {
			planetID := int(data[i+2]) | (int(data[i+3]) << 8)
			mineralAmount := int(data[i+6]) | (int(data[i+7]) << 8)
			eb.PacketsCaptured = append(eb.PacketsCaptured, PacketCapturedEvent{
				PlanetID:      planetID,
				MineralAmount: mineralAmount,
			})
		}
	}

	// Search for mineral packet produced events (0xD3)
	// Format: D3 00 SS SS SS DD (6 bytes)
	//   Bytes 2-3: Source planet ID (16-bit LE) - NOTE: encoding not fully confirmed
	//   Byte 4: Repeat of source low byte
	//   Byte 5: Destination planet ID (low byte only observed)
	for i := 0; i < len(data)-5; i++ {
		if data[i] == EventTypeMineralPacketProduced && data[i+1] == 0x00 {
			sourcePlanetID := int(data[i+2]) | (int(data[i+3]) << 8)
			destPlanetID := int(data[i+5]) // Only low byte observed
			eb.PacketsProduced = append(eb.PacketsProduced, MineralPacketProducedEvent{
				SourcePlanetID:      sourcePlanetID,
				DestinationPlanetID: destPlanetID,
			})
		}
	}

	// Search for packet bombardment events (0xD8)
	// Format: D8 00 PP PP XX MM MM 00 DD (9 bytes)
	//   Bytes 2-3: Planet ID (16-bit LE)
	//   Byte 4: Unknown (often same as planet low byte)
	//   Bytes 5-6: Mineral amount in kT (16-bit LE)
	//   Byte 7: Unknown (always 0x00 observed)
	//   Byte 8: Colonists killed / 100
	for i := 0; i < len(data)-8; i++ {
		if data[i] == EventTypePacketBombardment && data[i+1] == 0x00 {
			planetID := int(data[i+2]) | (int(data[i+3]) << 8)
			mineralAmount := int(data[i+5]) | (int(data[i+6]) << 8)
			colonistsKilled := int(data[i+8]) * 100
			eb.PacketBombardments = append(eb.PacketBombardments, PacketBombardmentEvent{
				PlanetID:        planetID,
				MineralAmount:   mineralAmount,
				ColonistsKilled: colonistsKilled,
			})
		}
	}
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

// Cargo transfer direction constants
const (
	CargoTransferLoad   = 0x10 // Load: from target to fleet
	CargoTransferUnload = 0x00 // Unload: from fleet to target
)

// ManualSmallLoadUnloadTaskBlock represents small load/unload task (Type 1)
// Used for cargo transfers where amounts fit in single bytes (-128 to 127 kT each)
// Target can be a planet or another fleet
//
// For fleet-to-fleet transfers, amounts are signed:
//   - Positive = load (receive from target)
//   - Negative = unload (give to target)
type ManualSmallLoadUnloadTaskBlock struct {
	GenericBlock

	FleetNumber  int // Fleet performing the transfer (0-indexed)
	TargetNumber int // Target planet or fleet number (0-indexed)
	TaskByte     int // Raw task/flags byte for analysis
	CargoMask    int // Bitmask of cargo types present (bit 0=Iron, 1=Bor, 2=Germ, 3=Colonists)

	// Cargo amounts (signed byte each, -128 to 127 kT)
	// For fleet-to-fleet: positive = load, negative = unload
	Ironium   int
	Boranium  int
	Germanium int
	Colonists int
}

// NewManualSmallLoadUnloadTaskBlock creates a ManualSmallLoadUnloadTaskBlock from a GenericBlock
func NewManualSmallLoadUnloadTaskBlock(b GenericBlock) *ManualSmallLoadUnloadTaskBlock {
	block := &ManualSmallLoadUnloadTaskBlock{GenericBlock: b}
	block.decode()
	return block
}

func (b *ManualSmallLoadUnloadTaskBlock) decode() {
	data := b.Decrypted
	if len(data) < 10 {
		return
	}

	// Bytes 0-1: Fleet number (16-bit)
	b.FleetNumber = int(encoding.Read16(data, 0))

	// Bytes 2-3: Target number (planet or fleet, 16-bit)
	b.TargetNumber = int(encoding.Read16(data, 2))

	// Byte 4: Task/flags byte (direction and other flags)
	b.TaskByte = int(data[4])

	// Byte 5: Cargo type bitmask
	b.CargoMask = int(data[5])

	// Bytes 6-9: Cargo amounts (signed bytes for fleet-to-fleet transfers)
	b.Ironium = int(int8(data[6]))
	b.Boranium = int(int8(data[7]))
	b.Germanium = int(int8(data[8]))
	b.Colonists = int(int8(data[9]))
}

// IsLoad returns true if this is a load operation (target -> fleet)
// Bit 4 (0x10) of TaskByte indicates load direction
func (b *ManualSmallLoadUnloadTaskBlock) IsLoad() bool {
	return (b.TaskByte & CargoTransferLoad) != 0
}

// IsUnload returns true if this is an unload operation (fleet -> target)
func (b *ManualSmallLoadUnloadTaskBlock) IsUnload() bool {
	return !b.IsLoad()
}

// HasIronium returns true if ironium is being transferred
func (b *ManualSmallLoadUnloadTaskBlock) HasIronium() bool {
	return (b.CargoMask & 0x01) != 0
}

// HasBoranium returns true if boranium is being transferred
func (b *ManualSmallLoadUnloadTaskBlock) HasBoranium() bool {
	return (b.CargoMask & 0x02) != 0
}

// HasGermanium returns true if germanium is being transferred
func (b *ManualSmallLoadUnloadTaskBlock) HasGermanium() bool {
	return (b.CargoMask & 0x04) != 0
}

// HasColonists returns true if colonists are being transferred
func (b *ManualSmallLoadUnloadTaskBlock) HasColonists() bool {
	return (b.CargoMask & 0x08) != 0
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
