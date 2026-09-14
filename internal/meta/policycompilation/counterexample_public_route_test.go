package policycompilation

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestCounterexamplePublicGenerationCheckAndExecution(t *testing.T) {
	source, input := counterexampleExecutionFixture(t)
	work := t.TempDir()
	project := filepath.Join(work, "project")
	if err := os.Mkdir(project, 0o700); err != nil {
		t.Fatal(err)
	}
	policyPath := writeRevisionOperationInput(t, project, "policy.gooo", source)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	raw := counterexampleExecutionJSON(t, input)
	observed, err := ObserveGoooPolicyCounterexample(ctx, PolicyRevisionOperationContract(),
		policyPath, source, "metapolicycompilation", "metapolicycompilation", raw)
	if err != nil {
		t.Fatal(err)
	}
	assertCounterexampleExecutionWitness(t, observed, input, raw)
	if !sameResult(input.Counterexample.Observed, observed.Operation.Observation.Baseline.FirstResults[0]) {
		t.Fatal("original counterexample was not corroborated by native execution")
	}
	request, err := DecodePolicyRevisionObservationRequest([]byte(observed.RevisionRequestJSON))
	if err != nil {
		t.Fatal(err)
	}
	binary := buildCounterexamplePublicCLI(t, ctx, work)
	proposal, proposalBytes := generateCounterexamplePublicProposal(t, ctx, binary, work, project, policyPath, request)
	if proposal.Condition != request.Condition || proposal.FromDecision != request.FromDecision ||
		proposal.ToDecision != request.ToDecision || proposal.Original.SourceDigest != DigestBytes(source) ||
		proposal.Candidate.SourceDigest != observed.Proposal.CandidatePolicy.SourceDigest ||
		proposal.Candidate.SemanticDigest != observed.Proposal.CandidatePolicy.SemanticDigest ||
		!reflect.DeepEqual(proposal.ChangedCoordinates, observed.Proposal.ChangedCoordinates) {
		t.Fatal("public generation did not consume the counterexample-derived request")
	}
	candidatePath := filepath.Join(work, "first", "candidate.gooo")
	candidate := readCounterexamplePublicArtifact(t, candidatePath)
	if !bytes.Equal(candidate, []byte(observed.Proposal.CandidateSource)) {
		t.Fatal("public candidate differs from the Gooo-bound execution candidate")
	}
	runCounterexamplePublicCLI(t, ctx, binary, work, "check", candidatePath)
	manifest, judge := compileCounterexamplePublicCandidate(t, ctx, binary, work, candidatePath)
	assertCounterexamplePublicExecution(t, ctx, observed, input, proposalBytes, manifest, judge)
	if !bytes.Equal(source, readCounterexamplePublicArtifact(t, policyPath)) {
		t.Fatal("public route changed the original source")
	}
	entries, err := os.ReadDir(project)
	if err != nil || len(entries) != 1 {
		t.Fatalf("public route changed the input project inventory: %v", err)
	}
}

func compileCounterexamplePublicCandidate(t *testing.T, ctx context.Context, binary, work, candidatePath string) (PublicGenerationManifest, []byte) {
	t.Helper()
	outputRoot := filepath.Join(work, "generated")
	raw := runCounterexamplePublicCLI(t, ctx, binary, work,
		"generate", candidatePath, "--out", outputRoot, "--profile", PublicProfileID,
		"--profile-package", "metapolicycompilation", "--profile-namespace", "metapolicycompilation",
		"--profile-project-root", filepath.Dir(candidatePath), "--json")
	var manifest PublicGenerationManifest
	if err := decodeStrictJSON(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(outputRoot)
	if err != nil || len(entries) != 4 || len(manifest.GeneratedFiles) != len(entries) ||
		manifest.ExecutionObserved || manifest.CurrentConformance != PublicGenerationConformanceUnknown {
		t.Fatalf("compilation artifact or execution boundary changed: %v", err)
	}
	judge := readCounterexamplePublicArtifact(t, filepath.Join(outputRoot, "judge.go"))
	if manifest.GeneratedJudgeDigest != DigestBytes(judge) {
		t.Fatal("public manifest does not identify the emitted judge bytes")
	}
	return manifest, judge
}

func assertCounterexamplePublicExecution(t *testing.T, ctx context.Context, report PolicyCounterexampleExecution, input PolicyCounterexampleExecutionInput, proposalBytes []byte, manifest PublicGenerationManifest, judge []byte) {
	t.Helper()
	observed := report.Operation.Observation
	if manifest.SourceDigest != observed.CandidatePolicy.SourceDigest ||
		manifest.SemanticDigest != observed.CandidatePolicy.SemanticDigest ||
		manifest.GeneratedJudgeDigest != observed.Candidate.GeneratedJudgeDigest {
		t.Fatal("public compilation differs from the source/semantic/generated execution identity")
	}
	cases := make([]Case, 0, len(input.Cases))
	for _, pair := range input.Cases {
		cases = append(cases, pair.Candidate)
	}
	results, err := ExecuteGeneratedBatch(ctx, judge, cases)
	if err != nil || len(results) != len(cases) {
		t.Fatalf("execute the public-generated judge: %v", err)
	}
	for index, result := range results {
		if !sameResult(result, observed.Candidate.FirstResults[index]) {
			t.Fatalf("public-generated decision differs for case %s", result.CaseID)
		}
	}
	facts := make(map[string]any)
	facts["evidence_class"] = EvidenceSyntheticFixture
	facts["gooo_binding"] = report.Operation.Binding
	facts["input_artifact_digest"] = report.InputArtifactDigest
	facts["revision_request_artifact_digest"] = report.Operation.RequestArtifactDigest
	facts["public_proposal_digest"] = DigestBytes(proposalBytes)
	facts["candidate_source_digest"] = manifest.SourceDigest
	facts["candidate_semantic_digest"] = manifest.SemanticDigest
	facts["public_generated_judge_digest"] = DigestBytes(judge)
	facts["public_generated_artifacts"] = manifest.GeneratedFiles
	facts["execution_counts"] = observed.Counts
	facts["public_fresh_results"] = results
	facts["admission"] = observed.Admission
	facts["improvement"] = observed.Improvement
	raw, err := json.Marshal(facts)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("COUNTEREXAMPLE_PUBLIC_ROUTE_WITNESS=%s", raw)
}
