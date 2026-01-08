# Fleet Blocks (Types 16 & 17)

## Overview

Fleet data is stored in two block types:
- **Type 16 (FleetBlock)**: Full fleet data for owned fleets
- **Type 17 (PartialFleetBlock)**: Scanned/observed fleet data

Both share the same header format but differ in what data follows.

## Fleet Header (6 bytes minimum)

```
Bytes 0-1: Fleet ID + Owner (packed)
   Bits 0-8:   Fleet number (0-511)
   Bits 9-12:  Owner player (0-15)
   Bits 13-15: Unused

Bytes 2-3: iPlayer (int16)
   Owner player index (redundant with bits 9-12)

Byte 4: det (detection/kind level)
   Detection level determining how much data follows

Byte 5: Flags byte
   See Flags Byte section below
```

### Fleet ID Encoding (Bytes 0-1)

```c
fleetNum = d[0] + ((d[1] & 0x01) << 8);  // 9 bits
owner = (d[1] >> 1) & 0x0F;               // 4 bits
```

## Detection Level (Byte 4)

The `det` field controls what information is included:

| Value | Level    | Description                              |
|-------|----------|------------------------------------------|
| 0     | None     | Minimal data                             |
| 1-6   | Partial  | Varying levels of scanned information    |
| 7     | Full     | Complete fleet information (owned fleet) |

## Flags Byte (0x05)

```
Byte 0x05 = 0bUUUU_DCBA
```

| Bit | Field        | Description                                      |
|-----|--------------|--------------------------------------------------|
| 0   | `fInclude`   | Include in reports/selection                     |
| 1   | `fRepOrders` | Repeat waypoint orders when complete             |
| 2   | `fDead`      | Fleet has been destroyed                         |
| 3   | `fByteCsh`   | Ship counts use 1 byte (0=2 bytes, 1=1 byte)     |
| 4-7 | (unused)     | Not persisted - always 0 in file format          |

### Bits 4-7 Analysis

In the full in-memory `FLEET` structure, bits 12-15
(corresponding to bits 4-7 of byte 0x05) contain runtime flags:

| Bit | In-Memory Flag   | Description                           |
|-----|------------------|---------------------------------------|
| 4   | `fDone`          | Fleet processing complete for turn    |
| 5   | `fBombed`        | Fleet bombed a planet this turn       |
| 6   | `fHereAllTurn`   | Fleet stayed at location all turn     |
| 7   | `fNoHeal`        | Fleet cannot heal/repair              |

**These are NOT persisted to the file format.** The `WriteFleet` function
zeroes these bits when constructing the `FLEETSOME` structure for file output.
They are recalculated each turn during game processing.

**Source:** Analysis of `WriteFleet` assembly at 1070:81c6 and `MarkFleet` at 1070:885e in stars.exe

## Variable Data Following Header

### Full Fleet (Type 16, det=7)

After the 6-byte header:

```
Bytes 6-7:   idPlanet (int16) - Planet fleet is orbiting (-1 if in space)
Bytes 8-11:  pt (POINT) - X,Y coordinates
Bytes 12+:   Ship counts and cargo data (variable length based on fByteCsh)
```

#### Ship Counts Section

A bitmask word indicates which ship design slots have ships:

```
Bytes N-N+1: grMask (uint16) - Bit i set if design slot i has ships
```

For each bit set in grMask:
- If `fByteCsh=1`: 1 byte ship count (0-255)
- If `fByteCsh=0`: 2 byte ship count (0-65535)

#### Cargo Section

If fleet has cargo capacity:
- 4 bytes: Ironium (int32)
- 4 bytes: Boranium (int32)
- 4 bytes: Germanium (int32)
- 4 bytes: Colonists (int32)
- 4 bytes: Fuel (int32)

### Partial Fleet (Type 17)

Limited information based on scan level.
Header only or with partial position data.

## Related Structures

### FLEETSOME (file format, 12 bytes)

```c
typedef struct _fleetsome {
    int16_t id;      /* +0x0000 */
    int16_t iPlayer; /* +0x0002 */
    union {
        struct {
            uint16_t det : 8;        /* +0x0004 */
            uint16_t fInclude : 1;   /* bit 8 */
            uint16_t fRepOrders : 1; /* bit 9 */
            uint16_t fDead : 1;      /* bit 10 */
            uint16_t fByteCsh : 1;   /* bit 11 */
            uint16_t unused : 4;     /* bits 12-15 - NOT PERSISTED */
        };
    }; /* +0x0004 */
    int16_t idPlanet; /* +0x0006 */
    POINT pt;         /* +0x0008 */
} FLEETSOME;
```

### FLEET (in-memory, 124 bytes)

Full runtime structure with additional fields:

```c
typedef struct _fleet {
    // ... id, iPlayer fields same as FLEETSOME ...
    union {
        struct {
            uint16_t det : 8;
            uint16_t fInclude : 1;
            uint16_t fRepOrders : 1;
            uint16_t fDead : 1;
            uint16_t fDone : 1;        /* Runtime only */
            uint16_t fBombed : 1;      /* Runtime only */
            uint16_t fHereAllTurn : 1; /* Runtime only */
            uint16_t fNoHeal : 1;      /* Runtime only */
            uint16_t fMark : 1;        /* Runtime only */
        };
    };
    int16_t idPlanet;
    POINT pt;
    int16_t rgcsh[16];     /* Ship counts per design slot */
    // ... cargo, orders, and other runtime data ...
} FLEET;
```

## References

- `types.h`: FLEET, FLEETSOME, FLEETID structures
- `file.h`: FReadFleet function
- `save.h`: WriteFleet, MarkFleet functions
- `WriteFleet` @ 1070:81c6
- `MarkFleet` @ 1070:885e
