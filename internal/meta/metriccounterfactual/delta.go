package metriccounterfactual

import artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"

func ComputeDelta(before, after State) Delta {
	delta := Delta{
		DirectFolders:    after.Totals.DirectFolders - before.Totals.DirectFolders,
		DirectFiles:      after.Totals.DirectFiles - before.Totals.DirectFiles,
		RecursiveFolders: after.Totals.RecursiveFolders - before.Totals.RecursiveFolders,
		RecursiveFiles:   after.Totals.RecursiveFiles - before.Totals.RecursiveFiles,
		GoFiles:          after.Totals.GoFiles - before.Totals.GoFiles,
		GoooFiles:        after.Totals.GoooFiles - before.Totals.GoooFiles,
		GoLines:          after.Totals.GoLines - before.Totals.GoLines,
		GoooLines:        after.Totals.GoooLines - before.Totals.GoooLines,
	}
	beforeFiles := make(map[string]FileMetric, len(before.Files))
	afterFiles := make(map[string]FileMetric, len(after.Files))
	for _, file := range before.Files {
		beforeFiles[file.Path] = file
	}
	for _, file := range after.Files {
		afterFiles[file.Path] = file
		previous, found := beforeFiles[file.Path]
		if !found || !artifact.Equal(previous, file) {
			delta.ChangedFiles++
		}
	}
	for file := range beforeFiles {
		if _, found := afterFiles[file]; !found {
			delta.ChangedFiles++
		}
	}
	beforeDirectories := make(map[string]DirectoryMetric, len(before.Directories))
	afterDirectories := make(map[string]DirectoryMetric, len(after.Directories))
	for _, directory := range before.Directories {
		beforeDirectories[directory.Path] = directory
	}
	for _, directory := range after.Directories {
		afterDirectories[directory.Path] = directory
		previous, found := beforeDirectories[directory.Path]
		if !found || !artifact.Equal(previous, directory) {
			delta.ChangedDirectories++
		}
	}
	for directory := range beforeDirectories {
		if _, found := afterDirectories[directory]; !found {
			delta.ChangedDirectories++
		}
	}
	return delta
}
