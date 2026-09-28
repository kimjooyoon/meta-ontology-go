package languagesemantic

import (
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func validateSyntaxReceipt(receipt syntaxReceipt, expectedHead, root string, raw []byte) error {
	if receipt.Schema != languagesyntax.ReportSchema {
		return fmt.Errorf("syntax evidence schema is unknown")
	}
	if receipt.Source.ExpectedHeadSHA != expectedHead {
		return fmt.Errorf("syntax evidence head does not match the semantic subject")
	}
	if receipt.Decision != "PASS" || receipt.Resolution != "EXACT" {
		return fmt.Errorf("syntax evidence decision is not explicit PASS / EXACT")
	}
	if !receipt.Source.ObservationKnown || !receipt.Source.ConceptBound {
		return fmt.Errorf("syntax evidence is not dynamically bound")
	}
	upstream, err := validateUpstreamSyntaxArtifact(raw, expectedHead, root)
	if err != nil {
		return err
	}
	if len(receipt.Cases) != len(upstream.Cases) {
		return fmt.Errorf("syntax evidence case projection contains %d cases, want %d",
			len(receipt.Cases), len(upstream.Cases))
	}
	if err := validateSyntaxPackages(receipt.Source.PackageUnits); err != nil {
		return err
	}
	if err := validateSyntaxCases(receipt.Cases); err != nil {
		return err
	}
	if err := validateSyntaxCaseProjection(receipt.Cases, upstream.Cases); err != nil {
		return err
	}
	if err := validateSyntaxPackageProjection(receipt.Source.PackageUnits, upstream.Source.PackageUnits); err != nil {
		return err
	}
	if receipt.RepositoryWrites != 0 || receipt.MutationAuthorized || upstream.RepositoryWrites != 0 ||
		upstream.MutationAuthorized || upstream.Source.ConceptRepositoryWrites != 0 {
		return fmt.Errorf("syntax evidence crossed the read-only effect boundary")
	}
	return nil
}

func validateSyntaxPackages(units []syntaxPackageUnit) error {
	seen := make(map[string]struct{}, len(units))
	for _, unit := range units {
		if unit.ID == "" || unit.Path == "" {
			return fmt.Errorf("syntax evidence package-unit identity is incomplete")
		}
		if _, exists := seen[unit.ID]; exists {
			return fmt.Errorf("syntax evidence package-unit identity is duplicated")
		}
		seen[unit.ID] = struct{}{}
	}
	return nil
}

func validateSyntaxCaseProjection(cases []syntaxCase, upstream []languagesyntax.CaseResult) error {
	if len(cases) != len(upstream) {
		return fmt.Errorf("syntax evidence case projection is incomplete")
	}
	for index := range upstream {
		got, want := cases[index].Definition, upstream[index].Definition
		if got.ID != want.ID || got.Path != want.Path || got.Kind != want.Kind {
			return fmt.Errorf("syntax evidence case projection %d is not canonical", index)
		}
	}
	return nil
}

func validateSyntaxPackageProjection(units []syntaxPackageUnit, upstream []languagesyntax.PackageDefinition) error {
	if len(units) != len(upstream) {
		return fmt.Errorf("syntax evidence package-unit projection contains %d units, want %d", len(units), len(upstream))
	}
	for index := range upstream {
		got, want := units[index], upstream[index]
		if got.ID != want.ID || got.Path != want.Path || !reflect.DeepEqual(got.Members, want.Members) ||
			got.Entry != want.Entry || got.ReportSchema != want.ReportSchema || got.MetaReducer != want.MetaReducer ||
			got.SourceFilesIndicator != want.SourceFilesIndicator || got.ExecutionIndicator != want.ExecutionIndicator {
			return fmt.Errorf("syntax evidence package-unit projection %d is not canonical", index)
		}
	}
	return nil
}
