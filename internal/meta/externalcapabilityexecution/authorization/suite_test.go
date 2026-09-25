package authorization

import "testing"

func TestAuthorizationConformanceSuite(t *testing.T) {
	suite := RunSuite(fixtureInput())
	if suite.Passed != 20 || suite.Total != 20 || suite.CoverageBPS != 10000 {
		t.Fatalf("unexpected suite coverage: %#v", suite)
	}
	if suite.AuthorizedCases != 1 || suite.UnknownCases != 7 || suite.DeniedCases != 12 {
		t.Fatalf("unexpected suite partition: %#v", suite)
	}
}

func TestKnownMismatchDenies(t *testing.T) {
	receipt := runCase(fixtureInput(), "operation-mismatch")
	if receipt.Decision != DecisionDenied || receipt.Resolution != ResolutionExact {
		t.Fatalf("known mismatch did not deny exactly: %#v", receipt)
	}
}
