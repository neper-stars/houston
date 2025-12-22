package blocks

import (
	"github.com/neper-stars/houston/encoding"
)

// Battle tactics
const (
	TacticDisengage           = 0
	TacticDisengageIfChallenged = 1
	TacticMinimizeDamage      = 2
	TacticMaximizeNetDamage   = 3
	TacticMaximizeDamageRatio = 4
	TacticMaximizeDamage      = 5
)

// Battle target types
const (
	TargetNone        = 0
	TargetAny         = 1
	TargetStarbase    = 2
	TargetArmedShips  = 3
	TargetBombers     = 4 // Also freighters
	TargetUnarmedShips = 5
	TargetFuelTransports = 6
	TargetFreighters  = 7
)

// Attack who values
const (
	AttackNobody          = 0
	AttackEnemies         = 1
	AttackNeutralAndEnemies = 2
	AttackEveryone        = 3
	AttackPlayerBase      = 4 // Player ID = value - 4
)

// BattlePlanBlock represents a fleet battle plan configuration (Type 30)
type BattlePlanBlock struct {
	GenericBlock

	OwnerPlayerId   int    // Owner player (0-15)
	PlanId          int    // Plan ID (0-15)
	Tactic          int    // Battle tactic (0-5)
	DumpCargo       bool   // Dump cargo before battle
	PrimaryTarget   int    // Primary target type (0-7)
	SecondaryTarget int    // Secondary target type (0-7)
	AttackWho       int    // Attack policy (0-19)
	Name            string // Plan name
	Deleted         bool   // True if this plan was deleted
}

// NewBattlePlanBlock creates a BattlePlanBlock from a GenericBlock
func NewBattlePlanBlock(b GenericBlock) *BattlePlanBlock {
	bpb := &BattlePlanBlock{
		GenericBlock: b,
	}
	bpb.decode()
	return bpb
}

func (bpb *BattlePlanBlock) decode() {
	data := bpb.Decrypted
	if len(data) < 4 {
		return
	}

	word0 := encoding.Read16(data, 0)
	bpb.OwnerPlayerId = int(word0 & 0x0F)          // Bits 0-3
	bpb.PlanId = int((word0 >> 4) & 0x0F)          // Bits 4-7
	bpb.Tactic = int((word0 >> 8) & 0x0F)          // Bits 8-11
	bpb.DumpCargo = (word0 & 0x8000) != 0          // Bit 15

	word1 := encoding.Read16(data, 2)
	bpb.PrimaryTarget = int(word1 & 0x0F)          // Bits 0-3
	bpb.SecondaryTarget = int((word1 >> 4) & 0x0F) // Bits 4-7
	bpb.AttackWho = int(word1 >> 8)                // Bits 8-15

	// If exactly 4 bytes, the plan was deleted
	if len(data) == 4 {
		bpb.Deleted = true
		return
	}

	// Decode the plan name
	if len(data) > 4 {
		nameBytes := data[4:]
		decoded, err := encoding.DecodeStarsString(nameBytes)
		if err == nil {
			bpb.Name = decoded
		}
	}
}

// TacticName returns a human-readable name for the tactic
func (bpb *BattlePlanBlock) TacticName() string {
	switch bpb.Tactic {
	case TacticDisengage:
		return "Disengage"
	case TacticDisengageIfChallenged:
		return "Disengage if Challenged"
	case TacticMinimizeDamage:
		return "Minimize Damage"
	case TacticMaximizeNetDamage:
		return "Maximize Net Damage"
	case TacticMaximizeDamageRatio:
		return "Maximize Damage Ratio"
	case TacticMaximizeDamage:
		return "Maximize Damage"
	default:
		return "Unknown"
	}
}

// TargetPlayer returns the specific player ID if AttackWho >= AttackPlayerBase
func (bpb *BattlePlanBlock) TargetPlayer() int {
	if bpb.AttackWho >= AttackPlayerBase {
		return bpb.AttackWho - AttackPlayerBase
	}
	return -1
}

// PrimaryTargetName returns a human-readable name for the primary target
func (bpb *BattlePlanBlock) PrimaryTargetName() string {
	return targetName(bpb.PrimaryTarget)
}

// SecondaryTargetName returns a human-readable name for the secondary target
func (bpb *BattlePlanBlock) SecondaryTargetName() string {
	return targetName(bpb.SecondaryTarget)
}

// targetName returns a human-readable name for a target type
func targetName(t int) string {
	switch t {
	case TargetNone:
		return "None/Disengage"
	case TargetAny:
		return "Any"
	case TargetStarbase:
		return "Starbase"
	case TargetArmedShips:
		return "Armed Ships"
	case TargetBombers:
		return "Bombers/Freighters"
	case TargetUnarmedShips:
		return "Unarmed Ships"
	case TargetFuelTransports:
		return "Fuel Transports"
	case TargetFreighters:
		return "Freighters"
	default:
		return "Unknown"
	}
}

// AttackWhoName returns a human-readable name for the attack policy
func (bpb *BattlePlanBlock) AttackWhoName() string {
	switch bpb.AttackWho {
	case AttackNobody:
		return "Nobody"
	case AttackEnemies:
		return "Enemies"
	case AttackNeutralAndEnemies:
		return "Neutrals & Enemies"
	case AttackEveryone:
		return "Everyone"
	default:
		if bpb.AttackWho >= AttackPlayerBase {
			return "Specific Player"
		}
		return "Unknown"
	}
}

// BattleBlock represents a battle record (Type 31)
// Structure not fully documented - preserves raw data for analysis
type BattleBlock struct {
	GenericBlock
}

// NewBattleBlock creates a BattleBlock from a GenericBlock
func NewBattleBlock(b GenericBlock) *BattleBlock {
	return &BattleBlock{GenericBlock: b}
}

// BattleContinuationBlock represents battle continuation data (Type 39)
// Structure not fully documented - preserves raw data for analysis
type BattleContinuationBlock struct {
	GenericBlock
}

// NewBattleContinuationBlock creates a BattleContinuationBlock from a GenericBlock
func NewBattleContinuationBlock(b GenericBlock) *BattleContinuationBlock {
	return &BattleContinuationBlock{GenericBlock: b}
}

// SetFleetBattlePlanBlock represents setting a fleet's battle plan (Type 42)
// Structure not fully documented - preserves raw data for analysis
type SetFleetBattlePlanBlock struct {
	GenericBlock
}

// NewSetFleetBattlePlanBlock creates a SetFleetBattlePlanBlock from a GenericBlock
func NewSetFleetBattlePlanBlock(b GenericBlock) *SetFleetBattlePlanBlock {
	return &SetFleetBattlePlanBlock{GenericBlock: b}
}
