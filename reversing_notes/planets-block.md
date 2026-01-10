# PlanetsBlock (Type 7)

This block contains universe configuration and the planet list. It maps directly to the `GAME` structure (64 bytes) plus trailing planet coordinate data.

## GAME Structure (64 bytes)

From `types.h:434-464`:

```c
typedef struct _game
{
    int32_t lid;         /* +0x0000 */  // Game ID (unique identifier)
    int16_t mdSize;      /* +0x0004 */  // Universe size mode
    int16_t mdDensity;   /* +0x0006 */  // Planet density mode
    int16_t cPlayer;     /* +0x0008 */  // Number of players
    int16_t cPlanMax;    /* +0x000a */  // Number of planets
    int16_t mdStartDist; /* +0x000c */  // Starting distance mode
    int16_t fDirty;      /* +0x000e */  // Runtime dirty flag (not persisted)
    union {
        uint16_t wCrap;
        struct {
            uint16_t fExtraFuel : 1;   // Max minerals
            uint16_t fSlowTech : 1;    // Slow tech advances
            uint16_t fSinglePlr : 1;   // Single player game
            uint16_t fTutorial : 1;    // Tutorial mode
            uint16_t fAisBand : 1;     // Computer alliances
            uint16_t fBBSPlay : 1;     // Accelerated BBS play
            uint16_t fVisScores : 1;   // Public scores
            uint16_t fNoRandom : 1;    // No random events
            uint16_t fClumping : 1;    // Galaxy clumping
            uint16_t wGen : 3;         // Generation counter
            uint16_t unused : 4;
        };
    }; /* +0x0010 */
    uint16_t turn;       /* +0x0012 */  // Current turn number
    uint8_t rgvc[12];    /* +0x0014 */  // Victory conditions (12 bytes)
    char szName[32];     /* +0x0020 */  // Game name (null-padded)
} GAME;
```

## Field Details

### Bytes 0-3: `lid` (Game ID)

A unique 32-bit identifier generated when a new game is created. Used to verify file integrity - when loading a player .m file, the game checks that its `lid` matches the host .hst file to ensure files belong to the same game instance.

### Bytes 4-5: `mdSize` (Universe Size Mode)

| Value | Size   |
|-------|--------|
| 0     | Tiny   |
| 1     | Small  |
| 2     | Medium |
| 3     | Large  |
| 4     | Huge   |

### Bytes 6-7: `mdDensity` (Planet Density Mode)

| Value | Density |
|-------|---------|
| 0     | Sparse  |
| 1     | Normal  |
| 2     | Dense   |
| 3     | Packed  |

### Bytes 8-9: `cPlayer` (Player Count)

Number of players in the game (1-16).

### Bytes 10-11: `cPlanMax` (Planet Count)

Total number of planets in the universe.

### Bytes 12-13: `mdStartDist` (Starting Distance Mode)

An index indicating how far apart players' homeworlds are placed. Higher values = greater initial separation.

### Bytes 14-15: `fDirty` (Runtime Flag)

**Runtime-only field.** Tracks whether the game data has unsaved modifications. This is NOT meaningful in persisted files and should be ignored on read (typically 0 in saved files).

### Bytes 16-17: Game Settings Flags

Bitmask of game configuration options:

| Bit | Flag        | Description              |
|-----|-------------|--------------------------|
| 0   | fExtraFuel  | Max Minerals enabled     |
| 1   | fSlowTech   | Slow Tech Advances       |
| 2   | fSinglePlr  | Single Player game       |
| 3   | fTutorial   | Tutorial mode            |
| 4   | fAisBand    | Computer Alliances       |
| 5   | fBBSPlay    | Accelerated BBS Play     |
| 6   | fVisScores  | Public Scores            |
| 7   | fNoRandom   | No Random Events         |
| 8   | fClumping   | Galaxy Clumping          |
| 9-11| wGen        | Generation counter (3 bits) |
| 12-15| unused    | Reserved                 |

### Bytes 18-19: `turn` (Turn Number)

Current game year/turn. In .xy files (initial game state), this is 0.

### Bytes 20-31: `rgvc[12]` (Victory Conditions)

12-byte array encoding victory condition settings. Each byte: bit 7 = enabled, bits 0-6 = threshold index.

| Index | Condition                    | Formula (idx → value)     |
|-------|------------------------------|---------------------------|
| 0     | Owns % of planets            | idx*5+20 → 20-100%        |
| 1     | Attain tech level X          | idx+8 → 8-26              |
| 2     | In Y tech fields             | idx+2 → 2-6               |
| 3     | Exceed score                 | idx*1000+1000 → 1k-20k    |
| 4     | Exceed 2nd place by %        | idx*10+20 → 20-300%       |
| 5     | Production capacity (k)      | idx*10+10 → 10-500        |
| 6     | Own capital ships            | idx*10+10 → 10-300        |
| 7     | Highest score after N years  | idx*10+30 → 30-900        |
| 8     | Meet N criteria              | Direct value 1-7          |
| 9     | Min years before winner      | idx*10+30 → 30-500        |
| **10**| **Padding**                  | **Always 0 (unused)**     |
| **11**| **Padding**                  | **Always 0 (unused)**     |

Indices 10-11 are padding bytes for array alignment and are never used by the game.

### Bytes 32-63: `szName[32]` (Game Name)

Game name string, null-padded to 32 bytes.

## Trailing Planet Data

After the 64-byte GAME structure, there are `cPlanMax * 4` bytes of planet coordinate data (NOT included in block size).

Each planet is encoded as 4 bytes (little-endian uint32):

```
Bits 31-22 (10 bits): Planet name ID (index into planet names table)
Bits 21-10 (12 bits): Y coordinate (absolute)
Bits  9-0  (10 bits): X offset from previous planet
```

X coordinates use delta encoding: first planet X = 1000 + offset, subsequent planets X = previous_X + offset.

## Source

- `GAME` structure: types.h:434-464
- Game settings flags: types.h:446-459
