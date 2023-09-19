package houston

type Block interface {
	BlockTypeID() BlockTypeID
	BlockSize() BlockSize
	BlockData() BlockData
	DecryptedData() DecryptedData
}

type BlockTypeID uint16
type BlockSize uint16
type BlockData []byte
type DecryptedData []byte

// GenericBlock is the most basic implementation
// of the Block interface
// It will have no more methods available
// If you have a generic block this means we have not implemented
// a more complex one for this block ID
// Certainly because we have not enough technical information
// to do so...
type GenericBlock struct {
	Type      BlockTypeID
	Size      BlockSize
	Data      BlockData
	Decrypted DecryptedData
}

func (b GenericBlock) BlockTypeID() BlockTypeID {
	return b.Type
}

func (b GenericBlock) BlockSize() BlockSize {
	return b.Size
}

func (b GenericBlock) BlockData() BlockData {
	return b.Data
}

func (b GenericBlock) DecryptedData() DecryptedData {
	return b.Decrypted
}
