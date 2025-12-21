package blocks

// Block is the interface that all block types must implement
type Block interface {
	BlockTypeID() BlockTypeID
	BlockSize() BlockSize
	BlockData() BlockData
	DecryptedData() DecryptedData
}

// BlockTypeID represents the type identifier of a block (6 bits)
type BlockTypeID uint16

// BlockSize represents the size of a block's data (10 bits)
type BlockSize uint16

// BlockData represents the raw encrypted block data
type BlockData []byte

// DecryptedData represents the decrypted block data
type DecryptedData []byte
