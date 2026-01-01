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

	/*
		// if you only need the header, let's say because you only want
		// to see if an .xN file is being submitted or just saved
		// then you can call FileHeader() on the raw FileData
		fh, err := fd.FileHeader()
		if err != nil {
			fmt.Println("failed to parse file header", err.Error())
			os.Exit(1)
		}

		// the FileHeader struct has handy methods to get useful info
		// like the:
		//   - player index (beware player1 will be numbered 0)
		//   - the turn number
		//   - the year (2400 + turn number)
		fmt.Println("Player #:", fh.PlayerIndex())
		fmt.Println("Turn Submitted:", fh.TurnSubmitted())
		fmt.Println("Year:", fh.Year())
	*/

	// if you want to iterate over all blocks
	bl, err := fd.BlockList()
	if err != nil {
		fmt.Println("failed to iterate over blocks:", err.Error())
		os.Exit(1)
	}
	fmt.Println("************")
	fmt.Println("Blocks found")
	fmt.Println("************")
	for _, b := range bl {
		switch b.BlockTypeID() {
		case hs.FileHeaderBlockType:
			fh, ok := b.(hs.FileHeader)
			if !ok {
				fmt.Println("failed to assert fileHeader...")
				os.Exit(1)
			}
			fmt.Println("File Header found")
			fmt.Println("	Year is:", fh.Year())
			fmt.Println("	Player index is:", fh.PlayerIndex())
			fmt.Println("	Turn Submitted:", fh.TurnSubmitted())
			fmt.Printf("	Game ID is: %d\n", fh.GameID)
		case hs.PlayerBlockType:
			pb, ok := b.(hs.PlayerBlock)
			if !ok {
				fmt.Println("failed to assert player block...")
				os.Exit(1)
			}
			if pb.Valid {
				fmt.Println("Player Block found")
				// fmt.Println("Hashed password bytes:", pb.HashedPass())
				// fmt.Println("Hashed password:", pb.HashedPass().Uint32())
				fmt.Println("	player index:", pb.PlayerNumber)
				fmt.Println("	race plural name:", pb.NamePlural)
				fmt.Println("	race singular name:", pb.NameSingular)
			} else {
				fmt.Println("empty player block... nothing to report")
			}

		default:
			// use the decrypted version
			// fmt.Println("-->", b.BlockTypeID(), b.BlockSize(), b.DecryptedData())
		}
	}
}
