package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/executor"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/judge"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/producer"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func main() {
	sourcePath := flag.String("source", "", "Gooo source value contract")
	contractPath := flag.String("contract", "", "validator expectation contract (compatibility input)")
	headSHA := flag.String("head-sha", "", "exact subject commit")
	executionID := flag.String("execution-id", "", "witness execution identifier")
	receiptDir := flag.String("receipt-dir", "", "receipt output directory")
	outputPath := flag.String("output", "", "report output path")
	flag.Parse()
	if *sourcePath == "" || *headSHA == "" || *executionID == "" || *receiptDir == "" || *outputPath == "" || *contractPath == "" {
		fail("-source, -contract, -head-sha, -execution-id, -output, and -receipt-dir are required")
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err.Error())
	}
	contractRaw, err := os.ReadFile(*contractPath)
	if err != nil {
		fail(err.Error())
	}
	var validatorContract model.Contract
	if err := json.Unmarshal(contractRaw, &validatorContract); err != nil {
		fail(fmt.Sprintf("decode validator expectation contract: %v", err))
	}
	if err := model.ValidateContract(validatorContract); err != nil {
		fail(err.Error())
	}
	report, err := buildReport(source, *headSHA, *executionID)
	if err != nil {
		fail(err.Error())
	}
	if err := writeReceipts(*receiptDir, report.Cases); err != nil {
		fail(err.Error())
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(*outputPath, append(raw, '\n'), 0o644); err != nil {
		fail(err.Error())
	}
	fmt.Printf("bounded transformation witness: %s cases=%d/%d claims=%d transitions=%d provisional=%d auth=%d executed=%d observed=%d unknown-scopes=%d\n",
		report.Decision, report.Summary.CasesSatisfied, report.Summary.CasesTotal, report.Summary.UniqueClaimInstances,
		report.Summary.AcceptedTransitions, report.Summary.ProvisionalReceipts, report.Summary.AuthorizationReceipts,
		report.Summary.ExecutedEffects, report.Summary.IndependentlyObservedEffects, report.Summary.UnknownEffectScopes)
}

func buildReport(source []byte, headSHA, executionID string) (model.Report, error) {
	if !model.ValidHead(headSHA) {
		return model.Report{}, fmt.Errorf("invalid head sha %q", headSHA)
	}
	if !model.ValidExecutionID(headSHA, executionID) {
		return model.Report{}, fmt.Errorf("invalid execution id %q", executionID)
	}
	sourceCases, err := producer.Discover(source)
	if err != nil {
		return model.Report{}, err
	}
	contract := model.CanonicalContract()
	semanticDigest, err := semanticSourceDigest(source)
	if err != nil {
		return model.Report{}, err
	}
	report := model.Report{Schema: model.ReportSchema, HeadSHA: headSHA, ExecutionID: executionID, SourcePath: model.SourcePath,
		SourceDigest: model.DigestBytes(source), SemanticSourceDigest: semanticDigest, ContractDigest: model.ValueContractDigest(),
		ValidatorContractDigest: model.ValidatorContractDigest(), DenominatorID: model.DenominatorID, DenominatorTotal: len(sourceCases),
		Decision: model.DecisionPass, Resolution: model.ResolutionExact, Reason: "ALL_BOUNDED_CASES_SATISFIED",
		NotClaimed: []string{"arbitrary transformation authorization", "general invariant preservation outside bounded-fixture-input-domain-v1", "repository promotion authority"}}
	for _, sourceCase := range sourceCases {
		expectation, ok := model.ValidatorExpectationFor(contract, sourceCase.CaseID)
		if !ok {
			return model.Report{}, fmt.Errorf("source case %q has no labeled validator expectation", sourceCase.CaseID)
		}
		provisional, err := producer.Build(source, headSHA, sourceCase.CaseID)
		if err != nil {
			return model.Report{}, err
		}
		provisional.ExecutionID = executionID
		provisional.AuthorizationDigest = model.AuthorizationDigest(provisional)
		provisional = model.SealReceipt(provisional)
		authorization := judge.Judge(provisional, source)
		receipt := provisional
		judgment := authorization
		authorizationReceiptDigest := ""
		if authorization.Independent && authorization.Decision == model.DecisionAllowed {
			authorizationReceiptDigest = authorization.AuthorizationDigest
		}
		independentlyObservedEffects := 0
		if authorization.Independent && authorization.Decision == model.DecisionAllowed && provisional.Evidence.EffectIntent == "approved-artifact" {
			root := os.Getenv("RUNNER_TEMP")
			if root == "" {
				root = os.TempDir()
			}
			effect, err := executor.Emit(provisional, authorization, headSHA, executor.Path(root, "approved-artifact-"+executionID))
			if err != nil {
				return model.Report{}, err
			}
			receipt.Phase = model.ReceiptExecuted
			receipt.Effects = []model.Effect{effect}
			receipt.TempArtifactWriteAuthorized = true
			receipt = model.SealReceipt(receipt)
			judgment = judge.Judge(receipt, source)
			if judgment.Independent {
				if _, err := executor.Observe(effect); err != nil {
					return model.Report{}, err
				}
				independentlyObservedEffects = 1
			}
		}
		satisfied := judgment.Independent && judgment.Decision == expectation.ExpectedDecision && judgment.Resolution == expectation.ExpectedResolution &&
			judgment.Reason == expectation.ExpectedReason && judgment.Status == expectation.ExpectedStatus && len(receipt.Effects) == expectation.ExpectedEffectCount
		if !satisfied {
			if report.Decision == model.DecisionPass {
				report.Decision, report.Resolution, report.Reason = model.DecisionFailClosed, model.ResolutionLower,
					fmt.Sprintf("CASE=%s;STAGE=%s;STEP=%s;REASON=%s", sourceCase.CaseID, judgmentStage(judgment), judgmentStep(judgment), judgment.Reason)
			}
		}
		report.Cases = append(report.Cases, model.CaseResult{Expectation: expectation, ProvisionalReceiptDigest: provisional.Digest, AuthorizationReceiptDigest: authorizationReceiptDigest, ExecutedEffects: len(receipt.Effects), IndependentlyObservedEffects: independentlyObservedEffects, Receipt: receipt, Judgment: judgment, Satisfied: satisfied})
	}
	report.Summary = summarize(report.Cases)
	report.Indicators = judge.Indicators(report.Summary)
	return model.SealReport(report), nil
}

func summarize(cases []model.CaseResult) model.Summary {
	summary := model.Summary{CasesTotal: len(cases), SourceDerivedCases: len(cases), BoundedInputDomainDenominator: len(cases),
		BoundedInputDomainObservations: len(cases), ClaimTemplates: len(model.CanonicalValueSpecs()),
		CorrectionCount: 12, CorrectionDenominator: 12, RepositoryNetContentObserved: false, RepositoryNetContentUnchanged: false,
		RepositoryNetStatusObserved: false, RepositoryNetStatusUnchanged: false,
		RepositoryNetContentState: model.RepositoryNetContentStateUnknown, RepositoryNetSnapshotDenominator: 1, RepositoryActualOrTransientWrites: model.UnknownEffectScope, RepositoryWrites: -1,
		RepositoryPathAuthorization: false, AmbientProcessAuthority: model.UnknownEffectScope}
	for _, item := range cases {
		if item.Satisfied {
			summary.CasesSatisfied++
		}
		switch item.Judgment.Decision {
		case model.DecisionAllowed:
			summary.AuthorizedCases++
		case model.DecisionRefuted:
			summary.RefutedCases++
		case model.DecisionBlocked:
			summary.OpenCases++
		}
		summary.ClaimsTotal += len(item.Receipt.Claims)
		summary.UniqueClaimInstances += len(item.Receipt.Claims)
		summary.DischargedClaims += item.Judgment.DischargedClaims
		summary.RefutedClaims += item.Judgment.RefutedClaims
		summary.OpenClaims += item.Judgment.OpenClaims
		for _, claim := range item.Receipt.Claims {
			summary.TransitionEvents += len(claim.Transitions)
			summary.AcceptedTransitions += len(claim.Transitions)
		}
		summary.ProvisionalReceipts += boolInt(item.ProvisionalReceiptDigest != "")
		if item.AuthorizationReceiptDigest != "" {
			summary.AuthorizationReceipts++
		}
		if item.Receipt.Phase == model.ReceiptExecuted {
			summary.ExecutedEffects += item.ExecutedEffects
			summary.TempArtifactWriteAuthorized = true
		}
		summary.IndependentlyObservedEffects += item.IndependentlyObservedEffects
		for _, effect := range item.Receipt.Effects {
			if effect.Kind == model.EffectApproved {
				summary.ApprovedArtifactEffects++
			}
			if effect.RepositoryActualOrTransientWrites == model.UnknownEffectScope {
				summary.UnknownEffectScopes++
			}
		}
		summary.RepositoryNetContentUnchanged = summary.RepositoryNetContentUnchanged && item.Receipt.RepositoryNetContentUnchanged
		summary.RepositoryMutationAuthorized |= boolInt(item.Receipt.RepositoryMutationAuthorized)
		if item.Receipt.RepositoryWritesObserved {
			if summary.RepositoryWrites < 0 {
				summary.RepositoryWrites = 0
			}
			summary.RepositoryWrites += item.Receipt.RepositoryWrites
		}
	}
	if summary.CasesTotal > 0 {
		summary.CoverageBPS = summary.CasesSatisfied * 10_000 / summary.CasesTotal
		summary.InputDomainCoverageBPS = summary.BoundedInputDomainObservations * 10_000 / summary.BoundedInputDomainDenominator
	}
	return summary
}

func writeReceipts(directory string, cases []model.CaseResult) error {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	for _, item := range cases {
		raw, err := json.MarshalIndent(item.Receipt, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, item.Receipt.CaseID+".json"), append(raw, '\n'), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func semanticSourceDigest(source []byte) (string, error) {
	file, diagnostics := syntax.ParseFile(model.SourcePath, string(source))
	if diagnostics.HasErrors() {
		return "", fmt.Errorf("semantic source syntax: %s", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return "", err
	}
	return "sha256:" + ir.StableHash(), nil
}

func judgmentStage(judgment model.Judgment) string {
	if judgment.Decision == model.DecisionBlocked {
		return "REGRESSION"
	}
	return "ADJUDICATION"
}
func judgmentStep(judgment model.Judgment) string {
	if judgment.Decision == model.DecisionBlocked {
		return "execute-replay"
	}
	return "derive-claim-state"
}
