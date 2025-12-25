package store

import "github.com/neper-stars/houston/blocks"

// BattlePlanEntity represents a battle plan.
type BattlePlanEntity struct {
	meta EntityMeta

	// Identification
	Owner  int // Player index (0-15)
	PlanId int // Plan ID (0-15)

	// Plan settings
	Name            string
	Tactic          int  // Battle tactic (0-5)
	DumpCargo       bool // Dump cargo before battle
	PrimaryTarget   int  // Primary target type (0-7)
	SecondaryTarget int  // Secondary target type (0-7)
	AttackWho       int  // Attack policy (0-19)
	Deleted         bool

	// Raw block (preserved for re-encoding)
	battlePlanBlock *blocks.BattlePlanBlock
}

// Meta returns the entity metadata.
func (bp *BattlePlanEntity) Meta() *EntityMeta {
	return &bp.meta
}

// RawBlocks returns the original blocks.
func (bp *BattlePlanEntity) RawBlocks() []blocks.Block {
	if bp.battlePlanBlock != nil {
		return []blocks.Block{*bp.battlePlanBlock}
	}
	return nil
}

// SetDirty marks the entity as modified.
func (bp *BattlePlanEntity) SetDirty() {
	bp.meta.Dirty = true
}

// newBattlePlanEntityFromBlock creates a BattlePlanEntity from a BattlePlanBlock.
func newBattlePlanEntityFromBlock(bpb *blocks.BattlePlanBlock, source *FileSource) *BattlePlanEntity {
	entity := &BattlePlanEntity{
		meta: EntityMeta{
			Key: EntityKey{
				Type:   EntityTypeBattlePlan,
				Owner:  bpb.OwnerPlayerId,
				Number: bpb.PlanId,
			},
			BestSource: source,
			Quality:    QualityFull,
			Turn:       source.Turn,
		},
		Owner:           bpb.OwnerPlayerId,
		PlanId:          bpb.PlanId,
		Name:            bpb.Name,
		Tactic:          bpb.Tactic,
		DumpCargo:       bpb.DumpCargo,
		PrimaryTarget:   bpb.PrimaryTarget,
		SecondaryTarget: bpb.SecondaryTarget,
		AttackWho:       bpb.AttackWho,
		Deleted:         bpb.Deleted,
		battlePlanBlock: bpb,
	}
	entity.meta.AddSource(source)
	return entity
}
