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
		observed = append(observed, FileObservation{
			Path: path, GoooLines: sourceLines(raw), SourceDigest: digestBytes(raw),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return observed, nil
}
