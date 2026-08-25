package languagesemantic

import (
	"fmt"
)

func validateSyntaxReceipt(receipt syntaxReceipt, expectedHead string) error {
	if receipt.Schema != "gooo/language-syntax-roundtrip/v1" {
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
	if receipt.Summary.Satisfied != expectedSyntaxCases || receipt.Summary.Total != expectedSyntaxCases ||
		receipt.Summary.ValidCases != expectedSyntaxValid || receipt.Summary.InvalidCases != expectedSyntaxInvalid ||
		receipt.Summary.GoooLines != expectedSyntaxLines {
		return fmt.Errorf("syntax evidence denominator does not match %d cases / %d valid / %d invalid / %d lines",
			expectedSyntaxCases, expectedSyntaxValid, expectedSyntaxInvalid, expectedSyntaxLines)
	}
	if len(receipt.Source.GoooFiles) != expectedSyntaxFiles {
		return fmt.Errorf("syntax evidence contains %d Gooo files, want %d",
			len(receipt.Source.GoooFiles), expectedSyntaxFiles)
	}
	if err := validateSyntaxPackages(receipt.Source.PackageUnits); err != nil {
		return err
	}
	if err := validateSyntaxCases(receipt.Cases); err != nil {
		return err
	}
	if receipt.RepositoryWrites != 0 || receipt.MutationAuthorized {
		return fmt.Errorf("syntax evidence crossed the read-only effect boundary")
	}
	return nil
}

func validateSyntaxPackages(units []syntaxPackageUnit) error {
	if len(units) != expectedPackageUnits {
		return fmt.Errorf("syntax evidence contains %d package units, want %d", len(units), expectedPackageUnits)
	}
	unit := units[0]
	if unit.ID != "billing-package" || unit.Path != "examples/billing-package" || unit.Entry != "PayOrder" ||
		unit.ReportSchema != "gooo/language-package-execution-report/v1" ||
		unit.MetaReducer != "languagepackageexecution.Evaluate" ||
		unit.SourceFilesIndicator != "PACKAGE_SOURCE_FILES" || unit.ExecutionIndicator != "PACKAGE_EXECUTIONS" ||
		len(unit.Members) != expectedPackageFiles || unit.Members[0] != "examples/billing-package/activity.gooo" ||
		unit.Members[1] != "examples/billing-package/entities.gooo" {
		return fmt.Errorf("syntax evidence package-unit binding is not canonical")
	}
	return nil
}
