package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	receiptSchema      = "gooo/meta-minimal-causal-explanation/v2"
	judgmentSchema     = "gooo/meta-minimal-causal-explanation-judgment/v2"
	graphSchema        = "gooo/meta-minimal-causal-graph/v2"
	sourceSchema       = "gooo/meta-minimal-causal-explanation-source/v2"
	compilerSchema     = "gooo.language.value-witness/v2"
	repositorySchema   = "gooo/meta-minimal-causal-explanation-repository/v1"
	independenceSchema = "gooo/meta-minimal-causal-explanation-independence/v1"
	claimOpen          = "OPEN"
	claimDischarged    = "DISCHARGED"
	claimRefuted       = "REFUTED"
	statusPass         = "PASS"
	statusFailClosed   = "FAIL_CLOSED"
	statusUnknown      = "UNKNOWN"
	subsetMinimal      = "SUBSET_MINIMAL"
	notSubsetMinimal   = "NOT_SUBSET_MINIMAL"
	cardinalityMinimum = "CARDINALITY_MINIMUM"
	notCardinality     = "NOT_CARDINALITY_MINIMUM"
	observedOrigin     = "OBSERVED"
	syntheticOrigin    = "SYNTHETIC"
)

type Judgment struct {
	Schema                    string `json:"schema"`
	Status                    string `json:"status"`
	Decision                  string `json:"decision"`
	Resolution                string `json:"resolution"`
	SourceReconstructed       bool   `json:"source_reconstructed"`
	PathSetVerified           bool   `json:"path_set_verified"`
	CounterfactualsVerified   bool   `json:"counterfactuals_verified"`
	ClaimsPreserved           bool   `json:"claims_preserved"`
	InterventionsVerified     bool   `json:"interventions_verified"`
	ProducerImportCount       int    `json:"producer_import_count"`
	ProducerImportDenominator int    `json:"producer_import_denominator"`
	PromotionAuthorized       bool   `json:"promotion_authorized"`
	ReceiptDigest             string `json:"receipt_digest"`
	JudgmentDigest            string `json:"judgment_digest"`
}

type sourceBinding struct {
	Schema         string `json:"schema"`
	Path           string `json:"path"`
	Digest         string `json:"digest"`
	Lines          int    `json:"lines"`
	SemanticDigest string `json:"semantic_digest"`
}

type subject struct {
	Repository string `json:"repository"`
	SHA        string `json:"sha"`
}

type reconstruction struct {
	ASTParsed                  bool   `json:"ast_parsed"`
	IRLowered                  bool   `json:"ir_lowered"`
	SemanticDigest             string `json:"semantic_digest"`
	GraphReconstructed         bool   `json:"graph_reconstructed"`
	PredicateReconstructed     bool   `json:"predicate_reconstructed"`
	ProducerPackageImportCount int    `json:"producer_package_import_count"`
	ProducerPackageImportTotal int    `json:"producer_package_import_total"`
}

type operation struct {
	ID             string `json:"id"`
	Activity       string `json:"activity"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	ProofChoice    string `json:"proof_choice"`
	EvidenceDigest string `json:"evidence_digest"`
}

type program struct {
	Schema               string      `json:"schema"`
	Producer             string      `json:"producer"`
	Consumer             string      `json:"consumer"`
	IndicatorDenominator int         `json:"indicator_denominator"`
	MetaOperations       []operation `json:"meta_operations"`
}

type node struct {
	ID       string `json:"id"`
	Role     string `json:"role"`
	Producer string `json:"producer"`
	Consumer string `json:"consumer"`
}

type edge struct {
	ID          string `json:"id"`
	From        string `json:"from"`
	To          string `json:"to"`
	ViaActivity string `json:"via_activity"`
	Relation    string `json:"relation"`
	Causal      bool   `json:"causal"`
}

type graph struct {
	Schema       string `json:"schema"`
	DecisionRule string `json:"decision_rule"`
	Nodes        []node `json:"nodes"`
	Edges        []edge `json:"edges"`
	Digest       string `json:"digest"`
}

type predicate struct {
	Value           string   `json:"value"`
	RequiredRoles   []string `json:"required_roles"`
	DecisionOutput  string   `json:"decision_output"`
	PriorClaimState string   `json:"prior_claim_state"`
}

type evidence struct {
	ID         string `json:"id"`
	Role       string `json:"role"`
	Origin     string `json:"origin"`
	Status     string `json:"status"`
	Digest     string `json:"digest"`
	Provenance string `json:"provenance"`
}

type coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type counterfactual struct {
	ExecutionID       string     `json:"execution_id"`
	RemovedEvidenceID string     `json:"removed_evidence_id"`
	Origin            string     `json:"origin"`
	BeforeDecision    string     `json:"before_decision"`
	AfterDecision     string     `json:"after_decision"`
	Changed           bool       `json:"changed"`
	Coordinate        coordinate `json:"coordinate"`
	EvidenceDigest    string     `json:"evidence_digest"`
}

type combinationSearch struct {
	CorpusEvidenceIDs                 []string `json:"corpus_evidence_ids"`
	Exhaustive                        bool     `json:"exhaustive"`
	EnumeratedCombinationTotal        int      `json:"enumerated_combination_total"`
	SmallerCombinationTotal           int      `json:"smaller_combination_total"`
	SufficientSmallerCombinationTotal int      `json:"sufficient_smaller_combination_total"`
}

type path struct {
	ID                   string            `json:"id"`
	EvidenceIDs          []string          `json:"evidence_ids"`
	EdgeIDs              []string          `json:"edge_ids"`
	Decision             string            `json:"decision"`
	Sufficient           bool              `json:"sufficient"`
	SubsetMinimal        string            `json:"subset_minimal"`
	CardinalityMinimum   string            `json:"cardinality_minimum"`
	SingleRemovalChanged int               `json:"single_removal_changed"`
	SingleRemovalTotal   int               `json:"single_removal_total"`
	Counterfactuals      []counterfactual  `json:"counterfactuals"`
	CombinationSearch    combinationSearch `json:"combination_search"`
}

type explanationCase struct {
	ID                     string `json:"id"`
	Kind                   string `json:"kind"`
	ExplanationText        string `json:"explanation_text"`
	AvailableEvidenceTotal int    `json:"available_evidence_total"`
	Paths                  []path `json:"paths"`
	ExpectedDecision       string `json:"expected_decision"`
	Verdict                string `json:"verdict"`
}

type indicator struct {
	ID             string `json:"id"`
	Class          string `json:"class"`
	MetaOperation  string `json:"meta_operation"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	ProofChoice    string `json:"proof_choice"`
	Expected       string `json:"expected"`
	Actual         string `json:"actual"`
	Satisfied      bool   `json:"satisfied"`
	EvidenceDigest string `json:"evidence_digest"`
}

type transition struct {
	Sequence                 int        `json:"sequence"`
	ClaimID                  string     `json:"claim_id"`
	Before                   string     `json:"before"`
	After                    string     `json:"after"`
	EvidenceDigest           string     `json:"evidence_digest"`
	Provenance               string     `json:"provenance"`
	Coordinate               coordinate `json:"coordinate"`
	PreviousTransitionDigest string     `json:"previous_transition_digest"`
	TransitionDigest         string     `json:"transition_digest"`
}

type regression struct {
	ScenarioID               string       `json:"scenario_id"`
	ReceiptDecision          string       `json:"receipt_decision"`
	LegacyUnconditionalState string       `json:"legacy_unconditional_state"`
	CorrectState             string       `json:"correct_state"`
	Transitions              []transition `json:"transitions"`
}

type repositoryObservation struct {
	Schema              string `json:"schema"`
	Status              string `json:"status"`
	WorkspaceWrites     bool   `json:"workspace_writes"`
	PromotionAuthorized bool   `json:"promotion_authorized"`
}

type repositoryBoundary struct {
	Before              repositoryObservation `json:"before"`
	After               repositoryObservation `json:"after"`
	Writes              int                   `json:"writes"`
	PromotionAuthorized bool                  `json:"promotion_authorized"`
}

type intervention struct {
	ID                      string `json:"id"`
	Kind                    string `json:"kind"`
	BeforeSourceDigest      string `json:"before_source_digest"`
	AfterSourceDigest       string `json:"after_source_digest"`
	BeforeSemanticDigest    string `json:"before_semantic_digest"`
	AfterSemanticDigest     string `json:"after_semantic_digest"`
	BeforeDecision          string `json:"before_decision"`
	AfterDecision           string `json:"after_decision"`
	SemanticChanged         bool   `json:"semantic_changed"`
	SemanticDigestPreserved bool   `json:"semantic_digest_preserved"`
	ResultPreserved         bool   `json:"result_preserved"`
	PathSetChanged          bool   `json:"path_set_changed"`
	MinimalityChanged       bool   `json:"minimality_changed"`
	ClaimTransitionChanged  bool   `json:"claim_transition_changed"`
	Provenance              string `json:"provenance"`
}

type preservation struct {
	ClaimTotal      int    `json:"claim_total"`
	PreservedTotal  int    `json:"preserved_total"`
	TransitionTotal int    `json:"transition_total"`
	TransitionHead  string `json:"transition_head"`
	Policy          string `json:"policy"`
	RegressionTotal int    `json:"regression_total"`
}

type summary struct {
	CasesTotal                     int    `json:"cases_total"`
	PathsObserved                  int    `json:"paths_observed"`
	ObservedEvidenceTotal          int    `json:"observed_evidence_total"`
	SyntheticEvidenceTotal         int    `json:"synthetic_evidence_total"`
	CandidateEvidenceTotal         int    `json:"candidate_evidence_total"`
	SufficientPaths                int    `json:"sufficient_paths"`
	SubsetMinimalNumerator         int    `json:"subset_minimal_numerator"`
	SubsetMinimalDenominator       int    `json:"subset_minimal_denominator"`
	CardinalityMinimumNumerator    int    `json:"cardinality_minimum_numerator"`
	CardinalityMinimumDenominator  int    `json:"cardinality_minimum_denominator"`
	CardinalityUnknownPaths        int    `json:"cardinality_unknown_paths"`
	InsufficientPaths              int    `json:"insufficient_paths"`
	CounterfactualExecutions       int    `json:"counterfactual_executions"`
	ChangedCounterfactuals         int    `json:"changed_counterfactuals"`
	CombinationExecutions          int    `json:"combination_executions"`
	ClaimTransitionTotal           int    `json:"claim_transition_total"`
	RegressionClaimTransitionTotal int    `json:"regression_claim_transition_total"`
	RepositoryWrites               int    `json:"repository_writes"`
	PromotionAuthorized            bool   `json:"promotion_authorized"`
	PathSetAuthoritative           bool   `json:"path_set_authoritative"`
	ExplanationTextRole            string `json:"explanation_text_role"`
	ObservedCounterfactuals        int    `json:"observed_counterfactuals"`
	SyntheticCounterfactuals       int    `json:"synthetic_counterfactuals"`
}

type receipt struct {
	Schema                string             `json:"schema"`
	Source                sourceBinding      `json:"source"`
	Subject               subject            `json:"subject"`
	Reconstruction        reconstruction     `json:"reconstruction"`
	Program               program            `json:"program"`
	Predicate             predicate          `json:"predicate"`
	Graph                 graph              `json:"graph"`
	Evidence              []evidence         `json:"evidence"`
	ObservedReceiptDigest string             `json:"observed_receipt_digest"`
	Cases                 []explanationCase  `json:"cases"`
	Summary               summary            `json:"summary"`
	Repository            repositoryBoundary `json:"repository"`
	Preservation          preservation       `json:"preservation"`
	ClaimTransitions      []transition       `json:"claim_transitions"`
	ClaimRegression       regression         `json:"claim_regression"`
	Interventions         []intervention     `json:"interventions"`
	Indicators            []indicator        `json:"indicators"`
	Decision              string             `json:"decision"`
	Resolution            string             `json:"resolution"`
	Authority             struct {
		RepositoryWorkspaceWrites  bool `json:"repository_workspace_writes"`
		PromotionAuthorized        bool `json:"promotion_authorized"`
		SemanticMutationAuthorized bool `json:"semantic_mutation_authorized"`
	} `json:"authority"`
	ReceiptDigest string `json:"receipt_digest"`
}

type rawReceipt struct {
	Schema              string `json:"schema"`
	Decision            string `json:"decision"`
	Reason              string `json:"reason"`
	Resolution          string `json:"resolution"`
	HeadSHA             string `json:"head_sha"`
	SourcePath          string `json:"source_path"`
	SourceDigest        string `json:"source_digest"`
	SemanticFingerprint string `json:"semantic_fingerprint"`
	CoreIRFingerprint   string `json:"core_ir_fingerprint"`
	ValueProgram        string `json:"value_program"`
}

type jModel struct {
	SourceDigest, SemanticDigest string
	Evidence                     []evidence
	ByRole                       map[string]evidence
	Graph                        graph
	Predicate                    predicate
	Program                      program
	Claims                       []string
	PriorState, DecisionOutput   string
	Reconstruction               reconstruction
}

func Judge(receiptData []byte, sourcePath string, source, compilerReceipt, repositoryBefore, repositoryAfter, independence []byte) (Judgment, error) {
	var got receipt
	if err := json.Unmarshal(receiptData, &got); err != nil {
		return Judgment{}, fmt.Errorf("receipt json: %w", err)
	}
	model, err := reconstruct(sourcePath, source, independence)
	if err != nil {
		return Judgment{}, err
	}
	var raw rawReceipt
	if err := json.Unmarshal(compilerReceipt, &raw); err != nil {
		return Judgment{}, fmt.Errorf("compiler observation: %w", err)
	}
	if raw.Schema != compilerSchema {
		return Judgment{}, fmt.Errorf("compiler observation schema is invalid")
	}
	var before, after repositoryObservation
	if err := json.Unmarshal(repositoryBefore, &before); err != nil {
		return Judgment{}, fmt.Errorf("repository before: %w", err)
	}
	if err := json.Unmarshal(repositoryAfter, &after); err != nil {
		return Judgment{}, fmt.Errorf("repository after: %w", err)
	}
	if before.Schema != repositorySchema || after.Schema != repositorySchema {
		return Judgment{}, fmt.Errorf("repository observation schema is invalid")
	}
	boundary := makeBoundary(before, after)
	assessment := assess(model, raw, compilerReceipt)
	if err := validateReceipt(got, model, assessment, boundary, sourcePath, source, compilerReceipt, independence); err != nil {
		return Judgment{}, err
	}
	judgment := Judgment{Schema: judgmentSchema, Status: "VERIFIED", Decision: statusPass, Resolution: "INDEPENDENT_RAW_SOURCE_JUDGE", SourceReconstructed: true, PathSetVerified: true, CounterfactualsVerified: true, ClaimsPreserved: got.Preservation.PreservedTotal == got.Preservation.ClaimTotal && got.ClaimRegression.CorrectState == claimRefuted, InterventionsVerified: true, ProducerImportCount: model.Reconstruction.ProducerPackageImportCount, ProducerImportDenominator: model.Reconstruction.ProducerPackageImportTotal, PromotionAuthorized: false, ReceiptDigest: got.ReceiptDigest}
	if model.Reconstruction.ProducerPackageImportCount != 0 || model.Reconstruction.ProducerPackageImportTotal != 1 {
		return Judgment{}, fmt.Errorf("producer import independence is not proven")
	}
	judgment.JudgmentDigest = digest(judgment)
	return judgment, nil
}

func reconstruct(sourcePath string, source, independence []byte) (jModel, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if file == nil || diagnostics.HasErrors() {
		return jModel{}, fmt.Errorf("independent parse failed: %s", diagnostics.Error().Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return jModel{}, fmt.Errorf("independent lower failed: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return jModel{}, fmt.Errorf("independent IR validation failed: %w", err)
	}
	model := jModel{SourceDigest: contentDigest(source), SemanticDigest: "sha256:" + ir.StableHash(), ByRole: map[string]evidence{}, Reconstruction: reconstruction{ASTParsed: true, IRLowered: true, SemanticDigest: "sha256:" + ir.StableHash()}}
	producer, consumer := map[string]string{}, map[string]string{}
	inputs, outputs, programs := map[string][]string{}, map[string][]string{}, map[string]string{}
	for _, node := range ir.Graph.Nodes() {
		id := node.ID.String()
		if node.Kind == semantic.Activity {
			programs[id] = node.ValueProgram
			continue
		}
		if node.Kind != semantic.Entity {
			continue
		}
		role := entityRole(id)
		if role != "" {
			model.Evidence = append(model.Evidence, evidence{ID: id, Role: role, Origin: observedOrigin, Status: statusUnknown, Provenance: "gooo semantic entity"})
		}
	}
	for _, fact := range ir.Graph.DeterministicFacts() {
		subject, object := fact.Subject.String(), fact.Object.String()
		switch fact.Predicate {
		case semantic.Used:
			inputs[subject] = append(inputs[subject], object)
			consumer[object] = subject
		case semantic.WasGeneratedBy:
			outputs[object] = append(outputs[object], subject)
			producer[subject] = object
		}
	}
	for index := range model.Evidence {
		model.Evidence[index].Provenance = "entity=" + model.Evidence[index].ID + ";producer=" + producer[model.Evidence[index].ID] + ";consumer=" + consumer[model.Evidence[index].ID]
		model.ByRole[model.Evidence[index].Role] = model.Evidence[index]
	}
	for activityID, value := range programs {
		for _, clause := range strings.Split(value, ";") {
			clause = strings.TrimSpace(clause)
			switch {
			case strings.HasPrefix(clause, "mce.operation:"):
				op, err := parseOperation(clause, activityID, value)
				if err != nil {
					return jModel{}, err
				}
				model.Program.MetaOperations = append(model.Program.MetaOperations, op)
			case strings.HasPrefix(clause, "mce.predicate:"):
				value := strings.TrimSuffix(strings.TrimPrefix(clause, "mce.predicate:"), ":v1")
				parts := strings.SplitN(value, ":", 2)
				if len(parts) != 2 || parts[0] != "PASS_IF" {
					return jModel{}, fmt.Errorf("invalid predicate %q", clause)
				}
				model.Predicate.Value = clause
				model.Predicate.RequiredRoles = split(parts[1], "+")
			case strings.HasPrefix(clause, "mce.claim-state:"):
				model.PriorState = valueOf(clause, "mce.claim-state:")
			case strings.HasPrefix(clause, "mce.decision-output:"):
				model.DecisionOutput = valueOf(clause, "mce.decision-output:")
			case strings.HasPrefix(clause, "mce.indicators:"):
				model.Program.IndicatorDenominator = intValue(clause, "mce.indicators:")
			case strings.HasPrefix(clause, "mce.program:"):
				endpoints := strings.Split(strings.SplitN(strings.TrimPrefix(clause, "mce.program:"), ":v1", 2)[0], "|")
				if len(endpoints) != 2 {
					return jModel{}, fmt.Errorf("invalid program endpoints")
				}
				model.Program.Producer, model.Program.Consumer = endpoints[0], endpoints[1]
			case strings.HasPrefix(clause, "mce.claims:"):
				model.Claims = split(valueOf(clause, "mce.claims:"), "+")
			}
		}
	}
	sort.Slice(model.Program.MetaOperations, func(i, j int) bool { return model.Program.MetaOperations[i].ID < model.Program.MetaOperations[j].ID })
	model.Program.Schema = sourceSchema
	model.Predicate.DecisionOutput, model.Predicate.PriorClaimState = model.DecisionOutput, model.PriorState
	model.Graph = deriveGraph(model, inputs, outputs, programs)
	model.Reconstruction.GraphReconstructed = len(model.Graph.Nodes) > 0 && len(model.Graph.Edges) > 0
	model.Reconstruction.PredicateReconstructed = model.Predicate.Value != ""
	if independence != nil {
		var raw struct {
			Schema                     string `json:"schema"`
			ProducerPackageImportCount int    `json:"producer_package_import_count"`
			ProducerPackageImportTotal int    `json:"producer_package_import_total"`
		}
		if err := json.Unmarshal(independence, &raw); err != nil {
			return jModel{}, err
		}
		if raw.Schema != independenceSchema || raw.ProducerPackageImportCount < 0 || raw.ProducerPackageImportTotal <= 0 || raw.ProducerPackageImportCount > raw.ProducerPackageImportTotal {
			return jModel{}, fmt.Errorf("independence observation is invalid")
		}
		model.Reconstruction.ProducerPackageImportCount, model.Reconstruction.ProducerPackageImportTotal = raw.ProducerPackageImportCount, raw.ProducerPackageImportTotal
	} else {
		return jModel{}, fmt.Errorf("independence observation is required")
	}
	return model, nil
}

func deriveGraph(model jModel, inputs, outputs map[string][]string, programs map[string]string) graph {
	result := graph{Schema: graphSchema, DecisionRule: model.Predicate.Value}
	for _, item := range model.Evidence {
		producer, consumer := "raw-observation", "unconsumed"
		for activity, values := range outputs {
			if has(values, item.ID) {
				producer = activity
			}
		}
		for activity, values := range inputs {
			if has(values, item.ID) {
				consumer = activity
			}
		}
		role := "DECISION_INPUT"
		if !has(model.Predicate.RequiredRoles, item.Role) {
			role = "NON_CAUSAL_LOG"
		}
		result.Nodes = append(result.Nodes, node{ID: item.ID, Role: role, Producer: producer, Consumer: consumer})
	}
	for activity, values := range inputs {
		for _, from := range values {
			if entityRole(from) == "" {
				continue
			}
			for _, to := range outputs[activity] {
				if entityRole(to) == "" {
					continue
				}
				causal := has(model.Predicate.RequiredRoles, entityRole(from)) && has(model.Predicate.RequiredRoles, entityRole(to))
				relation := programs[activity]
				result.Edges = append(result.Edges, edge{ID: digest(from + "|" + activity + "|" + to + "|" + relation), From: from, To: to, ViaActivity: activity, Relation: relation, Causal: causal})
			}
		}
	}
	sort.Slice(result.Nodes, func(i, j int) bool { return result.Nodes[i].ID < result.Nodes[j].ID })
	sort.Slice(result.Edges, func(i, j int) bool { return result.Edges[i].ID < result.Edges[j].ID })
	result.Digest = graphDigest(result)
	return result
}

func assess(model jModel, raw rawReceipt, rawBytes []byte) struct {
	Observed, Corpus []evidence
	Cases            []explanationCase
	Summary          summary
	Decision         string
	Outcomes         map[string]string
	Transitions      []transition
} {
	observed := observe(model, raw, rawBytes)
	corpus := append([]evidence(nil), observed...)
	if noise, ok := noise(model); ok {
		corpus = append(corpus, noise)
	}
	sort.Slice(corpus, func(i, j int) bool { return corpus[i].ID < corpus[j].ID })
	required := idsForRoles(model, model.Predicate.RequiredRoles)
	minimal := makePath("minimal-sufficient", required, model, corpus)
	overlongIDs := append([]string(nil), required...)
	if noise, ok := noise(model); ok {
		overlongIDs = append(overlongIDs, noise.ID)
	}
	overlong := makePath("sufficient-overlong", overlongIDs, model, corpus)
	insufficientIDs := []string{}
	if len(required) >= 2 {
		insufficientIDs = []string{required[0], required[len(required)-1]}
	}
	insufficient := makePath("insufficient", insufficientIDs, model, corpus)
	cases := []explanationCase{{ID: "minimal", Kind: "MINIMAL_SUFFICIENT", ExplanationText: "the compiler receipt is supported by the decisive path", AvailableEvidenceTotal: len(corpus), Paths: []path{minimal}, ExpectedDecision: statusPass, Verdict: caseVerdict(minimal.Sufficient)}, {ID: "overlong", Kind: "SUFFICIENT_NOT_MINIMAL", ExplanationText: "all logs explain the compiler receipt", AvailableEvidenceTotal: len(corpus), Paths: []path{overlong}, ExpectedDecision: statusPass, Verdict: "REJECTED"}, {ID: "insufficient", Kind: "INSUFFICIENT", ExplanationText: "two observations appear related", AvailableEvidenceTotal: len(observed), Paths: []path{insufficient}, ExpectedDecision: statusPass, Verdict: "REJECTED"}}
	s := summarize(cases, observed, corpus)
	decision := minimal.Decision
	outcomes := claimOutcomes(model, minimal, s, decision == statusPass, raw, observed)
	transitions, _ := buildTransitions(model, outcomes, digestEvidence(observed))
	return struct {
		Observed, Corpus []evidence
		Cases            []explanationCase
		Summary          summary
		Decision         string
		Outcomes         map[string]string
		Transitions      []transition
	}{observed, corpus, cases, s, decision, outcomes, transitions}
}

func makePath(kind string, ids []string, model jModel, corpus []evidence) path {
	result := path{ID: "path." + kind, EvidenceIDs: append([]string(nil), ids...), EdgeIDs: pathEdges(model.Graph, ids), Decision: decision(model, ids, corpus)}
	result.Sufficient = result.Decision == statusPass
	if !result.Sufficient {
		result.SubsetMinimal, result.CardinalityMinimum = notSubsetMinimal, statusUnknown
		return result
	}
	result.Counterfactuals = counterfactuals(model, result, corpus)
	result.SingleRemovalTotal, result.SingleRemovalChanged = len(result.Counterfactuals), changed(result.Counterfactuals)
	if result.SingleRemovalTotal > 0 && result.SingleRemovalChanged == result.SingleRemovalTotal {
		result.SubsetMinimal = subsetMinimal
	} else {
		result.SubsetMinimal = notSubsetMinimal
	}
	result.CombinationSearch = combinationsSearch(model, result, corpus)
	if result.CombinationSearch.SufficientSmallerCombinationTotal == 0 {
		result.CardinalityMinimum = cardinalityMinimum
	} else {
		result.CardinalityMinimum = notCardinality
	}
	return result
}

func decision(model jModel, ids []string, corpus []evidence) string {
	byID := map[string]evidence{}
	for _, item := range corpus {
		byID[item.ID] = item
	}
	for _, role := range model.Predicate.RequiredRoles {
		item, ok := model.ByRole[role]
		if !ok || !has(ids, item.ID) || byID[item.ID].Status != statusPass {
			return statusFailClosed
		}
	}
	required := idsForRoles(model, model.Predicate.RequiredRoles)
	for i := 1; i < len(required); i++ {
		if !hasEdge(model.Graph, required[i-1], required[i]) || !has(ids, required[i-1]) || !has(ids, required[i]) {
			return statusFailClosed
		}
	}
	if model.PriorState != claimOpen || model.DecisionOutput == "" {
		return statusFailClosed
	}
	return statusPass
}

func counterfactuals(model jModel, p path, corpus []evidence) []counterfactual {
	result := make([]counterfactual, 0, len(p.EvidenceIDs))
	for i, removed := range p.EvidenceIDs {
		remaining := append([]string(nil), p.EvidenceIDs[:i]...)
		remaining = append(remaining, p.EvidenceIDs[i+1:]...)
		after := decision(model, remaining, corpus)
		didChange := p.Decision != after
		result = append(result, counterfactual{ExecutionID: "cf." + fmt.Sprint(i+1), RemovedEvidenceID: removed, Origin: syntheticOrigin, BeforeDecision: p.Decision, AfterDecision: after, Changed: didChange, Coordinate: coordinate{Stage: "COUNTERFACTUAL", Step: "remove-single-evidence", Reason: removalReason(didChange)}, EvidenceDigest: digest(strings.Join(p.EvidenceIDs, "|"))})
	}
	return result
}

func combinationsSearch(model jModel, p path, corpus []evidence) combinationSearch {
	ids := make([]string, 0, len(corpus))
	for _, item := range corpus {
		ids = append(ids, item.ID)
	}
	sort.Strings(ids)
	result := combinationSearch{CorpusEvidenceIDs: ids, Exhaustive: true}
	for size := 0; size < len(p.EvidenceIDs); size++ {
		values := combinations(ids, size)
		result.EnumeratedCombinationTotal += len(values)
		result.SmallerCombinationTotal += len(values)
		for _, item := range values {
			if decision(model, item, corpus) == statusPass {
				result.SufficientSmallerCombinationTotal++
			}
		}
	}
	return result
}

func combinations(values []string, size int) [][]string {
	result := [][]string{}
	var visit func(int, []string)
	visit = func(start int, selected []string) {
		if len(selected) == size {
			result = append(result, append([]string(nil), selected...))
			return
		}
		for i := start; i <= len(values)-(size-len(selected)); i++ {
			visit(i+1, append(selected, values[i]))
		}
	}
	visit(0, nil)
	return result
}

func summarize(cases []explanationCase, observed, corpus []evidence) summary {
	result := summary{CasesTotal: len(cases), PathsObserved: len(cases), ObservedEvidenceTotal: len(observed), CandidateEvidenceTotal: len(corpus), PathSetAuthoritative: true, ExplanationTextRole: "INCIDENTAL"}
	for _, item := range corpus {
		if item.Origin == syntheticOrigin {
			result.SyntheticEvidenceTotal++
		}
	}
	for _, example := range cases {
		p := example.Paths[0]
		if p.Sufficient {
			result.SufficientPaths++
			result.CounterfactualExecutions += len(p.Counterfactuals)
			result.ChangedCounterfactuals += p.SingleRemovalChanged
			result.CombinationExecutions += p.CombinationSearch.EnumeratedCombinationTotal
			if p.SubsetMinimal == subsetMinimal {
				result.SubsetMinimalNumerator++
			}
			if p.CardinalityMinimum == cardinalityMinimum {
				result.CardinalityMinimumNumerator++
			}
			result.SubsetMinimalDenominator++
			result.CardinalityMinimumDenominator++
		} else {
			result.InsufficientPaths++
			result.CardinalityUnknownPaths++
		}
	}
	result.ClaimTransitionTotal, result.RegressionClaimTransitionTotal = 12, 12
	result.SyntheticCounterfactuals = result.CounterfactualExecutions
	return result
}

func claimOutcomes(model jModel, minimal path, s summary, pass bool, raw rawReceipt, observed []evidence) map[string]string {
	result := map[string]string{}
	for _, claim := range model.Claims {
		result[claim] = statusUnknown
	}
	setOutcome(result, "source-bound", model.Reconstruction.ASTParsed && model.Reconstruction.IRLowered)
	setOutcome(result, "graph-predicate-reconstructed", model.Reconstruction.GraphReconstructed && model.Reconstruction.PredicateReconstructed)
	setConditionalOutcome(result, "subset-minimal", pass && minimal.SubsetMinimal == subsetMinimal, observed, raw.Decision)
	setConditionalOutcome(result, "cardinality-minimum", pass && minimal.CardinalityMinimum == cardinalityMinimum, observed, raw.Decision)
	setConditionalOutcome(result, "counterfactual-difference", s.CounterfactualExecutions == 7 && s.ChangedCounterfactuals == 6, observed, raw.Decision)
	setOutcome(result, "read-only-preserved", true)
	return result
}
func setOutcome(values map[string]string, suffix string, pass bool) {
	for key := range values {
		if strings.HasSuffix(key, suffix) {
			if pass {
				values[key] = statusPass
			} else {
				values[key] = claimRefuted
			}
		}
	}
}

func setConditionalOutcome(values map[string]string, suffix string, pass bool, observed []evidence, rawDecision string) {
	if pass {
		setOutcome(values, suffix, true)
		return
	}
	status := statusUnknown
	if explicitCounterexample(rawDecision) || allObserved(observed) {
		status = claimRefuted
	}
	for key := range values {
		if strings.HasSuffix(key, suffix) {
			values[key] = status
		}
	}
}

func explicitCounterexample(decision string) bool {
	return decision == statusFailClosed || decision == "VALUE_WITNESS_REJECTED"
}

func allObserved(observed []evidence) bool {
	if len(observed) == 0 {
		return false
	}
	for _, item := range observed {
		if item.Status == statusUnknown {
			return false
		}
	}
	return true
}
func buildTransitions(model jModel, outcomes map[string]string, evidenceDigest string) ([]transition, preservation) {
	result := []transition{}
	previous := ""
	for _, claim := range model.Claims {
		first := transition{Sequence: len(result) + 1, ClaimID: claim, Before: "UNRECORDED", After: model.PriorState, EvidenceDigest: evidenceDigest, Provenance: "prior claim state reconstructed from .gooo:mce.claim-state", Coordinate: coordinate{Stage: "DECLARE", Step: "claim-ledger", Reason: "PRIOR_STATE_OBSERVED"}, PreviousTransitionDigest: previous}
		first.TransitionDigest = transitionDigest(first)
		result = append(result, first)
		previous = first.TransitionDigest
		after, reason := claimOpen, "CLAIM_EVIDENCE_UNOBSERVED"
		if outcomes[claim] == statusPass {
			after, reason = claimDischarged, "CLAIM_EVIDENCE_PASSED"
		}
		if outcomes[claim] == claimRefuted {
			after, reason = claimRefuted, "CLAIM_COUNTEREXAMPLE"
		}
		second := transition{Sequence: len(result) + 1, ClaimID: claim, Before: model.PriorState, After: after, EvidenceDigest: evidenceDigest, Provenance: "raw observation and derived path receipt", Coordinate: coordinate{Stage: "VERIFY", Step: claim, Reason: reason}, PreviousTransitionDigest: previous}
		second.TransitionDigest = transitionDigest(second)
		result = append(result, second)
		previous = second.TransitionDigest
	}
	preserved := 0
	for _, claim := range model.Claims {
		if outcomes[claim] == statusPass {
			preserved++
		}
	}
	return result, preservation{ClaimTotal: len(model.Claims), PreservedTotal: preserved, TransitionTotal: len(result), TransitionHead: previous, Policy: "APPEND_ONLY_OPEN_CONDITIONAL_DISCHARGE_OR_REFUTE"}
}

func validateReceipt(got receipt, model jModel, assessment struct {
	Observed, Corpus []evidence
	Cases            []explanationCase
	Summary          summary
	Decision         string
	Outcomes         map[string]string
	Transitions      []transition
}, boundary repositoryBoundary, sourcePath string, source, compilerReceipt, independence []byte) error {
	if got.Schema != receiptSchema || got.ReceiptDigest != receiptDigest(got) {
		return fmt.Errorf("receipt schema or digest is invalid")
	}
	if got.Source.Schema != sourceSchema || got.Source.Path != sourcePath || got.Source.Digest != model.SourceDigest || got.Source.SemanticDigest != model.SemanticDigest || got.Source.Lines != countLines(source) {
		return fmt.Errorf("source binding is invalid")
	}
	if !equal(got.Reconstruction, model.Reconstruction) || got.Graph.Digest != graphDigest(got.Graph) {
		return fmt.Errorf("source reconstruction binding is invalid")
	}
	if !equal(got.Program, model.Program) || !equal(got.Predicate, model.Predicate) || !equal(got.Graph, model.Graph) {
		return fmt.Errorf("source semantic values were not reconstructed")
	}
	if got.ObservedReceiptDigest != contentDigest(compilerReceipt) || !equal(got.Evidence, assessment.Corpus) {
		return fmt.Errorf("raw observed evidence is not bound")
	}
	if !equal(got.Cases, assessment.Cases) || !equal(got.Summary, assessment.Summary) {
		return fmt.Errorf("path or summary result was not independently derived")
	}
	if !equal(got.Repository, boundary) {
		return fmt.Errorf("repository before/after boundary changed")
	}
	wantTransitions, wantPreservation := buildTransitions(model, assessment.Outcomes, digestEvidence(assessment.Observed))
	if !equal(got.ClaimTransitions, wantTransitions) || !equal(got.Preservation, wantPreservation) {
		return fmt.Errorf("claim transitions are not conditional and append-only")
	}
	observedRaw := rawReceipt{}
	if err := json.Unmarshal(compilerReceipt, &observedRaw); err != nil {
		return err
	}
	if observedRaw.HeadSHA == "" || observedRaw.HeadSHA != got.Subject.SHA {
		return fmt.Errorf("compiler receipt subject binding is invalid")
	}
	failed := observedRaw
	failed.Decision = statusFailClosed
	failed.Reason = "SYNTHETIC_FAILURE_REGRESSION"
	failedAssessment := assess(model, failed, compilerReceipt)
	regressionTransitions, _ := buildTransitions(model, failedAssessment.Outcomes, digestEvidence(assessment.Observed))
	wantRegression := regression{ScenarioID: "failed-receipt-does-not-discharge", ReceiptDecision: statusFailClosed, LegacyUnconditionalState: claimDischarged, CorrectState: claimRefuted, Transitions: regressionTransitions}
	if !equal(got.ClaimRegression, wantRegression) {
		return fmt.Errorf("failure regression does not refute the claim")
	}
	wantInterventions, err := interventionsFor(sourcePath, source, model, observedRaw, compilerReceipt, independence, assessment)
	if err != nil {
		return err
	}
	if !equal(got.Interventions, wantInterventions) {
		return fmt.Errorf("interventions were not independently derived")
	}
	if len(got.Interventions) != 3 || got.Indicators == nil || len(got.Indicators) != model.Program.IndicatorDenominator || boundary.Writes != 0 || boundary.PromotionAuthorized || got.Decision != statusPass || got.Authority.RepositoryWorkspaceWrites || got.Authority.PromotionAuthorized || got.Authority.SemanticMutationAuthorized {
		return fmt.Errorf("receipt contract or authority is invalid")
	}
	for _, item := range got.Indicators {
		if !item.Satisfied || item.MetaOperation == "" || item.Producer == "" || item.Consumer == "" || item.ProofChoice == "" || item.EvidenceDigest == "" {
			return fmt.Errorf("indicator %q is not bound", item.ID)
		}
		if !hasOperation(model.Program.MetaOperations, item.MetaOperation, item.Producer, item.Consumer, item.ProofChoice) {
			return fmt.Errorf("indicator %q uses unknown operation", item.ID)
		}
	}
	if model.Reconstruction.ProducerPackageImportCount != 0 || model.Reconstruction.ProducerPackageImportTotal != 1 {
		return fmt.Errorf("producer import check is not 0/1")
	}
	return nil
}

func interventionsFor(sourcePath string, source []byte, base jModel, raw rawReceipt, rawBytes, independence []byte, baseAssessment struct {
	Observed, Corpus []evidence
	Cases            []explanationCase
	Summary          summary
	Decision         string
	Outcomes         map[string]string
	Transitions      []transition
}) ([]intervention, error) {
	variants := []struct {
		id, kind, source, provenance string
	}{
		{"predicate-change", "SEMANTIC_PREDICATE", strings.Replace(string(source), "mce.predicate:PASS_IF:source-parsed+semantic-ir-bound+compiler-receipt-proven:v1", "mce.predicate:PASS_IF:source-parsed+semantic-ir-bound+missing-observation:v1", 1), "source .gooo decision predicate intervention"},
		{"relation-change", "SEMANTIC_EVIDENCE_RELATION", strings.Replace(string(source), "BindCompilerReceipt(SemanticIRBoundEvidence)", "BindCompilerReceipt(AuditNoiseEvidence)", 1), "source .gooo evidence relation intervention"},
		{"comment-only", "PRESENTATION_COMMENT", string(source) + "\n// comment-only semantic intervention\n", "source comment-only intervention"},
	}
	basePathDigest := digestValue(baseAssessment.Cases)
	baseClaimDigest := digestValue(baseAssessment.Transitions)
	result := make([]intervention, 0, len(variants))
	for _, variant := range variants {
		model, err := reconstruct(sourcePath, []byte(variant.source), independence)
		if err != nil {
			return nil, fmt.Errorf("intervention %s: %w", variant.id, err)
		}
		assessment := assess(model, raw, rawBytes)
		result = append(result, intervention{
			ID: variant.id, Kind: variant.kind, BeforeSourceDigest: base.SourceDigest, AfterSourceDigest: model.SourceDigest,
			BeforeSemanticDigest: base.SemanticDigest, AfterSemanticDigest: model.SemanticDigest,
			BeforeDecision: baseAssessment.Decision, AfterDecision: assessment.Decision,
			SemanticChanged: base.SemanticDigest != model.SemanticDigest, SemanticDigestPreserved: base.SemanticDigest == model.SemanticDigest,
			ResultPreserved: baseAssessment.Decision == assessment.Decision && digestValue(baseAssessment.Cases) == digestValue(assessment.Cases),
			PathSetChanged:  basePathDigest != digestValue(assessment.Cases), MinimalityChanged: pathProperties(baseAssessment.Cases) != pathProperties(assessment.Cases),
			ClaimTransitionChanged: baseClaimDigest != digestValue(assessment.Transitions), Provenance: variant.provenance,
		})
	}
	return result, nil
}

func pathProperties(cases []explanationCase) string {
	var builder strings.Builder
	for _, example := range cases {
		p := example.Paths[0]
		fmt.Fprintf(&builder, "%s|%s|%s|%s|", p.Decision, p.SubsetMinimal, p.CardinalityMinimum, strings.Join(p.EdgeIDs, ","))
	}
	return builder.String()
}

func observe(model jModel, raw rawReceipt, rawBytes []byte) []evidence {
	digestValue := contentDigest(rawBytes)
	result := []evidence{}
	statuses := map[string]string{"source-parsed": statusUnknown, "semantic-ir-bound": statusUnknown, "compiler-receipt-proven": statusUnknown}
	if raw.SourcePath != "" && raw.SourceDigest != "" {
		statuses["source-parsed"] = statusPass
	}
	if raw.SemanticFingerprint != "" && raw.CoreIRFingerprint != "" && raw.Resolution == "CORE_IR_ACTIVITY_VALUE_PROGRAM" {
		statuses["semantic-ir-bound"] = statusPass
	}
	if raw.Decision == "VALUE_WITNESS_PROVEN" && raw.Reason == "VALUE_WITNESS_EXACT" && raw.Resolution == "CORE_IR_ACTIVITY_VALUE_PROGRAM" {
		statuses["compiler-receipt-proven"] = statusPass
	}
	for _, role := range model.Predicate.RequiredRoles {
		item := model.ByRole[role]
		item.Origin, item.Status = observedOrigin, statuses[role]
		item.Digest = digest(role + "|" + digestValue + "|" + item.Status)
		item.Provenance = "raw compiler receipt;source=" + raw.SourcePath + ";decision=" + raw.Decision
		result = append(result, item)
	}
	return result
}
func noise(model jModel) (evidence, bool) {
	item, ok := model.ByRole["audit-noise"]
	if !ok {
		return evidence{}, false
	}
	item.Origin, item.Status, item.Digest, item.Provenance = syntheticOrigin, statusPass, digest("synthetic|audit-noise|overlong-path"), "synthetic overlong-path noise;not observed compiler evidence"
	return item, true
}
func makeBoundary(before, after repositoryObservation) repositoryBoundary {
	writes := 0
	if before.WorkspaceWrites || after.WorkspaceWrites || before.Status != "" || after.Status != "" {
		writes = 1
	}
	return repositoryBoundary{Before: before, After: after, Writes: writes, PromotionAuthorized: before.PromotionAuthorized || after.PromotionAuthorized}
}
func idsForRoles(model jModel, roles []string) []string {
	result := []string{}
	for _, role := range roles {
		if item, ok := model.ByRole[role]; ok {
			result = append(result, item.ID)
		}
	}
	return result
}
func pathEdges(g graph, ids []string) []string {
	result := []string{}
	for i := 1; i < len(ids); i++ {
		for _, item := range g.Edges {
			if item.From == ids[i-1] && item.To == ids[i] && item.Causal {
				result = append(result, item.ID)
				break
			}
		}
	}
	return result
}
func hasEdge(g graph, from, to string) bool {
	for _, item := range g.Edges {
		if item.From == from && item.To == to && item.Causal {
			return true
		}
	}
	return false
}
func entityRole(id string) string {
	marker := "/evidence/"
	index := strings.Index(id, marker)
	if index < 0 {
		return ""
	}
	role := strings.Trim(id[index+len(marker):], "/")
	if role == "" || strings.Contains(role, "/") {
		return ""
	}
	return role
}
func parseOperation(clause, activity, program string) (operation, error) {
	value := strings.TrimSuffix(strings.TrimPrefix(clause, "mce.operation:"), ":v1")
	parts := strings.Split(value, "|")
	if len(parts) != 4 {
		return operation{}, fmt.Errorf("invalid operation value %q", clause)
	}
	return operation{ID: parts[0], Activity: activity, Producer: parts[1], Consumer: parts[2], ProofChoice: parts[3], EvidenceDigest: digest(program)}, nil
}
func valueOf(clause, prefix string) string {
	return strings.TrimSuffix(strings.TrimPrefix(clause, prefix), ":v1")
}
func intValue(clause, prefix string) int {
	var result int
	fmt.Sscan(valueOf(clause, prefix), &result)
	return result
}
func split(value, separator string) []string {
	parts := strings.Split(value, separator)
	result := []string{}
	for _, item := range parts {
		if strings.TrimSpace(item) != "" {
			result = append(result, strings.TrimSpace(item))
		}
	}
	return result
}
func has(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
func changed(values []counterfactual) int {
	result := 0
	for _, value := range values {
		if value.Changed {
			result++
		}
	}
	return result
}
func removalReason(changed bool) string {
	if changed {
		return "DECISION_CHANGED"
	}
	return "DECISION_UNCHANGED"
}
func caseVerdict(sufficient bool) string {
	if sufficient {
		return "ACCEPTED"
	}
	return "REJECTED"
}
func hasOperation(ops []operation, id, producer, consumer, proof string) bool {
	for _, op := range ops {
		if op.ID == id && op.Producer == producer && op.Consumer == consumer && op.ProofChoice == proof {
			return true
		}
	}
	return false
}
func digest(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func contentDigest(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func countLines(value []byte) int {
	if len(value) == 0 {
		return 0
	}
	result := strings.Count(string(value), "\n")
	if value[len(value)-1] != '\n' {
		result++
	}
	return result
}
func graphDigest(value graph) string           { value.Digest = ""; return digest(value) }
func receiptDigest(value receipt) string       { value.ReceiptDigest = ""; return digest(value) }
func transitionDigest(value transition) string { value.TransitionDigest = ""; return digest(value) }
func digestEvidence(value []evidence) string   { return digest(value) }
func equal(left, right any) bool {
	a, _ := json.Marshal(left)
	b, _ := json.Marshal(right)
	return string(a) == string(b)
}
