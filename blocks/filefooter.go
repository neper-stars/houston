package blocks

import (
	"github.com/neper-stars/houston/encoding"
)

// FileFooterBlock represents the end-of-file marker block (Type 0)
// It contains a checksum for the file, except for .h# files which have no checksum
type FileFooterBlock struct {
	GenericBlock
	Checksum uint16
}

// NewFileFooterBlock creates a FileFooterBlock from a GenericBlock
func NewFileFooterBlock(b GenericBlock) *FileFooterBlock {
	fb := &FileFooterBlock{
		GenericBlock: b,
	}

	// .h# files have no checksum in their footer (size 0)
	// Other files have a 2-byte checksum
	if len(b.Data) >= 2 {
		fb.Checksum = encoding.Read16(b.Data, 0)
	}

	return fb
}

// HasChecksum returns true if the footer contains a checksum
func (fb *FileFooterBlock) HasChecksum() bool {
	return len(fb.Data) >= 2
}
