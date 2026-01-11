package store

import (
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

	// Skip surface minerals section (variable length - preserve original encoding)
	// Note: Changing population/surface minerals to different byte lengths would require
	// rebuilding the entire block, which is complex. For now, we preserve the original
	// encoding and only update fixed sections.
	if pb.HasSurfaceMinerals && index < len(data) {
		contentsLengths := data[index]
		index++
		ironLen := encoding.VarLenByteCount(encoding.ExtractVarLenField(contentsLengths, 0))
		boraLen := encoding.VarLenByteCount(encoding.ExtractVarLenField(contentsLengths, 2))
		germLen := encoding.VarLenByteCount(encoding.ExtractVarLenField(contentsLengths, 4))
		popLen := encoding.VarLenByteCount(encoding.ExtractVarLenField(contentsLengths, 6))

		// Update surface minerals in-place using the same byte lengths as original
		if index+ironLen <= len(data) {
			encoding.WriteVarLenFixedSize(data, index, pb.Ironium, ironLen)
			index += ironLen
		}
		if index+boraLen <= len(data) {
			encoding.WriteVarLenFixedSize(data, index, pb.Boranium, boraLen)
			index += boraLen
		}
		if index+germLen <= len(data) {
			encoding.WriteVarLenFixedSize(data, index, pb.Germanium, germLen)
			index += germLen
		}
		if index+popLen <= len(data) {
			encoding.WriteVarLenFixedSize(data, index, pb.Population, popLen)
			index += popLen
		}
	}

	// Update installations section (8 bytes, fixed)
	if pb.HasInstallations && index+8 <= len(data) {
		// First dword: population change + mines + factories
		dword1 := uint32(pb.DeltaPop&0xFF) |
			uint32((pb.Mines&0xFFF)<<8) |
			uint32((pb.Factories&0xFFF)<<20)
		encoding.Write32(data, index, dword1)

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
		// Note: index not incremented as it's not used after this point
	}

	// Starbase and route sections are preserved from original data

	return data, nil
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
