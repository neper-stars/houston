# MessageBlock (Type 40)

Player-to-player messages are stored in MessageBlocks.

## Structure

The file format is the in-memory `MSGPLR` structure written directly to disk.

```
Offset  Size  Field         Original Name    Description
------  ----  -----         -------------    -----------
0-1     2     (garbage)     lpmsgplrNext     Low word of linked list pointer - IGNORE
2-3     2     (garbage)     (high word)      High word of linked list pointer - IGNORE
4-5     2     Sender ID     iPlrFrom         Sender player index (0-15)
6-7     2     Recipient ID  iPlrTo           0 = broadcast, 1-16 = specific player
8-9     2     InReplyTo     iInRe            Message ID being replied to (for threading)
10-11   2     Length        cLen             Message byte count (negative = ASCII)
12+     Var   Message       rgbMsg           Stars! encoded message string
```

## Field Details

### Bytes 0-3: Linked List Pointer (Garbage)

The game writes the in-memory `MSGPLR` structure directly to file, including
the linked list pointer used for in-memory message management. These bytes
have no meaning in the file and should be ignored when reading.

### Bytes 8-9: iInRe (In-Reply-To)

This field is used for message threading. When a player replies to a message,
this contains the ID/index of the original message being replied to. A value
of 0 typically means the message is not a reply.

Note: Previously documented as "typically 3=reply, 4=normal" - those were
coincidentally the actual message IDs in test data, not semantic flags.

### Bytes 10-11: cLen (Message Length)

- **Positive value**: Message uses Stars! compressed string encoding
- **Negative value**: Message is plain ASCII (use `~cLen` or `-cLen-1` for actual length)

## Notes

- HST files do not contain message blocks
- In .x files, sender is always the file's player
- Player IDs in messages are offset: 0 = Everyone, 1-16 = Players 1-16

## Source

Decompiled from `IO::WriteRt` at 1070:947c and `MSG::FFinishPlrMsgEntry` at
1030:9bd6 in stars26jrc3.exe (Stars! 2.60j RC3).

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
