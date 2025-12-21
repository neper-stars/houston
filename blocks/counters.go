package blocks

import (
	"github.com/neper-stars/houston/encoding"
)

// CountersBlock represents game object counters (Type 32)
// Contains the count of planets and fleets in the game
type CountersBlock struct {
	GenericBlock

	PlanetCount int // Total number of planets
	FleetCount  int // Total number of fleets
}

// NewCountersBlock creates a CountersBlock from a GenericBlock
func NewCountersBlock(b GenericBlock) *CountersBlock {
	cb := &CountersBlock{
		GenericBlock: b,
	}
	cb.decode()
	return cb
}

func (cb *CountersBlock) decode() {
	data := cb.Decrypted
	if len(data) < 4 {
		return
	}

	cb.PlanetCount = int(encoding.Read16(data, 0))
	cb.FleetCount = int(encoding.Read16(data, 2))
}
