package workspacesplit

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

// FileOwner names authored input rather than supplying a reconstructed definition.
type FileOwner struct {
	ManifestPath string
	Request      workspaceinput.SplitRequest
}

// DefinitionObservation records the exact serialized input, including formatting.
// ManifestPath retains the caller's reader address, as in OwnedProposal.
type DefinitionObservation struct {
	ManifestPath string
	SHA256       string
	ByteLength   int
}

var ErrDefinitionChanged = errors.New("workspace definition changed during preparation")

// PrepareFiles reads every supplied definition, prepares the shared source change,
// and rereads the definitions before returning. Publication still needs current
// input and address checks; these sequential reads are not a filesystem snapshot.
func PrepareFiles(reader workspaceinput.Reader, requested []FileOwner) (Plan, error) {
	if reader == nil || len(requested) == 0 || len(requested) > MaxOwners {
		return Plan{}, fmt.Errorf("shared split requires a reader and 1 to %d explicit owners", MaxOwners)
	}
	owners := make([]Owner, 0, len(requested))
	inputs := make([]DefinitionObservation, 0, len(requested))
	total := 0
	for _, request := range requested {
		raw, input, err := readDefinitionInput(reader, request.ManifestPath)
		if err != nil {
			return Plan{}, err
		}
		total += len(raw)
		if total > workspaceinput.MaxSourceSetBytes {
			return Plan{}, fmt.Errorf("workspace definitions exceed the aggregate byte limit")
		}
		definition, err := workspaceinput.Decode(raw)
		if err != nil {
			return Plan{}, fmt.Errorf("workspace %q: %w", request.ManifestPath, err)
		}
		owners = append(owners, Owner{ManifestPath: request.ManifestPath, Definition: definition, Request: request.Request})
		inputs = append(inputs, input)
	}
	plan, err := Prepare(reader, owners)
	if err != nil {
		return Plan{}, err
	}
	for _, input := range inputs {
		_, current, err := readDefinitionInput(reader, input.ManifestPath)
		if err != nil {
			return Plan{}, err
		}
		if current != input {
			return Plan{}, fmt.Errorf("%w: %s", ErrDefinitionChanged, input.ManifestPath)
		}
	}
	plan.Definitions = inputs
	return plan, nil
}

func readDefinitionInput(reader workspaceinput.Reader, path string) ([]byte, DefinitionObservation, error) {
	raw, err := reader.ReadFile(path)
	if err != nil {
		return nil, DefinitionObservation{}, fmt.Errorf("workspace %q: %w", path, err)
	}
	if len(raw) > workspaceinput.MaxDefinitionBytes {
		return nil, DefinitionObservation{}, fmt.Errorf("workspace %q exceeds the definition byte limit", path)
	}
	return raw, DefinitionObservation{ManifestPath: path, SHA256: fmt.Sprintf("sha256:%x", sha256.Sum256(raw)), ByteLength: len(raw)}, nil
}
