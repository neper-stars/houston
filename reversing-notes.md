# Stars! block structure reversing Notes
 
##  Research Complete Event Format (7 bytes)

  50 00 FE FF LL FF FF
  │  │  └───┘ │  └───┘
  │  │    │   │    └─ Field (0-5), NextField
  │  │    │   └─ Level (1-26)
  │  │    └─ Fixed value 0xFFFE
  │  └─ Flags (0x00)
  └─ Event type (0x50 = EventTypeResearchComplete)

### Notes
  2. 0xFFFE is "no planet" - Production events have planet IDs at bytes 2-3; research is player-global, so it uses -2/0xFFFE as a "no planet" marker. This is consistent with Stars! event structure, not a wasted fixed value.
  3. Byte 6 is NextField - This is the field where research will continue, not a validation repeat. In our samples Field == NextField because players continued in the same field, but they're different data fields.

## Research Event Format (confirmed)

  50 00 FE FF LL CF NF
  │  │  └───┘ │  │  └─ Next research field (0-5)
  │  │    │   │  └──── Completed field (0-5)
  │  │    │   └─────── Level achieved (1-26)
  │  │    └─────────── 0xFFFE = "no planet" (global event)
  │  └──────────────── Flags
  └────────────────────Event type (0x50)

  ResearchChangeBlock Format (new)

  Byte 0: Research budget percentage (0-100)
  Byte 1: (next_field << 4) | current_field

### Notes

  1. 0xFFFE = "no planet" - Research is global, so it uses -2 where planet IDs go in production events
  2. Byte 6 in research events = NextField - Where research continues after completion (not a validation repeat)
  3. ResearchChangeBlock - Encodes both the budget % and field changes in 2 bytes using nibble packing
