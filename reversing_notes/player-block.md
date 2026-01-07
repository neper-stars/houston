# PlayerBlock (Type 6)

The PlayerBlock contains player data including race settings, research, diplomatic relations, and production templates.

## Basic Structure

When `FullDataFlag` is set, `FullDataBytes` (104 bytes starting at offset 8) contains race settings:

| Offset | Size | Field                                                                         |
|--------|------|-------------------------------------------------------------------------------|
| 8-16   | 9    | Habitability ranges                                                           |
| 17     | 1    | Growth rate (max population growth %, typically 1-20)                         |
| 18-23  | 6    | Tech levels (Energy, Weapons, Propulsion, Construction, Electronics, Biotech) |

## Player Flags (offset 0x54)

Player state flags are stored at offset 0x54 (84 decimal) in the PlayerBlock as a 16-bit value.

### Binary Format

```
Bits 0-4:   State flags
Bits 5-15:  Unused (always 0)
```

### Flag Definitions

| Bit | Mask | Name     | Description                         |
|-----|------|----------|-------------------------------------|
| 0   | 0x01 | Dead     | Player has been eliminated          |
| 1   | 0x02 | Crippled | Player is crippled (definition TBD) |
| 2   | 0x04 | Cheater  | Cheater flag detected               |
| 3   | 0x08 | Learned  | Unknown purpose                     |
| 4   | 0x10 | Hacker   | Hacker flag detected                |

### Notes

- The Cheater and Hacker flags may be set by the game when certain exploit conditions are detected
- The Crippled flag purpose needs further investigation (possibly related to victory conditions)

---

## Zip Production Queue (offset 0x56)

The "Zip Production" feature allows players to define production templates that can be quickly applied to any planet. The Default template (Q1) is auto-applied to newly conquered planets.

### Storage Location

**In PlayerBlock (Type 6, M files):**
- Offset 0x56 (86 decimal), 26 bytes total
- Only the Default queue (Q1) is stored; other custom queues appear to be client-side only

**In SaveAndSubmitBlock (Type 46, X files):**
- Variable size: 2 + (2 × itemCount) bytes
- Contains the zip prod order being submitted

### Binary Format

```
FR NN [II II] [II II] ... [padding]
│  │  └─────────────────┘
│  │           └─── Items (2 bytes each, up to 12 items)
│  └────────────── Item count (0-12)
└───────────────── fNoResearch flag (0 or 1)
```

### fNoResearch Flag (offset 0)

Controls how the planet contributes to research:

| Value | GUI Label                      | Behavior                                                    |
|-------|--------------------------------|-------------------------------------------------------------|
| 0     | "Contribute to Research"       | Uses global research percentage (normal contribution)       |
| 1     | "Don't contribute to Research" | Only leftover resources after production go to research     |

When `fNoResearch=1`, the planet prioritizes production - research only receives resources remaining after the production queue has been fully processed for the year.

**Source:** Field `fNoResearch` in `ZIPPRODQ1` structure (types.h:2331)

### Item Encoding

Each item is a 16-bit little-endian value with format `(Count << 6) | ItemId`:

```
Bits 0-5:   Item ID (0-6 for auto-build items)
Bits 6-15:  Count (0-1023, max settable in GUI is 1020)
```

**IMPORTANT:** This differs from ProductionQueueBlock which uses `(ItemId << 10) | Count`. ZipProd has the fields reversed!

### Auto-Build Item IDs

| ID | Item               |
|----|--------------------|
| 0  | Auto Mines         |
| 1  | Auto Factories     |
| 2  | Auto Defenses      |
| 3  | Auto Alchemy       |
| 4  | Auto Min Terraform |
| 5  | Auto Max Terraform |
| 6  | Auto Packets       |

### Example Decoding

Raw data: `00 07 C0 02 81 4B 02 FF 43 00 04 FF C5 05 06 6F`

```
fNoResearch: 0x00 (uses global research percentage)
Item count: 7

Item 0: 0x02C0 → ID=(0x02C0 & 0x3F)=0, Count=(0x02C0 >> 6)=11  → AutoMines(11)
Item 1: 0x4B81 → ID=1, Count=302  → AutoFactories(302)
Item 2: 0xFF02 → ID=2, Count=1020 → AutoDefenses(1020)
Item 3: 0x0043 → ID=3, Count=1   → AutoAlchemy(1)
Item 4: 0xFF04 → ID=4, Count=1020 → AutoMinTerraform(1020)
Item 5: 0x05C5 → ID=5, Count=23  → AutoMaxTerraform(23)
Item 6: 0x6F06 → ID=6, Count=444 → AutoPackets(444)
```

### Notes

- **Items CAN repeat**: The same auto-build item type can appear multiple times with different counts
- **Maximum 12 items**: The queue is limited to 12 items
- Count of 1 for AutoAlchemy may indicate "enabled" since alchemy doesn't have a meaningful quantity limit

---

## Player Relations Storage

After turn generation, diplomatic relations are stored in the PlayerBlock within the player's own M file.

**Location**: In PlayerBlock, after FullDataBytes (at offset 0x70), a length-prefixed array stores relations.

**Format:**
```
LL [R0] [R1] [R2] ... [R(LL-1)]
│   └────────────────────────── Relation to player i (0=Neutral, 1=Friend, 2=Enemy)
└────────────────────────────── Length (number of entries)
```

**IMPORTANT: Different encoding from order files!**

| Value | Order File (Type 38) | M File Storage |
|-------|----------------------|----------------|
| 0     | Friend               | Neutral        |
| 1     | Neutral              | Friend         |
| 2     | Enemy                | Enemy          |

Friend and Neutral are **swapped** between order files and M file storage.

**Storage rules:**
- `PlayerRelations[i]` = relation to player `i`
- Array length varies by player - indices beyond array length default to Neutral
- Player's relation to self (own index) is stored as Neutral (0)

**Example from 3-player game:**
```
P0 (Hobbits):   set P1=Friend, P2=Neutral
  Stored: [02] [00 01] = length=2, [0]=Neutral(self), [1]=Friend(P1)
  P2 defaults to Neutral (not stored)

P1 (Halflings): set P0=Neutral, P2=Enemy
  Stored: [03] [00 00 02] = length=3, [0]=Neutral(P0), [1]=Neutral(self), [2]=Enemy(P2)

P2 (Orcs):      set P0=Friend, P1=Enemy
  Stored: [02] [01 02] = length=2, [0]=Friend(P0), [1]=Enemy(P1)
  Self defaults to Neutral (not stored)
```

---

## AI Player Configuration

In PlayerBlock (Type 6), byte 7 encodes AI settings:

```
Bit 0: Always 1
Bit 1: AI enabled (0=off, 1=on)
Bits 2-3: AI skill level
  00 = Easy
  01 = Standard
  10 = Harder
  11 = Expert
Bit 4: Always 0
Bits 5-7: Mode (flip when set to Human Inactive)
```

**Special Values**:
- AI password "viewai" = bytes [238, 171, 77, 9] (0xEEAB4D09)
- Human(Inactive) password = [255, 255, 255, 255] (bit-inverted from blank)

---

## Lesser Race Traits (LRT) Bitmask

14 traits encoded in 2 bytes at PlayerBlock offset 78-79:

| Bit   | Short | Full Name                |
|-------|-------|--------------------------|
| 0     | IFE   | Improved Fuel Efficiency |
| 1     | TT    | Total Terraforming       |
| 2     | ARM   | Advanced Remote Mining   |
| 3     | ISB   | Improved Starbases       |
| 4     | GR    | Generalised Research     |
| 5     | UR    | Ultimate Recycling       |
| 6     | MA    | Mineral Alchemy          |
| 7     | NRSE  | No Ram Scoop Engines     |
| 8     | CE    | Cheap Engines            |
| 9     | OBRM  | Only Basic Remote Mining |
| 10    | NAS   | No Advanced Scanners     |
| 11    | LSP   | Low Starting Population  |
| 12    | BET   | Bleeding Edge Technology |
| 13    | RS    | Regenerating Shields     |
| 14-15 | -     | Unused                   |

---

## Client-Side Zip Queue Storage (Stars.ini)

Custom zip queue definitions (Q2, Q3, Q4 names and contents) are stored in `Stars.ini`, typically at `C:\Windows\Stars.ini` (or under Wine: `~/.wine/drive_c/windows/Stars.ini`).

**INI Section: `[ZipOrders]`**

```ini
[ZipOrders]
ZipOrdersP1=agaeaabeaaceaaeiaafiaagiaa<Default>
ZipOrdersP2=abaajbZO1
ZipOrdersP3=abbajbZO2
ZipOrdersP4=acaeaabeaaZO3
ZipOrdersP5=
```

**Format**: `ZipOrdersP{n}=[encoded_data][QueueName]`
- `n` = Queue slot number (1-4: Default, Q2, Q3, Q4)
- `[encoded_data]` = Base-11 encoded queue items (lowercase letters a-k)
- `[QueueName]` = Queue name appended directly after encoded data

The encoding uses lowercase letters where 'a'=0, 'b'=1, ..., 'k'=10.

**Count Encoding (Base-11)**:
```
count = (high_char - 'a') × 11 + (low_char - 'a')
```

Examples:
- `aa` = 0×11 + 0 = 0 (no limit / empty)
- `ab` = 0×11 + 1 = 1
- `jb` = 9×11 + 1 = 100
