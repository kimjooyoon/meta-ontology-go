package evidencequorum

import (
	"reflect"
	"sort"
)

const supportedValue = "SUPPORTS"

func Evaluate(input Input) Report {
	report := Report{
		Schema:         ReportSchema,
		Scope:          Scope,
		HeadSHA:        input.HeadSHA,
		SourcePath:     input.SourcePath,
		SourceDigest:   digestBytes(input.Source),
		ContractDigest: digestJSON(input.Contract),
		NotClaimed: []string{
			"confidence-weighted voting",
			"full Byzantine consensus",
			"full compiler semantic correctness",
			"identity or trust of an evidence producer",
			"repository mutation or side effects",
		},
	}
	for _, raw := range allReceiptBytes(input) {
		if receipt, err := DecodeReceipt(raw); err == nil && verifyReceipt(receipt) {
			report.ReceiptDigests = append(report.ReceiptDigests, receipt.Digest)
		}
	}
	sort.Strings(report.ReceiptDigests)
	for index, definition := range input.Contract.Cases {
		receipts := input.Receipts
		if len(input.CaseReceipts) == len(input.Contract.Cases) {
			receipts = input.CaseReceipts[index]
		}
		report.Cases = append(report.Cases, evaluateCase(input.Contract, definition, input, receipts))
	}
	report.Summary = summarize(report.Cases, input.Contract)
	report.Decision = DecisionClosed
	report.Resolution = ResolutionInvariant
	report.Reason = "EVIDENCE_QUORUM_CONTRACT_MISMATCH"
	if reflect.DeepEqual(input.Contract, CanonicalContract()) && validHead(input.HeadSHA) &&
		input.SourcePath == input.Contract.SourcePath && len(input.Source) > 0 &&
		report.Summary.CasesSatisfied == input.Contract.FixedCaseDenominator {
		report.Decision = DecisionPass
		report.Resolution = ResolutionExact
		report.Reason = "EVIDENCE_QUORUM_CONTRACT_SATISFIED"
	}
	report.Indicators = indicators(report)
	report.Proofs = proofs(report)
	report.Digest = reportDigest(report)
	return report
}

type evidenceGroup struct {
	ids    []string
	roles  []string
	values map[string]bool
}

func evaluateCase(contract Contract, definition CaseDefinition, input Input, receipts [][]byte) CaseResult {
	result := CaseResult{ID: definition.ID, Status: "NOT_SATISFIED",
		ExpectedDecision: definition.ExpectedDecision}
	claim := contract.Claim
	allEvidence := make([]Evidence, 0)
	receiptDigests := make([]string, 0)
	for _, raw := range receipts {
		receipt, err := DecodeReceipt(raw)
		if err != nil || !verifyReceipt(receipt) {
			return finishCase(result, definition, claim, StatusRefuted, ResolutionInvariant,
				"EVIDENCE_RECEIPT_INVALID", nil, receiptDigests, "receipt-integrity")
		}
		if !receiptMatches(contract, receipt, input) {
			return finishCase(result, definition, claim, StatusRefuted, ResolutionInvariant,
				"EVIDENCE_RECEIPT_PROVENANCE_MISMATCH", nil, receiptDigests, "receipt-provenance")
		}
		receiptDigests = append(receiptDigests, receipt.Digest)
		allEvidence = append(allEvidence, receipt.Evidence...)
	}
	sort.Slice(allEvidence, func(i, j int) bool { return allEvidence[i].ID < allEvidence[j].ID })
	result.RawEvidence = len(allEvidence)
	groups := map[string]*evidenceGroup{}
	for _, evidence := range allEvidence {
		if !evidenceMatches(contract, evidence, input) || evidence.Value != supportedValue && evidence.Value != "CONTRADICTS" {
			return finishCase(result, definition, claim, StatusRefuted, ResolutionInvariant,
				"EVIDENCE_VALUE_OR_PROVENANCE_INVALID", nil, receiptDigests, "evidence-integrity")
		}
		group := groups[evidence.OriginGroup]
		if group == nil {
			group = &evidenceGroup{values: map[string]bool{}}
			groups[evidence.OriginGroup] = group
		}
		group.ids = append(group.ids, evidence.ID)
		group.roles = appendUnique(group.roles, evidence.Role)
		group.values[evidence.Value] = true
	}
	result.IndependentGroups = len(groups)
	for _, group := range groups {
		result.DuplicateEvidence += len(group.ids) - 1
		if len(group.values) > 1 {
			result.ConflictGroups++
		}
	}
	result.Groups = groupResults(groups)
	if result.ConflictGroups > 0 || hasMultipleValues(groups) {
		return finishCase(result, definition, claim, StatusRefuted, ResolutionInvariant,
			"QUORUM_CONFLICT", result.Groups, receiptDigests, "conflicting-values")
	}
	if result.IndependentGroups < contract.MinimumIndependentGroups || !hasRequiredRoles(groups, contract.RequiredRoles) {
		return finishCase(result, definition, claim, StatusOpen, ResolutionLower,
			"QUORUM_INSUFFICIENT_INDEPENDENT_GROUPS", result.Groups, receiptDigests,
			"minimum-independent-groups")
	}
	return finishCase(result, definition, claim, StatusDischarged, ResolutionExact,
		"QUORUM_CLAIM_DISCHARGED", result.Groups, receiptDigests, "independent-group-count")
}

func allReceiptBytes(input Input) [][]byte {
	if len(input.CaseReceipts) != len(input.Contract.Cases) {
		return input.Receipts
	}
	var result [][]byte
	for _, receipts := range input.CaseReceipts {
		result = append(result, receipts...)
	}
	return result
}

func receiptMatches(contract Contract, receipt Receipt, input Input) bool {
	return receipt.Schema == ReceiptSchema && receipt.HeadSHA == input.HeadSHA &&
		receipt.SourcePath == input.SourcePath && receipt.SourceDigest == digestBytes(input.Source) &&
		receipt.Producer == contract.Claim.Producer && receipt.Consumer == contract.Claim.Consumer &&
		receipt.MetaOperation == contract.Claim.MetaOperation && receipt.ProofChoice == contract.Claim.ProofChoice &&
		receipt.Decision == DecisionPass && receipt.Resolution == ResolutionExact &&
		receipt.RepositoryWrites == 0 && !receipt.MutationAuthority
}

func evidenceMatches(contract Contract, evidence Evidence, input Input) bool {
	return evidence.ID != "" && evidence.ClaimID == contract.Claim.ID && evidence.OriginGroup != "" &&
		evidence.Role != "" && contains(contract.RequiredRoles, evidence.Role) &&
		evidence.Producer == contract.Claim.Producer && evidence.Consumer == contract.Claim.Consumer &&
		evidence.MetaOperation == contract.Claim.MetaOperation && evidence.ProofChoice == contract.Claim.ProofChoice &&
		evidence.SourcePath == input.SourcePath && evidence.SourceDigest == digestBytes(input.Source)
}

func finishCase(result CaseResult, definition CaseDefinition, claim ClaimDefinition, status, resolution, reason string,
	groups []GroupResult, receipts []string, step string) CaseResult {
	result.ExpectedResolution = definition.ExpectedResolution
	result.ExpectedReason = definition.ExpectedReason
	result.Groups = groups
	result.ObservedDecision = DecisionClosed
	result.ObservedResolution = resolution
	result.ObservedReason = reason
	result.Coordinate = Coordinate{Stage: "QUORUM_DECISION", Step: step, Reason: reason}
	if status == StatusDischarged {
		result.ObservedDecision = DecisionPass
	}
	result.Claims = []ClaimResult{{ID: claim.ID, Producer: claim.Producer, Consumer: claim.Consumer,
		MetaOperation: claim.MetaOperation, ProofChoice: claim.ProofChoice, Status: status,
		Reason: reason, Coordinate: result.Coordinate, EvidenceDigests: append([]string{}, receipts),
		Transitions: []ClaimTransition{{From: "OPEN", To: status, Coordinate: result.Coordinate}}}}
	if result.ObservedDecision == definition.ExpectedDecision && result.ObservedResolution == definition.ExpectedResolution &&
		result.ObservedReason == definition.ExpectedReason && status == definition.ExpectedStatus {
		result.Status = "SATISFIED"
	}
	return result
}

func groupResults(groups map[string]*evidenceGroup) []GroupResult {
	result := make([]GroupResult, 0, len(groups))
	for origin, group := range groups {
		values := make([]string, 0, len(group.values))
		for value := range group.values {
			values = append(values, value)
		}
		sort.Strings(group.ids)
		sort.Strings(group.roles)
		sort.Strings(values)
		result = append(result, GroupResult{OriginGroup: origin, EvidenceIDs: group.ids,
			Roles: group.roles, Values: values, Independent: true})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].OriginGroup < result[j].OriginGroup })
	return result
}

func hasMultipleValues(groups map[string]*evidenceGroup) bool {
	values := map[string]bool{}
	for _, group := range groups {
		for value := range group.values {
			values[value] = true
		}
	}
	return len(values) > 1
}

func hasRequiredRoles(groups map[string]*evidenceGroup, required []string) bool {
	roles := map[string]bool{}
	for _, group := range groups {
		for role := range group.roles {
			roles[role] = true
		}
	}
	for _, role := range required {
		if !roles[role] {
			return false
		}
	}
	return true
}

func appendUnique(values []string, value string) []string {
	if contains(values, value) {
		return values
	}
	return append(values, value)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
