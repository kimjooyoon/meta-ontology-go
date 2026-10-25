package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type workspaceSplitRegistration struct {
	Root      string                                    `json:"root"`
	Directory string                                    `json:"directory"`
	Profile   syntax.EntityFieldsSupport                `json:"workspace_profile"`
	Operation generation.OperationInputContractEvidence `json:"operation"`
	Inputs    []workspaceSplitRegistrationInput         `json:"inputs"`
	Outputs   []workspaceSplitRegistrationOutput        `json:"outputs"`
}

type workspaceSplitRegistrationInput struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type workspaceSplitRegistrationOutput struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Source string `json:"source"`
}

func prepareRegisteredWorkspaceSplit(options workspaceSourceSplitOptions, reader SourceReader) (receipt workspaceSourceSplitReceipt, err error) {
	defer func() {
		if err != nil {
			receipt = workspaceSourceSplitReceipt{Schema: "gooo/workspace-source-split-receipt/v1",
				Decision: "FAIL_CLOSED", ManifestPath: options.manifestPath, ManifestDigest: receipt.ManifestDigest}
		}
	}()
	requested := options.registryRoot
	options.registryRoot = ""
	receipt, err = prepareWorkspaceSourceSplit(options, reader)
	if err != nil {
		return receipt, err
	}
	receipt.Decision = "FAIL_CLOSED"
	root, directory, err := splitRegistrationLocation(requested, options.manifestPath)
	if err != nil {
		return receipt, err
	}
	partition, err := workspaceRegistrationPartition(options, reader)
	if err != nil {
		return receipt, err
	}
	if !reflect.DeepEqual(partition.Files, receipt.Result.Files) {
		return receipt, fmt.Errorf("workspace and syntax registration disagree on physical source partition")
	}
	candidate, err := syntaxregistration.PreparePartitionRegistrationCandidate(os.DirFS(root), directory, partition)
	if err != nil {
		return receipt, err
	}
	registration, err := describeSplitRegistration(root, directory, partition, candidate)
	if err != nil {
		return receipt, err
	}
	receipt.Registration, receipt.Decision = registration, "PROPOSED"
	return receipt, nil
}

func splitRegistrationLocation(requested, manifest string) (string, string, error) {
	root, err := filepath.Abs(requested)
	if err == nil {
		root, err = filepath.EvalSymlinks(root)
	}
	if err != nil {
		return "", "", err
	}
	directory, err := filepath.Abs(filepath.Dir(manifest))
	if err == nil {
		directory, err = filepath.EvalSymlinks(directory)
	}
	if err != nil || (root != directory && !pathContained(root, directory)) {
		return "", "", fmt.Errorf("workspace must be inside the syntax registry root")
	}
	relative, err := filepath.Rel(root, directory)
	return root, filepath.ToSlash(relative), err
}

func workspaceRegistrationPartition(options workspaceSourceSplitOptions, reader SourceReader) (sourcebundle.Partition, error) {
	manifest, _, err := loadWorkspaceRenameManifest(reader, options.manifestPath)
	if err != nil {
		return sourcebundle.Partition{}, err
	}
	loaded, err := loadPackageSources(reader, options.manifestPath, manifest)
	if err != nil {
		return sourcebundle.Partition{}, err
	}
	for _, pkg := range loaded.Packages {
		if pkg.Path != options.packagePath {
			continue
		}
		if len(pkg.Imports) != 0 {
			return sourcebundle.Partition{}, fmt.Errorf("registered source partition requires a source-bundle adapter for imports")
		}
		sources := make([]sourcebundle.Source, len(pkg.Sources))
		for i, source := range pkg.Sources {
			sources[i] = sourcebundle.Source{Filename: source.Filename, Content: source.Content}
		}
		parsed, err := sourcebundle.ParseWithEntityFieldsSupport(sources, syntax.EntityFieldsV4Support())
		if err != nil {
			return sourcebundle.Partition{}, err
		}
		// Match package execution's existing profile selection; registry consumers
		// independently retain and replay their own declared profiles.
		support := syntax.EntityFieldsSupportForFile(parsed.File, syntax.EntityFieldsV4Support())
		return sourcebundle.PreparePartitionWithEntityFieldsSupport(sources, options.source, options.before, options.into, support)
	}
	return sourcebundle.Partition{}, fmt.Errorf("registered source package is absent")
}
