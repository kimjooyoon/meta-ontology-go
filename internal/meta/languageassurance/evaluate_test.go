package languageassurance

import "testing"

const testSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestAssuranceDecisionMatrix(t *testing.T) {
	tests := []struct {
		name                                                                  string
		selfMinting, roleConflict, missing, missingSnapshot, mismatchSnapshot bool
		input, output                                                         Decision
		decision, reason                                                      string
		wantSelf, wantRole, wantLaundering, wantUnknown                       int
		wantSnapshot, wantSnapshotPaths                                       int
	}{{name: "independent", decision: CandidateAllowLimited, reason: ReasonBoundaryClear}, {name: "self-minting", selfMinting: true, decision: CandidateBlock, reason: ReasonGovernanceViolation, wantSelf: 1}, {name: "role-conflict", roleConflict: true, decision: CandidateBlock, reason: ReasonGovernanceViolation, wantRole: 1}, {name: "unknown-fixed-point", input: DecisionUnknown, output: DecisionFixedPoint, decision: CandidateFailClosed, reason: ReasonTopDecisionUnknown, wantLaundering: 1, wantUnknown: 1}, {name: "unknown-block", input: DecisionUnknown, output: DecisionBlock, decision: CandidateFailClosed, reason: ReasonTopDecisionUnknown, wantUnknown: 1}, {name: "missing-evidence", missing: true, decision: CandidateFailClosed, reason: ReasonEvidenceUnknown, wantLaundering: -1, wantUnknown: -1}, {name: "missing-snapshot", missingSnapshot: true, decision: CandidateFailClosed, reason: ReasonEvidenceUnknown, wantSnapshot: -1, wantSnapshotPaths: -1}, {name: "snapshot-mismatch", mismatchSnapshot: true, decision: CandidateBlock, reason: ReasonSnapshotMismatch, wantSnapshot: 6666, wantSnapshotPaths: 1}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transaction := independentTransaction()
			if test.selfMinting {
				transaction.AuthorityRoutes[0].PromotedBy = "contract-author"
				transaction.RoleBindings[0].Roles = append(transaction.RoleBindings[0].Roles, RolePromoter)
			}
			if test.roleConflict {
				transaction.RoleBindings[1].Roles = append(transaction.RoleBindings[1].Roles, RoleImplementer)
			}
			if test.missing {
				transaction.DecisionTransitions = nil
			} else if test.input != "" {
				transaction.DecisionTransitions[0].Input, transaction.DecisionTransitions[0].Output = test.input, test.output
			}
			if test.missingSnapshot {
				transaction.SnapshotBindings = nil
			}
			if test.mismatchSnapshot {
				transaction.SnapshotBindings[2].SubjectSHA = "0000000000000000000000000000000000000000"
			}
			transaction = withTestReceipt(t, transaction)
			wantSnapshot := test.wantSnapshot
			if wantSnapshot == 0 {
				wantSnapshot = 10000
			}
			report := evaluateForTest(t, transaction)
			if report.AssuranceDecision != AssurancePartial || report.CandidateDecision != test.decision || report.CandidateReason != test.reason || report.Summary.Operating != 12 || report.Summary.DenominatorTotal != 12 || report.Summary.ImplementationCoverageBPS != 10000 {
				t.Fatalf("decision=%s/%s reason=%s coverage=%d/%d", report.AssuranceDecision, report.CandidateDecision, report.CandidateReason, report.Summary.Operating, report.Summary.DenominatorTotal)
			}
			if metricValue(report.Summary.SelfMintingPaths) != test.wantSelf || metricValue(report.Summary.RoleConflictPaths) != test.wantRole || metricValue(report.Summary.UnknownLaunderingPaths) != test.wantLaundering || metricValue(report.Summary.UnknownTopDecisions) != test.wantUnknown || metricValue(report.Summary.ExactSnapshotBindingBPS) != wantSnapshot || metricValue(report.Summary.SnapshotMismatchPaths) != test.wantSnapshotPaths {
				t.Fatalf("summary=%+v", report.Summary)
			}
		})
	}
}

func evaluateForTest(t *testing.T, transaction Transaction) Report {
	t.Helper()
	report, err := Evaluate(testSHA, transaction)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateForSubject(report, testSHA); err != nil {
		t.Fatal(err)
	}
	return report
}

func metricValue(value *int) int {
	if value == nil {
		return -1
	}
	return *value
}

func independentTransaction() Transaction {
	return Transaction{Schema: TransactionSchema, TransactionID: "independent", AuthorityRoutes: []AuthorityRoute{{RuleID: "promotion-v1", AuthoredBy: "contract-author", PromotedBy: "promoter"}}, RoleBindings: []RoleBinding{{Principal: "contract-author", Roles: []Role{RoleContractAuthor}}, {Principal: "promoter", Roles: []Role{RolePromoter}}}, DecisionTransitions: []DecisionTransition{{ID: "promotion", Input: DecisionPass, Output: DecisionAuthorized}}, SnapshotBindings: []SnapshotBinding{{EvidenceID: "authority_routes", SubjectSHA: testSHA}, {EvidenceID: "role_bindings", SubjectSHA: testSHA}, {EvidenceID: "decision_transitions", SubjectSHA: testSHA}}}
}
