# BattleBlock (Type 31) and BattleContinuationBlock (Type 39)

**Source**: Decompiled from `stars26jrc3.exe` using Ghidra. Structure names from NB09 CodeView debug symbols.

The BattleBlock contains the battle VCR recording data. The structure was previously misunderstood - the header is **14 bytes** (not 18), and action records are **variable size** (not fixed 22 bytes).

## BTLDATA Header (14 bytes)

From the Stars! binary `BTLDATA` structure:

```c
typedef struct _btldata {
    uint16_t id;        // +0x00: Battle identifier
    uint8_t  cplr;      // +0x02: Number of players involved
    uint8_t  ctok;      // +0x03: Total stack count (TOK = token/stack)
    uint16_t grfPlr;    // +0x04: Player bitmask (bit N = player N involved)
    uint16_t cbData;    // +0x06: Total data size in bytes
    uint16_t idPlanet;  // +0x08: Planet ID (-1 = deep space, signed)
    POINT    pt;        // +0x0a: X,Y coordinates (4 bytes)
    TOK      rgtok[0];  // +0x0e: Stack array starts here
} BTLDATA;  // Total header: 14 bytes
```

| Offset | Size | Field    | Description                                        |
|--------|------|----------|----------------------------------------------------|
| 0x00   | 2    | id       | Battle identifier (uint16 LE)                      |
| 0x02   | 1    | cplr     | Number of players involved in battle               |
| 0x03   | 1    | ctok     | Total number of stacks                             |
| 0x04   | 2    | grfPlr   | Player bitmask - bit N set = player N is in battle |
| 0x06   | 2    | cbData   | Total block data size                              |
| 0x08   | 2    | idPlanet | Planet ID (int16 LE, -1 = deep space)              |
| 0x0a   | 2    | x        | X coordinate                                       |
| 0x0c   | 2    | y        | Y coordinate                                       |

**Note**: The old interpretation of bytes 14-17 as attacker/defender counts was incorrect. Those bytes are part of the first TOK structure.

## TOK Stack Structure (29 bytes each)

Each participating stack is defined by a `TOK` structure (0x1d = 29 bytes):

```c
typedef struct _tok {
    uint16_t id;          // +0x00: Fleet/Planet ID
    uint8_t  iplr;        // +0x02: Owner player ID (0-15)
    uint8_t  grobj;       // +0x03: Object type (1=starbase, other=fleet)
    uint8_t  ishdef;      // +0x04: Ship design ID
    uint8_t  brc;         // +0x05: Battle grid position (row*11 + col)
    uint8_t  initBase;    // +0x06: Base initiative
    uint8_t  initMin;     // +0x07: Minimum initiative
    uint8_t  initMac;     // +0x08: Maximum initiative
    uint8_t  itokTarget;  // +0x09: Target stack index
    uint8_t  pctCloak;    // +0x0a: Cloak percentage
    uint8_t  pctJam;      // +0x0b: Jammer percentage
    uint8_t  pctBC;       // +0x0c: Battle computer percentage
    uint8_t  pctCap;      // +0x0d: Capacitor percentage
    uint8_t  pctBeamDef;  // +0x0e: Beam deflector percentage
    uint16_t wt;          // +0x0f: Mass/weight
    uint16_t dpShield;    // +0x11: Shield hitpoints
    uint16_t csh;         // +0x13: Ship count
    DV       dv;          // +0x15: Armor damage value (2 bytes)
    uint16_t mdTarget;    // +0x17: Target mode bitfield
    // +0x19-0x1c: Additional fields (5 bytes)
} TOK;  // Total: 29 bytes (0x1d)
```

| Offset | Size | Field      | Description                           |
|--------|------|------------|---------------------------------------|
| 0x00   | 2    | id         | Fleet or planet ID                    |
| 0x02   | 1    | iplr       | Owner player (0-15)                   |
| 0x03   | 1    | grobj      | 1 = starbase, other = fleet           |
| 0x04   | 1    | ishdef     | Ship design index                     |
| 0x05   | 1    | brc        | Grid position (encoded)               |
| 0x06   | 1    | initBase   | Base initiative value                 |
| 0x07   | 1    | initMin    | Minimum initiative                    |
| 0x08   | 1    | initMac    | Maximum initiative                    |
| 0x09   | 1    | itokTarget | Target stack index                    |
| 0x0a   | 1    | pctCloak   | Cloak % (0-100)                       |
| 0x0b   | 1    | pctJam     | Jammer %                              |
| 0x0c   | 1    | pctBC      | Battle computer %                     |
| 0x0d   | 1    | pctCap     | Capacitor %                           |
| 0x0e   | 1    | pctBeamDef | Beam deflector %                      |
| 0x0f   | 2    | wt         | Ship mass                             |
| 0x11   | 2    | dpShield   | Shield HP                             |
| 0x13   | 2    | csh        | Ship count in stack                   |
| 0x15   | 2    | dv         | **VERIFIED** Damage state (DV struct) |
| 0x17   | 2    | mdTarget   | Target mode bits                      |
| 0x19   | 5    | -          | Additional fields                     |

## BTLREC Action Records (VARIABLE SIZE!)

**IMPORTANT**: Action records are NOT fixed 22-byte chunks. Each `BTLREC` has variable size:

```
Record size = 6 + (ctok × 8) bytes
```

Where `ctok` is the number of kill events in that action.

```c
typedef struct _btlrec {
    uint8_t  itok;      // +0x00: Acting stack index
    uint8_t  brcDest;   // +0x01: Destination grid position
    int16_t  ctok;      // +0x02: Kill record count (determines size!)
    // +0x04: Bitfield (16 bits):
    //   bits 0-3:  iRound (round number 0-15)
    //   bits 4-7:  dzDis (distance moved)
    //   bits 8-15: itokAttack (target stack index)
    KILL     rgkill[0]; // +0x06: Array of kill records
} BTLREC;  // Base: 6 bytes + ctok×8 bytes
```

| Offset | Size | Field    | Description                                          |
|--------|------|----------|------------------------------------------------------|
| 0x00   | 1    | itok     | Acting stack index (0 to ctok-1)                     |
| 0x01   | 1    | brcDest  | Destination grid position                            |
| 0x02   | 2    | ctok     | Number of KILL records following                     |
| 0x04   | 2    | bitfield | Round (4 bits) + distance (4 bits) + target (8 bits) |
| 0x06   | N×8  | rgkill   | Array of KILL structures                             |

## KILL Structure (8 bytes each)

Each kill event within a BTLREC:

```c
typedef struct _kill {
    uint8_t  itok;      // +0x00: Target stack index
    uint8_t  grfWeapon; // +0x01: Weapon type flags (0x01, 0x04, 0xC4 observed)
    uint16_t cshKill;   // +0x02: Number of ships killed
    uint16_t dpShield;  // +0x04: Shield damage dealt
    DV       dv;        // +0x06: Target's damage STATE after attack (DV struct)
} KILL;  // Total: 8 bytes
```

| Offset | Size | Field     | Status       | Description                                          |
|--------|------|-----------|--------------|------------------------------------------------------|
| 0x00   | 1    | itok      | **VERIFIED** | Target stack that was hit                            |
| 0x01   | 1    | grfWeapon | Partial      | Weapon type flags (values 0x01, 0x04, 0xC4 observed) |
| 0x02   | 2    | cshKill   | **VERIFIED** | Ships destroyed (matches VCR display)                |
| 0x04   | 2    | dpShield  | **VERIFIED** | Shield damage dealt (matches VCR display)            |
| 0x06   | 2    | dv        | **VERIFIED** | Target's damage STATE after attack (DV struct)       |

**IMPORTANT: The `dv` field contains the target's damage STATE, not damage dealt!**

The VCR-displayed "armor damage" is calculated from weapon damage vs shields/armor, not read from the `dv` field. The `dv` stores how damaged the target is AFTER the attack.

Example interpretation of observed values:
- Phase 9: dv=868 (0x0364) → pctSh=100, pctDp=6 → target at 1.2% armor damage
- Phase 12: dv=15076 (0x3AE4) → pctSh=100, pctDp=117 → target at 23.4% armor damage
- Phase 61: dv=63972 (0xF9E4) → pctSh=100, pctDp=499 → target at 99.8% armor damage (nearly dead)

## DV (Damage Value) Structure (2 bytes) - **VERIFIED FROM DECOMPILATION**

The DV structure stores the **damage STATE** of a stack (not the damage dealt). It's a bit-packed 16-bit value:

```c
// From stars26jrc3.exe decompilation (FDamageTok @ 10f0:81d4)
typedef struct _dv {
    union {
        uint16_t dp;      // Raw 16-bit value
        struct {
            uint16_t pctSh : 7;  // Bits 0-6: % of ships with partial damage (0-100+)
            uint16_t pctDp : 9;  // Bits 7-15: Armor damage % (0-499, capped)
        };
    };
} DV;
```

**How DV is encoded** (from FDamageTok):
```c
if (pctDp > 499) pctDp = 499;  // Cap at 499%
dv = (pctDp << 7) | (pctSh & 0x7F);
```

**How DV is decoded to remaining armor** (from LdpFromItokDv @ 10e8:07a8):
```c
// Get base armor per ship from ship definition (HUL.dp at offset 0x38)
int baseArmor = shdef->hul.dp;
int shipCount = tok->csh;

// Total armor capacity
long totalArmor = baseArmor * shipCount;

if (dv != 0) {
    // Ships with distributed damage
    int affectedShips = (shipCount * (dv & 0x7F)) / 100;
    if (affectedShips < 1) affectedShips = 1;

    // Damage to subtract
    int dmg = (baseArmor * (dv >> 7)) / 10 * affectedShips / 50;
    totalArmor -= dmg;
}
return totalArmor;  // Remaining armor HP
```

**Corrected interpretation of observed values:**
- dv=868 (0x0364): pctSh=100, pctDp=6 → 100% ships have 6×10/500=1.2% armor damage
- dv=15076 (0x3AE4): pctSh=100, pctDp=117 → 100% ships have 23.4% armor damage
- dv=63972 (0xF9E4): pctSh=100, pctDp=499 → 100% ships have 99.8% armor damage (nearly dead)

**Key insight**: The DV in KILL records stores the TARGET's damage state AFTER the attack, not the damage dealt. The VCR calculates "armor damage dealt" from weapon stats - it's not stored directly in the record.

## Battle Damage Formulas - **VERIFIED FROM DECOMPILATION**

**Source**: Decompiled from `stars26jrc3.exe` - functions `FDamageTok`, `RegenShield`, `CTorpHit`, `DpFromPtokBrcToBrc`

### Shield Damage (FDamageTok @ 10f0:81d4)

Shields are a pool shared across all ships in a stack:
```c
// TOK offsets: 0x11 = dpShield (per ship), 0x13 = csh (ship count)
long totalShields = dpShield * shipCount;

if (damage < totalShields) {
    // Shields absorb all damage
    dpShield = (totalShields - damage) / shipCount;
    remainingDamage = 0;
} else {
    // Shields destroyed, excess goes to armor
    remainingDamage = damage - totalShields;
    dpShield = 0;
}
```

### Armor Damage (FDamageTok @ 10f0:81d4)

Armor damage is distributed across ships, with damaged ships killed first:
```c
// Get existing damage state from DV (offset 0x15)
int pctDp = dv >> 7;       // Armor damage % (0-499)
int pctSh = dv & 0x7F;     // % of ships already damaged

// Ships with existing damage
int cshDamaged = (shipCount * pctSh) / 100;
int damagePerShip = (baseArmor * pctDp) / 500;

// 1. Kill damaged ships first (they have less armor remaining)
int remainingArmor = baseArmor - damagePerShip;
while (damage >= remainingArmor && cshDamaged > 0) {
    damage -= remainingArmor;
    cshDamaged--;
    shipCount--;
    killCount++;
}

// 2. Kill undamaged ships
while (damage >= baseArmor && shipCount > 0) {
    damage -= baseArmor;
    shipCount--;
    killCount++;
}

// 3. Distribute remaining damage to survivors
if (damage > 0 && shipCount > 0) {
    pctDp = min((damage * 500) / baseArmor, 499);
    pctSh = 100;  // All survivors now damaged
} else if (cshDamaged > 0) {
    pctSh = (cshDamaged * 100) / shipCount;
    // pctDp stays the same
}

// Pack new DV
dv = (pctDp << 7) | (pctSh & 0x7F);
```

### Shield Regeneration (RegenShield @ 10f0:3c16)

Shields regenerate 10% per battle round:
```c
int maxShield = DpShieldOfShdef(shdef, player);
int regen = maxShield / 10;  // 10% regeneration
dpShield = min(dpShield + regen, maxShield);
```

### Torpedo Hit Calculation (CTorpHit @ 10f0:6790)

Torpedo accuracy is modified by jammer vs battle computer:
```c
int pctHit = baseAccuracy;  // From weapon stats

if (targetJammer > attackerBattleComp) {
    // Jammer reduces hit chance
    pctHit = pctHit * (100 - (targetJammer - attackerBattleComp)) / 100;
} else {
    // Battle computer increases hit chance
    int bonus = attackerBattleComp - targetJammer;
    pctHit = 100 - (100 - pctHit) * (100 - bonus) / 100;
}

if (pctHit < 1) pctHit = 1;

// For small salvos (<200), roll each torpedo
// For large salvos, use average: hits = torpedoes * pctHit / 100
```

### Beam Damage with Range (DpFromPtokBrcToBrc @ 10f0:4d2e)

Beam weapons lose effectiveness at range:
```c
int baseDamage = weaponDamage * slotCount;

if (range > 0 && weaponRange > 0) {
    // Damage falloff with distance
    int pctFalloff = (range * 100) / weaponRange;
    baseDamage = baseDamage * (100 - pctFalloff) / 100;
}

// Beam deflector reduces damage
if (targetBeamDeflect > 0) {
    baseDamage = baseDamage * (100 - targetBeamDeflect) / 100;
}
```

## Block Continuation (Type 39)

When battle data exceeds 1024 bytes (0x400), the game splits it across multiple blocks:

1. **First block (Type 31)**: Header + up to 35 stacks + initial action records
2. **Continuation blocks (Type 39)**: Additional stacks and/or action records

From `WriteBattles` in the binary:
```c
if (cbData >= 0x400) {
    // Write header + stacks first (max 0x22 = 34 stacks per block)
    WriteRt(0x1f, ctok * 0x1d + 0x0e, lpbtldata);  // Type 31

    // Write remaining stacks in continuation blocks
    while (remaining_stacks > 0) {
        WriteRt(0x27, min(remaining * 0x1d, 0x3f7), data);  // Type 39
    }

    // Write action records in continuation blocks
    WriteRt(0x27, action_data_size, action_data);  // Type 39
}
```

**Status**: Type 39 has not yet been observed in test data. It appears to be used only for very long battles (many stacks, high movement ships, 16 rounds).

## Grid Position Encoding

Battle grid positions use base-11 encoding for a 10×10 grid:

```
position = col × 11 + row
```

Where col and row are 0-9. Decoding:
```
col = position / 11
row = position % 11
```

Examples:
| Encoded | Decimal | Grid (col, row)   |
|---------|---------|-------------------|
| 0x00    | 0       | (0, 0)            |
| 0x25    | 37      | (3, 4)            |
| 0x52    | 82      | (7, 5)            |
| 0x6D    | 109     | (9, 10) - invalid |

## Round and Phase Calculation

**Phase counting**: Total phases = number of action records + 1 (setup phase). This is fully decoded and verified against real game files.

**Round calculation**: Scan BTLREC records for the `iRound` field (bits 0-3 of offset 0x04). Rounds = max_round + 1.

## Example Battle Layout

```
Offset  Content
------  -------
0x000   BTLDATA header (14 bytes)
0x00E   TOK[0] - Stack 0 (29 bytes)
0x02B   TOK[1] - Stack 1 (29 bytes)
0x048   TOK[2] - Stack 2 (29 bytes)
...
0x0XX   BTLREC[0] - First action (6+ bytes)
0x0YY   BTLREC[1] - Second action (6+ bytes)
...
```

**Example from test data (Hobbits vs Halflings at Redmond):**
- Battle ID: 1
- Rounds: 16 (stored as 15)
- Total stacks: 6 (4 attackers + 2 defenders)
- Planet: 392 (Redmond)
- Location: (1943, 2087)
- Attacker losses: 1, Defender losses: 2

## Limitations and Unknown Encodings

1. **Initial placement**: Round 0 may use different encoding for initial stack positions (not the standard `[round][stack][stack][type]` pattern)

2. **Fractional movement**: Stacks with movement > 1 (e.g., 1¾) get multiple phases per round. The encoding for these additional phases is not fully decoded.
