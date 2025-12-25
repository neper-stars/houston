package store

import "github.com/neper-stars/houston/blocks"

// ProductionItem represents a single item in a production queue.
type ProductionItem struct {
	ItemId          int // Item type ID (0-63)
	Count           int // Quantity to build (0-1023)
	CompletePercent int // Completion percentage (0-4095)
	ItemType        int // 2=standard, 4=custom design
}

// IsAutoItem returns true if this is an auto-build item.
func (pi *ProductionItem) IsAutoItem() bool {
	return pi.ItemType == blocks.ProductionItemTypeStandard && pi.ItemId <= 6
}

// IsShipDesign returns true if this is a custom ship/starbase design.
func (pi *ProductionItem) IsShipDesign() bool {
	return pi.ItemType == blocks.ProductionItemTypeCustom
}

// ProductionQueueEntity represents a planet's production queue.
type ProductionQueueEntity struct {
	meta EntityMeta

	// Identification
	PlanetNumber int // Associated planet

	// Queue contents
	Items []ProductionItem

	// Raw block (preserved for re-encoding)
	queueBlock *blocks.ProductionQueueBlock
}

// Meta returns the entity metadata.
func (pq *ProductionQueueEntity) Meta() *EntityMeta {
	return &pq.meta
}

// RawBlocks returns the original blocks.
func (pq *ProductionQueueEntity) RawBlocks() []blocks.Block {
	if pq.queueBlock != nil {
		return []blocks.Block{*pq.queueBlock}
	}
	return nil
}

// SetDirty marks the entity as modified.
func (pq *ProductionQueueEntity) SetDirty() {
	pq.meta.Dirty = true
}

// QueueLength returns the number of items in the queue.
func (pq *ProductionQueueEntity) QueueLength() int {
	return len(pq.Items)
}

// GetItem returns the queue item at the given index, or nil if out of range.
func (pq *ProductionQueueEntity) GetItem(index int) *ProductionItem {
	if index >= 0 && index < len(pq.Items) {
		return &pq.Items[index]
	}
	return nil
}

// AddItem adds an item to the queue (marks dirty).
func (pq *ProductionQueueEntity) AddItem(itemId, count, itemType int) {
	pq.Items = append(pq.Items, ProductionItem{
		ItemId:   itemId,
		Count:    count,
		ItemType: itemType,
	})
	pq.SetDirty()
}

// Clear removes all items from the queue (marks dirty).
func (pq *ProductionQueueEntity) Clear() {
	pq.Items = nil
	pq.SetDirty()
}

// newProductionQueueEntityFromBlock creates a ProductionQueueEntity from a ProductionQueueBlock.
func newProductionQueueEntityFromBlock(pqb *blocks.ProductionQueueBlock, planetNumber int, source *FileSource) *ProductionQueueEntity {
	items := make([]ProductionItem, len(pqb.Items))
	for i, item := range pqb.Items {
		items[i] = ProductionItem{
			ItemId:          item.ItemId,
			Count:           item.Count,
			CompletePercent: item.CompletePercent,
			ItemType:        item.ItemType,
		}
	}

	entity := &ProductionQueueEntity{
		meta: EntityMeta{
			Key: EntityKey{
				Type:   EntityTypeProductionQueue,
				Owner:  -1, // Production queues are keyed by planet, not owner
				Number: planetNumber,
			},
			BestSource: source,
			Quality:    QualityFull,
			Turn:       source.Turn,
		},
		PlanetNumber: planetNumber,
		Items:        items,
		queueBlock:   pqb,
	}
	entity.meta.AddSource(source)
	return entity
}
