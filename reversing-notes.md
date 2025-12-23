##  Research Complete Event Format (7 bytes)

  50 00 FE FF LL FF FF
  │  │  └───┘ │  └───┘
  │  │    │   │    └─ Field (0-5), NextField
  │  │    │   └─ Level (1-26)
  │  │    └─ Fixed value 0xFFFE
  │  └─ Flags (0x00)
  └─ Event type (0x50 = EventTypeResearchComplete)

  1. 0xFFFE is "no planet" - Production events have planet IDs at bytes 2-3; research is player-global, so it uses -2/0xFFFE as a "no planet" marker. This is consistent with Stars! event structure, not a wasted fixed value.
  2. Byte 6 is NextField - This is the field where research will continue, not a validation repeat. In our samples Field == NextField because players continued in the same field, but they're different data fields.
