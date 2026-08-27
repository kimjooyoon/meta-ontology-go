package languageproofartifactverifier

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const caseEnvelopePolicyOperation = "SELECT_LOWEST_DECLARED_RANK"

type caseEnvelopePolicy struct {
	Observation CaseEnvelopePolicyObservation
	ranks       map[string]int
}

type caseEnvelopePolicyError struct {
	Step   string
	Reason string
	Detail string
}

func (e *caseEnvelopePolicyError) Error() string {
	return fmt.Sprintf("%s: %s", e.Reason, e.Detail)
}

// parseCaseEnvelopePolicy is the consumer's independent policy lens. It
// parses and lowers the actual .gooo source, then reads only the declared
// value programs. It never consults the validator's fixed expectation rows.
func parseCaseEnvelopePolicy(raw []byte) (caseEnvelopePolicy, error) {
	policy := caseEnvelopePolicy{Observation: CaseEnvelopePolicyObservation{RawSourceDigest: digestBytes(raw)}, ranks: map[string]int{}}
	if len(raw) == 0 {
		return policy, &caseEnvelopePolicyError{Step: "parse-policy", Reason: "CASE_ENVELOPE_POLICY_MISSING", Detail: "policy source is empty"}
	}
	file, diagnostics := syntax.ParseFile("proof-carrying-case-envelope-policy.gooo", string(raw))
	if file == nil || diagnostics.HasErrors() || file.Package == nil || file.Namespace == nil {
		return policy, &caseEnvelopePolicyError{Step: "parse-policy", Reason: "CASE_ENVELOPE_POLICY_MALFORMED", Detail: "policy source is not parseable"}
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return policy, &caseEnvelopePolicyError{Step: "lower-policy", Reason: "CASE_ENVELOPE_POLICY_MALFORMED", Detail: err.Error()}
	}
	var operation string
	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	policyOutputs := map[string]string{}
	for _, declaration := range declarations {
		if activity, ok := declaration.(*syntax.ActivityDecl); ok {
			policyOutputs[activity.Name] = activity.Output
		}
	}
	rows := make([]PolicyIssueRow, 0, 11)
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity {
			continue
		}
		switch {
		case strings.HasPrefix(node.ValueProgram, "proof.case-envelope.issue;"):
			if policyOutputs[node.Name] != "CaseEnvelopePolicy" {
				return policy, &caseEnvelopePolicyError{Step: "validate-policy", Reason: "CASE_ENVELOPE_POLICY_ROLE_MISMATCH", Detail: node.Name}
			}
			fields, fieldErr := parseCaseEnvelopeFields(node.ValueProgram, "issue", []string{"kind", "rank"})
			if fieldErr != nil {
				return policy, fieldErr
			}
			rank, parseErr := strconv.Atoi(fields["rank"])
			if parseErr != nil || rank < 1 {
				return policy, &caseEnvelopePolicyError{Step: "parse-policy", Reason: "CASE_ENVELOPE_POLICY_MALFORMED", Detail: "issue rank is not a positive integer"}
			}
			if !knownCaseEnvelopeIssueKind(fields["kind"]) {
				return policy, &caseEnvelopePolicyError{Step: "validate-policy", Reason: "CASE_ENVELOPE_POLICY_UNKNOWN_ISSUE", Detail: fields["kind"]}
			}
			rows = append(rows, PolicyIssueRow{Kind: fields["kind"], Rank: rank})
		case strings.HasPrefix(node.ValueProgram, "proof.case-envelope.reduction;"):
			if policyOutputs[node.Name] != "CaseEnvelopePolicy" {
				return policy, &caseEnvelopePolicyError{Step: "validate-policy", Reason: "CASE_ENVELOPE_POLICY_ROLE_MISMATCH", Detail: node.Name}
			}
			fields, fieldErr := parseCaseEnvelopeFields(node.ValueProgram, "reduction", []string{"operation", "missing", "duplicate", "unknown"})
			if fieldErr != nil {
				return policy, fieldErr
			}
			if operation != "" {
				return policy, &caseEnvelopePolicyError{Step: "validate-policy", Reason: "CASE_ENVELOPE_POLICY_DUPLICATE_REDUCTION", Detail: "multiple reduction operations are declared"}
			}
			if fields["operation"] != caseEnvelopePolicyOperation || fields["missing"] != "FAIL_CLOSED" || fields["duplicate"] != "FAIL_CLOSED" || fields["unknown"] != "FAIL_CLOSED" {
				return policy, &caseEnvelopePolicyError{Step: "validate-policy", Reason: "CASE_ENVELOPE_POLICY_MALFORMED", Detail: "reduction operation is not the fail-closed lowest-rank policy"}
			}
			operation = fields["operation"]
		case strings.HasPrefix(node.ValueProgram, "proof.case-envelope."):
			return policy, &caseEnvelopePolicyError{Step: "parse-policy", Reason: "CASE_ENVELOPE_POLICY_UNKNOWN_PROGRAM", Detail: node.ValueProgram}
		}
	}
	if len(rows) != CaseEnvelopePolicyRowTotal {
		return policy, &caseEnvelopePolicyError{Step: "validate-policy", Reason: "CASE_ENVELOPE_POLICY_ROW_COUNT_MISMATCH", Detail: fmt.Sprintf("got %d rows, want %d", len(rows), CaseEnvelopePolicyRowTotal)}
	}
	if operation == "" {
		return policy, &caseEnvelopePolicyError{Step: "validate-policy", Reason: "CASE_ENVELOPE_POLICY_REDUCTION_MISSING", Detail: "no reduction operation is declared"}
	}
	for _, row := range rows {
		if row.Kind == "" {
			return policy, &caseEnvelopePolicyError{Step: "validate-policy", Reason: "CASE_ENVELOPE_POLICY_MALFORMED", Detail: "issue kind is empty"}
		}
		if _, exists := policy.ranks[row.Kind]; exists {
			return policy, &caseEnvelopePolicyError{Step: "validate-policy", Reason: "CASE_ENVELOPE_POLICY_DUPLICATE_ISSUE", Detail: row.Kind}
		}
		for existingKind, existingRank := range policy.ranks {
			if existingRank == row.Rank {
				return policy, &caseEnvelopePolicyError{Step: "validate-policy", Reason: "CASE_ENVELOPE_POLICY_DUPLICATE_RANK", Detail: fmt.Sprintf("rank %d is used by %s and %s", row.Rank, existingKind, row.Kind)}
			}
		}
		if row.Rank > CaseEnvelopePolicyRowTotal {
			return policy, &caseEnvelopePolicyError{Step: "validate-policy", Reason: "CASE_ENVELOPE_POLICY_RANK_SET_MISMATCH", Detail: fmt.Sprintf("rank %d is outside 1..%d", row.Rank, CaseEnvelopePolicyRowTotal)}
		}
		policy.ranks[row.Kind] = row.Rank
	}
	for rank := 1; rank <= CaseEnvelopePolicyRowTotal; rank++ {
		found := false
		for _, declared := range policy.ranks {
			if declared == rank {
				found = true
				break
			}
		}
		if !found {
			return policy, &caseEnvelopePolicyError{Step: "validate-policy", Reason: "CASE_ENVELOPE_POLICY_RANK_SET_MISMATCH", Detail: fmt.Sprintf("rank %d is missing", rank)}
		}
	}
	sort.Slice(rows, func(left, right int) bool { return rows[left].Rank < rows[right].Rank })
	policy.Observation.IssueRows = rows
	policy.Observation.UniqueIssueRows = len(policy.ranks)
	policy.Observation.UniqueRankRows = len(policy.ranks)
	policy.Observation.SelectionOperation = operation
	policy.Observation.SemanticDigest = digestValue(struct {
		Rows      []PolicyIssueRow `json:"rows"`
		Operation string           `json:"operation"`
	}{rows, operation})
	return policy, nil
}

func knownCaseEnvelopeIssueKind(kind string) bool {
	switch kind {
	case "EXTERNAL_EVIDENCE_ABSENT", "OPERATION_EVIDENCE_MISSING", "OPERATION_ATTACHMENT_MISSING", "EVIDENCE_DIGEST_MISMATCH", "LEDGER_MISMATCH", "CLAIM_STRUCTURE_MISMATCH", "SOURCE_RECONSTRUCTION_MISMATCH", "OPERATION_ATTACHMENT_DIGEST_MISMATCH", "OPERATION_RECONSTRUCTION_MISMATCH", "INVARIANT_EVIDENCE_NOT_PRESERVED", "RECIPE_MISMATCH":
		return true
	default:
		return false
	}
}

func parseCaseEnvelopeFields(program, kind string, required []string) (map[string]string, error) {
	parts := strings.SplitN(program, ";", 2)
	if len(parts) != 2 || parts[0] != "proof.case-envelope."+kind {
		return nil, &caseEnvelopePolicyError{Step: "parse-policy", Reason: "CASE_ENVELOPE_POLICY_MALFORMED", Detail: "invalid " + kind + " meta-code"}
	}
	fields := map[string]string{}
	for _, field := range strings.Split(parts[1], ";") {
		key, value, ok := strings.Cut(field, "=")
		if !ok || key == "" || value == "" {
			return nil, &caseEnvelopePolicyError{Step: "parse-policy", Reason: "CASE_ENVELOPE_POLICY_MALFORMED", Detail: "invalid field " + field}
		}
		if _, exists := fields[key]; exists {
			return nil, &caseEnvelopePolicyError{Step: "parse-policy", Reason: "CASE_ENVELOPE_POLICY_DUPLICATE_FIELD", Detail: key}
		}
		fields[key] = value
	}
	for _, key := range required {
		if fields[key] == "" {
			return nil, &caseEnvelopePolicyError{Step: "parse-policy", Reason: "CASE_ENVELOPE_POLICY_MISSING_FIELD", Detail: key}
		}
	}
	allowed := map[string]bool{}
	for _, key := range required {
		allowed[key] = true
	}
	for key := range fields {
		if !allowed[key] {
			return nil, &caseEnvelopePolicyError{Step: "validate-policy", Reason: "CASE_ENVELOPE_POLICY_EXTRA_FIELD", Detail: key}
		}
	}
	return fields, nil
}

func policyObservation(policy caseEnvelopePolicy, issues []caseEnvelopeIssue, selected caseEnvelopeIssue) CaseEnvelopePolicyObservation {
	result := policy.Observation
	result.ObservedIssueSet = make([]string, 0, len(issues))
	seen := map[string]bool{}
	for _, issue := range issues {
		if issue.Kind != "" && !seen[issue.Kind] {
			result.ObservedIssueSet = append(result.ObservedIssueSet, issue.Kind)
			seen[issue.Kind] = true
		}
	}
	sort.Strings(result.ObservedIssueSet)
	result.ObservedIssueMembershipDigest = digestValue(result.ObservedIssueSet)
	result.ObservedIssueCount = len(result.ObservedIssueSet)
	if selected.Kind == "" || selected.Kind == "NO_CASE_ENVELOPE_ISSUE" {
		result.SelectedIssue = "NONE"
		result.SelectedRank = 0
	} else {
		result.SelectedIssue = selected.Kind
		result.SelectedRank = policy.ranks[selected.Kind]
	}
	return result
}

func policyFailureObservation(raw []byte) CaseEnvelopePolicyObservation {
	return CaseEnvelopePolicyObservation{RawSourceDigest: digestBytes(raw)}
}

func policyObservationShapeOK(observation CaseEnvelopePolicyObservation, sourceDigest string) bool {
	return policyRowsShapeOK(observation, sourceDigest) && observation.SelectedIssue == "NONE" && observation.SelectedRank == 0 &&
		len(observation.ObservedIssueSet) == 0 && observation.ObservedIssueCount == 0 && observation.ObservedIssueMembershipDigest == digestValue([]string{})
}

func policyRowsShapeOK(observation CaseEnvelopePolicyObservation, sourceDigest string) bool {
	if observation.RawSourceDigest != sourceDigest || !validDigest(observation.RawSourceDigest) || !validDigest(observation.SemanticDigest) ||
		len(observation.IssueRows) != CaseEnvelopePolicyRowTotal || observation.UniqueIssueRows != CaseEnvelopePolicyRowTotal || observation.UniqueRankRows != CaseEnvelopePolicyRowTotal ||
		observation.SelectionOperation != caseEnvelopePolicyOperation {
		return false
	}
	kinds := map[string]bool{}
	ranks := map[int]bool{}
	for _, row := range observation.IssueRows {
		if row.Kind == "" || kinds[row.Kind] || row.Rank < 1 || row.Rank > CaseEnvelopePolicyRowTotal || ranks[row.Rank] {
			return false
		}
		kinds[row.Kind], ranks[row.Rank] = true, true
	}
	return len(kinds) == CaseEnvelopePolicyRowTotal && len(ranks) == CaseEnvelopePolicyRowTotal
}

func policyRank(policy caseEnvelopePolicy, kind string) (int, bool) {
	rank, ok := policy.ranks[kind]
	return rank, ok
}

// CheckCaseEnvelopePolicy is a small CI-facing probe for policy fixtures. It
// exposes the same parser/lowerer and selection boundary used by the
// consumer, including typed fail-closed coordinates for malformed policy and
// observed issue kinds absent from the source declaration.
func CheckCaseEnvelopePolicy(raw []byte, observedIssue string) (CaseEnvelopePolicyObservation, Coordinate, error) {
	policy, err := parseCaseEnvelopePolicy(raw)
	if err != nil {
		coordinate := Coordinate{"CONSUME_POLICY", "parse-policy", "CASE_ENVELOPE_POLICY_INVALID"}
		var policyErr *caseEnvelopePolicyError
		if errors.As(err, &policyErr) {
			coordinate = Coordinate{"CONSUME_POLICY", policyErr.Step, policyErr.Reason}
		}
		return policy.Observation, coordinate, err
	}
	if observedIssue != "" {
		if _, ok := policyRank(policy, observedIssue); !ok {
			err := &caseEnvelopePolicyError{Step: "select-issue", Reason: "CASE_ENVELOPE_POLICY_UNKNOWN_ISSUE", Detail: observedIssue}
			return policyObservation(policy, []caseEnvelopeIssue{{Kind: observedIssue}}, caseEnvelopeIssue{Kind: "CASE_ENVELOPE_POLICY_UNKNOWN_ISSUE"}), Coordinate{"CONSUME_POLICY", err.Step, err.Reason}, err
		}
		issue := caseEnvelopeIssue{Kind: observedIssue}
		return policyObservation(policy, []caseEnvelopeIssue{issue}, issue), Coordinate{}, nil
	}
	return policyObservation(policy, nil, caseEnvelopeIssue{}), Coordinate{}, nil
}

// PolicyFixtureObservation is emitted by the CI-only fixture path. It is
// produced by the same artifact verifier kernel as normal cases, with an
// optional synthetic observed issue used only to exercise the unknown-issue
// reduction boundary.
type PolicyFixtureObservation struct {
	Decision              string                        `json:"decision"`
	Resolution            string                        `json:"resolution"`
	Reason                string                        `json:"reason"`
	Coordinate            Coordinate                    `json:"coordinate"`
	ArtifactDigest        string                        `json:"artifact_digest"`
	Policy                CaseEnvelopePolicyObservation `json:"policy"`
	PolicyClaimStatus     string                        `json:"policy_claim_status"`
	PolicyClaimResolution string                        `json:"policy_claim_resolution"`
	PolicyClaimReason     string                        `json:"policy_claim_reason"`
	PolicyClaimCoordinate Coordinate                    `json:"policy_claim_coordinate"`
	PolicyClaimEvidence   []string                      `json:"policy_claim_evidence_digests"`
	ReceiptDigest         string                        `json:"receipt_digest"`
}

func policyFixtureReceiptDigest(receipt PolicyFixtureObservation) string {
	receipt.ReceiptDigest = ""
	return digestValue(receipt)
}

// VerifyCaseEnvelopePolicyFixture runs the complete verifyArtifact path so a
// malformed policy cannot be accepted by a detached parser probe. The
// observedIssue parameter is only for the unknown-observed-issue fixture and
// is injected before the same policy reducer used by production verification.
func VerifyCaseEnvelopePolicyFixture(raw, source, operation, recipe []byte, head, observedIssue string) PolicyFixtureObservation {
	result := verifyArtifactWithObservedIssue(raw, source, operation, recipe, head, ProofPhasePreliminary, observedIssue)
	fixture := PolicyFixtureObservation{Decision: result.Decision, Resolution: result.Resolution, Reason: result.Reason,
		Coordinate: result.Coordinate, ArtifactDigest: result.ArtifactDigest, Policy: result.Policy}
	for _, claim := range result.Claims {
		if claim.ID == "case-envelope-policy-bound" {
			fixture.PolicyClaimStatus = claim.Status
			fixture.PolicyClaimResolution = claim.Resolution
			fixture.PolicyClaimReason = claim.Reason
			fixture.PolicyClaimCoordinate = claim.Coordinate
			fixture.PolicyClaimEvidence = append([]string(nil), claim.EvidenceDigests...)
			break
		}
	}
	fixture.ReceiptDigest = policyFixtureReceiptDigest(fixture)
	return fixture
}
