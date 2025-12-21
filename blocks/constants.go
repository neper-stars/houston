// Package blocks contains all Stars! file block type definitions and parsers.
//
// Stars! game files are composed of a sequence of blocks, each with a 16-bit header
// containing a 6-bit type ID and 10-bit size, followed by the block data.
// Most blocks (except FileHeader and FileFooter) are encrypted using a
// custom XOR-based encryption scheme.
//
// Block types are organized into several categories:
//   - File structure blocks (0, 8, 9): File header, footer, and hash
//   - Player/game data (6, 7, 32, 45): Player info, planets list, counters, scores
//   - Planet blocks (13, 14, 35): Full and partial planet data, planet changes
//   - Fleet blocks (16, 17, 21, 23, 24, 37, 44): Fleet data and operations
//   - Design blocks (26, 27): Ship and starbase designs
//   - Waypoint blocks (3, 4, 5, 10, 19, 20): Fleet movement orders
//   - Production blocks (28, 29): Planet production queues
//   - Battle blocks (30, 31, 39, 42): Battle plans and records
//   - Message blocks (33, 40): Player messages and filters
//   - Other blocks: Events, objects, research, passwords, etc.
package blocks

const (
	// FileFooterBlockType (0) marks the end of a Stars! file.
	// This block has zero size and no data. It is NOT encrypted.
	// Every valid Stars! file ends with this block.
	FileFooterBlockType BlockTypeID = iota // 0

	// ManualSmallLoadUnloadTaskBlockType (1) represents a small cargo transfer task.
	// Used when a fleet is ordered to load or unload a small amount of cargo
	// (minerals, colonists) at a planet or another fleet.
	ManualSmallLoadUnloadTaskBlockType // 1

	// ManualMediumLoadUnloadTaskBlockType (2) represents a medium cargo transfer task.
	// Similar to type 1 but for medium-sized cargo transfers.
	ManualMediumLoadUnloadTaskBlockType // 2

	// WaypointDeleteBlockType (3) removes a waypoint from a fleet's orders.
	// Contains the fleet number and waypoint index to delete.
	WaypointDeleteBlockType // 3

	// WaypointAddBlockType (4) adds a new waypoint to a fleet's orders.
	// Contains destination coordinates and task information.
	WaypointAddBlockType // 4

	// WaypointChangeTaskBlockType (5) modifies the task at an existing waypoint.
	// Used to change what a fleet does when it arrives at a waypoint.
	WaypointChangeTaskBlockType // 5

	// PlayerBlockType (6) contains player information including race data.
	// Found in .m and .h files. Contains the player's race name, password hash,
	// and various player-specific settings.
	PlayerBlockType // 6

	// PlanetsBlockType (7) contains the universe's planet list.
	// This block has a special structure: the main block data is followed by
	// 4 bytes per planet containing name ID and coordinates.
	// Planet X coordinates are stored as offsets from the previous planet.
	PlanetsBlockType // 7

	// FileHeaderBlockType (8) is the first block in every Stars! file.
	// Contains file metadata: magic number, game ID, turn number, player index,
	// encryption salt, and various flags. This block is NOT encrypted.
	FileHeaderBlockType // 8

	// FileHashBlockType (9) contains player identification data.
	// Includes serial number and hardware fingerprint for detecting multi-accounting.
	FileHashBlockType // 9

	// WaypointRepeatOrdersBlockType (10) sets a fleet to repeat its waypoint orders.
	// When enabled, the fleet will loop through its waypoints continuously.
	WaypointRepeatOrdersBlockType // 10

	// UnknownBlock11BlockType (11) has an unknown purpose.
	// Preserved for completeness but not yet decoded.
	UnknownBlock11BlockType // 11

	// EventsBlockType (12) contains game event notifications.
	// Events include battles, discoveries, production completions, etc.
	EventsBlockType // 12

	// PlanetBlockType (13) contains full planet data for an owned planet.
	// Includes environment, minerals, population, installations, starbase info,
	// and other detailed planet information.
	PlanetBlockType // 13

	// PartialPlanetBlockType (14) contains partial planet data.
	// Used for planets the player has scanned but doesn't own.
	// Contains only the information visible to the scanning player.
	PartialPlanetBlockType // 14

	// UnknownBlock15BlockType (15) has an unknown purpose.
	UnknownBlock15BlockType // 15

	// FleetBlockType (16) contains full fleet data for an owned fleet.
	// Includes ship composition, cargo, fuel, damage, battle plan, and waypoints.
	FleetBlockType // 16

	// PartialFleetBlockType (17) contains partial fleet data.
	// Used for fleets the player can see but doesn't own.
	// The amount of information varies based on scanning capability.
	PartialFleetBlockType // 17

	// UnknownBlock18BlockType (18) has an unknown purpose.
	UnknownBlock18BlockType // 18

	// WaypointTaskBlockType (19) defines the task at a waypoint.
	// Tasks include colonize, remote mining, transport, patrol, etc.
	WaypointTaskBlockType // 19

	// WaypointBlockType (20) contains waypoint position and basic orders.
	// Each fleet has a list of waypoints defining its movement path.
	WaypointBlockType // 20

	// FleetNameBlockType (21) contains a custom fleet name.
	// Players can rename fleets from the default "Fleet #N" names.
	FleetNameBlockType // 21

	// UnknownBlock22BlockType (22) has an unknown purpose.
	UnknownBlock22BlockType // 22

	// MoveShipsBlockType (23) transfers ships between fleets.
	// Used when splitting ships from one fleet to another.
	MoveShipsBlockType // 23

	// FleetSplitBlockType (24) splits a fleet into two fleets.
	// Contains the fleet number being split.
	FleetSplitBlockType // 24

	// ManualLargeLoadUnloadTaskBlockType (25) represents a large cargo transfer task.
	// Similar to types 1 and 2 but for large cargo transfers.
	ManualLargeLoadUnloadTaskBlockType // 25

	// DesignBlockType (26) contains a ship or starbase design.
	// Includes hull type, components, armor, mass, and build statistics.
	// Players can have up to 16 ship designs and 10 starbase designs.
	DesignBlockType // 26

	// DesignChangeBlockType (27) modifies or deletes an existing design.
	// Can either update a design's components or mark it as deleted.
	DesignChangeBlockType // 27

	// ProductionQueueBlockType (28) contains a planet's production queue.
	// Lists items being built: ships, starbases, defenses, factories, etc.
	ProductionQueueBlockType // 28

	// ProductionQueueChangeBlockType (29) modifies a planet's production queue.
	// Contains the planet ID and the new queue contents.
	ProductionQueueChangeBlockType // 29

	// BattlePlanBlockType (30) defines fleet battle tactics.
	// Players can create up to 16 battle plans specifying targeting,
	// engagement rules, and tactical preferences.
	BattlePlanBlockType // 30

	// BattleBlockType (31) contains battle record data.
	// Records the outcome of space combat for replay in the battle viewer.
	BattleBlockType // 31

	// CountersBlockType (32) contains game object counts.
	// Tracks the total number of planets and fleets in the game.
	CountersBlockType // 32

	// MessagesFilterBlockType (33) contains message filter settings.
	// Players can filter which types of messages they receive.
	MessagesFilterBlockType // 33

	// ResearchChangeBlockType (34) changes research priorities.
	// Sets which technology field to research and resource allocation.
	ResearchChangeBlockType // 34

	// PlanetChangeBlockType (35) modifies planet settings.
	// Changes production, population routing, or other planet options.
	PlanetChangeBlockType // 35

	// ChangePasswordBlockType (36) changes the player's race password.
	// Contains the new password hash.
	ChangePasswordBlockType // 36

	// FleetsMergeBlockType (37) merges multiple fleets into one.
	// Contains the target fleet and list of fleets to merge into it.
	FleetsMergeBlockType // 37

	// PlayersRelationChangeBlockType (38) changes diplomatic relations.
	// Sets whether another player is friend, enemy, or neutral.
	PlayersRelationChangeBlockType // 38

	// BattleContinuationBlockType (39) continues battle record data.
	// Large battles may span multiple blocks.
	BattleContinuationBlockType // 39

	// MessageBlockType (40) contains a player-to-player message.
	// Includes sender, recipient(s), and the message text.
	// Messages use a special Stars! string encoding.
	MessageBlockType // 40

	// AiHFileRecordBlockType (41) contains AI host file records.
	// Used in computer player game management.
	AiHFileRecordBlockType // 41

	// SetFleetBattlePlanBlockType (42) assigns a battle plan to a fleet.
	// Links a fleet to one of the player's defined battle plans.
	SetFleetBattlePlanBlockType // 42

	// ObjectBlockType (43) contains game objects like minefields and wormholes.
	// A multipurpose block with different subtypes:
	//   - Minefields: position, mine count, type (standard/heavy/speed bump)
	//   - Wormholes: endpoints, stability, which players have seen/used it
	//   - Mystery Traders: position, destination, items for sale
	//   - Mineral Packets: position, trajectory, contents
	ObjectBlockType // 43

	// RenameFleetBlockType (44) renames a fleet.
	// Changes a fleet's display name from the default.
	RenameFleetBlockType // 44

	// PlayerScoresBlockType (45) contains player score data.
	// Tracks various scoring metrics for victory condition evaluation.
	PlayerScoresBlockType // 45

	// SaveAndSubmitBlockType (46) marks a turn as submitted.
	// Indicates the player has finished their turn and submitted orders.
	SaveAndSubmitBlockType // 46
)

// BlockTypeName returns a human-readable name for a block type ID.
func BlockTypeName(id BlockTypeID) string {
	names := map[BlockTypeID]string{
		FileFooterBlockType:                 "FileFooter",
		ManualSmallLoadUnloadTaskBlockType:  "ManualSmallLoadUnloadTask",
		ManualMediumLoadUnloadTaskBlockType: "ManualMediumLoadUnloadTask",
		WaypointDeleteBlockType:             "WaypointDelete",
		WaypointAddBlockType:                "WaypointAdd",
		WaypointChangeTaskBlockType:         "WaypointChangeTask",
		PlayerBlockType:                     "Player",
		PlanetsBlockType:                    "Planets",
		FileHeaderBlockType:                 "FileHeader",
		FileHashBlockType:                   "FileHash",
		WaypointRepeatOrdersBlockType:       "WaypointRepeatOrders",
		UnknownBlock11BlockType:             "Unknown11",
		EventsBlockType:                     "Events",
		PlanetBlockType:                     "Planet",
		PartialPlanetBlockType:              "PartialPlanet",
		UnknownBlock15BlockType:             "Unknown15",
		FleetBlockType:                      "Fleet",
		PartialFleetBlockType:               "PartialFleet",
		UnknownBlock18BlockType:             "Unknown18",
		WaypointTaskBlockType:               "WaypointTask",
		WaypointBlockType:                   "Waypoint",
		FleetNameBlockType:                  "FleetName",
		UnknownBlock22BlockType:             "Unknown22",
		MoveShipsBlockType:                  "MoveShips",
		FleetSplitBlockType:                 "FleetSplit",
		ManualLargeLoadUnloadTaskBlockType:  "ManualLargeLoadUnloadTask",
		DesignBlockType:                     "Design",
		DesignChangeBlockType:               "DesignChange",
		ProductionQueueBlockType:            "ProductionQueue",
		ProductionQueueChangeBlockType:      "ProductionQueueChange",
		BattlePlanBlockType:                 "BattlePlan",
		BattleBlockType:                     "Battle",
		CountersBlockType:                   "Counters",
		MessagesFilterBlockType:             "MessagesFilter",
		ResearchChangeBlockType:             "ResearchChange",
		PlanetChangeBlockType:               "PlanetChange",
		ChangePasswordBlockType:             "ChangePassword",
		FleetsMergeBlockType:                "FleetsMerge",
		PlayersRelationChangeBlockType:      "PlayersRelationChange",
		BattleContinuationBlockType:         "BattleContinuation",
		MessageBlockType:                    "Message",
		AiHFileRecordBlockType:              "AiHFileRecord",
		SetFleetBattlePlanBlockType:         "SetFleetBattlePlan",
		ObjectBlockType:                     "Object",
		RenameFleetBlockType:                "RenameFleet",
		PlayerScoresBlockType:               "PlayerScores",
		SaveAndSubmitBlockType:              "SaveAndSubmit",
	}
	if name, ok := names[id]; ok {
		return name
	}
	return "Unknown"
}
