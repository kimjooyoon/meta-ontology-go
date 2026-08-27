package audienceresolution

import (
	"fmt"
	"strings"
)

// ValidateReceipt only checks producer shape. It is intentionally not the
// independent evidence boundary: the consumer reopens source and artifacts.
func ValidateReceipt(receipt Receipt) error {
	if receipt.Schema != ReceiptSchema || receipt.ContractID == "" || receipt.Subject == "" || !receipt.Provisional {
		return fmt.Errorf("receipt identity or provisional state is invalid")
	}
	if receipt.Decision != "PASS" && receipt.Decision != "UNKNOWN" && receipt.Decision != "REFUTED" {
		return fmt.Errorf("receipt decision is unknown")
	}
	if (receipt.Resolution != "EXACT" && receipt.Resolution != "LOWER_RESOLUTION" && receipt.Resolution != "INVARIANT_ONLY") ||
		receipt.Reason == "" || !validDigest(receipt.Digest) || !validDigest(receipt.FactsDigest) {
		return fmt.Errorf("receipt resolution or digest is invalid")
	}
	copy := receipt
	copy.Digest = ""
	if receiptDigest(copy) != receipt.Digest {
		return fmt.Errorf("receipt digest mismatch")
	}
	if len(receipt.Indicators) != len(indicatorSpecs()) || len(receipt.Views) != 3 || len(receipt.Counterexamples) != 2 || len(receipt.ClaimTransitions) != 36 {
		return fmt.Errorf("receipt cardinalities are invalid")
	}
	if receipt.Summary.Coordinates.Total <= 0 || receipt.Summary.Coordinates.Satisfied < 0 ||
		receipt.Summary.Coordinates.Satisfied > receipt.Summary.Coordinates.Total ||
		receipt.Summary.Coordinates.BasisPoints != basisPoints(receipt.Summary.Coordinates.Satisfied, receipt.Summary.Coordinates.Total) ||
		receipt.Summary.SourceDenominator <= 0 || receipt.Summary.DistinctPropositions <= 0 {
		return fmt.Errorf("receipt summary is invalid")
	}
	if err := validateIndicators(receipt.Indicators); err != nil {
		return err
	}
	if err := validateViews(receipt); err != nil {
		return err
	}
	if err := validateCounterexamples(receipt.Counterexamples); err != nil {
		return err
	}
	previous := digestBytes([]byte("gooo://audience-resolution/claim-event/genesis"))
	for _, transition := range receipt.ClaimTransitions {
		if transition.IndicatorID == "" || transition.Audience == "" || transition.Proposition == "" || transition.PropositionDigest == "" ||
			transition.TargetAddress == "" || transition.Before != "OPEN" || (transition.After != "OPEN" && transition.After != "DISCHARGED" && transition.After != "REFUTED") ||
			(transition.Visibility != "VISIBLE" && transition.Visibility != "OMITTED") || !validDigest(transition.EvidenceDigest) ||
			!validDigest(transition.EventDigest) || transition.PreviousEventDigest != previous || transition.Producer == "" || transition.Consumer == "" ||
			transition.Stage == "" || transition.Step == "" || transition.Reason == "" {
			return fmt.Errorf("claim transition %q is incomplete", transition.IndicatorID)
		}
		if claimEventDigest(transition) != transition.EventDigest {
			return fmt.Errorf("claim transition %q digest mismatch", transition.IndicatorID)
		}
		previous = transition.EventDigest
	}
	return nil
}

func validateIndicators(values []Indicator) error {
	specs := indicatorSpecs()
	for index, value := range values {
		spec := specs[index]
		metadataMatches := value.Producer == spec.Producer && value.Consumer == spec.Consumer && value.MetaOperation == spec.MetaOperation &&
			value.ProofChoice == spec.ProofChoice && value.Stage == spec.Stage && value.Step == spec.Step && value.Reason == spec.Reason
		if value.EvidenceStatus == EvidenceUnknown && value.Observed == 0 {
			metadataMatches = value.Producer != "" && value.Consumer != "" && value.MetaOperation != "" && value.ProofChoice != "" && value.Stage != "" && value.Step != "" && value.Reason != ""
		}
		if value.ID != spec.ID || value.Class != spec.Class || !metadataMatches || value.ClaimBefore != "OPEN" || value.Expected != 1 || value.Observed < 0 || value.Observed > 1 ||
			(value.Satisfied && (value.Observed != 1 || value.ClaimAfter != "DISCHARGED")) || (!value.Satisfied && value.Observed == 1) {
			return fmt.Errorf("indicator %q is not canonical", value.ID)
		}
		if value.ClaimAfter != "OPEN" && value.ClaimAfter != "DISCHARGED" && value.ClaimAfter != "REFUTED" {
			return fmt.Errorf("indicator %q has invalid claim state", value.ID)
		}
	}
	return nil
}

func validateViews(receipt Receipt) error {
	if receipt.Views[0].Audience != "USER" || receipt.Views[1].Audience != "TOOL_AUTHOR" || receipt.Views[2].Audience != "GOVERNOR" {
		return fmt.Errorf("audience order is not canonical")
	}
	governor := receipt.Views[2]
	for index, view := range receipt.Views {
		if view.Resolution == "" || view.Required != governor.Required || view.Total != len(view.CoordinateIDs) || view.Visible < 0 || view.Visible > view.Required ||
			view.OmittedCoordinateCount != view.Required-view.Visible || view.Satisfied < 0 || view.Satisfied > view.Total ||
			view.BasisPoints != basisPoints(view.Satisfied, view.Total) || view.GlobalDecision != receipt.Decision || view.LocalDecision == "" ||
			view.LocalResolution == "" || view.LocalReason == "" || (view.InheritedStatus != "LOCALLY_VERIFIED" && view.InheritedStatus != "INHERITED_NOT_LOCALLY_VERIFIED") ||
			view.SubjectRequired != receipt.Summary.Coordinates.Total || view.SubjectSatisfied < 0 || view.SubjectSatisfied > view.SubjectRequired {
			return fmt.Errorf("audience view %q is not canonical", view.Audience)
		}
		if view.LocalDecision == "PASS" && view.SubjectSatisfied != view.SubjectRequired {
			return fmt.Errorf("audience view %q claims PASS without local subject evidence", view.Audience)
		}
		if view.LocalDecision == "UNKNOWN" && view.SubjectSatisfied == view.SubjectRequired && view.GlobalDecision != "REFUTED" {
			return fmt.Errorf("audience view %q claims UNKNOWN despite complete local subject evidence", view.Audience)
		}
		if index > 0 && !prefix(receipt.Views[index-1].CoordinateIDs, view.CoordinateIDs) {
			return fmt.Errorf("audience coordinate policy is not nested")
		}
	}
	return nil
}

func validateCounterexamples(values []CounterexampleResult) error {
	seen := map[string]bool{}
	for _, value := range values {
		if seen[value.ID] || value.ID == "" || value.Global == "PASS" || !value.ExecutionValidated || len(value.Views) != 3 ||
			value.Proposition == "" || !validDigest(value.PropositionDigest) || value.TargetAddress == "" || value.Stage == "" || value.Step == "" || value.Reason == "" ||
			value.BeforeClaim != "OPEN" || (value.AfterClaim != "OPEN" && value.AfterClaim != "REFUTED") || !validDigest(value.ContentDigest) {
			return fmt.Errorf("counterexample %q is not executed", value.ID)
		}
		seen[value.ID] = true
	}
	if !seen["counterexample.missing-information"] || !seen["counterexample.decision-contradiction"] {
		return fmt.Errorf("counterexample execution set is incomplete")
	}
	return nil
}

func prefix(shorter, longer []string) bool {
	if len(longer) < len(shorter) {
		return false
	}
	for index, value := range shorter {
		if value != longer[index] {
			return false
		}
	}
	return true
}

func validDigest(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	for _, character := range value[7:] {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}
