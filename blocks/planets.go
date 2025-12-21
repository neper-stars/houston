package blocks

import (
	"encoding/binary"
	"fmt"

	"github.com/neper-stars/houston/data"
)

// Planet is a struct holding the planet
// position (x, y) and the NameID
type Planet struct {
	ID        int
	DisplayId int
	NameID    uint32
	Name      string
	Y         uint32
	X         uint32
}

// PlanetsBlock is an implementation of Block
// that provides accessors for the game rules
// and the universe content (list of planets)
type PlanetsBlock struct {
	GenericBlock
	Valid            bool
	UniverseSize     uint16
	Density          uint16
	PlayerCount      uint16
	PlanetCount      uint16
	StartingDistance uint32
	GameSettings     uint16
	GameName         string
	Planets          []Planet
}

func (p *PlanetsBlock) GetPlanetCount() int {
	if int(p.PlanetCount) != len(p.Planets) {
		fmt.Println("planet count mismatch")
	}
	return len(p.Planets)
}

func NewPlanetsBlock(b GenericBlock) *PlanetsBlock {
	block := PlanetsBlock{
		GenericBlock: b,
	}
	data := b.DecryptedData()
	if len(data) < 64 {
		// not enough data to form a specialized block
		return &block
	}

	block.Valid = true
	block.UniverseSize = binary.LittleEndian.Uint16(data[4:6])
	block.Density = binary.LittleEndian.Uint16(data[6:8])
	block.PlayerCount = binary.LittleEndian.Uint16(data[8:10])
	block.PlanetCount = binary.LittleEndian.Uint16(data[10:12])
	block.StartingDistance = binary.LittleEndian.Uint32(data[12:16])
	/*
	   Determined bits of game settings:

	   Max Minerals       - 0000000000000001
	   Slow tech advances - 0000000000000010
	   Accel. BBS play    - 0000000000100000
	   No random events   - 0000000010000000
	   Computer Alliances - 0000000000010000
	   Public scores      - 0000000001000000
	   Galaxy Clumping    - 0000000100000000
	   Single player (?)  - 0000000000000100

	   What is bit 4?       0000000000001000
	*/
	block.GameSettings = binary.LittleEndian.Uint16(data[16:18])
	// TODO bytes 18-31
	block.GameName = string(data[32:64])

	return &block
}

func (p *PlanetsBlock) ParsePlanetsData(d []byte) {
	x := uint32(1000)

	for i := 0; i < int(p.PlanetCount); i++ {
		planetData := binary.LittleEndian.Uint32(d[i*4 : (i+1)*4])

		nameID := planetData >> 22      // First 10 bits
		y := (planetData >> 10) & 0xFFF // Middle 12 bits
		xOffset := planetData & 0x3FF   // Last 10 bits
		x += xOffset

		planet := Planet{
			ID:        i,
			DisplayId: i + 1,
			NameID:    nameID,
			Name:      data.PlanetNames[nameID],
			Y:         y,
			X:         x,
		}

		p.Planets = append(p.Planets, planet)
	}
}

func (p *PlanetsBlock) GetPlanets() []Planet {
	return p.Planets
}
