# Planet Blocks (Types 13 & 14)

## Overview

Planet data is stored in two block types:
- **Type 13 (PlanetBlock)**: Full planet data for owned/colonized planets
- **Type 14 (PartialPlanetBlock)**: Scanned/observed planet data

Both share the same header format (RTPLANET) but differ in what data follows.

## RTPLANET Header (4 bytes)

From decompiled code (types.h:1687-1708):

```
Bytes 0-1: Planet ID + Owner
   Bits 0-10:  Planet ID (0-2047)
   Bits 11-15: Owner player (0-15 for player, 31 = no owner)

Bytes 2-3: Flags word
   Bits 0-6:   det (detection level, 7 bits)
   Bit 7:      fHomeworld
   Bit 8:      fInclude
   Bit 9:      fStarbase
   Bit 10:     fIncEVO (Include original environment values = terraformed)
   Bit 11:     fIncImp (Include improvements/installations)
   Bit 12:     fIsArtifact
   Bit 13:     fIncSurfMin (Include surface minerals)
   Bit 14:     fRouting (Has fleet route set)
   Bit 15:     fFirstYear (First year planet is visible to player)
```

**Source:** `RTPLANET` structure in types.h and `IO::WritePlanet` function

## Detection Level (det) Field

The `det` field (bits 0-6) controls what information is visible:

| Value | Level       | Description                                       |
|-------|-------------|---------------------------------------------------|
| 1     | Pen Scan    | Basic visibility (planet exists, maybe position)  |
| 2     | Special     | Used for special cases (avoids starbase updates)  |
| 3     | Normal Scan | Standard scan - can see starbase, some details    |
| 4+    | Full        | Owner can see all planet details                  |
| 7     | Maximum     | Complete information (used in ship part defaults) |

The value increases as more information becomes available. Higher values include all lower-level information.

**Source:** `IO::MarkPlanet` function and WritePlanet logic

## Flag Descriptions

| Flag        | Bit | Mask   | Description                                               |
|-------------|-----|--------|-----------------------------------------------------------|
| fHomeworld  | 7   | 0x0080 | This is a player's homeworld                              |
| fInclude    | 8   | 0x0100 | Planet is included in scans/reports                       |
| fStarbase   | 9   | 0x0200 | Planet has a starbase                                     |
| fIncEVO     | 10  | 0x0400 | Original environment values included (terraformed planet) |
| fIncImp     | 11  | 0x0800 | Installations data (8 bytes) is included                  |
| fIsArtifact | 12  | 0x1000 | Planet has an ancient artifact                            |
| fIncSurfMin | 13  | 0x2000 | Surface minerals data is included                         |
| fRouting    | 14  | 0x4000 | Fleet routing destination is set                          |
| fFirstYear  | 15  | 0x8000 | First year this planet is visible to the player           |

## Environment Section

If `det >= 2`, environment data follows the header:

```
Byte 4:     Length encoding for fractional mineral concentrations
            Length = 1 + (bits 0-1) + (bits 2-3) + (bits 4-5)
Bytes 5+:   Variable-length mineral concentration data
            Then: 3 bytes mineral concentrations (Ir, Bo, Ge)
            Then: 3 bytes environment (Gravity, Temp, Radiation)
            If fIncEVO: 3 bytes original environment
            If owned: 2 bytes population/defense estimates
```

## Surface Minerals Section

If `fIncSurfMin` is set:

```
Byte N:     Length encoding (2 bits each for Ir, Bo, Ge, Pop)
            Each 2-bit field: 0=skip, 1=1 byte, 2=2 bytes, 3=4 bytes
Bytes N+1+: Variable-length values for each resource
```

## Installations Section (8 bytes)

If `fIncImp` is set, 8 bytes are copied directly from PLANET offset 0x14:

```
Bytes 0-3 (32-bit packed):
   Bits 0-7:    iDeltaPop (population change indicator)
   Bits 8-19:   cMines (12-bit mine count, 0-4095)
   Bits 20-31:  cFactories (12-bit factory count, 0-4095)

Bytes 4-7 (32-bit packed):
   Bits 0-11:   cDefenses (12-bit defense count, 0-4095)
   Bits 12-16:  iScanner (5-bit planetary scanner ID, 0-31)
   Bits 17-21:  unused5 (5 bits, always 0)
   Bit 22:      fArtifact (has artifact - also in header)
   Bit 23:      fNoResearch (don't contribute to research)
   Bits 24-31:  unused2 (8 bits, always 0)
```

### iScanner Values

The `iScanner` field (5 bits) indicates the planetary scanner installed:

| ID        | Scanner Name                    |
|-----------|---------------------------------|
| 0         | None                            |
| 1-N       | Various planetary scanner types |
| 31 (0x1F) | No scanner (special value)      |

### fNoResearch Flag

When set, the planet does not contribute to research using the global percentage. Instead, only leftover resources after production are contributed.

**Source:** Analysis from player-block.md and `ZIPPRODQ1.fNoResearch`

## Starbase Section

If `fStarbase` is set:
- **Type 13 (full planet)**: 4 bytes of starbase data
  - Byte 0 low nibble: Starbase design index (0-15)
  - Byte 0 high nibble: Damage info
  - Byte 2: Mass driver destination planet ID
  - Other bytes: Additional starbase info
- **Type 14 (partial planet)**: 1 byte with design index only

## Route Section

If `fRouting` is set (Type 13 only):
- 2 bytes: Route destination planet ID

## Turn Number

Last 2 bytes (if present): Turn number when planet was last seen
