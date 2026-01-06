package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/cmd/houston/blockdetail"
	"github.com/neper-stars/houston/parser"
)

type blocksCommand struct {
	Detailed bool   `short:"d" long:"detailed" description:"Show detailed ASCII schema for each block"`
	Filter   string `short:"f" long:"filter" description:"Filter by block type IDs (comma-separated, e.g. '8,6' for FileHeader and Player)"`
	Args     struct {
		File string `positional-arg-name:"file" description:"Stars! game file to read" required:"true"`
	} `positional-args:"yes"`
}

func (c *blocksCommand) Execute(args []string) error {
	fileBytes, err := os.ReadFile(c.Args.File)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fd := parser.FileData(fileBytes)

	blockList, err := fd.BlockList()
	if err != nil {
		return fmt.Errorf("failed to parse blocks: %w", err)
	}

	// Parse filter if provided
	filterSet := make(map[blocks.BlockTypeID]bool)
	if c.Filter != "" {
		for _, part := range strings.Split(c.Filter, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			typeID, err := strconv.ParseUint(part, 10, 16)
			if err != nil {
				return fmt.Errorf("invalid block type ID in filter: %q", part)
			}
			filterSet[blocks.BlockTypeID(typeID)] = true
		}
	}

	fmt.Printf("File: %s (%d bytes)\n", c.Args.File, len(fileBytes))
	fmt.Printf("Blocks: %d\n\n", len(blockList))

	for i, block := range blockList {
		// Apply filter if set
		if len(filterSet) > 0 && !filterSet[block.BlockTypeID()] {
			continue
		}
		if c.Detailed {
			fmt.Print(blockdetail.FormatDetailed(block, i))
			fmt.Println()
		} else {
			typeID := block.BlockTypeID()
			typeName := blocks.BlockTypeName(typeID)
			size := block.BlockSize()

			fmt.Printf("Block %d: %s (type=%d, size=%d)\n", i, typeName, typeID, size)

			decrypted := block.DecryptedData()
			if len(decrypted) > 0 {
				fmt.Printf("  Data: %s\n", hex.EncodeToString(decrypted))
			}

			printBlockDetails(block)
			fmt.Println()
		}
	}

	return nil
}

func printBlockDetails(block blocks.Block) {
	switch b := block.(type) {
	case blocks.FileHeader:
		fmt.Printf("  GameID: %d, Turn: %d (Year %d), Player: %d\n",
			b.GameID, b.Turn, b.Year(), b.PlayerIndex())
	case blocks.PlanetsBlock:
		fmt.Printf("  PlanetCount: %d\n", b.GetPlanetCount())
	case blocks.PlanetBlock:
		fmt.Printf("  PlanetNumber: %d, Owner: %d\n", b.PlanetNumber, b.Owner)
	case blocks.PartialPlanetBlock:
		fmt.Printf("  PlanetNumber: %d, Owner: %d\n", b.PlanetNumber, b.Owner)
	case blocks.FleetBlock:
		fmt.Printf("  FleetNumber: %d, Owner: %d, X: %d, Y: %d\n",
			b.FleetNumber, b.Owner, b.X, b.Y)
	case blocks.PartialFleetBlock:
		fmt.Printf("  FleetNumber: %d, Owner: %d, X: %d, Y: %d\n",
			b.FleetNumber, b.Owner, b.X, b.Y)
	case blocks.DesignBlock:
		fmt.Printf("  DesignNumber: %d, HullID: %d, Name: %s\n",
			b.DesignNumber, b.HullId, b.Name)
	case blocks.CountersBlock:
		fmt.Printf("  Planets: %d, Fleets: %d\n", b.PlanetCount, b.FleetCount)
	case blocks.MessageBlock:
		fmt.Printf("  From: %d, To: %d\n", b.SenderId, b.ReceiverId)
	case blocks.ObjectBlock:
		fmt.Printf("  ObjectType: %d, Owner: %d\n", b.ObjectType, b.Owner)
	}
}

func addBlocksCommand(parser *flags.Parser) {
	_, err := parser.AddCommand("blocks",
		"Display blocks in a Stars! file",
		"Reads a Stars! game file and displays its decrypted blocks.\n\n"+
			"This tool is useful for debugging and understanding Stars! file structure.\n"+
			"It displays each block with its type ID and hex-encoded decrypted data.\n"+
			"For certain block types (FileHeader, Planets, Planet, Fleet, Design),\n"+
			"it also shows the parsed structure.",
		&blocksCommand{})
	if err != nil {
		panic(err)
	}
}
