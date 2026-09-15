package sourceauthorityupstream

import (
	"context"
	"errors"
	"strings"
	"testing"
)

const gomacroTitle = "## gomacro - interactive Go interpreter and debugger with generics and macros"

type staticFetcher struct {
	data []byte
	err  error
}

func (fetcher staticFetcher) Fetch(context.Context, string) ([]byte, error) {
	return fetcher.data, fetcher.err
}

func TestRunSuiteUsesFixedThreeCaseDenominator(t *testing.T) {
	suite := RunSuite(context.Background(), strings.Repeat("a", 40), staticFetcher{data: []byte(gomacroTitle + "\nignored")})
	if suite.Decision != "PASS" || suite.Resolution != ResolutionExact {
		t.Fatalf("decision=%s resolution=%s", suite.Decision, suite.Resolution)
	}
	if suite.Summary.CasesTotal != 3 || suite.Summary.CasesPassed != 3 || suite.Summary.CoverageBPS != 10000 {
		t.Fatalf("summary=%+v", suite.Summary)
	}
	if suite.Summary.ExactAllow != 1 || suite.Summary.FailClosed != 2 {
		t.Fatalf("case split=%+v", suite.Summary)
	}
	for _, result := range suite.Cases {
		if !result.Passed || result.Receipt.RepositoryWrites != 0 || result.Receipt.PromotionCreditBPS != 0 {
			t.Fatalf("case %s=%+v", result.ID, result)
		}
	}
}

func TestFetchFailureLowersResolutionAndBlocks(t *testing.T) {
	receipt := Observe(context.Background(), GomacroPolicy(), GomacroRequest(strings.Repeat("b", 40)), staticFetcher{err: errors.New("offline")})
	if receipt.Observation != ObservationUnknown || receipt.Resolution != ResolutionInvariantOnly {
		t.Fatalf("observation=%s resolution=%s", receipt.Observation, receipt.Resolution)
	}
	if receipt.Enforcement != EnforcementBlock || receipt.Reason != ReasonFetchFailed {
		t.Fatalf("enforcement=%s reason=%s", receipt.Enforcement, receipt.Reason)
	}
}

func TestUnknownReceiptNeverAllows(t *testing.T) {
	request := GomacroRequest(strings.Repeat("c", 40))
	request.Authority.Repository = "cosmos72/not-gomacro"
	receipt := Observe(context.Background(), GomacroPolicy(), request, staticFetcher{data: []byte(gomacroTitle)})
	if receipt.Observation != ObservationUnknown || receipt.Enforcement != EnforcementBlock {
		t.Fatalf("unknown was laundered: %+v", receipt)
	}
	if receipt.Reason != ReasonAuthorityScopeMismatch || receipt.Snapshot != nil {
		t.Fatalf("authority mismatch crossed fetch boundary: %+v", receipt)
	}
}
