package counterexamplefirstjudge

import (
	"fmt"
	"strings"

	cf "github.com/kimjooyoon/meta-ontology-go/internal/meta/counterexamplefirst"
)

func Evaluate(input cf.JudgeInput) cf.Report {
	if !cf.ValidContract(input.Contract) || input.SourcePath != input.Contract.SourcePath ||
		input.Corpus.Schema != cf.CorpusSchema || input.Corpus.Version != 1 ||
		len(input.Corpus.Scenarios) != cf.CaseCount {
		return closedReport(input, "COUNTEREXAMPLE_INPUT_CONTRACT_UNKNOWN", "LOWER_RESOLUTION")
	}
	if err := independentSourceCheck(input.Source, input.Contract); err != nil {
		return closedReport(input, err.Error(), "LOWER_RESOLUTION")
	}
	if len(input.Receipts) != cf.CaseCount {
		return closedReport(input, "COUNTEREXAMPLE_RECEIPT_COUNT_MISMATCH", "LOWER_RESOLUTION")
	}
	byID := make(map[string]cf.Scenario, len(input.Corpus.Scenarios))
	for _, scenario := range input.Corpus.Scenarios {
		byID[scenario.ID] = scenario
	}
	verified := 0
	for _, spec := range input.Contract.Cases {
		scenario, ok := byID[spec.ID]
		if !ok {
			return closedReport(input, "COUNTEREXAMPLE_SCENARIO_MISSING", "LOWER_RESOLUTION")
		}
		var receipt *cf.DecisionReceipt
		for index := range input.Receipts {
			if input.Receipts[index].ScenarioID == spec.ID {
				receipt = &input.Receipts[index]
				break
			}
		}
		if receipt == nil {
			return closedReport(input, "COUNTEREXAMPLE_RECEIPT_MISSING", "LOWER_RESOLUTION")
		}
		expected := independentlyExpectedReceipt(input.Contract, input.HeadSHA, input.SourcePath, input.Source, spec, scenario)
		if !sameReceipt(*receipt, expected) {
			return closedReportWithReceipts(input, input.Receipts, "COUNTEREXAMPLE_RECEIPT_MISMATCH", "EXACT", verified)
		}
		verified++
	}
	summary := summarize(input, verified)
	indicators := makeIndicators(summary, input.Contract)
	decision, resolution, reason := "FAIL_CLOSED", "EXACT", "COUNTEREXAMPLE_JUDGE_CONTRACT_MISMATCH"
	if allSatisfied(indicators) && summary.CasesSatisfied == summary.CasesTotal {
		decision, reason = "PASS", "COUNTEREXAMPLE_FIRST_CONTRACT_OBSERVED"
	}
	report := cf.Report{Schema: cf.ReportSchema, ContractID: input.Contract.ID, HeadSHA: input.HeadSHA,
		Decision: decision, Resolution: resolution, Reason: reason, Denominator: input.Contract.Fixed,
		Summary: summary, Indicators: indicators, Receipts: input.Receipts,
		NotClaimed: append([]string{}, input.Contract.NotClaimed...), TamperedRejected: 0}
	report.Digest = cf.ReportDigest(report)
	return report
}

func independentlyExpectedReceipt(contract cf.Contract, head, sourcePath string, source []byte, spec cf.CaseSpec, scenario cf.Scenario) cf.DecisionReceipt {
	decision, resolution, reason := "FAIL_CLOSED", "LOWER_RESOLUTION", "COUNTEREXAMPLE_REQUIRED"
	coordinate := cf.Coordinate{Stage: "COUNTEREXAMPLE", Step: "minimum-required", Reason: reason}
	if scenario.Counterexample != nil {
		if isUnknown(*scenario.Counterexample) {
			decision, reason = "UNKNOWN", "COUNTEREXAMPLE_COORDINATE_UNKNOWN"
			coordinate = cf.Coordinate{Stage: "UNKNOWN", Step: "UNKNOWN", Reason: "UNKNOWN"}
		} else if !independentMinimal(*scenario.Counterexample) {
			resolution, reason = "EXACT", "COUNTEREXAMPLE_NOT_MINIMAL"
			coordinate = cf.Coordinate{Stage: "COUNTEREXAMPLE", Step: "shrink", Reason: reason}
		} else if !independentResolution(*scenario.Counterexample, scenario.Resolution) {
			reason = "COUNTEREXAMPLE_UNRESOLVED"
			coordinate = cf.Coordinate{Stage: "RESOLUTION", Step: "await-proof", Reason: reason}
		} else {
			decision, resolution, reason = "PASS", "EXACT", "COUNTEREXAMPLE_RESOLVED"
			coordinate = cf.Coordinate{Stage: "COMPILE_DECISION", Step: "promote-after-resolution", Reason: reason}
		}
	}
	decisionInput := cf.DecisionInput{CandidateID: scenario.Candidate.ID,
		SuccessExampleDigest: cf.DigestBytes([]byte(scenario.Candidate.SuccessExample)), RequiredBeforeCompile: true}
	if scenario.Counterexample != nil {
		decisionInput.CounterexampleID = scenario.Counterexample.ID
		decisionInput.CounterexampleDigest = cf.CounterexampleDigest(*scenario.Counterexample)
	}
	if scenario.Resolution != nil {
		decisionInput.ResolutionID = scenario.Resolution.ID
		decisionInput.ResolutionDigest = cf.ResolutionDigest(*scenario.Resolution)
	}
	receipt := cf.DecisionReceipt{Schema: cf.ReceiptSchema, ContractID: contract.ID, HeadSHA: head,
		SourcePath: sourcePath, SourceDigest: cf.DigestBytes(source), ScenarioID: scenario.ID,
		Producer: contract.Producer, Consumer: contract.Consumer, MetaOperation: spec.MetaOperation,
		ProofChoice: spec.ProofChoice, Decision: decision, Resolution: resolution, Reason: reason,
		Coordinate: coordinate, DecisionInput: decisionInput,
		ClaimTransitions: independentTransitions(contract, spec, scenario, coordinate, reason),
		Effects:          cf.Effects{RepositoryWrites: 0, MutationAuthority: false}}
	receipt.Digest = cf.ReceiptDigest(receipt)
	return receipt
}

func independentMinimal(value cf.Counterexample) bool {
	if value.ID == "" || !value.Failing || !value.Minimal || value.Size < 1 || len(value.ShrinkTrace) == 0 {
		return false
	}
	last := value.Size + 1
	for _, step := range value.ShrinkTrace {
		if step.FromSize != last || step.ToSize < 1 || step.ToSize >= step.FromSize || !step.PreservesFailure {
			return false
		}
		last = step.ToSize
	}
	return last == value.Size
}

func independentResolution(counterexample cf.Counterexample, evidence *cf.ResolutionEvidence) bool {
	if evidence == nil {
		return false
	}
	return evidence.ID != "" && evidence.CounterexampleID == counterexample.ID && evidence.Accepted &&
		evidence.Stage == "RESOLUTION" && evidence.Step == "prove-repair" &&
		evidence.Reason == "RESOLUTION_EVIDENCE_ACCEPTED" &&
		evidence.ProofChoice == "COUNTEREXAMPLE_RESOLUTION" &&
		evidence.MetaOperation == "resolve-minimal-counterexample" &&
		evidence.Producer == "counterexample-resolution-witness" &&
		evidence.Consumer == cf.ProducerID
}

func isUnknown(value cf.Counterexample) bool {
	return value.Stage == "UNKNOWN" && value.Step == "UNKNOWN" && value.Reason == "UNKNOWN"
}

func independentTransitions(contract cf.Contract, spec cf.CaseSpec, scenario cf.Scenario, coordinate cf.Coordinate, reason string) []cf.ClaimTransition {
	firstStatus, secondStatus, thirdStatus := "PASS", "PASS", "PASS"
	if scenario.Counterexample == nil {
		firstStatus, secondStatus, thirdStatus = "BLOCK", "UNKNOWN", "BLOCK"
	} else if isUnknown(*scenario.Counterexample) {
		firstStatus, secondStatus, thirdStatus = "UNKNOWN", "UNKNOWN", "UNKNOWN"
	} else if !independentMinimal(*scenario.Counterexample) {
		firstStatus, secondStatus, thirdStatus = "FAIL", "BLOCK", "BLOCK"
	} else if !independentResolution(*scenario.Counterexample, scenario.Resolution) {
		secondStatus, thirdStatus = "BLOCK", "BLOCK"
	}
	firstStage, firstStep, firstReason := "COUNTEREXAMPLE", "minimum-required", "COUNTEREXAMPLE_REQUIRED"
	if scenario.Counterexample != nil {
		firstStage, firstStep, firstReason = "COUNTEREXAMPLE", "shrink", scenario.Counterexample.Reason
		if isUnknown(*scenario.Counterexample) {
			firstStage, firstStep, firstReason = "UNKNOWN", "UNKNOWN", "UNKNOWN"
		}
	}
	secondStage, secondStep, secondReason := "RESOLUTION", "await-proof", "COUNTEREXAMPLE_UNRESOLVED"
	if scenario.Resolution != nil {
		secondStage, secondStep, secondReason = scenario.Resolution.Stage, scenario.Resolution.Step, scenario.Resolution.Reason
	}
	makeTransition := func(sequence int, from, to, status, stage, step, transitionReason string) cf.ClaimTransition {
		return cf.ClaimTransition{Sequence: sequence, From: from, To: to, Status: status,
			Stage: stage, Step: step, Reason: transitionReason, Producer: contract.Producer,
			Consumer: contract.Consumer, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice,
			EvidenceDigest: cf.DigestValue([]string{scenario.ID, from, to, status, stage, step, transitionReason})}
	}
	return []cf.ClaimTransition{
		makeTransition(1, "CANDIDATE", "COUNTEREXAMPLE", firstStatus, firstStage, firstStep, firstReason),
		makeTransition(2, "COUNTEREXAMPLE", "RESOLUTION", secondStatus, secondStage, secondStep, secondReason),
		makeTransition(3, "RESOLUTION", "COMPILE_DECISION", thirdStatus, coordinate.Stage, coordinate.Step, coordinate.Reason),
	}
}

func sameReceipt(actual, expected cf.DecisionReceipt) bool {
	return actual.Digest == expected.Digest && actual.Schema == expected.Schema &&
		actual.ContractID == expected.ContractID && actual.HeadSHA == expected.HeadSHA &&
		actual.SourcePath == expected.SourcePath && actual.SourceDigest == expected.SourceDigest &&
		actual.ScenarioID == expected.ScenarioID && actual.Producer == expected.Producer &&
		actual.Consumer == expected.Consumer && actual.MetaOperation == expected.MetaOperation &&
		actual.ProofChoice == expected.ProofChoice && actual.Decision == expected.Decision &&
		actual.Resolution == expected.Resolution && actual.Reason == expected.Reason &&
		actual.Coordinate == expected.Coordinate &&
		cf.DigestValue(actual.DecisionInput) == cf.DigestValue(expected.DecisionInput) &&
		cf.DigestValue(actual.ClaimTransitions) == cf.DigestValue(expected.ClaimTransitions) &&
		actual.Effects == expected.Effects && actual.Digest == cf.ReceiptDigest(actual)
}

func summarize(input cf.JudgeInput, verified int) cf.Summary {
	byID := make(map[string]cf.Scenario, len(input.Corpus.Scenarios))
	for _, scenario := range input.Corpus.Scenarios {
		byID[scenario.ID] = scenario
	}
	summary := cf.Summary{CasesTotal: len(input.Contract.Cases), CounterexamplesRequired: len(input.Contract.Cases),
		ReceiptsVerified: verified, ProducerDependencies: input.ProducerDependencies}
	for _, spec := range input.Contract.Cases {
		scenario := byID[spec.ID]
		if scenario.Counterexample != nil {
			summary.CounterexamplesObserved++
			if independentMinimal(*scenario.Counterexample) {
				summary.MinimalCounterexamples++
			}
			if independentResolution(*scenario.Counterexample, scenario.Resolution) {
				summary.ResolutionsObserved++
				summary.PromotionsAfterResolution++
			}
			if isUnknown(*scenario.Counterexample) {
				summary.UnknownCoordinatesPreserved++
			}
		} else {
			summary.SuccessOnlyBlocks++
		}
	}
	summary.ClaimTransitionsPreserved = verified * 3
	summary.DeterministicReplays = 1
	summary.CasesSatisfied = verified
	return summary
}

func makeIndicators(summary cf.Summary, contract cf.Contract) []cf.Indicator {
	metric := func(id, class, proof, operation string, value, target, denominator int) cf.Indicator {
		return cf.Indicator{ID: id, Class: class, Producer: contract.Producer, Consumer: contract.Consumer,
			ProofChoice: proof, MetaOperation: operation, Value: value, Target: target,
			Denominator: denominator, Satisfied: value == target}
	}
	return []cf.Indicator{
		metric("counterexample.required", "FOUNDATION", "COUNTEREXAMPLE_REQUIRED", "require-counterexample-before-compile", summary.CounterexamplesRequired, cf.CaseCount, cf.CaseCount),
		metric("counterexample.minimal", "DRIVER", "COUNTEREXAMPLE_SHRINKING", "verify-local-minimality", summary.MinimalCounterexamples, 1, 1),
		metric("resolution.bound", "COHERENCE", "COUNTEREXAMPLE_RESOLUTION", "bind-resolution-to-counterexample", summary.ResolutionsObserved, 1, 1),
		metric("decision.after-resolution", "OUTCOME", "COUNTEREXAMPLE_RESOLUTION", "promote-after-resolution", summary.PromotionsAfterResolution, 1, 1),
		metric("success-only.blocked", "GUARDRAIL", "COUNTEREXAMPLE_REQUIRED", "block-success-example-only", summary.SuccessOnlyBlocks, 1, 1),
		metric("unknown.coordinate-preserved", "GUARDRAIL", "UNKNOWN_PRESERVATION", "preserve-unknown-coordinate", summary.UnknownCoordinatesPreserved, 1, 1),
		metric("claim.transition-closure", "COHERENCE", "COUNTEREXAMPLE_RESOLUTION", "preserve-claim-transitions", summary.ClaimTransitionsPreserved, cf.TransitionCount, cf.TransitionCount),
		metric("receipt.independent-verification", "OUTCOME", "INDEPENDENT_JUDGMENT", "verify-decision-receipts", summary.ReceiptsVerified, cf.CaseCount, cf.CaseCount),
		metric("producer.dependencies", "GUARDRAIL", "INDEPENDENT_JUDGMENT", "separate-producer-from-consumer", summary.ProducerDependencies, 0, 1),
		metric("effects.repository-writes", "GUARDRAIL", "READ_ONLY", "deny-repository-writes", summary.RepositoryWrites, 0, 1),
	}
}

func allSatisfied(values []cf.Indicator) bool {
	if len(values) != cf.IndicatorCount {
		return false
	}
	for _, value := range values {
		if !value.Satisfied {
			return false
		}
	}
	return true
}

func closedReport(input cf.JudgeInput, reason, resolution string) cf.Report {
	return closedReportWithReceipts(input, nil, reason, resolution, 0)
}

func closedReportWithReceipts(input cf.JudgeInput, receipts []cf.DecisionReceipt, reason, resolution string, verified int) cf.Report {
	denominator := input.Contract.Fixed
	if denominator.Version == "" {
		denominator = cf.CanonicalContract().Fixed
	}
	report := cf.Report{Schema: cf.ReportSchema, ContractID: input.Contract.ID, HeadSHA: input.HeadSHA,
		Decision: "FAIL_CLOSED", Resolution: resolution, Reason: reason, Denominator: denominator,
		Summary: cf.Summary{CasesTotal: cf.CaseCount, ReceiptsVerified: verified,
			ProducerDependencies: input.ProducerDependencies}, Receipts: receipts,
		NotClaimed: append([]string{}, cf.CanonicalContract().NotClaimed...), Digest: ""}
	report.Digest = cf.ReportDigest(report)
	return report
}

func independentSourceCheck(source []byte, contract cf.Contract) error {
	lines := strings.Split(strings.ReplaceAll(string(source), "\r\n", "\n"), "\n")
	required := []string{
		"package " + contract.Package,
		"namespace " + contract.Namespace,
		"entity CompilationClaim id \"gooo://counterexample-first/entity/compilation-claim\"",
		"entity MinimalCounterexample id \"gooo://counterexample-first/entity/minimal-counterexample\"",
		"entity ResolutionEvidence id \"gooo://counterexample-first/entity/resolution-evidence\"",
		"entity CompilationDecision id \"gooo://counterexample-first/entity/compilation-decision\"",
		"activity DiscoverMinimalCounterexample(CompilationClaim) -> MinimalCounterexample",
		"activity BindResolutionEvidence(MinimalCounterexample) -> ResolutionEvidence",
		"activity PromoteOnlyAfterResolution(ResolutionEvidence) -> CompilationDecision",
	}
	for _, want := range required {
		found := false
		for _, line := range lines {
			if strings.TrimSpace(line) == want {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("COUNTEREXAMPLE_SOURCE_UNKNOWN:%s", want)
		}
	}
	return nil
}
