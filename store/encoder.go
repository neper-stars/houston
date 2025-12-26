package store

import (
	"encoding/binary"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

// BlockEncoder provides methods to encode entities back into blocks.
type BlockEncoder struct{}

// NewBlockEncoder creates a new block encoder.
func NewBlockEncoder() *BlockEncoder {
	return &BlockEncoder{}
}

// EncodeBlock creates a complete block including the 2-byte header.
// Returns the raw bytes ready for encryption (for data blocks) or direct writing (for header/footer).
func (e *BlockEncoder) EncodeBlock(typeID blocks.BlockTypeID, data []byte) []byte {
	size := len(data)
	if size > 1023 {
		size = 1023 // Max 10 bits
	}

	// Header: type (6 bits) << 10 | size (10 bits)
	header := (uint16(typeID) << 10) | uint16(size)

	result := make([]byte, 2+len(data))
	binary.LittleEndian.PutUint16(result[0:2], header)
	copy(result[2:], data)

	return result
}

// EncodeFleetBlock encodes a FleetEntity back to block data.
// This is complex due to the variable-length encoding, so for now we prefer
// using the preserved raw data when available.
func (e *BlockEncoder) EncodeFleetBlock(fleet *FleetEntity) ([]byte, error) {
	// If the fleet has the original block data, use it
	if fleet.fleetBlock != nil && !fleet.Meta().Dirty {
		return fleet.fleetBlock.DecryptedData(), nil
	}

	// TODO: Full encoding of dirty fleets
	// For now, return the original data if available
	if fleet.fleetBlock != nil {
		return encodeModifiedFleet(fleet)
	}

	return nil, ErrNoRawBlockData
}

// encodeModifiedFleet attempts to encode a modified fleet.
// This creates new block data based on the entity's current state.
func encodeModifiedFleet(fleet *FleetEntity) ([]byte, error) {
	fb := fleet.fleetBlock
	if fb == nil {
		return nil, ErrNoRawBlockData
	}

	// Start with the original decrypted data as a template
	data := make([]byte, len(fb.DecryptedData()))
	copy(data, fb.DecryptedData())

	// Update cargo values if the fleet has cargo data
	if fb.HasCargo() {
		// Find the cargo section offset by replaying the decode
		index := 14 // After fixed header

		// Skip ship counts
		for bit := 0; bit < 16; bit++ {
			if (fb.ShipTypes & (1 << bit)) != 0 {
				if fb.ShipCountTwoBytes {
					index += 2
				} else {
					index++
				}
			}
		}

		// Now at cargo section - encode new cargo values
		if index+2 <= len(data) {
			// Calculate the length indicators for the new cargo values
			contentsLengths := uint16(0)
			contentsLengths |= uint16(encoding.ByteLengthForInt(fleet.ironium)) << 0
			contentsLengths |= uint16(encoding.ByteLengthForInt(fleet.boranium)) << 2
			contentsLengths |= uint16(encoding.ByteLengthForInt(fleet.germanium)) << 4
			contentsLengths |= uint16(encoding.ByteLengthForInt(fleet.population)) << 6
			contentsLengths |= uint16(encoding.ByteLengthForInt(fleet.fuel)) << 8

			binary.LittleEndian.PutUint16(data[index:], contentsLengths)
			index += 2

			// Write each cargo value with variable length encoding
			index = writeVarLen(data, index, fleet.ironium)
			index = writeVarLen(data, index, fleet.boranium)
			index = writeVarLen(data, index, fleet.germanium)
			index = writeVarLen(data, index, fleet.population)
			index = writeVarLen(data, index, fleet.fuel)
		}
	}

	return data, nil
}

// writeVarLen writes a variable-length encoded integer and returns the new index.
func writeVarLen(data []byte, index int, value int64) int {
	byteLen := encoding.ByteLengthForInt(value)
	switch byteLen {
	case 0:
		// Zero uses no bytes
		return index
	case 1:
		if index < len(data) {
			data[index] = byte(value & 0xFF)
		}
		return index + 1
	case 2:
		if index+1 < len(data) {
			binary.LittleEndian.PutUint16(data[index:], uint16(value))
		}
		return index + 2
	case 3: // Actually means 4 bytes
		if index+3 < len(data) {
			binary.LittleEndian.PutUint32(data[index:], uint32(value))
		}
		return index + 4
	}
	return index
}

// EncodeProductionQueueBlock encodes a ProductionQueueEntity back to block data.
func (e *BlockEncoder) EncodeProductionQueueBlock(pq *ProductionQueueEntity) ([]byte, error) {
	if pq.queueBlock != nil && !pq.Meta().Dirty {
		return pq.queueBlock.DecryptedData(), nil
	}

	// Encode the production queue items
	// Format: Each item is 4 bytes
	// Bits 0-5: item ID (6 bits)
	// Bits 6-15: count (10 bits)
	// Bits 16-27: complete percent (12 bits)
	// Bits 28-31: item type (4 bits, but only 2-4 used)

	data := make([]byte, len(pq.Items)*4)
	for i, item := range pq.Items {
		// Pack the item into 4 bytes
		packed := uint32(item.ItemId & 0x3F)               // 6 bits
		packed |= uint32(item.Count&0x3FF) << 6            // 10 bits
		packed |= uint32(item.CompletePercent&0xFFF) << 16 // 12 bits
		packed |= uint32(item.ItemType&0x0F) << 28         // 4 bits

		binary.LittleEndian.PutUint32(data[i*4:], packed)
	}

	return data, nil
}

// EncodeBattlePlanBlock encodes a BattlePlanEntity back to block data.
func (e *BlockEncoder) EncodeBattlePlanBlock(bp *BattlePlanEntity) ([]byte, error) {
	if bp.battlePlanBlock != nil && !bp.Meta().Dirty {
		return bp.battlePlanBlock.DecryptedData(), nil
	}

	// For now, return original data if available
	if bp.battlePlanBlock != nil {
		return bp.battlePlanBlock.DecryptedData(), nil
	}

	return nil, ErrNoRawBlockData
}

// GetRawBlockData returns the raw decrypted block data for an entity if available.
// This is the preferred method when the entity hasn't been modified.
func GetRawBlockData(entity Entity) (blocks.BlockTypeID, []byte, bool) {
	rawBlocks := entity.RawBlocks()
	if len(rawBlocks) == 0 {
		return 0, nil, false
	}

	// Return the first block's data
	block := rawBlocks[0]
	return block.BlockTypeID(), block.DecryptedData(), true
}

// ComputeRaceFooter calculates the correct footer checksum for decrypted race data.
//
// The algorithm
//  1. Take decrypted PlayerBlock data up to (but not including) the nibble-packed names
//  2. Decode the singular and plural race names to ASCII
//  3. Pad each name to 15 characters with a leading 0
//  4. Interleave the name bytes: singular[0:2], plural[0:2], singular[2:4], plural[2:4], ...
//  5. checkSum1 = XOR of all even-indexed bytes
//  6. checkSum2 = XOR of all odd-indexed bytes
//  7. Return (checkSum2 << 8) | checkSum1
func ComputeRaceFooter(decryptedData []byte, singularName, pluralName string) uint16 {
	// Find where the name data starts
	// Structure: first 8 bytes are header, then if fullDataFlag is set at byte 6 bit 2,
	// there's 0x68 bytes of full data, then playerRelations, then names
	fullDataFlag := (decryptedData[6] & 0x04) != 0
	index := 8
	if fullDataFlag {
		index = 0x70 // 112 bytes: 8 header + 0x68 (104) full data
		playerRelationsLength := int(decryptedData[index])
		index += 1 + playerRelationsLength
	}

	// Data length is everything before the names section
	dataLength := index

	// Build the checksum data array
	var dData []byte
	dData = append(dData, decryptedData[:dataLength]...)

	// Prepare singular name: leading 0, ASCII bytes, padded to 16 total
	singularOrd := make([]byte, 16)
	singularOrd[0] = 0
	for i, c := range singularName {
		if i < 15 {
			singularOrd[i+1] = byte(c)
		}
	}

	// Prepare plural name: leading 0, ASCII bytes, padded to 16 total
	pluralOrd := make([]byte, 16)
	pluralOrd[0] = 0
	for i, c := range pluralName {
		if i < 15 {
			pluralOrd[i+1] = byte(c)
		}
	}

	// Interleave: add pairs from singular, then pairs from plural
	for i := 0; i < 16; i += 2 {
		dData = append(dData, singularOrd[i], singularOrd[i+1])
		dData = append(dData, pluralOrd[i], pluralOrd[i+1])
	}

	// Compute checksums
	var checkSum1, checkSum2 byte
	for i := 0; i < len(dData); i += 2 {
		checkSum1 ^= dData[i]
	}
	for i := 1; i < len(dData); i += 2 {
		checkSum2 ^= dData[i]
	}

	return uint16(checkSum1) | uint16(checkSum2)<<8
}
