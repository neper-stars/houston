package blocks

import (
	"github.com/neper-stars/houston/encoding"
)

// Production queue item types
const (
	ProductionItemTypeStandard = 2 // Standard game items
	ProductionItemTypeCustom   = 4 // Custom ship/starbase designs
)

// Standard production item IDs (when ItemType == ProductionItemTypeStandard)
const (
	ProductionItemAutoMines        = 0
	ProductionItemAutoFactories    = 1
	ProductionItemAutoDefenses     = 2
	ProductionItemAutoAlchemy      = 3
	ProductionItemAutoMinTerraform = 4
	ProductionItemAutoMaxTerraform = 5
	ProductionItemAutoPackets      = 6
	ProductionItemFactory          = 7
	ProductionItemMine             = 8
	ProductionItemDefense          = 9
	ProductionItemMineralAlchemy   = 11
	ProductionItemPacketIronium    = 14
	ProductionItemPacketBoranium   = 15
	ProductionItemPacketGermanium  = 16
	ProductionItemPacketMixed      = 17
	ProductionItemScanner          = 27
)

// QueueItem represents a single item in the production queue
type QueueItem struct {
	ItemId          int // Item type ID (0-63)
	Count           int // Quantity to build (0-1023)
	CompletePercent int // Completion percentage (0-4095)
	ItemType        int // 2=standard, 4=custom design
}

// IsAutoItem returns true if this is an auto-build item
func (qi *QueueItem) IsAutoItem() bool {
	return qi.ItemType == ProductionItemTypeStandard && qi.ItemId <= 6
}

// IsShipDesign returns true if this is a custom ship/starbase design
func (qi *QueueItem) IsShipDesign() bool {
	return qi.ItemType == ProductionItemTypeCustom
}

// ProductionQueueBlock represents a planet's production queue (Type 28)
type ProductionQueueBlock struct {
	GenericBlock

	Items []QueueItem
}

// NewProductionQueueBlock creates a ProductionQueueBlock from a GenericBlock
func NewProductionQueueBlock(b GenericBlock) *ProductionQueueBlock {
	pqb := &ProductionQueueBlock{
		GenericBlock: b,
		Items:        make([]QueueItem, 0),
	}
	pqb.decode()
	return pqb
}

func (pqb *ProductionQueueBlock) decode() {
	data := pqb.Decrypted

	// Each queue item is 4 bytes
	for i := 0; i+4 <= len(data); i += 4 {
		chunk1 := encoding.Read16(data, i)
		chunk2 := encoding.Read16(data, i+2)

		item := QueueItem{
			ItemId:          int(chunk1 >> 10),   // Top 6 bits
			Count:           int(chunk1 & 0x3FF), // Bottom 10 bits
			CompletePercent: int(chunk2 >> 4),    // Top 12 bits
			ItemType:        int(chunk2 & 0x0F),  // Bottom 4 bits
		}

		pqb.Items = append(pqb.Items, item)
	}
}

// QueueLength returns the number of items in the queue
func (pqb *ProductionQueueBlock) QueueLength() int {
	return len(pqb.Items)
}

// GetItem returns the queue item at the given index, or nil if out of range
func (pqb *ProductionQueueBlock) GetItem(index int) *QueueItem {
	if index >= 0 && index < len(pqb.Items) {
		return &pqb.Items[index]
	}
	return nil
}

// Encode returns the raw block data bytes (without the 2-byte block header).
// Each queue item is packed into 4 bytes.
func (pqb *ProductionQueueBlock) Encode() []byte {
	data := make([]byte, len(pqb.Items)*4)
	for i, item := range pqb.Items {
		// Pack into 4 bytes:
		// chunk1: ItemId (6 bits) << 10 | Count (10 bits)
		// chunk2: CompletePercent (12 bits) << 4 | ItemType (4 bits)
		chunk1 := uint16((item.ItemId&0x3F)<<10) | uint16(item.Count&0x3FF)
		chunk2 := uint16((item.CompletePercent&0xFFF)<<4) | uint16(item.ItemType&0x0F)

		encoding.Write16(data, i*4, chunk1)
		encoding.Write16(data, i*4+2, chunk2)
	}
	return data
}

// ProductionQueueChangeBlock represents a production queue modification (Type 29)
// It includes the planet ID and the new queue contents
type ProductionQueueChangeBlock struct {
	GenericBlock

	PlanetId int         // Planet ID (11 bits, 0-2047)
	Items    []QueueItem // Queue items (same format as ProductionQueueBlock)
}

// NewProductionQueueChangeBlock creates a ProductionQueueChangeBlock from a GenericBlock
func NewProductionQueueChangeBlock(b GenericBlock) *ProductionQueueChangeBlock {
	pqcb := &ProductionQueueChangeBlock{
		GenericBlock: b,
		Items:        make([]QueueItem, 0),
	}
	pqcb.decode()
	return pqcb
}

func (pqcb *ProductionQueueChangeBlock) decode() {
	data := pqcb.Decrypted
	if len(data) < 2 {
		return
	}

	// First 2 bytes: Planet ID (11 bits)
	pqcb.PlanetId = int(encoding.Read16(data, 0) & 0x7FF)

	// Remaining bytes are queue items (4 bytes each)
	for i := 2; i+4 <= len(data); i += 4 {
		chunk1 := encoding.Read16(data, i)
		chunk2 := encoding.Read16(data, i+2)

		item := QueueItem{
			ItemId:          int(chunk1 >> 10),   // Top 6 bits
			Count:           int(chunk1 & 0x3FF), // Bottom 10 bits
			CompletePercent: int(chunk2 >> 4),    // Top 12 bits
			ItemType:        int(chunk2 & 0x0F),  // Bottom 4 bits
		}

		pqcb.Items = append(pqcb.Items, item)
	}
}

// QueueLength returns the number of items in the queue
func (pqcb *ProductionQueueChangeBlock) QueueLength() int {
	return len(pqcb.Items)
}

// GetItem returns the queue item at the given index, or nil if out of range
func (pqcb *ProductionQueueChangeBlock) GetItem(index int) *QueueItem {
	if index >= 0 && index < len(pqcb.Items) {
		return &pqcb.Items[index]
	}
	return nil
}

// Encode returns the raw block data bytes (without the 2-byte block header).
// First 2 bytes are the planet ID, then 4 bytes per queue item.
func (pqcb *ProductionQueueChangeBlock) Encode() []byte {
	data := make([]byte, 2+len(pqcb.Items)*4)

	// Planet ID (11 bits)
	encoding.Write16(data, 0, uint16(pqcb.PlanetId&0x7FF))

	// Queue items
	for i, item := range pqcb.Items {
		offset := 2 + i*4
		chunk1 := uint16((item.ItemId&0x3F)<<10) | uint16(item.Count&0x3FF)
		chunk2 := uint16((item.CompletePercent&0xFFF)<<4) | uint16(item.ItemType&0x0F)

		encoding.Write16(data, offset, chunk1)
		encoding.Write16(data, offset+2, chunk2)
	}
	return data
}
