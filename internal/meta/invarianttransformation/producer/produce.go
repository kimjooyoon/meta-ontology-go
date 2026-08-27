package producer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
)

const (
	replayUnavailableReason = "REGRESSION_REPLAY_RECIPE_UNAVAILABLE"
	replayExecutionReason   = "REGRESSION_REPLAY_EXECUTION_FAILED"
	replayMismatchReason    = "REGRESSION_REPLAY_MISMATCH"
)

// Build emits a source-bound receipt. The bounded fixture executes the
// candidate once, then executes the source-declared replay recipe a second
// time when that recipe is executable. It never writes the repository.
func Build(source []byte, headSHA, caseID string) (model.Receipt, error) {
	contract := model.CanonicalContract()
	spec, ok := caseSpec(contract, caseID)
	if !ok {
		return model.Receipt{}, fmt.Errorf("unknown invariant transformation case %q", caseID)
	}
	if !model.ValidHead(headSHA) {
		return model.Receipt{}, fmt.Errorf("invalid head sha %q", headSHA)
	}
	fixture, err := parseSourceFixture(source, spec)
	if err != nil {
		return model.Receipt{}, err
	}

	sourceDigest := model.DigestBytes(source)
	semanticBefore := model.SemanticDigest(fixture.Input)
	semanticAfter := model.SemanticDigest(fixture.CandidateResult)
	expectedSemantic := model.SemanticDigest(fixture.Expected)
	candidateDigest := model.CandidateDigest(fixture.CandidateOperation, fixture.Input, fixture.CandidateResult)
	replayOutput, replayErr := executeReplay(fixture.ReplayRecipe, fixture.Input)
	evidence := model.TransformationEvidence{
		SourceDigest: sourceDigest, InputValue: fixture.Input, CandidateOperation: fixture.CandidateOperation,
		CandidateResult: fixture.CandidateResult, ExpectedValue: fixture.Expected, Invariant: fixture.Invariant,
		CandidateDigest: candidateDigest, SemanticBeforeDigest: semanticBefore, SemanticAfterDigest: semanticAfter,
		ExpectedSemanticDigest: expectedSemantic, ReplayRecipe: fixture.ReplayRecipe,
		BaselineInputValue: fixture.Input, BaselineOperation: fixture.CandidateOperation,
		BaselineOutput: fixture.CandidateResult, BaselineDigest: candidateDigest,
		ReplayCount: 1, RegressionWitnessPresent: false,
	}
	if replayErr != nil {
		evidence.ReplayFailureStage = "REGRESSION"
		evidence.ReplayFailureStep = "execute-replay"
		evidence.ReplayFailureReason = replayFailureReason(fixture.ReplayRecipe, replayErr)
	} else {
		replayDigest := model.CandidateDigest(fixture.ReplayRecipe, fixture.Input, replayOutput)
		replaySemanticDigest := model.SemanticDigest(replayOutput)
		evidence.ReplayInputValue = fixture.Input
		evidence.ReplayOperation = fixture.ReplayRecipe
		evidence.ReplayOutput = replayOutput
		evidence.ReplayDigest = replayDigest
		evidence.ReplaySemanticDigest = replaySemanticDigest
		evidence.ReplayEvidenceDigest = model.ReplayDigest(candidateDigest, replayDigest)
		evidence.ReplayCount = 2
		evidence.RegressionWitnessPresent = candidateDigest == replayDigest && semanticAfter == replaySemanticDigest
	}

	postconditionDigest := model.PostconditionDigest(semanticBefore, semanticAfter, expectedSemantic)
	regressionDigest := ""
	if evidence.ReplayCount == 2 {
		regressionDigest = evidence.ReplayEvidenceDigest
	}
	statuses := map[string]string{
		"precondition": model.StatusDischarged, "transformation": model.StatusDischarged,
		"postcondition": model.StatusDischarged, "regression-witness": model.StatusDischarged,
	}
	reasons := map[string]string{
		"precondition": "EXACT_SOURCE_SNAPSHOT", "transformation": "TRANSFORMATION_OBSERVED",
		"postcondition": "SEMANTIC_POSTCONDITION_PRESERVED", "regression-witness": "REGRESSION_REPLAY_MATCHED",
	}
	if fixture.CandidateResult != fixture.Expected {
		statuses["postcondition"] = model.StatusRefuted
		reasons["postcondition"] = "SEMANTIC_POSTCONDITION_REFUTED"
	}
	if evidence.ReplayCount != 2 {
		statuses["regression-witness"] = model.StatusOpen
		reasons["regression-witness"] = evidence.ReplayFailureReason
	} else if !evidence.RegressionWitnessPresent {
		statuses["regression-witness"] = model.StatusRefuted
		reasons["regression-witness"] = replayMismatchReason
	} else if fixture.CandidateResult != fixture.Expected {
		statuses["regression-witness"] = model.StatusRefuted
		reasons["regression-witness"] = "REGRESSION_REPLAY_REFUTED"
	}

	claims := make([]model.Claim, 0, len(contract.Values))
	values := make([]model.MetaValue, 0, len(contract.Values))
	for _, valueSpec := range contract.Values {
		evidenceDigest := evidenceFor(valueSpec.ID, sourceDigest, candidateDigest, postconditionDigest, regressionDigest)
		coordinate := model.Coordinate{Stage: valueSpec.Coordinate.Stage, Step: valueSpec.Coordinate.Step, Reason: reasons[valueSpec.ID]}
		claim := model.Claim{ID: valueSpec.ID, Status: statuses[valueSpec.ID], Reason: reasons[valueSpec.ID], Coordinate: coordinate,
			EvidenceDigests: evidenceDigests(evidenceDigest), Transitions: []model.Transition{{From: model.StatusOpen, To: statuses[valueSpec.ID], Coordinate: coordinate}}}
		claims = append(claims, claim)
		values = append(values, model.MetaValue{ID: valueSpec.ID, Kind: valueSpec.Kind, Value: statuses[valueSpec.ID], EvidenceDigest: evidenceDigest,
			Producer: valueSpec.Producer, Consumer: valueSpec.Consumer, MetaOperation: valueSpec.MetaOperation, ProofChoice: valueSpec.ProofChoice, Coordinate: coordinate})
	}
	decision, resolution, reason := deriveDecision(claims)
	receipt := model.Receipt{Schema: model.ReceiptSchema, CaseID: spec.ID, CaseKind: spec.Kind, HeadSHA: headSHA, SourcePath: model.SourcePath,
		SourceDigest: sourceDigest, ContractDigest: model.Digest(contract), Producer: model.ProducerID, Consumer: model.ConsumerID,
		MetaOperation: model.AuthorityOp, ProofChoice: model.ProofRegression, Values: values, Claims: claims, Evidence: evidence,
		Decision: decision, Resolution: resolution, Reason: reason, Effects: []model.Effect{}, RepositoryWrites: 0,
		MutationAuthority: false, AuthorityScope: model.AuthorityScope}
	if fixture.ApprovedArtifact {
		effect, err := recordApprovedArtifact(fixture)
		if err != nil {
			return model.Receipt{}, err
		}
		receipt.Effects = append(receipt.Effects, effect)
	}
	return model.SealReceipt(receipt), nil
}

func caseSpec(contract model.Contract, id string) (model.CaseSpec, bool) {
	for _, spec := range contract.Cases {
		if spec.ID == id {
			return spec, true
		}
	}
	return model.CaseSpec{}, false
}

func executeReplay(recipe string, input int64) (int64, error) {
	if recipe == "unavailable" {
		return 0, fmt.Errorf("%s", replayUnavailableReason)
	}
	return executeCandidate(recipe, input)
}

func replayFailureReason(recipe string, err error) string {
	if recipe == "unavailable" {
		return replayUnavailableReason
	}
	return replayExecutionReason
}

func approvedArtifactPath() string {
	root := os.Getenv("RUNNER_TEMP")
	if root == "" {
		root = os.TempDir()
	}
	return filepath.Join(root, "gooo-invariant-transformation-approved-artifact.bin")
}

func approvedArtifactBytes(fixture sourceFixture) []byte {
	return []byte(fmt.Sprintf("gooo approved artifact\ncase=%s\ninput=%d\noperation=%s\noutput=%d\n",
		fixture.CaseID, fixture.Input, fixture.CandidateOperation, fixture.CandidateResult))
}

func recordApprovedArtifact(fixture sourceFixture) (model.Effect, error) {
	path := approvedArtifactPath()
	data := approvedArtifactBytes(fixture)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return model.Effect{}, fmt.Errorf("write approved artifact: %w", err)
	}
	return model.Effect{Kind: model.EffectApproved, ArtifactID: "gooo://invariant-transformation/artifact/approved",
		ArtifactDigest: model.DigestBytes(data), ArtifactPath: path, ArtifactSize: len(data), Producer: model.ProducerID,
		Consumer: model.ConsumerID, MetaOperation: "record-approved-artifact-effect", RepositoryWrites: 0, MutationAuthority: false}, nil
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
