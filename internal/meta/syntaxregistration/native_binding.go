package syntaxregistration

import (
	"io/fs"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

type NativeBindingInfo struct {
	ActivityID      string
	OutputID        string
	InputActivityID string
	InputOutputID   string
	SourceDigest    string
	SemanticDigest  string
}

// NativeBinding describes only the exact compiled Gooo contract, not authority.
func NativeBinding() (NativeBindingInfo, error) {
	binding, err := bindContract()
	if err != nil {
		return NativeBindingInfo{}, err
	}
	return NativeBindingInfo{ActivityID: binding.activities[0], OutputID: binding.outputs[0],
		InputActivityID: binding.inputActivity, InputOutputID: binding.inputOutput,
		SourceDigest: binding.source, SemanticDigest: binding.semantic}, nil
}

func RequestDigest(request Request) string { return digestValue(request) }

// ObservePresence reports an exact declaration match. It never infers a case
// from names, a score or prose, and does not accept contradictory registration.
func ObservePresence(repository fs.FS, request Request) (bool, error) {
	snapshot, source, err := InspectInputs(repository, request)
	if err != nil {
		return false, err
	}
	if snapshot != request.SnapshotDigest || source != request.SourceDigest {
		return false, failure("UNKNOWN", "observe-registration", "REGISTRATION_INPUT_STALE",
			"STALE", "refresh-exact-input-binding")
	}
	raw, err := fs.ReadFile(repository, corpusPath)
	if err != nil {
		return false, err
	}
	var registry languagesyntax.Registry
	if err := decodeStrict(raw, &registry); err != nil {
		return false, err
	}
	for _, observed := range registry.Cases {
		if observed.ID != request.Case.ID && observed.Path != request.Case.Path {
			continue
		}
		if !reflect.DeepEqual(observed, request.Case) {
			return false, failure("REFUTED", "observe-registration", "REGISTRATION_CASE_IDENTITY_CONTRADICTION",
				"", "report-registration-counterexample")
		}
		return true, nil
	}
	return false, nil
}
