package blocks

import (
	"github.com/neper-stars/houston/encoding"
)

// DesignSlot represents a component slot in a ship/starbase design
type DesignSlot struct {
	Category uint16 // Tech category bitmask
	ItemId   int    // Item within category
	Count    int    // Quantity installed
}

// DesignBlock represents a ship or starbase design (Type 26)
type DesignBlock struct {
	GenericBlock

	// Design metadata
	IsFullDesign  bool // True if this is a full design with components
	IsTransferred bool // True if design received from another player
	IsStarbase    bool // True if starbase, false if ship
	DesignNumber  int  // 0-15

	// Hull and appearance
	HullId int // Hull type ID
	Pic    int // Picture ID

	// Ship/starbase properties (calculated for full designs)
	Mass         int   // Total mass
	FuelCapacity int   // Total fuel capacity
	Armor        int   // Armor value (0-65535)
	SlotCount    int   // Number of component slots
	TurnDesigned int   // Turn when design was created
	TotalBuilt   int64 // Total ships built of this design
	TotalRemaining int64 // Ships still in existence

	// Components (only for full designs)
	Slots []DesignSlot

	// Design name
	Name      string
	NameBytes []byte

	// Bug detection flags
	ColonizerModuleBug bool
	SpaceDocBug        bool
}

// NewDesignBlock creates a DesignBlock from a GenericBlock
func NewDesignBlock(b GenericBlock) (*DesignBlock, error) {
	db := &DesignBlock{
		GenericBlock: b,
		Slots:        make([]DesignSlot, 0),
	}
	if err := db.decode(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DesignBlock) decode() error {
	data := db.Decrypted
	if len(data) < 6 {
		return nil
	}

	// Byte 0: First control byte
	// Bits 0-1: Must be 0b11 (0x03)
	// Bit 2: isFullDesign flag
	// Bits 3-7: Must be 0
	db.IsFullDesign = (data[0] & 0x04) == 0x04

	// Byte 1: Second control byte
	// Bit 0: Must be 1
	// Bit 1: Must be 0
	// Bits 2-5: designNumber (0-15)
	// Bit 6: isStarbase flag
	// Bit 7: isTransferred flag
	db.IsTransferred = (data[1] & 0x80) == 0x80
	db.IsStarbase = (data[1] & 0x40) == 0x40
	db.DesignNumber = int((data[1] & 0x3C) >> 2)

	// Byte 2: Hull ID
	db.HullId = int(data[2] & 0xFF)

	// Byte 3: Picture ID
	db.Pic = int(data[3] & 0xFF)

	index := 4

	if db.IsFullDesign {
		if len(data) < 17 {
			return nil
		}

		// Bytes 4-5: Armor (16-bit)
		db.Armor = int(encoding.Read16(data, 4))

		// Byte 6: Slot count
		db.SlotCount = int(data[6] & 0xFF)

		// Bytes 7-8: Turn designed (16-bit)
		db.TurnDesigned = int(encoding.Read16(data, 7))

		// Bytes 9-12: Total built (32-bit)
		db.TotalBuilt = int64(encoding.Read32(data, 9))

		// Bytes 13-16: Total remaining (32-bit)
		db.TotalRemaining = int64(encoding.Read32(data, 13))

		index = 17

		// Component slots (4 bytes each)
		db.Mass = 0 // Will be calculated from components
		for i := 0; i < db.SlotCount; i++ {
			if index+4 > len(data) {
				break
			}

			slot := DesignSlot{
				Category: encoding.Read16(data, index),
				ItemId:   int(data[index+2] & 0xFF),
				Count:    int(data[index+3] & 0xFF),
			}
			db.Slots = append(db.Slots, slot)
			index += 4

			// Check for colonizer module bug
			if slot.ItemId == 0 && slot.Count == 0 && slot.Category == 4096 {
				db.ColonizerModuleBug = true
			}

			// Check for space dock bug
			if db.IsStarbase && db.HullId == 33 && slot.ItemId == 11 &&
				slot.Category == 8 && slot.Count >= 22 && db.Armor >= 49518 {
				db.SpaceDocBug = true
			}
		}
	} else {
		// Brief design: no components
		// Bytes 4-5: Mass (16-bit)
		db.Mass = int(encoding.Read16(data, 4))
		index = 6
	}

	// Design name (variable length Stars! encoded string)
	if index < len(data) {
		nameLen := int(data[index] & 0xFF)
		if index+1+nameLen <= len(data) {
			db.NameBytes = make([]byte, 1+nameLen)
			copy(db.NameBytes, data[index:index+1+nameLen])

			// Decode the name using Stars! string encoding
			decoded, err := encoding.DecodeStarsString(db.NameBytes)
			if err == nil {
				db.Name = decoded
			}
		}
	}

	return nil
}

// GetSlot returns the slot at the given index, or nil if out of range
func (db *DesignBlock) GetSlot(index int) *DesignSlot {
	if index >= 0 && index < len(db.Slots) {
		return &db.Slots[index]
	}
	return nil
}

// ShipCount returns the number of ships remaining of this design
func (db *DesignBlock) ShipCount() int64 {
	return db.TotalRemaining
}

// DesignChangeBlock represents a design modification or deletion (Type 27)
// It has 2 extra bytes at the beginning compared to DesignBlock
type DesignChangeBlock struct {
	GenericBlock

	// If IsDelete is true, this block represents a design deletion
	IsDelete       bool
	DesignToDelete int  // Design number to delete (0-15)
	IsStarbase     bool // True if deleting a starbase design

	// If IsDelete is false, the design data follows (same as DesignBlock)
	Design *DesignBlock
}

// NewDesignChangeBlock creates a DesignChangeBlock from a GenericBlock
func NewDesignChangeBlock(b GenericBlock) (*DesignChangeBlock, error) {
	dcb := &DesignChangeBlock{
		GenericBlock: b,
	}
	if err := dcb.decode(); err != nil {
		return nil, err
	}
	return dcb, nil
}

func (dcb *DesignChangeBlock) decode() error {
	data := dcb.Decrypted
	if len(data) < 2 {
		return nil
	}

	// Check if this is a deletion (first nibble is 0)
	if data[0]%16 == 0 {
		dcb.IsDelete = true
		dcb.DesignToDelete = int(data[1] % 16)
		dcb.IsStarbase = (data[1]>>4)%2 == 1
		return nil
	}

	// Otherwise, this is a design change - skip first 2 bytes and parse as DesignBlock
	if len(data) < 4 {
		return nil
	}

	// Create a modified data slice without the first 2 bytes
	designData := make([]byte, len(data)-2)
	copy(designData, data[2:])

	// The bit 0 of byte 1 must be set for the DesignBlock decoder
	// Some files have this bit unset, which is a known issue
	if (designData[1] & 0x01) != 0x01 {
		designData[1] |= 0x01
	}

	// Create a GenericBlock with the modified data
	designGenericBlock := GenericBlock{
		Type:      DesignBlockType,
		Size:      BlockSize(len(designData)),
		Data:      BlockData(designData),
		Decrypted: DecryptedData(designData),
	}

	var err error
	dcb.Design, err = NewDesignBlock(designGenericBlock)
	return err
}
