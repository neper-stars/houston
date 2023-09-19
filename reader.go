package houston

import (
	"os"
	"io"
)

func ReadRawFile(fName string, fileData *FileData) error {
	f, err := os.Open(fName)
	if err != nil {
		return err
	}
	*fileData, err = io.ReadAll(f)
	if err != nil {
		return err
	}
	return nil
}
