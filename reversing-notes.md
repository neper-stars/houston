# Stars! Block Structure Reversing Notes

## Event Types (in EventsBlock, Type 12)

### Production Events (planet-specific)

| Type | Name              | Format                                     |
|------|-------------------|--------------------------------------------|
| 0x26 | Population Change | `26 00 PP PP ...`                          |
| 0x35 | Defenses Built    | `35 00 PP PP PP PP` (5 bytes)              |
| 0x36 | Factories Built   | `36 00 PP PP CC PP PP` (6 bytes, CC=count) |
| 0x37 | Mineral Alchemy   | `37 00 PP PP PP PP` (5 bytes)              |
| 0x38 | Mines Built       | `38 00 PP PP CC PP PP` (6 bytes, CC=count) |
| 0x3E | Queue Empty       | `3E 00 PP PP PP PP` (5 bytes)              |

Where `PP PP` = Planet ID (16-bit little-endian)

### Global Events (not planet-specific)

| Type | Name                       | Format                              |
|------|----------------------------|-------------------------------------|
| 0x50 | Research Complete          | `50 00 FE FF LL CF NF` (7 bytes)    |
| 0x57 | Terraformable Planet Found | `57 FF ?? ?? ?? ?? GG GG` (8 bytes) |
| 0x5F | Tech Benefit               | `5F FF CC II II XX XX` (7 bytes)    |

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

| ID | Field         |
|----|---------------|
| 0  | Energy        |
| 1  | Weapons       |
| 2  | Propulsion    |
| 3  | Construction  |
| 4  | Electronics   |
| 5  | Biotechnology |

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
| ID | Item               |
|----|--------------------|
| 0  | Auto Mines         |
| 1  | Auto Factories     |
| 2  | Auto Defenses      |
| 3  | Auto Alchemy       |
| 4  | Auto Min Terraform |
| 5  | Auto Max Terraform |
| 6  | Auto Packets       |
| 7  | Factory            |
| 8  | Mine               |
| 9  | Defense            |
| 11 | Mineral Alchemy    |

### ChangePasswordBlock (Type 36) - 4 bytes

```
HH HH HH HH
└─────────┘
     └───── New password hash (uint32 little-endian)
```

This order changes the player's race password. The hash is computed using the Stars! password hashing algorithm (see Password System section below).

**Special values:**
- Hash = 0: Removes password (no password required)
- Hash > 0: Sets password to one that hashes to this value

**Example from test data:**
- `7A 2D 00 00` → Hash = 11642 = HashRacePassword("hob")

### PlanetChangeBlock (Type 35) - 6 bytes

```
PP PP FF XX XX XX
└───┘ │  └─────┘
  │   │     └─ Additional settings (TBD)
  │   └─────── Flags byte
  └──────────── Planet ID (11 bits)
```

Flags byte (byte 2):
| Bit      | Meaning                                        |
|----------|------------------------------------------------|
| 7 (0x80) | Contribute only leftover resources to research |
| 0-6      | TBD                                            |

### PlayersRelationChangeBlock (Type 38) - 2 bytes

```
RR PP
│  └── Target player index (0-15)
└───── Relation type
```

**Relation types:**
| Value | Relation |
|-------|----------|
| 0     | Friend   |
| 1     | Neutral  |
| 2     | Enemy    |

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
| Value | Type                            |
|-------|---------------------------------|
| 1     | Planet                          |
| 2     | Fleet                           |
| 4     | Deep Space                      |
| 8     | Wormhole/Salvage/Mystery Trader |

**Waypoint tasks:**
| Value | Task             |
|-------|------------------|
| 0     | None             |
| 1     | Transport        |
| 2     | Colonize         |
| 3     | Remote Mining    |
| 4     | Merge with Fleet |
| 5     | Scrap Fleet      |
| 6     | Lay Mines        |
| 7     | Patrol           |
| 8     | Route            |
| 9     | Transfer         |

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
| Value | Action              |
|-------|---------------------|
| 0     | No Action           |
| 1     | Load All Available  |
| 2     | Unload All          |
| 3     | Load Exactly N kT   |
| 4     | Unload Exactly N kT |
| 5     | Fill Up to N%       |
| 6     | Wait for N%         |
| 7     | Drop and Load       |
| 8     | Set Amount To N kT  |

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
| Value | Range                |
|-------|----------------------|
| 0     | 50 ly (default)      |
| 1     | 100 ly               |
| 2     | 150 ly               |
| ...   | (value + 1) × 50 ly  |
| 10    | 550 ly               |
| 11    | Any enemy (infinite) |

Example from test data:
- `09 00 01 00 5D 05 34 05 01 00 67 14 00 00 01` → Patrol range 1 = 100 ly
- `09 00 02 00 58 05 83 05 02 00 67 14 00 00 02` → Patrol range 2 = 150 ly

---

## M File Blocks

### PlayerBlock (Type 6)

When `FullDataFlag` is set, `FullDataBytes` (104 bytes starting at offset 8) contains race settings:

| Offset | Size | Field                                                                         |
|--------|------|-------------------------------------------------------------------------------|
| 8-16   | 9    | Habitability ranges                                                           |
| 17     | 1    | Growth rate (max population growth %, typically 1-20)                         |
| 18-23  | 6    | Tech levels (Energy, Weapons, Propulsion, Construction, Electronics, Biotech) |

---

## General Notes

1. **Planet ID encoding**: Usually 11 bits (0-2047), stored in first 2 bytes with other flags in upper bits

2. **"No planet" marker**: Global events (like research) use `0xFFFE` (-2 signed) where planet-specific events have planet IDs

3. **Nibble packing**: Stars! developers pack multiple small values into single bytes using nibbles (4 bits each), e.g., ResearchChangeBlock encodes two field IDs in one byte

4. **Data validation**: Rather than repeating data for validation, Stars! uses checksums. When bytes appear to repeat, they likely represent different data that happens to have the same value in test samples

---

## Object Block (Type 43)

The Object Block is a multipurpose block for map objects with several subtypes:

| ObjectType | Name           | Description                 |
|------------|----------------|-----------------------------|
| 0          | Minefield      | Player-owned minefields     |
| 1          | Packet/Salvage | Mineral packets and salvage |
| 2          | Wormhole       | Wormholes                   |
| 3          | Mystery Trader | The Mystery Trader ship     |

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

| Warp | Raw Byte | Hex  |
|------|----------|------|
| 5    | 196      | 0xC4 |
| 6    | 200      | 0xC8 |
| 7    | 204      | 0xCC |
| 8    | 208      | 0xD0 |
| 9    | 212      | 0xD4 |
| 10   | 216      | 0xD8 |
| 11   | 220      | 0xDC |
| 12   | 224      | 0xE0 |
| 13   | 228      | 0xE4 |

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

### BattleBlock (Type 31) Structure - UPDATED FROM BINARY ANALYSIS

**Source**: Decompiled from `stars26jrc3.exe` using Ghidra. Structure names from NB09 CodeView debug symbols.

The BattleBlock contains the battle VCR recording data. The structure was previously misunderstood - the header is **14 bytes** (not 18), and action records are **variable size** (not fixed 22 bytes).

#### BTLDATA Header (14 bytes)

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

#### TOK Stack Structure (29 bytes each)

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

| Offset | Size | Field      | Description                 |
|--------|------|------------|-----------------------------|
| 0x00   | 2    | id         | Fleet or planet ID          |
| 0x02   | 1    | iplr       | Owner player (0-15)         |
| 0x03   | 1    | grobj      | 1 = starbase, other = fleet |
| 0x04   | 1    | ishdef     | Ship design index           |
| 0x05   | 1    | brc        | Grid position (encoded)     |
| 0x06   | 1    | initBase   | Base initiative value       |
| 0x07   | 1    | initMin    | Minimum initiative          |
| 0x08   | 1    | initMac    | Maximum initiative          |
| 0x09   | 1    | itokTarget | Target stack index          |
| 0x0a   | 1    | pctCloak   | Cloak % (0-100)             |
| 0x0b   | 1    | pctJam     | Jammer %                    |
| 0x0c   | 1    | pctBC      | Battle computer %           |
| 0x0d   | 1    | pctCap     | Capacitor %                 |
| 0x0e   | 1    | pctBeamDef | Beam deflector %            |
| 0x0f   | 2    | wt         | Ship mass                   |
| 0x11   | 2    | dpShield   | Shield HP                   |
| 0x13   | 2    | csh        | Ship count in stack         |
| 0x15   | 2    | dv         | **VERIFIED** Damage state (DV struct) |
| 0x17   | 2    | mdTarget   | Target mode bits            |
| 0x19   | 5    | -          | Additional fields           |

#### BTLREC Action Records (VARIABLE SIZE!)

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

#### KILL Structure (8 bytes each)

Each kill event within a BTLREC:

```c
typedef struct _kill {
    uint8_t  itok;      // +0x00: Target stack index
    uint8_t  grfWeapon; // +0x01: Weapon type flags (0x01, 0x04, 0xC4 observed)
    uint16_t cshKill;   // +0x02: Number of ships killed
    uint16_t dpShield;  // +0x04: Shield damage dealt
    DV       dv;        // +0x06: Unknown purpose (NOT armor damage!)
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

#### DV (Damage Value) Structure (2 bytes) - **VERIFIED FROM DECOMPILATION**

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

#### Battle Damage Formulas - **VERIFIED FROM DECOMPILATION**

**Source**: Decompiled from `stars26jrc3.exe` - functions `FDamageTok`, `RegenShield`, `CTorpHit`, `DpFromPtokBrcToBrc`

##### Shield Damage (FDamageTok @ 10f0:81d4)

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

##### Armor Damage (FDamageTok @ 10f0:81d4)

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

##### Shield Regeneration (RegenShield @ 10f0:3c16)

Shields regenerate 10% per battle round:
```c
int maxShield = DpShieldOfShdef(shdef, player);
int regen = maxShield / 10;  // 10% regeneration
dpShield = min(dpShield + regen, maxShield);
```

##### Torpedo Hit Calculation (CTorpHit @ 10f0:6790)

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

##### Beam Damage with Range (DpFromPtokBrcToBrc @ 10f0:4d2e)

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

#### Block Continuation (Type 39)

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

#### Grid Position Encoding

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

#### Round Calculation

The header does NOT reliably store round count. Calculate from action data:
1. Scan BTLREC records for the `iRound` field (bits 0-3 of offset 0x04)
2. Find maximum round number
3. Rounds = max_round + 1

#### Example Battle Layout

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
| Offset | Pattern       | Interpretation                       |
|--------|---------------|--------------------------------------|
| 0x176  | `0E 00 64 04` | 14 shield + 4 armor (torpedo hit)    |
| 0x1C0  | `10 00 64 04` | 16 shield + 4 armor (laser hit)      |
| 0x1B4  | `64 78 01 46` | Cumulative 120 dmg, Stack 1 at (6,4) |
| 0x23C  | `64 B6 01 47` | Cumulative 182 dmg, Stack 1 at (6,5) |

**Damage markers**:
- `0x64` = armor damage marker (or cumulative state)
- `0xE4` = alternative marker (possibly torpedo-specific)

**Grid position encoding**:
```
position = col × 11 + row
```

Where col and row are 0-9 on a 10×10 grid. The base-11 encoding provides unique values for all grid cells.

**Confirmed position mappings (from battle-02 analysis):**
| Encoded | Decimal | Grid Position | Description                  |
|---------|---------|---------------|------------------------------|
| 0x52    | 82      | (7, 5)        | Stalwart Defender initial    |
| 0x51    | 81      | (7, 4)        | Stalwart Defender final      |
| 0x46    | 70      | (6, 4)        | Stalwart Defender mid-battle |
| 0x47    | 71      | (6, 5)        | Stalwart Defender moving     |
| 0x25    | 37      | (3, 4)        | Cruiser phase 4              |
| 0x3A    | 58      | (5, 3)        | Cruiser phase 28             |

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

| File Type | Extension | Footer Data                     |
|-----------|-----------|---------------------------------|
| M files   | .m1-.m16  | Turn number (from FileHeader)   |
| XY files  | .xy       | PlayerCount (from PlanetsBlock) |
| X files   | .x1-.x16  | None (footer size 0)            |
| H files   | .h1-.h16  | None (footer size 0)            |

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

---

## Password System

Stars! uses a weak 32-bit hash for race passwords. The algorithm is trivially reversible through brute force, and many collisions exist.

### Password Hash Algorithm

```go
func HashRacePassword(password string) uint32 {
    bytes := []byte(password)
    hash := uint32(bytes[0])  // Start with first character

    for i, b := range bytes[1:] {
        if i % 2 == 0 {
            hash = (hash * uint32(b)) & 0xFFFFFFFF  // Odd positions: multiply
        } else {
            hash = (hash + uint32(b)) & 0xFFFFFFFF  // Even positions: add
        }
    }
    return hash
}
```

**Algorithm breakdown:**
1. Initialize hash with ASCII value of first character
2. For each subsequent character at 1-based position `p`:
   - If `p` is odd (2nd, 4th, 6th...): multiply hash by character value
   - If `p` is even (3rd, 5th, 7th...): add character value to hash
3. All operations are modulo 2³² (32-bit overflow)

**Example:**
```
Password: "hob"
  h = 104
  hash = 104
  o (position 1, odd): hash = 104 * 111 = 11544
  b (position 2, even): hash = 11544 + 98 = 11642
Result: 11642 (0x00002D7A)
```

### Hash Weaknesses

The algorithm is extremely weak:
1. **32-bit output**: Only 4 billion possible hashes
2. **Multiplicative structure**: Creates many collisions
3. **No salt**: Same password always produces same hash
4. **Sequential dependency**: Short passwords have small hash space

**Collision example for hash 11642 ("hob"):**
- "hob" (original)
- "awc" (found by brute force)
- Many others exist

### Password Storage

**In PlayerBlock (M files):**
- Offset 12-15 within block data
- 4 bytes, uint32 little-endian
- Hash = 0 means no password set

**In ChangePasswordBlock (X files):**
- See Type 36 documentation above
- 4 bytes, uint32 little-endian
- Hash = 0 removes the password

### Race File Password

Race files (.r1-.r16) also store password hashes in the PlayerBlock at the same offset (bytes 12-15). Modifying the password requires recalculating the race file integrity hash (see "Race File Integrity Hash" section below).

### Brute Force Performance

With parallel implementation on modern hardware:
- 5-character alphanumeric (36^5 = 60M combinations): < 1 second
- 6-character alphanumeric (36^6 = 2B combinations): ~30 seconds
- Due to collisions, valid alternative passwords are typically found quickly

---

## Race File Checksum (SOLVED!)

Race files (.r1-.r16) have a 16-bit checksum in the FileFooter that validates the race data.

### Race File Structure

```
[FileHeader]     16 bytes (Type 8) - contains salt value
[PlayerBlock]    Variable (Type 6, encrypted) - contains race data
[FileFooter]     4 bytes (Type 0) - 2-byte header + 2-byte checksum
```

### Encryption Parameters (Race Files)

Race files use specific encryption parameters:
- Salt: From FileHeader
- Game ID: 0
- Turn: 0
- Player Index: 31
- Offset: 0

### Checksum Algorithm

The checksum is computed from decrypted PlayerBlock data plus interleaved race names:

```go
func ComputeRaceFooter(decryptedData []byte, singularName, pluralName string) uint16 {
    // Find where the name data starts
    fullDataFlag := (decryptedData[6] & 0x04) != 0
    index := 8
    if fullDataFlag {
        index = 0x70 // 112 bytes: 8 header + 0x68 (104) full data
        playerRelationsLength := int(decryptedData[index])
        index += 1 + playerRelationsLength
    }

    // Data length is everything before the names section
    dataLength := index

    // Build the checksum data array
    var dData []byte
    dData = append(dData, decryptedData[:dataLength]...)

    // Prepare singular name: leading 0, ASCII bytes, padded to 16 total
    singularOrd := make([]byte, 16)
    singularOrd[0] = 0
    for i, c := range singularName {
        if i < 15 {
            singularOrd[i+1] = byte(c)
        }
    }

    // Prepare plural name: leading 0, ASCII bytes, padded to 16 total
    pluralOrd := make([]byte, 16)
    pluralOrd[0] = 0
    for i, c := range pluralName {
        if i < 15 {
            pluralOrd[i+1] = byte(c)
        }
    }

    // Interleave: add pairs from singular, then pairs from plural
    for i := 0; i < 16; i += 2 {
        dData = append(dData, singularOrd[i], singularOrd[i+1])
        dData = append(dData, pluralOrd[i], pluralOrd[i+1])
    }

    // Compute checksums
    var checkSum1, checkSum2 byte
    for i := 0; i < len(dData); i += 2 {
        checkSum1 ^= dData[i]
    }
    for i := 1; i < len(dData); i += 2 {
        checkSum2 ^= dData[i]
    }

    return uint16(checkSum1) | uint16(checkSum2)<<8
}
```

**Algorithm steps:**
1. Take decrypted PlayerBlock data up to (but not including) the nibble-packed names
2. Decode singular and plural race names to ASCII
3. Pad each name to 15 characters with a leading 0 (16 bytes total)
4. Interleave name bytes in pairs: singular[0:2], plural[0:2], singular[2:4], plural[2:4], ...
5. XOR all even-indexed bytes → checkSum1
6. XOR all odd-indexed bytes → checkSum2
7. Return `(checkSum2 << 8) | checkSum1`

### Password Location

Password hash is stored at bytes 12-15 of decrypted PlayerBlock data:
- Hash = 0x00000000: No password
- Hash > 0: Password set (see Password System section above)

### Password Removal

To remove a password from a race file:
1. Decrypt the PlayerBlock using race file encryption parameters
2. Zero out bytes 12-15 (password hash)
3. Parse the race names from decrypted data
4. Re-encrypt the modified PlayerBlock
5. Recalculate the footer checksum using the algorithm above
6. Update the FileFooter with the new checksum

**Implementation:** `houston race-password <file>` command and `racefixer.RemovePasswordBytes()` function.

### Testing

Verified against 39 race files in `testdata/scenario-racefiles/`:
- All .r1 and .r2 files
- Files with and without passwords
- Various race names (short, long, special characters)
- Different race settings (PRT, LRT, habitat, etc.)

### FileHashBlock (Type 9) - Copy Protection

X files contain a FileHashBlock (Type 9) with 17 bytes of copy protection data.

```
Offset  Size  Field
------  ----  -----
0-1     2     Unknown (always 0x001E observed)
2-5     4     Serial number hash (uint32 LE)
6-9     4     C: drive volume label hash (uint32 LE)
10-11   2     C: drive date/time hash
12-13   2     D: drive volume label hash
14-15   2     D: drive date/time hash
16      1     C: and D: drive size in 100's of MB
```

**Purpose**: Validates installation disk info to detect if a turn file was edited on a different machine. This triggers the "Copy Protection Activated When Editing an Ally's Turn File" bug.

---

## MessageBlock (Type 40)

Player-to-player messages are stored in MessageBlocks.

```
Offset  Size  Field
------  ----  -----
0-1     2     Unknown
2-3     2     Unknown
4-5     2     Sender ID (16-bit LE)
6-7     2     Recipient ID (16-bit LE, 0 = "Everyone")
8-9     2     Unknown
10-11   2     Message byte count
12+     Var   Stars! encoded message string
```

**Notes**:
- HST files do not contain message blocks
- In .x files, sender is always the file's player
- Player IDs in messages are offset: 0 = Everyone, 1-16 = Players 1-16

---

## Mystery Trader Items

The Mystery Trader (ObjectType 3 in Block 43) can offer 13 different items, encoded as a bitmask:

| Bit         | Item                     |
|-------------|--------------------------|
| 0 (value=0) | Research (initial state) |
| 0           | Multi Cargo Pod          |
| 1           | Multi Function Pod       |
| 2           | Langston Shield          |
| 3           | Mega Poly Shell          |
| 4           | Alien Miner              |
| 5           | Hush-a-Boom              |
| 6           | Anti Matter Torpedo      |
| 7           | Multi Contained Munition |
| 8           | Mini Morph               |
| 9           | Enigma Pulsar            |
| 10          | Genesis Device           |
| 11          | Jump Gate                |
| 12          | Ship/MT Lifeboat         |

---

## Known Exploits and Bugs

these exploits can be detected and fixed:

### Cheap Colonizer
**Detection**: Design block (Type 26/27) where a slot has `itemId=0, itemCategory=4096 (colonization), itemCount=0`
**Cause**: Colonization module removed but slot left as "empty colonization" instead of truly empty
**Effect**: Ship can colonize without colonization module
**Fix**: Set itemCategory to 0 for the slot

### Space Dock Armor Overflow
**Detection**: Space Dock (hullId=33) with ISB+RS race traits, >21 SuperLatanium in armor slot, armor value >= 49518
**Cause**: Buffer overflow when calculating armor
**Effect**: Massively increased armor values
**Fix**: Cap SuperLatanium count at 21, recalculate armor

### 10th Starbase Bug
**Detection**: Last player in game has a starbase in design slot 10 (0-indexed: slot 9)
**Cause**: Unknown internal Stars! limitation
**Effect**: Game crash if Player 1's Fleet 1 refuels at this starbase
**Fix**: Cannot be fixed automatically, only warned

### Friendly Fire Battle Plan
**Detection**: Battle Plan block (Type 30) where `planNumber == 0` (Default) and `attackWho > 3`
**Cause**: Default battle plan set to attack specific player instead of enemy/neutral
**Effect**: Own ships fire on each other
**Fix**: Reset attackWho to 2 (Neutral/Enemy)

### SS Pop Steal
**Detection**: Waypoint Change block (Type 5) with transport task targeting own population
**Cause**: Exploit in transport order handling
**Effect**: Population theft via transport orders
**Fix**: Reset transport task to do nothing

### 32k Merge Bug
**Detection**: Fleet merge where combined ship count > 32767 for any design slot
**Cause**: 16-bit signed integer overflow
**Effect**: Data corruption, potential crash
**Fix**: Cancel merge by resetting task to "No Action"

### Mineral Upload Exploit
**Detection**: Manual load/unload block transferring minerals from own planet to enemy fleet, exceeding fleet cargo capacity
**Cause**: Improper validation of upload target ownership
**Effect**: Free resource generation
**Fix**: Cancel the order

### Cheap Starbase
**Detection**: Starbase design change (Type 27) for a starbase under construction (completePercent > 0)
**Cause**: Editing starbase design while partially built
**Effect**: Resources/items duplicated or exploited
**Fix**: Reset starbase design slots to empty

---

## Zip Production Queue (Player Block offset 0x56)

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
FF NN [II II] [II II] ... [padding]
│  │  └─────────────────┘
│  │           └─── Items (2 bytes each, up to 7 items)
│  └────────────── Item count (0-7)
└───────────────── Flags byte (purpose TBD, usually 0x00)
```

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
Flags: 0x00
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

- **Items CAN repeat**: The same auto-build item type can appear multiple times with different counts (e.g., AutoMines(1) followed by AutoMines(2))
- **Maximum 12 items**: The queue is limited to 12 items. In the GUI, zip queues are populated by "importing" from a planet's actual production queue, and only the first 12 items are imported
- The "Contribute only leftover resources to research" checkbox state location is TBD (possibly in Flags byte or elsewhere in player data)
- Count of 1 for AutoAlchemy may indicate "enabled" since alchemy doesn't have a meaningful quantity limit
- Multiple SaveAndSubmit blocks may appear in X files, potentially for different queue slots or sequential updates

### Client-Side Storage (Stars.ini)

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
- `[QueueName]` = Queue name appended directly after encoded data (no separator)
  - Default queue uses angle brackets: `<Default>`
  - Custom queues use plain text: `ZO1`, `zZoO`, etc. (can contain any characters including lowercase)

**Encoded data length** is determined by the header:
```
length = 2 + (item_count × 4)
```
- Header (2 chars) + 4 chars per item
- Example: `ab` (1 item) → 2 + 4 = 6 chars, then queue name follows

#### Encoding Format

The encoding uses lowercase letters where 'a'=0, 'b'=1, ..., 'k'=10.

**Header**: `a[item_count_char]`
- `ab` = 1 item ('b'=1)
- `ac` = 2 items ('c'=2)
- `ag` = 7 items ('g'=6, possibly 0-indexed or special case for Default)

**Item Encoding** (varies by queue size):

*Single-item queues*: `[type_char]a[count_high][count_low]`
- Type: 'a'=AutoMines, 'b'=AutoFactories, 'c'=AutoDefenses, etc.
- Count: 2 chars in base-11

*Multi-item queues*: `[type_char]e[count_high][count_low]` per item
- Type: same as above
- Flag 'e' instead of 'a' for multi-item queues
- Count: 2 chars in base-11

**Count Encoding (Base-11)**:
```
count = (high_char - 'a') × 11 + (low_char - 'a')
```

Examples:
- `aa` = 0×11 + 0 = 0 (no limit / empty)
- `ab` = 0×11 + 1 = 1
- `jb` = 9×11 + 1 = 100
- `ba` = 1×11 + 0 = 11

#### Detailed Examples

**Example 1: `abaajbzZoO`** (1 item: AutoMines(100), name "zZoO")
```
ab  aajb  zZoO
^^  ^^^^  ^^^^
│   │     └── Queue name (everything after encoded data)
│   └── Item 1: type 'a'=AutoMines, flag 'a', count 'jb'=100
└── Header: 'a' prefix + 'b'=1 item
```
Length: 2 + (1 × 4) = 6 chars → `abaajb`, then `zZoO` is the name

**Example 2: `acaeaabeaaZO3`** (2 items: AutoMines(0), AutoFactories(0), name "ZO3")
```
ac  aeaa  beaa  ZO3
^^  ^^^^  ^^^^  ^^^
│   │     │     └── Queue name
│   │     └── Item 2: type 'b'=AutoFactories, flag 'e', count 'aa'=0
│   └── Item 1: type 'a'=AutoMines, flag 'e', count 'aa'=0
└── Header: 'a' prefix + 'c'=2 items
```
Length: 2 + (2 × 4) = 10 chars → `acaeaabeaa`, then `ZO3` is the name

#### Summary Table

| INI Value    | Decoded                                 |
|--------------|-----------------------------------------|
| `abaajb`     | 1 item: AutoMines(100)                  |
| `abbajb`     | 1 item: AutoFactories(100)              |
| `acaeaabeaa` | 2 items: AutoMines(0), AutoFactories(0) |

#### Notes

- The Default queue (`<Default>`) encoding may differ slightly from custom queues
- Empty entries (`ZipOrdersP3=`) indicate no custom queue defined for that slot
- Data is only saved when the game client exits
- This explains why custom queue names persist between sessions despite not being in the game files

---

## Player Flags (Player Block offset 0x54)

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

## Serial Number Validation

Stars! serial numbers use base-36 encoding (A-Z = 0-25, 0-9 = 26-35).

**Format**: 8 characters, e.g., "SAH62J1E"

**Valid Series Letters** (first character after processing):
- S (18), W (22), C (2), E (4), G (6)

**Valid Number Range**: 100 to 1,500,000

**Character Position Processing**:
- Positions 0, 1, 4, 7, 3 contribute to series/number
- Positions 2, 5, 6 are checksum digits
- XOR with 0x15 applied for values < 0x20

---

## Complete Block Type List

| ID | Type                            | Notes                                           |
|----|---------------------------------|-------------------------------------------------|
| 0  | FileFooterBlock                 | Year (.M/.HST), Checksum XOR (.R), null (.X/.H) |
| 1  | ManualSmallLoadUnloadTaskBlock  |                                                 |
| 2  | ManualMediumLoadUnloadTaskBlock |                                                 |
| 3  | WaypointDeleteBlock             |                                                 |
| 4  | WaypointAddBlock                |                                                 |
| 5  | WaypointChangeTaskBlock         |                                                 |
| 6  | PlayerBlock                     |                                                 |
| 7  | PlanetsBlock                    |                                                 |
| 8  | FileHeaderBlock                 | Unencrypted                                     |
| 9  | FileHashBlock                   | Copy protection                                 |
| 10 | WaypointRepeatOrdersBlock       |                                                 |
| 11 | Unknown                         | Never observed                                  |
| 12 | EventsBlock                     |                                                 |
| 13 | PlanetBlock                     | Full planet data                                |
| 14 | PartialPlanetBlock              | Scanned planet data                             |
| 15 | Unknown                         | Never observed                                  |
| 16 | FleetBlock                      | Full fleet data                                 |
| 17 | PartialFleetBlock               | Scanned fleet data                              |
| 18 | Unknown                         | Never observed                                  |
| 19 | WaypointTaskBlock               | In .M/.HST files                                |
| 20 | WaypointBlock                   | In .M/.HST files                                |
| 21 | FleetNameBlock                  | Custom fleet names                              |
| 22 | Unknown                         | Never observed                                  |
| 23 | MoveShipsBlock                  |                                                 |
| 24 | FleetSplitBlock                 |                                                 |
| 25 | ManualLargeLoadUnloadTaskBlock  |                                                 |
| 26 | DesignBlock                     | Ship/starbase design                            |
| 27 | DesignChangeBlock               | Design modification in .X                       |
| 28 | ProductionQueueBlock            |                                                 |
| 29 | ProductionQueueChangeBlock      |                                                 |
| 30 | BattlePlanBlock                 |                                                 |
| 31 | BattleBlock                     | VCR data (partially decoded)                    |
| 32 | CountersBlock                   | Game counters                                   |
| 33 | MessagesFilterBlock             | Message filtering prefs                         |
| 34 | ResearchChangeBlock             |                                                 |
| 35 | PlanetChangeBlock               |                                                 |
| 36 | ChangePasswordBlock             | .X files / Password in .HST                     |
| 37 | FleetsMergeBlock                |                                                 |
| 38 | PlayersRelationChangeBlock      |                                                 |
| 39 | BattleContinuationBlock         | Extended battle data (not decoded)              |
| 40 | MessageBlock                    | Player messages                                 |
| 41 | AIRecordBlock                   | AI records in .H files (not decoded)            |
| 42 | SetFleetBattlePlanBlock         |                                                 |
| 43 | ObjectBlock                     | Minefields, packets, wormholes, MT              |
| 44 | RenameFleetBlock                |                                                 |
| 45 | PlayerScoresBlock               | Victory condition tracking                      |
| 46 | SaveAndSubmitBlock              | Turn submission                                 |

---

## Design Slot Item Categories

In DesignBlock (Type 26) and DesignChangeBlock (Type 27), each slot is encoded as 4 bytes:

```
CC CC II NN
└───┘ │  └── Count (number of items)
  │   └───── ItemId (0-indexed within category)
  └───────── Category (16-bit LE, item type)
```

The Category field indicates the **type of item equipped** in the slot, not the hull's slot definition:

| Value  | Category    | Description                                                    |
|--------|-------------|----------------------------------------------------------------|
| 0x0000 | Empty       | No item equipped                                               |
| 0x0001 | Engine      | Engines (Trans-Star 10, NHRS, Galaxy, etc.)                    |
| 0x0002 | Scanner     | Ship scanners (Bat, Rhino, Mole, Possum, etc.)                 |
| 0x0004 | Shield      | Shields (Mole-skin, Cow-hide, Bear, Gorilla, etc.)             |
| 0x0008 | Armor       | Armor (Tritanium, Kelarium, Neutronium, etc.)                  |
| 0x0010 | BeamWeapon  | Beam weapons (Laser, X-Ray, Phaser, etc.)                      |
| 0x0020 | Torpedo     | Torpedoes (Alpha, Beta, Delta, Omega, etc.)                    |
| 0x0040 | Bomb        | Bombs (LadyFinger, M-70, Smart, etc.)                          |
| 0x0080 | MiningRobot | Mining robots (Midget, Mini, Maxi, etc.)                       |
| 0x0100 | MineLayer   | Mine layers (Mine40, Heavy50, Speed20, etc.)                   |
| 0x0200 | Orbital     | Orbital devices (Stargates, Mass Drivers)                      |
| 0x0400 | Planetary   | Planetary scanners (Viewer, Scoper, Snooper)                   |
| 0x0800 | Electrical  | Electrical devices (Cloaks, Jammers, Capacitors)               |
| 0x1000 | Mechanical  | Mechanical devices (Cargo Pod, Fuel Tank, Colonization Module) |

### ItemId Indexing

ItemId is **0-indexed** within each category. To convert to game constants (which are typically 1-indexed), add 1:

```
scannerConstant = slot.ItemId + 1
```

**Example**: A slot with `Category=0x0002, ItemId=1, Count=1` means:
- Category 0x0002 = Scanner
- ItemId 1 → Scanner constant 2 = Rhino Scanner
- Count 1 = One scanner equipped

### Scanner Detection Example

```go
for _, slot := range design.Slots {
    if slot.Category == ItemCategoryScanner && slot.Count > 0 {
        scannerID := slot.ItemId + 1  // Convert 0-indexed to 1-indexed
        stats := ShipScannerStats[scannerID]
        // Use stats.NormalRange, stats.PenetratingRange
    }
}
```

---

## Block Type Mappings from Binary Analysis

From decompiling the Stars! binary (stars26jrc3.exe), the following block type mappings were confirmed:

### Writing Functions

Blocks are written using two key functions:
- `WriteMemRt(rt, cb, data)` - Writes to memory log buffer (for X files, player orders)
- `WriteRt(rt, cb, data)` - Writes directly to file stream (for M/H files)

Block header format: `(size & 0x3FF) | (type << 10)` (10-bit size + 6-bit type)

### Confirmed Block Type to Function Mappings

| Type      | Block Name                 | Writing Function                | Purpose                             |
|-----------|----------------------------|---------------------------------|-------------------------------------|
| 1         | ManualSmallLoadUnloadTask  | LogMakeValidXfer (qty < 0x80)   | Small cargo transfer (signed byte)  |
| 2         | ManualMediumLoadUnloadTask | LogMakeValidXfer (qty < 0x8000) | Medium cargo transfer (signed word) |
| 3         | WaypointDelete             | LogChangeFleet                  | Remove waypoint from fleet          |
| 4         | WaypointAdd                | LogChangeFleet                  | Add new waypoint to fleet           |
| 5         | WaypointChangeTask         | LogChangeFleet                  | Modify task at existing waypoint    |
| 23 (0x17) | MoveShips                  | LogMakeValidXferf               | Transfer ships between fleets       |
| 24 (0x18) | FleetSplit                 | LogSplitFleet                   | Split a fleet into two              |
| 25 (0x19) | ManualLargeLoadUnloadTask  | LogMakeValidXfer (large)        | Large cargo transfer (32-bit)       |
| 27 (0x1b) | DesignChange               | LogChangeShDef                  | Ship/starbase design changes        |
| 29 (0x1d) | ProductionQueueChange      | LogChangePlanet                 | Production queue orders             |
| 35 (0x23) | PlanetChange               | LogChangePlanet                 | Planet setting changes              |
| 37 (0x25) | FleetsMerge                | LogMergeFleet                   | Merge multiple fleets               |
| 38 (0x26) | PlayersRelationChange      | LogChangeRelations              | Diplomatic relation changes         |
| 44 (0x2c) | RenameFleet                | LogChangeName                   | Fleet renaming                      |

### Unknown Block Types

Block types 11, 15, 18, and 22 were not found in the standard game operation functions. Possible explanations:
- Used only in specific edge cases (AI host file records, tutorial mode)
- Legacy/deprecated types no longer written by modern game versions
- Used only in M-file read operations but never written to X files

### Key Structures

From the NB09 debug symbols in the binary:

```c
// RTSHIPINT - Simple fleet/waypoint reference (4 bytes)
typedef struct _rtshipint {
    int16_t id;   // Fleet ID
    int16_t i;    // Index (e.g., waypoint index)
} RTSHIPINT;

// RTSHIPINT2 - Extended fleet/waypoint reference (6 bytes)
typedef struct _rtshipint2 {
    int16_t id;   // Fleet ID
    int16_t i;    // Index 1
    int16_t i2;   // Index 2
} RTSHIPINT2;

// LOGXFER - Cargo transfer log entry (24 bytes)
typedef struct _logxfer {
    int16_t id;           // Target ID
    int16_t grobj;        // Object type (planet=1, fleet=2)
    int32_t rgdItem[5];   // Quantities: Ironium, Boranium, Germanium, Colonists, Fuel
} LOGXFER;

// RTCHGNAME - Rename record (variable size)
typedef struct _rtchgname {
    int16_t grobj;        // Object type
    int16_t id;           // Object ID
    uint8_t rgb[1];       // Compressed name data
} RTCHGNAME;
```

