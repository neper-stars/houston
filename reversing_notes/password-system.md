# Password System

Stars! uses a weak 32-bit hash for race passwords. The algorithm is trivially reversible through brute force, and many collisions exist.

## Password Hash Algorithm

```go
func HashRacePassword(password string) uint32 {
    bytes := []byte(password)
    hash := uint32(bytes[0])  // Start with first character

    for i, b := range bytes[1:] {
        if i % 2 == 0 {
            hash = (hash * uint32(b)) & 0xFFFFFFFF  // Odd positions: multiply
        } else {
            hash = (hash + uint32(b)) & 0xFFFFFFFF  // Even positions: add
        }
    }
    return hash
}
```

**Algorithm breakdown:**
1. Initialize hash with ASCII value of first character
2. For each subsequent character at 1-based position `p`:
   - If `p` is odd (2nd, 4th, 6th...): multiply hash by character value
   - If `p` is even (3rd, 5th, 7th...): add character value to hash
3. All operations are modulo 2³² (32-bit overflow)

**Example:**
```
Password: "hob"
  h = 104
  hash = 104
  o (position 1, odd): hash = 104 * 111 = 11544
  b (position 2, even): hash = 11544 + 98 = 11642
Result: 11642 (0x00002D7A)
```

## Hash Weaknesses

The algorithm is extremely weak:
1. **32-bit output**: Only 4 billion possible hashes
2. **Multiplicative structure**: Creates many collisions
3. **No salt**: Same password always produces same hash
4. **Sequential dependency**: Short passwords have small hash space

**Collision example for hash 11642 ("hob"):**
- "hob" (original)
- "awc" (found by brute force)
- Many others exist

## Password Storage

**In PlayerBlock (M files):**
- Offset 12-15 within block data
- 4 bytes, uint32 little-endian
- Hash = 0 means no password set

**In ChangePasswordBlock (X files):**
- See Type 36 documentation in [x-file-blocks.md](x-file-blocks.md)
- 4 bytes, uint32 little-endian
- Hash = 0 removes the password

## Brute Force Performance

With parallel implementation on modern hardware:
- 5-character alphanumeric (36^5 = 60M combinations): < 1 second
- 6-character alphanumeric (36^6 = 2B combinations): ~30 seconds
- Due to collisions, valid alternative passwords are typically found quickly

---

# Race File Checksum

Race files (.r1-.r16) have a 16-bit checksum in the FileFooter that validates the race data.

## Race File Structure

```
[FileHeader]     16 bytes (Type 8) - contains salt value
[PlayerBlock]    Variable (Type 6, encrypted) - contains race data
[FileFooter]     4 bytes (Type 0) - 2-byte header + 2-byte checksum
```

## Encryption Parameters (Race Files)

Race files use specific encryption parameters:
- Salt: From FileHeader
- Game ID: 0
- Turn: 0
- Player Index: 31
- Offset: 0

## Checksum Algorithm

The checksum is computed from decrypted PlayerBlock data plus interleaved race names:

```go
func ComputeRaceFooter(decryptedData []byte, singularName, pluralName string) uint16 {
    // Find where the name data starts
    fullDataFlag := (decryptedData[6] & 0x04) != 0
    index := 8
    if fullDataFlag {
        index = 0x70 // 112 bytes: 8 header + 0x68 (104) full data
        playerRelationsLength := int(decryptedData[index])
        index += 1 + playerRelationsLength
    }

    // Data length is everything before the names section
    dataLength := index

    // Build the checksum data array
    var dData []byte
    dData = append(dData, decryptedData[:dataLength]...)

    // Prepare singular name: leading 0, ASCII bytes, padded to 16 total
    singularOrd := make([]byte, 16)
    singularOrd[0] = 0
    for i, c := range singularName {
        if i < 15 {
            singularOrd[i+1] = byte(c)
        }
    }

    // Prepare plural name: leading 0, ASCII bytes, padded to 16 total
    pluralOrd := make([]byte, 16)
    pluralOrd[0] = 0
    for i, c := range pluralName {
        if i < 15 {
            pluralOrd[i+1] = byte(c)
        }
    }

    // Interleave: add pairs from singular, then pairs from plural
    for i := 0; i < 16; i += 2 {
        dData = append(dData, singularOrd[i], singularOrd[i+1])
        dData = append(dData, pluralOrd[i], pluralOrd[i+1])
    }

    // Compute checksums
    var checkSum1, checkSum2 byte
    for i := 0; i < len(dData); i += 2 {
        checkSum1 ^= dData[i]
    }
    for i := 1; i < len(dData); i += 2 {
        checkSum2 ^= dData[i]
    }

    return uint16(checkSum1) | uint16(checkSum2)<<8
}
```

**Algorithm steps:**
1. Take decrypted PlayerBlock data up to (but not including) the nibble-packed names
2. Decode singular and plural race names to ASCII
3. Pad each name to 15 characters with a leading 0 (16 bytes total)
4. Interleave name bytes in pairs: singular[0:2], plural[0:2], singular[2:4], plural[2:4], ...
5. XOR all even-indexed bytes → checkSum1
6. XOR all odd-indexed bytes → checkSum2
7. Return `(checkSum2 << 8) | checkSum1`

## Password Removal

To remove a password from a race file:
1. Decrypt the PlayerBlock using race file encryption parameters
2. Zero out bytes 12-15 (password hash)
3. Parse the race names from decrypted data
4. Re-encrypt the modified PlayerBlock
5. Recalculate the footer checksum using the algorithm above
6. Update the FileFooter with the new checksum

**Implementation:** `houston race-password <file>` command and `racefixer.RemovePasswordBytes()` function.

## Testing

Verified against 39 race files in `testdata/scenario-racefiles/`:
- All .r1 and .r2 files
- Files with and without passwords
- Various race names (short, long, special characters)
- Different race settings (PRT, LRT, habitat, etc.)

---

## Serial Number Validation

Stars! serial numbers use base-36 encoding (A-Z = 0-25, 0-9 = 26-35).

**Format**: 8 characters, e.g., "SAH62J1E"

**Valid Series Letters** (first character after processing):
- S (18), W (22), C (2), E (4), G (6)

**Valid Number Range**: 100 to 1,500,000

**Character Position Processing**:
- Positions 0, 1, 4, 7, 3 contribute to series/number
- Positions 2, 5, 6 are checksum digits
- XOR with 0x15 applied for values < 0x20
