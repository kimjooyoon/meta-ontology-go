// Package workspacesplit couples explicitly supplied workspace consumers of one
// source partition. The host owns discovery, address/symlink identity, freshness
// and atomic publication of every returned member.
package workspacesplit

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

const MaxOwners = 64

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

type File struct {
	Path   string
	Source packageruntime.SplitSourceFile
}

type Plan struct {
	Owners []OwnedProposal
	Files  []File
}

// Prepare returns one joint draft or no draft. Equal content at different source
// addresses does not make those inputs the same occurrence. Addresses are lexical
// paths in the supplied reader's namespace; filesystem alias resolution is a host
// obligation. This constructor neither discovers omitted owners nor writes files.
func Prepare(reader workspaceinput.Reader, owners []Owner) (Plan, error) {
	if len(owners) == 0 || len(owners) > MaxOwners {
		return Plan{}, fmt.Errorf("shared split requires 1 to %d explicit owners", MaxOwners)
	}
	plan := Plan{}
	manifests := make(map[string]bool, len(owners))
	sourceBytes := 0
	for _, owner := range owners {
		proposal, err := prepareOwner(reader, owner, manifests)
		if err != nil {
			return Plan{}, err
		}
		files, err := sharedFiles(owner, proposal.Split)
		if err != nil {
			return Plan{}, err
		}
		if err := coupleFiles(plan.Files, files); err != nil {
			return Plan{}, err
		}
		sourceBytes += proposedSourceBytes(proposal.Split)
		if sourceBytes > workspaceinput.MaxSourceSetBytes {
			return Plan{}, fmt.Errorf("shared split proposals exceed the aggregate source byte limit")
		}
		if len(plan.Files) == 0 {
			plan.Files = files
		}
		plan.Owners = append(plan.Owners, proposal)
	}
	for _, file := range plan.Files {
		if manifests[file.Path] {
			return Plan{}, fmt.Errorf("workspace definition and split source addresses overlap")
		}
	}
	return plan, nil
}
