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
)

// Verify independently parses and lowers the raw Gooo source. Report cases,
// denominator members, claim transitions, and guardrails are all replayed
// against this local wire model rather than treated as authority.
func Verify(input Input) Verification {
	verification := Verification{Schema: VerificationSchema, HeadSHA: input.HeadSHA, Producer: "denominatorevolution.Evaluate", Consumer: "denominatorevolutionverify.Verify", SourceDigest: digestBytes(input.SourceRaw), NotClaimed: []string{"improvement rate", "aggregate estimate", "projected coverage", "semantic quality", "repository mutation"}, RepositoryWrites: input.RepositorySnapshot.ChangedPaths, MutationAuthority: false, RepositorySnapshot: input.RepositorySnapshot}
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
		verifyCheck("contract-schema", "FOUNDATION", "bind-contract-version", "FOUNDATION", "decode-contract", contractOK && contract.Schema == ContractSchema && contract.Version == 1 && len(contract.Denominator.Obligations) == DenominatorSize && len(contract.Cases) == CaseCount, "contract v1 with 5 expected members and 3 expected cases", contract.Schema),
		verifyCheck("report-schema", "FOUNDATION", "bind-report-version", "FOUNDATION", "decode-report", reportOK && report.Schema == ReportSchema, ReportSchema, report.Schema),
		verifyCheck("report-digest", "REGRESSION", "recompute-report-digest", "REPLAY", "compare-report-digest", reportOK && report.Digest == reportDigest(report), "recomputed digest", report.Digest),
		verifyCheck("contract-digest", "FOUNDATION", "bind-contract-digest", "FOUNDATION", "compare-contract-digest", reportOK && report.ContractDigest == verification.ContractDigest, verification.ContractDigest, report.ContractDigest),
		verifyCheck("source-digest", "FOUNDATION", "bind-source-digest", "FOUNDATION", "compare-source-digest", reportOK && len(input.SourceRaw) > 0 && report.SourceDigest == verification.SourceDigest, verification.SourceDigest, report.SourceDigest),
		verifyCheck("repository-snapshot", "REGRESSION", "bind-ci-observation", "EFFECT", "compare-before-after-snapshot", reportOK && report.RepositorySnapshot == input.RepositorySnapshot && report.RepositoryWrites == input.RepositorySnapshot.ChangedPaths, snapshotText(input.RepositorySnapshot), snapshotText(report.RepositorySnapshot)),
	}

	projection, wire, sourceErr := parseSource(input.SourceRaw)
	checks = append(checks, verifyCheck("source-reconstruction", "FOUNDATION", "parse-and-lower-source", "FOUNDATION", "reconstruct-source-wire", sourceErr == nil && projection.Exact && projection.CaseCount == CaseCount && projection.ObligationCount == DenominatorSize, "source-derived 5 obligations and 3 cases", sourceErrorText(projection, sourceErr)))
	if reportOK && sourceErr == nil {
		base, cases, ledger, emitted := sourceInputs(wire, input.RepositorySnapshot)
		sourceSpecs := make([]CaseSpec, 0, len(wire.Cases))
		for _, value := range wire.Cases {
			sourceSpecs = append(sourceSpecs, value.Spec)
		}
		verification.Guardrails = makeGuardrails(forbiddenEstimateObserved(emitted), input.RepositorySnapshot.ChangedPaths)
		verification.ClaimLedger = ledger
		checks = append(checks, verifySourceProjection(report, projection)...)
		checks = append(checks, verifyDenominator(report, contract, base)...)
		checks = append(checks, verifyCases(report, contract, sourceSpecs, cases, base, ledger)...)
		checks = append(checks, verifyClaims(report, ledger, emitted)...)
		checks = append(checks, verifySummary(report, len(cases), len(ledger), verification.Guardrails)...)
	} else {
		verification.Guardrails = makeGuardrails(forbiddenEstimateObserved(nil), input.RepositorySnapshot.ChangedPaths)
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

func sourceInputs(wire sourceWire, snapshot RepositorySnapshot) (Denominator, []Case, []ClaimLedgerEntry, []EmittedClaim) {
	base := Denominator{Version: wire.Version, Obligations: append([]Obligation(nil), wire.Obligations...)}
	base.Digest = denominatorDigest(base)
	legal := cloneDenominator(base, wire.SuccessorVersion)
	for _, change := range wire.Deletions {
		legal.Obligations = removeObligation(legal.Obligations, change.ObligationID)
	}
	if mutation, ok := mutationFor(wire, "legal-advance"); ok {
		legal.Obligations = append(legal.Obligations, mutation)
	}
	legal.Digest = denominatorDigest(legal)
	unauthorized := cloneDenominator(base, wire.SuccessorVersion)
	if mutation, ok := mutationFor(wire, "unauthorized-change"); ok {
		unauthorized.Obligations = append(unauthorized.Obligations, mutation)
	}
	unauthorized.Digest = denominatorDigest(unauthorized)
	unknown := cloneDenominator(base, wire.UnknownPredecessorVersion)
	unknown.Digest = denominatorDigest(unknown)
	unknownSuccessor := cloneDenominator(base, wire.UnknownSuccessorVersion)
	unknownSuccessor.Digest = denominatorDigest(unknownSuccessor)
	ledger := sealLedger(wire.Claims)
	guards := makeGuardrails(forbiddenEstimateObserved(wire.EmittedClaims), snapshot.ChangedPaths)
	receipt := Receipt{Schema: ReceiptSchema, ID: wire.ReceiptID, Producer: "denominatorevolution.Evaluate", Consumer: "denominatorevolutionverify.Verify", Predecessor: Ref{Version: base.Version, Digest: base.Digest}, Successor: Ref{Version: legal.Version, Digest: legal.Digest}, Additions: append([]Change(nil), wire.Additions...), Deletions: append([]Change(nil), wire.Deletions...), Decision: wire.ReceiptDecision, Reason: wire.ReceiptReason, Coordinate: wire.ReceiptCoordinate, RepositoryWrites: snapshot.ChangedPaths, MutationAuthority: false, RepositorySnapshot: snapshot, Guardrails: guards}
	receipt.Digest = receiptDigest(receipt)
	result := []Case{{ID: wire.Cases[0].Spec.ID, Kind: wire.Cases[0].Spec.Kind, ExpectedDecision: wire.Cases[0].Spec.ExpectedDecision, ExpectedResolution: wire.Cases[0].Spec.ExpectedResolution, ExpectedReason: wire.Cases[0].Spec.ExpectedReason, FromClaim: wire.Cases[0].Spec.FromClaim, ToClaim: wire.Cases[0].Spec.ToClaim, Predecessor: base, Successor: legal, Receipt: &receipt, Coordinate: Coordinate{Stage: wire.Cases[0].Spec.Stage, Step: wire.Cases[0].Spec.Step, Reason: wire.Cases[0].Spec.ExpectedReason}}, {ID: wire.Cases[1].Spec.ID, Kind: wire.Cases[1].Spec.Kind, ExpectedDecision: wire.Cases[1].Spec.ExpectedDecision, ExpectedResolution: wire.Cases[1].Spec.ExpectedResolution, ExpectedReason: wire.Cases[1].Spec.ExpectedReason, FromClaim: wire.Cases[1].Spec.FromClaim, ToClaim: wire.Cases[1].Spec.ToClaim, Predecessor: base, Successor: unauthorized, Coordinate: Coordinate{Stage: wire.Cases[1].Spec.Stage, Step: wire.Cases[1].Spec.Step, Reason: wire.Cases[1].Spec.ExpectedReason}}, {ID: wire.Cases[2].Spec.ID, Kind: wire.Cases[2].Spec.Kind, ExpectedDecision: wire.Cases[2].Spec.ExpectedDecision, ExpectedResolution: wire.Cases[2].Spec.ExpectedResolution, ExpectedReason: wire.Cases[2].Spec.ExpectedReason, FromClaim: wire.Cases[2].Spec.FromClaim, ToClaim: wire.Cases[2].Spec.ToClaim, Predecessor: unknown, Successor: unknownSuccessor, Coordinate: Coordinate{Stage: wire.Cases[2].Spec.Stage, Step: wire.Cases[2].Spec.Step, Reason: wire.Cases[2].Spec.ExpectedReason}}}
	return base, result, ledger, append([]EmittedClaim(nil), wire.EmittedClaims...)
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
func mutationFor(wire sourceWire, caseID string) (Obligation, bool) {
	for _, value := range wire.Mutations {
		if value.CaseID == caseID {
			return value.Obligation, true
		}
	}
	return Obligation{}, false
}
func sealLedger(values []ClaimLedgerEntry) []ClaimLedgerEntry {
	result := append([]ClaimLedgerEntry(nil), values...)
	previous := ""
	for index := range result {
		result[index].PreviousDigest = previous
		result[index].Digest = claimLedgerDigest(result[index])
		previous = result[index].Digest
	}
	return result
}
func claimLedgerDigest(value ClaimLedgerEntry) string { value.Digest = ""; return digestValue(value) }

func verifySourceProjection(report Report, expected SourceProjection) []Check {
	ok := reflect.DeepEqual(report.SourceProjection, expected)
	return []Check{verifyCheck("source-projection", "FOUNDATION", "bind-source-wire", "FOUNDATION", "compare-lowered-source", ok, "exact ParseFile -> bidir.Lower projection", sourceProjectionText(report.SourceProjection))}
}

func verifyDenominator(report Report, contract Contract, base Denominator) []Check {
	denominatorOK := reflect.DeepEqual(report.Denominator, base) && report.Denominator.Digest == denominatorDigest(report.Denominator)
	contractOK := contract.Denominator.Version == base.Version && reflect.DeepEqual(contract.Denominator.Obligations, base.Obligations)
	return []Check{verifyCheck("fixed-denominator", "FOUNDATION", "bind-fixed-denominator", "FOUNDATION", "compare-member-set", denominatorOK && contractOK, "5 source-derived members at "+DenominatorVersion, report.Denominator.Version+" / "+report.Denominator.Digest)}
}

func verifyCases(report Report, contract Contract, sourceSpecs []CaseSpec, expected []Case, base Denominator, ledger []ClaimLedgerEntry) []Check {
	if len(report.Cases) != CaseCount || len(contract.Cases) != CaseCount {
		return []Check{verifyCheck("case-cardinality", "FOUNDATION", "bind-case-corpus", "FOUNDATION", "compare-case-count", false, "3 source-derived cases", intText(len(report.Cases)))}
	}
	checks := []Check{}
	for index, value := range report.Cases {
		spec := contract.Cases[index]
		want := expected[index]
		decision, resolution, reason := independentDecision(want, base)
		fromClaim, toClaim := ledger[index].PriorState, ledger[index].NextState
		basic := value.ID == want.ID && value.Kind == want.Kind && value.Status == "SATISFIED" && spec.ID == want.ID && value.ExpectedDecision == want.ExpectedDecision && value.ExpectedResolution == want.ExpectedResolution && value.ExpectedReason == want.ExpectedReason && value.ObservedDecision == decision && value.ObservedResolution == resolution && value.ObservedReason == reason && value.FromClaim == fromClaim && value.ToClaim == toClaim && reflect.DeepEqual(value.Predecessor, want.Predecessor) && reflect.DeepEqual(value.Successor, want.Successor) && len(value.Checks) == 8
		checks = append(checks, verifyCheck(value.ID, spec.ProofChoice, spec.MetaOperation, spec.Stage, spec.Step, basic && index < len(sourceSpecs) && reflect.DeepEqual(spec, sourceSpecs[index]), spec.ExpectedDecision+" / "+spec.ExpectedReason, value.ObservedDecision+" / "+value.ObservedReason))
		if value.ID == "legal-advance" {
			checks = append(checks, verifyLegalReceipt(value, want, base)...)
		}
		if value.ID == "unauthorized-change" {
			checks = append(checks, verifyUnauthorized(value, decision, reason)...)
		}
		if value.ID == "unknown-predecessor" {
			checks = append(checks, verifyUnknown(value, decision, resolution, reason, fromClaim, toClaim)...)
		}
	}
	return checks
}

func independentDecision(value Case, base Denominator) (string, string, string) {
	predKnown := value.Predecessor.Version == base.Version && value.Predecessor.Digest == base.Digest && reflect.DeepEqual(value.Predecessor.Obligations, base.Obligations) && value.Predecessor.Digest == denominatorDigest(value.Predecessor)
	successorValid := value.Successor.Digest == denominatorDigest(value.Successor)
	if !predKnown {
		return "FAIL_CLOSED", "LOWER_RESOLUTION", "PREDECESSOR_DIGEST_UNKNOWN"
	}
	if value.Kind == "AUTHORIZED_MIGRATION" && value.Receipt != nil && value.Receipt.Decision == "ADVANCE" && value.Receipt.Digest == receiptDigest(*value.Receipt) && successorValid && guardrailsConform(value.Receipt.Guardrails) {
		return "ADVANCE", "EXACT", "DENOMINATOR_ADVANCE_AUTHORIZED"
	}
	return "BLOCK", "INVARIANT_ONLY", "MIGRATION_RECEIPT_MISSING"
}

func verifyLegalReceipt(value Case, want Case, base Denominator) []Check {
	receipt := value.Receipt
	ok := receipt != nil && reflect.DeepEqual(receipt, want.Receipt) && receipt.Predecessor.Version == base.Version && receipt.Successor.Version == value.Successor.Version && receipt.Decision == "ADVANCE" && receipt.Reason == "DENOMINATOR_ADVANCE_AUTHORIZED" && receipt.Digest == receiptDigest(*receipt) && sameChanges(receipt.Additions, want.Receipt.Additions) && sameChanges(receipt.Deletions, want.Receipt.Deletions)
	return []Check{verifyCheck("legal-migration-receipt", "COHERENCE", "accept-migration-receipt", "MIGRATE", "replay-receipt", ok, "source-bound v1 -> v2 receipt with explicit observations", receiptText(receipt))}
}
func verifyUnauthorized(value Case, decision, reason string) []Check {
	ok := value.Receipt == nil && decision == "BLOCK" && reason == "MIGRATION_RECEIPT_MISSING"
	return []Check{verifyCheck("unauthorized-no-receipt", "REGRESSION", "reject-unreceipted-denominator-change", "DECIDE", "reject-missing-receipt", ok, "no receipt and BLOCK", decision+" / "+receiptText(value.Receipt))}
}
func verifyUnknown(value Case, decision, resolution, reason, fromClaim, toClaim string) []Check {
	ok := value.Receipt == nil && value.Predecessor.Version != DenominatorVersion && decision == "FAIL_CLOSED" && resolution == "LOWER_RESOLUTION" && reason == "PREDECESSOR_DIGEST_UNKNOWN" && fromClaim == "OPEN" && toClaim == "OPEN"
	return []Check{verifyCheck("unknown-predecessor", "FOUNDATION", "fail-closed-unknown-predecessor", "RESOLVE", "lookup-predecessor", ok, "FAIL_CLOSED / LOWER_RESOLUTION / OPEN -> OPEN", decision+" / "+resolution+" / "+fromClaim+" -> "+toClaim)}
}

func verifyClaims(report Report, expected []ClaimLedgerEntry, emitted []EmittedClaim) []Check {
	ok := reflect.DeepEqual(report.ClaimLedger, expected) && reflect.DeepEqual(report.EmittedClaims, emitted)
	return []Check{verifyCheck("persistent-claim-ledger", "COHERENCE", "bind-persistent-claim-ledger", "DECIDE", "replay-open-state-transitions", ok, "3 sealed OPEN-origin ledger entries", intText(len(report.ClaimLedger))+" entries")}
}

func verifySummary(report Report, cases, ledger int, guards []Guardrail) []Check {
	summary := report.Summary
	ok := summary.CasesSatisfied == CaseCount && summary.CasesTotal == cases && summary.FixedDenominatorNumerator == DenominatorSize && summary.FixedDenominatorDenominator == DenominatorSize && summary.LegalAdvanceNumerator == 1 && summary.LegalAdvanceDenominator == 1 && summary.UnauthorizedRejectionNumerator == 1 && summary.UnauthorizedRejectionDenominator == 1 && summary.UnknownPredecessorNumerator == 1 && summary.UnknownPredecessorDenominator == 1 && summary.AdditionReasonNumerator == 1 && summary.AdditionReasonDenominator == 1 && summary.DeletionReasonNumerator == 1 && summary.DeletionReasonDenominator == 1 && summary.SourceCasesNumerator == cases && summary.SourceCasesDenominator == CaseCount && summary.PersistentClaimsNumerator == ledger && summary.PersistentClaimsDenominator == CaseCount && summary.GuardrailObservationsNumerator == guardrailCount && summary.GuardrailObservationsDenominator == guardrailCount && reflect.DeepEqual(summary.Guardrails, guards) && guardrailsConform(summary.Guardrails)
	return []Check{verifyCheck("exact-summary", "FOUNDATION", "bind-exact-counters", "SUMMARY", "compare-numerators-and-denominators", ok, "fixed 5/5, source cases 3/3, persistent claims 3/3, guardrails 2/2", "summary counters and explicit observations")}
}

func makeGuardrails(forbidden, writes int) []Guardrail {
	return []Guardrail{newGuardrail(guardrailForbiddenEstimate, forbidden, 0), newGuardrail(guardrailRepositoryWrites, writes, 0)}
}
func newGuardrail(id string, observed, allowed int) Guardrail {
	conforms := observed <= allowed
	numerator := 0
	if conforms {
		numerator = 1
	}
	return Guardrail{ID: id, Direction: "AT_MOST", Observed: observed, AllowedMax: allowed, ConformanceNumerator: numerator, ConformanceDenominator: 1, Conforms: conforms}
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
	return true
}
func sameChanges(left, right []Change) bool {
	if len(left) != len(right) {
		return false
	}
	sort.Slice(left, func(i, j int) bool { return left[i].ObligationID < left[j].ObligationID })
	sort.Slice(right, func(i, j int) bool { return right[i].ObligationID < right[j].ObligationID })
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
