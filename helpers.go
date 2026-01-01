package houston

import (
	"errors"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

// Order represents a file submitted by a player with
// his turn orders
// Order file can be just a save file or a
// full "submitted" file.
// stars will only process submitted files
type Order struct {
	fd     *parser.FileData
	Header *blocks.FileHeader
}

func NewOrder(fd *parser.FileData) (*Order, error) {
	if fd == nil {
		return nil, errors.New("file data is required")
	}
	o := Order{
		fd: fd,
	}
	header, err := o.fd.FileHeader()
	if err != nil {
		return nil, err
	}

	o.Header = header

	return &o, nil
}

func NewOrderFromFile(fName string) (*Order, error) {
	var fd parser.FileData
	if err := parser.ReadRawFile(fName, &fd); err != nil {
		return nil, err
	}

	return NewOrder(&fd)
}

func NewOrderFromBytes(data []byte) (*Order, error) {
	var fd parser.FileData = data
	return NewOrder(&fd)
}

func (o *Order) TurnSubmitted() bool {
	if o.Header == nil {
		panic("cannot test turn submitted on a nil Header")
	}

	return o.Header.TurnSubmitted()
}

func (o *Order) Year() int {
	if o.Header == nil {
		panic("cannot work on a nil Header")
	}
	return o.Header.Year()
}
