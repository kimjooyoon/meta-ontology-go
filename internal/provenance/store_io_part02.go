package provenance

import (
	"os"
)

func closeFile(file *os.File, point storageFaultPoint) error {
	if err := failStorageOperation(point); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
