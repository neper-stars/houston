# Block Type 11: WaypointChangeTask

This block modifies the task type at a specific waypoint for a fleet.
It seems to be a lightweight alternative to Block Type 5 (WaypointChangeTask)
This is only a theory as this blocktype info was recovered from reverse
engineering the stars! 2.6jrc3.exe but we never encountered this block
type in all our testdata...

## File Layout (6 bytes)

| Offset | Size | Type      | Field         | Description                                 |
|--------|------|-----------|---------------|---------------------------------------------|
| 0x00   | 2    | uint16 LE | FleetID       | Fleet identifier                            |
| 0x02   | 2    | uint16 LE | WaypointIndex | Index into fleet's waypoint array (0-based) |
| 0x04   | 2    | uint16 LE | TaskType      | New task type (0-9)                         |

## Task Type Values

| Value | Name          | Description               |
|-------|---------------|---------------------------|
| 0     | None          | No special task           |
| 1     | Transport     | Load/unload cargo         |
| 2     | Colonize      | Colonize planet           |
| 3     | Remote Mining | Remote mine planet        |
| 4     | Merge         | Merge with another fleet  |
| 5     | Scrap Fleet   | Scrap the fleet           |
| 6     | Lay Mines     | Lay minefields            |
| 7     | Patrol        | Patrol area               |
| 8     | Route         | Auto-route                |
| 9     | Transfer      | Transfer to another fleet |

## Memory Layout

The task is stored in the ORDER structure within PLORD (fleet waypoint container).

### FLEET Structure (relevant fields)

```c
typedef struct _fleet {
    // ... other fields ...
    int16_t cord;        /* +0x0062 */ // Waypoint count
    PLORD *lpplord;      /* +0x0064 */ // Pointer to waypoint array
    // ... other fields ...
} FLEET;
```

### PLORD Structure (waypoint container)

```c
typedef struct PLORD {
    uint16_t flags;      /* +0x0000 */ // cbItem:8, fMark:1, ht:3, cAlloc:4
    uint8_t iordMax;     /* +0x0002 */ // Max waypoints
    uint8_t iordMac;     /* +0x0003 */ // Current waypoint count
    ORDER rgord[1];      /* +0x0004 */ // Flexible array of waypoints
} PLORD;
```

### ORDER Structure (18 bytes per waypoint)

```c
typedef struct _order {
    POINT pt;            /* +0x0000 */ // Destination coordinates (4 bytes)
    int16_t id;          /* +0x0004 */ // Target object ID (2 bytes)
    union {
        struct {
            uint16_t grTask : 4;       // bits 0-3: Task type (0-9)
            uint16_t iWarp : 4;        // bits 4-7: Warp speed
            uint16_t grobj : 4;        // bits 8-11: Target object type
            uint16_t fValidTask : 1;   // bit 12: Valid task flag
            uint16_t fNoAutoTrack : 1; // bit 13: No auto track flag
            uint16_t fUnused : 2;      // bits 14-15: Unused
        };
    };                   /* +0x0006 */ // Flags word (2 bytes)
    union {
        TASKXPORT txp;      // Transport task data
        TASKLAYMINES tlm;   // Lay mines task data
        TASKPATROL tptl;    // Patrol task data
        TASKSELL tsell;     // Sell/transfer task data
    };                   /* +0x0008 */ // Task-specific data (10 bytes)
} ORDER;  // Total: 18 bytes (0x12)
```

### How Block Type 11 Modifies Memory

The decompiled code from `FRunLogRecord` (address 1048:a38c):

```c
case 0xb:  // Block Type 11
    pFVar22 = UTIL::LpflFromId(*(short *)lpb);  // Get fleet from ID
    if (pFVar22 == NULL) return 0;

    // Validate waypoint index
    iVar16 = *(int *)((int)lpb + 2);  // WaypointIndex
    if (fleet->cord <= iVar16) return 0;  // Must be < waypoint count

    // Validate task type
    if (9 < *(int *)((int)lpb + 4)) return 0;  // Must be <= 9

    // Modify grTask field (low 4 bits of flags word)
    // Access: lpplord + (waypointIndex * 18) + 10
    // This resolves to: rgord[waypointIndex].flags (offset 6 within ORDER)
    // Note: +10 from PLORD base = +6 within ORDER due to 4-byte PLORD header
    *(uint *)(*(int *)(fleet + 0x64) + waypointIndex * 0x12 + 10) =
        (*(uint *)(...) & 0xfff0) | (taskType & 0xf);
```

The offset calculation:
- `fleet + 0x64` = lpplord pointer (PLORD base)
- `lpplord + waypointIndex * 0x12 + 10` = `lpplord->rgord[waypointIndex]` + offset 6
- Offset 6 within ORDER is the flags word containing grTask

## Validation Rules

1. **Fleet must exist**: `LpflFromId(FleetID)` must return a valid fleet pointer
2. **Valid waypoint index**: `WaypointIndex < fleet.cord` (waypoint count)
3. **Valid task type**: `TaskType <= 9`

## Related Block Types

| Type | Name                 | Purpose                                          |
|------|----------------------|--------------------------------------------------|
| 5    | WaypointChangeTask   | Changes task with full task-specific data        |
| 10   | WaypointRepeatOrders | Sets/clears repeat orders flag for a fleet       |
| 11   | **This block**       | Lightweight task type change (no task data)      |
| 19   | WaypointTask         | Full waypoint task definition with cargo amounts |
| 20   | Waypoint             | Waypoint position and basic orders               |

## Use Case

Block Type 11 appears to be a **lightweight task change** that only
modifies the task type without including task-specific parameters
(unlike Type 5 which includes transport cargo amounts, mine laying counts,
patrol settings, etc.).

This would be useful when:
- Changing a waypoint from one simple task to another
- Clearing a task (setting to 0 = None)
- The task-specific parameters are not needed or should remain unchanged

## Source

Decompiled from `FRunLogRecord` function at address 1048:a38c
in stars26jrc3.exe (Stars! 2.60j RC3).
