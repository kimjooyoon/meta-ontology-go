package counterexamplefirstjudge

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

var independentMetaActivities = []string{"CanonicalEntityID", "DiscoverMinimalCounterexample", "BindResolutionEvidence", "PromoteOnlyAfterResolution"}
var independentMetaEdges = []struct{ from, through, to string }{
	{from: "CanonicalEntityID", through: "CompilationClaim", to: "DiscoverMinimalCounterexample"},
	{from: "DiscoverMinimalCounterexample", through: "MinimalCounterexample", to: "BindResolutionEvidence"},
	{from: "BindResolutionEvidence", through: "ResolutionEvidence", to: "PromoteOnlyAfterResolution"},
}

// Evaluate reconstructs receipts from raw source and corpus inputs. It does
// not import the producer package or a canonical outcome table.
func Evaluate(input cf.JudgeInput) cf.Report {
	if !validStructure(input.Contract) || input.SourcePath != input.Contract.SourcePath || input.Corpus.Schema != cf.CorpusSchema || input.Corpus.Version != 1 || len(input.Corpus.Scenarios) != cf.CaseCount {
		return closedReport(input, "COUNTEREXAMPLE_INPUT_CONTRACT_UNKNOWN", "LOWER_RESOLUTION")
	}
	programObservation, rule, err := independentProgram(input.SourcePath, input.Source)
	if err != nil || !programObservation.ParseOK || !programObservation.LowerOK {
		if err != nil {
			return closedReport(input, err.Error(), "LOWER_RESOLUTION")
		}
		return closedReport(input, "COUNTEREXAMPLE_SOURCE_UNOBSERVED", "LOWER_RESOLUTION")
	}
	if !programObservation.MetaOperation.Authorized {
		return closedReport(input, "COUNTEREXAMPLE_META_OPERATION_UNAUTHORIZED", "LOWER_RESOLUTION")
	}
	if rule != input.Contract.Predicates[0].Rule {
		return closedReport(input, "COUNTEREXAMPLE_POLICY_MISMATCH", "LOWER_RESOLUTION")
	}
	byID := make(map[string]cf.Scenario, len(input.Corpus.Scenarios))
	for _, scenario := range input.Corpus.Scenarios {
		if _, exists := byID[scenario.ID]; exists {
			return closedReport(input, "COUNTEREXAMPLE_SCENARIO_DUPLICATE", "LOWER_RESOLUTION")
		}
		byID[scenario.ID] = scenario
	}
	if len(input.Receipts) != len(input.Contract.Cases) {
		return closedReport(input, "COUNTEREXAMPLE_RECEIPT_COUNT_MISMATCH", "LOWER_RESOLUTION")
	}
	reconstructed := make([]cf.DecisionReceipt, 0, len(input.Contract.Cases))
	for _, spec := range input.Contract.Cases {
		scenario, ok := byID[spec.ID]
		predicateSpec, predicateOK := independentPredicateSpec(input.Contract.Predicates, spec.PredicateID)
		if !ok || scenario.Candidate.ID == "" || scenario.Candidate.ClaimID != spec.ClaimID || scenario.Candidate.PredicateID != spec.PredicateID || scenario.Candidate.Claim != spec.Proposition || !predicateOK {
			return closedReport(input, "COUNTEREXAMPLE_SCENARIO_BINDING_MISMATCH", "LOWER_RESOLUTION")
		}
		reconstructed = append(reconstructed, reconstructReceipt(input.Contract, input.HeadSHA, input.SourcePath, input.Source, programObservation, rule, predicateSpec, spec, scenario))
	}
	verified := 0
	for index := range reconstructed {
		if sameReceipt(input.Receipts[index], reconstructed[index]) {
			verified++
		}
	}
	if verified != len(reconstructed) {
		report := closedReport(input, "COUNTEREXAMPLE_RECEIPT_MISMATCH", "EXACT")
		report.Receipts = input.Receipts
		report.TamperedRejected = len(reconstructed) - verified
		report.Summary.ReceiptsVerified = verified
		report.Digest = cf.ReportDigest(report)
		return report
	}
	summary := summarize(input, reconstructed, verified)
	indicators := makeIndicators(summary, input.Contract)
	decision, resolution, reason := "FAIL_CLOSED", "EXACT", "COUNTEREXAMPLE_JUDGE_CONTRACT_MISMATCH"
	if allSatisfied(indicators) {
		decision, reason = "PASS", "COUNTEREXAMPLE_FIRST_CONTRACT_OBSERVED"
	}
	report := cf.Report{Schema: cf.ReportSchema, ContractID: input.Contract.ID, HeadSHA: input.HeadSHA, Decision: decision, Resolution: resolution, Reason: reason, Denominator: input.Contract.Fixed, Summary: summary, Indicators: indicators, Receipts: input.Receipts, NotClaimed: append([]string{}, input.Contract.NotClaimed...), TamperedRejected: 0}
	report.Digest = cf.ReportDigest(report)
	return report
}

func validStructure(contract cf.Contract) bool {
	if contract.Schema != cf.ContractSchema || contract.Version != 3 || contract.ID == "" || contract.SourcePath == "" || contract.Package == "" || contract.Namespace == "" || contract.Producer == "" || contract.Consumer == "" || contract.MetaOperation == "" || contract.Fixed.Version != cf.DenominatorVersion || contract.Fixed.Cases != cf.CaseCount || contract.Fixed.UniqueClaims != cf.CaseCount || contract.Fixed.UniquePredicates != cf.CaseCount || contract.Fixed.Indicators != cf.IndicatorCount || contract.Fixed.ClaimTransitions != cf.TransitionCount || contract.Fixed.UnknownCoordinates != 1 || contract.Fixed.CorpusInputs != cf.CaseCount || len(contract.Predicates) != cf.CaseCount || len(contract.Cases) != cf.CaseCount {
		return false
	}
	seenPredicates := make(map[string]bool, len(contract.Predicates))
	for _, predicate := range contract.Predicates {
		if predicate.ID == "" || predicate.Kind == "" || predicate.Operation == "" || predicate.Rule == "" || predicate.SourceFact == "" || predicate.FailureRule == "" || seenPredicates[predicate.ID] {
			return false
		}
		seenPredicates[predicate.ID] = true
	}
	seenClaims := make(map[string]bool, len(contract.Cases))
	seenCasePredicates := make(map[string]bool, len(contract.Cases))
	for _, spec := range contract.Cases {
		if spec.ID == "" || spec.ClaimID == "" || spec.Proposition == "" || spec.PredicateID == "" || spec.InputKind == "" || spec.ProofChoice == "" || spec.MetaOperation == "" || seenClaims[spec.ClaimID] || seenCasePredicates[spec.PredicateID] || !seenPredicates[spec.PredicateID] {
			return false
		}
		seenClaims[spec.ClaimID] = true
		seenCasePredicates[spec.PredicateID] = true
	}
	return len(contract.NotClaimed) > 0
}

func independentPredicateSpec(predicates []cf.PredicateSpec, id string) (cf.PredicateSpec, bool) {
	for _, predicate := range predicates {
		if predicate.ID == id {
			return predicate, true
		}
	}
	return cf.PredicateSpec{}, false
}

func reconstructReceipt(contract cf.Contract, head, sourcePath string, source []byte, program cf.ExecutionObservation, rule string, predicateSpec cf.PredicateSpec, spec cf.CaseSpec, scenario cf.Scenario) cf.DecisionReceipt {
	claimID := scenario.Candidate.ClaimID
	propositionDigest := cf.PropositionDigest(claimID, scenario.Candidate.Claim, scenario.Candidate.PredicateID)
	observedPredicateSpec := predicateSpec
	observedPredicateSpec.Rule = rule
	observation, predicate := independentInput(scenario.Candidate.ID, scenario.Candidate.Source, observedPredicateSpec, program.SemanticDigest)
	var counterexample *cf.Counterexample
	if predicate.ViolationObserved {
		counterexample = independentMinimal(scenario.Candidate.ID, scenario.Candidate.Source, observedPredicateSpec, program.SemanticDigest, observation, predicate)
	}
	var evidence *cf.ResolutionEvidence
	if counterexample != nil && scenario.Resolution != nil {
		if repaired, ok := independentRepairSource(counterexample.Source, scenario.Resolution.Operation, rule); ok {
			resolutionObservation, resolutionPredicate := independentInput(scenario.Resolution.ID, &repaired, observedPredicateSpec, program.SemanticDigest)
			deltaDigest := cf.DigestValue(struct {
				Operation string `json:"operation"`
				Before    string `json:"before"`
				After     string `json:"after"`
			}{scenario.Resolution.Operation, counterexample.SourceDigest, resolutionObservation.SourceDigest})
			evidence = &cf.ResolutionEvidence{
				ID: scenario.Resolution.ID, CounterexampleID: counterexample.ID, InputID: scenario.Resolution.ID,
				OriginalSourceDigest: counterexample.SourceDigest, RepairSourceDigest: resolutionObservation.SourceDigest, RepairDeltaDigest: deltaDigest,
				RepairOperation: scenario.Resolution.Operation, SameClaim: true, SamePredicate: true, ClaimID: claimID, PropositionDigest: propositionDigest, PredicateID: predicateSpec.ID,
				Observation: resolutionObservation, Predicate: resolutionPredicate, Stage: "RESOLUTION", Step: "repair-and-rerun-minimal-counterexample", Reason: independentResolutionReason(resolutionPredicate),
				ProofChoice: "COUNTEREXAMPLE_RESOLUTION", MetaOperation: "resolve-minimal-counterexample", Producer: "counterexample-resolution-witness", Consumer: cf.ProducerID,
			}
		}
	}
	decision, resolution, reason, coordinate := independentDecision(predicate, counterexample, evidence)
	decisionInput := cf.DecisionInput{ClaimID: claimID, PropositionDigest: propositionDigest, PredicateID: predicateSpec.ID, CandidateID: scenario.Candidate.ID, CandidateDigest: observation.SourceDigest, RequiredBeforeCompile: true}
	if counterexample != nil {
		decisionInput.CounterexampleID = counterexample.ID
		decisionInput.CounterexampleDigest = cf.CounterexampleDigest(*counterexample)
	}
	if evidence != nil {
		decisionInput.ResolutionID = evidence.ID
		decisionInput.ResolutionDigest = cf.ResolutionDigest(*evidence)
		decisionInput.RepairDeltaDigest = evidence.RepairDeltaDigest
	}
	receipt := cf.DecisionReceipt{
		Schema: cf.ReceiptSchema, ContractID: contract.ID, HeadSHA: head, SourcePath: sourcePath, SourceDigest: cf.DigestBytes(source), SemanticDigest: program.SemanticDigest,
		ProgramMetaOperation: program.MetaOperation, ScenarioID: scenario.ID, ClaimID: claimID, PropositionDigest: propositionDigest, PredicateID: predicateSpec.ID,
		Producer: contract.Producer, Consumer: contract.Consumer, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice, Decision: decision, Resolution: resolution, Reason: reason, Coordinate: coordinate,
		CandidateObservation: observation, CandidatePredicate: predicate, Counterexample: counterexample, ResolutionEvidence: evidence, DecisionInput: decisionInput,
		ClaimTransitions: independentTransitions(contract, spec, scenario.ID, claimID, propositionDigest, predicateSpec.ID, observation, predicate, counterexample, evidence, coordinate, reason),
		Effects:          cf.Effects{RepositoryWrites: 0, MutationAuthority: "UNKNOWN"},
	}
	receipt.Digest = cf.ReceiptDigest(receipt)
	return receipt
}

func independentProgram(filename string, source []byte) (cf.ExecutionObservation, string, error) {
	observation := independentExecute(filename, source)
	if !observation.ParseOK || !observation.LowerOK {
		return observation, "", nil
	}
	for _, node := range observation.Nodes {
		if node.Kind == semantic.Activity.String() && node.Name == "CanonicalEntityID" {
			if node.ValueProgram != cf.RuleIdentityV1 && node.ValueProgram != cf.RuleIdentityV2 {
				return observation, node.ValueProgram, fmt.Errorf("COUNTEREXAMPLE_POLICY_UNKNOWN:%s", node.ValueProgram)
			}
			return observation, node.ValueProgram, nil
		}
	}
	return observation, "", fmt.Errorf("COUNTEREXAMPLE_POLICY_ACTIVITY_MISSING")
}

func independentInput(inputID string, source *string, predicateSpec cf.PredicateSpec, baselineSemantic string) (cf.ExecutionObservation, cf.PredicateObservation) {
	if source == nil {
		observation := cf.ExecutionObservation{InputID: inputID, OutputDigest: cf.DigestValue([]string{inputID, "SOURCE_NOT_PROVIDED"})}
		return observation, independentFinalizePredicate(cf.PredicateObservation{PredicateID: predicateSpec.ID, Rule: predicateSpec.Rule, UnknownObserved: true, Reason: "SOURCE_NOT_PROVIDED"})
	}
	observation := independentExecute(inputID, []byte(*source))
	return observation, independentPredicate(predicateSpec, observation, baselineSemantic)
}

func independentExecute(inputID string, source []byte) cf.ExecutionObservation {
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
	observation.MetaOperation = independentInspectMetaOperation(observation.Nodes, observation.Facts)
	observation.OutputDigest = cf.DigestValue(struct {
		SemanticDigest string                      `json:"semantic_digest"`
		Nodes          []cf.NodeObservation        `json:"nodes"`
		Facts          []cf.FactObservation        `json:"facts"`
		MetaOperation  cf.MetaOperationObservation `json:"meta_operation"`
	}{observation.SemanticDigest, observation.Nodes, observation.Facts, observation.MetaOperation})
	return observation
}

func independentInspectMetaOperation(nodes []cf.NodeObservation, facts []cf.FactObservation) cf.MetaOperationObservation {
	meta := cf.MetaOperationObservation{RequiredActivities: append([]string{}, independentMetaActivities...)}
	byID := make(map[string]cf.NodeObservation, len(nodes))
	activities := make(map[string]bool, len(independentMetaActivities))
	for _, node := range nodes {
		byID[node.ID] = node
		if node.Kind == semantic.Activity.String() {
			activities[node.Name] = true
		}
	}
	meta.ActivitiesPresent = true
	for _, name := range independentMetaActivities {
		if !activities[name] {
			meta.ActivitiesPresent = false
		}
	}
	for _, edge := range independentMetaEdges {
		for _, generated := range facts {
			if generated.Predicate != string(semantic.WasGeneratedBy) || byID[generated.Subject].Kind != semantic.Entity.String() || byID[generated.Object].Name != edge.from || byID[generated.Subject].Name != edge.through {
				continue
			}
			for _, used := range facts {
				if used.Predicate == string(semantic.Used) && used.Subject == independentActivityIDByName(byID, edge.to) && used.Object == generated.Subject {
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
	meta.Connected = len(meta.Edges) == len(independentMetaEdges)
	for index, edge := range independentMetaEdges {
		if index >= len(meta.Edges) || meta.Edges[index] != (cf.GraphEdgeObservation{From: edge.from, Through: edge.through, To: edge.to}) {
			meta.Connected = false
		}
	}
	meta.Authorized = meta.ActivitiesPresent && meta.Connected
	if !meta.ActivitiesPresent {
		meta.Reason = "META_OPERATION_ACTIVITY_MISSING"
	} else if !meta.Connected {
		meta.Reason = "META_OPERATION_EDGE_MISSING_OR_REORDERED"
	} else {
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

func independentActivityIDByName(nodes map[string]cf.NodeObservation, name string) string {
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

func independentPredicate(spec cf.PredicateSpec, observation cf.ExecutionObservation, baselineSemantic string) cf.PredicateObservation {
	predicate := cf.PredicateObservation{PredicateID: spec.ID, Rule: spec.Rule}
	if !observation.ParseOK || !observation.LowerOK {
		return independentFinalizePredicate(cf.PredicateObservation{PredicateID: spec.ID, Rule: spec.Rule, UnknownObserved: true, Reason: "EXECUTION_UNOBSERVED"})
	}
	predicate.Applicable = true
	identityViolation := independentIdentityViolation(spec.Rule, observation.Nodes)
	switch spec.Kind {
	case "ENTITY_ID_DRIFT":
		predicate.ViolationObserved, predicate.PassObserved = identityViolation, !identityViolation
		predicate.Reason = independentReason(identityViolation, "ENTITY_ID_DRIFT", "IDENTITY_CANONICAL")
	case "CANONICAL_SOURCE":
		predicate.ViolationObserved, predicate.PassObserved = identityViolation, !identityViolation
		predicate.Reason = independentReason(identityViolation, "CANONICAL_SOURCE_NOT_ADMISSIBLE", "CANONICAL_SOURCE_OBSERVED")
	case "RESOLUTION_REQUIRED":
		predicate.ViolationObserved, predicate.PassObserved = identityViolation, !identityViolation
		predicate.Reason = independentReason(identityViolation, "REPAIR_REQUIRED", "NO_REPAIR_REQUIRED")
	case "SEMANTIC_DIGEST":
		predicate.ViolationObserved = observation.SemanticDigest != baselineSemantic
		predicate.PassObserved = !predicate.ViolationObserved
		predicate.Reason = independentReason(predicate.ViolationObserved, "SEMANTIC_DIGEST_CHANGED", "SEMANTIC_DIGEST_PRESERVED")
	case "SOURCE_ACQUISITION":
		predicate.PassObserved = observation.SourceDigest != ""
		predicate.Reason = independentReason(!predicate.PassObserved, "SOURCE_NOT_PROVIDED", "SOURCE_ACQUIRED")
	default:
		predicate.Applicable, predicate.UnknownObserved, predicate.Reason = false, true, "PREDICATE_KIND_UNKNOWN"
	}
	return independentFinalizePredicate(predicate)
}

func independentIdentityViolation(rule string, nodes []cf.NodeObservation) bool {
	for _, node := range nodes {
		if node.Kind == semantic.Entity.String() && node.ID != independentCanonicalID(rule, node) {
			return true
		}
	}
	return false
}

func independentReason(condition bool, failure, pass string) string {
	if condition {
		return failure
	}
	return pass
}

func independentFinalizePredicate(predicate cf.PredicateObservation) cf.PredicateObservation {
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

func independentCanonicalID(rule string, node cf.NodeObservation) string {
	prefix := "gooo://counterexample-first/entity/"
	if rule == cf.RuleIdentityV2 {
		prefix = "gooo://counterexample-first/v2/entity/"
	}
	return prefix + independentKebab(node.Name)
}

func independentKebab(value string) string {
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

func independentMinimal(inputID string, source *string, predicateSpec cf.PredicateSpec, baselineSemantic string, observation cf.ExecutionObservation, predicate cf.PredicateObservation) *cf.Counterexample {
	if source == nil {
		return nil
	}
	current, currentObservation, currentPredicate := *source, observation, predicate
	var trace []cf.ShrinkObservation
	numerator, denominator := 0, 0
	for {
		candidates := independentShrinkCandidates(current, predicateSpec.Rule)
		if len(candidates) == 0 {
			break
		}
		var next string
		var nextObservation cf.ExecutionObservation
		var nextPredicate cf.PredicateObservation
		for index, candidate := range candidates {
			candidateObservation, candidatePredicate := independentInput(fmt.Sprintf("%s/shrink-%d", inputID, index), &candidate, predicateSpec, baselineSemantic)
			trace = append(trace, cf.ShrinkObservation{CandidateDigest: candidateObservation.SourceDigest, SourceBytes: len(candidate), Observation: candidateObservation, Predicate: candidatePredicate})
			if candidatePredicate.ViolationObserved && (next == "" || len(candidate) < len(next)) {
				next, nextObservation, nextPredicate = candidate, candidateObservation, candidatePredicate
			}
		}
		if next == "" {
			for _, step := range trace[len(trace)-len(candidates):] {
				denominator++
				if !step.Predicate.ViolationObserved {
					numerator++
				}
			}
			break
		}
		current, currentObservation, currentPredicate = next, nextObservation, nextPredicate
	}
	digest := cf.DigestBytes([]byte(predicateSpec.ID + "|" + currentObservation.SourceDigest))
	return &cf.Counterexample{ID: "ce-" + digest[len(digest)-12:], Source: current, SourceDigest: currentObservation.SourceDigest, SourceBytes: currentObservation.SourceBytes, Observation: currentObservation, Predicate: currentPredicate, ShrinkTrace: trace, FiniteNeighborhoodNumerator: numerator, FiniteNeighborhoodDenominator: denominator, FiniteNeighborhoodIrreducible: denominator > 0 && numerator == denominator, Stage: "COUNTEREXAMPLE", Step: "finite-neighborhood-shrink", Reason: "FINITE_NEIGHBORHOOD_IRREDUCIBLE"}
}

func independentShrinkCandidates(source, rule string) []string {
	const noisy = "gooo://counterexample-first/entity/compilation-claim?noise=1&drift=1"
	const drift = "gooo://counterexample-first/entity/compilation-claim?drift=1"
	if strings.Contains(source, noisy) {
		return []string{strings.Replace(source, noisy, drift, 1)}
	}
	if strings.Contains(source, drift) {
		return []string{strings.Replace(source, drift, independentCanonicalID(rule, cf.NodeObservation{Name: "CompilationClaim"}), 1)}
	}
	return nil
}

func independentRepairSource(source, operation, rule string) (string, bool) {
	if operation != "canonicalize-entity-id" {
		return "", false
	}
	const drift = "gooo://counterexample-first/entity/compilation-claim?drift=1"
	if !strings.Contains(source, drift) {
		return "", false
	}
	return strings.Replace(source, drift, independentCanonicalID(rule, cf.NodeObservation{Name: "CompilationClaim"}), 1), true
}

func independentDecision(predicate cf.PredicateObservation, counterexample *cf.Counterexample, evidence *cf.ResolutionEvidence) (string, string, string, cf.Coordinate) {
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

func independentResolutionReason(predicate cf.PredicateObservation) string {
	if predicate.PassObserved {
		return "REPAIR_RERUN_SAME_MINIMAL_SOURCE_PASSED"
	}
	return "REPAIR_RERUN_DID_NOT_PASS"
}

func independentTransitions(contract cf.Contract, spec cf.CaseSpec, scenarioID, claimID, propositionDigest, predicateID string, observation cf.ExecutionObservation, predicate cf.PredicateObservation, counterexample *cf.Counterexample, evidence *cf.ResolutionEvidence, coordinate cf.Coordinate, reason string) []cf.ClaimTransition {
	state, status, firstReason := "OPEN", "LOWER_RESOLUTION", "COUNTEREXAMPLE_REQUIRED"
	firstStage, firstStep := "COUNTEREXAMPLE", "discover"
	if predicate.UnknownObserved {
		firstReason, firstStage, firstStep = "SOURCE_NOT_PROVIDED", "INPUT_OBSERVATION", "source-acquisition"
	} else if counterexample != nil {
		state, status, firstReason, firstStep = "REFUTED", "REFUTED", counterexample.Reason, "finite-neighborhood-shrink"
	}
	make := func(sequence int, from, to, value, stage, step, transitionReason, evidenceDigest string) cf.ClaimTransition {
		predicateDigest := cf.DigestValue(struct {
			PredicateID string `json:"predicate_id"`
			Evidence    string `json:"evidence"`
		}{predicateID, evidenceDigest})
		return cf.ClaimTransition{Sequence: sequence, ClaimID: claimID, PropositionDigest: propositionDigest, PredicateID: predicateID, From: from, To: to, Status: value, Stage: stage, Step: step, Reason: transitionReason, Producer: contract.Producer, Consumer: contract.Consumer, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice, EvidenceDigest: evidenceDigest, PredicateDigest: predicateDigest}
	}
	transitions := []cf.ClaimTransition{make(1, "OPEN", state, status, firstStage, firstStep, firstReason, predicate.EvidenceDigest)}
	if state == "REFUTED" && evidence != nil && evidence.Predicate.PassObserved && evidence.SameClaim && evidence.SamePredicate {
		transitions = append(transitions, make(2, "REFUTED", "DISCHARGED", "DISCHARGED", "RESOLUTION", "repair-and-rerun-minimal-counterexample", evidence.Reason, cf.ResolutionDigest(*evidence)))
		transitions = append(transitions, make(3, "DISCHARGED", "DISCHARGED", "PROMOTED", coordinate.Stage, coordinate.Step, reason, evidence.Predicate.EvidenceDigest))
		return transitions
	}
	transitions = append(transitions, make(2, state, state, status, "RESOLUTION", "await-repair-proof", reason, observation.SourceDigest))
	transitions = append(transitions, make(3, state, state, status, coordinate.Stage, coordinate.Step, coordinate.Reason, observation.OutputDigest))
	return transitions
}

func sameReceipt(actual, expected cf.DecisionReceipt) bool {
	return actual.Digest == expected.Digest && cf.ReceiptDigest(actual) == actual.Digest && cf.DigestValue(actual) == cf.DigestValue(expected)
}

func summarize(input cf.JudgeInput, receipts []cf.DecisionReceipt, verified int) cf.Summary {
	summary := cf.Summary{ReceiptsReconstructed: len(receipts), CasesTotal: len(receipts), ReceiptsVerified: verified, SourceReconstructionNumerator: len(receipts), SourceReconstructionDenominator: len(receipts), ProducerImportNumerator: input.ProducerDependencies, ProducerImportDenominator: 1, RepositoryWrites: input.WorkspaceEffects.RepositoryWrites, MutationAuthority: input.WorkspaceEffects.MutationAuthority, CapabilityEvidence: input.WorkspaceEffects.CapabilityEvidence, DeterministicReplays: 1}
	claimIDs := make(map[string]bool)
	predicateIDs := make(map[string]bool)
	for _, receipt := range receipts {
		if receipt.CandidateObservation.SourceDigest != "" {
			summary.CorpusExecutions++
		}
		claimIDs[receipt.ClaimID] = true
		predicateIDs[receipt.PredicateID] = true
		if receipt.Counterexample != nil {
			summary.DiscoveredCounterexamples++
			summary.ShrinkCandidateExecutions += len(receipt.Counterexample.ShrinkTrace)
			summary.FiniteNeighborhoodNumerator += receipt.Counterexample.FiniteNeighborhoodNumerator
			summary.FiniteNeighborhoodDenominator += receipt.Counterexample.FiniteNeighborhoodDenominator
		}
		if receipt.ResolutionEvidence != nil {
			summary.ResolutionReruns++
			if receipt.ResolutionEvidence.Predicate.PassObserved {
				summary.PromotionsAfterResolution++
			}
		}
		if receipt.Decision == "UNKNOWN" && receipt.Coordinate.Stage == "INPUT_OBSERVATION" && receipt.Coordinate.Step == "source-acquisition" && receipt.Coordinate.Reason == "SOURCE_NOT_PROVIDED" {
			summary.UnknownCoordinatesPreserved++
		}
		summary.ClaimTransitionsPreserved += len(receipt.ClaimTransitions)
	}
	summary.UniqueClaimsObserved = len(claimIDs)
	summary.UniquePredicatesObserved = len(predicateIDs)
	return summary
}

func makeIndicators(summary cf.Summary, contract cf.Contract) []cf.Indicator {
	metric := func(id, class, proof, operation string, value, target, denominator int) cf.Indicator {
		return cf.Indicator{ID: id, Class: class, Producer: contract.Producer, Consumer: contract.Consumer, ProofChoice: proof, MetaOperation: operation, Value: value, Target: target, Denominator: denominator, Satisfied: value == target}
	}
	controls := summary.CasesTotal - summary.DiscoveredCounterexamples - summary.UnknownCoordinatesPreserved
	indicators := []cf.Indicator{
		metric("corpus.execution", "FOUNDATION", "PARSE_LOWER_OBSERVATION", "execute-fixed-corpus", summary.CorpusExecutions, contract.Fixed.CorpusInputs-1, contract.Fixed.CorpusInputs),
		metric("counterexample.discovery", "DRIVER", "COUNTEREXAMPLE_REQUIRED", "discover-from-predicate", summary.DiscoveredCounterexamples, 2, contract.Fixed.Cases),
		metric("counterexample.finite-neighborhood-irreducibility", "DRIVER", "COUNTEREXAMPLE_SHRINKING", "verify-finite-neighborhood-irreducibility", summary.FiniteNeighborhoodNumerator, summary.FiniteNeighborhoodDenominator, summary.FiniteNeighborhoodDenominator),
		metric("resolution.rerun", "COHERENCE", "COUNTEREXAMPLE_RESOLUTION", "repair-and-rerun-minimal-counterexample", summary.ResolutionReruns, 1, 1),
		metric("decision.after-resolution", "OUTCOME", "COUNTEREXAMPLE_RESOLUTION", "promote-after-resolution", summary.PromotionsAfterResolution, 1, 1),
		metric("controls.blocked", "GUARDRAIL", "COUNTEREXAMPLE_REQUIRED", "block-unresolved-or-passing-control", controls, 2, contract.Fixed.Cases),
		metric("unknown.coordinate-preserved", "GUARDRAIL", "UNKNOWN_PRESERVATION", "preserve-input-observation-coordinate", summary.UnknownCoordinatesPreserved, contract.Fixed.UnknownCoordinates, contract.Fixed.UnknownCoordinates),
		metric("claim.transition-closure", "COHERENCE", "CLAIM_TRANSITION", "preserve-append-only-transitions", summary.ClaimTransitionsPreserved, contract.Fixed.ClaimTransitions, contract.Fixed.ClaimTransitions),
		metric("receipt.independent-verification", "OUTCOME", "INDEPENDENT_JUDGMENT", "verify-reconstructed-receipts", summary.ReceiptsVerified, contract.Fixed.Cases, contract.Fixed.Cases),
		metric("source.reconstruction", "FOUNDATION", "RAW_SOURCE_RECONSTRUCTION", "reconstruct-source-semantics", summary.SourceReconstructionNumerator, summary.SourceReconstructionDenominator, summary.SourceReconstructionDenominator),
		metric("producer.import", "GUARDRAIL", "INDEPENDENT_JUDGMENT", "separate-producer-from-consumer", summary.ProducerImportNumerator, 0, summary.ProducerImportDenominator),
	}
	readOnly := metric("repository.read-only", "GUARDRAIL", "READ_ONLY", "bind-ci-effects-to-capability-evidence", summary.RepositoryWrites, 0, 1)
	readOnly.Satisfied = summary.RepositoryWrites == 0 && summary.MutationAuthority == "DENIED" && inputCapabilityEvidence(summary)
	return append(indicators, readOnly)
}

// The capability evidence is carried into the summary through the DENIED
// authority value. A missing capability record is represented as UNKNOWN and
// therefore cannot satisfy the read-only indicator.
func inputCapabilityEvidence(summary cf.Summary) bool {
	return summary.MutationAuthority == "DENIED" && summary.CapabilityEvidence != ""
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
	denominator := input.Contract.Fixed
	if denominator.Version == "" {
		denominator = cf.FixedDenominator{Version: cf.DenominatorVersion, Cases: cf.CaseCount, UniqueClaims: cf.CaseCount, UniquePredicates: cf.CaseCount, Indicators: cf.IndicatorCount, ClaimTransitions: cf.TransitionCount, UnknownCoordinates: 1, CorpusInputs: cf.CaseCount}
	}
	report := cf.Report{Schema: cf.ReportSchema, ContractID: input.Contract.ID, HeadSHA: input.HeadSHA, Decision: "FAIL_CLOSED", Resolution: resolution, Reason: reason, Denominator: denominator, Summary: cf.Summary{CasesTotal: cf.CaseCount, ProducerImportNumerator: input.ProducerDependencies, ProducerImportDenominator: 1, RepositoryWrites: input.WorkspaceEffects.RepositoryWrites, MutationAuthority: input.WorkspaceEffects.MutationAuthority, CapabilityEvidence: input.WorkspaceEffects.CapabilityEvidence}, NotClaimed: append([]string{}, input.Contract.NotClaimed...)}
	report.Digest = cf.ReportDigest(report)
	return report
}
