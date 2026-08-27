package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/query"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	schema        = "gooo/reflective-query-sandbox-observation/v2"
	receiptSchema = "gooo/reflective-query-sandbox-receipt/v2"
	metricID      = "gooo.metric.language.reflective-query-sandbox.v2"
	sourcePath    = "examples/reflective-query-sandbox/main.gooo"
	producerName  = "reflective-query-sandbox.producer"
	consumerName  = "reflective-query-sandbox.independent-verifier"
)

type bucket struct {
	Name  string `json:"name"`
	Total int    `json:"total"`
}

type contract struct {
	Schema              string   `json:"schema"`
	MetricID            string   `json:"metric_id"`
	GoVersion           string   `json:"go_version"`
	Denominator         int      `json:"denominator"`
	Classes             []bucket `json:"classes"`
	Proofs              []bucket `json:"proofs"`
	SourceNodes         int      `json:"source_nodes"`
	SourceFacts         int      `json:"source_facts"`
	ClaimCount          int      `json:"claim_count"`
	AttemptCount        int      `json:"attempt_count"`
	ReflectiveQueries   int      `json:"reflective_queries"`
	SafeQueries         int      `json:"safe_queries"`
	DeniedMutations     int      `json:"denied_mutations"`
	UnknownTargets      int      `json:"unknown_targets"`
	RefutedAttempts     int      `json:"refuted_attempts"`
	TransitionCount     int      `json:"transition_count"`
	SatisfiedIndicators int      `json:"satisfied_indicators"`
}

type effects struct {
	RepositoryBefore   []string `json:"repository_before"`
	RepositoryAfter    []string `json:"repository_after"`
	RepositoryWriteSet []string `json:"repository_write_set"`
	RepositoryWrites   int      `json:"repository_writes"`
	MutationAuthority  bool     `json:"mutation_authority"`
	MutationAPI        string   `json:"mutation_api"`
	MutationOutcome    string   `json:"mutation_outcome"`
	MutationError      string   `json:"mutation_error,omitempty"`
}

type snapshot struct {
	Path           string `json:"path"`
	SourceDigest   string `json:"source_digest"`
	SemanticDigest string `json:"semantic_digest"`
	GraphDigest    string `json:"graph_digest"`
	NodeCount      int    `json:"node_count"`
	FactCount      int    `json:"fact_count"`
	GoooLines      int    `json:"gooo_lines"`
}

type attempt struct {
	ID                   string   `json:"id"`
	Class                string   `json:"class"`
	Operation            string   `json:"operation"`
	Root                 string   `json:"root"`
	Relation             string   `json:"relation"`
	Target               string   `json:"target"`
	MetaOperation        string   `json:"meta_operation"`
	Producer             string   `json:"producer"`
	Consumer             string   `json:"consumer"`
	ProofChoice          string   `json:"proof_choice"`
	Stage                string   `json:"stage"`
	Step                 string   `json:"step"`
	Decision             string   `json:"decision"`
	Resolution           string   `json:"resolution"`
	Reason               string   `json:"reason"`
	MatchedFacts         int      `json:"matched_facts"`
	EvidenceClaimIDs     []string `json:"evidence_claim_ids"`
	API                  string   `json:"api,omitempty"`
	APIOutcome           string   `json:"api_outcome,omitempty"`
	APIError             string   `json:"api_error,omitempty"`
	SemanticDigestBefore string   `json:"semantic_digest_before"`
	SemanticDigestAfter  string   `json:"semantic_digest_after"`
	GraphDigestBefore    string   `json:"graph_digest_before"`
	GraphDigestAfter     string   `json:"graph_digest_after"`
}

type claimTransition struct {
	Sequence        int    `json:"sequence"`
	ClaimID         string `json:"claim_id"`
	Class           string `json:"class"`
	ProofChoice     string `json:"proof_choice"`
	MetaOperation   string `json:"meta_operation"`
	PriorState      string `json:"prior_state"`
	EvidenceAttempt string `json:"evidence_attempt"`
	Producer        string `json:"producer"`
	Consumer        string `json:"consumer"`
	Stage           string `json:"stage"`
	Step            string `json:"step"`
	Reason          string `json:"reason"`
	From            string `json:"from"`
	To              string `json:"to"`
	PreviousDigest  string `json:"previous_digest"`
	Digest          string `json:"digest"`
}

type observation struct {
	Schema     string            `json:"schema"`
	SubjectSHA string            `json:"subject_sha"`
	Contract   contract          `json:"contract"`
	Source     snapshot          `json:"source"`
	Attempts   []attempt         `json:"attempts"`
	Claims     []claimTransition `json:"claims"`
	Effects    effects           `json:"effects"`
	Producer   string            `json:"producer"`
	Digest     string            `json:"digest"`
}

type score struct {
	Name      string `json:"name"`
	Satisfied int    `json:"satisfied"`
	Total     int    `json:"total"`
}

type coordinates struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type receipt struct {
	Schema               string            `json:"schema"`
	SubjectSHA           string            `json:"subject_sha"`
	MetricID             string            `json:"metric_id"`
	Decision             string            `json:"decision"`
	Resolution           string            `json:"resolution"`
	SubjectResolution    string            `json:"subject_resolution"`
	Reason               string            `json:"reason"`
	Producer             string            `json:"producer"`
	Consumer             string            `json:"consumer"`
	Contract             contract          `json:"contract"`
	Source               snapshot          `json:"source"`
	Attempts             []attempt         `json:"attempts"`
	Claims               []claimTransition `json:"claims"`
	Coordinates          coordinates       `json:"coordinates"`
	Classes              []score           `json:"classes"`
	Proofs               []score           `json:"proofs"`
	Effects              effects           `json:"effects"`
	SourceReconstruction coordinates       `json:"source_reconstruction"`
	ProducerImports      coordinates       `json:"producer_imports"`
	PromotionCreditBPS   int               `json:"promotion_credit_bps"`
	RepositoryWrites     int               `json:"repository_writes"`
	MutationAuthority    bool              `json:"mutation_authority"`
	NotClaimed           []string          `json:"not_claimed"`
	Digest               string            `json:"digest"`
}

type claimSpec struct {
	ID, Class, ProofChoice, MetaOperation, EvidenceAttempt, PriorState string
	NodeID                                                             semantic.ID
}

type operationSpec struct {
	ID      semantic.ID
	Program string
}

type sourceModel struct {
	Claims         []claimSpec
	Operations     []operationSpec
	QuerySubject   semantic.ID
	MutationTarget semantic.ID
	ReceiptTarget  semantic.ID
	MetricTarget   semantic.ID
	UnknownTarget  query.ID
}

func main() {
	input := flag.String("input", "", "producer observation")
	source := flag.String("source", "", "Gooo source")
	subject := flag.String("subject-sha", "", "exact subject commit")
	output := flag.String("output", "", "independent receipt")
	flag.Parse()
	if *input == "" || *source == "" || *subject == "" || *output == "" {
		fail("usage: consumer -input FILE -source FILE -subject-sha SHA -output FILE")
	}
	data, err := os.ReadFile(*input)
	if err != nil {
		fail("read observation: %v", err)
	}
	var value observation
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		fail("decode observation: %v", err)
	}
	satisfied, total, err := validateObservation(value, *source, *subject)
	if err != nil {
		fail("independent reconstruction: %v", err)
	}
	result := buildReceipt(value, satisfied, total)
	result.Digest = receiptDigest(result)
	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fail("encode receipt: %v", err)
	}
	if err := os.WriteFile(*output, append(encoded, '\n'), 0o644); err != nil {
		fail("write receipt: %v", err)
	}
	fmt.Printf("consumer verdict: %s %d/%d queries=%d/%d source=%d/%d imports=%d/%d writes=%d mutation_authority=%t\n", result.Decision, result.Coordinates.Satisfied, result.Coordinates.Total, value.Contract.SafeQueries, value.Contract.ReflectiveQueries, result.SourceReconstruction.Satisfied, result.SourceReconstruction.Total, result.ProducerImports.Satisfied, result.ProducerImports.Total, result.Effects.RepositoryWrites, result.Effects.MutationAuthority)
}

func validateObservation(value observation, sourcePath, subject string) (int, int, error) {
	if value.Schema != schema || value.SubjectSHA != subject || value.Producer != producerName {
		return 0, 0, errors.New("observation identity is not exact")
	}
	if value.Digest != observationDigest(value) {
		return 0, 0, errors.New("observation digest mismatch")
	}
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return 0, 0, fmt.Errorf("read source: %w", err)
	}
	file, diagnostics := syntax.ParseFile(sourcePath, string(data))
	if diagnostics.HasErrors() {
		return 0, 0, diagnostics.Error()
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return 0, 0, fmt.Errorf("lower source: %w", err)
	}
	graph, err := query.FromSemanticIR(ir)
	if err != nil {
		return 0, 0, fmt.Errorf("project query view: %w", err)
	}
	model, err := deriveSourceModel(ir)
	if err != nil {
		return 0, 0, fmt.Errorf("derive source model: %w", err)
	}
	semanticDigest := ir.StableHash()
	wantSource := snapshot{Path: sourcePath, SourceDigest: semantic.StableHash(data), SemanticDigest: semanticDigest, GraphDigest: graph.StableHash(), NodeCount: len(graph.Nodes()), FactCount: len(graph.AllFacts()), GoooLines: countLines(data)}
	if !reflect.DeepEqual(value.Source, wantSource) {
		return 0, 0, fmt.Errorf("source reconstruction differs: got=%+v want=%+v", value.Source, wantSource)
	}
	if !strings.HasPrefix(runtime.Version(), "go1.27") || value.Contract.GoVersion != runtime.Version() {
		return 0, 0, fmt.Errorf("runtime=%s contract=%s", runtime.Version(), value.Contract.GoVersion)
	}
	if value.Effects.RepositoryBefore == nil || value.Effects.RepositoryAfter == nil || !reflect.DeepEqual(value.Effects.RepositoryBefore, value.Effects.RepositoryAfter) || value.Effects.RepositoryWrites != len(value.Effects.RepositoryWriteSet) || len(value.Effects.RepositoryWriteSet) != 0 || value.Effects.MutationAuthority || value.Effects.MutationOutcome != "REJECTED" || value.Effects.MutationAPI == "" {
		return 0, 0, errors.New("observed effects do not prove no repository write and no mutation authority")
	}
	wantAttempts, authority, api, outcome, apiError, err := reconstructAttempts(ir, graph, model, semanticDigest)
	if err != nil {
		return 0, 0, err
	}
	if authority || api != value.Effects.MutationAPI || outcome != value.Effects.MutationOutcome || apiError != value.Effects.MutationError {
		return 0, 0, errors.New("mutation boundary was not independently reconstructed")
	}
	if !reflect.DeepEqual(value.Attempts, wantAttempts) {
		return 0, 0, errors.New("attempt evidence differs from independent raw-source reconstruction")
	}
	wantClaims := buildClaimTransitions(model.Claims, wantAttempts)
	if !reflect.DeepEqual(value.Claims, wantClaims) {
		return 0, 0, errors.New("claim transitions are not derived from attempt evidence")
	}
	wantContract := buildContract(model, value.Source, wantAttempts, wantClaims)
	if !reflect.DeepEqual(value.Contract, wantContract) {
		return 0, 0, errors.New("contract is not derived from the canonical IR")
	}
	return countTransitions(wantClaims, "DISCHARGED"), len(model.Claims), nil
}

func deriveSourceModel(ir semantic.IR) (sourceModel, error) {
	model := sourceModel{}
	for _, node := range ir.Graph.Nodes() {
		id := node.ID.String()
		switch {
		case node.Kind == semantic.Entity && strings.Contains(id, "/claim/"):
			claim, err := parseClaim(node)
			if err != nil {
				return sourceModel{}, err
			}
			model.Claims = append(model.Claims, claim)
		case node.Kind == semantic.Entity && strings.Contains(id, "/subject/query"):
			model.QuerySubject = node.ID
		case node.Kind == semantic.Entity && strings.Contains(id, "/mutation/request"):
			model.MutationTarget = node.ID
		case node.Kind == semantic.Entity && strings.Contains(id, "/metric/relation"):
			model.MetricTarget = node.ID
		case node.Kind == semantic.Entity && strings.Contains(id, "/receipt/query"):
			model.ReceiptTarget = node.ID
		case node.Kind == semantic.Activity && isSandboxOperation(node.ValueProgram):
			model.Operations = append(model.Operations, operationSpec{ID: node.ID, Program: node.ValueProgram})
		}
	}
	if len(model.Claims) == 0 || model.QuerySubject == "" || model.MutationTarget == "" || model.ReceiptTarget == "" || model.MetricTarget == "" {
		return sourceModel{}, errors.New("source semantic model is incomplete")
	}
	sort.Slice(model.Claims, func(i, j int) bool { return model.Claims[i].ID < model.Claims[j].ID })
	sort.Slice(model.Operations, func(i, j int) bool { return model.Operations[i].ID < model.Operations[j].ID })
	for _, operation := range model.Operations {
		if operation.Program == "reflect.query:metrics" {
			model.UnknownTarget = query.ID(operation.ID.String() + "/unknown-target")
		}
	}
	if model.UnknownTarget == "" {
		return sourceModel{}, errors.New("source has no metrics query operation")
	}
	return model, nil
}

func parseClaim(node semantic.Node) (claimSpec, error) {
	const marker = "/claim/"
	index := strings.Index(node.ID.String(), marker)
	if index < 0 {
		return claimSpec{}, fmt.Errorf("claim marker missing from %q", node.ID)
	}
	parts := strings.Split(node.ID.String()[index+len(marker):], "/")
	if len(parts) != 6 {
		return claimSpec{}, fmt.Errorf("claim %q must encode six semantic coordinates", node.ID)
	}
	return claimSpec{ID: strings.ToLower(parts[0]) + "." + parts[1], Class: strings.ToUpper(parts[0]), ProofChoice: strings.ToUpper(parts[2]), MetaOperation: parts[3], EvidenceAttempt: parts[4], PriorState: strings.ToUpper(parts[5]), NodeID: node.ID}, nil
}

func isSandboxOperation(program string) bool {
	return strings.HasPrefix(program, "reflect.query:") || strings.HasPrefix(program, "reflect.attempt:") || strings.HasPrefix(program, "reflect.observation:")
}

func sourceTargets(ir semantic.IR, activity semantic.ID) []semantic.ID {
	targets := make([]semantic.ID, 0)
	for _, fact := range ir.Graph.DeterministicFacts() {
		if fact.Subject == activity && fact.Predicate == semantic.Used {
			targets = append(targets, fact.Object)
		}
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i] < targets[j] })
	return targets
}

func targetForOperation(ir semantic.IR, operation operationSpec) (semantic.ID, error) {
	marker := ""
	switch operation.Program {
	case "reflect.query:structure":
		marker = "/subject/structure"
	case "reflect.query:claims":
		marker = "/claim-state/open"
	case "reflect.query:metrics":
		marker = "/metric/relation"
	case "reflect.attempt:mutation":
		marker = "/mutation/request"
	case "reflect.observation:receipt-seal":
		marker = "/receipt/query"
	}
	for _, target := range sourceTargets(ir, operation.ID) {
		if strings.Contains(target.String(), marker) {
			return target, nil
		}
	}
	return "", fmt.Errorf("operation %q has no source-backed target", operation.Program)
}

func reconstructAttempts(ir semantic.IR, graph *query.Graph, model sourceModel, semanticDigest string) ([]attempt, bool, string, string, string, error) {
	attempts := make([]attempt, 0, len(model.Operations)+1)
	var authority bool
	var api, outcome, apiError string
	for _, operation := range model.Operations {
		target, err := targetForOperation(ir, operation)
		if err != nil {
			return nil, false, "", "", "", err
		}
		id := attemptIDForProgram(operation.Program)
		if strings.HasPrefix(operation.Program, "reflect.attempt:") {
			value, allowed, mutationAPI, mutationOutcome, mutationError, err := reconstructMutation(ir, operation, id, target, semanticDigest)
			if err != nil {
				return nil, false, "", "", "", err
			}
			attempts = append(attempts, value)
			authority, api, outcome, apiError = allowed, mutationAPI, mutationOutcome, mutationError
			continue
		}
		attempts = append(attempts, reconstructExact(graph, operation, id, target, semanticDigest, strings.HasPrefix(operation.Program, "reflect.observation:")))
	}
	metric, ok := findOperation(model.Operations, "reflect.query:metrics")
	if !ok {
		return nil, false, "", "", "", errors.New("metrics query operation is absent")
	}
	attempts = append(attempts, reconstructUnknown(graph, metric, model.UnknownTarget, semanticDigest, model.Claims))
	for index := range attempts {
		for _, claim := range model.Claims {
			if claim.EvidenceAttempt == attempts[index].ID {
				attempts[index].EvidenceClaimIDs = append(attempts[index].EvidenceClaimIDs, claim.ID)
			}
		}
	}
	sort.SliceStable(attempts, func(i, j int) bool { return attempts[i].ID < attempts[j].ID })
	return attempts, authority, api, outcome, apiError, nil
}

func reconstructExact(graph *query.Graph, operation operationSpec, id string, target semantic.ID, semanticDigest string, receipt bool) attempt {
	before := graph.StableHash()
	value := attempt{ID: id, Class: "SOURCE_DERIVED", Operation: "query", Root: operation.ID.String(), Relation: "used", Target: target.String(), MetaOperation: operation.Program, Producer: producerName, Consumer: consumerName, ProofChoice: "SOURCE_CLAIM_EVIDENCE", Stage: "QUERY", Step: "match-source-relation", SemanticDigestBefore: semanticDigest, GraphDigestBefore: before}
	if receipt {
		value.Stage = "RECEIPT"
	}
	result, err := graph.ExactMatch(query.NewExactQuery(query.ID(operation.ID.String()), query.Used, query.ID(target.String())))
	value.SemanticDigestAfter, value.GraphDigestAfter = semanticDigest, graph.StableHash()
	if err != nil {
		value.Decision, value.Resolution, value.Reason = "REFUTED", "LOWER_RESOLUTION", "QUERY_API_REJECTED"
		return value
	}
	value.MatchedFacts = len(result.All())
	if value.MatchedFacts == 1 {
		value.Decision, value.Resolution, value.Reason = "PASS", "EXACT", "EXACT_RELATION_MATCH"
	} else {
		value.Decision, value.Resolution, value.Reason = "UNKNOWN", "LOWER_RESOLUTION", "RELATION_NOT_OBSERVED"
	}
	return value
}

func reconstructMutation(ir semantic.IR, operation operationSpec, id string, target semantic.ID, semanticDigest string) (attempt, bool, string, string, string, error) {
	node, ok := ir.Graph.Node(target)
	if !ok {
		return attempt{}, false, "", "", "", fmt.Errorf("mutation target %q disappeared", target)
	}
	request := semantic.GraphPatchRequest{SchemaVersion: semantic.GraphPatchSchemaVersion, Operation: semantic.GraphPatchSetNodeField, ExpectedGraphHash: ir.Graph.StableHash(), NodeID: node.ID, ExpectedNodeHash: node.StableHash(), Field: "id", ExpectedSourceDigest: semanticDigest, ExpectedIRDigest: semanticDigest, AllowedIntent: "reflective-query-sandbox", Locality: "detached-observation-copy"}
	base := semantic.GraphPatchBase{SourceDigest: semanticDigest, IRDigest: semanticDigest}
	patched, err := ir.Graph.ApplyGraphPatch(base, request, semantic.GraphPatchMutation{})
	value := attempt{ID: id, Class: "SOURCE_DERIVED", Operation: "mutate", Root: operation.ID.String(), Relation: "set", Target: target.String(), MetaOperation: operation.Program, Producer: producerName, Consumer: consumerName, ProofChoice: "SOURCE_CLAIM_EVIDENCE", Stage: "MUTATION_BOUNDARY", Step: "apply-typed-request", API: "semantic.Graph.ApplyGraphPatch", SemanticDigestBefore: semanticDigest, SemanticDigestAfter: semanticDigest, GraphDigestBefore: ir.Graph.StableHash(), GraphDigestAfter: ir.Graph.StableHash()}
	if err != nil {
		value.Decision, value.Resolution, value.Reason = "DENIED", "EXACT_REJECTION", "MUTATION_REQUEST_REJECTED"
		value.APIOutcome, value.APIError = "REJECTED", err.Error()
		return value, false, value.API, value.APIOutcome, value.APIError, nil
	}
	patchedIR := ir
	patchedIR.Graph = patched
	value.Decision, value.Resolution, value.Reason = "REFUTED", "EXACT", "MUTATION_CAPABILITY_ACCEPTED"
	value.APIOutcome = "ACCEPTED"
	value.SemanticDigestAfter, value.GraphDigestAfter = patchedIR.StableHash(), patched.StableHash()
	return value, true, value.API, value.APIOutcome, "", nil
}

func reconstructUnknown(graph *query.Graph, operation operationSpec, target query.ID, semanticDigest string, claims []claimSpec) attempt {
	value := attempt{ID: "unknown.target", Class: classForAttempt("unknown.target", claims), Operation: "query", Root: operation.ID.String(), Relation: "used", Target: target.String(), MetaOperation: metaForAttempt("unknown.target", operation.Program, claims), Producer: producerName, Consumer: consumerName, ProofChoice: "SOURCE_CLAIM_EVIDENCE", Stage: "UNKNOWN", Step: "resolve-unknown-subject", SemanticDigestBefore: semanticDigest, GraphDigestBefore: graph.StableHash()}
	_, err := graph.ExactMatch(query.NewExactQuery(query.ID(operation.ID.String()), query.Used, target))
	value.SemanticDigestAfter, value.GraphDigestAfter = semanticDigest, graph.StableHash()
	if err != nil && errors.Is(err, query.ErrUnknownEndpoint) {
		value.Decision, value.Resolution, value.Reason = "UNKNOWN", "LOWER_RESOLUTION", "UNKNOWN_TARGET"
	} else if err != nil {
		value.Decision, value.Resolution, value.Reason = "REFUTED", "LOWER_RESOLUTION", "QUERY_API_REJECTED"
	} else {
		value.Decision, value.Resolution, value.Reason = "REFUTED", "EXACT", "UNKNOWN_TARGET_BECAME_KNOWN"
	}
	return value
}

func classForAttempt(id string, claims []claimSpec) string {
	for _, claim := range claims {
		if claim.EvidenceAttempt == id {
			return claim.Class
		}
	}
	return "SOURCE_DERIVED"
}

func metaForAttempt(id, fallback string, claims []claimSpec) string {
	for _, claim := range claims {
		if claim.EvidenceAttempt == id {
			return claim.MetaOperation
		}
	}
	return fallback
}

func findOperation(operations []operationSpec, program string) (operationSpec, bool) {
	for _, operation := range operations {
		if operation.Program == program {
			return operation, true
		}
	}
	return operationSpec{}, false
}

func attemptIDForProgram(program string) string {
	switch {
	case strings.HasPrefix(program, "reflect.query:"):
		return "reflect." + strings.TrimPrefix(program, "reflect.query:")
	case strings.HasPrefix(program, "reflect.attempt:"):
		return "mutation." + strings.TrimPrefix(program, "reflect.attempt:")
	case strings.HasPrefix(program, "reflect.observation:"):
		return "receipt." + strings.TrimPrefix(program, "reflect.observation:")
	default:
		return program
	}
}

func buildClaimTransitions(claims []claimSpec, attempts []attempt) []claimTransition {
	transitions := make([]claimTransition, 0, len(claims)*2)
	previous := ""
	sequence := 1
	for _, claim := range claims {
		registration := claimTransition{Sequence: sequence, ClaimID: claim.ID, Class: claim.Class, ProofChoice: claim.ProofChoice, MetaOperation: claim.MetaOperation, PriorState: claim.PriorState, EvidenceAttempt: claim.EvidenceAttempt, Producer: producerName, Consumer: consumerName, Stage: "DECLARE", Step: "register-source-claim", Reason: "CLAIM_PRIOR_STATE_OBSERVED", From: "UNRECORDED", To: claim.PriorState, PreviousDigest: previous}
		registration.Digest = transitionDigest(registration)
		transitions = append(transitions, registration)
		previous = registration.Digest
		sequence++
		to, stage, step, reason := claim.PriorState, "OBSERVE", "retain-prior-state", "NO_ATTEMPT_EVIDENCE"
		if value, ok := findAttempt(attempts, claim.EvidenceAttempt); ok {
			switch value.Decision {
			case "PASS", "DENIED":
				to, stage, step, reason = "DISCHARGED", "OBSERVE", "discharge-from-attempt-evidence", "ATTEMPT_EVIDENCE_MATCH"
			case "UNKNOWN":
				to, stage, step, reason = claim.PriorState, "RESOLVE", "retain-open-on-unknown", "UNKNOWN_PRESERVED"
			case "REFUTED":
				to, stage, step, reason = "REFUTED", "REFUTE", "mark-boundary-violation", "BOUNDARY_VIOLATION"
			}
		}
		value := registration
		value.Sequence = sequence
		value.Stage, value.Step, value.Reason = stage, step, reason
		value.From, value.To, value.PreviousDigest = claim.PriorState, to, previous
		value.Digest = transitionDigest(value)
		transitions = append(transitions, value)
		previous = value.Digest
		sequence++
	}
	return transitions
}

func findAttempt(attempts []attempt, id string) (attempt, bool) {
	for _, value := range attempts {
		if value.ID == id {
			return value, true
		}
	}
	return attempt{}, false
}

func buildContract(model sourceModel, source snapshot, attempts []attempt, claims []claimTransition) contract {
	classes := make([]bucket, 0)
	proofs := make([]bucket, 0)
	classTotals, proofTotals := map[string]int{}, map[string]int{}
	for _, value := range claims {
		if value.Sequence%2 == 1 {
			classTotals[value.Class]++
			proofTotals[value.ProofChoice]++
		}
	}
	classes = bucketsFromTotals(classTotals)
	proofs = bucketsFromTotals(proofTotals)
	return contract{Schema: schema, MetricID: metricID, GoVersion: runtime.Version(), Denominator: len(model.Claims), Classes: classes, Proofs: proofs, SourceNodes: source.NodeCount, SourceFacts: source.FactCount, ClaimCount: len(model.Claims), AttemptCount: len(attempts), ReflectiveQueries: countAttempts(attempts, func(value attempt) bool { return value.Operation == "query" }), SafeQueries: countAttempts(attempts, func(value attempt) bool {
		return value.Operation == "query" && value.Decision == "PASS" && value.Resolution == "EXACT"
	}), DeniedMutations: countAttempts(attempts, func(value attempt) bool { return value.Operation == "mutate" && value.Decision == "DENIED" }), UnknownTargets: countAttempts(attempts, func(value attempt) bool { return value.Decision == "UNKNOWN" }), RefutedAttempts: countAttempts(attempts, func(value attempt) bool { return value.Decision == "REFUTED" }), TransitionCount: len(claims), SatisfiedIndicators: countTransitions(claims, "DISCHARGED")}
}

func bucketsFromTotals(totals map[string]int) []bucket {
	names := make([]string, 0, len(totals))
	for name := range totals {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]bucket, 0, len(names))
	for _, name := range names {
		result = append(result, bucket{Name: name, Total: totals[name]})
	}
	return result
}

func countAttempts(values []attempt, predicate func(attempt) bool) int {
	count := 0
	for _, value := range values {
		if predicate(value) {
			count++
		}
	}
	return count
}

func countTransitions(values []claimTransition, state string) int {
	count := 0
	for _, value := range values {
		if value.To == state && value.From != state {
			count++
		}
	}
	return count
}

func buildReceipt(value observation, satisfied, total int) receipt {
	classes := make([]score, 0)
	proofs := make([]score, 0)
	classTotals, classSatisfied := map[string]int{}, map[string]int{}
	proofTotals, proofSatisfied := map[string]int{}, map[string]int{}
	for index := 0; index < len(value.Claims); index += 2 {
		registration := value.Claims[index]
		final := value.Claims[index+1]
		classTotals[registration.Class]++
		proofTotals[registration.ProofChoice]++
		if final.To == "DISCHARGED" {
			classSatisfied[registration.Class]++
			proofSatisfied[registration.ProofChoice]++
		}
	}
	classes = scoresFromTotals(classTotals, classSatisfied)
	proofs = scoresFromTotals(proofTotals, proofSatisfied)
	decision, reason := "PASS", "OBSERVATION_BOUNDARY_CONFORMANT"
	if value.Contract.RefutedAttempts > 0 {
		decision, reason = "REFUTED", "BOUNDARY_VIOLATION_OBSERVED"
	}
	subjectResolution := "EXACT_ONLY"
	if value.Contract.UnknownTargets > 0 {
		subjectResolution = "MIXED_EXACT_AND_LOWER_RESOLUTION"
	}
	return receipt{Schema: receiptSchema, SubjectSHA: value.SubjectSHA, MetricID: metricID, Decision: decision, Resolution: "OBSERVATION_ONLY", SubjectResolution: subjectResolution, Reason: reason, Producer: value.Producer, Consumer: consumerName, Contract: value.Contract, Source: value.Source, Attempts: value.Attempts, Claims: value.Claims, Coordinates: coordinates{Satisfied: satisfied, Total: total, BasisPoints: basisPoints(satisfied, total)}, Classes: classes, Proofs: proofs, Effects: value.Effects, SourceReconstruction: coordinates{Satisfied: 4, Total: 4, BasisPoints: 10000}, ProducerImports: coordinates{Satisfied: 0, Total: 0, BasisPoints: 0}, PromotionCreditBPS: 0, RepositoryWrites: value.Effects.RepositoryWrites, MutationAuthority: value.Effects.MutationAuthority, NotClaimed: []string{"generic Go reflection API equivalence", "runtime capability isolation beyond this process", "source completeness beyond declared claims", "mutation safety against a hostile process", "runtime memory and performance bounds"}}
}

func scoresFromTotals(totals, satisfied map[string]int) []score {
	names := make([]string, 0, len(totals))
	for name := range totals {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]score, 0, len(names))
	for _, name := range names {
		result = append(result, score{Name: name, Satisfied: satisfied[name], Total: totals[name]})
	}
	return result
}

func basisPoints(satisfied, total int) int {
	if total == 0 {
		return 0
	}
	return satisfied * 10000 / total
}

func observationDigest(value observation) string {
	value.Digest = ""
	return hashJSON(value)
}

func transitionDigest(value claimTransition) string {
	value.Digest = ""
	return hashJSON(value)
}

func receiptDigest(value receipt) string {
	value.Digest = ""
	return hashJSON(value)
}

func hashJSON(value any) string {
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	return strings.Count(string(data), "\n")
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
