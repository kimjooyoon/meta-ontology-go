package bodypathstream

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFileInputMissingPathRetainsFilesystemCause(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent")
	_, err := readFileInput(path, 128)
	var cause *os.PathError
	if !errors.Is(err, os.ErrNotExist) || !errors.As(err, &cause) || cause.Path != path {
		t.Fatal("missing input lost its filesystem cause", err)
	}
}
