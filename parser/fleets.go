package parser

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
)

// FleetInfo combines fleet data with its display name.
// This provides a higher-level view of fleet data by associating
// FleetNameBlocks with their corresponding FleetBlocks.
type FleetInfo struct {
	Fleet         *blocks.FleetBlock  // The fleet data
	CustomName    string              // Custom name if set, empty otherwise
	HasCustomName bool                // True if fleet has a custom name
	PrimaryDesign *blocks.DesignBlock // Primary design (first in ShipTypes bitmask)
}

// Name returns the fleet's display name.
// If the fleet has a custom name, it returns that.
// Otherwise, it generates a default name using the primary design name
// and fleet number in the format "{designName} #{fleetNumber+1}".
// If no design is available, returns "Fleet #{fleetNumber+1}".
func (fi *FleetInfo) Name() string {
	if fi.HasCustomName {
		return fi.CustomName
	}
	designName := ""
	if fi.PrimaryDesign != nil {
		designName = fi.PrimaryDesign.Name
	}
	if designName == "" {
		return fmt.Sprintf("Fleet #%d", fi.Fleet.FleetNumber+1)
	}
	return fmt.Sprintf("%s #%d", designName, fi.Fleet.FleetNumber+1)
}

// DesignMap maps design slot numbers (0-15) to DesignBlocks.
type DesignMap map[int]*blocks.DesignBlock

// ExtractDesigns extracts all ship designs from a block list.
// Returns a map from design slot number to DesignBlock.
func ExtractDesigns(blockList []blocks.Block) DesignMap {
	designs := make(DesignMap)
	for _, block := range blockList {
		if db, ok := block.(blocks.DesignBlock); ok {
			if !db.IsStarbase {
				designs[db.DesignNumber] = &db
			}
		}
	}
	return designs
}

// getPrimaryDesignSlot returns the first (lowest) design slot set in the ShipTypes bitmask.
// Returns -1 if no designs are set.
func getPrimaryDesignSlot(shipTypes uint16) int {
	for i := 0; i < 16; i++ {
		if (shipTypes & (1 << i)) != 0 {
			return i
		}
	}
	return -1
}

// ExtractFleets extracts all fleets from a block list, associating
// FleetNameBlocks with their corresponding FleetBlocks.
//
// In Stars! M files, a FleetNameBlock (Type 21) immediately follows
// the FleetBlock (Type 16) it names. Fleets without a following
// FleetNameBlock have auto-generated names.
//
// This version does not resolve design names. Use ExtractFleetsWithDesigns
// for automatic design name resolution.
func ExtractFleets(blockList []blocks.Block) []*FleetInfo {
	return ExtractFleetsWithDesigns(blockList, nil)
}

// ExtractFleetsWithDesigns extracts all fleets from a block list, associating
// FleetNameBlocks with their corresponding FleetBlocks and resolving design names.
//
// If designs is nil, design names will not be resolved (PrimaryDesign will be nil).
// Use ExtractDesigns to get a DesignMap from the same block list.
func ExtractFleetsWithDesigns(blockList []blocks.Block, designs DesignMap) []*FleetInfo {
	var fleets []*FleetInfo
	var lastFleetInfo *FleetInfo

	for _, block := range blockList {
		switch b := block.(type) {
		case blocks.FleetBlock:
			info := &FleetInfo{
				Fleet: &b,
			}
			// Resolve primary design
			if designs != nil {
				slot := getPrimaryDesignSlot(b.ShipTypes)
				if slot >= 0 {
					info.PrimaryDesign = designs[slot]
				}
			}
			fleets = append(fleets, info)
			lastFleetInfo = info
		case blocks.FleetNameBlock:
			// FleetNameBlock follows the FleetBlock it names
			if lastFleetInfo != nil {
				lastFleetInfo.CustomName = b.Name
				lastFleetInfo.HasCustomName = true
			}
			lastFleetInfo = nil // Clear after assigning name
		case blocks.WaypointBlock, blocks.WaypointTaskBlock:
			// Waypoint blocks can appear between FleetBlock and FleetNameBlock
			// Don't clear lastFleetInfo
		default:
			// Any other block type ends the fleet context
			lastFleetInfo = nil
		}
	}

	return fleets
}

// ExtractFleetsMap returns a map of fleet number to FleetInfo for quick lookup.
func ExtractFleetsMap(blockList []blocks.Block) map[int]*FleetInfo {
	fleets := ExtractFleets(blockList)
	result := make(map[int]*FleetInfo, len(fleets))
	for _, fi := range fleets {
		result[fi.Fleet.FleetNumber] = fi
	}
	return result
}

// ExtractFleetsMapWithDesigns returns a map of fleet number to FleetInfo with design resolution.
func ExtractFleetsMapWithDesigns(blockList []blocks.Block, designs DesignMap) map[int]*FleetInfo {
	fleets := ExtractFleetsWithDesigns(blockList, designs)
	result := make(map[int]*FleetInfo, len(fleets))
	for _, fi := range fleets {
		result[fi.Fleet.FleetNumber] = fi
	}
	return result
}

// ExtractAllFleetInfo is a convenience function that extracts both designs and fleets
// in one call, returning fleets with fully resolved names.
func ExtractAllFleetInfo(blockList []blocks.Block) []*FleetInfo {
	designs := ExtractDesigns(blockList)
	return ExtractFleetsWithDesigns(blockList, designs)
}
