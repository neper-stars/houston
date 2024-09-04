package main

import (
	"fmt"
	"os"

	hs "github.com/neper-stars/houston"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Println("you should provide a filename as a positional argument")
		os.Exit(1)
	}
	// create an empty FileData struct
	var fd hs.FileData
	// ask houston to populate the data inside the struct
	// by reading from a given filename
	// this may return an error for a lot of different reasons
	// so always check your corners....
	if err := hs.ReadRawFile(os.Args[1], &fd); err != nil {
		fmt.Println("failed to open file:", err)
		os.Exit(1)
	}

	bl, err := fd.BlockList()
	if err != nil {
		fmt.Println("failed to iterate over blocks:", err.Error())
		os.Exit(1)
	}
	for _, b := range bl {
		switch b.BlockTypeID() {
		// we only care about the player block
		case hs.PlayerBlockType:
			pb, ok := b.(hs.PlayerBlock)
			if !ok {
				fmt.Println("failed to assert player block...")
				os.Exit(1)
			}
			if pb.Valid {
				fmt.Println("Player Block found")
				fmt.Println("Hashed password bytes:", pb.HashedPass())
				fmt.Println("Hashed password:", pb.HashedPass().Uint32())
				// searchArea := "aAbBcCdDeEfFgGhHiIjJkKlLmMnNoOpPqQrRsStTuUvVwWxXyYzZ+!.,"
				searchArea := "abcdefghijklmnopqrstuvwxyz"
				hs.GuessRacePassword(pb.HashedPass().Uint32(), 8, 1, searchArea, true)
			} else {
				fmt.Println("empty player block... nothing to report")
			}
		}
	}
}
