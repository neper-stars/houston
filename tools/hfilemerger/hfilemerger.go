// Package hfilemerger provides functionality to merge data from multiple H (history) files.
//
// H files contain historical game data that can be merged to combine information
// from different players or different turns. This is useful for maintaining
// comprehensive game histories.
//
// The library operates entirely in memory - callers are responsible for reading files
// from and writing files to their storage (disk, database, etc.).
//
// Example usage:
//
//	merger := hfilemerger.New()
//	if err := merger.AddH("player1", player1HData); err != nil {
//	    log.Fatal(err)
//	}
//	if err := merger.AddM("player1m", player1MData); err != nil {
//	    log.Fatal(err)
//	}
//	if err := merger.Merge(); err != nil {
//	    log.Fatal(err)
//	}
//	// Get merged data for each H entry
//	mergedData := merger.GetMergedData("player1")
package hfilemerger

import (
	"fmt"
	"io"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

// FileEntry represents a single file's data.
type FileEntry struct {
	Name         string
	OriginalData []byte
	Blocks       []blocks.Block
	IsHFile      bool
}

// Merger handles merging multiple H files.
type Merger struct {
	entries   map[string]*FileEntry
	hNames    []string // preserve order of H files
	mNames    []string // preserve order of M files
	gameID    uint32
	planets   map[int]*PlanetInfo
	players   [16]*blocks.PlayerBlock
	designs   [16][16]*DesignInfo
	starbases [16][10]*DesignInfo

	// State
	merged bool
}

// PlanetInfo tracks planet data across turns.
type PlanetInfo struct {
	Latest            *blocks.PartialPlanetBlock
	LatestEnvironment *blocks.PartialPlanetBlock
	LatestStarbase    *blocks.PartialPlanetBlock
	LatestTurn        int
}

// DesignInfo tracks design data.
type DesignInfo struct {
	Player int
	Block  *blocks.DesignBlock
	Turn   int
}

// MergeResult contains the results of a merge operation.
type MergeResult struct {
	HEntriesProcessed int
	MEntriesProcessed int
	PlanetsMerged     int
	DesignsMerged     int
	Warnings          []string
}

// New creates a new Merger.
func New() *Merger {
	return &Merger{
		entries: make(map[string]*FileEntry),
		planets: make(map[int]*PlanetInfo),
	}
}

// AddH adds H file data to be merged.
// The name parameter is a unique identifier for this entry.
func (m *Merger) AddH(name string, data []byte) error {
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

	// Validate game ID
	if len(m.entries) == 0 {
		m.gameID = header.GameID
	} else if header.GameID != m.gameID {
		return fmt.Errorf("game ID mismatch in %s (expected %d, got %d)", name, m.gameID, header.GameID)
	}

	entry := &FileEntry{
		Name:         name,
		OriginalData: data,
		Blocks:       blockList,
		IsHFile:      true,
	}

	m.entries[name] = entry
	m.hNames = append(m.hNames, name)

	return nil
}

// AddHReader adds H file data from an io.Reader.
func (m *Merger) AddHReader(name string, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}
	return m.AddH(name, data)
}

// AddM adds M file data to incorporate design data from.
// M files contribute data but are not modified by the merge.
func (m *Merger) AddM(name string, data []byte) error {
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

	// Validate game ID
	if len(m.entries) == 0 {
		m.gameID = header.GameID
	} else if header.GameID != m.gameID {
		return fmt.Errorf("game ID mismatch in %s (expected %d, got %d)", name, m.gameID, header.GameID)
	}

	entry := &FileEntry{
		Name:         name,
		OriginalData: data,
		Blocks:       blockList,
		IsHFile:      false,
	}

	m.entries[name] = entry
	m.mNames = append(m.mNames, name)

	return nil
}

// AddMReader adds M file data from an io.Reader.
func (m *Merger) AddMReader(name string, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}
	return m.AddM(name, data)
}

// HEntryCount returns the number of H entries added.
func (m *Merger) HEntryCount() int {
	return len(m.hNames)
}

// MEntryCount returns the number of M entries added.
func (m *Merger) MEntryCount() int {
	return len(m.mNames)
}

// HNames returns the names of all H entries in order.
func (m *Merger) HNames() []string {
	return m.hNames
}

// MNames returns the names of all M entries in order.
func (m *Merger) MNames() []string {
	return m.mNames
}

// Merge performs the merge operation on all added entries.
func (m *Merger) Merge() (*MergeResult, error) {
	if len(m.hNames) == 0 {
		return nil, fmt.Errorf("no H entries given")
	}

	if m.merged {
		return nil, fmt.Errorf("already merged")
	}

	// Process all entries
	for _, name := range m.hNames {
		entry := m.entries[name]
		m.processEntry(entry)
	}
	for _, name := range m.mNames {
		entry := m.entries[name]
		m.processEntry(entry)
	}

	m.merged = true

	result := &MergeResult{
		HEntriesProcessed: len(m.hNames),
		MEntriesProcessed: len(m.mNames),
		PlanetsMerged:     len(m.planets),
	}

	// Count designs
	for i := 0; i < 16; i++ {
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

// GetMergedData returns the merged data for a specific H entry.
// Note: Currently returns original data as block encoding is not yet implemented.
func (m *Merger) GetMergedData(name string) ([]byte, error) {
	if !m.merged {
		return nil, fmt.Errorf("must call Merge() first")
	}

	entry, ok := m.entries[name]
	if !ok {
		return nil, fmt.Errorf("entry %s not found", name)
	}

	if !entry.IsHFile {
		return nil, fmt.Errorf("entry %s is not an H file", name)
	}

	// TODO: Implement block encoding to rebuild file with merged data
	return entry.OriginalData, nil
}

// WriteMergedData writes the merged data for a specific H entry to an io.Writer.
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

// GetDesigns returns the merged ship design data.
func (m *Merger) GetDesigns() [16][16]*DesignInfo {
	return m.designs
}

// GetStarbases returns the merged starbase design data.
func (m *Merger) GetStarbases() [16][10]*DesignInfo {
	return m.starbases
}

// GetGameID returns the game ID from the merged entries.
func (m *Merger) GetGameID() uint32 {
	return m.gameID
}

func (m *Merger) processEntry(entry *FileEntry) {
	var fileTurn int

	for _, block := range entry.Blocks {
		switch b := block.(type) {
		case blocks.FileHeader:
			fileTurn = int(b.Turn)

		case blocks.PartialPlanetBlock:
			m.processPlanet(&b, fileTurn)

		case blocks.PlanetBlock:
			ppb := b.PartialPlanetBlock
			m.processPlanet(&ppb, fileTurn)

		case blocks.PlayerBlock:
			m.processPlayer(&b)
		}
	}
}

func (m *Merger) processPlanet(block *blocks.PartialPlanetBlock, fileTurn int) {
	planetNum := block.PlanetNumber
	info := m.planets[planetNum]
	if info == nil {
		info = &PlanetInfo{LatestTurn: -1}
		m.planets[planetNum] = info
	}

	turn := fileTurn
	if block.Turn > 0 {
		turn = block.Turn
	}

	if turn > info.LatestTurn {
		info.Latest = block
		info.LatestTurn = turn
	}

	if block.CanSeeEnvironment() {
		if info.LatestEnvironment == nil || turn > info.LatestEnvironment.Turn {
			info.LatestEnvironment = block
		}
	}

	if block.HasStarbase {
		if info.LatestStarbase == nil || turn > info.LatestStarbase.Turn {
			info.LatestStarbase = block
		}
	}
}

func (m *Merger) processPlayer(block *blocks.PlayerBlock) {
	if m.players[block.PlayerNumber] == nil {
		m.players[block.PlayerNumber] = block
	}
}
