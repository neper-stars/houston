// Package playerchanger provides functionality to modify player attributes in Stars! game files.
//
// This package can be used to change player attributes such as AI/human status,
// which is useful for taking over abandoned positions or debugging games.
//
// The library operates entirely in memory - callers are responsible for reading files
// from and writing files to their storage (disk, database, etc.).
//
// Example usage:
//
//	data, _ := os.ReadFile("Game.hst")
//	info, err := playerchanger.ReadPlayersFromBytes("Game.hst", data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, player := range info.Players {
//	    fmt.Printf("Player %d: %s (%s)\n", player.Number, player.Name, player.Status)
//	}
//
//	// Change player 0 to AI (CA Expert)
//	modified, result, err := playerchanger.ChangeToAIBytes(data, 0, store.AIExpertCA)
package playerchanger

import (
	"fmt"
	"io"
	"os"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
	"github.com/neper-stars/houston/store"
)

// PlayerInfo contains information about a player.
type PlayerInfo struct {
	Number              int
	Name                string
	PluralName          string
	ShipDesignCount     int
	StarbaseDesignCount int
	OwnedPlanets        int // Number of planets owned by this player
	Fleets              int
	Status              string // Human, Human (Inactive), or AI (XX)
	StatusType          store.PlayerStatus
	AIExpertType        store.AIExpertType // Only valid if StatusType == PlayerStatusAI
	Block               *blocks.PlayerBlock
}

// FileInfo contains information about players in a file.
type FileInfo struct {
	Filename   string
	Size       int
	BlockCount int
	GameID     uint32
	Turn       uint16
	Year       int
	Players    []PlayerInfo
}

// ChangeResult contains the result of a player change operation.
type ChangeResult struct {
	Success        bool
	Message        string
	PreviousStatus string
	NewStatus      string
}

// ReadPlayers reads player information from a game file.
func ReadPlayers(filename string) (*FileInfo, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ReadPlayersFromBytes(filename, fileBytes)
}

// ReadPlayersFromReader reads player information from an io.Reader.
func ReadPlayersFromReader(name string, r io.Reader) (*FileInfo, error) {
	fileBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	return ReadPlayersFromBytes(name, fileBytes)
}

// ReadPlayersFromBytes reads player information from file data.
// The name parameter is used for display purposes only.
func ReadPlayersFromBytes(name string, fileBytes []byte) (*FileInfo, error) {
	fd := parser.FileData(fileBytes)

	header, err := fd.FileHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to parse file header: %w", err)
	}

	blockList, err := fd.BlockList()
	if err != nil {
		return nil, fmt.Errorf("failed to parse blocks: %w", err)
	}

	info := &FileInfo{
		Filename:   name,
		Size:       len(fileBytes),
		BlockCount: len(blockList),
		GameID:     header.GameID,
		Turn:       header.Turn,
		Year:       header.Year(),
		Players:    make([]PlayerInfo, 0),
	}

	// Create a temporary store to use its player status detection
	gs := store.New()
	if err := gs.AddFile(name, fileBytes); err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	for _, block := range blockList {
		if p, ok := block.(blocks.PlayerBlock); ok {
			// Get the player entity from store for status info
			playerEntity, _ := gs.Player(p.PlayerNumber)

			// Get actual owned planet count from store
			ownedPlanets := gs.PlanetsByOwner(p.PlayerNumber)

			pi := PlayerInfo{
				Number:              p.PlayerNumber,
				Name:                p.NameSingular,
				PluralName:          p.NamePlural,
				ShipDesignCount:     p.ShipDesignCount,
				StarbaseDesignCount: p.StarbaseDesignCount,
				OwnedPlanets:        len(ownedPlanets),
				Fleets:              p.Fleets,
				Block:               &p,
			}

			if playerEntity != nil {
				pi.Status = playerEntity.GetStatusString()
				pi.StatusType = playerEntity.GetStatus()
				pi.AIExpertType = playerEntity.GetAIExpertType()
			} else {
				pi.Status = "Unknown"
				pi.StatusType = store.PlayerStatusHuman
			}

			info.Players = append(info.Players, pi)
		}
	}

	return info, nil
}

// GetPlayer returns information about a specific player.
func (fi *FileInfo) GetPlayer(number int) *PlayerInfo {
	for i := range fi.Players {
		if fi.Players[i].Number == number {
			return &fi.Players[i]
		}
	}
	return nil
}

// PlayerCount returns the number of players found.
func (fi *FileInfo) PlayerCount() int {
	return len(fi.Players)
}

// ChangeToAIBytes changes a player to AI control and returns the modified data.
// The expertType parameter specifies which AI expert to use.
func ChangeToAIBytes(data []byte, playerNumber int, expertType store.AIExpertType) ([]byte, *ChangeResult, error) {
	gs := store.New()
	if err := gs.AddFile("game.hst", data); err != nil {
		return nil, nil, fmt.Errorf("failed to parse file: %w", err)
	}

	player, ok := gs.Player(playerNumber)
	if !ok {
		return nil, nil, fmt.Errorf("player %d not found", playerNumber)
	}

	previousStatus := player.GetStatusString()

	if err := player.ChangeToAI(expertType); err != nil {
		return nil, nil, fmt.Errorf("failed to change player: %w", err)
	}

	// Regenerate the HST file with the modified player
	modified, err := gs.RegenerateHSTFile()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to regenerate file: %w", err)
	}

	result := &ChangeResult{
		Success:        true,
		Message:        fmt.Sprintf("Changed player %d from %s to AI (%s)", playerNumber, previousStatus, expertType.ShortName()),
		PreviousStatus: previousStatus,
		NewStatus:      fmt.Sprintf("AI (%s)", expertType.ShortName()),
	}

	return modified, result, nil
}

// ChangeToAIReader changes a player to AI control from data in an io.Reader.
func ChangeToAIReader(r io.Reader, playerNumber int, expertType store.AIExpertType) ([]byte, *ChangeResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read data: %w", err)
	}
	return ChangeToAIBytes(data, playerNumber, expertType)
}

// ChangeToHumanBytes changes a player to human control and returns the modified data.
func ChangeToHumanBytes(data []byte, playerNumber int) ([]byte, *ChangeResult, error) {
	gs := store.New()
	if err := gs.AddFile("game.hst", data); err != nil {
		return nil, nil, fmt.Errorf("failed to parse file: %w", err)
	}

	player, ok := gs.Player(playerNumber)
	if !ok {
		return nil, nil, fmt.Errorf("player %d not found", playerNumber)
	}

	previousStatus := player.GetStatusString()

	if err := player.ChangeToHuman(); err != nil {
		return nil, nil, fmt.Errorf("failed to change player: %w", err)
	}

	// Regenerate the HST file with the modified player
	modified, err := gs.RegenerateHSTFile()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to regenerate file: %w", err)
	}

	result := &ChangeResult{
		Success:        true,
		Message:        fmt.Sprintf("Changed player %d from %s to Human", playerNumber, previousStatus),
		PreviousStatus: previousStatus,
		NewStatus:      "Human",
	}

	return modified, result, nil
}

// ChangeToHumanReader changes a player to human control from data in an io.Reader.
func ChangeToHumanReader(r io.Reader, playerNumber int) ([]byte, *ChangeResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read data: %w", err)
	}
	return ChangeToHumanBytes(data, playerNumber)
}

// ChangeToInactiveBytes changes a player to Human (Inactive) and returns the modified data.
func ChangeToInactiveBytes(data []byte, playerNumber int) ([]byte, *ChangeResult, error) {
	gs := store.New()
	if err := gs.AddFile("game.hst", data); err != nil {
		return nil, nil, fmt.Errorf("failed to parse file: %w", err)
	}

	player, ok := gs.Player(playerNumber)
	if !ok {
		return nil, nil, fmt.Errorf("player %d not found", playerNumber)
	}

	previousStatus := player.GetStatusString()

	if err := player.ChangeToInactive(); err != nil {
		return nil, nil, fmt.Errorf("failed to change player: %w", err)
	}

	// Regenerate the HST file with the modified player
	modified, err := gs.RegenerateHSTFile()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to regenerate file: %w", err)
	}

	result := &ChangeResult{
		Success:        true,
		Message:        fmt.Sprintf("Changed player %d from %s to Human (Inactive)", playerNumber, previousStatus),
		PreviousStatus: previousStatus,
		NewStatus:      "Human (Inactive)",
	}

	return modified, result, nil
}

// ChangeToInactiveReader changes a player to Human (Inactive) from data in an io.Reader.
func ChangeToInactiveReader(r io.Reader, playerNumber int) ([]byte, *ChangeResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read data: %w", err)
	}
	return ChangeToInactiveBytes(data, playerNumber)
}
