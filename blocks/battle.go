package blocks

import (
	"fmt"

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

// BattleStack represents a "stack" in a Stars! battle.
// A stack is a group of ships of the same design that move and fight as a unit.
// When a fleet enters battle, it splits into stacks - one per ship design.
// Example: A fleet with 3 Scouts and 5 Destroyers becomes 2 stacks.
type BattleStack struct {
	DesignID  int // Ship design ID (references the player's ship design list)
	ShipCount int // Number of ships in this stack
	OwnerID   int // Owner: 0 = Side 1, 1 = Side 2 (derived from position in stack list)
}

// BattleActionType represents the type of battle action
type BattleActionType int

const (
	ActionMove   BattleActionType = iota // Stack movement
	ActionFire                           // Stack fires weapons
	ActionDamage                         // Stack takes damage
)

// BattleRecordEvent represents a decoded battle event within an action record.
// Each 22-byte action record can contain multiple events.
// Note: This is distinct from BattleEvent in misc.go which represents battle
// notifications in the events block.
type BattleRecordEvent struct {
	Round      int              // Round number (0-15)
	StackID    int              // Acting stack (0-5)
	ActionType BattleActionType // Type of action
	TargetID   int              // Target stack for fire actions
	GridX      int              // Battle grid X position (0-9)
	GridY      int              // Battle grid Y position (0-9)
	Damage     int              // Damage amount for damage events
}

// String returns a human-readable description of the event
func (e BattleRecordEvent) String() string {
	switch e.ActionType {
	case ActionMove:
		return fmt.Sprintf("Round %2d: Stack %d moves to (%d,%d)", e.Round, e.StackID, e.GridX, e.GridY)
	case ActionFire:
		return fmt.Sprintf("Round %2d: Stack %d fires", e.Round, e.StackID)
	case ActionDamage:
		return fmt.Sprintf("Round %2d: Stack %d takes %d damage at (%d,%d)", e.Round, e.StackID, e.Damage, e.GridX, e.GridY)
	default:
		return fmt.Sprintf("Round %2d: Stack %d: unknown action", e.Round, e.StackID)
	}
}

// BattleAction represents a single action record in the battle (22 bytes).
// Each record can contain multiple events (movement, firing, damage).
type BattleAction struct {
	RawData []byte              // Raw 22-byte action data
	Events  []BattleRecordEvent // Decoded events from this action
}

// BattlePhase represents one stack's turn in the battle, matching the
// "Phase X of Y" display in the Stars! Battle VCR viewer.
// Each phase shows a single stack acting: moving, firing, and/or dealing damage.
type BattlePhase struct {
	PhaseNum int // Phase number (1-based, as shown in Battle VCR)
	Round    int // Round number (0-15)
	StackID  int // Acting stack ID (0-5)
	GridX    int // Stack's grid X position (0-9)
	GridY    int // Stack's grid Y position (0-9)
	Fired    bool // True if stack fired weapons this phase
	TargetID int  // Target stack ID if fired (-1 if no target)
	Damage   int  // Total damage dealt this phase
}

// String returns a human-readable description of the phase
func (p BattlePhase) String() string {
	action := "waits"
	if p.Fired {
		if p.Damage > 0 {
			action = fmt.Sprintf("fires at Stack %d for %d damage", p.TargetID, p.Damage)
		} else {
			action = "fires (no damage)"
		}
	}
	return fmt.Sprintf("Phase %d (Round %d): Stack %d at (%d,%d) %s",
		p.PhaseNum, p.Round, p.StackID, p.GridX, p.GridY, action)
}

// BattleBlock represents a battle record (Type 31)
// Structure:
//   - Header: 18 bytes
//   - Stack definitions: 29 bytes each
//   - Action records: 22 bytes each
type BattleBlock struct {
	GenericBlock

	// Header fields (18 bytes)
	BattleID        int // Battle identifier
	Rounds          int // Number of battle rounds (stored as rounds-1)
	Side1Stacks     int // Number of stacks on side 1 (viewer's side)
	TotalStacks     int // Total number of stacks in battle
	RecordedSize    int // Block size as recorded in header (bytes 6-7)
	PlanetID        int // Planet where battle occurred
	X               int // X coordinate
	Y               int // Y coordinate
	AttackerStacks  int // Attacker stack count
	DefenderStacks  int // Defender stack count
	AttackerLosses  int // Ships lost by attacker
	DefenderLosses  int // Ships lost by defender

	// Parsed data
	Stacks  []BattleStack  // Stack definitions
	Actions []BattleAction // Action records
}

const (
	battleHeaderSize     = 18
	battleStackSize      = 29
	battleActionSize     = 22
	battleStackMarker    = 0x41
)

// NewBattleBlock creates a BattleBlock from a GenericBlock
func NewBattleBlock(b GenericBlock) *BattleBlock {
	bb := &BattleBlock{GenericBlock: b}
	bb.decode()
	return bb
}

func (bb *BattleBlock) decode() {
	data := bb.Decrypted
	if len(data) < battleHeaderSize {
		return
	}

	// Decode header (18 bytes)
	bb.BattleID = int(data[0])
	bb.Rounds = int(data[1]) + 1 // Stored as rounds-1
	bb.Side1Stacks = int(data[2])
	bb.TotalStacks = int(data[3])
	// bytes 4-5: unknown
	bb.RecordedSize = int(encoding.Read16(data, 6))
	bb.PlanetID = int(encoding.Read16(data, 8))
	bb.X = int(encoding.Read16(data, 10))
	bb.Y = int(encoding.Read16(data, 12))
	bb.AttackerStacks = int(data[14])
	bb.DefenderStacks = int(data[15])
	bb.AttackerLosses = int(data[16])
	bb.DefenderLosses = int(data[17])

	// Decode stack definitions
	stackStart := battleHeaderSize
	for s := 0; s < bb.TotalStacks && stackStart+battleStackSize <= len(data); s++ {
		stack := data[stackStart : stackStart+battleStackSize]
		bs := bb.decodeStack(stack, s)
		bb.Stacks = append(bb.Stacks, bs)
		stackStart += battleStackSize
	}

	// Decode action records
	actionStart := battleHeaderSize + bb.TotalStacks*battleStackSize
	numActions := (len(data) - actionStart) / battleActionSize
	for i := 0; i < numActions; i++ {
		offset := actionStart + i*battleActionSize
		if offset+battleActionSize > len(data) {
			break
		}
		actionData := data[offset : offset+battleActionSize]
		action := BattleAction{
			RawData: actionData,
			Events:  bb.decodeActionEvents(actionData),
		}
		bb.Actions = append(bb.Actions, action)
	}
}

// decodeActionEvents parses individual events from a 22-byte action record.
// Each action record can contain multiple events encoded as:
//   - [round][stack][stack][0x04] = movement
//   - [round][stack][stack][0xC4] = fire
//   - [0x64|0xE4][damage][stack][position] = damage taken
//   - [stack][position] = position update
func (bb *BattleBlock) decodeActionEvents(actionData []byte) []BattleRecordEvent {
	var events []BattleRecordEvent
	currentRound := 0
	pos := 0

	for pos < len(actionData)-1 {
		b := actionData[pos]

		// Pattern 1: Round marker followed by stack actions
		// Format: [round] [stack] [stack/target] [type]
		if b <= 15 && pos+3 < len(actionData) {
			stack1 := actionData[pos+1]
			stack2 := actionData[pos+2]
			actionType := actionData[pos+3]

			if stack1 <= 5 && stack2 <= 5 {
				if actionType == 0x04 {
					// Movement action
					events = append(events, BattleRecordEvent{
						Round:      int(b),
						StackID:    int(stack1),
						ActionType: ActionMove,
					})
					currentRound = int(b)
					pos += 4
					continue
				} else if actionType == 0xC4 {
					// Fire action
					events = append(events, BattleRecordEvent{
						Round:      int(b),
						StackID:    int(stack1),
						ActionType: ActionFire,
					})
					currentRound = int(b)
					pos += 4
					continue
				}
			}
		}

		// Pattern 2: Damage marker
		// Format: [0x64 or 0xE4] [amount] [stack] [position]
		if (b == 0x64 || b == 0xE4) && pos+3 < len(actionData) {
			dmgAmount := int(actionData[pos+1])
			stackID := int(actionData[pos+2])
			position := int(actionData[pos+3])

			if stackID <= 5 && position >= 0x40 && position <= 0x99 {
				gridX, gridY := posToGrid(position)
				events = append(events, BattleRecordEvent{
					Round:      currentRound,
					StackID:    stackID,
					ActionType: ActionDamage,
					Damage:     dmgAmount,
					GridX:      gridX,
					GridY:      gridY,
				})
				pos += 4
				continue
			}
		}

		// Pattern 3: Stack + Position (position update)
		// Format: [stack] [position]
		if b <= 5 && pos+1 < len(actionData) {
			position := int(actionData[pos+1])
			if position >= 0x40 && position <= 0x99 {
				gridX, gridY := posToGrid(position)
				events = append(events, BattleRecordEvent{
					Round:      currentRound,
					StackID:    int(b),
					ActionType: ActionMove,
					GridX:      gridX,
					GridY:      gridY,
				})
				pos += 2
				continue
			}
		}

		pos++
	}

	return events
}

// posToGrid converts a position byte (0x40-0x99) to grid coordinates.
// The Stars! battle grid is 10x10, with 0x40 being (0,0).
func posToGrid(pos int) (int, int) {
	if pos < 0x40 || pos > 0x99 {
		return -1, -1
	}
	return (pos - 0x40) % 10, (pos - 0x40) / 10
}

func (bb *BattleBlock) decodeStack(stack []byte, stackIndex int) BattleStack {
	bs := BattleStack{}

	// Find the 0x41 marker in the stack data
	markerPos := -1
	for i := 0; i < len(stack); i++ {
		if stack[i] == battleStackMarker {
			markerPos = i
			break
		}
	}

	if markerPos >= 1 {
		bs.DesignID = int(stack[markerPos-1])
	}
	if markerPos >= 0 && markerPos+1 < len(stack) {
		bs.ShipCount = int(stack[markerPos+1])
	}

	// Determine owner: stacks 0 to Side1Stacks-1 belong to side 1 (viewer)
	// stacks Side1Stacks to TotalStacks-1 belong to side 2 (enemy)
	if stackIndex < bb.Side1Stacks {
		bs.OwnerID = 0 // Side 1 (viewer)
	} else {
		bs.OwnerID = 1 // Side 2 (enemy)
	}

	return bs
}

// Side2Stacks returns the number of stacks on side 2 (enemy)
func (bb *BattleBlock) Side2Stacks() int {
	return bb.TotalStacks - bb.Side1Stacks
}

// AllEvents returns all decoded battle events across all action records.
func (bb *BattleBlock) AllEvents() []BattleRecordEvent {
	var events []BattleRecordEvent
	for _, action := range bb.Actions {
		events = append(events, action.Events...)
	}
	return events
}

// EventsByRound returns battle events grouped by round number.
func (bb *BattleBlock) EventsByRound() map[int][]BattleRecordEvent {
	byRound := make(map[int][]BattleRecordEvent)
	for _, event := range bb.AllEvents() {
		byRound[event.Round] = append(byRound[event.Round], event)
	}
	return byRound
}

// Phases returns the battle as a sequence of phases matching the Battle VCR display.
// Each phase represents one stack's turn to act within a round.
// Phases are extracted by scanning the raw action data for phase markers.
// The pattern [round][stack][stack][0x04|0xC4] marks each phase boundary.
//
// Known limitations:
//   - Detects approximately 60% of phases shown in the Battle VCR
//   - Early rounds (0-1) may have lower detection rates
//   - Later rounds (5+) are detected reliably
//   - Some phases may use different encoding not yet fully decoded
//
// The returned phases are accurate but not necessarily complete.
// Use RawActionData() for custom analysis if needed.
func (bb *BattleBlock) Phases() []BattlePhase {
	var phases []BattlePhase
	phaseNum := 1

	// Concatenate all action raw data for scanning
	var allData []byte
	for _, action := range bb.Actions {
		allData = append(allData, action.RawData...)
	}

	if len(allData) < 4 {
		return nil
	}

	// Track stack positions and last detected offset
	stackPositions := make(map[int][2]int)
	lastOffset := -4 // Minimum gap between phase markers to avoid overlapping matches

	// Scan for phase markers: [round][stack1][stack2][type]
	// where type = 0x04 (move) or 0xC4 (fire)
	for i := 0; i < len(allData)-3; i++ {
		round := int(allData[i])
		stack1 := int(allData[i+1])
		stack2 := int(allData[i+2])
		actionType := allData[i+3]

		// Check if this is a valid phase marker
		if round > 15 || stack1 > 5 || stack2 > 5 {
			continue
		}
		if actionType != 0x04 && actionType != 0xC4 {
			continue
		}

		// Avoid overlapping detections (false positives from within other data)
		if i-lastOffset < 4 {
			continue
		}

		// Found a phase marker
		phase := BattlePhase{
			PhaseNum: phaseNum,
			Round:    round,
			StackID:  stack1,
			TargetID: -1,
			Fired:    actionType == 0xC4,
		}
		lastOffset = i

		// Look for position data after the marker
		// Format: [marker 4 bytes] ... [stack][position] ...
		// Scan ahead for stack+position pairs
		for j := i + 4; j < len(allData)-1 && j < i+12; j++ {
			s := int(allData[j])
			pos := int(allData[j+1])
			if s <= 5 && pos >= 0x40 && pos <= 0x99 {
				phase.GridX, phase.GridY = posToGrid(pos)
				stackPositions[s] = [2]int{phase.GridX, phase.GridY}
				break
			}
		}

		// If we didn't find a position, use last known
		if phase.GridX == 0 && phase.GridY == 0 {
			if pos, ok := stackPositions[stack1]; ok {
				phase.GridX, phase.GridY = pos[0], pos[1]
			}
		}

		// Look for damage markers (0x64 or 0xE4) following this phase
		for j := i + 4; j < len(allData)-3 && j < i+12; j++ {
			if allData[j] == 0x64 || allData[j] == 0xE4 {
				dmgAmount := int(allData[j+1])
				targetStack := int(allData[j+2])
				if targetStack <= 5 {
					phase.Damage += dmgAmount
					if phase.TargetID == -1 {
						phase.TargetID = targetStack
					}
				}
				break
			}
		}

		phases = append(phases, phase)
		phaseNum++
	}

	return phases
}

// RawActionData returns the concatenated raw bytes from all action records.
// This can be used for custom analysis of the battle recording format.
func (bb *BattleBlock) RawActionData() []byte {
	var data []byte
	for _, action := range bb.Actions {
		data = append(data, action.RawData...)
	}
	return data
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
// Found in X files when player assigns a battle plan to a fleet
// Format: 4 bytes
//   Bytes 0-1: Fleet number (9 bits, little-endian)
//   Bytes 2-3: Battle plan index (little-endian)
type SetFleetBattlePlanBlock struct {
	GenericBlock

	FleetNumber     int // Fleet number (0-indexed, display is +1)
	BattlePlanIndex int // Battle plan index (0=Default, 1-4=custom plans)
}

// NewSetFleetBattlePlanBlock creates a SetFleetBattlePlanBlock from a GenericBlock
func NewSetFleetBattlePlanBlock(b GenericBlock) *SetFleetBattlePlanBlock {
	sfbp := &SetFleetBattlePlanBlock{GenericBlock: b}
	sfbp.decode()
	return sfbp
}

func (sfbp *SetFleetBattlePlanBlock) decode() {
	data := sfbp.Decrypted
	if len(data) < 4 {
		return
	}

	// Fleet number is 9 bits (like other fleet blocks)
	sfbp.FleetNumber = int(data[0]&0xFF) + (int(data[1]&0x01) << 8)
	// Battle plan index
	sfbp.BattlePlanIndex = int(data[2]&0xFF) + (int(data[3]) << 8)
}
