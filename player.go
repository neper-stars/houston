package houston

import (
	"errors"
)

var ErrInvalidPlayerBlock = errors.New("invalid player block")

// HashedPass is a 4-byte clump
type HashedPass []byte

func (h HashedPass) Uint32() uint32 {
	return read32(h, 0)
}

type PlayerBlock struct {
	GenericBlock
	Valid bool
}

// HashedPass returns the hashed password from inside
// the PlayerBlock
// This can be used has a source for the GuessRacePassword function
// by doing hashed.Uint32()
func (p PlayerBlock) HashedPass() HashedPass {
	// the hashed password is stored at offset 12
	// of the decrypted data, and it is 4 bytes long
	return []byte(p.DecryptedData()[12:16])
}

func NewPlayerBlock(b GenericBlock) *PlayerBlock {
	p := PlayerBlock{
		GenericBlock: b,
	}
	if len(b.DecryptedData()) >= 16 {
		p.Valid = true
	}
	return &p
}
