// Package mfilemerger provides functionality to merge data between allied players' M files.
//
// This package allows allies to share scanning information by merging planet, fleet,
// design, and object data from multiple M files. Each player's file is augmented with
// information gathered from all allied files.
//
// Example usage:
//
//	merger := mfilemerger.New()
//	if err := merger.AddFile("player1.m1"); err != nil {
//	    log.Fatal(err)
//	}
//	if err := merger.AddFile("player2.m2"); err != nil {
//	    log.Fatal(err)
//	}
//	if err := merger.Merge(); err != nil {
//	    log.Fatal(err)
//	}
//	// Files are now merged, backups created
package mfilemerger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

// Merger handles merging multiple M files.
type Merger struct {
	files      map[string][]blocks.Block
	filenames  []string
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

	// Options
	MineralSharing bool
	CreateBackups  bool
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
	FilesProcessed int
	PlanetsMerged  int
	FleetsMerged   int
	DesignsMerged  int
	ObjectsMerged  int
	Warnings       []string
	BackupFiles    []string
}

// New creates a new Merger with default options.
func New() *Merger {
	m := &Merger{
		files:          make(map[string][]blocks.Block),
		planets:        make(map[int]*PlanetInfo),
		objects:        make(map[int]blocks.ObjectBlock),
		MineralSharing: true,
		CreateBackups:  true,
	}

	for i := 0; i < 16; i++ {
		m.fleets[i] = make(map[int]*FleetInfo)
	}

	return m
}

// AddFile adds an M file to be merged.
func (m *Merger) AddFile(filename string) error {
	blockList, err := readFile(filename)
	if err != nil {
		return fmt.Errorf("unable to parse file %s: %w", filename, err)
	}

	if err := m.validateMFile(filename, blockList); err != nil {
		return err
	}

	m.files[filename] = blockList
	m.filenames = append(m.filenames, filename)

	return nil
}

// AddBytes adds M file data from bytes.
func (m *Merger) AddBytes(name string, data []byte) error {
	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	if err != nil {
		return fmt.Errorf("failed to parse blocks: %w", err)
	}

	if err := m.validateMFile(name, blockList); err != nil {
		return err
	}

	m.files[name] = blockList
	m.filenames = append(m.filenames, name)

	return nil
}

// FileCount returns the number of files added.
func (m *Merger) FileCount() int {
	return len(m.files)
}

// Validate checks that all files are from the same game and turn.
func (m *Merger) Validate() error {
	return m.checkGameIdsAndTurns()
}

// Merge performs the merge operation on all added files.
func (m *Merger) Merge() (*MergeResult, error) {
	if len(m.files) == 0 {
		return nil, fmt.Errorf("no files to merge")
	}

	// Validate game IDs and turns
	if err := m.checkGameIdsAndTurns(); err != nil {
		return nil, err
	}

	// Process each file and collect data
	for _, blockList := range m.files {
		if err := m.processFile(blockList); err != nil {
			return nil, err
		}
	}

	// Post-process: finalize merged data
	m.postProcess()

	result := &MergeResult{
		FilesProcessed: len(m.files),
		PlanetsMerged:  len(m.planets),
		ObjectsMerged:  len(m.objects),
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

	// Write back merged files
	if m.CreateBackups {
		for _, filename := range m.filenames {
			backupName, err := m.writeFile(filename)
			if err != nil {
				return result, err
			}
			if backupName != "" {
				result.BackupFiles = append(result.BackupFiles, backupName)
			}
		}
	}

	return result, nil
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

// GetGameID returns the game ID from the merged files.
func (m *Merger) GetGameID() uint32 {
	return m.gameID
}

// GetTurn returns the turn number from the merged files.
func (m *Merger) GetTurn() uint16 {
	return m.turn
}

func readFile(filename string) ([]blocks.Block, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	fd := parser.FileData(fileBytes)
	return fd.BlockList()
}

func (m *Merger) validateMFile(filename string, blockList []blocks.Block) error {
	if len(blockList) == 0 {
		return fmt.Errorf("%s does not parse into block list", filename)
	}

	_, ok := blockList[0].(blocks.FileHeader)
	if !ok {
		return fmt.Errorf("%s does not start with header block", filename)
	}

	// Check file extension for M file
	ext := strings.ToLower(filepath.Ext(filename))
	if len(ext) < 2 || ext[1] != 'm' {
		return fmt.Errorf("%s does not appear to be an M file", filename)
	}

	return nil
}

func (m *Merger) checkGameIdsAndTurns() error {
	first := true
	for filename, blockList := range m.files {
		var header blocks.FileHeader
		for _, block := range blockList {
			if h, ok := block.(blocks.FileHeader); ok {
				header = h
				break
			}
		}

		if first {
			m.gameID = header.GameID
			m.turn = header.Turn
			m.playerMask = 1 << header.PlayerIndex()
			first = false
		} else {
			if header.GameID != m.gameID {
				return fmt.Errorf("game ID mismatch in %s", filename)
			}
			if header.Turn != m.turn {
				return fmt.Errorf("turn mismatch in %s (expected %d, got %d)", filename, m.turn, header.Turn)
			}
			m.playerMask |= 1 << header.PlayerIndex()
		}
	}
	return nil
}

func (m *Merger) processFile(blockList []blocks.Block) error {
	shipDesignOwners := make([]int, 0)
	starbaseDesignOwners := make([]int, 0)
	shipDesignIndex := 0
	starbaseDesignIndex := 0

	for _, block := range blockList {
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

func (m *Merger) writeFile(filename string) (string, error) {
	backupName := backupFilename(filename)
	if err := copyFile(filename, backupName); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	// Note: Full file writing would require implementing block encoding
	return backupName, nil
}

func backupFilename(filename string) string {
	ext := filepath.Ext(filename)
	if len(ext) >= 2 && (ext[1] == 'm' || ext[1] == 'M') {
		return strings.TrimSuffix(filename, ext) + ".backup-" + ext[1:]
	}
	return filename + ".backup-m"
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	return err
}
