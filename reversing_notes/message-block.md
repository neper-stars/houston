# MessageBlock (Type 40)

Player-to-player messages are stored in MessageBlocks.

## Structure

```
Offset  Size  Field
------  ----  -----
0-1     2     Unknown
2-3     2     Unknown
4-5     2     Sender ID (16-bit LE)
6-7     2     Recipient ID (16-bit LE, 0 = "Everyone")
8-9     2     Unknown
10-11   2     Message byte count
12+     Var   Stars! encoded message string
```

## Notes

- HST files do not contain message blocks
- In .x files, sender is always the file's player
- Player IDs in messages are offset: 0 = Everyone, 1-16 = Players 1-16

---

# Mystery Trader Items

The Mystery Trader (ObjectType 3 in [ObjectBlock](object-block.md)) can offer 13 different items, encoded as a bitmask:

| Bit         | Item                     |
|-------------|--------------------------|
| 0 (value=0) | Research (initial state) |
| 0           | Multi Cargo Pod          |
| 1           | Multi Function Pod       |
| 2           | Langston Shield          |
| 3           | Mega Poly Shell          |
| 4           | Alien Miner              |
| 5           | Hush-a-Boom              |
| 6           | Anti Matter Torpedo      |
| 7           | Multi Contained Munition |
| 8           | Mini Morph               |
| 9           | Enigma Pulsar            |
| 10          | Genesis Device           |
| 11          | Jump Gate                |
| 12          | Ship/MT Lifeboat         |
