# FileHeader Block (Type 8)

The FileHeader block is the first block in all Stars! files (M, X, HST, H, Race). It contains 16 bytes of metadata including game ID, version, turn number, and various flags.

## Structure (16 bytes)

From decompiled code (RTBOF in types.h):

```
Bytes 0-3:   Magic number ("J3D1" or "J3J3")
Bytes 4-7:   Game ID (32-bit)
Bytes 8-9:   Version data (encoded)
Bytes 10-11: Turn number
Bytes 12-13: Player data (salt + player index)
Byte 14:     File type (dt)
Byte 15:     Flags
```

## Version Data (bytes 8-9)

```
Bits 0-4:   Increment (5 bits, 0-31)
Bits 5-11:  Minor version (7 bits, 0-127)
Bits 12-15: Major version (4 bits, 0-15)
```

Stars! 2.60j RC4 reports version 2.83.0.

## Player Data (bytes 12-13)

```
Bits 0-4:   Player index (5 bits, 0-31)
            - 0-15 for players
            - 31 for race files
Bits 5-15:  Salt (11 bits, 0-2047) - used for encryption
```

## File Type (byte 14)

| Value | Extension | Description                   | Houston Constant |
|-------|-----------|-------------------------------|------------------|
| 0     | .xy       | Universe definition file      | FileTypeXY       |
| 1     | .x#       | Turn order file (submitted)   | FileTypeX        |
| 2     | .hst      | Host file                     | FileTypeHST      |
| 3     | .m#       | Player turn file              | FileTypeM        |
| 4     | .h#       | History file                  | FileTypeH        |
| 5     | .r#       | Race file                     | FileTypeRace     |

**Source:** `dt` field in RTBOF structure

**Notes:**
- XY files use dt=0 (previously thought to be "Unknown")
- XY files also use playerIndex=31 (same as race files)

## Flags Byte (byte 15)

| Bit | Mask | Field        | Houston Name     | Description                              |
|-----|------|--------------|------------------|------------------------------------------|
| 0   | 0x01 | fDone        | TurnSubmitted    | Turn has been submitted                  |
| 1   | 0x02 | fInUse       | HostUsing        | File is currently in use by host         |
| 2   | 0x04 | fMulti       | MultipleTurns    | Multiplayer game / multiple turns        |
| 3   | 0x08 | fGameOverMan | GameOver         | Game has ended                           |
| 4   | 0x10 | fCrippled    | Shareware (*)    | Crippled/demo mode                       |
| 5-7 | 0xE0 | wGen         | (Not documented) | Generation counter (3 bits, 0-7)         |

**(*) Note:** Houston labels bit 4 as "Shareware" but the decompiled source uses `fCrippled`.

## wGen (Generation Counter)

The `wGen` field (bits 5-7) is a 3-bit generation counter (values 0-7). It tracks file versioning for consistency checks between the host and players.

### Storage Location

In the GAME structure, `wGen` lives in bits 9-11 of `game.wCrap` (offset 0x10):
```c
typedef struct _game {
    // ...
    union {
        uint16_t wCrap;
        struct {
            uint16_t fExtraFuel : 1;   // bit 0
            uint16_t fSlowTech : 1;    // bit 1
            uint16_t fSinglePlr : 1;   // bit 2
            uint16_t fTutorial : 1;    // bit 3
            uint16_t fAisBand : 1;     // bit 4
            uint16_t fBBSPlay : 1;     // bit 5
            uint16_t fVisScores : 1;   // bit 6
            uint16_t fNoRandom : 1;    // bit 7
            uint16_t fClumping : 1;    // bit 8
            uint16_t wGen : 3;         // bits 9-11 ← HERE
            uint16_t unused : 4;       // bits 12-15
        };
    };  /* +0x0010 */
} GAME;
```

When writing files, `wGen` is extracted as `game.wCrap >> 9` and placed in bits 5-7 of byte 15.

### Validation Behavior

**Critical for generating files:** wGen validation depends on file type!

From `io_loadgame.c`:
```c
if (((dt_00 != 1) || (((uint)_DATA::game.wCrap >> 3 & 1) != 0)) ||
   ((uint)rtbof.flags8 >> 0xd == ((uint)_DATA::game.wCrap >> 9 & 7)))
  goto success;
FileError(0x1d);  // Error: wGen mismatch
```

| File Type | dt | wGen Validated? | Notes |
|-----------|----|-----------------|-----------------------------------------|
| X file    | 1  | **YES**         | Must match game state when host loads   |
| HST file  | 2  | No              | dt != 1, bypasses check                 |
| M file    | 3  | No              | dt != 1, bypasses check                 |

**For M files:** Since `dt != 1` is always TRUE, Stars! jumps to success immediately without checking wGen. **Any value 0-7 is accepted.**

**For X files:** When the host loads player orders, wGen from the X file MUST match the wGen in the host's game state. This prevents players from submitting orders for a different turn.

### Practical Guidance for File Generation

#### M Files (dt=3)
When houston generates an M file:
- **wGen can be any value (0-7)** - Stars! does not validate it for M files
- For consistency, consider:
  - Using 0 for fresh/new files
  - Copying wGen from the corresponding HST file if available
  - The value doesn't affect gameplay

#### X Files (dt=1) - CRITICAL
When houston generates an X file (player orders):
- **wGen MUST match the M file wGen for the same turn**
- The host validates wGen when loading X files
- Mismatched wGen causes FileError(0x1d) - file rejection

**Solution:** Read wGen from the player's M file and copy it directly to the X file header.

```go
// Example: Generating X file
mFileHeader := readFileHeader("game.m1")
xFileHeader.wGen = mFileHeader.wGen  // Copy directly!
xFileHeader.Turn = mFileHeader.Turn  // Same turn
```

**Why this works:** The client workflow is:
1. Player loads M file → wGen stored in game state
2. Player makes orders
3. Player submits X file → same wGen written to header
4. Host loads X file → validates wGen matches its game state

Since both client and host loaded the same M file (or host generated it), wGen will match.

### Observed Values

Sample wGen values from test files (same game, sequential turns):

| Year | Flags (byte 15) | wGen |
|------|-----------------|------|
| 2400 | 0x00            | 0    |
| 2401 | 0xA0            | 5    |
| 2402 | 0x40            | 2    |
| 2403 | 0x40            | 2    |
| 2410 | 0x40            | 2    |
| 2420 | 0x20            | 1    |
| 2430 | 0xA0            | 5    |

The pattern is not simply `turn % 8` - it appears to be based on internal game state that changes during turn processing.

## RTBOF Structure (from types.h)

```c
typedef struct _rtbof {
    char rgid[4];           /* +0x0000  Magic number */
    int32_t lidGame;        /* +0x0004  Game ID */
    uint16_t wVersion;      /* +0x0008  Version data */
    uint16_t turn;          /* +0x000a  Turn number */
    union {
        int16_t iPlayer : 5;    /* Player index (bits 0-4) */
        int16_t lSaltTime : 11; /* Salt (bits 5-15) */
    };                      /* +0x000c  Player data */
    union {
        uint16_t dt : 8;           /* File type (byte 14) */
        uint16_t fDone : 1;        /* Bit 8 (byte 15, bit 0) */
        uint16_t fInUse : 1;       /* Bit 9 (byte 15, bit 1) */
        uint16_t fMulti : 1;       /* Bit 10 (byte 15, bit 2) */
        uint16_t fGameOverMan : 1; /* Bit 11 (byte 15, bit 3) */
        uint16_t fCrippled : 1;    /* Bit 12 (byte 15, bit 4) */
        uint16_t wGen : 3;         /* Bits 13-15 (byte 15, bits 5-7) */
    };                      /* +0x000e  File type + flags */
} RTBOF;
```

## Magic Numbers

| Magic  | Description                              |
|--------|------------------------------------------|
| J3D1   | Standard encrypted Stars! file           |
| J3J3   | Race file format                         |
