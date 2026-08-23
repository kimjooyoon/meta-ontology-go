package languageassurance

import "testing"

const testSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestIndependentTransactionIsOnlyLimitedAllow(t *testing.T) {
	report := evaluateForTest(t, independentTransaction())
	if report.AssuranceDecision != AssurancePartial || report.CandidateDecision != CandidateAllowLimited {
		t.Fatalf("decisions = %s/%s", report.AssuranceDecision, report.CandidateDecision)
	}
	if report.Summary.Operating != 3 || report.Summary.DenominatorTotal != 12 || report.Summary.ImplementationCoverageBPS != 2500 {
		t.Fatalf("coverage = %d/%d (%d bps)", report.Summary.Operating, report.Summary.DenominatorTotal, report.Summary.ImplementationCoverageBPS)
	}
}

func TestSelfMintingPathBlocks(t *testing.T) {
	transaction := independentTransaction()
	transaction.AuthorityRoutes[0].PromotedBy = "contract-author"
	transaction.RoleBindings[0].Roles = append(transaction.RoleBindings[0].Roles, RolePromoter)
	report := evaluateForTest(t, transaction)
	if report.CandidateDecision != CandidateBlock || observed(t, report.Summary.SelfMintingPaths) != 1 {
		t.Fatalf("decision=%s self_minting=%d", report.CandidateDecision, observed(t, report.Summary.SelfMintingPaths))
	}
}

func TestRoleConflictPathBlocks(t *testing.T) {
	transaction := independentTransaction()
	transaction.RoleBindings[1].Roles = append(transaction.RoleBindings[1].Roles, RoleImplementer)
	report := evaluateForTest(t, transaction)
	if report.CandidateDecision != CandidateBlock || observed(t, report.Summary.RoleConflictPaths) != 1 {
		t.Fatalf("decision=%s role_conflicts=%d", report.CandidateDecision, observed(t, report.Summary.RoleConflictPaths))
	}
}

func TestUnknownTopDecisionFailsClosed(t *testing.T) {
	transaction := independentTransaction()
	transaction.DecisionTransitions[0].Input = DecisionUnknown
	transaction.DecisionTransitions[0].Output = DecisionFixedPoint
	report := evaluateForTest(t, transaction)
	if report.CandidateDecision != CandidateFailClosed || report.CandidateReason != ReasonTopDecisionUnknown {
		t.Fatalf("decision=%s reason=%s", report.CandidateDecision, report.CandidateReason)
	}
	if observed(t, report.Summary.UnknownLaunderingPaths) != 1 || observed(t, report.Summary.UnknownTopDecisions) != 1 {
		t.Fatal("unknown decision was not preserved")
	}
}

func TestUnknownToBlockStillFailsClosedWithoutLaundering(t *testing.T) {
	transaction := independentTransaction()
	transaction.DecisionTransitions[0].Input = DecisionUnknown
	transaction.DecisionTransitions[0].Output = DecisionBlock
	report := evaluateForTest(t, transaction)
	if report.CandidateDecision != CandidateFailClosed || observed(t, report.Summary.UnknownLaunderingPaths) != 0 {
		t.Fatalf("decision=%s laundering=%d", report.CandidateDecision, observed(t, report.Summary.UnknownLaunderingPaths))
	}
}

func TestMissingEvidenceIsUnknownNotZero(t *testing.T) {
	transaction := independentTransaction()
	transaction.DecisionTransitions = nil
	report := evaluateForTest(t, transaction)
	if report.CandidateDecision != CandidateFailClosed || report.Summary.UnknownLaunderingPaths != nil {
		t.Fatalf("decision=%s laundering=%v", report.CandidateDecision, report.Summary.UnknownLaunderingPaths)
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

func observed(t *testing.T, value *int) int {
	t.Helper()
	if value == nil {
		t.Fatal("value is unresolved")
	}
	return *value
}

func independentTransaction() Transaction {
	return Transaction{
		Schema: TransactionSchema, TransactionID: "independent",
		AuthorityRoutes: []AuthorityRoute{{RuleID: "promotion-v1", AuthoredBy: "contract-author", PromotedBy: "promoter"}},
		RoleBindings: []RoleBinding{
			{Principal: "contract-author", Roles: []Role{RoleContractAuthor}},
			{Principal: "promoter", Roles: []Role{RolePromoter}},
		},
		DecisionTransitions: []DecisionTransition{{ID: "promotion", Input: DecisionPass, Output: DecisionAuthorized}},
	}
}
