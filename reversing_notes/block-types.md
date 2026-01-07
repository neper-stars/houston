# Complete Block Type List

| ID | Type                            | Notes                                           |
|----|---------------------------------|-------------------------------------------------|
| 0  | FileFooterBlock                 | Year (.M/.HST), Checksum XOR (.R), null (.X/.H) |
| 1  | ManualSmallLoadUnloadTaskBlock  |                                                 |
| 2  | ManualMediumLoadUnloadTaskBlock |                                                 |
| 3  | WaypointDeleteBlock             |                                                 |
| 4  | WaypointAddBlock                |                                                 |
| 5  | WaypointChangeTaskBlock         |                                                 |
| 6  | PlayerBlock                     |                                                 |
| 7  | PlanetsBlock                    |                                                 |
| 8  | FileHeaderBlock                 | Unencrypted                                     |
| 9  | FileHashBlock                   | Copy protection                                 |
| 10 | WaypointRepeatOrdersBlock       |                                                 |
| 11 | Unknown                         | Never observed                                  |
| 12 | EventsBlock                     |                                                 |
| 13 | PlanetBlock                     | Full planet data                                |
| 14 | PartialPlanetBlock              | Scanned planet data                             |
| 15 | Unknown                         | Never observed                                  |
| 16 | FleetBlock                      | Full fleet data                                 |
| 17 | PartialFleetBlock               | Scanned fleet data                              |
| 18 | Unknown                         | Never observed                                  |
| 19 | WaypointTaskBlock               | In .M/.HST files                                |
| 20 | WaypointBlock                   | In .M/.HST files                                |
| 21 | FleetNameBlock                  | Custom fleet names                              |
| 22 | Unknown                         | Never observed                                  |
| 23 | MoveShipsBlock                  |                                                 |
| 24 | FleetSplitBlock                 |                                                 |
| 25 | ManualLargeLoadUnloadTaskBlock  |                                                 |
| 26 | DesignBlock                     | Ship/starbase design                            |
| 27 | DesignChangeBlock               | Design modification in .X                       |
| 28 | ProductionQueueBlock            |                                                 |
| 29 | ProductionQueueChangeBlock      |                                                 |
| 30 | BattlePlanBlock                 |                                                 |
| 31 | BattleBlock                     | VCR data                                        |
| 32 | CountersBlock                   | Game counters                                   |
| 33 | MessagesFilterBlock             | Message filtering prefs                         |
| 34 | ResearchChangeBlock             |                                                 |
| 35 | PlanetChangeBlock               |                                                 |
| 36 | ChangePasswordBlock             | .X files / Password in .HST                     |
| 37 | FleetsMergeBlock                |                                                 |
| 38 | PlayersRelationChangeBlock      |                                                 |
| 39 | BattleContinuationBlock         | Extended battle data                            |
| 40 | MessageBlock                    | Player messages                                 |
| 41 | AIRecordBlock                   | AI records in .H files (not decoded)            |
| 42 | SetFleetBattlePlanBlock         |                                                 |
| 43 | ObjectBlock                     | Minefields, packets, wormholes, MT              |
| 44 | RenameFleetBlock                |                                                 |
| 45 | PlayerScoresBlock               | Victory condition tracking                      |
| 46 | SaveAndSubmitBlock              | Turn submission                                 |

---

## Design Slot Item Categories

In DesignBlock (Type 26) and DesignChangeBlock (Type 27), each slot is encoded as 4 bytes:

```
CC CC II NN
└───┘ │  └── Count (number of items)
  │   └───── ItemId (0-indexed within category)
  └───────── Category (16-bit LE, item type)
```

The Category field indicates the **type of item equipped** in the slot, not the hull's slot definition:

| Value  | Category    | Description                                                    |
|--------|-------------|----------------------------------------------------------------|
| 0x0000 | Empty       | No item equipped                                               |
| 0x0001 | Engine      | Engines (Trans-Star 10, NHRS, Galaxy, etc.)                    |
| 0x0002 | Scanner     | Ship scanners (Bat, Rhino, Mole, Possum, etc.)                 |
| 0x0004 | Shield      | Shields (Mole-skin, Cow-hide, Bear, Gorilla, etc.)             |
| 0x0008 | Armor       | Armor (Tritanium, Kelarium, Neutronium, etc.)                  |
| 0x0010 | BeamWeapon  | Beam weapons (Laser, X-Ray, Phaser, etc.)                      |
| 0x0020 | Torpedo     | Torpedoes (Alpha, Beta, Delta, Omega, etc.)                    |
| 0x0040 | Bomb        | Bombs (LadyFinger, M-70, Smart, etc.)                          |
| 0x0080 | MiningRobot | Mining robots (Midget, Mini, Maxi, etc.)                       |
| 0x0100 | MineLayer   | Mine layers (Mine40, Heavy50, Speed20, etc.)                   |
| 0x0200 | Orbital     | Orbital devices (Stargates, Mass Drivers)                      |
| 0x0400 | Planetary   | Planetary scanners (Viewer, Scoper, Snooper)                   |
| 0x0800 | Electrical  | Electrical devices (Cloaks, Jammers, Capacitors)               |
| 0x1000 | Mechanical  | Mechanical devices (Cargo Pod, Fuel Tank, Colonization Module) |

### ItemId Indexing

ItemId is **0-indexed** within each category. To convert to game constants (which are typically 1-indexed), add 1:

```
scannerConstant = slot.ItemId + 1
```

**Example**: A slot with `Category=0x0002, ItemId=1, Count=1` means:
- Category 0x0002 = Scanner
- ItemId 1 → Scanner constant 2 = Rhino Scanner
- Count 1 = One scanner equipped

### Scanner Detection Example

```go
for _, slot := range design.Slots {
    if slot.Category == ItemCategoryScanner && slot.Count > 0 {
        scannerID := slot.ItemId + 1  // Convert 0-indexed to 1-indexed
        stats := ShipScannerStats[scannerID]
        // Use stats.NormalRange, stats.PenetratingRange
    }
}
```

---

## Block Type Mappings from Binary Analysis

From decompiling the Stars! binary (stars26jrc3.exe), the following block type mappings were confirmed:

### Writing Functions

Blocks are written using two key functions:
- `WriteMemRt(rt, cb, data)` - Writes to memory log buffer (for X files, player orders)
- `WriteRt(rt, cb, data)` - Writes directly to file stream (for M/H files)

Block header format: `(size & 0x3FF) | (type << 10)` (10-bit size + 6-bit type)

### Confirmed Block Type to Function Mappings

| Type      | Block Name                 | Writing Function                | Purpose                             |
|-----------|----------------------------|---------------------------------|-------------------------------------|
| 1         | ManualSmallLoadUnloadTask  | LogMakeValidXfer (qty < 0x80)   | Small cargo transfer (signed byte)  |
| 2         | ManualMediumLoadUnloadTask | LogMakeValidXfer (qty < 0x8000) | Medium cargo transfer (signed word) |
| 3         | WaypointDelete             | LogChangeFleet                  | Remove waypoint from fleet          |
| 4         | WaypointAdd                | LogChangeFleet                  | Add new waypoint to fleet           |
| 5         | WaypointChangeTask         | LogChangeFleet                  | Modify task at existing waypoint    |
| 23 (0x17) | MoveShips                  | LogMakeValidXferf               | Transfer ships between fleets       |
| 24 (0x18) | FleetSplit                 | LogSplitFleet                   | Split a fleet into two              |
| 25 (0x19) | ManualLargeLoadUnloadTask  | LogMakeValidXfer (large)        | Large cargo transfer (32-bit)       |
| 27 (0x1b) | DesignChange               | LogChangeShDef                  | Ship/starbase design changes        |
| 29 (0x1d) | ProductionQueueChange      | LogChangePlanet                 | Production queue orders             |
| 35 (0x23) | PlanetChange               | LogChangePlanet                 | Planet setting changes              |
| 37 (0x25) | FleetsMerge                | LogMergeFleet                   | Merge multiple fleets               |
| 38 (0x26) | PlayersRelationChange      | LogChangeRelations              | Diplomatic relation changes         |
| 44 (0x2c) | RenameFleet                | LogChangeName                   | Fleet renaming                      |

### Unknown Block Types

Block types 11, 15, 18, and 22 were not found in the standard game operation functions. Possible explanations:
- Used only in specific edge cases (AI host file records, tutorial mode)
- Legacy/deprecated types no longer written by modern game versions
- Used only in M-file read operations but never written to X files

### Key Structures

From the NB09 debug symbols in the binary:

```c
// RTSHIPINT - Simple fleet/waypoint reference (4 bytes)
typedef struct _rtshipint {
    int16_t id;   // Fleet ID
    int16_t i;    // Index (e.g., waypoint index)
} RTSHIPINT;

// RTSHIPINT2 - Extended fleet/waypoint reference (6 bytes)
typedef struct _rtshipint2 {
    int16_t id;   // Fleet ID
    int16_t i;    // Index 1
    int16_t i2;   // Index 2
} RTSHIPINT2;

// LOGXFER - Cargo transfer log entry (24 bytes)
typedef struct _logxfer {
    int16_t id;           // Target ID
    int16_t grobj;        // Object type (planet=1, fleet=2)
    int32_t rgdItem[5];   // Quantities: Ironium, Boranium, Germanium, Colonists, Fuel
} LOGXFER;

// RTCHGNAME - Rename record (variable size)
typedef struct _rtchgname {
    int16_t grobj;        // Object type
    int16_t id;           // Object ID
    uint8_t rgb[1];       // Compressed name data
} RTCHGNAME;
```
