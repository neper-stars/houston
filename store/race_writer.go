package store

import (
	"encoding/binary"
	"math/rand"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
	"github.com/neper-stars/houston/race"
)

// CreateRaceFile generates a complete .r1-.r16 file from a race configuration.
// This creates the file from scratch (not from an existing source).
// The playerSlot should be 1-16 (the extension suffix).
func CreateRaceFile(r *race.Race, playerSlot int) ([]byte, error) {
	writer := NewFileWriter()

	// 1. Create FileHeader for race file
	headerData := createRaceFileHeaderData(playerSlot)
	headerBlock := blocks.GenericBlock{
		Type: blocks.FileHeaderBlockType,
		Size: blocks.BlockSize(len(headerData)),
		Data: headerData,
	}
	header, err := blocks.NewFileHeader(headerBlock)
	if err != nil {
		return nil, err
	}
	result := writer.WriteHeader(header)

	// Get salt from header for encryption
	salt := header.Salt()

	// 2. Init encryption (gameId=0, turn=0, playerIndex=31)
	writer.InitEncryption(salt, 0, 0, 31, 0)

	// 3. Build PlayerBlock from Race
	playerBlockData := encodeRaceToPlayerBlock(r, playerSlot)
	result = append(result, writer.WriteEncryptedBlock(blocks.PlayerBlockType, playerBlockData)...)

	// 4. Compute and write footer
	footerData := ComputeRaceFooter(playerBlockData, r.SingularName, r.PluralName)
	result = append(result, writer.WriteFooter(true, footerData)...)

	return result, nil
}

// createRaceFileHeaderData creates the raw 16-byte file header data for a race file.
// Note: Race files use playerIndex=31 in the header for encryption, not the player slot.
// The player slot is indicated by the file extension (.r1, .r2, etc.).
func createRaceFileHeaderData(playerSlot int) []byte {
	_ = playerSlot // Player slot is only used for filename, not stored in header
	data := make([]byte, 16)

	// Magic: "J3J3"
	data[0] = 'J'
	data[1] = '3'
	data[2] = 'J'
	data[3] = '3'

	// GameID: 0 for race files
	binary.LittleEndian.PutUint32(data[4:8], 0)

	// VersionData: v2.7.1 = (2 << 12) | (7 << 5) | 1 = 0x20E1
	versionData := uint16((2 << 12) | (7 << 5) | 1)
	binary.LittleEndian.PutUint16(data[8:10], versionData)

	// Turn: 0 for race files
	binary.LittleEndian.PutUint16(data[10:12], 0)

	// PlayerData: (salt << 5) | playerIndex
	// Race files use playerIndex=31 for encryption (not the player slot)
	// Generate a random salt (11 bits)
	salt := uint16(rand.Intn(2048))
	playerData := (salt << 5) | uint16(31) // Always 31 for race files
	binary.LittleEndian.PutUint16(data[12:14], playerData)

	// Flags: 0 for race files (bytes 14-15)
	data[14] = 0
	data[15] = 0

	return data
}

// encodeRaceToPlayerBlock encodes a race configuration to the PlayerBlock format.
// Race files always have FullDataFlag set.
func encodeRaceToPlayerBlock(r *race.Race, playerSlot int) []byte {
	// Calculate the size we need:
	// - 8 bytes: player header
	// - 104 bytes (0x68): full race data
	// - 1 byte: player relations length (0 for race files)
	// - variable: singular name (encoded)
	// - variable: plural name (encoded)

	singularEncoded := encoding.EncodeStarsString(r.SingularName)
	pluralEncoded := encoding.EncodeStarsString(r.PluralName)

	// Total size: 8 + 104 + 1 + len(singularEncoded) + len(pluralEncoded) + padding
	totalSize := 8 + 104 + 1 + len(singularEncoded) + len(pluralEncoded)
	// Add padding byte if plural name is empty (for 16-bit alignment)
	if len(pluralEncoded) == 1 { // Empty string encodes to just the length byte
		totalSize++
	}

	data := make([]byte, totalSize)
	index := 0

	// Byte 0: Player number (slot)
	data[index] = byte(playerSlot)
	index++

	// Byte 1: Ship design count (0 for race files)
	data[index] = 0
	index++

	// Bytes 2-3: Planets (0 for race files) - 10 bits, high bits in byte 3
	data[index] = 0
	index++
	data[index] = 0
	index++

	// Bytes 4-5: Fleets (0) in low 10 bits, Starbase design count (0) in high 4 bits
	data[index] = 0
	index++
	data[index] = 0
	index++

	// Byte 6: Logo (high 5 bits) | FullDataFlag (bit 2) | fixed bits (bits 0-1 = 3)
	// For race files: FullDataFlag = 1
	logo := r.Icon & 0x1F // 5 bits for logo (0-31)
	byte6 := byte((logo << 3) | 0x04 | 0x03)
	data[index] = byte6
	index++

	// Byte 7: AI settings (for race files, typically 0x01 = human player)
	// Bit 0 = 1 (always), Bit 1 = AI enabled (0), rest = 0
	data[index] = 0x01
	index++

	// Now the full data bytes (104 bytes starting at index 8)
	fullDataStart := index

	// Bytes 8-9: Homeworld (0 for race files)
	binary.LittleEndian.PutUint16(data[fullDataStart:], 0)
	// Bytes 10-11: Rank (0 for race files)
	binary.LittleEndian.PutUint16(data[fullDataStart+2:], 0)
	// Bytes 12-15: Password hash (0 for no password, or encode password)
	if r.Password != "" {
		// TODO: Implement password hashing
		binary.LittleEndian.PutUint32(data[fullDataStart+4:], 0)
	} else {
		binary.LittleEndian.PutUint32(data[fullDataStart+4:], 0)
	}

	// Bytes 16-24: Habitability (9 bytes)
	data[fullDataStart+8] = encodeHabCenter(r.GravityCenter, r.GravityImmune)
	data[fullDataStart+9] = encodeHabCenter(r.TemperatureCenter, r.TemperatureImmune)
	data[fullDataStart+10] = encodeHabCenter(r.RadiationCenter, r.RadiationImmune)
	data[fullDataStart+11] = encodeHabLow(r.GravityCenter, r.GravityWidth, r.GravityImmune)
	data[fullDataStart+12] = encodeHabLow(r.TemperatureCenter, r.TemperatureWidth, r.TemperatureImmune)
	data[fullDataStart+13] = encodeHabLow(r.RadiationCenter, r.RadiationWidth, r.RadiationImmune)
	data[fullDataStart+14] = encodeHabHigh(r.GravityCenter, r.GravityWidth, r.GravityImmune)
	data[fullDataStart+15] = encodeHabHigh(r.TemperatureCenter, r.TemperatureWidth, r.TemperatureImmune)
	data[fullDataStart+16] = encodeHabHigh(r.RadiationCenter, r.RadiationWidth, r.RadiationImmune)

	// Byte 25: Growth rate
	data[fullDataStart+17] = byte(r.GrowthRate)

	// Bytes 26-31: Tech levels (0 for race files - starting levels)
	data[fullDataStart+18] = 0 // Energy
	data[fullDataStart+19] = 0 // Weapons
	data[fullDataStart+20] = 0 // Propulsion
	data[fullDataStart+21] = 0 // Construction
	data[fullDataStart+22] = 0 // Electronics
	data[fullDataStart+23] = 0 // Biotech

	// Bytes 32-55: Tech points (24 bytes, all 0 for race files)
	for i := 24; i < 48; i++ {
		data[fullDataStart+i] = 0
	}

	// Bytes 56-57: Research settings
	data[fullDataStart+48] = 0  // ResearchPercentage (default 0%)
	data[fullDataStart+49] = 0  // CurrentResearchField (0) | NextResearchField (0)

	// Bytes 58-61: Unknown/reserved (set to 0)
	for i := 50; i < 54; i++ {
		data[fullDataStart+i] = 0
	}

	// Bytes 62-68: Production settings
	data[fullDataStart+54] = byte(r.ColonistsPerResource / 100) // ResourcePerColonist (in hundreds)
	data[fullDataStart+55] = byte(r.FactoryOutput)
	data[fullDataStart+56] = byte(r.FactoryCost)
	data[fullDataStart+57] = byte(r.FactoryCount)
	data[fullDataStart+58] = byte(r.MineOutput)
	data[fullDataStart+59] = byte(r.MineCost)
	data[fullDataStart+60] = byte(r.MineCount)

	// Byte 69: Leftover points option
	data[fullDataStart+61] = byte(r.LeftoverPointsOn)

	// Bytes 70-75: Research costs (0=expensive, 1=normal, 2=cheap)
	data[fullDataStart+62] = byte(r.ResearchEnergy)
	data[fullDataStart+63] = byte(r.ResearchWeapons)
	data[fullDataStart+64] = byte(r.ResearchPropulsion)
	data[fullDataStart+65] = byte(r.ResearchConstruction)
	data[fullDataStart+66] = byte(r.ResearchElectronics)
	data[fullDataStart+67] = byte(r.ResearchBiotech)

	// Bytes 76-77: PRT
	data[fullDataStart+68] = byte(r.PRT)
	data[fullDataStart+69] = 0 // Always 0

	// Bytes 78-79: LRT
	binary.LittleEndian.PutUint16(data[fullDataStart+70:], r.LRT)

	// Byte 80: Unknown (set to 0)
	data[fullDataStart+72] = 0

	// Byte 81: Checkboxes
	var checkBoxes byte = 0
	if r.TechsStartHigh {
		checkBoxes |= 0x20 // bit 5
	}
	if r.FactoriesUseLessGerm {
		checkBoxes |= 0x80 // bit 7
	}
	data[fullDataStart+73] = checkBoxes

	// Bytes 82-83: MT Items (0 for race files)
	binary.LittleEndian.PutUint16(data[fullDataStart+74:], 0)

	// Bytes 84-103: Remaining full data (20 bytes, set to 0)
	for i := 76; i < 96; i++ {
		data[fullDataStart+i] = 0
	}

	// Move index past full data section
	index = fullDataStart + 104 // 0x68 = 104 bytes

	// Byte 112 (0x70): Player relations length (0 for race files)
	data[index] = 0
	index++

	// Singular name (encoded)
	copy(data[index:], singularEncoded)
	index += len(singularEncoded)

	// Plural name (encoded)
	copy(data[index:], pluralEncoded)
	index += len(pluralEncoded)

	// Padding if plural name is empty
	if len(pluralEncoded) == 1 {
		data[index] = 0
		index++
	}

	return data[:index]
}

// encodeHabCenter encodes a habitability center value.
// Returns 255 for immune, otherwise the center value in the internal format.
func encodeHabCenter(center int, immune bool) byte {
	if immune {
		return 255
	}
	return byte(center)
}

// encodeHabLow encodes the low bound of a habitability range.
func encodeHabLow(center, width int, immune bool) byte {
	if immune {
		return 255
	}
	low := center - width
	if low < 0 {
		low = 0
	}
	return byte(low)
}

// encodeHabHigh encodes the high bound of a habitability range.
func encodeHabHigh(center, width int, immune bool) byte {
	if immune {
		return 255
	}
	high := center + width
	if high > 100 {
		high = 100
	}
	return byte(high)
}
