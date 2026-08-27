package counterexamplefirstcompiler

import (
	"fmt"
	"strings"

	cf "github.com/kimjooyoon/meta-ontology-go/internal/meta/counterexamplefirst"
)

func Compile(contract cf.Contract, head, sourcePath string, source []byte, corpus cf.ScenarioCorpus) ([]cf.DecisionReceipt, error) {
	if !cf.ValidContract(contract) || sourcePath != contract.SourcePath ||
		corpus.Schema != cf.CorpusSchema || corpus.Version != 1 || len(corpus.Scenarios) != cf.CaseCount {
		return nil, fmt.Errorf("COUNTEREXAMPLE_INPUT_CONTRACT_UNKNOWN")
	}
	if err := validateSource(source, contract); err != nil {
		return nil, err
	}
	byID := make(map[string]cf.Scenario, len(corpus.Scenarios))
	for _, scenario := range corpus.Scenarios {
		byID[scenario.ID] = scenario
	}
	receipts := make([]cf.DecisionReceipt, 0, len(contract.Cases))
	for _, spec := range contract.Cases {
		scenario, ok := byID[spec.ID]
		if !ok || scenario.Candidate.ID == "" {
			return nil, fmt.Errorf("COUNTEREXAMPLE_SCENARIO_MISSING:%s", spec.ID)
		}
		receipts = append(receipts, compileScenario(contract, head, sourcePath, source, spec, scenario))
	}
	return receipts, nil
}

func compileScenario(contract cf.Contract, head, sourcePath string, source []byte, spec cf.CaseSpec, scenario cf.Scenario) cf.DecisionReceipt {
	decision, resolution, reason := "FAIL_CLOSED", "LOWER_RESOLUTION", "COUNTEREXAMPLE_REQUIRED"
	coordinate := cf.Coordinate{Stage: "COUNTEREXAMPLE", Step: "minimum-required", Reason: reason}
	if scenario.Counterexample != nil && unknownCoordinate(*scenario.Counterexample) {
		decision, reason = "UNKNOWN", "COUNTEREXAMPLE_COORDINATE_UNKNOWN"
		coordinate = cf.Coordinate{Stage: "UNKNOWN", Step: "UNKNOWN", Reason: "UNKNOWN"}
	} else if scenario.Counterexample != nil && !minimal(*scenario.Counterexample) {
		resolution, reason = "EXACT", "COUNTEREXAMPLE_NOT_MINIMAL"
		coordinate = cf.Coordinate{Stage: "COUNTEREXAMPLE", Step: "shrink", Reason: reason}
	} else if scenario.Counterexample != nil && !resolved(*scenario.Counterexample, scenario.Resolution) {
		reason = "COUNTEREXAMPLE_UNRESOLVED"
		coordinate = cf.Coordinate{Stage: "RESOLUTION", Step: "await-proof", Reason: reason}
	} else if scenario.Counterexample != nil {
		decision, resolution, reason = "PASS", "EXACT", "COUNTEREXAMPLE_RESOLVED"
		coordinate = cf.Coordinate{Stage: "COMPILE_DECISION", Step: "promote-after-resolution", Reason: reason}
	}
	transitions := transitions(contract, spec, scenario, coordinate, decision, reason)
	input := cf.DecisionInput{
		CandidateID:           scenario.Candidate.ID,
		SuccessExampleDigest:  cf.DigestBytes([]byte(scenario.Candidate.SuccessExample)),
		RequiredBeforeCompile: true,
	}
	if scenario.Counterexample != nil {
		input.CounterexampleID = scenario.Counterexample.ID
		input.CounterexampleDigest = cf.CounterexampleDigest(*scenario.Counterexample)
	}
	if scenario.Resolution != nil {
		input.ResolutionID = scenario.Resolution.ID
		input.ResolutionDigest = cf.ResolutionDigest(*scenario.Resolution)
	}
	receipt := cf.DecisionReceipt{
		Schema: cf.ReceiptSchema, ContractID: contract.ID, HeadSHA: head,
		SourcePath: sourcePath, SourceDigest: cf.DigestBytes(source), ScenarioID: scenario.ID,
		Producer: cf.ProducerID, Consumer: cf.ConsumerID, MetaOperation: spec.MetaOperation,
		ProofChoice: spec.ProofChoice, Decision: decision, Resolution: resolution, Reason: reason,
		Coordinate: coordinate, DecisionInput: input, ClaimTransitions: transitions,
		Effects: cf.Effects{RepositoryWrites: 0, MutationAuthority: false},
	}
	receipt.Digest = cf.ReceiptDigest(receipt)
	return receipt
}

func minimal(value cf.Counterexample) bool {
	if !value.Failing || !value.Minimal || value.ID == "" || value.Size <= 0 {
		return false
	}
	previous := value.Size + 1
	for _, step := range value.ShrinkTrace {
		if step.FromSize != previous || step.ToSize <= 0 || step.ToSize >= step.FromSize || !step.PreservesFailure {
			return false
		}
		previous = step.ToSize
	}
	return len(value.ShrinkTrace) > 0 && previous == value.Size
}

func resolved(counterexample cf.Counterexample, evidence *cf.ResolutionEvidence) bool {
	return evidence != nil && evidence.ID != "" && evidence.CounterexampleID == counterexample.ID &&
		evidence.Stage == "RESOLUTION" && evidence.Step == "prove-repair" &&
		evidence.Reason == "RESOLUTION_EVIDENCE_ACCEPTED" && evidence.Accepted &&
		evidence.ProofChoice == "COUNTEREXAMPLE_RESOLUTION" &&
		evidence.MetaOperation == "resolve-minimal-counterexample" &&
		evidence.Producer == "counterexample-resolution-witness" &&
		evidence.Consumer == cf.ProducerID
}

func unknownCoordinate(value cf.Counterexample) bool {
	return value.Stage == "UNKNOWN" && value.Step == "UNKNOWN" && value.Reason == "UNKNOWN"
}

func transitions(contract cf.Contract, spec cf.CaseSpec, scenario cf.Scenario, coordinate cf.Coordinate, decision, reason string) []cf.ClaimTransition {
	statusOne, statusTwo, statusThree := "PASS", "PASS", "PASS"
	if scenario.Counterexample == nil {
		statusOne, statusTwo, statusThree = "BLOCK", "UNKNOWN", "BLOCK"
	} else if unknownCoordinate(*scenario.Counterexample) {
		statusOne, statusTwo, statusThree = "UNKNOWN", "UNKNOWN", "UNKNOWN"
	} else if !minimal(*scenario.Counterexample) {
		statusOne, statusTwo, statusThree = "FAIL", "BLOCK", "BLOCK"
	} else if !resolved(*scenario.Counterexample, scenario.Resolution) {
		statusTwo, statusThree = "BLOCK", "BLOCK"
	}
	base := func(sequence int, from, to, status, stage, step, transitionReason string) cf.ClaimTransition {
		return cf.ClaimTransition{Sequence: sequence, From: from, To: to, Status: status,
			Stage: stage, Step: step, Reason: transitionReason, Producer: contract.Producer,
			Consumer: contract.Consumer, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice,
			EvidenceDigest: cf.DigestValue([]string{scenario.ID, from, to, status, stage, step, transitionReason})}
	}
	firstStage, firstStep, firstReason := "COUNTEREXAMPLE", "minimum-required", "COUNTEREXAMPLE_REQUIRED"
	secondStage, secondStep, secondReason := "RESOLUTION", "await-proof", "COUNTEREXAMPLE_UNRESOLVED"
	thirdStage, thirdStep, thirdReason := "COMPILE_DECISION", "promote-after-resolution", reason
	if scenario.Counterexample != nil {
		firstStage, firstStep, firstReason = "COUNTEREXAMPLE", "shrink", scenario.Counterexample.Reason
		if unknownCoordinate(*scenario.Counterexample) {
			firstStage, firstStep, firstReason = "UNKNOWN", "UNKNOWN", "UNKNOWN"
		}
	}
	if scenario.Resolution != nil {
		secondStage, secondStep, secondReason = scenario.Resolution.Stage, scenario.Resolution.Step, scenario.Resolution.Reason
	}
	return []cf.ClaimTransition{
		base(1, "CANDIDATE", "COUNTEREXAMPLE", statusOne, firstStage, firstStep, firstReason),
		base(2, "COUNTEREXAMPLE", "RESOLUTION", statusTwo, secondStage, secondStep, secondReason),
		base(3, "RESOLUTION", "COMPILE_DECISION", statusThree, coordinate.Stage, coordinate.Step, coordinate.Reason),
	}
}

func validateSource(source []byte, contract cf.Contract) error {
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
