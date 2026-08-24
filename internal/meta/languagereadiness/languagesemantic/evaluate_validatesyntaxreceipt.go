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
	if receipt.Summary.Satisfied != 19 || receipt.Summary.Total != 19 || receipt.Summary.ValidCases != 17 || receipt.Summary.InvalidCases != 2 || receipt.Summary.GoooLines != 262 {
		return fmt.Errorf("syntax evidence denominator does not match 18 cases / 16 files / 245 lines")
	}
	if len(receipt.Source.GoooFiles) != expectedSources {
		return fmt.Errorf("syntax evidence contains %d Gooo files, want %d", len(receipt.Source.GoooFiles), expectedSources)
	}
	if receipt.RepositoryWrites != 0 || receipt.MutationAuthorized {
		return fmt.Errorf("syntax evidence crossed the read-only effect boundary")
	}
	return nil
}
