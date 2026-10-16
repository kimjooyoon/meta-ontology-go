package candidatedelivery

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"slices"
)

// ReadArchiveMembers reads caller-selected fixed names without extracting paths.
func ReadArchiveMembers(raw []byte, wanted []string) (map[string][]byte, error) {
	if len(raw) > MaxArchiveBytes {
		return nil, fmt.Errorf("delivery archive oversized")
	}
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, err
	}
	if len(z.File) > 128 {
		return nil, fmt.Errorf("delivery member population oversized")
	}
	files := make(map[string][]byte, len(wanted))
	for _, entry := range z.File {
		if !slices.Contains(wanted, entry.Name) {
			continue
		}
		if _, found := files[entry.Name]; found {
			return nil, fmt.Errorf("duplicate delivery member")
		}
		stream, err := entry.Open()
		if err != nil {
			return nil, err
		}
		data, readErr := io.ReadAll(io.LimitReader(stream, (8<<20)+1))
		closeErr := stream.Close()
		if readErr != nil || closeErr != nil || len(data) > 8<<20 {
			return nil, fmt.Errorf("delivery member unavailable or oversized")
		}
		files[entry.Name] = data
	}
	return files, nil
}
