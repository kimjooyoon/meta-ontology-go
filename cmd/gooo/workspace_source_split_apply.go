package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

type workspaceSourceSplitApplication struct {
	ProposalSHA256       string `json:"proposal_sha256"`
	State                string `json:"state"`
	OutputManifestDigest string `json:"output_manifest_digest,omitempty"`
}

func applyWorkspaceSourceSplit(options workspaceSourceSplitOptions, ops atomicFileOps) (workspaceSourceSplitReceipt, error) {
	receipt := workspaceSourceSplitReceipt{Schema: "gooo/workspace-source-split-receipt/v1",
		Decision: "FAIL_CLOSED", ManifestPath: options.manifestPath}
	raw, err := readPackageReceipt(options.applyReceipt)
	if err != nil {
		return receipt, err
	}
	receipt.Application = &workspaceSourceSplitApplication{ProposalSHA256: workspaceDigest(raw), State: "NOT_APPLIED"}
	proposal, err := decodeWorkspaceSourceSplitProposal(raw)
	if err != nil {
		return receipt, err
	}
	proposalOptions := workspaceSourceSplitOptions{
		packagePath: proposal.Result.PackagePath, source: proposal.Result.InputFilename,
		before: proposal.Result.BeforeID, into: proposal.Result.Files[1].Filename, manifestPath: options.manifestPath}
	if proposal.Discovery != nil {
		proposalOptions.discoveryRoot = proposal.Discovery.Root
	} else {
		for _, consumer := range proposal.Consumers {
			proposalOptions.consumerWorkspaces = append(proposalOptions.consumerWorkspaces, consumer.ManifestPath)
		}
	}
	current, err := currentWorkspaceSourceSplitProposal(proposalOptions, proposal)
	if err != nil {
		return receipt, err
	}
	current.ManifestPath, current.Application = options.manifestPath, receipt.Application
	current.Decision = "FAIL_CLOSED"
	writes, err := sourceSplitAllWrites(current)
	if err != nil {
		return current, err
	}
	priorCheck := ops.beforePublish
	ops.beforePublish = func() error {
		return recheckSourceSplitProjection(proposalOptions, proposal, current, writes, priorCheck)
	}
	if err := writeAtomicFilesWithOps(writes, ops); err != nil {
		current.Application.State = "STATE_DIFFERS"
		if checkAtomicPreconditions(writes) == nil {
			current.Application.State = "PRE_STATE_RETAINED"
		}
		return current, err
	}
	if err := verifyAppliedSourceSplit(current); err != nil {
		current.Application.State = "VERIFICATION_FAILED"
		return current, err
	}
	current.Decision, current.Application.State = "APPLIED", "OUTPUT_VERIFIED"
	return current, nil
}

func currentWorkspaceSourceSplitProposal(options workspaceSourceSplitOptions, proposal workspaceSourceSplitReceipt) (workspaceSourceSplitReceipt, error) {
	current, err := prepareWorkspaceSourceSplit(options, OSFileReader{})
	if err != nil {
		return current, err
	}
	current.ManifestPath = proposal.ManifestPath
	if !reflect.DeepEqual(sourceSplitPublicReceipt(current), sourceSplitPublicReceipt(proposal)) {
		return current, fmt.Errorf("saved split proposal differs from current workspace or canonical partition")
	}
	return current, nil
}

func decodeWorkspaceSourceSplitProposal(raw []byte) (workspaceSourceSplitReceipt, error) {
	var proposal workspaceSourceSplitReceipt
	if err := completeness.CheckJSON(raw, packageReceiptMaxBytes); err != nil {
		return proposal, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&proposal); err != nil {
		return proposal, err
	}
	if proposal.Schema != "gooo/workspace-source-split-receipt/v1" || proposal.Decision != "PROPOSED" ||
		proposal.Application != nil || proposal.Error != "" || proposal.Result == nil ||
		proposal.UpdatedWorkspace == nil || len(proposal.Result.Files) != 2 ||
		len(proposal.Consumers) > maxSplitConsumerWorkspaces {
		return proposal, fmt.Errorf("apply requires a complete, unapplied source split proposal")
	}
	for _, consumer := range proposal.Consumers {
		if consumer.ManifestPath == "" || consumer.ManifestDigest == "" || consumer.UpdatedWorkspace == nil ||
			consumer.Result == nil || len(consumer.Result.Files) != 2 {
			return proposal, fmt.Errorf("apply requires complete consumer proposals")
		}
	}
	if proposal.Discovery != nil && (proposal.Discovery.Root == "" ||
		proposal.Discovery.Candidates != "recursive-json-files" || len(proposal.Discovery.Definitions) == 0) {
		return proposal, fmt.Errorf("apply requires a complete workspace discovery scope")
	}
	return proposal, nil
}

func sourceSplitWrites(receipt workspaceSourceSplitReceipt) ([]atomicWrite, error) {
	manifestPath, err := filepath.Abs(receipt.ManifestPath)
	if err != nil {
		return nil, err
	}
	root, err := filepath.EvalSymlinks(filepath.Dir(manifestPath))
	if err != nil {
		return nil, err
	}
	manifestPath = filepath.Join(root, filepath.Base(manifestPath))
	manifestBefore, err := captureOutputSnapshot(manifestPath)
	if err != nil {
		return nil, err
	}
	if !manifestBefore.exists || workspaceDigest(manifestBefore.data) != receipt.ManifestDigest {
		return nil, fmt.Errorf("workspace manifest changed before split application")
	}
	writes := make([]atomicWrite, 0, 3)
	for i, file := range receipt.Result.Files {
		target, err := sourceSplitTarget(root, file.Filename)
		if err != nil {
			return nil, err
		}
		before, err := captureOutputSnapshot(target)
		if err != nil {
			return nil, fmt.Errorf("inspect split target %q: %w", file.Filename, err)
		}
		if (i == 0 && (!before.exists || workspaceDigest(before.data) != receipt.Result.InputSHA256)) ||
			(i == 1 && before.exists) {
			return nil, fmt.Errorf("split target precondition changed: %s", file.Filename)
		}
		if i == 1 {
			before.mode = writes[0].before.mode
		}
		writes = append(writes, atomicWrite{path: target, data: []byte(file.Source), before: &before})
	}
	manifestData, err := workspaceinput.Encode(*receipt.UpdatedWorkspace)
	if err != nil {
		return nil, err
	}
	return append(writes, atomicWrite{path: manifestPath, data: manifestData, before: &manifestBefore}), nil
}

func sourceSplitTarget(root, filename string) (string, error) {
	relative, err := workspaceSourcePath(filename)
	if err != nil {
		return "", err
	}
	target := filepath.Join(root, relative)
	parent, err := filepath.EvalSymlinks(filepath.Dir(target))
	if err != nil {
		return "", err
	}
	if parent != root && !pathContained(root, parent) {
		return "", fmt.Errorf("split target parent is outside the workspace")
	}
	return filepath.Join(parent, filepath.Base(target)), nil
}

func verifyAppliedSourceSplit(receipt workspaceSourceSplitReceipt) error {
	digest, err := verifySourceSplitOutput(receipt, OSFileReader{})
	if err == nil {
		receipt.Application.OutputManifestDigest = digest
	}
	return err
}

func verifySourceSplitOutput(receipt workspaceSourceSplitReceipt, reader SourceReader) (string, error) {
	digest, err := verifySingleSourceSplitOutput(receipt, reader)
	if err != nil {
		return "", err
	}
	for _, consumer := range receipt.Consumers {
		if _, err := verifySingleSourceSplitOutput(consumer.receipt(), reader); err != nil {
			return "", fmt.Errorf("consumer workspace %q: %w", consumer.ManifestPath, err)
		}
	}
	return digest, nil
}

func verifySingleSourceSplitOutput(receipt workspaceSourceSplitReceipt, reader SourceReader) (string, error) {
	manifest, digest, err := loadWorkspaceRenameManifest(reader, receipt.ManifestPath)
	if err != nil {
		return "", err
	}
	expected, err := workspaceinput.Encode(*receipt.UpdatedWorkspace)
	if err != nil {
		return "", err
	}
	if digest != workspaceDigest(expected) {
		return "", fmt.Errorf("workspace manifest differs from the proposal")
	}
	loaded, err := loadPackageSources(reader, receipt.ManifestPath, manifest)
	if err != nil {
		return "", err
	}
	view, err := packageruntime.Describe(loaded)
	if err != nil {
		return "", err
	}
	if view.ImageDigest != receipt.Result.OutputImageDigest {
		return "", fmt.Errorf("package image differs from the proposal")
	}
	return digest, nil
}
