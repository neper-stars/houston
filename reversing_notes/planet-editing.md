# Why Planet Edits Revert when only editing .mN and .hst files

The game uses a log-based reconstruction system, not direct file reads.

## Turn Generation Flow

  1. FGenerateTurn allocates a TURNSERIAL structure (16 bytes per player) at generate_turn.c:134-138
  2. FLoadLogFile() (io_loadgame.c:1520-1688) loads your .m1/.h1 files but:
    - Extracts an embedded TURNSERIAL from the log file header (lines 1583-1590)
    - This contains: lSerial (registration serial) + pbEnv (11-byte hardware fingerprint)
  3. FRunLogFile() (io_loadgame.c:2624-2652) replays all log records to reconstruct game state:
  for (; iCur < _DATA::imemLogCur; iCur = iCur + (lprts->flags1 & 0x3ffU) + 2) {
      FRunLogRecord(record_type, size, data);
  }
  3. This processes XFER records, COLDROP records, planet changes - every planet state is derived from these records, not from the binary file layout.
  4. FValidSerialLong() (serial_encoding.c:61-183) validates the TURNSERIAL. If validation fails, the player is flagged as a cheater (flag 0x40000).

## The Problem

When you edit .hst, .m1, .h1 files directly:
  - The log records still contain the original data
  - The TURNSERIAL checksum hasn't been updated
  - FRunLogFile() replays the original records, overwriting your edits

### The "Master Copy"

The log file contains serialized records that are replayed in order. Each record has:
  - Type code (bits 10-15 of flags1)
  - Size (bits 0-9)
  - Data payload

Planet data exists as individual records, not as a contiguous block you can simply patch.

## To Successfully Edit Planet Data

You need to:

  1. Parse the log records within the .m1/.h1 files
  2. Find the specific planet record(s)
  3. Modify the record data
  4. Update any checksums (the TURNSERIAL XOR checksum at byte 15)

The records are the source of truth, not the file's binary layout.

Here's the complete planet log record format:

## Log Record Header

Each log record uses a 16-bit flags1 word:
  - Bits 0-9: Record size (0-1023 bytes)
  - Bits 10-15: Record type code (rt)

### RTPLANET Structure (4 bytes base)

  typedef struct _rtplanet {
      union {
          struct {
              int16_t id : 11;      // Planet ID (0-2047)
              int16_t iPlayer : 5;  // Player ID (0-31)
          };
      }; // +0x0000
      union {
          struct {
              uint16_t det : 7;           // Detail level
              uint16_t fHomeworld : 1;
              uint16_t fInclude : 1;
              uint16_t fStarbase : 1;
              uint16_t fIncEVO : 1;
              uint16_t fIncImp : 1;       // Includes installations
              uint16_t fIsArtifact : 1;
              uint16_t fIncSurfMin : 1;   // Includes surface minerals
              uint16_t fRouting : 1;
              uint16_t fFirstYear : 1;
          };
      }; // +0x0002
  } RTPLANET;  // 4 bytes

### Full PLANET Structure (56 bytes)

  Offset  Size  Field               Description
  ------  ----  -----               -----------
  0x00    2     id                  Planet ID
  0x02    2     iPlayer             Owner player ID
  0x04    2     flags               det(8) + various flags(8)
  0x06    3     rgpctMinLevel[3]    Mineral concentration % (Ir, Bo, Ge)
  0x09    3     rgMinConc[3]        Mineral concentrations
  0x0C    3     rgEnvVar[3]         Environment (Grav, Temp, Rad)
  0x0F    3     rgEnvVarOrig[3]     Original environment
  0x12    2     uGuesses            uPopGuess(12) + uDefGuess(4)
  0x14    8     rgbImp              INSTALLATIONS (packed bitfield)
  0x1C    16    rgwtMin[4]          Surface minerals (4x int32)
  0x2C    4     lStarbase           Starbase info (packed)
  0x30    2     wRouting            Routing target
  0x32    2     turn                Last update turn
  0x34    4     lpplprod            Production queue pointer

Installation Encoding (rgbImp - 8 bytes at offset 0x14)

This is the critical part for mines, factories, and defenses:

### Bit Layout (64 bits total):

  Bits 0-7:    iDeltaPop     (8 bits)  - Population delta
  Bits 8-19:   cDefenses     (12 bits) - Defense count (0-4095)
  Bits 20-31:  cMines        (12 bits) - Mine count (0-4095)
  Bits 32-36:  iScanner      (5 bits)  - Scanner type index
  Bits 37-41:  unused        (5 bits)
  Bits 42-53:  cFactories    (12 bits) - Factory count (0-4095)
  Bit  54:     fArtifact     (1 bit)
  Bit  55:     fNoResearch   (1 bit)
  Bits 56-63:  unused        (8 bits)

  Byte-by-byte breakdown:
  Byte 0 (0x14): iDeltaPop[7:0]
  Byte 1 (0x15): cDefenses[3:0] | iDeltaPop (overflow handling)
  Byte 2 (0x16): cDefenses[11:4]
  Byte 3 (0x17): cMines[7:0]
  Byte 4 (0x18): cMines[11:8] | iScanner[3:0]
  Byte 5 (0x19): unused[4:0] | iScanner[4]
  Byte 6 (0x1A): cFactories[7:0]
  Byte 7 (0x1B): flags | cFactories[11:8]

  Surface Minerals (rgwtMin - 16 bytes at offset 0x1C)

  0x1C: int32_t ironium    (kT on surface)
  0x20: int32_t boranium   (kT on surface)
  0x24: int32_t germanium  (kT on surface)
  0x28: int32_t unused     (always 0?)

  Transfer Records (COLDROP)

  For colonist/cargo transfers to planets:

  typedef struct _coldrop {
      int16_t idFleetSrc;   // +0x00 Source fleet ID
      int16_t idPlr;        // +0x02 Player ID
      int16_t idPlanetDst;  // +0x04 Destination planet ID
      uint16_t flags;       // +0x06 fCanColonize bit
      int32_t cColonist;    // +0x08 Colonist count
  } COLDROP;  // 12 bytes

  Transfer Records (RTXFER variants)

  typedef struct _rtxfer {
      uint16_t id1;         // +0x00 Source ID
      uint16_t id2;         // +0x02 Destination ID
      uint8_t grobj1 : 4;   // +0x04 Source type (planet=1, fleet=2)
      uint8_t grobj2 : 4;   //       Dest type
      uint8_t grbitItems;   // +0x05 Bitmask of items transferred
      char rgcQuan[];       // +0x06 Variable: quantities per item
  } RTXFER;

  grbitItems bitmask:
  - Bit 0: Ironium
  - Bit 1: Boranium
  - Bit 2: Germanium
  - Bit 3: Colonists
  - Bit 4: Fuel

To Edit Planet Data Successfully

  You need to:
  1. Find the RTPLANET record for your target planet (match by id field)
  2. Modify the rgbImp bitfield for installations
  3. Modify rgwtMin for surface minerals
  4. Update any COLDROP/RTXFER records that affect that planet
  5. Recalculate the TURNSERIAL checksum (XOR of bytes 0-14 stored in byte 15)

  The log records are packed sequentially in the file after the RTLOGHDR (17 bytes) and RTBOF (16 bytes) headers.
  
## Log Records ARE Standard Block Types

  The log records are NOT a separate structure - they ARE the standard block types (0-46). The .M files contain blocks that get accumulated into the lpLog buffer and replayed.

  Where Planet Data Lives
  ┌────────────┬────────────────────┬──────────────────────────────────────────────────────────┐
  │ Block Type │        Name        │                         Contains                         │
  ├────────────┼────────────────────┼──────────────────────────────────────────────────────────┤
  │ 13 (0x0D)  │ PlanetBlock        │ Full planet data (owned planets)                         │
  ├────────────┼────────────────────┼──────────────────────────────────────────────────────────┤
  │ 14 (0x0E)  │ PartialPlanetBlock │ Scanned/visible planet data                              │
  ├────────────┼────────────────────┼──────────────────────────────────────────────────────────┤
  │ 35 (0x23)  │ PlanetChangeBlock  │ Planet setting changes (routes, research flag, driver %) │
  └────────────┴────────────────────┴──────────────────────────────────────────────────────────┘
  File Structure

  .M files (e.g., game.m1) contain in order:
  1. FileHeaderBlock (8) - RTBOF structure (16 bytes)
  2. FileHashBlock (9) - TURNSERIAL validation (17 bytes)
  3. Sequence of blocks - These are the "log records"
  4. FileFooterBlock (0) - EOF marker

  How FLoadLogFile() Works

  io_loadgame.c:1592-1598:

  for (iCur = 0; iCur < cbLog; iCur = iCur + (hdrCur & 0x3FF) + 2) {
      ReadRt();                                    // Read next block
      fmemmove(lpLog + iCur, &hdrCur, 2);         // Copy header to lpLog
      fmemmove(lpLog + iCur + 2, &rgbCur, size);  // Copy data to lpLog
  }

  It reads blocks using ReadRt() and copies them sequentially into lpLog.

  How FRunLogFile() Replays

  io_loadgame.c:2644:

  FRunLogRecord(hdr >> 10, hdr & 0x3FF, data);
  //            ^^^^^^^^   ^^^^^^^^^^^  ^^^^
  //            block type  size        payload

  FRunLogRecord() is a giant switch on block type (see all_funcs.c:28652+):
  - Case 0x23 (35): PlanetChangeBlock - modifies planet settings
  - Case 0x17 (23): MoveShipsBlock - transfers between fleets
  - Case 0x1d (29): ProductionQueueChangeBlock
  - etc.

## In Houston

Every primitive is implemented in houston for this. It works.
