package metriccounterfactual

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

func Measure(root string) (State, error) {
	directories := []string{"."}
	files := make([]FileMetric, 0)
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, current)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink is not measurable: %s", relative)
		}
		if entry.IsDir() {
			if relative != "." {
				directories = append(directories, relative)
			}
			return nil
		}
		content, err := os.ReadFile(current)
		if err != nil {
			return err
		}
		files = append(files, FileMetric{
			Path: relative, Language: languageForPath(relative),
			Lines: artifact.CountLines(content), Bytes: len(content),
			Digest: artifact.ContentDigest(content),
		})
		return nil
	})
	if err != nil {
		return State{}, err
	}
	sort.Strings(directories)
	sort.Slice(files, func(left, right int) bool { return files[left].Path < files[right].Path })
	metrics, totals := aggregateMetrics(directories, files)
	return SealState(State{
		Schema: StateSchema, Files: files, Directories: metrics,
		Totals: totals, RootPolicy: ProjectRootPolicy(),
	})
}
