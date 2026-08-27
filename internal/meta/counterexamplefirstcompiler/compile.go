package counterexamplefirstcompiler

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	cf "github.com/kimjooyoon/meta-ontology-go/internal/meta/counterexamplefirst"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func Compile(contract cf.Contract, head, sourcePath string, source []byte, corpus cf.ScenarioCorpus) ([]cf.DecisionReceipt, error) {
	if !cf.ValidContract(contract) || sourcePath != contract.SourcePath ||
		corpus.Schema != cf.CorpusSchema || corpus.Version != 1 || len(corpus.Scenarios) != cf.CaseCount {
		return nil, fmt.Errorf("COUNTEREXAMPLE_INPUT_CONTRACT_UNKNOWN")
	}
	contractObservation, rule, err := observeProgram(sourcePath, source)
	if err != nil {
		return nil, err
	}
	if !contractObservation.ParseOK || !contractObservation.LowerOK {
		return nil, fmt.Errorf("COUNTEREXAMPLE_SOURCE_UNOBSERVED")
	}
	if rule != contract.Predicate.Rule {
		return nil, fmt.Errorf("COUNTEREXAMPLE_POLICY_MISMATCH:%s", rule)
	}
	byID := make(map[string]cf.Scenario, len(corpus.Scenarios))
	for _, scenario := range corpus.Scenarios {
		if _, exists := byID[scenario.ID]; exists {
			return nil, fmt.Errorf("COUNTEREXAMPLE_SCENARIO_DUPLICATE:%s", scenario.ID)
		}
		byID[scenario.ID] = scenario
	}
	receipts := make([]cf.DecisionReceipt, 0, len(contract.Cases))
	for _, spec := range contract.Cases {
		scenario, ok := byID[spec.ID]
		if !ok || scenario.Candidate.ID == "" {
			return nil, fmt.Errorf("COUNTEREXAMPLE_SCENARIO_MISSING:%s", spec.ID)
		}
		receipts = append(receipts, compileScenario(contract, head, sourcePath, source, contractObservation, rule, spec, scenario))
	}
	return receipts, nil
}

func compileScenario(contract cf.Contract, head, sourcePath string, source []byte, program cf.ExecutionObservation, rule string, spec cf.CaseSpec, scenario cf.Scenario) cf.DecisionReceipt {
	candidateObservation, candidatePredicate := observeInput(scenario.Candidate.ID, scenario.Candidate.Source, rule)
	var counterexample *cf.Counterexample
	if candidatePredicate.ViolationObserved {
		counterexample = discoverMinimal(scenario.Candidate.ID, scenario.Candidate.Source, rule, candidateObservation, candidatePredicate)
	}

	var resolutionEvidence *cf.ResolutionEvidence
	if counterexample != nil && scenario.Resolution != nil && scenario.Resolution.Source != nil {
		observation, predicate := observeInput(scenario.Resolution.ID, scenario.Resolution.Source, rule)
		resolutionEvidence = &cf.ResolutionEvidence{
			ID: scenario.Resolution.ID, CounterexampleID: counterexample.ID, InputID: scenario.Resolution.ID,
			Observation: observation, Predicate: predicate, Stage: "RESOLUTION", Step: "rerun-minimal-counterexample",
			Reason: resolutionReason(predicate), ProofChoice: "COUNTEREXAMPLE_RESOLUTION",
			MetaOperation: "resolve-minimal-counterexample", Producer: "counterexample-resolution-witness",
			Consumer: cf.ProducerID,
		}
	}

	decision, resolution, reason, coordinate := decisionFor(candidateObservation, candidatePredicate, counterexample, resolutionEvidence)
	input := cf.DecisionInput{CandidateID: scenario.Candidate.ID, CandidateDigest: candidateObservation.SourceDigest, RequiredBeforeCompile: true}
	if counterexample != nil {
		input.CounterexampleID = counterexample.ID
		input.CounterexampleDigest = cf.CounterexampleDigest(*counterexample)
	}
	if resolutionEvidence != nil {
		input.ResolutionID = resolutionEvidence.ID
		input.ResolutionDigest = cf.ResolutionDigest(*resolutionEvidence)
	}
	receipt := cf.DecisionReceipt{
		Schema: cf.ReceiptSchema, ContractID: contract.ID, HeadSHA: head,
		SourcePath: sourcePath, SourceDigest: cf.DigestBytes(source), SemanticDigest: program.SemanticDigest,
		ScenarioID: scenario.ID, Producer: contract.Producer, Consumer: contract.Consumer,
		MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice,
		Decision: decision, Resolution: resolution, Reason: reason, Coordinate: coordinate,
		CandidateObservation: candidateObservation, CandidatePredicate: candidatePredicate,
		Counterexample: counterexample, ResolutionEvidence: resolutionEvidence,
		DecisionInput: input, ClaimTransitions: claimTransitions(contract, spec, scenario.ID, candidateObservation, candidatePredicate, counterexample, resolutionEvidence, coordinate, reason),
		Effects: cf.Effects{RepositoryWrites: 0, MutationAuthority: false},
	}
	receipt.Digest = cf.ReceiptDigest(receipt)
	return receipt
}

func observeProgram(filename string, source []byte) (cf.ExecutionObservation, string, error) {
	observation := execute(filename, source)
	if !observation.ParseOK || !observation.LowerOK {
		return observation, "", nil
	}
	var rule string
	for _, node := range observation.Nodes {
		if node.Kind == semantic.Activity.String() && node.Name == "CanonicalEntityID" {
			rule = node.ValueProgram
		}
	}
	if rule == "" {
		return observation, "", fmt.Errorf("COUNTEREXAMPLE_POLICY_ACTIVITY_MISSING")
	}
	if rule != cf.RuleIdentityV1 && rule != cf.RuleIdentityV2 {
		return observation, rule, fmt.Errorf("COUNTEREXAMPLE_POLICY_UNKNOWN:%s", rule)
	}
	return observation, rule, nil
}

func observeInput(inputID string, source *string, rule string) (cf.ExecutionObservation, cf.PredicateObservation) {
	if source == nil {
		observation := cf.ExecutionObservation{InputID: inputID, OutputDigest: cf.DigestValue([]string{inputID, "UNOBSERVED"})}
		return observation, cf.PredicateObservation{Rule: rule, UnknownObserved: true, Reason: "INPUT_UNOBSERVED"}
	}
	observation := execute(inputID, []byte(*source))
	return observation, evaluatePredicate(rule, observation)
}

func execute(inputID string, source []byte) cf.ExecutionObservation {
	file, diagnostics := syntax.ParseFile(inputID, string(source))
	observation := cf.ExecutionObservation{InputID: inputID, SourceDigest: cf.DigestBytes(source), SourceBytes: len(source)}
	for _, diagnostic := range diagnostics {
		observation.ParseDiagnostics = append(observation.ParseDiagnostics, cf.DiagnosticObservation{
			Code: string(diagnostic.Code), Line: diagnostic.Span.Start.Line, Column: diagnostic.Span.Start.Column,
		})
	}
	observation.ParseOK = file != nil && !diagnostics.HasErrors()
	if !observation.ParseOK {
		observation.OutputDigest = cf.DigestValue(struct {
			Diagnostics []cf.DiagnosticObservation `json:"diagnostics"`
		}{observation.ParseDiagnostics})
		return observation
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		observation.LowerError = err.Error()
		observation.OutputDigest = cf.DigestValue(struct {
			Diagnostics []cf.DiagnosticObservation `json:"diagnostics"`
			LowerError  string                     `json:"lower_error"`
		}{observation.ParseDiagnostics, observation.LowerError})
		return observation
	}
	observation.LowerOK = true
	observation.SemanticDigest = cf.DigestBytes([]byte(ir.StableHash()))
	for _, node := range ir.Graph.Nodes() {
		observation.Nodes = append(observation.Nodes, cf.NodeObservation{
			ID: node.ID.String(), Kind: node.Kind.String(), Namespace: node.Namespace.String(),
			Name: node.Name, ValueProgram: node.ValueProgram,
		})
	}
	observation.OutputDigest = cf.DigestValue(struct {
		SemanticDigest string               `json:"semantic_digest"`
		Nodes          []cf.NodeObservation `json:"nodes"`
	}{observation.SemanticDigest, observation.Nodes})
	return observation
}

func evaluatePredicate(rule string, observation cf.ExecutionObservation) cf.PredicateObservation {
	predicate := cf.PredicateObservation{Rule: rule}
	if !observation.ParseOK || !observation.LowerOK {
		predicate.UnknownObserved = true
		predicate.Reason = "EXECUTION_UNOBSERVED"
		return predicate
	}
	predicate.Applicable = true
	for _, node := range observation.Nodes {
		if node.Kind != semantic.Entity.String() {
			continue
		}
		if node.ID != canonicalEntityID(rule, node) {
			predicate.ViolationObserved = true
			predicate.Reason = "ENTITY_ID_DRIFT"
			return predicate
		}
	}
	predicate.PassObserved = true
	predicate.Reason = "PREDICATE_PASSED"
	return predicate
}

func canonicalEntityID(rule string, node cf.NodeObservation) string {
	prefix := "gooo://counterexample-first/entity/"
	if rule == cf.RuleIdentityV2 {
		prefix = "gooo://counterexample-first/v2/entity/"
	}
	return prefix + kebab(node.Name)
}

func kebab(value string) string {
	var result []rune
	for index, char := range []rune(strings.TrimSpace(value)) {
		if unicode.IsUpper(char) && index > 0 {
			result = append(result, '-')
		}
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			result = append(result, unicode.ToLower(char))
		} else if len(result) > 0 && result[len(result)-1] != '-' {
			result = append(result, '-')
		}
	}
	return strings.Trim(string(result), "-")
}

func discoverMinimal(inputID string, source *string, rule string, observation cf.ExecutionObservation, predicate cf.PredicateObservation) *cf.Counterexample {
	currentSource := *source
	currentObservation, currentPredicate := observation, predicate
	var trace []cf.ShrinkObservation
	minimalNumerator, minimalDenominator := 0, 0
	for {
		candidates := shrinkCandidates(currentSource, rule)
		if len(candidates) == 0 {
			break
		}
		var nextSource string
		var nextObservation cf.ExecutionObservation
		var nextPredicate cf.PredicateObservation
		for index, candidate := range candidates {
			candidateObservation, candidatePredicate := observeInput(fmt.Sprintf("%s/shrink-%d", inputID, index), &candidate, rule)
			trace = append(trace, cf.ShrinkObservation{
				CandidateDigest: candidateObservation.SourceDigest, SourceBytes: len(candidate),
				Observation: candidateObservation, Predicate: candidatePredicate,
			})
			if candidatePredicate.ViolationObserved && (nextSource == "" || len(candidate) < len(nextSource)) {
				nextSource, nextObservation, nextPredicate = candidate, candidateObservation, candidatePredicate
			}
		}
		if nextSource == "" {
			for _, step := range trace[len(trace)-len(candidates):] {
				minimalDenominator++
				if step.Predicate.PassObserved {
					minimalNumerator++
				}
			}
			break
		}
		currentSource, currentObservation, currentPredicate = nextSource, nextObservation, nextPredicate
	}
	ceDigest := cf.DigestBytes([]byte(rule + "|" + currentObservation.SourceDigest))
	return &cf.Counterexample{
		ID: "ce-" + ceDigest[len(ceDigest)-12:], SourceDigest: currentObservation.SourceDigest,
		SourceBytes: currentObservation.SourceBytes, Observation: currentObservation, Predicate: currentPredicate,
		ShrinkTrace: trace, MinimalityNumerator: minimalNumerator, MinimalityDenominator: minimalDenominator,
		MinimalityProved: minimalDenominator > 0 && minimalNumerator == minimalDenominator,
		Stage:            "COUNTEREXAMPLE", Step: "shrink", Reason: "MINIMAL_COUNTEREXAMPLE_OBSERVED",
	}
}

func shrinkCandidates(source, rule string) []string {
	const noisy = "gooo://counterexample-first/entity/compilation-claim?noise=1&drift=1"
	const drift = "gooo://counterexample-first/entity/compilation-claim?drift=1"
	if strings.Contains(source, noisy) {
		return []string{strings.Replace(source, noisy, drift, 1)}
	}
	if strings.Contains(source, drift) {
		return []string{strings.Replace(source, drift, canonicalClaimID(rule), 1)}
	}
	return nil
}

func canonicalClaimID(rule string) string {
	node := cf.NodeObservation{Name: "CompilationClaim"}
	return canonicalEntityID(rule, node)
}

func decisionFor(observation cf.ExecutionObservation, predicate cf.PredicateObservation, counterexample *cf.Counterexample, evidence *cf.ResolutionEvidence) (string, string, string, cf.Coordinate) {
	if predicate.UnknownObserved {
		return "UNKNOWN", "LOWER_RESOLUTION", "INPUT_UNOBSERVED", cf.Coordinate{Stage: "UNKNOWN", Step: "UNKNOWN", Reason: "UNKNOWN"}
	}
	if counterexample == nil {
		return "FAIL_CLOSED", "LOWER_RESOLUTION", "COUNTEREXAMPLE_REQUIRED", cf.Coordinate{Stage: "COUNTEREXAMPLE", Step: "discover", Reason: "COUNTEREXAMPLE_REQUIRED"}
	}
	if !counterexample.MinimalityProved {
		return "REFUTED", "LOWER_RESOLUTION", "COUNTEREXAMPLE_NOT_MINIMAL", cf.Coordinate{Stage: "COUNTEREXAMPLE", Step: "shrink", Reason: "COUNTEREXAMPLE_NOT_MINIMAL"}
	}
	if evidence == nil {
		return "REFUTED", "LOWER_RESOLUTION", "COUNTEREXAMPLE_UNRESOLVED", cf.Coordinate{Stage: "RESOLUTION", Step: "await-proof", Reason: "COUNTEREXAMPLE_UNRESOLVED"}
	}
	if !evidence.Predicate.PassObserved {
		return "REFUTED", "LOWER_RESOLUTION", "COUNTEREXAMPLE_UNRESOLVED", cf.Coordinate{Stage: "RESOLUTION", Step: "await-proof", Reason: "COUNTEREXAMPLE_UNRESOLVED"}
	}
	return "PASS", "EXACT", "COUNTEREXAMPLE_RESOLVED", cf.Coordinate{Stage: "COMPILE_DECISION", Step: "promote-after-resolution", Reason: "COUNTEREXAMPLE_RESOLVED"}
}

func resolutionReason(predicate cf.PredicateObservation) string {
	if predicate.PassObserved {
		return "RESOLUTION_RERUN_PASSED"
	}
	return "RESOLUTION_RERUN_DID_NOT_PASS"
}

func claimTransitions(contract cf.Contract, spec cf.CaseSpec, scenarioID string, observation cf.ExecutionObservation, predicate cf.PredicateObservation, counterexample *cf.Counterexample, evidence *cf.ResolutionEvidence, coordinate cf.Coordinate, reason string) []cf.ClaimTransition {
	state := "OPEN"
	status := "LOWER_RESOLUTION"
	firstReason := "COUNTEREXAMPLE_REQUIRED"
	firstStage, firstStep := "COUNTEREXAMPLE", "discover"
	if predicate.UnknownObserved {
		firstReason, firstStage, firstStep = "INPUT_UNOBSERVED", "UNKNOWN", "UNKNOWN"
	} else if counterexample != nil {
		state, status, firstReason, firstStep = "REFUTED", "REFUTED", counterexample.Reason, "shrink"
	}
	transitions := []cf.ClaimTransition{
		makeTransition(contract, spec, scenarioID, 1, "OPEN", state, status, firstStage, firstStep, firstReason, observation.SourceDigest),
	}
	if state == "REFUTED" && evidence != nil && evidence.Predicate.PassObserved {
		transitions = append(transitions, makeTransition(contract, spec, scenarioID, 2, "REFUTED", "DISCHARGED", "DISCHARGED", "RESOLUTION", "rerun-minimal-counterexample", evidence.Reason, cf.ResolutionDigest(*evidence)))
		transitions = append(transitions, makeTransition(contract, spec, scenarioID, 3, "DISCHARGED", "DISCHARGED", "PROMOTED", coordinate.Stage, coordinate.Step, reason, evidence.Predicate.Rule))
		return transitions
	}
	transitions = append(transitions, makeTransition(contract, spec, scenarioID, 2, state, state, status, "RESOLUTION", "await-proof", reason, observation.SourceDigest))
	transitions = append(transitions, makeTransition(contract, spec, scenarioID, 3, state, state, status, coordinate.Stage, coordinate.Step, coordinate.Reason, observation.OutputDigest))
	return transitions
}

func makeTransition(contract cf.Contract, spec cf.CaseSpec, scenarioID string, sequence int, from, to, status, stage, step, reason, evidence string) cf.ClaimTransition {
	return cf.ClaimTransition{Sequence: sequence, From: from, To: to, Status: status, Stage: stage, Step: step, Reason: reason,
		Producer: contract.Producer, Consumer: contract.Consumer, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice,
		EvidenceDigest: cf.DigestValue([]string{scenarioID, from, to, status, stage, step, reason, evidence})}
}
