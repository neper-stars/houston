# Block Type 33: MessagesFilterBlock

## Overview

The MessagesFilterBlock stores the player's message filter preferences as a bitmap.
Each bit corresponds to a message type ID - when set, that message type is hidden
from the Messages window in the game client.

## Structure

| Offset | Size | Field           | Description                      |
|--------|------|-----------------|----------------------------------|
| 0-48   | 49   | bitfMsgFiltered | Bitmap of filtered message types |

**Total size:** 49 bytes (392 bits)

## Bit Addressing

The bitmap uses standard bit-array addressing:

```c
byteIndex = messageId / 8;
bitIndex  = messageId % 8;
isFiltered = (bitfMsgFiltered[byteIndex] & (1 << bitIndex)) != 0;
```

**Example:** To check if message type 0x50 (80 decimal) is filtered:
- Byte index: 80 / 8 = 10
- Bit index: 80 % 8 = 0
- Check: `bitfMsgFiltered[10] & 0x01`

## Capacity

- 49 bytes = 392 bits
- Stars! uses message IDs from 0 to approximately 0x190 (400)
- The 392-bit capacity is sufficient for all message types

## Related Game Variables

From decompiled code (`io_loadgame.c`):

| Variable           | Size     | Purpose                                     |
|--------------------|----------|---------------------------------------------|
| `bitfMsgFiltered`  | 49 bytes | Which message types to hide                 |
| `bitfMsgSent`      | 49 bytes | Which message types were sent this turn     |
| `fViewFilteredMsg` | 1 byte   | Toggle to show/hide filtered messages in UI |

## Initialization

On `MSG::ResetMessages()`, both bitmaps are cleared:

```c
memset(&bitfMsgFiltered, 0, 0x31);  // Clear all filters
memset(&bitfMsgSent, 0, 0x31);      // Clear sent tracking
```

## Message Navigation

The `IMsgNext()` function uses the filter bitmap to skip filtered messages:

```c
short MSG::IMsgNext(short fFilteredOnly) {
    // ...
    do {
        i++;
        // ...
        sVar1 = IdmGetMessageN(i);
    } while (((bitfMsgFiltered[sVar1 >> 3] & (1 << (sVar1 & 7))) == 0) == fFilteredOnly);
    // ...
}
```

## File Location

This block appears in:
- `.M` files (player turn files) - stores the player's filter preferences

## Code References

- `io_loadgame.c:2863` - Filter bit check in `IMsgNext()`
- `io_loadgame.c:4047-4048` - Bitmap initialization in `ResetMessages()`

## Message Type ID Mapping

Message IDs from `strings_uncompressed.c` (second string table, starting at array offset ~1418):

### Battle Messages (0x20-0x22, 0x8D-0xB8, 0x113-0x117)

| ID      | Hex         | Message Summary                                  |
|---------|-------------|--------------------------------------------------|
| 32      | 0x20        | Battle aftermath - great damage before destroyed |
| 33      | 0x21        | Battle - watched forces annihilate each other    |
| 34      | 0x22        | Battle - observed forces defeating               |
| 126     | 0x7E        | Battle took place - VCR recording                |
| 141-168 | 0x8D-0xA8   | Various battle outcome messages                  |
| 249-250 | 0xF9-0xFA   | Battle observed (not involved)                   |
| 275-279 | 0x113-0x117 | Additional battle messages                       |

### Colony/Population Messages (0x23-0x26, 0x40-0x41)

| ID  | Hex  | Message Summary                             |
|-----|------|---------------------------------------------|
| 35  | 0x23 | All colonists died - lost planet            |
| 36  | 0x24 | All colonists orbiting died - lost starbase |
| 37  | 0x25 | Population decreased (from X to Y)          |
| 38  | 0x26 | Population decreased due to overcrowding    |
| 100 | 0x64 | Colonists jumped ship - lost planet         |
| 101 | 0x65 | Colonists abandoned starbase                |

### Fleet/Fuel Messages (0x27-0x2A, 0x8B, 0xF3)

| ID  | Hex  | Message Summary                |
|-----|------|--------------------------------|
| 39  | 0x27 | Fleet ran out of fuel          |
| 40  | 0x28 | Waypoint destroyed/disappeared |
| 41  | 0x29 | Fleet ducked behind planet     |
| 42  | 0x2A | Fleet outran scanners          |
| 139 | 0x8B | Out of fuel - speed decreased  |
| 243 | 0xF3 | Ram scoops produced fuel       |

### Cargo Transfer Messages (0x2B-0x2E, 0x42-0x4D, 0x79-0x7A)

| ID      | Hex       | Message Summary          |
|---------|-----------|--------------------------|
| 43-46   | 0x2B-0x2E | Load/unload/beam cargo   |
| 66-77   | 0x42-0x4D | Fleet-to-fleet transfers |
| 121-122 | 0x79-0x7A | Load from fleet          |

### Production Messages (0x2F-0x3E, 0x7C, 0xCD-0xCF)

| ID      | Hex       | Message Summary         |
|---------|-----------|-------------------------|
| 47-52   | 0x2F-0x34 | Starbase built ship(s)  |
| 53-54   | 0x35-0x36 | Built factories         |
| 55-56   | 0x37-0x38 | Built mines             |
| 57-58   | 0x39-0x3A | Built defenses          |
| 62-63   | 0x3E-0x3F | Production queue empty  |
| 124     | 0x7C      | Built planetary scanner |
| 205-207 | 0xCD-0xCF | Built starbase          |

### Research Messages (0x50, 0x5F, 0x78, 0x136)

| ID  | Hex   | Message Summary                   |
|-----|-------|-----------------------------------|
| 80  | 0x50  | Research completed - tech level X |
| 95  | 0x5F  | Tech breakthrough benefit         |
| 120 | 0x78  | Breakthrough - new hull type      |
| 310 | 0x136 | Research completed (alternate)    |

### Colonization Messages (0x51-0x58)

| ID    | Hex       | Message Summary                     |
|-------|-----------|-------------------------------------|
| 81    | 0x51      | Colonize order - not in orbit       |
| 82    | 0x52      | Colonize - planet already populated |
| 83    | 0x53      | Colonize - no colonists             |
| 84    | 0x54      | Colonize - no colony module         |
| 85-88 | 0x55-0x58 | Various colonization errors         |

### Scrap Fleet Messages (0x59-0x5D, 0x140-0x143)

| ID      | Hex         | Message Summary              |
|---------|-------------|------------------------------|
| 89-93   | 0x59-0x5D   | Fleet dismantled messages    |
| 320-323 | 0x140-0x143 | Fleet dismantled (alternate) |

### Strange Artifact (0x5E)

| ID | Hex  | Message Summary                         |
|----|------|-----------------------------------------|
| 94 | 0x5E | Strange artifact found - research boost |

### Bombing Messages (0x60-0x73, 0x8F-0x90)

| ID      | Hex       | Message Summary      |
|---------|-----------|----------------------|
| 96-105  | 0x60-0x69 | Your bombing runs    |
| 106-115 | 0x6A-0x73 | Enemy bombing you    |
| 143-144 | 0x8F-0x90 | Killed all colonists |

### Remote Mining Messages (0x75-0x77)

| ID  | Hex  | Message Summary                  |
|-----|------|----------------------------------|
| 117 | 0x75 | No mining modules                |
| 118 | 0x76 | Planet inhabited - cancel mining |
| 119 | 0x77 | Mining in deep space - canceled  |

### Terraforming Messages (0x7B, 0xBD, 0x12C-0x12D)

| ID      | Hex         | Message Summary              |
|---------|-------------|------------------------------|
| 123     | 0x7B        | Terraforming improved planet |
| 189     | 0xBD        | Remote terraforming complete |
| 300-301 | 0x12C-0x12D | Improved planet value        |

### Comet Strike Messages (0x83-0x8A)

| ID  | Hex  | Message Summary                  |
|-----|------|----------------------------------|
| 131 | 0x83 | Small comet (unowned)            |
| 132 | 0x84 | Medium comet (unowned)           |
| 133 | 0x85 | Large comet (unowned)            |
| 134 | 0x86 | Huge comet (unowned)             |
| 135 | 0x87 | Small comet (owned, 25% deaths)  |
| 136 | 0x88 | Medium comet (owned, 45% deaths) |
| 137 | 0x89 | Large comet (owned, 65% deaths)  |
| 138 | 0x8A | Huge comet (owned, 85% deaths)   |

### Alchemy Message (0x8C)

| ID  | Hex  | Message Summary                |
|-----|------|--------------------------------|
| 140 | 0x8C | Scientists transmuted minerals |

### Minefield Messages (0xBE-0xCC)

| ID      | Hex       | Message Summary                |
|---------|-----------|--------------------------------|
| 190-204 | 0xBE-0xCC | Mine sweeping, mine hits, etc. |

### Mass Driver/Packet Messages (0xD1-0xDA)

| ID      | Hex       | Message Summary                         |
|---------|-----------|-----------------------------------------|
| 209-218 | 0xD1-0xDA | Packet production, capture, bombardment |

### Stargate Messages (0xDE-0xEB)

| ID      | Hex       | Message Summary                    |
|---------|-----------|------------------------------------|
| 222-235 | 0xDE-0xEB | Stargate usage, errors, casualties |

### Planet Discovery Messages (0xA9-0xAE)

| ID      | Hex       | Message Summary                          |
|---------|-----------|------------------------------------------|
| 169     | 0xA9      | Home planet introduction                 |
| 170     | 0xAA      | Found occupied planet                    |
| 171-174 | 0xAB-0xAE | Found new planet (habitable/not/unknown) |

### Victory/Death Messages (0xB5-0xBC)

| ID      | Hex       | Message Summary              |
|---------|-----------|------------------------------|
| 181-184 | 0xB5-0xB8 | Winner declared, player dead |
| 187-188 | 0xBB-0xBC | Race eliminated messages     |

### Mystery Trader Messages (0x108-0x110, 0x12B)

| ID      | Hex         | Message Summary             |
|---------|-------------|-----------------------------|
| 264-271 | 0x108-0x10F | Mystery Trader interactions |
| 272     | 0x110       | Mystery Trader vanished     |
| 299     | 0x12B       | Mystery Trader detected     |

### Anti-Cheat/Punishment Messages (0x100-0x107)

| ID  | Hex   | Message Summary                                 |
|-----|-------|-------------------------------------------------|
| 256 | 0x100 | Population suspects usurper (productivity -20%) |
| 257 | 0x101 | Colonists suspect wrong emperor                 |
| 258 | 0x102 | Fleet refused to move                           |
| 259 | 0x103 | Fleet captains staged strike                    |
| 260 | 0x104 | Fleet defected                                  |
| 261 | 0x105 | Crew sold cargo on black market                 |
| 262 | 0x106 | Freedom fighters destroyed mines                |
| 263 | 0x107 | Freedom fighters stole minerals                 |

### Race Hacking Detection (0x117)

| ID  | Hex   | Message Summary               |
|-----|-------|-------------------------------|
| 279 | 0x117 | Race definition tampered with |

### Tips (0x7F-0x82)

| ID      | Hex       | Message Summary       |
|---------|-----------|-----------------------|
| 127-130 | 0x7F-0x82 | Various gameplay tips |

## Houston Implementation

```go
// Check if a message type is filtered
if mfb.IsFiltered(0x50) {
    // Message type 0x50 (research complete) is hidden
}

// Get all filtered message IDs
filteredIds := mfb.GetFilteredMessageIds()

// Set/clear a filter
mfb.SetFiltered(0x50, true)  // Hide research complete messages
mfb.SetFiltered(0x50, false) // Show research complete messages
```
