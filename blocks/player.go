package blocks

import (
	"errors"

	"github.com/neper-stars/houston/encoding"
)

var ErrInvalidPlayerBlock = errors.New("invalid player block")

// HashedPass is a 4-byte clump
type HashedPass []byte

func (h HashedPass) Uint32() uint32 {
	return encoding.Read32(h, 0)
}

type PlayerBlock struct {
	GenericBlock
	Valid               bool
	PlayerNumber        int
	NamePlural          string
	NameSingular        string
	ShipDesignCount     int
	Planets             int
	Fleets              int
	StarbaseDesignCount int
	Logo                int
	FullDataFlag        bool
	Byte7               byte
	FullDataBytes       []byte
	PlayerRelations     []byte
	// Race settings (from FullDataBytes when FullDataFlag is true)
	GrowthRate int // Max population growth rate percentage (typically 1-20)
}

// HashedPass returns the hashed password from inside
// the PlayerBlock
// This can be used has a source for the GuessRacePassword function
// by doing hashed.Uint32()
func (p *PlayerBlock) HashedPass() HashedPass {
	// the hashed password is stored at offset 12
	// of the decrypted data, and it is 4 bytes long
	return []byte(p.DecryptedData()[12:16])
}

func (p *PlayerBlock) decode() error {
	// Ensure that there is enough data to decode
	if len(p.Decrypted) < 8 {
		return errors.New("unexpected player data size")
	}

	p.PlayerNumber = int(p.Decrypted[0])
	p.ShipDesignCount = int(p.Decrypted[1])
	p.Planets = int(p.Decrypted[2]) + (int(p.Decrypted[3]) & 0x03 << 8)

	if int(p.Decrypted[3])&0xFC != 0 {
		return errors.New("unexpected player values")
	}

	p.Fleets = int(p.Decrypted[4]) + (int(p.Decrypted[5]) & 0x03 << 8)
	p.StarbaseDesignCount = int(p.Decrypted[5]) >> 4

	if int(p.Decrypted[5])&0x0C != 0 {
		return errors.New("unexpected player values")
	}

	p.Logo = int(p.Decrypted[6]) >> 3
	p.FullDataFlag = (int(p.Decrypted[6]) & 0x04) != 0

	if int(p.Decrypted[6])&0x03 != 3 {
		return errors.New("unexpected player values")
	}

	p.Byte7 = p.Decrypted[7]

	index := 8
	if p.FullDataFlag {
		p.FullDataBytes = make([]byte, 0x68)
		copy(p.FullDataBytes, p.Decrypted[8:8+0x68])
		// Decode race settings from FullDataBytes
		p.GrowthRate = int(p.FullDataBytes[17])
		index = 0x70
		playerRelationsLength := int(p.Decrypted[index]) & 0xFF
		p.PlayerRelations = make([]byte, playerRelationsLength)
		copy(p.PlayerRelations, p.Decrypted[index+1:index+1+playerRelationsLength])
		index += 1 + playerRelationsLength
	}

	// Decode the singular name
	singularNameLength := int(p.Decrypted[index]) & 0xFF
	nameBytesSingular := make([]byte, singularNameLength+1)
	copy(nameBytesSingular, p.Decrypted[index:index+singularNameLength+1])

	var err error
	p.NameSingular, err = encoding.DecodeStarsString(nameBytesSingular)
	if err != nil {
		return err
	}

	index += singularNameLength + 1

	// Decode plural name (if exist)
	pluralNameLength := int(p.Decrypted[index]) & 0xFF
	nameBytesPlural := make([]byte, pluralNameLength+1)
	copy(nameBytesPlural, p.Decrypted[index:index+pluralNameLength+1])

	p.NamePlural, err = encoding.DecodeStarsString(nameBytesPlural)
	if err != nil {
		return err
	}

	index += pluralNameLength + 1
	// If no plural name skip another byte because of 16-bit alignment
	if pluralNameLength == 0 {
		index++
	}
	return nil
}

func NewPlayerBlock(b GenericBlock) (*PlayerBlock, error) {
	p := &PlayerBlock{
		GenericBlock: b,
	}
	if len(b.DecryptedData()) >= 16 {
		p.Valid = true
	}

	if err := p.decode(); err != nil {
		return nil, err
	}

	return p, nil
}

// Stored relation values in M files (different from order file encoding)
const (
	StoredRelationNeutral = 0
	StoredRelationFriend  = 1
	StoredRelationEnemy   = 2
)

// GetRelationTo returns this player's diplomatic relation to another player.
// Returns the stored relation value (0=Neutral, 1=Friend, 2=Enemy).
// Relations beyond the stored array length default to Neutral (0).
// Returns -1 only if playerIndex is negative.
func (p *PlayerBlock) GetRelationTo(playerIndex int) int {
	if playerIndex < 0 {
		return -1
	}
	if len(p.PlayerRelations) == 0 || playerIndex >= len(p.PlayerRelations) {
		// Relations not explicitly stored default to Neutral
		return StoredRelationNeutral
	}
	return int(p.PlayerRelations[playerIndex])
}

// GetRelationName returns the human-readable name for a stored relation value
func GetRelationName(storedRelation int) string {
	switch storedRelation {
	case StoredRelationNeutral:
		return "Neutral"
	case StoredRelationFriend:
		return "Friend"
	case StoredRelationEnemy:
		return "Enemy"
	default:
		return "Unknown"
	}
}
