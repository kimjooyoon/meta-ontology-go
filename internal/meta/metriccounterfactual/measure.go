package metriccounterfactual

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
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
			Lines: countLines(content), Bytes: len(content), Digest: contentDigest(content),
		})
		return nil
	})
	if err != nil {
		return State{}, err
	}
	sort.Strings(directories)
	sort.Slice(files, func(left, right int) bool { return files[left].Path < files[right].Path })
	metrics := make([]DirectoryMetric, len(directories))
	index := make(map[string]int, len(directories))
	for position, directory := range directories {
		metrics[position].Path = directory
		index[directory] = position
	}
	for _, directory := range directories {
		if directory == "." {
			continue
		}
		parent := path.Dir(directory)
		metrics[index[parent]].DirectFolders++
		for _, ancestor := range ancestorPaths(parent) {
			metrics[index[ancestor]].RecursiveFolders++
		}
	}
	for _, file := range files {
		parent := path.Dir(file.Path)
		metrics[index[parent]].DirectFiles++
		for _, ancestor := range ancestorPaths(parent) {
			target := &metrics[index[ancestor]]
			target.RecursiveFiles++
			addLanguage(target, file)
		}
	}
	root := metrics[index["."]]
	return SealState(State{
		Schema: StateSchema, Files: files, Directories: metrics,
		Totals: Totals{
			DirectFolders: root.DirectFolders, DirectFiles: root.DirectFiles,
			RecursiveFolders: root.RecursiveFolders, RecursiveFiles: root.RecursiveFiles,
			GoFiles: root.GoFiles, GoooFiles: root.GoooFiles,
			GoLines: root.GoLines, GoooLines: root.GoooLines,
		},
		RootPolicy: ProjectRootPolicy(),
	})
}

func ancestorPaths(directory string) []string {
	result := []string{directory}
	for directory != "." {
		directory = path.Dir(directory)
		result = append(result, directory)
	}
	return result
}

func languageForPath(file string) string {
	switch {
	case strings.HasSuffix(file, ".gooo"):
		return "gooo"
	case strings.HasSuffix(file, ".go"):
		return "go"
	default:
		return "other"
	}
}

func addLanguage(directory *DirectoryMetric, file FileMetric) {
	if file.Language == "go" {
		directory.GoFiles++
		directory.GoLines += file.Lines
	}
	if file.Language == "gooo" {
		directory.GoooFiles++
		directory.GoooLines += file.Lines
	}
}
