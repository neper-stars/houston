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

### PlayersRelationChangeBlock (Type 38) - 2 bytes

```
RR PP
│  └── Target player index (0-15)
└───── Relation type
```

**Relation types:**
| Value | Relation |
|-------|----------|
| 0 | Friend |
| 1 | Neutral |
| 2 | Enemy |

This order sets the diplomatic relation with another player. Diplomatic relations are one-way: Player A can consider Player B a friend while Player B considers Player A an enemy.

**Example from test data:**
- `00 01` → Set player 1 (Halflings) as Friend
- `02 00` → Set player 0 (Hobbits) as Enemy

### Player Relations Storage (in PlayerBlock, M files)

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
|-------|---------------------|----------------|
| 0 | Friend | Neutral |
| 1 | Neutral | Friend |
| 2 | Enemy | Enemy |

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

### SetFleetBattlePlanBlock (Type 42) - 4 bytes

```
FF FF PP PP
└───┘ └───┘
  │     └─── Battle plan index (16-bit LE)
  └───────── Fleet number (9 bits)
```

This order assigns a battle plan to a fleet. Battle plan index 0 is the default plan; indices 1-4 correspond to custom battle plans.

Example from test data:
- `02 00 01 00` → Fleet 2 (Long Range Scout+ #3) assigned to plan 1 (Kill Starbase)

### WaypointRepeatOrdersBlock (Type 10) - 4 bytes

```
FF FF WW XX
└───┘ │  └─── Unknown/flags
  │   └────── Waypoint index where repeat starts
  └────────── Fleet number (9 bits)
```

This order enables "Repeat Orders" for a fleet, causing it to loop back to a specified waypoint after completing its route.

Example from test data:
- `09 00 01 00` → Fleet 9 (Fleet 2) repeats from waypoint 1

### RenameFleetBlock (Type 44) - Variable length

```
FF FF UU UU LL [encoded name...]
└───┘ └───┘ │  └───────────────── Stars! encoded string (LL bytes)
  │     │   └──────────────────── Name length
  │     └──────────────────────── Unknown (often 0x0002)
  └────────────────────────────── Fleet number (16-bit LE)
```

This order renames a fleet from its default name to a custom name.

**Name encoding**: Uses Stars! string format where the first byte is the length, followed by the encoded characters using the standard Stars! character compression.

**Example from test data:**
- `00 00 02 00 06 C2 D5 7D EA AE 2F` → Fleet 0 renamed to "Scoutty"
  - Fleet number: 0
  - Unknown: 2
  - Name length: 6 (but encodes to 7-char "Scoutty" due to compression)
  - Encoded name: `C2 D5 7D EA AE 2F`

### FleetNameBlock (Type 21) - Variable length (M files)

```
LL [encoded name bytes...]
│  └───────────────────────── Stars! encoded string (LL bytes)
└──────────────────────────── Name length
```

This block appears in M files **only for fleets with custom names**. It immediately precedes the FleetBlock whose name it contains (positional association - no fleet number in the block itself).

**Important:** If no FleetNameBlock precedes a FleetBlock, the game auto-generates the fleet name (e.g., "Long Range Scout #1", "Armed Probe #2") based on the ship design and fleet number.

**Example from test data (results/game.m1):**
- `06 C2 D5 7D EA AE 2F` → "Scoutty"
  - Appears before the FleetBlock for the renamed fleet
  - Other fleets without custom names have no preceding FleetNameBlock

**Relationship between Type 44 and Type 21:**
- Type 44 (RenameFleetBlock): Order in X file to rename a fleet
- Type 21 (FleetNameBlock): Stored custom name in M file after turn generation

### WaypointChangeTaskBlock (Type 5) - Variable length

Base format (12 bytes):
```
FF FF WW XX XX XX YY YY TT TT WK TY
└───┘ │  └─────┘ └─────┘ └───┘ │  └── Target type (low nibble)
  │   │     │       │      │   └───── Warp (high nibble) + Task (low nibble)
  │   │     │       │      └───────── Target ID (9 bits)
  │   │     │       └──────────────── Y coordinate (16-bit LE)
  │   │     └──────────────────────── X coordinate (16-bit LE)
  │   └────────────────────────────── Waypoint number
  └────────────────────────────────── Fleet number (9 bits)
```

**Target types:**
| Value | Type |
|-------|------|
| 1 | Planet |
| 2 | Fleet |
| 4 | Deep Space |
| 8 | Wormhole/Salvage/Mystery Trader |

**Waypoint tasks:**
| Value | Task |
|-------|------|
| 0 | None |
| 1 | Transport |
| 2 | Colonize |
| 3 | Remote Mining |
| 4 | Merge with Fleet |
| 5 | Scrap Fleet |
| 6 | Lay Mines |
| 7 | Patrol |
| 8 | Route |
| 9 | Transfer |

**Special warp value:**
- 11 = Use Stargate

#### Transport Task Extension (bytes 12-19)

When task = 1 (Transport), additional bytes encode cargo orders:

```
For each cargo type (Ironium, Boranium, Germanium, Colonists):
  VV AA
  │  └── Action (high nibble): action << 4
  └───── Value (amount in kT or percentage)
```

**Transport actions:**
| Value | Action |
|-------|--------|
| 0 | No Action |
| 1 | Load All Available |
| 2 | Unload All |
| 3 | Load Exactly N kT |
| 4 | Unload Exactly N kT |
| 5 | Fill Up to N% |
| 6 | Wait for N% |
| 7 | Drop and Load |
| 8 | Set Amount To N kT |

Example from test data:
- `03 00 01 00 0B 05 4F 05 01 20 51 18 12 30 00 10 32 50`
- Fleet 3 (Large Freighter #4), waypoint 1 at (1291, 1359)
- Task: Transport to salvage
- Ironium: `12 30` → value=18, action=3 → "Load Exactly 18 kT"
- Boranium: `00 10` → value=0, action=1 → "Load All Available"
- Germanium: `32 50` → value=50, action=5 → "Fill Up to 50%"

#### Patrol Task Extension (byte 14)

When task = 7 (Patrol), byte 14 encodes the intercept range:

```
... RR
    └── Patrol range (0-11)
```

**Patrol range values:**
| Value | Range |
|-------|-------|
| 0 | 50 ly (default) |
| 1 | 100 ly |
| 2 | 150 ly |
| ... | (value + 1) × 50 ly |
| 10 | 550 ly |
| 11 | Any enemy (infinite) |

Example from test data:
- `09 00 01 00 5D 05 34 05 01 00 67 14 00 00 01` → Patrol range 1 = 100 ly
- `09 00 02 00 58 05 83 05 02 00 67 14 00 00 02` → Patrol range 2 = 150 ly

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

### Fleet Scrapped at Starbase (0x5A)

```
5A FF PP PP FI MM
│  │  └───┘ │  └─── Fleet mass / 7 (encoded)
│  │    │   └────── Fleet index (0-based, display is +1)
│  │    └────────── Planet ID (16-bit LE) with starbase
│  └─────────────── Flags (0x02 observed)
└────────────────── Event type
```

This event is generated when a fleet is dismantled at a starbase. The encoding is identical to 0x59 (scrapped at planet) but has a key semantic difference:

**Mineral recovery difference**:
- 0x59 (at planet): Encoded value = minerals recovered (100% recovery)
- 0x5A (at starbase): Encoded value = fleet mass, minerals recovered ≈ 55% of mass

The Stars! game applies a ~55% recovery rate when scrapping at a starbase versus 100% at a planet. The event stores the total fleet mass, and the client calculates the actual minerals recovered for display.

**Decoding formula**: `fleetMass = encodedByte * 7`

Example from test data:
- `5A 02 69 00 0B 06` → Fleet 11 (#12) scrapped at planet 105 (Hurl) starbase, fleet mass 42kT (6 × 7)
- Message shows: "23kT recovered" = 42kT × 54.8% recovery rate

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
1       1     Rounds field (see note below)
2       1     Side 1 stack count
3       1     Total stack count (Side 2 = total - side1)
4-5     2     Unknown (always 0x0003 observed)
6-7     2     Block size (16-bit LE, self-referencing)
8-9     2     Planet ID (16-bit LE)
10-11   2     X coordinate (16-bit LE)
12-13   2     Y coordinate (16-bit LE)
14      1     Attacker stack count
15      1     Defender stack count
16      1     Attacker losses (matches game display)
17      1     Unknown17 (NOT defender losses - see analysis below)
```

**IMPORTANT: Header Field Analysis (battle-01 vs battle-02)**

| Field | Battle-01 | Battle-02 | Notes |
|-------|-----------|-----------|-------|
| byte[1] (Rounds) | 15 → 16 rounds | 0 → 1 round? | Action data shows 10 rounds! |
| byte[16] (AttackerLosses) | 1 | 0 | Matches game messages ✓ |
| byte[17] (Unknown17) | 2 | 2 | NOT defender losses (both show 1 in game) |

**byte[1] interpretation issue:**
- Battle-01: byte[1]=15, action data has rounds 0-15 (16 rounds) ✓
- Battle-02: byte[1]=0, but action data has rounds 0-9 (10 rounds!) ✗

The naive interpretation `Rounds = byte[1] + 1` only works for some battles. Battle-02 (a "continuation" battle where ships carried damage from a previous turn) has byte[1]=0 but clearly 10 rounds of action data.

**Solution implemented**: The parser now calculates rounds by scanning action data for phase markers `[round][stack1][stack2][0x04|0xC4]` and finding the maximum round number. This gives accurate results for both battles:
- Battle-01: max round 15 → 16 rounds ✓
- Battle-02: max round 9 → 10 rounds ✓

The header byte[1] is unreliable and should not be trusted for round count.

**byte[17] is NOT defender losses:**
Both battles have byte[17]=2, but actual defender losses from game messages:
- Battle-01: 1 enemy ship lost (game shows "losing one of their own")
- Battle-02: 1 enemy ship lost (Stalwart Defender destroyed at phase 61)

The true defender losses are stored in BattleEvent (Type 0x4F in EventsBlock) as `enemyLosses`, not in the BattleBlock header.

#### Stack Definitions (29 bytes each)

Each participating stack has a 29-byte definition containing:
- Ship design information
- Ship count
- Initial stats (armor, shields, initiative)

The 0x41 or 0x58 byte serves as a marker within the stack structure:
- Byte at marker-1: Design ID
- Byte at marker+1: Initiative value
- Following bytes: Ship count, movement, and other stats

**Stack mapping from battle-02 (Hobbits vs Halflings):**
| Stack | Design | Marker | Init | Ship | Armor | Shields |
|-------|--------|--------|------|------|-------|---------|
| 0 | 0x0A | 0x41 | 6 | Cruiser | 975 | 100 |
| 1 | 0x04 | 0x58 | 4 | Stalwart Defender | 275 | 0 |
| 2 | 0x08 | 0x41 | 0 | Super-Fuel Xport | 12 | 0 |
| 3 | 0x01 | 0x41 | 0 | Long Range Scout | 20 | 0 |

Note: Marker 0x58 appears on enemy stack (Halflings), while 0x41 appears on friendly stacks (Hobbits). This may indicate player ownership.

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

**Damage encoding** (from battle-02 analysis):

Two damage formats have been identified:

1. **Shield + Armor damage** (when hitting target with shields):
```
[shield_dmg] 00 64 [armor_dmg] ...
```
- Shield damage values observed: 14 (torpedo), 16 (X-Ray Laser), 2 (deflection)
- The 16 matches X-Ray Laser power rating exactly

2. **Cumulative damage state**:
```
64 [cumulative] [stack] [position]
```
- Large values (91, 120, 153, 182, 211, 240) track total damage to a stack
- Progression shows ~29 damage per laser salvo (2 X-Ray Lasers × ~14.5)

**Observed patterns from battle-02:**
| Offset | Pattern | Interpretation |
|--------|---------|----------------|
| 0x176 | `0E 00 64 04` | 14 shield + 4 armor (torpedo hit) |
| 0x1C0 | `10 00 64 04` | 16 shield + 4 armor (laser hit) |
| 0x1B4 | `64 78 01 46` | Cumulative 120 dmg, Stack 1 at (6,4) |
| 0x23C | `64 B6 01 47` | Cumulative 182 dmg, Stack 1 at (6,5) |

**Damage markers**:
- `0x64` = armor damage marker (or cumulative state)
- `0xE4` = alternative marker (possibly torpedo-specific)

**Grid position encoding**:
```
position = col × 11 + row
```

Where col and row are 0-9 on a 10×10 grid. The base-11 encoding provides unique values for all grid cells.

**Confirmed position mappings (from battle-02 analysis):**
| Encoded | Decimal | Grid Position | Description |
|---------|---------|---------------|-------------|
| 0x52 | 82 | (7, 5) | Stalwart Defender initial |
| 0x51 | 81 | (7, 4) | Stalwart Defender final |
| 0x46 | 70 | (6, 4) | Stalwart Defender mid-battle |
| 0x47 | 71 | (6, 5) | Stalwart Defender moving |
| 0x25 | 37 | (3, 4) | Cruiser phase 4 |
| 0x3A | 58 | (5, 3) | Cruiser phase 28 |

**Decoding formula**: `col = position / 11`, `row = position % 11`

**Stack position updates** use `[stack][position]` pairs (2 bytes) without the full phase marker, appearing throughout the action data.

#### Limitations and Unknown Encodings

1. **Initial placement**: Round 0 may use different encoding for initial stack positions (not the standard `[round][stack][stack][type]` pattern)

2. **Fractional movement**: Stacks with movement > 1 (e.g., 1¾) get multiple phases per round. The encoding for these additional phases is not fully decoded.

3. **Complete phase count**: Test battle shows 68 phases in VCR but only ~40 are detected via pattern matching. The remaining ~28 phases use encoding patterns not yet identified.

### BattleContinuationBlock (Type 39) - Investigation

**Status**: Not yet observed in test data.

**Theory 1: Battle data overflow**
Stars! battles are limited to 16 rounds. With a maximum of ~256 phases possible (16 rounds × 16 stacks × high movement), the action data could exceed what fits in a single BattleBlock. Type 39 may be used to store overflow action records when the 22-byte action data exceeds block size limits.

**Theory 2: Multi-turn battle continuation**
When ships survive a battle and fight again on the next turn with carried damage, this could be flagged differently. However, analysis of battle-02 (a continuation of battle-01) shows:
- No Type 39 block present
- Uses standard Type 31 BattleBlock
- Same BattleID (1) as battle-01
- byte[1] = 0 (unlike battle-01's byte[1] = 15)

**Findings from test data:**
- Neither battle-01 nor battle-02 contain Type 39 blocks
- Both fit in single Type 31 blocks (854 and 754 bytes respectively)
- "Continuation" in the game sense (ships with prior damage) doesn't require Type 39

**Likely conclusion**: Type 39 is for literal data continuation when a battle's action records don't fit in a single block, not for game-level battle continuation between turns. More test data needed with very long battles (many stacks, high movement ships, 16 rounds) to observe Type 39 in use.

---

## File Structure

### File Footer Data

Each Stars! file type has different footer data (this is NOT a checksum - just metadata):

| File Type | Extension | Footer Data |
|-----------|-----------|-------------|
| M files | .m1-.m16 | Turn number (from FileHeader) |
| XY files | .xy | PlayerCount (from PlanetsBlock) |
| X files | .x1-.x16 | None (footer size 0) |
| H files | .h1-.h16 | None (footer size 0) |

The footer data is stored as a 16-bit little-endian value in the FileFooter block when present. Despite being called "checksum" in some documentation, these values are simply copies of existing metadata, not computed integrity checks.

### PlanetsBlock Trailing Data

The PlanetsBlock (Type 7) has a unique structure: after the encrypted 64-byte block data, there are additional bytes for planet coordinates that are **stored unencrypted**.

```
[Block Header 2 bytes] [Block Data 64 bytes, encrypted] [Planet Data N×4 bytes, unencrypted]
```

- Block data (64 bytes): Contains universe settings, player count, planet count, game name - **encrypted**
- Trailing planet data (4 bytes per planet): Contains packed planet coordinates and name IDs - **unencrypted**

Each planet entry (4 bytes, little-endian uint32):
```
Bits 31-22 (10 bits): Planet name ID (index into planet names table)
Bits 21-10 (12 bits): Y coordinate (absolute)
Bits  9-0  (10 bits): X offset from previous planet (first planet uses base 1000)
```

This is the only known case where data following an encrypted block is stored unencrypted.

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
