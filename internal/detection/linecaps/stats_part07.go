package linecaps

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/moduleboundary"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

type metricModuleOwners struct {
	root        string
	byDirectory map[string]metricModuleOwner
	byRoot      map[string]sourcepolicy.ApplicabilityReason
}

type metricModuleOwner struct {
	root   string
	reason sourcepolicy.ApplicabilityReason
}

func (owners *metricModuleOwners) directory(path string) metricModuleOwner {
	directory := filepath.Dir(path)
	if owner, ok := owners.byDirectory[directory]; ok {
		return owner
	}
	root, err := moduleboundary.Directory(owners.root, directory)
	if err != nil {
		owner := metricModuleOwner{reason: sourcepolicy.ApplicabilityReasonModuleOwnerUnavailable}
		owners.byDirectory[directory] = owner
		return owner
	}
	relative, err := filepath.Rel(owners.root, root)
	if err != nil {
		owner := metricModuleOwner{reason: sourcepolicy.ApplicabilityReasonModuleOwnerUnavailable}
		owners.byDirectory[directory] = owner
		return owner
	}
	if relative == "." {
		relative = ""
	}
	relative = filepath.ToSlash(relative)
	reason, ok := owners.byRoot[root]
	if !ok {
		supported, supportErr := moduleboundary.LocalToolchainSupported(root)
		switch {
		case supportErr != nil:
			reason = sourcepolicy.ApplicabilityReasonModuleOwnerUnavailable
		case !supported:
			reason = sourcepolicy.ApplicabilityReasonModuleToolchainUnavailable
		}
		owners.byRoot[root] = reason
	}
	owner := metricModuleOwner{root: relative, reason: reason}
	owners.byDirectory[directory] = owner
	return owner
}

func (owners *metricModuleOwners) owner(path string, symlink bool) metricModuleOwner {
	owner := owners.directory(path)
	if !symlink || owner.reason != "" {
		return owner
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		owner.reason = sourcepolicy.ApplicabilityReasonModuleOwnerUnavailable
		return owner
	}
	target := owners.directory(resolved)
	if target.reason != "" || target.root != owner.root {
		owner.reason = sourcepolicy.ApplicabilityReasonCrossModuleSource
	}
	return owner
}

func measureFileMetric(fullPath, relative, extension string) (FileMetric, error) {
	metric := FileMetric{Path: relative, Language: FileLanguageOther}
	switch extension {
	case ".go":
		metric.Language = FileLanguageGo
	case ".gooo":
		metric.Language = FileLanguageGooo
	default:
		return metric, nil
	}
	source, err := os.ReadFile(fullPath)
	if err != nil {
		return FileMetric{}, err
	}
	metric.Lines = lineCount(source)
	return metric, nil
}

func accumulateFileMetric(directory *directoryNode, metric FileMetric) {
	switch metric.Language {
	case FileLanguageGo:
		directory.goFiles++
		directory.goLines += metric.Lines
	case FileLanguageGooo:
		directory.goooFiles++
		directory.goooLines += metric.Lines
	}
}

func collectMetricFile(absRoot, path string, symlink bool, directories map[string]*directoryNode,
	files *[]FileMetric, owners *metricModuleOwners) error {
	relative, err := filepath.Rel(absRoot, path)
	if err != nil {
		return err
	}
	relative = filepath.ToSlash(relative)
	parent := filepath.ToSlash(filepath.Dir(relative))
	ensureDirectoryNode(directories, parent)
	directories[parent].directFiles++
	extension := strings.ToLower(filepath.Ext(relative))
	metric, err := measureFileMetric(path, relative, extension)
	if err != nil {
		return err
	}
	if metric.Language == FileLanguageGo {
		owner := owners.owner(path, symlink)
		metric.ModuleRoot, metric.ModuleBoundaryReason = owner.root, owner.reason
	}
	*files = append(*files, metric)
	accumulateFileMetric(directories[parent], metric)
	return nil
}
