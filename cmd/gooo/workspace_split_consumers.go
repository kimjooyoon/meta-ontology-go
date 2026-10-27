package main

import (
	"fmt"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

const maxSplitConsumerWorkspaces = workspacesplit.MaxOwners - 1

type workspaceSourceSplitConsumer struct {
	ManifestPath     string                      `json:"manifest_path"`
	ManifestDigest   string                      `json:"manifest_digest"`
	UpdatedWorkspace *workspaceManifest          `json:"updated_workspace"`
	Result           *packageruntime.SourceSplit `json:"result"`
}

func (consumer workspaceSourceSplitConsumer) receipt() workspaceSourceSplitReceipt {
	return workspaceSourceSplitReceipt{Schema: "gooo/workspace-source-split-receipt/v1", Decision: "PROPOSED",
		ManifestPath: consumer.ManifestPath, ManifestDigest: consumer.ManifestDigest,
		UpdatedWorkspace: consumer.UpdatedWorkspace, Result: consumer.Result}
}

func prepareSharedWorkspaceSourceSplit(options workspaceSourceSplitOptions, reader SourceReader) (workspaceSourceSplitReceipt, error) {
	receipt := workspaceSourceSplitReceipt{Schema: "gooo/workspace-source-split-receipt/v1",
		Decision: "FAIL_CLOSED", ManifestPath: options.manifestPath}
	requested, err := sharedSplitRequests(options, reader)
	if err != nil {
		return receipt, err
	}
	plan, err := workspacesplit.PrepareFiles(reader, requested)
	if err != nil {
		return receipt, err
	}
	for i, owner := range plan.Owners {
		consumer := workspaceSourceSplitConsumer{ManifestPath: owner.ManifestPath,
			ManifestDigest: plan.Definitions[i].SHA256, UpdatedWorkspace: &owner.Split.Definition, Result: &owner.Split.Result}
		if i == 0 {
			receipt = consumer.receipt()
		} else {
			receipt.Consumers = append(receipt.Consumers, consumer)
		}
	}
	return receipt, nil
}

func sharedSplitRequests(options workspaceSourceSplitOptions, reader SourceReader) ([]workspacesplit.FileOwner, error) {
	if len(options.consumerWorkspaces) >= workspacesplit.MaxOwners {
		return nil, fmt.Errorf("too many consumer workspaces")
	}
	request := workspaceinput.SplitRequest{PackagePath: options.packagePath, Source: options.source,
		BeforeID: options.before, Into: options.into}
	requested := []workspacesplit.FileOwner{{ManifestPath: options.manifestPath, Request: request}}
	for _, name := range options.consumerWorkspaces {
		definition, _, err := loadWorkspaceRenameManifest(reader, name)
		if err != nil {
			return nil, err
		}
		relative, err := sharedSplitRequest(options, name, definition)
		if err != nil {
			return nil, err
		}
		requested = append(requested, workspacesplit.FileOwner{ManifestPath: name, Request: relative})
	}
	return requested, nil
}

func sharedSplitRequest(options workspaceSourceSplitOptions, manifestPath string,
	definition workspaceManifest) (workspaceinput.SplitRequest, error) {
	request := workspaceinput.SplitRequest{BeforeID: options.before}
	var err error
	request.Source, err = relativeSplitSource(options.manifestPath, manifestPath, options.source)
	if err != nil {
		return request, err
	}
	request.Into, err = relativeSplitSource(options.manifestPath, manifestPath, options.into)
	if err != nil {
		return request, err
	}
	for _, pkg := range definition.Packages {
		for _, source := range pkg.Sources {
			normalized, err := workspaceinput.SourcePath(source)
			if err != nil {
				return request, err
			}
			if filepath.ToSlash(normalized) == request.Source {
				request.PackagePath = pkg.Path
				return request, nil
			}
		}
	}
	return request, fmt.Errorf("consumer workspace %q does not declare the selected source", manifestPath)
}

func relativeSplitSource(primary, consumer, original string) (string, error) {
	source, err := workspaceinput.SourcePath(original)
	if err != nil {
		return "", err
	}
	absolute, err := filepath.Abs(filepath.Join(filepath.Dir(primary), source))
	if err != nil {
		return "", err
	}
	root, err := filepath.Abs(filepath.Dir(consumer))
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(root, absolute)
	if err != nil {
		return "", err
	}
	value, err := workspaceinput.SourcePath(filepath.ToSlash(relative))
	return filepath.ToSlash(value), err
}
