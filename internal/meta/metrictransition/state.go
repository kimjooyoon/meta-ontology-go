package metrictransition

import (
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
)

func buildState(inputs inputSet) (RepositoryState, error) {
	source, err := buildPlane("LOGICAL_SOURCE", inputs.report.Directories)
	if err != nil {
		return RepositoryState{}, err
	}
	storageName, storageDirectories := "PHYSICAL_STORAGE", inputs.report.StorageDirectories
	if len(storageDirectories) == 0 {
		storageName, storageDirectories = "LOGICAL_SOURCE_FALLBACK", inputs.report.Directories
	}
	storage, err := buildPlane(storageName, storageDirectories)
	if err != nil {
		return RepositoryState{}, err
	}
	policy, err := rootPolicy(inputs.metrics, storage.Root)
	if err != nil {
		return RepositoryState{}, err
	}
	files := make([]LanguageFile, 0)
	for _, file := range inputs.report.Files {
		language := string(file.Language)
		if language == "go" || language == "gooo" {
			files = append(files, LanguageFile{Path: file.Path, Language: language, Lines: file.Lines})
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	state := RepositoryState{Schema: StateSchema, Repository: inputs.report.Repository, CommitSHA: inputs.report.CommitSHA, RootPolicy: policy, Source: source, Storage: storage, LanguageFiles: files}
	return sealState(state)
}

func buildPlane(name string, metrics []linecaps.DirectoryMetric) (MetricPlane, error) {
	directories := make([]DirectoryState, 0, len(metrics))
	var root Counts
	foundRoot := false
	for _, metric := range metrics {
		counts := countsOf(metric)
		directories = append(directories, DirectoryState{Path: metric.Path, SubjectKind: string(metric.SubjectKind), Counts: counts})
		if metric.Path == "." {
			root, foundRoot = counts, true
		}
	}
	if !foundRoot {
		return MetricPlane{}, fmt.Errorf("metric plane %s has no project root", name)
	}
	sort.Slice(directories, func(i, j int) bool { return directories[i].Path < directories[j].Path })
	return MetricPlane{Name: name, Root: root, Directories: directories}, nil
}

func countsOf(metric linecaps.DirectoryMetric) Counts {
	return Counts{DirectFolders: metric.DirectFolders, DirectFiles: metric.DirectFiles, RecursiveFolders: metric.RecursiveFolders, RecursiveFiles: metric.RecursiveFiles, GoFiles: metric.GoFiles, GoooFiles: metric.GoooFiles, GoLines: metric.GoLines, GoooLines: metric.GoooLines}
}
