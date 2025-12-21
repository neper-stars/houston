package parser

import (
	"io"
	"os"
)

// ReadRawFile reads an entire file into a FileData struct
func ReadRawFile(fName string, fileData *FileData) error {
	f, err := os.Open(fName)
	if err != nil {
		return err
	}
	defer f.Close()
	*fileData, err = io.ReadAll(f)
	if err != nil {
		return err
	}
	return nil
}
