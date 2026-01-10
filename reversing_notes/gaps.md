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

### FileHashBlock (Type 9)

| Field     | Location  | Notes                       |
|-----------|-----------|-----------------------------|
| `Unknown` | Bytes 0-1 | Possibly flags or player ID |

---

## Summary Statistics

| Category                   | Count |
|----------------------------|-------|
| Completely unknown blocks  | 3     |
| Blocks with unknown fields | 1     |
| Total unknown fields/bytes | ~1    |

---

## Resolved Mechanisms

### PlayerBlock (Type 6) - fCrippled Flag - RESOLVED

**Status: DEPRECATED (confirmed)**

The `fCrippled` flag (bit 1, 0x02) at player offset 0x54 is a **legacy flag** that
is checked for backward compatibility but **never set** in Stars! 2.60j RC3.

**Evidence:**
1. Exhaustive search found no `| 2` or `| 0x02` operations on player flags
2. fHacker (which IS set for cheating) does NOT trigger the tech cap
3. Only fCrippled and fCheater trigger the tech cap (lines 81620-81622)
4. fCheater IS set for file sharing (matching homeworld coordinates)

**Conclusion:** The punishment system evolved:
- **Old:** fCrippled → tech cap at 9 (no longer set, checked for old saves)
- **New:** fHacker → race value degradation (growth rate decreased)
- **Added:** fCheater → tech cap at 9 (for file sharing detection)

See `player-block.md` for full documentation.

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
