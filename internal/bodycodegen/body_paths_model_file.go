package bodycodegen

import (
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/fileopen"
)

// Read the single artifact before schema dispatch. Older metadata formats keep
// their smaller size bound in the dispatcher after the schema is known.
func readStructuralModelBytes(name string) ([]byte, error) {
	f, err := fileopen.ReadOnly(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > 512<<10 {
		return nil, fmt.Errorf("bounded regular structural metadata required")
	}
	raw, err := io.ReadAll(io.LimitReader(f, (512<<10)+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > 512<<10 {
		return nil, fmt.Errorf("structural metadata grew beyond bound")
	}
	return raw, nil
}
