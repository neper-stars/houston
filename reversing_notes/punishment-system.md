# Stars! Anti-Cheat and Punishment System

This document describes the complete anti-cheat detection and
punishment mechanisms in Stars! 2.60j RC3.

## Overview

Stars! implements several mechanisms to detect and punish players who cheat,
plus a shareware limitation system:

1. **Shareware Limitation** (fCrippled flag) - Unregistered shareware players have tech capped at 9
2. **Serial Piracy Detection** (fCheater flag) - Detects when multiple players share one registration
3. **Race Hacking Detection** (fHacker flag) - Detects modified race files with illegal advantages

## Player Flags (Offset 0x54)

| Bit | Flag      | Value | Status   | Description                              |
|-----|-----------|-------|----------|------------------------------------------|
| 0   | fDead     | 0x01  | Active   | Player has been eliminated               |
| 1   | fCrippled | 0x02  | Legacy   | Unregistered shareware player            |
| 2   | fCheater  | 0x04  | Active   | Serial number piracy detected            |
| 3   | fLearned  | 0x08  | Unused   | Cleared on load, never read              |
| 4   | fHacker   | 0x10  | Active   | Race file modification detected          |

---

## Detection Mechanisms

### 1. Serial Piracy Detection (fCheater)

**Trigger:** Two players have the same serial number but different hardware fingerprints.

**Code Location:** `generate_turn.c:214-246`, `IPlrAlsoCheater @ 1018:07aa`

**How it works:**

Each player's .m file contains a FileHashBlock (type 9, 17 bytes) with:

| Offset | Size | Field | Description |
|--------|------|-------|-------------|
| 0-3 | 4 | lSerial | 32-bit registration serial number |
| 4-14 | 11 | pbEnv | Hardware fingerprint (used in detection) |
| 15-16 | 2 | pbEnv tail | Hardware fingerprint (NOT used in detection) |

**Note:** Some documentation incorrectly lists bytes 0-1 as "Unknown" and lSerial at
offset 2-5. Analysis of `FValidSerialLong` and `IPlrAlsoCheater` confirms that bytes 0-3
are validated as the 32-bit serial number: `FValidSerialLong(CONCAT22(block[2:4], block[0:2]))`.

During turn generation, the game compares all player pairs:
```
For each player pair (A, B):
    If lSerial[A] == lSerial[B]:                    // Same registration
        If pbEnv[A][0:11] != pbEnv[B][0:11]:        // Different hardware
            Set fCheater on player A
            Set fCheater on player B
```

**Validation:** Before comparison, `FValidSerialLong(lSerial)` validates the serial number.
Valid serials satisfy: `(lSerial / 36^4) ∈ {2, 4, 6, 18, 22}`

**Purpose:** Detects when two different people are using the same purchased registration
on different computers - i.e., sharing/pirating the serial number.

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

3. **Serial Piracy Detection (fCheater):**
   - Tech cap at 9 (same as shareware)
   - Production penalties (80% and 50%)
   - Random negative events (~75% chance)
   - Detects multiple players using same registration serial on different hardware

---

## Key Functions

| Function          | Purpose                                | Location         |
|-------------------|----------------------------------------|------------------|
| IPlrAlsoCheater   | Find player with matching serial       | 1018:07aa        |
| FValidSerialLong  | Validate 32-bit serial number          | 1070:48c4        |
| SpankTheCheaters  | Apply turn-based penalties to cheaters | 10f0:192a        |
| CAdvantagePoints  | Calculate race value                   | 10e0:444c        |
| FSendPlrMsg       | Send player message                    | 1030:7ee8        |
| Random            | Random number generation               | 1040:16d2        |

---

## Source Files

- `decompiled/punishment.h` - Header with flag definitions and prototypes
- `decompiled/punishment.c` - Extracted punishment code with pseudocode
- `decompiled/all_funcs.c` - Full decompiled source
