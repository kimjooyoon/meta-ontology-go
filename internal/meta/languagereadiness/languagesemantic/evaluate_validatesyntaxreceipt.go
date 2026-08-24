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
	if receipt.Summary.Satisfied != 17 || receipt.Summary.Total != 17 || receipt.Summary.ValidCases != 15 || receipt.Summary.InvalidCases != 2 || receipt.Summary.GoooLines != 225 {
		return fmt.Errorf("syntax evidence denominator does not match 17 cases / 15 files / 225 lines")
	}
	if len(receipt.Source.GoooFiles) != expectedSources {
		return fmt.Errorf("syntax evidence contains %d Gooo files, want %d", len(receipt.Source.GoooFiles), expectedSources)
	}
	if receipt.RepositoryWrites != 0 || receipt.MutationAuthorized {
		return fmt.Errorf("syntax evidence crossed the read-only effect boundary")
	}
	return nil
}
