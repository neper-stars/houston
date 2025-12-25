package store

import "github.com/neper-stars/houston/blocks"

// EventsEntity represents the events block for a turn.
// Events are stored per-source rather than merged, as they're turn-specific.
type EventsEntity struct {
	// Production events
	ProductionEvents []blocks.ProductionEvent

	// Research events
	ResearchEvents []blocks.ResearchCompleteEvent

	// Tech benefits gained
	TechBenefits []blocks.TechBenefitEvent

	// Terraformable planets found
	TerraformablePlanets []blocks.TerraformablePlanetFoundEvent

	// Population changes
	PopulationChanges []blocks.PopulationChangeEvent

	// Packet events
	PacketsCaptured    []blocks.PacketCapturedEvent
	PacketsProduced    []blocks.MineralPacketProducedEvent
	PacketBombardments []blocks.PacketBombardmentEvent

	// Construction events
	StarbasesBuilt []blocks.StarbaseBuiltEvent

	// Random events
	CometStrikes     []blocks.CometStrikeEvent
	StrangeArtifacts []blocks.StrangeArtifactEvent

	// Colony events
	NewColonies []blocks.NewColonyEvent

	// Fleet scrapped events
	FleetsScrapped           []blocks.FleetScrappedEvent
	FleetsScrappedAtStarbase []blocks.FleetScrappedAtStarbaseEvent
	FleetsScrappedInSpace    []blocks.FleetScrappedInSpaceEvent

	// Battle events
	Battles []blocks.BattleEvent

	// Source tracking
	Source *FileSource
	Turn   uint16

	// Raw block
	eventsBlock *blocks.EventsBlock
}

// newEventsEntityFromBlock creates an EventsEntity from an EventsBlock.
func newEventsEntityFromBlock(eb *blocks.EventsBlock, source *FileSource) *EventsEntity {
	return &EventsEntity{
		ProductionEvents:         eb.ProductionEvents,
		ResearchEvents:           eb.ResearchEvents,
		TechBenefits:             eb.TechBenefits,
		TerraformablePlanets:     eb.TerraformablePlanets,
		PopulationChanges:        eb.PopulationChanges,
		PacketsCaptured:          eb.PacketsCaptured,
		PacketsProduced:          eb.PacketsProduced,
		PacketBombardments:       eb.PacketBombardments,
		StarbasesBuilt:           eb.StarbasesBuilt,
		CometStrikes:             eb.CometStrikes,
		StrangeArtifacts:         eb.StrangeArtifacts,
		NewColonies:              eb.NewColonies,
		FleetsScrapped:           eb.FleetsScrapped,
		FleetsScrappedAtStarbase: eb.FleetsScrappedAtStarbase,
		FleetsScrappedInSpace:    eb.FleetsScrappedInSpace,
		Battles:                  eb.Battles,
		Source:                   source,
		Turn:                     source.Turn,
		eventsBlock:              eb,
	}
}

// HasBattles returns true if there were battles this turn.
func (e *EventsEntity) HasBattles() bool {
	return len(e.Battles) > 0
}

// HasResearch returns true if research was completed this turn.
func (e *EventsEntity) HasResearch() bool {
	return len(e.ResearchEvents) > 0
}

// TotalEvents returns the total number of events.
func (e *EventsEntity) TotalEvents() int {
	return len(e.ProductionEvents) +
		len(e.ResearchEvents) +
		len(e.TechBenefits) +
		len(e.TerraformablePlanets) +
		len(e.PopulationChanges) +
		len(e.PacketsCaptured) +
		len(e.PacketsProduced) +
		len(e.PacketBombardments) +
		len(e.StarbasesBuilt) +
		len(e.CometStrikes) +
		len(e.StrangeArtifacts) +
		len(e.NewColonies) +
		len(e.FleetsScrapped) +
		len(e.FleetsScrappedAtStarbase) +
		len(e.FleetsScrappedInSpace) +
		len(e.Battles)
}
