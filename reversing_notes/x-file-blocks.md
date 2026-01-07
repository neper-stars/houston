# X File Order Blocks

X files (.x1 through .x16) contain player orders submitted for turn generation. These blocks encode commands like research changes, fleet orders, diplomatic changes, etc.

## ResearchChangeBlock (Type 34) - 2 bytes

```
BB FF
│  └─ (next_field << 4) | current_field
└──── Research budget percentage (0-100)
```

Example: `0F 25` = 15% budget, current=Biotechnology(5), next=Propulsion(2)

## ProductionQueueChangeBlock (Type 29)

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

## ChangePasswordBlock (Type 36) - 4 bytes

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

## PlanetChangeBlock (Type 35) - 6 bytes

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

## PlayersRelationChangeBlock (Type 38) - 2 bytes

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

## SetFleetBattlePlanBlock (Type 42) - 4 bytes

```
FF FF PP PP
└───┘ └───┘
  │     └─── Battle plan index (16-bit LE)
  └───────── Fleet number (9 bits)
```

This order assigns a battle plan to a fleet. Battle plan index 0 is the default plan; indices 1-4 correspond to custom battle plans.

Example from test data:
- `02 00 01 00` → Fleet 2 (Long Range Scout+ #3) assigned to plan 1 (Kill Starbase)

## WaypointRepeatOrdersBlock (Type 10) - 4 bytes

```
FF FF WW XX
└───┘ │  └─── Unknown/flags
  │   └────── Waypoint index where repeat starts
  └────────── Fleet number (9 bits)
```

This order enables "Repeat Orders" for a fleet, causing it to loop back to a specified waypoint after completing its route.

Example from test data:
- `09 00 01 00` → Fleet 9 (Fleet 2) repeats from waypoint 1

## RenameFleetBlock (Type 44) - Variable length

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

## WaypointChangeTaskBlock (Type 5) - Variable length

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

### Transport Task Extension (bytes 12-19)

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

### Patrol Task Extension (byte 14)

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
