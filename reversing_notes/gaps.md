# Knowledge Gaps in Block Parsing

This document tracks unknown fields and undecoded blocks in the Stars! file format.

## Completely Unknown Block Types

| Block Type | Name           | Status                       |
|------------|----------------|------------------------------|
| 15         | UnknownBlock15 | Purpose unknown, not decoded |
| 18         | UnknownBlock18 | Purpose unknown, not decoded |
| 22         | UnknownBlock22 | Purpose unknown, not decoded |

These blocks are preserved as raw data but their structure and purpose have not been determined.

**Note:** A scan of all testdata files (`.m*`, `.x*`, `.h*`, `.hst`, `.xy`) found no occurrences of these block types. They may be very rare, used only in specific scenarios, or possibly obsolete.

To search for these blocks in your own files, run: `mise run find-unknown-blocks`

---

## Blocks with Unknown Fields

### PlayerBlock (Type 6)

| Field     | Location                         | Notes                |
|-----------|----------------------------------|----------------------|
| `Learned` | PlayerFlags (offset 0x54, bit 3) | Unknown purpose flag |

### FileHashBlock (Type 9)

| Field     | Location  | Notes                       |
|-----------|-----------|-----------------------------|
| `Unknown` | Bytes 0-1 | Possibly flags or player ID |

### SaveAndSubmitBlock (Type 46)

| Field  | Location | Notes                     |
|--------|----------|---------------------------|
| Byte 0 | Flags    | Purpose TBD, usually 0x00 |

---

## Summary Statistics

| Category                   | Count |
|----------------------------|-------|
| Completely unknown blocks  | 3     |
| Blocks with unknown fields | 3     |
| Total unknown fields/bytes | ~4    |

---

## Notes

- All unknown fields are preserved during round-trip encoding to maintain file integrity
- Some "unknown" fields may be padding or reserved for future use by the original game
- Fields marked "TBD" have partial understanding but need confirmation

## How to Help

If you discover the purpose of any unknown field:

1. Update the relevant block struct in `blocks/`
2. Add proper field name and documentation
3. Update tests to verify the new understanding
4. Remove the entry from this document
