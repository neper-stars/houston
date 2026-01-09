# ObjectBlock (Type 43)

The Object Block is a multipurpose block for map objects with several subtypes:

| ObjectType | Name           | Description                 |
|------------|----------------|-----------------------------|
| 0          | Minefield      | Player-owned minefields     |
| 1          | Packet/Salvage | Mineral packets and salvage |
| 2          | Wormhole       | Wormholes                   |
| 3          | Mystery Trader | The Mystery Trader ship     |

## Common Header (6 bytes)

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

## Minefield (ObjectType 0) - 18 bytes

```
OO OO XX XX YY YY CC CC CC CC VV VV TT DD NN NN TN TN
└───┘ └───┘ └───┘ └─────────┘ └───┘ │  │  └───┘ └───┘
  │     │     │         │       │   │  │    │     └─ Turn number (16-bit LE)
  │     │     │         │       │   │  │    └─────── grbitPlrNow: current visibility mask
  │     │     │         │       │   │  └──────────── fDetonate: detonating flag
  │     │     │         │       │   └─────────────── iType: mine type (0-2)
  │     │     │         │       └─────────────────── grbitPlr: visibility mask (16-bit LE)
  │     │     │         └─────────────────────────── cMines: mine count (32-bit LE)
  │     │     └───────────────────────────────────── Y position
  │     └─────────────────────────────────────────── X position
  └───────────────────────────────────────────────── Object ID (see above)
```

### Minefield Fields

| Offset | Size | Field         | Original Name  | Description                           |
|--------|------|---------------|----------------|---------------------------------------|
| 6-9    | 4    | MineCount     | `cMines`       | Number of mines in minefield          |
| 10-11  | 2    | CanSeeBits    | `grbitPlr`     | Player visibility bitmask (historical)|
| 12     | 1    | MinefieldType | `iType`        | 0=standard, 1=heavy, 2=speed bump     |
| 13     | 1    | Detonating    | `fDetonate`    | 1 if minefield is detonating          |
| 14-15  | 2    | CurrentSeeBits| `grbitPlrNow`  | Current turn visibility bitmask       |
| 16-17  | 2    | TurnNumber    | `turn`         | Turn when last updated                |

**grbitPlr vs grbitPlrNow**: `grbitPlr` tracks which players have ever detected this minefield.
`grbitPlrNow` tracks which players can currently see it this turn (may differ due to scanner range changes).

---

## Mineral Packet (ObjectType 1) - 18 bytes

```
OO OO XX XX YY YY DD SS II II BB BB GG GG WD WD TN TN
└───┘ └───┘ └───┘ │  │  └───┘ └───┘ └───┘ └───┘ └───┘
  │     │     │   │  │    │     │     │     │     └─ Turn number (16-bit LE)
  │     │     │   │  │    │     │     │     └─────── wtMax|iDecayRate (see below)
  │     │     │   │  │    │     │     └─────────────  Germanium kT (16-bit LE)
  │     │     │   │  │    │     └───────────────────  Boranium kT (16-bit LE)
  │     │     │   │  │    └─────────────────────────  Ironium kT (16-bit LE)
  │     │     │   │  └──────────────────────────────  Speed byte
  │     │     │   └─────────────────────────────────  Destination planet ID (8-bit)
  │     │     └─────────────────────────────────────  Y position
  │     └───────────────────────────────────────────  X position
  └─────────────────────────────────────────────────  Object ID (see above)
```

### Bytes 14-15: wtMax | iDecayRate

This 16-bit word contains two fields from the `THPACK` structure:

```
Bits 0-13:  wtMax (14 bits)      - Maximum weight/capacity in kT (0-16383)
Bits 14-15: iDecayRate (2 bits)  - Decay rate index (0-3)
```

The decay rate likely controls how fast the packet loses minerals over time or distance.

### Warp Speed Encoding (byte 7)

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

## Salvage Object (ObjectType 1, variant)

Salvage objects share the same ObjectType as mineral packets but have different field meanings for bytes 6-7:

```
OO OO XX XX YY YY FF SF II II BB BB GG GG WD WD TN TN
└───┘ └───┘ └───┘ │  │  └───┘ └───┘ └───┘ └───┘ └───┘
  │     │     │   │  │    │     │     │     │     └─ Turn number (16-bit LE)
  │     │     │   │  │    │     │     │     └─────── wtMax|iDecayRate (same as packet)
  │     │     │   │  │    │     │     └─────────────  Germanium kT (16-bit LE)
  │     │     │   │  │    │     └───────────────────  Boranium kT (16-bit LE)
  │     │     │   │  │    └─────────────────────────  Ironium kT (16-bit LE)
  │     │     │   │  └──────────────────────────────  Source/Fleet byte (see below)
  │     │     │   └─────────────────────────────────  0xFF marker (salvage indicator)
  │     │     └─────────────────────────────────────  Y position
  │     └───────────────────────────────────────────  X position
  └─────────────────────────────────────────────────  Object ID (see above)
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

## Wormhole (ObjectType 2) - 18 bytes

```
OO OO XX XX YY YY SM SM VV VV TT TT PP PP -- -- TN TN
└───┘ └───┘ └───┘ └───┘ └───┘ └───┘ └───┘ └───┘ └───┘
  │     │     │     │     │     │     │     │     └─ Turn number (16-bit LE)
  │     │     │     │     │     │     │     └─────── Padding (unused)
  │     │     │     │     │     │     └─────────────  idPartner: target wormhole ID
  │     │     │     │     │     └───────────────────  grbitPlrTrav: traversal mask
  │     │     │     │     └─────────────────────────  grbitPlr: visibility mask
  │     │     │     └───────────────────────────────  Stability/movement word (see below)
  │     │     └─────────────────────────────────────  Y position
  │     └───────────────────────────────────────────  X position
  └─────────────────────────────────────────────────  Object ID (see above)
```

### Wormhole Fields

| Offset | Size | Field           | Original Name   | Description                        |
|--------|------|-----------------|-----------------|------------------------------------|
| 6-7    | 2    | StabilityWord   | (bitfield)      | See breakdown below                |
| 8-9    | 2    | CanSeeBits      | `grbitPlr`      | Player visibility bitmask          |
| 10-11  | 2    | BeenThroughBits | `grbitPlrTrav`  | Players who have traversed         |
| 12-13  | 2    | TargetId        | `idPartner`     | Partner wormhole ID                |
| 14-15  | 2    | (padding)       | -               | Unused (THWORM is only 8 bytes)    |
| 16-17  | 2    | TurnNumber      | `turn`          | Turn when last updated             |

### Bytes 6-7: Stability/Movement Word

This 16-bit word from the `THWORM` structure:

```
Bits 0-1:   iStable (2 bits)     - Stability index (0-3)
Bits 2-11:  cLastMove (10 bits)  - Turns since last movement (0-1023)
Bit 12:     fDestKnown (1 bit)   - Destination known to players
Bit 13:     fInclude (1 bit)     - Include in display flag
Bits 14-15: (unused)
```

**Note**: Houston currently treats byte 6 as a raw "Stability" value with thresholds (32, 40, 60, etc.).
The actual encoding may combine iStable with cLastMove to compute displayed stability.

### Bytes 14-15: Padding

The `THWORM` structure is only 8 bytes, but the THING union allocates 10 bytes (sized to largest member).
Bytes 14-15 are unused padding for wormholes and should be preserved as-is for round-trip encoding.

---

## Mystery Trader (ObjectType 3) - 18 bytes

```
OO OO XX XX YY YY DX DX DY DY WW -- MM MM II II TN TN
└───┘ └───┘ └───┘ └───┘ └───┘ │  │  └───┘ └───┘ └───┘
  │     │     │     │     │   │  │    │     │     └─ Turn number (16-bit LE)
  │     │     │     │     │   │  │    │     └─────── grbitTrader: item bits
  │     │     │     │     │   │  │    └─────────────  grbitPlr: met player mask
  │     │     │     │     │   │  └──────────────────  (unused byte)
  │     │     │     │     │   └─────────────────────  iWarp: warp speed (4 bits)
  │     │     │     │     └─────────────────────────  ptDest.y: destination Y
  │     │     │     └───────────────────────────────  ptDest.x: destination X
  │     │     └─────────────────────────────────────  Y position
  │     └───────────────────────────────────────────  X position
  └─────────────────────────────────────────────────  Object ID (see above)
```

### Mystery Trader Fields

| Offset | Size | Field    | Original Name   | Description                    |
|--------|------|----------|-----------------|--------------------------------|
| 6-7    | 2    | XDest    | `ptDest.x`      | Destination X coordinate       |
| 8-9    | 2    | YDest    | `ptDest.y`      | Destination Y coordinate       |
| 10     | 1    | Warp     | `iWarp`         | Warp speed (low 4 bits only)   |
| 11     | 1    | (unused) | -               | Unused byte                    |
| 12-13  | 2    | MetBits  | `grbitPlr`      | Bitmask of players trader met  |
| 14-15  | 2    | ItemBits | `grbitTrader`   | Bitmask of items carried       |
| 16-17  | 2    | TurnNo   | `turn`          | Turn number                    |

### Item Bits (grbitTrader)

| Bit | Value  | Item                     |
|-----|--------|--------------------------|
| 0   | 0x0001 | Multi Cargo Pod          |
| 1   | 0x0002 | Multi Function Pod       |
| 2   | 0x0004 | Langston Shield          |
| 3   | 0x0008 | Mega Poly Shell          |
| 4   | 0x0010 | Alien Miner              |
| 5   | 0x0020 | Hush-a-Boom              |
| 6   | 0x0040 | Anti Matter Torpedo      |
| 7   | 0x0080 | Multi Contained Munition |
| 8   | 0x0100 | Mini Morph               |
| 9   | 0x0200 | Enigma Pulsar            |
| 10  | 0x0400 | Genesis Device           |
| 11  | 0x0800 | Jump Gate                |
| 12  | 0x1000 | Ship/MT Lifeboat         |

**Note**: A value of 0 for ItemBits means the trader is offering Research.

---

## Source

Structures from decompiled `THING`, `THMINE`, `THPACK`, `THWORM`, and `THTRADER` in stars26jrc3.exe.
