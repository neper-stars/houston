# PlayerBlock (Type 6)

The PlayerBlock contains player data including race settings, research, diplomatic relations, and production templates.

## Header Structure (bytes 0x00-0x0F)

| Offset   | Size   | Field             | Description                                             |
| -------- | ------ | ----------------- | ------------------------------------------------------  |
| 0x00     | 1      | iPlayer           | Player number (0-15)                                    |
| 0x01     | 1      | cShDef            | Ship design count                                       |
| 0x02     | 2      | cPlanet           | Planet count (low 10 bits)                              |
| 0x04     | 2      | cFleet/cshdefSB   | Fleet count (bits 0-11) + Starbase designs (bits 12-15) |
| 0x06     | 2      | wMdPlr            | Player mode flags (Logo, AI settings, etc.)             |
| 0x08     | 2      | idPlanetHome      | Home planet ID (int16_t)                                |
| 0x0A     | 2      | wScore            | Player ranking position (uint16_t) - see note           |
| 0x0C     | 4      | lSalt             | Password hash                                           |

**Source:** PLAYER structure in types.h:2648

### idPlanetHome (offset 0x08)

The ID of the player's homeworld planet. This is set when the game is created and typically doesn't change.

### wScore (offset 0x0A) - Actually Rank!

**IMPORTANT:** Despite being named "wScore" in the decompiled source, this field stores the
player's **ranking position** (1=1st place, 2=2nd place, etc.), NOT their actual score.

The actual Score value displayed in the "Player Scores" dialog is computed
client-side (see below for exact formula obtained from decompilation)
and is NOT stored in the file.

**Verified examples (see testdata screenshots for proof):**
- scenario-basic/game.m1: wScore=0 (Year 2400, not yet ranked)
- scenario-minefield/game.m1: wScore=1 (Rank 1st, Score 27 in UI - see scores.png)
- scenario-history/game.m2: wScore=2 (Rank 2nd, Score 29 in UI - see scores.png)

**Houston field name:** `Rank` (to reflect actual meaning, not misleading source name)

---

## Score Calculation Formula

The player's Score is computed at runtime by `CalcPlayerScore()`
in UTIL segment. It is NOT stored in game files.

**Source:** Decompiled from `UTIL::CalcPlayerScore` at MEMORY_UTIL:0x58a6

### Formula Components

```
Score = PlanetPopScore + Resources/30 + Starbases×3 + TechScore + ShipScore
```

### 1. Planet Population Score

For each owned planet:
```
popScore = min(6, (population + 999) / 1000)
```

**IMPORTANT - Population Units:**
- Population in the planet data is stored in units of **100 colonists**
- A stored value of 1041 represents 104,100 colonists
- The formula uses the **stored value**, NOT actual colonists
- Example: stored=1041 gives (1041+999)/1000 = 2 points

**MYSTERY - Off-by-One Discrepancy (Houston findings):**
Test data shows expected popScore is consistently +1 higher than the formula produces:
- History scenario: formula gives 2, expected is 3
- Minefield scenario: formula gives 2, expected is 3

This suggests the actual formula might be:
```
popScore = 1 + sum_per_planet(min(6, (population + 999) / 1000))
```

The +1 base is NOT visible in the decompiled CalcPlayerScore function at 1038:58a6.
The source of this discrepancy is currently unknown - could be:
- A different code path for score display
- Version difference between 2.60j RC3 and tested builds
- Post-processing applied elsewhere

**RESOLUTION (Decompiler findings):**
The +1 base formula is confirmed working. Verified with scenario-singleplayer/2483:
- 11 planets with populations ranging from 484,000 to 1,118,800
- Stored values (pop/100): 4840 to 11188
- All cap at 6 except Gladiolus (4840 → 5 points)
- Total: 1 (base) + 60 + 5 = 66 ✓

- Maximum 6 points per planet (at 6,000+ stored population)

### 2. Resource Score

```
resourceScore = totalResources / 30
```
Where `totalResources` = sum of `CResourcesAtPlanet()` for all owned planets.

### 3. Starbase Score

```
starbaseScore = countStarbases × 3
```

Only starbases with **non-zero cargo capacity** (`wtCargoMax != 0`) are counted.

**Important:** This means **Orbital Fort** (cargo capacity = 0) does NOT count
towards the starbase score, only actual starbases with docking capability.

| Hull ID | Name          | Cargo Capacity | Counts for Score? |
|---------|---------------|----------------|-------------------|
| 32      | Orbital Fort  | 0              | No                |
| 33      | Space Dock    | 200            | Yes               |
| 34      | Space Station | 65535          | Yes               |
| 35      | Ultra Station | 65535          | Yes               |
| 36      | Death Star    | 65535          | Yes               |

### 4. Tech Level Score

For each of 6 tech fields (Energy, Weapons, Propulsion, Construction, Electronics, Biotech):

| Tech Level   | Points per Level   | Formula          |
| ------------ | ------------------ | ---------------- |
| 0-3          | level              | `+level`         |
| 4-6          | 5, 7, 9            | `+level×2 - 3`   |
| 7-9          | 12, 15, 18         | `+level×3 - 9`   |
| 10+          | 22, 26, 30, ...    | `+level×4 - 18`  |

**DISCREPANCY NOTE (Houston findings):**
Test data (minefield scenario) shows:
- Tiered formula gives TechScore = 20 (for tech levels summing to 19)
- Expected TechScore = 19 (raw sum of levels)

This suggests the game might use **raw sum of tech levels** rather than the
tiered formula shown in the decompiled code. Further investigation needed.

**RESOLUTION (Decompiler findings):**
The tiered formula IS correct. At low tech levels (sum ~19), tiered ≈ raw,
but at high tech levels the difference is massive:
- Tech sum 19 → tiered ~20 (nearly equal to raw)
- Tech sum 76 → tiered ~197 (vs raw 76!)

Verified with scenario-singleplayer/2483 (Score 838):
- Pop: 66 + Resources: 567 + Starbases: 3 + Tech: 197 + Ships: 5 = 838 ✓

Example for tech levels [13,13,13,13,12,12] = sum 76:
- Tiered: 34+34+34+34+30+30 = 196 points

**Note:** Tech score is skipped if the player is dead (bit 0 of player flags set).

### 5. Ship Score

Ships are categorized by combat power (from `LComputePower()`):

| Category | Power Range      | Description        |
|----------|------------------|--------------------|
| Unarmed  | power = 0        | No weapons         |
| Escort   | 0 < power < 2000 | Light combat ships |
| Capital  | power ≥ 2000     | Heavy combat ships |

Ship scoring:

```
unarmedCapped = min(unarmedCount, planetCount)
escortCapped  = min(escortCount, planetCount)

shipScore = unarmedCapped/2 + escortCapped + capitalScore
```

Capital ship score uses diminishing returns:

```
if capitalCount > 0:
    capitalScore = (planetCount × capitalCount) / (planetCount + capitalCount)
```

### LComputePower() - Ship Combat Power Calculation

Calculates a ship design's combat power rating for score categorization.

**Source:** Decompiled from `UTIL::LComputePower` at MEMORY_UTIL:0x0b32

#### Function Signature

```c
int32_t LComputePower(SHDEF *lpshdef);
```

#### Algorithm Overview

The function iterates through all hull slots and sums damage potential from:
- Beam weapons (grhst = 0x10)
- Torpedoes (grhst = 0x20)
- Bombs (grhst = 0x40)
- Electrical specials (grhst = 0x800, items 12/13)

#### Part Type Constants (grhst)

| Value  | Type       | Description                                          |
|--------|------------|------------------------------------------------------|
| 0x01   | Engine     | Ship engines                                         |
| 0x02   | Armor      | Ship armor                                           |
| 0x10   | Beam       | Beam weapons (phasers, lasers, etc.)                 |
| 0x20   | Torpedo    | Torpedo launchers                                    |
| 0x40   | Bomb       | Planetary bombs                                      |
| 0x80   | Mining     | Mining robots (Robo-Miner, etc.)                     |
| 0x800  | Electrical | Electrical equipment (cloaks, computers, capacitors) |
| 0x1000 | Mechanical | Mechanical equipment (Maneuver Jet, Overthruster)    |

#### Damage Calculations

**Beams (grhst = 0x10):**
```c
beamPower = beam.dp × count × (beam.dRangeMax + 3)
if (beam.grfAbilities & 1) {  // Sapper/gating beam
    beamPower = beamPower / 2;  // Halved for sappers
}
dpBeams += beamPower;
```

**Torpedoes (grhst = 0x20):**
```c
torpPower = torp.dp × count × (torp.dRangeMax - 2)
dpTorps += torpPower;
```

**Bombs (grhst = 0x40):**
```c
bombPower = (bomb.dDmgCol + bomb.dDmgBldg) × count × 2
dpBombs += bombPower;
```

**Electrical Items (grhst = 0x800, item IDs 12 or 13):**

The code checks for electrical items with ID 0x0C (12) or 0x0D (13):

| ID | Name             | grAbility | Multiplier     |
|----|------------------|-----------|----------------|
| 12 | Jammer 50        | 50        | 1.50× per unit |
| 13 | Energy Capacitor | 10        | 1.10× per unit |

**Note:** It's unclear why Jammer 50 (a defensive jammer) is included here.
This may be a bug in the original code - logically only Energy Capacitor (13)
and Flux Capacitor (14) should boost beam damage, but 14 is NOT checked.

```c
// pctCap starts at 1000 (representing 100.0%)
pctCap = 1000;
for each matching item:
    pctCap = pctCap × (100 + item.grAbility) / 100;
// Cap at 255% (0xff)
if (pctCap > 255) pctCap = 255;
// Applied to beam damage
dpBeams = dpBeams × pctCap / 100;
```

#### Speed Bonus

Ship speed adds to beam damage effectiveness:
```c
speed = SpdOfShip(NULL, 0, NULL, 0, lpshdef);  // Score calc uses design only
speedBonus = dpBeams × (speed - 4);
```

Speed 4 is the baseline (no bonus). Each point above 4 multiplies beam power.

#### Final Power

```c
totalPower = dpBombs + dpBeams + speedBonus + dpTorps
```

#### Score Categorization

The returned power value determines ship category:

| Power  | Category | Score Weight                   |
|--------|----------|--------------------------------|
| 0      | Unarmed  | count/2, capped at planetCount |
| 1-1999 | Escort   | count, capped at planetCount   |
| ≥2000  | Capital  | diminishing returns formula    |

#### Key Insights

1. **Range matters**: Beam range adds +3, torpedo range subtracts -2 from
   damage calculation, favoring long-range beams.

2. **Sapper penalty**: Beam weapons with sapper/gating ability (bit 0 of
   grfAbilities) have their power halved for scoring.

3. **Capacitor stacking**: Multiple capacitors multiply together, e.g.,
   two Energy Capacitors = 1.1 × 1.1 = 1.21× beam damage. Capped at 255%.

4. **Speed bonus**: Faster ships get bonus beam damage based on
   `(speed - 4)`, so speed 4 gives no bonus, speed 8 gives 4× beam factor.

5. **Bomb power**: Bombs contribute total colonist + building damage × 2 × count,
   making heavy bombers score as powerful warships.

6. **Missing Flux Capacitor**: The code checks items 12 and 13 but NOT 14
   (Flux Capacitor). This appears to be a bug - Flux Capacitor doesn't
   contribute to power scoring despite being a beam damage booster.

### SpdOfShip() - Ship Speed Calculation

Calculates combat speed for a ship design.

**Source:** Decompiled from `BATTLE::SpdOfShip` at MEMORY_BATTLE

#### Function Signature

```c
int16_t SpdOfShip(FLEET *lpfl, int16_t ishdef, TOK *ptok,
                  int16_t fDumpCargo, SHDEF *lpshdef);
```

For score calculation, called as: `SpdOfShip(NULL, 0, NULL, 0, lpshdef)`

#### Algorithm

**Step 1: Count Speed-Boosting Equipment**

Iterates hull slots to find:
- Engine type and count
- Thrusters (grhst=0x800, item 4): +1 speed each
- Maneuver jets (grhst=0x1000, item 7): +1 speed each
- Overthruster (grhst=0x1000, item 8): +2 speed each
- Robo-Ultra-Miner (grhst=0x80, item 6): +0.5 speed each
- Interspace-10 engine (id=8): +0.5 speed each

**NOTE:** The Robo-Ultra-Miner speed bonus is suspect - a mining robot providing combat
speed is unusual. Earlier decompilation notes incorrectly called this "Sub-light motor",
but item 6 in category 0x80 (Mining) is definitively the Robo-Ultra-Miner. This may be:
- A decompilation error (wrong item ID or category)
- An undocumented game feature
- Dead code that never executes in practice
Needs in-game verification.

**Step 2: Determine Base Warp**

For ramscoop engines (IDs 7, 8, 9, 14, 15):
```c
baseWarp = 10;
```

For other engines:
```c
// Find highest warp where fuel usage ≤ 120mg
baseWarp = 9;
while (baseWarp > 0 && engine.fuelUsage[baseWarp] > 120) {
    baseWarp--;
}
```

**Step 3: Calculate Base Speed**

```c
speed = baseWarp - 4 + thrusters + (halfThrusters + 1) / 2;
```

**Step 4: Race Bonus**

War Monger races (PRT = 2) get +2 speed bonus.

**Step 5: Mass Penalty**

```c
finalSpeed = speed - (mass / 70) / engineCount;
```

Where `mass` is hull empty weight (+ cargo if fleet context).

**Step 6: Clamp Result**

```c
if (finalSpeed > 8) finalSpeed = 8;
if (finalSpeed < 0) finalSpeed = 0;
```

#### Speed Equipment Summary

| Slot Type     | Item ID | Name             | Speed Bonus |
|---------------|---------|------------------|-------------|
| 0x800 (Elec)  | 4       | Thruster         | +1.0        |
| 0x80 (Mining) | 6       | Robo-Ultra-Miner | +0.5        |
| 0x1000        | 7       | Maneuver Jet     | +1.0        |
| 0x1000        | 8       | Overthruster     | +2.0        |
| Engine        | 8       | Interspace-10    | +0.5        |

**Note:** The Robo-Ultra-Miner speed bonus is hardcoded in SpdOfShip, not in
the part data. This may be an undocumented feature. Needs in-game verification
on a Mini-Miner with Robo-Ultra-Miner equipped.

#### Hardcoded Warp-10 Engines

The code checks for engine IDs 7, 8, 9, 14, 15 specifically:

| ID | Name                       |
|----|----------------------------|
| 7  | Trans-Galactic Drive       |
| 8  | Interspace-10              |
| 9  | Enigma Pulsar              |
| 14 | Trans-Galactic Super Scoop |
| 15 | Trans-Galactic Mizer Scoop |

**Note:** This list does NOT include all ramscoop engines. The following are
ramscoops but are NOT in the hardcoded warp-10 list:
- ID 11: Radiating Hydro-Ram Scoop
- ID 12: Sub-Galactic Fuel Scoop
- ID 13: Trans-Galactic Fuel Scoop
- ID 16: Galaxy Scoop

For binary compatibility, Houston should use these exact IDs (7, 8, 9, 14, 15)
rather than trying to detect "ramscoops" generically.

#### Key Insights

1. **Score context**: When called from LComputePower, fleet/token are NULL,
   so cargo weight and race bonuses don't apply - only hull design matters.

2. **Mass penalty**: Heavier ships are slower. Each 70kT of mass costs
   1 speed point per engine. More engines = less penalty.

3. **Speed cap**: Maximum speed is 8, minimum is 0.

4. **Half-thrusters**: Robo-Ultra-Miner and Interspace-10 each give +0.5,
   rounded: 1 gives +1, 2 gives +1, 3 gives +2, etc.

5. **Robo-Ultra-Miner speed bonus**: The code hardcodes a +0.5 combat speed
   bonus for Mining slot (grhst 0x80) item ID 6 (Robo-Ultra-Miner). This is
   NOT stored in the part data - it's a special case in SpdOfShip. This appears
   to be an undocumented feature. Can be verified in-game by comparing combat
   speed of a Mini-Miner with Robo-Ultra-Miner vs other mining robots equipped.

### Example Calculation

From scenario-history (Player 2, Rank 2nd, Score 29):
- Planets: 1
- Starbases: 1
- Unarmed Ships: 5
- Escort Ships: 2
- Capital Ships: 0
- Tech Levels: 18 total (likely 3+3+3+3+3+3)

```
PlanetPopScore ≈ 1-6 (depending on population)
Resources/30   ≈ 143/30 = 4
Starbases×3    = 1×3 = 3
TechScore      = 3+3+3+3+3+3 = 18 (at level 3 each)
Unarmed/2      = min(5,1)/2 = 0
Escort         = min(2,1) = 1
Capital        = 0
─────────────────────────────
Total          ≈ 29 ✓
```

### CResourcesAtPlanet() - Detailed Resource Calculation

The `totalResources` value in the score formula is computed by summing
`CResourcesAtPlanet()` for every owned planet. This function is critical
to understanding how planetary production translates to score.

**Source:** Decompiled from `PLANET::CResourcesAtPlanet` at MEMORY_PLANET:0x788e
(`stars-decompile/decompiled/io_loadgame.c:526-604`)

#### Function Signature

```c
int16_t CResourcesAtPlanet(PLANET *lppl, int16_t iplr);
```

#### Algorithm

**Step 1: Zero Population Check**

If the planet has no colonists, return 0 resources.

**Step 2: Get Race Stats**

```c
popEfficiency = GetRaceStat(player, 0);   // rgAttr[0] - colonist resource output
factEfficiency = GetRaceStat(player, 1);  // rgAttr[1] - factory efficiency
prt = GetRaceStat(player, 14);            // rgAttr[14] - Primary Racial Trait
```

**Step 3: Overcrowding Adjustment**

If population exceeds max capacity:
```
effectivePop = (actualPop - maxPop) / 2 + maxPop
```
Excess population above maximum only contributes at **50% efficiency**.

**Step 4: Resource Calculation (Two Paths)**

**Path A: Alternate Reality (AR) Race (PRT = 8)**

AR races don't use factories - they use orbital bases instead:
```
energyTech = max(1, player.rgTech[0])
resources = floor(sqrt((energyTech × population) / popEfficiency))
```

**Path B: Standard Races (All Other PRTs)**

```
popContribution = population / popEfficiency
factories = min(actualFactories, maxOperableFactories)
factoryContribution = (factories × factEfficiency + 9) / 10
resources = popContribution + factoryContribution
```

**Step 5: Minimum Guarantee**

If the calculated value is 0, return 1. Every inhabited planet
produces at least 1 resource.

#### Race Stat Indices (rgAttr Array)

| Index     | Name               | Description                                       |
|-----------|--------------------|---------------------------------------------------|
| 0         | Pop Efficiency     | Resources per 100 colonists (divisor)             |
| 1         | Factory Efficiency | Factory output multiplier (resources per factory) |
| 3         | Factories Operate  | Factories operable per 100 colonists              |
| 4         | Mine Efficiency    | Mine output multiplier                            |
| 14 (0x0E) | PRT                | Primary Racial Trait (0-9)                        |

### CMaxOperableFactories() - Operable Factory Limit

Calculates how many factories the current population can operate.

**Source:** Decompiled from `PLANET::CMaxOperableFactories` at MEMORY_PLANET:0x7618

#### Function Signature

```c
int16_t CMaxOperableFactories(PLANET *lppl, int16_t iplr, int16_t fNextYear);
```

#### Algorithm

```c
// Get the absolute max factories for this planet
maxFactories = CMaxFactories(lppl, iplr);

// Get race's "factories per 100 colonists" setting
factoriesOperate = GetRaceStat(player, 3);  // rgAttr[3]

// Get current population (optionally including next year's growth)
pop = planet.population;
if (fNextYear) {
    pop += ChgPopFromPlanet(lppl, 0);  // Add expected growth
}

// Calculate operable factories
maxOperable = (pop * factoriesOperate) / 100;

// Cap at planet's max factories
if (maxOperable > maxFactories) {
    maxOperable = maxFactories;
}

// Minimum of 1 (unless AR race)
if (maxOperable < 1) {
    maxOperable = 1;
}

// AR races can't operate factories
if (prt == 8) {
    maxOperable = 0;
}

return maxOperable;
```

#### Formula

```
maxOperable = min(CMaxFactories, (Population × FactoriesOperate) / 100)
```

Where `FactoriesOperate` = `rgAttr[3]` (typically 10-25, meaning 10-25 factories per 100 colonists).

### CMaxFactories() - Planet Factory Capacity

Calculates the maximum number of factories a planet can support.

**Source:** Decompiled from `PLANET::CMaxFactories` at MEMORY_PLANET:0x755c

#### Function Signature

```c
int16_t CMaxFactories(PLANET *lppl, int16_t iplr);
```

#### Algorithm

```c
// Get max population for this planet
maxPop = CalcPlanetMaxPop(lppl->id, iplr);

// Get race's "factories per 100 colonists" setting
factoriesOperate = GetRaceStat(player, 3);  // rgAttr[3]

// Calculate max factories based on max population
maxFactories = (maxPop * factoriesOperate) / 100;

// Minimum of 10 factories
if (maxFactories < 10) {
    maxFactories = 10;
}

// AR races can't have factories
if (prt == 8) {
    maxFactories = 0;
}

return maxFactories;
```

#### Formula

```
maxFactories = max(10, (CalcPlanetMaxPop × FactoriesOperate) / 100)
```

#### Key Insight

The `FactoriesOperate` setting (rgAttr[3]) serves dual purpose:
1. **Max factories** on a planet = maxPop × factoriesOperate / 100
2. **Operable factories** = currentPop × factoriesOperate / 100

This means a race with higher `FactoriesOperate` can build AND operate more
factories per colonist, making it a powerful economic multiplier.

#### Primary Racial Trait (PRT) Values

| Value | Abbreviation | Name                  |
|-------|--------------|-----------------------|
| 0     | HE           | Hyper Expansion       |
| 1     | SS           | Super Stealth         |
| 2     | WM           | War Monger            |
| 3     | CA           | Claim Adjuster        |
| 4     | IS           | Inner Strength        |
| 5     | SD           | Space Demolition      |
| 6     | PP           | Packet Physics        |
| 7     | IT           | Interstellar Traveler |
| 8     | AR           | Alternate Reality     |
| 9     | JOAT         | Jack of All Trades    |

#### Key Observations

1. **Quadratic factory scaling**: Factory contribution uses `factories²`,
   making additional factories increasingly valuable (diminishing returns
   on a per-factory basis, but total contribution grows quadratically).

2. **AR races are unique**: They ignore factories entirely and use a
   square root formula based on Energy tech level. This makes Energy
   research critical for AR resource production.

3. **Overcrowding penalty**: Stuffing more colonists than max pop
   only yields 50% efficiency on the excess - not worthless, but
   suboptimal compared to spreading across multiple planets.

4. **Minimum guarantee**: Even a nearly-dead colony with 100 colonists
   and no factories will produce 1 resource, ensuring it contributes
   *something* to the score.

### CalcPlanetMaxPop() - Maximum Population Calculation

Calculates the maximum population a planet can support for a given race.

**Source:** Decompiled from `PLANET::CalcPlanetMaxPop` at MEMORY_PLANET:0x7096

#### Function Signature

```c
int32_t CalcPlanetMaxPop(int16_t idpl, int16_t iplr);
```

#### Algorithm

**Step 1: Lookup Planet Data**

Retrieves planet structure by ID using `FLookupPlanet()`.

**Step 2: Check for Alternate Reality (AR) Race**

```c
prt = GetRaceStat(player, 14);  // PRT
if (prt == 8) {  // AR race
    if (planet.iPlayer != iplr || !planet.fStarbase) {
        return 0;  // AR needs own starbase to have population
    }
    // Max pop = starbase hull capacity × 4
    maxPop = (huldef[starbase.ihul].baseCapacity - 0x20) * 4;
}
```

AR races can only have population at planets with their own starbases.
Max pop is determined by the starbase hull capacity, not planet habitability.

**Step 3: Standard Race Calculation**

```c
pctDesire = PctPlanetDesirability(planet, iplr);

if (pctDesire < 5) {
    maxPop = 500;  // Minimum for barely habitable
} else {
    maxPop = pctDesire * 100;  // Base: 100 colonists per % desirability
}
```

**Step 4: PRT Modifiers**

| PRT                       | Value | Modifier                     | Effect        |
|---------------------------|-------|------------------------------|---------------|
| HE (Hyper Expansion)      | 0     | `maxPop = maxPop - maxPop/2` | -50% capacity |
| JOAT (Jack of All Trades) | 9     | `maxPop = maxPop + maxPop/5` | +20% capacity |

**Step 5: LRT Modifier**

```c
if (GetRaceGrbit(player, 9) != 0) {  // OBRM (Only Basic Remote Mining)
    maxPop = maxPop + maxPop/10;     // +10% bonus
}
```

#### Summary Formula

For standard races:
```
maxPop = PctPlanetDesirability × 100 × PRT_modifier × LRT_modifier
```

Where:
- HE: PRT_modifier = 0.5
- JOAT: PRT_modifier = 1.2
- Others: PRT_modifier = 1.0
- OBRM: LRT_modifier = 1.1
- Others: LRT_modifier = 1.0

### PctPlanetDesirability() - Habitability Calculation

Calculates how desirable a planet is for a race based on environment match.

**Source:** Decompiled from `PLANET::PctPlanetDesirability` at MEMORY_PLANET:0x6e1e

#### Function Signature

```c
int16_t PctPlanetDesirability(PLANET *lppl, int16_t iPlr);
```

#### Environment Variables

Three environment factors are evaluated (index 0-2):

| Index | Factor      | Planet Offset | Race Min           | Race Ideal | Race Max |
|-------|-------------|---------------|--------------------|------------|----------|
| 0     | Gravity     | lppl + 0x0C   | iplr×0xC0 + 0x59B2 | + 0x59B5   | + 0x59B8 |
| 1     | Temperature | lppl + 0x0D   | iplr×0xC0 + 0x59B3 | + 0x59B6   | + 0x59B9 |
| 2     | Radiation   | lppl + 0x0E   | iplr×0xC0 + 0x59B4 | + 0x59B7   | + 0x59BA |

Values are signed bytes: -50 to +50 (representing 0.12g to 8.00g, -200°C to +200°C, 0mR to 100mR).

#### Algorithm

For each of the 3 environment factors:

**Case A: Immune to Factor** (max value < 0)
```c
pctPos += 10000;  // Full contribution
```

**Case B: Planet Outside Habitable Range**
```c
if (planetValue < raceMin || planetValue > raceMax) {
    penalty = min(15, distance_from_nearest_boundary);
    pctNeg += penalty;
}
```

**Case C: Planet Within Habitable Range**
```c
d = distance_from_ideal_to_boundary;
pctVar = (abs(planetValue - idealValue) * 100) / d;
pctPos += (100 - pctVar)²;  // Squared for exponential falloff

// Additional penalty if beyond race's "preferred" zone
if (dPenalty > 0) {
    pctMod = pctMod * (d*2 - dPenalty) / (d*2);
}
```

#### Final Result

```c
if (pctNeg == 0) {
    // Habitable: return positive percentage
    result = sqrt(pctPos / 3) * pctMod / 100;
} else {
    // Uninhabitable: return negative penalty
    result = -pctNeg;
}
```

#### Return Values

| Value | Meaning                                              |
|-------|------------------------------------------------------|
| > 0   | Habitable, percentage desirability (0-100+)          |
| = 0   | Marginal habitability                                |
| < 0   | Uninhabitable, negative = penalty points (up to -45) |

#### Key Insights

1. **Immunity bonus**: Races immune to an environment factor (max = -1)
   get full 10000 points for that factor, making immunity very powerful.

2. **Hard boundaries**: If ANY factor is outside min-max range, the planet
   is uninhabitable (returns negative). All 3 factors must be in range.

3. **Squared falloff**: Being closer to ideal is exponentially better due
   to the `(100 - pctVar)²` term.

4. **15-point cap**: Uninhabitability penalty is capped at 15 per factor
   (max total: -45).

### Hull Structures and Cost Data

Hull definitions are used in score calculation to determine if a starbase
counts towards the score (only if `wtCargoMax != 0`).

**Source:** Data from `rghuldef[32]` and `rghuldefSB[5]` in `parts.c`

#### HUL Structure (123 bytes)

Base hull data structure, embedded in both HULDEF and SHDEF:

```c
typedef struct _hul {
    int16_t ihuldef;        // +0x00: Hull definition index
    char rgTech[6];         // +0x02: Tech requirements [Ene,Wep,Prop,Con,Ele,Bio]
    char szClass[32];       // +0x08: Hull class name
    uint16_t wtEmpty;       // +0x28: Empty mass (kT)
    uint16_t resCost;       // +0x2A: Resource cost
    uint16_t rgwtOreCost[3];// +0x2C: Mineral costs [Ir, Bo, Ge]
    int16_t ibmp;           // +0x32: Bitmap index
    uint16_t wtCargoMax;    // +0x34: Cargo capacity (kT) - USED IN SCORE CHECK
    uint16_t wtFuelMax;     // +0x36: Fuel capacity (mg)
    uint16_t dp;            // +0x38: Base armor (damage points)
    HS rghs[16];            // +0x3A: Hull slots (4 bytes each)
    uint8_t chs;            // +0x7A: Number of slots
} HUL;  // Total: 123 bytes (0x7B)
```

#### HULDEF Structure (143 bytes)

Hull definition with additional metadata:

```c
typedef struct _huldef {
    HUL hul;                // +0x00: Base hull data
    uint16_t init : 6;      // +0x7B: Initiative bonus
    uint16_t imdAttack : 4; //        Attack mode
    uint16_t imdCategory : 4;//       Category (freighter, warship, etc.)
    uint16_t unused : 2;
    uint16_t wrcCargo;      // +0x7D: Cargo widget resource
    uint8_t rgbrc[16];      // +0x7F: Slot bitmap resource codes
} HULDEF;  // Total: 143 bytes (0x8F)
```

#### LphuldefFromId() - Hull Definition Lookup

**Source:** Decompiled from `PARTS::LphuldefFromId` at MEMORY_PARTS:0x512c

```c
HULDEF* LphuldefFromId(int16_t id) {
    if (id < 0x20) {  // Ship hulls: 0-31
        return &rghuldef[id];  // At segment offset 0x29F0
    } else {          // Starbase hulls: 32-36
        return &rghuldefSB[id - 0x20];
    }
}
```

#### GetTrueHullCost() - Extract Hull Costs

**Source:** Decompiled from `UTIL::GetTrueHullCost` at MEMORY_UTIL:0x5dba

```c
void GetTrueHullCost(int16_t iPlayer, HUL *lphul, uint16_t *rgCost) {
    // Copy mineral costs
    rgCost[0] = lphul->rgwtOreCost[0];  // Ironium
    rgCost[1] = lphul->rgwtOreCost[1];  // Boranium
    rgCost[2] = lphul->rgwtOreCost[2];  // Germanium
    // Copy resource cost
    rgCost[3] = lphul->resCost;
}
```

#### Starbase Hull Data

| ID | Name          | Tech Req | resCost | Ir  | Bo  | Ge  | Cargo | DP   |
|----|---------------|----------|---------|-----|-----|-----|-------|------|
| 32 | Orbital Fort  | Con 0    | 80      | 24  | 0   | 34  | 0     | 100  |
| 33 | Space Dock    | Con 4    | 200     | 40  | 10  | 50  | 200   | 250  |
| 34 | Space Station | Con 0    | 1200    | 240 | 160 | 500 | 65535 | 500  |
| 35 | Ultra Station | Con 12   | 1200    | 240 | 160 | 600 | 65535 | 1000 |
| 36 | Death Star    | Con 17   | 1500    | 240 | 160 | 700 | 65535 | 1500 |

---

## Full Data Section (starts at offset 0x10)

When `FullDataFlag` is set (bit 2 of byte 0x06), the Full Data section contains race settings:

| Offset | Size | Field                                                                         |
|--------|------|-------------------------------------------------------------------------------|
| 0x10   | 9    | Habitability ranges                                                           |
| 0x19   | 1    | Growth rate (max population growth %, typically 1-20)                         |
| 0x1A   | 6    | Tech levels (Energy, Weapons, Propulsion, Construction, Electronics, Biotech) |

## Player Flags (offset 0x54)

Player state flags are stored at offset 0x54 (84 decimal) in the PlayerBlock as a 16-bit value.

### Binary Format

```
Bits 0-4:   State flags
Bits 5-15:  Unused (always 0)
```

### Flag Definitions

| Bit | Mask | Name     | Description                              |
|-----|------|----------|------------------------------------------|
| 0   | 0x01 | Dead     | Player has been eliminated               |
| 1   | 0x02 | Crippled | Player is crippled (definition TBD)      |
| 2   | 0x04 | Cheater  | Cheater flag detected                    |
| 3   | 0x08 | Learned  | **Deprecated** - cleared on load, unused |
| 4   | 0x10 | Hacker   | Hacker flag detected                     |

### Notes

- The Cheater and Hacker flags may be set by the game
  when certain exploit conditions are detected
- The Crippled flag purpose needs further investigation
  (possibly related to victory conditions)

### fLearned Flag Analysis (Bit 3)

**Status: Deprecated/Unused**

The `fLearned` flag (bit 3, mask 0x08) exists in the PLAYER structure at offset
0x54 but is not used by the game. Analysis of the decompiled code reveals:

**Evidence:**

1. **Explicitly cleared on load:** The player loading function at `1070:03e4`
   (FUN_1070_03e4 in all_funcs.c:43111) always clears this bit after loading
   player data from a file:
   ```c
   // all_funcs.c:43164-43165
   pbVar1 = (byte *)((int)param_1 + 0x54);
   *pbVar1 = *pbVar1 & 0xf7;  // Clear bit 3
   ```

2. **Never set:** No code in the decompiled sources sets this bit (no `| 0x08`
   or `| 8` operations on the player flags field at offset 0x54).

3. **Never read:** No code tests this bit to make any decisions.

**Conclusion:**

This flag was likely used in an earlier version of Stars! but was deprecated.
The game preserves the structure field for file format compatibility but
explicitly ignores it by clearing the bit when loading players. Even if a
file contains this bit set, it will be cleared and have no effect.

**Source:** Decompiled from `FUN_1070_03e4` (player loading) at 1070:03e4
in stars26jrc3.exe, confirmed by exhaustive search for bit operations.

---

## Zip Production Queue (offset 0x56)

The "Zip Production" feature allows players to define production
templates that can be quickly applied to any planet.
The Default template (Q1) is auto-applied to newly conquered planets.

### Storage Location

**In PlayerBlock (Type 6, M files):**
- Offset 0x56 (86 decimal), 26 bytes total
- Only the Default queue (Q1) is stored; other custom queues
  are client-side only, they are stored in the Stars.ini file

**In SaveAndSubmitBlock (Type 46, X files):**
- Variable size: 2 + (2 × itemCount) bytes
- Contains the zip prod order being submitted

### Binary Format

```
FR NN [II II] [II II] ... [padding]
│  │  └─────────────────┘
│  │           └─── Items (2 bytes each, up to 12 items)
│  └────────────── Item count (0-12)
└───────────────── fNoResearch flag (0 or 1)
```

### fNoResearch Flag (offset 0)

Controls how the planet contributes to research:

| Value | GUI Label                      | Behavior                                                    |
|-------|--------------------------------|-------------------------------------------------------------|
| 0     | "Contribute to Research"       | Uses global research percentage (normal contribution)       |
| 1     | "Don't contribute to Research" | Only leftover resources after production go to research     |

When `fNoResearch=1`, the planet prioritizes production,
research only receives resources remaining after the production
queue has been fully processed for the year.

**Source:** Field `fNoResearch` in `ZIPPRODQ1` structure (types.h:2331)

### Item Encoding

Each item is a 16-bit little-endian value with format `(Count << 6) | ItemId`:

```
Bits 0-5:   Item ID (0-6 for auto-build items)
Bits 6-15:  Count (0-1023, max settable in GUI is 1020)
```

**IMPORTANT:** This differs from ProductionQueueBlock which
uses `(ItemId << 10) | Count`. ZipProd has the fields reversed!

### Auto-Build Item IDs

| ID | Item               |
|----|--------------------|
| 0  | Auto Mines         |
| 1  | Auto Factories     |
| 2  | Auto Defenses      |
| 3  | Auto Alchemy       |
| 4  | Auto Min Terraform |
| 5  | Auto Max Terraform |
| 6  | Auto Packets       |

### Example Decoding

Raw data: `00 07 C0 02 81 4B 02 FF 43 00 04 FF C5 05 06 6F`

```
fNoResearch: 0x00 (uses global research percentage)
Item count: 7

Item 0: 0x02C0 → ID=(0x02C0 & 0x3F)=0, Count=(0x02C0 >> 6)=11  → AutoMines(11)
Item 1: 0x4B81 → ID=1, Count=302  → AutoFactories(302)
Item 2: 0xFF02 → ID=2, Count=1020 → AutoDefenses(1020)
Item 3: 0x0043 → ID=3, Count=1   → AutoAlchemy(1)
Item 4: 0xFF04 → ID=4, Count=1020 → AutoMinTerraform(1020)
Item 5: 0x05C5 → ID=5, Count=23  → AutoMaxTerraform(23)
Item 6: 0x6F06 → ID=6, Count=444 → AutoPackets(444)
```

### Notes

- **Items CAN repeat**: The same auto-build item type can
  appear multiple times with different counts
- **Maximum 12 items**: The queue is limited to 12 items
- Count of 1 for AutoAlchemy may indicate "enabled" since
  alchemy doesn't have a meaningful quantity limit

---

## Client-Side Zip Queue Storage (Stars.ini)

Custom zip queue definitions (Q2, Q3, Q4 names and contents)
are stored in `Stars.ini`, typically at `C:\Windows\Stars.ini`
(or under Wine: `~/.wine/drive_c/windows/Stars.ini`).

**INI Section: `[ZipOrders]`**

```ini
[ZipOrders]
ZipOrdersP1=agaeaabeaaceaaeiaafiaagiaa<Default>
ZipOrdersP2=abaajbZO1
ZipOrdersP3=abbajbZO2
ZipOrdersP4=acaeaabeaaZO3
ZipOrdersP5=
```

**Format**: `ZipOrdersP{n}=[encoded_data][QueueName]`
- `n` = Queue slot number (1-4: Default, Q2, Q3, Q4)
- `[encoded_data]` = Base-11 encoded queue items (lowercase letters a-k)
- `[QueueName]` = Queue name appended directly after encoded data

The encoding uses lowercase letters where 'a'=0, 'b'=1, ..., 'k'=10.

**Count Encoding (Base-11)**:

```
count = (high_char - 'a') × 11 + (low_char - 'a')
```

Examples:
- `aa` = 0×11 + 0 = 0 (no limit / empty)
- `ab` = 0×11 + 1 = 1
- `jb` = 9×11 + 1 = 100

---

## Player Relations Storage

After turn generation, diplomatic relations are stored
in the PlayerBlock within the player's own M file.

**Location**: In PlayerBlock, after FullDataBytes (at offset 0x70),
a length-prefixed array stores relations.

**Format:**
```
LL [R0] [R1] [R2] ... [R(LL-1)]
│   └────────────────────────── Relation to player i (0=Neutral, 1=Friend, 2=Enemy)
└────────────────────────────── Length (number of entries)
```

**IMPORTANT: Different encoding from order files!**

| Value | Order File (Type 38) | M File Storage |
|-------|----------------------|----------------|
| 0     | Friend               | Neutral        |
| 1     | Neutral              | Friend         |
| 2     | Enemy                | Enemy          |

Friend and Neutral are **swapped** between order files and M file storage.

**Storage rules:**
- `PlayerRelations[i]` = relation to player `i`
- Array length varies by player - indices beyond array length default to Neutral
- Player's relation to self (own index) is stored as Neutral (0)

**Example from 3-player game:**
```
P0 (Hobbits):   set P1=Friend, P2=Neutral
  Stored: [02] [00 01] = length=2, [0]=Neutral(self), [1]=Friend(P1)
  P2 defaults to Neutral (not stored)

P1 (Halflings): set P0=Neutral, P2=Enemy
  Stored: [03] [00 00 02] = length=3, [0]=Neutral(P0), [1]=Neutral(self), [2]=Enemy(P2)

P2 (Orcs):      set P0=Friend, P1=Enemy
  Stored: [02] [01 02] = length=2, [0]=Friend(P0), [1]=Enemy(P1)
  Self defaults to Neutral (not stored)
```

---

## AI Player Configuration

In PlayerBlock (Type 6), byte 7 encodes AI settings:

```
Bit 0: Always 1
Bit 1: AI enabled (0=off, 1=on)
Bits 2-3: AI skill level
  00 = Easy
  01 = Standard
  10 = Harder
  11 = Expert
Bit 4: Always 0
Bits 5-7: Mode (flip when set to Human Inactive)
```

**Special Values**:
- AI password "viewai" = bytes [238, 171, 77, 9] (0xEEAB4D09)
- Human(Inactive) password = [255, 255, 255, 255] (bit-inverted from blank)

---

## Lesser Race Traits (LRT) Bitmask

14 traits encoded in 2 bytes at PlayerBlock offset 78-79:

| Bit   | Short | Full Name                |
|-------|-------|--------------------------|
| 0     | IFE   | Improved Fuel Efficiency |
| 1     | TT    | Total Terraforming       |
| 2     | ARM   | Advanced Remote Mining   |
| 3     | ISB   | Improved Starbases       |
| 4     | GR    | Generalised Research     |
| 5     | UR    | Ultimate Recycling       |
| 6     | MA    | Mineral Alchemy          |
| 7     | NRSE  | No Ram Scoop Engines     |
| 8     | CE    | Cheap Engines            |
| 9     | OBRM  | Only Basic Remote Mining |
| 10    | NAS   | No Advanced Scanners     |
| 11    | LSP   | Low Starting Population  |
| 12    | BET   | Bleeding Edge Technology |
| 13    | RS    | Regenerating Shields     |
| 14-15 | -     | Unused                   |

