# Stars! Anti-Cheat and Punishment System

This document describes the complete anti-cheat detection and
punishment mechanisms in Stars! 2.60j RC3.

## Overview

Stars! implements several mechanisms to detect and punish players who cheat,
plus a shareware limitation system:

1. **Shareware Limitation** (fCrippled flag) - Unregistered shareware players have tech capped at 9
2. **File Sharing Detection** (fCheater flag) - Detects when players share save files
3. **Race Hacking Detection** (fHacker flag) - Detects modified race files with illegal advantages

## Player Flags (Offset 0x54)

| Bit | Flag      | Value | Status   | Description                              |
|-----|-----------|-------|----------|------------------------------------------|
| 0   | fDead     | 0x01  | Active   | Player has been eliminated               |
| 1   | fCrippled | 0x02  | Legacy   | Unregistered shareware player            |
| 2   | fCheater  | 0x04  | Active   | File sharing detected                    |
| 3   | fLearned  | 0x08  | Unused   | Cleared on load, never read              |
| 4   | fHacker   | 0x10  | Active   | Race file modification detected          |

---

## Detection Mechanisms

### 1. File Sharing Detection (fCheater)

**Trigger:** Two players have identical homeworld coordinates.

**Code Location:** `all_funcs.c:71947-71967`

**How it works:**
```
For each player pair (A, B):
    If homeworld_coords[A] == homeworld_coords[B]:
        Set fCheater on player A
        Set fCheater on player B
```

When players share .m files (e.g., to see each other's view), their homeworld coordinates will match, triggering detection.

### 2. Race Hacking Detection (fHacker)

**Trigger:** Race value calculation returns < 500.

**Code Location:** `all_funcs.c:72141-72177`

**How it works:**
```
race_value = FUN_10e0_3356(player)  // Calculate race point value

If race_value < 500:
    Set fHacker flag
    Send notification to all players

    // Attempt to fix the race:
    While race_value < 500 AND growth_rate > 1%:
        growth_rate--
        race_value = recalculate()

    // Last resort - zero tech levels
    If race_value still < 500:
        Zero tech levels 8-13
```

### 3. Shareware Limitation (fCrippled)

**Purpose:** Identifies unregistered shareware players.

**Status:** Flag is READ but never SET in Stars! 2.60j RC3 (post-shareware release).

This flag indicates a player using an unregistered shareware copy of Stars!. The same flag
exists in the FileHeader block. When set, the player's tech levels are capped at 9 instead
of the normal maximum of 25.

Stars! 2.60j RC3 is a post-shareware version, so this flag is never set by the game itself.
However, the game still checks this flag for backward compatibility with save files from
the shareware era.

---

## Punishment Details

### fCheater Punishments

| Punishment      | Details                         | Code Location             |
|-----------------|---------------------------------|---------------------------|
| Tech Cap        | Max tech level 9 (vs 25 normal) | `all_funcs.c:81620-81622` |
| Production -20% | Output multiplied by 4/5        | `all_funcs.c:77292-77298` |
| Production -50% | Conditional halving             | `all_funcs.c:17879-17882` |
| Random Events   | ~75% chance of negative events  | `all_funcs.c:73342-73358` |
| Messages        | Extra punishment every 8 turns  | `all_funcs.c:71974-71984` |

#### Tech Level Cap (Line 81620-81622)
```c
if ((tech > 25) ||
    ((player_flags & PLAYER_FLAG_CRIPPLED) && tech > 9) ||
    ((player_flags & PLAYER_FLAG_CHEATER) && tech > 9)) {
    // Block tech advancement
}
```

#### Production Penalty 80% (Line 77292-77298)
```c
if (player_flags & PLAYER_FLAG_CHEATER) {
    production = (production * 4) / 5;  // 80% of original
}
```

#### Random Event Punishment (Line 73342-73358)
```c
if (player_flags & PLAYER_FLAG_CHEATER) {
    // Stagger by turn: skip if (player_id % 8) == (turn % 8) AND turn > 10
    if (random_check(4) == 0) {  // 3/4 chance
        trigger_negative_event();
    }
}
```

### fHacker Punishments

| Punishment         | Details                         | Code Location             |
|--------------------|---------------------------------|---------------------------|
| Growth Degradation | Reduced until race value >= 500 | `all_funcs.c:72163-72168` |
| Tech Zeroing       | Levels 8-13 zeroed if needed    | `all_funcs.c:72170-72176` |
| Notification       | All players informed            | `all_funcs.c:72147-72151` |

**Important:** fHacker does NOT trigger the tech level cap (only fCrippled and fCheater do).

#### Growth Rate Degradation (Line 72163-72168)
```c
while (race_value < 500 && growth_rate > 1) {
    growth_rate--;  // Decrease by 1%
    race_value = recalculate();
}
```

### fCrippled Limitations (Shareware)

Shareware players have the following limitation:

| Limitation | Details                         | Code Location             |
|------------|---------------------------------|---------------------------|
| Tech Cap   | Max tech level 9 (vs 25 normal) | `all_funcs.c:81620-81622` |

This is the same tech cap applied to cheaters, but for a different reason: shareware
players were limited to encourage purchasing the full version.

---

## Message Strings

From `strings_uncompressed.c`:

| ID     | Message                                                            |
|--------|--------------------------------------------------------------------|
| 0x015a | "\\s has degraded \\p from a value of \\i% to \\i%."               |
| 0x015b | "\\s is currently unable to degrade the value of \\p beyond \\i%." |
| 0x0182 | "Hacked race discovered. \\L race statistics have been altered..." |

---

## System Summary

The punishment/limitation system has three distinct purposes:

1. **Shareware Limitation (fCrippled):**
   - Tech cap at 9
   - For unregistered shareware players
   - Not set in 2.60j (post-shareware), but still checked for old saves

2. **Race Hacking Detection (fHacker):**
   - Growth rate degradation (reduced until race value >= 500)
   - Tech zeroing as fallback
   - Affects race statistics, NOT tech advancement
   - Does NOT trigger tech cap

3. **File Sharing Detection (fCheater):**
   - Tech cap at 9 (same as shareware)
   - Production penalties (80% and 50%)
   - Random negative events (~75% chance)
   - Detects players sharing .m files via homeworld coordinate matching

---

## Key Functions

| Function      | Purpose                            | Location        |
|---------------|------------------------------------|-----------------|
| FUN_1018_050e | Find matching cheater by homeworld | Line 2366       |
| FUN_10e0_3356 | Calculate race value               | Called at 72142 |
| FUN_1030_766a | Send player message                | Line 11892      |
| FUN_1040_1676 | Random number check                | Called at 73345 |

---

## Source Files

- `decompiled/punishment.h` - Header with flag definitions and prototypes
- `decompiled/punishment.c` - Extracted punishment code with pseudocode
- `decompiled/all_funcs.c` - Full decompiled source
