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
