package externaloraclehumility

import "strings"

const (
	statusSatisfied   = "SATISFIED"
	statusUnsatisfied = "UNSATISFIED"
	statusUnknown     = "UNKNOWN"
	agreementYes      = "AGREES"
	agreementNo       = "DISAGREES"
	agreementUnknown  = "UNKNOWN"
)

type inspected struct {
	sourceReceipt, sourceDeclarations, receiptRoles string
	references                                      [3]string
	relations, agreement, authority, transitions    string
	agreementState, reason                          string
}

func Judge(input Input) Report {
	state := inspect(input)
	agreement := state.agreementState
	decision, resolution, effect, reason := decisionFor(state)
	transitions := makeTransitions(agreement)
	indicators := makeIndicators(state, decision, transitions)
	report := Report{
		Schema: ReportSchema, SubjectSHA: input.Subject, Decision: decision,
		Resolution: resolution, Reason: reason, ReferenceAgreement: agreement,
		SemanticAuthority: input.Contract.Source.Authority, AuthorityGrant: "NONE",
		EnforcementEffect: effect, Total: len(indicators), Indicators: indicators,
		Transitions: transitions, Receipt: input.Receipt,
		Producer: "independent-judge", Consumer: "semantic-authority-governor",
		MetaOperation: "separate-reference-agreement-from-authority",
		ProofChoice:   "REGRESSION", Stage: "govern", Step: "authority-boundary",
	}
	for _, indicator := range indicators {
		switch indicator.Status {
		case statusSatisfied:
			report.Completed++
		case statusUnknown:
			report.UnknownIndicators++
		}
	}
	report.BasisPoints = report.Completed * 10000 / report.Total
	report.ReportDigest = ""
	report.ReportDigest = Digest(report)
	return report
}

func inspect(input Input) inspected {
	state := inspected{agreementState: agreementUnknown}
	contractOK := validateContract(input.Contract) == nil
	sourceDigest := DigestBytes(input.Source)
	state.sourceDeclarations = statusFor(contractOK && sourceDigest == input.Contract.Source.SHA256 && sourceShapeExact(input.Source), "SOURCE_DECLARATIONS_EXACT")
	state.sourceReceipt = statusFor(contractOK && receiptExact(input.Receipt, input.Subject, input.Contract, sourceDigest), "SOURCE_RECEIPT_EXACT")
	state.receiptRoles = statusFor(receiptRolesExact(input.Receipt), "RECEIPT_ROLES_EXACT")
	state.references, state.relations, state.agreement, state.agreementState, state.reason = inspectReferences(input.Contract, input.Evidence)
	state.authority = statusFor(input.Contract.Source.Authority == "GOOO_SOURCE_INTENT" && input.Evidence.Schema == EvidenceSchema, "SOURCE_AUTHORITY_RETAINED")
	state.transitions = statusSatisfied
	return state
}

func statusFor(ok bool, _ string) string {
	if ok {
		return statusSatisfied
	}
	return statusUnsatisfied
}

func decisionFor(state inspected) (string, string, string, string) {
	if state.sourceReceipt != statusSatisfied || state.sourceDeclarations != statusSatisfied || state.receiptRoles != statusSatisfied {
		return "FAIL_CLOSED", "EXACT", "BLOCK", "SOURCE_RECEIPT_OR_DECLARATIONS_NOT_BOUND"
	}
	switch state.agreementState {
	case agreementYes:
		return "REFERENCE_AGREEMENT_OBSERVED", "EXACT", "NO_EFFECT", "EXTERNAL_AGREEMENT_IS_COMPARATIVE_ONLY"
	case agreementNo:
		return "FAIL_CLOSED", "EXACT", "BLOCK", state.reason
	default:
		return "FAIL_CLOSED", "UNKNOWN", "BLOCK", "EXTERNAL_REFERENCE_ABSENT_FAILS_CLOSED"
	}
}

func sourceShapeExact(source []byte) bool {
	text := string(source)
	fragments := []string{
		"package externaloraclehumility\n",
		"namespace externaloraclehumility\n",
		`entity SemanticSubject id "gooo://external-oracle-humility/entity/semantic-subject"`,
		`activity ObserveGoooSource(SemanticSubject) -> SourceReceipt`,
		`activity CompareExternalReference(SourceReceipt) -> ReferenceAgreement`,
		`activity RefuseExternalSemanticAuthority(ReferenceAgreement) -> AuthorityBoundary`,
		`activity PersistClaimTransition(AuthorityBoundary) -> ClaimTransition`,
		`activity EmitReadOnlyHumilityReceipt(ClaimTransition) -> HumilityReceipt`,
	}
	for _, fragment := range fragments {
		if !strings.Contains(text, fragment) {
			return false
		}
	}
	return true
}

func receiptExact(receipt SourceReceipt, subject string, contract Contract, sourceDigest string) bool {
	return receipt.Schema == ReceiptSchema && receipt.SubjectSHA == subject &&
		receipt.SourcePath == contract.Source.Path && receipt.SourceSHA256 == sourceDigest &&
		receipt.Claims != nil && sameClaims(receipt.Claims, contract.Source.Claims)
}

func sameClaims(got, want []Claim) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range want {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

func receiptRolesExact(receipt SourceReceipt) bool {
	return receipt.Producer == "source-receipt-producer" && receipt.Consumer == "independent-judge" &&
		receipt.MetaOperation == "emit-source-receipt" && receipt.ProofChoice == "FOUNDATION" &&
		receipt.Stage == "observe" && receipt.Step == "source-receipt" && receipt.Reason == "GOOO_SOURCE_INTENT_BOUND"
}

func inspectReferences(contract Contract, evidence ReferenceEvidenceSet) ([3]string, string, string, string, string) {
	var statuses [3]string
	if evidence.Schema != EvidenceSchema {
		return statuses, statusUnsatisfied, statusUnsatisfied, agreementUnknown, "EXTERNAL_REFERENCE_SET_SCHEMA_MISMATCH"
	}
	if len(evidence.References) != len(contract.References) {
		return statuses, statusUnsatisfied, statusUnsatisfied, agreementNo, "EXTERNAL_REFERENCE_SET_CARDINALITY_MISMATCH"
	}
	byID := make(map[string]ReferenceEvidence, len(evidence.References))
	for _, reference := range evidence.References {
		if _, exists := byID[reference.ID]; exists {
			return statuses, statusUnsatisfied, statusUnsatisfied, agreementNo, "EXTERNAL_REFERENCE_SET_DUPLICATE_ID"
		}
		byID[reference.ID] = reference
	}
	anyUnknown, anyMismatch, allAgree := false, false, true
	for index, expected := range contract.References {
		actual, ok := byID[expected.ID]
		if !ok || !actual.Available {
			statuses[index], anyUnknown = statusUnknown, true
			allAgree = false
			continue
		}
		if actual.DocumentSHA256 != expected.DocumentSHA256 || actual.Relation != expected.Relation ||
			actual.Authority != expected.Authority || !actual.Agreement || actual.Proposition == "" || actual.EvidenceRole == "" {
			statuses[index], anyMismatch = statusUnsatisfied, true
			allAgree = false
			continue
		}
		statuses[index] = statusSatisfied
	}
	if anyUnknown {
		return statuses, statusFor(!anyMismatch, "REFERENCE_RELATIONS"), statusUnknown, agreementUnknown, "EXTERNAL_REFERENCE_EVIDENCE_UNAVAILABLE"
	}
	if anyMismatch {
		return statuses, statusUnsatisfied, statusSatisfied, agreementNo, "EXTERNAL_REFERENCE_DISAGREEMENT_OBSERVED"
	}
	if allAgree {
		return statuses, statusSatisfied, statusSatisfied, agreementYes, "EXTERNAL_REFERENCE_AGREEMENT_OBSERVED"
	}
	return statuses, statusUnknown, statusUnknown, agreementUnknown, "EXTERNAL_REFERENCE_STATE_UNKNOWN"
}

func makeTransitions(agreement string) []ClaimTransition {
	agreementAfter, agreementReason := "UNKNOWN", "reference absence remains unresolved"
	if agreement == agreementYes {
		agreementAfter, agreementReason = "AGREEMENT_OBSERVED", "primary references agree as comparison evidence"
	} else if agreement == agreementNo {
		agreementAfter, agreementReason = "DISAGREEMENT_OBSERVED", "reference mismatch is recorded without promotion"
	}
	return []ClaimTransition{
		{"reference-agreement", "compare", "reference-set", "OPEN", agreementAfter, agreementReason, "independent-judge", "claim-ledger", "classify-reference-agreement", "COHERENCE", true},
		{"source-intent-authority", "govern", "authority-boundary", "SOURCE_ONLY", "SOURCE_ONLY", "external evidence cannot promote Gooo source authority", "independent-judge", "semantic-authority-governor", "refuse-external-semantic-authority", "REGRESSION", true},
		{"official-semantic-state", "persist", "official-state", "UNCHANGED", "UNCHANGED", "read-only comparison has no official mutation", "independent-judge", "official-state", "preserve-official-semantic-state", "REGRESSION", true},
	}
}

func makeIndicators(state inspected, decision string, transitions []ClaimTransition) []Indicator {
	statuses := []string{state.sourceReceipt, state.sourceDeclarations, state.receiptRoles,
		state.references[0], state.references[1], state.references[2], state.relations,
		state.agreement, state.authority, statusFor(allPersisted(transitions), "TRANSITIONS"),
		statusSatisfied, statusFor(decision != "PASS", "NO_PASS"),
	}
	reasons := []string{"SOURCE_RECEIPT_EXACT", "SOURCE_DECLARATIONS_REREAD", "PRODUCER_CONSUMER_BOUND",
		"GOMACRO_REFERENCE_BOUND", "RACKET_REFERENCE_BOUND", "REPRODUCIBLE_BUILDS_REFERENCE_BOUND",
		"COMPARATIVE_RELATIONS_CHECKED", "REFERENCE_AGREEMENT_CLASSIFIED", "EXTERNAL_AUTHORITY_REFUSED",
		"CLAIM_TRANSITION_PERSISTED", "JUDGE_IS_READ_ONLY", "PASS_PROMOTION_REFUSED"}
	result := make([]Indicator, len(denominator))
	for index, criterion := range denominator {
		result[index] = Indicator{ID: criterion.ID, Class: criterion.Class, ProofChoice: criterion.ProofChoice,
			Producer: criterion.Producer, Consumer: criterion.Consumer, MetaOperation: criterion.MetaOperation,
			Stage: criterion.Stage, Step: criterion.Step, Unit: criterion.Unit, Relation: criterion.Relation,
			Status: statuses[index], Reason: reasons[index], Value: indicatorValue(statuses[index], criterion.Target), Target: criterion.Target}
	}
	return result
}

func indicatorValue(status string, target int) int {
	if status == statusSatisfied {
		return target
	}
	return 0
}

func allPersisted(transitions []ClaimTransition) bool {
	if len(transitions) != 3 {
		return false
	}
	for _, transition := range transitions {
		if !transition.Persisted {
			return false
		}
	}
	return true
}
