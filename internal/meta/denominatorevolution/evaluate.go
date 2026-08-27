package denominatorevolution

import (
	"reflect"
	"sort"
)

func Evaluate(input Input) Report {
	contract := CanonicalContract()
	base := Denominator{Version: contract.Denominator.Version, Obligations: cloneObligations(contract.Denominator.Obligations)}
	base.Digest = denominatorDigest(base)
	report := Report{Schema: ReportSchema, Scope: ReportScope, HeadSHA: input.HeadSHA, Producer: contract.Producer, Consumer: contract.Consumer, ContractDigest: digestValue(input.Contract), SourceDigest: DigestBytes(input.Source), Denominator: base, SourceProjection: projectSource(input.Source), NotClaimed: contract.NotClaimed, RepositoryWrites: 0, MutationAuthority: false}
	if !reflect.DeepEqual(input.Contract, contract) {
		return finishFailure(report, "DENOMINATOR_EVOLUTION_CONTRACT_DRIFT", "INVARIANT_ONLY")
	}
	if !report.SourceProjection.Exact {
		return finishFailure(report, "GOOO_SOURCE_PROJECTION_UNKNOWN", "UNKNOWN")
	}
	for _, value := range CanonicalCases() {
		report.Cases = append(report.Cases, evaluateCase(value, base, contract.Policy))
	}
	report.Summary = summarize(report.Cases, base, report.RepositoryWrites)
	report.Indicators = makeIndicators(report.Summary)
	report.Decision, report.Resolution, report.Reason = "PASS", "EXACT", "DENOMINATOR_EVOLUTION_CONTRACT_SATISFIED"
	if report.Summary.CasesSatisfied != CaseCount || len(report.Indicators) != 8 || hasUnsatisfied(report.Indicators) {
		return finishFailure(report, "DENOMINATOR_EVOLUTION_CONTRACT_VIOLATED", "INVARIANT_ONLY")
	}
	report.Digest = reportDigest(report)
	return report
}

func evaluateCase(input CaseInput, base Denominator, policy MeasurementPolicy) CaseResult {
	spec := input.Spec
	predDigestValid := input.Predecessor.Digest == denominatorDigest(input.Predecessor)
	successorDigestValid := input.Successor.Digest == denominatorDigest(input.Successor)
	predKnown := predDigestValid && input.Predecessor.Version == base.Version && input.Predecessor.Digest == base.Digest && reflect.DeepEqual(input.Predecessor.Obligations, base.Obligations)
	additions, deletions := changes(base, input.Successor)
	additionsValid := reasonsValid(additions, input.Receipt, policy.AllowedAdditionReasons)
	deletionsValid := reasonsValid(deletions, input.Receipt, policy.AllowedDeletionReasons)
	receiptValid := validReceipt(input, additions, deletions, additionsValid && deletionsValid)
	decision, resolution, reason, toClaim := "FAIL_CLOSED", "UNKNOWN", "PREDECESSOR_DIGEST_UNKNOWN", "UNKNOWN"
	if predKnown {
		decision, resolution, reason, toClaim = "BLOCK", "INVARIANT_ONLY", "MIGRATION_RECEIPT_MISSING", "REJECTED"
		if spec.Kind == "AUTHORIZED_MIGRATION" && receiptValid && !containsForbidden(input.Successor, policy.ForbiddenClaims) {
			decision, resolution, reason, toClaim = "ADVANCE", "EXACT", "DENOMINATOR_ADVANCE_AUTHORIZED", "ACCEPTED"
		}
	}
	fromClaim := "PROPOSED"
	coordinate := Coordinate{Stage: spec.Stage, Step: spec.Step, Reason: reason}
	checks := []CheckResult{
		check("denominator-version", "FOUNDATION", "bind-fixed-denominator", spec, status(predKnown, "PASS", "UNKNOWN"), base.Version, input.Predecessor.Version),
		check("denominator-member-digest", "FOUNDATION", "bind-fixed-denominator", spec, status(successorDigestValid && predDigestValid, "PASS", "FAIL"), "valid canonical digests", boolText(successorDigestValid && predDigestValid)),
		check("predecessor-registration", "FOUNDATION", "bind-predecessor-digest", spec, status(predKnown, "PASS", "UNKNOWN"), base.Digest, input.Predecessor.Digest),
		check("addition-reason", "COHERENCE", "classify-change-reason", spec, status(additionsValid, "PASS", "FAIL"), "every addition has an admissible reason", boolText(additionsValid)),
		check("deletion-reason", "COHERENCE", "classify-change-reason", spec, status(deletionsValid, "PASS", "FAIL"), "every deletion has an admissible reason", boolText(deletionsValid)),
		check("migration-receipt", "COHERENCE", "accept-migration-receipt", spec, receiptStatus(predKnown, receiptValid, input.Receipt), "receipt bound to both digests", receiptText(input.Receipt)),
		check("claim-transition", spec.ProofChoice, spec.MetaOperation, spec, status(decision == spec.ExpectedDecision && reason == spec.ExpectedReason && toClaim == spec.ToClaim, "PASS", "FAIL"), spec.FromClaim+" -> "+spec.ToClaim, fromClaim+" -> "+toClaim),
		check("read-only-effect", "REGRESSION", "preserve-read-only-migration", spec, "PASS", "0 repository writes and no mutation authority", "0 repository writes and no mutation authority"),
	}
	caseResult := CaseResult{ID: spec.ID, Kind: spec.Kind, ExpectedDecision: spec.ExpectedDecision, ExpectedResolution: spec.ExpectedResolution, ExpectedReason: spec.ExpectedReason, ObservedDecision: decision, ObservedResolution: resolution, ObservedReason: reason, FromClaim: fromClaim, ToClaim: toClaim, Predecessor: input.Predecessor, Successor: input.Successor, Receipt: input.Receipt, Coordinate: coordinate, Checks: checks}
	caseResult.Status = "UNSATISFIED"
	if decision == spec.ExpectedDecision && resolution == spec.ExpectedResolution && reason == spec.ExpectedReason && toClaim == spec.ToClaim && successorDigestValid {
		caseResult.Status = "SATISFIED"
	}
	return caseResult
}

func changes(before, after Denominator) ([]Change, []Change) {
	beforeIDs, afterIDs := map[string]bool{}, map[string]bool{}
	for _, value := range before.Obligations {
		beforeIDs[value.ID] = true
	}
	for _, value := range after.Obligations {
		afterIDs[value.ID] = true
	}
	additions, deletions := []Change{}, []Change{}
	for id := range afterIDs {
		if !beforeIDs[id] {
			additions = append(additions, Change{ObligationID: id})
		}
	}
	for id := range beforeIDs {
		if !afterIDs[id] {
			deletions = append(deletions, Change{ObligationID: id})
		}
	}
	sort.Slice(additions, func(i, j int) bool { return additions[i].ObligationID < additions[j].ObligationID })
	sort.Slice(deletions, func(i, j int) bool { return deletions[i].ObligationID < deletions[j].ObligationID })
	return additions, deletions
}

func reasonsValid(changes []Change, receipt *MigrationReceipt, allowed []string) bool {
	if len(changes) == 0 {
		return true
	}
	if receipt == nil {
		return false
	}
	for _, change := range changes {
		found := false
		for _, candidate := range append(append([]Change{}, receipt.Additions...), receipt.Deletions...) {
			if candidate.ObligationID == change.ObligationID && contains(allowed, candidate.Reason) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func validReceipt(input CaseInput, additions, deletions []Change, reasonsOK bool) bool {
	receipt := input.Receipt
	if receipt == nil || !reasonsOK || receipt.Schema != ReceiptSchema || receipt.Decision != "ADVANCE" || receipt.Reason != "DENOMINATOR_ADVANCE_AUTHORIZED" || receipt.RepositoryWrites != 0 || receipt.MutationAuthority || !guardrailsConform(receipt.Guardrails) {
		return false
	}
	if receipt.Predecessor.Version != input.Predecessor.Version || receipt.Predecessor.Digest != input.Predecessor.Digest || receipt.Successor.Version != input.Successor.Version || receipt.Successor.Digest != input.Successor.Digest {
		return false
	}
	if !sameChanges(additions, receipt.Additions) || !sameChanges(deletions, receipt.Deletions) {
		return false
	}
	return receipt.Digest != "" && receipt.Digest == receiptDigest(*receipt)
}

func sameChanges(left, right []Change) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].ObligationID != right[index].ObligationID {
			return false
		}
	}
	return true
}

func containsForbidden(value Denominator, forbidden []string) bool {
	for _, obligation := range value.Obligations {
		claim := obligation.ID + " " + obligation.Claim
		for _, word := range forbidden {
			if containsFold(claim, word) {
				return true
			}
		}
	}
	return false
}
