# Stars! Block Structure Reversing Notes

Documentation of the Stars! game file format, derived from binary analysis of `stars26jrc3.exe` using Ghidra, test data analysis, and community research.

## Block Documentation

### M File Blocks (Game State)

| Block Type | File                                                          | Description                                                  |
|------------|---------------------------------------------------------------|--------------------------------------------------------------|
| 6          | [player-block.md](player-block.md)                            | Player data, race settings, relations, production templates  |
| 12         | [events-block.md](events-block.md)                            | Turn events (production, research, battles, etc.)            |
| 31         | [battle-block.md](battle-block.md)                            | Battle VCR recording data                                    |
| 39         | [battle-block.md](battle-block.md#block-continuation-type-39) | Battle continuation (overflow data)                          |
| 40         | [message-block.md](message-block.md)                          | Player-to-player messages                                    |
| 43         | [object-block.md](object-block.md)                            | Map objects (minefields, packets, wormholes, Mystery Trader) |

### X File Blocks (Orders)

| Block Type | File                                                                                  | Description                  |
|------------|---------------------------------------------------------------------------------------|------------------------------|
| 5          | [x-file-blocks.md](x-file-blocks.md#waypointchangetaskblock-type-5---variable-length) | Waypoint task changes        |
| 10         | [x-file-blocks.md](x-file-blocks.md#waypointrepeatordersblock-type-10---4-bytes)      | Repeat orders                |
| 29         | [x-file-blocks.md](x-file-blocks.md#productionqueuechangeblock-type-29)               | Production queue changes     |
| 34         | [x-file-blocks.md](x-file-blocks.md#researchchangeblock-type-34---2-bytes)            | Research changes             |
| 35         | [x-file-blocks.md](x-file-blocks.md#planetchangeblock-type-35---6-bytes)              | Planet settings              |
| 36         | [x-file-blocks.md](x-file-blocks.md#changepasswordblock-type-36---4-bytes)            | Password changes             |
| 38         | [x-file-blocks.md](x-file-blocks.md#playersrelationchangeblock-type-38---2-bytes)     | Diplomatic relations         |
| 42         | [x-file-blocks.md](x-file-blocks.md#setfleetbattleplanblock-type-42---4-bytes)        | Fleet battle plan assignment |
| 44         | [x-file-blocks.md](x-file-blocks.md#renamefleetblock-type-44---variable-length)       | Fleet renaming               |

## Reference Documentation

| File                                     | Description                                              |
|------------------------------------------|----------------------------------------------------------|
| [block-types.md](block-types.md)         | Complete list of all 46 block types                      |
| [file-structure.md](file-structure.md)   | File format details (headers, footers, encryption)       |
| [password-system.md](password-system.md) | Password hashing, race file checksums, serial validation |
| [exploits.md](exploits.md)               | Known game exploits and detection methods                |

## General Notes

1. **Planet ID encoding**: Usually 11 bits (0-2047), stored in first 2 bytes with other flags in upper bits

2. **"No planet" marker**: Global events (like research) use `0xFFFE` (-2 signed) where planet-specific events have planet IDs

3. **Nibble packing**: Stars! developers pack multiple small values into single bytes using nibbles (4 bits each)

4. **Block header format**: `(size & 0x3FF) | (type << 10)` (10-bit size + 6-bit type)

5. **Encryption**: Most block data is encrypted. See [file-structure.md](file-structure.md) for details on which blocks are encrypted.

## Verification Status

- **VERIFIED**: Confirmed against game decompilation and/or VCR screenshots
- **Partial**: Some fields verified, others inferred
- **TBD**: Needs further investigation

## Contributing

When adding new documentation:
1. Create a new `.md` file for significant new block types
2. Update this README with links to new files
3. Include verification status for each field
4. Add examples from test data where possible
