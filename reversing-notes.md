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
