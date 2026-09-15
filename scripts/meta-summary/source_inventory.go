package main

import (
	"encoding/json"
	"fmt"
	"sort"
)

func decodeSourceInventory(data []byte) (sourceInventory, error) {
	var document sourceMetricsEvidence
	if err := json.Unmarshal(data, &document); err != nil {
		return sourceInventory{}, fmt.Errorf("decode source metrics: %w", err)
	}
	if document.Meta.Schema != "gooo/indicator-report/v3" ||
		document.Meta.Policy.Schema != "gooo/source-policy/v1" || len(document.Directories) == 0 {
		return sourceInventory{}, fmt.Errorf("source metrics inventory foundation is incomplete")
	}
	root, ok := sourceRootDirectory(document.Directories)
	if !ok {
		return sourceInventory{}, fmt.Errorf("source metrics root directory is missing")
	}
	readme, err := sourceReadme(document.Meta.Indicators)
	if err != nil {
		return sourceInventory{}, err
	}
	candidates := sourceCandidates(document.Meta.Indicators)
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].MetricID != candidates[j].MetricID {
			return candidates[i].MetricID < candidates[j].MetricID
		}
		return candidates[i].Subject < candidates[j].Subject
	})
	return sourceInventory{
		RegularFiles: root.RecursiveFiles, DirectoriesIncludingRoot: len(document.Directories),
		DescendantDirectories: root.RecursiveFolders, GoFiles: root.GoFiles, GoLines: root.GoLines,
		GoooFiles: root.GoooFiles, GoooLines: root.GoooLines, RootReadme: readme,
		Thresholds: sourceThresholds{GoFile: document.Meta.Policy.MaxFileLines, GoooFile: document.Meta.Policy.MaxFileLines, Function: document.Meta.Policy.MaxFunctionLines},
		Candidates: candidates,
	}, nil
}

func sourceRootDirectory(directories []sourceDirectoryEvidence) (sourceDirectoryEvidence, bool) {
	for _, directory := range directories {
		if directory.Path == "." {
			return directory, true
		}
	}
	return sourceDirectoryEvidence{}, false
}
