package workspaceexecution

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func applyWorkspaceRenameFiles(manifest *packageruntime.Manifest,
	changes map[string]map[string][]WorkspaceRenameEdit) ([]WorkspaceRenameFile, error) {
	var files []WorkspaceRenameFile
	for i := range manifest.Packages {
		pkg := &manifest.Packages[i]
		for j := range pkg.Sources {
			source := &pkg.Sources[j]
			edits := changes[pkg.Path][source.Filename]
			if len(edits) == 0 {
				continue
			}
			before := source.Content
			after, err := applyWorkspaceSourceEdits(before, edits)
			if err != nil {
				return nil, fmt.Errorf("package %s source %s: %w", pkg.Path, source.Filename, err)
			}
			source.Content = after
			files = append(files, WorkspaceRenameFile{pkg.Path, source.Filename, sourceSHA256([]byte(before)),
				sourceSHA256([]byte(after)), edits, after})
		}
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].PackagePath != files[j].PackagePath {
			return files[i].PackagePath < files[j].PackagePath
		}
		return files[i].Filename < files[j].Filename
	})
	return files, nil
}

func applyWorkspaceSourceEdits(source string, edits []WorkspaceRenameEdit) (string, error) {
	slices.SortFunc(edits, func(a, b WorkspaceRenameEdit) int { return a.Start - b.Start })
	var out strings.Builder
	previous := 0
	for _, edit := range edits {
		if edit.Start < previous || edit.End < edit.Start || edit.End > len(source) || source[edit.Start:edit.End] != edit.Before {
			return "", fmt.Errorf("overlapping or stale source edits")
		}
		out.WriteString(source[previous:edit.Start])
		out.WriteString(edit.After)
		previous = edit.End
	}
	out.WriteString(source[previous:])
	return out.String(), nil
}
