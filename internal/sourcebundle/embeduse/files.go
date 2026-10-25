package embeduse

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

func (o *observer) prepareProjection(sources []sourcebundle.Source) error {
	if len(sources) == 0 {
		return nil
	}
	if _, err := sourcebundle.Digest(sources); err != nil {
		return err
	}
	for _, source := range sources {
		address, err := o.address(source.Filename)
		if err != nil {
			return err
		}
		backing, err := o.backing([]byte(source.Content))
		if err != nil {
			return err
		}
		if _, duplicate := o.projections[address]; duplicate {
			return fmt.Errorf("projected source address is duplicated")
		}
		o.projections[address] = backing
		o.projectedNames = append(o.projectedNames, source.Filename)
	}
	return nil
}

func (o *observer) backing(raw []byte) (string, error) {
	o.serial++
	name := filepath.Join(o.work, fmt.Sprintf("%d.txt", o.serial))
	if err := os.WriteFile(name, raw, 0o600); err != nil {
		return "", err
	}
	return name, nil
}

func (o *observer) read(name string) ([]byte, int, error) {
	address, err := o.address(name)
	if err != nil {
		return nil, 0, err
	}
	path, projected := o.projections[address]
	if !projected {
		path = address
	}
	raw, err := readBounded(path)
	if err != nil {
		return nil, 0, err
	}
	member := sourcebundle.Member{Filename: name, SHA256: fmt.Sprintf("%x", sha256.Sum256(raw)), ByteLength: len(raw)}
	if index, seen := o.inputs[name]; seen {
		if o.report.Files[index] != (File{Member: member, Projected: projected}) {
			return nil, 0, fmt.Errorf("source changed during Go embed observation: %s", name)
		}
		return raw, index, nil
	}
	o.bytes += len(raw)
	if o.bytes > sourcebundle.MaxTotalBytes || len(o.inputs) >= maxFiles {
		return nil, 0, fmt.Errorf("Go embed observation exceeds input budget")
	}
	index := len(o.report.Files)
	o.inputs[name] = index
	o.report.Files = append(o.report.Files, File{Member: member, Projected: projected})
	return raw, index, nil
}

func readBounded(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("Go embed input is not a regular file: %s", path)
	}
	raw, err := io.ReadAll(io.LimitReader(file, sourcebundle.MaxSourceBytes+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > sourcebundle.MaxSourceBytes {
		return nil, fmt.Errorf("Go embed input exceeds file byte limit")
	}
	return raw, nil
}

func (o *observer) recheckInputs() error {
	for _, file := range o.report.Files {
		if _, _, err := o.read(file.Filename); err != nil {
			return err
		}
	}
	return nil
}
