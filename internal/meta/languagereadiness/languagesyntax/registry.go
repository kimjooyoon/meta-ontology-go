package languagesyntax

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagepackageexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
)

func expectedRegistry() Registry {
	valid := func(id, path string) CaseDefinition {
		return CaseDefinition{ID: id, Path: path, Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability}
	}
	invalid := func(id, path, diagnostic string) CaseDefinition {
		return CaseDefinition{ID: id, Path: path, Kind: KindInvalid, ExpectedDecision: DecisionClosed, ExpectedDiagnostic: diagnostic, ProofChoice: "REGRESSION", MetaOperation: "reject-invalid-syntax", Scope: ScopeLanguageCapability}
	}
	governance := func(id, path string) CaseDefinition {
		return CaseDefinition{ID: id, Path: path, Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeGovernanceObservation}
	}
	entityFields := valid("entity-fields-v1", "examples/entity-fields-v1/main.gooo")
	entityFields.EntityFields = true
	ciTimeCausality := valid("ci-time-causality", "examples/ci-time-causality/main.gooo")
	ciTimeCausality.ImplicitActivityPorts = true
	packageUnit := PackageDefinition{ID: "billing-package", Path: "examples/billing-package", Members: []string{"examples/billing-package/activity.gooo", "examples/billing-package/entities.gooo"}, Entry: "PayOrder", ReportSchema: languagepackageexecution.ReportSchema, MetaReducer: "languagepackageexecution.Evaluate", SourceFilesIndicator: "PACKAGE_SOURCE_FILES", ExecutionIndicator: "PACKAGE_EXECUTIONS"}
	symbolicUnit := PackageDefinition{ID: "symbolic-invocation-schema", Path: "examples/symbolic-invocation-schema", Members: []string{"examples/symbolic-invocation-schema/activity.gooo", "examples/symbolic-invocation-schema/entities.gooo", "examples/symbolic-invocation-schema/reader-request.gooo"}, Entry: "Checkout", ReportSchema: languagepackageexecution.ReportSchema, MetaReducer: "languagepackageexecution.Evaluate", SourceFilesIndicator: "PACKAGE_SOURCE_FILES", ExecutionIndicator: "PACKAGE_EXECUTIONS"}
	selfImprovementObservationUnit := PackageDefinition{ID: "self-improvement-observation", Path: "examples/self-improvement-observation", Members: []string{"examples/self-improvement-observation/observation.gooo"}, Entry: "DeclareOperationIntent", ReportSchema: languagepackageexecution.ReportSchema, MetaReducer: "languagepackageexecution.Evaluate", SourceFilesIndicator: "PACKAGE_SOURCE_FILES", ExecutionIndicator: "PACKAGE_EXECUTIONS"}
	partialReuseUnit := PackageDefinition{ID: "self-improvement-partial-reuse", Path: "examples/self-improvement-partial-reuse", Members: []string{"examples/self-improvement-partial-reuse/main.gooo"}, Entry: "CreateReceipt", ReportSchema: languagepackageexecution.ReportSchema, MetaReducer: "languagepackageexecution.Evaluate", SourceFilesIndicator: "PACKAGE_SOURCE_FILES", ExecutionIndicator: "PACKAGE_EXECUTIONS"}
	return Registry{Schema: RegistrySchema, Cases: []CaseDefinition{
		valid("billing", "examples/billing/main.gooo"),
		valid("language-test-pass", "examples/language-test/main.gooo"),
		valid("language-test-failing-assertion", "examples/language-test/failing.gooo"),
		valid("bootstrap", "examples/bootstrap/main.gooo"),
		valid("conformance", "examples/conformance/main.gooo"),
		valid("compiler-self-improvement", "examples/compiler-self-improvement/main.gooo"),
		valid("compiler-self-improvement-operation-envelope", "examples/compiler-self-improvement/operation-envelope.gooo"),
		valid("causal-ci-selection", "examples/causal-ci-selection/main.gooo"),
		valid("causal-ci-selection-semantic-intervention", "examples/causal-ci-selection/semantic-intervention.gooo"),
		valid("causal-ci-selection-nonsemantic-intervention", "examples/causal-ci-selection/nonsemantic-intervention.gooo"),
		valid("causal-ci-selection-contradiction-intervention", "examples/causal-ci-selection/contradiction-intervention.gooo"),
		valid("external-capability-authorization-policy", "examples/external-capability-execution/authorization/policy.gooo"),
		valid("activity-cardinality-resolution", "examples/activity-cardinality-resolution/main.gooo"),
		valid("meta-actionability", "examples/meta-actionability/main.gooo"),
		valid("meta-binding-coverage", "examples/meta-binding-coverage/main.gooo"),
		valid("meta-operation-artifact-coverage", "examples/meta-operation-artifact-coverage/main.gooo"),
		valid("meta-policy-compilation", "examples/meta-policy-compilation/policy.gooo"),
		valid("metric-meta-program-closure", "examples/metric-meta-program-closure/main.gooo"),
		valid("metric-meta-program", "examples/metric-meta-program/main.gooo"),
		valid("root-readme-indicator", "examples/root-readme-indicator/main.gooo"),
		valid("self-improvement", "examples/self-improvement/main.gooo"),
		valid("self-improvement-minimal-loop", "examples/self-improvement-minimal-loop/main.gooo"),
		valid("self-improvement-operation-envelope", "examples/self-improvement-minimal-loop/operation-envelope.gooo"),
		valid("self-improvement-candidate", "examples/self-improvement/candidate.gooo"),
		valid("self-improvement-transport", "examples/self-improvement/transport.gooo"),
		valid("language-value-witness", "examples/language-value-witness/main.gooo"),
		valid("language-operation-catalog", "examples/language-operation-catalog/main.gooo"),
		valid("language-operation-catalog-unknown", "examples/language-operation-catalog/unknown.gooo"),
		valid("claim-resolution-tuple", "cmd/gooo/testdata/claim-resolution/main.gooo"),
		valid("claim-dependency-causality", "cmd/gooo/testdata/claim-dependency/main.gooo"),
		valid("roundtrip-minimal", "internal/detection/roundtrip/testdata/minimal.gooo"),
		valid("directory-kind-ontology", "internal/meta/directorykind/ontology.gooo"),
		valid("directory-partition-ontology", "internal/meta/directorypartition/ontology.gooo"),
		valid("repository-projection-repair", "examples/repository-projection-repair/main.gooo"),
		valid("opentofu-observation", "examples/opentofu-observation/main.gooo"),
		valid("public-self-observation-discovery-policy", "examples/self-improvement-discovery/discovery.gooo"),
		valid("public-self-observation-discovery-project", "examples/self-improvement-discovery/project.gooo"),
		invalid("unknown-keyword", "examples/language-syntax-roundtrip/unknown-keyword.txt", "parse.unexpected-token"),
		invalid("unterminated-string", "examples/language-syntax-roundtrip/unterminated-string.txt", "lex.unterminated-string"),
		invalid("source-execution-invalid", "examples/language-source-execution/invalid.gooo", "parse.unexpected-token"),
		valid("source-splitter-conformance", "examples/source-splitter-conformance/main.gooo"),
		valid("rollback-integrity-activation", "examples/rollback-integrity-activation/main.gooo"),
		valid("vertical-slice-closure-activation", "examples/vertical-slice-closure-activation/main.gooo"),
		valid("workgraph", "examples/workgraph/main.gooo"),
		valid("external-conformance-activation", "examples/external-conformance-activation/main.gooo"),
		valid("gooo-release-publication", "examples/gooo-release-publication/main.gooo"),
		valid("ci-plan", "examples/ci-plan/main.gooo"),
		valid("ci-effort-observation", "examples/ci-effort-observation/main.gooo"),
		ciTimeCausality,
		valid("reproducibility-semantics", "examples/reproducibility-semantics/main.gooo"),
		entityFields,
		valid("temporal-transition-ticket", "examples/temporal-transition-ticket/main.gooo"),
		governance("live-governance-snapshot", "examples/live-governance-snapshot/main.gooo"),
	}, PackageUnits: []PackageDefinition{packageUnit, symbolicUnit, selfImprovementObservationUnit, partialReuseUnit}, MetaSources: []string{"internal/meta/entityfields/entity-fields-meta.gooo"}}
}

func decodeRegistry(raw []byte) (Registry, error) {
	registry := Registry{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&registry); err != nil {
		return registry, fmt.Errorf("decode language syntax registry: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return registry, fmt.Errorf("decode language syntax registry: trailing content")
	}
	if err := validateCaseScopes(registry); err != nil {
		return registry, err
	}
	if !reflect.DeepEqual(registry, expectedRegistry()) {
		return registry, fmt.Errorf("language syntax registry mismatch")
	}
	return registry, nil
}

func CapabilityCaseTotal() int {
	return FixedCapabilityTotal
}

func validateCaseScopes(registry Registry) error {
	if len(registry.Cases) != FixedTotal || FixedCapabilityTotal+FixedGovernanceTotal != FixedTotal {
		return fmt.Errorf("language syntax scope denominator mismatch")
	}
	capability, governance := 0, 0
	governanceIDs, governancePaths := []string{}, []string{}
	for _, definition := range registry.Cases {
		switch definition.Scope {
		case ScopeLanguageCapability:
			capability++
		case ScopeGovernanceObservation:
			governance++
			governanceIDs = append(governanceIDs, definition.ID)
			governancePaths = append(governancePaths, definition.Path)
		default:
			return fmt.Errorf("language syntax case %q has missing or unknown scope", definition.ID)
		}
	}
	if capability != FixedCapabilityTotal || governance != FixedGovernanceTotal ||
		len(governanceIDs) != 1 || governanceIDs[0] != "live-governance-snapshot" ||
		len(governancePaths) != 1 || governancePaths[0] != "examples/live-governance-snapshot/main.gooo" {
		return fmt.Errorf("language syntax scope partition mismatch")
	}
	return nil
}

func unresolvedCases(source Source) []CaseResult {
	results := make([]CaseResult, 0, totalCases)
	for _, definition := range expectedRegistry().Cases {
		item := CaseResult{Definition: definition, Evidence: replay.Result{ObservedDecision: replay.DecisionUnknown}, Status: "UNRESOLVED"}
		item.EvidenceDigest = caseDigest(item, source)
		results = append(results, item)
	}
	return results
}
