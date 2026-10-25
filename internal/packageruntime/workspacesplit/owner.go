package workspacesplit

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

type Owner struct {
	ManifestPath string
	Definition   workspaceinput.Definition
	Request      workspaceinput.SplitRequest
}

type OwnedProposal struct {
	ManifestPath   string
	Split          workspaceinput.SplitProposal
	DefinitionJSON []byte
}

func prepareOwner(reader *sourceObservations, owner Owner, manifests map[string]bool) (OwnedProposal, error) {
	if strings.TrimSpace(owner.ManifestPath) == "" {
		return OwnedProposal{}, fmt.Errorf("workspace definition address is required")
	}
	address, err := filepath.Abs(owner.ManifestPath)
	if err != nil {
		return OwnedProposal{}, err
	}
	if manifests[address] {
		return OwnedProposal{}, fmt.Errorf("duplicate workspace definition address")
	}
	manifests[address] = true
	reader.consumer = address
	split, err := workspaceinput.PrepareSplit(reader, owner.ManifestPath, owner.Definition, owner.Request)
	if err != nil {
		return OwnedProposal{}, fmt.Errorf("workspace %q: %w", owner.ManifestPath, err)
	}
	raw, err := workspaceinput.Encode(split.Definition)
	if err != nil {
		return OwnedProposal{}, err
	}
	return OwnedProposal{ManifestPath: owner.ManifestPath, Split: split, DefinitionJSON: raw}, nil
}

func proposedSourceBytes(proposal workspaceinput.SplitProposal) int {
	size := 0
	for _, pkg := range proposal.Result.Manifest.Packages {
		for _, source := range pkg.Sources {
			size += len(source.Content)
		}
	}
	return size
}
