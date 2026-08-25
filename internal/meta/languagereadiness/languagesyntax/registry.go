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
		return CaseDefinition{ID: id, Path: path, Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax"}
	}
	invalid := func(id, path, diagnostic string) CaseDefinition {
		return CaseDefinition{ID: id, Path: path, Kind: KindInvalid, ExpectedDecision: DecisionClosed,
			ExpectedDiagnostic: diagnostic, ProofChoice: "REGRESSION", MetaOperation: "reject-invalid-syntax"}
	}
	packageUnit := PackageDefinition{ID: "billing-package", Path: "examples/billing-package",
		Members: []string{"examples/billing-package/activity.gooo", "examples/billing-package/entities.gooo"},
		Entry:   "PayOrder", ReportSchema: languagepackageexecution.ReportSchema,
		MetaReducer: "languagepackageexecution.Evaluate", SourceFilesIndicator: "PACKAGE_SOURCE_FILES",
		ExecutionIndicator: "PACKAGE_EXECUTIONS"}
	return Registry{Schema: RegistrySchema, Cases: []CaseDefinition{
		valid("billing", "examples/billing/main.gooo"),
		valid("bootstrap", "examples/bootstrap/main.gooo"),
		valid("conformance", "examples/conformance/main.gooo"),
		valid("meta-actionability", "examples/meta-actionability/main.gooo"),
		valid("meta-binding-coverage", "examples/meta-binding-coverage/main.gooo"),
		valid("meta-operation-artifact-coverage", "examples/meta-operation-artifact-coverage/main.gooo"),
		valid("metric-meta-program-closure", "examples/metric-meta-program-closure/main.gooo"),
		valid("metric-meta-program", "examples/metric-meta-program/main.gooo"),
		valid("root-readme-indicator", "examples/root-readme-indicator/main.gooo"),
		valid("self-improvement", "examples/self-improvement/main.gooo"),
		valid("roundtrip-minimal", "internal/detection/roundtrip/testdata/minimal.gooo"),
		valid("directory-kind-ontology", "internal/meta/directorykind/ontology.gooo"),
		valid("directory-partition-ontology", "internal/meta/directorypartition/ontology.gooo"),
		invalid("unknown-keyword", "examples/language-syntax-roundtrip/unknown-keyword.txt", "parse.unexpected-token"),
		invalid("unterminated-string", "examples/language-syntax-roundtrip/unterminated-string.txt", "lex.unterminated-string"),
		invalid("source-execution-invalid", "examples/language-source-execution/invalid.gooo", "parse.unexpected-token"),
		valid("source-splitter-conformance", "examples/source-splitter-conformance/main.gooo"),
		valid("rollback-integrity-activation", "examples/rollback-integrity-activation/main.gooo"),
		valid("vertical-slice-closure-activation", "examples/vertical-slice-closure-activation/main.gooo"),
		valid("external-conformance-activation", "examples/external-conformance-activation/main.gooo"),
	}, PackageUnits: []PackageDefinition{packageUnit}}
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
	if !reflect.DeepEqual(registry, expectedRegistry()) {
		return registry, fmt.Errorf("language syntax registry mismatch")
	}
	return registry, nil
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
