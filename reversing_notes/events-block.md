# EventsBlock (Type 12)

The EventsBlock contains a list of events that occurred during the turn. Each event has a type byte followed by type-specific data.

## Production Events (planet-specific)

| Type | Name              | Format                                     |
|------|-------------------|--------------------------------------------|
| 0x26 | Population Change | `26 00 PP PP ...`                          |
| 0x35 | Defenses Built    | `35 00 PP PP PP PP` (5 bytes)              |
| 0x36 | Factories Built   | `36 00 PP PP CC PP PP` (6 bytes, CC=count) |
| 0x37 | Mineral Alchemy   | `37 00 PP PP PP PP` (5 bytes)              |
| 0x38 | Mines Built       | `38 00 PP PP CC PP PP` (6 bytes, CC=count) |
| 0x3E | Queue Empty       | `3E 00 PP PP PP PP` (5 bytes)              |

Where `PP PP` = Planet ID (16-bit little-endian)

## Global Events (not planet-specific)

| Type | Name                       | Format                              |
|------|----------------------------|-------------------------------------|
| 0x50 | Research Complete          | `50 00 FE FF LL CF NF` (7 bytes)    |
| 0x57 | Terraformable Planet Found | `57 FF ?? ?? ?? ?? GG GG` (8 bytes) |
| 0x5F | Tech Benefit               | `5F FF CC II II XX XX` (7 bytes)    |

### Research Complete Event (0x50)

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

### Terraformable Planet Found Event (0x57)

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

### Tech Benefit Event (0x5F)

```
5F FF CC II II XX XX
│  │  │  └───┘ └───┘
│  │  │    │     └─ Extra data
│  │  │    └─────── Item ID (16-bit)
│  │  └──────────── Category
│  └─────────────── Flags
└────────────────── Event type (0x5F)
```

## Research Field IDs

| ID | Field         |
|----|---------------|
| 0  | Energy        |
| 1  | Weapons       |
| 2  | Propulsion    |
| 3  | Construction  |
| 4  | Electronics   |
| 5  | Biotechnology |

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

See [battle-block.md](battle-block.md) for the detailed battle recording format.
