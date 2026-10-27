package replay

import (
	"io/fs"
	"strings"
)

func Observe(repository fs.FS) ([]FileObservation, error) {
	observed := []FileObservation{}
	err := fs.WalkDir(repository, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".gooo") {
			return nil
		}
		raw, err := fs.ReadFile(repository, path)
		if err != nil {
			return err
		}
		observed = append(observed, ObserveSource(path, raw))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return observed, nil
}

// ObserveSource describes exact supplied bytes. The caller owns the physical or
// proposed address; this observation does not assert that a file was installed.
func ObserveSource(path string, raw []byte) FileObservation {
	return FileObservation{Path: path, GoooLines: sourceLines(raw), SourceDigest: digestBytes(raw)}
}
