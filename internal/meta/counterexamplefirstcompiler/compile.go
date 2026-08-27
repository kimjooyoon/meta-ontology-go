package counterexamplefirstcompiler

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	cf "github.com/kimjooyoon/meta-ontology-go/internal/meta/counterexamplefirst"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

var requiredMetaActivities = []string{
	"CanonicalEntityID",
	"DiscoverMinimalCounterexample",
	"BindResolutionEvidence",
	"PromoteOnlyAfterResolution",
}

var requiredMetaEdges = []struct {
	from    string
	through string
	to      string
}{
	{from: "CanonicalEntityID", through: "CompilationClaim", to: "DiscoverMinimalCounterexample"},
	{from: "DiscoverMinimalCounterexample", through: "MinimalCounterexample", to: "BindResolutionEvidence"},
	{from: "BindResolutionEvidence", through: "ResolutionEvidence", to: "PromoteOnlyAfterResolution"},
}

func Compile(contract cf.Contract, head, sourcePath string, source []byte, corpus cf.ScenarioCorpus) ([]cf.DecisionReceipt, error) {
	if !cf.ValidContract(contract) || sourcePath != contract.SourcePath || corpus.Schema != cf.CorpusSchema || corpus.Version != 1 || len(corpus.Scenarios) != cf.CaseCount {
		return nil, fmt.Errorf("COUNTEREXAMPLE_INPUT_CONTRACT_UNKNOWN")
	}
	program, rule, err := observeProgram(sourcePath, source)
	if err != nil {
		return nil, err
	}
	if !program.ParseOK || !program.LowerOK {
		return nil, fmt.Errorf("COUNTEREXAMPLE_SOURCE_UNOBSERVED")
	}
	if !program.MetaOperation.Authorized {
		return nil, fmt.Errorf("COUNTEREXAMPLE_META_OPERATION_UNAUTHORIZED:%s", program.MetaOperation.Reason)
	}
	if rule != contract.Predicates[0].Rule {
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
		if !ok || scenario.Candidate.ID == "" || scenario.Candidate.ClaimID != spec.ClaimID || scenario.Candidate.PredicateID != spec.PredicateID || scenario.Candidate.Claim != spec.Proposition {
			return nil, fmt.Errorf("COUNTEREXAMPLE_SCENARIO_BINDING_MISMATCH:%s", spec.ID)
		}
		predicateSpec, ok := predicateByID(contract.Predicates, spec.PredicateID)
		if !ok {
			return nil, fmt.Errorf("COUNTEREXAMPLE_PREDICATE_MISSING:%s", spec.PredicateID)
		}
		receipts = append(receipts, compileScenario(contract, head, sourcePath, source, program, rule, predicateSpec, spec, scenario))
	}
	return receipts, nil
}

func predicateByID(predicates []cf.PredicateSpec, id string) (cf.PredicateSpec, bool) {
	for _, predicate := range predicates {
		if predicate.ID == id {
			return predicate, true
		}
	}
	return cf.PredicateSpec{}, false
}

func compileScenario(contract cf.Contract, head, sourcePath string, source []byte, program cf.ExecutionObservation, rule string, predicateSpec cf.PredicateSpec, spec cf.CaseSpec, scenario cf.Scenario) cf.DecisionReceipt {
	claimID := scenario.Candidate.ClaimID
	propositionDigest := cf.PropositionDigest(claimID, scenario.Candidate.Claim, scenario.Candidate.PredicateID)
	observedPredicateSpec := predicateSpec
	observedPredicateSpec.Rule = rule
	candidateObservation, candidatePredicate := observeInput(scenario.Candidate.ID, scenario.Candidate.Source, observedPredicateSpec, program.SemanticDigest)
	if !program.MetaOperation.Authorized {
		return unauthorizedReceipt(contract, head, sourcePath, source, program, spec, scenario, claimID, propositionDigest, candidateObservation, candidatePredicate)
	}
	var counterexample *cf.Counterexample
	if candidatePredicate.ViolationObserved {
		counterexample = discoverMinimal(scenario.Candidate.ID, scenario.Candidate.Source, observedPredicateSpec, program.SemanticDigest, candidateObservation, candidatePredicate)
	}

	var resolutionEvidence *cf.ResolutionEvidence
	if counterexample != nil && scenario.Resolution != nil {
		if repaired, ok := repairSource(counterexample.Source, scenario.Resolution.Operation, rule); ok {
			observation, predicate := observeInput(scenario.Resolution.ID, &repaired, observedPredicateSpec, program.SemanticDigest)
			deltaDigest := cf.DigestValue(struct {
				Operation string `json:"operation"`
				Before    string `json:"before"`
				After     string `json:"after"`
			}{scenario.Resolution.Operation, counterexample.SourceDigest, observation.SourceDigest})
			resolutionEvidence = &cf.ResolutionEvidence{
				ID: scenario.Resolution.ID, CounterexampleID: counterexample.ID, InputID: scenario.Resolution.ID,
				OriginalSourceDigest: counterexample.SourceDigest, RepairSourceDigest: observation.SourceDigest,
				RepairDeltaDigest: deltaDigest, RepairOperation: scenario.Resolution.Operation,
				SameClaim: true, SamePredicate: true, ClaimID: claimID, PropositionDigest: propositionDigest, PredicateID: predicateSpec.ID,
				Observation: observation, Predicate: predicate, Stage: "RESOLUTION", Step: "repair-and-rerun-minimal-counterexample",
				Reason: resolutionReason(predicate), ProofChoice: "COUNTEREXAMPLE_RESOLUTION",
				MetaOperation: "resolve-minimal-counterexample", Producer: "counterexample-resolution-witness", Consumer: cf.ProducerID,
			}
		}
	}

	decision, resolution, reason, coordinate := decisionFor(candidatePredicate, counterexample, resolutionEvidence)
	input := cf.DecisionInput{ClaimID: claimID, PropositionDigest: propositionDigest, PredicateID: predicateSpec.ID, CandidateID: scenario.Candidate.ID, CandidateDigest: candidateObservation.SourceDigest, RequiredBeforeCompile: true}
	if counterexample != nil {
		input.CounterexampleID = counterexample.ID
		input.CounterexampleDigest = cf.CounterexampleDigest(*counterexample)
	}
	if resolutionEvidence != nil {
		input.ResolutionID = resolutionEvidence.ID
		input.ResolutionDigest = cf.ResolutionDigest(*resolutionEvidence)
		input.RepairDeltaDigest = resolutionEvidence.RepairDeltaDigest
	}
	receipt := cf.DecisionReceipt{
		Schema: cf.ReceiptSchema, ContractID: contract.ID, HeadSHA: head,
		SourcePath: sourcePath, SourceDigest: cf.DigestBytes(source), SemanticDigest: program.SemanticDigest,
		ProgramMetaOperation: program.MetaOperation, ScenarioID: scenario.ID, ClaimID: claimID,
		PropositionDigest: propositionDigest, PredicateID: predicateSpec.ID, Producer: contract.Producer, Consumer: contract.Consumer,
		MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice, Decision: decision, Resolution: resolution,
		Reason: reason, Coordinate: coordinate, CandidateObservation: candidateObservation, CandidatePredicate: candidatePredicate,
		Counterexample: counterexample, ResolutionEvidence: resolutionEvidence, DecisionInput: input,
		ClaimTransitions: claimTransitions(contract, spec, scenario.ID, claimID, propositionDigest, predicateSpec.ID, candidateObservation, candidatePredicate, counterexample, resolutionEvidence, coordinate, reason),
		Effects:          cf.Effects{RepositoryWrites: 0, MutationAuthority: "UNKNOWN", CapabilityEvidence: ""},
	}
	receipt.Digest = cf.ReceiptDigest(receipt)
	return receipt
}

func unauthorizedReceipt(contract cf.Contract, head, sourcePath string, source []byte, program cf.ExecutionObservation, spec cf.CaseSpec, scenario cf.Scenario, claimID, propositionDigest string, observation cf.ExecutionObservation, predicate cf.PredicateObservation) cf.DecisionReceipt {
	coordinate := cf.Coordinate{Stage: "META_OPERATION", Step: "graph-authorization", Reason: "META_OPERATION_NOT_AUTHORIZED"}
	reason := "META_OPERATION_NOT_AUTHORIZED"
	receipt := cf.DecisionReceipt{
		Schema: cf.ReceiptSchema, ContractID: contract.ID, HeadSHA: head, SourcePath: sourcePath,
		SourceDigest: cf.DigestBytes(source), SemanticDigest: program.SemanticDigest, ProgramMetaOperation: program.MetaOperation,
		ScenarioID: scenario.ID, ClaimID: claimID, PropositionDigest: propositionDigest, PredicateID: spec.PredicateID,
		Producer: contract.Producer, Consumer: contract.Consumer, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice,
		Decision: "UNKNOWN", Resolution: "LOWER_RESOLUTION", Reason: reason, Coordinate: coordinate,
		CandidateObservation: observation, CandidatePredicate: predicate, DecisionInput: cf.DecisionInput{
			ClaimID: claimID, PropositionDigest: propositionDigest, PredicateID: spec.PredicateID, CandidateID: scenario.Candidate.ID,
			CandidateDigest: observation.SourceDigest, RequiredBeforeCompile: true,
		},
		ClaimTransitions: claimTransitions(contract, spec, scenario.ID, claimID, propositionDigest, spec.PredicateID, observation, predicate, nil, nil, coordinate, reason),
		Effects:          cf.Effects{RepositoryWrites: 0, MutationAuthority: "UNKNOWN"},
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

func observeInput(inputID string, source *string, predicateSpec cf.PredicateSpec, baselineSemantic string) (cf.ExecutionObservation, cf.PredicateObservation) {
	if source == nil {
		observation := cf.ExecutionObservation{InputID: inputID, OutputDigest: cf.DigestValue([]string{inputID, "SOURCE_NOT_PROVIDED"})}
		return observation, finalizePredicate(cf.PredicateObservation{PredicateID: predicateSpec.ID, Rule: predicateSpec.Rule, UnknownObserved: true, Reason: "SOURCE_NOT_PROVIDED"})
	}
	observation := execute(inputID, []byte(*source))
	return observation, evaluatePredicate(predicateSpec, observation, baselineSemantic)
}

func execute(inputID string, source []byte) cf.ExecutionObservation {
	file, diagnostics := syntax.ParseFile(inputID, string(source))
	observation := cf.ExecutionObservation{InputID: inputID, SourceDigest: cf.DigestBytes(source), SourceBytes: len(source)}
	for _, diagnostic := range diagnostics {
		observation.ParseDiagnostics = append(observation.ParseDiagnostics, cf.DiagnosticObservation{Code: string(diagnostic.Code), Line: diagnostic.Span.Start.Line, Column: diagnostic.Span.Start.Column})
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
		observation.Nodes = append(observation.Nodes, cf.NodeObservation{ID: node.ID.String(), Kind: node.Kind.String(), Namespace: node.Namespace.String(), Name: node.Name, ValueProgram: node.ValueProgram})
	}
	for _, fact := range ir.Graph.AllFacts() {
		observation.Facts = append(observation.Facts, cf.FactObservation{Subject: fact.Subject.String(), Predicate: fact.Predicate.String(), Object: fact.Object.String(), Status: fact.Status.String()})
	}
	observation.MetaOperation = inspectMetaOperation(observation.Nodes, observation.Facts)
	observation.OutputDigest = cf.DigestValue(struct {
		SemanticDigest string                      `json:"semantic_digest"`
		Nodes          []cf.NodeObservation        `json:"nodes"`
		Facts          []cf.FactObservation        `json:"facts"`
		MetaOperation  cf.MetaOperationObservation `json:"meta_operation"`
	}{observation.SemanticDigest, observation.Nodes, observation.Facts, observation.MetaOperation})
	return observation
}

func inspectMetaOperation(nodes []cf.NodeObservation, facts []cf.FactObservation) cf.MetaOperationObservation {
	meta := cf.MetaOperationObservation{RequiredActivities: append([]string{}, requiredMetaActivities...)}
	byID := make(map[string]cf.NodeObservation, len(nodes))
	activities := make(map[string]bool, len(requiredMetaActivities))
	for _, node := range nodes {
		byID[node.ID] = node
		if node.Kind == semantic.Activity.String() {
			activities[node.Name] = true
		}
	}
	meta.ActivitiesPresent = true
	for _, name := range requiredMetaActivities {
		if !activities[name] {
			meta.ActivitiesPresent = false
		}
	}
	for _, edge := range requiredMetaEdges {
		for _, generated := range facts {
			if generated.Predicate != string(semantic.WasGeneratedBy) || byID[generated.Subject].Kind != semantic.Entity.String() || byID[generated.Object].Name != edge.from || byID[generated.Subject].Name != edge.through {
				continue
			}
			for _, used := range facts {
				if used.Predicate == string(semantic.Used) && used.Subject == activityIDByName(byID, edge.to) && used.Object == generated.Subject {
					meta.Edges = append(meta.Edges, cf.GraphEdgeObservation{From: edge.from, Through: edge.through, To: edge.to})
				}
			}
		}
	}
	for _, edge := range meta.Edges {
		if len(meta.ActivityOrder) == 0 || meta.ActivityOrder[len(meta.ActivityOrder)-1] != edge.From {
			meta.ActivityOrder = append(meta.ActivityOrder, edge.From)
		}
		if len(meta.ActivityOrder) == 0 || meta.ActivityOrder[len(meta.ActivityOrder)-1] != edge.To {
			meta.ActivityOrder = append(meta.ActivityOrder, edge.To)
		}
	}
	meta.Connected = len(meta.Edges) == len(requiredMetaEdges)
	for index, edge := range requiredMetaEdges {
		if index >= len(meta.Edges) || meta.Edges[index] != (cf.GraphEdgeObservation{From: edge.from, Through: edge.through, To: edge.to}) {
			meta.Connected = false
		}
	}
	meta.Authorized = meta.ActivitiesPresent && meta.Connected
	switch {
	case !meta.ActivitiesPresent:
		meta.Reason = "META_OPERATION_ACTIVITY_MISSING"
	case !meta.Connected:
		meta.Reason = "META_OPERATION_EDGE_MISSING_OR_REORDERED"
	default:
		meta.Reason = "META_OPERATION_GRAPH_AUTHORIZED"
	}
	meta.Digest = cf.DigestValue(struct {
		Required []string                  `json:"required"`
		Order    []string                  `json:"order"`
		Edges    []cf.GraphEdgeObservation `json:"edges"`
		Present  bool                      `json:"present"`
		Linked   bool                      `json:"linked"`
		Reason   string                    `json:"reason"`
	}{meta.RequiredActivities, meta.ActivityOrder, meta.Edges, meta.ActivitiesPresent, meta.Connected, meta.Reason})
	return meta
}

func activityIDByName(nodes map[string]cf.NodeObservation, name string) string {
	ids := make([]string, 0, 1)
	for id, node := range nodes {
		if node.Kind == semantic.Activity.String() && node.Name == name {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	if len(ids) == 0 {
		return ""
	}
	return ids[0]
}

func evaluatePredicate(spec cf.PredicateSpec, observation cf.ExecutionObservation, baselineSemantic string) cf.PredicateObservation {
	predicate := cf.PredicateObservation{PredicateID: spec.ID, Rule: spec.Rule}
	if !observation.ParseOK || !observation.LowerOK {
		return finalizePredicate(cf.PredicateObservation{PredicateID: spec.ID, Rule: spec.Rule, UnknownObserved: true, Reason: "EXECUTION_UNOBSERVED"})
	}
	predicate.Applicable = true
	identityViolation := observesIdentityViolation(spec.Rule, observation.Nodes)
	switch spec.Kind {
	case "ENTITY_ID_DRIFT":
		predicate.ViolationObserved = identityViolation
		predicate.PassObserved = !identityViolation
		predicate.Reason = chooseReason(identityViolation, "ENTITY_ID_DRIFT", "IDENTITY_CANONICAL")
	case "CANONICAL_SOURCE":
		predicate.ViolationObserved = identityViolation
		predicate.PassObserved = !identityViolation
		predicate.Reason = chooseReason(identityViolation, "CANONICAL_SOURCE_NOT_ADMISSIBLE", "CANONICAL_SOURCE_OBSERVED")
	case "RESOLUTION_REQUIRED":
		predicate.ViolationObserved = identityViolation
		predicate.PassObserved = !identityViolation
		predicate.Reason = chooseReason(identityViolation, "REPAIR_REQUIRED", "NO_REPAIR_REQUIRED")
	case "SEMANTIC_DIGEST":
		predicate.ViolationObserved = observation.SemanticDigest != baselineSemantic
		predicate.PassObserved = !predicate.ViolationObserved
		predicate.Reason = chooseReason(predicate.ViolationObserved, "SEMANTIC_DIGEST_CHANGED", "SEMANTIC_DIGEST_PRESERVED")
	case "SOURCE_ACQUISITION":
		predicate.PassObserved = observation.SourceDigest != ""
		predicate.Reason = chooseReason(predicate.PassObserved == false, "SOURCE_NOT_PROVIDED", "SOURCE_ACQUIRED")
	default:
		predicate.UnknownObserved = true
		predicate.Applicable = false
		predicate.Reason = "PREDICATE_KIND_UNKNOWN"
	}
	return finalizePredicate(predicate)
}

func observesIdentityViolation(rule string, nodes []cf.NodeObservation) bool {
	for _, node := range nodes {
		if node.Kind == semantic.Entity.String() && node.ID != canonicalEntityID(rule, node) {
			return true
		}
	}
	return false
}

func chooseReason(condition bool, failure, pass string) string {
	if condition {
		return failure
	}
	return pass
}

func finalizePredicate(predicate cf.PredicateObservation) cf.PredicateObservation {
	predicate.EvidenceDigest = cf.DigestValue(struct {
		PredicateID       string `json:"predicate_id"`
		Rule              string `json:"rule"`
		Applicable        bool   `json:"applicable"`
		ViolationObserved bool   `json:"violation_observed"`
		PassObserved      bool   `json:"pass_observed"`
		UnknownObserved   bool   `json:"unknown_observed"`
		Reason            string `json:"reason"`
	}{predicate.PredicateID, predicate.Rule, predicate.Applicable, predicate.ViolationObserved, predicate.PassObserved, predicate.UnknownObserved, predicate.Reason})
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

func discoverMinimal(inputID string, source *string, predicateSpec cf.PredicateSpec, baselineSemantic string, observation cf.ExecutionObservation, predicate cf.PredicateObservation) *cf.Counterexample {
	if source == nil {
		return nil
	}
	currentSource := *source
	currentObservation, currentPredicate := observation, predicate
	var trace []cf.ShrinkObservation
	numerator, denominator := 0, 0
	for {
		candidates := shrinkCandidates(currentSource, predicateSpec.Rule)
		if len(candidates) == 0 {
			break
		}
		var nextSource string
		var nextObservation cf.ExecutionObservation
		var nextPredicate cf.PredicateObservation
		for index, candidate := range candidates {
			candidateObservation, candidatePredicate := observeInput(fmt.Sprintf("%s/shrink-%d", inputID, index), &candidate, predicateSpec, baselineSemantic)
			trace = append(trace, cf.ShrinkObservation{CandidateDigest: candidateObservation.SourceDigest, SourceBytes: len(candidate), Observation: candidateObservation, Predicate: candidatePredicate})
			if candidatePredicate.ViolationObserved && (nextSource == "" || len(candidate) < len(nextSource)) {
				nextSource, nextObservation, nextPredicate = candidate, candidateObservation, candidatePredicate
			}
		}
		if nextSource == "" {
			for _, step := range trace[len(trace)-len(candidates):] {
				denominator++
				if !step.Predicate.ViolationObserved {
					numerator++
				}
			}
			break
		}
		currentSource, currentObservation, currentPredicate = nextSource, nextObservation, nextPredicate
	}
	ceDigest := cf.DigestBytes([]byte(predicateSpec.ID + "|" + currentObservation.SourceDigest))
	return &cf.Counterexample{
		ID: "ce-" + ceDigest[len(ceDigest)-12:], Source: currentSource, SourceDigest: currentObservation.SourceDigest,
		SourceBytes: currentObservation.SourceBytes, Observation: currentObservation, Predicate: currentPredicate, ShrinkTrace: trace,
		FiniteNeighborhoodNumerator: numerator, FiniteNeighborhoodDenominator: denominator, FiniteNeighborhoodIrreducible: denominator > 0 && numerator == denominator,
		Stage: "COUNTEREXAMPLE", Step: "finite-neighborhood-shrink", Reason: "FINITE_NEIGHBORHOOD_IRREDUCIBLE",
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
	return canonicalEntityID(rule, cf.NodeObservation{Name: "CompilationClaim"})
}

func repairSource(source, operation, rule string) (string, bool) {
	if operation != "canonicalize-entity-id" {
		return "", false
	}
	const drift = "gooo://counterexample-first/entity/compilation-claim?drift=1"
	if !strings.Contains(source, drift) {
		return "", false
	}
	return strings.Replace(source, drift, canonicalClaimID(rule), 1), true
}

func decisionFor(predicate cf.PredicateObservation, counterexample *cf.Counterexample, evidence *cf.ResolutionEvidence) (string, string, string, cf.Coordinate) {
	if predicate.UnknownObserved {
		return "UNKNOWN", "LOWER_RESOLUTION", "SOURCE_NOT_PROVIDED", cf.Coordinate{Stage: "INPUT_OBSERVATION", Step: "source-acquisition", Reason: "SOURCE_NOT_PROVIDED"}
	}
	if counterexample == nil {
		return "FAIL_CLOSED", "LOWER_RESOLUTION", "COUNTEREXAMPLE_REQUIRED", cf.Coordinate{Stage: "COUNTEREXAMPLE", Step: "discover", Reason: "COUNTEREXAMPLE_REQUIRED"}
	}
	if !counterexample.FiniteNeighborhoodIrreducible {
		return "REFUTED", "LOWER_RESOLUTION", "FINITE_NEIGHBORHOOD_NOT_IRREDUCIBLE", cf.Coordinate{Stage: "COUNTEREXAMPLE", Step: "finite-neighborhood-shrink", Reason: "FINITE_NEIGHBORHOOD_NOT_IRREDUCIBLE"}
	}
	if evidence == nil || !evidence.Predicate.PassObserved || !evidence.SameClaim || !evidence.SamePredicate || evidence.RepairDeltaDigest == "" {
		return "REFUTED", "LOWER_RESOLUTION", "COUNTEREXAMPLE_UNRESOLVED", cf.Coordinate{Stage: "RESOLUTION", Step: "await-repair-proof", Reason: "COUNTEREXAMPLE_UNRESOLVED"}
	}
	return "PASS", "EXACT", "COUNTEREXAMPLE_RESOLVED", cf.Coordinate{Stage: "COMPILE_DECISION", Step: "promote-after-resolution", Reason: "COUNTEREXAMPLE_RESOLVED"}
}

func resolutionReason(predicate cf.PredicateObservation) string {
	if predicate.PassObserved {
		return "REPAIR_RERUN_SAME_MINIMAL_SOURCE_PASSED"
	}
	return "REPAIR_RERUN_DID_NOT_PASS"
}

func claimTransitions(contract cf.Contract, spec cf.CaseSpec, scenarioID, claimID, propositionDigest, predicateID string, observation cf.ExecutionObservation, predicate cf.PredicateObservation, counterexample *cf.Counterexample, evidence *cf.ResolutionEvidence, coordinate cf.Coordinate, reason string) []cf.ClaimTransition {
	state, status, firstReason := "OPEN", "LOWER_RESOLUTION", "COUNTEREXAMPLE_REQUIRED"
	firstStage, firstStep := "COUNTEREXAMPLE", "discover"
	if predicate.UnknownObserved {
		firstReason, firstStage, firstStep = "SOURCE_NOT_PROVIDED", "INPUT_OBSERVATION", "source-acquisition"
	} else if counterexample != nil {
		state, status, firstReason, firstStep = "REFUTED", "REFUTED", counterexample.Reason, "finite-neighborhood-shrink"
	}
	transitions := []cf.ClaimTransition{makeTransition(contract, spec, scenarioID, claimID, propositionDigest, predicateID, 1, "OPEN", state, status, firstStage, firstStep, firstReason, predicate.EvidenceDigest)}
	if state == "REFUTED" && evidence != nil && evidence.Predicate.PassObserved && evidence.SameClaim && evidence.SamePredicate {
		transitions = append(transitions, makeTransition(contract, spec, scenarioID, claimID, propositionDigest, predicateID, 2, "REFUTED", "DISCHARGED", "DISCHARGED", "RESOLUTION", "repair-and-rerun-minimal-counterexample", evidence.Reason, cf.ResolutionDigest(*evidence)))
		transitions = append(transitions, makeTransition(contract, spec, scenarioID, claimID, propositionDigest, predicateID, 3, "DISCHARGED", "DISCHARGED", "PROMOTED", coordinate.Stage, coordinate.Step, reason, evidence.Predicate.EvidenceDigest))
		return transitions
	}
	observationEvidence := observation.SourceDigest
	if observationEvidence == "" {
		observationEvidence = observation.OutputDigest
	}
	transitions = append(transitions, makeTransition(contract, spec, scenarioID, claimID, propositionDigest, predicateID, 2, state, state, status, "RESOLUTION", "await-repair-proof", reason, observationEvidence))
	transitions = append(transitions, makeTransition(contract, spec, scenarioID, claimID, propositionDigest, predicateID, 3, state, state, status, coordinate.Stage, coordinate.Step, coordinate.Reason, observation.OutputDigest))
	return transitions
}

func makeTransition(contract cf.Contract, spec cf.CaseSpec, scenarioID, claimID, propositionDigest, predicateID string, sequence int, from, to, status, stage, step, reason, evidence string) cf.ClaimTransition {
	predicateDigest := cf.DigestValue(struct {
		PredicateID string `json:"predicate_id"`
		Evidence    string `json:"evidence"`
	}{predicateID, evidence})
	return cf.ClaimTransition{Sequence: sequence, ClaimID: claimID, PropositionDigest: propositionDigest, PredicateID: predicateID, From: from, To: to, Status: status, Stage: stage, Step: step, Reason: reason,
		Producer: contract.Producer, Consumer: contract.Consumer, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice,
		EvidenceDigest: evidence, PredicateDigest: predicateDigest}
}
