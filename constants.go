package houston

const (
	FileFooterBlockType                 BlockTypeID = iota // 0
	ManualSmallLoadUnloadTaskBlockType                     // 1
	ManualMediumLoadUnloadTaskBlockType                    // 2
	WaypointDeleteBlockType                                // 3
	WaypointAddBlockType                                   // 4
	WaypointChangeTaskBlockType                            // 5
	PlayerBlockType                                        // 6
	PlanetsBlockType                                       // 7
	FileHeaderBlockType                                    // 8
	FileHashBlockType                                      // 9
	WaypointRepeatOrdersBlockType                          // 10
	UnknownBlock11BlockType                                // 11
	EventsBlockType                                        // 12
	PlanetBlockType                                        // 13
	PartialPlanetBlockType                                 // 14
	UnknownBlock15BlockType                                // 15
	FleetBlockType                                         // 16
	PartialFleetBlockType                                  // 17
	UnknownBlock18BlockType                                // 18
	WaypointTaskBlockType                                  // 19
	WaypointBlockType                                      // 20
	FleetNameBlockType                                     // 21
	UnknownBlock22BlockType                                // 22
	MoveShipsBlockType                                     // 23
	FleetSplitBlockType                                    // 24
	ManualLargeLoadUnloadTaskBlockType                     // 25
	DesignBlockType                                        // 26
	DesignChangeBlockType                                  // 27
	ProductionQueueBlockType                               // 28
	ProductionQueueChangeBlockType                         // 29
	BattlePlanBlockType                                    // 30
	BattleBlockType                                        // 31
	CountersBlockType                                      // 32
	MessagesFilterBlockType                                // 33
	ResearchChangeBlockType                                // 34
	PlanetChangeBlockType                                  // 35
	ChangePasswordBlockType                                // 36
	FleetsMergeBlockType                                   // 37
	PlayersRelationChangeBlockType                         // 38
	BattleContinuationBlockType                            // 39
	MessageBlockType                                       // 40
	AiHFileRecordBlockType                                 // 41
	SetFleetBattlePlanBlockType                            // 42
	ObjectBlockType                                        // 43
	RenameFleetBlocType                                    // 44
	PlayerScoresBlockType                                  // 45
	SaveAndSubmitBlockType                                 // 46
)
