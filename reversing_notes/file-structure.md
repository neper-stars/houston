# File Structure

This document covers the general structure of Stars! game files.

## File Footer Data

Each Stars! file type has different footer data (this is NOT a checksum - just metadata):

| File Type | Extension | Footer Data                     |
|-----------|-----------|---------------------------------|
| M files   | .m1-.m16  | Turn number (from FileHeader)   |
| XY files  | .xy       | PlayerCount (from PlanetsBlock) |
| X files   | .x1-.x16  | None (footer size 0)            |
| H files   | .h1-.h16  | None (footer size 0)            |

The footer data is stored as a 16-bit little-endian value in the FileFooter block when present. Despite being called "checksum" in some documentation, these values are simply copies of existing metadata, not computed integrity checks.

---

## PlanetsBlock Trailing Data

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

## FileHashBlock (Type 9) - Copy Protection

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
