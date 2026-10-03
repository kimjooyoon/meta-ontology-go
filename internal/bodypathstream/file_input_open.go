package bodypathstream

import (
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/fileopen"
)

func openFileInput(path string) (*os.File, error) {
	return fileopen.ReadOnly(path)
}
