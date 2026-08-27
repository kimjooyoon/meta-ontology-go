package denominatorevolution

import (
	"reflect"
	"sort"
)

func cloneObligations(value []Obligation) []Obligation {
	return append([]Obligation(nil), value...)
}

// Evaluate is the producer-side witness. Source parsing and lowering happen
// before any case is built; the JSON contract is used only as an expectation
// against the source-derived wire model.
func Evaluate(input Input) Report {
	contract := input.Contract
	snapshot := input.RepositorySnapshot
	report := Report{
		Schema: ReportSchema, Scope: ReportScope, HeadSHA: input.HeadSHA,
		Producer: contract.Producer, Consumer: contract.Consumer,
		ContractDigest: digestValue(contract), SourceDigest: DigestBytes(input.Source),
		NotClaimed: contract.NotClaimed, RepositoryWrites: snapshot.ChangedPaths,
		MutationAuthority: false, RepositorySnapshot: snapshot,
	}

	projection, wire, err := parseSource(input.Source)
	report.SourceProjection = projection
	if err != nil {
		return finishFailure(report, "GOOO_SOURCE_PROJECTION_UNKNOWN", "LOWER_RESOLUTION")
	}
	base, cases, ledger, emitted := sourceCaseInputs(wire, snapshot)
	report.Denominator = base
	report.ClaimLedger = ledger
	report.EmittedClaims = emitted
	for _, value := range cases {
		report.Cases = append(report.Cases, evaluateCase(value, base, contract.Policy))
	}
	report.Summary = summarize(report.Cases, base, len(wire.Cases), len(ledger), forbiddenEstimateObserved(emitted), snapshot.ChangedPaths)
	report.Indicators = makeIndicators(report.Summary)

	if !sourceContractMatches(wire, contract) {
		return finishFailure(report, "DENOMINATOR_EVOLUTION_CONTRACT_DRIFT", "INVARIANT_ONLY")
	}
	report.Decision, report.Resolution, report.Reason = "PASS", "EXACT", "DENOMINATOR_EVOLUTION_CONTRACT_SATISFIED"
	if report.Summary.CasesSatisfied != CaseCount || len(report.Indicators) != CheckCount || hasUnsatisfied(report.Indicators) {
		return finishFailure(report, "DENOMINATOR_EVOLUTION_CONTRACT_VIOLATED", "INVARIANT_ONLY")
	}
	report.Digest = reportDigest(report)
	return report
}

func sourceCaseInputs(wire sourceWire, snapshot RepositorySnapshot) (Denominator, []CaseInput, []ClaimLedgerEntry, []EmittedClaim) {
	base := Denominator{Version: wire.Version, Obligations: cloneObligations(wire.Obligations)}
	base.Digest = denominatorDigest(base)

	legal := cloneDenominator(base, wire.SuccessorVersion)
	for _, change := range wire.Deletions {
		legal.Obligations = removeObligation(legal.Obligations, change.ObligationID)
	}
	if mutation, ok := sourceMutationFor(wire, "legal-advance"); ok {
		legal.Obligations = append(legal.Obligations, mutation)
	}
	legal.Digest = denominatorDigest(legal)

	unauthorized := cloneDenominator(base, wire.SuccessorVersion)
	if mutation, ok := sourceMutationFor(wire, "unauthorized-change"); ok {
		unauthorized.Obligations = append(unauthorized.Obligations, mutation)
	}
	unauthorized.Digest = denominatorDigest(unauthorized)

	unknown := cloneDenominator(base, wire.UnknownPredecessorVersion)
	unknown.Digest = denominatorDigest(unknown)
	unknownSuccessor := cloneDenominator(base, wire.UnknownSuccessorVersion)
	unknownSuccessor.Digest = denominatorDigest(unknownSuccessor)

	ledger := sealClaimLedger(wire.Claims)
	emitted := append([]EmittedClaim(nil), wire.EmittedClaims...)
	guardrails := makeGuardrails(forbiddenEstimateObserved(emitted), snapshot.ChangedPaths)
	receipt := MigrationReceipt{
		Schema: ReceiptSchema, ID: wire.ReceiptID, Producer: "denominatorevolution.Evaluate", Consumer: "denominatorevolutionverify.Verify",
		Predecessor: DenominatorRef{Version: base.Version, Digest: base.Digest}, Successor: DenominatorRef{Version: legal.Version, Digest: legal.Digest},
		Additions: append([]Change(nil), wire.Additions...), Deletions: append([]Change(nil), wire.Deletions...),
		Decision: wire.ReceiptDecision, Reason: wire.ReceiptReason, Coordinate: wire.ReceiptCoordinate,
		RepositoryWrites: snapshot.ChangedPaths, MutationAuthority: false, RepositorySnapshot: snapshot, Guardrails: guardrails,
	}
	receipt.Digest = receiptDigest(receipt)
	inputs := []CaseInput{
		{Spec: wire.Cases[0].Spec, Predecessor: base, Successor: legal, Receipt: &receipt, Claim: ledger[0]},
		{Spec: wire.Cases[1].Spec, Predecessor: base, Successor: unauthorized, Claim: ledger[1]},
		{Spec: wire.Cases[2].Spec, Predecessor: unknown, Successor: unknownSuccessor, Claim: ledger[2]},
	}
	return base, inputs, ledger, emitted
}

func sourceMutationFor(wire sourceWire, caseID string) (Obligation, bool) {
	for _, value := range wire.Mutations {
		if value.CaseID == caseID {
			return value.Obligation, true
		}
	}
	return Obligation{}, false
}

func cloneDenominator(value Denominator, version string) Denominator {
	return Denominator{Version: version, Obligations: cloneObligations(value.Obligations)}
}

func removeObligation(values []Obligation, id string) []Obligation {
	result := make([]Obligation, 0, len(values))
	for _, value := range values {
		if value.ID != id {
			result = append(result, value)
		}
	}
	return result
}

func sealClaimLedger(values []ClaimLedgerEntry) []ClaimLedgerEntry {
	result := append([]ClaimLedgerEntry(nil), values...)
	previous := ""
	for index := range result {
		result[index].PreviousDigest = previous
		result[index].Digest = claimLedgerDigest(result[index])
		previous = result[index].Digest
	}
	return result
}

func sourceContractMatches(wire sourceWire, contract Contract) bool {
	if contract.Denominator.Version != wire.Version || !reflect.DeepEqual(contract.Denominator.Obligations, wire.Obligations) || len(contract.Cases) != len(wire.Cases) {
		return false
	}
	for index := range wire.Cases {
		if !reflect.DeepEqual(contract.Cases[index], wire.Cases[index].Spec) {
			return false
		}
	}
	return contract.Policy.NoAggregateEstimates && containsAll(contract.Policy.ForbiddenClaims, []string{"improvement rate", "aggregate estimate", "projected coverage"}) && containsAll(contract.Policy.AllowedAdditionReasons, []string{"NEW_MEASURABLE_OBLIGATION"}) && containsAll(contract.Policy.AllowedDeletionReasons, []string{"DEPRECATED_DUPLICATE"})
}

func containsAll(values, wanted []string) bool {
	for _, value := range wanted {
		found := false
		for _, candidate := range values {
			if candidate == value {
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

func evaluateCase(input CaseInput, base Denominator, policy MeasurementPolicy) CaseResult {
	spec := input.Spec
	predDigestValid := input.Predecessor.Digest == denominatorDigest(input.Predecessor)
	successorDigestValid := input.Successor.Digest == denominatorDigest(input.Successor)
	predKnown := predDigestValid && input.Predecessor.Version == base.Version && input.Predecessor.Digest == base.Digest && reflect.DeepEqual(input.Predecessor.Obligations, base.Obligations)
	additions, deletions := changes(base, input.Successor)
	additionsValid := reasonsValid(additions, input.Receipt, policy.AllowedAdditionReasons)
	deletionsValid := reasonsValid(deletions, input.Receipt, policy.AllowedDeletionReasons)
	receiptValid := validReceipt(input, additions, deletions, additionsValid && deletionsValid)
	decision, resolution, reason := "FAIL_CLOSED", "LOWER_RESOLUTION", "PREDECESSOR_DIGEST_UNKNOWN"
	if predKnown {
		decision, resolution, reason = "BLOCK", "INVARIANT_ONLY", "MIGRATION_RECEIPT_MISSING"
		if spec.Kind == "AUTHORIZED_MIGRATION" && receiptValid && !containsForbidden(input.Successor, policy.ForbiddenClaims) {
			decision, resolution, reason = "ADVANCE", "EXACT", "DENOMINATOR_ADVANCE_AUTHORIZED"
		}
	}
	fromClaim, toClaim := input.Claim.PriorState, input.Claim.NextState
	coordinate := Coordinate{Stage: spec.Stage, Step: spec.Step, Reason: reason}
	checks := []CheckResult{
		check("denominator-version", "FOUNDATION", "bind-fixed-denominator", spec, status(predKnown, "PASS", "UNKNOWN"), base.Version, input.Predecessor.Version),
		check("denominator-member-digest", "FOUNDATION", "bind-fixed-denominator", spec, status(successorDigestValid && predDigestValid, "PASS", "FAIL"), "valid source-derived digests", boolText(successorDigestValid && predDigestValid)),
		check("predecessor-registration", "FOUNDATION", "bind-predecessor-digest", spec, status(predKnown, "PASS", "UNKNOWN"), base.Digest, input.Predecessor.Digest),
		check("addition-reason", "COHERENCE", "classify-change-reason", spec, status(additionsValid, "PASS", "FAIL"), "every addition has an admissible reason", boolText(additionsValid)),
		check("deletion-reason", "COHERENCE", "classify-change-reason", spec, status(deletionsValid, "PASS", "FAIL"), "every deletion has an admissible reason", boolText(deletionsValid)),
		check("migration-receipt", "COHERENCE", "accept-migration-receipt", spec, receiptStatus(predKnown, receiptValid, input.Receipt), "receipt bound to both digests", receiptText(input.Receipt)),
		check("claim-transition", spec.ProofChoice, spec.MetaOperation, spec, status(decision == spec.ExpectedDecision && reason == spec.ExpectedReason && fromClaim == spec.FromClaim && toClaim == spec.ToClaim, "PASS", "FAIL"), spec.FromClaim+" -> "+spec.ToClaim, fromClaim+" -> "+toClaim),
		check("read-only-effect", "REGRESSION", "preserve-read-only-migration", spec, status(input.Receipt == nil || input.Receipt.RepositoryWrites == 0, "PASS", "FAIL"), "repository snapshot changed paths <= 0", repositoryWritesText(input.Receipt)),
	}
	caseResult := CaseResult{ID: spec.ID, Kind: spec.Kind, ExpectedDecision: spec.ExpectedDecision, ExpectedResolution: spec.ExpectedResolution, ExpectedReason: spec.ExpectedReason, ObservedDecision: decision, ObservedResolution: resolution, ObservedReason: reason, FromClaim: fromClaim, ToClaim: toClaim, Predecessor: input.Predecessor, Successor: input.Successor, Receipt: input.Receipt, Coordinate: coordinate, Checks: checks}
	caseResult.Status = "UNSATISFIED"
	if decision == spec.ExpectedDecision && resolution == spec.ExpectedResolution && reason == spec.ExpectedReason && fromClaim == spec.FromClaim && toClaim == spec.ToClaim && successorDigestValid {
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
