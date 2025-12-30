package store

import (
	"errors"
	"os"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
	"github.com/neper-stars/houston/race"
)

var (
	ErrNotRaceFile   = errors.New("not a race file")
	ErrNoPlayerBlock = errors.New("no player block found in race file")
)

// ParseRaceData parses race file data and returns the Race configuration.
func ParseRaceData(data []byte) (*race.Race, error) {
	// Parse blocks
	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	if err != nil {
		return nil, err
	}

	// Find the PlayerBlock
	for _, block := range blockList {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			return PlayerBlockToRace(&pb), nil
		}
	}

	return nil, ErrNoPlayerBlock
}

// ValidateRaceData parses and validates race file data, returning any validation errors.
// Returns the parsed Race (even if invalid) and any validation errors.
func ValidateRaceData(data []byte) (*race.Race, []race.ValidationError) {
	r, err := ParseRaceData(data)
	if err != nil {
		return nil, []race.ValidationError{{
			Field:   "File",
			Message: err.Error(),
		}}
	}

	errs := race.Validate(r, true)
	return r, errs
}

// ParseRaceFile reads and parses a race file (.r1-.r16) from disk.
func ParseRaceFile(filename string) (*race.Race, error) {
	if DetectFileType(filename) != SourceTypeRFile {
		return nil, ErrNotRaceFile
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return ParseRaceData(data)
}

// ValidateRaceFile reads, parses and validates a race file from disk.
// Returns the parsed Race (even if invalid) and any validation errors.
func ValidateRaceFile(filename string) (*race.Race, []race.ValidationError) {
	if DetectFileType(filename) != SourceTypeRFile {
		return nil, []race.ValidationError{{
			Field:   "File",
			Message: ErrNotRaceFile.Error(),
		}}
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, []race.ValidationError{{
			Field:   "File",
			Message: err.Error(),
		}}
	}

	return ValidateRaceData(data)
}
