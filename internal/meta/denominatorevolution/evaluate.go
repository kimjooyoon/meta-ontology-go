package denominatorevolution

import (
	"reflect"
	"sort"
)

const (
	allowedAdditionReason = "NEW_MEASURABLE_OBLIGATION"
	allowedDeletionReason = "DEPRECATED_DUPLICATE"
	noReceipt             = "none"
)

var forbiddenPropositions = []string{"improvement rate", "aggregate estimate", "projected coverage"}

// Evaluate is the producer-side calculation. The source supplies members,
// proposal references, reasons, and receipt binding material. It supplies no
// decision, resolution, claim state, or expected result.
func Evaluate(input Input) Report {
	contract := input.Contract
	snapshot := input.RepositorySnapshot
	report := Report{
		Schema: ReportSchema, Scope: ReportScope, HeadSHA: input.HeadSHA,
		Producer: "denominatorevolution.Evaluate", Consumer: "denominatorevolutionverify.Verify",
		ContractDigest: digestValue(contract), SourceDigest: DigestBytes(input.Source),
		NotClaimed: append([]string(nil), contract.NotClaimed...), AggregateMetrics: []string{},
		RepositoryWrites: snapshot.ChangedPaths, MutationAuthority: false, RepositorySnapshot: snapshot,
	}

	projection, wire, err := parseSource(input.Source)
	report.SourceProjection = projection
	if err != nil {
		return finishFailure(report, "GOOO_SOURCE_PROJECTION_UNKNOWN", "LOWER_RESOLUTION")
	}
	base, records, inputs := sourceCaseInputs(wire, snapshot)
	report.Denominator = base
	report.DenominatorRecords = records
	for _, value := range inputs {
		report.Cases = append(report.Cases, evaluateCase(value, base))
	}
	report.ClaimLedger = sealClaimLedger(report.Cases)
	report.EmittedClaims = emitClaims(report.Cases)
	guards := makeGuardrails(forbiddenEstimateObserved(report.EmittedClaims), snapshot.ChangedPaths, projection.ForbiddenPropositionPresent)
	for index := range report.Cases {
		if report.Cases[index].Receipt != nil {
			report.Cases[index].Receipt.Guardrails = guards
			report.Cases[index].Receipt.Digest = receiptDigest(*report.Cases[index].Receipt)
		}
	}
	report.Summary = summarize(report.Cases, base, records, len(wire.Proposals), len(report.ClaimLedger), guards)
	report.Indicators = makeIndicators(report.Summary)

	if !contractExpectationMatches(report, contract) {
		return finishFailure(report, "DENOMINATOR_EVOLUTION_CONTRACT_DRIFT", "INVARIANT_ONLY")
	}
	if report.Summary.CasesSatisfied != CaseCount || len(report.Indicators) != CheckCount || hasUnsatisfied(report.Indicators) {
		return finishFailure(report, "DENOMINATOR_EVOLUTION_CONTRACT_VIOLATED", "INVARIANT_ONLY")
	}
	report.Decision, report.Resolution, report.Reason = "PASS", "EXACT", "DENOMINATOR_EVOLUTION_CONTRACT_SATISFIED"
	report.Digest = reportDigest(report)
	return report
}

func sourceCaseInputs(wire sourceWire, snapshot RepositorySnapshot) (Denominator, []DenominatorRecord, []CaseInput) {
	base := Denominator{Version: wire.Version, Obligations: cloneObligations(wire.Obligations)}
	base.Digest = denominatorDigest(base)
	registry := map[DenominatorRef]Denominator{{Version: base.Version, Digest: base.Digest}: base}

	records := []DenominatorRecord{}
	records = append(records, newDenominatorRecord("denominator-v1", base, nil))
	inputs := make([]CaseInput, 0, len(wire.Proposals))
	for _, proposal := range wire.Proposals {
		predecessor, known := registry[proposal.Predecessor]
		if !known {
			predecessor = Denominator{Version: proposal.Predecessor.Version, Digest: proposal.Predecessor.Digest}
		}
		successor, changeSetValid := applyChanges(base, proposal.Successor, wire.Additions, wire.Deletions)
		var receipt *MigrationReceipt
		if proposal.ReceiptID != noReceipt && proposal.ReceiptID == wire.Receipt.ID {
			receipt = materializeReceipt(wire.Receipt, snapshot)
		}
		inputs = append(inputs, CaseInput{ID: proposal.ID, Predecessor: predecessor, Successor: successor, PredecessorRef: proposal.Predecessor, SuccessorVersion: proposal.Successor, ReceiptID: proposal.ReceiptID, Receipt: receipt, ReceiptBoundPrev: wire.Receipt.BoundPrev, ReceiptBoundNext: wire.Receipt.BoundNext, Additions: append([]Change(nil), wire.Additions...), Deletions: append([]Change(nil), wire.Deletions...), PredecessorKnown: known, ChangeSetValid: changeSetValid})
		if known && proposal.Successor == SuccessorVersion && changeSetValid && len(records) == 1 {
			ref := DenominatorRef{Version: base.Version, Digest: base.Digest}
			records = append(records, newDenominatorRecord("denominator-v2", successor, &ref))
		}
	}
	return base, records, inputs
}

func newDenominatorRecord(id string, denominator Denominator, predecessor *DenominatorRef) DenominatorRecord {
	record := DenominatorRecord{ID: id, Version: denominator.Version, Predecessor: predecessor, Denominator: denominator, FixedMemberNumerator: len(denominator.Obligations), FixedMemberDenominator: DenominatorSize, Immutable: true}
	record.Digest = denominatorRecordDigest(record)
	return record
}

func materializeReceipt(material sourceReceiptMaterial, snapshot RepositorySnapshot) *MigrationReceipt {
	return &MigrationReceipt{
		Schema: ReceiptSchema, ID: material.ID, Producer: "denominatorevolution.Evaluate", Consumer: "denominatorevolutionverify.Verify",
		Predecessor: material.Predecessor, Successor: material.Successor,
		Additions: append([]Change(nil), material.Additions...), Deletions: append([]Change(nil), material.Deletions...),
		Coordinate:       Coordinate{Stage: "MIGRATE", Step: "bind-receipt-to-both-versions", Reason: "receipt binds computed predecessor and successor digests"},
		RepositoryWrites: snapshot.ChangedPaths, MutationAuthority: false, RepositorySnapshot: snapshot,
		Guardrails: makeGuardrails(0, snapshot.ChangedPaths, true),
	}
}

func applyChanges(base Denominator, version string, additions, deletions []Change) (Denominator, bool) {
	result := cloneDenominator(base, version)
	valid := true
	for _, change := range deletions {
		count := 0
		for _, obligation := range result.Obligations {
			if obligation.ID == change.ObligationID {
				count++
			}
		}
		if count != 1 {
			valid = false
			continue
		}
		result.Obligations = removeObligation(result.Obligations, change.ObligationID)
	}
	for _, change := range additions {
		if change.Member == nil || hasObligation(result.Obligations, change.ObligationID) || change.Member.ID != change.ObligationID {
			valid = false
			continue
		}
		result.Obligations = append(result.Obligations, *change.Member)
	}
	result.Digest = denominatorDigest(result)
	if !changeSetValid(base, result, additions, deletions) {
		valid = false
	}
	return result, valid
}

func changeSetValid(base, successor Denominator, additions, deletions []Change) bool {
	if len(additions) != 1 || len(deletions) != 1 || !reasonAllowed(additions[0].Reason, true) || !reasonAllowed(deletions[0].Reason, false) {
		return false
	}
	addition, deletion := additions[0], deletions[0]
	if addition.Member == nil || addition.Member.ID != addition.ObligationID || hasObligation(base.Obligations, addition.ObligationID) || !hasObligation(base.Obligations, deletion.ObligationID) || addition.ObligationID == deletion.ObligationID {
		return false
	}
	actualAdditions, actualDeletions := changes(base, successor)
	return sameChangeIDs(actualAdditions, additions) && sameChangeIDs(actualDeletions, deletions)
}

func changes(before, after Denominator) ([]Change, []Change) {
	beforeIDs, afterIDs := map[string]Obligation{}, map[string]Obligation{}
	for _, value := range before.Obligations {
		beforeIDs[value.ID] = value
	}
	for _, value := range after.Obligations {
		afterIDs[value.ID] = value
	}
	additions, deletions := []Change{}, []Change{}
	for id, value := range afterIDs {
		if _, ok := beforeIDs[id]; !ok {
			member := value
			additions = append(additions, Change{ObligationID: id, Member: &member})
		}
	}
	for id := range beforeIDs {
		if _, ok := afterIDs[id]; !ok {
			deletions = append(deletions, Change{ObligationID: id})
		}
	}
	sort.Slice(additions, func(i, j int) bool { return additions[i].ObligationID < additions[j].ObligationID })
	sort.Slice(deletions, func(i, j int) bool { return deletions[i].ObligationID < deletions[j].ObligationID })
	return additions, deletions
}

func evaluateCase(input CaseInput, base Denominator) CaseResult {
	predKnown := input.PredecessorKnown && input.PredecessorRef == (DenominatorRef{Version: base.Version, Digest: base.Digest}) && input.Predecessor.Digest == denominatorDigest(input.Predecessor) && reflect.DeepEqual(input.Predecessor.Obligations, base.Obligations)
	receiptValid := predKnown && input.ChangeSetValid && receiptBindingValid(input)
	decision, resolution, reason := "FAIL_CLOSED", "LOWER_RESOLUTION", "PREDECESSOR_DIGEST_UNKNOWN"
	if predKnown {
		decision, resolution, reason = "BLOCK", "INVARIANT_ONLY", "MIGRATION_RECEIPT_MISSING"
		if input.ReceiptID != noReceipt && receiptValid {
			decision, resolution, reason = "ADVANCE", "EXACT", "DENOMINATOR_ADVANCE_AUTHORIZED"
		}
	}
	claimID, fromClaim, toClaim, coordinate := derivedClaim(decision, resolution, reason)
	if input.Receipt != nil {
		if receiptValid {
			input.Receipt.Decision, input.Receipt.Reason = "ADVANCE", "DENOMINATOR_ADVANCE_AUTHORIZED"
		} else {
			input.Receipt.Decision, input.Receipt.Reason = "BLOCK", "MIGRATION_RECEIPT_MISSING"
		}
		input.Receipt.Digest = receiptDigest(*input.Receipt)
	}
	kind := "UNKNOWN_PREDECESSOR"
	if predKnown {
		if input.ReceiptID == noReceipt {
			kind = "REGISTERED_WITHOUT_RECEIPT"
		} else {
			kind = "REGISTERED_WITH_RECEIPT"
		}
	}
	checks := []CheckResult{
		check("denominator-version", "FOUNDATION", "bind-fixed-denominator", coordinate, status(predKnown, "PASS", "UNKNOWN"), DenominatorVersion, input.Predecessor.Version),
		check("denominator-member-digest", "FOUNDATION", "bind-fixed-denominator", coordinate, status(input.Predecessor.Digest == denominatorDigest(input.Predecessor) && input.Successor.Digest == denominatorDigest(input.Successor), "PASS", "FAIL"), "source-derived digests", input.Predecessor.Digest+" / "+input.Successor.Digest),
		check("predecessor-registration", "FOUNDATION", "bind-predecessor-digest", coordinate, status(predKnown, "PASS", "UNKNOWN"), base.Version+" / "+base.Digest, input.Predecessor.Version+" / "+input.Predecessor.Digest),
		check("addition-reason", "COHERENCE", "classify-change-reason", coordinate, status(input.ChangeSetValid && len(input.Additions) == 1 && reasonAllowed(input.Additions[0].Reason, true), "PASS", "FAIL"), allowedAdditionReason, changeText(input.Additions)),
		check("deletion-reason", "COHERENCE", "classify-change-reason", coordinate, status(input.ChangeSetValid && len(input.Deletions) == 1 && reasonAllowed(input.Deletions[0].Reason, false), "PASS", "FAIL"), allowedDeletionReason, changeText(input.Deletions)),
		check("migration-receipt", "COHERENCE", "accept-migration-receipt", coordinate, receiptStatus(predKnown, receiptValid), "computed receipt bound to both digests", receiptText(input.Receipt)),
		check("claim-transition", claimProof(decision), claimOperation(decision), coordinate, "PASS", "OPEN -> "+toClaim, fromClaim+" -> "+toClaim),
		check("read-only-effect", "REGRESSION", "preserve-read-only-migration", coordinate, status(input.Receipt == nil || input.Receipt.RepositoryWrites == 0, "PASS", "FAIL"), "repository snapshot changed paths <= 0", repositoryWritesText(input.Receipt)),
	}
	return CaseResult{ID: input.ID, Kind: kind, Status: "SATISFIED", ObservedDecision: decision, ObservedResolution: resolution, ObservedReason: reason, ClaimID: claimID, FromClaim: fromClaim, ToClaim: toClaim, Predecessor: input.Predecessor, Successor: input.Successor, Receipt: input.Receipt, Coordinate: coordinate, Checks: checks}
}

func receiptBindingValid(input CaseInput) bool {
	receipt := input.Receipt
	if receipt == nil || receipt.Schema != ReceiptSchema || receipt.ID == "" || receipt.Producer == "" || receipt.Consumer == "" || receipt.RepositoryWrites != 0 || receipt.MutationAuthority || !guardrailsConform(receipt.Guardrails) {
		return false
	}
	if receipt.Predecessor != input.PredecessorRef || receipt.Successor != (DenominatorRef{Version: input.Successor.Version, Digest: input.Successor.Digest}) || receipt.Predecessor != input.ReceiptBoundPrev || receipt.Successor != input.ReceiptBoundNext {
		return false
	}
	return sameChangeSet(input.Additions, receipt.Additions) && sameChangeSet(input.Deletions, receipt.Deletions)
}

func derivedClaim(decision, resolution, reason string) (string, string, string, Coordinate) {
	if decision == "FAIL_CLOSED" {
		return "predecessor-resolution", "OPEN", "OPEN", Coordinate{Stage: "RESOLVE", Step: "lookup-predecessor", Reason: reason}
	}
	if decision == "ADVANCE" && resolution == "EXACT" {
		return "denominator-advance-authorized", "OPEN", "DISCHARGED", Coordinate{Stage: "DECIDE", Step: "apply-migration-receipt", Reason: reason}
	}
	return "migration-authorized", "OPEN", "REFUTED", Coordinate{Stage: "DECIDE", Step: "reject-invalid-receipt", Reason: reason}
}

func sealClaimLedger(cases []CaseResult) []ClaimLedgerEntry {
	result := make([]ClaimLedgerEntry, 0, len(cases))
	previous := ""
	for index, value := range cases {
		entry := ClaimLedgerEntry{Sequence: index + 1, ClaimID: value.ClaimID, PriorState: value.FromClaim, NextState: value.ToClaim, Stage: value.Coordinate.Stage, Step: value.Coordinate.Step, Reason: value.Coordinate.Reason, PreviousDigest: previous}
		entry.Digest = claimLedgerDigest(entry)
		result = append(result, entry)
		previous = entry.Digest
	}
	return result
}

func emitClaims(cases []CaseResult) []EmittedClaim {
	result := make([]EmittedClaim, 0, len(cases))
	for _, value := range cases {
		class := "MIGRATION_AUTHORIZATION"
		if value.ObservedDecision == "FAIL_CLOSED" {
			class = "PREDECESSOR_RESOLUTION"
		}
		result = append(result, EmittedClaim{ID: value.ClaimID, Class: class, State: value.ToClaim})
	}
	return result
}

func contractExpectationMatches(report Report, contract Contract) bool {
	if contract.Denominator.Version != report.Denominator.Version || !reflect.DeepEqual(contract.Denominator.Obligations, report.Denominator.Obligations) || len(contract.Cases) != len(report.Cases) || len(report.AggregateMetrics) != 0 || !contract.Policy.NoAggregateEstimates || !containsAll(contract.Policy.ForbiddenClaims, forbiddenPropositions) || !contains(contract.Policy.AllowedAdditionReasons, allowedAdditionReason) || !contains(contract.Policy.AllowedDeletionReasons, allowedDeletionReason) {
		return false
	}
	for index, expected := range contract.Cases {
		observed := report.Cases[index]
		if observed.ID != expected.ID || observed.Kind != expected.Kind || observed.ObservedDecision != expected.ExpectedDecision || observed.ObservedResolution != expected.ExpectedResolution || observed.ObservedReason != expected.ExpectedReason || observed.FromClaim != expected.FromClaim || observed.ToClaim != expected.ToClaim {
			return false
		}
	}
	return true
}

func cloneObligations(value []Obligation) []Obligation { return append([]Obligation(nil), value...) }

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

func hasObligation(values []Obligation, id string) bool {
	for _, value := range values {
		if value.ID == id {
			return true
		}
	}
	return false
}

func reasonAllowed(reason string, addition bool) bool {
	if addition {
		return reason == allowedAdditionReason
	}
	return reason == allowedDeletionReason
}

func sameChangeIDs(left, right []Change) bool {
	if len(left) != len(right) {
		return false
	}
	leftCopy := append([]Change(nil), left...)
	rightCopy := append([]Change(nil), right...)
	sort.Slice(leftCopy, func(i, j int) bool { return leftCopy[i].ObligationID < leftCopy[j].ObligationID })
	sort.Slice(rightCopy, func(i, j int) bool { return rightCopy[i].ObligationID < rightCopy[j].ObligationID })
	for index := range leftCopy {
		if leftCopy[index].ObligationID != rightCopy[index].ObligationID {
			return false
		}
	}
	return true
}

func containsAll(values, wanted []string) bool {
	for _, value := range wanted {
		if !containsFoldAny(values, value) {
			return false
		}
	}
	return true
}

func containsFoldAny(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func contains(values []string, wanted string) bool { return containsFoldAny(values, wanted) }
