# WaypointRepeatOrdersBlock (Type 10)

This block enables or disables the "Repeat Orders" flag on a fleet, causing the fleet to loop through its waypoints when it reaches the end.

## Format (4 bytes)

```
FF FO WE 00
└───┘ │  └── Padding (unused)
  │   └───── Waypoint + Enable (packed byte)
  └───────── Fleet ID (9 bits, little-endian)
```

### Bytes 0-1: Fleet ID

```
Byte 0: Low 8 bits of fleet number
Byte 1: Bit 0 = high bit of fleet number (bit 8)
        Bits 1-7 = unused
```

Fleet number is 9 bits (0-511).

### Byte 2: Waypoint + Enable (Packed)

```
Byte 2 = 0bWWWW_WWWE
         └──────┘└─ Enable flag (bit 0)
            └────── Starting waypoint index (bits 1-7)
```

| Bits | Field              | Description                                   |
|------|--------------------|-----------------------------------------------|
| 0    | Enable             | 1 = enable repeat orders, 0 = disable         |
| 1-7  | RepeatFromWaypoint | Waypoint index where repeat loop starts (0-127)|

**Decoding**:
```
EnableRepeat = byte2 & 0x01
RepeatFromWaypoint = (byte2 >> 1) & 0x7F
```

**Encoding**:
```
byte2 = (RepeatFromWaypoint << 1) | (EnableRepeat & 0x01)
```

### Byte 3: Padding

Unused. The decompiled code's bit manipulation (`<< 9 >> 8` on 16-bit value) discards this byte entirely. Should be set to 0 for safety.

## Game Logic

The game reads the enable flag from byte 2 bit 0 and compares it to the fleet's current `fRepOrders` flag (bit 9 of the FLEET flags word at offset +0x04). If they differ, it toggles `fRepOrders`.

```c
// From FRunLogRecord case 10:
// Toggle fRepOrders if (byte2 bit 0) != current fRepOrders
puVar5 = (uint *)(fleet + 4);  // Fleet flags word
*puVar5 = *puVar5 ^ ((((byte2 << 1) ^ flags_byte) & 2) << 8);
```

## Related Structures

### FLEET flags (offset +0x04)

```c
struct {
    uint16_t det : 8;        // Detection level
    uint16_t fInclude : 1;   // Include in reports
    uint16_t fRepOrders : 1; // Repeat waypoint orders (bit 9)
    uint16_t fDead : 1;      // Fleet destroyed
    uint16_t fByteCsh : 1;   // Ship counts use 1 byte
    // ...
};
```

## Source

Decompiled from `FRunLogRecord` case 10 at 1048:a38c in stars26jrc3.exe.
