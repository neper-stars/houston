# Stars! Maximum Defense Calculations

This document describes how Stars! 2.60j RC3 calculates the maximum number of
defenses a planet can build and operate.

## Overview

There are two related defense limits:

1. **CMaxDefenses** - Absolute maximum defenses based on planet habitability
2. **CMaxOperableDefenses** - Population-limited defenses that can actually operate

Both limits are relevant when building defenses:
- You can only build up to CMaxDefenses
- Only CMaxOperableDefenses will actually function

## CMaxDefenses

### Formula

```
CMaxDefenses = clamp(habitability% * 4, 10, 100)
```

Where:
- `habitability%` = Planet habitability for the owning player (0-100)
- Result is clamped to range [10, 100]
- Alternate Reality (AR) race returns 0 (no planetary defenses)

### Code Location

- **Address:** 1048:5714 (FUN_1048_5714)
- **Source:** all_funcs.c lines 26540-26559

### Raw Decompiled Code

```c
int __cdecl16far FUN_1048_5714(undefined2 param_1, undefined2 param_2, int param_3)
{
    int iVar1;
    int local_6;

    local_6 = FUN_1048_5080(param_1, param_2, param_3);  // PctPlanetDesirability
    local_6 = local_6 * 4;
    if (local_6 < 10) {
        local_6 = 10;
    }
    if (100 < local_6) {
        local_6 = 100;
    }
    iVar1 = FUN_10e0_253a((int)&c_common::vtickTooltip1stVis + param_3 * 0xc0, 0xe);
    if (iVar1 == 8) {  // PRT_ALTERNATE_REALITY
        local_6 = 0;
    }
    return local_6;
}
```

### Examples

| Habitability | Calculation | CMaxDefenses |
|--------------|-------------|--------------|
| 100%         | 100 * 4 = 400 → 100 | 100 |
| 50%          | 50 * 4 = 200 → 100 | 100 |
| 25%          | 25 * 4 = 100 | 100 |
| 15%          | 15 * 4 = 60 | 60 |
| 5%           | 5 * 4 = 20 | 20 |
| 0%           | 0 * 4 = 0 → 10 | 10 |
| AR race      | N/A | 0 |

---

## CMaxOperableDefenses

### Formula

```
pop_limit = (population + 24) / 25
CMaxOperableDefenses = min(CMaxDefenses, min(1000, pop_limit))
```

Where:
- `population` = Current planet population (in game units)
- `pop_limit` = Maximum defenses based on population
- Result is capped at 1000
- Alternate Reality (AR) race returns 0

### Code Location

- **Address:** 1048:5768 (FUN_1048_5768)
- **Source:** all_funcs.c lines 26571-26608

### Raw Decompiled Code

```c
int __cdecl16far FUN_1048_5768(undefined4 param_1, int param_3, int param_4)
{
    int iVar1;
    uint uVar2;
    int iVar3;
    bool bVar4;
    long lVar5;
    uint local_c;
    int local_a;
    int local_8;
    int local_4;

    iVar1 = FUN_1048_5714((int)param_1, param_1._2_2_, param_3);  // CMaxDefenses
    local_c = *(uint *)((int)param_1 + 0x28);   // population low
    local_a = *(int *)((int)param_1 + 0x2a);    // population high
    if (param_4 != 0) {  // fNextYear
        iVar3 = local_a;
        uVar2 = FUN_1038_4b42((int)param_1, param_1._2_2_, 0);  // ChgPopFromPlanet
        bVar4 = CARRY2(local_c, uVar2);
        local_c = local_c + uVar2;
        local_a = local_a + iVar3 + (uint)bVar4;
    }
    lVar5 = FUN_1118_0c28(local_c + 0x18, local_a + (uint)(0xffe7 < local_c), 0x19, 0);  // div
    local_8 = (int)lVar5;
    if (1000 < lVar5) {
        local_8 = 1000;  // cap at 1000
    }
    local_4 = local_8;
    if (iVar1 < local_8) {
        local_4 = iVar1;  // cap by CMaxDefenses
    }
    iVar1 = FUN_10e0_253a((int)&c_common::vtickTooltip1stVis + param_3 * 0xc0, 0xe);
    if (iVar1 == 8) {  // PRT_ALTERNATE_REALITY
        local_4 = 0;
    }
    return local_4;
}
```

### Examples

| Habitability | Population | CMaxDefenses | Pop Limit | CMaxOperableDefenses |
|--------------|------------|--------------|-----------|----------------------|
| 100%         | 10,000     | 100          | 401       | 100                  |
| 100%         | 1,000      | 100          | 40        | 40                   |
| 15%          | 50,000     | 60           | 2,001→1000| 60                   |
| 15%          | 500        | 60           | 20        | 20                   |
| 0%           | 5,000      | 10           | 201       | 10                   |
| 25%          | 100        | 100          | 4         | 4                    |
| AR race      | Any        | 0            | N/A       | 0                    |

---

## Population Unit Note

Stars! uses 100 colonists = 1 population unit internally:
- "10,000 colonists" displayed = 1,000,000 internal population units
- The division by 25 operates on the internal units

The formula `(pop + 24) / 25` means:
- 25 internal units → 1 defense
- Which is 0.25 colonists → 1 defense (displayed)
- Or 1 colonist can support 4 defenses

---

## Comparison with Factories/Mines

Defense calculations follow a similar pattern to factories and mines, but with key differences:

### CMaxOperableFactories

```c
max_factories = cMaxOperFactories_perPop * population / 10000
```
- Uses race-specific `cMaxOperFactories_perPop` (typically 10-25)
- Efficiency varies by race traits

### CMaxOperableMines

```c
max_mines = cMaxOperMines_perPop * population / 10000
```
- Uses race-specific `cMaxOperMines_perPop` (typically 10-25)
- Efficiency varies by race traits

### CMaxOperableDefenses

```c
max_defenses = (population + 24) / 25
```
- **Fixed rate** (1 defense per 25 pop)
- **Not affected by race traits** (except AR gets 0)
- Has habitability-based cap (CMaxDefenses)

---

## Alternate Reality (AR) Race

The Alternate Reality race (PRT = 8) has **no planetary defenses**:

- Both CMaxDefenses and CMaxOperableDefenses return 0
- AR lives in orbit (starbases) rather than on planetary surfaces
- Defense calculations check for AR at the end of each function

Code pattern:
```c
prt = GetPlayerPRT(player);  // FUN_10e0_253a(..., 0x0e)
if (prt == 8) {  // PRT_ALTERNATE_REALITY
    result = 0;
}
```

---

## Key Functions

| Function | Address | Purpose |
|----------|---------|---------|
| CMaxDefenses | 1048:5714 | Habitability-based defense limit |
| CMaxOperableDefenses | 1048:5768 | Population-based operable limit |
| PctPlanetDesirability | 1048:5080 | Get planet habitability % |
| ChgPopFromPlanet | 1038:4b42 | Get projected population change |
| GetPlayerField | 10e0:253a | Read player structure field |
| DivLong | 1118:0c28 | 32-bit division helper |

---

## Houston Implementation

```go
// CMaxDefenses calculates the absolute maximum defenses based on habitability
func CMaxDefenses(habitability int, prt PRT) int {
    if prt == PRT_AR {
        return 0
    }
    maxDef := habitability * 4
    if maxDef < 10 {
        maxDef = 10
    }
    if maxDef > 100 {
        maxDef = 100
    }
    return maxDef
}

// CMaxOperableDefenses calculates population-limited operable defenses
func CMaxOperableDefenses(habitability int, population int, prt PRT) int {
    if prt == PRT_AR {
        return 0
    }
    maxDef := CMaxDefenses(habitability, prt)
    popLimit := (population + 24) / 25
    if popLimit > 1000 {
        popLimit = 1000
    }
    if maxDef < popLimit {
        return maxDef
    }
    return popLimit
}
```

---

## Source Files

- `decompiled/max_defenses.h` - Raw header with Ghidra function names
- `decompiled/max_defenses.c` - Raw decompiled code
- `decompiled/max_defenses_cleaned.h` - Cleaned header with readable names
- `decompiled/max_defenses_cleaned.c` - Cleaned implementation
- `decompiled/all_funcs.c` - Full decompiled source (lines 26540-26608)
