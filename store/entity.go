package store

import "github.com/neper-stars/houston/blocks"

// EntityType identifies the kind of game entity.
type EntityType int

const (
	EntityTypeFleet EntityType = iota
	EntityTypePlanet
	EntityTypeDesign
	EntityTypeStarbaseDesign
	EntityTypePlayer
	EntityTypeObject
	EntityTypeBattlePlan
	EntityTypeProductionQueue
	EntityTypeMessage
	EntityTypeWaypoint
)

// String returns a human-readable entity type name.
func (t EntityType) String() string {
	switch t {
	case EntityTypeFleet:
		return "Fleet"
	case EntityTypePlanet:
		return "Planet"
	case EntityTypeDesign:
		return "Design"
	case EntityTypeStarbaseDesign:
		return "StarbaseDesign"
	case EntityTypePlayer:
		return "Player"
	case EntityTypeObject:
		return "Object"
	case EntityTypeBattlePlan:
		return "BattlePlan"
	case EntityTypeProductionQueue:
		return "ProductionQueue"
	case EntityTypeMessage:
		return "Message"
	case EntityTypeWaypoint:
		return "Waypoint"
	default:
		return "Unknown"
	}
}

// EntityKey is a unique identifier for an entity.
type EntityKey struct {
	Type   EntityType
	Owner  int // Player index (0-15), -1 for unowned (e.g., planets)
	Number int // Entity number within owner/type
}

// EntityMeta contains common metadata for all entities.
type EntityMeta struct {
	Key        EntityKey
	BestSource *FileSource   // Source of best data
	AllSources []*FileSource // All sources that mentioned this entity
	Quality    DataQuality
	Turn       uint16 // Turn when last updated
	Dirty      bool   // Modified since load
}

// AddSource adds a source to AllSources if not already present.
func (m *EntityMeta) AddSource(source *FileSource) {
	for _, s := range m.AllSources {
		if s.ID == source.ID {
			return
		}
	}
	m.AllSources = append(m.AllSources, source)
}

// Entity is the interface all entity types implement.
type Entity interface {
	Meta() *EntityMeta
	RawBlocks() []blocks.Block
	SetDirty()
}

// EntityCollection manages a set of entities of the same type.
type EntityCollection[T Entity] struct {
	byKey   map[EntityKey]T
	byOwner map[int][]T
	all     []T
}

// NewEntityCollection creates a new empty collection.
func NewEntityCollection[T Entity]() *EntityCollection[T] {
	return &EntityCollection[T]{
		byKey:   make(map[EntityKey]T),
		byOwner: make(map[int][]T),
	}
}

// Get retrieves an entity by its key.
func (c *EntityCollection[T]) Get(key EntityKey) (T, bool) {
	entity, ok := c.byKey[key]
	return entity, ok
}

// GetByOwnerAndNumber retrieves an entity by owner and number.
func (c *EntityCollection[T]) GetByOwnerAndNumber(entityType EntityType, owner, number int) (T, bool) {
	return c.Get(EntityKey{Type: entityType, Owner: owner, Number: number})
}

// Add adds or updates an entity in the collection.
func (c *EntityCollection[T]) Add(entity T) {
	key := entity.Meta().Key
	existing, exists := c.byKey[key]

	if exists {
		// Update in place - find and replace in slices
		for i, e := range c.all {
			if e.Meta().Key == key {
				c.all[i] = entity
				break
			}
		}
		owner := key.Owner
		for i, e := range c.byOwner[owner] {
			if e.Meta().Key == key {
				c.byOwner[owner][i] = entity
				break
			}
		}
	} else {
		c.all = append(c.all, entity)
		c.byOwner[key.Owner] = append(c.byOwner[key.Owner], entity)
	}

	c.byKey[key] = entity
	_ = existing // silence unused variable warning
}

// ByOwner returns all entities owned by a specific player.
func (c *EntityCollection[T]) ByOwner(owner int) []T {
	return c.byOwner[owner]
}

// All returns all entities in the collection.
func (c *EntityCollection[T]) All() []T {
	return c.all
}

// Count returns the number of entities.
func (c *EntityCollection[T]) Count() int {
	return len(c.all)
}

// DirtyEntities returns all entities that have been modified.
func (c *EntityCollection[T]) DirtyEntities() []T {
	var dirty []T
	for _, entity := range c.all {
		if entity.Meta().Dirty {
			dirty = append(dirty, entity)
		}
	}
	return dirty
}

// ResetDirtyFlags clears the dirty flag on all entities.
func (c *EntityCollection[T]) ResetDirtyFlags() {
	for _, entity := range c.all {
		entity.Meta().Dirty = false
	}
}
