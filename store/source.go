package store

import (
	"strings"
	"time"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

// FileSourceType indicates the type of Stars! file.
type FileSourceType int

const (
	SourceTypeUnknown FileSourceType = iota
	SourceTypeMFile                  // Game state file (.m*)
	SourceTypeXFile                  // Orders file (.x*)
	SourceTypeHFile                  // History file (.h*)
	SourceTypeXYFile                 // Universe file (.xy)
	SourceTypeRFile                  // Race file (.r*)
)

// String returns a human-readable source type name.
func (t FileSourceType) String() string {
	switch t {
	case SourceTypeMFile:
		return "M-File"
	case SourceTypeXFile:
		return "X-File"
	case SourceTypeHFile:
		return "H-File"
	case SourceTypeXYFile:
		return "XY-File"
	case SourceTypeRFile:
		return "R-File"
	default:
		return "Unknown"
	}
}

// FileSource tracks where data came from.
type FileSource struct {
	ID          string         // Unique identifier (filename or custom ID)
	Type        FileSourceType // M, X, or H file
	PlayerIndex int            // Player index (0-15) from file header
	Turn        uint16         // Turn number
	GameID      uint32         // Game ID (for validation)
	RawData     []byte         // Original file bytes (preserved for re-parsing)
	Blocks      []blocks.Block // Parsed blocks
	Header      *blocks.FileHeader
	AddedAt     time.Time // When this source was added
}

// DetectFileType determines the file type from the filename.
func DetectFileType(filename string) FileSourceType {
	lower := strings.ToLower(filename)

	// Check for .xy first (before .x* check)
	if strings.HasSuffix(lower, ".xy") {
		return SourceTypeXYFile
	}

	// Check for .m*, .x*, .h*, .r* patterns
	for i := len(lower) - 1; i >= 0; i-- {
		if lower[i] == '.' {
			if i+1 < len(lower) {
				switch lower[i+1] {
				case 'm':
					return SourceTypeMFile
				case 'x':
					return SourceTypeXFile
				case 'h':
					return SourceTypeHFile
				case 'r':
					return SourceTypeRFile
				}
			}
			break
		}
	}

	return SourceTypeUnknown
}

// ParseSource parses raw file data into a FileSource.
func ParseSource(id string, data []byte) (*FileSource, error) {
	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	if err != nil {
		return nil, err
	}

	source := &FileSource{
		ID:      id,
		Type:    DetectFileType(id),
		RawData: data,
		Blocks:  blockList,
		AddedAt: time.Now(),
	}

	// Extract header info
	for _, block := range blockList {
		if fh, ok := block.(blocks.FileHeader); ok {
			source.Header = &fh
			source.PlayerIndex = fh.PlayerIndex()
			source.Turn = fh.Turn
			source.GameID = fh.GameID
			break
		}
	}

	return source, nil
}

// Reparse re-parses the raw data. Useful after format understanding improves.
func (fs *FileSource) Reparse() error {
	fd := parser.FileData(fs.RawData)
	blockList, err := fd.BlockList()
	if err != nil {
		return err
	}
	fs.Blocks = blockList

	// Re-extract header
	for _, block := range blockList {
		if fh, ok := block.(blocks.FileHeader); ok {
			fs.Header = &fh
			fs.PlayerIndex = fh.PlayerIndex()
			fs.Turn = fh.Turn
			fs.GameID = fh.GameID
			break
		}
	}

	return nil
}
