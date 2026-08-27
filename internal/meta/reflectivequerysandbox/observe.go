package reflectivequerysandbox

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/query"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func Observe(path string, subjectSHA, repositoryBeforePath, repositoryAfterPath string) (Observation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Observation{}, fmt.Errorf("read source: %w", err)
	}
	file, diagnostics := syntax.ParseFile(path, string(data))
	if diagnostics.HasErrors() {
		return Observation{}, diagnostics.Error()
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Observation{}, fmt.Errorf("lower source: %w", err)
	}
	graph, err := query.FromSemanticIR(ir)
	if err != nil {
		return Observation{}, fmt.Errorf("project query view: %w", err)
	}
	model, err := deriveSourceModel(ir)
	if err != nil {
		return Observation{}, fmt.Errorf("derive source contract: %w", err)
	}

	semanticDigest := ir.StableHash()
	source := Snapshot{
		Path: SourcePath, SourceDigest: semantic.StableHash(data), SemanticDigest: semanticDigest,
		GraphDigest: graph.StableHash(), NodeCount: len(graph.Nodes()), FactCount: len(graph.AllFacts()), GoooLines: countLines(data),
	}
	effects, err := observeEffects(repositoryBeforePath, repositoryAfterPath)
	if err != nil {
		return Observation{}, err
	}
	attempts, mutationAuthority, mutationAPI, mutationOutcome, mutationError, err := buildAttempts(ir, graph, model, semanticDigest, source.GraphDigest, effects)
	if err != nil {
		return Observation{}, err
	}
	effects.MutationAuthority = mutationAuthority
	effects.MutationAPI = mutationAPI
	effects.MutationOutcome = mutationOutcome
	effects.MutationError = mutationError
	receiptMaterial := receiptMaterialDigest(source, attempts, effects)
	for index := range attempts {
		if attempts[index].ID == "receipt.seal" {
			attempts[index].ObservedMaterialDigest = receiptMaterial
		}
	}
	claims := buildClaimTransitions(model.Claims, attempts, effects, receiptMaterial)
	return Observation{
		Schema: Schema, SubjectSHA: subjectSHA, Contract: buildContract(model, source, attempts, claims),
		Source: source, Attempts: attempts, Claims: claims, Effects: effects, ReceiptMaterialDigest: receiptMaterial, Producer: ProducerName,
	}, nil
}

func buildAttempts(ir semantic.IR, graph *query.Graph, model sourceModel, semanticDigest, graphDigest string, effects Effects) ([]Attempt, bool, string, string, string, error) {
	attempts := make([]Attempt, 0, len(model.Operations)+1)
	var mutationAuthority bool
	var mutationAPI, mutationOutcome, mutationError string
	for _, operation := range model.Operations {
		target, err := targetForOperation(ir, operation, model)
		if err != nil {
			return nil, false, "", "", "", err
		}
		attemptID := attemptIDForProgram(operation.Program)
		switch {
		case strings.HasPrefix(operation.Program, "reflect.query:"):
			attempts = append(attempts, exactAttempt(graph, operation, attemptID, target, semanticDigest, graphDigest, "QUERY", model.Claims))
		case operation.Program == "reflect.observation:repository-net":
			attempts = append(attempts, repositoryAttempt(operation, attemptID, target, semanticDigest, effects, model.Claims))
		case strings.HasPrefix(operation.Program, "reflect.observation:"):
			attempts = append(attempts, exactAttempt(graph, operation, attemptID, target, semanticDigest, graphDigest, "RECEIPT", model.Claims))
		case strings.HasPrefix(operation.Program, "reflect.attempt:"):
			attempt, authority, api, outcome, apiError, err := mutationAttempt(ir, operation, attemptID, target, semanticDigest, graphDigest, model, model.Claims)
			if err != nil {
				return nil, false, "", "", "", err
			}
			attempts = append(attempts, attempt)
			mutationAuthority, mutationAPI, mutationOutcome, mutationError = authority, api, outcome, apiError
		}
	}
	metricOperation, ok := findOperation(model.Operations, "reflect.query:metrics")
	if !ok {
		return nil, false, "", "", "", fmt.Errorf("source declares no metrics query operation")
	}
	unknown := unknownAttempt(graph, metricOperation, model.UnknownTarget, semanticDigest, graphDigest, model.Claims)
	attempts = append(attempts, unknown)
	for index := range attempts {
		for _, claim := range model.Claims {
			if claim.EvidenceAttempt == attempts[index].ID {
				attempts[index].EvidenceClaimIDs = append(attempts[index].EvidenceClaimIDs, claim.ID)
			}
		}
	}
	sort.SliceStable(attempts, func(i, j int) bool { return attempts[i].ID < attempts[j].ID })
	return attempts, mutationAuthority, mutationAPI, mutationOutcome, mutationError, nil
}

func exactAttempt(graph *query.Graph, operation operationSpec, id string, target semantic.ID, semanticDigest, graphDigest, stage string, claims []claimSpec) Attempt {
	before := graph.StableHash()
	attempt := Attempt{
		ID: id, Class: classForAttempt(id, claims), Operation: "query", Root: operation.ID.String(), Relation: "used", Target: target.String(),
		MetaOperation: metaForAttempt(id, operation.Program, claims), Producer: ProducerName, Consumer: ConsumerName,
		ProofChoice: proofForAttempt(id, claims), Stage: stage, Step: "match-source-relation",
		SemanticDigestBefore: semanticDigest, GraphDigestBefore: before,
	}
	result, err := graph.ExactMatch(query.NewExactQuery(query.ID(operation.ID.String()), query.Used, query.ID(target.String())))
	after := graph.StableHash()
	attempt.SemanticDigestAfter, attempt.GraphDigestAfter = semanticDigest, after
	attempt.ObservedMaterialDigest = after
	if err != nil {
		attempt.Decision, attempt.Resolution, attempt.Reason = "UNKNOWN", "LOWER_RESOLUTION", "QUERY_API_ERROR"
		return attempt
	}
	attempt.MatchedFacts = len(result.All())
	if attempt.MatchedFacts == 1 {
		attempt.Decision, attempt.Resolution, attempt.Reason = "PASS", "EXACT", "EXACT_RELATION_MATCH"
	} else {
		attempt.Decision, attempt.Resolution, attempt.Reason = "UNKNOWN", "LOWER_RESOLUTION", "RELATION_NOT_OBSERVED"
	}
	return attempt
}

func mutationAttempt(ir semantic.IR, operation operationSpec, id string, target semantic.ID, semanticDigest, graphDigest string, model sourceModel, claims []claimSpec) (Attempt, bool, string, string, string, error) {
	node, ok := ir.Graph.Node(target)
	if !ok {
		return Attempt{}, false, "", "", "", fmt.Errorf("mutation target %q disappeared from semantic IR", target)
	}
	field := tail(model.MutationField)
	payload := tail(model.MutationPayload)
	intent := tail(model.MutationIntent)
	locality := tail(model.MutationLocality)
	fieldHash := ""
	if field != "id" {
		var fieldErr error
		fieldHash, fieldErr = semantic.NodeFieldHash(node, field)
		if fieldErr != nil {
			return Attempt{}, false, "", "", "", fmt.Errorf("derive mutation field: %w", fieldErr)
		}
	}
	beforeSemantic := ir.StableHash()
	beforeGraph := ir.Graph.StableHash()
	request := semantic.GraphPatchRequest{
		SchemaVersion: semantic.GraphPatchSchemaVersion, Operation: semantic.GraphPatchSetNodeField,
		ExpectedGraphHash: beforeGraph, NodeID: node.ID, ExpectedNodeHash: node.StableHash(), Field: field, ExpectedFieldHash: fieldHash,
		ExpectedSourceDigest: semanticDigest, ExpectedIRDigest: semanticDigest,
		AllowedIntent: intent, Locality: locality,
	}
	base := semantic.GraphPatchBase{SourceDigest: semanticDigest, IRDigest: semanticDigest}
	patched, err := ir.Graph.ApplyGraphPatch(base, request, semantic.GraphPatchMutation{Name: payload})
	afterSemantic := ir.StableHash()
	afterGraph := ir.Graph.StableHash()
	api := "semantic.Graph.ApplyGraphPatch"
	attempt := Attempt{
		ID: id, Class: classForAttempt(id, claims), Operation: "mutate", Root: operation.ID.String(), Relation: "set", Target: target.String(),
		MetaOperation: metaForAttempt(id, operation.Program, claims), Producer: ProducerName, Consumer: ConsumerName,
		ProofChoice: proofForAttempt(id, claims), Stage: "MUTATION_BOUNDARY", Step: "apply-typed-request",
		API: api, SemanticDigestBefore: beforeSemantic, SemanticDigestAfter: afterSemantic,
		GraphDigestBefore: beforeGraph, GraphDigestAfter: afterGraph,
		OriginalSemanticDigestAfter: afterSemantic, OriginalGraphDigestAfter: afterGraph,
	}
	if err != nil {
		attempt.APIErrorCode = mutationErrorCode(err)
		attempt.APIOutcome, attempt.APIError = "REJECTED", err.Error()
		if afterSemantic != beforeSemantic || afterGraph != beforeGraph {
			attempt.Decision, attempt.Resolution, attempt.Reason = "REFUTED", "EXACT", "MUTATION_API_CHANGED_OR_PARTIALLY_MUTATED"
			return attempt, true, api, "REJECTED", err.Error(), nil
		}
		var conflict semantic.GraphPatchConflict
		if errors.As(err, &conflict) && conflict.Code == semantic.PatchImmutableField && conflict.Detail == field {
			attempt.Decision, attempt.Resolution, attempt.Reason = "DENIED", "EXACT_REJECTION", "IMMUTABLE_FIELD_REJECTED"
			return attempt, false, api, "REJECTED", err.Error(), nil
		}
		attempt.Decision, attempt.Resolution, attempt.Reason = "UNKNOWN", "LOWER_RESOLUTION", "MUTATION_API_ERROR_"+strings.ToUpper(string(attempt.APIErrorCode))
		return attempt, false, api, "ERROR", err.Error(), nil
	}
	patchedIR := ir
	patchedIR.Graph = patched
	attempt.SemanticDigestAfter, attempt.GraphDigestAfter = patchedIR.StableHash(), patched.StableHash()
	attempt.ReturnedSemanticDigest, attempt.ReturnedGraphDigest = attempt.SemanticDigestAfter, attempt.GraphDigestAfter
	attempt.Decision, attempt.Resolution, attempt.Reason = "REFUTED", "EXACT", "MUTATION_CAPABILITY_ACCEPTED"
	attempt.APIOutcome = "ACCEPTED"
	return attempt, true, api, "ACCEPTED", "", nil
}

func tail(id semantic.ID) string {
	parts := strings.Split(strings.TrimSuffix(id.String(), "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func mutationErrorCode(err error) string {
	var conflict semantic.GraphPatchConflict
	if errors.As(err, &conflict) {
		return string(conflict.Code)
	}
	return "unknown"
}

func unknownAttempt(graph *query.Graph, operation operationSpec, target query.ID, semanticDigest, graphDigest string, claims []claimSpec) Attempt {
	before := graph.StableHash()
	attempt := Attempt{
		ID: "unknown.target", Class: classForAttempt("unknown.target", claims), Operation: "query", Root: operation.ID.String(), Relation: "used", Target: target.String(),
		MetaOperation: metaForAttempt("unknown.target", operation.Program, claims), Producer: ProducerName, Consumer: ConsumerName,
		ProofChoice: proofForAttempt("unknown.target", claims), Stage: "UNKNOWN", Step: "resolve-unknown-subject", SemanticDigestBefore: semanticDigest, GraphDigestBefore: before,
	}
	_, err := graph.ExactMatch(query.NewExactQuery(query.ID(operation.ID.String()), query.Used, target))
	attempt.SemanticDigestAfter, attempt.GraphDigestAfter = semanticDigest, graph.StableHash()
	attempt.ObservedMaterialDigest = attempt.GraphDigestAfter
	if err != nil && errors.Is(err, query.ErrUnknownEndpoint) {
		attempt.Decision, attempt.Resolution, attempt.Reason = "UNKNOWN", "LOWER_RESOLUTION", "UNKNOWN_TARGET"
	} else if err != nil {
		attempt.Decision, attempt.Resolution, attempt.Reason = "UNKNOWN", "LOWER_RESOLUTION", "QUERY_API_ERROR"
	} else {
		attempt.Decision, attempt.Resolution, attempt.Reason = "REFUTED", "EXACT", "UNKNOWN_TARGET_BECAME_KNOWN"
	}
	return attempt
}

func proofForAttempt(id string, claims []claimSpec) string {
	for _, claim := range claims {
		if claim.EvidenceAttempt == id {
			return claim.ProofChoice
		}
	}
	return "SOURCE_DERIVED"
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

func observeEffects(beforePath, afterPath string) (Effects, error) {
	before, err := readStatusLines(beforePath)
	if err != nil {
		return Effects{}, err
	}
	after, err := readStatusLines(afterPath)
	if err != nil {
		return Effects{}, err
	}
	return Effects{RepositoryStatusBefore: before, RepositoryStatusAfter: after, NetRepositoryChanges: changedLines(before, after)}, nil
}

func repositoryAttempt(operation operationSpec, id string, target semantic.ID, semanticDigest string, effects Effects, claims []claimSpec) Attempt {
	material := semantic.StableHashString(strings.Join(effects.RepositoryStatusBefore, "\n") + "\x00" + strings.Join(effects.RepositoryStatusAfter, "\n"))
	value := Attempt{ID: id, Class: classForAttempt(id, claims), Operation: "repository", Root: operation.ID.String(), Relation: "net", Target: target.String(), MetaOperation: metaForAttempt(id, operation.Program, claims), Producer: ProducerName, Consumer: ConsumerName, ProofChoice: proofForAttempt(id, claims), Stage: "REPOSITORY", Step: "compare-status-snapshots", SemanticDigestBefore: semanticDigest, SemanticDigestAfter: semanticDigest, GraphDigestBefore: material, GraphDigestAfter: material, ObservedMaterialDigest: material}
	if len(effects.NetRepositoryChanges) == 0 && len(effects.RepositoryStatusBefore) == len(effects.RepositoryStatusAfter) {
		value.Decision, value.Resolution, value.Reason = "PASS", "EXACT", "NET_REPOSITORY_CHANGES_EMPTY"
	} else {
		value.Decision, value.Resolution, value.Reason = "REFUTED", "EXACT", "NET_REPOSITORY_CHANGES_OBSERVED"
	}
	return value
}

func readStatusLines(path string) ([]string, error) {
	if path == "" {
		return []string{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read repository status %q: %w", path, err)
	}
	text := strings.TrimSuffix(string(data), "\n")
	if text == "" {
		return []string{}, nil
	}
	return strings.Split(text, "\n"), nil
}

func changedLines(before, after []string) []string {
	left, right := make(map[string]struct{}), make(map[string]struct{})
	for _, line := range before {
		left[line] = struct{}{}
	}
	for _, line := range after {
		right[line] = struct{}{}
	}
	changed := make([]string, 0)
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

func countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	return strings.Count(string(data), "\n")
}
