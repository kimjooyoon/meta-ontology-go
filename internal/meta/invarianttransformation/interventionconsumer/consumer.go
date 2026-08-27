package interventionconsumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	reportSchema                   = "gooo/invariant-transformation-intervention-report/v1"
	consumerSchema                 = "gooo/invariant-transformation-intervention-consumer/v1"
	denominatorID                  = "gooo/invariant-transformation-intervention-denominator/v1"
	semanticDenominatorID          = "gooo/invariant-transformation-intervention-semantic-denominator/v1"
	nonSemanticDenominatorID       = "gooo/invariant-transformation-intervention-nonsemantic-denominator/v1"
	semanticCaseID                 = "semantic-source-intervention"
	nonSemanticCaseID              = "nonsemantic-source-intervention"
	semanticClaimID                = "semantic-intervention-claim"
	nonSemanticClaimID             = "nonsemantic-intervention-claim"
	interventionStage              = "INTERVENTION"
	semanticStep                   = "compare-semantic-projection-and-decision"
	semanticReason                 = "SEMANTIC_PROJECTION_AND_DECISION_CHANGED"
	nonSemanticStep                = "compare-nonsemantic-projection-and-decision"
	nonSemanticReason              = "NONSEMANTIC_PROJECTION_AND_DECISION_PRESERVED"
	semanticContradictionReason    = "SEMANTIC_INTERVENTION_CONTRADICTED"
	nonSemanticContradictionReason = "NONSEMANTIC_INTERVENTION_CONTRADICTED"
	evidenceUnobservableReason     = "INTERVENTION_EVIDENCE_UNOBSERVABLE"
	failClosedDecision             = "FAIL_CLOSED"
	preservedCaseID                = "preserved-translation"
	expectedSemanticMutation       = "expected=4"
	originalSemanticMutation       = "expected=3"
	nonSemanticInterventionLabel   = "comment-and-whitespace-only"
)

// DependencyBoundary is the production-only import evidence emitted by CI.
// The consumer accepts the evidence as data; it does not import the producer
// package whose dependency boundary it checks.
type DependencyBoundary struct {
	ProducerDependencyImports        int `json:"producer_dependency_imports"`
	AllowedProducerDependencyImports int `json:"allowed_producer_dependency_imports"`
}

// Audit is a source-bound consumer result. The fixed intervention counts are
// intentionally separate: reconstructed_cases is not a success score.
type Audit struct {
	Schema                           string `json:"schema"`
	HeadSHA                          string `json:"head_sha"`
	ProducerDependencyImports        int    `json:"producer_dependency_imports"`
	AllowedProducerDependencyImports int    `json:"allowed_producer_dependency_imports"`
	ReconstructedCases               int    `json:"reconstructed_cases"`
	ExpectedCases                    int    `json:"expected_cases"`
	CoherentTamperRejected           int    `json:"coherent_tamper_rejected"`
	ExpectedCoherentTamperRejections int    `json:"expected_coherent_tamper_rejections"`
	Decision                         string `json:"decision"`
	Resolution                       string `json:"resolution"`
	Reason                           string `json:"reason"`
	RepositoryWrites                 int    `json:"repository_writes"`
	MutationAuthority                bool   `json:"mutation_authority"`
	Digest                           string `json:"digest"`
}

type coordinateWire struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type transitionWire struct {
	From       string         `json:"from"`
	To         string         `json:"to"`
	Coordinate coordinateWire `json:"coordinate"`
}

type claimWire struct {
	ID          string           `json:"id"`
	Status      string           `json:"status"`
	Resolution  string           `json:"resolution"`
	Reason      string           `json:"reason"`
	Coordinate  coordinateWire   `json:"coordinate"`
	Transitions []transitionWire `json:"transitions"`
}

type projectionWire struct {
	Activity            string `json:"activity"`
	CaseID              string `json:"case_id"`
	Input               int64  `json:"input"`
	CandidateOperation  string `json:"candidate_operation"`
	CandidateResult     int64  `json:"candidate_result"`
	Expected            int64  `json:"expected"`
	Invariant           string `json:"invariant"`
	RegressionAvailable bool   `json:"regression_available"`
	ApprovedArtifact    bool   `json:"approved_artifact"`
}

type judgmentWire struct {
	Decision         string `json:"decision"`
	Resolution       string `json:"resolution"`
	Reason           string `json:"reason"`
	Status           string `json:"status"`
	Independent      bool   `json:"independent"`
	CheckedClaims    int    `json:"checked_claims"`
	DischargedClaims int    `json:"discharged_claims"`
	OpenClaims       int    `json:"open_claims"`
	RefutedClaims    int    `json:"refuted_claims"`
	Effects          int    `json:"effects"`
}

type sliceDenominatorWire struct {
	ID             string `json:"id"`
	CasesTotal     int    `json:"cases_total"`
	CasesSatisfied int    `json:"cases_satisfied"`
	CoverageBPS    int    `json:"coverage_bps"`
}

type fixedDenominatorWire struct {
	ID             string               `json:"id"`
	CasesTotal     int                  `json:"cases_total"`
	SemanticChange sliceDenominatorWire `json:"semantic_change"`
	NonSemantic    sliceDenominatorWire `json:"nonsemantic_change"`
}

type failureWire struct {
	CaseID string `json:"case_id"`
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type caseWire struct {
	ID                        string           `json:"id"`
	Kind                      string           `json:"kind"`
	SourceEdit                string           `json:"source_edit"`
	BaselineProjection        projectionWire   `json:"baseline_projection"`
	MutatedProjection         projectionWire   `json:"mutated_projection"`
	BaselineProjectionDigest  string           `json:"baseline_projection_digest"`
	MutatedProjectionDigest   string           `json:"mutated_projection_digest"`
	BaselineSourceDigest      string           `json:"baseline_source_digest"`
	MutatedSourceDigest       string           `json:"mutated_source_digest"`
	BaselineReceiptDigest     string           `json:"baseline_receipt_digest"`
	MutatedReceiptDigest      string           `json:"mutated_receipt_digest"`
	BaselineReceiptDecision   string           `json:"baseline_receipt_decision"`
	MutatedReceiptDecision    string           `json:"mutated_receipt_decision"`
	BaselineJudgment          judgmentWire     `json:"baseline_judgment"`
	MutatedJudgment           judgmentWire     `json:"mutated_judgment"`
	BaselineClaimTransitions  []transitionWire `json:"baseline_claim_transitions"`
	MutatedClaimTransitions   []transitionWire `json:"mutated_claim_transitions"`
	RawSourceDigestChanged    bool             `json:"raw_source_digest_changed"`
	ReceiptChanged            bool             `json:"receipt_changed"`
	SemanticProjectionEqual   bool             `json:"semantic_projection_equal"`
	DecisionEqual             bool             `json:"decision_equal"`
	ResolutionEqual           bool             `json:"resolution_equal"`
	ReasonEqual               bool             `json:"reason_equal"`
	DecisionChanged           bool             `json:"decision_changed"`
	ClaimTransitionsEqual     bool             `json:"claim_transitions_equal"`
	EvidenceObservable        bool             `json:"evidence_observable"`
	RepositoryWritesZero      bool             `json:"repository_writes_zero"`
	BaselineRepositoryWrites  int              `json:"baseline_repository_writes"`
	MutatedRepositoryWrites   int              `json:"mutated_repository_writes"`
	BaselineMutationAuthority bool             `json:"baseline_mutation_authority"`
	MutatedMutationAuthority  bool             `json:"mutated_mutation_authority"`
	Claim                     claimWire        `json:"claim"`
	Satisfied                 bool             `json:"satisfied"`
}

type reportWire struct {
	Schema            string               `json:"schema"`
	HeadSHA           string               `json:"head_sha"`
	SourcePath        string               `json:"source_path"`
	SourceDigest      string               `json:"source_digest"`
	Denominator       fixedDenominatorWire `json:"denominator"`
	CaseCount         int                  `json:"case_count"`
	Cases             []caseWire           `json:"cases"`
	Decision          string               `json:"decision"`
	Resolution        string               `json:"resolution"`
	Reason            string               `json:"reason"`
	RepositoryWrites  int                  `json:"repository_writes"`
	MutationAuthority bool                 `json:"mutation_authority"`
	Failure           *failureWire         `json:"failure,omitempty"`
	Digest            string               `json:"digest"`
}

type sourceFixture struct {
	Activity            string
	CaseID              string
	Input               int64
	CandidateOperation  string
	CandidateResult     int64
	Expected            int64
	Invariant           string
	RegressionAvailable bool
	ApprovedArtifact    bool
}

// VerifyReport reconstructs both interventions from the original .gooo
// source. It never calls producer.Build or imports the producer intervention
// package. The full source-derived wire report comparison also rejects a
// self-consistent, resealed report whose claims and top-level fields were
// changed together.
func VerifyReport(raw, source []byte, headSHA string, dependency DependencyBoundary) (Audit, error) {
	if dependency.ProducerDependencyImports != 0 || dependency.AllowedProducerDependencyImports != 0 {
		return Audit{}, fmt.Errorf("consumer producer dependency boundary is not 0/0")
	}
	observed, err := decodeReport(raw)
	if err != nil {
		return Audit{}, err
	}
	expected, err := reconstructExpected(source, headSHA)
	if err != nil {
		return Audit{}, err
	}
	if !reflect.DeepEqual(observed, expected) {
		return Audit{}, fmt.Errorf("consumer rejected source-incoherent intervention report")
	}

	tampered := coherentTamper(observed)
	coherentTamperRejected := 0
	if !reflect.DeepEqual(tampered, expected) {
		coherentTamperRejected = 1
	}
	if coherentTamperRejected != 1 {
		return Audit{}, fmt.Errorf("consumer failed coherent resealed tamper regression")
	}

	audit := Audit{
		Schema:                           consumerSchema,
		HeadSHA:                          headSHA,
		ProducerDependencyImports:        dependency.ProducerDependencyImports,
		AllowedProducerDependencyImports: dependency.AllowedProducerDependencyImports,
		ReconstructedCases:               len(expected.Cases),
		ExpectedCases:                    2,
		CoherentTamperRejected:           coherentTamperRejected,
		ExpectedCoherentTamperRejections: 1,
		Decision:                         expected.Decision,
		Resolution:                       expected.Resolution,
		Reason:                           expected.Reason,
		RepositoryWrites:                 expected.RepositoryWrites,
		MutationAuthority:                expected.MutationAuthority,
	}
	audit.Digest = sealAudit(audit)
	return audit, nil
}

func decodeReport(raw []byte) (reportWire, error) {
	var report reportWire
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil {
		return reportWire{}, fmt.Errorf("decode intervention report: %w", err)
	}
	if report.Schema != reportSchema || !model.ValidHead(report.HeadSHA) || report.SourcePath != model.SourcePath ||
		!model.ValidDigest(report.SourceDigest) || report.Digest == "" || report.Digest != sealReport(report).Digest {
		return reportWire{}, fmt.Errorf("intervention report identity or digest is invalid")
	}
	return report, nil
}

func reconstructExpected(source []byte, headSHA string) (reportWire, error) {
	if !model.ValidHead(headSHA) {
		return reportWire{}, fmt.Errorf("invalid head sha %q", headSHA)
	}
	baseline, err := parseFixture(source)
	if err != nil {
		return reportWire{}, err
	}
	if baseline.CaseID != preservedCaseID || baseline.Expected != 3 {
		return reportWire{}, fmt.Errorf("source is not the fixed preserved intervention fixture")
	}
	semanticSource, err := mutateSemantic(source)
	if err != nil {
		return reportWire{}, err
	}
	nonSemanticSource, err := mutateNonSemantic(source)
	if err != nil {
		return reportWire{}, err
	}
	semantic, err := reconstructCase(source, semanticSource, headSHA, semanticCaseID, "SEMANTIC", "semantic-value-change", semanticClaimID, semanticStep)
	if err != nil {
		return reportWire{}, err
	}
	nonSemantic, err := reconstructCase(source, nonSemanticSource, headSHA, nonSemanticCaseID, "NON_SEMANTIC", nonSemanticInterventionLabel, nonSemanticClaimID, nonSemanticStep)
	if err != nil {
		return reportWire{}, err
	}
	cases := []caseWire{semantic, nonSemantic}
	decision, resolution, reason, failure := deriveReport(cases)
	repositoryWrites, mutationAuthority := effectTotals(cases)
	report := reportWire{
		Schema: reportSchema, HeadSHA: headSHA, SourcePath: model.SourcePath, SourceDigest: model.DigestBytes(source),
		Denominator: fixedDenominatorWire{
			ID: denominatorID, CasesTotal: 2,
			SemanticChange: sliceDenominatorWire{ID: semanticDenominatorID, CasesTotal: 1, CasesSatisfied: boolInt(semantic.Satisfied), CoverageBPS: boolInt(semantic.Satisfied) * 10_000},
			NonSemantic:    sliceDenominatorWire{ID: nonSemanticDenominatorID, CasesTotal: 1, CasesSatisfied: boolInt(nonSemantic.Satisfied), CoverageBPS: boolInt(nonSemantic.Satisfied) * 10_000},
		},
		CaseCount: 2, Cases: cases, Decision: decision, Resolution: resolution,
		Reason: reason, Failure: failure, RepositoryWrites: repositoryWrites, MutationAuthority: mutationAuthority,
	}
	return sealReport(report), nil
}

func deriveReport(cases []caseWire) (string, string, string, *failureWire) {
	if len(cases) == 2 && cases[0].Satisfied && cases[1].Satisfied {
		return model.DecisionPass, model.ResolutionExact, "FIXED_INTERVENTION_CONTRACT_SATISFIED", nil
	}
	for _, item := range cases {
		if item.Satisfied {
			continue
		}
		failure := &failureWire{CaseID: item.ID, Stage: item.Claim.Coordinate.Stage, Step: item.Claim.Coordinate.Step, Reason: item.Claim.Reason}
		resolution := item.Claim.Resolution
		if resolution == "" {
			resolution = model.ResolutionLower
		}
		return failClosedDecision, resolution, "CASE=" + failure.CaseID + ";STAGE=" + failure.Stage + ";STEP=" + failure.Step + ";REASON=" + failure.Reason, failure
	}
	failure := &failureWire{CaseID: "intervention-denominator", Stage: interventionStage, Step: "adjudicate-fixed-cases", Reason: evidenceUnobservableReason}
	return failClosedDecision, model.ResolutionLower, "CASE=" + failure.CaseID + ";STAGE=" + failure.Stage + ";STEP=" + failure.Step + ";REASON=" + failure.Reason, failure
}

func effectTotals(cases []caseWire) (int, bool) {
	writes := 0
	mutationAuthority := false
	for _, item := range cases {
		writes += item.BaselineRepositoryWrites + item.MutatedRepositoryWrites
		mutationAuthority = mutationAuthority || item.BaselineMutationAuthority || item.MutatedMutationAuthority
	}
	return writes, mutationAuthority
}

func reconstructCase(source, mutated []byte, headSHA, id, kind, edit, claimID, step string) (caseWire, error) {
	baselineProjection, err := parseProjection(source)
	if err != nil {
		return caseWire{}, err
	}
	mutatedProjection, err := parseProjection(mutated)
	if err != nil {
		return caseWire{}, err
	}
	baselineReceipt, err := buildReceipt(source, headSHA)
	if err != nil {
		return caseWire{}, err
	}
	mutatedReceipt, err := buildReceipt(mutated, headSHA)
	if err != nil {
		return caseWire{}, err
	}
	baselineJudgment := judgmentFromReceipt(baselineReceipt)
	mutatedJudgment := judgmentFromReceipt(mutatedReceipt)
	baselineProjectionWire := projectionFromFixture(baselineProjection)
	mutatedProjectionWire := projectionFromFixture(mutatedProjection)
	baselineTransitions := transitionsFromReceipt(baselineReceipt)
	mutatedTransitions := transitionsFromReceipt(mutatedReceipt)
	semanticEqual := reflect.DeepEqual(baselineProjectionWire, mutatedProjectionWire)
	decisionEqual := baselineJudgment.Decision == mutatedJudgment.Decision && baselineReceipt.Decision == mutatedReceipt.Decision
	resolutionEqual := baselineJudgment.Resolution == mutatedJudgment.Resolution && baselineReceipt.Resolution == mutatedReceipt.Resolution
	reasonEqual := baselineJudgment.Reason == mutatedJudgment.Reason && baselineReceipt.Reason == mutatedReceipt.Reason
	rawDigestChanged := baselineReceipt.SourceDigest != mutatedReceipt.SourceDigest
	receiptChanged := baselineReceipt.Digest != mutatedReceipt.Digest
	transitionsEqual := reflect.DeepEqual(baselineTransitions, mutatedTransitions)
	writesZero := baselineReceipt.RepositoryWrites == 0 && mutatedReceipt.RepositoryWrites == 0 &&
		!baselineReceipt.MutationAuthority && !mutatedReceipt.MutationAuthority
	observationSatisfied := (kind == "SEMANTIC" && rawDigestChanged && receiptChanged && !semanticEqual && !decisionEqual && !resolutionEqual && !reasonEqual && !transitionsEqual && writesZero &&
		baselineJudgment.Decision == model.DecisionAllowed && mutatedJudgment.Decision == model.DecisionRefuted && mutatedJudgment.Reason == "SEMANTIC_POSTCONDITION_REFUTED") ||
		(kind == "NON_SEMANTIC" && rawDigestChanged && receiptChanged && semanticEqual && decisionEqual && resolutionEqual && reasonEqual && transitionsEqual && writesZero &&
			baselineJudgment.Decision == model.DecisionAllowed && mutatedJudgment.Decision == model.DecisionAllowed)
	claimReason := semanticReason
	if kind != "SEMANTIC" {
		claimReason = nonSemanticReason
	}
	claimResolution := model.ResolutionExact
	claimStatus := model.StatusDischarged
	if !observationSatisfied {
		claimResolution = model.ResolutionInvariant
		claimStatus = model.StatusRefuted
		if kind == "SEMANTIC" {
			claimReason = semanticContradictionReason
		} else {
			claimReason = nonSemanticContradictionReason
		}
	}
	coordinate := coordinateWire{Stage: interventionStage, Step: step, Reason: claimReason}
	claim := claimWire{ID: claimID, Status: claimStatus, Resolution: claimResolution, Reason: claimReason, Coordinate: coordinate,
		Transitions: []transitionWire{{From: model.StatusOpen, To: claimStatus, Coordinate: coordinate}}}
	return caseWire{
		ID: id, Kind: kind, SourceEdit: edit, BaselineProjection: baselineProjectionWire, MutatedProjection: mutatedProjectionWire,
		BaselineProjectionDigest: model.Digest(baselineProjectionWire), MutatedProjectionDigest: model.Digest(mutatedProjectionWire),
		BaselineSourceDigest: baselineReceipt.SourceDigest, MutatedSourceDigest: mutatedReceipt.SourceDigest,
		BaselineReceiptDigest: baselineReceipt.Digest, MutatedReceiptDigest: mutatedReceipt.Digest,
		BaselineReceiptDecision: baselineReceipt.Decision, MutatedReceiptDecision: mutatedReceipt.Decision,
		BaselineJudgment: judgmentFromModel(baselineJudgment), MutatedJudgment: judgmentFromModel(mutatedJudgment),
		BaselineClaimTransitions: baselineTransitions, MutatedClaimTransitions: mutatedTransitions,
		RawSourceDigestChanged: rawDigestChanged, ReceiptChanged: receiptChanged, SemanticProjectionEqual: semanticEqual,
		DecisionEqual: decisionEqual, ResolutionEqual: resolutionEqual, ReasonEqual: reasonEqual, DecisionChanged: !decisionEqual,
		ClaimTransitionsEqual: transitionsEqual, EvidenceObservable: baselineJudgment.Independent && mutatedJudgment.Independent,
		RepositoryWritesZero: writesZero, BaselineRepositoryWrites: baselineReceipt.RepositoryWrites, MutatedRepositoryWrites: mutatedReceipt.RepositoryWrites,
		BaselineMutationAuthority: baselineReceipt.MutationAuthority, MutatedMutationAuthority: mutatedReceipt.MutationAuthority,
		Claim: claim, Satisfied: observationSatisfied,
	}, nil
}

func buildReceipt(source []byte, headSHA string) (model.Receipt, error) {
	fixture, err := parseFixture(source)
	if err != nil {
		return model.Receipt{}, err
	}
	contract := model.CanonicalContract()
	sourceDigest := model.DigestBytes(source)
	semanticBefore := model.SemanticDigest(fixture.Input)
	semanticAfter := model.SemanticDigest(fixture.CandidateResult)
	expectedSemantic := model.SemanticDigest(fixture.Expected)
	candidateDigest := model.CandidateDigest(fixture.CandidateOperation, fixture.Input, fixture.CandidateResult)
	postconditionDigest := model.PostconditionDigest(semanticBefore, semanticAfter, expectedSemantic)
	regressionDigest := ""
	if fixture.RegressionAvailable {
		regressionDigest = model.ReplayDigest(semanticBefore, semanticAfter)
	}
	evidence := model.TransformationEvidence{SourceDigest: sourceDigest, InputValue: fixture.Input, CandidateOperation: fixture.CandidateOperation,
		CandidateResult: fixture.CandidateResult, ExpectedValue: fixture.Expected, Invariant: fixture.Invariant, CandidateDigest: candidateDigest,
		SemanticBeforeDigest: semanticBefore, SemanticAfterDigest: semanticAfter, ExpectedSemanticDigest: expectedSemantic,
		RegressionWitnessPresent: fixture.RegressionAvailable, ReplayCount: 0}
	if fixture.RegressionAvailable {
		evidence.ReplayBeforeDigest = semanticBefore
		evidence.ReplayAfterDigest = semanticAfter
		evidence.ReplayCount = 1
	}
	statuses := map[string]string{"precondition": model.StatusDischarged, "transformation": model.StatusDischarged, "postcondition": model.StatusDischarged, "regression-witness": model.StatusDischarged}
	reasons := map[string]string{"precondition": "EXACT_SOURCE_SNAPSHOT", "transformation": "TRANSFORMATION_OBSERVED", "postcondition": "SEMANTIC_POSTCONDITION_PRESERVED", "regression-witness": "REGRESSION_REPLAY_MATCHED"}
	if fixture.CandidateResult != fixture.Expected {
		statuses["postcondition"] = model.StatusRefuted
		reasons["postcondition"] = "SEMANTIC_POSTCONDITION_REFUTED"
	}
	if !fixture.RegressionAvailable {
		statuses["regression-witness"] = model.StatusOpen
		reasons["regression-witness"] = "REGRESSION_WITNESS_MISSING"
	} else if fixture.CandidateResult != fixture.Expected {
		statuses["regression-witness"] = model.StatusRefuted
		reasons["regression-witness"] = "REGRESSION_REPLAY_REFUTED"
	}
	claims := make([]model.Claim, 0, len(contract.Values))
	values := make([]model.MetaValue, 0, len(contract.Values))
	for _, valueSpec := range contract.Values {
		evidenceDigest := evidenceFor(valueSpec.ID, sourceDigest, candidateDigest, postconditionDigest, regressionDigest)
		claimCoordinate := model.Coordinate{Stage: valueSpec.Coordinate.Stage, Step: valueSpec.Coordinate.Step, Reason: reasons[valueSpec.ID]}
		claim := model.Claim{ID: valueSpec.ID, Status: statuses[valueSpec.ID], Reason: reasons[valueSpec.ID], Coordinate: claimCoordinate,
			EvidenceDigests: evidenceDigests(evidenceDigest), Transitions: []model.Transition{{From: model.StatusOpen, To: statuses[valueSpec.ID], Coordinate: claimCoordinate}}}
		claims = append(claims, claim)
		values = append(values, model.MetaValue{ID: valueSpec.ID, Kind: valueSpec.Kind, Value: statuses[valueSpec.ID], EvidenceDigest: evidenceDigest,
			Producer: valueSpec.Producer, Consumer: valueSpec.Consumer, MetaOperation: valueSpec.MetaOperation, ProofChoice: valueSpec.ProofChoice, Coordinate: claimCoordinate})
	}
	decision, resolution, reason := deriveDecision(claims)
	receipt := model.Receipt{Schema: model.ReceiptSchema, CaseID: fixture.CaseID, CaseKind: "PRESERVED", HeadSHA: headSHA, SourcePath: model.SourcePath,
		SourceDigest: sourceDigest, ContractDigest: model.Digest(contract), Producer: model.ProducerID, Consumer: model.ConsumerID, MetaOperation: model.AuthorityOp,
		ProofChoice: model.ProofRegression, Values: values, Claims: claims, Evidence: evidence, Decision: decision, Resolution: resolution, Reason: reason,
		Effects: []model.Effect{}, RepositoryWrites: 0, MutationAuthority: false}
	if fixture.ApprovedArtifact {
		receipt.Effects = append(receipt.Effects, model.Effect{Kind: model.EffectApproved, ArtifactID: "gooo://invariant-transformation/artifact/approved", ArtifactDigest: candidateDigest,
			Producer: model.ProducerID, Consumer: model.ConsumerID, MetaOperation: "record-approved-artifact-effect", RepositoryWrites: 0, MutationAuthority: false})
	}
	return model.SealReceipt(receipt), nil
}

func parseFixture(source []byte) (sourceFixture, error) {
	file, diagnostics := syntax.ParseFile(model.SourcePath, string(source))
	if diagnostics.HasErrors() {
		return sourceFixture{}, fmt.Errorf("consumer source syntax: %s", diagnostics.Error())
	}
	var found *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || activity.Name != "PreservedTranslation" {
			continue
		}
		if found != nil {
			return sourceFixture{}, fmt.Errorf("duplicate source activity %q", activity.Name)
		}
		found = activity
	}
	if found == nil || len(found.Parameters) != 0 || found.Result.Name != "Transformation" || !found.ValueProgramPresent {
		return sourceFixture{}, fmt.Errorf("preserved source activity is not executable")
	}
	fields, err := decodeFields(found.ValueProgram)
	if err != nil {
		return sourceFixture{}, err
	}
	input, err := strconv.ParseInt(fields["input"], 10, 64)
	if err != nil {
		return sourceFixture{}, fmt.Errorf("source input is not int64: %w", err)
	}
	expected, err := strconv.ParseInt(fields["expected"], 10, 64)
	if err != nil {
		return sourceFixture{}, fmt.Errorf("source expected is not int64: %w", err)
	}
	candidateResult, err := evaluateAdd(fields["candidate"], input)
	if err != nil {
		return sourceFixture{}, err
	}
	if fields["case"] != preservedCaseID || fields["invariant"] != "candidate-output-equals-expected" ||
		(fields["replay"] != "present" && fields["replay"] != "missing") || (fields["effect"] != "none" && fields["effect"] != "approved-artifact") {
		return sourceFixture{}, fmt.Errorf("source fixture is outside the bounded intervention contract")
	}
	return sourceFixture{Activity: found.Name, CaseID: fields["case"], Input: input, CandidateOperation: fields["candidate"], CandidateResult: candidateResult,
		Expected: expected, Invariant: fields["invariant"], RegressionAvailable: fields["replay"] == "present", ApprovedArtifact: fields["effect"] == "approved-artifact"}, nil
}

func parseProjection(source []byte) (sourceFixture, error) { return parseFixture(source) }

func projectionFromFixture(fixture sourceFixture) projectionWire {
	return projectionWire{Activity: fixture.Activity, CaseID: fixture.CaseID, Input: fixture.Input, CandidateOperation: fixture.CandidateOperation,
		CandidateResult: fixture.CandidateResult, Expected: fixture.Expected, Invariant: fixture.Invariant, RegressionAvailable: fixture.RegressionAvailable,
		ApprovedArtifact: fixture.ApprovedArtifact}
}

func decodeFields(program string) (map[string]string, error) {
	parts := strings.Split(program, ";")
	if len(parts) != 7 {
		return nil, fmt.Errorf("fixture computes value has %d fields, want 7", len(parts))
	}
	fields := make(map[string]string, len(parts))
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || key == "" || value == "" {
			return nil, fmt.Errorf("fixture field %q is malformed", part)
		}
		if _, exists := fields[key]; exists {
			return nil, fmt.Errorf("fixture field %q is duplicated", key)
		}
		fields[key] = value
	}
	for _, key := range []string{"case", "input", "candidate", "expected", "invariant", "replay", "effect"} {
		if fields[key] == "" {
			return nil, fmt.Errorf("fixture field %q is missing", key)
		}
	}
	return fields, nil
}

func evaluateAdd(operation string, input int64) (int64, error) {
	name, operandText, ok := strings.Cut(operation, ":")
	if !ok || name != "add" || operandText == "" || strings.Contains(operandText, ":") {
		return 0, fmt.Errorf("candidate operation %q is unsupported", operation)
	}
	operand, err := strconv.ParseInt(operandText, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("candidate operand is not int64: %w", err)
	}
	const maxInt64 = int64(1<<63 - 1)
	const minInt64 = -maxInt64 - 1
	if (operand > 0 && input > maxInt64-operand) || (operand < 0 && input < minInt64-operand) {
		return 0, fmt.Errorf("candidate operation %q overflows int64", operation)
	}
	return input + operand, nil
}

func mutateSemantic(source []byte) ([]byte, error) {
	old := []byte("case=preserved-translation;input=2;candidate=add:1;" + originalSemanticMutation)
	newValue := []byte("case=preserved-translation;input=2;candidate=add:1;" + expectedSemanticMutation)
	if bytes.Count(source, old) != 1 {
		return nil, fmt.Errorf("semantic intervention target count is not 1")
	}
	return bytes.Replace(source, old, newValue, 1), nil
}

func mutateNonSemantic(source []byte) ([]byte, error) {
	return append(append([]byte{}, source...), []byte("\n\n// non-semantic intervention: comment and whitespace only\n")...), nil
}

func transitionsFromReceipt(receipt model.Receipt) []transitionWire {
	result := make([]transitionWire, 0, len(receipt.Claims))
	for _, claim := range receipt.Claims {
		for _, transition := range claim.Transitions {
			result = append(result, transitionFromModel(transition))
		}
	}
	return result
}

func judgmentFromReceipt(receipt model.Receipt) model.Judgment {
	judgment := model.Judgment{Independent: true, CheckedClaims: len(receipt.Claims), Effects: len(receipt.Effects)}
	for _, claim := range receipt.Claims {
		switch claim.Status {
		case model.StatusDischarged:
			judgment.DischargedClaims++
		case model.StatusOpen:
			judgment.OpenClaims++
		case model.StatusRefuted:
			judgment.RefutedClaims++
		}
	}
	judgment.Decision, judgment.Resolution, judgment.Reason = deriveDecision(receipt.Claims)
	switch judgment.Decision {
	case model.DecisionAllowed:
		judgment.Status = model.StatusDischarged
	case model.DecisionBlocked:
		judgment.Status = model.StatusOpen
	default:
		judgment.Status = model.StatusRefuted
	}
	return judgment
}

func deriveDecision(claims []model.Claim) (string, string, string) {
	for _, claim := range claims {
		if claim.Status == model.StatusRefuted {
			return model.DecisionRefuted, model.ResolutionInvariant, claim.Reason
		}
	}
	for _, claim := range claims {
		if claim.Status == model.StatusOpen {
			return model.DecisionBlocked, model.ResolutionLower, claim.Reason
		}
	}
	return model.DecisionAllowed, model.ResolutionExact, "ALL_INVARIANTS_DISCHARGED"
}

func evidenceFor(id, sourceDigest, candidate, postcondition, regression string) string {
	switch id {
	case "precondition":
		return sourceDigest
	case "transformation":
		return candidate
	case "postcondition":
		return postcondition
	case "regression-witness":
		return regression
	default:
		return ""
	}
}

func evidenceDigests(digest string) []string {
	if digest == "" {
		return []string{}
	}
	return []string{digest}
}

func coherentTamper(report reportWire) reportWire {
	tampered := report
	tampered.Cases = append([]caseWire(nil), report.Cases...)
	tampered.Cases[1] = report.Cases[1]
	tampered.Cases[1].RawSourceDigestChanged = false
	tampered.Cases[1].ReceiptChanged = false
	tampered.Cases[1].DecisionEqual = false
	tampered.Cases[1].ResolutionEqual = false
	tampered.Cases[1].ReasonEqual = false
	tampered.Cases[1].DecisionChanged = true
	tampered.Cases[1].ClaimTransitionsEqual = false
	tampered.Cases[1].Claim.Status = model.StatusRefuted
	tampered.Cases[1].Claim.Resolution = model.ResolutionInvariant
	tampered.Cases[1].Claim.Reason = nonSemanticContradictionReason
	tampered.Cases[1].Claim.Coordinate.Reason = nonSemanticContradictionReason
	tampered.Cases[1].Claim.Transitions = []transitionWire{{From: model.StatusOpen, To: model.StatusRefuted, Coordinate: tampered.Cases[1].Claim.Coordinate}}
	tampered.Cases[1].Satisfied = false
	tampered.Denominator.NonSemantic.CasesSatisfied = 0
	tampered.Denominator.NonSemantic.CoverageBPS = 0
	tampered.Decision = failClosedDecision
	tampered.Resolution = model.ResolutionInvariant
	tampered.Reason = "CASE=" + nonSemanticCaseID + ";STAGE=" + interventionStage + ";STEP=" + nonSemanticStep + ";REASON=" + nonSemanticContradictionReason
	tampered.Failure = &failureWire{CaseID: nonSemanticCaseID, Stage: interventionStage, Step: nonSemanticStep, Reason: nonSemanticContradictionReason}
	return sealReport(tampered)
}

func sealReport(report reportWire) reportWire {
	report.Digest = ""
	report.Digest = model.Digest(report)
	return report
}

func sealAudit(audit Audit) string {
	audit.Digest = ""
	return model.Digest(audit)
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func coordinateFromModel(coordinate model.Coordinate) coordinateWire {
	return coordinateWire{Stage: coordinate.Stage, Step: coordinate.Step, Reason: coordinate.Reason}
}

func transitionFromModel(transition model.Transition) transitionWire {
	return transitionWire{From: transition.From, To: transition.To, Coordinate: coordinateFromModel(transition.Coordinate)}
}

func judgmentFromModel(judgment model.Judgment) judgmentWire {
	return judgmentWire{Decision: judgment.Decision, Resolution: judgment.Resolution, Reason: judgment.Reason, Status: judgment.Status,
		Independent: judgment.Independent, CheckedClaims: judgment.CheckedClaims, DischargedClaims: judgment.DischargedClaims,
		OpenClaims: judgment.OpenClaims, RefutedClaims: judgment.RefutedClaims, Effects: judgment.Effects}
}
