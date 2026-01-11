package blocks

// MessageTypeID represents a message type ID used in the MessagesFilterBlock bitmap.
// These IDs correspond to entries in the Stars! message string table.
type MessageTypeID int

// Message type ID constants for in-game turn messages.
// These can be used with MessagesFilterBlock.IsFiltered() and SetFiltered().
const (
	// Battle messages
	MsgBattleAftermath     MessageTypeID = 0x20 // Battle aftermath - great damage before destroyed
	MsgBattleWatched       MessageTypeID = 0x21 // Battle - watched forces annihilate each other
	MsgBattleObserved      MessageTypeID = 0x22 // Battle - observed forces defeating
	MsgBattleVCR           MessageTypeID = 0x7E // Battle took place - VCR recording available
	MsgBattleOutcomeStart  MessageTypeID = 0x8D // Start of battle outcome messages
	MsgBattleOutcomeEnd    MessageTypeID = 0xA8 // End of battle outcome messages
	MsgBattleNotInvolved1  MessageTypeID = 0xF9 // Colony observed battle (not involved)
	MsgBattleNotInvolved2  MessageTypeID = 0xFA // Fleet observed battle (not involved)
	MsgBattleExtendedStart MessageTypeID = 0x113
	MsgBattleExtendedEnd   MessageTypeID = 0x117

	// Colony/Population messages
	MsgColonistsDied       MessageTypeID = 0x23 // All colonists died - lost planet
	MsgColonistsOrbitDied  MessageTypeID = 0x24 // All colonists orbiting died - lost starbase
	MsgPopulationDecreased MessageTypeID = 0x25 // Population decreased (from X to Y)
	MsgPopulationOvercrowd MessageTypeID = 0x26 // Population decreased due to overcrowding
	MsgColonistsJumpedShip MessageTypeID = 0x64 // Colonists jumped ship - lost planet
	MsgColonistsAbandoned  MessageTypeID = 0x65 // Colonists abandoned starbase

	// Fleet/Fuel messages
	MsgFleetOutOfFuel        MessageTypeID = 0x27 // Fleet ran out of fuel
	MsgWaypointDestroyed     MessageTypeID = 0x28 // Waypoint destroyed/disappeared
	MsgFleetDuckedBehind     MessageTypeID = 0x29 // Fleet ducked behind planet
	MsgFleetOutranScanners   MessageTypeID = 0x2A // Fleet outran scanners
	MsgFleetFuelSpeedReduced MessageTypeID = 0x8B // Out of fuel - speed decreased
	MsgRamScoopFuel          MessageTypeID = 0xF3 // Ram scoops produced fuel

	// Cargo transfer messages
	MsgCargoLoaded        MessageTypeID = 0x2B // Fleet loaded cargo from planet
	MsgCargoBeamedFrom    MessageTypeID = 0x2C // Fleet beamed cargo from planet
	MsgCargoUnloaded      MessageTypeID = 0x2D // Fleet unloaded cargo to planet
	MsgCargoBeamedTo      MessageTypeID = 0x2E // Fleet beamed cargo to planet
	MsgFleetTransferStart MessageTypeID = 0x42 // Start of fleet-to-fleet transfer messages
	MsgFleetTransferEnd   MessageTypeID = 0x4D // End of fleet-to-fleet transfer messages
	MsgCargoFromFleet1    MessageTypeID = 0x79 // Loaded cargo from another fleet
	MsgCargoFromFleet2    MessageTypeID = 0x7A // Beamed cargo from another fleet

	// Production messages
	MsgStarbaseBuiltShip    MessageTypeID = 0x2F // Starbase built a ship
	MsgStarbaseBuiltShips   MessageTypeID = 0x30 // Starbase built multiple ships
	MsgStarbaseBuiltRouted1 MessageTypeID = 0x31 // Built ship routed to planet
	MsgStarbaseBuiltRouted2 MessageTypeID = 0x32 // Built ships routed to planet
	MsgStarbaseBuiltNoFuel1 MessageTypeID = 0x33 // Built ship not routed (no fuel)
	MsgStarbaseBuiltNoFuel2 MessageTypeID = 0x34 // Built ships not routed (no fuel)
	MsgBuiltFactory         MessageTypeID = 0x35 // Built a factory
	MsgBuiltFactories       MessageTypeID = 0x36 // Built factories
	MsgBuiltMine            MessageTypeID = 0x37 // Built a mine
	MsgBuiltMines           MessageTypeID = 0x38 // Built mines
	MsgBuiltDefense         MessageTypeID = 0x39 // Built a defense
	MsgBuiltDefenses        MessageTypeID = 0x3A // Built defenses
	MsgQueueCompleted       MessageTypeID = 0x3E // Production queue completed
	MsgQueueEmpty           MessageTypeID = 0x3F // Production queue is empty
	MsgBuiltScanner         MessageTypeID = 0x7C // Built planetary scanner
	MsgBuiltStarbase1       MessageTypeID = 0xCD // Built a starbase
	MsgBuiltStarbase2       MessageTypeID = 0xCE // Built starbase (with ship capacity)
	MsgBuiltStarbase3       MessageTypeID = 0xCF // Built starbase (unlimited capacity)

	// Research messages
	MsgResearchComplete  MessageTypeID = 0x50  // Research completed - tech level X
	MsgTechBenefit       MessageTypeID = 0x5F  // Tech breakthrough benefit
	MsgNewHullUnlocked   MessageTypeID = 0x78  // Breakthrough - new hull type
	MsgResearchComplete2 MessageTypeID = 0x136 // Research completed (alternate)

	// Colonization error messages
	MsgColonizeNotInOrbit  MessageTypeID = 0x51 // Colonize order - not in orbit
	MsgColonizeAlreadyPop  MessageTypeID = 0x52 // Colonize - planet already populated
	MsgColonizeNoColonists MessageTypeID = 0x53 // Colonize - no colonists
	MsgColonizeNoModule    MessageTypeID = 0x54 // Colonize - no colony module
	MsgColonizeErrorsStart MessageTypeID = 0x55 // Start of other colonization errors
	MsgColonizeErrorsEnd   MessageTypeID = 0x58 // End of colonization errors

	// Scrap messages
	MsgFleetScrappedStart  MessageTypeID = 0x59  // Fleet scrapped for minerals
	MsgFleetScrappedEnd    MessageTypeID = 0x5D  // End of scrap messages
	MsgFleetScrapped2Start MessageTypeID = 0x140 // Fleet scrapped (alternate)
	MsgFleetScrapped2End   MessageTypeID = 0x143 // Fleet scrapped (alternate end)

	// Strange artifact
	MsgArtifactFound MessageTypeID = 0x5E // Strange artifact found - research boost

	// Bombing messages
	MsgYourBombingStart  MessageTypeID = 0x60 // Start of your bombing messages
	MsgYourBombingEnd    MessageTypeID = 0x69 // End of your bombing messages
	MsgEnemyBombingStart MessageTypeID = 0x6A // Start of enemy bombing messages
	MsgEnemyBombingEnd   MessageTypeID = 0x73 // End of enemy bombing messages
	MsgBombingKilledAll1 MessageTypeID = 0x8F // Bombing killed all enemy colonists
	MsgBombingKilledAll2 MessageTypeID = 0x90 // Enemy bombing killed all your colonists

	// Remote mining messages
	MsgMiningNoModules       MessageTypeID = 0x75 // No mining modules
	MsgMiningPlanetInhabited MessageTypeID = 0x76 // Planet inhabited - cancel mining
	MsgMiningDeepSpace       MessageTypeID = 0x77 // Mining in deep space - canceled

	// Terraforming messages
	MsgTerraforming         MessageTypeID = 0x7B  // Terraforming improved planet
	MsgRemoteTerraform      MessageTypeID = 0xBD  // Remote terraforming complete
	MsgPlanetValueImproved1 MessageTypeID = 0x12C // Improved planet value
	MsgPlanetValueImproved2 MessageTypeID = 0x12D // Unable to improve further

	// Comet strike messages
	MsgCometSmallUnowned  MessageTypeID = 0x83 // Small comet (unowned planet)
	MsgCometMediumUnowned MessageTypeID = 0x84 // Medium comet (unowned)
	MsgCometLargeUnowned  MessageTypeID = 0x85 // Large comet (unowned)
	MsgCometHugeUnowned   MessageTypeID = 0x86 // Huge comet (unowned)
	MsgCometSmallOwned    MessageTypeID = 0x87 // Small comet (owned, 25% deaths)
	MsgCometMediumOwned   MessageTypeID = 0x88 // Medium comet (owned, 45% deaths)
	MsgCometLargeOwned    MessageTypeID = 0x89 // Large comet (owned, 65% deaths)
	MsgCometHugeOwned     MessageTypeID = 0x8A // Huge comet (owned, 85% deaths)

	// Alchemy message
	MsgAlchemy MessageTypeID = 0x8C // Scientists transmuted minerals

	// Minefield messages
	MsgMinefieldStart MessageTypeID = 0xBE // Start of minefield messages
	MsgMinefieldEnd   MessageTypeID = 0xCC // End of minefield messages

	// Mass driver/packet messages
	MsgPacketStart MessageTypeID = 0xD1 // Start of packet messages
	MsgPacketEnd   MessageTypeID = 0xDA // End of packet messages

	// Stargate messages
	MsgStargateStart MessageTypeID = 0xDE // Start of stargate messages
	MsgStargateEnd   MessageTypeID = 0xEB // End of stargate messages

	// Planet discovery messages
	MsgHomePlanet          MessageTypeID = 0xA9 // Home planet introduction
	MsgPlanetOccupied      MessageTypeID = 0xAA // Found occupied planet
	MsgPlanetNotHabitable  MessageTypeID = 0xAB // Found uninhabitable planet
	MsgPlanetHabitable     MessageTypeID = 0xAC // Found habitable planet
	MsgPlanetUnknown       MessageTypeID = 0xAD // Found planet (unknown habitability)
	MsgPlanetTerraformable MessageTypeID = 0xAE // Found terraformable planet

	// Victory/death messages
	MsgWinnerDeclared  MessageTypeID = 0xB5 // Winner declared (other player)
	MsgYouWin          MessageTypeID = 0xB6 // You have won
	MsgYouWinShared    MessageTypeID = 0xB7 // You won (shared victory)
	MsgYouAreDead      MessageTypeID = 0xB8 // You are dead
	MsgRaceEliminated1 MessageTypeID = 0xBB // Race eliminated
	MsgRaceEliminated2 MessageTypeID = 0xBC // All rivals eliminated

	// Mystery Trader messages
	MsgMysteryTraderStart    MessageTypeID = 0x108 // Start of Mystery Trader messages
	MsgMysteryTraderEnd      MessageTypeID = 0x10F // End of Mystery Trader messages
	MsgMysteryTraderVanished MessageTypeID = 0x110 // Mystery Trader vanished
	MsgMysteryTraderDetected MessageTypeID = 0x12B // Mystery Trader detected

	// Anti-cheat/punishment messages
	MsgCheatUsurper        MessageTypeID = 0x100 // Population suspects usurper (fCheater)
	MsgCheatWrongEmperor   MessageTypeID = 0x101 // Colonists suspect wrong emperor
	MsgCheatFleetRefused   MessageTypeID = 0x102 // Fleet refused to move
	MsgCheatFleetStrike    MessageTypeID = 0x103 // Fleet captains staged strike
	MsgCheatFleetDefected  MessageTypeID = 0x104 // Fleet defected
	MsgCheatCargoSold      MessageTypeID = 0x105 // Crew sold cargo on black market
	MsgCheatMinesDestroyed MessageTypeID = 0x106 // Freedom fighters destroyed mines
	MsgCheatMineralsStolen MessageTypeID = 0x107 // Freedom fighters stole minerals
	MsgCheatRaceTampered   MessageTypeID = 0x117 // Race definition tampered with (fHacker)

	// Gameplay tips
	MsgTip1 MessageTypeID = 0x7F // Tip: filter messages
	MsgTip2 MessageTypeID = 0x80 // Tip: add waypoints
	MsgTip3 MessageTypeID = 0x81 // Tip: design ships
	MsgTip4 MessageTypeID = 0x82 // Tip: popup help
)

// MessageCategory returns the category name for a message type ID.
func MessageCategory(id MessageTypeID) string {
	switch {
	// Battle messages
	case id >= MsgBattleAftermath && id <= MsgBattleObserved:
		return "Battle aftermath"
	case id >= MsgBattleOutcomeStart && id <= MsgBattleOutcomeEnd:
		return "Battle outcome"
	case id >= MsgBattleExtendedStart && id <= MsgBattleExtendedEnd:
		return "Battle"
	case id == MsgBattleVCR:
		return "Battle VCR"
	case id == MsgBattleNotInvolved1 || id == MsgBattleNotInvolved2:
		return "Battle observed"

	// Colony/Population
	case id >= MsgColonistsDied && id <= MsgPopulationOvercrowd:
		return "Colony death/population"
	case id == MsgColonistsJumpedShip || id == MsgColonistsAbandoned:
		return "Colony abandoned"

	// Fleet/Fuel
	case id >= MsgFleetOutOfFuel && id <= MsgFleetOutranScanners:
		return "Fleet waypoint/fuel"
	case id == MsgFleetFuelSpeedReduced:
		return "Fleet out of fuel"
	case id == MsgRamScoopFuel:
		return "Ram scoop fuel"

	// Cargo transfer
	case id >= MsgCargoLoaded && id <= MsgCargoBeamedTo:
		return "Cargo load/unload"
	case id >= MsgFleetTransferStart && id <= MsgFleetTransferEnd:
		return "Fleet transfer"
	case id == MsgCargoFromFleet1 || id == MsgCargoFromFleet2:
		return "Cargo from fleet"

	// Production
	case id >= MsgStarbaseBuiltShip && id <= MsgStarbaseBuiltNoFuel2:
		return "Starbase built ship"
	case id == MsgBuiltFactory || id == MsgBuiltFactories:
		return "Built factories"
	case id == MsgBuiltMine || id == MsgBuiltMines:
		return "Built mines"
	case id == MsgBuiltDefense || id == MsgBuiltDefenses:
		return "Built defenses"
	case id == MsgQueueCompleted || id == MsgQueueEmpty:
		return "Queue empty"
	case id == MsgBuiltScanner:
		return "Built scanner"
	case id >= MsgBuiltStarbase1 && id <= MsgBuiltStarbase3:
		return "Built starbase"

	// Research
	case id == MsgResearchComplete || id == MsgResearchComplete2:
		return "Research complete"
	case id == MsgTechBenefit:
		return "Tech benefit"
	case id == MsgNewHullUnlocked:
		return "New hull unlocked"

	// Colonization
	case id >= MsgColonizeNotInOrbit && id <= MsgColonizeErrorsEnd:
		return "Colonization error"

	// Scrap
	case id >= MsgFleetScrappedStart && id <= MsgFleetScrappedEnd:
		return "Fleet scrapped"
	case id >= MsgFleetScrapped2Start && id <= MsgFleetScrapped2End:
		return "Fleet scrapped"

	// Artifact
	case id == MsgArtifactFound:
		return "Artifact found"

	// Bombing
	case id >= MsgYourBombingStart && id <= MsgYourBombingEnd:
		return "Your bombing"
	case id >= MsgEnemyBombingStart && id <= MsgEnemyBombingEnd:
		return "Enemy bombing"
	case id == MsgBombingKilledAll1 || id == MsgBombingKilledAll2:
		return "Bombing annihilation"

	// Remote mining
	case id >= MsgMiningNoModules && id <= MsgMiningDeepSpace:
		return "Remote mining error"

	// Terraforming
	case id == MsgTerraforming:
		return "Terraforming"
	case id == MsgRemoteTerraform:
		return "Remote terraform"
	case id == MsgPlanetValueImproved1 || id == MsgPlanetValueImproved2:
		return "Planet value improved"

	// Comet
	case id >= MsgCometSmallUnowned && id <= MsgCometHugeOwned:
		return "Comet strike"

	// Alchemy
	case id == MsgAlchemy:
		return "Mineral alchemy"

	// Minefields
	case id >= MsgMinefieldStart && id <= MsgMinefieldEnd:
		return "Minefield"

	// Mass driver/packets
	case id >= MsgPacketStart && id <= MsgPacketEnd:
		return "Mass driver/packet"

	// Stargate
	case id >= MsgStargateStart && id <= MsgStargateEnd:
		return "Stargate"

	// Planet discovery
	case id == MsgHomePlanet:
		return "Home planet"
	case id >= MsgPlanetOccupied && id <= MsgPlanetTerraformable:
		return "Planet discovery"

	// Victory/death
	case id >= MsgWinnerDeclared && id <= MsgYouAreDead:
		return "Victory/death"
	case id == MsgRaceEliminated1 || id == MsgRaceEliminated2:
		return "Race eliminated"

	// Mystery Trader
	case id >= MsgMysteryTraderStart && id <= MsgMysteryTraderEnd:
		return "Mystery Trader"
	case id == MsgMysteryTraderVanished:
		return "MT vanished"
	case id == MsgMysteryTraderDetected:
		return "MT detected"

	// Anti-cheat punishment
	case id >= MsgCheatUsurper && id <= MsgCheatMineralsStolen:
		return "Anti-cheat punishment"
	case id == MsgCheatRaceTampered:
		return "Race hack detection"

	// Tips
	case id >= MsgTip1 && id <= MsgTip4:
		return "Gameplay tip"

	default:
		return ""
	}
}
