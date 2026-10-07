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
	dimensions, err := profileDimensions(model)
	if err != nil {
		return ProfileModel{}, err
	}
	model.Dimensions = dimensions
	if err := validateProfile(model); err != nil {
		return ProfileModel{}, err
	}
	return model, nil
}

func profileDimensions(model ProfileModel) ([]dimensionSpec, error) {
	vector, exists := model.Activities["AssembleDomainCompletenessVector"]
	if !exists || vector.Output != "DomainCompletenessVector" ||
		vector.ValueProgram != "gooo.metric.domain-completeness.vector:v1" {
		return nil, fmt.Errorf("profile vector activity is missing or invalid")
	}

	const profilePrefix = "gooo://meta/domain-completeness"
	result := make([]dimensionSpec, 0, len(vector.Inputs))
	seenOutputs := make(map[string]bool, len(vector.Inputs))
	seenActivities := make(map[string]bool, len(vector.Inputs))
	for _, output := range vector.Inputs {
		if seenOutputs[output] {
			return nil, fmt.Errorf("profile vector repeats dimension %q", output)
		}
		seenOutputs[output] = true
		entity, exists := model.Entities[output]
		if !exists || !strings.HasPrefix(entity.ID, profilePrefix+"/dimension/") {
			return nil, fmt.Errorf("profile vector input %q is not a declared dimension", output)
		}
		slug := strings.TrimPrefix(entity.ID, profilePrefix+"/dimension/")
		if slug == "" || strings.Contains(slug, "/") {
			return nil, fmt.Errorf("profile dimension %q has an invalid stable ID", output)
		}
		id := strings.ReplaceAll(slug, "-", "_")
		activityName := "Measure" + output
		activity, exists := model.Activities[activityName]
		if !exists || activity.Output != output || len(activity.Inputs) != 1 || seenActivities[activityName] {
			return nil, fmt.Errorf("profile dimension %q has no unique matching measurement activity", output)
		}
		metricID := activity.ValueProgram
		runtime, supported := dimensionRuntimes[metricID]
		if !supported {
			return nil, fmt.Errorf("profile dimension %q declares unsupported metric %q", output, metricID)
		}
		if metricID != "gooo.metric.domain-completeness."+slug+".v1" {
			return nil, fmt.Errorf("profile dimension %q metric does not match its stable ID", output)
		}
		seenActivities[activityName] = true
		evidence, exists := model.Entities[activity.Inputs[0]]
		stem := strings.TrimSuffix(strings.TrimSuffix(slug, "-coverage"), "-integrity")
		if !exists || evidence.ID != profilePrefix+"/evidence/"+stem {
			return nil, fmt.Errorf("profile measurement %q is not bound to its source evidence type", activityName)
		}
		result = append(result, dimensionSpec{
			ID: id, MetricID: metricID, Evidence: activity.Inputs[0], Unit: runtime.Unit,
		})
	}
	for name := range model.Activities {
		if strings.HasPrefix(name, "Measure") && !seenActivities[name] {
			return nil, fmt.Errorf("profile measurement %q is not connected to the vector", name)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("profile vector has no dimensions")
	}
	return result, nil
}

func validateProfile(model ProfileModel) error {
	prefix := "gooo://meta/domain-completeness"
	expectedEntities := map[string]string{
		"BoundaryEvidence":                   prefix + "/evidence/boundary",
		"ComparisonBaseline":                 prefix + "/comparison/baseline",
		"ComparisonDelta":                    prefix + "/comparison/delta",
		"BoundaryCoverage":                   prefix + "/dimension/boundary-coverage",
		"Boolean":                            prefix + "/type/boolean",
		"DeclarationEvidence":                prefix + "/evidence/declaration",
		"DeclarationCoverage":                prefix + "/dimension/declaration-coverage",
		"DomainCompletenessReceipt":          ReceiptSchema,
		"DomainCompletenessVector":           prefix + "/vector",
		"DomainProfile":                      ProfileID,
		"ExcludedGeneralPurposeCompleteness": prefix + "/excluded/general-purpose-completeness",
		"ExcludedExternalAdoption":           prefix + "/excluded/external-adoption",
		"Integer":                            prefix + "/type/integer",
		"GenerationEvidence":                 prefix + "/evidence/generation",
		"GenerationCoverage":                 prefix + "/dimension/generation-coverage",
		"ProvenanceEvidence":                 prefix + "/evidence/provenance",
		"ProvenanceIntegrity":                prefix + "/dimension/provenance-integrity",
		"ReadOnlySystemBudget":               prefix + "/budget/read-only-zero-human-actions",
		"ReverseObservationEvidence":         prefix + "/evidence/reverse-observation",
		"ReverseObservationCoverage":         prefix + "/dimension/reverse-observation-coverage",
		"Text":                               prefix + "/type/text",
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
	expectedDimensions := 0
	for _, id := range expectedEntities {
		if strings.HasPrefix(id, prefix+"/dimension/") {
			expectedDimensions++
		}
	}
	if len(model.Dimensions) != expectedDimensions {
		return fmt.Errorf("profile vector dimension denominator is %d, want %d", len(model.Dimensions), expectedDimensions)
	}
	expectedActivities := map[string]Activity{
		"EmitDomainCompletenessReceipt": {
			Name: "EmitDomainCompletenessReceipt", Inputs: []string{"DomainCompletenessVector"},
			Output: "DomainCompletenessReceipt", ValueProgram: "gooo.receipt.domain-completeness:v1",
		},
		"ReplayDomainCompletenessReceipt": {
			Name: "ReplayDomainCompletenessReceipt", Inputs: []string{"DomainCompletenessReceipt"},
			Output: "DomainCompletenessReceipt", ValueProgram: "gooo.replay.domain-completeness:v1",
		},
		"CompareDomainCompletenessVectors": {
			Name: "CompareDomainCompletenessVectors", Inputs: []string{"DomainCompletenessReceipt", "ComparisonBaseline"},
			Output: "ComparisonDelta", ValueProgram: "gooo.metric.domain-completeness.vector-delta.v1",
		},
		"ClassifyDomainCompleteness": {
			Name: "ClassifyDomainCompleteness", Inputs: []string{"Integer", "Integer", "Integer", "Boolean"},
			Output: "Text",
		},
		"FindPriorDomainCompletenessReceipt": {
			Name: "FindPriorDomainCompletenessReceipt", Inputs: []string{"DomainProfile"},
			Output: "ComparisonBaseline", ValueProgram: "gooo.evidence.latest-compatible-domain-receipt.v1",
		},
		"AssembleDomainCompletenessVector": model.Activities["AssembleDomainCompletenessVector"],
	}
	if len(model.Activities) != len(expectedActivities)+len(model.Dimensions) {
		return fmt.Errorf("profile activity denominator is %d, want %d", len(model.Activities), len(expectedActivities)+len(model.Dimensions))
	}
	for name, expected := range expectedActivities {
		actual, exists := model.Activities[name]
		programMismatch := actual.ValueProgram != expected.ValueProgram
		if name == "ClassifyDomainCompleteness" {
			programMismatch = strings.TrimSpace(actual.ValueProgram) == ""
		}
		if !exists || actual.Output != expected.Output || programMismatch ||
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
