package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type dimensionSpec struct {
	ID       string
	MetricID string
	Evidence string
	Unit     string
	Activity string
}

var dimensions = []dimensionSpec{
	{ID: "declaration_coverage", MetricID: "gooo.metric.domain-completeness.declaration-coverage.v1", Evidence: "DeclarationEvidence", Unit: "declarations", Activity: "MeasureDeclarationCoverage"},
	{ID: "generation_coverage", MetricID: "gooo.metric.domain-completeness.generation-coverage.v1", Evidence: "GenerationEvidence", Unit: "use_cases", Activity: "MeasureGenerationCoverage"},
	{ID: "reverse_observation_coverage", MetricID: "gooo.metric.domain-completeness.reverse-observation-coverage.v1", Evidence: "ReverseObservationEvidence", Unit: "observations", Activity: "MeasureReverseObservationCoverage"},
	{ID: "use_case_coverage", MetricID: "gooo.metric.domain-completeness.use-case-coverage.v1", Evidence: "UseCaseEvidence", Unit: "use_cases", Activity: "MeasureUseCaseCoverage"},
	{ID: "boundary_coverage", MetricID: "gooo.metric.domain-completeness.boundary-coverage.v1", Evidence: "BoundaryEvidence", Unit: "boundaries", Activity: "MeasureBoundaryCoverage"},
	{ID: "provenance_integrity", MetricID: "gooo.metric.domain-completeness.provenance-integrity.v1", Evidence: "ProvenanceEvidence", Unit: "bindings", Activity: "MeasureProvenanceIntegrity"},
}

func compileProfile(path string, source []byte) (ProfileModel, error) {
	file, diagnostics := syntax.ParseFile(path, string(source))
	if file == nil || diagnostics.HasErrors() {
		return ProfileModel{}, fmt.Errorf("profile syntax has %d diagnostic(s)", len(diagnostics))
	}
	if file.Package == nil || file.Package.Name != "domaincompleteness" ||
		file.Namespace == nil || file.Namespace.Name != "domain_completeness" {
		return ProfileModel{}, fmt.Errorf("profile package or namespace is invalid")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return ProfileModel{}, fmt.Errorf("lower profile: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return ProfileModel{}, fmt.Errorf("validate profile: %w", err)
	}
	model := ProfileModel{
		Package: file.Package.Name, Namespace: file.Namespace.Name,
		Entities: make(map[string]Entity), Activities: make(map[string]Activity),
		SemanticHash: ir.StableHash(), SourceDigest: digestBytes(source),
	}
	for _, declaration := range file.Decls {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			if value.Name == "" || value.ID == "" {
				return ProfileModel{}, fmt.Errorf("profile entity has empty name or ID")
			}
			if _, exists := model.Entities[value.Name]; exists {
				return ProfileModel{}, fmt.Errorf("duplicate profile entity %q", value.Name)
			}
			model.Entities[value.Name] = Entity{Name: value.Name, ID: value.ID}
		case *syntax.ActivityDecl:
			if _, exists := model.Activities[value.Name]; exists {
				return ProfileModel{}, fmt.Errorf("duplicate profile activity %q", value.Name)
			}
			inputs := make([]string, len(value.Inputs))
			for index, input := range value.Inputs {
				inputs[index] = input.Name
			}
			model.Activities[value.Name] = Activity{
				Name: value.Name, Inputs: inputs, Output: value.Output,
				ValueProgram: value.ValueProgram,
			}
		}
	}
	if err := validateProfile(model); err != nil {
		return ProfileModel{}, err
	}
	return model, nil
}

func validateProfile(model ProfileModel) error {
	prefix := "gooo://meta/domain-completeness"
	expectedEntities := map[string]string{
		"BoundaryEvidence":                   prefix + "/evidence/boundary",
		"ComparisonBaseline":                 prefix + "/comparison/baseline",
		"ComparisonDelta":                    prefix + "/comparison/delta",
		"BoundaryCoverage":                   prefix + "/dimension/boundary-coverage",
		"DeclarationEvidence":                prefix + "/evidence/declaration",
		"DeclarationCoverage":                prefix + "/dimension/declaration-coverage",
		"DomainCompletenessReceipt":          ReceiptSchema,
		"DomainCompletenessVector":           prefix + "/vector",
		"DomainProfile":                      ProfileID,
		"ExcludedGeneralPurposeCompleteness": prefix + "/excluded/general-purpose-completeness",
		"ExcludedExternalAdoption":           prefix + "/excluded/external-adoption",
		"GenerationEvidence":                 prefix + "/evidence/generation",
		"GenerationCoverage":                 prefix + "/dimension/generation-coverage",
		"ProvenanceEvidence":                 prefix + "/evidence/provenance",
		"ProvenanceIntegrity":                prefix + "/dimension/provenance-integrity",
		"ReadOnlySystemBudget":               prefix + "/budget/read-only-zero-human-actions",
		"ReverseObservationEvidence":         prefix + "/evidence/reverse-observation",
		"ReverseObservationCoverage":         prefix + "/dimension/reverse-observation-coverage",
		"UseCaseEvidence":                    prefix + "/evidence/use-case",
		"UseCaseCoverage":                    prefix + "/dimension/use-case-coverage",
	}
	if len(model.Entities) != len(expectedEntities) {
		return fmt.Errorf("profile entity denominator is %d, want %d", len(model.Entities), len(expectedEntities))
	}
	for name, id := range expectedEntities {
		entity, exists := model.Entities[name]
		if !exists || entity.ID != id {
			return fmt.Errorf("profile entity %q must bind ID %q", name, id)
		}
	}
	expectedActivities := make(map[string]Activity)
	for _, dimension := range dimensions {
		expectedActivities[dimension.Activity] = Activity{
			Name: dimension.Activity, Inputs: []string{dimension.Evidence},
			Output: title(dimension.ID), ValueProgram: dimension.MetricID,
		}
	}
	expectedActivities["AssembleDomainCompletenessVector"] = Activity{
		Name:   "AssembleDomainCompletenessVector",
		Inputs: []string{"DeclarationCoverage", "GenerationCoverage", "ReverseObservationCoverage", "UseCaseCoverage", "BoundaryCoverage", "ProvenanceIntegrity"},
		Output: "DomainCompletenessVector", ValueProgram: "gooo.metric.domain-completeness.vector:v1",
	}
	expectedActivities["EmitDomainCompletenessReceipt"] = Activity{
		Name: "EmitDomainCompletenessReceipt", Inputs: []string{"DomainCompletenessVector"},
		Output: "DomainCompletenessReceipt", ValueProgram: "gooo.receipt.domain-completeness:v1",
	}
	expectedActivities["ReplayDomainCompletenessReceipt"] = Activity{
		Name: "ReplayDomainCompletenessReceipt", Inputs: []string{"DomainCompletenessReceipt"},
		Output: "DomainCompletenessReceipt", ValueProgram: "gooo.replay.domain-completeness:v1",
	}
	expectedActivities["CompareDomainCompletenessVectors"] = Activity{
		Name: "CompareDomainCompletenessVectors", Inputs: []string{"DomainCompletenessReceipt", "ComparisonBaseline"},
		Output: "ComparisonDelta", ValueProgram: "gooo.metric.domain-completeness.vector-delta.v1",
	}
	expectedActivities["FindPriorDomainCompletenessReceipt"] = Activity{
		Name: "FindPriorDomainCompletenessReceipt", Inputs: []string{"DomainProfile"},
		Output: "ComparisonBaseline", ValueProgram: "gooo.evidence.latest-compatible-domain-receipt.v1",
	}
	if len(model.Activities) != len(expectedActivities) {
		return fmt.Errorf("profile activity denominator is %d, want %d", len(model.Activities), len(expectedActivities))
	}
	for name, expected := range expectedActivities {
		actual, exists := model.Activities[name]
		if !exists || actual.Output != expected.Output || actual.ValueProgram != expected.ValueProgram ||
			!equalStrings(actual.Inputs, expected.Inputs) {
			return fmt.Errorf("profile activity %q does not match the receipt contract", name)
		}
	}
	return nil
}

func renderProfile(model ProfileModel) string {
	entities := make([]Entity, 0, len(model.Entities))
	for _, entity := range model.Entities {
		entities = append(entities, entity)
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i].Name < entities[j].Name })
	activities := make([]Activity, 0, len(model.Activities))
	for _, activity := range model.Activities {
		activities = append(activities, activity)
	}
	sort.Slice(activities, func(i, j int) bool { return activities[i].Name < activities[j].Name })
	var source strings.Builder
	fmt.Fprintf(&source, "package %s", model.Package)
	source.WriteByte('\n')
	fmt.Fprintf(&source, "namespace %s", model.Namespace)
	source.WriteByte('\n')
	source.WriteByte('\n')
	for _, entity := range entities {
		fmt.Fprintf(&source, "entity %s id %q", entity.Name, entity.ID)
		source.WriteByte('\n')
	}
	source.WriteByte('\n')
	for _, activity := range activities {
		fmt.Fprintf(&source, "activity %s(%s) -> %s computes %q",
			activity.Name, strings.Join(activity.Inputs, ", "), activity.Output, activity.ValueProgram)
		source.WriteByte('\n')
	}
	return source.String()
}

func semanticHash(path string, source []byte) (string, error) {
	file, diagnostics := syntax.ParseFile(path, string(source))
	if file == nil || diagnostics.HasErrors() {
		return "", fmt.Errorf("generated profile syntax has %d diagnostic(s)", len(diagnostics))
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return "", fmt.Errorf("lower generated profile: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return "", fmt.Errorf("validate generated profile: %w", err)
	}
	return ir.StableHash(), nil
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func title(value string) string {
	parts := strings.Split(value, "_")
	for index, part := range parts {
		if part == "" {
			continue
		}
		parts[index] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
