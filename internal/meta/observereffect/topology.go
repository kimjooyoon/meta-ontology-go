package observereffect

import (
	"os"
	"path/filepath"
	"strings"
)

type topologyExpectation struct {
	Path             string
	Workflow         string
	Upstream         string
	TriggerBlock     string
	PullRequestBlock string
	Concurrency      string
}

var topologyExpectations = []topologyExpectation{
	{
		Path:             ".github/workflows/transformation-effect.yml",
		Workflow:         "transformation-effect",
		Upstream:         "CI",
		TriggerBlock:     "workflow_run:\n    workflows: [CI]\n    types: [completed]\n    branches: [dev, main]",
		PullRequestBlock: "pull_request:\n    branches: [dev, main]",
		Concurrency:      "transformation-effect-${{ github.event_name }}-${{ github.event.pull_request.number || github.event.workflow_run.head_branch }}",
	},
	{
		Path:         ".github/workflows/self-improvement-cycle.yml",
		Workflow:     "self-improvement-cycle",
		Upstream:     "CI",
		TriggerBlock: "workflow_run:\n    workflows: [\"CI\"]\n    types: [completed]\n    branches: [dev, main]",
		Concurrency:  "self-improvement-cycle-${{ github.event.workflow_run.head_branch }}",
	},
	{
		Path:         ".github/workflows/source-subject-witness.yml",
		Workflow:     "source-subject-witness",
		Upstream:     "CI",
		TriggerBlock: "workflow_run:\n    workflows: [\"CI\"]\n    types: [completed]\n    branches: [dev, main]",
		Concurrency:  "source-subject-witness-${{ github.event.workflow_run.head_branch }}",
	},
	{
		Path:         ".github/workflows/metric-transition.yml",
		Workflow:     "metric-transition",
		Upstream:     "CI",
		TriggerBlock: "workflow_run:\n    workflows: [CI]\n    types: [completed]\n    branches: [dev, main]",
		Concurrency:  "metric-transition-${{ github.event.workflow_run.head_branch }}",
	},
	{
		Path:             ".github/workflows/self-improvement-language-observation.yml",
		Workflow:         "self-improvement-language-observation",
		Upstream:         "Language example experiment",
		TriggerBlock:     "workflow_run:\n    workflows: [\"Language example experiment\"]\n    types: [completed]\n    branches: [dev]",
		PullRequestBlock: "pull_request:\n    branches: [dev]",
		Concurrency:      "self-improvement-language-observation-${{ github.event_name }}-${{ github.event.pull_request.number || github.event.workflow_run.head_branch }}",
	},
}

func buildTopology(root string) TopologyEvidence {
	subscribers := make([]TopologySubscriber, 0, len(topologyExpectations))
	filtered := 0
	exact := true
	for _, expected := range topologyExpectations {
		payload, err := os.ReadFile(filepath.Join(root, expected.Path))
		actual := "missing"
		status := "FAIL"
		reason := "TRIGGER_TOPOLOGY_MISMATCH_FAIL_CLOSED"
		if err == nil {
			actual = "workflow_run branch filter and concurrency present"
			triggerOK := strings.Contains(string(payload), expected.TriggerBlock) && strings.Contains(string(payload), "cancel-in-progress: true")
			pullRequestOK := expected.PullRequestBlock == "" || strings.Contains(string(payload), expected.PullRequestBlock)
			concurrencyOK := strings.Contains(string(payload), "group: "+expected.Concurrency)
			if triggerOK && pullRequestOK && concurrencyOK {
				status = "PASS"
				reason = "BRANCH_FILTER_AND_IDENTITY_CONCURRENCY_EXACT"
				filtered++
			}
		}
		if status != "PASS" {
			exact = false
		}
		subscribers = append(subscribers, TopologySubscriber{
			Workflow: expected.Workflow, Upstream: expected.Upstream,
			Expected: expected.TriggerBlock + "; group: " + expected.Concurrency,
			Actual:   actual, Concurrency: expected.Concurrency, Status: status,
			Producer: "observer-effect-ledger", Consumer: "observer-effect-judge",
			MetaOperation: "audit-workflow-run-topology", ProofChoice: "REGRESSION",
			Stage: "AUDIT", Step: "inspect-workflow-trigger", Reason: reason,
		})
	}
	return TopologyEvidence{
		Scope:                         "STATIC_TRIGGER_TOPOLOGY",
		WorkflowRunSubscribersAudited: len(subscribers), WorkflowRunSubscribersExpected: 5,
		BranchFilteredSubscribersBefore: 0,
		BranchFilteredSubscribers:       filtered, BranchFilteredSubscribersExpected: 5,
		DuplicatePROObservationPathsBefore: 2, DuplicatePROObservationPathsAfter: 1,
		ExpectedSkippedCIChildRunsPerPRCompletionBefore: 4,
		ExpectedSkippedCIChildRunsPerPRCompletionAfter:  0,
		Subscribers: subscribers,
		CausalEdges: topologyCausalEdges(), Exact: exact && len(subscribers) == 5 && filtered == 5,
		Producer: "observer-effect-ledger", Consumer: "observer-effect-judge",
		MetaOperation: "audit-trigger-topology", ProofChoice: "REGRESSION",
		Stage: "AUDIT", Step: "compare-trigger-graph", Reason: topologyReason(exact),
	}
}

func topologyCausalEdges() []CausalEdge {
	return []CausalEdge{
		{
			ID: "ci-completion-to-skipped-child", From: "CI.workflow_run.completed",
			To: "downstream.workflow_run.object", Relation: "branch-filter-removes-pr-fanout",
			Before: 4, After: 0, Producer: "observer-effect-ledger", Consumer: "observer-effect-judge",
			MetaOperation: "explain-ci-child-run-suppression", ProofChoice: "COHERENCE",
			Stage: "AUDIT", Step: "trace-ci-completion-edge", Reason: "PR_COMPLETION_IS_NOT_DEV_OR_MAIN_PUSH",
		},
		{
			ID: "language-pr-to-observation", From: "Language example experiment PR completion",
			To:       "self-improvement-language-observation",
			Relation: "direct-pr-path-retained-workflow-run-pr-path-filtered",
			Before:   2, After: 1, Producer: "observer-effect-ledger", Consumer: "observer-effect-judge",
			MetaOperation: "explain-language-observation-deduplication", ProofChoice: "COHERENCE",
			Stage: "AUDIT", Step: "trace-language-observation-edge", Reason: "DEV_BRANCH_FILTER_LEAVES_DIRECT_PULL_REQUEST_PATH",
		},
		{
			ID: "same-pr-stale-commit-to-cancellation", From: "same pull_request identity",
			To: "pull_request concurrency group", Relation: "cancel-in-progress",
			Before: 0, After: 1, Producer: "observer-effect-ledger", Consumer: "observer-effect-judge",
			MetaOperation: "explain-stale-pr-cancellation", ProofChoice: "REGRESSION",
			Stage: "AUDIT", Step: "trace-pr-concurrency-edge", Reason: "PR_NUMBER_SEPARATES_DIFFERENT_PRS",
		},
		{
			ID: "same-branch-stale-run-to-cancellation", From: "same workflow_run head_branch",
			To: "workflow_run concurrency group", Relation: "cancel-in-progress",
			Before: 0, After: 1, Producer: "observer-effect-ledger", Consumer: "observer-effect-judge",
			MetaOperation: "explain-stale-branch-cancellation", ProofChoice: "REGRESSION",
			Stage: "AUDIT", Step: "trace-workflow-run-concurrency-edge", Reason: "HEAD_BRANCH_SEPARATES_DEV_AND_MAIN",
		},
	}
}

func topologyReason(exact bool) string {
	if exact {
		return "FIVE_SUBSCRIBERS_AND_FIVE_BRANCH_FILTERS_EXACT"
	}
	return "TRIGGER_TOPOLOGY_MISMATCH_FAIL_CLOSED"
}

func runnerScopedEvidence() RunnerScopedEvidence {
	evidence := RunnerScopedEvidence{
		Scope: "RUNNER_SCOPED", Classification: "HISTORICAL_FIXTURE", Status: "OPEN",
		Source:         "review-supplied historical Actions API snapshot",
		ObservationRef: "dev SHA #540 latest 100 workflow_run objects",
		ObservedAt:     "UNKNOWN", Query: "NOT_CAPTURED", SubjectSHA: "dev SHA #540",
		SkippedWorkflowRuns: 59, QueuedWorkflowRuns: 41,
		TimeDependent: true, CurrentEvidence: false, IncludedInFixedDenominator: false,
		Producer: "review-fixture", Consumer: "observer-effect-judge",
		MetaOperation: "classify-historical-runner-snapshot", ProofChoice: "FOUNDATION",
		Stage: "OBSERVE", Step: "historical-actions-api-snapshot",
		Reason: "HISTORICAL_FIXTURE_OPEN_NOT_CURRENT_EVIDENCE",
	}
	evidence.EvidenceDigest = DigestValue([]any{
		evidence.Classification, evidence.Status, evidence.ObservationRef,
		evidence.SkippedWorkflowRuns, evidence.QueuedWorkflowRuns,
		evidence.ObservedAt, evidence.Query, evidence.SubjectSHA,
	})
	return evidence
}

func guardianExpectation() GuardianExpectation {
	return GuardianExpectation{
		Scope: "CI_TRUST_ROOT", Code: "CI-ROOT-OF-TRUST-001",
		ExpectedDecision: "FAIL_CLOSED", ExpectedRoute: "BOOTSTRAP_EXPECTED_NEGATIVE",
		RequiredContext: false, IncludedInFixedDenominator: false,
		Producer: "CI guardian", Consumer: "ci-proof",
		MetaOperation: "classify-protected-workflow-change", ProofChoice: "REGRESSION",
		Stage: "GUARD", Step: "evaluate-ci-root-of-trust",
		Reason: "PR_CONTROLLED_WORKFLOW_POLICY_CHANGE",
	}
}
