package generation

import (
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

// The legacy checker still owns exactly its four original native bindings.
// The fifth binding is independently pinned to its compiled Gooo input ABI.
func DefaultRegistry() []Binding {
	binding, valid := nativeRegistrationBinding()
	if !valid {
		return nil
	}
	return append(legacyDefaultRegistry(), binding)
}

func nativeRegistrationBinding() (Binding, bool) {
	evidence, err := syntaxregistration.NativeBinding()
	if err != nil {
		return Binding{}, false
	}
	source := strings.TrimPrefix(evidence.SourceDigest, "sha256:")
	semantic := strings.TrimPrefix(evidence.SemanticDigest, "sha256:")
	if !validDigest(source) || !validDigest(semantic) ||
		evidence.ActivityID == "" || evidence.InputActivityID == "" || evidence.InputOutputID == "" {
		return Binding{}, false
	}
	return Binding{Operation: sourcepolicy.OperationRegisterSyntax,
		Activity: evidence.ActivityID, Output: evidence.OutputID,
		InputSubjectKind:          sourcepolicy.SubjectKindRegistrationRequest,
		InputContractSourceDigest: source, InputContractSemanticDigest: semantic,
		IndependenceGroupID: "syntax-registration", ProofChoice: ProofCoherence,
		Executor:  "scripts/meta-execution:registration-worker",
		Evaluator: "scripts/meta-execution:registration-conformance",
		RequiredIndicatorIDs: []string{"registration.artifact-completeness/v1",
			"registration.execution-identity/v1", "registration.native-conformance/v1", "registration.replay/v1"},
		ReceiptRequired: true, Priority: 5}, true
}

func registryIndex(registry []Binding) (map[sourcepolicy.Operation]Binding, bool) {
	var legacy []Binding
	var registration []Binding
	for _, binding := range registry {
		if binding.Operation == sourcepolicy.OperationRegisterSyntax {
			registration = append(registration, binding)
		} else {
			legacy = append(legacy, binding)
		}
	}
	expected, known := nativeRegistrationBinding()
	if !known || len(registration) != 1 || !reflect.DeepEqual(registration[0], expected) {
		return nil, false
	}
	index, valid := legacyRegistryIndex(legacy)
	if !valid {
		return nil, false
	}
	index[expected.Operation] = registration[0]
	return index, true
}

func BindingForOperation(registry []Binding, operation sourcepolicy.Operation) (Binding, bool) {
	index, valid := registryIndex(registry)
	if !valid {
		return Binding{}, false
	}
	binding, exists := index[operation]
	return binding, exists
}
