package workspacesplit

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

// InputObservation records a source read, not an external definition read or a
// filesystem snapshot. Consumers are the definitions that used this occurrence.
type InputObservation struct {
	Path, SHA256 string
	ByteLength   int
	Consumers    []string
}

func sharedFiles(owner Owner, proposal workspaceinput.SplitProposal) ([]File, error) {
	files := make([]File, len(proposal.Result.Files))
	for i, source := range proposal.Result.Files {
		relative, err := workspaceinput.SourcePath(source.Filename)
		if err != nil {
			return nil, err
		}
		address, err := filepath.Abs(filepath.Join(filepath.Dir(owner.ManifestPath), relative))
		if err != nil {
			return nil, err
		}
		files[i] = File{Path: address, Source: source}
	}
	return files, nil
}

func coupleFiles(expected, candidate []File) error {
	if len(expected) == 0 {
		return nil
	}
	if len(expected) != len(candidate) {
		return fmt.Errorf("shared consumers disagree on split member count")
	}
	for i, file := range expected {
		other := candidate[i]
		if file.Path != other.Path || file.Source.Source != other.Source.Source ||
			file.Source.SHA256 != other.Source.SHA256 || !slices.Equal(file.Source.Origins, other.Source.Origins) {
			return fmt.Errorf("shared consumers disagree on source address, partition bytes or origins")
		}
	}
	return nil
}

func checkMemberAddresses(plan Plan, manifests map[string]bool) error {
	for _, file := range plan.Files {
		if manifests[file.Path] {
			return fmt.Errorf("workspace definition and split source addresses overlap: %s", file.Path)
		}
	}
	for _, input := range plan.Inputs {
		if manifests[input.Path] {
			return fmt.Errorf("workspace definition would overwrite an observed source: %s", input.Path)
		}
	}
	return nil
}
