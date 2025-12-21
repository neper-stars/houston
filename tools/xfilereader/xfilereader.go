// Package xfilereader provides functionality to read and validate X (turn order) files.
//
// X files contain player orders for a turn, including waypoint changes,
// production queue modifications, research changes, and other commands.
//
// Example usage:
//
//	info, err := xfilereader.ReadFile("player1.x1")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Turn %d orders for player %d\n", info.Turn, info.PlayerIndex)
//	for _, order := range info.Orders {
//	    fmt.Println(order.Description)
//	}
package xfilereader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

// FileInfo contains information about an X file.
type FileInfo struct {
	Filename    string
	Size        int
	BlockCount  int
	GameID      uint32
	Turn        uint16
	Year        int
	PlayerIndex int
	IsSubmitted bool
	Orders      []Order
	BlockCounts map[string]int
}

// Order represents a single order in the X file.
type Order struct {
	Type        string
	Description string
	Block       blocks.Block
}

// ReadFile reads an X file and returns its contents.
func ReadFile(filename string) (*FileInfo, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ReadBytes(filename, fileBytes)
}

// ReadBytes parses X file data and returns its contents.
func ReadBytes(filename string, fileBytes []byte) (*FileInfo, error) {
	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if len(ext) < 2 || ext[1] != 'x' {
		return nil, fmt.Errorf("%s does not appear to be an X file", filename)
	}

	fd := parser.FileData(fileBytes)

	// Parse header
	header, err := fd.FileHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to parse file header: %w", err)
	}

	// Parse all blocks
	blockList, err := fd.BlockList()
	if err != nil {
		return nil, fmt.Errorf("failed to parse blocks: %w", err)
	}

	info := &FileInfo{
		Filename:    filename,
		Size:        len(fileBytes),
		BlockCount:  len(blockList),
		GameID:      header.GameID,
		Turn:        header.Turn,
		Year:        header.Year(),
		PlayerIndex: header.PlayerIndex(),
		Orders:      make([]Order, 0),
		BlockCounts: make(map[string]int),
	}

	// Process blocks
	for _, block := range blockList {
		order := extractOrder(block)
		if order != nil {
			info.Orders = append(info.Orders, *order)
		}

		// Count block types
		typeName := getBlockTypeName(block)
		info.BlockCounts[typeName]++

		// Check for submission
		if _, ok := block.(blocks.SaveAndSubmitBlock); ok {
			info.IsSubmitted = true
		}
	}

	return info, nil
}

// Validate checks if the X file is valid.
func (fi *FileInfo) Validate() error {
	if fi.BlockCount == 0 {
		return fmt.Errorf("file contains no blocks")
	}
	return nil
}

// OrderCount returns the number of orders in the file.
func (fi *FileInfo) OrderCount() int {
	return len(fi.Orders)
}

// GetOrders returns orders filtered by type.
func (fi *FileInfo) GetOrders(orderType string) []Order {
	var filtered []Order
	for _, order := range fi.Orders {
		if order.Type == orderType {
			filtered = append(filtered, order)
		}
	}
	return filtered
}

func extractOrder(block blocks.Block) *Order {
	switch b := block.(type) {
	case blocks.WaypointAddBlock:
		return &Order{
			Type:        "WaypointAdd",
			Description: fmt.Sprintf("Fleet %d: add waypoint to (%d, %d)", b.FleetNumber, b.X, b.Y),
			Block:       block,
		}

	case blocks.WaypointDeleteBlock:
		return &Order{
			Type:        "WaypointDelete",
			Description: fmt.Sprintf("Fleet %d: delete waypoint %d", b.FleetNumber, b.WaypointNumber),
			Block:       block,
		}

	case blocks.WaypointChangeTaskBlock:
		return &Order{
			Type:        "WaypointChangeTask",
			Description: fmt.Sprintf("Fleet %d: change task at waypoint %d", b.FleetNumber, b.WaypointNumber),
			Block:       block,
		}

	case blocks.ProductionQueueChangeBlock:
		return &Order{
			Type:        "ProductionQueueChange",
			Description: fmt.Sprintf("Planet %d: update production queue (%d items)", b.PlanetId, len(b.Items)),
			Block:       block,
		}

	case blocks.DesignChangeBlock:
		if b.IsDelete {
			designType := "ship"
			if b.IsStarbase {
				designType = "starbase"
			}
			return &Order{
				Type:        "DesignChange",
				Description: fmt.Sprintf("Delete %s design %d", designType, b.DesignToDelete),
				Block:       block,
			}
		}
		return &Order{
			Type:        "DesignChange",
			Description: "Update design",
			Block:       block,
		}

	case blocks.FleetSplitBlock:
		return &Order{
			Type:        "FleetSplit",
			Description: fmt.Sprintf("Split fleet %d", b.FleetNumber),
			Block:       block,
		}

	case blocks.FleetsMergeBlock:
		return &Order{
			Type:        "FleetsMerge",
			Description: fmt.Sprintf("Merge %d fleets into fleet %d", len(b.FleetsToMerge), b.FleetNumber),
			Block:       block,
		}

	case blocks.ResearchChangeBlock:
		return &Order{
			Type:        "ResearchChange",
			Description: "Change research priority",
			Block:       block,
		}

	case blocks.PlanetChangeBlock:
		return &Order{
			Type:        "PlanetChange",
			Description: "Change planet settings",
			Block:       block,
		}

	case blocks.ChangePasswordBlock:
		return &Order{
			Type:        "ChangePassword",
			Description: "Change race password",
			Block:       block,
		}

	case blocks.PlayersRelationChangeBlock:
		return &Order{
			Type:        "PlayersRelationChange",
			Description: "Change player relation",
			Block:       block,
		}

	case blocks.SaveAndSubmitBlock:
		return &Order{
			Type:        "SaveAndSubmit",
			Description: "Turn submitted",
			Block:       block,
		}
	}

	return nil
}

func getBlockTypeName(block blocks.Block) string {
	switch block.(type) {
	case blocks.FileHeader:
		return "FileHeader"
	case blocks.FileFooterBlock:
		return "FileFooter"
	case blocks.WaypointAddBlock:
		return "WaypointAdd"
	case blocks.WaypointDeleteBlock:
		return "WaypointDelete"
	case blocks.WaypointChangeTaskBlock:
		return "WaypointChangeTask"
	case blocks.ProductionQueueChangeBlock:
		return "ProductionQueueChange"
	case blocks.DesignChangeBlock:
		return "DesignChange"
	case blocks.FleetSplitBlock:
		return "FleetSplit"
	case blocks.FleetsMergeBlock:
		return "FleetsMerge"
	case blocks.ResearchChangeBlock:
		return "ResearchChange"
	case blocks.PlanetChangeBlock:
		return "PlanetChange"
	case blocks.ChangePasswordBlock:
		return "ChangePassword"
	case blocks.PlayersRelationChangeBlock:
		return "PlayersRelationChange"
	case blocks.SaveAndSubmitBlock:
		return "SaveAndSubmit"
	default:
		return "Other"
	}
}
