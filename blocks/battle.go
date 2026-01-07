package blocks

import (
	"fmt"

	"github.com/neper-stars/houston/encoding"
)

// DV represents a damage value structure (2 bytes) that stores the damage STATE
// of a stack (not the damage dealt). It's a bit-packed 16-bit value from decompilation.
//
// Structure (from stars26jrc3.exe FDamageTok @ 10f0:81d4):
//
//	Bits 0-6 (pctSh): % of ships with partial damage (0-100+)
//	Bits 7-15 (pctDp): Armor damage % (0-499, capped)
//
// Encoding: dv = (pctDp << 7) | (pctSh & 0x7F)
type DV uint16

// NewDV creates a DV value from percentages.
//   - armorDamagePercent: Armor damage as a percentage (0.0 to 99.8)
//   - shipsAffectedPercent: Percentage of ships with damage (0 to 100)
//
// Example: NewDV(50.0, 100) creates a DV representing 50% armor damage on all ships.
func NewDV(armorDamagePercent float64, shipsAffectedPercent int) DV {
	// Convert percentage to internal pctDp (scaled by 5, capped at 499)
	pctDp := int(armorDamagePercent * 5)
	if pctDp > 499 {
		pctDp = 499
	}
	if pctDp < 0 {
		pctDp = 0
	}

	// pctSh is stored directly (0-127)
	pctSh := shipsAffectedPercent
	if pctSh > 127 {
		pctSh = 127
	}
	if pctSh < 0 {
		pctSh = 0
	}

	return DV((pctDp << 7) | (pctSh & 0x7F))
}

// PctSh returns the percentage of ships with partial damage (bits 0-6).
// Value range is 0-100+ (can exceed 100 in some edge cases).
func (dv DV) PctSh() int {
	return int(dv & 0x7F)
}

// PctDp returns the armor damage percentage (bits 7-15).
// Value range is 0-499 (capped at 499 = 99.8% damage).
func (dv DV) PctDp() int {
	return int(dv >> 7)
}

// ArmorDamagePercent returns the armor damage as a human-readable percentage.
// The pctDp value is scaled: actual % = pctDp / 5.
func (dv DV) ArmorDamagePercent() float64 {
	return float64(dv.PctDp()) / 5.0
}

// IsZero returns true if this DV represents no damage state.
func (dv DV) IsZero() bool {
	return dv == 0
}

// String returns a human-readable representation of the damage state.
func (dv DV) String() string {
	if dv.IsZero() {
		return "undamaged"
	}
	return fmt.Sprintf("%.1f%% armor damage (%d%% ships affected)",
		dv.ArmorDamagePercent(), dv.PctSh())
}

// Battle tactics
const (
	TacticDisengage             = 0
	TacticDisengageIfChallenged = 1
	TacticMinimizeDamage        = 2
	TacticMaximizeNetDamage     = 3
	TacticMaximizeDamageRatio   = 4
	TacticMaximizeDamage        = 5
)

// Battle target types
const (
	TargetNone           = 0
	TargetAny            = 1
	TargetStarbase       = 2
	TargetArmedShips     = 3
	TargetBombers        = 4 // Also freighters
	TargetUnarmedShips   = 5
	TargetFuelTransports = 6
	TargetFreighters     = 7
)

// Attack who values
const (
	AttackNobody            = 0
	AttackEnemies           = 1
	AttackNeutralAndEnemies = 2
	AttackEveryone          = 3
	AttackPlayerBase        = 4 // Player ID = value - 4
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
	bpb.OwnerPlayerId = int(word0 & 0x0F) // Bits 0-3
	bpb.PlanId = int((word0 >> 4) & 0x0F) // Bits 4-7
	bpb.Tactic = int((word0 >> 8) & 0x0F) // Bits 8-11
	bpb.DumpCargo = (word0 & 0x8000) != 0 // Bit 15

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

// Encode returns the raw block data bytes (without the 2-byte block header).
func (bpb *BattlePlanBlock) Encode() []byte {
	// Build word0: owner (4 bits) | planId (4 bits) | tactic (4 bits) | reserved (3 bits) | dumpCargo (1 bit)
	var word0 = uint16(bpb.OwnerPlayerId & 0x0F)
	word0 |= uint16((bpb.PlanId & 0x0F) << 4)
	word0 |= uint16((bpb.Tactic & 0x0F) << 8)
	if bpb.DumpCargo {
		word0 |= 0x8000
	}

	// Build word1: primaryTarget (4 bits) | secondaryTarget (4 bits) | attackWho (8 bits)
	var word1 = uint16(bpb.PrimaryTarget & 0x0F)
	word1 |= uint16((bpb.SecondaryTarget & 0x0F) << 4)
	word1 |= uint16((bpb.AttackWho & 0xFF) << 8)

	// If deleted, only return the 4-byte header
	if bpb.Deleted {
		data := make([]byte, 4)
		encoding.Write16(data, 0, word0)
		encoding.Write16(data, 2, word1)
		return data
	}

	// Encode the plan name
	nameEncoded := encoding.EncodeStarsString(bpb.Name)
	data := make([]byte, 4+len(nameEncoded))
	encoding.Write16(data, 0, word0)
	encoding.Write16(data, 2, word1)
	copy(data[4:], nameEncoded)

	return data
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

// BattleStack represents a "stack" in a Stars! battle (TOK structure, 29 bytes).
// A stack is a group of ships of the same design that move and fight as a unit.
// When a fleet enters battle, it splits into stacks - one per ship design.
// Example: A fleet with 3 Scouts and 5 Destroyers becomes 2 stacks.
//
// Structure from Stars! binary (TOK, 0x1d = 29 bytes):
//
//	+0x00: uint16 id         - Fleet/Planet ID
//	+0x02: uint8  iplr       - Owner player ID (0-15)
//	+0x03: uint8  grobj      - Object type (1=starbase, other=fleet)
//	+0x04: uint8  ishdef     - Ship design ID
//	+0x05: uint8  brc        - Battle grid position
//	+0x06: uint8  initBase   - Base initiative
//	+0x07: uint8  initMin    - Min initiative
//	+0x08: uint8  initMac    - Max initiative
//	+0x09: uint8  itokTarget - Target stack index
//	+0x0a: uint8  pctCloak   - Cloak percentage
//	+0x0b: uint8  pctJam     - Jammer percentage
//	+0x0c: uint8  pctBC      - Battle computer percentage
//	+0x0d: uint8  pctCap     - Capacitor percentage
//	+0x0e: uint8  pctBeamDef - Beam deflector percentage
//	+0x0f: uint16 wt         - Mass/weight
//	+0x11: uint16 dpShield   - Shield hitpoints
//	+0x13: uint16 csh        - Ship count
//	+0x15: uint16 dv         - Armor damage value
//	+0x17: uint16 mdTarget   - Target mode bitfield
//	+0x19-0x1c: additional fields
type BattleStack struct {
	// Identification
	FleetOrPlanetID int  // ID of fleet or planet (offset 0x00)
	OwnerPlayerID   int  // Owner player 0-15 (offset 0x02)
	IsStarbase      bool // True if starbase (grobj==1), false if fleet (offset 0x03)
	DesignID        int  // Ship design ID (offset 0x04)

	// Position and targeting
	GridPosition int // Battle grid position (offset 0x05)
	TargetStack  int // Target stack index (offset 0x09)

	// Initiative
	InitiativeBase int // Base initiative (offset 0x06)
	InitiativeMin  int // Minimum initiative (offset 0x07)
	InitiativeMax  int // Maximum initiative (offset 0x08)

	// Combat modifiers (percentages)
	CloakPercent       int // Cloak percentage (offset 0x0a)
	JammerPercent      int // Jammer percentage (offset 0x0b)
	BattleCompPercent  int // Battle computer percentage (offset 0x0c)
	CapacitorPercent   int // Capacitor percentage (offset 0x0d)
	BeamDeflectPercent int // Beam deflector percentage (offset 0x0e)

	// Stats
	Mass        int // Ship mass/weight (offset 0x0f)
	ShieldHP    int // Shield hitpoints (offset 0x11)
	ShipCount   int // Number of ships in stack (offset 0x13)
	DamageState DV  // Damage state at battle start (offset 0x15, DV struct)
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

// KillRecord represents damage dealt in a single attack (KILL structure, 8 bytes).
//
// Structure from Stars! binary (KILL):
//
//	+0x00: uint8  itok      - Target stack index
//	+0x01: uint8  grfWeapon - Weapon type flags
//	+0x02: uint16 cshKill   - Ships destroyed (verified against VCR)
//	+0x04: uint16 dpShield  - Shield damage dealt (verified against VCR)
//	+0x06: DV     dv        - Target's damage STATE after attack (not damage dealt!)
//
// Note: The armor damage displayed in the VCR is calculated from weapon stats,
// NOT read from the KILL record. The DV field stores the cumulative damage state.
type KillRecord struct {
	StackID      int // Target stack that took damage
	WeaponFlags  int // Weapon type flags (0x01=beam, 0x04=torp, 0xC4 observed)
	ShipsKilled  int // Number of ships destroyed
	ShieldDamage int // Shield damage dealt
	TargetDV     DV  // Target's damage state AFTER this attack
}

// BattleAction represents a single action record in the battle.
// Each record can contain multiple events (movement, firing, damage).
type BattleAction struct {
	RawData []byte              // Raw action data (variable size)
	Events  []BattleRecordEvent // Decoded events from this action
	Kills   []KillRecord        // Ships destroyed in this action
}

// BattlePhase represents one stack's turn in the battle, matching the
// "Phase X of Y" display in the Stars! Battle VCR viewer.
// Each phase shows a single stack acting: moving, firing, and/or dealing damage.
type BattlePhase struct {
	PhaseNum int  // Phase number (1-based, as shown in Battle VCR)
	Round    int  // Round number (0-15)
	StackID  int  // Acting stack ID (0-5)
	GridX    int  // Stack's grid X position (0-9)
	GridY    int  // Stack's grid Y position (0-9)
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
//
// Structure from Stars! binary (BTLDATA):
//
//	Header (14 bytes):
//	  +0x00: uint16 id       - Battle identifier
//	  +0x02: uint8  cplr     - Player count involved in battle
//	  +0x03: uint8  ctok     - Total stack count
//	  +0x04: uint16 grfPlr   - Player bitmask (bit N = player N involved)
//	  +0x06: uint16 cbData   - Total data size in bytes
//	  +0x08: uint16 idPlanet - Planet ID (-1 if in deep space)
//	  +0x0a: uint16 x        - X coordinate
//	  +0x0c: uint16 y        - Y coordinate
//	  +0x0e: TOK rgtok[]     - Stack array starts here
//
//	Stack definitions (TOK): 29 bytes each
//	Action records (BTLREC): variable size (ctok*8 + 6 bytes each)
type BattleBlock struct {
	GenericBlock

	// Header fields (14 bytes from BTLDATA)
	BattleID      int    // Battle identifier (+0x00)
	PlayerCount   int    // Number of players involved (+0x02)
	TotalStacks   int    // Total stack count (+0x03)
	PlayerBitmask uint16 // Which players are involved (+0x04)
	RecordedSize  int    // Total data size (+0x06)
	PlanetID      int    // Planet ID, -1 if deep space (+0x08)
	X             int    // X coordinate (+0x0a)
	Y             int    // Y coordinate (+0x0c)

	// Derived fields
	Rounds int // Calculated from action data

	// Parsed data
	Stacks  []BattleStack  // Stack definitions (TOK array)
	Actions []BattleAction // Action records (BTLREC array)
}

const (
	battleHeaderSize = 14 // BTLDATA header is 14 bytes, not 18
	battleStackSize  = 29 // TOK structure is 0x1d = 29 bytes
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

	// Decode header (14 bytes from BTLDATA structure)
	bb.BattleID = int(encoding.Read16(data, 0))        // +0x00: uint16 id
	bb.PlayerCount = int(data[2])                      // +0x02: uint8 cplr
	bb.TotalStacks = int(data[3])                      // +0x03: uint8 ctok
	bb.PlayerBitmask = encoding.Read16(data, 4)        // +0x04: uint16 grfPlr
	bb.RecordedSize = int(encoding.Read16(data, 6))    // +0x06: uint16 cbData
	bb.PlanetID = int(int16(encoding.Read16(data, 8))) // +0x08: uint16 idPlanet (signed, -1 = deep space)
	bb.X = int(encoding.Read16(data, 10))              // +0x0a: uint16 x
	bb.Y = int(encoding.Read16(data, 12))              // +0x0c: uint16 y

	// Decode stack definitions (TOK array starts at offset 0x0e)
	stackStart := battleHeaderSize
	for s := 0; s < bb.TotalStacks && stackStart+battleStackSize <= len(data); s++ {
		stack := data[stackStart : stackStart+battleStackSize]
		bs := bb.decodeStack(stack)
		bb.Stacks = append(bb.Stacks, bs)
		stackStart += battleStackSize
	}

	// Action records (BTLREC) follow the stacks
	// Each BTLREC has variable size: base 6 bytes + ctok*8 bytes for kills
	actionStart := battleHeaderSize + bb.TotalStacks*battleStackSize
	bb.decodeActionRecords(data[actionStart:])

	// Calculate actual rounds from action data
	bb.Rounds = bb.calculateRoundsFromActions()
}

// decodeActionRecords parses the BTLREC array which has variable-size records.
// Each BTLREC structure:
//
//	+0x00: uint8  itok       - Acting stack index
//	+0x01: uint8  brcDest    - Destination grid position
//	+0x02: int16  ctok       - Kill record count
//	+0x04: uint16 bitfield   - iRound:4, dzDis:4, itokAttack:8
//	+0x06: KILL[] rgkill     - Array of kill records (8 bytes each)
//
// Record size = 6 + (ctok * 8) bytes
func (bb *BattleBlock) decodeActionRecords(data []byte) {
	offset := 0
	for offset+6 <= len(data) {
		// Read BTLREC header
		itok := int(data[offset])
		brcDest := int(data[offset+1])
		ctok := int(int16(encoding.Read16(data, offset+2)))
		bitfield := encoding.Read16(data, offset+4)

		iRound := int(bitfield & 0x0F)
		_ = int((bitfield >> 4) & 0x0F) // dzDis - distance moved, unused for now
		itokAttack := int((bitfield >> 8) & 0xFF)

		// Calculate record size
		recordSize := 6
		if ctok > 0 {
			recordSize += ctok * 8 // Each KILL is 8 bytes
		}

		if offset+recordSize > len(data) {
			break
		}

		// Create action from this record
		action := BattleAction{
			RawData: data[offset : offset+recordSize],
		}

		// Add a move/fire event based on the record
		event := BattleRecordEvent{
			Round:   iRound,
			StackID: itok,
		}

		// Decode grid position from brcDest
		if isValidPosition(brcDest) {
			event.GridX, event.GridY = posToGrid(brcDest)
			event.ActionType = ActionMove
		}

		// If there's a target, it's a fire action
		if itokAttack < bb.TotalStacks && itokAttack != itok {
			event.ActionType = ActionFire
			event.TargetID = itokAttack
		}

		action.Events = append(action.Events, event)

		// Decode kill records if present
		// KILL record structure (8 bytes):
		//   +0x00: uint8  itok       - target stack index that took damage
		//   +0x01: uint8  grfWeapon  - weapon flags (0x01=beam, 0x04=torp, 0xC4 observed)
		//   +0x02: uint16 cshKill    - number of ships destroyed (verified against VCR)
		//   +0x04: uint16 dpShield   - shield damage dealt (verified against VCR)
		//   +0x06: DV     dv         - target's damage STATE after attack (not damage dealt!)
		if ctok > 0 {
			killOffset := offset + 6
			for k := 0; k < ctok && killOffset+8 <= len(data); k++ {
				killItok := int(data[killOffset])
				grfWeapon := int(data[killOffset+1])
				cshKill := int(encoding.Read16(data, killOffset+2))
				dpShield := int(encoding.Read16(data, killOffset+4))
				targetDV := DV(encoding.Read16(data, killOffset+6))

				// Track all damage records (not just kills)
				action.Kills = append(action.Kills, KillRecord{
					StackID:      killItok,
					WeaponFlags:  grfWeapon,
					ShipsKilled:  cshKill,
					ShieldDamage: dpShield,
					TargetDV:     targetDV,
				})

				if cshKill > 0 || dpShield > 0 {
					dmgEvent := BattleRecordEvent{
						Round:      iRound,
						StackID:    killItok,
						ActionType: ActionDamage,
						Damage:     dpShield + cshKill*100, // Approximate damage
					}
					action.Events = append(action.Events, dmgEvent)
				}
				killOffset += 8
			}
		}

		bb.Actions = append(bb.Actions, action)
		offset += recordSize

		// Safety check - if ctok is negative or we're not advancing, break
		if recordSize <= 0 {
			break
		}
	}
}

// calculateRoundsFromActions finds the maximum round number from decoded action
// events and returns the total round count. Rounds are 0-indexed in the data,
// so we add 1 to convert to count.
func (bb *BattleBlock) calculateRoundsFromActions() int {
	maxRound := -1
	for _, action := range bb.Actions {
		for _, event := range action.Events {
			if event.Round > maxRound {
				maxRound = event.Round
			}
		}
	}

	if maxRound >= 0 {
		return maxRound + 1 // Convert 0-indexed to count
	}
	return 1 // Default to 1 round if no actions found
}

// posToGrid converts a position byte to grid coordinates.
// The Stars! battle grid is 10x10.
// Position encoding: position = col*11 + row
// Example: (7,5) = 7*11 + 5 = 82 = 0x52
func posToGrid(pos int) (int, int) {
	col := pos / 11
	row := pos % 11
	if col > 9 || row > 9 {
		return -1, -1
	}
	return col, row
}

// isValidPosition checks if a byte value represents a valid battle grid position.
// Valid positions encode (col, row) where 0 <= col <= 9 and 0 <= row <= 9.
// Position = col*11 + row, so valid range is 0 to 108.
func isValidPosition(pos int) bool {
	col := pos / 11
	row := pos % 11
	return col <= 9 && row <= 9
}

func (bb *BattleBlock) decodeStack(stack []byte) BattleStack {
	bs := BattleStack{}

	if len(stack) < battleStackSize {
		return bs
	}

	// Decode TOK structure (29 bytes)
	bs.FleetOrPlanetID = int(encoding.Read16(stack, 0)) // +0x00: uint16 id
	bs.OwnerPlayerID = int(stack[2])                    // +0x02: uint8 iplr
	bs.IsStarbase = stack[3] == 1                       // +0x03: uint8 grobj (1=starbase)
	bs.DesignID = int(stack[4])                         // +0x04: uint8 ishdef
	bs.GridPosition = int(stack[5])                     // +0x05: uint8 brc
	bs.InitiativeBase = int(stack[6])                   // +0x06: uint8 initBase
	bs.InitiativeMin = int(stack[7])                    // +0x07: uint8 initMin
	bs.InitiativeMax = int(stack[8])                    // +0x08: uint8 initMac
	bs.TargetStack = int(stack[9])                      // +0x09: uint8 itokTarget
	bs.CloakPercent = int(stack[10])                    // +0x0a: uint8 pctCloak
	bs.JammerPercent = int(stack[11])                   // +0x0b: uint8 pctJam
	bs.BattleCompPercent = int(stack[12])               // +0x0c: uint8 pctBC
	bs.CapacitorPercent = int(stack[13])                // +0x0d: uint8 pctCap
	bs.BeamDeflectPercent = int(stack[14])              // +0x0e: uint8 pctBeamDef
	bs.Mass = int(encoding.Read16(stack, 15))           // +0x0f: uint16 wt
	bs.ShieldHP = int(encoding.Read16(stack, 17))       // +0x11: uint16 dpShield
	bs.ShipCount = int(encoding.Read16(stack, 19))      // +0x13: uint16 csh
	bs.DamageState = DV(encoding.Read16(stack, 21))     // +0x15: DV (damage state)

	return bs
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

// StacksForPlayer returns the number of stacks belonging to a specific player.
func (bb *BattleBlock) StacksForPlayer(playerID int) int {
	count := 0
	for _, stack := range bb.Stacks {
		if stack.OwnerPlayerID == playerID {
			count++
		}
	}
	return count
}

// StackCountByPlayer returns a map of player ID to stack count.
func (bb *BattleBlock) StackCountByPlayer() map[int]int {
	counts := make(map[int]int)
	for _, stack := range bb.Stacks {
		counts[stack.OwnerPlayerID]++
	}
	return counts
}

// CasualtiesForPlayer returns the total ships lost by a specific player.
func (bb *BattleBlock) CasualtiesForPlayer(playerID int) int {
	total := 0
	for _, action := range bb.Actions {
		for _, kill := range action.Kills {
			// Map stack ID to player
			if kill.StackID < len(bb.Stacks) && bb.Stacks[kill.StackID].OwnerPlayerID == playerID {
				total += kill.ShipsKilled
			}
		}
	}
	return total
}

// CasualtiesByPlayer returns a map of player ID to total ships lost.
func (bb *BattleBlock) CasualtiesByPlayer() map[int]int {
	casualties := make(map[int]int)
	for _, action := range bb.Actions {
		for _, kill := range action.Kills {
			// Map stack ID to player
			if kill.StackID < len(bb.Stacks) {
				playerID := bb.Stacks[kill.StackID].OwnerPlayerID
				casualties[playerID] += kill.ShipsKilled
			}
		}
	}
	return casualties
}

// TotalPhases returns the total number of phases as displayed in the Battle VCR.
// The game shows "Phase X of Y" where Y = len(Actions) + 1.
//
// The +1 accounts for the initial "setup" phase (Phase 1 in the VCR) which displays
// all stacks at their starting positions before any movement occurs. This setup phase
// has no corresponding action record - the initial positions come from the stack
// definitions (TOK structures). Each subsequent phase (2 through Y) corresponds to
// one action record (BTLREC structure).
func (bb *BattleBlock) TotalPhases() int {
	return len(bb.Actions) + 1
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
			if s <= 5 && isValidPosition(pos) {
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

// Encode returns the raw block data bytes (without the 2-byte block header).
func (bcb *BattleContinuationBlock) Encode() []byte {
	// Preserve raw data since structure is not fully documented
	if bcb.Decrypted != nil {
		return bcb.Decrypted
	}
	return bcb.Data
}

// SetFleetBattlePlanBlock represents setting a fleet's battle plan (Type 42)
// Found in X files when player assigns a battle plan to a fleet
// Format: 4 bytes
//
//	Bytes 0-1: Fleet number (9 bits, little-endian)
//	Bytes 2-3: Battle plan index (little-endian)
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

// Encode returns the raw block data bytes (without the 2-byte block header).
func (sfbp *SetFleetBattlePlanBlock) Encode() []byte {
	data := make([]byte, 4)
	data[0] = byte(sfbp.FleetNumber & 0xFF)
	data[1] = byte((sfbp.FleetNumber >> 8) & 0x01)
	data[2] = byte(sfbp.BattlePlanIndex & 0xFF)
	data[3] = byte((sfbp.BattlePlanIndex >> 8) & 0xFF)
	return data
}
