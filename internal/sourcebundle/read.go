package sourcebundle

import (
	"fmt"
	"io"
	"io/fs"
)

// Read returns every named member or no source set. Names are relative to fsys.
// The caller owns selection and snapshot/freshness of a mutable filesystem.
func Read(fsys fs.FS, names []string) ([]Source, error) {
	if fsys == nil || len(names) == 0 || len(names) > MaxSources {
		return nil, fmt.Errorf("source bundle requires a filesystem and 1 to %d members", MaxSources)
	}
	seen := map[string]bool{}
	for _, name := range names {
		if err := checkName(name, seen); err != nil {
			return nil, err
		}
	}
	sources, total := make([]Source, 0, len(names)), 0
	for _, name := range names {
		raw, err := readMember(fsys, name, min(MaxSourceBytes, MaxTotalBytes-total))
		if err != nil {
			return nil, fmt.Errorf("source bundle member %q: %w", name, err)
		}
		total += len(raw)
		sources = append(sources, Source{Filename: name, Content: string(raw)})
	}
	return sources, nil
}

func readMember(fsys fs.FS, name string, limit int) ([]byte, error) {
	file, err := fsys.Open(name)
	if err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(io.LimitReader(file, int64(limit)+1))
	closeErr := file.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if len(raw) > limit {
		return nil, fmt.Errorf("source bundle byte limit exceeded")
	}
	return raw, nil
}
