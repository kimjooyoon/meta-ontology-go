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
	schema        = "gooo/reflective-query-sandbox-observation/v3"
	receiptSchema = "gooo/reflective-query-sandbox-receipt/v3"
	metricID      = "gooo.metric.language.reflective-query-sandbox.v3"
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
	RepositoryStatusBefore []string `json:"repository_status_before"`
	RepositoryStatusAfter  []string `json:"repository_status_after"`
	NetRepositoryChanges   []string `json:"net_repository_changes"`
	MutationAuthority      bool     `json:"mutation_authority"`
	MutationAPI            string   `json:"mutation_api"`
	MutationOutcome        string   `json:"mutation_outcome"`
	MutationError          string   `json:"mutation_error,omitempty"`
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
	ID                          string   `json:"id"`
	Class                       string   `json:"class"`
	Operation                   string   `json:"operation"`
	Root                        string   `json:"root"`
	Relation                    string   `json:"relation"`
	Target                      string   `json:"target"`
	MetaOperation               string   `json:"meta_operation"`
	Producer                    string   `json:"producer"`
	Consumer                    string   `json:"consumer"`
	ProofChoice                 string   `json:"proof_choice"`
	Stage                       string   `json:"stage"`
	Step                        string   `json:"step"`
	Decision                    string   `json:"decision"`
	Resolution                  string   `json:"resolution"`
	Reason                      string   `json:"reason"`
	MatchedFacts                int      `json:"matched_facts"`
	EvidenceClaimIDs            []string `json:"evidence_claim_ids"`
	API                         string   `json:"api,omitempty"`
	APIOutcome                  string   `json:"api_outcome,omitempty"`
	APIError                    string   `json:"api_error,omitempty"`
	APIErrorCode                string   `json:"api_error_code,omitempty"`
	ObservedMaterialDigest      string   `json:"observed_material_digest,omitempty"`
	SemanticDigestBefore        string   `json:"semantic_digest_before"`
	SemanticDigestAfter         string   `json:"semantic_digest_after"`
	GraphDigestBefore           string   `json:"graph_digest_before"`
	GraphDigestAfter            string   `json:"graph_digest_after"`
	OriginalSemanticDigestAfter string   `json:"original_semantic_digest_after"`
	OriginalGraphDigestAfter    string   `json:"original_graph_digest_after"`
	ReturnedSemanticDigest      string   `json:"returned_semantic_digest,omitempty"`
	ReturnedGraphDigest         string   `json:"returned_graph_digest,omitempty"`
}
type claimTransition struct {
	Sequence               int    `json:"sequence"`
	ClaimID                string `json:"claim_id"`
	Class                  string `json:"class"`
	ProofChoice            string `json:"proof_choice"`
	MetaOperation          string `json:"meta_operation"`
	PriorState             string `json:"prior_state"`
	EvidenceAttempt        string `json:"evidence_attempt"`
	PredicateID            string `json:"predicate_id"`
	Producer               string `json:"producer"`
	Consumer               string `json:"consumer"`
	Stage                  string `json:"stage"`
	Step                   string `json:"step"`
	Reason                 string `json:"reason"`
	From                   string `json:"from"`
	To                     string `json:"to"`
	PreviousDigest         string `json:"previous_digest"`
	Digest                 string `json:"digest"`
	ObservedMaterialDigest string `json:"observed_material_digest,omitempty"`
}
type observation struct {
	Schema                string            `json:"schema"`
	SubjectSHA            string            `json:"subject_sha"`
	Contract              contract          `json:"contract"`
	Source                snapshot          `json:"source"`
	Attempts              []attempt         `json:"attempts"`
	Claims                []claimTransition `json:"claims"`
	Effects               effects           `json:"effects"`
	ReceiptMaterialDigest string            `json:"receipt_material_digest"`
	Producer              string            `json:"producer"`
	Digest                string            `json:"digest"`
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
type importBoundary struct {
	ForbiddenImportsObserved int    `json:"forbidden_imports_observed"`
	MaximumAllowed           int    `json:"maximum_allowed"`
	EvidenceDigest           string `json:"evidence_digest"`
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
	ImportBoundary       importBoundary    `json:"import_boundary"`
	PromotionCreditBPS   int               `json:"promotion_credit_bps"`
	MutationAuthority    bool              `json:"mutation_authority"`
	NotClaimed           []string          `json:"not_claimed"`
	Digest               string            `json:"digest"`
}
type claimSpec struct {
	ID, Class, PredicateID, ProofChoice, MetaOperation, EvidenceAttempt, PriorState string
	NodeID                                                                          semantic.ID
}
type operationSpec struct {
	ID      semantic.ID
	Program string
}
type sourceModel struct {
	Claims           []claimSpec
	Operations       []operationSpec
	QuerySubject     semantic.ID
	MutationTarget   semantic.ID
	MutationField    semantic.ID
	MutationPayload  semantic.ID
	MutationIntent   semantic.ID
	MutationLocality semantic.ID
	RepositoryTarget semantic.ID
	ReceiptTarget    semantic.ID
	MetricTarget     semantic.ID
	UnknownTarget    query.ID
}

func main() {
	input := flag.String("input", "", "producer observation")
	source := flag.String("source", "", "Gooo source")
	subject := flag.String("subject-sha", "", "exact subject commit")
	importsFile := flag.String("producer-imports-evidence", "", "raw go list evidence")
	maximum := flag.Int("producer-imports-maximum", -1, "maximum forbidden imports")
	output := flag.String("output", "", "independent receipt")
	flag.Parse()
	if *input == "" || *source == "" || *subject == "" || *importsFile == "" || *maximum < 0 || *output == "" {
		fail("usage: consumer -input FILE -source FILE -subject-sha SHA -producer-imports-evidence FILE -producer-imports-maximum N -output FILE")
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
	importData, err := os.ReadFile(*importsFile)
	if err != nil {
		fail("read import evidence: %v", err)
	}
	forbidden := 0
	for _, line := range strings.Split(string(importData), "\n") {
		if strings.Contains(line, "/internal/meta/reflectivequerysandbox") {
			forbidden++
		}
	}
	if forbidden > *maximum {
		fail("producer import boundary exceeded: observed=%d maximum=%d", forbidden, *maximum)
	}
	satisfied, total, reconstruction, err := validateObservation(value, *source, *subject)
	if err != nil {
		fail("independent reconstruction: %v", err)
	}
	imports := importBoundary{ForbiddenImportsObserved: forbidden, MaximumAllowed: *maximum, EvidenceDigest: bytesDigest(importData)}
	result := buildReceipt(value, satisfied, total, reconstruction, coordinates{Satisfied: 1, Total: 1, BasisPoints: 10000}, imports)
	result.Digest = receiptDigest(result)
	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fail("encode receipt: %v", err)
	}
	if err := os.WriteFile(*output, append(encoded, '\n'), 0o644); err != nil {
		fail("write receipt: %v", err)
	}
	fmt.Printf("consumer verdict: %s %d/%d queries=%d/%d source=%d/%d imports=%d/%d net_repository_changes=%d mutation_authority=%t\n", result.Decision, result.Coordinates.Satisfied, result.Coordinates.Total, value.Contract.SafeQueries, value.Contract.ReflectiveQueries, result.SourceReconstruction.Satisfied, result.SourceReconstruction.Total, result.ProducerImports.Satisfied, result.ProducerImports.Total, len(result.Effects.NetRepositoryChanges), result.Effects.MutationAuthority)
}

func validateObservation(value observation, sourcePath, subject string) (int, int, coordinates, error) {
	if value.Schema != schema || value.SubjectSHA != subject || value.Producer != producerName {
		return 0, 0, coordinates{}, errors.New("observation identity is not exact")
	}
	if value.Digest != observationDigest(value) {
		return 0, 0, coordinates{}, errors.New("observation digest mismatch")
	}
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return 0, 0, coordinates{}, fmt.Errorf("read source: %w", err)
	}
	file, diagnostics := syntax.ParseFile(sourcePath, string(data))
	if diagnostics.HasErrors() {
		return 0, 0, coordinates{}, diagnostics.Error()
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return 0, 0, coordinates{}, fmt.Errorf("lower source: %w", err)
	}
	graph, err := query.FromSemanticIR(ir)
	if err != nil {
		return 0, 0, coordinates{}, fmt.Errorf("project query view: %w", err)
	}
	model, err := deriveSourceModel(ir)
	if err != nil {
		return 0, 0, coordinates{}, fmt.Errorf("derive source model: %w", err)
	}
	semanticDigest := ir.StableHash()
	wantSource := snapshot{Path: sourcePath, SourceDigest: semantic.StableHash(data), SemanticDigest: semanticDigest, GraphDigest: graph.StableHash(), NodeCount: len(graph.Nodes()), FactCount: len(graph.AllFacts()), GoooLines: countLines(data)}
	if !reflect.DeepEqual(value.Source, wantSource) {
		return 0, 0, coordinates{}, fmt.Errorf("source reconstruction differs: got=%+v want=%+v", value.Source, wantSource)
	}
	if !strings.HasPrefix(runtime.Version(), "go1.27") || value.Contract.GoVersion != runtime.Version() {
		return 0, 0, coordinates{}, fmt.Errorf("runtime=%s contract=%s", runtime.Version(), value.Contract.GoVersion)
	}
	if value.Effects.RepositoryStatusBefore == nil || value.Effects.RepositoryStatusAfter == nil || !reflect.DeepEqual(value.Effects.RepositoryStatusBefore, value.Effects.RepositoryStatusAfter) || !reflect.DeepEqual(changedLines(value.Effects.RepositoryStatusBefore, value.Effects.RepositoryStatusAfter), value.Effects.NetRepositoryChanges) || len(value.Effects.NetRepositoryChanges) != 0 || value.Effects.MutationAuthority || value.Effects.MutationOutcome != "REJECTED" || value.Effects.MutationAPI == "" {
		return 0, 0, coordinates{}, errors.New("observed effects do not prove the reported net repository relation and mutation boundary")
	}
	wantAttempts, authority, api, outcome, apiError, err := reconstructAttempts(ir, graph, model, semanticDigest, value.Effects)
	if err != nil {
		return 0, 0, coordinates{}, err
	}
	if authority || api != value.Effects.MutationAPI || outcome != value.Effects.MutationOutcome || apiError != value.Effects.MutationError {
		return 0, 0, coordinates{}, errors.New("mutation boundary was not independently reconstructed")
	}
	material := receiptMaterialDigest(wantSource, wantAttempts, value.Effects)
	for index := range wantAttempts {
		if wantAttempts[index].ID == "receipt.seal" {
			wantAttempts[index].ObservedMaterialDigest = material
		}
	}
	if value.ReceiptMaterialDigest != material {
		return 0, 0, coordinates{}, errors.New("receipt material digest mismatch")
	}
	if !reflect.DeepEqual(value.Attempts, wantAttempts) {
		return 0, 0, coordinates{}, errors.New("attempt evidence differs from independent raw-source reconstruction")
	}
	wantClaims := buildClaimTransitions(model.Claims, wantAttempts, value.Effects, material)
	if !reflect.DeepEqual(value.Claims, wantClaims) {
		return 0, 0, coordinates{}, errors.New("claim transitions are not independently predicate-evaluated")
	}
	wantContract := buildContract(model, wantSource, wantAttempts, wantClaims)
	if !reflect.DeepEqual(value.Contract, wantContract) {
		return 0, 0, coordinates{}, errors.New("contract is not derived from canonical IR")
	}
	checks := []bool{reflect.DeepEqual(value.Source, wantSource), reflect.DeepEqual(value.Effects.RepositoryStatusBefore, value.Effects.RepositoryStatusAfter) && len(value.Effects.NetRepositoryChanges) == 0, reflect.DeepEqual(value.Attempts, wantAttempts), reflect.DeepEqual(value.Claims, wantClaims), reflect.DeepEqual(value.Contract, wantContract), value.ReceiptMaterialDigest == material}
	return countTransitions(wantClaims, "DISCHARGED"), len(model.Claims), reconstructionCoordinates(checks), nil
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
		case node.Kind == semantic.Entity && strings.Contains(id, "/mutation/field/"):
			model.MutationField = node.ID
		case node.Kind == semantic.Entity && strings.Contains(id, "/mutation/payload/"):
			model.MutationPayload = node.ID
		case node.Kind == semantic.Entity && strings.Contains(id, "/mutation/intent/"):
			model.MutationIntent = node.ID
		case node.Kind == semantic.Entity && strings.Contains(id, "/mutation/locality/"):
			model.MutationLocality = node.ID
		case node.Kind == semantic.Entity && strings.Contains(id, "/repository/status/net"):
			model.RepositoryTarget = node.ID
		case node.Kind == semantic.Entity && strings.Contains(id, "/metric/relation"):
			model.MetricTarget = node.ID
		case node.Kind == semantic.Entity && strings.Contains(id, "/receipt/query"):
			model.ReceiptTarget = node.ID
		case node.Kind == semantic.Activity && isSandboxOperation(node.ValueProgram):
			model.Operations = append(model.Operations, operationSpec{ID: node.ID, Program: node.ValueProgram})
		}
	}
	if len(model.Claims) == 0 || model.QuerySubject == "" || model.MutationTarget == "" || model.MutationField == "" || model.MutationPayload == "" || model.MutationIntent == "" || model.MutationLocality == "" || model.RepositoryTarget == "" || model.ReceiptTarget == "" || model.MetricTarget == "" {
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
	marker := "/claim/"
	index := strings.Index(node.ID.String(), marker)
	if index < 0 {
		return claimSpec{}, fmt.Errorf("claim marker missing from %q", node.ID)
	}
	parts := strings.Split(node.ID.String()[index+len(marker):], "/")
	if len(parts) != 7 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" || parts[4] == "" || parts[5] == "" || parts[6] == "" {
		return claimSpec{}, fmt.Errorf("claim %q must encode seven nonempty semantic coordinates", node.ID)
	}
	class, predicate, proof, prior := strings.ToUpper(parts[0]), parts[2], strings.ToUpper(parts[3]), strings.ToUpper(parts[6])
	if class != "OUTCOME" && class != "DRIVER" && class != "GUARDRAIL" {
		return claimSpec{}, fmt.Errorf("claim %q has unsupported class", node.ID)
	}
	if proof != "FOUNDATION" && proof != "COHERENCE" && proof != "REGRESSION" {
		return claimSpec{}, fmt.Errorf("claim %q has unsupported proof", node.ID)
	}
	if prior != "OPEN" || !allowedPredicate(predicate) {
		return claimSpec{}, fmt.Errorf("claim %q has unsupported prior/predicate", node.ID)
	}
	return claimSpec{ID: strings.ToLower(parts[0]) + "." + parts[1], Class: class, PredicateID: predicate, ProofChoice: proof, MetaOperation: parts[4], EvidenceAttempt: parts[5], PriorState: prior, NodeID: node.ID}, nil
}
func allowedPredicate(predicate string) bool {
	switch predicate {
	case "query-relation-exact", "semantic-digest-equal", "graph-digest-equal", "query-projection-stable", "receipt-observation-digest-verified", "claim-ledger-chained", "unknown-subject-preserved", "immutable-mutation-rejected", "mutation-boundary-rejected", "net-repository-changes-empty":
		return true
	}
	return false
}
func isSandboxOperation(program string) bool {
	return strings.HasPrefix(program, "reflect.query:") || strings.HasPrefix(program, "reflect.attempt:") || strings.HasPrefix(program, "reflect.observation:")
}
func sourceTargets(ir semantic.IR, activity semantic.ID) []semantic.ID {
	result := []semantic.ID{}
	for _, fact := range ir.Graph.DeterministicFacts() {
		if fact.Subject == activity && fact.Predicate == semantic.Used {
			result = append(result, fact.Object)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
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
	case "reflect.observation:repository-net":
		marker = "/repository/status/net"
	}
	for _, target := range sourceTargets(ir, operation.ID) {
		if marker == "" || strings.Contains(target.String(), marker) {
			return target, nil
		}
	}
	return "", fmt.Errorf("operation %q has no source-backed target", operation.Program)
}

func reconstructAttempts(ir semantic.IR, graph *query.Graph, model sourceModel, semanticDigest string, fx effects) ([]attempt, bool, string, string, string, error) {
	values := make([]attempt, 0, len(model.Operations)+1)
	authority, api, outcome, apiError := false, "", "", ""
	for _, operation := range model.Operations {
		target, err := targetForOperation(ir, operation)
		if err != nil {
			return nil, false, "", "", "", err
		}
		id := attemptIDForProgram(operation.Program)
		switch {
		case strings.HasPrefix(operation.Program, "reflect.attempt:"):
			value, allowed, mutationAPI, mutationOutcome, mutationError, err := reconstructMutation(ir, operation, id, target, semanticDigest, model, model.Claims)
			if err != nil {
				return nil, false, "", "", "", err
			}
			values = append(values, value)
			authority, api, outcome, apiError = allowed, mutationAPI, mutationOutcome, mutationError
		case operation.Program == "reflect.observation:repository-net":
			values = append(values, repositoryAttempt(operation, id, target, semanticDigest, fx, model.Claims))
		default:
			values = append(values, reconstructExact(graph, operation, id, target, semanticDigest, strings.HasPrefix(operation.Program, "reflect.observation:"), model.Claims))
		}
	}
	metric, ok := findOperation(model.Operations, "reflect.query:metrics")
	if !ok {
		return nil, false, "", "", "", errors.New("metrics query operation is absent")
	}
	values = append(values, reconstructUnknown(graph, metric, model.UnknownTarget, semanticDigest, model.Claims))
	for index := range values {
		for _, claim := range model.Claims {
			if claim.EvidenceAttempt == values[index].ID {
				values[index].EvidenceClaimIDs = append(values[index].EvidenceClaimIDs, claim.ID)
			}
		}
	}
	sort.SliceStable(values, func(i, j int) bool { return values[i].ID < values[j].ID })
	return values, authority, api, outcome, apiError, nil
}
func reconstructExact(graph *query.Graph, operation operationSpec, id string, target semantic.ID, semanticDigest string, receipt bool, claims []claimSpec) attempt {
	before := graph.StableHash()
	value := attempt{ID: id, Class: classForAttempt(id, claims), Operation: "query", Root: operation.ID.String(), Relation: "used", Target: target.String(), MetaOperation: metaForAttempt(id, operation.Program, claims), Producer: producerName, Consumer: consumerName, ProofChoice: proofForAttempt(id, claims), Stage: "QUERY", Step: "match-source-relation", SemanticDigestBefore: semanticDigest, GraphDigestBefore: before}
	if receipt {
		value.Stage = "RECEIPT"
	}
	result, err := graph.ExactMatch(query.NewExactQuery(query.ID(operation.ID.String()), query.Used, query.ID(target.String())))
	value.SemanticDigestAfter, value.GraphDigestAfter = semanticDigest, graph.StableHash()
	value.ObservedMaterialDigest = value.GraphDigestAfter
	if err != nil {
		value.Decision, value.Resolution, value.Reason = "UNKNOWN", "LOWER_RESOLUTION", "QUERY_API_ERROR"
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
func reconstructMutation(ir semantic.IR, operation operationSpec, id string, target semantic.ID, semanticDigest string, model sourceModel, claims []claimSpec) (attempt, bool, string, string, string, error) {
	node, ok := ir.Graph.Node(target)
	if !ok {
		return attempt{}, false, "", "", "", fmt.Errorf("mutation target %q disappeared", target)
	}
	field, payload, intent, locality := tail(model.MutationField), tail(model.MutationPayload), tail(model.MutationIntent), tail(model.MutationLocality)
	fieldHash := ""
	var err error
	if field != "id" {
		fieldHash, err = semantic.NodeFieldHash(node, field)
		if err != nil {
			return attempt{}, false, "", "", "", err
		}
	}
	beforeSemantic, beforeGraph := ir.StableHash(), ir.Graph.StableHash()
	request := semantic.GraphPatchRequest{SchemaVersion: semantic.GraphPatchSchemaVersion, Operation: semantic.GraphPatchSetNodeField, ExpectedGraphHash: beforeGraph, NodeID: node.ID, ExpectedNodeHash: node.StableHash(), Field: field, ExpectedFieldHash: fieldHash, ExpectedSourceDigest: semanticDigest, ExpectedIRDigest: semanticDigest, AllowedIntent: intent, Locality: locality}
	base := semantic.GraphPatchBase{SourceDigest: semanticDigest, IRDigest: semanticDigest}
	patched, callErr := ir.Graph.ApplyGraphPatch(base, request, semantic.GraphPatchMutation{Name: payload})
	afterSemantic, afterGraph := ir.StableHash(), ir.Graph.StableHash()
	value := attempt{ID: id, Class: classForAttempt(id, claims), Operation: "mutate", Root: operation.ID.String(), Relation: "set", Target: target.String(), MetaOperation: metaForAttempt(id, operation.Program, claims), Producer: producerName, Consumer: consumerName, ProofChoice: proofForAttempt(id, claims), Stage: "MUTATION_BOUNDARY", Step: "apply-typed-request", API: "semantic.Graph.ApplyGraphPatch", SemanticDigestBefore: beforeSemantic, SemanticDigestAfter: afterSemantic, GraphDigestBefore: beforeGraph, GraphDigestAfter: afterGraph, OriginalSemanticDigestAfter: afterSemantic, OriginalGraphDigestAfter: afterGraph}
	if callErr != nil {
		value.APIErrorCode = mutationErrorCode(callErr)
		value.APIOutcome, value.APIError = "REJECTED", callErr.Error()
		if afterSemantic != beforeSemantic || afterGraph != beforeGraph {
			value.Decision, value.Resolution, value.Reason = "REFUTED", "EXACT", "MUTATION_API_CHANGED_OR_PARTIALLY_MUTATED"
			return value, true, value.API, value.APIOutcome, value.APIError, nil
		}
		var conflict semantic.GraphPatchConflict
		if errors.As(callErr, &conflict) && conflict.Code == semantic.PatchImmutableField && conflict.Detail == field {
			value.Decision, value.Resolution, value.Reason = "DENIED", "EXACT_REJECTION", "IMMUTABLE_FIELD_REJECTED"
			return value, false, value.API, value.APIOutcome, value.APIError, nil
		}
		value.Decision, value.Resolution, value.Reason = "UNKNOWN", "LOWER_RESOLUTION", "MUTATION_API_ERROR_"+strings.ToUpper(value.APIErrorCode)
		return value, false, value.API, "ERROR", value.APIError, nil
	}
	patchedIR := ir
	patchedIR.Graph = patched
	value.ReturnedSemanticDigest, value.ReturnedGraphDigest = patchedIR.StableHash(), patched.StableHash()
	value.SemanticDigestAfter, value.GraphDigestAfter = value.ReturnedSemanticDigest, value.ReturnedGraphDigest
	value.Decision, value.Resolution, value.Reason, value.APIOutcome = "REFUTED", "EXACT", "MUTATION_CAPABILITY_ACCEPTED", "ACCEPTED"
	return value, true, value.API, value.APIOutcome, "", nil
}
func tail(id semantic.ID) string {
	parts := strings.Split(strings.TrimSuffix(id.String(), "/"), "/")
	return parts[len(parts)-1]
}
func mutationErrorCode(err error) string {
	var conflict semantic.GraphPatchConflict
	if errors.As(err, &conflict) {
		return string(conflict.Code)
	}
	return "unknown"
}
func repositoryAttempt(operation operationSpec, id string, target semantic.ID, semanticDigest string, fx effects, claims []claimSpec) attempt {
	material := semantic.StableHashString(strings.Join(fx.RepositoryStatusBefore, "\n") + "\x00" + strings.Join(fx.RepositoryStatusAfter, "\n"))
	value := attempt{ID: id, Class: classForAttempt(id, claims), Operation: "repository", Root: operation.ID.String(), Relation: "net", Target: target.String(), MetaOperation: metaForAttempt(id, operation.Program, claims), Producer: producerName, Consumer: consumerName, ProofChoice: proofForAttempt(id, claims), Stage: "REPOSITORY", Step: "compare-status-snapshots", SemanticDigestBefore: semanticDigest, SemanticDigestAfter: semanticDigest, GraphDigestBefore: material, GraphDigestAfter: material, ObservedMaterialDigest: material}
	if len(fx.NetRepositoryChanges) == 0 && stringSliceEqual(fx.RepositoryStatusBefore, fx.RepositoryStatusAfter) {
		value.Decision, value.Resolution, value.Reason = "PASS", "EXACT", "NET_REPOSITORY_CHANGES_EMPTY"
	} else {
		value.Decision, value.Resolution, value.Reason = "REFUTED", "EXACT", "NET_REPOSITORY_CHANGES_OBSERVED"
	}
	return value
}
func reconstructUnknown(graph *query.Graph, operation operationSpec, target query.ID, semanticDigest string, claims []claimSpec) attempt {
	value := attempt{ID: "unknown.target", Class: classForAttempt("unknown.target", claims), Operation: "query", Root: operation.ID.String(), Relation: "used", Target: target.String(), MetaOperation: metaForAttempt("unknown.target", operation.Program, claims), Producer: producerName, Consumer: consumerName, ProofChoice: proofForAttempt("unknown.target", claims), Stage: "UNKNOWN", Step: "resolve-unknown-subject", SemanticDigestBefore: semanticDigest, GraphDigestBefore: graph.StableHash()}
	_, err := graph.ExactMatch(query.NewExactQuery(query.ID(operation.ID.String()), query.Used, target))
	value.SemanticDigestAfter, value.GraphDigestAfter = semanticDigest, graph.StableHash()
	value.ObservedMaterialDigest = value.GraphDigestAfter
	if err != nil && errors.Is(err, query.ErrUnknownEndpoint) {
		value.Decision, value.Resolution, value.Reason = "UNKNOWN", "LOWER_RESOLUTION", "UNKNOWN_TARGET"
	} else if err != nil {
		value.Decision, value.Resolution, value.Reason = "UNKNOWN", "LOWER_RESOLUTION", "QUERY_API_ERROR"
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
func proofForAttempt(id string, claims []claimSpec) string {
	for _, claim := range claims {
		if claim.EvidenceAttempt == id {
			return claim.ProofChoice
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
func findOperation(values []operationSpec, program string) (operationSpec, bool) {
	for _, value := range values {
		if value.Program == program {
			return value, true
		}
	}
	return operationSpec{}, false
}
func attemptIDForProgram(program string) string {
	switch {
	case strings.HasPrefix(program, "reflect.query:"):
		return "reflect." + strings.TrimPrefix(program, "reflect.query:")
	case strings.HasPrefix(program, "reflect.attempt:"):
		return "mutation.attempt"
	case strings.HasPrefix(program, "reflect.observation:"):
		if strings.HasSuffix(program, ":repository-net") {
			return "repository.net"
		}
		return "receipt.seal"
	default:
		return program
	}
}

type predicateResult struct{ To, Stage, Step, Reason, Material string }

func buildClaimTransitions(claims []claimSpec, attempts []attempt, fx effects, material string) []claimTransition {
	values := make([]claimTransition, 0, len(claims)*2)
	previous := ""
	sequence := 1
	for _, claim := range claims {
		registration := claimTransition{Sequence: sequence, ClaimID: claim.ID, Class: claim.Class, ProofChoice: claim.ProofChoice, MetaOperation: claim.MetaOperation, PriorState: claim.PriorState, EvidenceAttempt: claim.EvidenceAttempt, PredicateID: claim.PredicateID, Producer: producerName, Consumer: consumerName, Stage: "DECLARE", Step: "register-source-claim", Reason: "CLAIM_PRIOR_STATE_OBSERVED", From: "UNRECORDED", To: claim.PriorState, PreviousDigest: previous, ObservedMaterialDigest: semantic.StableHashString(claim.ID + "|" + claim.PredicateID + "|" + claim.PriorState)}
		registration.Digest = transitionDigest(registration)
		values = append(values, registration)
		previous = registration.Digest
		result := evaluateClaim(claim, attempts, fx, material)
		final := registration
		final.Sequence = sequence + 1
		final.Stage, final.Step, final.Reason, final.From, final.To, final.PreviousDigest, final.ObservedMaterialDigest = result.Stage, result.Step, result.Reason, claim.PriorState, result.To, previous, result.Material
		if claim.PredicateID == "claim-ledger-chained" && final.PreviousDigest != registration.Digest {
			final.Stage, final.Step, final.Reason, final.To, final.ObservedMaterialDigest = "RESOLVE", "verify-claim-ledger-chain", "CLAIM_LEDGER_CHAIN_BROKEN", claim.PriorState, registration.Digest
		}
		final.Digest = transitionDigest(final)
		values = append(values, final)
		previous = final.Digest
		sequence += 2
	}
	return values
}
func evaluateClaim(claim claimSpec, attempts []attempt, fx effects, material string) predicateResult {
	value, ok := findAttempt(attempts, claim.EvidenceAttempt)
	if !ok {
		return predicateResult{To: claim.PriorState, Stage: "RESOLVE", Step: "resolve-missing-evidence", Reason: "EVIDENCE_MISSING"}
	}
	contradiction := func(reason string) predicateResult {
		return predicateResult{To: "REFUTED", Stage: "REFUTE", Step: "predicate-contradiction", Reason: reason, Material: value.ObservedMaterialDigest}
	}
	observationError := func() predicateResult {
		return predicateResult{To: claim.PriorState, Stage: "RESOLVE", Step: "predicate-observation-error", Reason: value.Reason, Material: value.ObservedMaterialDigest}
	}
	exact := value.Decision == "PASS" && value.Resolution == "EXACT" && value.MatchedFacts == 1
	sameSemantic := value.SemanticDigestBefore != "" && value.SemanticDigestBefore == value.SemanticDigestAfter
	sameGraph := value.GraphDigestBefore != "" && value.GraphDigestBefore == value.GraphDigestAfter
	switch claim.PredicateID {
	case "query-relation-exact":
		if exact {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "evaluate-query-relation", Reason: "EXACT_RELATION_MATCH", Material: value.GraphDigestAfter}
		}
		if value.Decision == "REFUTED" {
			return contradiction("QUERY_RELATION_CONTRADICTION")
		}
		return observationError()
	case "semantic-digest-equal":
		if sameSemantic {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "compare-semantic-digest", Reason: "SEMANTIC_DIGEST_EQUAL", Material: semantic.StableHashString(value.SemanticDigestBefore + "|" + value.SemanticDigestAfter)}
		}
		if value.SemanticDigestBefore != "" && value.SemanticDigestAfter != "" {
			return contradiction("SEMANTIC_DIGEST_CHANGED")
		}
		return observationError()
	case "graph-digest-equal":
		if sameGraph {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "compare-graph-digest", Reason: "GRAPH_DIGEST_EQUAL", Material: value.GraphDigestAfter}
		}
		if value.GraphDigestBefore != "" && value.GraphDigestAfter != "" {
			return contradiction("GRAPH_DIGEST_CHANGED")
		}
		return observationError()
	case "query-projection-stable":
		if exact && sameGraph {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "verify-query-projection", Reason: "QUERY_MATCH_AND_GRAPH_STABLE", Material: value.GraphDigestAfter}
		}
		if value.Decision == "REFUTED" {
			return contradiction("QUERY_PROJECTION_CONTRADICTION")
		}
		return observationError()
	case "receipt-observation-digest-verified":
		if value.ObservedMaterialDigest != "" && material != "" && value.ObservedMaterialDigest == material {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "verify-receipt-material-digest", Reason: "RECEIPT_MATERIAL_DIGEST_VERIFIED", Material: material}
		}
		if value.ObservedMaterialDigest != "" && material != "" {
			return contradiction("RECEIPT_MATERIAL_DIGEST_MISMATCH")
		}
		return observationError()
	case "claim-ledger-chained":
		if value.ObservedMaterialDigest != "" && value.Decision == "PASS" {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "verify-claim-ledger-evidence", Reason: "CLAIM_LEDGER_EVIDENCE_PRESENT", Material: value.ObservedMaterialDigest}
		}
		if value.Decision == "REFUTED" {
			return contradiction("CLAIM_LEDGER_EVIDENCE_CONTRADICTION")
		}
		return observationError()
	case "unknown-subject-preserved":
		if value.Decision == "UNKNOWN" && value.Resolution == "LOWER_RESOLUTION" && value.Stage == "UNKNOWN" && value.Step == "resolve-unknown-subject" && value.Reason == "UNKNOWN_TARGET" {
			return predicateResult{To: claim.PriorState, Stage: "RESOLVE", Step: "retain-open-on-unknown", Reason: "UNKNOWN_PRESERVED", Material: value.GraphDigestAfter}
		}
		if value.Decision == "REFUTED" {
			return contradiction("UNKNOWN_SUBJECT_BOUNDARY_CONTRADICTION")
		}
		return observationError()
	case "immutable-mutation-rejected":
		if immutableRejection(value) {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "verify-immutable-rejection", Reason: "IMMUTABLE_FIELD_REJECTED", Material: value.OriginalGraphDigestAfter}
		}
		if value.Decision == "REFUTED" || value.APIOutcome == "ACCEPTED" {
			return contradiction("IMMUTABLE_MUTATION_BOUNDARY_VIOLATED")
		}
		return observationError()
	case "mutation-boundary-rejected":
		if immutableRejection(value) && !fx.MutationAuthority && fx.MutationOutcome == "REJECTED" {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "verify-mutation-boundary", Reason: "TYPED_REJECTION_AND_GRAPH_UNCHANGED", Material: value.OriginalGraphDigestAfter}
		}
		if fx.MutationAuthority || value.APIOutcome == "ACCEPTED" || value.Decision == "REFUTED" {
			return contradiction("MUTATION_AUTHORITY_OBSERVED")
		}
		return observationError()
	case "net-repository-changes-empty":
		if reflectRepositoryNet(fx) && value.Decision == "PASS" && value.Resolution == "EXACT" {
			return predicateResult{To: "DISCHARGED", Stage: "OBSERVE", Step: "verify-net-repository-status", Reason: "NET_REPOSITORY_CHANGES_EMPTY", Material: value.ObservedMaterialDigest}
		}
		if len(fx.NetRepositoryChanges) > 0 {
			return contradiction("NET_REPOSITORY_CHANGES_OBSERVED")
		}
		return observationError()
	}
	return predicateResult{To: claim.PriorState, Stage: "RESOLVE", Step: "reject-unknown-predicate", Reason: "PREDICATE_NOT_ALLOWED"}
}
func immutableRejection(value attempt) bool {
	return value.Decision == "DENIED" && value.Resolution == "EXACT_REJECTION" && value.APIOutcome == "REJECTED" && value.APIErrorCode == string(semantic.PatchImmutableField) && value.GraphDigestBefore != "" && value.GraphDigestBefore == value.OriginalGraphDigestAfter && value.SemanticDigestBefore == value.OriginalSemanticDigestAfter && value.ReturnedGraphDigest == ""
}
func reflectRepositoryNet(fx effects) bool {
	return fx.RepositoryStatusBefore != nil && fx.RepositoryStatusAfter != nil && len(fx.NetRepositoryChanges) == 0 && stringSliceEqual(fx.RepositoryStatusBefore, fx.RepositoryStatusAfter)
}

func changedLines(before, after []string) []string {
	left, right := map[string]struct{}{}, map[string]struct{}{}
	for _, line := range before {
		left[line] = struct{}{}
	}
	for _, line := range after {
		right[line] = struct{}{}
	}
	changed := []string{}
	for line := range left {
		if _, ok := right[line]; !ok {
			changed = append(changed, "BEFORE:"+line)
		}
	}
	for line := range right {
		if _, ok := left[line]; !ok {
			changed = append(changed, "AFTER:"+line)
		}
	}
	sort.Strings(changed)
	return changed
}
func stringSliceEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
func findAttempt(values []attempt, id string) (attempt, bool) {
	for _, value := range values {
		if value.ID == id {
			return value, true
		}
	}
	return attempt{}, false
}

type materialValue struct {
	Source   snapshot  `json:"source"`
	Attempts []attempt `json:"attempts"`
	Effects  effects   `json:"effects"`
}

func receiptMaterialDigest(source snapshot, values []attempt, fx effects) string {
	copyValues := append([]attempt(nil), values...)
	for index := range copyValues {
		copyValues[index].ObservedMaterialDigest = ""
	}
	return hashJSON(materialValue{Source: source, Attempts: copyValues, Effects: fx})
}

func buildContract(model sourceModel, source snapshot, attempts []attempt, claims []claimTransition) contract {
	classTotals, proofTotals := map[string]int{}, map[string]int{}
	for index, value := range claims {
		if index%2 == 0 {
			classTotals[value.Class]++
			proofTotals[value.ProofChoice]++
		}
	}
	return contract{Schema: schema, MetricID: metricID, GoVersion: runtime.Version(), Denominator: len(model.Claims), Classes: buckets(classTotals), Proofs: buckets(proofTotals), SourceNodes: source.NodeCount, SourceFacts: source.FactCount, ClaimCount: len(model.Claims), AttemptCount: len(attempts), ReflectiveQueries: countAttempts(attempts, func(value attempt) bool { return value.Operation == "query" }), SafeQueries: countAttempts(attempts, func(value attempt) bool {
		return value.Operation == "query" && value.Decision == "PASS" && value.Resolution == "EXACT"
	}), DeniedMutations: countAttempts(attempts, func(value attempt) bool { return value.Operation == "mutate" && value.Decision == "DENIED" }), UnknownTargets: countAttempts(attempts, func(value attempt) bool { return value.Decision == "UNKNOWN" }), RefutedAttempts: countAttempts(attempts, func(value attempt) bool { return value.Decision == "REFUTED" }), TransitionCount: len(claims), SatisfiedIndicators: countTransitions(claims, "DISCHARGED")}
}
func buckets(totals map[string]int) []bucket {
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
		if value.To == state && value.From != value.To {
			count++
		}
	}
	return count
}
func buildReceipt(value observation, satisfied, total int, sourceReconstruction, producerImports coordinates, imports importBoundary) receipt {
	classTotals, classSatisfied, proofTotals, proofSatisfied := map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}
	for index := 0; index < len(value.Claims); index += 2 {
		registration, final := value.Claims[index], value.Claims[index+1]
		classTotals[registration.Class]++
		proofTotals[registration.ProofChoice]++
		if final.To == "DISCHARGED" {
			classSatisfied[registration.Class]++
			proofSatisfied[registration.ProofChoice]++
		}
	}
	decision, reason := "PASS", "OBSERVATION_BOUNDARY_CONFORMANT"
	for _, claim := range value.Claims {
		if claim.To == "REFUTED" {
			decision, reason = "REFUTED", "BOUNDARY_VIOLATION_OBSERVED"
			break
		}
	}
	subjectResolution := "EXACT_ONLY"
	if value.Contract.UnknownTargets > 0 {
		subjectResolution = "MIXED_EXACT_AND_LOWER_RESOLUTION"
	}
	return receipt{Schema: receiptSchema, SubjectSHA: value.SubjectSHA, MetricID: metricID, Decision: decision, Resolution: "OBSERVATION_ONLY", SubjectResolution: subjectResolution, Reason: reason, Producer: value.Producer, Consumer: consumerName, Contract: value.Contract, Source: value.Source, Attempts: value.Attempts, Claims: value.Claims, Coordinates: coordinates{Satisfied: satisfied, Total: total, BasisPoints: basisPoints(satisfied, total)}, Classes: scores(classTotals, classSatisfied), Proofs: scores(proofTotals, proofSatisfied), Effects: value.Effects, SourceReconstruction: sourceReconstruction, ProducerImports: producerImports, ImportBoundary: imports, PromotionCreditBPS: 0, MutationAuthority: value.Effects.MutationAuthority, NotClaimed: []string{"generic Go reflection API equivalence", "runtime capability isolation beyond this process", "source completeness beyond declared claims", "mutation safety against a hostile process", "runtime memory and performance bounds"}}
}
func scores(totals, satisfied map[string]int) []score {
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
func reconstructionCoordinates(checks []bool) coordinates {
	satisfied := 0
	for _, check := range checks {
		if check {
			satisfied++
		}
	}
	return coordinates{Satisfied: satisfied, Total: len(checks), BasisPoints: basisPoints(satisfied, len(checks))}
}
func observationDigest(value observation) string    { value.Digest = ""; return hashJSON(value) }
func transitionDigest(value claimTransition) string { value.Digest = ""; return hashJSON(value) }
func receiptDigest(value receipt) string            { value.Digest = ""; return hashJSON(value) }
func hashJSON(value any) string {
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func bytesDigest(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	return strings.Count(string(data), "\n")
}
func fail(format string, args ...any) { fmt.Fprintf(os.Stderr, format+"\n", args...); os.Exit(1) }
