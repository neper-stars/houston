# Houston

This is a golang library to read and manipulate Stars! game files.

It is intended to be used as a library for the Neper project to allow easy
hosting a multiplayer games with a nice desktop client (Astrum).

This library would not exit without the inspiration of multiple people
who put the groundwork for understanding the stars file format. I "only"
reverserd a few more bits and bytes from a few blocks. Not all is complete
per exemple the battle recording still is a mistery even if some progress
was made.

At the moment we understand most of the blocks, we now how to decypher the
data.

The houston library exposes many additional tools:

  - a map renderer that can ready an .mN file
    (with its corresponding .xy file and generate a nice
    svg map with many optional displays.
  - a .mN file merger
  - a password recovery bruteforcer in case a player drops from the game
    and we have a replacement player...

The library has been optimized for readability with two different layers.

  - The first, low-level layer (blocks) exposes
    the inner guts of the file structure for Stars! game files.
  - The second, a high level API (store) insulates the user from the
    intricaties of the file format and tries to propose a high level
    logical API to "view" and manipulate the game "state"

# Acknowldgements:

As said above this lib would not exist without the inspiration from:

  - https://github.com/stars-4x/starsapi
  - https://github.com/ricks03/TotalHost

Those two projects were tremendously usefull in understanding the stars!
encryption, the stars fileheader algorithm and many different packets
bits that were resisting analysis.

