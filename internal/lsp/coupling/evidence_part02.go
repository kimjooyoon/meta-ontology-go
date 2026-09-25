package coupling

import (
	"fmt"
)

func validateExplanation(envelope Envelope, explanation Explanation) error {
	if err := validateIdentity(explanation.CodeSymbolID, "code symbol"); err != nil {
		return fmt.Errorf("code symbol ID: %w", err)
	}
	if err := validateIdentity(explanation.SemanticOwnerID, "semantic owner"); err != nil {
		return fmt.Errorf("semantic owner ID: %w", err)
	}
	if err := validateBoundLocation(explanation.Origin); err != nil {
		return fmt.Errorf("origin: %w", err)
	}
	if explanation.Origin.URI != envelope.Document.URI {
		return fmt.Errorf("origin URI does not match document URI")
	}
	if err := validateBoundLocation(explanation.Target); err != nil {
		return fmt.Errorf("target: %w", err)
	}
	if !explanation.Claim.valid() {
		return fmt.Errorf("invalid change claim %q", explanation.Claim)
	}
	if !explanation.Status.valid() {
		return fmt.Errorf("invalid explanation status %q", explanation.Status)
	}
	if explanation.Status == OutcomePass && explanation.Reason != "" {
		return fmt.Errorf("PASS explanation cannot have a reason")
	}
	if explanation.Status != OutcomePass && !explanation.Reason.valid() {
		return fmt.Errorf("non-PASS explanation requires a known reason")
	}
	if explanation.Status == OutcomePass && len(explanation.CausalSpans) == 0 {
		return fmt.Errorf("PASS explanation requires causal spans")
	}
	seen := make(map[string]struct{}, len(explanation.CausalSpans))
	for index, span := range explanation.CausalSpans {
		if err := validateIdentity(span.StableID, "causal span"); err != nil {
			return fmt.Errorf("causal span %d ID: %w", index, err)
		}
		if err := validateSourceMapBinding(span.SourceMapID, span.SourceMapDigest); err != nil {
			return fmt.Errorf("causal span %d source map: %w", index, err)
		}
		if !exactText(span.URI) {
			return fmt.Errorf("causal span %d URI is required", index)
		}
		if err := validateRange(span.Range); err != nil {
			return fmt.Errorf("causal span %d: %w", index, err)
		}
		if span.Ordinal < 0 {
			return fmt.Errorf("causal span %d has negative ordinal", index)
		}
		if _, exists := seen[span.StableID]; exists {
			return fmt.Errorf("duplicate causal span %q", span.StableID)
		}
		seen[span.StableID] = struct{}{}
	}
	return nil
}
func validateBoundLocation(location BoundLocation) error {
	if err := validateIdentity(location.StableID, "location"); err != nil {
		return fmt.Errorf("stable ID: %w", err)
	}
	if err := validateSourceMapBinding(location.SourceMapID, location.SourceMapDigest); err != nil {
		return err
	}
	if !exactText(location.URI) {
		return fmt.Errorf("URI is required")
	}
	return validateRange(location.Range)
}
