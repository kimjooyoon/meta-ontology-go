package syntaxregistration

import (
	"fmt"
	"io/fs"
	"path"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

// PartitionRegistrationCandidate carries one proposed write set: both source
// fragments, the corpus and their native registry owners. Inputs retain exact
// observed bytes, including unchanged siblings. It grants no application authority
// and does not establish a global filesystem snapshot or native ABI admission.
type PartitionRegistrationCandidate struct {
	Projection languagesyntax.PartitionRegistration
	Inputs     map[string][]byte
	Outputs    map[string][]byte
}

// PreparePartitionRegistrationCandidate resolves native ownership from the
// registry's finite source expressions, then constructs and rechecks all edits.
func PreparePartitionRegistrationCandidate(repository fs.FS, directory string,
	partition sourcebundle.Partition) (PartitionRegistrationCandidate, error) {
	nativeInputs, err := readPartitionNativeInputs(repository)
	if err != nil {
		return PartitionRegistrationCandidate{}, err
	}
	var before languagesyntax.Registry
	if err := decodeStrict(nativeInputs[corpusPath], &before); err != nil {
		return PartitionRegistrationCandidate{}, err
	}
	projection, err := languagesyntax.PreparePartitionRegistration(repository, nativeInputs[corpusPath], directory, partition)
	if err != nil {
		return PartitionRegistrationCandidate{}, err
	}
	inputs, err := capturePartitionSources(repository, directory, partition, before, projection)
	if err != nil {
		return PartitionRegistrationCandidate{}, err
	}
	outputs, err := generatePartitionNative(nativeInputs, before, projection.Registry)
	if err != nil {
		return PartitionRegistrationCandidate{}, err
	}
	outputs[corpusPath] = projection.Corpus
	for _, file := range partition.Files {
		outputs[path.Join(directory, file.Filename)] = []byte(file.Source)
	}
	current, err := readPartitionNativeInputs(repository)
	if err != nil {
		return PartitionRegistrationCandidate{}, err
	}
	if !reflect.DeepEqual(current, nativeInputs) {
		return PartitionRegistrationCandidate{}, fmt.Errorf("native registry inputs changed during partition preparation")
	}
	current, err = readPartitionNamedInputs(repository, sortedPaths(inputs))
	if err != nil {
		return PartitionRegistrationCandidate{}, err
	}
	if !reflect.DeepEqual(current, inputs) {
		return PartitionRegistrationCandidate{}, fmt.Errorf("source consumer inputs changed during partition preparation")
	}
	for name, raw := range nativeInputs {
		inputs[name] = raw
	}
	return PartitionRegistrationCandidate{Projection: projection, Inputs: inputs, Outputs: outputs}, nil
}

// Validation reconstructs the complete candidate; changing or resealing returned
// content cannot make an altered write set match current source preparation.
func ValidatePartitionRegistrationCandidate(repository fs.FS, directory string, partition sourcebundle.Partition,
	candidate PartitionRegistrationCandidate) error {
	expected, err := PreparePartitionRegistrationCandidate(repository, directory, partition)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(expected, candidate) {
		return fmt.Errorf("partition registration candidate differs from current preparation")
	}
	return nil
}
