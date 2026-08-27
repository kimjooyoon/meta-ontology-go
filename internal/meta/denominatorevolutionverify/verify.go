package denominatorevolutionverify

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
)

const (
	guardrailForbiddenEstimate = "gooo.guardrail.denominator.forbidden-estimate.v1"
	guardrailRepositoryWrites  = "gooo.guardrail.denominator.repository-writes.v1"
	guardrailCount             = 2
	allowedAdditionReason      = "NEW_MEASURABLE_OBLIGATION"
	allowedDeletionReason      = "DEPRECATED_DUPLICATE"
	noReceipt                  = "none"
)

var forbiddenPropositions = []string{"improvement rate", "aggregate estimate", "projected coverage"}

type caseInput struct {
	ID               string
	Predecessor      Denominator
	Successor        Denominator
	PredecessorRef   Ref
	SuccessorVersion string
	ReceiptID        string
	Receipt          *Receipt
	ReceiptBoundPrev Ref
	ReceiptBoundNext Ref
	Additions        []Change
	Deletions        []Change
	PredecessorKnown bool
	ChangeSetValid   bool
}

// Verify independently reconstructs all source inputs and derives all case
// results. The producer report is evidence to compare, never an authority.
func Verify(input Input) Verification {
	verification := Verification{Schema: VerificationSchema, HeadSHA: input.HeadSHA, Producer: "denominatorevolution.Evaluate", Consumer: "denominatorevolutionverify.Verify", SourceDigest: digestBytes(input.SourceRaw), NotClaimed: []string{"improvement rate", "aggregate estimate", "projected coverage", "semantic quality", "repository mutation"}, AggregateMetrics: []string{}, RepositoryWrites: input.RepositorySnapshot.ChangedPaths, MutationAuthority: false, RepositorySnapshot: input.RepositorySnapshot}
	var contract Contract
	var report Report
	contractOK := decode(input.ContractRaw, &contract)
	reportOK := decode(input.ReportRaw, &report)
	if contractOK {
		verification.ContractDigest = digestValue(contract)
	} else {
		verification.ContractDigest = digestBytes(input.ContractRaw)
	}
	if reportOK {
		verification.ReportDigest = report.Digest
	}
	checks := []Check{
		verifyCheck("contract-schema", "FOUNDATION", "bind-contract-version", "FOUNDATION", "decode-contract", contractOK && contract.Schema == ContractSchema && contract.Version == 1 && len(contract.Denominator.Obligations) == DenominatorSize && len(contract.Cases) == CaseCount, "contract v1 expectation with 5 members and 3 cases", contract.Schema),
		verifyCheck("report-schema", "FOUNDATION", "bind-report-version", "FOUNDATION", "decode-report", reportOK && report.Schema == ReportSchema, ReportSchema, report.Schema),
		verifyCheck("report-digest", "REGRESSION", "recompute-report-digest", "REPLAY", "compare-report-digest", reportOK && report.Digest == reportDigest(report), "recomputed report digest", report.Digest),
		verifyCheck("contract-digest", "FOUNDATION", "bind-contract-digest", "FOUNDATION", "compare-contract-digest", reportOK && report.ContractDigest == verification.ContractDigest, verification.ContractDigest, report.ContractDigest),
		verifyCheck("source-digest", "FOUNDATION", "bind-source-digest", "FOUNDATION", "compare-source-digest", reportOK && len(input.SourceRaw) > 0 && report.SourceDigest == verification.SourceDigest, verification.SourceDigest, report.SourceDigest),
		verifyCheck("repository-snapshot", "REGRESSION", "bind-ci-observation", "EFFECT", "compare-before-after-snapshot", reportOK && report.RepositorySnapshot == input.RepositorySnapshot && report.RepositoryWrites == input.RepositorySnapshot.ChangedPaths, snapshotText(input.RepositorySnapshot), snapshotText(report.RepositorySnapshot)),
	}

	projection, wire, sourceErr := parseSource(input.SourceRaw)
	checks = append(checks, verifyCheck("source-reconstruction", "FOUNDATION", "parse-and-lower-source", "FOUNDATION", "reconstruct-source-inputs", sourceErr == nil && projection.Exact && projection.CaseCount == CaseCount && projection.ObligationCount == DenominatorSize, "source-derived 5 obligations and 3 proposals", sourceErrorText(projection, sourceErr)))
	if reportOK && sourceErr == nil {
		base, records, inputs := sourceInputs(wire, input.RepositorySnapshot, projection.ForbiddenPropositionPresent)
		computed := make([]Case, 0, len(inputs))
		for _, value := range inputs {
			computed = append(computed, evaluateCase(value, base))
		}
		ledger := sealLedger(computed)
		emitted := emitClaims(computed)
		guards := makeGuardrails(forbiddenEstimateObserved(emitted), input.RepositorySnapshot.ChangedPaths, projection.ForbiddenPropositionPresent)
		for index := range computed {
			if computed[index].Receipt != nil {
				computed[index].Receipt.Guardrails = guards
				computed[index].Receipt.Digest = receiptDigest(*computed[index].Receipt)
			}
		}
		verification.Guardrails = guards
		verification.ClaimLedger = ledger
		verification.EmittedClaims = emitted
		verification.DenominatorRecords = records
		checks = append(checks, verifySourceProjection(report, projection)...)
		checks = append(checks, verifyDenominator(report, base, records)...)
		checks = append(checks, verifyCases(report, contract, computed)...)
		checks = append(checks, verifyClaims(report, ledger, emitted)...)
		checks = append(checks, verifySummary(report, len(computed), len(ledger), records, guards)...)
		checks = append(checks, verifyArtifactSchema(report, projection)...)
	} else {
		verification.Guardrails = makeGuardrails(0, input.RepositorySnapshot.ChangedPaths, false)
	}
	verification.Checks = checks
	verification.Decision, verification.Resolution, verification.Reason = "PASS", "EXACT", "INDEPENDENT_DENOMINATOR_DECISION_VERIFIED"
	if !contractOK || !reportOK || sourceErr != nil || hasFailure(checks) || report.HeadSHA != input.HeadSHA || report.Decision != "PASS" || report.Resolution != "EXACT" {
		verification.Decision, verification.Resolution, verification.Reason = "FAIL_CLOSED", "INVARIANT_ONLY", "INDEPENDENT_DENOMINATOR_DECISION_REJECTED"
	}
	verification.Digest = verificationDigest(verification)
	return verification
}

func decode(raw []byte, value any) bool {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(value) == nil
}

func sourceInputs(wire sourceWire, snapshot RepositorySnapshot, propositionPresent bool) (Denominator, []DenominatorRecord, []caseInput) {
	base := Denominator{Version: wire.Version, Obligations: append([]Obligation(nil), wire.Obligations...)}
	base.Digest = denominatorDigest(base)
	registry := map[Ref]Denominator{{Version: base.Version, Digest: base.Digest}: base}
	records := []DenominatorRecord{newRecord("denominator-v1", base, nil)}
	inputs := make([]caseInput, 0, len(wire.Proposals))
	for _, proposal := range wire.Proposals {
		predecessor, known := registry[proposal.Predecessor]
		if !known {
			predecessor = Denominator{Version: proposal.Predecessor.Version, Digest: proposal.Predecessor.Digest}
		}
		successor, changeSetValid := applyChanges(base, proposal.Successor, wire.Additions, wire.Deletions)
		var receipt *Receipt
		if proposal.ReceiptID != noReceipt && proposal.ReceiptID == wire.Receipt.ID {
			receipt = materializeReceipt(wire.Receipt, snapshot, propositionPresent)
		}
		inputs = append(inputs, caseInput{ID: proposal.ID, Predecessor: predecessor, Successor: successor, PredecessorRef: proposal.Predecessor, SuccessorVersion: proposal.Successor, ReceiptID: proposal.ReceiptID, Receipt: receipt, ReceiptBoundPrev: wire.Receipt.BoundPrev, ReceiptBoundNext: wire.Receipt.BoundNext, Additions: append([]Change(nil), wire.Additions...), Deletions: append([]Change(nil), wire.Deletions...), PredecessorKnown: known, ChangeSetValid: changeSetValid})
		if known && proposal.Successor == SuccessorVersion && changeSetValid && len(records) == 1 {
			ref := Ref{Version: base.Version, Digest: base.Digest}
			records = append(records, newRecord("denominator-v2", successor, &ref))
		}
	}
	return base, records, inputs
}

func newRecord(id string, denominator Denominator, predecessor *Ref) DenominatorRecord {
	record := DenominatorRecord{ID: id, Version: denominator.Version, Predecessor: predecessor, Denominator: denominator, FixedMemberNumerator: len(denominator.Obligations), FixedMemberDenominator: DenominatorSize, Immutable: true}
	record.Digest = recordDigest(record)
	return record
}

func materializeReceipt(material sourceReceiptMaterial, snapshot RepositorySnapshot, propositionPresent bool) *Receipt {
	return &Receipt{Schema: ReceiptSchema, ID: material.ID, Producer: "denominatorevolution.Evaluate", Consumer: "denominatorevolutionverify.Verify", Predecessor: material.Predecessor, Successor: material.Successor, Additions: append([]Change(nil), material.Additions...), Deletions: append([]Change(nil), material.Deletions...), Coordinate: Coordinate{Stage: "MIGRATE", Step: "bind-receipt-to-both-versions", Reason: "receipt binds computed predecessor and successor digests"}, RepositoryWrites: snapshot.ChangedPaths, MutationAuthority: false, RepositorySnapshot: snapshot, Guardrails: makeGuardrails(0, snapshot.ChangedPaths, propositionPresent)}
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
	if len(additions) != 1 || len(deletions) != 1 || additions[0].Reason != allowedAdditionReason || deletions[0].Reason != allowedDeletionReason {
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

func evaluateCase(input caseInput, base Denominator) Case {
	predKnown := input.PredecessorKnown && input.PredecessorRef == (Ref{Version: base.Version, Digest: base.Digest}) && input.Predecessor.Digest == denominatorDigest(input.Predecessor) && reflect.DeepEqual(input.Predecessor.Obligations, base.Obligations)
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
	checks := []Check{
		check("denominator-version", "FOUNDATION", "bind-fixed-denominator", coordinate, status(predKnown, "PASS", "UNKNOWN"), DenominatorVersion, input.Predecessor.Version),
		check("denominator-member-digest", "FOUNDATION", "bind-fixed-denominator", coordinate, status(input.Predecessor.Digest == denominatorDigest(input.Predecessor) && input.Successor.Digest == denominatorDigest(input.Successor), "PASS", "FAIL"), "source-derived digests", input.Predecessor.Digest+" / "+input.Successor.Digest),
		check("predecessor-registration", "FOUNDATION", "bind-predecessor-digest", coordinate, status(predKnown, "PASS", "UNKNOWN"), base.Version+" / "+base.Digest, input.Predecessor.Version+" / "+input.Predecessor.Digest),
		check("addition-reason", "COHERENCE", "classify-change-reason", coordinate, status(input.ChangeSetValid && len(input.Additions) == 1 && input.Additions[0].Reason == allowedAdditionReason, "PASS", "FAIL"), allowedAdditionReason, changeText(input.Additions)),
		check("deletion-reason", "COHERENCE", "classify-change-reason", coordinate, status(input.ChangeSetValid && len(input.Deletions) == 1 && input.Deletions[0].Reason == allowedDeletionReason, "PASS", "FAIL"), allowedDeletionReason, changeText(input.Deletions)),
		check("migration-receipt", "COHERENCE", "accept-migration-receipt", coordinate, receiptStatus(predKnown, receiptValid), "computed receipt bound to both digests", receiptText(input.Receipt)),
		check("claim-transition", claimProof(decision), claimOperation(decision), coordinate, "PASS", "OPEN -> "+toClaim, fromClaim+" -> "+toClaim),
		check("read-only-effect", "REGRESSION", "preserve-read-only-migration", coordinate, status(input.Receipt == nil || input.Receipt.RepositoryWrites == 0, "PASS", "FAIL"), "repository snapshot changed paths <= 0", receiptWrites(input.Receipt)),
	}
	return Case{ID: input.ID, Kind: kind, Status: "SATISFIED", ObservedDecision: decision, ObservedResolution: resolution, ObservedReason: reason, ClaimID: claimID, FromClaim: fromClaim, ToClaim: toClaim, Predecessor: input.Predecessor, Successor: input.Successor, Receipt: input.Receipt, Coordinate: coordinate, Checks: checks}
}

func receiptBindingValid(input caseInput) bool {
	receipt := input.Receipt
	if receipt == nil || receipt.Schema != ReceiptSchema || receipt.ID == "" || receipt.Producer == "" || receipt.Consumer == "" || receipt.RepositoryWrites != 0 || receipt.MutationAuthority || !guardrailsConform(receipt.Guardrails) {
		return false
	}
	if receipt.Predecessor != input.PredecessorRef || receipt.Successor != (Ref{Version: input.Successor.Version, Digest: input.Successor.Digest}) || receipt.Predecessor != input.ReceiptBoundPrev || receipt.Successor != input.ReceiptBoundNext {
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

func sealLedger(cases []Case) []ClaimLedgerEntry {
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

func emitClaims(cases []Case) []EmittedClaim {
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

func verifySourceProjection(report Report, expected SourceProjection) []Check {
	ok := reflect.DeepEqual(report.SourceProjection, expected)
	return []Check{verifyCheck("source-projection", "FOUNDATION", "bind-source-inputs", "FOUNDATION", "compare-lowered-source", ok, "exact ParseFile -> bidir.Lower input projection", sourceProjectionText(report.SourceProjection))}
}

func verifyDenominator(report Report, base Denominator, records []DenominatorRecord) []Check {
	ok := reflect.DeepEqual(report.Denominator, base) && reflect.DeepEqual(report.DenominatorRecords, records) && len(records) == 2 && records[0].Digest == recordDigest(records[0]) && records[1].Digest == recordDigest(records[1])
	return []Check{verifyCheck("fixed-denominator", "FOUNDATION", "bind-fixed-denominator", "FOUNDATION", "compare-member-set", ok, "5 source-derived members at "+DenominatorVersion+" plus immutable v2 record", report.Denominator.Version+" / "+report.Denominator.Digest)}
}

func verifyCases(report Report, contract Contract, expected []Case) []Check {
	if len(report.Cases) != CaseCount || len(contract.Cases) != CaseCount || len(expected) != CaseCount {
		return []Check{verifyCheck("case-cardinality", "FOUNDATION", "bind-case-corpus", "FOUNDATION", "compare-case-count", false, "3 source-derived cases", intText(len(report.Cases)))}
	}
	checks := []Check{}
	for index, want := range expected {
		observed := report.Cases[index]
		exp := contract.Cases[index]
		ok := reflect.DeepEqual(observed, want) && want.ID == exp.ID && want.Kind == exp.Kind && want.ObservedDecision == exp.ExpectedDecision && want.ObservedResolution == exp.ExpectedResolution && want.ObservedReason == exp.ExpectedReason && want.FromClaim == exp.FromClaim && want.ToClaim == exp.ToClaim
		checks = append(checks, verifyCheck(want.ID, exp.ProofChoice, exp.MetaOperation, exp.Stage, exp.Step, ok, exp.ExpectedDecision+" / "+exp.ExpectedReason, want.ObservedDecision+" / "+want.ObservedReason))
	}
	return checks
}

func verifyClaims(report Report, expected []ClaimLedgerEntry, emitted []EmittedClaim) []Check {
	ok := reflect.DeepEqual(report.ClaimLedger, expected) && reflect.DeepEqual(report.EmittedClaims, emitted) && validLedger(expected)
	return []Check{verifyCheck("persistent-claim-ledger", "COHERENCE", "bind-persistent-claim-ledger", "DECIDE", "replay-calculated-state-transitions", ok, "3 calculated OPEN-origin ledger entries", intText(len(report.ClaimLedger))+" entries")}
}

func verifySummary(report Report, cases, ledger int, records []DenominatorRecord, guards []Guardrail) []Check {
	summary := report.Summary
	ok := summary.CasesSatisfied == CaseCount && summary.CasesTotal == cases && summary.FixedDenominatorNumerator == DenominatorSize && summary.FixedDenominatorDenominator == DenominatorSize && summary.LegalAdvanceNumerator == 1 && summary.LegalAdvanceDenominator == 1 && summary.UnauthorizedRejectionNumerator == 1 && summary.UnauthorizedRejectionDenominator == 1 && summary.UnknownPredecessorNumerator == 1 && summary.UnknownPredecessorDenominator == 1 && summary.AdditionReasonNumerator == 1 && summary.AdditionReasonDenominator == 1 && summary.DeletionReasonNumerator == 1 && summary.DeletionReasonDenominator == 1 && summary.SourceCasesNumerator == cases && summary.SourceCasesDenominator == CaseCount && summary.PersistentClaimsNumerator == ledger && summary.PersistentClaimsDenominator == CaseCount && summary.GuardrailObservationsNumerator == guardrailCount && summary.GuardrailObservationsDenominator == guardrailCount && summary.VersionRecordsNumerator == len(records) && summary.VersionRecordsDenominator == 2 && summary.V1NonretroactiveNumerator == 1 && summary.V1NonretroactiveDenominator == 1 && reflect.DeepEqual(summary.Guardrails, guards) && guardrailsConform(summary.Guardrails)
	return []Check{verifyCheck("exact-summary", "FOUNDATION", "bind-exact-counters", "SUMMARY", "compare-numerators-and-denominators", ok, "fixed 5/5, source cases 3/3, persistent claims 3/3, records 2/2, guardrails 2/2", "summary counters and explicit observations")}
}

func verifyArtifactSchema(report Report, projection SourceProjection) []Check {
	ok := projection.ForbiddenPropositionPresent && len(report.AggregateMetrics) == 0 && len(report.NotClaimed) > 0 && !containsAggregateMetric(report)
	return []Check{verifyCheck("artifact-no-aggregate", "REGRESSION", "reject-aggregate-artifact", "GUARD", "inspect-artifact-schema", ok, "no aggregate metric artifact is generated", "aggregate_metrics="+intText(len(report.AggregateMetrics)))}
}

func containsAggregateMetric(report Report) bool {
	for _, value := range report.AggregateMetrics {
		if value != "" {
			return true
		}
	}
	return false
}

func makeGuardrails(forbidden, writes int, propositionPresent bool) []Guardrail {
	return []Guardrail{newGuardrail(guardrailForbiddenEstimate, propositionPresent, forbidden, 0), newGuardrail(guardrailRepositoryWrites, false, writes, 0)}
}

func newGuardrail(id string, propositionPresent bool, observed, allowed int) Guardrail {
	conforms := observed <= allowed
	numerator := 0
	if conforms {
		numerator = 1
	}
	return Guardrail{ID: id, Direction: "AT_MOST", PropositionPresent: propositionPresent, Observed: observed, AllowedMax: allowed, ConformanceNumerator: numerator, ConformanceDenominator: 1, Conforms: conforms}
}

func forbiddenEstimateObserved(values []EmittedClaim) int {
	count := 0
	for _, value := range values {
		if value.Class == "FORBIDDEN_ESTIMATE" && value.State == "ASSERTED" {
			count++
		}
	}
	return count
}

func guardrailsConform(values []Guardrail) bool {
	if len(values) != guardrailCount {
		return false
	}
	ids := []string{guardrailForbiddenEstimate, guardrailRepositoryWrites}
	for index, value := range values {
		conforms := value.Observed <= value.AllowedMax
		numerator := 0
		if conforms {
			numerator = 1
		}
		if value.ID != ids[index] || value.Direction != "AT_MOST" || value.AllowedMax != 0 || value.ConformanceDenominator != 1 || value.ConformanceNumerator != numerator || value.Conforms != conforms {
			return false
		}
	}
	return values[0].PropositionPresent
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

func validLedger(values []ClaimLedgerEntry) bool {
	previous := ""
	for index, value := range values {
		if value.Sequence != index+1 || value.PriorState != "OPEN" || value.PreviousDigest != previous || value.Digest != claimLedgerDigest(value) {
			return false
		}
		previous = value.Digest
	}
	return true
}

func cloneDenominator(value Denominator, version string) Denominator {
	return Denominator{Version: version, Obligations: append([]Obligation(nil), value.Obligations...)}
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
