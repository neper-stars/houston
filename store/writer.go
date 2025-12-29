package store

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/crypto"
)

var (
	ErrNoSourceForPlayer = errors.New("no source file found for player")
	ErrNoRawBlockData    = errors.New("no raw block data available for encoding")
)

// FileWriter handles writing Stars! game files.
type FileWriter struct {
	encryptor *crypto.Encryptor
	encoder   *BlockEncoder
}

// NewFileWriter creates a new file writer.
func NewFileWriter() *FileWriter {
	return &FileWriter{
		encryptor: crypto.NewEncryptor(),
		encoder:   NewBlockEncoder(),
	}
}

// WriteHeader encodes a file header block (not encrypted).
func (w *FileWriter) WriteHeader(header *blocks.FileHeader) []byte {
	// File header is NOT encrypted - use raw block data
	data := header.BlockData()
	return w.encoder.EncodeBlock(blocks.FileHeaderBlockType, data)
}

// WriteFooter encodes a file footer block (not encrypted).
// The footerData parameter is the 2-byte footer metadata (0 for X/H files).
func (w *FileWriter) WriteFooter(hasFooterData bool, footerData uint16) []byte {
	var data []byte
	if hasFooterData {
		data = make([]byte, 2)
		binary.LittleEndian.PutUint16(data, footerData)
	}
	return w.encoder.EncodeBlock(blocks.FileFooterBlockType, data)
}

// InitEncryption initializes encryption with game parameters.
func (w *FileWriter) InitEncryption(salt, gameId, turn, playerIndex, shareware int) {
	w.encryptor.InitEncryption(salt, gameId, turn, playerIndex, shareware)
}

// WriteEncryptedBlock writes a block with encryption.
func (w *FileWriter) WriteEncryptedBlock(typeID blocks.BlockTypeID, decryptedData []byte) []byte {
	encryptedData := w.encryptor.EncryptBytes(decryptedData)
	return w.encoder.EncodeBlock(typeID, encryptedData)
}

// GenerateMFile generates an M file for a specific player from the GameStore.
// It rebuilds the file from the stored entities and their raw block data.
func (gs *GameStore) GenerateMFile(playerIndex int) ([]byte, error) {
	// Find the source file for this player to get encryption parameters
	var sourceFile *FileSource
	for _, source := range gs.sources {
		if source.PlayerIndex == playerIndex && source.Type == SourceTypeMFile {
			sourceFile = source
			break
		}
	}

	if sourceFile == nil {
		return nil, fmt.Errorf("%w: player %d", ErrNoSourceForPlayer, playerIndex)
	}

	return gs.generateFileFromSource(sourceFile)
}

// GenerateXFile generates an X file (orders file) for a specific player.
// X files contain only the change blocks (commands) for submission.
func (gs *GameStore) GenerateXFile(playerIndex int) ([]byte, error) {
	// Find a source file for this player to get encryption parameters
	var sourceFile *FileSource
	for _, source := range gs.sources {
		if source.PlayerIndex == playerIndex {
			sourceFile = source
			break
		}
	}

	if sourceFile == nil {
		return nil, fmt.Errorf("%w: player %d", ErrNoSourceForPlayer, playerIndex)
	}

	return gs.generateXFileFromSource(sourceFile)
}

// GenerateXYFile generates an XY file (universe file).
func (gs *GameStore) GenerateXYFile() ([]byte, error) {
	// Find an XY file source
	var sourceFile *FileSource
	for _, source := range gs.sources {
		if source.Type == SourceTypeXYFile {
			sourceFile = source
			break
		}
	}

	if sourceFile == nil {
		return nil, errors.New("no XY file source found")
	}

	return gs.generateXYFileFromSource(sourceFile)
}

// GenerateHFile generates an H file (history file) for a specific player.
func (gs *GameStore) GenerateHFile(playerIndex int) ([]byte, error) {
	// Find an H file source for this player
	var sourceFile *FileSource
	for _, source := range gs.sources {
		if source.PlayerIndex == playerIndex && source.Type == SourceTypeHFile {
			sourceFile = source
			break
		}
	}

	if sourceFile == nil {
		return nil, fmt.Errorf("%w: player %d", ErrNoSourceForPlayer, playerIndex)
	}

	return gs.generateHFileFromSource(sourceFile)
}

// GenerateRFile generates an R file (race file) for a specific player slot.
// Race files are identified by their extension (.r1-.r16).
func (gs *GameStore) GenerateRFile(playerSlot int) ([]byte, error) {
	// Find an R file source for this player slot
	var sourceFile *FileSource
	for _, source := range gs.sources {
		if source.PlayerIndex == playerSlot && source.Type == SourceTypeRFile {
			sourceFile = source
			break
		}
	}

	if sourceFile == nil {
		return nil, fmt.Errorf("%w: player slot %d", ErrNoSourceForPlayer, playerSlot)
	}

	return gs.generateRFileFromSource(sourceFile)
}

// generateFileFromSource generates a complete file from a source template.
// If entities have been modified (marked dirty), their blocks are re-encoded.
func (gs *GameStore) generateFileFromSource(source *FileSource) ([]byte, error) {
	writer := NewFileWriter()
	encoder := NewBlockEncoder()
	var result []byte

	header := source.Header
	if header == nil {
		return nil, ErrNoHeader
	}

	// Write file header (not encrypted)
	result = append(result, writer.WriteHeader(header)...)

	// Initialize encryption
	shareware := 0
	if header.Shareware() {
		shareware = 1
	}
	writer.InitEncryption(header.Salt(), int(header.GameID), int(header.Turn), header.PlayerIndex(), shareware)

	// Track current fleet and planet for association
	var currentFleetKey *EntityKey
	var lastPlanetNumber int = -1

	// Write all blocks from the source, replacing dirty entities with re-encoded data
	for _, block := range source.Blocks {
		typeID := block.BlockTypeID()

		// Skip header (already written) and footer (written at end)
		if typeID == blocks.FileHeaderBlockType {
			continue
		}
		if typeID == blocks.FileFooterBlockType {
			continue
		}

		var decrypted []byte

		// Check if this block corresponds to a dirty entity and needs re-encoding
		switch b := block.(type) {
		case blocks.FleetBlock:
			currentFleetKey = &EntityKey{Type: EntityTypeFleet, Owner: b.Owner, Number: b.FleetNumber}
			lastPlanetNumber = -1
			if fleet, ok := gs.Fleets.Get(*currentFleetKey); ok && fleet.Meta().Dirty {
				if encoded, err := encoder.EncodeFleetBlock(fleet); err == nil {
					decrypted = encoded
				} else {
					decrypted = block.DecryptedData()
				}
			} else {
				decrypted = block.DecryptedData()
			}

		case blocks.PartialFleetBlock:
			currentFleetKey = &EntityKey{Type: EntityTypeFleet, Owner: b.Owner, Number: b.FleetNumber}
			lastPlanetNumber = -1
			if fleet, ok := gs.Fleets.Get(*currentFleetKey); ok && fleet.Meta().Dirty {
				if encoded, err := encoder.EncodeFleetBlock(fleet); err == nil {
					decrypted = encoded
				} else {
					decrypted = block.DecryptedData()
				}
			} else {
				decrypted = block.DecryptedData()
			}

		case blocks.PlanetBlock:
			lastPlanetNumber = b.PlanetNumber
			currentFleetKey = nil
			decrypted = block.DecryptedData()

		case blocks.PartialPlanetBlock:
			lastPlanetNumber = b.PlanetNumber
			currentFleetKey = nil
			decrypted = block.DecryptedData()

		case blocks.ProductionQueueBlock:
			// Find the production queue entity for the last planet
			if lastPlanetNumber >= 0 {
				if pq, ok := gs.ProductionQueues.GetByOwnerAndNumber(EntityTypeProductionQueue, -1, lastPlanetNumber); ok && pq.Meta().Dirty {
					if encoded, err := encoder.EncodeProductionQueueBlock(pq); err == nil {
						decrypted = encoded
					} else {
						decrypted = block.DecryptedData()
					}
				} else {
					decrypted = block.DecryptedData()
				}
			} else {
				decrypted = block.DecryptedData()
			}

		case blocks.BattlePlanBlock:
			key := EntityKey{Type: EntityTypeBattlePlan, Owner: source.PlayerIndex, Number: b.PlanId}
			if bp, ok := gs.BattlePlans.Get(key); ok && bp.Meta().Dirty {
				if encoded, err := encoder.EncodeBattlePlanBlock(bp); err == nil {
					decrypted = encoded
				} else {
					decrypted = block.DecryptedData()
				}
			} else {
				decrypted = block.DecryptedData()
			}

		case blocks.WaypointBlock, blocks.WaypointTaskBlock:
			// For waypoints, check if the associated fleet is dirty
			// If so, we might need to re-encode (for now, use original data)
			decrypted = block.DecryptedData()

		case blocks.ObjectBlock:
			currentFleetKey = nil
			lastPlanetNumber = -1
			decrypted = block.DecryptedData()

		default:
			// For all other block types, use original data
			decrypted = block.DecryptedData()
		}

		result = append(result, writer.WriteEncryptedBlock(typeID, decrypted)...)

		// Handle special case for PlanetsBlock (extra trailing data)
		if pb, ok := block.(blocks.PlanetsBlock); ok {
			if pb.Valid && len(pb.RawPlanetsData) > 0 {
				// The planets data follows the block, encrypted separately
				encryptedPlanets := writer.encryptor.EncryptBytes(pb.RawPlanetsData)
				result = append(result, encryptedPlanets...)
			}
		}
	}

	// Write file footer with turn number as footer data
	footerData := mFileFooterData(header)
	result = append(result, writer.WriteFooter(true, footerData)...)

	return result, nil
}

// generateXFileFromSource generates an X file (orders) from a source template.
func (gs *GameStore) generateXFileFromSource(source *FileSource) ([]byte, error) {
	writer := NewFileWriter()
	var result []byte

	header := source.Header
	if header == nil {
		return nil, ErrNoHeader
	}

	// Write file header (not encrypted)
	result = append(result, writer.WriteHeader(header)...)

	// Initialize encryption
	shareware := 0
	if header.Shareware() {
		shareware = 1
	}
	writer.InitEncryption(header.Salt(), int(header.GameID), int(header.Turn), header.PlayerIndex(), shareware)

	// If source is already an X file, preserve all blocks for round-trip
	// Otherwise, filter to command blocks only and add SaveAndSubmit
	if source.Type == SourceTypeXFile {
		// Preserve all blocks from X file source (round-trip)
		for _, block := range source.Blocks {
			typeID := block.BlockTypeID()

			// Skip header and footer
			if typeID == blocks.FileHeaderBlockType || typeID == blocks.FileFooterBlockType {
				continue
			}

			decrypted := block.DecryptedData()
			result = append(result, writer.WriteEncryptedBlock(typeID, decrypted)...)
		}
	} else {
		// Generate X file from M file - filter to command blocks
		for _, block := range source.Blocks {
			typeID := block.BlockTypeID()

			// Include only command blocks in X files
			if isCommandBlock(typeID) {
				decrypted := block.DecryptedData()
				result = append(result, writer.WriteEncryptedBlock(typeID, decrypted)...)
			}
		}

		// Write SaveAndSubmit block to mark as submitted
		saveSubmitData := []byte{} // Empty data for basic submission
		result = append(result, writer.WriteEncryptedBlock(blocks.SaveAndSubmitBlockType, saveSubmitData)...)
	}

	// Write file footer (X files have no footer data)
	result = append(result, writer.WriteFooter(false, 0)...)

	return result, nil
}

// generateXYFileFromSource generates an XY file (universe) from a source template.
func (gs *GameStore) generateXYFileFromSource(source *FileSource) ([]byte, error) {
	writer := NewFileWriter()
	var result []byte

	header := source.Header
	if header == nil {
		return nil, ErrNoHeader
	}

	// Write file header (not encrypted)
	result = append(result, writer.WriteHeader(header)...)

	// Initialize encryption
	shareware := 0
	if header.Shareware() {
		shareware = 1
	}
	writer.InitEncryption(header.Salt(), int(header.GameID), int(header.Turn), header.PlayerIndex(), shareware)

	// Find PlayerCount from PlanetsBlock for footer data
	var playerCount uint16
	for _, block := range source.Blocks {
		if pb, ok := block.(blocks.PlanetsBlock); ok {
			playerCount = pb.PlayerCount
			break
		}
	}

	// Write all blocks from source
	for _, block := range source.Blocks {
		typeID := block.BlockTypeID()

		// Skip header and footer
		if typeID == blocks.FileHeaderBlockType || typeID == blocks.FileFooterBlockType {
			continue
		}

		decrypted := block.DecryptedData()
		result = append(result, writer.WriteEncryptedBlock(typeID, decrypted)...)

		// Handle PlanetsBlock trailing data (stored unencrypted in file format)
		if pb, ok := block.(blocks.PlanetsBlock); ok {
			if pb.Valid && len(pb.RawPlanetsData) > 0 {
				result = append(result, pb.RawPlanetsData...)
			}
		}
	}

	// Write file footer with PlayerCount as footer data
	footerData := xyFileFooterData(playerCount)
	result = append(result, writer.WriteFooter(true, footerData)...)

	return result, nil
}

// generateHFileFromSource generates an H file (history) from a source template.
func (gs *GameStore) generateHFileFromSource(source *FileSource) ([]byte, error) {
	writer := NewFileWriter()
	var result []byte

	header := source.Header
	if header == nil {
		return nil, ErrNoHeader
	}

	// Write file header (not encrypted)
	result = append(result, writer.WriteHeader(header)...)

	// Initialize encryption
	shareware := 0
	if header.Shareware() {
		shareware = 1
	}
	writer.InitEncryption(header.Salt(), int(header.GameID), int(header.Turn), header.PlayerIndex(), shareware)

	// Write all blocks from source
	for _, block := range source.Blocks {
		typeID := block.BlockTypeID()

		// Skip header and footer
		if typeID == blocks.FileHeaderBlockType || typeID == blocks.FileFooterBlockType {
			continue
		}

		decrypted := block.DecryptedData()
		result = append(result, writer.WriteEncryptedBlock(typeID, decrypted)...)
	}

	// Write file footer (H files have no footer data)
	result = append(result, writer.WriteFooter(false, 0)...)

	return result, nil
}

// generateRFileFromSource generates an R file (race file) from a source template.
// Race file structure: FileHeader (16 bytes) + PlayerBlock (encrypted) + FileFooter (2 bytes).
// The footer is a checksum computed from the decrypted race data.
func (gs *GameStore) generateRFileFromSource(source *FileSource) ([]byte, error) {
	writer := NewFileWriter()
	var result []byte

	header := source.Header
	if header == nil {
		return nil, ErrNoHeader
	}

	// Write file header (not encrypted)
	result = append(result, writer.WriteHeader(header)...)

	// Initialize encryption for race files: gameId=0, turn=0, playerIndex=31
	writer.InitEncryption(header.Salt(), 0, 0, 31, 0)

	// Find the PlayerBlock and track decrypted data for footer computation
	var decryptedPlayerBlockData []byte
	var singularName, pluralName string

	for _, block := range source.Blocks {
		typeID := block.BlockTypeID()

		// Skip header and footer
		if typeID == blocks.FileHeaderBlockType || typeID == blocks.FileFooterBlockType {
			continue
		}

		decrypted := block.DecryptedData()
		result = append(result, writer.WriteEncryptedBlock(typeID, decrypted)...)

		// Capture PlayerBlock data for footer computation
		if typeID == blocks.PlayerBlockType {
			decryptedPlayerBlockData = decrypted
			// Parse race names from PlayerBlock
			if pb, ok := block.(*blocks.PlayerBlock); ok {
				singularName = pb.NameSingular
				pluralName = pb.NamePlural
			}
		}
	}

	// Write file footer with computed race checksum
	footerData := rFileFooterData(decryptedPlayerBlockData, singularName, pluralName)
	result = append(result, writer.WriteFooter(true, footerData)...)

	return result, nil
}

// isCommandBlock returns true if the block type is a command/order block for X files.
func isCommandBlock(typeID blocks.BlockTypeID) bool {
	switch typeID {
	case blocks.WaypointDeleteBlockType,
		blocks.WaypointAddBlockType,
		blocks.WaypointChangeTaskBlockType,
		blocks.WaypointRepeatOrdersBlockType,
		blocks.ManualSmallLoadUnloadTaskBlockType,
		blocks.ManualMediumLoadUnloadTaskBlockType,
		blocks.ManualLargeLoadUnloadTaskBlockType,
		blocks.MoveShipsBlockType,
		blocks.FleetSplitBlockType,
		blocks.FleetsMergeBlockType,
		blocks.RenameFleetBlockType,
		blocks.DesignChangeBlockType,
		blocks.ProductionQueueChangeBlockType,
		blocks.BattlePlanBlockType,
		blocks.SetFleetBattlePlanBlockType,
		blocks.ResearchChangeBlockType,
		blocks.PlanetChangeBlockType,
		blocks.ChangePasswordBlockType,
		blocks.PlayersRelationChangeBlockType,
		blocks.MessageBlockType,
		blocks.MessagesFilterBlockType:
		return true
	default:
		return false
	}
}

// mFileFooterData returns the footer data for an M file.
// M file footer data = Turn number (from header).
func mFileFooterData(header *blocks.FileHeader) uint16 {
	return header.Turn
}

// xyFileFooterData returns the footer data for an XY file.
// XY file footer data = PlayerCount (from PlanetsBlock).
func xyFileFooterData(playerCount uint16) uint16 {
	return playerCount
}

// rFileFooterData returns the footer data for an R file (race file).
// R file footer = blocks.ComputeRaceFooter(decrypted PlayerBlock data, singular name, plural name).
func rFileFooterData(decryptedPlayerBlockData []byte, singularName, pluralName string) uint16 {
	return blocks.ComputeRaceFooter(decryptedPlayerBlockData, singularName, pluralName)
}

// RegenerateMFile creates a new M file with any modified entities re-encoded.
// This is the primary method for saving changes back to a file.
func (gs *GameStore) RegenerateMFile(playerIndex int) ([]byte, error) {
	// Find the source file for this player
	var sourceFile *FileSource
	for _, source := range gs.sources {
		if source.PlayerIndex == playerIndex && source.Type == SourceTypeMFile {
			sourceFile = source
			break
		}
	}

	if sourceFile == nil {
		return nil, fmt.Errorf("%w: player %d", ErrNoSourceForPlayer, playerIndex)
	}

	return gs.regenerateWithChanges(sourceFile)
}

// regenerateWithChanges regenerates a file, replacing dirty entities with re-encoded blocks.
func (gs *GameStore) regenerateWithChanges(source *FileSource) ([]byte, error) {
	writer := NewFileWriter()
	var result []byte

	header := source.Header
	if header == nil {
		return nil, ErrNoHeader
	}

	// Write file header
	result = append(result, writer.WriteHeader(header)...)

	// Initialize encryption
	shareware := 0
	if header.Shareware() {
		shareware = 1
	}
	writer.InitEncryption(header.Salt(), int(header.GameID), int(header.Turn), header.PlayerIndex(), shareware)

	// Track which entities we've replaced
	replacedFleets := make(map[EntityKey]bool)
	replacedQueues := make(map[EntityKey]bool)

	// Process all blocks from the source
	for _, block := range source.Blocks {
		typeID := block.BlockTypeID()

		// Skip header and footer
		if typeID == blocks.FileHeaderBlockType || typeID == blocks.FileFooterBlockType {
			continue
		}

		var decrypted []byte

		switch b := block.(type) {
		case blocks.FleetBlock:
			key := EntityKey{Type: EntityTypeFleet, Owner: b.Owner, Number: b.FleetNumber}
			if fleet, ok := gs.Fleets.Get(key); ok && fleet.Meta().Dirty {
				encoded, err := writer.encoder.EncodeFleetBlock(fleet)
				if err == nil {
					decrypted = encoded
					replacedFleets[key] = true
				}
			}
		case blocks.PartialFleetBlock:
			key := EntityKey{Type: EntityTypeFleet, Owner: b.Owner, Number: b.FleetNumber}
			if fleet, ok := gs.Fleets.Get(key); ok && fleet.Meta().Dirty {
				encoded, err := writer.encoder.EncodeFleetBlock(fleet)
				if err == nil {
					decrypted = encoded
					replacedFleets[key] = true
				}
			}
		case blocks.ProductionQueueBlock:
			// Production queues need special handling - we need to know the planet number
			// For now, use original data
		}

		// Use original data if not replaced
		if decrypted == nil {
			decrypted = block.DecryptedData()
		}

		result = append(result, writer.WriteEncryptedBlock(typeID, decrypted)...)

		// Handle PlanetsBlock trailing data (stored unencrypted in file format)
		if pb, ok := block.(blocks.PlanetsBlock); ok {
			if pb.Valid && len(pb.RawPlanetsData) > 0 {
				result = append(result, pb.RawPlanetsData...)
			}
		}
	}

	_ = replacedQueues // silence unused warning

	// Write footer with turn number as footer data
	footerData := mFileFooterData(header)
	result = append(result, writer.WriteFooter(true, footerData)...)

	return result, nil
}
