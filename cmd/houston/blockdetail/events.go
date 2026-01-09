package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
)

func init() {
	RegisterFormatter(blocks.EventsBlockType, FormatEvents)
}

// eventTypeName returns human-readable name for event type
func eventTypeName(eventType int) string {
	names := map[int]string{
		blocks.EventTypeDefensesBuilt:            "DefensesBuilt",
		blocks.EventTypeFactoriesBuilt:           "FactoriesBuilt",
		blocks.EventTypeMineralAlchemyBuilt:      "MineralAlchemy",
		blocks.EventTypeMinesBuilt:               "MinesBuilt",
		blocks.EventTypeQueueEmpty:               "QueueEmpty",
		blocks.EventTypePopulationChange:         "PopulationChange",
		blocks.EventTypeResearchComplete:         "ResearchComplete",
		blocks.EventTypeTerraformablePlanetFound: "TerraformableFound",
		blocks.EventTypeTechBenefit:              "TechBenefit",
		blocks.EventTypePacketCaptured:           "PacketCaptured",
		blocks.EventTypeMineralPacketProduced:    "PacketProduced",
		blocks.EventTypePacketBombardment:        "PacketBombardment",
		blocks.EventTypeStarbaseBuilt:            "StarbaseBuilt",
		blocks.EventTypeNewColony:                "NewColony",
		blocks.EventTypeStrangeArtifact:          "StrangeArtifact",
		blocks.EventTypeFleetScrapped:            "FleetScrapped",
		blocks.EventTypeFleetScrappedAtStarbase:  "FleetScrappedAtStarbase",
		blocks.EventTypeFleetScrappedInSpace:     "FleetScrappedInSpace",
		blocks.EventTypeBattle:                   "Battle",
	}
	if name, ok := names[eventType]; ok {
		return name
	}
	// Handle comet strike event types (0x83-0x8a)
	if eventType >= blocks.EventTypeCometStrikeFirst && eventType <= blocks.EventTypeCometStrikeLast {
		return "CometStrike"
	}
	return fmt.Sprintf("Unknown(0x%02X)", eventType)
}

// FormatEvents provides detailed view for EventsBlock (type 12)
func FormatEvents(block blocks.Block, index int) string {
	width := DefaultWidth
	eb, ok := block.(blocks.EventsBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := eb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 5 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Event counts summary
	totalEvents := len(eb.ProductionEvents) + len(eb.ResearchEvents) + len(eb.TechBenefits) +
		len(eb.TerraformablePlanets) + len(eb.PopulationChanges) + len(eb.PacketsCaptured) +
		len(eb.PacketsProduced) + len(eb.PacketBombardments) + len(eb.StarbasesBuilt) +
		len(eb.CometStrikes) + len(eb.NewColonies) + len(eb.StrangeArtifacts) +
		len(eb.FleetsScrapped) + len(eb.FleetsScrappedAtStarbase) + len(eb.FleetsScrappedInSpace) +
		len(eb.Battles)

	fields = append(fields, fmt.Sprintf("Total parsed events: %d", totalEvents))
	fields = append(fields, fmt.Sprintf("Raw data size: %d bytes", len(d)))

	// Production Events
	if len(eb.ProductionEvents) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Production Events (%d) ──", len(eb.ProductionEvents)))
		for i, ev := range eb.ProductionEvents {
			prefix := TreeBranch
			if i == len(eb.ProductionEvents)-1 {
				prefix = TreeEnd
			}
			countStr := ""
			if ev.Count > 0 {
				countStr = fmt.Sprintf(" (count=%d)", ev.Count)
			}
			fields = append(fields, fmt.Sprintf("  %s 0x%02X %s @ Planet #%d%s",
				prefix, ev.EventType, eventTypeName(ev.EventType), ev.PlanetID+1, countStr))
		}
	}

	// Research Events
	if len(eb.ResearchEvents) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Research Complete Events (%d) ──", len(eb.ResearchEvents)))
		for i, ev := range eb.ResearchEvents {
			prefix := TreeBranch
			if i == len(eb.ResearchEvents)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s %s Level %d achieved, next: %s",
				prefix, blocks.ResearchFieldName(ev.Field), ev.Level, blocks.ResearchFieldName(ev.NextField)))
		}
	}

	// Tech Benefits
	if len(eb.TechBenefits) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Tech Benefits (%d) ──", len(eb.TechBenefits)))
		for i, ev := range eb.TechBenefits {
			prefix := TreeBranch
			if i == len(eb.TechBenefits)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s Category %d, Item ID %d",
				prefix, ev.Category, ev.ItemID))
		}
	}

	// Terraformable Planets Found
	if len(eb.TerraformablePlanets) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Terraformable Planets Found (%d) ──", len(eb.TerraformablePlanets)))
		for i, ev := range eb.TerraformablePlanets {
			prefix := TreeBranch
			if i == len(eb.TerraformablePlanets)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s Growth rate: %.1f%% (encoded=%d)",
				prefix, ev.GrowthRatePercent, ev.GrowthRateEncoded))
		}
	}

	// Population Changes
	if len(eb.PopulationChanges) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Population Changes (%d) ──", len(eb.PopulationChanges)))
		for i, ev := range eb.PopulationChanges {
			prefix := TreeBranch
			if i == len(eb.PopulationChanges)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s Planet #%d: %d colonists",
				prefix, ev.PlanetID+1, ev.Amount))
		}
	}

	// Packets Captured
	if len(eb.PacketsCaptured) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Packets Captured (%d) ──", len(eb.PacketsCaptured)))
		for i, ev := range eb.PacketsCaptured {
			prefix := TreeBranch
			if i == len(eb.PacketsCaptured)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s Planet #%d: %d kT minerals",
				prefix, ev.PlanetID+1, ev.MineralAmount))
		}
	}

	// Packets Produced
	if len(eb.PacketsProduced) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Packets Produced (%d) ──", len(eb.PacketsProduced)))
		for i, ev := range eb.PacketsProduced {
			prefix := TreeBranch
			if i == len(eb.PacketsProduced)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s Source #%d -> Dest #%d",
				prefix, ev.SourcePlanetID+1, ev.DestinationPlanetID+1))
		}
	}

	// Packet Bombardments
	if len(eb.PacketBombardments) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Packet Bombardments (%d) ──", len(eb.PacketBombardments)))
		for i, ev := range eb.PacketBombardments {
			prefix := TreeBranch
			if i == len(eb.PacketBombardments)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s Planet #%d: %d kT, %d killed",
				prefix, ev.PlanetID+1, ev.MineralAmount, ev.ColonistsKilled))
		}
	}

	// Starbases Built
	if len(eb.StarbasesBuilt) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Starbases Built (%d) ──", len(eb.StarbasesBuilt)))
		for i, ev := range eb.StarbasesBuilt {
			prefix := TreeBranch
			if i == len(eb.StarbasesBuilt)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s Planet #%d (design=%d)",
				prefix, ev.PlanetID+1, ev.DesignInfo))
		}
	}

	// Comet Strikes
	if len(eb.CometStrikes) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Comet Strikes (%d) ──", len(eb.CometStrikes)))
		for i, ev := range eb.CometStrikes {
			prefix := TreeBranch
			if i == len(eb.CometStrikes)-1 {
				prefix = TreeEnd
			}
			ownerStr := "unowned"
			if ev.IsOwnedPlanet() {
				ownerStr = fmt.Sprintf("owned, %d%% deaths", ev.DeathPercent())
			}
			habChanges := ev.ChangedHabNames()
			habStr := "none"
			if len(habChanges) > 0 {
				habStr = fmt.Sprintf("%v", habChanges)
			}
			fields = append(fields, fmt.Sprintf("  %s Planet #%d: %s comet (%s)",
				prefix, ev.PlanetID+1, ev.CometSizeName(), ownerStr))
			fields = append(fields, fmt.Sprintf("       Hab changes: %s (mask=0x%02X)",
				habStr, ev.HabChangeMask))
		}
	}

	// New Colonies
	if len(eb.NewColonies) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── New Colonies (%d) ──", len(eb.NewColonies)))
		for i, ev := range eb.NewColonies {
			prefix := TreeBranch
			if i == len(eb.NewColonies)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s Planet #%d", prefix, ev.PlanetID+1))
		}
	}

	// Strange Artifacts
	if len(eb.StrangeArtifacts) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Strange Artifacts (%d) ──", len(eb.StrangeArtifacts)))
		for i, ev := range eb.StrangeArtifacts {
			prefix := TreeBranch
			if i == len(eb.StrangeArtifacts)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s Planet #%d: +%d %s research",
				prefix, ev.PlanetID+1, ev.BoostAmount, blocks.ResearchFieldName(ev.ResearchField)))
		}
	}

	// Fleets Scrapped
	if len(eb.FleetsScrapped) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Fleets Scrapped at Planet (%d) ──", len(eb.FleetsScrapped)))
		for i, ev := range eb.FleetsScrapped {
			prefix := TreeBranch
			if i == len(eb.FleetsScrapped)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s Fleet #%d @ Planet #%d: %d kT recovered",
				prefix, ev.FleetIndex+1, ev.PlanetID+1, ev.MineralAmount))
		}
	}

	// Fleets Scrapped at Starbase
	if len(eb.FleetsScrappedAtStarbase) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Fleets Scrapped at Starbase (%d) ──", len(eb.FleetsScrappedAtStarbase)))
		for i, ev := range eb.FleetsScrappedAtStarbase {
			prefix := TreeBranch
			if i == len(eb.FleetsScrappedAtStarbase)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s Fleet #%d @ Planet #%d: %d kT mass",
				prefix, ev.FleetIndex+1, ev.PlanetID+1, ev.FleetMass))
		}
	}

	// Fleets Scrapped in Space
	if len(eb.FleetsScrappedInSpace) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Fleets Scrapped in Space (%d) ──", len(eb.FleetsScrappedInSpace)))
		for i, ev := range eb.FleetsScrappedInSpace {
			prefix := TreeBranch
			if i == len(eb.FleetsScrappedInSpace)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("  %s Salvage object ID=0x%04X (subtype=0x%02X)",
				prefix, ev.SalvageObjectID, ev.Subtype))
		}
	}

	// Battles
	if len(eb.Battles) > 0 {
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("── Battles (%d) ──", len(eb.Battles)))
		for i, ev := range eb.Battles {
			prefix := TreeBranch
			if i == len(eb.Battles)-1 {
				prefix = TreeEnd
			}
			result := "draw"
			if ev.YouSurvived && !ev.EnemySurvived {
				result = "victory"
			} else if !ev.YouSurvived && ev.EnemySurvived {
				result = "defeat"
			}
			recStr := ""
			if ev.HasRecording {
				recStr = " [recording]"
			}
			fields = append(fields, fmt.Sprintf("  %s Planet #%d vs Player %d: %s%s",
				prefix, ev.PlanetID+1, ev.EnemyPlayer+1, result, recStr))
			fields = append(fields, fmt.Sprintf("       Your forces: %d (lost %d) | Enemy forces: %d (lost %d)",
				ev.YourForces, ev.YourLosses, ev.EnemyForces, ev.EnemyLosses))
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	if len(eb.ProductionEvents) > 0 {
		fields = append(fields, fmt.Sprintf("  Production events: %d", len(eb.ProductionEvents)))
	}
	if len(eb.ResearchEvents) > 0 {
		fields = append(fields, fmt.Sprintf("  Research events: %d", len(eb.ResearchEvents)))
	}
	if len(eb.Battles) > 0 {
		fields = append(fields, fmt.Sprintf("  Battles: %d", len(eb.Battles)))
	}
	if totalEvents == 0 {
		fields = append(fields, "  No events parsed (may contain unknown event types)")
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
