package store

import "github.com/neper-stars/houston/blocks"

// Object type constants
const (
	ObjectTypeMinefield = 0
	ObjectTypePacket    = 1
	ObjectTypeWormhole  = 2
	ObjectTypeTrader    = 3
)

// ObjectEntity represents a map object (minefield, packet, wormhole, trader, salvage).
type ObjectEntity struct {
	meta EntityMeta

	// Core identification
	Number     int
	Owner      int // -1 for unowned
	ObjectType int // 0=minefield, 1=packet, 2=wormhole, 3=trader

	// Position
	X, Y int

	// Minefield-specific fields
	MineCount     int64
	MinefieldType int // 0=standard, 1=heavy, 2=speed bump
	Detonating    bool

	// Wormhole-specific fields
	WormholeId      int
	TargetId        int
	BeenThroughBits uint16
	CanSeeBits      uint16
	Stability       int

	// Mystery Trader-specific fields
	XDest    int
	YDest    int
	Warp     int
	MetBits  uint16
	ItemBits uint16
	TurnNo   int

	// Packet-specific fields
	DestinationPlanetID int
	Ironium             int
	Boranium            int
	Germanium           int
	PacketSpeed         int

	// Salvage-specific fields
	IsSalvage     bool
	SourceFleetID int

	// Raw block (preserved for re-encoding)
	objectBlock *blocks.ObjectBlock
}

// Meta returns the entity metadata.
func (o *ObjectEntity) Meta() *EntityMeta {
	return &o.meta
}

// RawBlocks returns the original blocks.
func (o *ObjectEntity) RawBlocks() []blocks.Block {
	if o.objectBlock != nil {
		return []blocks.Block{*o.objectBlock}
	}
	return nil
}

// SetDirty marks the entity as modified.
func (o *ObjectEntity) SetDirty() {
	o.meta.Dirty = true
}

// IsMinefield returns true if this is a minefield.
func (o *ObjectEntity) IsMinefield() bool {
	return o.ObjectType == ObjectTypeMinefield
}

// IsPacket returns true if this is a mineral packet (not salvage).
func (o *ObjectEntity) IsPacket() bool {
	return o.ObjectType == ObjectTypePacket && !o.IsSalvage
}

// IsWormhole returns true if this is a wormhole.
func (o *ObjectEntity) IsWormhole() bool {
	return o.ObjectType == ObjectTypeWormhole
}

// IsTrader returns true if this is a mystery trader.
func (o *ObjectEntity) IsTrader() bool {
	return o.ObjectType == ObjectTypeTrader
}

// GetCargo returns the cargo for packets/salvage as a Cargo struct.
func (o *ObjectEntity) GetCargo() Cargo {
	return Cargo{
		Ironium:   int64(o.Ironium),
		Boranium:  int64(o.Boranium),
		Germanium: int64(o.Germanium),
	}
}

// newObjectEntityFromBlock creates an ObjectEntity from an ObjectBlock.
func newObjectEntityFromBlock(ob *blocks.ObjectBlock, source *FileSource) *ObjectEntity {
	// Skip count objects
	if ob.IsCountObject {
		return nil
	}

	entity := &ObjectEntity{
		meta: EntityMeta{
			Key: EntityKey{
				Type:   EntityTypeObject,
				Owner:  ob.Owner,
				Number: ob.Number,
			},
			BestSource: source,
			Quality:    QualityFull,
			Turn:       source.Turn,
		},
		Number:              ob.Number,
		Owner:               ob.Owner,
		ObjectType:          ob.ObjectType,
		X:                   ob.X,
		Y:                   ob.Y,
		MineCount:           ob.MineCount,
		MinefieldType:       ob.MinefieldType,
		Detonating:          ob.Detonating,
		WormholeId:          ob.WormholeId,
		TargetId:            ob.TargetId,
		BeenThroughBits:     ob.BeenThroughBits,
		CanSeeBits:          ob.CanSeeBits,
		Stability:           ob.Stability,
		XDest:               ob.XDest,
		YDest:               ob.YDest,
		Warp:                ob.Warp,
		MetBits:             ob.MetBits,
		ItemBits:            ob.ItemBits,
		TurnNo:              ob.TurnNo,
		DestinationPlanetID: ob.DestinationPlanetID,
		Ironium:             ob.Ironium,
		Boranium:            ob.Boranium,
		Germanium:           ob.Germanium,
		PacketSpeed:         ob.PacketSpeed,
		IsSalvage:           ob.IsSalvageObject,
		SourceFleetID:       ob.SourceFleetID,
		objectBlock:         ob,
	}
	entity.meta.AddSource(source)
	return entity
}
