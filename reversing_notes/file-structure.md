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

## PlanetsBlock Structure (Type 7)

The PlanetsBlock (Type 7) has a unique structure: after the encrypted 64-byte block data, there are additional bytes for planet coordinates that are **stored unencrypted**.

```
[Block Header 2 bytes] [Block Data 64 bytes, encrypted] [Planet Data N×4 bytes, unencrypted]
```

### Block Data Layout (64 bytes, encrypted)

| Offset | Size | Field             | Description                                       |
|--------|------|-------------------|---------------------------------------------------|
| 0-3    | 4    | lid               | Game ID / serial number                           |
| 4-5    | 2    | UniverseSize      | 0=Tiny, 1=Small, 2=Medium, 3=Large, 4=Huge        |
| 6-7    | 2    | Density           | 0=Sparse, 1=Normal, 2=Dense, 3=Packed             |
| 8-9    | 2    | PlayerCount       | Number of players (1-16)                          |
| 10-11  | 2    | PlanetCount       | Total number of planets                           |
| 12-15  | 4    | StartingDistance  | Player homeworld separation                       |
| 16-17  | 2    | GameSettings      | Game options bitmask (see GameSetting* constants) |
| 18-19  | 2    | Turn              | Current turn number (0 in XY files)               |
| 20-31  | 12   | VictoryConditions | Victory condition settings (see below)            |
| 32-63  | 32   | GameName          | Game name, null-padded                            |

### Victory Conditions Array (12 bytes)

Each byte in the VictoryConditions array encodes:
- Bit 7 (0x80): Enabled flag (condition is active)
- Bits 0-6 (0x7F): Threshold index value

The `GetVCVal()` function converts the index to actual values using formulas:

| Index | Max Idx | Formula             | Value Range | Victory Condition                                   |
|-------|---------|---------------------|-------------|-----------------------------------------------------|
| 0     | 16      | `idx * 5 + 20`      | 20-100%     | Owns % of planets                                   |
| 1     | 18      | `idx + 8`           | 8-26        | Attains tech level X                                |
| 2     | 4       | `idx + 2`           | 2-6         | **in Y tech fields** (2nd value for tech condition) |
| 3     | 19      | `idx * 1000 + 1000` | 1k-20k      | Exceeds score                                       |
| 4     | 28      | `idx * 10 + 20`     | 20-300%     | Exceeds 2nd place by %                              |
| 5     | 49      | `idx * 10 + 10`     | 10-500      | Production capacity (thousands)                     |
| 6     | 29      | `idx * 10 + 10`     | 10-300      | Owns capital ships                                  |
| 7     | 87      | `idx * 10 + 30`     | 30-900      | Highest score after N years                         |
| 8     | -       | counts enabled      | 1-7         | Meets N of above criteria                           |
| 9     | 47      | `idx * 10 + 30`     | 30-500      | Minimum years before winner declared                |
| 10-11 | -       | -                   | -           | Reserved                                            |

**Note:** Index 8 uses a special formula that counts how many conditions (indices 0-7, excluding 2) have their enabled bit set, then caps the value at that count.

**Source:** Decompiled from stars26jrc3.exe:
- `vrgvcMax[]` array at 1078:b5a8
- `GetVCVal()` function at 1078:b710
- `GetVCCheck()` function at 1078:b60c

### Trailing Planet Data (unencrypted)

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
0-1     2     Unknown (possibly flags or player ID)
2-5     4     Serial number (uint32 LE) - decoded from registration string
6-16    11    Hardware fingerprint (pbEnv) - machine identification data
```

**Purpose**: Validates installation disk info to detect if a turn file was edited on a different machine. This triggers the "Copy Protection Activated When Editing an Ally's Turn File" bug.

### Serial Number Encoding

The serial number stored at bytes 2-5 is **not a hash** - it's a deterministic transformation of the user's registration string.

**Registration String Format:**
- 28 characters using: `A-Z`, `a-z`, `0-9`, `-`, `*`
- The visible portion is typically just the first 8 characters
- Each character encodes 6 bits (28 chars × 6 bits = 168 bits = 21 bytes)

**Character Encoding (6 bits each):**

| Characters | Value Range | Formula         |
|------------|-------------|-----------------|
| A-Z        | 0-25        | `ch - 'A'`      |
| a-z        | 26-51       | `ch - 'a' + 26` |
| 0-9        | 52-61       | `ch - '0' + 52` |
| `-`        | 62          | 0x3E            |
| `*`        | 63          | 0x3F            |

**Decoding Process:**
1. Base64-like decode 28 chars → 21 bytes
2. Apply shuffle permutation using `vrgbShuffleSerial[21]`
3. Extract: bytes 0-3 = `lSerial`, bytes 4-14 = `pbEnv`

**Shuffle Table** (at address 1020:2870):
```
vrgbShuffleSerial[21] = {
    0x0b, 0x04, 0x05, 0x10, 0x11, 0x0c, 0x13, 0x0f,
    0x0a, 0x01, 0x0e, 0x0d, 0x03, 0x12, 0x02, 0x14,
    0x09, 0x07, 0x00, 0x08, 0x06
};
```

The shuffle maps: `output[vrgbShuffleSerial[i]] = decoded[i]`

### Hardware Fingerprint (pbEnv)

The 11-byte hardware fingerprint at bytes 6-16 contains:

| Offset | Size | Field                             |
|--------|------|-----------------------------------|
| 0-3    | 4    | C: drive volume label             |
| 4-5    | 2    | C: drive date/time                |
| 6-8    | 3    | D: drive volume label             |
| 9      | 1    | D: drive date/time                |
| 10     | 1    | Combined drive sizes (100s of MB) |

### Key Functions

| Function              | Address   | Purpose                                   |
|-----------------------|-----------|-------------------------------------------|
| `FSerialAndEnvFromSz` | 1020:2aec | Decode serial string → (lSerial, pbEnv)   |
| `FormatSerialAndEnv`  | 1020:2886 | Encode (lSerial, pbEnv) → serial string   |
| `FValidSerialLong`    | 1070:48c4 | Validate decoded serial value             |
| `LongFromSerialCh`    | 1038:6280 | Convert single char (older 8-char format) |

**Valid Serial Types:** After dividing `lSerial` by 36^4, the type must be one of: `0x02, 0x04, 0x06, 0x12, 0x16`
