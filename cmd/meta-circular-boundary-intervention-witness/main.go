package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundary"
	consumer "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundaryconsumer"
	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundarycontract"
)

const interventionSchema = "gooo/meta-circular-boundary-causality/v1"

func main() {
	sourcePath := flag.String("source", producer.ExpectedSourcePath, "Gooo source to intervene")
	headSHA := flag.String("head-sha", "", "exact 40-character commit SHA")
	grantPath := flag.String("grant", "", "raw external grant artifact path")
	effectPath := flag.String("effect-evidence", "", "raw workspace effect artifact path")
	output := flag.String("output", "", "causality report output path")
	flag.Parse()

	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fatal(err)
	}
	grant, err := os.ReadFile(*grantPath)
	if err != nil {
		fatal(err)
	}
	effect, err := os.ReadFile(*effectPath)
	if err != nil {
		fatal(err)
	}
	baseInput := contract.Input{Path: *sourcePath, HeadSHA: *headSHA, Source: source, GrantEvidence: grant, EffectEvidence: effect}
	baseline, err := evaluateWithExecution(baseInput, "intervention-baseline")
	if err != nil {
		fatal(fmt.Errorf("baseline: %w", err))
	}

	semanticRequest := replaceOnce(source,
		[]byte("id=explicit-read-only-capability|description=source|request=READ_ONLY|request_execution=true"),
		[]byte("id=explicit-read-only-capability|description=source|request=READ_ONLY|request_execution=false"))
	grantChange := replaceFirst(grant, []byte(`"scope": "READ_ONLY"`), []byte(`"scope": "WRITE"`))
	descriptionForgery := replaceOnce(source,
		[]byte("id=description-only|description=source|request=NONE|request_execution=true"),
		[]byte("id=description-only|description=source|request=NONE|description_authority=GRANTED|request_execution=true"))
	graphConnection := replaceOnce(source,
		[]byte("activity GrantReadOnlyMetaCapability(SelfDescription) -> ReadOnlyCapability"),
		[]byte("activity GrantReadOnlyMetaCapability(MetaOperation) -> ReadOnlyCapability"))
	commentOnly := append(append([]byte(nil), source...), '\n', '/', '/', ' ', 'c', 'o', 'm', 'm', 'e', 'n', 't', '-', 'o', 'n', 'l', 'y', ' ', 'i', 'n', 't', 'e', 'r', 'v', 'e', 'n', 't', '\n')

	cases := []contract.CausalityCase{
		intervention(baseInput, baseline, "request-only", "SEMANTIC_REQUEST", "explicit-read-only-capability", semanticRequest, grant, true, false),
		intervention(baseInput, baseline, "grant-change", "SEMANTIC_GRANT", "explicit-read-only-capability", source, grantChange, false, true),
		intervention(baseInput, baseline, "description-only-forgery", "SEMANTIC_DESCRIPTION_FORGERY", "description-only", descriptionForgery, grant, true, false),
		intervention(baseInput, baseline, "comment-only", "NON_SEMANTIC_COMMENT", "explicit-read-only-capability", commentOnly, grant, false, false),
		intervention(baseInput, baseline, "graph-connection", "SEMANTIC_GRAPH", "explicit-read-only-capability", graphConnection, grant, true, false),
	}
	report := contract.CausalityReport{Schema: interventionSchema, Cases: cases}
	report.Summary.Total = len(cases)
	for _, item := range cases {
		if item.Passed {
			report.Summary.Passed++
		}
	}
	if report.Summary.Total > 0 {
		report.Summary.CoverageBPS = report.Summary.Passed * 10_000 / report.Summary.Total
	}
	if report.Summary.Passed != report.Summary.Total {
		fatal(fmt.Errorf("semantic-causality contract failed: %d/%d", report.Summary.Passed, report.Summary.Total))
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fatal(err)
	}
	encoded = append(encoded, '\n')
	if *output == "" {
		_, _ = os.Stdout.Write(encoded)
		return
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*output, encoded, 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("semantic-causality: %d/%d (%d BPS)\n", report.Summary.Passed, report.Summary.Total, report.Summary.CoverageBPS)
}

func evaluateWithExecution(input contract.Input, label string) (contract.Report, error) {
	initial := producer.Evaluate(input)
	for _, item := range initial.Cases {
		if item.Observation.Authorization != producer.AuthorizationGranted || !item.Attempt.RequestExecution {
			continue
		}
		artifact, err := producer.ExecuteReadOnlyMetaOperation(initial.Source, item.Grant, item.Definition.ID)
		if err != nil {
			return contract.Report{}, err
		}
		root, err := os.MkdirTemp("", "meta-circular-boundary-"+label+"-")
		if err != nil {
			return contract.Report{}, err
		}
		if err := producer.WriteExecutionArtifact(root, artifact); err != nil {
			return contract.Report{}, err
		}
		input.ExecutionArtifacts = append(input.ExecutionArtifacts, artifact)
	}
	return producer.Evaluate(input), nil
}

func intervention(baseInput contract.Input, baseline contract.Report, id, kind, target string, source, grant []byte, expectedSemanticChange, expectedGrantChange bool) contract.CausalityCase {
	input := baseInput
	input.Source = source
	input.GrantEvidence = grant
	intervened, err := evaluateWithExecution(input, "intervention-"+id)
	if err != nil {
		fatal(fmt.Errorf("%s: %w", id, err))
	}
	consumerAccepted := consumer.Judge(intervened, input) == nil
	baselineCase := caseByID(baseline, target)
	intervenedCase := caseByID(intervened, target)
	baselineReceipt := baselineCase.Receipt
	intervenedReceipt := intervenedCase.Receipt
	semanticChanged := baseline.Source.SemanticDigest != intervened.Source.SemanticDigest
	sourceChanged := baseline.Source.SourceDigest != intervened.Source.SourceDigest
	rawInputChanged := sourceChanged || baseline.Source.GrantArtifactDigest != intervened.Source.GrantArtifactDigest
	grantChanged := baselineCase.Observation.GrantDigest != intervenedCase.Observation.GrantDigest
	graphChanged := baseline.Source.Graph.Digest != intervened.Source.Graph.Digest
	baselineDescriptionClaim := claimTransition(baselineReceipt, "description")
	intervenedDescriptionClaim := claimTransition(intervenedReceipt, "description")
	baselineAuthorizationClaim := claimTransition(baselineReceipt, "authorization")
	intervenedAuthorizationClaim := claimTransition(intervenedReceipt, "authorization")
	baselineExecutionClaim := claimTransition(baselineReceipt, "execution")
	intervenedExecutionClaim := claimTransition(intervenedReceipt, "execution")
	claimsChanged := baselineDescriptionClaim.After != intervenedDescriptionClaim.After || baselineDescriptionClaim.PropositionDigest != intervenedDescriptionClaim.PropositionDigest || baselineAuthorizationClaim.After != intervenedAuthorizationClaim.After || baselineAuthorizationClaim.PropositionDigest != intervenedAuthorizationClaim.PropositionDigest || baselineExecutionClaim.After != intervenedExecutionClaim.After || baselineExecutionClaim.PropositionDigest != intervenedExecutionClaim.PropositionDigest
	semanticOutputsPreserved := !semanticChanged && !grantChanged && !graphChanged && baselineCase.Observation.Decision == intervenedCase.Observation.Decision && baselineCase.Observation.Authorization == intervenedCase.Observation.Authorization && baselineCase.Observation.Execution == intervenedCase.Observation.Execution && baselineCase.Observation.GrantDigest == intervenedCase.Observation.GrantDigest && baselineCase.Observation.OutputDigest == intervenedCase.Observation.OutputDigest && baselineCase.Observation.DescriptionEscalated == intervenedCase.Observation.DescriptionEscalated && !claimsChanged
	semanticOutputsChanged := baselineCase.Observation.Decision != intervenedCase.Observation.Decision || baselineCase.Observation.Authorization != intervenedCase.Observation.Authorization || baselineCase.Observation.Execution != intervenedCase.Observation.Execution || baselineCase.Observation.GrantDigest != intervenedCase.Observation.GrantDigest || baselineCase.Observation.OutputDigest != intervenedCase.Observation.OutputDigest || baselineCase.Observation.DescriptionEscalated != intervenedCase.Observation.DescriptionEscalated || claimsChanged
	passed := consumerAccepted && rawInputChanged && semanticChanged == expectedSemanticChange && grantChanged == expectedGrantChange
	if kind == "NON_SEMANTIC_COMMENT" {
		passed = passed && semanticOutputsPreserved
	} else {
		passed = passed && semanticOutputsChanged && !semanticOutputsPreserved
	}
	return contract.CausalityCase{
		ID: id, Kind: kind, TargetCaseID: target,
		BaselineSourceDigest: baseline.Source.SourceDigest, IntervenedSourceDigest: intervened.Source.SourceDigest,
		BaselineSemanticDigest: baseline.Source.SemanticDigest, IntervenedSemanticDigest: intervened.Source.SemanticDigest,
		BaselineGrantDigest: baselineCase.Observation.GrantDigest, IntervenedGrantDigest: intervenedCase.Observation.GrantDigest,
		BaselineOutputDigest: baselineCase.Observation.OutputDigest, IntervenedOutputDigest: intervenedCase.Observation.OutputDigest,
		BaselineDescriptionPropositionDigest: baselineDescriptionClaim.PropositionDigest, IntervenedDescriptionPropositionDigest: intervenedDescriptionClaim.PropositionDigest,
		BaselineAuthorizationPropositionDigest: baselineAuthorizationClaim.PropositionDigest, IntervenedAuthorizationPropositionDigest: intervenedAuthorizationClaim.PropositionDigest,
		BaselineExecutionPropositionDigest: baselineExecutionClaim.PropositionDigest, IntervenedExecutionPropositionDigest: intervenedExecutionClaim.PropositionDigest,
		BaselineDescriptionEscalated: baselineCase.Observation.DescriptionEscalated, IntervenedDescriptionEscalated: intervenedCase.Observation.DescriptionEscalated,
		SourceChanged: sourceChanged, SemanticChanged: semanticChanged, RawInputChanged: rawInputChanged, GrantChanged: grantChanged, GraphChanged: graphChanged, ExpectedSemanticChange: expectedSemanticChange,
		ConsumerAccepted:     consumerAccepted,
		BaselineCaseDecision: baselineCase.Observation.Decision, IntervenedCaseDecision: intervenedCase.Observation.Decision,
		BaselineAuthorization: baselineCase.Observation.Authorization, IntervenedAuthorization: intervenedCase.Observation.Authorization,
		BaselineExecution: baselineCase.Observation.Execution, IntervenedExecution: intervenedCase.Observation.Execution,
		BaselineReceiptDecision: baselineReceipt.Decision, IntervenedReceiptDecision: intervenedReceipt.Decision,
		BaselineAuthorizationEvidenceDigest: baselineReceipt.AuthorizationEvidenceDigest, IntervenedAuthorizationEvidenceDigest: intervenedReceipt.AuthorizationEvidenceDigest,
		BaselineAuthorizationClaim: baselineAuthorizationClaim.After, IntervenedAuthorizationClaim: intervenedAuthorizationClaim.After,
		BaselineExecutionClaim: baselineExecutionClaim.After, IntervenedExecutionClaim: intervenedExecutionClaim.After,
		BaselineDescriptionClaim: baselineDescriptionClaim.After, IntervenedDescriptionClaim: intervenedDescriptionClaim.After,
		SemanticOutputsPreserved: semanticOutputsPreserved, Passed: passed,
	}
}

func replaceOnce(value, old, replacement []byte) []byte {
	if bytes.Count(value, old) != 1 {
		fatal(fmt.Errorf("intervention needle count is not one: %q", old))
	}
	return bytes.Replace(value, old, replacement, 1)
}

func replaceFirst(value, old, replacement []byte) []byte {
	if !bytes.Contains(value, old) {
		fatal(fmt.Errorf("intervention needle is absent: %q", old))
	}
	return bytes.Replace(value, old, replacement, 1)
}

func caseByID(report contract.Report, id string) contract.CaseResult {
	for _, item := range report.Cases {
		if item.Definition.ID == id {
			return item
		}
	}
	fatal(fmt.Errorf("missing causality case %q", id))
	return contract.CaseResult{}
}

func claimTransition(receipt contract.Receipt, suffix string) contract.ClaimTransition {
	for _, transition := range receipt.ClaimTransitions {
		if transition.ClaimID == receipt.CaseID+"."+suffix {
			return transition
		}
	}
	fatal(fmt.Errorf("missing %s claim in %q receipt", suffix, receipt.CaseID))
	return contract.ClaimTransition{}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
