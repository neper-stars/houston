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

## Mineral Packet (ObjectType 1) - 18 bytes

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
