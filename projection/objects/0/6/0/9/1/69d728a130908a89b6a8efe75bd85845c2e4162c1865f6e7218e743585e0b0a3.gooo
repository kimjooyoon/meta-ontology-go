package main

import (
	"path/filepath"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func collectMetricTargets(opts options) ([]string, error) {
	report, err := linecaps.AnalyzeLineMetrics(opts.root)
	if err != nil {
		return nil, err
	}
	policy := sourcepolicy.Default()
	policy.MaxFileLines = opts.maxLines
	policy.MaxDirectDirectoryIn = opts.maxEntries
	meta, err := linecaps.EvaluateLineMetricIndicators(report, policy)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0)
	for _, indicator := range meta.Failed() {
		if indicator.Operation != sourcepolicy.OperationSplitGo && indicator.Operation != sourcepolicy.OperationSplitGooo {
			continue
		}
		paths = append(paths, filepath.Clean(filepath.Join(opts.root, filepath.FromSlash(indicator.Subject))))
	}
	sort.Strings(paths)
	return dedupe(paths), nil
}
