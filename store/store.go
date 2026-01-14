package store

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/neper-stars/houston/blocks"
)

var (
	ErrGameIDMismatch = errors.New("game ID mismatch")
	ErrNoHeader       = errors.New("file has no header block")
)

// GameStore aggregates game state from multiple Stars! files.
type GameStore struct {
	// Game identification
	GameID   uint32
	GameName string
	Turn     uint16

	// Sources (preserved for re-parsing and tracking)
	sources     map[string]*FileSource
	sourceOrder []string // Preserve add order

	// Conflict resolution
	resolver ConflictResolver

	// Universe data (from PlanetsBlock)
	planetNames      map[int]string // Planet number -> name
	UniverseSize     uint16         // 0=Tiny, 1=Small, 2=Medium, 3=Large, 4=Huge
	Density          uint16         // 0=Sparse, 1=Normal, 2=Dense, 3=Packed
	PlayerCount      uint16         // Number of players in the game
	PlanetCount      uint16         // Total number of planets
	StartingDistance uint16         // Starting distance mode (mdStartDist)
	GameSettings     uint16         // Game options bitmask

	// Victory conditions (from PlanetsBlock)
	VictoryConditions blocks.DecodedVictoryConditions

	// Entity collections
	Fleets           *EntityCollection[*FleetEntity]
	Designs          *EntityCollection[*DesignEntity]
	Planets          *EntityCollection[*PlanetEntity]
	Players          *EntityCollection[*PlayerEntity]
	Objects          *EntityCollection[*ObjectEntity]
	BattlePlans      *EntityCollection[*BattlePlanEntity]
	ProductionQueues *EntityCollection[*ProductionQueueEntity]

	// Non-entity collections (not using EntityCollection pattern)
	Messages []*MessageEntity
	Events   []*EventsEntity
}

// New creates an empty GameStore with default conflict resolution.
func New() *GameStore {
	return NewWithResolver(&DefaultResolver{})
}

// NewWithResolver creates a GameStore with custom conflict resolution.
func NewWithResolver(resolver ConflictResolver) *GameStore {
	return &GameStore{
		sources:          make(map[string]*FileSource),
		sourceOrder:      make([]string, 0),
		resolver:         resolver,
		planetNames:      make(map[int]string),
		Fleets:           NewEntityCollection[*FleetEntity](),
		Designs:          NewEntityCollection[*DesignEntity](),
		Planets:          NewEntityCollection[*PlanetEntity](),
		Players:          NewEntityCollection[*PlayerEntity](),
		Objects:          NewEntityCollection[*ObjectEntity](),
		BattlePlans:      NewEntityCollection[*BattlePlanEntity](),
		ProductionQueues: NewEntityCollection[*ProductionQueueEntity](),
	}
}

// AddFile parses and merges data from a file.
func (gs *GameStore) AddFile(name string, data []byte) error {
	source, err := ParseSource(name, data)
	if err != nil {
		return err
	}

	if err := gs.validateSource(source); err != nil {
		return err
	}

	// Store the source
	if _, exists := gs.sources[name]; !exists {
		gs.sourceOrder = append(gs.sourceOrder, name)
	}
	gs.sources[name] = source

	// Update game info from first file
	if gs.GameID == 0 && source.Header != nil {
		gs.GameID = source.GameID
		gs.Turn = source.Turn
	}

	// Update turn if this file is newer
	if source.Turn > gs.Turn {
		gs.Turn = source.Turn
	}

	// Merge entities from this source
	return gs.mergeSource(source)
}

// AddFileReader adds from an io.Reader.
func (gs *GameStore) AddFileReader(name string, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return gs.AddFile(name, data)
}

// AddFileWithXY loads a game file and automatically loads the companion XY file
// if the input is an M or H file (to get planet coordinates).
func (gs *GameStore) AddFileWithXY(filename string) error {
	return gs.AddFileWithXYFromFS(filename, osFS{})
}

// AddFileWithXYFromFS loads a file with optional companion XY file using a filesystem interface.
func (gs *GameStore) AddFileWithXYFromFS(filename string, fs FileSystem) error {
	// First, try to load companion XY file for M/H files
	xyFile := findCompanionXYFile(filename, fs)
	if xyFile != "" {
		// Load XY file first to get planet coordinates
		data, err := fs.ReadFile(xyFile)
		if err == nil {
			// Ignore errors - just try to load
			_ = gs.AddFile(xyFile, data)
		}
	}

	// Now load the main file
	data, err := fs.ReadFile(filename)
	if err != nil {
		return err
	}
	return gs.AddFile(filename, data)
}

// FileSystem interface for abstracting file operations.
type FileSystem interface {
	ReadFile(filename string) ([]byte, error)
	Stat(filename string) (bool, error)
}

// osFS implements FileSystem using os package.
type osFS struct{}

func (osFS) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (osFS) Stat(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err != nil {
		return false, err
	}
	return true, nil
}

// findCompanionXYFile finds the XY file for a given M or H file.
// Returns empty string if not found or not applicable.
func findCompanionXYFile(filename string, fs FileSystem) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ""
	}

	// Check if this is an M or H file (e.g., .m1, .h1, .M1, .H1)
	extLower := strings.ToLower(ext)
	if len(extLower) < 2 {
		return ""
	}

	fileType := extLower[1] // 'm', 'h', 'x', etc.
	if fileType != 'm' && fileType != 'h' {
		return "" // Only M and H files need companion XY
	}

	// Build the XY filename
	baseName := strings.TrimSuffix(filename, ext)
	xyFile := baseName + ".xy"

	// Check if it exists
	if ok, _ := fs.Stat(xyFile); ok {
		return xyFile
	}

	// Try uppercase
	xyFile = baseName + ".XY"
	if ok, _ := fs.Stat(xyFile); ok {
		return xyFile
	}

	return ""
}

// validateSource checks that the source is compatible with the store.
func (gs *GameStore) validateSource(source *FileSource) error {
	if source.Header == nil {
		return ErrNoHeader
	}

	// First file sets the game ID
	if gs.GameID == 0 {
		return nil
	}

	// Subsequent files must match
	if source.GameID != gs.GameID {
		return ErrGameIDMismatch
	}

	return nil
}

// mergeSource extracts and merges entities from a source.
func (gs *GameStore) mergeSource(source *FileSource) error {
	// First pass: Extract planet names from PlanetsBlock, designs, players, battle plans, messages, and events
	// Brief designs (IsFullDesign=false) are stored separately - they belong to enemy players,
	// and we'll associate them with the correct owner when processing enemy fleets.
	briefDesigns := make(map[int]*blocks.DesignBlock) // keyed by design slot
	messageIndex := 0
	for _, block := range source.Blocks {
		switch b := block.(type) {
		case blocks.PlanetsBlock:
			gs.mergePlanetsBlock(&b, source)
		case blocks.DesignBlock:
			if b.IsFullDesign {
				// Full designs belong to the file's player
				gs.mergeDesign(&b, source)
			} else {
				// Brief designs are for enemy ships - defer ownership assignment
				briefDesigns[b.DesignNumber] = &b
			}
		case blocks.PlayerBlock:
			gs.mergePlayer(&b, source)
		case blocks.BattlePlanBlock:
			gs.mergeBattlePlan(&b, source)
		case blocks.MessageBlock:
			gs.Messages = append(gs.Messages, newMessageEntityFromBlock(&b, messageIndex, source))
			messageIndex++
		case blocks.EventsBlock:
			gs.Events = append(gs.Events, newEventsEntityFromBlock(&b, source))
		case blocks.PlayerScoresBlock:
			gs.mergePlayerScores(&b, source)
		}
	}

	// Second pass: Extract fleets, planets, objects, production queues, and waypoints
	// Note: FleetNameBlock (type 21) follows the FleetBlock it names, so we track
	// the last fleet and apply the name when we see FleetNameBlock.
	// We also track enemy design slot usage to associate brief designs with correct owners.
	enemyDesignSlots := make(map[int]map[int]bool) // owner -> set of design slots used
	var currentFleet *FleetEntity
	var waypointIndex int
	var lastPlanetNumber = -1
	for _, block := range source.Blocks {
		switch b := block.(type) {
		case blocks.FleetBlock:
			currentFleet = gs.mergeFleet(&b.PartialFleetBlock, nil, source)
			// Track enemy design slot usage
			if b.Owner != source.PlayerIndex {
				if enemyDesignSlots[b.Owner] == nil {
					enemyDesignSlots[b.Owner] = make(map[int]bool)
				}
				for slot := 0; slot < 16; slot++ {
					if (b.ShipTypes & (1 << slot)) != 0 {
						enemyDesignSlots[b.Owner][slot] = true
					}
				}
			}
			waypointIndex = 0
			lastPlanetNumber = -1
		case blocks.PartialFleetBlock:
			currentFleet = gs.mergeFleet(&b, nil, source)
			// Track enemy design slot usage
			if b.Owner != source.PlayerIndex {
				if enemyDesignSlots[b.Owner] == nil {
					enemyDesignSlots[b.Owner] = make(map[int]bool)
				}
				for slot := 0; slot < 16; slot++ {
					if (b.ShipTypes & (1 << slot)) != 0 {
						enemyDesignSlots[b.Owner][slot] = true
					}
				}
			}
			waypointIndex = 0
			lastPlanetNumber = -1
		case blocks.FleetNameBlock:
			// FleetNameBlock follows the FleetBlock it names
			if currentFleet != nil {
				currentFleet.CustomName = b.Name
				currentFleet.HasCustomName = true
				currentFleet.nameBlock = &b
			}
		case blocks.WaypointBlock:
			if currentFleet != nil {
				wp := newWaypointEntityFromBlock(&b, currentFleet.Owner, currentFleet.FleetNumber, waypointIndex, source)
				currentFleet.AddWaypoint(wp)
				waypointIndex++
			}
		case blocks.WaypointTaskBlock:
			if currentFleet != nil {
				wp := newWaypointEntityFromTaskBlock(&b, currentFleet.Owner, currentFleet.FleetNumber, waypointIndex, source)
				currentFleet.AddWaypoint(wp)
				waypointIndex++
			}
		case blocks.PlanetBlock:
			gs.mergePlanet(&b.PartialPlanetBlock, source)
			lastPlanetNumber = b.PlanetNumber
			currentFleet = nil // Planet block ends fleet context
		case blocks.PartialPlanetBlock:
			gs.mergePlanet(&b, source)
			lastPlanetNumber = b.PlanetNumber
			currentFleet = nil
		case blocks.ObjectBlock:
			gs.mergeObject(&b, source)
			lastPlanetNumber = -1
			currentFleet = nil
		case blocks.ProductionQueueBlock:
			if lastPlanetNumber >= 0 {
				gs.mergeProductionQueue(&b, lastPlanetNumber, source)
			}
		}
	}

	// Third pass: Associate brief designs with enemy players
	// Brief designs are scanned enemy ship designs - we now know which enemy players
	// use which design slots, so we can assign the correct owner.
	for owner, slots := range enemyDesignSlots {
		for slot := range slots {
			// Check if we have a brief design for this slot
			if briefDesign, ok := briefDesigns[slot]; ok {
				// Create a design entity for this enemy player
				gs.mergeDesignForOwner(briefDesign, owner, source)
			}
		}
	}

	return nil
}

// mergePlanetsBlock extracts planet names, coordinates, and universe info.
func (gs *GameStore) mergePlanetsBlock(pb *blocks.PlanetsBlock, source *FileSource) {
	if !pb.Valid {
		return
	}

	// Store game name if not set
	if gs.GameName == "" {
		gs.GameName = pb.GameName
	}

	// Store universe settings (only set once from first PlanetsBlock)
	if gs.PlanetCount == 0 {
		gs.UniverseSize = pb.UniverseSize
		gs.Density = pb.Density
		gs.PlayerCount = pb.PlayerCount
		gs.PlanetCount = pb.PlanetCount
		gs.StartingDistance = pb.StartingDistance
		gs.GameSettings = pb.GameSettings
		gs.VictoryConditions = pb.GetVictoryConditions()
	}

	// Extract planet names and create/update planet entities with coordinates
	for _, planet := range pb.Planets {
		gs.planetNames[planet.ID] = planet.Name

		// Create a minimal planet entity with coordinates if it doesn't exist
		key := EntityKey{
			Type:   EntityTypePlanet,
			Owner:  -1, // Unowned by default
			Number: planet.ID,
		}

		if existing, ok := gs.Planets.Get(key); ok {
			// Update existing planet with coordinates if missing
			if existing.X == 0 && existing.Y == 0 {
				existing.X = int(planet.X)
				existing.Y = int(planet.Y)
			}
			if existing.Name == "" {
				existing.Name = planet.Name
			}
		} else {
			// Create new minimal planet entity with coordinates
			entity := &PlanetEntity{
				meta: EntityMeta{
					Key:        key,
					BestSource: source,
					Quality:    QualityMinimal,
					Turn:       source.Turn,
				},
				PlanetNumber: planet.ID,
				Owner:        -1,
				Name:         planet.Name,
				X:            int(planet.X),
				Y:            int(planet.Y),
			}
			entity.meta.AddSource(source)
			gs.Planets.Add(entity)
		}
	}
}

// mergeDesign merges a design into the store.
func (gs *GameStore) mergeDesign(db *blocks.DesignBlock, source *FileSource) {
	entity := newDesignEntityFromBlock(db, source)
	key := entity.Meta().Key

	if existing, ok := gs.Designs.Get(key); ok {
		if gs.resolver.ShouldReplace(existing, entity) {
			existing.Meta().AddSource(source)
			gs.Designs.Add(entity)
		} else {
			existing.Meta().AddSource(source)
		}
	} else {
		gs.Designs.Add(entity)
	}
}

// mergeDesignForOwner merges a design into the store with a specific owner.
// This is used for brief designs (scanned enemy ships) where the owner
// is determined from the fleet that uses the design, not the file's PlayerIndex.
func (gs *GameStore) mergeDesignForOwner(db *blocks.DesignBlock, owner int, source *FileSource) {
	entityType := EntityTypeDesign
	if db.IsStarbase {
		entityType = EntityTypeStarbaseDesign
	}

	// Brief designs (scanned) have lower quality than full designs
	quality := QualityPartial

	entity := &DesignEntity{
		meta: EntityMeta{
			Key: EntityKey{
				Type:   entityType,
				Owner:  owner,
				Number: db.DesignNumber,
			},
			BestSource: source,
			Quality:    quality,
			Turn:       source.Turn,
		},
		DesignNumber: db.DesignNumber,
		Owner:        owner,
		IsStarbase:   db.IsStarbase,
		Name:         db.Name,
		HullId:       db.HullId,
		designBlock:  db,
	}
	entity.meta.AddSource(source)

	key := entity.Meta().Key
	if existing, ok := gs.Designs.Get(key); ok {
		if gs.resolver.ShouldReplace(existing, entity) {
			existing.Meta().AddSource(source)
			gs.Designs.Add(entity)
		} else {
			existing.Meta().AddSource(source)
		}
	} else {
		gs.Designs.Add(entity)
	}
}

// mergeFleet merges a fleet into the store and returns it for waypoint association.
func (gs *GameStore) mergeFleet(fb *blocks.PartialFleetBlock, nameBlock *blocks.FleetNameBlock, source *FileSource) *FleetEntity {
	entity := newFleetEntityFromBlock(fb, source)

	// Associate name if present
	if nameBlock != nil {
		entity.CustomName = nameBlock.Name
		entity.HasCustomName = true
		entity.nameBlock = nameBlock
	}

	// Resolve primary design
	primarySlot := getPrimaryDesignSlot(fb.ShipTypes)
	if primarySlot >= 0 {
		designKey := EntityKey{
			Type:   EntityTypeDesign,
			Owner:  fb.Owner,
			Number: primarySlot,
		}
		if design, ok := gs.Designs.Get(designKey); ok {
			entity.PrimaryDesign = design
		}
	}

	key := entity.Meta().Key

	existing, ok := gs.Fleets.Get(key)
	if !ok {
		gs.Fleets.Add(entity)
		return entity
	}

	existing.Meta().AddSource(source)
	if gs.resolver.ShouldReplace(existing, entity) {
		gs.Fleets.Add(entity)
		return entity
	}
	return existing
}

// mergePlanet merges a planet into the store.
func (gs *GameStore) mergePlanet(pb *blocks.PartialPlanetBlock, source *FileSource) {
	entity := newPlanetEntityFromBlock(pb, source)

	// Resolve name from planetNames
	if name, ok := gs.planetNames[pb.PlanetNumber]; ok {
		entity.Name = name
	}

	// Try to find existing planet by number (owner may differ)
	// First check with exact key, then search by number
	key := entity.Meta().Key
	var existing *PlanetEntity
	var found bool

	if existing, found = gs.Planets.Get(key); !found {
		// Try to find by planet number with owner=-1 (from XY file)
		unownedKey := EntityKey{
			Type:   EntityTypePlanet,
			Owner:  -1,
			Number: pb.PlanetNumber,
		}
		existing, found = gs.Planets.Get(unownedKey)
	}

	if found {
		// Preserve coordinates if the new entity doesn't have them
		if entity.X == 0 && entity.Y == 0 && (existing.X != 0 || existing.Y != 0) {
			entity.X = existing.X
			entity.Y = existing.Y
		}
		// Preserve name if the new entity doesn't have it
		if entity.Name == "" && existing.Name != "" {
			entity.Name = existing.Name
		}

		if gs.resolver.ShouldReplace(existing, entity) {
			existing.Meta().AddSource(source)
			// Remove old entry if owner changed
			if existing.Owner != entity.Owner {
				gs.Planets.Remove(existing.Meta().Key)
			}
			gs.Planets.Add(entity)
		} else {
			existing.Meta().AddSource(source)
			// Still update coordinates if missing
			if existing.X == 0 && existing.Y == 0 {
				existing.X = entity.X
				existing.Y = entity.Y
			}
		}
	} else {
		gs.Planets.Add(entity)
	}
}

// mergePlayer merges a player into the store.
func (gs *GameStore) mergePlayer(pb *blocks.PlayerBlock, source *FileSource) {
	entity := newPlayerEntityFromBlock(pb, source)
	key := entity.Meta().Key

	if existing, ok := gs.Players.Get(key); ok {
		if gs.resolver.ShouldReplace(existing, entity) {
			existing.Meta().AddSource(source)
			gs.Players.Add(entity)
		} else {
			existing.Meta().AddSource(source)
		}
	} else {
		gs.Players.Add(entity)
	}
}

// mergeObject merges an object into the store.
func (gs *GameStore) mergeObject(ob *blocks.ObjectBlock, source *FileSource) {
	entity := newObjectEntityFromBlock(ob, source)
	if entity == nil {
		return // Count objects are skipped
	}

	key := entity.Meta().Key

	if existing, ok := gs.Objects.Get(key); ok {
		if gs.resolver.ShouldReplace(existing, entity) {
			existing.Meta().AddSource(source)
			gs.Objects.Add(entity)
		} else {
			existing.Meta().AddSource(source)
		}
	} else {
		gs.Objects.Add(entity)
	}
}

// mergeBattlePlan merges a battle plan into the store.
func (gs *GameStore) mergeBattlePlan(bpb *blocks.BattlePlanBlock, source *FileSource) {
	entity := newBattlePlanEntityFromBlock(bpb, source)
	key := entity.Meta().Key

	if existing, ok := gs.BattlePlans.Get(key); ok {
		if gs.resolver.ShouldReplace(existing, entity) {
			existing.Meta().AddSource(source)
			gs.BattlePlans.Add(entity)
		} else {
			existing.Meta().AddSource(source)
		}
	} else {
		gs.BattlePlans.Add(entity)
	}
}

// mergeProductionQueue merges a production queue into the store.
func (gs *GameStore) mergeProductionQueue(pqb *blocks.ProductionQueueBlock, planetNumber int, source *FileSource) {
	entity := newProductionQueueEntityFromBlock(pqb, planetNumber, source)
	key := entity.Meta().Key

	if existing, ok := gs.ProductionQueues.Get(key); ok {
		if gs.resolver.ShouldReplace(existing, entity) {
			existing.Meta().AddSource(source)
			gs.ProductionQueues.Add(entity)
		} else {
			existing.Meta().AddSource(source)
		}
	} else {
		gs.ProductionQueues.Add(entity)
	}
}

// Sources returns all added file sources in add order.
func (gs *GameStore) Sources() []*FileSource {
	result := make([]*FileSource, 0, len(gs.sourceOrder))
	for _, name := range gs.sourceOrder {
		result = append(result, gs.sources[name])
	}
	return result
}

// Source returns a specific source by ID.
func (gs *GameStore) Source(id string) (*FileSource, bool) {
	source, ok := gs.sources[id]
	return source, ok
}

// SourceCount returns the number of sources.
func (gs *GameStore) SourceCount() int {
	return len(gs.sources)
}

// PlanetName returns the name of a planet by number.
func (gs *GameStore) PlanetName(planetNumber int) string {
	return gs.planetNames[planetNumber]
}

// HasChanges returns true if any entity has been modified.
func (gs *GameStore) HasChanges() bool {
	if len(gs.Fleets.DirtyEntities()) > 0 {
		return true
	}
	if len(gs.Designs.DirtyEntities()) > 0 {
		return true
	}
	if len(gs.Planets.DirtyEntities()) > 0 {
		return true
	}
	if len(gs.Players.DirtyEntities()) > 0 {
		return true
	}
	if len(gs.Objects.DirtyEntities()) > 0 {
		return true
	}
	if len(gs.BattlePlans.DirtyEntities()) > 0 {
		return true
	}
	if len(gs.ProductionQueues.DirtyEntities()) > 0 {
		return true
	}
	return false
}

// ResetDirtyFlags clears all dirty flags.
func (gs *GameStore) ResetDirtyFlags() {
	gs.Fleets.ResetDirtyFlags()
	gs.Designs.ResetDirtyFlags()
	gs.Planets.ResetDirtyFlags()
	gs.Players.ResetDirtyFlags()
	gs.Objects.ResetDirtyFlags()
	gs.BattlePlans.ResetDirtyFlags()
	gs.ProductionQueues.ResetDirtyFlags()
}

// Fleet returns a fleet by owner and number.
func (gs *GameStore) Fleet(owner, number int) (*FleetEntity, bool) {
	return gs.Fleets.GetByOwnerAndNumber(EntityTypeFleet, owner, number)
}

// FleetsByOwner returns all fleets owned by a player.
func (gs *GameStore) FleetsByOwner(owner int) []*FleetEntity {
	return gs.Fleets.ByOwner(owner)
}

// AllFleets returns all fleets in the store.
func (gs *GameStore) AllFleets() []*FleetEntity {
	return gs.Fleets.All()
}

// Design returns a ship design by owner and slot.
func (gs *GameStore) Design(owner, slot int) (*DesignEntity, bool) {
	return gs.Designs.GetByOwnerAndNumber(EntityTypeDesign, owner, slot)
}

// StarbaseDesign returns a starbase design by owner and slot.
func (gs *GameStore) StarbaseDesign(owner, slot int) (*DesignEntity, bool) {
	return gs.Designs.GetByOwnerAndNumber(EntityTypeStarbaseDesign, owner, slot)
}

// DesignsByOwner returns all designs (both ship and starbase) owned by a player.
func (gs *GameStore) DesignsByOwner(owner int) []*DesignEntity {
	return gs.Designs.ByOwner(owner)
}

// ShipDesignsByOwner returns only ship designs owned by a player.
func (gs *GameStore) ShipDesignsByOwner(owner int) []*DesignEntity {
	all := gs.Designs.ByOwner(owner)
	result := make([]*DesignEntity, 0, len(all))
	for _, d := range all {
		if !d.IsStarbase {
			result = append(result, d)
		}
	}
	return result
}

// StarbaseDesignsByOwner returns only starbase designs owned by a player.
func (gs *GameStore) StarbaseDesignsByOwner(owner int) []*DesignEntity {
	all := gs.Designs.ByOwner(owner)
	result := make([]*DesignEntity, 0, len(all))
	for _, d := range all {
		if d.IsStarbase {
			result = append(result, d)
		}
	}
	return result
}

// AllDesigns returns all designs in the store.
func (gs *GameStore) AllDesigns() []*DesignEntity {
	return gs.Designs.All()
}

// Planet returns a planet by number.
// Note: Planets use Owner=-1 for unowned planets, so we search by number only.
func (gs *GameStore) Planet(number int) (*PlanetEntity, bool) {
	// First try to find unowned planet
	if planet, ok := gs.Planets.GetByOwnerAndNumber(EntityTypePlanet, -1, number); ok {
		return planet, true
	}
	// Then search all planets for this number
	for _, planet := range gs.Planets.All() {
		if planet.PlanetNumber == number {
			return planet, true
		}
	}
	return nil, false
}

// PlanetForSave returns a planet by number, prioritizing dirty planets.
// This is used when regenerating files to ensure we find the modified entity
// even if it's stored with a different owner key than the source block.
func (gs *GameStore) PlanetForSave(number int) (*PlanetEntity, bool) {
	// First, search all planets for a dirty one with this number
	for _, planet := range gs.Planets.All() {
		if planet.PlanetNumber == number && planet.Meta().Dirty {
			return planet, true
		}
	}
	// If no dirty planet found, return any planet with this number
	return gs.Planet(number)
}

// PlanetsByOwner returns all planets owned by a player.
// Use owner=-1 for unowned planets.
func (gs *GameStore) PlanetsByOwner(owner int) []*PlanetEntity {
	return gs.Planets.ByOwner(owner)
}

// AllPlanets returns all planets in the store.
func (gs *GameStore) AllPlanets() []*PlanetEntity {
	return gs.Planets.All()
}

// PlanetByName returns a planet by name (case-sensitive).
func (gs *GameStore) PlanetByName(name string) (*PlanetEntity, bool) {
	for _, planet := range gs.Planets.All() {
		if planet.Name == name {
			return planet, true
		}
	}
	return nil, false
}

// VisiblePlanets returns all planets visible to a player (detection level > 0).
// This filters out planets from the XY file that the player hasn't detected.
func (gs *GameStore) VisiblePlanets() []*PlanetEntity {
	var result []*PlanetEntity
	for _, planet := range gs.Planets.All() {
		if planet.DetectionLevel > blocks.DetNotVisible {
			result = append(result, planet)
		}
	}
	return result
}

// VisiblePlanetsByOwner returns planets owned by a player that are visible.
// For the player's own planets, this returns all owned planets.
// For opponent planets, this only returns planets with detection level > 0.
func (gs *GameStore) VisiblePlanetsByOwner(owner int) []*PlanetEntity {
	var result []*PlanetEntity
	for _, planet := range gs.Planets.ByOwner(owner) {
		if planet.DetectionLevel > blocks.DetNotVisible {
			result = append(result, planet)
		}
	}
	return result
}

// VisibleOpponentPlanets returns all opponent planets (owned by another player, not unowned)
// that are visible (detection level > 0).
func (gs *GameStore) VisibleOpponentPlanets(playerNumber int) []*PlanetEntity {
	var result []*PlanetEntity
	for _, planet := range gs.Planets.All() {
		// Owner >= 0 means owned by a player (Owner == -1 means unowned)
		if planet.Owner >= 0 && planet.Owner != playerNumber && planet.DetectionLevel > blocks.DetNotVisible {
			result = append(result, planet)
		}
	}
	return result
}

// CanSeePopulation returns true if the detection level allows seeing population data.
// Population is visible at detection level >= DetNormalScan (3) for enemy planets,
// or always for owned planets.
func (gs *GameStore) CanSeePopulation(planet *PlanetEntity, viewerPlayerNumber int) bool {
	if planet.Owner == viewerPlayerNumber {
		return true
	}
	return planet.DetectionLevel >= blocks.DetNormalScan
}

// CanSeeMinerals returns true if the detection level allows seeing mineral data.
// Minerals are visible at detection level >= DetNormalScan (3) for enemy planets,
// or always for owned planets.
func (gs *GameStore) CanSeeMinerals(planet *PlanetEntity, viewerPlayerNumber int) bool {
	if planet.Owner == viewerPlayerNumber {
		return true
	}
	return planet.DetectionLevel >= blocks.DetNormalScan
}

// CanSeeInstallations returns true if the detection level allows seeing installations.
// Installations (mines, factories, defenses) require DetFull (4+).
func (gs *GameStore) CanSeeInstallations(planet *PlanetEntity, viewerPlayerNumber int) bool {
	if planet.Owner == viewerPlayerNumber {
		return true
	}
	return planet.DetectionLevel >= blocks.DetFull
}

// Player returns a player by index.
func (gs *GameStore) Player(index int) (*PlayerEntity, bool) {
	return gs.Players.GetByOwnerAndNumber(EntityTypePlayer, index, index)
}

// AllPlayers returns all players in the store.
func (gs *GameStore) AllPlayers() []*PlayerEntity {
	return gs.Players.All()
}

// Object returns an object by owner and number.
func (gs *GameStore) Object(owner, number int) (*ObjectEntity, bool) {
	return gs.Objects.GetByOwnerAndNumber(EntityTypeObject, owner, number)
}

// ObjectsByOwner returns all objects owned by a player.
func (gs *GameStore) ObjectsByOwner(owner int) []*ObjectEntity {
	return gs.Objects.ByOwner(owner)
}

// AllObjects returns all objects in the store.
func (gs *GameStore) AllObjects() []*ObjectEntity {
	return gs.Objects.All()
}

// Minefields returns all minefield objects.
func (gs *GameStore) Minefields() []*ObjectEntity {
	var result []*ObjectEntity
	for _, obj := range gs.Objects.All() {
		if obj.IsMinefield() {
			result = append(result, obj)
		}
	}
	return result
}

// Wormholes returns all wormhole objects.
func (gs *GameStore) Wormholes() []*ObjectEntity {
	var result []*ObjectEntity
	for _, obj := range gs.Objects.All() {
		if obj.IsWormhole() {
			result = append(result, obj)
		}
	}
	return result
}

// Packets returns all mineral packet objects.
func (gs *GameStore) Packets() []*ObjectEntity {
	var result []*ObjectEntity
	for _, obj := range gs.Objects.All() {
		if obj.IsPacket() {
			result = append(result, obj)
		}
	}
	return result
}

// Salvage returns all salvage objects.
func (gs *GameStore) Salvage() []*ObjectEntity {
	var result []*ObjectEntity
	for _, obj := range gs.Objects.All() {
		if obj.IsSalvage {
			result = append(result, obj)
		}
	}
	return result
}

// BattlePlan returns a battle plan by owner and plan ID.
func (gs *GameStore) BattlePlan(owner, planId int) (*BattlePlanEntity, bool) {
	return gs.BattlePlans.GetByOwnerAndNumber(EntityTypeBattlePlan, owner, planId)
}

// BattlePlansByOwner returns all battle plans owned by a player.
func (gs *GameStore) BattlePlansByOwner(owner int) []*BattlePlanEntity {
	return gs.BattlePlans.ByOwner(owner)
}

// AllBattlePlans returns all battle plans in the store.
func (gs *GameStore) AllBattlePlans() []*BattlePlanEntity {
	return gs.BattlePlans.All()
}

// ProductionQueue returns a production queue by planet number.
func (gs *GameStore) ProductionQueue(planetNumber int) (*ProductionQueueEntity, bool) {
	return gs.ProductionQueues.GetByOwnerAndNumber(EntityTypeProductionQueue, -1, planetNumber)
}

// AllProductionQueues returns all production queues in the store.
func (gs *GameStore) AllProductionQueues() []*ProductionQueueEntity {
	return gs.ProductionQueues.All()
}

// AllMessages returns all messages in the store.
func (gs *GameStore) AllMessages() []*MessageEntity {
	return gs.Messages
}

// MessagesBySender returns all messages from a specific sender.
func (gs *GameStore) MessagesBySender(senderId int) []*MessageEntity {
	var result []*MessageEntity
	for _, msg := range gs.Messages {
		if msg.SenderId == senderId {
			result = append(result, msg)
		}
	}
	return result
}

// AllEvents returns all events in the store.
func (gs *GameStore) AllEvents() []*EventsEntity {
	return gs.Events
}

// EventsForTurn returns events for a specific turn.
func (gs *GameStore) EventsForTurn(turn uint16) []*EventsEntity {
	var result []*EventsEntity
	for _, evt := range gs.Events {
		if evt.Turn == turn {
			result = append(result, evt)
		}
	}
	return result
}

// HasGameSetting checks if a specific game setting flag is enabled.
// Use with data.GameSetting* constants.
func (gs *GameStore) HasGameSetting(flag int) bool {
	return (int(gs.GameSettings) & flag) != 0
}

// UniverseSizeName returns the human-readable name for the universe size.
func (gs *GameStore) UniverseSizeName() string {
	names := []string{"Tiny", "Small", "Medium", "Large", "Huge"}
	if int(gs.UniverseSize) < len(names) {
		return names[gs.UniverseSize]
	}
	return "Unknown"
}

// DensityName returns the human-readable name for the planet density.
func (gs *GameStore) DensityName() string {
	names := []string{"Sparse", "Normal", "Dense", "Packed"}
	if int(gs.Density) < len(names) {
		return names[gs.Density]
	}
	return "Unknown"
}

// mergePlayerScores associates PlayerScoresBlock data with the appropriate player.
func (gs *GameStore) mergePlayerScores(psb *blocks.PlayerScoresBlock, source *FileSource) {
	player, ok := gs.Player(psb.PlayerID)
	if !ok {
		// Player not yet loaded - this can happen if PlayerScoresBlock comes
		// before PlayerBlock in the file. We'll ignore it for now as the score
		// can still be accessed by iterating through blocks later.
		return
	}

	// Only update if we don't have score data or this is from a newer turn
	if player.StoredScore == nil || psb.Turn > player.StoredScore.Turn {
		player.StoredScore = &StoredScore{
			Score:        psb.Score,
			Resources:    psb.Resources,
			Planets:      psb.Planets,
			Starbases:    psb.Starbases,
			UnarmedShips: psb.UnarmedShips,
			EscortShips:  psb.EscortShips,
			CapitalShips: psb.CapitalShips,
			TechLevels:   psb.TechLevels,
			Rank:         psb.Rank,
			Turn:         psb.Turn,
		}
	}
}

// PlayerScore returns the stored score for a player, or nil if not available.
// This returns the authoritative score as calculated by the game itself.
func (gs *GameStore) PlayerScore(playerNumber int) *StoredScore {
	player, ok := gs.Player(playerNumber)
	if !ok {
		return nil
	}
	return player.StoredScore
}
