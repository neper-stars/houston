// Package playerchanger provides functionality to modify player attributes in Stars! game files.
//
// This package can be used to change player attributes such as AI/human status,
// which is useful for taking over abandoned positions or debugging games.
//
// Example usage:
//
//	info, err := playerchanger.ReadPlayers("Game.hst")
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
	Success    bool
	BackupFile string
	Message    string
}

// ReadPlayers reads player information from a game file.
func ReadPlayers(filename string) (*FileInfo, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ReadPlayersFromBytes(filename, fileBytes)
}

// ReadPlayersFromBytes reads player information from file data.
func ReadPlayersFromBytes(filename string, fileBytes []byte) (*FileInfo, error) {
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
		Filename:   filename,
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

// ChangeToAI changes a player to AI control.
func ChangeToAI(filename string, playerNumber int) (*ChangeResult, error) {
	info, err := ReadPlayers(filename)
	if err != nil {
		return nil, err
	}

	player := info.GetPlayer(playerNumber)
	if player == nil {
		return nil, fmt.Errorf("player %d not found", playerNumber)
	}

	// Create backup
	backupName := filename + ".backup"
	if err := copyFile(filename, backupName); err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	result := &ChangeResult{
		BackupFile: backupName,
		Message:    "player modification not yet fully implemented",
	}

	return result, nil
}

// ChangeToHuman changes a player to human control.
func ChangeToHuman(filename string, playerNumber int) (*ChangeResult, error) {
	info, err := ReadPlayers(filename)
	if err != nil {
		return nil, err
	}

	player := info.GetPlayer(playerNumber)
	if player == nil {
		return nil, fmt.Errorf("player %d not found", playerNumber)
	}

	// Create backup
	backupName := filename + ".backup"
	if err := copyFile(filename, backupName); err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	result := &ChangeResult{
		BackupFile: backupName,
		Message:    "player modification not yet fully implemented",
	}

	return result, nil
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
