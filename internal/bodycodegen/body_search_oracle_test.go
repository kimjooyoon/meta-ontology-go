package bodycodegen

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestIRBodySearchCallerCancellationStopsProviderRound(t *testing.T) {
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entered <- struct{}{}
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	defer server.Close()
	defer close(release)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	plan := irBodySearchPlan([]IRBodyFillCandidate{{ID: "identity", Expression: "input"}, {ID: "zero", Expression: "0"}},
		[]IRBodyFillTestCase{{Input: -1, Expected: 0}}, nil, 2)
	source := readIRBodySearchFixture(t)
	done := make(chan error, 1)
	go func() {
		_, err := GenerateWithIRBodySearch(ctx, "cancel.gooo", source, "ClampNegativeToZero", plan, server.URL+"/v1/systemone", "")
		done <- err
	}()
	select {
	case <-entered:
		cancel()
	case <-time.After(5 * time.Second):
		t.Fatal("provider round was not entered")
	}
	select {
	case err := <-done:
		var failure *IRBodySearchError
		if !errors.Is(err, context.Canceled) || !errors.As(err, &failure) || failure.Receipt.StopReason != "CALLER_CANCELLED" ||
			len(failure.Receipt.Attempts) != 0 || failure.Receipt.AttemptedCandidates != 0 || failure.Receipt.ProviderOperations != 1 {
			t.Fatalf("caller cancellation became a fallback or scored attempt: %v", err)
		}
		completeness := SearchFailureCompletenessReceipt("ClampNegativeToZero", source, err.Error(), failure.Receipt)
		for _, dimension := range completeness.Dimensions {
			if dimension.ID == "external_network_boundary" {
				if dimension.Status != "UNKNOWN" {
					t.Fatalf("canceled provider operation was reported as a proven no-network run: %#v", dimension)
				}
				return
			}
		}
		t.Fatal("external network boundary dimension is missing")
	case <-time.After(5 * time.Second):
		t.Fatal("caller cancellation did not stop the search")
	}
}

func TestIRBodySearchSharedProviderBudgetSkipsLaterNetworkRounds(t *testing.T) {
	var calls atomic.Int32
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	defer server.Close()
	defer close(release)
	plan := irBodySearchPlan([]IRBodyFillCandidate{
		{ID: "identity", Expression: "input"}, {ID: "zero", Expression: "0"}, {ID: "negate", Expression: "-input"},
	}, []IRBodyFillTestCase{{Input: -1, Expected: 0}, {Input: 1, Expected: 1}}, nil, 2)
	result, err := generateWithIRBodySearchBudget(context.Background(), "budget.gooo", readIRBodySearchFixture(t),
		"ClampNegativeToZero", plan, server.URL+"/v1/systemone", "", 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodySearch
	if calls.Load() != 1 || receipt.SelectedCandidateID != "zero" || len(receipt.Attempts) != 2 || receipt.ProviderOperations != 1 ||
		receipt.ProviderBudgetMS != 250 || receipt.ProviderBudgetUsedMS <= 0 {
		t.Fatalf("budget did not bound network rounds while preserving deterministic search: %#v, calls=%d", receipt, calls.Load())
	}
	for _, attempt := range receipt.Attempts {
		if attempt.Decision == nil || attempt.Decision.Mode == "laya" || attempt.Decision.FallbackReason != "SEARCH_PROVIDER_BUDGET_EXHAUSTED" {
			t.Fatalf("in-flight or skipped budget exhaustion was mislabeled: %#v", attempt)
		}
	}
}

func TestIRBodySearchProviderOperationsCountConfiguredResolveRounds(t *testing.T) {
	plan := irBodySearchPlan([]IRBodyFillCandidate{
		{ID: "identity", Expression: "input"}, {ID: "zero", Expression: "0"}, {ID: "negate", Expression: "-input"},
	}, []IRBodyFillTestCase{{Input: -1, Expected: 0}}, nil, 2)
	var requests []searchRequestSnapshot
	server := newIRBodySearchServer(t, func(call int, snapshot searchRequestSnapshot) string {
		requests = append(requests, snapshot)
		return []string{"identity", "zero"}[call]
	})
	defer server.Close()
	result, err := GenerateWithIRBodySearch(context.Background(), "provider-operations.gooo", readIRBodySearchFixture(t),
		"ClampNegativeToZero", plan, server.URL+"/v1/systemone", "")
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodySearch
	if len(requests) != 2 || receipt.ProviderOperations != 2 || len(receipt.Attempts) != 2 ||
		receipt.AttemptedCandidates != 2 || receipt.EvaluatedCandidates != 2 || receipt.SelectedCandidateID != "zero" {
		t.Fatalf("provider operation count does not match completed configured Resolve rounds: requests=%d receipt=%#v", len(requests), receipt)
	}
	for _, dimension := range result.Report.CompletenessReceipt.Dimensions {
		if dimension.ID == "external_network_boundary" {
			if dimension.Status != "UNKNOWN" {
				t.Fatalf("configured provider operations were reported as a proven no-network run: %#v", dimension)
			}
			return
		}
	}
	t.Fatal("external network boundary dimension is missing")
}

func TestIRBodySearchWhitespaceEndpointUsesNoProviderBudget(t *testing.T) {
	plan := irBodySearchPlan([]IRBodyFillCandidate{{ID: "identity", Expression: "input"}, {ID: "zero", Expression: "0"}},
		[]IRBodyFillTestCase{{Input: -1, Expected: 0}}, nil, 2)
	result, err := GenerateWithIRBodySearch(context.Background(), "blank-endpoint.gooo", readIRBodySearchFixture(t),
		"ClampNegativeToZero", plan, " \t\n ", "")
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodySearch
	if receipt.ProviderOperations != 0 || receipt.ProviderBudgetUsedMS != 0 || receipt.SelectedCandidateID != "zero" || len(receipt.Attempts) != 2 {
		t.Fatalf("whitespace-only endpoint consumed provider operations or budget: %#v", receipt)
	}
	if decision := receipt.Attempts[0].Decision; decision == nil || decision.Mode != "deterministic_fallback" || decision.FallbackReason != "NOT_CONFIGURED" {
		t.Fatalf("whitespace-only endpoint was not normalized to an unconfigured provider: %#v", decision)
	}
	for _, dimension := range result.Report.CompletenessReceipt.Dimensions {
		if dimension.ID == "external_network_boundary" {
			if dimension.Status != "PASS" {
				t.Fatalf("whitespace-only endpoint did not establish zero provider operations: %#v", dimension)
			}
			return
		}
	}
	t.Fatal("external network boundary dimension is missing")
}

func TestIRBodySearchFinalBodyMatchesCompiledGoOracle(t *testing.T) {
	source := []byte("package sample\nnamespace sample\nentity Integer id \"bodycodegen://entity/integer\"\n" +
		"activity Difference(Integer) -> Integer computes \"return (input - __GOOO_BODY_HOLE_floor__) * 2\"\n")
	plan := irBodySearchPlan([]IRBodyFillCandidate{
		{ID: "wrong", Expression: "input + 2"}, {ID: "right", Expression: "input + 1"},
	}, []IRBodyFillTestCase{{Input: -3, Expected: -2}, {Input: 0, Expected: -2}, {Input: 1, Expected: -2}},
		[]IRBodyFillTestCase{{Input: math.MinInt64, Expected: -2}, {Input: math.MaxInt64, Expected: -2}}, 2)
	result, err := GenerateWithIRBodySearch(context.Background(), "compound.gooo", source, "Difference", plan, "", "")
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodySearch
	if receipt.SelectedCandidateID != "right" || receipt.TrainingPassed != 3 || receipt.HoldoutPassed != 2 ||
		!slices.Equal(receipt.TrainingCaseResults, receipt.Attempts[1].CaseResults) {
		t.Fatalf("final emitted training evidence did not match the selected body: %#v", receipt)
	}
	oracleSource := "package main\nimport \"fmt\"\n" + withoutPackage(t, result.Source) + "\nfunc main() {\n"
	var want []string
	for _, testCase := range append(slices.Clone(receipt.TrainingCaseResults), receipt.HoldoutCaseResults...) {
		oracleSource += fmt.Sprintf("fmt.Println(Difference(%d))\n", testCase.Input)
		want = append(want, fmt.Sprint(testCase.Actual))
	}
	oracleSource += "}\n"
	if got := strings.TrimSpace(runBodyFillGoOracle(t, oracleSource)); got != strings.Join(want, "\n") {
		t.Fatalf("compiled Go diverged from the search score: %q, want %q", got, strings.Join(want, "\n"))
	}
}

func TestIntegerCaseEvaluatorHonorsCallerCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results, passed, err := evaluateIntegerCasesContext(ctx, []byte("invalid source must never be parsed"), "Never",
		[]IRBodyFillTestCase{{Input: 1, Expected: 1}})
	if !errors.Is(err, context.Canceled) || results != nil || passed != 0 {
		t.Fatalf("canceled evaluator did work: results=%v passed=%d err=%v", results, passed, err)
	}
}
