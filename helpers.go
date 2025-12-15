package houston

import (
	"errors"
)

// Order represents a file submitted by a player with
// his turn orders
// Order file can be just a save file or a
// full "submitted" file.
// stars will only process submitted files
type Order struct {
	fd     *FileData
	Header *FileHeader
}

func NewOrder(fd *FileData) (*Order, error) {
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
	var fd FileData
	if err := ReadRawFile(fName, &fd); err != nil {
		return nil, err
	}

	return NewOrder(&fd)
}

func NewOrderFromBytes(data []byte) (*Order, error) {
	var fd FileData
	fd = data
	return NewOrder(&fd)
}

func (o *Order) TurnSubmitted() bool {
	if o.Header == nil {
		panic("cannot test turn submitted on a nil Header")
	}

	return o.Header.TurnSubmitted()
}
