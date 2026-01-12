package store

import (
	"fmt"
	"os"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

// DebugEncoding enables debug logging for encoding operations.
// Set to true to log to houston_debug.log in current directory.
var DebugEncoding = false

func debugLog(format string, args ...interface{}) {
	if !DebugEncoding {
		return
	}
	f, err := os.OpenFile("houston_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = fmt.Fprintf(f, format, args...)
}

// BlockEncoder provides methods to encode entities back into blocks.
type BlockEncoder struct{}

// NewBlockEncoder creates a new block encoder.
func NewBlockEncoder() *BlockEncoder {
	return &BlockEncoder{}
}

// EncodeBlock creates a complete block including the 2-byte header.
// Returns the raw bytes ready for encryption (for data blocks) or direct writing (for header/footer).
func (e *BlockEncoder) EncodeBlock(typeID blocks.BlockTypeID, data []byte) []byte {
	return blocks.EncodeBlockWithHeader(typeID, data)
}

// EncodeFleetBlock encodes a FleetEntity back to block data.
// Uses the block's Encode() method if the fleet has been modified.
func (e *BlockEncoder) EncodeFleetBlock(fleet *FleetEntity) ([]byte, error) {
	// If the fleet has the original block data and hasn't been modified, use it
	if fleet.fleetBlock != nil && !fleet.Meta().Dirty {
		return fleet.fleetBlock.DecryptedData(), nil
	}

	// If the fleet has been modified but has a block, encode it
	if fleet.fleetBlock != nil {
		// Update the block with current values and encode
		fleet.fleetBlock.Ironium = fleet.ironium
		fleet.fleetBlock.Boranium = fleet.boranium
		fleet.fleetBlock.Germanium = fleet.germanium
		fleet.fleetBlock.Population = fleet.population
		fleet.fleetBlock.Fuel = fleet.fuel
		return fleet.fleetBlock.Encode(), nil
	}

	return nil, ErrNoRawBlockData
}

// EncodeProductionQueueBlock encodes a ProductionQueueEntity back to block data.
// Delegates to ProductionQueueBlock.Encode().
func (e *BlockEncoder) EncodeProductionQueueBlock(pq *ProductionQueueEntity) ([]byte, error) {
	if pq.queueBlock != nil && !pq.Meta().Dirty {
		return pq.queueBlock.DecryptedData(), nil
	}

	// Create a ProductionQueueBlock from the entity items and encode it
	if pq.queueBlock != nil {
		// Update the block with current items
		pq.queueBlock.Items = make([]blocks.QueueItem, len(pq.Items))
		for i, item := range pq.Items {
			pq.queueBlock.Items[i] = blocks.QueueItem{
				ItemId:          item.ItemId,
				Count:           item.Count,
				CompletePercent: item.CompletePercent,
				ItemType:        item.ItemType,
			}
		}
		return pq.queueBlock.Encode(), nil
	}

	// No block available, create data directly
	data := make([]byte, len(pq.Items)*4)
	for i, item := range pq.Items {
		chunk1 := uint16((item.ItemId&0x3F)<<10) | uint16(item.Count&0x3FF)
		chunk2 := uint16((item.CompletePercent&0xFFF)<<4) | uint16(item.ItemType&0x0F)
		encoding.Write16(data, i*4, chunk1)
		encoding.Write16(data, i*4+2, chunk2)
	}
	return data, nil
}

// EncodeBattlePlanBlock encodes a BattlePlanEntity back to block data.
// Delegates to BattlePlanBlock.Encode().
func (e *BlockEncoder) EncodeBattlePlanBlock(bp *BattlePlanEntity) ([]byte, error) {
	if bp.battlePlanBlock != nil && !bp.Meta().Dirty {
		return bp.battlePlanBlock.DecryptedData(), nil
	}

	// If we have a block, update and encode
	if bp.battlePlanBlock != nil {
		return bp.battlePlanBlock.Encode(), nil
	}

	return nil, ErrNoRawBlockData
}

// EncodePlanetBlock encodes a PlanetEntity back to block data.
// Updates the block fields and re-encodes modified sections.
func (e *BlockEncoder) EncodePlanetBlock(planet *PlanetEntity) ([]byte, error) {
	// If the planet has the original block data and hasn't been modified, use it
	if planet.planetBlock != nil && !planet.Meta().Dirty {
		return planet.planetBlock.DecryptedData(), nil
	}

	// If the planet has been modified but has a block, update and encode
	if planet.planetBlock != nil {
		pb := planet.planetBlock

		// Sync entity values back to the block
		pb.Owner = planet.Owner
		pb.IsHomeworld = planet.IsHomeworld
		pb.DetectionLevel = planet.DetectionLevel
		pb.Include = planet.Include
		pb.HasStarbase = planet.HasStarbase
		pb.HasArtifact = planet.HasArtifact
		pb.IsTerraformed = planet.IsTerraformed
		pb.HasInstallations = planet.HasInstallations
		pb.FirstYear = planet.FirstYear

		// Environment values
		pb.IroniumConc = planet.IroniumConc
		pb.BoraniumConc = planet.BoraniumConc
		pb.GermaniumConc = planet.GermaniumConc
		pb.Gravity = planet.Gravity
		pb.Temperature = planet.Temperature
		pb.Radiation = planet.Radiation
		pb.OrigGravity = planet.OrigGravity
		pb.OrigTemperature = planet.OrigTemperature
		pb.OrigRadiation = planet.OrigRadiation

		// Surface minerals
		pb.Ironium = planet.Ironium
		pb.Boranium = planet.Boranium
		pb.Germanium = planet.Germanium
		pb.Population = planet.Population

		// Installations
		pb.Mines = planet.Mines
		pb.Factories = planet.Factories
		pb.Defenses = planet.Defenses
		pb.DeltaPop = planet.DeltaPop
		pb.ScannerID = planet.ScannerID
		pb.InstArtifact = planet.InstArtifact
		pb.NoResearch = planet.NoResearch

		// Starbase
		pb.StarbaseDesign = planet.StarbaseDesign

		// Route
		pb.RouteTarget = planet.RouteTarget

		// Encode with in-place modifications
		return e.encodePlanetBlockInPlace(pb)
	}

	return nil, ErrNoRawBlockData
}

// EncodePlanetBlockFromSource encodes a planet using the SOURCE block's structure
// with the entity's modified values. This is essential when regenerating files because
// each file (M, HST, H) has its own block structure for the same planet.
// The entity's planetBlock reference may point to a different file's block.
func (e *BlockEncoder) EncodePlanetBlockFromSource(sourceBlock *blocks.PartialPlanetBlock, planet *PlanetEntity) ([]byte, error) {
	if sourceBlock == nil || len(sourceBlock.Decrypted) < 4 {
		return nil, ErrNoRawBlockData
	}

	debugLog("EncodePlanetBlockFromSource: planet #%d, entity.Mines=%d, entity.Factories=%d, entity.Defenses=%d, entity.HasInstallations=%v\n",
		planet.PlanetNumber, planet.Mines, planet.Factories, planet.Defenses, planet.HasInstallations)
	debugLog("  sourceBlock.HasInstallations=%v, sourceBlock.Mines=%d, len(sourceBlock.Decrypted)=%d\n",
		sourceBlock.HasInstallations, sourceBlock.Mines, len(sourceBlock.Decrypted))

	// Create a temporary copy of the source block to modify
	// We don't want to mutate the original source block
	pb := *sourceBlock
	pb.Decrypted = make([]byte, len(sourceBlock.Decrypted))
	copy(pb.Decrypted, sourceBlock.Decrypted)

	// Apply entity values to the copy
	pb.Owner = planet.Owner
	pb.IsHomeworld = planet.IsHomeworld
	pb.DetectionLevel = planet.DetectionLevel
	pb.Include = planet.Include
	pb.HasStarbase = planet.HasStarbase
	pb.HasArtifact = planet.HasArtifact
	pb.IsTerraformed = planet.IsTerraformed
	pb.HasInstallations = planet.HasInstallations
	pb.FirstYear = planet.FirstYear

	// Environment values
	pb.IroniumConc = planet.IroniumConc
	pb.BoraniumConc = planet.BoraniumConc
	pb.GermaniumConc = planet.GermaniumConc
	pb.Gravity = planet.Gravity
	pb.Temperature = planet.Temperature
	pb.Radiation = planet.Radiation
	pb.OrigGravity = planet.OrigGravity
	pb.OrigTemperature = planet.OrigTemperature
	pb.OrigRadiation = planet.OrigRadiation

	// Surface minerals
	pb.Ironium = planet.Ironium
	pb.Boranium = planet.Boranium
	pb.Germanium = planet.Germanium
	pb.Population = planet.Population / 100 // Convert back to file units (100s of colonists)

	// Installations
	pb.Mines = planet.Mines
	pb.Factories = planet.Factories
	pb.Defenses = planet.Defenses
	pb.DeltaPop = planet.DeltaPop
	pb.ScannerID = planet.ScannerID
	pb.InstArtifact = planet.InstArtifact
	pb.NoResearch = planet.NoResearch

	// Starbase
	pb.StarbaseDesign = planet.StarbaseDesign

	// Route
	pb.RouteTarget = planet.RouteTarget

	debugLog("  after copy: pb.Mines=%d, pb.Factories=%d, pb.Defenses=%d, pb.HasInstallations=%v\n",
		pb.Mines, pb.Factories, pb.Defenses, pb.HasInstallations)

	return e.encodePlanetBlockInPlace(&pb)
}

// encodePlanetBlockInPlace modifies the decrypted block data in-place with updated values.
// This handles the complex variable-length planet block format.
func (e *BlockEncoder) encodePlanetBlockInPlace(pb *blocks.PartialPlanetBlock) ([]byte, error) {
	if len(pb.Decrypted) < 4 {
		return nil, ErrNoRawBlockData
	}

	// Make a copy of the decrypted data to modify
	data := make([]byte, len(pb.Decrypted))
	copy(data, pb.Decrypted)

	// Update header (bytes 0-3)
	// Bytes 0-1: Planet number and owner
	data[0] = byte(pb.PlanetNumber & 0xFF)
	ownerBits := 31 // No owner
	if pb.Owner >= 0 && pb.Owner <= 15 {
		ownerBits = pb.Owner
	}
	data[1] = byte((pb.PlanetNumber>>8)&0x07) | byte(ownerBits<<3)

	// Bytes 2-3: Flags word
	var flags uint16
	flags = uint16(pb.DetectionLevel & 0x7F)
	if pb.IsHomeworld {
		flags |= 0x0080
	}
	if pb.Include {
		flags |= 0x0100
	}
	if pb.HasStarbase {
		flags |= 0x0200
	}
	if pb.IsTerraformed {
		flags |= 0x0400
	}
	if pb.HasInstallations {
		flags |= 0x0800
	}
	if pb.HasArtifact {
		flags |= 0x1000
	}
	if pb.HasSurfaceMinerals {
		flags |= 0x2000
	}
	if pb.HasRoute {
		flags |= 0x4000
	}
	if pb.FirstYear {
		flags |= 0x8000
	}
	encoding.Write16(data, 2, flags)

	// Walk through the variable sections to find offsets
	index := 4

	// Environment section
	if pb.CanSeeEnvironment() && index < len(data) {
		// Skip the pre-environment length byte(s)
		preEnvLengthByte := int(data[index] & 0xFF)
		preEnvLength := 1
		preEnvLength += preEnvLengthByte & 0x03
		preEnvLength += (preEnvLengthByte & 0x0C) >> 2
		preEnvLength += (preEnvLengthByte & 0x30) >> 4
		index += preEnvLength

		// Update mineral concentrations (3 bytes)
		if index+3 <= len(data) {
			data[index] = byte(pb.IroniumConc & 0xFF)
			data[index+1] = byte(pb.BoraniumConc & 0xFF)
			data[index+2] = byte(pb.GermaniumConc & 0xFF)
			index += 3
		}

		// Update gravity, temperature, radiation (3 bytes)
		if index+3 <= len(data) {
			data[index] = byte(pb.Gravity & 0xFF)
			data[index+1] = byte(pb.Temperature & 0xFF)
			data[index+2] = byte(pb.Radiation & 0xFF)
			index += 3
		}

		// Update original values if terraformed
		if pb.IsTerraformed && index+3 <= len(data) {
			data[index] = byte(pb.OrigGravity & 0xFF)
			data[index+1] = byte(pb.OrigTemperature & 0xFF)
			data[index+2] = byte(pb.OrigRadiation & 0xFF)
			index += 3
		}

		// Skip estimates if owned
		if pb.Owner >= 0 && index+2 <= len(data) {
			index += 2
		}
	}

	// Surface minerals section (variable length)
	// We need to properly encode values with the right byte lengths
	if pb.HasSurfaceMinerals && index < len(data) {
		// Calculate old byte lengths from original contents byte
		oldContentsLengths := data[index]
		oldIronLen := encoding.VarLenByteCount(encoding.ExtractVarLenField(oldContentsLengths, 0))
		oldBoraLen := encoding.VarLenByteCount(encoding.ExtractVarLenField(oldContentsLengths, 2))
		oldGermLen := encoding.VarLenByteCount(encoding.ExtractVarLenField(oldContentsLengths, 4))
		oldPopLen := encoding.VarLenByteCount(encoding.ExtractVarLenField(oldContentsLengths, 6))
		oldTotalLen := 1 + oldIronLen + oldBoraLen + oldGermLen + oldPopLen // 1 for contents byte

		// Calculate new required byte lengths
		newIronEnc := encoding.ByteLengthForInt(pb.Ironium)
		newBoraEnc := encoding.ByteLengthForInt(pb.Boranium)
		newGermEnc := encoding.ByteLengthForInt(pb.Germanium)
		newPopEnc := encoding.ByteLengthForInt(pb.Population)
		newIronLen := encoding.VarLenByteCount(newIronEnc)
		newBoraLen := encoding.VarLenByteCount(newBoraEnc)
		newGermLen := encoding.VarLenByteCount(newGermEnc)
		newPopLen := encoding.VarLenByteCount(newPopEnc)
		newTotalLen := 1 + newIronLen + newBoraLen + newGermLen + newPopLen

		// Build new contents byte
		newContentsLengths := byte(newIronEnc | (newBoraEnc << 2) | (newGermEnc << 4) | (newPopEnc << 6))

		// Check if size changed
		sizeDelta := newTotalLen - oldTotalLen
		if sizeDelta != 0 {
			// Need to rebuild the data array with new size
			surfaceMineralsStart := index
			afterSurfaceMinerals := index + oldTotalLen
			restOfData := data[afterSurfaceMinerals:]

			newData := make([]byte, len(data)+sizeDelta)
			// Copy header and environment sections
			copy(newData[:surfaceMineralsStart], data[:surfaceMineralsStart])
			// We'll write the new surface minerals section below
			// Copy the rest (installations, starbase, route)
			copy(newData[surfaceMineralsStart+newTotalLen:], restOfData)
			data = newData
		}

		// Write surface minerals section
		data[index] = newContentsLengths
		index++
		index = encoding.WriteVarLen(data, index, pb.Ironium)
		index = encoding.WriteVarLen(data, index, pb.Boranium)
		index = encoding.WriteVarLen(data, index, pb.Germanium)
		index = encoding.WriteVarLen(data, index, pb.Population)
	}

	// Update installations section (8 bytes, fixed)
	// If data array is too small but entity has installations, expand it
	debugLog("  installations: HasInstallations=%v, index=%d, len(data)=%d\n", pb.HasInstallations, index, len(data))
	debugLog("  installations values: Mines=%d, Factories=%d, Defenses=%d, ScannerID=%d\n", pb.Mines, pb.Factories, pb.Defenses, pb.ScannerID)
	if pb.HasInstallations {
		if index+8 > len(data) {
			debugLog("  expanding data array from %d to %d\n", len(data), index+8)
			// Need to expand array for installations
			newData := make([]byte, index+8)
			copy(newData, data)
			data = newData
		}
		// First dword: population change + mines + factories
		dword1 := uint32(pb.DeltaPop&0xFF) |
			uint32((pb.Mines&0xFFF)<<8) |
			uint32((pb.Factories&0xFFF)<<20)
		encoding.Write32(data, index, dword1)
		debugLog("  wrote dword1=0x%08X at index %d\n", dword1, index)

		// Second dword: defenses + scanner + flags
		dword2 := uint32(pb.Defenses&0xFFF) |
			uint32((pb.ScannerID&0x1F)<<12)
		if pb.InstArtifact {
			dword2 |= 1 << 22
		}
		if pb.NoResearch {
			dword2 |= 1 << 23
		}
		encoding.Write32(data, index+4, dword2)
		debugLog("  wrote dword2=0x%08X at index %d\n", dword2, index+4)
		index += 8
	}

	// Starbase and route sections are preserved from original data

	debugLog("  final: len(data)=%d, index=%d\n", len(data), index)
	return data, nil
}

// EncodePlayerBlock encodes a PlayerEntity back to block data.
// Updates the block fields and re-encodes.
func (e *BlockEncoder) EncodePlayerBlock(player *PlayerEntity) ([]byte, error) {
	// If the player has the original block data and hasn't been modified, use it
	if player.playerBlock != nil && !player.Meta().Dirty {
		return player.playerBlock.DecryptedData(), nil
	}

	// If the player has been modified but has a block, encode it
	if player.playerBlock != nil {
		// The playerBlock fields are updated in-place by ChangeToAI/ChangeToHuman/etc
		// so we can just encode it directly
		return player.playerBlock.Encode()
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
