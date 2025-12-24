# Stars! Block Structure Reversing Notes

## Event Types (in EventsBlock, Type 12)

### Production Events (planet-specific)

| Type | Name | Format |
|------|------|--------|
| 0x26 | Population Change | `26 00 PP PP ...` |
| 0x35 | Defenses Built | `35 00 PP PP PP PP` (5 bytes) |
| 0x36 | Factories Built | `36 00 PP PP CC PP PP` (6 bytes, CC=count) |
| 0x37 | Mineral Alchemy | `37 00 PP PP PP PP` (5 bytes) |
| 0x38 | Mines Built | `38 00 PP PP CC PP PP` (6 bytes, CC=count) |
| 0x3E | Queue Empty | `3E 00 PP PP PP PP` (5 bytes) |

Where `PP PP` = Planet ID (16-bit little-endian)

### Global Events (not planet-specific)

| Type | Name | Format |
|------|------|--------|
| 0x50 | Research Complete | `50 00 FE FF LL CF NF` (7 bytes) |
| 0x57 | Terraformable Planet Found | `57 FF ?? ?? ?? ?? GG GG` (8 bytes) |
| 0x5F | Tech Benefit | `5F FF CC II II XX XX` (7 bytes) |

#### Research Complete Event (0x50)

```
50 00 FE FF LL CF NF
│  │  └───┘ │  │  └─ Next research field (0-5)
│  │    │   │  └──── Completed field (0-5)
│  │    │   └─────── Level achieved (1-26)
│  │    └─────────── 0xFFFE = "no planet" (research is global)
│  └──────────────── Flags (0x00)
└────────────────────Event type (0x50)
```

**Key insight**: `0xFFFE` is NOT a fixed/magic value - it's the "no planet" marker. Production events have planet IDs at bytes 2-3; research is player-global so it uses -2/0xFFFE instead. This is consistent with Stars! event structure.

#### Terraformable Planet Found Event (0x57)

```
57 FF ?? ?? ?? ?? GG GG
│  │              └───┘
│  │                └─── Growth rate encoded (16-bit LE)
│  └──────────────────── Flags
└────────────────────── Event type (0x57)
```

**Growth rate encoding**: The last 2 bytes encode the potential growth rate after terraforming.
- Formula: `growth_percent = encoded_value / 332`
- Example: `0x0380` (896) → 896 / 332 = 2.70%

**Planet identification**: Bytes 2-5 contain planet reference data (encoding TBD). The corresponding partial planet block (Type 14) will have extended 21-byte format with terraforming potential data.

#### Tech Benefit Event (0x5F)

```
5F FF CC II II XX XX
│  │  │  └───┘ └───┘
│  │  │    │     └─ Extra data
│  │  │    └─────── Item ID (16-bit)
│  │  └──────────── Category
│  └─────────────── Flags
└────────────────── Event type (0x5F)
```

### Research Field IDs

| ID | Field |
|----|-------|
| 0 | Energy |
| 1 | Weapons |
| 2 | Propulsion |
| 3 | Construction |
| 4 | Electronics |
| 5 | Biotechnology |

---

## X File Order Blocks

### ResearchChangeBlock (Type 34) - 2 bytes

```
BB FF
│  └─ (next_field << 4) | current_field
└──── Research budget percentage (0-100)
```

Example: `0F 25` = 15% budget, current=Biotechnology(5), next=Propulsion(2)

### ProductionQueueChangeBlock (Type 29)

```
PP PP [II II CC CC] [II II CC CC] ...
└───┘ └───────────┘
  │         └─ Queue items (4 bytes each)
  └─────────── Planet ID (11 bits)
```

Each queue item (4 bytes):
```
Chunk1 (16-bit): ItemId(6 bits) | Count(10 bits)
Chunk2 (16-bit): CompletePercent(12 bits) | ItemType(4 bits)
```

Item types:
- `2` = Standard game items
- `4` = Custom ship/starbase designs

Standard item IDs:
| ID | Item |
|----|------|
| 0 | Auto Mines |
| 1 | Auto Factories |
| 2 | Auto Defenses |
| 3 | Auto Alchemy |
| 4 | Auto Min Terraform |
| 5 | Auto Max Terraform |
| 6 | Auto Packets |
| 7 | Factory |
| 8 | Mine |
| 9 | Defense |
| 11 | Mineral Alchemy |

### PlanetChangeBlock (Type 35) - 6 bytes

```
PP PP FF XX XX XX
└───┘ │  └─────┘
  │   │     └─ Additional settings (TBD)
  │   └─────── Flags byte
  └──────────── Planet ID (11 bits)
```

Flags byte (byte 2):
| Bit | Meaning |
|-----|---------|
| 7 (0x80) | Contribute only leftover resources to research |
| 0-6 | TBD |

---

## M File Blocks

### PlayerBlock (Type 6)

When `FullDataFlag` is set, `FullDataBytes` (104 bytes starting at offset 8) contains race settings:

| Offset | Size | Field |
|--------|------|-------|
| 8-16 | 9 | Habitability ranges |
| 17 | 1 | Growth rate (max population growth %, typically 1-20) |
| 18-23 | 6 | Tech levels (Energy, Weapons, Propulsion, Construction, Electronics, Biotech) |

---

## General Notes

1. **Planet ID encoding**: Usually 11 bits (0-2047), stored in first 2 bytes with other flags in upper bits

2. **"No planet" marker**: Global events (like research) use `0xFFFE` (-2 signed) where planet-specific events have planet IDs

3. **Nibble packing**: Stars! developers pack multiple small values into single bytes using nibbles (4 bits each), e.g., ResearchChangeBlock encodes two field IDs in one byte

4. **Data validation**: Rather than repeating data for validation, Stars! uses checksums. When bytes appear to repeat, they likely represent different data that happens to have the same value in test samples

---

## Object Block (Type 43)

The Object Block is a multipurpose block for map objects with several subtypes:

| ObjectType | Name | Description |
|------------|------|-------------|
| 0 | Minefield | Player-owned minefields |
| 1 | Packet/Salvage | Mineral packets and salvage |
| 2 | Wormhole | Wormholes |
| 3 | Mystery Trader | The Mystery Trader ship |

### Common Header (6 bytes)

```
OO OO XX XX YY YY
└───┘ └───┘ └───┘
  │     │     └─── Y coordinate (16-bit LE)
  │     └───────── X coordinate (16-bit LE)
  └─────────────── Object ID word
```

Object ID word breakdown:
- Bits 0-8: Object number (9 bits, 0-511)
- Bits 9-12: Owner player index (4 bits, 0-15)
- Bits 13-15: Object type (3 bits, 0-3)

### Mineral Packet (ObjectType 1) - 18 bytes

```
OO OO XX XX YY YY DD SS II II BB BB GG GG ?? ?? ?? ??
└───┘ └───┘ └───┘ │  │  └───┘ └───┘ └───┘ └─────────┘
  │     │     │   │  │    │     │     │         └─ Unknown (4 bytes)
  │     │     │   │  │    │     │     └─────────── Germanium kT (16-bit LE)
  │     │     │   │  │    │     └───────────────── Boranium kT (16-bit LE)
  │     │     │   │  │    └─────────────────────── Ironium kT (16-bit LE)
  │     │     │   │  └──────────────────────────── Speed byte
  │     │     │   └─────────────────────────────── Destination planet ID (8-bit)
  │     │     └─────────────────────────────────── Y position
  │     └───────────────────────────────────────── X position
  └─────────────────────────────────────────────── Object ID (see above)
```

#### Warp Speed Encoding (byte 7)

The speed byte encodes warp speed using the formula:

```
rawByte = (warp - 5) * 4 + 196
warp = (rawByte >> 2) - 44
```

| Warp | Raw Byte | Hex |
|------|----------|-----|
| 5 | 196 | 0xC4 |
| 6 | 200 | 0xC8 |
| 7 | 204 | 0xCC |
| 8 | 208 | 0xD0 |
| 9 | 212 | 0xD4 |
| 10 | 216 | 0xD8 |
| 11 | 220 | 0xDC |
| 12 | 224 | 0xE0 |
| 13 | 228 | 0xE4 |

**Note**: The lower 2 bits of the speed byte appear to always be 0. Upper bits may contain additional flags (TBD).

### Salvage Object (ObjectType 1, variant)

Salvage objects share the same ObjectType as mineral packets but have different field meanings for bytes 6-7:

```
OO OO XX XX YY YY ?? SF II II BB BB GG GG ?? ?? ?? ??
└───┘ └───┘ └───┘ │  │  └───┘ └───┘ └───┘ └─────────┘
  │     │     │   │  │    │     │     │         └─ Unknown (4 bytes)
  │     │     │   │  │    │     │     └─────────── Germanium kT (16-bit LE)
  │     │     │   │  │    │     └───────────────── Boranium kT (16-bit LE)
  │     │     │   │  │    └─────────────────────── Ironium kT (16-bit LE)
  │     │     │   │  └──────────────────────────── Source/Fleet byte (see below)
  │     │     │   └─────────────────────────────── Unknown (0xFF for salvage)
  │     │     └─────────────────────────────────── Y position
  │     └───────────────────────────────────────── X position
  └─────────────────────────────────────────────── Object ID (see above)
```

**Source/Fleet byte (byte 7)**: For salvage from scrapped fleets:
- Low nibble (bits 0-3): Source fleet ID (0-indexed, display is ID+1)
- High nibble (bits 4-7): Flags (0x8 observed)

**Distinguishing packets vs salvage**:
- Mineral packets: byte 6 = destination planet ID, byte 7 = warp speed
- Salvage: byte 6 = 0xFF, byte 7 = fleet source info

Example from test data:
- `00 20 15 05 42 05 FF 83 06 00 00 00 04 00 02 00 00 00`
- Position: (1301, 1346)
- Byte 7: `0x83` → Fleet ID = 3 (displayed as #4), flags = 0x8
- Minerals: 6kT Fe, 0kT Bo, 4kT Ge

---

## Additional Event Types

### Mineral Packet Produced (0xD3)

```
D3 FF SS SS DD DD
│  │  └───┘ └───┘
│  │    │     └─── Destination planet ID (16-bit LE)
│  │    └───────── Source planet ID (16-bit LE)
│  └────────────── Flags (0xFF = global event marker)
└──────────────── Event type
```

This event is generated when a mass driver launches a mineral packet.

### Packet Captured (0xD5)

```
D5 FF PP PP MM MM
│  │  └───┘ └───┘
│  │    │     └─── Mineral amount in kT (16-bit LE)
│  │    └───────── Planet ID that captured the packet (16-bit LE)
│  └────────────── Flags
└──────────────── Event type
```

This event is generated when a planet with a mass driver catches an incoming mineral packet.

### Population Change (0x26)

```
26 00 PP PP AA AA ...
│  │  └───┘ └───┘
│  │    │     └─── Amount (encoded, likely in hundreds of colonists)
│  │    └───────── Planet ID (16-bit LE)
│  └────────────── Flags
└──────────────── Event type
```

This event tracks population changes on planets (growth, decay, or transfers).

### Packet Bombardment (0xD8)

```
D8 00 PP PP XX MM MM 00 DD
│  │  └───┘ │  └───┘ │  └─── Colonists killed / 100
│  │    │   │    │   └────── Unknown (always 0x00)
│  │    │   │    └────────── Mineral amount in kT (16-bit LE)
│  │    │   └─────────────── Unknown (often same as planet low byte)
│  │    └─────────────────── Planet ID (16-bit LE)
│  └──────────────────────── Flags (0x00)
└─────────────────────────── Event type
```

This event is generated when a mineral packet hits a planet that cannot catch it (no mass driver or insufficient mass driver level).

Example from test data:
- `D8 00 BA 00 BA F0 00 00 36` → Planet 186, 240kT, 5400 colonists killed (54 × 100)
- `D8 00 BA 00 BA FC 00 00 64` → Planet 186, 252kT, 10000 colonists killed (100 × 100)

### Starbase Built (0xCD)

```
CD 00 PP PP XX DD
│  │  └───┘ │  └─── Design info (exact encoding TBD)
│  │    │   └────── Unknown (repeat of planet low byte)
│  │    └────────── Planet ID (16-bit LE)
│  └─────────────── Flags (0x00)
└────────────────── Event type
```

This event is generated when a planet finishes constructing a new starbase.

Example from test data:
- `CD 00 BA 00 BA 12` → Planet 186 (Ice Patch) built a starbase, design info = 18

---

## Random Events

Random events occur when the game has "Random Events" enabled. These include comet strikes, artifact discoveries, and other special occurrences.

### Comet Strike (0x86)

```
86 SS PP PP PP PP
│  │  └───┘ └───┘
│  │    │     └─── Planet ID repeated (16-bit LE)
│  │    └───────── Planet ID (16-bit LE)
│  └────────────── Subtype/flags (0x02 observed)
└────────────────── Event type
```

This event is generated when a comet strikes a planet. The comet embeds minerals in the planet and radically alters its environment.

Example from test data:
- `86 02 E5 01 E5 01` → Planet 485 (Burgoyne) was struck by a comet, subtype = 2

### Strange Artifact (0x5E)

```
5E 02 FE FF PP PP FF BB
│  │  └───┘ └───┘ │  └─── Boost amount (research resources)
│  │    │     │   └────── Research field (0=Energy, 1=Weapons, etc.)
│  │    │     └────────── Planet ID (16-bit LE)
│  │    └──────────────── 0xFFFE marker
│  └───────────────────── Flags (0x02 observed)
└────────────────────────Event type
```

This event is generated when colonists settling a new planet find a strange artifact that boosts research.

Example from test data:
- `5E 02 FE FF 3E 01 00 6B` → Planet 318 (Harris) found artifact, Energy (0) +107 resources

---

## Colony Events

### New Colony Established (0x1C)

```
1C 00 PP PP XX XX PP PP PP PP
│  │  └───┘ └───┘ └─────────┘
│  │    │     │        └───── Planet ID repeated twice (confirmation)
│  │    │     └────────────── Extra data (possibly fleet info)
│  │    └──────────────────── Planet ID (16-bit LE)
│  └───────────────────────── Flags (0x00)
└──────────────────────────── Event type
```

This event is generated when colonists establish a new colony on a planet.

Example from test data:
- `1C 00 3E 01 0A 02 3E 01 3E 01` → New colony on planet 318 (Harris)

---

## Fleet Events

### Fleet Scrapped (0x59)

```
59 FF PP PP FI MM
│  │  └───┘ │  └─── Mineral amount / 7 (encoded)
│  │    │   └────── Fleet index (0-based, display is +1)
│  │    └────────── Planet ID (16-bit LE) where minerals deposited
│  └─────────────── Flags (possibly design ID or subtype)
└────────────────── Event type
```

This event is generated when a fleet is scrapped/dismantled at a planet, depositing minerals.

**Mineral encoding**: The mineral amount in kT is encoded as `value / 7`. To decode: `mineralAmount = encodedByte * 7`

Example from test data:
- `59 12 3E 01 0A 04` → Fleet 10 (Santa Maria #11) scrapped at planet 318 (Harris), 28kT minerals (4 × 7)

### Fleet Scrapped in Space (0x5B)

```
5B SS LL LL OO OO
│  │  └───┘ └───┘
│  │    │     └─── Object reference (salvage object ID word)
│  │    └───────── Location marker: 0xFFFA = "in deep space"
│  └────────────── Subtype/flags (0x06 observed)
└────────────────── Event type
```

This event is generated when a fleet is scrapped/dismantled in deep space (not at a planet). The salvage is left floating in space as a salvage object.

**Key differences from 0x59 (scrapped at planet)**:
- 0x59: Minerals deposited directly at planet, encoded in event
- 0x5B: Salvage object created in space, event references the object

**Location marker**: `0xFFFA` (-6 signed) indicates "in deep space", similar to `0xFFFE` (-2) for "no planet" but distinct.

**Object reference**: The last 2 bytes match the salvage object's ID word, linking the event to the salvage object which contains:
- Position (X, Y)
- Mineral amounts (Ironium, Boranium, Germanium)
- Source fleet ID (in byte 7, low nibble)

Example from test data:
- `5B 06 FA FF 00 20` → Fleet scrapped in space, salvage object `0x2000`
- Salvage object at (1301, 1346) with 6kT Fe, 0kT Bo, 4kT Ge
- Fleet ID in salvage byte 7: `0x83` → low nibble 3 = fleet #4 ("Teamster #4")

**Game message**: "Teamster #4 has been dismantled. The scrap was left in deep space."

---

## Battle Events

### Battle Occurred (0x4F)

```
4F FF FF PP PP OO YF EF YL EL
│  │  │  └───┘ │  │  │  │  └── Enemy losses (ships lost)
│  │  │    │   │  │  │  └───── Your losses (ships lost)
│  │  │    │   │  │  └──────── Enemy forces (stacks in battle)
│  │  │    │   │  └─────────── Your forces (stacks in battle)
│  │  │    │   └────────────── Outcome byte (see below)
│  │  │    └────────────────── Planet ID (16-bit LE) where battle occurred
│  │  └──────────────────────── Unknown (0xFF observed)
│  └────────────────────────── Flags (0xFF = global event)
└───────────────────────────── Event type
```

This event is generated when a battle takes place. The battle VCR recording is stored in a separate BattleBlock (Type 31).

**Outcome byte breakdown**:
- Low nibble (bits 0-3): Enemy player index (0-15)
- High nibble (bits 4-7): Outcome flags
  - Bit 4 (0x10): Enemy survived (not completely wiped out)
  - Bit 5 (0x20): Battle recording available

**Example from test data**:
- `4F FF FF 88 01 31 04 02 01 01` → Battle at planet 392 (Redmond)
  - Planet ID: 0x0188 = 392
  - Outcome byte: 0x31
    - Enemy player: 0x1 = Player 1 (Halflings)
    - Flags: 0x3 = both bits set (enemy survived, has recording)
  - Your forces: 4
  - Enemy forces: 2
  - Your losses: 1
  - Enemy losses: 1

**Game message**: "A battle took place at Redmond against the Halflings. Neither your 4 forces nor the enemy's 2 forces were completely wiped out. You lost 1 and the enemy lost 1."

### BattleBlock (Type 31) Structure

The BattleBlock contains the battle VCR recording data with three sections:

#### Header (18 bytes)

```
Offset  Size  Field
------  ----  -----
0       1     Battle ID (usually 1)
1       1     Rounds - 1 (0-indexed, add 1 for actual round count)
2       1     Side 1 stack count
3       1     Total stack count (Side 2 = total - side1)
4-5     2     Unknown (always 3 observed)
6-7     2     Block size (16-bit LE, self-referencing)
8-9     2     Planet ID (16-bit LE)
10-11   2     X coordinate (16-bit LE)
12-13   2     Y coordinate (16-bit LE)
14      1     Attacker stack count
15      1     Defender stack count
16      1     Attacker losses
17      1     Defender losses
```

#### Stack Definitions (29 bytes each)

Each participating stack has a 29-byte definition containing:
- Ship design information
- Ship count
- Initial stats (armor, shields, initiative)

The 0x41 byte serves as a marker within the stack structure:
- Byte at marker-1: Design ID
- Byte at marker+1: Ship count
- Following bytes: Initiative, movement, and other stats

#### Action Records (22 bytes each)

Battle actions are recorded in fixed 22-byte records containing:
- Position and state data
- Damage information (0x64 = base 100, 0xE4 = variant)
- Stack references and movement
- Round numbers (0-15 for 16 rounds)

**Example structure:**
```
Total size: 854 bytes
├── Header:      18 bytes   (0x000 - 0x011)
├── Stack defs: 174 bytes   (0x012 - 0x0BF) = 6 stacks × 29 bytes
└── Actions:    660 bytes   (0x0C0 - 0x355) = 30 actions × 22 bytes
```

**Example from test data (Hobbits vs Halflings at Redmond):**
- Battle ID: 1
- Rounds: 16 (stored as 15)
- Total stacks: 6 (4 attackers + 2 defenders)
- Planet: 392 (Redmond)
- Location: (1943, 2087)
- Attacker losses: 1, Defender losses: 2

#### Battle VCR Phase Detection

The Battle VCR displays "Phase X of Y, Round Z of 16" where each phase represents one stack's turn to act. Phase markers are encoded within the 22-byte action records using this pattern:

```
[round][stack][stack][type]
   │      │      │      └── 0x04 = MOVE, 0xC4 = FIRE
   │      │      └───────── Target stack (0-5) or self
   │      └──────────────── Acting stack (0-5)
   └─────────────────────── Round number (0-15)
```

**Detection accuracy**: ~60% of phases can be detected with this pattern. Later rounds (5-15) are reliably detected; early rounds have lower detection rates.

**Damage markers** follow phase markers:
- `0x64` or `0xE4` = damage indicator
- Next byte: damage amount
- Next byte: target stack (0-5)
- Next byte: grid position

**Grid position encoding**:
```
position = 0x40 + x + y*10
```

Where x and y are 0-9 on a 10×10 grid. Position 0x40 = (0,0), position 0x99 = (9,9).

**Note**: The Y-axis may be inverted between data encoding (Y=0 at top) and UI display (Y=0 at bottom).

**Stack position updates** use `[stack][position]` pairs (2 bytes) without the full phase marker, appearing throughout the action data.

#### Limitations and Unknown Encodings

1. **Initial placement**: Round 0 may use different encoding for initial stack positions (not the standard `[round][stack][stack][type]` pattern)

2. **Fractional movement**: Stacks with movement > 1 (e.g., 1¾) get multiple phases per round. The encoding for these additional phases is not fully decoded.

3. **Complete phase count**: Test battle shows 68 phases in VCR but only ~40 are detected via pattern matching. The remaining ~28 phases use encoding patterns not yet identified.

---

## Client-Generated Messages

Some messages displayed in the Stars! client are not stored in the M file but are dynamically generated based on game state analysis.

### Packet Collision Warnings

Warning messages like "A mass packet appears to be on a collision course with [Planet], which currently is unable to safely catch the packet" are **not stored as events** in the file.

Instead, the client:
1. Reads packet objects (position, destination, warp speed)
2. Reads destination planet data (mass driver capability)
3. Calculates if the planet can safely catch packets at that speed
4. Dynamically generates warning messages for any mismatches

This reduces file size by avoiding redundant data - the warning condition can be derived from existing packet and planet information.

### Enemy Planet Discovery Messages

Messages like "You have found a planet occupied by someone else. [Planet] is currently owned by the [Race]" are **not stored as events** in the file.

Instead, the client:
1. Reads PartialPlanetBlocks from the current turn
2. Checks owner field for each planet (owner > 0 means enemy-owned)
3. Compares with previous turn data to identify newly discovered enemy planets
4. Dynamically generates discovery messages for each new sighting

Example from test data:
- 3 enemy planets (IDs 392, 411, 412) owned by Player 1 (Halflings)
- Planet IDs NOT present in EventsBlock
- Client generates 3 "found enemy planet" messages from PartialPlanetBlock data
