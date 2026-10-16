package conceptoperation

import (
	"fmt"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

// DeclaredCohortDigest binds declared identities and roles before observations.
// Observed values, status and evidence digests do not define the denominator.
func DeclaredCohortDigest(baseline Receipt) (string, error) {
	contract := append([]OperationBinding(nil), baseline.Expected...)
	for index := range contract {
		contract[index].Actual, contract[index].Status = "", ""
		contract[index].EvidenceDigest, contract[index].OperationDigest = "", ""
	}
	return artifact.Digest(contract)
}

// DeclareObservations exposes canonical known facts as an explicit observation
// input. Callers may omit rows or append unbound IDs; neither adds authority.
func DeclareObservations(baseline Receipt, strategy []byte) (ObservationDocument, error) {
	if err := VerifyReceipt(baseline, baseline.Repository, baseline.SubjectSHA); err != nil {
		return ObservationDocument{}, err
	}
	if baseline.Status != "VERIFIED" {
		return ObservationDocument{}, fmt.Errorf("observations require a verified canonical producer")
	}
	bindings, err := observationBindings(strategy)
	if err != nil {
		return ObservationDocument{}, err
	}
	cohort, err := DeclaredCohortDigest(baseline)
	if err != nil {
		return ObservationDocument{}, err
	}
	result := ObservationDocument{Schema: ObservationDocumentSchema, Repository: baseline.Repository,
		SubjectSHA: baseline.SubjectSHA, ProgramDigest: baseline.ProgramDigest, DeclaredCohortDigest: cohort,
		Observations: make([]OperationObservation, 0, baseline.ExpectedCount)}
	for _, expected := range baseline.Expected {
		binding, ok := bindings[expected.IndicatorID]
		if !ok {
			return ObservationDocument{}, fmt.Errorf("declared operation %q has no canonical observation", expected.Operation)
		}
		result.Observations = append(result.Observations, OperationObservation{expected.Operation, &binding})
	}
	return result, nil
}

func observationBindings(strategy []byte) (map[string]metricstrategy.Binding, error) {
	var plan metricstrategy.Plan
	if err := decodeExact(strategy, &plan); err != nil {
		return nil, err
	}
	result := make(map[string]metricstrategy.Binding)
	for _, binding := range plan.Bindings {
		if metricstrategy.IsConceptOperationBinding(binding.IndicatorID) {
			if _, exists := result[binding.IndicatorID]; exists {
				return nil, fmt.Errorf("canonical observation binding is ambiguous")
			}
			result[binding.IndicatorID] = binding
		}
	}
	return result, nil
}

func validateObservationRows(document ObservationDocument, baseline Receipt,
	strategy []byte) ([]string, []string, error) {
	if len(document.Observations) > bodycodegen.RepeatedFieldLimit ||
		len(baseline.Expected) > bodycodegen.RepeatedFieldLimit || document.Observations == nil {
		return nil, nil, fmt.Errorf("operation observations exceed the declared repeated-field profile")
	}
	bindings, err := observationBindings(strategy)
	if err != nil {
		return nil, nil, err
	}
	expected := make([]string, 0, baseline.ExpectedCount)
	byOperation := make(map[string]metricstrategy.Binding, baseline.ExpectedCount)
	for _, operation := range baseline.Expected {
		expected = append(expected, operation.Operation)
		byOperation[operation.Operation] = bindings[operation.IndicatorID]
	}
	observed := make([]string, 0, len(document.Observations))
	for _, row := range document.Observations {
		if row.Operation == "" || len(row.Operation) > 1024 || !utf8.ValidString(row.Operation) {
			return nil, nil, fmt.Errorf("observed operation requires a bounded UTF-8 identity")
		}
		binding, known := byOperation[row.Operation]
		if known && (row.Binding == nil || !artifact.Equal(*row.Binding, binding)) {
			return nil, nil, fmt.Errorf("observed binding differs from source operation %q", row.Operation)
		}
		if !known && row.Binding != nil {
			return nil, nil, fmt.Errorf("outside-cohort observation cannot claim an executable binding")
		}
		observed = append(observed, row.Operation)
	}
	return expected, observed, nil
}
