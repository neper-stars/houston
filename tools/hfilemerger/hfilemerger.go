// Package hfilemerger provides functionality to merge data from multiple H (history) files.
//
// H files contain historical game data that can be merged to combine information
// from different players or different turns. This is useful for maintaining
// comprehensive game histories.
//
// Example usage:
//
//	merger := hfilemerger.New()
//	if err := merger.AddHFile("player1.h1"); err != nil {
//	    log.Fatal(err)
//	}
//	if err := merger.AddMFile("player1.m1"); err != nil {
//	    log.Fatal(err)
//	}
//	if err := merger.Merge(); err != nil {
//	    log.Fatal(err)
//	}
package hfilemerger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

// Merger handles merging multiple H files.
type Merger struct {
	files    map[string][]blocks.Block
	hFiles   []string
	mFiles   []string
	gameID   uint32
	planets  map[int]*PlanetInfo
	players  [16]*blocks.PlayerBlock
	designs  [16][16]*DesignInfo
	starbases [16][10]*DesignInfo

	CreateBackups bool
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
	HFilesProcessed int
	MFilesProcessed int
	PlanetsMerged   int
	DesignsMerged   int
	Warnings        []string
	BackupFiles     []string
}

// New creates a new Merger with default options.
func New() *Merger {
	return &Merger{
		files:         make(map[string][]blocks.Block),
		planets:       make(map[int]*PlanetInfo),
		CreateBackups: true,
	}
}

// AddHFile adds an H file to be merged.
func (m *Merger) AddHFile(filename string) error {
	blockList, err := readFile(filename)
	if err != nil {
		return fmt.Errorf("unable to parse file %s: %w", filename, err)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if len(ext) < 2 || ext[1] != 'h' {
		return fmt.Errorf("%s does not appear to be an H file", filename)
	}

	m.files[filename] = blockList
	m.hFiles = append(m.hFiles, filename)

	return nil
}

// AddMFile adds an M file to incorporate design data from.
func (m *Merger) AddMFile(filename string) error {
	blockList, err := readFile(filename)
	if err != nil {
		return fmt.Errorf("unable to parse file %s: %w", filename, err)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if len(ext) < 2 || ext[1] != 'm' {
		return fmt.Errorf("%s does not appear to be an M file", filename)
	}

	m.files[filename] = blockList
	m.mFiles = append(m.mFiles, filename)

	return nil
}

// HFileCount returns the number of H files added.
func (m *Merger) HFileCount() int {
	return len(m.hFiles)
}

// MFileCount returns the number of M files added.
func (m *Merger) MFileCount() int {
	return len(m.mFiles)
}

// Validate checks that all files are from the same game.
func (m *Merger) Validate() error {
	return m.checkGameIds()
}

// Merge performs the merge operation on all added files.
func (m *Merger) Merge() (*MergeResult, error) {
	if len(m.hFiles) == 0 {
		return nil, fmt.Errorf("no H files given")
	}

	// Validate game IDs
	if err := m.checkGameIds(); err != nil {
		return nil, err
	}

	// Process all files
	for _, blockList := range m.files {
		m.processFile(blockList)
	}

	result := &MergeResult{
		HFilesProcessed: len(m.hFiles),
		MFilesProcessed: len(m.mFiles),
		PlanetsMerged:   len(m.planets),
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

	// Write back H files
	if m.CreateBackups {
		for _, filename := range m.hFiles {
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

// GetDesigns returns the merged ship design data.
func (m *Merger) GetDesigns() [16][16]*DesignInfo {
	return m.designs
}

// GetStarbases returns the merged starbase design data.
func (m *Merger) GetStarbases() [16][10]*DesignInfo {
	return m.starbases
}

// GetGameID returns the game ID from the merged files.
func (m *Merger) GetGameID() uint32 {
	return m.gameID
}

func readFile(filename string) ([]blocks.Block, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	fd := parser.FileData(fileBytes)
	return fd.BlockList()
}

func (m *Merger) checkGameIds() error {
	first := true
	for filename, blockList := range m.files {
		for _, block := range blockList {
			if h, ok := block.(blocks.FileHeader); ok {
				if first {
					m.gameID = h.GameID
					first = false
				} else if h.GameID != m.gameID {
					return fmt.Errorf("game ID mismatch in %s", filename)
				}
				break
			}
		}
	}
	return nil
}

func (m *Merger) processFile(blockList []blocks.Block) {
	var fileTurn int

	for _, block := range blockList {
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

func (m *Merger) writeFile(filename string) (string, error) {
	backupName := backupFilename(filename)
	if err := copyFile(filename, backupName); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	return backupName, nil
}

func backupFilename(filename string) string {
	ext := filepath.Ext(filename)
	if len(ext) >= 2 && (ext[1] == 'h' || ext[1] == 'H') {
		return strings.TrimSuffix(filename, ext) + ".backup-" + ext[1:]
	}
	return filename + ".backup-h"
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
