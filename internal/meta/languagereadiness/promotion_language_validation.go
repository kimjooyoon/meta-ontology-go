package languagereadiness

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagediagnosticprovenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func validateLanguageEvidence(bundle PromotionEvidence, expectedHeadSHA string) error {
	if err := languagesyntax.Validate(bundle.Syntax, expectedHeadSHA); err != nil {
		return fmt.Errorf("verify language syntax roundtrip: %w", err)
	}
	if bundle.Syntax.Decision != languagesyntax.DecisionPass {
		return fmt.Errorf("FAIL_CLOSED: language syntax decision %q", bundle.Syntax.Decision)
	}
	if err := languagediagnosticprovenance.Validate(bundle.Diagnostic, expectedHeadSHA); err != nil {
		return fmt.Errorf("verify diagnostic provenance: %w", err)
	}
	return nil
}
