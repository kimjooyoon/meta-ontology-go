package extractor

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"strings"
)

// Snapshot source bytes once for both variants. External module-cache inputs
// remain unbound; this is not an immutable transitive dependency attestation.
func callbackObservationModuleSources(ctx context.Context, root, logical string, baseline map[string][]byte) (map[string][]byte, error) {
	if !fs.ValidPath(logical) {
		return nil, fmt.Errorf("callback observation source path is not relative")
	}
	files := map[string][]byte{}
	tree := os.DirFS(root)
	err := fs.WalkDir(tree, ".", func(name string, entry fs.DirEntry, walkError error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if walkError != nil {
			return walkError
		}
		if name == ".git" {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("callback observation refuses non-regular module input %s", name)
		}
		raw, err := fs.ReadFile(tree, name)
		if err != nil {
			return err
		}
		files[name] = raw
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files["go.mod"]) == 0 {
		return nil, fmt.Errorf("callback observation module snapshot has no go.mod")
	}
	prefix := ""
	if directory := path.Dir(logical); directory != "." {
		prefix = directory + "/"
	}
	observed := map[string][]byte{}
	for name, raw := range files {
		if strings.HasPrefix(name, prefix) {
			observed[strings.TrimPrefix(name, prefix)] = raw
		}
	}
	if !maps.EqualFunc(observed, baseline, bytes.Equal) {
		return nil, fmt.Errorf("callback observation package changed during module snapshot")
	}
	return files, nil
}
