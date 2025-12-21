// Package mfilemerger provides functionality to merge data between allied players' M files.
//
// This package allows allies to share scanning information by merging planet, fleet,
// design, and object data from multiple M files. Each player's file is augmented with
// information gathered from all allied files.
//
// The library operates entirely in memory - callers are responsible for reading files
// from and writing files to their storage (disk, database, etc.).
//
// Example usage:
//
//	merger := mfilemerger.New()
//	if err := merger.Add("player1", player1Data); err != nil {
//	    log.Fatal(err)
//	}
//	if err := merger.Add("player2", player2Data); err != nil {
//	    log.Fatal(err)
//	}
//	if err := merger.Merge(); err != nil {
//	    log.Fatal(err)
//	}
//	// Get merged data for each player
//	mergedData1 := merger.GetMergedData("player1")
//	mergedData2 := merger.GetMergedData("player2")
package mfilemerger

import (
	"fmt"
	"io"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

// FileEntry represents a single M file's data.
type FileEntry struct {
	Name         string
	OriginalData []byte
	Blocks       []blocks.Block
	PlayerIndex  int
}

// Merger handles merging multiple M files.
type Merger struct {
	entries    map[string]*FileEntry
	names      []string // preserve order
	gameID     uint32
	turn       uint16
	playerMask int

	// Collected data
	players   [16]*blocks.PlayerBlock
	planets   map[int]*PlanetInfo
	fleets    [16]map[int]*FleetInfo
	designs   [16][16]*DesignInfo
	starbases [16][10]*DesignInfo
	objects   map[int]blocks.ObjectBlock

	// State
	merged bool
}

// PlanetInfo tracks the best available data for a planet.
type PlanetInfo struct {
	Best            *blocks.PartialPlanetBlock
	BestEnvironment *blocks.PartialPlanetBlock
	BestStarbase    *blocks.PartialPlanetBlock
	HasConflict     bool
}

// FleetInfo tracks the best available data for a fleet.
type FleetInfo struct {
	Best     *blocks.PartialFleetBlock
	Original *blocks.PartialFleetBlock
	HasMass  bool
	Kind     byte
}

// DesignInfo tracks the best available data for a design.
type DesignInfo struct {
	Player      int
	Block       *blocks.DesignBlock
	HasConflict bool
}

// MergeResult contains the results of a merge operation.
type MergeResult struct {
	EntriesProcessed int
	PlanetsMerged    int
	FleetsMerged     int
	DesignsMerged    int
	ObjectsMerged    int
	Warnings         []string
}

// New creates a new Merger.
func New() *Merger {
	m := &Merger{
		entries: make(map[string]*FileEntry),
		planets: make(map[int]*PlanetInfo),
		objects: make(map[int]blocks.ObjectBlock),
	}

	for i := 0; i < 16; i++ {
		m.fleets[i] = make(map[int]*FleetInfo)
	}

	return m
}

// Add adds M file data to be merged.
// The name parameter is a unique identifier for this entry (e.g., filename or player ID).
func (m *Merger) Add(name string, data []byte) error {
	if m.merged {
		return fmt.Errorf("cannot add after merge")
	}

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	if err != nil {
		return fmt.Errorf("failed to parse blocks: %w", err)
	}

	if len(blockList) == 0 {
		return fmt.Errorf("%s does not parse into block list", name)
	}

	header, ok := blockList[0].(blocks.FileHeader)
	if !ok {
		return fmt.Errorf("%s does not start with header block", name)
	}

	entry := &FileEntry{
		Name:         name,
		OriginalData: data,
		Blocks:       blockList,
		PlayerIndex:  header.PlayerIndex(),
	}

	// Validate game ID and turn
	if len(m.entries) == 0 {
		m.gameID = header.GameID
		m.turn = header.Turn
		m.playerMask = 1 << header.PlayerIndex()
	} else {
		if header.GameID != m.gameID {
			return fmt.Errorf("game ID mismatch in %s (expected %d, got %d)", name, m.gameID, header.GameID)
		}
		if header.Turn != m.turn {
			return fmt.Errorf("turn mismatch in %s (expected %d, got %d)", name, m.turn, header.Turn)
		}
		m.playerMask |= 1 << header.PlayerIndex()
	}

	m.entries[name] = entry
	m.names = append(m.names, name)

	return nil
}

// AddReader adds M file data from an io.Reader.
func (m *Merger) AddReader(name string, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}
	return m.Add(name, data)
}

// EntryCount returns the number of entries added.
func (m *Merger) EntryCount() int {
	return len(m.entries)
}

// Names returns the names of all added entries in order.
func (m *Merger) Names() []string {
	return m.names
}

// Merge performs the merge operation on all added entries.
func (m *Merger) Merge() (*MergeResult, error) {
	if len(m.entries) == 0 {
		return nil, fmt.Errorf("no entries to merge")
	}

	if m.merged {
		return nil, fmt.Errorf("already merged")
	}

	// Process each entry and collect data
	for _, name := range m.names {
		entry := m.entries[name]
		if err := m.processEntry(entry); err != nil {
			return nil, err
		}
	}

	// Post-process: finalize merged data
	m.postProcess()

	m.merged = true

	result := &MergeResult{
		EntriesProcessed: len(m.entries),
		PlanetsMerged:    len(m.planets),
		ObjectsMerged:    len(m.objects),
	}

	// Count fleets and designs
	for i := 0; i < 16; i++ {
		result.FleetsMerged += len(m.fleets[i])
		for j := 0; j < 16; j++ {
			if m.designs[i][j] != nil {
				result.DesignsMerged++
			}
		}
		for j := 0; j < 10; j++ {
			if m.starbases[i][j] != nil {
				result.DesignsMerged++
			}
		}
	}

	return result, nil
}

// GetMergedData returns the merged data for a specific entry.
// Note: Currently returns original data as block encoding is not yet implemented.
// Full implementation would rebuild the file with merged blocks.
func (m *Merger) GetMergedData(name string) ([]byte, error) {
	if !m.merged {
		return nil, fmt.Errorf("must call Merge() first")
	}

	entry, ok := m.entries[name]
	if !ok {
		return nil, fmt.Errorf("entry %s not found", name)
	}

	// TODO: Implement block encoding to rebuild file with merged data
	// For now, return original data
	return entry.OriginalData, nil
}

// WriteMergedData writes the merged data for a specific entry to an io.Writer.
func (m *Merger) WriteMergedData(name string, w io.Writer) error {
	data, err := m.GetMergedData(name)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// GetPlanets returns the merged planet data.
func (m *Merger) GetPlanets() map[int]*PlanetInfo {
	return m.planets
}

// GetFleets returns the merged fleet data for a player.
func (m *Merger) GetFleets(player int) map[int]*FleetInfo {
	if player < 0 || player >= 16 {
		return nil
	}
	return m.fleets[player]
}

// GetDesigns returns the merged ship design data.
func (m *Merger) GetDesigns() [16][16]*DesignInfo {
	return m.designs
}

// GetStarbases returns the merged starbase design data.
func (m *Merger) GetStarbases() [16][10]*DesignInfo {
	return m.starbases
}

// GetObjects returns the merged object data (minefields, wormholes, etc.).
func (m *Merger) GetObjects() map[int]blocks.ObjectBlock {
	return m.objects
}

// GetGameID returns the game ID from the merged entries.
func (m *Merger) GetGameID() uint32 {
	return m.gameID
}

// GetTurn returns the turn number from the merged entries.
func (m *Merger) GetTurn() uint16 {
	return m.turn
}

func (m *Merger) processEntry(entry *FileEntry) error {
	shipDesignOwners := make([]int, 0)
	starbaseDesignOwners := make([]int, 0)
	shipDesignIndex := 0
	starbaseDesignIndex := 0

	for _, block := range entry.Blocks {
		switch b := block.(type) {
		case blocks.PlayerBlock:
			m.processPlayer(&b, &shipDesignOwners, &starbaseDesignOwners)

		case blocks.PartialPlanetBlock:
			m.processPlanet(&b)

		case blocks.PlanetBlock:
			ppb := b.PartialPlanetBlock
			m.processPlanet(&ppb)

		case blocks.DesignBlock:
			var owner int
			if b.IsStarbase {
				if starbaseDesignIndex < len(starbaseDesignOwners) {
					owner = starbaseDesignOwners[starbaseDesignIndex]
					starbaseDesignIndex++
				}
				m.processStarbaseDesign(owner, &b)
			} else {
				if shipDesignIndex < len(shipDesignOwners) {
					owner = shipDesignOwners[shipDesignIndex]
					shipDesignIndex++
				}
				m.processShipDesign(owner, &b)
			}

		case blocks.PartialFleetBlock:
			m.processFleet(&b)

		case blocks.FleetBlock:
			pfb := b.PartialFleetBlock
			m.processFleet(&pfb)

		case blocks.ObjectBlock:
			m.processObject(b)
		}
	}

	return nil
}

func (m *Merger) processPlayer(block *blocks.PlayerBlock, shipOwners, starbaseOwners *[]int) {
	playerNum := block.PlayerNumber
	if m.players[playerNum] == nil {
		m.players[playerNum] = block
	}

	for i := 0; i < block.ShipDesignCount; i++ {
		*shipOwners = append(*shipOwners, playerNum)
	}
	for i := 0; i < block.StarbaseDesignCount; i++ {
		*starbaseOwners = append(*starbaseOwners, playerNum)
	}
}

func (m *Merger) processPlanet(block *blocks.PartialPlanetBlock) {
	planetNum := block.PlanetNumber
	info := m.planets[planetNum]
	if info == nil {
		info = &PlanetInfo{}
		m.planets[planetNum] = info
	}

	if info.Best == nil {
		info.Best = block
	} else {
		if block.CanSeeEnvironment() && !info.Best.CanSeeEnvironment() {
			info.BestEnvironment = block
		}
		if block.HasStarbase && info.BestStarbase == nil {
			info.BestStarbase = block
		}
	}
}

func (m *Merger) processShipDesign(owner int, block *blocks.DesignBlock) {
	if owner < 0 || owner >= 16 || block.DesignNumber < 0 || block.DesignNumber >= 16 {
		return
	}

	info := m.designs[owner][block.DesignNumber]
	if info == nil {
		info = &DesignInfo{Player: owner}
		m.designs[owner][block.DesignNumber] = info
	}

	if info.Block == nil {
		info.Block = block
	} else if !info.Block.IsFullDesign && block.IsFullDesign {
		info.Block = block
	}
}

func (m *Merger) processStarbaseDesign(owner int, block *blocks.DesignBlock) {
	if owner < 0 || owner >= 16 || block.DesignNumber < 0 || block.DesignNumber >= 10 {
		return
	}

	info := m.starbases[owner][block.DesignNumber]
	if info == nil {
		info = &DesignInfo{Player: owner}
		m.starbases[owner][block.DesignNumber] = info
	}

	if info.Block == nil {
		info.Block = block
	} else if !info.Block.IsFullDesign && block.IsFullDesign {
		info.Block = block
	}
}

func (m *Merger) processFleet(block *blocks.PartialFleetBlock) {
	if block.Owner < 0 || block.Owner >= 16 {
		return
	}

	info := m.fleets[block.Owner][block.FleetNumber]
	if info == nil {
		info = &FleetInfo{}
		m.fleets[block.Owner][block.FleetNumber] = info
	}

	if info.Best == nil {
		info.Best = block
		info.Kind = block.KindByte
	} else if block.KindByte > info.Kind {
		info.Best = block
		info.Kind = block.KindByte
	}
}

func (m *Merger) processObject(block blocks.ObjectBlock) {
	if block.IsCountObject {
		return
	}

	objID := block.Number
	m.objects[objID] = block
}

func (m *Merger) postProcess() {
	for i := 0; i < 16; i++ {
		if m.players[i] == nil {
			continue
		}

		m.players[i].Fleets = len(m.fleets[i])

		shipCount := 0
		for j := 0; j < 16; j++ {
			if m.designs[i][j] != nil {
				shipCount++
			}
		}
		m.players[i].ShipDesignCount = shipCount

		starbaseCount := 0
		for j := 0; j < 10; j++ {
			if m.starbases[i][j] != nil {
				starbaseCount++
			}
		}
		m.players[i].StarbaseDesignCount = starbaseCount
	}
}
