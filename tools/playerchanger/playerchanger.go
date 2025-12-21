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
//	    fmt.Printf("Player %d: %s\n", player.Number, player.Name)
//	}
package playerchanger

import (
	"fmt"
	"io"
	"os"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

// PlayerInfo contains information about a player.
type PlayerInfo struct {
	Number            int
	Name              string
	PluralName        string
	ShipDesignCount   int
	StarbaseDesignCount int
	Planets           int
	Fleets            int
	Block             *blocks.PlayerBlock
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
	Success bool
	Message string
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

	for _, block := range blockList {
		if p, ok := block.(blocks.PlayerBlock); ok {
			pi := PlayerInfo{
				Number:            p.PlayerNumber,
				Name:              p.NameSingular,
				PluralName:        p.NamePlural,
				ShipDesignCount:   p.ShipDesignCount,
				StarbaseDesignCount: p.StarbaseDesignCount,
				Planets:           p.Planets,
				Fleets:            p.Fleets,
				Block:             &p,
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
func ChangeToAIBytes(data []byte, playerNumber int) ([]byte, *ChangeResult, error) {
	info, err := ReadPlayersFromBytes("", data)
	if err != nil {
		return nil, nil, err
	}

	player := info.GetPlayer(playerNumber)
	if player == nil {
		return nil, nil, fmt.Errorf("player %d not found", playerNumber)
	}

	// Note: Actual implementation would modify the player block data
	result := &ChangeResult{
		Message: "player modification not yet fully implemented",
	}

	return data, result, nil
}

// ChangeToAIReader changes a player to AI control from data in an io.Reader.
func ChangeToAIReader(r io.Reader, playerNumber int) ([]byte, *ChangeResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read data: %w", err)
	}
	return ChangeToAIBytes(data, playerNumber)
}

// ChangeToHumanBytes changes a player to human control and returns the modified data.
func ChangeToHumanBytes(data []byte, playerNumber int) ([]byte, *ChangeResult, error) {
	info, err := ReadPlayersFromBytes("", data)
	if err != nil {
		return nil, nil, err
	}

	player := info.GetPlayer(playerNumber)
	if player == nil {
		return nil, nil, fmt.Errorf("player %d not found", playerNumber)
	}

	// Note: Actual implementation would modify the player block data
	result := &ChangeResult{
		Message: "player modification not yet fully implemented",
	}

	return data, result, nil
}

// ChangeToHumanReader changes a player to human control from data in an io.Reader.
func ChangeToHumanReader(r io.Reader, playerNumber int) ([]byte, *ChangeResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read data: %w", err)
	}
	return ChangeToHumanBytes(data, playerNumber)
}
