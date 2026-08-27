package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/observereffect"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type Check struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type Judgment struct {
	Schema              string                         `json:"schema"`
	Producer            string                         `json:"producer"`
	Consumer            string                         `json:"consumer"`
	MetaOperation       string                         `json:"meta_operation"`
	ProofChoice         string                         `json:"proof_choice"`
	Decision            string                         `json:"decision"`
	SubjectDecision     string                         `json:"subject_decision"`
	Resolution          string                         `json:"resolution"`
	RepositoryWrites    int                            `json:"repository_writes"`
	MutationAuthority   bool                           `json:"mutation_authority"`
	PromotionAuthorized bool                           `json:"promotion_authorized"`
	Unknown             observereffect.Unknown         `json:"unknown"`
	Coordinate          observereffect.Unknown         `json:"coordinate"`
	Reason              string                         `json:"reason"`
	ClaimTransition     observereffect.ClaimTransition `json:"claim_transition"`
	Metrics             observereffect.Metrics         `json:"metrics"`
	Checks              []Check                        `json:"checks"`
	Digest              string                         `json:"digest"`
}

func judge(root string, report observereffect.Report, observationReceipt, effectReceipt observereffect.Receipt) (Judgment, error) {
	if report.Schema != observereffect.LedgerSchema || report.Experiment != observereffect.ExperimentName {
		return Judgment{}, fmt.Errorf("unexpected observer-effect ledger identity")
	}
	if !report.Source.GoooSource || !strings.HasSuffix(report.Source.Path, ".gooo") || report.Source.Digest == "" || len(report.Effects) != 4 || len(report.Indicators) != observereffect.FixedDenominator {
		return Judgment{}, fmt.Errorf("ledger does not contain the fixed experiment surface")
	}
	if err := validateCanonicalSource(root, report.Source); err != nil {
		return Judgment{}, err
	}
	if report.MutationAuthority || report.PromotionAuthorized || report.Authority.MutationAuthority || report.Authority.PromotionAuthorized {
		return Judgment{}, fmt.Errorf("ledger grants mutation or promotion authority")
	}
	if report.Coordinate != report.Unknown || report.Reason != report.Unknown.Reason {
		return Judgment{}, fmt.Errorf("unknown coordinate is not persistent")
	}
	if report.RepositoryWrites != report.Authority.RepositoryWrites {
		return Judgment{}, fmt.Errorf("repository write count is not authority-bound")
	}
	if err := validateTopology(root, report); err != nil {
		return Judgment{}, err
	}
	if report.Authority.OutputWrites != 0 {
		return Judgment{}, fmt.Errorf("observer output writes were claimed without instrumentation")
	}
	expectedSubject, expectedResolution := independentDecision(report)
	if report.Decision != expectedSubject || report.Resolution != expectedResolution {
		return Judgment{}, fmt.Errorf("decision is not derived from observed effects")
	}
	if err := validateEffects(report); err != nil {
		return Judgment{}, err
	}
	if err := validateCoordinates(report); err != nil {
		return Judgment{}, err
	}
	if err := validateIndicators(report); err != nil {
		return Judgment{}, err
	}
	if err := validateReceipts(report, observationReceipt, effectReceipt); err != nil {
		return Judgment{}, err
	}
	if report.Digest != independentReportDigest(report) {
		return Judgment{}, fmt.Errorf("ledger digest does not replay")
	}
	if report.EvidenceDigest != independentValueDigest([]any{report.Source, report.Observation, report.Effects, report.Unknown, report.ClaimTransition, report.Coordinates, report.Topology, report.RunnerScoped, report.Guardian}) {
		return Judgment{}, fmt.Errorf("ledger evidence digest does not replay")
	}
	if report.ClaimTransition.CurrentState != claimState(report.Decision) || !report.ClaimTransition.Persistent || report.ClaimTransition.Sequence != 2 {
		return Judgment{}, fmt.Errorf("persistent claim transition is inconsistent")
	}
	judgment := Judgment{
		Schema: observereffect.JudgmentSchema, Producer: "observer-effect-judge",
		Consumer: "ci-proof", MetaOperation: "independently-judge-effect-ledger",
		ProofChoice: "REGRESSION", Decision: "PASS", SubjectDecision: report.Decision,
		Resolution: report.Resolution, RepositoryWrites: report.RepositoryWrites,
		MutationAuthority: report.MutationAuthority, PromotionAuthorized: report.PromotionAuthorized,
		Unknown: report.Unknown, Coordinate: report.Coordinate, Reason: report.Reason,
		ClaimTransition: report.ClaimTransition, Metrics: report.Metrics,
		Checks: []Check{
			{ID: "judge.decision-recomputed", Status: "PASS", Reason: "EFFECTS_DERIVE_SUBJECT_DECISION"},
			{ID: "judge.fixed-denominator", Status: "PASS", Reason: "TWELVE_INDICATORS_RETAINED"},
			{ID: "judge.receipt-chain", Status: "PASS", Reason: "RECEIPTS_AND_LEDGER_DIGEST_BOUND"},
			{ID: "judge.authority", Status: "PASS", Reason: "MUTATION_AND_PROMOTION_DENIED"},
			{ID: "judge.coordinates", Status: "PASS", Reason: "EACH_DECLARED_COORDINATE_ADJUDICATED"},
			{ID: "judge.ci-root-of-trust", Status: "PASS", Reason: "EXPECTED_NEGATIVE_REPORTED_NOT_HIDDEN"},
		},
	}
	judgment.Digest = independentJudgmentDigest(judgment)
	return judgment, nil
}

func independentDecision(report observereffect.Report) (string, string) {
	if !report.Topology.Exact || report.RepositoryWrites != 0 || report.Observation.RepositoryStorage.Changed || coordinateStatus(report, "REPOSITORY_STORAGE") == "FAIL" || coordinateStatus(report, "ENVIRONMENT") == "FAIL" || coordinateStatus(report, "LOGICAL_TIME") == "FAIL" {
		return "FAIL_CLOSED", "EXACT"
	}
	if report.Unknown.Reason != "NONE" || coordinateStatus(report, "ENVIRONMENT") == "UNKNOWN" || coordinateStatus(report, "LOGICAL_TIME") == "UNKNOWN" || coordinateStatus(report, "OUTPUT") == "OPEN" {
		return "UNKNOWN", "LOWER_RESOLUTION"
	}
	return "OBSERVED", "EXACT"
}

func coordinateStatus(report observereffect.Report, coordinate string) string {
	for _, adjudication := range report.Coordinates {
		if adjudication.Coordinate == coordinate {
			return adjudication.Status
		}
	}
	return "UNKNOWN"
}

func validateCanonicalSource(root string, source observereffect.Source) error {
	path := source.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read canonical source: %w", err)
	}
	if independentBytesDigest(payload) != source.Digest || !strings.HasSuffix(path, ".gooo") {
		return fmt.Errorf("canonical source digest or suffix is not bound")
	}
	file, diagnostics := syntax.ParseFile(path, string(payload))
	if file == nil || diagnostics.HasErrors() || !source.CanonicalParse {
		return fmt.Errorf("canonical source parse was not observed")
	}
	ir, err := bidir.Lower(file)
	if err != nil || !source.CanonicalLowering || ir.StableHash() != source.SemanticDigest || !source.GoooSource {
		return fmt.Errorf("canonical source lowering was not observed")
	}
	expected := independentSemanticInterventions(path, payload, source.SemanticDigest)
	if !reflect.DeepEqual(expected, source.Interventions) {
		return fmt.Errorf("semantic/comment interventions are not canonical")
	}
	for _, intervention := range source.Interventions {
		if !intervention.SemanticInvariant || !intervention.ParseValid || !intervention.LoweringValid {
			return fmt.Errorf("semantic intervention %s did not preserve canonical meaning", intervention.Name)
		}
	}
	return nil
}

type independentInterventionCase struct {
	Name     string
	Mutation string
	Suffix   string
}

var independentInterventionCases = []independentInterventionCase{
	{
		Name: "comment-declaration-intervention", Mutation: "append comment-only entity and activity declarations",
		Suffix: "\n// entity Fake id \"gooo://fake/entity\"\n// activity Fake(Entity) -> Entity\n",
	},
	{
		Name: "quoted-text-comment-intervention", Mutation: "append quoted declaration-looking text inside a comment",
		Suffix: "\n// \"entity Fake\" \"activity Fake(Entity) -> Entity\"\n",
	},
}

func independentSemanticInterventions(filename string, payload []byte, baselineDigest string) []observereffect.SemanticIntervention {
	interventions := make([]observereffect.SemanticIntervention, 0, len(independentInterventionCases))
	for _, expected := range independentInterventionCases {
		mutated := append(append([]byte(nil), payload...), []byte(expected.Suffix)...)
		file, diagnostics := syntax.ParseFile(filename, string(mutated))
		parseValid := file != nil && !diagnostics.HasErrors()
		loweringValid := false
		mutatedDigest := ""
		if parseValid {
			ir, err := bidir.Lower(file)
			if err == nil {
				loweringValid = true
				mutatedDigest = ir.StableHash()
			}
		}
		interventions = append(interventions, observereffect.SemanticIntervention{
			Name: expected.Name, Mutation: expected.Mutation,
			ParseValid: parseValid, LoweringValid: loweringValid,
			BaselineDigest: baselineDigest, MutatedDigest: mutatedDigest,
			SemanticInvariant: parseValid && loweringValid && baselineDigest != "" && mutatedDigest == baselineDigest,
			Producer:          "observer-effect-ledger", Consumer: "observer-effect-judge",
			MetaOperation: "intervene-comment-and-quoted-text", ProofChoice: "REGRESSION",
			Stage: "BIND", Step: "parse-and-lower-intervention", Reason: "COMMENT_OR_QUOTED_TEXT_DID_NOT_CHANGE_SEMANTIC_IR",
		})
	}
	return interventions
}

func independentBytesDigest(payload []byte) string {
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}

type independentTopologyExpectation struct {
	Path             string
	Workflow         string
	Upstream         string
	TriggerBlock     string
	PullRequestBlock string
	Concurrency      string
}

var independentTopologyExpectations = []independentTopologyExpectation{
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

func validateTopology(root string, report observereffect.Report) error {
	topology := report.Topology
	if topology.Scope != "STATIC_TRIGGER_TOPOLOGY" || topology.WorkflowRunSubscribersAudited != 5 || topology.WorkflowRunSubscribersExpected != 5 || topology.BranchFilteredSubscribersBefore != 0 || topology.BranchFilteredSubscribersExpected != 5 || topology.DuplicatePROObservationPathsBefore != 2 || topology.DuplicatePROObservationPathsAfter != 1 || topology.ExpectedSkippedCIChildRunsPerPRCompletionBefore != 4 || topology.ExpectedSkippedCIChildRunsPerPRCompletionAfter != 0 {
		return fmt.Errorf("static topology metrics are not exact")
	}
	if len(topology.Subscribers) != len(independentTopologyExpectations) || len(topology.CausalEdges) != 4 {
		return fmt.Errorf("static topology surface is incomplete")
	}
	actualFiltered := 0
	for index, expected := range independentTopologyExpectations {
		payload, err := os.ReadFile(filepath.Join(root, expected.Path))
		actual := false
		if err == nil {
			text := string(payload)
			actual = strings.Contains(text, expected.TriggerBlock) && strings.Contains(text, "cancel-in-progress: true") && strings.Contains(text, "group: "+expected.Concurrency) && (expected.PullRequestBlock == "" || strings.Contains(text, expected.PullRequestBlock))
		}
		if actual {
			actualFiltered++
		}
		subscriber := topology.Subscribers[index]
		if subscriber.Workflow != expected.Workflow || subscriber.Upstream != expected.Upstream || subscriber.Status != map[bool]string{true: "PASS", false: "FAIL"}[actual] || subscriber.Expected == "" || subscriber.Actual == "" || subscriber.Concurrency != expected.Concurrency || subscriber.Producer == "" || subscriber.Consumer == "" || subscriber.MetaOperation == "" || subscriber.ProofChoice == "" || subscriber.Stage == "" || subscriber.Step == "" || subscriber.Reason == "" {
			return fmt.Errorf("workflow trigger topology subscriber %s is not independently bound", expected.Workflow)
		}
	}
	if topology.BranchFilteredSubscribers != actualFiltered || topology.Exact != (actualFiltered == len(independentTopologyExpectations)) {
		return fmt.Errorf("workflow trigger topology exactness is inconsistent")
	}
	edges := map[string][2]int{
		"ci-completion-to-skipped-child":        {4, 0},
		"language-pr-to-observation":            {2, 1},
		"same-pr-stale-commit-to-cancellation":  {0, 1},
		"same-branch-stale-run-to-cancellation": {0, 1},
	}
	for _, edge := range topology.CausalEdges {
		want, ok := edges[edge.ID]
		if !ok || edge.Before != want[0] || edge.After != want[1] || edge.Producer == "" || edge.Consumer == "" || edge.MetaOperation == "" || edge.ProofChoice == "" || edge.Stage == "" || edge.Step == "" || edge.Reason == "" {
			return fmt.Errorf("causal topology edge %s is not bound", edge.ID)
		}
		delete(edges, edge.ID)
	}
	if len(edges) != 0 {
		return fmt.Errorf("causal topology edges are incomplete")
	}
	runner := report.RunnerScoped
	expectedRunnerDigest := independentValueDigest([]any{
		runner.Classification, runner.Status, runner.ObservationRef,
		runner.SkippedWorkflowRuns, runner.QueuedWorkflowRuns,
		runner.ObservedAt, runner.Query, runner.SubjectSHA,
	})
	if runner.Scope != "RUNNER_SCOPED" || runner.Classification != "HISTORICAL_FIXTURE" || runner.Status != "OPEN" || runner.Source != "review-supplied historical Actions API snapshot" || runner.ObservationRef != "dev SHA #540 latest 100 workflow_run objects" || runner.ObservedAt != "UNKNOWN" || runner.Query != "NOT_CAPTURED" || runner.SubjectSHA != "dev SHA #540" || runner.EvidenceDigest != expectedRunnerDigest || runner.SkippedWorkflowRuns != 59 || runner.QueuedWorkflowRuns != 41 || !runner.TimeDependent || runner.CurrentEvidence || runner.IncludedInFixedDenominator || runner.Producer == "" || runner.Consumer == "" || runner.MetaOperation == "" || runner.ProofChoice == "" || runner.Stage == "" || runner.Step == "" || runner.Reason == "" {
		return fmt.Errorf("runner-scoped queue evidence is not isolated")
	}
	guardian := report.Guardian
	if guardian.Scope != "CI_TRUST_ROOT" || guardian.Code != "CI-ROOT-OF-TRUST-001" || guardian.ExpectedDecision != "FAIL_CLOSED" || guardian.ExpectedRoute != "BOOTSTRAP_EXPECTED_NEGATIVE" || guardian.RequiredContext || guardian.IncludedInFixedDenominator || guardian.Producer == "" || guardian.Consumer == "" || guardian.MetaOperation == "" || guardian.ProofChoice == "" || guardian.Stage == "" || guardian.Step == "" || guardian.Reason == "" {
		return fmt.Errorf("CI root-of-trust expected-negative evidence is hidden or malformed")
	}
	return nil
}

func validateEffects(report observereffect.Report) error {
	seen := make(map[string]bool, len(report.Effects))
	for _, effect := range report.Effects {
		if seen[effect.Domain] || effect.Producer != "observer-effect-ledger" || effect.Consumer != "observer-effect-judge" || effect.MetaOperation == "" || effect.ProofChoice == "" {
			return fmt.Errorf("effect metadata is not unique and bound")
		}
		seen[effect.Domain] = true
	}
	for _, domain := range []string{"REPOSITORY_STORAGE", "ENVIRONMENT", "LOGICAL_TIME", "OUTPUT"} {
		if !seen[domain] {
			return fmt.Errorf("effect domain %s is missing", domain)
		}
	}
	output := effectByDomain(report.Effects, "OUTPUT")
	if output.WriteCount != 0 || output.ObservedChanged || output.MutationAttempted || !output.Planned || output.BeforeDigest != "UNOBSERVED" || output.AfterDigest != "UNOBSERVED" || output.Status != "OPEN" || output.Stage != "EMIT_OUTPUT" || output.Step != "artifact-write" || output.Reason != "ACTUAL_OUTPUT_WRITES_NOT_INSTRUMENTED" {
		return fmt.Errorf("observer output effect is not honestly classified")
	}
	repository := effectByDomain(report.Effects, "REPOSITORY_STORAGE")
	if repository.Planned || repository.WriteCount != report.RepositoryWrites || repository.ObservedChanged != (report.RepositoryWrites != 0) || repository.MutationAttempted != (report.RepositoryWrites != 0) || repository.Status != report.Observation.RepositoryStorage.Status || repository.Stage != report.Observation.RepositoryStorage.Stage || repository.Step != report.Observation.RepositoryStorage.Step || repository.Reason != report.Observation.RepositoryStorage.Reason {
		return fmt.Errorf("repository storage effect is inconsistent")
	}
	for _, pair := range []struct {
		domain   string
		snapshot observereffect.SnapshotDelta
	}{
		{domain: "ENVIRONMENT", snapshot: report.Observation.Environment},
		{domain: "LOGICAL_TIME", snapshot: report.Observation.LogicalTime},
	} {
		effect := effectByDomain(report.Effects, pair.domain)
		if effect.Planned || effect.WriteCount != 0 || effect.ObservedChanged != pair.snapshot.Changed || effect.MutationAttempted || effect.Status != pair.snapshot.Status || effect.Stage != pair.snapshot.Stage || effect.Step != pair.snapshot.Step || effect.Reason != pair.snapshot.Reason {
			return fmt.Errorf("%s effect is inconsistent with its observation", pair.domain)
		}
	}
	return nil
}

func validateCoordinates(report observereffect.Report) error {
	if len(report.Coordinates) != 4 {
		return fmt.Errorf("coordinate adjudication denominator is not four")
	}
	seen := make(map[string]bool, len(report.Coordinates))
	for _, adjudication := range report.Coordinates {
		if seen[adjudication.Coordinate] || adjudication.Producer != "observer-effect-ledger" || adjudication.Consumer != "observer-effect-judge" || adjudication.MetaOperation == "" || adjudication.ProofChoice == "" || adjudication.Stage == "" || adjudication.Step == "" || adjudication.Reason == "" {
			return fmt.Errorf("coordinate adjudication metadata is not bound")
		}
		seen[adjudication.Coordinate] = true
	}
	for _, expected := range []struct {
		coordinate string
		snapshot   observereffect.SnapshotDelta
	}{
		{coordinate: "REPOSITORY_STORAGE", snapshot: report.Observation.RepositoryStorage},
		{coordinate: "ENVIRONMENT", snapshot: report.Observation.Environment},
		{coordinate: "LOGICAL_TIME", snapshot: report.Observation.LogicalTime},
	} {
		adjudication := coordinateByName(report.Coordinates, expected.coordinate)
		if adjudication.Status != expected.snapshot.Status || adjudication.Resolution != expected.snapshot.Resolution || adjudication.BeforeObserved != expected.snapshot.BeforeObserved || adjudication.AfterObserved != expected.snapshot.AfterObserved || adjudication.Stage != expected.snapshot.Stage || adjudication.Step != expected.snapshot.Step || adjudication.Reason != expected.snapshot.Reason {
			return fmt.Errorf("%s coordinate is not bound to its snapshot", expected.coordinate)
		}
		if expected.snapshot.BeforeDigest == "" || expected.snapshot.AfterDigest == "" {
			return fmt.Errorf("%s coordinate has no boundary digests", expected.coordinate)
		}
		expectedStatus := "PASS"
		if !expected.snapshot.BeforeObserved || !expected.snapshot.AfterObserved {
			expectedStatus = "UNKNOWN"
		} else if expected.snapshot.Changed {
			expectedStatus = "FAIL"
		}
		if expected.snapshot.Status != expectedStatus {
			return fmt.Errorf("%s coordinate status is constructed", expected.coordinate)
		}
		expectedResolution := "EXACT"
		if expectedStatus == "UNKNOWN" {
			expectedResolution = "LOWER_RESOLUTION"
		}
		if expected.snapshot.Resolution != expectedResolution {
			return fmt.Errorf("%s coordinate resolution is inconsistent", expected.coordinate)
		}
	}
	output := coordinateByName(report.Coordinates, "OUTPUT")
	if output.Status != "OPEN" || output.Resolution != "LOWER_RESOLUTION" || output.BeforeObserved || output.AfterObserved || output.Stage != "EMIT_OUTPUT" || output.Step != "artifact-write" || output.Reason != "ACTUAL_OUTPUT_WRITES_NOT_INSTRUMENTED" {
		return fmt.Errorf("output coordinate is not open and lower-resolution")
	}
	for _, coordinate := range []string{"REPOSITORY_STORAGE", "ENVIRONMENT", "LOGICAL_TIME", "OUTPUT"} {
		if !seen[coordinate] {
			return fmt.Errorf("coordinate %s is missing", coordinate)
		}
	}
	return nil
}

func coordinateByName(coordinates []observereffect.CoordinateAdjudication, name string) observereffect.CoordinateAdjudication {
	for _, coordinate := range coordinates {
		if coordinate.Coordinate == name {
			return coordinate
		}
	}
	return observereffect.CoordinateAdjudication{}
}

func validateIndicators(report observereffect.Report) error {
	expectedIDs := map[string]bool{
		"OEL-OBS-01": true, "OEL-OBS-02": true, "OEL-OBS-03": true, "OEL-OBS-04": true,
		"OEL-OBS-05": true, "OEL-OBS-06": true, "OEL-EFF-01": true, "OEL-EFF-02": true,
		"OEL-EFF-03": true, "OEL-EFF-04": true, "OEL-GOV-01": true, "OEL-GOV-02": true,
	}
	repository := effectByDomain(report.Effects, "REPOSITORY_STORAGE")
	environment := effectByDomain(report.Effects, "ENVIRONMENT")
	logicalTime := effectByDomain(report.Effects, "LOGICAL_TIME")
	output := effectByDomain(report.Effects, "OUTPUT")
	expectedStatus := map[string]string{
		"OEL-OBS-01": independentIndicatorStatus(report.Source.GoooSource && report.Source.CanonicalParse && report.Source.CanonicalLowering && report.Source.SemanticDigest != ""),
		"OEL-OBS-02": independentIndicatorStatus(report.Observation.RepositoryStorage.BeforeObserved),
		"OEL-OBS-03": independentIndicatorStatus(report.Observation.RepositoryStorage.AfterObserved),
		"OEL-OBS-04": independentIndicatorStatus(report.Observation.Environment.Status),
		"OEL-OBS-05": independentIndicatorStatus(report.Observation.LogicalTime.Status),
		"OEL-OBS-06": independentIndicatorStatus(len(report.Effects) == 4),
		"OEL-EFF-01": independentIndicatorStatus(repository.Status),
		"OEL-EFF-02": independentIndicatorStatus(environment.Status),
		"OEL-EFF-03": independentIndicatorStatus(logicalTime.Status),
		"OEL-EFF-04": independentIndicatorStatus(output.Status),
		"OEL-GOV-01": independentIndicatorStatus(!report.MutationAuthority && !report.PromotionAuthorized && !report.Authority.MutationAuthority && !report.Authority.PromotionAuthorized),
		"OEL-GOV-02": independentIndicatorStatus(report.ClaimTransition.Persistent && report.ClaimTransition.Sequence == 2),
	}
	ids := make(map[string]bool, len(report.Indicators))
	pass, observations, effects, guardrails := 0, 0, 0, 0
	for _, indicator := range report.Indicators {
		if ids[indicator.ID] || !expectedIDs[indicator.ID] || indicator.Producer == "" || indicator.Consumer == "" || indicator.MetaOperation == "" || indicator.ProofChoice == "" {
			return fmt.Errorf("indicator metadata is not bound")
		}
		ids[indicator.ID] = true
		if indicator.Status != expectedStatus[indicator.ID] {
			return fmt.Errorf("indicator %s is not independently adjudicated", indicator.ID)
		}
		if indicator.Status == "PASS" {
			pass++
		} else if indicator.Status != "FAIL" && indicator.Status != "UNKNOWN" {
			return fmt.Errorf("indicator %s has invalid status", indicator.ID)
		}
		switch indicator.Class {
		case "OBSERVATION":
			observations += boolInt(indicator.Status == "PASS")
		case "EFFECT":
			effects += boolInt(indicator.Status == "PASS")
		case "GUARDRAIL":
			guardrails += boolInt(indicator.Status == "PASS")
		default:
			return fmt.Errorf("indicator %s has invalid class", indicator.ID)
		}
	}
	if len(ids) != observereffect.FixedDenominator {
		return fmt.Errorf("fixed indicator set is incomplete")
	}
	if report.Metrics.FixedDenominator != observereffect.FixedDenominator || report.Metrics.Satisfied != pass || report.Metrics.CoverageBasisPoints != pass*10000/observereffect.FixedDenominator {
		return fmt.Errorf("fixed denominator metrics are not recomputed")
	}
	if report.Metrics.ObservationTotal != 6 || report.Metrics.EffectTotal != 4 || report.Metrics.GuardrailTotal != 2 {
		return fmt.Errorf("indicator denominators changed")
	}
	if report.Metrics.ObservationSatisfied != observations || report.Metrics.EffectSatisfied != effects || report.Metrics.GuardrailSatisfied != guardrails {
		return fmt.Errorf("indicator class metrics are not recomputed")
	}
	if report.Decision == "OBSERVED" && pass != observereffect.FixedDenominator {
		return fmt.Errorf("observed result did not satisfy all indicators")
	}
	return nil
}

func independentIndicatorStatus(value any) string {
	switch typed := value.(type) {
	case bool:
		if typed {
			return "PASS"
		}
		return "FAIL"
	case string:
		if typed == "PASS" {
			return "PASS"
		}
		if typed == "FAIL" {
			return "FAIL"
		}
	}
	return "UNKNOWN"
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func validateReceipts(report observereffect.Report, observationReceipt, effectReceipt observereffect.Receipt) error {
	for _, receipt := range []observereffect.Receipt{observationReceipt, effectReceipt} {
		if receipt.Schema != observereffect.ReceiptSchema || receipt.Producer != "observer-effect-ledger" || receipt.Consumer != "observer-effect-judge" || receipt.SubjectDigest != report.Source.Digest || receipt.Decision != report.Decision || receipt.Resolution != report.Resolution || receipt.RepositoryWrites != report.RepositoryWrites || receipt.MutationAuthority || receipt.Coordinate != report.Coordinate || receipt.Reason != report.Reason {
			return fmt.Errorf("receipt is not bound to the report")
		}
		if receipt.Digest != independentReceiptDigest(receipt) {
			return fmt.Errorf("receipt digest does not replay")
		}
	}
	if len(report.ReceiptDigests) != 2 || report.ReceiptDigests[0] != observationReceipt.Digest || report.ReceiptDigests[1] != effectReceipt.Digest {
		return fmt.Errorf("receipt digest list is not canonical")
	}
	if observationReceipt.Kind != "OBSERVATION_RESULT" || effectReceipt.Kind != "OBSERVER_EFFECT" {
		return fmt.Errorf("observation and effect receipts are not separated")
	}
	if observationReceipt.EvidenceDigest != report.EvidenceDigest || effectReceipt.EvidenceDigest != independentValueDigest(report.Effects) {
		return fmt.Errorf("receipt evidence is not bound to its role")
	}
	if observationReceipt.ClaimTransition != report.ClaimTransition || effectReceipt.ClaimTransition != report.ClaimTransition {
		return fmt.Errorf("receipt claim transition is not persistent")
	}
	return nil
}

func effectByDomain(effects []observereffect.Effect, domain string) observereffect.Effect {
	for _, effect := range effects {
		if effect.Domain == domain {
			return effect
		}
	}
	return observereffect.Effect{}
}

func claimState(decision string) string {
	switch decision {
	case "FAIL_CLOSED":
		return "REFUTED"
	case "UNKNOWN":
		return "UNKNOWN"
	default:
		return "SUPPORTED"
	}
}

func independentReceiptDigest(receipt observereffect.Receipt) string {
	receipt.Digest = ""
	return independentValueDigest(receipt)
}

func independentReportDigest(report observereffect.Report) string {
	report.Digest = ""
	return independentValueDigest(report)
}

func independentJudgmentDigest(judgment Judgment) string {
	judgment.Digest = ""
	return independentValueDigest(judgment)
}

func independentValueDigest(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}

func bytesReader(payload []byte) *bytes.Reader { return bytes.NewReader(payload) }
